use std::sync::atomic::Ordering;
use std::sync::Arc;
use std::time::Instant;

use tokio::sync::Mutex as TokioMutex;
use tonic::{Request, Response, Status};
use tracing::{info, warn};
use uuid::Uuid;

use guardrails_eval::{EvalContext, Verdict};

use proto_gen::platform::activity::v1::activity_service_client::ActivityServiceClient;
use proto_gen::platform::activity::v1::{ActionOutcome, ActionRecord, RecordActionRequest};
use proto_gen::platform::economics::v1::economics_service_client::EconomicsServiceClient;
use proto_gen::platform::economics::v1::{CheckBudgetRequest, RecordUsageRequest, UsageRecord};
use proto_gen::platform::governance::v1::data_governance_service_client::DataGovernanceServiceClient;
use proto_gen::platform::governance::v1::InspectEgressRequest;
use proto_gen::platform::human::v1::human_interaction_service_client::HumanInteractionServiceClient;
use proto_gen::platform::human::v1::{
    CreateHumanRequestRequest, GetHumanRequestRequest, HumanRequestStatus,
};
use proto_gen::platform::host_agent::v1::host_agent_api_service_server::HostAgentApiService;
use proto_gen::platform::host_agent::v1::{
    ActionVerdict, CheckHumanRequestRequest, CheckHumanRequestResponse, ExecuteToolRequest,
    ExecuteToolResponse, ListToolsRequest, ListToolsResponse, ReportActionResultRequest,
    ReportActionResultResponse, ReportProgressRequest, ReportProgressResponse,
    RequestHumanInputRequest, RequestHumanInputResponse,
};

use crate::sandbox::{SandboxEvent, SandboxManager};

/// gRPC implementation of the HostAgentAPIService — the agent-facing API exposed
/// inside each sandbox.
///
/// This is a **policy-only** engine: ExecuteTool evaluates guardrails and budget
/// but does NOT execute tools. The agent executes tools locally inside its
/// container, then calls ReportActionResult to record the outcome for auditing.
pub struct HostAgentApiServiceImpl {
    sandbox_manager: SandboxManager,
    his_client: Option<TokioMutex<HumanInteractionServiceClient<tonic::transport::Channel>>>,
    activity_client: Option<Arc<TokioMutex<ActivityServiceClient<tonic::transport::Channel>>>>,
    economics_client: Option<Arc<TokioMutex<EconomicsServiceClient<tonic::transport::Channel>>>>,
    governance_client: Option<Arc<TokioMutex<DataGovernanceServiceClient<tonic::transport::Channel>>>>,
}

impl std::fmt::Debug for HostAgentApiServiceImpl {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("HostAgentApiServiceImpl")
            .field("sandbox_manager", &self.sandbox_manager)
            .field("his_configured", &self.his_client.is_some())
            .field("activity_configured", &self.activity_client.is_some())
            .field("economics_configured", &self.economics_client.is_some())
            .field("governance_configured", &self.governance_client.is_some())
            .finish()
    }
}

impl HostAgentApiServiceImpl {
    pub fn new(
        sandbox_manager: SandboxManager,
        his_client: Option<HumanInteractionServiceClient<tonic::transport::Channel>>,
        activity_client: Option<ActivityServiceClient<tonic::transport::Channel>>,
        economics_client: Option<EconomicsServiceClient<tonic::transport::Channel>>,
        governance_client: Option<DataGovernanceServiceClient<tonic::transport::Channel>>,
    ) -> Self {
        Self {
            sandbox_manager,
            his_client: his_client.map(TokioMutex::new),
            activity_client: activity_client.map(|c| Arc::new(TokioMutex::new(c))),
            economics_client: economics_client.map(|c| Arc::new(TokioMutex::new(c))),
            governance_client: governance_client.map(|c| Arc::new(TokioMutex::new(c))),
        }
    }
}

#[tonic::async_trait]
impl HostAgentApiService for HostAgentApiServiceImpl {
    /// Evaluate guardrails and budget for a tool call. Returns a verdict
    /// (ALLOW/DENY/ESCALATE) and an action_id. The agent is responsible for
    /// executing the tool locally if allowed, then calling ReportActionResult.
    async fn execute_tool(
        &self,
        request: Request<ExecuteToolRequest>,
    ) -> Result<Response<ExecuteToolResponse>, Status> {
        // 1. Extract sandbox from metadata
        let sandbox = self
            .sandbox_manager
            .lookup_by_metadata(&request)
            .map_err(|e| Status::unauthenticated(e.to_string()))?;

        let req = request.into_inner();
        let action_id = Uuid::new_v4().to_string();

        info!(
            tool_name = %req.tool_name,
            sandbox_id = %sandbox.id,
            action_id = %action_id,
            justification = %req.justification,
            "agent requested tool evaluation"
        );

        // 2. Convert proto Struct parameters to serde_json::Value
        let proto_parameters = req.parameters.clone();
        let parameters = req
            .parameters
            .map(|s| {
                let json_bytes = serde_json::to_vec(&proto_struct_to_value(&s))
                    .unwrap_or_default();
                serde_json::from_slice(&json_bytes).unwrap_or(serde_json::Value::Null)
            })
            .unwrap_or(serde_json::Value::Null);

        // 3. Budget check — deny or escalate based on on_exceeded policy
        if let Some(economics_mutex) = &self.economics_client {
            let check_result = {
                let mut client = economics_mutex.lock().await;
                client
                    .check_budget(CheckBudgetRequest {
                        agent_id: sandbox.agent_id.clone(),
                        estimated_cost: 0.0,
                    })
                    .await
            };
            match check_result {
                Ok(resp) => {
                    let budget_resp = resp.into_inner();
                    if budget_resp.warning {
                        warn!(
                            agent_id = %sandbox.agent_id,
                            remaining = %budget_resp.remaining,
                            "budget warning threshold reached"
                        );
                    }
                    if !budget_resp.allowed {
                        let enforcement = budget_resp.enforcement_action.as_str();
                        match enforcement {
                            "request_increase" => {
                                warn!(
                                    agent_id = %sandbox.agent_id,
                                    tool_name = %req.tool_name,
                                    "budget exhausted — escalating for budget increase"
                                );
                                return Ok(Response::new(ExecuteToolResponse {
                                    verdict: ActionVerdict::Escalate as i32,
                                    result: None,
                                    denial_reason: String::new(),
                                    escalation_id: "budget_increase_required".to_string(),
                                    action_id: action_id.clone(),
                                }));
                            }
                            _ => {
                                // "halt" or unknown — hard deny
                                warn!(
                                    agent_id = %sandbox.agent_id,
                                    tool_name = %req.tool_name,
                                    "budget exhausted — denying tool execution"
                                );
                                return Ok(Response::new(ExecuteToolResponse {
                                    verdict: ActionVerdict::Deny as i32,
                                    result: None,
                                    denial_reason: "budget exhausted".to_string(),
                                    escalation_id: String::new(),
                                    action_id: action_id.clone(),
                                }));
                            }
                        }
                    }
                }
                Err(e) => {
                    warn!(error = %e, "budget check failed — allowing execution");
                }
            }
        }

        // 4. Build evaluation context
        let eval_ctx = EvalContext {
            tool_name: req.tool_name.clone(),
            parameters: parameters.clone(),
            agent_id: sandbox.agent_id.clone(),
            trust_level: String::new(),         // TODO: populate from agent metadata
            data_classification: String::new(), // TODO: populate from governance
        };

        // 5. Evaluate guardrails (read lock for hot-reload support)
        let eval_start = Instant::now();
        let verdict = {
            let evaluator = sandbox
                .evaluator
                .read()
                .map_err(|e| Status::internal(format!("evaluator lock poisoned: {e}")))?;
            evaluator.evaluate(&eval_ctx)
        };
        let eval_latency_us = eval_start.elapsed().as_micros() as i64;

        // 6. Map verdict to response fields (NO tool execution — policy only)
        let (action_verdict, denial_reason, escalation_id) = match &verdict {
            Verdict::Allow => (ActionVerdict::Allow, String::new(), String::new()),
            Verdict::Deny(reason) => (ActionVerdict::Deny, reason.clone(), String::new()),
            Verdict::Escalate(rule_id) => {
                (ActionVerdict::Escalate, String::new(), rule_id.clone())
            }
        };

        // 6b. DLP egress inspection — for allowed tool calls with destination/url parameters
        let (action_verdict, denial_reason, escalation_id) = if matches!(action_verdict, ActionVerdict::Allow) {
            if let Some(governance_mutex) = &self.governance_client {
                // Extract destination from parameters (check common field names).
                let destination = parameters.get("destination")
                    .or_else(|| parameters.get("url"))
                    .or_else(|| parameters.get("endpoint"))
                    .and_then(|v| v.as_str())
                    .unwrap_or_default();

                if !destination.is_empty() {
                    let content = parameters.get("content")
                        .or_else(|| parameters.get("body"))
                        .or_else(|| parameters.get("data"))
                        .map(|v| v.to_string())
                        .unwrap_or_default();

                    let inspect_result = {
                        let mut client = governance_mutex.lock().await;
                        client.inspect_egress(InspectEgressRequest {
                            agent_id: sandbox.agent_id.clone(),
                            destination: destination.to_string(),
                            content: content.into_bytes(),
                            content_type: "application/json".to_string(),
                        }).await
                    };

                    match inspect_result {
                        Ok(resp) => {
                            let egress_resp = resp.into_inner();
                            if !egress_resp.allowed {
                                warn!(
                                    agent_id = %sandbox.agent_id,
                                    tool_name = %req.tool_name,
                                    destination = %destination,
                                    reason = %egress_resp.reason,
                                    "DLP egress inspection denied"
                                );
                                (ActionVerdict::Deny, format!("DLP denied: {}", egress_resp.reason), String::new())
                            } else {
                                (action_verdict, denial_reason, escalation_id)
                            }
                        }
                        Err(e) => {
                            warn!(error = %e, "DLP egress inspection failed — allowing execution");
                            (action_verdict, denial_reason, escalation_id)
                        }
                    }
                } else {
                    (action_verdict, denial_reason, escalation_id)
                }
            } else {
                (action_verdict, denial_reason, escalation_id)
            }
        } else {
            (action_verdict, denial_reason, escalation_id)
        };

        // 7. Increment actions counter
        sandbox.actions_executed.fetch_add(1, Ordering::Relaxed);

        // 8. Emit action event on sandbox event channel
        let verdict_str = match &verdict {
            Verdict::Allow => "allow",
            Verdict::Deny(_) => "deny",
            Verdict::Escalate(_) => "escalate",
        };
        let _ = sandbox.event_tx.send(SandboxEvent::Action {
            action_id: action_id.clone(),
            tool_name: req.tool_name.clone(),
            verdict: verdict_str.to_string(),
            evaluation_latency_us: eval_latency_us,
        });

        info!(
            tool_name = %req.tool_name,
            action_id = %action_id,
            verdict = ?action_verdict,
            eval_latency_us = %eval_latency_us,
            "tool evaluation complete"
        );

        // 9. Fire-and-forget: persist evaluation record to Activity Store
        if let Some(activity_mutex) = &self.activity_client {
            let outcome = match &verdict {
                Verdict::Allow => ActionOutcome::Allowed as i32,
                Verdict::Deny(_) => ActionOutcome::Denied as i32,
                Verdict::Escalate(_) => ActionOutcome::Escalated as i32,
            };
            let guardrail_rule_id = match &verdict {
                Verdict::Escalate(rule_id) => rule_id.clone(),
                _ => String::new(),
            };
            let record = ActionRecord {
                record_id: action_id.clone(),
                workspace_id: sandbox.workspace_id.clone(),
                agent_id: sandbox.agent_id.clone(),
                task_id: String::new(),
                tool_name: req.tool_name.clone(),
                parameters: proto_parameters,
                result: None, // result comes later via ReportActionResult
                outcome,
                guardrail_rule_id,
                denial_reason: denial_reason.clone(),
                evaluation_latency_us: eval_latency_us,
                execution_latency_us: 0,
                recorded_at: None,
                tenant_id: String::new(),
            };
            let activity_mutex = activity_mutex.clone();
            tokio::spawn(async move {
                let mut client = activity_mutex.lock().await;
                if let Err(e) = client
                    .record_action(RecordActionRequest {
                        record: Some(record),
                    })
                    .await
                {
                    tracing::warn!(error = %e, "failed to persist action record to Activity Store");
                }
            });
        }

        // 10. Fire-and-forget: record tool usage to Economics Service
        if let Some(economics_mutex) = &self.economics_client {
            if matches!(verdict, Verdict::Allow) {
                let economics_mutex = economics_mutex.clone();
                let agent_id = sandbox.agent_id.clone();
                let workspace_id = sandbox.workspace_id.clone();
                let tool_name = req.tool_name.clone();
                tokio::spawn(async move {
                    let mut client = economics_mutex.lock().await;
                    let record = UsageRecord {
                        record_id: String::new(),
                        agent_id,
                        workspace_id,
                        resource_type: format!("tool:{tool_name}"),
                        quantity: 1.0,
                        unit: "invocations".to_string(),
                        cost: 0.0,
                        recorded_at: None,
                        tenant_id: String::new(),
                    };
                    if let Err(e) = client
                        .record_usage(RecordUsageRequest {
                            record: Some(record),
                        })
                        .await
                    {
                        tracing::warn!(error = %e, "failed to record usage to Economics Service");
                    }
                });
            }
        }

        Ok(Response::new(ExecuteToolResponse {
            verdict: action_verdict as i32,
            result: None, // policy-only: no tool result
            denial_reason,
            escalation_id,
            action_id,
        }))
    }

    /// Records the outcome of an agent-executed tool call for the audit trail.
    /// Called by the agent after executing a tool that was allowed by ExecuteTool.
    async fn report_action_result(
        &self,
        request: Request<ReportActionResultRequest>,
    ) -> Result<Response<ReportActionResultResponse>, Status> {
        let sandbox = self
            .sandbox_manager
            .lookup_by_metadata(&request)
            .map_err(|e| Status::unauthenticated(e.to_string()))?;

        let req = request.into_inner();
        info!(
            action_id = %req.action_id,
            sandbox_id = %sandbox.id,
            success = %req.success,
            "agent reported action result"
        );

        // Fire-and-forget: update the action record in Activity Store with the result
        if let Some(activity_mutex) = &self.activity_client {
            let record = ActionRecord {
                record_id: req.action_id.clone(),
                workspace_id: sandbox.workspace_id.clone(),
                agent_id: sandbox.agent_id.clone(),
                task_id: String::new(),
                tool_name: String::new(), // already recorded in ExecuteTool
                parameters: None,
                result: req.result,
                outcome: if req.success {
                    ActionOutcome::Allowed as i32
                } else {
                    ActionOutcome::Error as i32
                },
                guardrail_rule_id: String::new(),
                denial_reason: req.error_message,
                evaluation_latency_us: 0,
                execution_latency_us: 0,
                recorded_at: None,
                tenant_id: String::new(),
            };
            let activity_mutex = activity_mutex.clone();
            tokio::spawn(async move {
                let mut client = activity_mutex.lock().await;
                if let Err(e) = client
                    .record_action(RecordActionRequest {
                        record: Some(record),
                    })
                    .await
                {
                    tracing::warn!(error = %e, "failed to persist action result to Activity Store");
                }
            });
        }

        Ok(Response::new(ReportActionResultResponse {}))
    }

    async fn request_human_input(
        &self,
        request: Request<RequestHumanInputRequest>,
    ) -> Result<Response<RequestHumanInputResponse>, Status> {
        let sandbox = self
            .sandbox_manager
            .lookup_by_metadata(&request)
            .map_err(|e| Status::unauthenticated(e.to_string()))?;

        let his_mutex = self
            .his_client
            .as_ref()
            .ok_or_else(|| Status::unavailable("HIS not configured"))?;

        let req = request.into_inner();
        info!(
            sandbox_id = %sandbox.id,
            question = %req.question,
            "agent requesting human input"
        );

        let timeout_seconds = if req.timeout_seconds > 0 {
            req.timeout_seconds
        } else {
            300
        };

        let create_req = CreateHumanRequestRequest {
            workspace_id: sandbox.workspace_id.clone(),
            agent_id: sandbox.agent_id.clone(),
            question: req.question,
            options: req.options,
            context: req.context,
            timeout_seconds,
            r#type: 0,
            urgency: 0,
            task_id: String::new(),
        };

        let create_resp = {
            let mut client = his_mutex.lock().await;
            client
                .create_request(create_req)
                .await
                .map_err(|e| Status::internal(format!("HIS create_request failed: {e}")))?
        };

        let human_request = create_resp
            .into_inner()
            .request
            .ok_or_else(|| Status::internal("HIS returned empty request"))?;

        Ok(Response::new(RequestHumanInputResponse {
            request_id: human_request.request_id,
            response: String::new(),
            responder_id: String::new(),
            timed_out: false,
        }))
    }

    async fn check_human_request(
        &self,
        request: Request<CheckHumanRequestRequest>,
    ) -> Result<Response<CheckHumanRequestResponse>, Status> {
        let his_mutex = self
            .his_client
            .as_ref()
            .ok_or_else(|| Status::unavailable("HIS not configured"))?;

        let req = request.into_inner();
        info!(request_id = %req.request_id, "checking human request status");

        let get_resp = {
            let mut client = his_mutex.lock().await;
            client
                .get_request(GetHumanRequestRequest {
                    request_id: req.request_id,
                })
                .await
                .map_err(|e| Status::internal(format!("HIS get_request failed: {e}")))?
        };

        let hr = get_resp
            .into_inner()
            .request
            .ok_or_else(|| Status::not_found("human request not found"))?;

        let status_str = match HumanRequestStatus::try_from(hr.status)
            .unwrap_or(HumanRequestStatus::Unspecified)
        {
            HumanRequestStatus::Pending => "pending",
            HumanRequestStatus::Responded => "responded",
            HumanRequestStatus::Expired => "expired",
            HumanRequestStatus::Cancelled => "expired",
            _ => "pending",
        };

        Ok(Response::new(CheckHumanRequestResponse {
            status: status_str.to_string(),
            response: hr.response,
            responder_id: hr.responder_id,
        }))
    }

    async fn report_progress(
        &self,
        request: Request<ReportProgressRequest>,
    ) -> Result<Response<ReportProgressResponse>, Status> {
        let sandbox = self.sandbox_manager.lookup_by_metadata(&request).ok();

        let req = request.into_inner();
        info!(
            message = %req.message,
            percent_complete = %req.percent_complete,
            "agent reported progress"
        );

        if let Some(sandbox) = sandbox {
            let _ = sandbox.event_tx.send(SandboxEvent::Progress {
                message: req.message,
                percent_complete: req.percent_complete,
            });
        }

        Ok(Response::new(ReportProgressResponse {}))
    }

    /// Returns the tool definitions registered for this sandbox.
    async fn list_tools(
        &self,
        request: Request<ListToolsRequest>,
    ) -> Result<Response<ListToolsResponse>, Status> {
        let sandbox = self
            .sandbox_manager
            .lookup_by_metadata(&request)
            .map_err(|e| Status::unauthenticated(e.to_string()))?;

        info!(
            sandbox_id = %sandbox.id,
            tool_count = %sandbox.tool_definitions.len(),
            "listing tools for sandbox"
        );

        Ok(Response::new(ListToolsResponse {
            tools: sandbox.tool_definitions.clone(),
        }))
    }
}

/// Convert a prost_types::Struct into a serde_json::Value.
fn proto_struct_to_value(s: &prost_types::Struct) -> serde_json::Value {
    let map: serde_json::Map<String, serde_json::Value> = s
        .fields
        .iter()
        .map(|(k, v)| (k.clone(), proto_value_to_json(v)))
        .collect();
    serde_json::Value::Object(map)
}

fn proto_value_to_json(v: &prost_types::Value) -> serde_json::Value {
    use prost_types::value::Kind;
    match &v.kind {
        Some(Kind::NullValue(_)) => serde_json::Value::Null,
        Some(Kind::NumberValue(n)) => serde_json::json!(n),
        Some(Kind::StringValue(s)) => serde_json::Value::String(s.clone()),
        Some(Kind::BoolValue(b)) => serde_json::Value::Bool(*b),
        Some(Kind::StructValue(s)) => proto_struct_to_value(s),
        Some(Kind::ListValue(l)) => {
            serde_json::Value::Array(l.values.iter().map(proto_value_to_json).collect())
        }
        None => serde_json::Value::Null,
    }
}

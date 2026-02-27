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
use proto_gen::platform::human::v1::human_interaction_service_client::HumanInteractionServiceClient;
use proto_gen::platform::human::v1::{
    CreateHumanRequestRequest, GetHumanRequestRequest, HumanRequestStatus,
};
use proto_gen::platform::runtime::v1::agent_api_service_server::AgentApiService;
use proto_gen::platform::runtime::v1::{
    ActionVerdict, CheckHumanRequestRequest, CheckHumanRequestResponse, ExecuteToolRequest,
    ExecuteToolResponse, ReportProgressRequest, ReportProgressResponse,
    RequestHumanInputRequest, RequestHumanInputResponse,
};

use crate::sandbox::{SandboxEvent, SandboxManager};
use crate::tools::{ToolInterceptor, ToolRequest};

/// gRPC implementation of the AgentAPIService — the agent-facing API exposed
/// inside each sandbox.
pub struct AgentApiServiceImpl {
    sandbox_manager: SandboxManager,
    tool_interceptor: ToolInterceptor,
    his_client: Option<TokioMutex<HumanInteractionServiceClient<tonic::transport::Channel>>>,
    activity_client: Option<Arc<TokioMutex<ActivityServiceClient<tonic::transport::Channel>>>>,
    economics_client: Option<Arc<TokioMutex<EconomicsServiceClient<tonic::transport::Channel>>>>,
}

impl std::fmt::Debug for AgentApiServiceImpl {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("AgentApiServiceImpl")
            .field("sandbox_manager", &self.sandbox_manager)
            .field("tool_interceptor", &self.tool_interceptor)
            .field("his_configured", &self.his_client.is_some())
            .field("activity_configured", &self.activity_client.is_some())
            .field("economics_configured", &self.economics_client.is_some())
            .finish()
    }
}

impl AgentApiServiceImpl {
    pub fn new(
        sandbox_manager: SandboxManager,
        tool_interceptor: ToolInterceptor,
        his_client: Option<HumanInteractionServiceClient<tonic::transport::Channel>>,
        activity_client: Option<ActivityServiceClient<tonic::transport::Channel>>,
        economics_client: Option<EconomicsServiceClient<tonic::transport::Channel>>,
    ) -> Self {
        Self {
            sandbox_manager,
            tool_interceptor,
            his_client: his_client.map(TokioMutex::new),
            activity_client: activity_client.map(|c| Arc::new(TokioMutex::new(c))),
            economics_client: economics_client.map(|c| Arc::new(TokioMutex::new(c))),
        }
    }
}

#[tonic::async_trait]
impl AgentApiService for AgentApiServiceImpl {
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
        info!(
            tool_name = %req.tool_name,
            sandbox_id = %sandbox.id,
            justification = %req.justification,
            "agent requested tool execution"
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

        // 3. Budget check — deny if agent has exhausted budget, warn-and-allow on RPC failure
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
                    if !resp.into_inner().allowed {
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
                        }));
                    }
                }
                Err(e) => {
                    warn!(error = %e, "budget check failed — allowing execution");
                }
            }
        }

        // 4. Build evaluation context with real agent_id from sandbox state
        let eval_ctx = EvalContext {
            tool_name: req.tool_name.clone(),
            parameters: parameters.clone(),
            agent_id: sandbox.agent_id.clone(),
        };

        // 5. Evaluate guardrails using the sandbox's evaluator (read lock for hot-reload support)
        let eval_start = Instant::now();
        let evaluator = sandbox
            .evaluator
            .read()
            .map_err(|e| Status::internal(format!("evaluator lock poisoned: {e}")))?;
        let verdict = evaluator.evaluate(&eval_ctx);
        drop(evaluator); // release read lock before tool execution
        let eval_latency_us = eval_start.elapsed().as_micros() as i64;

        let (action_verdict, denial_reason, escalation_id, tool_result) = match &verdict {
            Verdict::Allow => {
                // 6. Execute tool through interceptor
                let tool_req = ToolRequest {
                    tool_name: req.tool_name.clone(),
                    parameters,
                };
                match self.tool_interceptor.intercept(
                    tool_req,
                    &sandbox.allowed_tools,
                    &sandbox.env_vars,
                ) {
                    Ok(result) => {
                        let proto_result = json_to_proto_struct(&result.output);
                        (ActionVerdict::Allow, String::new(), String::new(), Some(proto_result))
                    }
                    Err(e) => {
                        warn!(error = %e, "tool interceptor error");
                        (ActionVerdict::Deny, e.to_string(), String::new(), None)
                    }
                }
            }
            Verdict::Deny(reason) => {
                (ActionVerdict::Deny, reason.clone(), String::new(), None)
            }
            Verdict::Escalate(rule_id) => {
                (ActionVerdict::Escalate, String::new(), rule_id.clone(), None)
            }
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
            action_id: Uuid::new_v4().to_string(),
            tool_name: req.tool_name.clone(),
            verdict: verdict_str.to_string(),
            evaluation_latency_us: eval_latency_us,
        });

        info!(
            tool_name = %req.tool_name,
            verdict = ?action_verdict,
            eval_latency_us = %eval_latency_us,
            "tool execution evaluated"
        );

        // 9. Fire-and-forget: persist action record to Activity Store
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
                record_id: String::new(),
                workspace_id: sandbox.workspace_id.clone(),
                agent_id: sandbox.agent_id.clone(),
                task_id: String::new(),
                tool_name: req.tool_name.clone(),
                parameters: proto_parameters.clone(),
                result: tool_result.clone(),
                outcome,
                guardrail_rule_id,
                denial_reason: denial_reason.clone(),
                evaluation_latency_us: eval_latency_us,
                execution_latency_us: 0,
                recorded_at: None,
            };
            let activity_mutex = activity_mutex.clone();
            tokio::spawn(async move {
                let mut client = activity_mutex.lock().await;
                if let Err(e) = client.record_action(RecordActionRequest { record: Some(record) }).await {
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
            result: tool_result,
            denial_reason,
            escalation_id,
        }))
    }

    async fn request_human_input(
        &self,
        request: Request<RequestHumanInputRequest>,
    ) -> Result<Response<RequestHumanInputResponse>, Status> {
        // Extract sandbox from metadata to get workspace_id and agent_id
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
            300 // default 5 minutes
        };

        // Create the human request via HIS — return immediately (non-blocking).
        let create_req = CreateHumanRequestRequest {
            workspace_id: sandbox.workspace_id.clone(),
            agent_id: sandbox.agent_id.clone(),
            question: req.question,
            options: req.options,
            context: req.context,
            timeout_seconds,
            r#type: 0,   // UNSPECIFIED — HIS defaults to "question"
            urgency: 0,  // UNSPECIFIED — HIS defaults to "normal"
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

        // Return immediately with the request_id so the agent can poll via check_human_request.
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
        // Try to extract sandbox for event emission
        let sandbox = self.sandbox_manager.lookup_by_metadata(&request).ok();

        let req = request.into_inner();
        info!(
            message = %req.message,
            percent_complete = %req.percent_complete,
            "agent reported progress"
        );

        // Emit progress event if we have a sandbox reference
        if let Some(sandbox) = sandbox {
            let _ = sandbox.event_tx.send(SandboxEvent::Progress {
                message: req.message,
                percent_complete: req.percent_complete,
            });
        }

        Ok(Response::new(ReportProgressResponse {}))
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

/// Convert a serde_json::Value into a prost_types::Struct.
fn json_to_proto_struct(value: &serde_json::Value) -> prost_types::Struct {
    match value {
        serde_json::Value::Object(map) => prost_types::Struct {
            fields: map
                .iter()
                .map(|(k, v)| (k.clone(), json_to_proto_value(v)))
                .collect(),
        },
        _ => prost_types::Struct {
            fields: std::collections::BTreeMap::new(),
        },
    }
}

fn json_to_proto_value(value: &serde_json::Value) -> prost_types::Value {
    use prost_types::value::Kind;
    prost_types::Value {
        kind: Some(match value {
            serde_json::Value::Null => Kind::NullValue(0),
            serde_json::Value::Bool(b) => Kind::BoolValue(*b),
            serde_json::Value::Number(n) => Kind::NumberValue(n.as_f64().unwrap_or(0.0)),
            serde_json::Value::String(s) => Kind::StringValue(s.clone()),
            serde_json::Value::Array(arr) => Kind::ListValue(prost_types::ListValue {
                values: arr.iter().map(json_to_proto_value).collect(),
            }),
            serde_json::Value::Object(_) => Kind::StructValue(json_to_proto_struct(value)),
        }),
    }
}

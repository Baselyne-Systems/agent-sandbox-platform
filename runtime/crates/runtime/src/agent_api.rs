use tonic::{Request, Response, Status};
use tracing::info;

use guardrails_eval::{EvalContext, Evaluator, Verdict};

use proto_gen::platform::runtime::v1::agent_api_service_server::AgentApiService;
use proto_gen::platform::runtime::v1::{
    ActionVerdict, ExecuteToolRequest, ExecuteToolResponse, ReportProgressRequest,
    ReportProgressResponse, RequestHumanInputRequest, RequestHumanInputResponse,
};

use crate::sandbox::SandboxManager;

/// gRPC implementation of the AgentAPIService — the agent-facing API exposed
/// inside each sandbox.
#[derive(Debug)]
pub struct AgentApiServiceImpl {
    _sandbox_manager: SandboxManager,
}

impl AgentApiServiceImpl {
    pub fn new(sandbox_manager: SandboxManager) -> Self {
        Self {
            _sandbox_manager: sandbox_manager,
        }
    }
}

#[tonic::async_trait]
impl AgentApiService for AgentApiServiceImpl {
    async fn execute_tool(
        &self,
        request: Request<ExecuteToolRequest>,
    ) -> Result<Response<ExecuteToolResponse>, Status> {
        let req = request.into_inner();
        info!(
            tool_name = %req.tool_name,
            justification = %req.justification,
            "agent requested tool execution"
        );

        // Convert proto Struct parameters to serde_json::Value for the evaluator.
        let parameters = req
            .parameters
            .map(|s| {
                let json_bytes = serde_json::to_vec(&proto_struct_to_value(&s))
                    .unwrap_or_default();
                serde_json::from_slice(&json_bytes).unwrap_or(serde_json::Value::Null)
            })
            .unwrap_or(serde_json::Value::Null);

        // Build the evaluation context.
        let eval_ctx = EvalContext {
            tool_name: req.tool_name.clone(),
            parameters,
            // In a real implementation, agent_id would come from the
            // authenticated sandbox context.
            agent_id: "unknown".to_string(),
        };

        // Create a stub evaluator with an empty policy.
        let stub_policy = guardrails_eval::CompiledPolicy {
            rules: Vec::new(),
        };
        let policy_bytes =
            serde_json::to_vec(&stub_policy).map_err(|e| Status::internal(e.to_string()))?;
        let evaluator =
            Evaluator::load(&policy_bytes).map_err(|e| Status::internal(e.to_string()))?;

        let verdict = evaluator.evaluate(&eval_ctx);

        let (action_verdict, denial_reason, escalation_id) = match verdict {
            Verdict::Allow => (ActionVerdict::Allow, String::new(), String::new()),
            Verdict::Deny(reason) => (ActionVerdict::Deny, reason, String::new()),
            Verdict::Escalate(ctx) => (ActionVerdict::Escalate, String::new(), ctx),
        };

        info!(
            tool_name = %req.tool_name,
            verdict = ?action_verdict,
            "tool execution evaluated"
        );

        Ok(Response::new(ExecuteToolResponse {
            verdict: action_verdict as i32,
            result: None,
            denial_reason,
            escalation_id,
        }))
    }

    async fn request_human_input(
        &self,
        _request: Request<RequestHumanInputRequest>,
    ) -> Result<Response<RequestHumanInputResponse>, Status> {
        Err(Status::unimplemented(
            "RequestHumanInput is not yet implemented",
        ))
    }

    async fn report_progress(
        &self,
        request: Request<ReportProgressRequest>,
    ) -> Result<Response<ReportProgressResponse>, Status> {
        let req = request.into_inner();
        info!(
            message = %req.message,
            percent_complete = %req.percent_complete,
            "agent reported progress"
        );

        Ok(Response::new(ReportProgressResponse {}))
    }
}

/// Convert a prost_types::Struct into a serde_json::Value.
///
/// This is a lightweight helper; a production implementation would use a more
/// robust conversion or a dedicated crate.
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

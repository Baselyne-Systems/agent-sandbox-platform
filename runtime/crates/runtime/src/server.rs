use std::collections::HashMap;
use std::pin::Pin;
use std::sync::atomic::Ordering;

use tokio_stream::wrappers::BroadcastStream;
use tokio_stream::{Stream, StreamExt};
use tonic::{Request, Response, Status};
use tracing::info;

use proto_gen::platform::runtime::v1::runtime_service_server::RuntimeService;
use proto_gen::platform::runtime::v1::{
    CreateSandboxRequest, CreateSandboxResponse, DestroySandboxRequest, DestroySandboxResponse,
    GetSandboxStatusRequest, GetSandboxStatusResponse, SandboxEvent as ProtoSandboxEvent,
    SandboxState as ProtoSandboxState, StreamEventsRequest,
    UpdateSandboxGuardrailsRequest, UpdateSandboxGuardrailsResponse,
};
use proto_gen::platform::runtime::v1::{ActionEvent, ActionVerdict, LifecycleEvent};

use crate::sandbox::{CreateSandboxParams, SandboxEvent, SandboxManager, SandboxStatus};

/// gRPC implementation of the RuntimeService — called by the control-plane
/// Workspace Service to manage sandboxes on this host.
#[derive(Debug)]
pub struct RuntimeServiceImpl {
    sandbox_manager: SandboxManager,
    advertise_endpoint: String,
}

impl RuntimeServiceImpl {
    pub fn new(sandbox_manager: SandboxManager, advertise_endpoint: String) -> Self {
        Self {
            sandbox_manager,
            advertise_endpoint,
        }
    }
}

#[tonic::async_trait]
impl RuntimeService for RuntimeServiceImpl {
    type StreamEventsStream =
        Pin<Box<dyn Stream<Item = Result<ProtoSandboxEvent, Status>> + Send + 'static>>;

    async fn create_sandbox(
        &self,
        request: Request<CreateSandboxRequest>,
    ) -> Result<Response<CreateSandboxResponse>, Status> {
        let req = request.into_inner();
        info!(
            workspace_id = %req.workspace_id,
            agent_id = %req.agent_id,
            "creating sandbox"
        );

        let (allowed_tools, env_vars) = req
            .spec
            .as_ref()
            .map(|s| (s.allowed_tools.clone(), s.env_vars.clone()))
            .unwrap_or_else(|| (vec![], HashMap::new()));

        let params = CreateSandboxParams {
            workspace_id: req.workspace_id.clone(),
            agent_id: req.agent_id.clone(),
            allowed_tools,
            env_vars,
            compiled_guardrails: req.compiled_guardrails,
        };

        let state = self
            .sandbox_manager
            .create(params)
            .map_err(|e| Status::internal(e.to_string()))?;

        let agent_api_endpoint = self.advertise_endpoint.clone();

        info!(
            sandbox_id = %state.id,
            "sandbox created"
        );

        Ok(Response::new(CreateSandboxResponse {
            sandbox_id: state.id.clone(),
            agent_api_endpoint,
        }))
    }

    async fn destroy_sandbox(
        &self,
        request: Request<DestroySandboxRequest>,
    ) -> Result<Response<DestroySandboxResponse>, Status> {
        let req = request.into_inner();
        info!(
            sandbox_id = %req.sandbox_id,
            reason = %req.reason,
            "destroying sandbox"
        );

        self.sandbox_manager
            .destroy(&req.sandbox_id)
            .map_err(|e| Status::not_found(e.to_string()))?;

        Ok(Response::new(DestroySandboxResponse {}))
    }

    async fn get_sandbox_status(
        &self,
        request: Request<GetSandboxStatusRequest>,
    ) -> Result<Response<GetSandboxStatusResponse>, Status> {
        let req = request.into_inner();
        info!(sandbox_id = %req.sandbox_id, "getting sandbox status");

        let sandbox = self
            .sandbox_manager
            .get_sandbox(&req.sandbox_id)
            .map_err(|e| Status::not_found(e.to_string()))?;

        let state = {
            let status = sandbox.status.lock().map_err(|e| {
                Status::internal(format!("lock poisoned: {e}"))
            })?;
            match *status {
                SandboxStatus::Starting => ProtoSandboxState::Starting,
                SandboxStatus::Running => ProtoSandboxState::Running,
                SandboxStatus::Stopped => ProtoSandboxState::Stopped,
                SandboxStatus::Failed => ProtoSandboxState::Failed,
            }
        };

        let started_at = sandbox
            .created_at
            .duration_since(std::time::UNIX_EPOCH)
            .ok()
            .map(|d| prost_types::Timestamp {
                seconds: d.as_secs() as i64,
                nanos: d.subsec_nanos() as i32,
            });

        Ok(Response::new(GetSandboxStatusResponse {
            sandbox_id: sandbox.id.clone(),
            state: state as i32,
            memory_used_mb: 0,
            cpu_used_millicores: 0,
            actions_executed: sandbox.actions_executed.load(Ordering::Relaxed) as i32,
            started_at,
        }))
    }

    async fn stream_events(
        &self,
        request: Request<StreamEventsRequest>,
    ) -> Result<Response<Self::StreamEventsStream>, Status> {
        let req = request.into_inner();
        info!(sandbox_id = %req.sandbox_id, "streaming events");

        let sandbox = self
            .sandbox_manager
            .get_sandbox(&req.sandbox_id)
            .map_err(|e| Status::not_found(e.to_string()))?;

        let sandbox_id = sandbox.id.clone();
        let rx = sandbox.event_tx.subscribe();
        let stream = BroadcastStream::new(rx);

        let output = stream.filter_map(move |result| {
            let sid = sandbox_id.clone();
            match result {
                Ok(event) => {
                    let now = std::time::SystemTime::now()
                        .duration_since(std::time::UNIX_EPOCH)
                        .unwrap_or_default();
                    let timestamp = Some(prost_types::Timestamp {
                        seconds: now.as_secs() as i64,
                        nanos: now.subsec_nanos() as i32,
                    });

                    let proto_event = match event {
                        SandboxEvent::Action {
                            action_id,
                            tool_name,
                            verdict,
                            evaluation_latency_us,
                        } => {
                            let verdict_enum = match verdict.as_str() {
                                "allow" => ActionVerdict::Allow,
                                "deny" => ActionVerdict::Deny,
                                "escalate" => ActionVerdict::Escalate,
                                _ => ActionVerdict::Unspecified,
                            };
                            ProtoSandboxEvent {
                                sandbox_id: sid,
                                timestamp,
                                event: Some(
                                    proto_gen::platform::runtime::v1::sandbox_event::Event::Action(
                                        ActionEvent {
                                            action_id,
                                            tool_name,
                                            verdict: verdict_enum as i32,
                                            evaluation_latency_us,
                                        },
                                    ),
                                ),
                            }
                        }
                        SandboxEvent::Lifecycle { new_state, reason } => {
                            let state_enum = match new_state.as_str() {
                                "starting" => ProtoSandboxState::Starting,
                                "running" => ProtoSandboxState::Running,
                                "stopped" => ProtoSandboxState::Stopped,
                                "failed" => ProtoSandboxState::Failed,
                                _ => ProtoSandboxState::Unspecified,
                            };
                            ProtoSandboxEvent {
                                sandbox_id: sid,
                                timestamp,
                                event: Some(
                                    proto_gen::platform::runtime::v1::sandbox_event::Event::Lifecycle(
                                        LifecycleEvent {
                                            new_state: state_enum as i32,
                                            reason,
                                        },
                                    ),
                                ),
                            }
                        }
                        SandboxEvent::Progress { .. } => {
                            // Progress events are not in the proto schema,
                            // map as lifecycle with reason containing message
                            return None;
                        }
                    };
                    Some(Ok(proto_event))
                }
                Err(tokio_stream::wrappers::errors::BroadcastStreamRecvError::Lagged(n)) => {
                    tracing::warn!(skipped = n, "event stream lagged, skipping missed events");
                    None
                }
            }
        });

        Ok(Response::new(Box::pin(output)))
    }

    async fn update_sandbox_guardrails(
        &self,
        request: Request<UpdateSandboxGuardrailsRequest>,
    ) -> Result<Response<UpdateSandboxGuardrailsResponse>, Status> {
        let req = request.into_inner();
        info!(sandbox_id = %req.sandbox_id, "updating sandbox guardrails");

        self.sandbox_manager
            .update_guardrails(&req.sandbox_id, &req.compiled_guardrails)
            .map_err(|e| Status::internal(e.to_string()))?;

        Ok(Response::new(UpdateSandboxGuardrailsResponse {}))
    }
}

use std::pin::Pin;

use tokio_stream::Stream;
use tonic::{Request, Response, Status};
use tracing::info;

use proto_gen::platform::runtime::v1::runtime_service_server::RuntimeService;
use proto_gen::platform::runtime::v1::{
    CreateSandboxRequest, CreateSandboxResponse, DestroySandboxRequest, DestroySandboxResponse,
    GetSandboxStatusRequest, GetSandboxStatusResponse, SandboxEvent, SandboxState,
    StreamEventsRequest,
};

use crate::sandbox::SandboxManager;

/// gRPC implementation of the RuntimeService — called by the control-plane
/// Workspace Service to manage sandboxes on this host.
#[derive(Debug)]
pub struct RuntimeServiceImpl {
    sandbox_manager: SandboxManager,
}

impl RuntimeServiceImpl {
    pub fn new(sandbox_manager: SandboxManager) -> Self {
        Self { sandbox_manager }
    }
}

#[tonic::async_trait]
impl RuntimeService for RuntimeServiceImpl {
    type StreamEventsStream =
        Pin<Box<dyn Stream<Item = Result<SandboxEvent, Status>> + Send + 'static>>;

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

        let spec_summary = req
            .spec
            .as_ref()
            .map(|s| format!("{}MB/{}mc", s.memory_mb, s.cpu_millicores))
            .unwrap_or_else(|| "default".to_string());

        let handle = self
            .sandbox_manager
            .create(req.workspace_id.clone(), spec_summary)
            .map_err(|e| Status::internal(e.to_string()))?;

        let agent_api_endpoint = "localhost:50052".to_string();

        info!(
            sandbox_id = %handle.id,
            "sandbox created"
        );

        Ok(Response::new(CreateSandboxResponse {
            sandbox_id: handle.id,
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

        let handle = self
            .sandbox_manager
            .get_status(&req.sandbox_id)
            .map_err(|e| Status::not_found(e.to_string()))?;

        Ok(Response::new(GetSandboxStatusResponse {
            sandbox_id: handle.id,
            state: SandboxState::Running as i32,
            memory_used_mb: 0,
            cpu_used_millicores: 0,
            actions_executed: 0,
            started_at: None,
        }))
    }

    async fn stream_events(
        &self,
        _request: Request<StreamEventsRequest>,
    ) -> Result<Response<Self::StreamEventsStream>, Status> {
        Err(Status::unimplemented(
            "StreamEvents is not yet implemented",
        ))
    }
}

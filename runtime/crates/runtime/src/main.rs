mod agent_api;
mod sandbox;
mod server;
mod tools;

use std::net::SocketAddr;

use anyhow::Result;
use tonic::transport::Server;
use tracing::info;

use proto_gen::platform::runtime::v1::agent_api_service_server::AgentApiServiceServer;
use proto_gen::platform::runtime::v1::runtime_service_server::RuntimeServiceServer;

use crate::agent_api::AgentApiServiceImpl;
use crate::sandbox::SandboxManager;
use crate::server::RuntimeServiceImpl;

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize tracing with env-filter support (e.g. RUST_LOG=debug).
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| tracing_subscriber::EnvFilter::new("info")),
        )
        .init();

    let port: u16 = std::env::var("GRPC_PORT")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(50052);

    let addr: SocketAddr = format!("0.0.0.0:{port}").parse()?;

    let sandbox_manager = SandboxManager::new();
    let runtime_service = RuntimeServiceImpl::new(sandbox_manager.clone());
    let agent_api_service = AgentApiServiceImpl::new(sandbox_manager);

    info!("Runtime starting on :{port}");

    Server::builder()
        .add_service(RuntimeServiceServer::new(runtime_service))
        .add_service(AgentApiServiceServer::new(agent_api_service))
        .serve_with_shutdown(addr, async {
            tokio::signal::ctrl_c()
                .await
                .expect("failed to listen for ctrl+c");
            info!("Received ctrl+c, shutting down gracefully");
        })
        .await?;

    Ok(())
}

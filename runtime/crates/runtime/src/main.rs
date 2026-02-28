mod agent_api;
mod container;
mod sandbox;
mod server;

use std::net::SocketAddr;
use std::sync::Arc;

use anyhow::Result;
use tonic::transport::Server;
use tracing::info;

use proto_gen::platform::activity::v1::activity_service_client::ActivityServiceClient;
use proto_gen::platform::economics::v1::economics_service_client::EconomicsServiceClient;
use proto_gen::platform::human::v1::human_interaction_service_client::HumanInteractionServiceClient;
use proto_gen::platform::host_agent::v1::host_agent_api_service_server::HostAgentApiServiceServer;
use proto_gen::platform::host_agent::v1::host_agent_service_server::HostAgentServiceServer;

use crate::agent_api::HostAgentApiServiceImpl;
use crate::sandbox::SandboxManager;
use crate::server::HostAgentServiceImpl;

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

    // Optional HIS endpoint for human interaction forwarding.
    let his_client = match std::env::var("HIS_ENDPOINT") {
        Ok(endpoint) => {
            info!(endpoint = %endpoint, "connecting to HIS");
            let client = HumanInteractionServiceClient::connect(endpoint).await?;
            Some(client)
        }
        Err(_) => {
            info!("HIS_ENDPOINT not set — RequestHumanInput will return unavailable");
            None
        }
    };

    // Optional Activity Store endpoint for persisting action records.
    let activity_client = match std::env::var("ACTIVITY_ENDPOINT") {
        Ok(endpoint) => {
            info!(endpoint = %endpoint, "connecting to Activity Store");
            let client = ActivityServiceClient::connect(endpoint).await?;
            Some(client)
        }
        Err(_) => {
            info!("ACTIVITY_ENDPOINT not set — action records will not be persisted");
            None
        }
    };

    // Optional Economics Service endpoint for budget enforcement.
    let economics_client = match std::env::var("ECONOMICS_ENDPOINT") {
        Ok(endpoint) => {
            info!(endpoint = %endpoint, "connecting to Economics Service");
            let client = EconomicsServiceClient::connect(endpoint).await?;
            Some(client)
        }
        Err(_) => {
            info!("ECONOMICS_ENDPOINT not set — budget enforcement disabled");
            None
        }
    };

    let advertise_addr = std::env::var("ADVERTISE_ADDRESS").unwrap_or_else(|_| "localhost".to_string());
    let advertise_endpoint = format!("{advertise_addr}:{port}");
    info!(advertise_endpoint = %advertise_endpoint, "agent API advertise endpoint");

    // Initialize container runtime — DockerRuntime if ENABLE_DOCKER=true, otherwise Noop.
    let container_runtime: Arc<dyn container::ContainerRuntime> =
        if std::env::var("ENABLE_DOCKER").unwrap_or_default() == "true" {
            info!("Docker container runtime enabled");
            Arc::new(container::DockerRuntime::new()?)
        } else {
            info!("Docker disabled — using noop container runtime");
            Arc::new(container::NoopContainerRuntime)
        };

    let sandbox_manager = SandboxManager::new(container_runtime);
    let host_agent_service = HostAgentServiceImpl::new(sandbox_manager.clone(), advertise_endpoint);
    let host_agent_api_service =
        HostAgentApiServiceImpl::new(sandbox_manager, his_client, activity_client, economics_client);

    info!("Host Agent starting on :{port}");

    Server::builder()
        .add_service(HostAgentServiceServer::new(host_agent_service))
        .add_service(HostAgentApiServiceServer::new(host_agent_api_service))
        .serve_with_shutdown(addr, async {
            tokio::signal::ctrl_c()
                .await
                .expect("failed to listen for ctrl+c");
            info!("Received ctrl+c, shutting down gracefully");
        })
        .await?;

    Ok(())
}

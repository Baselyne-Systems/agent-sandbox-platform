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
use proto_gen::platform::compute::v1::compute_plane_service_client::ComputePlaneServiceClient;
use proto_gen::platform::compute::v1::{HeartbeatRequest, HostResources, RegisterHostRequest};
use proto_gen::platform::economics::v1::economics_service_client::EconomicsServiceClient;
use proto_gen::platform::governance::v1::data_governance_service_client::DataGovernanceServiceClient;
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

    // Optional Governance Service endpoint for DLP egress inspection.
    let governance_client = match std::env::var("GOVERNANCE_ENDPOINT") {
        Ok(endpoint) => {
            info!(endpoint = %endpoint, "connecting to Governance Service");
            let client = DataGovernanceServiceClient::connect(endpoint).await?;
            Some(client)
        }
        Err(_) => {
            info!("GOVERNANCE_ENDPOINT not set — DLP egress inspection disabled");
            None
        }
    };

    // Determine supported isolation tiers.
    let supported_tiers: Vec<String> = {
        let default_tiers = "standard,hardened".to_string();
        let mut tiers: Vec<String> = std::env::var("SUPPORTED_TIERS")
            .unwrap_or(default_tiers)
            .split(',')
            .map(|s| s.trim().to_string())
            .filter(|s| !s.is_empty())
            .collect();
        // Only include "isolated" if ISOLATED_RUNTIME is set.
        if std::env::var("ISOLATED_RUNTIME").is_ok() && !tiers.contains(&"isolated".to_string()) {
            tiers.push("isolated".to_string());
        }
        tiers
    };
    info!(supported_tiers = ?supported_tiers, "isolation tiers available");

    let advertise_addr = std::env::var("ADVERTISE_ADDRESS").unwrap_or_else(|_| "localhost".to_string());
    let advertise_endpoint = format!("{advertise_addr}:{port}");
    info!(advertise_endpoint = %advertise_endpoint, "agent API advertise endpoint");

    // Host resource configuration (from env vars, with sensible defaults).
    let total_memory_mb: i64 = std::env::var("TOTAL_MEMORY_MB")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(16384);
    let total_cpu_millicores: i32 = std::env::var("TOTAL_CPU_MILLICORES")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(8000);
    let total_disk_mb: i64 = std::env::var("TOTAL_DISK_MB")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(102400);
    info!(total_memory_mb, total_cpu_millicores, total_disk_mb, "host resource configuration");

    // Optional Compute Plane endpoint for host self-registration and heartbeats.
    let compute_registration: Option<(String, ComputePlaneServiceClient<tonic::transport::Channel>)> =
        match std::env::var("COMPUTE_ENDPOINT") {
            Ok(endpoint) => {
                info!(endpoint = %endpoint, "connecting to Compute Plane");
                match ComputePlaneServiceClient::connect(endpoint).await {
                    Ok(mut client) => {
                        let register_resp = client
                            .register_host(RegisterHostRequest {
                                address: advertise_endpoint.clone(),
                                total_resources: Some(HostResources {
                                    memory_mb: total_memory_mb,
                                    cpu_millicores: total_cpu_millicores,
                                    disk_mb: total_disk_mb,
                                }),
                                supported_tiers: supported_tiers.clone(),
                            })
                            .await;

                        match register_resp {
                            Ok(resp) => {
                                let host = resp.into_inner().host.expect("host in register response");
                                let host_id = host.host_id.clone();
                                info!(host_id = %host_id, "registered with Compute Plane");
                                Some((host_id, client))
                            }
                            Err(e) => {
                                tracing::error!(error = %e, "failed to register with Compute Plane — heartbeats disabled");
                                None
                            }
                        }
                    }
                    Err(e) => {
                        tracing::error!(error = %e, "failed to connect to Compute Plane — heartbeats disabled");
                        None
                    }
                }
            }
            Err(_) => {
                info!("COMPUTE_ENDPOINT not set — host registration and heartbeats disabled");
                None
            }
        };

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
        HostAgentApiServiceImpl::new(sandbox_manager.clone(), his_client, activity_client, economics_client, governance_client);

    // Spawn periodic heartbeat loop if registered with Compute Plane.
    let heartbeat_handle = if let Some((host_id, client)) = compute_registration {
        let heartbeat_sandbox_manager = sandbox_manager;
        let heartbeat_tiers = supported_tiers;
        let heartbeat_host_id = host_id.clone();

        let handle = tokio::spawn(async move {
            let mut interval = tokio::time::interval(std::time::Duration::from_secs(30));
            let mut client = client;

            // Per-sandbox resource defaults for estimation.
            let per_sandbox_memory: i64 = 512;
            let per_sandbox_cpu: i32 = 500;
            let per_sandbox_disk: i64 = 1024;

            loop {
                interval.tick().await;

                let active_sandboxes = heartbeat_sandbox_manager.active_count() as i32;
                let available_memory = (total_memory_mb - (active_sandboxes as i64 * per_sandbox_memory)).max(0);
                let available_cpu = (total_cpu_millicores - (active_sandboxes * per_sandbox_cpu)).max(0);
                let available_disk = (total_disk_mb - (active_sandboxes as i64 * per_sandbox_disk)).max(0);

                let result = client
                    .heartbeat(HeartbeatRequest {
                        host_id: heartbeat_host_id.clone(),
                        available_memory_mb: available_memory,
                        available_cpu_millicores: available_cpu,
                        available_disk_mb: available_disk,
                        active_sandboxes,
                        supported_tiers: heartbeat_tiers.clone(),
                    })
                    .await;

                match result {
                    Ok(resp) => {
                        let status = resp.into_inner().status;
                        // HostStatus::HostStatusDraining = 2
                        if status == 2 {
                            tracing::warn!(
                                host_id = %heartbeat_host_id,
                                "compute plane reports DRAINING status"
                            );
                        }
                    }
                    Err(e) => {
                        tracing::warn!(
                            error = %e,
                            host_id = %heartbeat_host_id,
                            "heartbeat failed — will retry next interval"
                        );
                    }
                }
            }
        });
        Some((host_id, handle))
    } else {
        None
    };

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

    // Clean up heartbeat loop after server stops.
    if let Some((host_id, handle)) = heartbeat_handle {
        handle.abort();
        info!(host_id = %host_id, "heartbeat loop stopped");
    }

    Ok(())
}

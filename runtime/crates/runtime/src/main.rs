mod agent_api;
mod container;
mod sandbox;
mod server;

use std::net::SocketAddr;
use std::sync::Arc;

use anyhow::Result;
use opentelemetry::trace::TracerProvider as _;
use opentelemetry_otlp::WithExportConfig;
use tonic::transport::Server;
use tracing::info;
use tracing_subscriber::layer::SubscriberExt;
use tracing_subscriber::util::SubscriberInitExt;

use proto_gen::platform::activity::v1::activity_service_client::ActivityServiceClient;
use proto_gen::platform::compute::v1::compute_plane_service_client::ComputePlaneServiceClient;
use proto_gen::platform::compute::v1::{HeartbeatRequest, HostResources, RegisterHostRequest};
use proto_gen::platform::economics::v1::economics_service_client::EconomicsServiceClient;
use proto_gen::platform::governance::v1::data_governance_service_client::DataGovernanceServiceClient;
use proto_gen::platform::host_agent::v1::host_agent_api_service_server::HostAgentApiServiceServer;
use proto_gen::platform::host_agent::v1::host_agent_service_server::HostAgentServiceServer;
use proto_gen::platform::human::v1::human_interaction_service_client::HumanInteractionServiceClient;

use crate::agent_api::HostAgentApiServiceImpl;
use crate::sandbox::SandboxManager;
use crate::server::HostAgentServiceImpl;

/// Resolves a gRPC endpoint: individual env var takes priority,
/// then CONTROL_PLANE + well-known port, then None.
fn resolve_endpoint(
    env_key: &str,
    control_plane: &Option<String>,
    default_port: u16,
) -> Option<String> {
    if let Ok(endpoint) = std::env::var(env_key) {
        return Some(endpoint);
    }
    control_plane
        .as_ref()
        .map(|cp| format!("http://{}:{}", cp, default_port))
}

/// Auto-detects host resources, with env var overrides.
fn detect_resources() -> (i64, i32, i64) {
    let detected_memory_mb = {
        let sys = sysinfo::System::new_with_specifics(
            sysinfo::RefreshKind::nothing().with_memory(sysinfo::MemoryRefreshKind::everything()),
        );
        (sys.total_memory() / (1024 * 1024)) as i64
    };

    let detected_cpu_millicores = std::thread::available_parallelism()
        .map(|n| n.get() as i32 * 1000)
        .unwrap_or(4000);

    let detected_disk_mb = {
        let disks = sysinfo::Disks::new_with_refreshed_list();
        disks
            .iter()
            .find(|d| d.mount_point() == std::path::Path::new("/"))
            .map(|d| (d.total_space() / (1024 * 1024)) as i64)
            .unwrap_or(102400)
    };

    let total_memory_mb = std::env::var("TOTAL_MEMORY_MB")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(detected_memory_mb);

    let total_cpu_millicores = std::env::var("TOTAL_CPU_MILLICORES")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(detected_cpu_millicores);

    let total_disk_mb = std::env::var("TOTAL_DISK_MB")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(detected_disk_mb);

    (total_memory_mb, total_cpu_millicores, total_disk_mb)
}

/// Auto-detects the local IP via the OS routing table.
/// Routes toward CONTROL_PLANE (or 8.8.8.8) to find the right interface.
fn detect_advertise_address(control_plane: &Option<String>) -> String {
    if let Ok(addr) = std::env::var("ADVERTISE_ADDRESS") {
        return addr;
    }

    let target = control_plane
        .as_ref()
        .map(|cp| format!("{}:80", cp))
        .unwrap_or_else(|| "8.8.8.8:80".to_string());

    std::net::UdpSocket::bind("0.0.0.0:0")
        .and_then(|socket| {
            socket.connect(&target)?;
            socket.local_addr()
        })
        .map(|local| local.ip().to_string())
        .unwrap_or_else(|_| "localhost".to_string())
}

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize tracing with env-filter support (e.g. RUST_LOG=debug).
    // Conditionally adds OTLP tracing export if OTEL_EXPORTER_OTLP_ENDPOINT is set.
    let env_filter = tracing_subscriber::EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| tracing_subscriber::EnvFilter::new("info"));

    let fmt_layer = tracing_subscriber::fmt::layer();

    let registry = tracing_subscriber::registry()
        .with(env_filter)
        .with(fmt_layer);

    if let Ok(otlp_endpoint) = std::env::var("OTEL_EXPORTER_OTLP_ENDPOINT") {
        let exporter = opentelemetry_otlp::SpanExporter::builder()
            .with_tonic()
            .with_endpoint(otlp_endpoint)
            .build()
            .expect("failed to create OTLP exporter");

        let tracer_provider = opentelemetry_sdk::trace::TracerProvider::builder()
            .with_batch_exporter(exporter, opentelemetry_sdk::runtime::Tokio)
            .with_resource(opentelemetry_sdk::Resource::new(vec![
                opentelemetry::KeyValue::new("service.name", "bulkhead-host-agent"),
            ]))
            .build();

        let tracer = tracer_provider.tracer("host-agent");
        let telemetry_layer = tracing_opentelemetry::layer().with_tracer(tracer);

        registry.with(telemetry_layer).init();
        info!("OpenTelemetry tracing enabled");
    } else {
        registry.init();
    }

    let port: u16 = std::env::var("GRPC_PORT")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(50052);

    let addr: SocketAddr = format!("0.0.0.0:{port}").parse()?;

    // CONTROL_PLANE: single address that derives all service endpoints.
    let control_plane: Option<String> = std::env::var("CONTROL_PLANE").ok();
    if let Some(ref cp) = control_plane {
        info!(control_plane = %cp, "deriving service endpoints from well-known ports");
    }

    // Connect to control-plane services. Each checks its individual env var
    // first, then falls back to CONTROL_PLANE + well-known port.
    let his_client = match resolve_endpoint("HIS_ENDPOINT", &control_plane, 50063) {
        Some(endpoint) => {
            info!(endpoint = %endpoint, "connecting to HIS");
            let client = HumanInteractionServiceClient::connect(endpoint).await?;
            Some(client)
        }
        None => {
            info!("HIS endpoint not configured — RequestHumanInput will return unavailable");
            None
        }
    };

    let activity_client = match resolve_endpoint("ACTIVITY_ENDPOINT", &control_plane, 50065) {
        Some(endpoint) => {
            info!(endpoint = %endpoint, "connecting to Activity Store");
            let client = ActivityServiceClient::connect(endpoint).await?;
            Some(client)
        }
        None => {
            info!("Activity endpoint not configured — action records will not be persisted");
            None
        }
    };

    let economics_client = match resolve_endpoint("ECONOMICS_ENDPOINT", &control_plane, 50066) {
        Some(endpoint) => {
            info!(endpoint = %endpoint, "connecting to Economics Service");
            let client = EconomicsServiceClient::connect(endpoint).await?;
            Some(client)
        }
        None => {
            info!("Economics endpoint not configured — budget enforcement disabled");
            None
        }
    };

    let governance_client = match resolve_endpoint("GOVERNANCE_ENDPOINT", &control_plane, 50064) {
        Some(endpoint) => {
            info!(endpoint = %endpoint, "connecting to Governance Service");
            let client = DataGovernanceServiceClient::connect(endpoint).await?;
            Some(client)
        }
        None => {
            info!("Governance endpoint not configured — DLP egress inspection disabled");
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

    let advertise_addr = detect_advertise_address(&control_plane);
    let advertise_endpoint = format!("{advertise_addr}:{port}");
    info!(
        advertise_endpoint = %advertise_endpoint,
        source = if std::env::var("ADVERTISE_ADDRESS").is_ok() { "env" } else { "auto-detected" },
        "advertise address"
    );

    // Host resource configuration (auto-detected from system, env vars override).
    let (total_memory_mb, total_cpu_millicores, total_disk_mb) = detect_resources();
    info!(
        total_memory_mb,
        total_cpu_millicores,
        total_disk_mb,
        memory_source = if std::env::var("TOTAL_MEMORY_MB").is_ok() {
            "env"
        } else {
            "auto-detected"
        },
        cpu_source = if std::env::var("TOTAL_CPU_MILLICORES").is_ok() {
            "env"
        } else {
            "auto-detected"
        },
        disk_source = if std::env::var("TOTAL_DISK_MB").is_ok() {
            "env"
        } else {
            "auto-detected"
        },
        "host resource configuration"
    );

    // Compute Plane for host self-registration and heartbeats.
    let compute_registration: Option<(
        String,
        ComputePlaneServiceClient<tonic::transport::Channel>,
    )> = match resolve_endpoint("COMPUTE_ENDPOINT", &control_plane, 50067) {
        Some(endpoint) => {
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
        None => {
            info!("Compute endpoint not configured — host registration and heartbeats disabled");
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
    let host_agent_api_service = HostAgentApiServiceImpl::new(
        sandbox_manager.clone(),
        his_client,
        activity_client,
        economics_client,
        governance_client,
    );

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
                let available_memory =
                    (total_memory_mb - (active_sandboxes as i64 * per_sandbox_memory)).max(0);
                let available_cpu =
                    (total_cpu_millicores - (active_sandboxes * per_sandbox_cpu)).max(0);
                let available_disk =
                    (total_disk_mb - (active_sandboxes as i64 * per_sandbox_disk)).max(0);

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

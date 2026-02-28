use std::collections::HashMap;

use anyhow::Result;
use async_trait::async_trait;
use tracing::{info, warn};

/// Pluggable container runtime for managing sandbox containers.
/// DockerRuntime uses bollard to talk to the Docker daemon.
/// NoopContainerRuntime is a pass-through for testing or non-Docker deployments.
#[async_trait]
pub trait ContainerRuntime: Send + Sync + std::fmt::Debug {
    /// Start a container for the given sandbox. Returns a container ID.
    async fn start_container(
        &self,
        sandbox_id: &str,
        image: &str,
        env_vars: HashMap<String, String>,
        memory_bytes: i64,
        cpu_quota: i64,
    ) -> Result<String>;

    /// Stop and remove a container.
    async fn stop_container(&self, container_id: &str) -> Result<()>;
}

/// Docker-based container runtime using bollard.
#[derive(Debug)]
pub struct DockerRuntime {
    client: bollard::Docker,
}

impl DockerRuntime {
    pub fn new() -> Result<Self> {
        let client = bollard::Docker::connect_with_local_defaults()
            .map_err(|e| anyhow::anyhow!("failed to connect to Docker: {e}"))?;
        info!("connected to Docker daemon");
        Ok(Self { client })
    }
}

#[async_trait]
impl ContainerRuntime for DockerRuntime {
    async fn start_container(
        &self,
        sandbox_id: &str,
        image: &str,
        env_vars: HashMap<String, String>,
        memory_bytes: i64,
        cpu_quota: i64,
    ) -> Result<String> {
        // 1. Pull image (best-effort — may already exist locally)
        use futures_util::StreamExt;
        info!(image = %image, sandbox_id = %sandbox_id, "pulling container image");
        let mut pull_stream = self.client.create_image(
            Some(bollard::image::CreateImageOptions {
                from_image: image,
                ..Default::default()
            }),
            None,
            None,
        );
        while let Some(result) = pull_stream.next().await {
            match result {
                Ok(_) => {}
                Err(e) => {
                    warn!(error = %e, image = %image, "image pull failed — attempting to use local image");
                    break;
                }
            }
        }

        // 2. Build env vars list (KEY=VALUE format) with Bulkhead injections
        let mut env: Vec<String> = env_vars
            .iter()
            .map(|(k, v)| format!("{k}={v}"))
            .collect();
        // Inject Bulkhead SDK connection vars
        if let Ok(endpoint) = std::env::var("ADVERTISE_ADDRESS") {
            let port = std::env::var("GRPC_PORT").unwrap_or_else(|_| "50052".to_string());
            env.push(format!("BULKHEAD_ENDPOINT={endpoint}:{port}"));
        }
        env.push(format!("BULKHEAD_SANDBOX_ID={sandbox_id}"));

        // 3. Create container with resource limits
        let container_name = format!("bulkhead-{sandbox_id}");
        let host_config = bollard::models::HostConfig {
            memory: Some(memory_bytes),
            cpu_quota: Some(cpu_quota),
            ..Default::default()
        };

        let config = bollard::container::Config {
            image: Some(image.to_string()),
            env: Some(env),
            host_config: Some(host_config),
            ..Default::default()
        };

        let create_resp = self
            .client
            .create_container(
                Some(bollard::container::CreateContainerOptions {
                    name: container_name.as_str(),
                    platform: None,
                }),
                config,
            )
            .await
            .map_err(|e| anyhow::anyhow!("failed to create container: {e}"))?;

        let container_id = create_resp.id;
        info!(container_id = %container_id, sandbox_id = %sandbox_id, "container created");

        // 4. Start container
        self.client
            .start_container::<String>(&container_id, None)
            .await
            .map_err(|e| anyhow::anyhow!("failed to start container: {e}"))?;

        info!(container_id = %container_id, sandbox_id = %sandbox_id, "container started");
        Ok(container_id)
    }

    async fn stop_container(&self, container_id: &str) -> Result<()> {
        info!(container_id = %container_id, "stopping container");

        // Stop with 10s timeout
        let _ = self
            .client
            .stop_container(
                container_id,
                Some(bollard::container::StopContainerOptions { t: 10 }),
            )
            .await;

        // Force remove
        self.client
            .remove_container(
                container_id,
                Some(bollard::container::RemoveContainerOptions {
                    force: true,
                    ..Default::default()
                }),
            )
            .await
            .map_err(|e| anyhow::anyhow!("failed to remove container: {e}"))?;

        info!(container_id = %container_id, "container removed");
        Ok(())
    }
}

/// No-op container runtime for testing or deployments without Docker.
#[derive(Debug, Clone)]
pub struct NoopContainerRuntime;

#[async_trait]
impl ContainerRuntime for NoopContainerRuntime {
    async fn start_container(
        &self,
        sandbox_id: &str,
        image: &str,
        _env_vars: HashMap<String, String>,
        _memory_bytes: i64,
        _cpu_quota: i64,
    ) -> Result<String> {
        info!(
            sandbox_id = %sandbox_id,
            image = %image,
            "noop container runtime — skipping container start"
        );
        Ok(format!("noop-{sandbox_id}"))
    }

    async fn stop_container(&self, container_id: &str) -> Result<()> {
        info!(
            container_id = %container_id,
            "noop container runtime — skipping container stop"
        );
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn noop_runtime_start_stop() {
        let runtime = NoopContainerRuntime;
        let container_id = runtime
            .start_container("sb-001", "python:3.12", HashMap::new(), 512 * 1024 * 1024, 100_000)
            .await
            .unwrap();
        assert!(container_id.contains("sb-001"));
        runtime.stop_container(&container_id).await.unwrap();
    }
}

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
    /// The `isolation_tier` parameter controls the security profile:
    /// - "standard": normal Docker (cgroups + namespaces)
    /// - "hardened": seccomp, read-only rootfs, no-new-privileges, dropped caps
    /// - "isolated": gVisor/Kata runtime + hardened options
    async fn start_container(
        &self,
        sandbox_id: &str,
        image: &str,
        env_vars: HashMap<String, String>,
        memory_bytes: i64,
        cpu_quota: i64,
        egress_allowlist: &[String],
        isolation_tier: &str,
    ) -> Result<String>;

    /// Stop and remove a container.
    async fn stop_container(&self, container_id: &str) -> Result<()>;

    /// Clean up any network/egress rules associated with this sandbox.
    async fn cleanup_egress_rules(&self, sandbox_id: &str) -> Result<()>;
}

/// Derive a short chain name from a sandbox ID (iptables chain names max 28 chars).
/// Format: BH-{first 12 chars of sandbox_id}
fn chain_name(sandbox_id: &str) -> String {
    let short = &sandbox_id[..sandbox_id.len().min(12)];
    format!("BH-{short}")
}

/// Run an iptables command. Returns Ok(()) on success, Err on failure.
async fn run_iptables(args: &[&str]) -> Result<()> {
    let output = tokio::process::Command::new("iptables")
        .args(args)
        .output()
        .await
        .map_err(|e| anyhow::anyhow!("failed to run iptables: {e}"))?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        return Err(anyhow::anyhow!(
            "iptables {:?} failed: {}",
            args,
            stderr.trim()
        ));
    }
    Ok(())
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

    /// Apply iptables egress rules for a sandbox container.
    /// Creates a per-sandbox chain with ACCEPT rules for allowed destinations
    /// and a default DROP at the end.
    async fn apply_egress_rules(
        &self,
        sandbox_id: &str,
        container_id: &str,
        egress_allowlist: &[String],
    ) -> Result<()> {
        // Get container IP from Docker inspect
        let inspect = self
            .client
            .inspect_container(container_id, None)
            .await
            .map_err(|e| anyhow::anyhow!("failed to inspect container: {e}"))?;

        let container_ip = inspect
            .network_settings
            .as_ref()
            .and_then(|ns| ns.ip_address.as_deref())
            .unwrap_or("");

        if container_ip.is_empty() {
            warn!(
                sandbox_id = %sandbox_id,
                "container has no IP address, skipping egress rules"
            );
            return Ok(());
        }

        let chain = chain_name(sandbox_id);
        info!(
            sandbox_id = %sandbox_id,
            chain = %chain,
            container_ip = %container_ip,
            rules = egress_allowlist.len(),
            "applying egress allowlist rules"
        );

        // 1. Create the per-sandbox chain
        if let Err(e) = run_iptables(&["-N", &chain]).await {
            warn!(error = %e, chain = %chain, "chain may already exist");
        }

        // 2. Add jump rule on FORWARD chain for traffic from this container
        if let Err(e) = run_iptables(&["-I", "FORWARD", "-s", container_ip, "-j", &chain]).await {
            warn!(error = %e, "failed to add FORWARD jump rule");
        }

        // 3. Allow established/related connections (return traffic)
        let _ = run_iptables(&[
            "-A",
            &chain,
            "-m",
            "conntrack",
            "--ctstate",
            "ESTABLISHED,RELATED",
            "-j",
            "ACCEPT",
        ])
        .await;

        // 4. Allow DNS (UDP and TCP port 53)
        let _ = run_iptables(&["-A", &chain, "-p", "udp", "--dport", "53", "-j", "ACCEPT"]).await;
        let _ = run_iptables(&["-A", &chain, "-p", "tcp", "--dport", "53", "-j", "ACCEPT"]).await;

        // 5. Allow traffic to the Runtime endpoint (so the SDK can call back)
        if let Ok(advertise) = std::env::var("ADVERTISE_ADDRESS") {
            let port = std::env::var("GRPC_PORT").unwrap_or_else(|_| "50052".to_string());
            let _ = run_iptables(&[
                "-A", &chain, "-d", &advertise, "-p", "tcp", "--dport", &port, "-j", "ACCEPT",
            ])
            .await;
        }

        // 6. Add ACCEPT rules for each allowed destination
        for dest in egress_allowlist {
            if let Err(e) = run_iptables(&["-A", &chain, "-d", dest, "-j", "ACCEPT"]).await {
                warn!(error = %e, dest = %dest, "failed to add egress ACCEPT rule");
            }
        }

        // 7. Default DROP at end of chain
        if let Err(e) = run_iptables(&["-A", &chain, "-j", "DROP"]).await {
            warn!(error = %e, "failed to add default DROP rule");
        }

        info!(sandbox_id = %sandbox_id, chain = %chain, "egress rules applied");
        Ok(())
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
        egress_allowlist: &[String],
        isolation_tier: &str,
    ) -> Result<String> {
        // 1. Pull image (best-effort — may already exist locally)
        use futures_util::StreamExt;
        info!(image = %image, sandbox_id = %sandbox_id, isolation_tier = %isolation_tier, "pulling container image");
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
        let mut env: Vec<String> = env_vars.iter().map(|(k, v)| format!("{k}={v}")).collect();
        // Inject Bulkhead SDK connection vars
        if let Ok(endpoint) = std::env::var("ADVERTISE_ADDRESS") {
            let port = std::env::var("GRPC_PORT").unwrap_or_else(|_| "50052".to_string());
            env.push(format!("BULKHEAD_ENDPOINT={endpoint}:{port}"));
        }
        env.push(format!("BULKHEAD_SANDBOX_ID={sandbox_id}"));

        // 3. Create container with resource limits and tier-specific security profile
        let container_name = format!("bulkhead-{sandbox_id}");
        let host_config = build_host_config(isolation_tier, memory_bytes, cpu_quota);

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

        // 5. Apply egress allowlist rules if specified
        if !egress_allowlist.is_empty() {
            if let Err(e) = self
                .apply_egress_rules(sandbox_id, &container_id, egress_allowlist)
                .await
            {
                warn!(
                    error = %e,
                    sandbox_id = %sandbox_id,
                    "failed to apply egress rules — container running without network restrictions"
                );
            }
        }

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

    async fn cleanup_egress_rules(&self, sandbox_id: &str) -> Result<()> {
        let chain = chain_name(sandbox_id);
        info!(sandbox_id = %sandbox_id, chain = %chain, "cleaning up egress rules");

        // Remove the FORWARD jump rule (best-effort, may not exist)
        let _ = run_iptables(&["-D", "FORWARD", "-j", &chain]).await;

        // Flush the chain
        let _ = run_iptables(&["-F", &chain]).await;

        // Delete the chain
        if let Err(e) = run_iptables(&["-X", &chain]).await {
            warn!(error = %e, chain = %chain, "failed to delete iptables chain");
        }

        Ok(())
    }
}

/// Build a Docker HostConfig based on the isolation tier.
///
/// - "standard": normal Docker (memory + cpu_quota only)
/// - "hardened": adds seccomp default profile, read-only rootfs, no-new-privileges, drops caps
/// - "isolated": uses gVisor/Kata runtime (ISOLATED_RUNTIME env var, default "runsc") + hardened opts
fn build_host_config(
    isolation_tier: &str,
    memory_bytes: i64,
    cpu_quota: i64,
) -> bollard::models::HostConfig {
    match isolation_tier {
        "hardened" => bollard::models::HostConfig {
            memory: Some(memory_bytes),
            cpu_quota: Some(cpu_quota),
            readonly_rootfs: Some(true),
            security_opt: Some(vec![
                "no-new-privileges:true".to_string(),
                "seccomp=default".to_string(),
            ]),
            cap_drop: Some(vec!["ALL".to_string()]),
            cap_add: Some(vec!["NET_BIND_SERVICE".to_string()]),
            ..Default::default()
        },
        "isolated" => {
            let runtime = std::env::var("ISOLATED_RUNTIME").unwrap_or_else(|_| "runsc".to_string());
            bollard::models::HostConfig {
                memory: Some(memory_bytes),
                cpu_quota: Some(cpu_quota),
                readonly_rootfs: Some(true),
                security_opt: Some(vec!["no-new-privileges:true".to_string()]),
                cap_drop: Some(vec!["ALL".to_string()]),
                cap_add: Some(vec!["NET_BIND_SERVICE".to_string()]),
                runtime: Some(runtime),
                ..Default::default()
            }
        }
        _ => {
            // "standard" or empty/unknown — current behavior
            bollard::models::HostConfig {
                memory: Some(memory_bytes),
                cpu_quota: Some(cpu_quota),
                ..Default::default()
            }
        }
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
        _egress_allowlist: &[String],
        _isolation_tier: &str,
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

    async fn cleanup_egress_rules(&self, sandbox_id: &str) -> Result<()> {
        info!(
            sandbox_id = %sandbox_id,
            "noop container runtime — skipping egress rule cleanup"
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
            .start_container(
                "sb-001",
                "python:3.12",
                HashMap::new(),
                512 * 1024 * 1024,
                100_000,
                &[],
                "standard",
            )
            .await
            .unwrap();
        assert!(container_id.contains("sb-001"));
        runtime.stop_container(&container_id).await.unwrap();
    }

    #[tokio::test]
    async fn noop_runtime_with_egress_allowlist() {
        let runtime = NoopContainerRuntime;
        let allowlist = vec!["api.example.com".to_string(), "10.0.0.0/8".to_string()];
        let container_id = runtime
            .start_container(
                "sb-002",
                "python:3.12",
                HashMap::new(),
                512 * 1024 * 1024,
                100_000,
                &allowlist,
                "standard",
            )
            .await
            .unwrap();
        assert!(container_id.contains("sb-002"));
        runtime.cleanup_egress_rules("sb-002").await.unwrap();
        runtime.stop_container(&container_id).await.unwrap();
    }

    #[test]
    fn chain_name_truncates() {
        assert_eq!(chain_name("abcdef123456789"), "BH-abcdef123456");
        assert_eq!(chain_name("short"), "BH-short");
    }

    #[test]
    fn build_host_config_standard() {
        let cfg = build_host_config("standard", 512 * 1024 * 1024, 100_000);
        assert_eq!(cfg.memory, Some(512 * 1024 * 1024));
        assert_eq!(cfg.cpu_quota, Some(100_000));
        assert!(cfg.readonly_rootfs.is_none());
        assert!(cfg.cap_drop.is_none());
        assert!(cfg.runtime.is_none());
    }

    #[test]
    fn build_host_config_hardened() {
        let cfg = build_host_config("hardened", 256 * 1024 * 1024, 50_000);
        assert_eq!(cfg.memory, Some(256 * 1024 * 1024));
        assert_eq!(cfg.readonly_rootfs, Some(true));
        assert_eq!(cfg.cap_drop, Some(vec!["ALL".to_string()]));
        assert_eq!(cfg.cap_add, Some(vec!["NET_BIND_SERVICE".to_string()]));
        let security_opts = cfg.security_opt.expect("hardened should set security_opt");
        assert!(security_opts.contains(&"no-new-privileges:true".to_string()));
        assert!(security_opts.contains(&"seccomp=default".to_string()));
        // Hardened does not set a custom runtime.
        assert!(cfg.runtime.is_none());
    }

    #[test]
    fn build_host_config_isolated() {
        let cfg = build_host_config("isolated", 1024 * 1024 * 1024, 200_000);
        assert_eq!(cfg.readonly_rootfs, Some(true));
        assert_eq!(cfg.cap_drop, Some(vec!["ALL".to_string()]));
        // Isolated sets a custom runtime (default: runsc).
        assert!(cfg.runtime.is_some());
        assert_eq!(cfg.runtime.unwrap(), "runsc");
    }

    #[test]
    fn build_host_config_unknown_tier_falls_back_to_standard() {
        let cfg = build_host_config("unknown", 512 * 1024 * 1024, 100_000);
        assert!(cfg.readonly_rootfs.is_none());
        assert!(cfg.cap_drop.is_none());
        assert!(cfg.runtime.is_none());
    }
}

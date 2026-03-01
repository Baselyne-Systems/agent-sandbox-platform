use std::collections::HashMap;
use std::sync::atomic::{AtomicU32, Ordering};
use std::sync::{Arc, Mutex, RwLock};
use std::time::SystemTime;

use anyhow::{anyhow, Result};
use tokio::sync::broadcast;
use tracing::info;
use uuid::Uuid;

use guardrails_eval::Evaluator;

use proto_gen::platform::host_agent::v1::ToolDefinition;

use crate::container::ContainerRuntime;

/// Status of a sandbox.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SandboxStatus {
    Starting,
    Running,
    Stopped,
    Failed,
}

/// Events emitted by a sandbox for streaming via HostAgentService.StreamEvents.
#[derive(Debug, Clone)]
pub enum SandboxEvent {
    Action {
        action_id: String,
        tool_name: String,
        verdict: String,
        evaluation_latency_us: i64,
    },
    Lifecycle {
        new_state: String,
        reason: String,
    },
    Progress {
        message: String,
        percent_complete: f32,
    },
}

/// Parameters for creating a new sandbox.
pub struct CreateSandboxParams {
    pub workspace_id: String,
    pub agent_id: String,
    pub allowed_tools: Vec<String>,
    pub env_vars: HashMap<String, String>,
    pub compiled_guardrails: Vec<u8>,
    pub container_image: String,
    pub egress_allowlist: Vec<String>,
    pub isolation_tier: String,
    /// MCP-compatible tool schemas registered for this sandbox.
    pub tool_definitions: Vec<ToolDefinition>,
}

/// Full per-sandbox state. Stored behind `Arc` for lock-free reads from async tasks.
pub struct SandboxState {
    /// Unique identifier for this sandbox.
    pub id: String,
    /// The workspace this sandbox belongs to.
    pub workspace_id: String,
    /// The agent running in this sandbox.
    pub agent_id: String,
    /// Current status.
    pub status: Mutex<SandboxStatus>,
    /// The guardrails evaluator loaded with the sandbox's compiled policy.
    /// Wrapped in RwLock to support hot-reload via UpdateSandboxGuardrails RPC.
    pub evaluator: RwLock<Evaluator>,
    /// Tools the agent is allowed to invoke.
    pub allowed_tools: Vec<String>,
    /// Environment variables injected into the sandbox.
    pub env_vars: HashMap<String, String>,
    /// Docker image for the sandbox container.
    pub container_image: String,
    /// Approved destination hosts/CIDRs for egress filtering.
    pub egress_allowlist: Vec<String>,
    /// Security isolation tier (standard, hardened, isolated).
    pub isolation_tier: String,
    /// MCP-compatible tool schemas registered for this sandbox.
    pub tool_definitions: Vec<ToolDefinition>,
    /// Container ID if a container was started for this sandbox.
    pub container_id: Mutex<Option<String>>,
    /// Counter of actions executed in this sandbox.
    pub actions_executed: AtomicU32,
    /// Broadcast sender for sandbox events (subscribe for StreamEvents).
    pub event_tx: broadcast::Sender<SandboxEvent>,
    /// When the sandbox was created.
    pub created_at: SystemTime,
}

impl std::fmt::Debug for SandboxState {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("SandboxState")
            .field("id", &self.id)
            .field("workspace_id", &self.workspace_id)
            .field("agent_id", &self.agent_id)
            .field("status", &self.status)
            .field("allowed_tools", &self.allowed_tools)
            .field(
                "actions_executed",
                &self.actions_executed.load(Ordering::Relaxed),
            )
            .field("created_at", &self.created_at)
            .finish()
    }
}

/// Manages the lifecycle of sandboxes on this host.
#[derive(Debug, Clone)]
pub struct SandboxManager {
    sandboxes: Arc<Mutex<HashMap<String, Arc<SandboxState>>>>,
    container_runtime: Arc<dyn ContainerRuntime>,
}

impl SandboxManager {
    /// Create a new sandbox manager with the given container runtime.
    pub fn new(container_runtime: Arc<dyn ContainerRuntime>) -> Self {
        Self {
            sandboxes: Arc::new(Mutex::new(HashMap::new())),
            container_runtime,
        }
    }

    /// Provision a new sandbox with full state, optionally starting a container.
    pub async fn create(&self, params: CreateSandboxParams) -> Result<Arc<SandboxState>> {
        let evaluator = Evaluator::load(&params.compiled_guardrails)
            .map_err(|e| anyhow!("failed to load guardrails: {e}"))?;

        let id = Uuid::new_v4().to_string();
        let (event_tx, _) = broadcast::channel(256);

        // Start container if an image is specified
        let container_id = if !params.container_image.is_empty() {
            let memory_bytes = 512 * 1024 * 1024; // default 512MB
            let cpu_quota = 100_000; // default 1 CPU
            let tier = if params.isolation_tier.is_empty() {
                "standard"
            } else {
                &params.isolation_tier
            };
            match self
                .container_runtime
                .start_container(
                    &id,
                    &params.container_image,
                    params.env_vars.clone(),
                    memory_bytes,
                    cpu_quota,
                    &params.egress_allowlist,
                    tier,
                )
                .await
            {
                Ok(cid) => Some(cid),
                Err(e) => {
                    return Err(anyhow!("failed to start container: {e}"));
                }
            }
        } else {
            None
        };

        let state = Arc::new(SandboxState {
            id: id.clone(),
            workspace_id: params.workspace_id.clone(),
            agent_id: params.agent_id.clone(),
            status: Mutex::new(SandboxStatus::Running),
            evaluator: RwLock::new(evaluator),
            allowed_tools: params.allowed_tools,
            env_vars: params.env_vars,
            container_image: params.container_image,
            egress_allowlist: params.egress_allowlist,
            isolation_tier: params.isolation_tier,
            tool_definitions: params.tool_definitions,
            container_id: Mutex::new(container_id),
            actions_executed: AtomicU32::new(0),
            event_tx: event_tx.clone(),
            created_at: SystemTime::now(),
        });

        info!(
            sandbox_id = %id,
            workspace_id = %params.workspace_id,
            agent_id = %params.agent_id,
            "sandbox provisioned"
        );

        // Send lifecycle started event (ignore error if no receivers)
        let _ = event_tx.send(SandboxEvent::Lifecycle {
            new_state: "running".into(),
            reason: "sandbox created".into(),
        });

        let mut sandboxes = self
            .sandboxes
            .lock()
            .map_err(|e| anyhow!("lock poisoned: {e}"))?;
        sandboxes.insert(id, state.clone());

        Ok(state)
    }

    /// Destroy (tear down) a sandbox by ID, stopping its container if one exists.
    pub async fn destroy(&self, sandbox_id: &str) -> Result<()> {
        let state = {
            let mut sandboxes = self
                .sandboxes
                .lock()
                .map_err(|e| anyhow!("lock poisoned: {e}"))?;
            sandboxes
                .remove(sandbox_id)
                .ok_or_else(|| anyhow!("sandbox not found: {sandbox_id}"))?
        };

        // Clean up egress rules before stopping the container
        if let Err(e) = self
            .container_runtime
            .cleanup_egress_rules(sandbox_id)
            .await
        {
            tracing::warn!(
                error = %e,
                sandbox_id = %sandbox_id,
                "failed to clean up egress rules — proceeding with sandbox teardown"
            );
        }

        // Stop container if one was started
        let container_id = state
            .container_id
            .lock()
            .map_err(|e| anyhow!("lock poisoned: {e}"))?
            .clone();
        if let Some(cid) = container_id {
            if let Err(e) = self.container_runtime.stop_container(&cid).await {
                tracing::warn!(
                    error = %e,
                    container_id = %cid,
                    sandbox_id = %sandbox_id,
                    "failed to stop container — proceeding with sandbox teardown"
                );
            }
        }

        // Update status
        if let Ok(mut status) = state.status.lock() {
            *status = SandboxStatus::Stopped;
        }
        // Send lifecycle stopped event
        let _ = state.event_tx.send(SandboxEvent::Lifecycle {
            new_state: "stopped".into(),
            reason: "sandbox destroyed".into(),
        });
        info!(sandbox_id = %sandbox_id, "sandbox destroyed");
        Ok(())
    }

    /// Retrieve sandbox state by ID.
    pub fn get_sandbox(&self, sandbox_id: &str) -> Result<Arc<SandboxState>> {
        let sandboxes = self
            .sandboxes
            .lock()
            .map_err(|e| anyhow!("lock poisoned: {e}"))?;

        sandboxes
            .get(sandbox_id)
            .cloned()
            .ok_or_else(|| anyhow!("sandbox not found: {sandbox_id}"))
    }

    /// Hot-reload guardrails for a running sandbox.
    pub fn update_guardrails(&self, sandbox_id: &str, compiled_guardrails: &[u8]) -> Result<()> {
        let sandbox = self.get_sandbox(sandbox_id)?;

        let new_evaluator = Evaluator::load(compiled_guardrails)
            .map_err(|e| anyhow!("failed to load guardrails: {e}"))?;

        let mut evaluator = sandbox
            .evaluator
            .write()
            .map_err(|e| anyhow!("evaluator lock poisoned: {e}"))?;
        *evaluator = new_evaluator;

        info!(sandbox_id = %sandbox_id, "guardrails updated");

        let _ = sandbox.event_tx.send(SandboxEvent::Lifecycle {
            new_state: "running".into(),
            reason: "guardrails updated".into(),
        });

        Ok(())
    }

    /// Returns the number of currently active sandboxes.
    pub fn active_count(&self) -> usize {
        self.sandboxes.lock().map(|s| s.len()).unwrap_or(0)
    }

    /// Look up a sandbox ID from gRPC request metadata.
    pub fn lookup_by_metadata<T>(&self, request: &tonic::Request<T>) -> Result<Arc<SandboxState>> {
        let sandbox_id = request
            .metadata()
            .get("x-sandbox-id")
            .and_then(|v| v.to_str().ok())
            .ok_or_else(|| anyhow!("missing x-sandbox-id metadata"))?;

        self.get_sandbox(sandbox_id)
    }
}

impl Default for SandboxManager {
    fn default() -> Self {
        Self::new(Arc::new(crate::container::NoopContainerRuntime))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::container::NoopContainerRuntime;

    fn empty_policy_bytes() -> Vec<u8> {
        serde_json::to_vec(&guardrails_eval::CompiledPolicy { rules: vec![] }).unwrap()
    }

    fn test_mgr() -> SandboxManager {
        SandboxManager::new(Arc::new(NoopContainerRuntime))
    }

    fn test_params(workspace_id: &str, agent_id: &str) -> CreateSandboxParams {
        CreateSandboxParams {
            workspace_id: workspace_id.to_string(),
            agent_id: agent_id.to_string(),
            allowed_tools: vec!["read_file".to_string(), "write_file".to_string()],
            env_vars: HashMap::from([("ENV".to_string(), "test".to_string())]),
            compiled_guardrails: empty_policy_bytes(),
            container_image: String::new(),
            egress_allowlist: vec![],
            isolation_tier: "standard".to_string(),
            tool_definitions: vec![],
        }
    }

    #[tokio::test]
    async fn create_and_destroy_sandbox() {
        let mgr = test_mgr();
        let state = mgr
            .create(test_params("ws-001", "agent-001"))
            .await
            .unwrap();
        assert_eq!(*state.status.lock().unwrap(), SandboxStatus::Running);
        assert_eq!(state.workspace_id, "ws-001");
        assert_eq!(state.agent_id, "agent-001");
        assert_eq!(state.allowed_tools, vec!["read_file", "write_file"]);

        let fetched = mgr.get_sandbox(&state.id).unwrap();
        assert_eq!(fetched.id, state.id);

        mgr.destroy(&state.id).await.unwrap();
        assert!(mgr.get_sandbox(&state.id).is_err());
    }

    #[tokio::test]
    async fn destroy_nonexistent_returns_error() {
        let mgr = test_mgr();
        assert!(mgr.destroy("no-such-sandbox").await.is_err());
    }

    #[tokio::test]
    async fn sandbox_has_event_channel() {
        let mgr = test_mgr();
        let state = mgr
            .create(test_params("ws-002", "agent-002"))
            .await
            .unwrap();

        let mut rx = state.event_tx.subscribe();
        let _ = state.event_tx.send(SandboxEvent::Progress {
            message: "test".into(),
            percent_complete: 50.0,
        });

        let event = rx.try_recv().unwrap();
        match event {
            SandboxEvent::Progress { message, .. } => assert_eq!(message, "test"),
            _ => panic!("unexpected event type"),
        }
    }

    #[tokio::test]
    async fn actions_executed_counter() {
        let mgr = test_mgr();
        let state = mgr
            .create(test_params("ws-003", "agent-003"))
            .await
            .unwrap();

        assert_eq!(state.actions_executed.load(Ordering::Relaxed), 0);
        state.actions_executed.fetch_add(1, Ordering::Relaxed);
        assert_eq!(state.actions_executed.load(Ordering::Relaxed), 1);
    }

    #[tokio::test]
    async fn update_guardrails_replaces_policy() {
        let mgr = test_mgr();
        let state = mgr
            .create(test_params("ws-005", "agent-005"))
            .await
            .unwrap();

        let policy_with_rule = serde_json::to_vec(&guardrails_eval::CompiledPolicy {
            rules: vec![guardrails_eval::CompiledRule {
                id: "r1".to_string(),
                name: "block-dangerous".to_string(),
                rule_type: guardrails_eval::RuleType::ToolFilter,
                condition: "dangerous_tool".to_string(),
                action: guardrails_eval::RuleAction::Deny,
                priority: 1,
                enabled: true,
                scope: None,
            }],
        })
        .unwrap();

        mgr.update_guardrails(&state.id, &policy_with_rule).unwrap();

        let ctx = guardrails_eval::EvalContext {
            tool_name: "dangerous_tool".to_string(),
            parameters: serde_json::Value::Null,
            agent_id: "agent-005".to_string(),
            trust_level: String::new(),
            data_classification: String::new(),
        };
        let evaluator = state.evaluator.read().unwrap();
        let verdict = evaluator.evaluate(&ctx);
        assert!(matches!(verdict, guardrails_eval::Verdict::Deny(_)));
    }

    #[test]
    fn update_guardrails_nonexistent_sandbox() {
        let mgr = test_mgr();
        let result = mgr.update_guardrails("no-such-sandbox", &empty_policy_bytes());
        assert!(result.is_err());
    }

    #[tokio::test]
    async fn update_guardrails_invalid_bytes() {
        let mgr = test_mgr();
        let state = mgr
            .create(test_params("ws-006", "agent-006"))
            .await
            .unwrap();
        let result = mgr.update_guardrails(&state.id, b"not valid json");
        assert!(result.is_err());
    }

    #[tokio::test]
    async fn update_guardrails_emits_lifecycle_event() {
        let mgr = test_mgr();
        let state = mgr
            .create(test_params("ws-007", "agent-007"))
            .await
            .unwrap();

        let mut rx = state.event_tx.subscribe();
        mgr.update_guardrails(&state.id, &empty_policy_bytes())
            .unwrap();

        let event = rx.try_recv().unwrap();
        match event {
            SandboxEvent::Lifecycle { reason, .. } => assert_eq!(reason, "guardrails updated"),
            _ => panic!("expected lifecycle event"),
        }
    }

    #[tokio::test]
    async fn destroy_sends_stopped_event() {
        let mgr = test_mgr();
        let state = mgr
            .create(test_params("ws-004", "agent-004"))
            .await
            .unwrap();

        let mut rx = state.event_tx.subscribe();
        mgr.destroy(&state.id).await.unwrap();

        let event = rx.try_recv().unwrap();
        match event {
            SandboxEvent::Lifecycle { new_state, .. } => assert_eq!(new_state, "stopped"),
            _ => panic!("expected lifecycle event"),
        }
    }

    #[tokio::test]
    async fn create_with_egress_allowlist() {
        let mgr = test_mgr();
        let mut params = test_params("ws-011", "agent-011");
        params.egress_allowlist = vec!["api.example.com".to_string(), "10.0.0.0/8".to_string()];
        let state = mgr.create(params).await.unwrap();

        assert_eq!(state.egress_allowlist.len(), 2);
        assert_eq!(state.egress_allowlist[0], "api.example.com");
        assert_eq!(state.egress_allowlist[1], "10.0.0.0/8");
    }

    #[tokio::test]
    async fn create_with_container_image() {
        let mgr = test_mgr();
        let mut params = test_params("ws-010", "agent-010");
        params.container_image = "python:3.12".to_string();
        let state = mgr.create(params).await.unwrap();

        let container_id = state.container_id.lock().unwrap().clone();
        assert!(
            container_id.is_some(),
            "container should be started for non-empty image"
        );
        assert!(container_id.unwrap().starts_with("noop-"));
    }

    #[tokio::test]
    async fn active_count_tracks_sandboxes() {
        let mgr = test_mgr();
        assert_eq!(mgr.active_count(), 0);

        let s1 = mgr
            .create(test_params("ws-100", "agent-100"))
            .await
            .unwrap();
        assert_eq!(mgr.active_count(), 1);

        let s2 = mgr
            .create(test_params("ws-101", "agent-101"))
            .await
            .unwrap();
        assert_eq!(mgr.active_count(), 2);

        mgr.destroy(&s1.id).await.unwrap();
        assert_eq!(mgr.active_count(), 1);

        mgr.destroy(&s2.id).await.unwrap();
        assert_eq!(mgr.active_count(), 0);
    }
}

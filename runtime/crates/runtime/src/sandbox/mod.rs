use std::collections::HashMap;
use std::sync::atomic::{AtomicU32, Ordering};
use std::sync::{Arc, Mutex, RwLock};
use std::time::SystemTime;

use anyhow::{anyhow, Result};
use tokio::sync::broadcast;
use tracing::info;
use uuid::Uuid;

use guardrails_eval::Evaluator;

/// Status of a sandbox.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SandboxStatus {
    Starting,
    Running,
    Stopped,
    Failed,
}

/// Events emitted by a sandbox for streaming via RuntimeService.StreamEvents.
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
            .field("actions_executed", &self.actions_executed.load(Ordering::Relaxed))
            .field("created_at", &self.created_at)
            .finish()
    }
}

/// Manages the lifecycle of sandboxes on this host.
#[derive(Debug, Clone)]
pub struct SandboxManager {
    sandboxes: Arc<Mutex<HashMap<String, Arc<SandboxState>>>>,
}

impl SandboxManager {
    /// Create a new, empty sandbox manager.
    pub fn new() -> Self {
        Self {
            sandboxes: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    /// Provision a new sandbox with full state.
    pub fn create(&self, params: CreateSandboxParams) -> Result<Arc<SandboxState>> {
        let evaluator = Evaluator::load(&params.compiled_guardrails)
            .map_err(|e| anyhow!("failed to load guardrails: {e}"))?;

        let id = Uuid::new_v4().to_string();
        let (event_tx, _) = broadcast::channel(256);

        let state = Arc::new(SandboxState {
            id: id.clone(),
            workspace_id: params.workspace_id.clone(),
            agent_id: params.agent_id.clone(),
            status: Mutex::new(SandboxStatus::Running),
            evaluator: RwLock::new(evaluator),
            allowed_tools: params.allowed_tools,
            env_vars: params.env_vars,
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

    /// Destroy (tear down) a sandbox by ID.
    pub fn destroy(&self, sandbox_id: &str) -> Result<()> {
        let mut sandboxes = self
            .sandboxes
            .lock()
            .map_err(|e| anyhow!("lock poisoned: {e}"))?;

        if let Some(state) = sandboxes.remove(sandbox_id) {
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
        } else {
            Err(anyhow!("sandbox not found: {sandbox_id}"))
        }
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
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn empty_policy_bytes() -> Vec<u8> {
        serde_json::to_vec(&guardrails_eval::CompiledPolicy { rules: vec![] }).unwrap()
    }

    fn test_params(workspace_id: &str, agent_id: &str) -> CreateSandboxParams {
        CreateSandboxParams {
            workspace_id: workspace_id.to_string(),
            agent_id: agent_id.to_string(),
            allowed_tools: vec!["read_file".to_string(), "write_file".to_string()],
            env_vars: HashMap::from([("ENV".to_string(), "test".to_string())]),
            compiled_guardrails: empty_policy_bytes(),
        }
    }

    #[test]
    fn create_and_destroy_sandbox() {
        let mgr = SandboxManager::new();
        let state = mgr.create(test_params("ws-001", "agent-001")).unwrap();
        assert_eq!(*state.status.lock().unwrap(), SandboxStatus::Running);
        assert_eq!(state.workspace_id, "ws-001");
        assert_eq!(state.agent_id, "agent-001");
        assert_eq!(state.allowed_tools, vec!["read_file", "write_file"]);

        let fetched = mgr.get_sandbox(&state.id).unwrap();
        assert_eq!(fetched.id, state.id);

        mgr.destroy(&state.id).unwrap();
        assert!(mgr.get_sandbox(&state.id).is_err());
    }

    #[test]
    fn destroy_nonexistent_returns_error() {
        let mgr = SandboxManager::new();
        assert!(mgr.destroy("no-such-sandbox").is_err());
    }

    #[test]
    fn sandbox_has_event_channel() {
        let mgr = SandboxManager::new();
        let state = mgr.create(test_params("ws-002", "agent-002")).unwrap();

        // Subscribe and verify we can receive events
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

    #[test]
    fn actions_executed_counter() {
        let mgr = SandboxManager::new();
        let state = mgr.create(test_params("ws-003", "agent-003")).unwrap();

        assert_eq!(state.actions_executed.load(Ordering::Relaxed), 0);
        state.actions_executed.fetch_add(1, Ordering::Relaxed);
        assert_eq!(state.actions_executed.load(Ordering::Relaxed), 1);
    }

    #[test]
    fn update_guardrails_replaces_policy() {
        let mgr = SandboxManager::new();
        let state = mgr.create(test_params("ws-005", "agent-005")).unwrap();

        // Build a policy with one deny rule
        let policy_with_rule = serde_json::to_vec(&guardrails_eval::CompiledPolicy {
            rules: vec![guardrails_eval::CompiledRule {
                id: "r1".to_string(),
                name: "block-dangerous".to_string(),
                rule_type: guardrails_eval::RuleType::ToolFilter,
                condition: "dangerous_tool".to_string(),
                action: guardrails_eval::RuleAction::Deny,
                priority: 1,
                enabled: true,
            }],
        })
        .unwrap();

        mgr.update_guardrails(&state.id, &policy_with_rule).unwrap();

        // Verify the new policy is active by evaluating
        let ctx = guardrails_eval::EvalContext {
            tool_name: "dangerous_tool".to_string(),
            parameters: serde_json::Value::Null,
            agent_id: "agent-005".to_string(),
        };
        let evaluator = state.evaluator.read().unwrap();
        let verdict = evaluator.evaluate(&ctx);
        assert!(matches!(verdict, guardrails_eval::Verdict::Deny(_)));
    }

    #[test]
    fn update_guardrails_nonexistent_sandbox() {
        let mgr = SandboxManager::new();
        let result = mgr.update_guardrails("no-such-sandbox", &empty_policy_bytes());
        assert!(result.is_err());
    }

    #[test]
    fn update_guardrails_invalid_bytes() {
        let mgr = SandboxManager::new();
        let state = mgr.create(test_params("ws-006", "agent-006")).unwrap();
        let result = mgr.update_guardrails(&state.id, b"not valid json");
        assert!(result.is_err());
    }

    #[test]
    fn update_guardrails_emits_lifecycle_event() {
        let mgr = SandboxManager::new();
        let state = mgr.create(test_params("ws-007", "agent-007")).unwrap();

        let mut rx = state.event_tx.subscribe();
        mgr.update_guardrails(&state.id, &empty_policy_bytes()).unwrap();

        let event = rx.try_recv().unwrap();
        match event {
            SandboxEvent::Lifecycle { reason, .. } => assert_eq!(reason, "guardrails updated"),
            _ => panic!("expected lifecycle event"),
        }
    }

    #[test]
    fn destroy_sends_stopped_event() {
        let mgr = SandboxManager::new();
        let state = mgr.create(test_params("ws-004", "agent-004")).unwrap();

        let mut rx = state.event_tx.subscribe();
        mgr.destroy(&state.id).unwrap();

        let event = rx.try_recv().unwrap();
        match event {
            SandboxEvent::Lifecycle { new_state, .. } => assert_eq!(new_state, "stopped"),
            _ => panic!("expected lifecycle event"),
        }
    }
}

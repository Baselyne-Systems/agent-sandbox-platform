use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::time::SystemTime;

use anyhow::{anyhow, Result};
use tracing::info;
use uuid::Uuid;

/// Status of a sandbox.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SandboxStatus {
    Starting,
    Running,
    Stopped,
    Failed,
}

/// A handle to a managed sandbox instance.
#[derive(Debug, Clone)]
pub struct SandboxHandle {
    /// Unique identifier for this sandbox.
    pub id: String,
    /// Current status.
    pub status: SandboxStatus,
    /// When the sandbox was created.
    pub created_at: SystemTime,
}

/// Manages the lifecycle of sandboxes on this host.
///
/// Uses an in-memory HashMap behind `Arc<Mutex<>>` for now. A production
/// implementation would integrate with the container/microVM runtime.
#[derive(Debug, Clone)]
pub struct SandboxManager {
    sandboxes: Arc<Mutex<HashMap<String, SandboxHandle>>>,
}

impl SandboxManager {
    /// Create a new, empty sandbox manager.
    pub fn new() -> Self {
        Self {
            sandboxes: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    /// Provision a new sandbox for the given workspace.
    pub fn create(&self, workspace_id: String, spec: String) -> Result<SandboxHandle> {
        let id = Uuid::new_v4().to_string();
        let handle = SandboxHandle {
            id: id.clone(),
            status: SandboxStatus::Running,
            created_at: SystemTime::now(),
        };

        info!(
            sandbox_id = %id,
            workspace_id = %workspace_id,
            spec = %spec,
            "sandbox provisioned"
        );

        let mut sandboxes = self
            .sandboxes
            .lock()
            .map_err(|e| anyhow!("lock poisoned: {e}"))?;
        sandboxes.insert(id, handle.clone());

        Ok(handle)
    }

    /// Destroy (tear down) a sandbox by ID.
    pub fn destroy(&self, sandbox_id: &str) -> Result<()> {
        let mut sandboxes = self
            .sandboxes
            .lock()
            .map_err(|e| anyhow!("lock poisoned: {e}"))?;

        if sandboxes.remove(sandbox_id).is_some() {
            info!(sandbox_id = %sandbox_id, "sandbox destroyed");
            Ok(())
        } else {
            Err(anyhow!("sandbox not found: {sandbox_id}"))
        }
    }

    /// Retrieve the current status of a sandbox.
    pub fn get_status(&self, sandbox_id: &str) -> Result<SandboxHandle> {
        let sandboxes = self
            .sandboxes
            .lock()
            .map_err(|e| anyhow!("lock poisoned: {e}"))?;

        sandboxes
            .get(sandbox_id)
            .cloned()
            .ok_or_else(|| anyhow!("sandbox not found: {sandbox_id}"))
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

    #[test]
    fn create_and_destroy_sandbox() {
        let mgr = SandboxManager::new();
        let handle = mgr
            .create("ws-001".to_string(), "256MB/500mc".to_string())
            .unwrap();
        assert_eq!(handle.status, SandboxStatus::Running);

        let status = mgr.get_status(&handle.id).unwrap();
        assert_eq!(status.id, handle.id);

        mgr.destroy(&handle.id).unwrap();
        assert!(mgr.get_status(&handle.id).is_err());
    }

    #[test]
    fn destroy_nonexistent_returns_error() {
        let mgr = SandboxManager::new();
        assert!(mgr.destroy("no-such-sandbox").is_err());
    }
}

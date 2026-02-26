use serde::{Deserialize, Serialize};
use thiserror::Error;
use tracing::debug;

/// Errors that can occur during guardrails evaluation.
#[derive(Debug, Error)]
pub enum EvalError {
    #[error("failed to deserialize compiled policy: {0}")]
    DeserializationError(String),

    #[error("evaluation error: {0}")]
    EvaluationError(String),
}

/// A compiled guardrails policy, deserialized from bytes produced by the
/// control-plane GuardrailsService.CompilePolicy RPC.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompiledPolicy {
    /// Serialized rule data. The exact format is opaque to callers; the
    /// evaluator knows how to interpret it.
    pub rules: Vec<u8>,
}

/// Context provided to the evaluator for each tool call.
#[derive(Debug, Clone)]
pub struct EvalContext {
    /// The name of the tool being invoked.
    pub tool_name: String,
    /// The parameters passed to the tool.
    pub parameters: serde_json::Value,
    /// The identity of the agent making the call.
    pub agent_id: String,
}

/// The outcome of evaluating a single action against the loaded policy.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Verdict {
    /// The action is permitted.
    Allow,
    /// The action is denied with the given reason.
    Deny(String),
    /// The action requires human escalation with the given context.
    Escalate(String),
}

/// The guardrails evaluator. Loads a compiled policy and evaluates incoming
/// tool-call requests against it.
#[derive(Debug)]
pub struct Evaluator {
    _policy: CompiledPolicy,
}

impl Evaluator {
    /// Deserialize a [`CompiledPolicy`] from bytes and return a ready evaluator.
    pub fn load(bytes: &[u8]) -> Result<Self, EvalError> {
        let policy: CompiledPolicy = serde_json::from_slice(bytes).map_err(|e| {
            EvalError::DeserializationError(e.to_string())
        })?;
        debug!(rules_len = policy.rules.len(), "loaded compiled policy");
        Ok(Self { _policy: policy })
    }

    /// Evaluate the given context against the loaded policy.
    ///
    /// Stub implementation: always returns [`Verdict::Allow`].
    pub fn evaluate(&self, ctx: &EvalContext) -> Verdict {
        debug!(
            tool = %ctx.tool_name,
            agent = %ctx.agent_id,
            "evaluating action — stub: allowing"
        );
        Verdict::Allow
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn stub_evaluator_allows_everything() {
        let policy = CompiledPolicy {
            rules: Vec::new(),
        };
        let bytes = serde_json::to_vec(&policy).unwrap();
        let evaluator = Evaluator::load(&bytes).unwrap();

        let ctx = EvalContext {
            tool_name: "read_file".to_string(),
            parameters: serde_json::json!({"path": "/tmp/test.txt"}),
            agent_id: "agent-001".to_string(),
        };

        assert_eq!(evaluator.evaluate(&ctx), Verdict::Allow);
    }
}

use serde::{Deserialize, Serialize};
use thiserror::Error;
use tracing::{debug, warn};

/// Errors that can occur during guardrails evaluation.
#[derive(Debug, Error)]
pub enum EvalError {
    #[error("failed to deserialize compiled policy: {0}")]
    DeserializationError(String),

    #[error("evaluation error: {0}")]
    EvaluationError(String),
}

/// The type of a guardrail rule.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum RuleType {
    ToolFilter,
    ParameterCheck,
    RateLimit,
    BudgetLimit,
}

/// The action to take when a rule matches.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum RuleAction {
    Allow,
    Deny,
    Escalate,
    Log,
}

/// Scope restricts which agents/tools/trust-levels/data-classifications a rule applies to.
/// Empty or None means "match all".
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct RuleScope {
    #[serde(default)]
    pub agent_ids: Vec<String>,
    #[serde(default)]
    pub tool_names: Vec<String>,
    #[serde(default)]
    pub trust_levels: Vec<String>,
    #[serde(default)]
    pub data_classifications: Vec<String>,
}

/// A single compiled guardrail rule.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompiledRule {
    pub id: String,
    pub name: String,
    pub rule_type: RuleType,
    /// For ToolFilter: comma-separated tool names (e.g. "exec,shell,sudo").
    /// For ParameterCheck: "field_name=forbidden_value".
    pub condition: String,
    pub action: RuleAction,
    /// Lower number = higher priority.
    pub priority: i32,
    pub enabled: bool,
    /// Optional scope — restricts when this rule is evaluated.
    #[serde(default)]
    pub scope: Option<RuleScope>,
}

/// A compiled guardrails policy, deserialized from bytes produced by the
/// control-plane GuardrailsService.CompilePolicy RPC.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompiledPolicy {
    pub rules: Vec<CompiledRule>,
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
    /// The trust level of the agent (e.g. "new", "established", "trusted").
    pub trust_level: String,
    /// The data classification relevant to this call (e.g. "public", "confidential").
    pub data_classification: String,
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
    policy: CompiledPolicy,
}

impl Evaluator {
    /// Deserialize a [`CompiledPolicy`] from bytes and return a ready evaluator.
    pub fn load(bytes: &[u8]) -> Result<Self, EvalError> {
        let policy: CompiledPolicy = serde_json::from_slice(bytes)
            .map_err(|e| EvalError::DeserializationError(e.to_string()))?;
        debug!(rules_count = policy.rules.len(), "loaded compiled policy");
        Ok(Self { policy })
    }

    /// Evaluate the given context against the loaded policy.
    ///
    /// Rules are sorted by priority (ascending — lower number = higher priority).
    /// The first matching enabled rule determines the verdict.
    /// If no rule matches, the default verdict is `Allow`.
    pub fn evaluate(&self, ctx: &EvalContext) -> Verdict {
        let mut rules: Vec<&CompiledRule> =
            self.policy.rules.iter().filter(|r| r.enabled).collect();
        rules.sort_by_key(|r| r.priority);

        for rule in rules {
            if !self.scope_matches(rule, ctx) {
                continue;
            }
            if self.matches_rule(rule, ctx) {
                debug!(
                    rule_id = %rule.id,
                    rule_name = %rule.name,
                    tool = %ctx.tool_name,
                    agent = %ctx.agent_id,
                    action = ?rule.action,
                    "rule matched"
                );
                return match &rule.action {
                    RuleAction::Allow => Verdict::Allow,
                    RuleAction::Deny => {
                        Verdict::Deny(format!("denied by rule '{}' ({})", rule.name, rule.id))
                    }
                    RuleAction::Escalate => Verdict::Escalate(rule.id.clone()),
                    RuleAction::Log => {
                        debug!(
                            rule_id = %rule.id,
                            tool = %ctx.tool_name,
                            "log-only rule matched, allowing"
                        );
                        Verdict::Allow
                    }
                };
            }
        }

        debug!(
            tool = %ctx.tool_name,
            agent = %ctx.agent_id,
            "no rule matched — default allow"
        );
        Verdict::Allow
    }

    /// Check whether the rule's scope applies to the current context.
    /// Empty scope fields mean "match all".
    fn scope_matches(&self, rule: &CompiledRule, ctx: &EvalContext) -> bool {
        let scope = match &rule.scope {
            Some(s) => s,
            None => return true, // No scope = applies globally
        };
        if !scope.agent_ids.is_empty() && !scope.agent_ids.contains(&ctx.agent_id) {
            return false;
        }
        if !scope.tool_names.is_empty() && !scope.tool_names.contains(&ctx.tool_name) {
            return false;
        }
        if !scope.trust_levels.is_empty()
            && !ctx.trust_level.is_empty()
            && !scope.trust_levels.contains(&ctx.trust_level)
        {
            return false;
        }
        if !scope.data_classifications.is_empty()
            && !ctx.data_classification.is_empty()
            && !scope
                .data_classifications
                .contains(&ctx.data_classification)
        {
            return false;
        }
        true
    }

    /// Check whether a single rule matches the given context.
    fn matches_rule(&self, rule: &CompiledRule, ctx: &EvalContext) -> bool {
        match rule.rule_type {
            RuleType::ToolFilter => {
                // condition is a comma-separated list of tool names
                rule.condition
                    .split(',')
                    .map(|s| s.trim())
                    .any(|name| name == ctx.tool_name)
            }
            RuleType::ParameterCheck => {
                // condition is "field_name=forbidden_value"
                if let Some((field, value)) = rule.condition.split_once('=') {
                    let field = field.trim();
                    let value = value.trim();
                    match ctx.parameters.get(field) {
                        Some(serde_json::Value::String(s)) => s == value,
                        Some(v) => v.to_string().trim_matches('"') == value,
                        None => false,
                    }
                } else {
                    warn!(
                        rule_id = %rule.id,
                        condition = %rule.condition,
                        "invalid parameter_check condition format, expected 'field=value'"
                    );
                    false
                }
            }
            RuleType::RateLimit | RuleType::BudgetLimit => {
                // Evaluated at control-plane level, skip in runtime evaluator.
                false
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_policy(rules: Vec<CompiledRule>) -> Vec<u8> {
        serde_json::to_vec(&CompiledPolicy { rules }).unwrap()
    }

    fn make_ctx(tool: &str, params: serde_json::Value) -> EvalContext {
        EvalContext {
            tool_name: tool.to_string(),
            parameters: params,
            agent_id: "agent-001".to_string(),
            trust_level: String::new(),
            data_classification: String::new(),
        }
    }

    #[test]
    fn deny_by_tool_name() {
        let bytes = make_policy(vec![CompiledRule {
            id: "r1".into(),
            name: "deny-exec".into(),
            rule_type: RuleType::ToolFilter,
            condition: "exec,shell".into(),
            action: RuleAction::Deny,
            priority: 10,
            enabled: true,
            scope: None,
        }]);
        let evaluator = Evaluator::load(&bytes).unwrap();

        let ctx = make_ctx("exec", serde_json::json!({}));
        assert!(matches!(evaluator.evaluate(&ctx), Verdict::Deny(_)));

        let ctx = make_ctx("shell", serde_json::json!({}));
        assert!(matches!(evaluator.evaluate(&ctx), Verdict::Deny(_)));

        // Non-matching tool should be allowed
        let ctx = make_ctx("read_file", serde_json::json!({}));
        assert_eq!(evaluator.evaluate(&ctx), Verdict::Allow);
    }

    #[test]
    fn allow_by_priority() {
        // Lower priority number wins. Allow rule (priority 1) should beat
        // deny rule (priority 10) for the same tool.
        let bytes = make_policy(vec![
            CompiledRule {
                id: "r-deny".into(),
                name: "deny-read".into(),
                rule_type: RuleType::ToolFilter,
                condition: "read_file".into(),
                action: RuleAction::Deny,
                priority: 10,
                enabled: true,
                scope: None,
            },
            CompiledRule {
                id: "r-allow".into(),
                name: "allow-read".into(),
                rule_type: RuleType::ToolFilter,
                condition: "read_file".into(),
                action: RuleAction::Allow,
                priority: 1,
                enabled: true,
                scope: None,
            },
        ]);
        let evaluator = Evaluator::load(&bytes).unwrap();

        let ctx = make_ctx("read_file", serde_json::json!({}));
        assert_eq!(evaluator.evaluate(&ctx), Verdict::Allow);
    }

    #[test]
    fn parameter_check_match() {
        let bytes = make_policy(vec![CompiledRule {
            id: "r-param".into(),
            name: "deny-tmp-path".into(),
            rule_type: RuleType::ParameterCheck,
            condition: "path=/etc/shadow".into(),
            action: RuleAction::Deny,
            priority: 5,
            enabled: true,
            scope: None,
        }]);
        let evaluator = Evaluator::load(&bytes).unwrap();

        // Matching parameter
        let ctx = make_ctx("read_file", serde_json::json!({"path": "/etc/shadow"}));
        assert!(matches!(evaluator.evaluate(&ctx), Verdict::Deny(_)));

        // Non-matching parameter
        let ctx = make_ctx("read_file", serde_json::json!({"path": "/tmp/safe.txt"}));
        assert_eq!(evaluator.evaluate(&ctx), Verdict::Allow);
    }

    #[test]
    fn no_match_default_allow() {
        let bytes = make_policy(vec![CompiledRule {
            id: "r1".into(),
            name: "deny-exec".into(),
            rule_type: RuleType::ToolFilter,
            condition: "exec".into(),
            action: RuleAction::Deny,
            priority: 10,
            enabled: true,
            scope: None,
        }]);
        let evaluator = Evaluator::load(&bytes).unwrap();

        let ctx = make_ctx("read_file", serde_json::json!({}));
        assert_eq!(evaluator.evaluate(&ctx), Verdict::Allow);
    }

    #[test]
    fn escalate_verdict() {
        let bytes = make_policy(vec![CompiledRule {
            id: "r-esc".into(),
            name: "escalate-delete".into(),
            rule_type: RuleType::ToolFilter,
            condition: "delete_file".into(),
            action: RuleAction::Escalate,
            priority: 5,
            enabled: true,
            scope: None,
        }]);
        let evaluator = Evaluator::load(&bytes).unwrap();

        let ctx = make_ctx("delete_file", serde_json::json!({}));
        match evaluator.evaluate(&ctx) {
            Verdict::Escalate(id) => assert_eq!(id, "r-esc"),
            other => panic!("expected Escalate, got {:?}", other),
        }
    }

    #[test]
    fn disabled_rule_skip() {
        let bytes = make_policy(vec![CompiledRule {
            id: "r1".into(),
            name: "deny-exec-disabled".into(),
            rule_type: RuleType::ToolFilter,
            condition: "exec".into(),
            action: RuleAction::Deny,
            priority: 1,
            enabled: false,
            scope: None,
        }]);
        let evaluator = Evaluator::load(&bytes).unwrap();

        let ctx = make_ctx("exec", serde_json::json!({}));
        assert_eq!(evaluator.evaluate(&ctx), Verdict::Allow);
    }

    #[test]
    fn log_action_treated_as_allow() {
        let bytes = make_policy(vec![CompiledRule {
            id: "r-log".into(),
            name: "log-exec".into(),
            rule_type: RuleType::ToolFilter,
            condition: "exec".into(),
            action: RuleAction::Log,
            priority: 5,
            enabled: true,
            scope: None,
        }]);
        let evaluator = Evaluator::load(&bytes).unwrap();

        let ctx = make_ctx("exec", serde_json::json!({}));
        assert_eq!(evaluator.evaluate(&ctx), Verdict::Allow);
    }

    #[test]
    fn empty_policy_allows_everything() {
        let bytes = make_policy(vec![]);
        let evaluator = Evaluator::load(&bytes).unwrap();

        let ctx = make_ctx("anything", serde_json::json!({"key": "value"}));
        assert_eq!(evaluator.evaluate(&ctx), Verdict::Allow);
    }

    #[test]
    fn scope_agent_id_match() {
        let bytes = make_policy(vec![CompiledRule {
            id: "r1".into(),
            name: "deny-exec-for-agent".into(),
            rule_type: RuleType::ToolFilter,
            condition: "exec".into(),
            action: RuleAction::Deny,
            priority: 10,
            enabled: true,
            scope: Some(RuleScope {
                agent_ids: vec!["agent-001".into()],
                ..Default::default()
            }),
        }]);
        let evaluator = Evaluator::load(&bytes).unwrap();

        // Matching agent — should deny
        let ctx = make_ctx("exec", serde_json::json!({}));
        assert!(matches!(evaluator.evaluate(&ctx), Verdict::Deny(_)));

        // Different agent — scope doesn't match, should allow
        let ctx = EvalContext {
            tool_name: "exec".to_string(),
            parameters: serde_json::json!({}),
            agent_id: "agent-999".to_string(),
            trust_level: String::new(),
            data_classification: String::new(),
        };
        assert_eq!(evaluator.evaluate(&ctx), Verdict::Allow);
    }

    #[test]
    fn scope_trust_level_match() {
        let bytes = make_policy(vec![CompiledRule {
            id: "r1".into(),
            name: "deny-exec-for-new".into(),
            rule_type: RuleType::ToolFilter,
            condition: "exec".into(),
            action: RuleAction::Deny,
            priority: 10,
            enabled: true,
            scope: Some(RuleScope {
                trust_levels: vec!["new".into()],
                ..Default::default()
            }),
        }]);
        let evaluator = Evaluator::load(&bytes).unwrap();

        // "new" trust level — should deny
        let ctx = EvalContext {
            tool_name: "exec".to_string(),
            parameters: serde_json::json!({}),
            agent_id: "agent-001".to_string(),
            trust_level: "new".to_string(),
            data_classification: String::new(),
        };
        assert!(matches!(evaluator.evaluate(&ctx), Verdict::Deny(_)));

        // "trusted" trust level — scope doesn't match
        let ctx = EvalContext {
            tool_name: "exec".to_string(),
            parameters: serde_json::json!({}),
            agent_id: "agent-001".to_string(),
            trust_level: "trusted".to_string(),
            data_classification: String::new(),
        };
        assert_eq!(evaluator.evaluate(&ctx), Verdict::Allow);
    }

    #[test]
    fn scope_no_scope_applies_globally() {
        let bytes = make_policy(vec![CompiledRule {
            id: "r1".into(),
            name: "deny-exec-global".into(),
            rule_type: RuleType::ToolFilter,
            condition: "exec".into(),
            action: RuleAction::Deny,
            priority: 10,
            enabled: true,
            scope: None,
        }]);
        let evaluator = Evaluator::load(&bytes).unwrap();

        // Should deny for any agent
        let ctx = EvalContext {
            tool_name: "exec".to_string(),
            parameters: serde_json::json!({}),
            agent_id: "any-agent".to_string(),
            trust_level: "trusted".to_string(),
            data_classification: "confidential".to_string(),
        };
        assert!(matches!(evaluator.evaluate(&ctx), Verdict::Deny(_)));
    }
}

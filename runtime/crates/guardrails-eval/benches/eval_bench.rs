use criterion::{black_box, criterion_group, criterion_main, Criterion};
use guardrails_eval::{
    CompiledPolicy, CompiledRule, EvalContext, Evaluator, RuleAction, RuleScope, RuleType,
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Serialize a `CompiledPolicy` to bytes (JSON), matching what
/// `Evaluator::load` expects.
fn policy_bytes(rules: Vec<CompiledRule>) -> Vec<u8> {
    serde_json::to_vec(&CompiledPolicy { rules }).unwrap()
}

/// Generate `n` ToolFilter/Deny rules. Rule `i` blocks tool `"tool_{i}"`.
fn gen_tool_filter_rules(n: usize) -> Vec<CompiledRule> {
    (0..n)
        .map(|i| CompiledRule {
            id: format!("r{i}"),
            name: format!("deny-tool-{i}"),
            rule_type: RuleType::ToolFilter,
            condition: format!("tool_{i}"),
            action: RuleAction::Deny,
            priority: i as i32,
            enabled: true,
            scope: None,
        })
        .collect()
}

/// Build a simple `EvalContext` for the given tool name.
fn ctx(tool: &str) -> EvalContext {
    EvalContext {
        tool_name: tool.to_string(),
        parameters: serde_json::json!({}),
        agent_id: "agent-001".to_string(),
        trust_level: "established".to_string(),
        data_classification: "internal".to_string(),
    }
}

// ---------------------------------------------------------------------------
// evaluate benchmarks
// ---------------------------------------------------------------------------

fn evaluate_benchmarks(c: &mut Criterion) {
    let mut group = c.benchmark_group("evaluate");

    // -- single_rule: 1 ToolFilter deny rule for "shell", evaluate "shell" -> DENY
    {
        let bytes = policy_bytes(vec![CompiledRule {
            id: "r0".into(),
            name: "deny-shell".into(),
            rule_type: RuleType::ToolFilter,
            condition: "shell".into(),
            action: RuleAction::Deny,
            priority: 0,
            enabled: true,
            scope: None,
        }]);
        let evaluator = Evaluator::load(&bytes).unwrap();
        let eval_ctx = ctx("shell");

        group.bench_function("single_rule", |b| {
            b.iter(|| evaluator.evaluate(black_box(&eval_ctx)))
        });
    }

    // -- ten_rules: 10 ToolFilter rules, match on the 10th
    {
        let rules = gen_tool_filter_rules(10);
        let bytes = policy_bytes(rules);
        let evaluator = Evaluator::load(&bytes).unwrap();
        let eval_ctx = ctx("tool_9");

        group.bench_function("ten_rules", |b| {
            b.iter(|| evaluator.evaluate(black_box(&eval_ctx)))
        });
    }

    // -- hundred_rules: 100 rules, match on the 100th
    {
        let rules = gen_tool_filter_rules(100);
        let bytes = policy_bytes(rules);
        let evaluator = Evaluator::load(&bytes).unwrap();
        let eval_ctx = ctx("tool_99");

        group.bench_function("hundred_rules", |b| {
            b.iter(|| evaluator.evaluate(black_box(&eval_ctx)))
        });
    }

    // -- five_hundred_rules: 500 rules, worst-case scan
    {
        let rules = gen_tool_filter_rules(500);
        let bytes = policy_bytes(rules);
        let evaluator = Evaluator::load(&bytes).unwrap();
        let eval_ctx = ctx("tool_499");

        group.bench_function("five_hundred_rules", |b| {
            b.iter(|| evaluator.evaluate(black_box(&eval_ctx)))
        });
    }

    // -- no_match: 10 rules, evaluate a tool that matches none -> default ALLOW
    {
        let rules = gen_tool_filter_rules(10);
        let bytes = policy_bytes(rules);
        let evaluator = Evaluator::load(&bytes).unwrap();
        let eval_ctx = ctx("nonexistent_tool");

        group.bench_function("no_match", |b| {
            b.iter(|| evaluator.evaluate(black_box(&eval_ctx)))
        });
    }

    // -- with_scope: rules with agent_id and trust_level scopes
    {
        let rules = vec![
            // Rule that won't match scope (different agent)
            CompiledRule {
                id: "r-other-agent".into(),
                name: "deny-other-agent".into(),
                rule_type: RuleType::ToolFilter,
                condition: "shell".into(),
                action: RuleAction::Deny,
                priority: 0,
                enabled: true,
                scope: Some(RuleScope {
                    agent_ids: vec!["agent-other".into()],
                    trust_levels: vec!["new".into()],
                    ..Default::default()
                }),
            },
            // Rule that will match scope
            CompiledRule {
                id: "r-scoped".into(),
                name: "deny-scoped".into(),
                rule_type: RuleType::ToolFilter,
                condition: "shell".into(),
                action: RuleAction::Deny,
                priority: 1,
                enabled: true,
                scope: Some(RuleScope {
                    agent_ids: vec!["agent-001".into()],
                    trust_levels: vec!["established".into()],
                    ..Default::default()
                }),
            },
        ];
        let bytes = policy_bytes(rules);
        let evaluator = Evaluator::load(&bytes).unwrap();
        let eval_ctx = ctx("shell");

        group.bench_function("with_scope", |b| {
            b.iter(|| evaluator.evaluate(black_box(&eval_ctx)))
        });
    }

    // -- parameter_check: ParameterCheck rule for "path=/etc/shadow"
    {
        let rules = vec![CompiledRule {
            id: "r-param".into(),
            name: "deny-shadow-path".into(),
            rule_type: RuleType::ParameterCheck,
            condition: "path=/etc/shadow".into(),
            action: RuleAction::Deny,
            priority: 0,
            enabled: true,
            scope: None,
        }];
        let bytes = policy_bytes(rules);
        let evaluator = Evaluator::load(&bytes).unwrap();
        let eval_ctx = EvalContext {
            tool_name: "read_file".to_string(),
            parameters: serde_json::json!({"path": "/etc/shadow"}),
            agent_id: "agent-001".to_string(),
            trust_level: "established".to_string(),
            data_classification: "confidential".to_string(),
        };

        group.bench_function("parameter_check", |b| {
            b.iter(|| evaluator.evaluate(black_box(&eval_ctx)))
        });
    }

    group.finish();
}

// ---------------------------------------------------------------------------
// load benchmarks
// ---------------------------------------------------------------------------

fn load_benchmarks(c: &mut Criterion) {
    let mut group = c.benchmark_group("load");

    // -- load_10_rules
    {
        let bytes = policy_bytes(gen_tool_filter_rules(10));
        group.bench_function("load_10_rules", |b| {
            b.iter(|| Evaluator::load(black_box(&bytes)).unwrap())
        });
    }

    // -- load_100_rules
    {
        let bytes = policy_bytes(gen_tool_filter_rules(100));
        group.bench_function("load_100_rules", |b| {
            b.iter(|| Evaluator::load(black_box(&bytes)).unwrap())
        });
    }

    // -- load_500_rules
    {
        let bytes = policy_bytes(gen_tool_filter_rules(500));
        group.bench_function("load_500_rules", |b| {
            b.iter(|| Evaluator::load(black_box(&bytes)).unwrap())
        });
    }

    group.finish();
}

// ---------------------------------------------------------------------------
// Criterion harness
// ---------------------------------------------------------------------------

criterion_group!(evaluate_group, evaluate_benchmarks);
criterion_group!(load_group, load_benchmarks);
criterion_main!(evaluate_group, load_group);

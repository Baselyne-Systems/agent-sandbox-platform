package e2e

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

func TestGuardrailPolicyPipeline(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	denyRule, err := guardrailsSvc.CreateRule(ctx, tenant, "deny-shell", "block shell access",
		models.RuleTypeToolFilter, "shell", models.RuleActionDeny, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create deny rule: %v", err)
	}

	allowRule, err := guardrailsSvc.CreateRule(ctx, tenant, "allow-web", "allow web search",
		models.RuleTypeToolFilter, "web_search", models.RuleActionAllow, 10, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create allow rule: %v", err)
	}

	gs, err := guardrailsSvc.CreateSet(ctx, tenant, "test-policy", "test policy set",
		[]string{denyRule.ID, allowRule.ID}, nil)
	if err != nil {
		t.Fatalf("create set: %v", err)
	}
	if len(gs.RuleIDs) != 2 {
		t.Fatalf("expected 2 rule IDs in set, got %d", len(gs.RuleIDs))
	}

	compiled, count, err := guardrailsSvc.CompilePolicy(ctx, tenant, []string{denyRule.ID, allowRule.ID})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 rules compiled, got %d", count)
	}

	var policy map[string]json.RawMessage
	if err := json.Unmarshal(compiled, &policy); err != nil {
		t.Fatalf("unmarshal policy: %v", err)
	}
	if _, ok := policy["rules"]; !ok {
		t.Fatal("compiled policy missing 'rules' key")
	}
}

func TestPolicySimulation(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	denyRule, err := guardrailsSvc.CreateRule(ctx, tenant, "deny-shell", "block shell",
		models.RuleTypeToolFilter, "shell", models.RuleActionDeny, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create deny rule: %v", err)
	}

	ruleIDs := []string{denyRule.ID}

	result, err := guardrailsSvc.SimulatePolicy(ctx, tenant, ruleIDs, "shell", nil, "")
	if err != nil {
		t.Fatalf("simulate shell: %v", err)
	}
	if result.Verdict != "deny" {
		t.Fatalf("expected deny for shell, got %s", result.Verdict)
	}

	result, err = guardrailsSvc.SimulatePolicy(ctx, tenant, ruleIDs, "web_search", nil, "")
	if err != nil {
		t.Fatalf("simulate web_search: %v", err)
	}
	if result.Verdict != "allow" {
		t.Fatalf("expected allow for web_search, got %s", result.Verdict)
	}
}

func TestRulePriorityOrdering(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	ruleA, err := guardrailsSvc.CreateRule(ctx, tenant, "allow-shell", "allow shell",
		models.RuleTypeToolFilter, "shell", models.RuleActionAllow, 10, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create allow rule: %v", err)
	}

	ruleB, err := guardrailsSvc.CreateRule(ctx, tenant, "deny-shell", "deny shell",
		models.RuleTypeToolFilter, "shell", models.RuleActionDeny, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create deny rule: %v", err)
	}

	ruleIDs := []string{ruleA.ID, ruleB.ID}

	result, err := guardrailsSvc.SimulatePolicy(ctx, tenant, ruleIDs, "shell", nil, "")
	if err != nil {
		t.Fatalf("simulate: %v", err)
	}
	if result.Verdict != "deny" {
		t.Fatalf("expected deny (lower priority wins), got %s", result.Verdict)
	}
	if result.MatchedRuleID != ruleB.ID {
		t.Fatalf("expected matched rule B, got %s", result.MatchedRuleID)
	}
}

func TestParameterCheckRule(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	rule, err := guardrailsSvc.CreateRule(ctx, tenant, "deny-prod", "deny production use",
		models.RuleTypeParameterCheck, "env=production", models.RuleActionDeny, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create param rule: %v", err)
	}

	ruleIDs := []string{rule.ID}

	result, err := guardrailsSvc.SimulatePolicy(ctx, tenant, ruleIDs, "deploy", map[string]string{"env": "production"}, "")
	if err != nil {
		t.Fatalf("simulate production: %v", err)
	}
	if result.Verdict != "deny" {
		t.Fatalf("expected deny for production, got %s", result.Verdict)
	}

	result, err = guardrailsSvc.SimulatePolicy(ctx, tenant, ruleIDs, "deploy", map[string]string{"env": "staging"}, "")
	if err != nil {
		t.Fatalf("simulate staging: %v", err)
	}
	if result.Verdict != "allow" {
		t.Fatalf("expected allow for staging, got %s", result.Verdict)
	}
}

func TestGuardrailSetResolution(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	rule1, err := guardrailsSvc.CreateRule(ctx, tenant, "rule-1", "d", models.RuleTypeToolFilter, "a", models.RuleActionDeny, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create rule 1: %v", err)
	}
	rule2, err := guardrailsSvc.CreateRule(ctx, tenant, "rule-2", "d", models.RuleTypeToolFilter, "b", models.RuleActionAllow, 2, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create rule 2: %v", err)
	}

	_, err = guardrailsSvc.CreateSet(ctx, tenant, "my-policy", "test", []string{rule1.ID, rule2.ID}, nil)
	if err != nil {
		t.Fatalf("create set: %v", err)
	}

	ids, err := guardrailsSvc.ResolveRuleIDs(ctx, tenant, "set:my-policy")
	if err != nil {
		t.Fatalf("resolve set: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 rule IDs from set resolution, got %d", len(ids))
	}

	directRef := rule1.ID + "," + rule2.ID
	ids, err = guardrailsSvc.ResolveRuleIDs(ctx, tenant, directRef)
	if err != nil {
		t.Fatalf("resolve direct: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 rule IDs from direct resolution, got %d", len(ids))
	}
	joined := strings.Join(ids, ",")
	if !strings.Contains(joined, rule1.ID) || !strings.Contains(joined, rule2.ID) {
		t.Fatalf("resolved IDs don't match: %v", ids)
	}
}

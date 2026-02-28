package guardrails

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// mockRepo is a hand-written in-memory Repository for testing.
type mockRepo struct {
	rules  map[string]*models.GuardrailRule
	nextID int
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		rules: make(map[string]*models.GuardrailRule),
	}
}

func (m *mockRepo) nextUUID() string {
	m.nextID++
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", m.nextID)
}

func (m *mockRepo) CreateRule(_ context.Context, rule *models.GuardrailRule) error {
	rule.ID = m.nextUUID()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = rule.CreatedAt
	cp := *rule
	cp.Labels = copyLabels(rule.Labels)
	m.rules[rule.ID] = &cp
	return nil
}

func (m *mockRepo) GetRule(_ context.Context, id string) (*models.GuardrailRule, error) {
	r, ok := m.rules[id]
	if !ok {
		return nil, nil
	}
	cp := *r
	cp.Labels = copyLabels(r.Labels)
	return &cp, nil
}

func (m *mockRepo) UpdateRule(_ context.Context, rule *models.GuardrailRule) error {
	_, ok := m.rules[rule.ID]
	if !ok {
		return ErrRuleNotFound
	}
	rule.UpdatedAt = time.Now()
	cp := *rule
	cp.Labels = copyLabels(rule.Labels)
	m.rules[rule.ID] = &cp
	return nil
}

func (m *mockRepo) DeleteRule(_ context.Context, id string) error {
	if _, ok := m.rules[id]; !ok {
		return ErrRuleNotFound
	}
	delete(m.rules, id)
	return nil
}

func (m *mockRepo) ListRules(_ context.Context, ruleType models.RuleType, enabledOnly bool, afterID string, limit int) ([]models.GuardrailRule, error) {
	var result []models.GuardrailRule
	for _, r := range m.rules {
		if ruleType != "" && r.Type != ruleType {
			continue
		}
		if enabledOnly && !r.Enabled {
			continue
		}
		if afterID != "" && r.ID <= afterID {
			continue
		}
		cp := *r
		cp.Labels = copyLabels(r.Labels)
		result = append(result, cp)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func copyLabels(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func TestCreateRule_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	rule, err := svc.CreateRule(context.Background(), "deny-exec", "Block exec calls",
		models.RuleTypeToolFilter, "tool == 'exec'", models.RuleActionDeny, 10, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.ID == "" {
		t.Error("expected rule ID to be set")
	}
	if rule.Name != "deny-exec" {
		t.Errorf("expected name 'deny-exec', got %q", rule.Name)
	}
	if !rule.Enabled {
		t.Error("expected rule to be enabled by default")
	}
	if rule.Labels == nil {
		t.Error("expected labels to be initialized")
	}
}

func TestCreateRule_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.CreateRule(ctx, "", "desc", models.RuleTypeToolFilter, "cond", models.RuleActionDeny, 0, nil, models.RuleScope{}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty name, got: %v", err)
	}
	if _, err := svc.CreateRule(ctx, "name", "desc", models.RuleTypeToolFilter, "", models.RuleActionDeny, 0, nil, models.RuleScope{}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty condition, got: %v", err)
	}
	if _, err := svc.CreateRule(ctx, "name", "desc", "bad_type", "cond", models.RuleActionDeny, 0, nil, models.RuleScope{}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for invalid type, got: %v", err)
	}
	if _, err := svc.CreateRule(ctx, "name", "desc", models.RuleTypeToolFilter, "cond", "bad_action", 0, nil, models.RuleScope{}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for invalid action, got: %v", err)
	}
}

func TestGetRule_Found(t *testing.T) {
	svc := NewService(newMockRepo())
	created, _ := svc.CreateRule(context.Background(), "r", "", models.RuleTypeRateLimit, "c", models.RuleActionLog, 0, nil, models.RuleScope{})
	got, err := svc.GetRule(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}
}

func TestGetRule_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetRule(context.Background(), "nonexistent-id")
	if !errors.Is(err, ErrRuleNotFound) {
		t.Errorf("expected ErrRuleNotFound, got: %v", err)
	}
}

func TestUpdateRule_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	created, _ := svc.CreateRule(ctx, "r", "", models.RuleTypeToolFilter, "c", models.RuleActionDeny, 5, nil, models.RuleScope{})

	created.Name = "updated"
	created.Priority = 20
	updated, err := svc.UpdateRule(ctx, created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "updated" {
		t.Errorf("expected name 'updated', got %q", updated.Name)
	}
	if updated.Priority != 20 {
		t.Errorf("expected priority 20, got %d", updated.Priority)
	}
}

func TestUpdateRule_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	_, err := svc.UpdateRule(ctx, &models.GuardrailRule{ID: "", Name: "n", Condition: "c", Type: models.RuleTypeToolFilter, Action: models.RuleActionDeny})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty ID, got: %v", err)
	}

	_, err = svc.UpdateRule(ctx, &models.GuardrailRule{ID: "x", Name: "", Condition: "c", Type: models.RuleTypeToolFilter, Action: models.RuleActionDeny})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty name, got: %v", err)
	}
}

func TestDeleteRule_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	rule, _ := svc.CreateRule(ctx, "r", "", models.RuleTypeToolFilter, "c", models.RuleActionDeny, 0, nil, models.RuleScope{})

	if err := svc.DeleteRule(ctx, rule.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.GetRule(ctx, rule.ID)
	if !errors.Is(err, ErrRuleNotFound) {
		t.Errorf("expected ErrRuleNotFound after delete, got: %v", err)
	}
}

func TestDeleteRule_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	err := svc.DeleteRule(context.Background(), "no-such-id")
	if !errors.Is(err, ErrRuleNotFound) {
		t.Errorf("expected ErrRuleNotFound, got: %v", err)
	}
}

func TestListRules_WithFilters(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	svc.CreateRule(ctx, "r1", "", models.RuleTypeToolFilter, "c", models.RuleActionDeny, 0, nil, models.RuleScope{})
	svc.CreateRule(ctx, "r2", "", models.RuleTypeRateLimit, "c", models.RuleActionLog, 0, nil, models.RuleScope{})
	svc.CreateRule(ctx, "r3", "", models.RuleTypeToolFilter, "c", models.RuleActionAllow, 0, nil, models.RuleScope{})

	// Filter by type
	rules, _, err := svc.ListRules(ctx, models.RuleTypeToolFilter, false, 50, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("expected 2 rules for tool_filter, got %d", len(rules))
	}

	// All rules
	rules, _, err = svc.ListRules(ctx, "", false, 50, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 3 {
		t.Errorf("expected 3 rules total, got %d", len(rules))
	}
}

func TestListRules_Pagination(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		svc.CreateRule(ctx, fmt.Sprintf("rule-%d", i), "", models.RuleTypeToolFilter, "c", models.RuleActionDeny, 0, nil, models.RuleScope{})
	}

	rules, nextToken, err := svc.ListRules(ctx, "", false, 3, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}
	if nextToken == "" {
		t.Error("expected a next page token")
	}

	rules2, nextToken2, err := svc.ListRules(ctx, "", false, 3, nextToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules2) != 2 {
		t.Fatalf("expected 2 rules on second page, got %d", len(rules2))
	}
	if nextToken2 != "" {
		t.Error("expected no next page token on last page")
	}
}

func TestCompilePolicy(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Create rules first so CompilePolicy can fetch them
	r1, _ := svc.CreateRule(ctx, "deny-exec", "Block exec", models.RuleTypeToolFilter, "exec,shell", models.RuleActionDeny, 10, nil, models.RuleScope{})
	r2, _ := svc.CreateRule(ctx, "log-read", "Log reads", models.RuleTypeToolFilter, "read_file", models.RuleActionLog, 20, nil, models.RuleScope{})
	r3, _ := svc.CreateRule(ctx, "check-path", "Check path", models.RuleTypeParameterCheck, "path=/etc/shadow", models.RuleActionDeny, 5, nil, models.RuleScope{})

	compiled, count, err := svc.CompilePolicy(ctx, []string{r1.ID, r2.ID, r3.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
	if len(compiled) == 0 {
		t.Error("expected non-empty compiled policy")
	}

	// Verify the output is a valid compiledPolicy JSON
	var policy compiledPolicy
	if err := json.Unmarshal(compiled, &policy); err != nil {
		t.Fatalf("failed to unmarshal compiled policy: %v", err)
	}
	if len(policy.Rules) != 3 {
		t.Errorf("expected 3 rules in policy, got %d", len(policy.Rules))
	}
	if policy.Rules[0].Name != "deny-exec" {
		t.Errorf("expected first rule name 'deny-exec', got %q", policy.Rules[0].Name)
	}
	if policy.Rules[0].RuleType != "tool_filter" {
		t.Errorf("expected rule_type 'tool_filter', got %q", policy.Rules[0].RuleType)
	}
	if policy.Rules[0].Action != "deny" {
		t.Errorf("expected action 'deny', got %q", policy.Rules[0].Action)
	}
	if !policy.Rules[0].Enabled {
		t.Error("expected rule to be enabled")
	}
}

func TestCompilePolicy_Empty(t *testing.T) {
	svc := NewService(newMockRepo())
	_, _, err := svc.CompilePolicy(context.Background(), nil)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty rule IDs, got: %v", err)
	}
}

func TestCompilePolicy_NotFoundRule(t *testing.T) {
	svc := NewService(newMockRepo())
	_, _, err := svc.CompilePolicy(context.Background(), []string{"nonexistent-id"})
	if !errors.Is(err, ErrRuleNotFound) {
		t.Errorf("expected ErrRuleNotFound for nonexistent rule, got: %v", err)
	}
}

func TestSimulatePolicy_ToolFilterMatch(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	rule, _ := svc.CreateRule(ctx, "deny-exec", "Block exec", models.RuleTypeToolFilter, "exec,shell", models.RuleActionDeny, 10, nil, models.RuleScope{})

	result, err := svc.SimulatePolicy(ctx, []string{rule.ID}, "exec", nil, "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != "deny" {
		t.Errorf("expected verdict 'deny', got %q", result.Verdict)
	}
	if result.MatchedRuleID != rule.ID {
		t.Errorf("expected matched rule ID %q, got %q", rule.ID, result.MatchedRuleID)
	}
	if result.MatchedRuleName != "deny-exec" {
		t.Errorf("expected matched rule name 'deny-exec', got %q", result.MatchedRuleName)
	}
}

func TestSimulatePolicy_ToolFilterNoMatch(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	rule, _ := svc.CreateRule(ctx, "deny-exec", "Block exec", models.RuleTypeToolFilter, "exec,shell", models.RuleActionDeny, 10, nil, models.RuleScope{})

	result, err := svc.SimulatePolicy(ctx, []string{rule.ID}, "read_file", nil, "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != "allow" {
		t.Errorf("expected verdict 'allow', got %q", result.Verdict)
	}
	if result.MatchedRuleID != "" {
		t.Errorf("expected no matched rule ID, got %q", result.MatchedRuleID)
	}
}

func TestSimulatePolicy_ParameterCheckMatch(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	rule, _ := svc.CreateRule(ctx, "block-shadow", "Block /etc/shadow", models.RuleTypeParameterCheck, "path=/etc/shadow", models.RuleActionDeny, 5, nil, models.RuleScope{})

	result, err := svc.SimulatePolicy(ctx, []string{rule.ID}, "read_file", map[string]string{"path": "/etc/shadow"}, "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != "deny" {
		t.Errorf("expected verdict 'deny', got %q", result.Verdict)
	}
	if result.MatchedRuleID != rule.ID {
		t.Errorf("expected matched rule ID %q, got %q", rule.ID, result.MatchedRuleID)
	}
}

func TestSimulatePolicy_PriorityOrdering(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// Lower priority number = higher precedence
	allowRule, _ := svc.CreateRule(ctx, "allow-read", "Allow reads", models.RuleTypeToolFilter, "read_file", models.RuleActionAllow, 1, nil, models.RuleScope{})
	svc.CreateRule(ctx, "deny-read", "Deny reads", models.RuleTypeToolFilter, "read_file", models.RuleActionDeny, 10, nil, models.RuleScope{})

	result, err := svc.SimulatePolicy(ctx, []string{allowRule.ID}, "read_file", nil, "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != "allow" {
		t.Errorf("expected verdict 'allow' (higher priority), got %q", result.Verdict)
	}
}

func TestSimulatePolicy_DisabledRulesSkipped(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	rule, _ := svc.CreateRule(ctx, "deny-exec", "Block exec", models.RuleTypeToolFilter, "exec", models.RuleActionDeny, 10, nil, models.RuleScope{})

	// Disable the rule
	rule.Enabled = false
	svc.UpdateRule(ctx, rule)

	result, err := svc.SimulatePolicy(ctx, []string{rule.ID}, "exec", nil, "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != "allow" {
		t.Errorf("expected verdict 'allow' (disabled rule), got %q", result.Verdict)
	}
}

func TestSimulatePolicy_EscalateVerdict(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	rule, _ := svc.CreateRule(ctx, "escalate-deploy", "Escalate deployments", models.RuleTypeToolFilter, "deploy", models.RuleActionEscalate, 5, nil, models.RuleScope{})

	result, err := svc.SimulatePolicy(ctx, []string{rule.ID}, "deploy", nil, "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != "escalate" {
		t.Errorf("expected verdict 'escalate', got %q", result.Verdict)
	}
}

func TestSimulatePolicy_EmptyRules(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.SimulatePolicy(context.Background(), nil, "exec", nil, "agent-1")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSimulatePolicy_EmptyToolName(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	rule, _ := svc.CreateRule(ctx, "r", "", models.RuleTypeToolFilter, "exec", models.RuleActionDeny, 0, nil, models.RuleScope{})
	_, err := svc.SimulatePolicy(ctx, []string{rule.ID}, "", nil, "agent-1")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

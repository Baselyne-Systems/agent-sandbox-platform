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
	sets   map[string]*models.GuardrailSet
	setsByName map[string]string // name -> ID
	nextID int
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		rules:      make(map[string]*models.GuardrailRule),
		sets:       make(map[string]*models.GuardrailSet),
		setsByName: make(map[string]string),
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

func (m *mockRepo) GetRule(_ context.Context, _ string, id string) (*models.GuardrailRule, error) {
	r, ok := m.rules[id]
	if !ok {
		return nil, nil
	}
	cp := *r
	cp.Labels = copyLabels(r.Labels)
	return &cp, nil
}

func (m *mockRepo) UpdateRule(_ context.Context, _ string, rule *models.GuardrailRule) error {
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

func (m *mockRepo) DeleteRule(_ context.Context, _ string, id string) error {
	if _, ok := m.rules[id]; !ok {
		return ErrRuleNotFound
	}
	delete(m.rules, id)
	return nil
}

func (m *mockRepo) ListRules(_ context.Context, _ string, ruleType models.RuleType, enabledOnly bool, afterID string, limit int) ([]models.GuardrailRule, error) {
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

func (m *mockRepo) CreateSet(_ context.Context, set *models.GuardrailSet) error {
	set.ID = m.nextUUID()
	set.CreatedAt = time.Now()
	set.UpdatedAt = set.CreatedAt
	cp := *set
	cp.RuleIDs = copyStringSlice(set.RuleIDs)
	cp.Labels = copyLabels(set.Labels)
	m.sets[set.ID] = &cp
	m.setsByName[set.Name] = set.ID
	return nil
}

func (m *mockRepo) GetSet(_ context.Context, _ string, id string) (*models.GuardrailSet, error) {
	s, ok := m.sets[id]
	if !ok {
		return nil, nil
	}
	cp := *s
	cp.RuleIDs = copyStringSlice(s.RuleIDs)
	cp.Labels = copyLabels(s.Labels)
	return &cp, nil
}

func (m *mockRepo) GetSetByName(_ context.Context, _ string, name string) (*models.GuardrailSet, error) {
	id, ok := m.setsByName[name]
	if !ok {
		return nil, nil
	}
	return m.GetSet(context.Background(), "", id)
}

func (m *mockRepo) UpdateSet(_ context.Context, _ string, set *models.GuardrailSet) error {
	existing, ok := m.sets[set.ID]
	if !ok {
		return ErrSetNotFound
	}
	// Remove old name mapping if name changed.
	delete(m.setsByName, existing.Name)
	set.UpdatedAt = time.Now()
	cp := *set
	cp.RuleIDs = copyStringSlice(set.RuleIDs)
	cp.Labels = copyLabels(set.Labels)
	m.sets[set.ID] = &cp
	m.setsByName[set.Name] = set.ID
	return nil
}

func (m *mockRepo) DeleteSet(_ context.Context, _ string, id string) error {
	s, ok := m.sets[id]
	if !ok {
		return ErrSetNotFound
	}
	delete(m.setsByName, s.Name)
	delete(m.sets, id)
	return nil
}

func (m *mockRepo) ListSets(_ context.Context, _ string, afterID string, limit int) ([]models.GuardrailSet, error) {
	var result []models.GuardrailSet
	for _, s := range m.sets {
		if afterID != "" && s.ID <= afterID {
			continue
		}
		cp := *s
		cp.RuleIDs = copyStringSlice(s.RuleIDs)
		cp.Labels = copyLabels(s.Labels)
		result = append(result, cp)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func copyStringSlice(s []string) []string {
	if s == nil {
		return nil
	}
	cp := make([]string, len(s))
	copy(cp, s)
	return cp
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
	rule, err := svc.CreateRule(context.Background(), "test-tenant", "deny-exec", "Block exec calls",
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

	if _, err := svc.CreateRule(ctx, "test-tenant", "", "desc", models.RuleTypeToolFilter, "cond", models.RuleActionDeny, 0, nil, models.RuleScope{}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty name, got: %v", err)
	}
	if _, err := svc.CreateRule(ctx, "test-tenant", "name", "desc", models.RuleTypeToolFilter, "", models.RuleActionDeny, 0, nil, models.RuleScope{}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty condition, got: %v", err)
	}
	if _, err := svc.CreateRule(ctx, "test-tenant", "name", "desc", "bad_type", "cond", models.RuleActionDeny, 0, nil, models.RuleScope{}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for invalid type, got: %v", err)
	}
	if _, err := svc.CreateRule(ctx, "test-tenant", "name", "desc", models.RuleTypeToolFilter, "cond", "bad_action", 0, nil, models.RuleScope{}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for invalid action, got: %v", err)
	}
}

func TestGetRule_Found(t *testing.T) {
	svc := NewService(newMockRepo())
	created, _ := svc.CreateRule(context.Background(), "test-tenant", "r", "", models.RuleTypeRateLimit, "c", models.RuleActionLog, 0, nil, models.RuleScope{})
	got, err := svc.GetRule(context.Background(), "test-tenant", created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}
}

func TestGetRule_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetRule(context.Background(), "test-tenant", "nonexistent-id")
	if !errors.Is(err, ErrRuleNotFound) {
		t.Errorf("expected ErrRuleNotFound, got: %v", err)
	}
}

func TestUpdateRule_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	created, _ := svc.CreateRule(ctx, "test-tenant", "r", "", models.RuleTypeToolFilter, "c", models.RuleActionDeny, 5, nil, models.RuleScope{})

	created.Name = "updated"
	created.Priority = 20
	updated, err := svc.UpdateRule(ctx, "test-tenant", created)
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

	_, err := svc.UpdateRule(ctx, "test-tenant", &models.GuardrailRule{ID: "", Name: "n", Condition: "c", Type: models.RuleTypeToolFilter, Action: models.RuleActionDeny})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty ID, got: %v", err)
	}

	_, err = svc.UpdateRule(ctx, "test-tenant", &models.GuardrailRule{ID: "x", Name: "", Condition: "c", Type: models.RuleTypeToolFilter, Action: models.RuleActionDeny})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty name, got: %v", err)
	}
}

func TestDeleteRule_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	rule, _ := svc.CreateRule(ctx, "test-tenant", "r", "", models.RuleTypeToolFilter, "c", models.RuleActionDeny, 0, nil, models.RuleScope{})

	if err := svc.DeleteRule(ctx, "test-tenant", rule.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.GetRule(ctx, "test-tenant", rule.ID)
	if !errors.Is(err, ErrRuleNotFound) {
		t.Errorf("expected ErrRuleNotFound after delete, got: %v", err)
	}
}

func TestDeleteRule_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	err := svc.DeleteRule(context.Background(), "test-tenant", "no-such-id")
	if !errors.Is(err, ErrRuleNotFound) {
		t.Errorf("expected ErrRuleNotFound, got: %v", err)
	}
}

func TestListRules_WithFilters(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	svc.CreateRule(ctx, "test-tenant", "r1", "", models.RuleTypeToolFilter, "c", models.RuleActionDeny, 0, nil, models.RuleScope{})
	svc.CreateRule(ctx, "test-tenant", "r2", "", models.RuleTypeRateLimit, "c", models.RuleActionLog, 0, nil, models.RuleScope{})
	svc.CreateRule(ctx, "test-tenant", "r3", "", models.RuleTypeToolFilter, "c", models.RuleActionAllow, 0, nil, models.RuleScope{})

	// Filter by type
	rules, _, err := svc.ListRules(ctx, "test-tenant", models.RuleTypeToolFilter, false, 50, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("expected 2 rules for tool_filter, got %d", len(rules))
	}

	// All rules
	rules, _, err = svc.ListRules(ctx, "test-tenant", "", false, 50, "")
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
		svc.CreateRule(ctx, "test-tenant", fmt.Sprintf("rule-%d", i), "", models.RuleTypeToolFilter, "c", models.RuleActionDeny, 0, nil, models.RuleScope{})
	}

	rules, nextToken, err := svc.ListRules(ctx, "test-tenant", "", false, 3, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}
	if nextToken == "" {
		t.Error("expected a next page token")
	}

	rules2, nextToken2, err := svc.ListRules(ctx, "test-tenant", "", false, 3, nextToken)
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
	r1, _ := svc.CreateRule(ctx, "test-tenant", "deny-exec", "Block exec", models.RuleTypeToolFilter, "exec,shell", models.RuleActionDeny, 10, nil, models.RuleScope{})
	r2, _ := svc.CreateRule(ctx, "test-tenant", "log-read", "Log reads", models.RuleTypeToolFilter, "read_file", models.RuleActionLog, 20, nil, models.RuleScope{})
	r3, _ := svc.CreateRule(ctx, "test-tenant", "check-path", "Check path", models.RuleTypeParameterCheck, "path=/etc/shadow", models.RuleActionDeny, 5, nil, models.RuleScope{})

	compiled, count, err := svc.CompilePolicy(ctx, "test-tenant", []string{r1.ID, r2.ID, r3.ID})
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
	_, _, err := svc.CompilePolicy(context.Background(), "test-tenant", nil)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty rule IDs, got: %v", err)
	}
}

func TestCompilePolicy_NotFoundRule(t *testing.T) {
	svc := NewService(newMockRepo())
	_, _, err := svc.CompilePolicy(context.Background(), "test-tenant", []string{"nonexistent-id"})
	if !errors.Is(err, ErrRuleNotFound) {
		t.Errorf("expected ErrRuleNotFound for nonexistent rule, got: %v", err)
	}
}

func TestSimulatePolicy_ToolFilterMatch(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	rule, _ := svc.CreateRule(ctx, "test-tenant", "deny-exec", "Block exec", models.RuleTypeToolFilter, "exec,shell", models.RuleActionDeny, 10, nil, models.RuleScope{})

	result, err := svc.SimulatePolicy(ctx, "test-tenant", []string{rule.ID}, "exec", nil, "agent-1")
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

	rule, _ := svc.CreateRule(ctx, "test-tenant", "deny-exec", "Block exec", models.RuleTypeToolFilter, "exec,shell", models.RuleActionDeny, 10, nil, models.RuleScope{})

	result, err := svc.SimulatePolicy(ctx, "test-tenant", []string{rule.ID}, "read_file", nil, "agent-1")
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

	rule, _ := svc.CreateRule(ctx, "test-tenant", "block-shadow", "Block /etc/shadow", models.RuleTypeParameterCheck, "path=/etc/shadow", models.RuleActionDeny, 5, nil, models.RuleScope{})

	result, err := svc.SimulatePolicy(ctx, "test-tenant", []string{rule.ID}, "read_file", map[string]string{"path": "/etc/shadow"}, "agent-1")
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
	allowRule, _ := svc.CreateRule(ctx, "test-tenant", "allow-read", "Allow reads", models.RuleTypeToolFilter, "read_file", models.RuleActionAllow, 1, nil, models.RuleScope{})
	svc.CreateRule(ctx, "test-tenant", "deny-read", "Deny reads", models.RuleTypeToolFilter, "read_file", models.RuleActionDeny, 10, nil, models.RuleScope{})

	result, err := svc.SimulatePolicy(ctx, "test-tenant", []string{allowRule.ID}, "read_file", nil, "agent-1")
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

	rule, _ := svc.CreateRule(ctx, "test-tenant", "deny-exec", "Block exec", models.RuleTypeToolFilter, "exec", models.RuleActionDeny, 10, nil, models.RuleScope{})

	// Disable the rule
	rule.Enabled = false
	svc.UpdateRule(ctx, "test-tenant", rule)

	result, err := svc.SimulatePolicy(ctx, "test-tenant", []string{rule.ID}, "exec", nil, "agent-1")
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

	rule, _ := svc.CreateRule(ctx, "test-tenant", "escalate-deploy", "Escalate deployments", models.RuleTypeToolFilter, "deploy", models.RuleActionEscalate, 5, nil, models.RuleScope{})

	result, err := svc.SimulatePolicy(ctx, "test-tenant", []string{rule.ID}, "deploy", nil, "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Verdict != "escalate" {
		t.Errorf("expected verdict 'escalate', got %q", result.Verdict)
	}
}

func TestSimulatePolicy_EmptyRules(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.SimulatePolicy(context.Background(), "test-tenant", nil, "exec", nil, "agent-1")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSimulatePolicy_EmptyToolName(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	rule, _ := svc.CreateRule(ctx, "test-tenant", "r", "", models.RuleTypeToolFilter, "exec", models.RuleActionDeny, 0, nil, models.RuleScope{})
	_, err := svc.SimulatePolicy(ctx, "test-tenant", []string{rule.ID}, "", nil, "agent-1")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- GuardrailSet tests ---

func TestCreateSet_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	set, err := svc.CreateSet(context.Background(), "test-tenant", "invoice-rules", "Rules for invoice processing", []string{"rule-1", "rule-2"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if set.ID == "" {
		t.Error("expected set ID to be set")
	}
	if set.Name != "invoice-rules" {
		t.Errorf("expected name 'invoice-rules', got %q", set.Name)
	}
	if len(set.RuleIDs) != 2 {
		t.Errorf("expected 2 rule IDs, got %d", len(set.RuleIDs))
	}
}

func TestCreateSet_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.CreateSet(ctx, "test-tenant", "", "desc", []string{"r1"}, nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty name, got: %v", err)
	}
	if _, err := svc.CreateSet(ctx, "test-tenant", "name", "desc", nil, nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for nil rule_ids, got: %v", err)
	}
	if _, err := svc.CreateSet(ctx, "test-tenant", "name", "desc", []string{}, nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty rule_ids, got: %v", err)
	}
}

func TestGetSet_Found(t *testing.T) {
	svc := NewService(newMockRepo())
	created, _ := svc.CreateSet(context.Background(), "test-tenant", "test-set", "", []string{"r1"}, nil)
	got, err := svc.GetSet(context.Background(), "test-tenant", created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}
}

func TestGetSet_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetSet(context.Background(), "test-tenant", "nonexistent-id")
	if !errors.Is(err, ErrSetNotFound) {
		t.Errorf("expected ErrSetNotFound, got: %v", err)
	}
}

func TestGetSetByName_Found(t *testing.T) {
	svc := NewService(newMockRepo())
	created, _ := svc.CreateSet(context.Background(), "test-tenant", "my-set", "", []string{"r1"}, nil)
	got, err := svc.GetSetByName(context.Background(), "test-tenant", "my-set")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}
}

func TestGetSetByName_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetSetByName(context.Background(), "test-tenant", "no-such-set")
	if !errors.Is(err, ErrSetNotFound) {
		t.Errorf("expected ErrSetNotFound, got: %v", err)
	}
}

func TestUpdateSet_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	created, _ := svc.CreateSet(ctx, "test-tenant", "s", "", []string{"r1"}, nil)

	created.Name = "updated"
	created.RuleIDs = []string{"r1", "r2", "r3"}
	updated, err := svc.UpdateSet(ctx, "test-tenant", created)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "updated" {
		t.Errorf("expected name 'updated', got %q", updated.Name)
	}
	if len(updated.RuleIDs) != 3 {
		t.Errorf("expected 3 rule IDs, got %d", len(updated.RuleIDs))
	}
}

func TestDeleteSet_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	set, _ := svc.CreateSet(ctx, "test-tenant", "s", "", []string{"r1"}, nil)
	if err := svc.DeleteSet(ctx, "test-tenant", set.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.GetSet(ctx, "test-tenant", set.ID)
	if !errors.Is(err, ErrSetNotFound) {
		t.Errorf("expected ErrSetNotFound after delete, got: %v", err)
	}
}

func TestListSets_Pagination(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		svc.CreateSet(ctx, "test-tenant", fmt.Sprintf("set-%d", i), "", []string{"r1"}, nil)
	}

	sets, nextToken, err := svc.ListSets(ctx, "test-tenant", 3, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sets) != 3 {
		t.Fatalf("expected 3 sets, got %d", len(sets))
	}
	if nextToken == "" {
		t.Error("expected a next page token")
	}

	sets2, nextToken2, err := svc.ListSets(ctx, "test-tenant", 3, nextToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sets2) != 2 {
		t.Fatalf("expected 2 sets on second page, got %d", len(sets2))
	}
	if nextToken2 != "" {
		t.Error("expected no next page token on last page")
	}
}

func TestResolveRuleIDs_SetPrefix(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	svc.CreateSet(ctx, "test-tenant", "invoice-rules", "", []string{"rule-a", "rule-b", "rule-c"}, nil)

	ids, err := svc.ResolveRuleIDs(ctx, "test-tenant", "set:invoice-rules")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 3 {
		t.Errorf("expected 3 rule IDs, got %d", len(ids))
	}
	if ids[0] != "rule-a" {
		t.Errorf("expected first ID 'rule-a', got %q", ids[0])
	}
}

func TestResolveRuleIDs_CommaSeparated(t *testing.T) {
	svc := NewService(newMockRepo())
	ids, err := svc.ResolveRuleIDs(context.Background(), "test-tenant", "r1, r2, r3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 3 {
		t.Errorf("expected 3 rule IDs, got %d", len(ids))
	}
	if ids[1] != "r2" {
		t.Errorf("expected second ID 'r2', got %q", ids[1])
	}
}

func TestResolveRuleIDs_SetNotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.ResolveRuleIDs(context.Background(), "test-tenant", "set:nonexistent")
	if !errors.Is(err, ErrSetNotFound) {
		t.Errorf("expected ErrSetNotFound, got: %v", err)
	}
}

//go:build integration

package guardrails

import (
	"context"
	"database/sql"
	"os"
	"sort"
	"testing"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/testutil"
)

var testDB *testutil.TestDB

func TestMain(m *testing.M) {
	testDB = testutil.MustSetupTestDB()
	code := m.Run()
	testDB.Cleanup()
	os.Exit(code)
}

func setup(t *testing.T) (*PostgresRepository, *sql.DB) {
	t.Helper()
	testutil.TruncateAll(t, testDB.DB)
	return NewPostgresRepository(testDB.DB), testDB.DB
}

func TestInteg_CreateAndGetRule_JSONBLabels(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	rule := &models.GuardrailRule{
		Name:        "block-shell",
		Description: "Block shell execution",
		Type:        models.RuleTypeToolFilter,
		Condition:   "tool_name == 'bash'",
		Action:      models.RuleActionDeny,
		Priority:    10,
		Enabled:     true,
		Labels:      map[string]string{"env": "prod", "severity": "high"},
	}
	if err := repo.CreateRule(ctx, rule); err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	if rule.ID == "" {
		t.Fatal("expected server-generated ID")
	}
	if rule.CreatedAt.IsZero() || rule.UpdatedAt.IsZero() {
		t.Fatal("expected server-generated timestamps")
	}

	got, err := repo.GetRule(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetRule: %v", err)
	}
	if got.Labels["env"] != "prod" || got.Labels["severity"] != "high" {
		t.Errorf("labels round-trip failed: %v", got.Labels)
	}
	if got.Type != models.RuleTypeToolFilter {
		t.Errorf("type = %q, want %q", got.Type, models.RuleTypeToolFilter)
	}
	if got.Action != models.RuleActionDeny {
		t.Errorf("action = %q, want %q", got.Action, models.RuleActionDeny)
	}
}

func TestInteg_UpdateRule_AllFields(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	rule := &models.GuardrailRule{
		Name:      "original",
		Type:      models.RuleTypeToolFilter,
		Condition: "old",
		Action:    models.RuleActionAllow,
		Priority:  1,
		Enabled:   true,
		Labels:    map[string]string{"v": "1"},
	}
	if err := repo.CreateRule(ctx, rule); err != nil {
		t.Fatalf("CreateRule: %v", err)
	}
	originalUpdatedAt := rule.UpdatedAt

	rule.Name = "updated"
	rule.Description = "now described"
	rule.Type = models.RuleTypeRateLimit
	rule.Condition = "new"
	rule.Action = models.RuleActionEscalate
	rule.Priority = 99
	rule.Enabled = false
	rule.Labels = map[string]string{"v": "2"}

	if err := repo.UpdateRule(ctx, rule); err != nil {
		t.Fatalf("UpdateRule: %v", err)
	}

	if !rule.UpdatedAt.After(originalUpdatedAt) {
		t.Error("updated_at should have changed")
	}

	got, err := repo.GetRule(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetRule: %v", err)
	}
	if got.Name != "updated" {
		t.Errorf("name = %q, want %q", got.Name, "updated")
	}
	if got.Type != models.RuleTypeRateLimit {
		t.Errorf("type = %q, want %q", got.Type, models.RuleTypeRateLimit)
	}
	if got.Enabled {
		t.Error("enabled should be false")
	}
	if got.Labels["v"] != "2" {
		t.Errorf("labels = %v, want v=2", got.Labels)
	}
}

func TestInteg_DeleteRule_Success(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	rule := &models.GuardrailRule{
		Name:      "delete-me",
		Type:      models.RuleTypeToolFilter,
		Condition: "any",
		Action:    models.RuleActionLog,
		Labels:    map[string]string{},
	}
	if err := repo.CreateRule(ctx, rule); err != nil {
		t.Fatalf("CreateRule: %v", err)
	}

	if err := repo.DeleteRule(ctx, rule.ID); err != nil {
		t.Fatalf("DeleteRule: %v", err)
	}

	got, err := repo.GetRule(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetRule: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil after delete, got %+v", got)
	}
}

func TestInteg_DeleteRule_NotFound(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	err := repo.DeleteRule(ctx, "00000000-0000-0000-0000-000000000000")
	if err != ErrRuleNotFound {
		t.Errorf("error = %v, want ErrRuleNotFound", err)
	}
}

func TestInteg_ListRules_TypeFilter(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	rules := []*models.GuardrailRule{
		{Name: "r1", Type: models.RuleTypeToolFilter, Condition: "c", Action: models.RuleActionDeny, Enabled: true, Labels: map[string]string{}},
		{Name: "r2", Type: models.RuleTypeRateLimit, Condition: "c", Action: models.RuleActionDeny, Enabled: true, Labels: map[string]string{}},
		{Name: "r3", Type: models.RuleTypeToolFilter, Condition: "c", Action: models.RuleActionDeny, Enabled: true, Labels: map[string]string{}},
	}
	for _, r := range rules {
		if err := repo.CreateRule(ctx, r); err != nil {
			t.Fatalf("CreateRule %s: %v", r.Name, err)
		}
	}

	filtered, err := repo.ListRules(ctx, models.RuleTypeToolFilter, false, "", 10)
	if err != nil {
		t.Fatalf("ListRules: %v", err)
	}
	if len(filtered) != 2 {
		t.Errorf("count = %d, want 2", len(filtered))
	}
}

func TestInteg_ListRules_EnabledOnly(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	rules := []*models.GuardrailRule{
		{Name: "enabled", Type: models.RuleTypeToolFilter, Condition: "c", Action: models.RuleActionDeny, Enabled: true, Labels: map[string]string{}},
		{Name: "disabled", Type: models.RuleTypeToolFilter, Condition: "c", Action: models.RuleActionDeny, Enabled: false, Labels: map[string]string{}},
	}
	for _, r := range rules {
		if err := repo.CreateRule(ctx, r); err != nil {
			t.Fatalf("CreateRule %s: %v", r.Name, err)
		}
	}

	enabled, err := repo.ListRules(ctx, "", true, "", 10)
	if err != nil {
		t.Fatalf("ListRules: %v", err)
	}
	if len(enabled) != 1 {
		t.Errorf("count = %d, want 1", len(enabled))
	}
	if enabled[0].Name != "enabled" {
		t.Errorf("name = %q, want %q", enabled[0].Name, "enabled")
	}
}

func TestInteg_ListRules_Pagination(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	var ids []string
	for i := 0; i < 5; i++ {
		r := &models.GuardrailRule{
			Name:      "rule-" + string(rune('A'+i)),
			Type:      models.RuleTypeToolFilter,
			Condition: "c",
			Action:    models.RuleActionDeny,
			Enabled:   true,
			Labels:    map[string]string{},
		}
		if err := repo.CreateRule(ctx, r); err != nil {
			t.Fatalf("CreateRule[%d]: %v", i, err)
		}
		ids = append(ids, r.ID)
	}
	sort.Strings(ids)

	page1, err := repo.ListRules(ctx, "", false, "", 2)
	if err != nil {
		t.Fatalf("ListRules page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 len = %d, want 2", len(page1))
	}

	page2, err := repo.ListRules(ctx, "", false, page1[1].ID, 2)
	if err != nil {
		t.Fatalf("ListRules page2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("page2 len = %d, want 2", len(page2))
	}

	page3, err := repo.ListRules(ctx, "", false, page2[1].ID, 2)
	if err != nil {
		t.Fatalf("ListRules page3: %v", err)
	}
	if len(page3) != 1 {
		t.Fatalf("page3 len = %d, want 1", len(page3))
	}
}

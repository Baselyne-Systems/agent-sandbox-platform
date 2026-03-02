//go:build integration

package guardrails

import (
	"context"
	"fmt"
	"testing"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

const benchTenant = "bench-tenant"

func benchSetup(b *testing.B) *PostgresRepository {
	b.Helper()
	_, err := testDB.DB.Exec(`TRUNCATE
		warm_pool_slots, warm_pool_configs,
		workspace_snapshots, delivery_channels, timeout_policies,
		action_records, human_requests, scoped_credentials, tasks,
		agents, guardrail_rules, guardrail_sets, usage_records, budgets,
		workspaces, hosts
		CASCADE`)
	if err != nil {
		b.Fatalf("truncate: %v", err)
	}
	return NewPostgresRepository(testDB.DB)
}

func BenchmarkDB_CreateRule(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		rule := &models.GuardrailRule{
			TenantID:    benchTenant,
			Name:        fmt.Sprintf("rule-%d", b.N),
			Description: "benchmark rule",
			Type:        models.RuleTypeToolFilter,
			Condition:   "tool_name == 'bash'",
			Action:      models.RuleActionDeny,
			Priority:    10,
			Enabled:     true,
			Labels:      map[string]string{"env": "bench", "severity": "high"},
			Scope: models.RuleScope{
				AgentIDs:  []string{"agent-1", "agent-2"},
				ToolNames: []string{"bash", "http_fetch"},
			},
		}
		if err := repo.CreateRule(ctx, rule); err != nil {
			b.Fatalf("CreateRule: %v", err)
		}
	}
}

func BenchmarkDB_GetRule(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	rule := &models.GuardrailRule{
		TenantID:  benchTenant,
		Name:      "get-rule",
		Type:      models.RuleTypeToolFilter,
		Condition: "tool_name == 'bash'",
		Action:    models.RuleActionDeny,
		Priority:  10,
		Enabled:   true,
		Labels:    map[string]string{"env": "bench"},
		Scope:     models.RuleScope{ToolNames: []string{"bash"}},
	}
	if err := repo.CreateRule(ctx, rule); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		got, err := repo.GetRule(ctx, benchTenant, rule.ID)
		if err != nil {
			b.Fatalf("GetRule: %v", err)
		}
		if got == nil {
			b.Fatal("expected rule, got nil")
		}
	}
}

func BenchmarkDB_ListRules_100(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		r := &models.GuardrailRule{
			TenantID:  benchTenant,
			Name:      fmt.Sprintf("list-rule-%d", i),
			Type:      models.RuleTypeToolFilter,
			Condition: "c",
			Action:    models.RuleActionDeny,
			Priority:  i,
			Enabled:   true,
			Labels:    map[string]string{"env": "bench"},
		}
		if err := repo.CreateRule(ctx, r); err != nil {
			b.Fatalf("seed[%d]: %v", i, err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		rules, err := repo.ListRules(ctx, benchTenant, "", false, "", 25)
		if err != nil {
			b.Fatalf("ListRules: %v", err)
		}
		if len(rules) != 25 {
			b.Fatalf("expected 25, got %d", len(rules))
		}
	}
}

func BenchmarkDB_UpdateRule(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	rule := &models.GuardrailRule{
		TenantID:  benchTenant,
		Name:      "update-rule",
		Type:      models.RuleTypeToolFilter,
		Condition: "old",
		Action:    models.RuleActionAllow,
		Priority:  1,
		Enabled:   true,
		Labels:    map[string]string{"v": "1"},
	}
	if err := repo.CreateRule(ctx, rule); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		rule.Name = fmt.Sprintf("updated-%d", b.N)
		rule.Condition = fmt.Sprintf("cond-%d", b.N)
		rule.Priority = b.N % 100
		rule.Labels = map[string]string{"v": fmt.Sprintf("%d", b.N)}
		if err := repo.UpdateRule(ctx, benchTenant, rule); err != nil {
			b.Fatalf("UpdateRule: %v", err)
		}
	}
}

func BenchmarkDB_BulkInsert_100(b *testing.B) {
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		repo := benchSetup(b)
		b.StartTimer()

		for i := 0; i < 100; i++ {
			r := &models.GuardrailRule{
				TenantID:  benchTenant,
				Name:      fmt.Sprintf("bulk-%d", i),
				Type:      models.RuleTypeToolFilter,
				Condition: "c",
				Action:    models.RuleActionDeny,
				Priority:  i,
				Enabled:   true,
				Labels:    map[string]string{"batch": "true"},
			}
			if err := repo.CreateRule(ctx, r); err != nil {
				b.Fatalf("CreateRule[%d]: %v", i, err)
			}
		}
	}
}

func BenchmarkDB_JSONBScopeRoundTrip(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		rule := &models.GuardrailRule{
			TenantID:    benchTenant,
			Name:        fmt.Sprintf("scope-rule-%d", b.N),
			Description: "full scope benchmark",
			Type:        models.RuleTypeParameterCheck,
			Condition:   "param.size > 1048576",
			Action:      models.RuleActionEscalate,
			Priority:    50,
			Enabled:     true,
			Labels: map[string]string{
				"env": "prod", "severity": "critical", "team": "security",
			},
			Scope: models.RuleScope{
				AgentIDs:            []string{"agent-a", "agent-b", "agent-c"},
				ToolNames:           []string{"http_fetch", "bash", "file_write", "db_query"},
				TrustLevels:         []string{"new", "established"},
				DataClassifications: []string{"pii", "financial", "medical"},
			},
		}
		if err := repo.CreateRule(ctx, rule); err != nil {
			b.Fatalf("CreateRule: %v", err)
		}

		got, err := repo.GetRule(ctx, benchTenant, rule.ID)
		if err != nil {
			b.Fatalf("GetRule: %v", err)
		}
		if len(got.Scope.AgentIDs) != 3 {
			b.Fatalf("scope.AgentIDs = %d, want 3", len(got.Scope.AgentIDs))
		}
		if len(got.Scope.DataClassifications) != 3 {
			b.Fatalf("scope.DataClassifications = %d, want 3", len(got.Scope.DataClassifications))
		}
	}
}

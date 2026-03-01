//go:build integration

package economics

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
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

func seedAgentForBench(b *testing.B, db *sql.DB, name string) string {
	b.Helper()
	var id string
	err := db.QueryRow(
		`INSERT INTO agents (name, owner_id, status, labels)
		 VALUES ($1, 'bench-owner', 'active', '{}')
		 RETURNING id`, name,
	).Scan(&id)
	if err != nil {
		b.Fatalf("seed agent %q: %v", name, err)
	}
	return id
}

func BenchmarkDB_InsertUsage(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		rec := &models.UsageRecord{
			TenantID:     benchTenant,
			AgentID:      "agent-usage",
			WorkspaceID:  "ws-usage",
			ResourceType: "llm_tokens",
			Unit:         "tokens",
			Quantity:     1500,
			Cost:         0.03,
		}
		if err := repo.InsertUsage(ctx, rec); err != nil {
			b.Fatalf("InsertUsage: %v", err)
		}
	}
}

func BenchmarkDB_UpsertBudget(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	now := time.Now().Truncate(time.Microsecond)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		budget := &models.Budget{
			TenantID:    benchTenant,
			AgentID:     "agent-upsert",
			Currency:    "USD",
			Limit:       100.0,
			Used:        0,
			PeriodStart: now,
			PeriodEnd:   now.Add(30 * 24 * time.Hour),
		}
		if err := repo.UpsertBudget(ctx, budget); err != nil {
			b.Fatalf("UpsertBudget: %v", err)
		}
	}
}

func BenchmarkDB_AddUsedAmount(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	now := time.Now().Truncate(time.Microsecond)
	budget := &models.Budget{
		TenantID:    benchTenant,
		AgentID:     "agent-add",
		Currency:    "USD",
		Limit:       1e9, // large limit to avoid overflow
		Used:        0,
		PeriodStart: now,
		PeriodEnd:   now.Add(30 * 24 * time.Hour),
	}
	if err := repo.UpsertBudget(ctx, budget); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := repo.AddUsedAmount(ctx, benchTenant, "agent-add", 0.01); err != nil {
			b.Fatalf("AddUsedAmount: %v", err)
		}
	}
}

func BenchmarkDB_AddUsedAmount_Concurrent(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	now := time.Now().Truncate(time.Microsecond)
	budget := &models.Budget{
		TenantID:    benchTenant,
		AgentID:     "agent-concurrent",
		Currency:    "USD",
		Limit:       1e12,
		Used:        0,
		PeriodStart: now,
		PeriodEnd:   now.Add(30 * 24 * time.Hour),
	}
	if err := repo.UpsertBudget(ctx, budget); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := repo.AddUsedAmount(ctx, benchTenant, "agent-concurrent", 0.001); err != nil {
				b.Fatalf("AddUsedAmount: %v", err)
			}
		}
	})
}

func BenchmarkDB_GetBudget(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	now := time.Now().Truncate(time.Microsecond)
	budget := &models.Budget{
		TenantID:         benchTenant,
		AgentID:          "agent-get",
		Currency:         "USD",
		Limit:            500.0,
		Used:             123.45,
		PeriodStart:      now,
		PeriodEnd:        now.Add(30 * 24 * time.Hour),
		OnExceeded:       "warn",
		WarningThreshold: 0.8,
	}
	if err := repo.UpsertBudget(ctx, budget); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		got, err := repo.GetBudget(ctx, benchTenant, "agent-get")
		if err != nil {
			b.Fatalf("GetBudget: %v", err)
		}
		if got == nil {
			b.Fatal("expected budget, got nil")
		}
	}
}

func BenchmarkDB_GetCostReport(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	now := time.Now()
	// Insert 500 usage records across 5 resource types.
	resourceTypes := []string{"llm_tokens", "compute_seconds", "storage_mb", "network_egress", "api_calls"}
	for i := 0; i < 500; i++ {
		rec := &models.UsageRecord{
			TenantID:     benchTenant,
			AgentID:      "agent-report",
			WorkspaceID:  fmt.Sprintf("ws-%d", i%10),
			ResourceType: resourceTypes[i%len(resourceTypes)],
			Unit:         "units",
			Quantity:     float64(i + 1),
			Cost:         float64(i+1) * 0.01,
		}
		if err := repo.InsertUsage(ctx, rec); err != nil {
			b.Fatalf("seed[%d]: %v", i, err)
		}
	}

	start := now.Add(-1 * time.Minute)
	end := now.Add(1 * time.Minute)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		costs, err := repo.GetCostReport(ctx, benchTenant, "agent-report", start, end)
		if err != nil {
			b.Fatalf("GetCostReport: %v", err)
		}
		if len(costs) != 5 {
			b.Fatalf("expected 5 resource types, got %d", len(costs))
		}
	}
}

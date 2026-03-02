//go:build integration

package activity

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

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

func benchWSUUID(n int) string {
	return fmt.Sprintf("c0000000-0000-0000-0000-%012d", n)
}

func BenchmarkDB_InsertAction(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	agentID := seedAgentForBench(b, testDB.DB, "insert-agent")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		rec := &models.ActionRecord{
			TenantID:    benchTenant,
			WorkspaceID: benchWSUUID(1),
			AgentID:     agentID,
			ToolName:    "http_fetch",
			Parameters:  json.RawMessage(`{"url":"https://example.com"}`),
			Result:      json.RawMessage(`{"status":200}`),
			Outcome:     models.ActionOutcomeAllowed,
		}
		if err := repo.InsertAction(ctx, rec); err != nil {
			b.Fatalf("InsertAction: %v", err)
		}
	}
}

func BenchmarkDB_InsertAction_LargeJSON(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	agentID := seedAgentForBench(b, testDB.DB, "large-json-agent")

	// Build a ~4KB JSON payload.
	largeValue := strings.Repeat("x", 3900)
	params := json.RawMessage(fmt.Sprintf(`{"data":"%s","count":42}`, largeValue))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		rec := &models.ActionRecord{
			TenantID:    benchTenant,
			WorkspaceID: benchWSUUID(2),
			AgentID:     agentID,
			ToolName:    "file_write",
			Parameters:  params,
			Outcome:     models.ActionOutcomeAllowed,
		}
		if err := repo.InsertAction(ctx, rec); err != nil {
			b.Fatalf("InsertAction: %v", err)
		}
	}
}

func BenchmarkDB_InsertAction_Concurrent(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	agentID := seedAgentForBench(b, testDB.DB, "concurrent-agent")
	var counter atomic.Int64

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := counter.Add(1)
			rec := &models.ActionRecord{
				TenantID:    benchTenant,
				WorkspaceID: benchWSUUID(int(n % 10)),
				AgentID:     agentID,
				ToolName:    "bash",
				Parameters:  json.RawMessage(`{"cmd":"ls"}`),
				Outcome:     models.ActionOutcomeAllowed,
			}
			if err := repo.InsertAction(ctx, rec); err != nil {
				b.Fatalf("InsertAction: %v", err)
			}
		}
	})
}

func BenchmarkDB_GetAction(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	agentID := seedAgentForBench(b, testDB.DB, "get-agent")

	rec := &models.ActionRecord{
		TenantID:    benchTenant,
		WorkspaceID: benchWSUUID(3),
		AgentID:     agentID,
		ToolName:    "http_fetch",
		Parameters:  json.RawMessage(`{"url":"https://example.com"}`),
		Result:      json.RawMessage(`{"status":200}`),
		Outcome:     models.ActionOutcomeAllowed,
	}
	if err := repo.InsertAction(ctx, rec); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		got, err := repo.GetAction(ctx, benchTenant, rec.ID)
		if err != nil {
			b.Fatalf("GetAction: %v", err)
		}
		if got == nil {
			b.Fatal("expected record, got nil")
		}
	}
}

func BenchmarkDB_QueryActions_ByWorkspace(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	agentID := seedAgentForBench(b, testDB.DB, "ws-query-agent")

	// Insert 500 records across 10 workspaces.
	for i := 0; i < 500; i++ {
		rec := &models.ActionRecord{
			TenantID:    benchTenant,
			WorkspaceID: benchWSUUID(i % 10),
			AgentID:     agentID,
			ToolName:    "test",
			Parameters:  json.RawMessage(`{}`),
			Outcome:     models.ActionOutcomeAllowed,
		}
		if err := repo.InsertAction(ctx, rec); err != nil {
			b.Fatalf("seed[%d]: %v", i, err)
		}
	}

	targetWS := benchWSUUID(0) // should have ~50 records

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		records, err := repo.QueryActions(ctx, benchTenant, QueryFilter{
			WorkspaceID: targetWS,
			Limit:       100,
		})
		if err != nil {
			b.Fatalf("QueryActions: %v", err)
		}
		if len(records) != 50 {
			b.Fatalf("expected 50, got %d", len(records))
		}
	}
}

func BenchmarkDB_QueryActions_TimeRange(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	agentID := seedAgentForBench(b, testDB.DB, "time-agent")

	// Insert records — all will have recorded_at = now().
	for i := 0; i < 200; i++ {
		rec := &models.ActionRecord{
			TenantID:    benchTenant,
			WorkspaceID: benchWSUUID(50),
			AgentID:     agentID,
			ToolName:    "test",
			Parameters:  json.RawMessage(`{}`),
			Outcome:     models.ActionOutcomeAllowed,
		}
		if err := repo.InsertAction(ctx, rec); err != nil {
			b.Fatalf("seed[%d]: %v", i, err)
		}
	}

	now := time.Now()
	start := now.Add(-1 * time.Minute)
	end := now.Add(1 * time.Minute)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		records, err := repo.QueryActions(ctx, benchTenant, QueryFilter{
			StartTime: &start,
			EndTime:   &end,
			Limit:     200,
		})
		if err != nil {
			b.Fatalf("QueryActions: %v", err)
		}
		if len(records) != 200 {
			b.Fatalf("expected 200, got %d", len(records))
		}
	}
}

func BenchmarkDB_AppendThroughput(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	agentID := seedAgentForBench(b, testDB.DB, "throughput-agent")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for i := 0; i < 1000; i++ {
			rec := &models.ActionRecord{
				TenantID:    benchTenant,
				WorkspaceID: benchWSUUID(99),
				AgentID:     agentID,
				ToolName:    "batch_tool",
				Parameters:  json.RawMessage(`{"i":` + fmt.Sprintf("%d", i) + `}`),
				Outcome:     models.ActionOutcomeAllowed,
			}
			if err := repo.InsertAction(ctx, rec); err != nil {
				b.Fatalf("InsertAction[%d]: %v", i, err)
			}
		}
	}
}

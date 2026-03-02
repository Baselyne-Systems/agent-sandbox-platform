//go:build integration

package human

import (
	"context"
	"database/sql"
	"fmt"
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

func benchHumanWSUUID(n int) string {
	return fmt.Sprintf("d0000000-0000-0000-0000-%012d", n)
}

func BenchmarkDB_CreateRequest(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	agentID := seedAgentForBench(b, testDB.DB, "create-req-agent")

	expires := time.Now().Add(time.Hour)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		req := &models.HumanRequest{
			TenantID:    benchTenant,
			WorkspaceID: benchHumanWSUUID(1),
			AgentID:     agentID,
			Question:    "Approve this action?",
			Options:     []string{"approve", "deny", "escalate"},
			Context:     "Invoice processing for order #1234",
			Status:      models.HumanRequestStatusPending,
			Type:        models.HumanRequestTypeApproval,
			Urgency:     models.HumanRequestUrgencyNormal,
			ExpiresAt:   &expires,
		}
		if err := repo.CreateRequest(ctx, req); err != nil {
			b.Fatalf("CreateRequest: %v", err)
		}
	}
}

func BenchmarkDB_GetRequest(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	agentID := seedAgentForBench(b, testDB.DB, "get-req-agent")

	req := &models.HumanRequest{
		TenantID:    benchTenant,
		WorkspaceID: benchHumanWSUUID(2),
		AgentID:     agentID,
		Question:    "Approve?",
		Options:     []string{"yes", "no"},
		Status:      models.HumanRequestStatusPending,
		Type:        models.HumanRequestTypeApproval,
		Urgency:     models.HumanRequestUrgencyNormal,
	}
	if err := repo.CreateRequest(ctx, req); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		got, err := repo.GetRequest(ctx, benchTenant, req.ID)
		if err != nil {
			b.Fatalf("GetRequest: %v", err)
		}
		if got == nil {
			b.Fatal("expected request, got nil")
		}
	}
}

func BenchmarkDB_RespondToRequest(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	agentID := seedAgentForBench(b, testDB.DB, "respond-req-agent")

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		req := &models.HumanRequest{
			TenantID:    benchTenant,
			WorkspaceID: benchHumanWSUUID(3),
			AgentID:     agentID,
			Question:    "Approve?",
			Options:     []string{"yes", "no"},
			Status:      models.HumanRequestStatusPending,
			Type:        models.HumanRequestTypeApproval,
			Urgency:     models.HumanRequestUrgencyNormal,
		}
		if err := repo.CreateRequest(ctx, req); err != nil {
			b.Fatalf("seed: %v", err)
		}
		b.StartTimer()

		if err := repo.RespondToRequest(ctx, benchTenant, req.ID, "approved", "admin-001"); err != nil {
			b.Fatalf("RespondToRequest: %v", err)
		}
	}
}

func BenchmarkDB_ListRequests_100(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	agentID := seedAgentForBench(b, testDB.DB, "list-req-agent")

	for i := 0; i < 100; i++ {
		req := &models.HumanRequest{
			TenantID:    benchTenant,
			WorkspaceID: benchHumanWSUUID(i % 10),
			AgentID:     agentID,
			Question:    fmt.Sprintf("Question %d?", i),
			Options:     []string{"yes", "no"},
			Status:      models.HumanRequestStatusPending,
			Type:        models.HumanRequestTypeQuestion,
			Urgency:     models.HumanRequestUrgencyNormal,
		}
		if err := repo.CreateRequest(ctx, req); err != nil {
			b.Fatalf("seed[%d]: %v", i, err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		requests, err := repo.ListRequests(ctx, benchTenant, "", "", "", 25)
		if err != nil {
			b.Fatalf("ListRequests: %v", err)
		}
		if len(requests) != 25 {
			b.Fatalf("expected 25, got %d", len(requests))
		}
	}
}

func BenchmarkDB_UpsertDeliveryChannel(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cfg := &models.DeliveryChannelConfig{
			TenantID:    benchTenant,
			UserID:      "user-bench",
			ChannelType: "slack",
			Endpoint:    fmt.Sprintf("https://hooks.slack.com/services/bench/%d", b.N),
			Enabled:     true,
		}
		if err := repo.UpsertDeliveryChannel(ctx, cfg); err != nil {
			b.Fatalf("UpsertDeliveryChannel: %v", err)
		}
	}
}

func BenchmarkDB_UpsertTimeoutPolicy(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		policy := &models.TimeoutPolicy{
			TenantID:          benchTenant,
			Scope:             "agent",
			ScopeID:           "agent-bench",
			TimeoutSecs:       300,
			Action:            "escalate",
			EscalationTargets: []string{"admin@example.com", "oncall@example.com"},
		}
		if err := repo.UpsertTimeoutPolicy(ctx, policy); err != nil {
			b.Fatalf("UpsertTimeoutPolicy: %v", err)
		}
	}
}

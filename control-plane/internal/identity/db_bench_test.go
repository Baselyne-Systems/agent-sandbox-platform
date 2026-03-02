//go:build integration

package identity

import (
	"context"
	"fmt"
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

func BenchmarkDB_CreateAgent(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		agent := &models.Agent{
			TenantID:    benchTenant,
			Name:        fmt.Sprintf("agent-%d", b.N),
			Description: "benchmark agent",
			OwnerID:     "owner-bench",
			Status:      models.AgentStatusActive,
			Labels:      map[string]string{"env": "bench", "tier": "gold"},
		}
		if err := repo.CreateAgent(ctx, agent); err != nil {
			b.Fatalf("CreateAgent: %v", err)
		}
	}
}

func BenchmarkDB_GetAgent(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	agent := &models.Agent{
		TenantID: benchTenant,
		Name:     "get-agent",
		OwnerID:  "owner-bench",
		Status:   models.AgentStatusActive,
		Labels:   map[string]string{"env": "bench"},
	}
	if err := repo.CreateAgent(ctx, agent); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		got, err := repo.GetAgent(ctx, benchTenant, agent.ID)
		if err != nil {
			b.Fatalf("GetAgent: %v", err)
		}
		if got == nil {
			b.Fatal("expected agent, got nil")
		}
	}
}

func BenchmarkDB_CreateAgent_Concurrent(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()
	var counter atomic.Int64

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := counter.Add(1)
			agent := &models.Agent{
				TenantID: benchTenant,
				Name:     fmt.Sprintf("concurrent-agent-%d", n),
				OwnerID:  "owner-bench",
				Status:   models.AgentStatusActive,
				Labels:   map[string]string{"env": "bench"},
			}
			if err := repo.CreateAgent(ctx, agent); err != nil {
				b.Fatalf("CreateAgent: %v", err)
			}
		}
	})
}

func BenchmarkDB_ListAgents_100(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		a := &models.Agent{
			TenantID: benchTenant,
			Name:     fmt.Sprintf("list-agent-%d", i),
			OwnerID:  "owner-bench",
			Status:   models.AgentStatusActive,
			Labels:   map[string]string{"env": "bench"},
		}
		if err := repo.CreateAgent(ctx, a); err != nil {
			b.Fatalf("seed[%d]: %v", i, err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		agents, err := repo.ListAgents(ctx, benchTenant, "", "", "", 25)
		if err != nil {
			b.Fatalf("ListAgents: %v", err)
		}
		if len(agents) != 25 {
			b.Fatalf("expected 25, got %d", len(agents))
		}
	}
}

func BenchmarkDB_MintCredential(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	agent := &models.Agent{
		TenantID: benchTenant,
		Name:     "cred-agent",
		OwnerID:  "owner-bench",
		Status:   models.AgentStatusActive,
		Labels:   map[string]string{},
	}
	if err := repo.CreateAgent(ctx, agent); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cred := &models.ScopedCredential{
			TenantID:  benchTenant,
			AgentID:   agent.ID,
			Scopes:    []string{"read", "write", "execute"},
			TokenHash: fmt.Sprintf("hash-%d", b.N),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := repo.CreateCredential(ctx, cred); err != nil {
			b.Fatalf("CreateCredential: %v", err)
		}
	}
}

func BenchmarkDB_DeactivateAgent(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		agent := &models.Agent{
			TenantID: benchTenant,
			Name:     fmt.Sprintf("deactivate-%d", b.N),
			OwnerID:  "owner-bench",
			Status:   models.AgentStatusActive,
			Labels:   map[string]string{},
		}
		if err := repo.CreateAgent(ctx, agent); err != nil {
			b.Fatalf("seed agent: %v", err)
		}
		for j := 0; j < 10; j++ {
			cred := &models.ScopedCredential{
				TenantID:  benchTenant,
				AgentID:   agent.ID,
				Scopes:    []string{"read"},
				TokenHash: fmt.Sprintf("hash-%d-%d", b.N, j),
				ExpiresAt: time.Now().Add(time.Hour),
			}
			if err := repo.CreateCredential(ctx, cred); err != nil {
				b.Fatalf("seed cred: %v", err)
			}
		}
		b.StartTimer()

		if err := repo.DeactivateAgent(ctx, benchTenant, agent.ID); err != nil {
			b.Fatalf("DeactivateAgent: %v", err)
		}
	}
}

func BenchmarkDB_JSONBLabelsRoundTrip(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	labels := make(map[string]string, 20)
	for i := 0; i < 20; i++ {
		labels[fmt.Sprintf("key-%d", i)] = fmt.Sprintf("value-%d-with-some-extra-data-for-size", i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		agent := &models.Agent{
			TenantID: benchTenant,
			Name:     fmt.Sprintf("labels-%d", b.N),
			OwnerID:  "owner-bench",
			Status:   models.AgentStatusActive,
			Labels:   labels,
		}
		if err := repo.CreateAgent(ctx, agent); err != nil {
			b.Fatalf("CreateAgent: %v", err)
		}

		got, err := repo.GetAgent(ctx, benchTenant, agent.ID)
		if err != nil {
			b.Fatalf("GetAgent: %v", err)
		}
		if len(got.Labels) != 20 {
			b.Fatalf("labels count = %d, want 20", len(got.Labels))
		}
	}
}

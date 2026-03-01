//go:build integration

package workspace

import (
	"context"
	"fmt"
	"testing"

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

func makeWorkspaceSpec() models.WorkspaceSpec {
	return models.WorkspaceSpec{
		MemoryMb:          1024,
		CpuMillicores:     2000,
		DiskMb:            4096,
		MaxDurationSecs:   7200,
		AllowedTools:      []string{"http_fetch", "bash", "file_write"},
		GuardrailPolicyID: "policy-bench",
		EnvVars:           map[string]string{"ENV": "bench", "API_KEY": "secret"},
		ContainerImage:    "ubuntu:22.04",
		EgressAllowlist:   []string{"*.example.com", "api.service.internal"},
		IsolationTier:     models.IsolationTierStandard,
	}
}

func BenchmarkDB_CreateWorkspace(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ws := &models.Workspace{
			TenantID: benchTenant,
			AgentID:  "agent-bench",
			TaskID:   "task-bench",
			Status:   models.WorkspaceStatusPending,
			Spec:     makeWorkspaceSpec(),
			HostID:   "host-bench",
		}
		if err := repo.CreateWorkspace(ctx, ws); err != nil {
			b.Fatalf("CreateWorkspace: %v", err)
		}
	}
}

func BenchmarkDB_GetWorkspace(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	ws := &models.Workspace{
		TenantID: benchTenant,
		AgentID:  "agent-get",
		TaskID:   "task-get",
		Status:   models.WorkspaceStatusRunning,
		Spec:     makeWorkspaceSpec(),
		HostID:   "host-get",
	}
	if err := repo.CreateWorkspace(ctx, ws); err != nil {
		b.Fatalf("seed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		got, err := repo.GetWorkspace(ctx, benchTenant, ws.ID)
		if err != nil {
			b.Fatalf("GetWorkspace: %v", err)
		}
		if got == nil {
			b.Fatal("expected workspace, got nil")
		}
	}
}

func BenchmarkDB_ListWorkspaces_100(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		ws := &models.Workspace{
			TenantID: benchTenant,
			AgentID:  fmt.Sprintf("agent-%d", i%5),
			Status:   models.WorkspaceStatusRunning,
			Spec: models.WorkspaceSpec{
				AllowedTools:    []string{},
				EnvVars:         map[string]string{},
				EgressAllowlist: []string{},
			},
		}
		if err := repo.CreateWorkspace(ctx, ws); err != nil {
			b.Fatalf("seed[%d]: %v", i, err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		workspaces, err := repo.ListWorkspaces(ctx, benchTenant, "", "", "", 25)
		if err != nil {
			b.Fatalf("ListWorkspaces: %v", err)
		}
		if len(workspaces) != 25 {
			b.Fatalf("expected 25, got %d", len(workspaces))
		}
	}
}

func BenchmarkDB_UpdateWorkspaceStatus(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	ws := &models.Workspace{
		TenantID: benchTenant,
		AgentID:  "agent-update",
		Status:   models.WorkspaceStatusPending,
		Spec: models.WorkspaceSpec{
			AllowedTools:    []string{},
			EnvVars:         map[string]string{},
			EgressAllowlist: []string{},
		},
	}
	if err := repo.CreateWorkspace(ctx, ws); err != nil {
		b.Fatalf("seed: %v", err)
	}

	statuses := []models.WorkspaceStatus{
		models.WorkspaceStatusCreating,
		models.WorkspaceStatusRunning,
		models.WorkspaceStatusPaused,
		models.WorkspaceStatusRunning,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		status := statuses[b.N%len(statuses)]
		if err := repo.UpdateWorkspaceStatus(ctx, benchTenant, ws.ID, status, "host-1", "10.0.0.1:8080", "sandbox-1"); err != nil {
			b.Fatalf("UpdateWorkspaceStatus: %v", err)
		}
	}
}

func BenchmarkDB_TerminateWorkspace(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		ws := &models.Workspace{
			TenantID: benchTenant,
			AgentID:  "agent-term",
			Status:   models.WorkspaceStatusRunning,
			Spec: models.WorkspaceSpec{
				AllowedTools:    []string{},
				EnvVars:         map[string]string{},
				EgressAllowlist: []string{},
			},
		}
		if err := repo.CreateWorkspace(ctx, ws); err != nil {
			b.Fatalf("seed: %v", err)
		}
		b.StartTimer()

		if err := repo.TerminateWorkspace(ctx, benchTenant, ws.ID, "benchmark shutdown"); err != nil {
			b.Fatalf("TerminateWorkspace: %v", err)
		}
	}
}

func BenchmarkDB_JSONBSpecRoundTrip(b *testing.B) {
	repo := benchSetup(b)
	ctx := context.Background()

	fullSpec := models.WorkspaceSpec{
		MemoryMb:           2048,
		CpuMillicores:      4000,
		DiskMb:             8192,
		MaxDurationSecs:    14400,
		AllowedTools:       []string{"http_fetch", "bash", "file_write", "file_read", "db_query", "code_exec"},
		GuardrailPolicyID:  "policy-full",
		EnvVars:            map[string]string{"ENV": "prod", "API_KEY": "k1", "DB_URL": "pg://host", "REGION": "us-east-1", "DEBUG": "false"},
		ContainerImage:     "custom-runtime:v2.1.0",
		EgressAllowlist:    []string{"*.example.com", "api.service.internal", "cdn.assets.io", "metrics.collector:9090"},
		IsolationTier:      models.IsolationTierHardened,
		DataClassification: "pii",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ws := &models.Workspace{
			TenantID: benchTenant,
			AgentID:  fmt.Sprintf("agent-spec-%d", b.N),
			TaskID:   "task-spec",
			Status:   models.WorkspaceStatusRunning,
			Spec:     fullSpec,
			HostID:   "host-spec",
		}
		if err := repo.CreateWorkspace(ctx, ws); err != nil {
			b.Fatalf("CreateWorkspace: %v", err)
		}

		got, err := repo.GetWorkspace(ctx, benchTenant, ws.ID)
		if err != nil {
			b.Fatalf("GetWorkspace: %v", err)
		}
		if len(got.Spec.AllowedTools) != 6 {
			b.Fatalf("AllowedTools = %d, want 6", len(got.Spec.AllowedTools))
		}
		if len(got.Spec.EnvVars) != 5 {
			b.Fatalf("EnvVars = %d, want 5", len(got.Spec.EnvVars))
		}
		if len(got.Spec.EgressAllowlist) != 4 {
			b.Fatalf("EgressAllowlist = %d, want 4", len(got.Spec.EgressAllowlist))
		}
	}
}

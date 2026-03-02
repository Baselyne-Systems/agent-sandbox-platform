//go:build integration

package workspace

import (
	"context"
	"database/sql"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/testutil"
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

func TestInteg_CreateAndGetWorkspace_JSONBRoundTrip(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	ws := &models.Workspace{
		TenantID: "test-tenant",
		AgentID:  "agent-001",
		TaskID:   "task-001",
		Status:   models.WorkspaceStatusRunning,
		Spec: models.WorkspaceSpec{
			MemoryMb:          1024,
			CpuMillicores:     2000,
			DiskMb:            4096,
			MaxDurationSecs:   7200,
			AllowedTools:      []string{"http_fetch", "bash", "file_write"},
			GuardrailPolicyID: "policy-001",
			EnvVars:           map[string]string{"API_KEY": "secret", "ENV": "prod"},
		},
		HostID: "host-001",
	}
	if err := repo.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	if ws.ID == "" {
		t.Fatal("expected server-generated ID")
	}
	if ws.CreatedAt.IsZero() || ws.UpdatedAt.IsZero() {
		t.Fatal("expected server-generated timestamps")
	}

	got, err := repo.GetWorkspace(ctx, "test-tenant", ws.ID)
	if err != nil {
		t.Fatalf("GetWorkspace: %v", err)
	}
	if got == nil {
		t.Fatal("expected workspace, got nil")
	}

	// AllowedTools JSONB round-trip
	if len(got.Spec.AllowedTools) != 3 {
		t.Errorf("AllowedTools len = %d, want 3", len(got.Spec.AllowedTools))
	}
	if got.Spec.AllowedTools[0] != "http_fetch" {
		t.Errorf("AllowedTools[0] = %q, want %q", got.Spec.AllowedTools[0], "http_fetch")
	}

	// EnvVars JSONB round-trip
	if got.Spec.EnvVars["API_KEY"] != "secret" || got.Spec.EnvVars["ENV"] != "prod" {
		t.Errorf("EnvVars round-trip failed: %v", got.Spec.EnvVars)
	}

	if got.Spec.MemoryMb != 1024 {
		t.Errorf("MemoryMb = %d, want 1024", got.Spec.MemoryMb)
	}
}

func TestInteg_GetWorkspace_NullableExpiresAt(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	// Without ExpiresAt
	wsNoExpiry := &models.Workspace{
		TenantID: "test-tenant",
		AgentID:  "agent-no-exp",
		Status:   models.WorkspaceStatusRunning,
		Spec: models.WorkspaceSpec{
			AllowedTools: []string{},
			EnvVars:      map[string]string{},
		},
	}
	if err := repo.CreateWorkspace(ctx, wsNoExpiry); err != nil {
		t.Fatalf("CreateWorkspace no expiry: %v", err)
	}

	got1, err := repo.GetWorkspace(ctx, "test-tenant", wsNoExpiry.ID)
	if err != nil {
		t.Fatalf("GetWorkspace no expiry: %v", err)
	}
	if got1.ExpiresAt != nil {
		t.Errorf("expected nil ExpiresAt, got %v", got1.ExpiresAt)
	}

	// With ExpiresAt
	expires := time.Now().Add(time.Hour).Truncate(time.Microsecond)
	wsWithExpiry := &models.Workspace{
		TenantID:  "test-tenant",
		AgentID:   "agent-with-exp",
		Status:    models.WorkspaceStatusRunning,
		ExpiresAt: &expires,
		Spec: models.WorkspaceSpec{
			AllowedTools: []string{},
			EnvVars:      map[string]string{},
		},
	}
	if err := repo.CreateWorkspace(ctx, wsWithExpiry); err != nil {
		t.Fatalf("CreateWorkspace with expiry: %v", err)
	}

	got2, err := repo.GetWorkspace(ctx, "test-tenant", wsWithExpiry.ID)
	if err != nil {
		t.Fatalf("GetWorkspace with expiry: %v", err)
	}
	if got2.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt, got nil")
	}
	if got2.ExpiresAt.Sub(expires).Abs() > time.Millisecond {
		t.Errorf("ExpiresAt = %v, want ~%v", got2.ExpiresAt, expires)
	}
}

func TestInteg_GetWorkspace_NotFound(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	got, err := repo.GetWorkspace(ctx, "test-tenant", "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestInteg_ListWorkspaces_Filters(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	workspaces := []struct {
		agentID string
		status  models.WorkspaceStatus
	}{
		{"agent-A", models.WorkspaceStatusRunning},
		{"agent-A", models.WorkspaceStatusTerminated},
		{"agent-B", models.WorkspaceStatusRunning},
	}
	for _, w := range workspaces {
		ws := &models.Workspace{
			TenantID: "test-tenant",
			AgentID:  w.agentID,
			Status:   w.status,
			Spec: models.WorkspaceSpec{
				AllowedTools: []string{},
				EnvVars:      map[string]string{},
			},
		}
		if err := repo.CreateWorkspace(ctx, ws); err != nil {
			t.Fatalf("CreateWorkspace: %v", err)
		}
	}

	// By agent
	byAgent, err := repo.ListWorkspaces(ctx, "test-tenant", "agent-A", "", "", 10)
	if err != nil {
		t.Fatalf("ListWorkspaces agent: %v", err)
	}
	if len(byAgent) != 2 {
		t.Errorf("agent filter count = %d, want 2", len(byAgent))
	}

	// By status
	byStatus, err := repo.ListWorkspaces(ctx, "test-tenant", "", models.WorkspaceStatusRunning, "", 10)
	if err != nil {
		t.Fatalf("ListWorkspaces status: %v", err)
	}
	if len(byStatus) != 2 {
		t.Errorf("status filter count = %d, want 2", len(byStatus))
	}
}

func TestInteg_ListWorkspaces_Pagination(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	var ids []string
	for i := 0; i < 5; i++ {
		ws := &models.Workspace{
			TenantID: "test-tenant",
			AgentID:  "agent-page",
			Status:   models.WorkspaceStatusRunning,
			Spec: models.WorkspaceSpec{
				AllowedTools: []string{},
				EnvVars:      map[string]string{},
			},
		}
		if err := repo.CreateWorkspace(ctx, ws); err != nil {
			t.Fatalf("CreateWorkspace[%d]: %v", i, err)
		}
		ids = append(ids, ws.ID)
	}
	sort.Strings(ids)

	page1, err := repo.ListWorkspaces(ctx, "test-tenant", "", "", "", 2)
	if err != nil {
		t.Fatalf("ListWorkspaces page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 len = %d, want 2", len(page1))
	}

	page2, err := repo.ListWorkspaces(ctx, "test-tenant", "", "", page1[1].ID, 2)
	if err != nil {
		t.Fatalf("ListWorkspaces page2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("page2 len = %d, want 2", len(page2))
	}

	page3, err := repo.ListWorkspaces(ctx, "test-tenant", "", "", page2[1].ID, 2)
	if err != nil {
		t.Fatalf("ListWorkspaces page3: %v", err)
	}
	if len(page3) != 1 {
		t.Fatalf("page3 len = %d, want 1", len(page3))
	}
}

func TestInteg_TerminateWorkspace_Success(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	ws := &models.Workspace{
		TenantID: "test-tenant",
		AgentID:  "agent-term",
		Status:   models.WorkspaceStatusRunning,
		Spec: models.WorkspaceSpec{
			AllowedTools: []string{},
			EnvVars:      map[string]string{},
		},
	}
	if err := repo.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	originalUpdatedAt := ws.UpdatedAt

	if err := repo.TerminateWorkspace(ctx, "test-tenant", ws.ID, "test shutdown"); err != nil {
		t.Fatalf("TerminateWorkspace: %v", err)
	}

	got, err := repo.GetWorkspace(ctx, "test-tenant", ws.ID)
	if err != nil {
		t.Fatalf("GetWorkspace: %v", err)
	}
	if got.Status != models.WorkspaceStatusTerminated {
		t.Errorf("status = %q, want %q", got.Status, models.WorkspaceStatusTerminated)
	}
	if !got.UpdatedAt.After(originalUpdatedAt) {
		t.Error("updated_at should have changed")
	}
}

func TestInteg_TerminateWorkspace_AlreadyTerminal(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	ws := &models.Workspace{
		TenantID: "test-tenant",
		AgentID:  "agent-already-term",
		Status:   models.WorkspaceStatusRunning,
		Spec: models.WorkspaceSpec{
			AllowedTools: []string{},
			EnvVars:      map[string]string{},
		},
	}
	if err := repo.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	if err := repo.TerminateWorkspace(ctx, "test-tenant", ws.ID, "first"); err != nil {
		t.Fatalf("first TerminateWorkspace: %v", err)
	}

	err := repo.TerminateWorkspace(ctx, "test-tenant", ws.ID, "second")
	if err != ErrWorkspaceAlreadyTerminal {
		t.Errorf("error = %v, want ErrWorkspaceAlreadyTerminal", err)
	}
}

func TestInteg_TerminateWorkspace_NotFound(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	err := repo.TerminateWorkspace(ctx, "test-tenant", "00000000-0000-0000-0000-000000000000", "reason")
	if err != ErrWorkspaceNotFound {
		t.Errorf("error = %v, want ErrWorkspaceNotFound", err)
	}
}

func TestInteg_UpdateWorkspaceStatus_RoundTrip(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	ws := &models.Workspace{
		TenantID: "test-tenant",
		AgentID:  "agent-update",
		Status:   models.WorkspaceStatusPending,
		Spec: models.WorkspaceSpec{
			AllowedTools: []string{},
			EnvVars:      map[string]string{},
		},
	}
	if err := repo.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Update to creating
	if err := repo.UpdateWorkspaceStatus(ctx, "test-tenant", ws.ID, models.WorkspaceStatusCreating, "", "", ""); err != nil {
		t.Fatalf("UpdateWorkspaceStatus creating: %v", err)
	}

	got, err := repo.GetWorkspace(ctx, "test-tenant", ws.ID)
	if err != nil {
		t.Fatalf("GetWorkspace: %v", err)
	}
	if got.Status != models.WorkspaceStatusCreating {
		t.Errorf("status = %q, want creating", got.Status)
	}

	// Update to running with host and sandbox info
	if err := repo.UpdateWorkspaceStatus(ctx, "test-tenant", ws.ID, models.WorkspaceStatusRunning, "host-abc", "runtime.host1:50052", "sandbox-xyz"); err != nil {
		t.Fatalf("UpdateWorkspaceStatus running: %v", err)
	}

	got, err = repo.GetWorkspace(ctx, "test-tenant", ws.ID)
	if err != nil {
		t.Fatalf("GetWorkspace: %v", err)
	}
	if got.Status != models.WorkspaceStatusRunning {
		t.Errorf("status = %q, want running", got.Status)
	}
	if got.HostID != "host-abc" {
		t.Errorf("host_id = %q, want 'host-abc'", got.HostID)
	}
	if got.HostAddress != "runtime.host1:50052" {
		t.Errorf("host_address = %q, want 'runtime.host1:50052'", got.HostAddress)
	}
	if got.SandboxID != "sandbox-xyz" {
		t.Errorf("sandbox_id = %q, want 'sandbox-xyz'", got.SandboxID)
	}
}

func TestInteg_UpdateWorkspaceStatus_NotFound(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	err := repo.UpdateWorkspaceStatus(ctx, "test-tenant", "00000000-0000-0000-0000-000000000000", models.WorkspaceStatusRunning, "h", "a", "s")
	if err != ErrWorkspaceNotFound {
		t.Errorf("error = %v, want ErrWorkspaceNotFound", err)
	}
}

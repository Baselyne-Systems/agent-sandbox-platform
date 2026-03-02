//go:build integration

package human

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/testutil"
)

// wsUUID generates a deterministic valid UUID for workspace IDs in tests.
func wsUUID(n int) string {
	return fmt.Sprintf("b0000000-0000-0000-0000-%012d", n)
}

const testTenant = "test-tenant"

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

func TestInteg_CreateAndGetRequest_JSONBOptions(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "human-agent")

	req := &models.HumanRequest{
		TenantID:    testTenant,
		WorkspaceID: wsUUID(1),
		AgentID:     agentID,
		Question:    "Approve this action?",
		Options:     []string{"approve", "deny", "escalate"},
		Context:     "Invoice #1234",
		Status:      models.HumanRequestStatusPending,
	}
	if err := repo.CreateRequest(ctx, req); err != nil {
		t.Fatalf("CreateRequest: %v", err)
	}

	if req.ID == "" {
		t.Fatal("expected server-generated ID")
	}
	if req.CreatedAt.IsZero() {
		t.Fatal("expected server-generated created_at")
	}

	got, err := repo.GetRequest(ctx, testTenant, req.ID)
	if err != nil {
		t.Fatalf("GetRequest: %v", err)
	}
	if got == nil {
		t.Fatal("expected request, got nil")
	}
	if len(got.Options) != 3 || got.Options[0] != "approve" || got.Options[1] != "deny" || got.Options[2] != "escalate" {
		t.Errorf("options round-trip failed: %v", got.Options)
	}
	if got.Question != "Approve this action?" {
		t.Errorf("question = %q, want %q", got.Question, "Approve this action?")
	}
	if got.Status != models.HumanRequestStatusPending {
		t.Errorf("status = %q, want %q", got.Status, models.HumanRequestStatusPending)
	}
}

func TestInteg_GetRequest_ExpiresAt(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "expires-agent")

	expires := time.Now().Add(time.Hour).Truncate(time.Microsecond)
	req := &models.HumanRequest{
		TenantID:    testTenant,
		WorkspaceID: wsUUID(2),
		AgentID:     agentID,
		Question:    "Timeout test",
		Options:     []string{"yes", "no"},
		Status:      models.HumanRequestStatusPending,
		ExpiresAt:   &expires,
	}
	if err := repo.CreateRequest(ctx, req); err != nil {
		t.Fatalf("CreateRequest: %v", err)
	}

	got, err := repo.GetRequest(ctx, testTenant, req.ID)
	if err != nil {
		t.Fatalf("GetRequest: %v", err)
	}
	if got.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt, got nil")
	}
	// PostgreSQL has microsecond precision
	if got.ExpiresAt.Sub(expires).Abs() > time.Millisecond {
		t.Errorf("ExpiresAt = %v, want ~%v", got.ExpiresAt, expires)
	}
}

func TestInteg_GetRequest_NotFound(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	got, err := repo.GetRequest(ctx, testTenant, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestInteg_RespondToRequest_Success(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "respond-agent")

	req := &models.HumanRequest{
		TenantID:    testTenant,
		WorkspaceID: wsUUID(3),
		AgentID:     agentID,
		Question:    "Approve?",
		Options:     []string{"yes", "no"},
		Status:      models.HumanRequestStatusPending,
	}
	if err := repo.CreateRequest(ctx, req); err != nil {
		t.Fatalf("CreateRequest: %v", err)
	}

	if err := repo.RespondToRequest(ctx, testTenant, req.ID, "approved", "admin-001"); err != nil {
		t.Fatalf("RespondToRequest: %v", err)
	}

	got, err := repo.GetRequest(ctx, testTenant, req.ID)
	if err != nil {
		t.Fatalf("GetRequest: %v", err)
	}
	if got.Status != models.HumanRequestStatusResponded {
		t.Errorf("status = %q, want %q", got.Status, models.HumanRequestStatusResponded)
	}
	if got.Response != "approved" {
		t.Errorf("response = %q, want %q", got.Response, "approved")
	}
	if got.ResponderID != "admin-001" {
		t.Errorf("responder_id = %q, want %q", got.ResponderID, "admin-001")
	}
	if got.RespondedAt == nil {
		t.Fatal("expected responded_at to be set")
	}
}

func TestInteg_RespondToRequest_AlreadyResponded(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "double-respond-agent")

	req := &models.HumanRequest{
		TenantID:    testTenant,
		WorkspaceID: wsUUID(4),
		AgentID:     agentID,
		Question:    "Approve?",
		Options:     []string{"yes"},
		Status:      models.HumanRequestStatusPending,
	}
	if err := repo.CreateRequest(ctx, req); err != nil {
		t.Fatalf("CreateRequest: %v", err)
	}

	if err := repo.RespondToRequest(ctx, testTenant, req.ID, "yes", "admin-1"); err != nil {
		t.Fatalf("first RespondToRequest: %v", err)
	}

	err := repo.RespondToRequest(ctx, testTenant, req.ID, "no", "admin-2")
	if err != ErrRequestNotPending {
		t.Errorf("error = %v, want ErrRequestNotPending", err)
	}
}

func TestInteg_RespondToRequest_NonExistent(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	err := repo.RespondToRequest(ctx, testTenant, "00000000-0000-0000-0000-000000000000", "yes", "admin")
	if err != ErrRequestNotPending {
		t.Errorf("error = %v, want ErrRequestNotPending", err)
	}
}

func TestInteg_ListRequests_Filters(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "list-agent")

	wsA := wsUUID(10)
	wsB := wsUUID(11)
	requests := []struct {
		wsID   string
		status models.HumanRequestStatus
	}{
		{wsA, models.HumanRequestStatusPending},
		{wsA, models.HumanRequestStatusResponded},
		{wsB, models.HumanRequestStatusPending},
	}
	for _, r := range requests {
		req := &models.HumanRequest{
			TenantID:    testTenant,
			WorkspaceID: r.wsID,
			AgentID:     agentID,
			Question:    "q",
			Options:     []string{"y"},
			Status:      r.status,
		}
		if err := repo.CreateRequest(ctx, req); err != nil {
			t.Fatalf("CreateRequest: %v", err)
		}
	}

	// Workspace filter
	byWs, err := repo.ListRequests(ctx, testTenant, wsA, "", "", 10)
	if err != nil {
		t.Fatalf("ListRequests workspace: %v", err)
	}
	if len(byWs) != 2 {
		t.Errorf("workspace filter count = %d, want 2", len(byWs))
	}

	// Status filter
	byStatus, err := repo.ListRequests(ctx, testTenant, "", models.HumanRequestStatusPending, "", 10)
	if err != nil {
		t.Fatalf("ListRequests status: %v", err)
	}
	if len(byStatus) != 2 {
		t.Errorf("status filter count = %d, want 2", len(byStatus))
	}
}

func TestInteg_ListRequests_Pagination(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "page-agent")

	var ids []string
	for i := 0; i < 5; i++ {
		req := &models.HumanRequest{
			TenantID:    testTenant,
			WorkspaceID: wsUUID(20),
			AgentID:     agentID,
			Question:    "q",
			Options:     []string{"y"},
			Status:      models.HumanRequestStatusPending,
		}
		if err := repo.CreateRequest(ctx, req); err != nil {
			t.Fatalf("CreateRequest[%d]: %v", i, err)
		}
		ids = append(ids, req.ID)
	}
	sort.Strings(ids)

	page1, err := repo.ListRequests(ctx, testTenant, "", "", "", 2)
	if err != nil {
		t.Fatalf("ListRequests page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 len = %d, want 2", len(page1))
	}

	page2, err := repo.ListRequests(ctx, testTenant, "", "", page1[1].ID, 2)
	if err != nil {
		t.Fatalf("ListRequests page2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("page2 len = %d, want 2", len(page2))
	}

	page3, err := repo.ListRequests(ctx, testTenant, "", "", page2[1].ID, 2)
	if err != nil {
		t.Fatalf("ListRequests page3: %v", err)
	}
	if len(page3) != 1 {
		t.Fatalf("page3 len = %d, want 1", len(page3))
	}
}

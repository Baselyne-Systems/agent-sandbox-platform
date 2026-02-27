//go:build integration

package identity

import (
	"context"
	"database/sql"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/testutil"
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

func TestInteg_CreateAndGetAgent_JSONBRoundTrip(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	agent := &models.Agent{
		Name:        "test-agent",
		Description: "integration test agent",
		OwnerID:     "owner-1",
		Status:      models.AgentStatusActive,
		Labels:      map[string]string{"env": "test", "tier": "gold"},
	}
	if err := repo.CreateAgent(ctx, agent); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	if agent.ID == "" {
		t.Fatal("expected server-generated ID")
	}
	if agent.CreatedAt.IsZero() {
		t.Fatal("expected server-generated created_at")
	}
	if agent.UpdatedAt.IsZero() {
		t.Fatal("expected server-generated updated_at")
	}

	got, err := repo.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("GetAgent: %v", err)
	}
	if got == nil {
		t.Fatal("expected agent, got nil")
	}
	if got.Name != "test-agent" {
		t.Errorf("name = %q, want %q", got.Name, "test-agent")
	}
	if got.Labels["env"] != "test" || got.Labels["tier"] != "gold" {
		t.Errorf("labels round-trip failed: %v", got.Labels)
	}
	if got.Status != models.AgentStatusActive {
		t.Errorf("status = %q, want %q", got.Status, models.AgentStatusActive)
	}
}

func TestInteg_GetAgent_NotFound(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	got, err := repo.GetAgent(ctx, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestInteg_ListAgents_Pagination(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	var ids []string
	for i := 0; i < 5; i++ {
		a := &models.Agent{
			Name:    "agent-" + string(rune('A'+i)),
			OwnerID: "owner-1",
			Status:  models.AgentStatusActive,
			Labels:  map[string]string{},
		}
		if err := repo.CreateAgent(ctx, a); err != nil {
			t.Fatalf("CreateAgent[%d]: %v", i, err)
		}
		ids = append(ids, a.ID)
	}
	sort.Strings(ids) // UUIDs sorted ascending = DB order

	// Page 1: first 2
	page1, err := repo.ListAgents(ctx, "", "", "", 2)
	if err != nil {
		t.Fatalf("ListAgents page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 len = %d, want 2", len(page1))
	}
	if page1[0].ID != ids[0] || page1[1].ID != ids[1] {
		t.Errorf("page1 IDs = [%s, %s], want [%s, %s]", page1[0].ID, page1[1].ID, ids[0], ids[1])
	}

	// Page 2: next 2
	page2, err := repo.ListAgents(ctx, "", "", page1[1].ID, 2)
	if err != nil {
		t.Fatalf("ListAgents page2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("page2 len = %d, want 2", len(page2))
	}
	if page2[0].ID != ids[2] || page2[1].ID != ids[3] {
		t.Errorf("page2 IDs = [%s, %s], want [%s, %s]", page2[0].ID, page2[1].ID, ids[2], ids[3])
	}

	// Page 3: last 1
	page3, err := repo.ListAgents(ctx, "", "", page2[1].ID, 2)
	if err != nil {
		t.Fatalf("ListAgents page3: %v", err)
	}
	if len(page3) != 1 {
		t.Fatalf("page3 len = %d, want 1", len(page3))
	}
	if page3[0].ID != ids[4] {
		t.Errorf("page3 ID = %s, want %s", page3[0].ID, ids[4])
	}
}

func TestInteg_ListAgents_Filters(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	agents := []*models.Agent{
		{Name: "a1", OwnerID: "owner-A", Status: models.AgentStatusActive, Labels: map[string]string{}},
		{Name: "a2", OwnerID: "owner-A", Status: models.AgentStatusInactive, Labels: map[string]string{}},
		{Name: "a3", OwnerID: "owner-B", Status: models.AgentStatusActive, Labels: map[string]string{}},
	}
	for _, a := range agents {
		if err := repo.CreateAgent(ctx, a); err != nil {
			t.Fatalf("CreateAgent %s: %v", a.Name, err)
		}
	}

	// Filter by owner
	byOwner, err := repo.ListAgents(ctx, "owner-A", "", "", 10)
	if err != nil {
		t.Fatalf("ListAgents by owner: %v", err)
	}
	if len(byOwner) != 2 {
		t.Errorf("by owner count = %d, want 2", len(byOwner))
	}

	// Filter by status
	byStatus, err := repo.ListAgents(ctx, "", models.AgentStatusActive, "", 10)
	if err != nil {
		t.Fatalf("ListAgents by status: %v", err)
	}
	if len(byStatus) != 2 {
		t.Errorf("by status count = %d, want 2", len(byStatus))
	}

	// Combined
	combined, err := repo.ListAgents(ctx, "owner-A", models.AgentStatusActive, "", 10)
	if err != nil {
		t.Fatalf("ListAgents combined: %v", err)
	}
	if len(combined) != 1 {
		t.Errorf("combined count = %d, want 1", len(combined))
	}
	if combined[0].Name != "a1" {
		t.Errorf("combined name = %q, want %q", combined[0].Name, "a1")
	}
}

func TestInteg_DeactivateAgent_Transaction(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	agent := &models.Agent{
		Name:    "deactivate-me",
		OwnerID: "owner-1",
		Status:  models.AgentStatusActive,
		Labels:  map[string]string{},
	}
	if err := repo.CreateAgent(ctx, agent); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	// Create two credentials
	for i := 0; i < 2; i++ {
		cred := &models.ScopedCredential{
			AgentID:   agent.ID,
			Scopes:    []string{"read"},
			TokenHash: "hash-" + string(rune('0'+i)),
			ExpiresAt: time.Now().Add(time.Hour),
		}
		if err := repo.CreateCredential(ctx, cred); err != nil {
			t.Fatalf("CreateCredential[%d]: %v", i, err)
		}
	}

	if err := repo.DeactivateAgent(ctx, agent.ID); err != nil {
		t.Fatalf("DeactivateAgent: %v", err)
	}

	got, err := repo.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("GetAgent: %v", err)
	}
	if got.Status != models.AgentStatusInactive {
		t.Errorf("status = %q, want %q", got.Status, models.AgentStatusInactive)
	}

	// Verify all credentials are revoked
	var revokedCount int
	err = testDB.DB.QueryRow(
		`SELECT COUNT(*) FROM scoped_credentials WHERE agent_id = $1 AND revoked = true`, agent.ID,
	).Scan(&revokedCount)
	if err != nil {
		t.Fatalf("count revoked: %v", err)
	}
	if revokedCount != 2 {
		t.Errorf("revoked count = %d, want 2", revokedCount)
	}
}

func TestInteg_DeactivateAgent_NotFound(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	err := repo.DeactivateAgent(ctx, "00000000-0000-0000-0000-000000000000")
	if err != ErrAgentNotFound {
		t.Errorf("error = %v, want ErrAgentNotFound", err)
	}
}

func TestInteg_DeactivateAgent_AlreadyInactive(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	agent := &models.Agent{
		Name:    "already-inactive",
		OwnerID: "owner-1",
		Status:  models.AgentStatusActive,
		Labels:  map[string]string{},
	}
	if err := repo.CreateAgent(ctx, agent); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	if err := repo.DeactivateAgent(ctx, agent.ID); err != nil {
		t.Fatalf("first DeactivateAgent: %v", err)
	}

	// Second deactivation should be idempotent
	if err := repo.DeactivateAgent(ctx, agent.ID); err != nil {
		t.Errorf("second DeactivateAgent: unexpected error: %v", err)
	}
}

func TestInteg_RevokeCredential_RowsAffected(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	agent := &models.Agent{
		Name:    "cred-agent",
		OwnerID: "owner-1",
		Status:  models.AgentStatusActive,
		Labels:  map[string]string{},
	}
	if err := repo.CreateAgent(ctx, agent); err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	cred := &models.ScopedCredential{
		AgentID:   agent.ID,
		Scopes:    []string{"write"},
		TokenHash: "revoke-me-hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := repo.CreateCredential(ctx, cred); err != nil {
		t.Fatalf("CreateCredential: %v", err)
	}

	if err := repo.RevokeCredential(ctx, cred.ID); err != nil {
		t.Fatalf("RevokeCredential: %v", err)
	}

	// Re-revoke should return ErrCredentialNotFound
	err := repo.RevokeCredential(ctx, cred.ID)
	if err != ErrCredentialNotFound {
		t.Errorf("error = %v, want ErrCredentialNotFound", err)
	}
}

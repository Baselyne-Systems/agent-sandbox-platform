//go:build integration

package economics

import (
	"context"
	"database/sql"
	"os"
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

func TestInteg_InsertUsage(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	rec := &models.UsageRecord{
		TenantID:     "test-tenant",
		AgentID:      "agent-001",
		WorkspaceID:  "ws-001",
		ResourceType: "llm_tokens",
		Unit:         "tokens",
		Quantity:     1500,
		Cost:         0.03,
	}
	if err := repo.InsertUsage(ctx, rec); err != nil {
		t.Fatalf("InsertUsage: %v", err)
	}

	if rec.ID == "" {
		t.Fatal("expected server-generated ID")
	}
	if rec.RecordedAt.IsZero() {
		t.Fatal("expected server-generated recorded_at")
	}
}

func TestInteg_GetBudget_NotFound(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	got, err := repo.GetBudget(ctx, "test-tenant", "nonexistent-agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestInteg_UpsertBudget_Insert(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Microsecond)
	budget := &models.Budget{
		TenantID:    "test-tenant",
		AgentID:     "agent-budget-1",
		Currency:    "USD",
		Limit:       100.0,
		Used:        0,
		PeriodStart: now,
		PeriodEnd:   now.Add(30 * 24 * time.Hour),
	}
	if err := repo.UpsertBudget(ctx, budget); err != nil {
		t.Fatalf("UpsertBudget insert: %v", err)
	}

	if budget.ID == "" {
		t.Fatal("expected server-generated ID")
	}

	got, err := repo.GetBudget(ctx, "test-tenant", "agent-budget-1")
	if err != nil {
		t.Fatalf("GetBudget: %v", err)
	}
	if got == nil {
		t.Fatal("expected budget, got nil")
	}
	if got.Limit != 100.0 {
		t.Errorf("limit = %f, want 100.0", got.Limit)
	}
	if got.Currency != "USD" {
		t.Errorf("currency = %q, want %q", got.Currency, "USD")
	}
}

func TestInteg_UpsertBudget_OnConflict(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Microsecond)
	budget := &models.Budget{
		TenantID:    "test-tenant",
		AgentID:     "agent-upsert",
		Currency:    "USD",
		Limit:       50.0,
		Used:        10.0,
		PeriodStart: now,
		PeriodEnd:   now.Add(30 * 24 * time.Hour),
	}
	if err := repo.UpsertBudget(ctx, budget); err != nil {
		t.Fatalf("UpsertBudget first: %v", err)
	}
	originalID := budget.ID

	// Upsert with new limit and currency
	budget.Limit = 200.0
	budget.Currency = "EUR"
	budget.Used = 25.0
	if err := repo.UpsertBudget(ctx, budget); err != nil {
		t.Fatalf("UpsertBudget second: %v", err)
	}

	// ID should persist (same row)
	if budget.ID != originalID {
		t.Errorf("ID changed: %s → %s", originalID, budget.ID)
	}

	got, err := repo.GetBudget(ctx, "test-tenant", "agent-upsert")
	if err != nil {
		t.Fatalf("GetBudget: %v", err)
	}
	if got.Limit != 200.0 {
		t.Errorf("limit = %f, want 200.0", got.Limit)
	}
	if got.Currency != "EUR" {
		t.Errorf("currency = %q, want %q", got.Currency, "EUR")
	}
	if got.Used != 25.0 {
		t.Errorf("used = %f, want 25.0", got.Used)
	}
}

func TestInteg_AddUsedAmount_Increment(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Microsecond)
	budget := &models.Budget{
		TenantID:    "test-tenant",
		AgentID:     "agent-increment",
		Currency:    "USD",
		Limit:       100.0,
		Used:        0,
		PeriodStart: now,
		PeriodEnd:   now.Add(30 * 24 * time.Hour),
	}
	if err := repo.UpsertBudget(ctx, budget); err != nil {
		t.Fatalf("UpsertBudget: %v", err)
	}

	if err := repo.AddUsedAmount(ctx, "test-tenant", "agent-increment", 25.0); err != nil {
		t.Fatalf("AddUsedAmount first: %v", err)
	}
	if err := repo.AddUsedAmount(ctx, "test-tenant", "agent-increment", 10.0); err != nil {
		t.Fatalf("AddUsedAmount second: %v", err)
	}

	got, err := repo.GetBudget(ctx, "test-tenant", "agent-increment")
	if err != nil {
		t.Fatalf("GetBudget: %v", err)
	}
	if got.Used != 35.0 {
		t.Errorf("used = %f, want 35.0", got.Used)
	}
}

func TestInteg_AddUsedAmount_NoBudget(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	err := repo.AddUsedAmount(ctx, "test-tenant", "nonexistent-agent", 10.0)
	if err != ErrBudgetNotFound {
		t.Errorf("error = %v, want ErrBudgetNotFound", err)
	}
}

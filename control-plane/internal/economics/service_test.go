package economics

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// mockRepo is a hand-written in-memory Repository for testing.
type mockRepo struct {
	usageRecords []*models.UsageRecord
	budgets      map[string]*models.Budget
	nextID       int
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		budgets: make(map[string]*models.Budget),
	}
}

func (m *mockRepo) nextUUID() string {
	m.nextID++
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", m.nextID)
}

func (m *mockRepo) InsertUsage(_ context.Context, record *models.UsageRecord) error {
	record.ID = m.nextUUID()
	record.RecordedAt = time.Now()
	cp := *record
	m.usageRecords = append(m.usageRecords, &cp)
	return nil
}

func (m *mockRepo) GetBudget(_ context.Context, agentID string) (*models.Budget, error) {
	b, ok := m.budgets[agentID]
	if !ok {
		return nil, nil
	}
	cp := *b
	return &cp, nil
}

func (m *mockRepo) UpsertBudget(_ context.Context, budget *models.Budget) error {
	if budget.ID == "" {
		budget.ID = m.nextUUID()
	}
	cp := *budget
	m.budgets[budget.AgentID] = &cp
	return nil
}

func (m *mockRepo) AddUsedAmount(_ context.Context, agentID string, amount float64) error {
	b, ok := m.budgets[agentID]
	if !ok {
		return ErrBudgetNotFound
	}
	b.Used += amount
	return nil
}

// --- RecordUsage tests ---

func TestRecordUsage_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	record, err := svc.RecordUsage(context.Background(), "agent-1", "ws-1", "compute", "seconds", 100, 0.50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.ID == "" {
		t.Error("expected record ID to be set")
	}
	if record.AgentID != "agent-1" {
		t.Errorf("expected agentID 'agent-1', got %q", record.AgentID)
	}
}

func TestRecordUsage_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.RecordUsage(ctx, "", "ws-1", "compute", "sec", 1, 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty agentID, got: %v", err)
	}
	if _, err := svc.RecordUsage(ctx, "a", "ws", "", "sec", 1, 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty resourceType, got: %v", err)
	}
	if _, err := svc.RecordUsage(ctx, "a", "ws", "compute", "", 1, 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty unit, got: %v", err)
	}
	if _, err := svc.RecordUsage(ctx, "a", "ws", "compute", "sec", 0, 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero quantity, got: %v", err)
	}
	if _, err := svc.RecordUsage(ctx, "a", "ws", "compute", "sec", 1, -1); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for negative cost, got: %v", err)
	}
}

func TestRecordUsage_UpdatesBudget(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Set up a budget first.
	_, err := svc.SetBudget(ctx, "agent-1", 100, "USD")
	if err != nil {
		t.Fatalf("failed to set budget: %v", err)
	}

	// Record usage with cost.
	_, err = svc.RecordUsage(ctx, "agent-1", "ws-1", "compute", "sec", 10, 5.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	budget, _ := svc.GetBudget(ctx, "agent-1")
	if budget.Used != 5.0 {
		t.Errorf("expected budget.Used=5.0, got %f", budget.Used)
	}
}

// --- GetBudget tests ---

func TestGetBudget_Found(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "agent-1", 100, "USD")

	budget, err := svc.GetBudget(ctx, "agent-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if budget.AgentID != "agent-1" {
		t.Errorf("expected agentID 'agent-1', got %q", budget.AgentID)
	}
	if budget.Limit != 100 {
		t.Errorf("expected limit 100, got %f", budget.Limit)
	}
}

func TestGetBudget_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetBudget(context.Background(), "nonexistent")
	if !errors.Is(err, ErrBudgetNotFound) {
		t.Errorf("expected ErrBudgetNotFound, got: %v", err)
	}
}

func TestGetBudget_EmptyID(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetBudget(context.Background(), "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty ID, got: %v", err)
	}
}

// --- SetBudget tests ---

func TestSetBudget_Create(t *testing.T) {
	svc := NewService(newMockRepo())
	budget, err := svc.SetBudget(context.Background(), "agent-1", 500, "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if budget.ID == "" {
		t.Error("expected budget ID to be set")
	}
	if budget.Limit != 500 {
		t.Errorf("expected limit 500, got %f", budget.Limit)
	}
	if budget.Currency != "USD" {
		t.Errorf("expected currency 'USD', got %q", budget.Currency)
	}
	if budget.Used != 0 {
		t.Errorf("expected used 0, got %f", budget.Used)
	}
}

func TestSetBudget_Update(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "agent-1", 100, "USD")

	// Simulate some usage.
	repo.budgets["agent-1"].Used = 25

	// Update budget — should preserve Used.
	budget, err := svc.SetBudget(ctx, "agent-1", 200, "EUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if budget.Limit != 200 {
		t.Errorf("expected limit 200, got %f", budget.Limit)
	}
	if budget.Used != 25 {
		t.Errorf("expected used 25 (preserved), got %f", budget.Used)
	}
}

func TestSetBudget_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.SetBudget(ctx, "", 100, "USD"); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty agentID, got: %v", err)
	}
	if _, err := svc.SetBudget(ctx, "a", 0, "USD"); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero limit, got: %v", err)
	}
	if _, err := svc.SetBudget(ctx, "a", 100, ""); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty currency, got: %v", err)
	}
}

// --- CheckBudget tests ---

func TestCheckBudget_Allowed(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "agent-1", 100, "USD")

	allowed, remaining, err := svc.CheckBudget(ctx, "agent-1", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected allowed=true")
	}
	if remaining != 100 {
		t.Errorf("expected remaining=100, got %f", remaining)
	}
}

func TestCheckBudget_Denied(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "agent-1", 100, "USD")
	repo.budgets["agent-1"].Used = 90

	allowed, remaining, err := svc.CheckBudget(ctx, "agent-1", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected allowed=false")
	}
	if remaining != 10 {
		t.Errorf("expected remaining=10, got %f", remaining)
	}
}

func TestCheckBudget_ExactLimit(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "agent-1", 100, "USD")
	repo.budgets["agent-1"].Used = 50

	allowed, remaining, err := svc.CheckBudget(ctx, "agent-1", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected allowed=true at exact limit")
	}
	if remaining != 50 {
		t.Errorf("expected remaining=50, got %f", remaining)
	}
}

func TestCheckBudget_NoBudget(t *testing.T) {
	svc := NewService(newMockRepo())
	allowed, _, err := svc.CheckBudget(context.Background(), "agent-1", 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected allowed=true when no budget exists")
	}
}

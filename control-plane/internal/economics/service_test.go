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

func (m *mockRepo) GetBudget(_ context.Context, _, agentID string) (*models.Budget, error) {
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

func (m *mockRepo) AddUsedAmount(_ context.Context, _, agentID string, amount float64) error {
	b, ok := m.budgets[agentID]
	if !ok {
		return ErrBudgetNotFound
	}
	b.Used += amount
	return nil
}

func (m *mockRepo) GetCostReport(_ context.Context, _, agentID string, start, end time.Time) ([]ResourceCost, error) {
	costs := make(map[string]*ResourceCost)
	for _, r := range m.usageRecords {
		if agentID != "" && r.AgentID != agentID {
			continue
		}
		if r.RecordedAt.Before(start) || r.RecordedAt.After(end) {
			continue
		}
		c, ok := costs[r.ResourceType]
		if !ok {
			c = &ResourceCost{ResourceType: r.ResourceType}
			costs[r.ResourceType] = c
		}
		c.TotalCost += r.Cost
		c.RecordCount++
	}
	var result []ResourceCost
	for _, c := range costs {
		result = append(result, *c)
	}
	return result, nil
}

// --- RecordUsage tests ---

func TestRecordUsage_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	record, err := svc.RecordUsage(context.Background(), "tenant-1", "agent-1", "ws-1", "compute", "seconds", 100, 0.50)
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

	if _, err := svc.RecordUsage(ctx, "tenant-1", "", "ws-1", "compute", "sec", 1, 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty agentID, got: %v", err)
	}
	if _, err := svc.RecordUsage(ctx, "tenant-1", "a", "ws", "", "sec", 1, 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty resourceType, got: %v", err)
	}
	if _, err := svc.RecordUsage(ctx, "tenant-1", "a", "ws", "compute", "", 1, 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty unit, got: %v", err)
	}
	if _, err := svc.RecordUsage(ctx, "tenant-1", "a", "ws", "compute", "sec", 0, 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero quantity, got: %v", err)
	}
	if _, err := svc.RecordUsage(ctx, "tenant-1", "a", "ws", "compute", "sec", 1, -1); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for negative cost, got: %v", err)
	}
}

func TestRecordUsage_UpdatesBudget(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Set up a budget first.
	_, err := svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "", 0)
	if err != nil {
		t.Fatalf("failed to set budget: %v", err)
	}

	// Record usage with cost.
	_, err = svc.RecordUsage(ctx, "tenant-1", "agent-1", "ws-1", "compute", "sec", 10, 5.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	budget, _ := svc.GetBudget(ctx, "tenant-1", "agent-1")
	if budget.Used != 5.0 {
		t.Errorf("expected budget.Used=5.0, got %f", budget.Used)
	}
}

// --- GetBudget tests ---

func TestGetBudget_Found(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "", 0)

	budget, err := svc.GetBudget(ctx, "tenant-1", "agent-1")
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
	_, err := svc.GetBudget(context.Background(), "tenant-1", "nonexistent")
	if !errors.Is(err, ErrBudgetNotFound) {
		t.Errorf("expected ErrBudgetNotFound, got: %v", err)
	}
}

func TestGetBudget_EmptyID(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetBudget(context.Background(), "tenant-1", "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty ID, got: %v", err)
	}
}

// --- SetBudget tests ---

func TestSetBudget_Create(t *testing.T) {
	svc := NewService(newMockRepo())
	budget, err := svc.SetBudget(context.Background(), "tenant-1", "agent-1", 500, "USD", "", 0)
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

	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "", 0)

	// Simulate some usage.
	repo.budgets["agent-1"].Used = 25

	// Update budget — should preserve Used.
	budget, err := svc.SetBudget(ctx, "tenant-1", "agent-1", 200, "EUR", "", 0)
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

	if _, err := svc.SetBudget(ctx, "tenant-1", "", 100, "USD", "", 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty agentID, got: %v", err)
	}
	if _, err := svc.SetBudget(ctx, "tenant-1", "a", 0, "USD", "", 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero limit, got: %v", err)
	}
	if _, err := svc.SetBudget(ctx, "tenant-1", "a", 100, "", "", 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty currency, got: %v", err)
	}
}

// --- CheckBudget tests ---

func TestCheckBudget_Allowed(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "", 0)

	result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected allowed=true")
	}
	if result.Remaining != 100 {
		t.Errorf("expected remaining=100, got %f", result.Remaining)
	}
}

func TestCheckBudget_Denied(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "", 0)
	repo.budgets["agent-1"].Used = 90

	result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Allowed {
		t.Error("expected allowed=false")
	}
	if result.Remaining != 10 {
		t.Errorf("expected remaining=10, got %f", result.Remaining)
	}
	if result.EnforcementAction != "halt" {
		t.Errorf("expected enforcement_action=halt, got %q", result.EnforcementAction)
	}
}

func TestCheckBudget_ExactLimit(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "", 0)
	repo.budgets["agent-1"].Used = 50

	result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected allowed=true at exact limit")
	}
	if result.Remaining != 50 {
		t.Errorf("expected remaining=50, got %f", result.Remaining)
	}
}

func TestCheckBudget_NoBudget(t *testing.T) {
	svc := NewService(newMockRepo())
	result, err := svc.CheckBudget(context.Background(), "tenant-1", "agent-1", 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected allowed=true when no budget exists")
	}
}

func TestCheckBudget_WarningThreshold(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "halt", 0.2) // warn at 20% remaining
	repo.budgets["agent-1"].Used = 85                                   // 15% remaining

	result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected allowed=true (15 remaining, only need 5)")
	}
	if !result.Warning {
		t.Error("expected warning=true (15 remaining < 20% of 100)")
	}
}

func TestCheckBudget_RequestIncrease(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "request_increase", 0)
	repo.budgets["agent-1"].Used = 100

	result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Allowed {
		t.Error("expected allowed=false")
	}
	if result.EnforcementAction != "request_increase" {
		t.Errorf("expected enforcement_action=request_increase, got %q", result.EnforcementAction)
	}
}

func TestCheckBudget_WarnMode(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "warn", 0)
	repo.budgets["agent-1"].Used = 100

	result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected allowed=true in warn mode")
	}
	if !result.Warning {
		t.Error("expected warning=true in warn mode when over budget")
	}
	if result.EnforcementAction != "warn" {
		t.Errorf("expected enforcement_action=warn, got %q", result.EnforcementAction)
	}
}

// --- GetCostReport tests ---

func TestGetCostReport_WithData(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	now := time.Now().UTC()
	// Insert some usage records.
	svc.RecordUsage(ctx, "tenant-1", "agent-1", "ws-1", "compute", "seconds", 100, 5.0)
	svc.RecordUsage(ctx, "tenant-1", "agent-1", "ws-1", "compute", "seconds", 200, 10.0)
	svc.RecordUsage(ctx, "tenant-1", "agent-1", "ws-1", "storage", "bytes", 500, 2.0)

	report, err := svc.GetCostReport(ctx, "tenant-1", "agent-1", now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalCost != 17.0 {
		t.Errorf("expected total cost 17.0, got %f", report.TotalCost)
	}
	if report.RecordCount != 3 {
		t.Errorf("expected record count 3, got %d", report.RecordCount)
	}
	if len(report.ByResourceType) != 2 {
		t.Errorf("expected 2 resource types, got %d", len(report.ByResourceType))
	}
}

func TestGetCostReport_Empty(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	now := time.Now().UTC()

	report, err := svc.GetCostReport(ctx, "tenant-1", "agent-1", now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalCost != 0 {
		t.Errorf("expected total cost 0, got %f", report.TotalCost)
	}
	if report.RecordCount != 0 {
		t.Errorf("expected record count 0, got %d", report.RecordCount)
	}
}

func TestGetCostReport_InvalidTimeRange(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	now := time.Now().UTC()

	_, err := svc.GetCostReport(ctx, "tenant-1", "", now.Add(time.Hour), now.Add(-time.Hour))
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for end before start, got: %v", err)
	}

	_, err = svc.GetCostReport(ctx, "tenant-1", "", now, now)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for equal start and end, got: %v", err)
	}
}

func TestGetCostReport_FiltersByAgent(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	now := time.Now().UTC()
	svc.RecordUsage(ctx, "tenant-1", "agent-1", "ws-1", "compute", "seconds", 100, 5.0)
	svc.RecordUsage(ctx, "tenant-1", "agent-2", "ws-2", "compute", "seconds", 200, 10.0)

	report, err := svc.GetCostReport(ctx, "tenant-1", "agent-1", now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalCost != 5.0 {
		t.Errorf("expected total cost 5.0 (agent-1 only), got %f", report.TotalCost)
	}
	if report.RecordCount != 1 {
		t.Errorf("expected record count 1, got %d", report.RecordCount)
	}
}

func TestGetCostReport_AllAgents(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	now := time.Now().UTC()
	svc.RecordUsage(ctx, "tenant-1", "agent-1", "ws-1", "compute", "seconds", 100, 5.0)
	svc.RecordUsage(ctx, "tenant-1", "agent-2", "ws-2", "compute", "seconds", 200, 10.0)

	report, err := svc.GetCostReport(ctx, "tenant-1", "", now.Add(-time.Minute), now.Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalCost != 15.0 {
		t.Errorf("expected total cost 15.0 (all agents), got %f", report.TotalCost)
	}
	if report.RecordCount != 2 {
		t.Errorf("expected record count 2, got %d", report.RecordCount)
	}
}

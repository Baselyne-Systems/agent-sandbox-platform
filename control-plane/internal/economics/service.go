package economics

import (
	"context"
	"errors"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

var (
	ErrBudgetNotFound = errors.New("budget not found")
	ErrInvalidInput   = errors.New("invalid input")
)

// Service implements economics business logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// RecordUsage validates and persists a usage record, then attempts to
// increment the agent's budget.Used amount (non-fatal if no budget exists).
func (s *Service) RecordUsage(ctx context.Context, agentID, workspaceID, resourceType, unit string, quantity, cost float64) (*models.UsageRecord, error) {
	if agentID == "" || resourceType == "" || unit == "" {
		return nil, ErrInvalidInput
	}
	if quantity <= 0 || cost < 0 {
		return nil, ErrInvalidInput
	}

	record := &models.UsageRecord{
		AgentID:      agentID,
		WorkspaceID:  workspaceID,
		ResourceType: resourceType,
		Unit:         unit,
		Quantity:     quantity,
		Cost:         cost,
	}
	if err := s.repo.InsertUsage(ctx, record); err != nil {
		return nil, err
	}

	// Best-effort: update budget.Used. Non-fatal if agent has no budget.
	if cost > 0 {
		err := s.repo.AddUsedAmount(ctx, agentID, cost)
		if err != nil && !errors.Is(err, ErrBudgetNotFound) {
			return nil, err
		}
	}

	return record, nil
}

// GetBudget returns the budget for the given agent.
func (s *Service) GetBudget(ctx context.Context, agentID string) (*models.Budget, error) {
	if agentID == "" {
		return nil, ErrInvalidInput
	}
	budget, err := s.repo.GetBudget(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if budget == nil {
		return nil, ErrBudgetNotFound
	}
	return budget, nil
}

// SetBudget creates or updates a budget for the given agent with a 30-day period.
func (s *Service) SetBudget(ctx context.Context, agentID string, limit float64, currency, onExceeded string, warningThreshold float64) (*models.Budget, error) {
	if agentID == "" || currency == "" {
		return nil, ErrInvalidInput
	}
	if limit <= 0 {
		return nil, ErrInvalidInput
	}
	if warningThreshold < 0 || warningThreshold > 1 {
		return nil, ErrInvalidInput
	}

	// Default on_exceeded to "halt" if not specified.
	if onExceeded == "" {
		onExceeded = "halt"
	}
	switch onExceeded {
	case "halt", "request_increase", "warn":
		// valid
	default:
		return nil, ErrInvalidInput
	}

	now := time.Now().UTC()
	budget := &models.Budget{
		AgentID:          agentID,
		Currency:         currency,
		Limit:            limit,
		Used:             0,
		PeriodStart:      now,
		PeriodEnd:        now.AddDate(0, 0, 30),
		OnExceeded:       onExceeded,
		WarningThreshold: warningThreshold,
	}

	// Check if budget already exists — preserve Used amount on update.
	existing, err := s.repo.GetBudget(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		budget.Used = existing.Used
	}

	if err := s.repo.UpsertBudget(ctx, budget); err != nil {
		return nil, err
	}
	return budget, nil
}

// CostReport holds the result of a cost report query.
type CostReport struct {
	TotalCost  float64
	RecordCount int
	ByResourceType []ResourceCost
}

// GetCostReport returns aggregated cost data for the given time window,
// optionally filtered by agent.
func (s *Service) GetCostReport(ctx context.Context, agentID string, start, end time.Time) (*CostReport, error) {
	if !start.Before(end) {
		return nil, ErrInvalidInput
	}

	costs, err := s.repo.GetCostReport(ctx, agentID, start, end)
	if err != nil {
		return nil, err
	}

	report := &CostReport{ByResourceType: costs}
	for _, c := range costs {
		report.TotalCost += c.TotalCost
		report.RecordCount += c.RecordCount
	}
	return report, nil
}

// BudgetCheckResult holds the full result of a budget check.
type BudgetCheckResult struct {
	Allowed           bool
	Remaining         float64
	EnforcementAction string // "halt", "request_increase", or "" when allowed
	Warning           bool   // true when remaining < warning_threshold * limit
}

// CheckBudget reads the agent's budget and returns whether the estimated
// cost fits within the remaining headroom, along with enforcement details.
func (s *Service) CheckBudget(ctx context.Context, agentID string, estimatedCost float64) (*BudgetCheckResult, error) {
	if agentID == "" {
		return nil, ErrInvalidInput
	}

	budget, err := s.repo.GetBudget(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if budget == nil {
		// No budget means no constraint — allow.
		return &BudgetCheckResult{Allowed: true}, nil
	}

	remaining := budget.Limit - budget.Used
	result := &BudgetCheckResult{
		Allowed:   remaining >= estimatedCost,
		Remaining: remaining,
	}

	// Check warning threshold.
	if budget.WarningThreshold > 0 && remaining < budget.WarningThreshold*budget.Limit {
		result.Warning = true
	}

	// If not allowed, determine enforcement action.
	if !result.Allowed {
		switch budget.OnExceeded {
		case "request_increase":
			result.EnforcementAction = "request_increase"
		case "warn":
			// Warn mode: allow the call but flag it.
			result.Allowed = true
			result.Warning = true
			result.EnforcementAction = "warn"
		default:
			// "halt" or unset — hard deny.
			result.EnforcementAction = "halt"
		}
	}

	return result, nil
}

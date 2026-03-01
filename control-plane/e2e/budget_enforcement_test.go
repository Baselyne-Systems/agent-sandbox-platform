package e2e

import (
	"context"
	"testing"
)

func TestBudgetFullCycle(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "budget-agent")

	budget, err := economicsSvc.SetBudget(ctx, tenant, agent.ID, 100.0, "USD", "halt", 0.0)
	if err != nil {
		t.Fatalf("set budget: %v", err)
	}
	if budget.Limit != 100.0 {
		t.Fatalf("expected limit 100, got %f", budget.Limit)
	}

	_, err = economicsSvc.RecordUsage(ctx, tenant, agent.ID, "ws-1", "compute", "hour", 1.0, 30.0)
	if err != nil {
		t.Fatalf("record usage 30: %v", err)
	}

	result, err := economicsSvc.CheckBudget(ctx, tenant, agent.ID, 10.0)
	if err != nil {
		t.Fatalf("check budget 10: %v", err)
	}
	if !result.Allowed {
		t.Fatal("expected allowed for $10 estimate")
	}

	_, err = economicsSvc.RecordUsage(ctx, tenant, agent.ID, "ws-1", "compute", "hour", 2.0, 50.0)
	if err != nil {
		t.Fatalf("record usage 50: %v", err)
	}

	result, err = economicsSvc.CheckBudget(ctx, tenant, agent.ID, 30.0)
	if err != nil {
		t.Fatalf("check budget 30: %v", err)
	}
	if result.Allowed {
		t.Fatal("expected denied for $30 estimate (would exceed $100 budget)")
	}
	if result.EnforcementAction != "halt" {
		t.Fatalf("expected enforcement 'halt', got %q", result.EnforcementAction)
	}
}

func TestBudgetWarningThreshold(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "warn-agent")

	_, err := economicsSvc.SetBudget(ctx, tenant, agent.ID, 100.0, "USD", "halt", 0.8)
	if err != nil {
		t.Fatalf("set budget: %v", err)
	}

	_, err = economicsSvc.RecordUsage(ctx, tenant, agent.ID, "ws-1", "compute", "hour", 3.0, 85.0)
	if err != nil {
		t.Fatalf("record usage: %v", err)
	}

	result, err := economicsSvc.CheckBudget(ctx, tenant, agent.ID, 5.0)
	if err != nil {
		t.Fatalf("check budget: %v", err)
	}
	if !result.Allowed {
		t.Fatal("expected allowed (85+5=90 < 100)")
	}
	if !result.Warning {
		t.Fatal("expected warning (85 > 80% threshold)")
	}
}

func TestBudgetOnExceededModes(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	tests := []struct {
		name         string
		onExceeded   string
		expectAllow  bool
		expectWarn   bool
		expectAction string
	}{
		{"halt", "halt", false, false, "halt"},
		{"warn", "warn", true, true, "warn"},
		{"request_increase", "request_increase", false, false, "request_increase"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clean(t)
			agent := registerAgent(t, ctx, tenant, "mode-"+tt.name)

			_, err := economicsSvc.SetBudget(ctx, tenant, agent.ID, 100.0, "USD", tt.onExceeded, 0.0)
			if err != nil {
				t.Fatalf("set budget: %v", err)
			}

			_, err = economicsSvc.RecordUsage(ctx, tenant, agent.ID, "ws-1", "compute", "hour", 5.0, 95.0)
			if err != nil {
				t.Fatalf("record usage: %v", err)
			}

			result, err := economicsSvc.CheckBudget(ctx, tenant, agent.ID, 20.0)
			if err != nil {
				t.Fatalf("check: %v", err)
			}
			if result.Allowed != tt.expectAllow {
				t.Fatalf("expected allowed=%v, got %v", tt.expectAllow, result.Allowed)
			}
			if result.Warning != tt.expectWarn {
				t.Fatalf("expected warning=%v, got %v", tt.expectWarn, result.Warning)
			}
			if result.EnforcementAction != tt.expectAction {
				t.Fatalf("expected enforcement %q, got %q", tt.expectAction, result.EnforcementAction)
			}
		})
	}
}

func TestBudgetPreservesUsedOnUpdate(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "preserve-agent")

	_, err := economicsSvc.SetBudget(ctx, tenant, agent.ID, 100.0, "USD", "halt", 0.0)
	if err != nil {
		t.Fatalf("set budget: %v", err)
	}

	_, err = economicsSvc.RecordUsage(ctx, tenant, agent.ID, "ws-1", "compute", "hour", 2.0, 40.0)
	if err != nil {
		t.Fatalf("record usage: %v", err)
	}

	_, err = economicsSvc.SetBudget(ctx, tenant, agent.ID, 200.0, "USD", "halt", 0.0)
	if err != nil {
		t.Fatalf("update budget: %v", err)
	}

	budget, err := economicsSvc.GetBudget(ctx, tenant, agent.ID)
	if err != nil {
		t.Fatalf("get budget: %v", err)
	}
	if budget.Limit != 200.0 {
		t.Fatalf("expected limit 200, got %f", budget.Limit)
	}
	if budget.Used < 39.0 || budget.Used > 41.0 {
		t.Fatalf("expected used ~40, got %f", budget.Used)
	}
}

func TestNoBudgetAllowsAll(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "no-budget-agent")

	result, err := economicsSvc.CheckBudget(ctx, tenant, agent.ID, 999999.0)
	if err != nil {
		t.Fatalf("check budget: %v", err)
	}
	if !result.Allowed {
		t.Fatal("expected allowed when no budget is set")
	}
}

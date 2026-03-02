package economics

import (
	"context"
	"testing"

	pb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/economics/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestHandler() *Handler {
	return NewHandler(NewService(newMockRepo()))
}

func TestHandler_RecordUsage_Success(t *testing.T) {
	h := newTestHandler()
	resp, err := h.RecordUsage(context.Background(), &pb.RecordUsageRequest{
		Record: &pb.UsageRecord{
			AgentId:      "agent-1",
			WorkspaceId:  "ws-1",
			ResourceType: "compute",
			Unit:         "cpu_seconds",
			Quantity:     120.5,
			Cost:         0.05,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RecordId == "" {
		t.Error("expected record ID")
	}
}

func TestHandler_RecordUsage_NilRecord(t *testing.T) {
	h := newTestHandler()
	_, err := h.RecordUsage(context.Background(), &pb.RecordUsageRequest{
		Record: nil,
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_RecordUsage_InvalidInput(t *testing.T) {
	h := newTestHandler()
	_, err := h.RecordUsage(context.Background(), &pb.RecordUsageRequest{
		Record: &pb.UsageRecord{
			AgentId:      "",
			ResourceType: "compute",
			Unit:         "cpu_seconds",
			Quantity:     1,
			Cost:         0.01,
		},
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_SetBudget_Success(t *testing.T) {
	h := newTestHandler()
	resp, err := h.SetBudget(context.Background(), &pb.SetBudgetRequest{
		AgentId:  "agent-1",
		Limit:    100.0,
		Currency: "USD",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Budget == nil {
		t.Fatal("expected budget in response")
	}
	if resp.Budget.AgentId != "agent-1" {
		t.Errorf("agent_id = %q, want 'agent-1'", resp.Budget.AgentId)
	}
	if resp.Budget.Limit != 100.0 {
		t.Errorf("limit = %f, want 100.0", resp.Budget.Limit)
	}
	if resp.Budget.Currency != "USD" {
		t.Errorf("currency = %q, want 'USD'", resp.Budget.Currency)
	}
	if resp.Budget.Used != 0 {
		t.Errorf("used = %f, want 0", resp.Budget.Used)
	}
	if resp.Budget.PeriodStart == nil {
		t.Error("expected period_start")
	}
	if resp.Budget.PeriodEnd == nil {
		t.Error("expected period_end")
	}
}

func TestHandler_SetBudget_InvalidInput(t *testing.T) {
	h := newTestHandler()
	_, err := h.SetBudget(context.Background(), &pb.SetBudgetRequest{
		AgentId:  "",
		Limit:    100.0,
		Currency: "USD",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_GetBudget_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	h.SetBudget(ctx, &pb.SetBudgetRequest{
		AgentId: "agent-1", Limit: 50.0, Currency: "USD",
	})

	resp, err := h.GetBudget(ctx, &pb.GetBudgetRequest{AgentId: "agent-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Budget.Limit != 50.0 {
		t.Errorf("limit = %f, want 50.0", resp.Budget.Limit)
	}
}

func TestHandler_GetBudget_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.GetBudget(context.Background(), &pb.GetBudgetRequest{
		AgentId: "nonexistent",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_CheckBudget_Allowed(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	h.SetBudget(ctx, &pb.SetBudgetRequest{
		AgentId: "agent-1", Limit: 100.0, Currency: "USD",
	})

	resp, err := h.CheckBudget(ctx, &pb.CheckBudgetRequest{
		AgentId:       "agent-1",
		EstimatedCost: 50.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Allowed {
		t.Error("expected allowed = true")
	}
	if resp.Remaining != 100.0 {
		t.Errorf("remaining = %f, want 100.0", resp.Remaining)
	}
}

func TestHandler_CheckBudget_Denied(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	h.SetBudget(ctx, &pb.SetBudgetRequest{
		AgentId: "agent-1", Limit: 10.0, Currency: "USD",
	})

	resp, err := h.CheckBudget(ctx, &pb.CheckBudgetRequest{
		AgentId:       "agent-1",
		EstimatedCost: 50.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Allowed {
		t.Error("expected allowed = false")
	}
}

func TestHandler_CheckBudget_NoBudget(t *testing.T) {
	h := newTestHandler()
	resp, err := h.CheckBudget(context.Background(), &pb.CheckBudgetRequest{
		AgentId:       "no-budget-agent",
		EstimatedCost: 1000.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No budget means no constraint — allow
	if !resp.Allowed {
		t.Error("expected allowed = true when no budget exists")
	}
}

func TestHandler_UsageThenBudgetCheck(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	h.SetBudget(ctx, &pb.SetBudgetRequest{
		AgentId: "agent-1", Limit: 10.0, Currency: "USD",
	})

	// Record usage that consumes 8.0
	h.RecordUsage(ctx, &pb.RecordUsageRequest{
		Record: &pb.UsageRecord{
			AgentId:      "agent-1",
			ResourceType: "compute",
			Unit:         "cpu_seconds",
			Quantity:     100,
			Cost:         8.0,
		},
	})

	// Budget should have 2.0 remaining
	resp, err := h.CheckBudget(ctx, &pb.CheckBudgetRequest{
		AgentId:       "agent-1",
		EstimatedCost: 5.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Allowed {
		t.Error("expected denied: estimated 5.0 > remaining 2.0")
	}
	if resp.Remaining != 2.0 {
		t.Errorf("remaining = %f, want 2.0", resp.Remaining)
	}
}

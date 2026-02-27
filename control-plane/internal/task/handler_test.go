package task

import (
	"context"
	"testing"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/task/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestHandler() *Handler {
	return NewHandler(NewService(newMockRepo()))
}

func createTestTask(t *testing.T, h *Handler) *pb.Task {
	t.Helper()
	resp, err := h.CreateTask(context.Background(), &pb.CreateTaskRequest{
		AgentId:           "agent-1",
		Goal:              "Process invoices",
		GuardrailPolicyId: "policy-1",
		WorkspaceConfig: &pb.TaskWorkspaceConfig{
			IsolationTier: "standard",
			MemoryMb:      512,
			CpuMillicores: 1000,
		},
		HumanInteraction: &pb.HumanInteractionConfig{
			EscalationTargets: []string{"admin@example.com"},
			TimeoutSecs:       600,
			TimeoutAction:     "escalate",
		},
		Budget: &pb.BudgetConfig{
			MaxCost:  10.0,
			Currency: "USD",
		},
		MaxDurationWithoutCheckinSecs: 300,
		Input:                         map[string]string{"file": "invoice.pdf"},
		Labels:                        map[string]string{"env": "test"},
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	return resp.Task
}

func TestHandler_CreateTask_Success(t *testing.T) {
	h := newTestHandler()
	task := createTestTask(t, h)

	if task.TaskId == "" {
		t.Error("expected task ID")
	}
	if task.AgentId != "agent-1" {
		t.Errorf("agent_id = %q, want 'agent-1'", task.AgentId)
	}
	if task.Goal != "Process invoices" {
		t.Errorf("goal = %q, want 'Process invoices'", task.Goal)
	}
	if task.Status != pb.TaskStatus_TASK_STATUS_PENDING {
		t.Errorf("status = %v, want PENDING", task.Status)
	}
	if task.GuardrailPolicyId != "policy-1" {
		t.Errorf("guardrail_policy_id = %q, want 'policy-1'", task.GuardrailPolicyId)
	}
	if task.WorkspaceConfig == nil {
		t.Fatal("expected workspace_config")
	}
	if task.WorkspaceConfig.IsolationTier != "standard" {
		t.Errorf("isolation_tier = %q, want 'standard'", task.WorkspaceConfig.IsolationTier)
	}
	if task.WorkspaceConfig.MemoryMb != 512 {
		t.Errorf("memory_mb = %d, want 512", task.WorkspaceConfig.MemoryMb)
	}
	if task.HumanInteraction == nil {
		t.Fatal("expected human_interaction")
	}
	if task.HumanInteraction.TimeoutSecs != 600 {
		t.Errorf("timeout_secs = %d, want 600", task.HumanInteraction.TimeoutSecs)
	}
	if task.Budget == nil {
		t.Fatal("expected budget")
	}
	if task.Budget.MaxCost != 10.0 {
		t.Errorf("max_cost = %f, want 10.0", task.Budget.MaxCost)
	}
	if task.MaxDurationWithoutCheckinSecs != 300 {
		t.Errorf("max_duration = %d, want 300", task.MaxDurationWithoutCheckinSecs)
	}
	if task.Input["file"] != "invoice.pdf" {
		t.Errorf("input[file] = %q, want 'invoice.pdf'", task.Input["file"])
	}
	if task.Labels["env"] != "test" {
		t.Errorf("labels[env] = %q, want 'test'", task.Labels["env"])
	}
	if task.CreatedAt == nil {
		t.Error("expected created_at timestamp")
	}
}

func TestHandler_CreateTask_Minimal(t *testing.T) {
	h := newTestHandler()
	resp, err := h.CreateTask(context.Background(), &pb.CreateTaskRequest{
		AgentId: "agent-1",
		Goal:    "Do something",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Task.TaskId == "" {
		t.Error("expected task ID")
	}
	if resp.Task.Status != pb.TaskStatus_TASK_STATUS_PENDING {
		t.Errorf("status = %v, want PENDING", resp.Task.Status)
	}
}

func TestHandler_CreateTask_InvalidInput(t *testing.T) {
	h := newTestHandler()
	tests := []struct {
		name string
		req  *pb.CreateTaskRequest
	}{
		{"empty agent_id", &pb.CreateTaskRequest{AgentId: "", Goal: "g"}},
		{"empty goal", &pb.CreateTaskRequest{AgentId: "a", Goal: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := h.CreateTask(context.Background(), tt.req)
			st, _ := status.FromError(err)
			if st.Code() != codes.InvalidArgument {
				t.Errorf("code = %v, want InvalidArgument", st.Code())
			}
		})
	}
}

func TestHandler_GetTask_Success(t *testing.T) {
	h := newTestHandler()
	created := createTestTask(t, h)

	resp, err := h.GetTask(context.Background(), &pb.GetTaskRequest{
		TaskId: created.TaskId,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Task.TaskId != created.TaskId {
		t.Errorf("ID mismatch: got %q, want %q", resp.Task.TaskId, created.TaskId)
	}
	if resp.Task.Goal != "Process invoices" {
		t.Errorf("goal = %q, want 'Process invoices'", resp.Task.Goal)
	}
}

func TestHandler_GetTask_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.GetTask(context.Background(), &pb.GetTaskRequest{
		TaskId: "nonexistent",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_ListTasks_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		h.CreateTask(ctx, &pb.CreateTaskRequest{
			AgentId: "agent-1", Goal: "goal",
		})
	}

	resp, err := h.ListTasks(ctx, &pb.ListTasksRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Tasks) != 3 {
		t.Errorf("tasks count = %d, want 3", len(resp.Tasks))
	}
}

func TestHandler_ListTasks_AgentFilter(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	h.CreateTask(ctx, &pb.CreateTaskRequest{AgentId: "agent-1", Goal: "g1"})
	h.CreateTask(ctx, &pb.CreateTaskRequest{AgentId: "agent-2", Goal: "g2"})
	h.CreateTask(ctx, &pb.CreateTaskRequest{AgentId: "agent-1", Goal: "g3"})

	resp, err := h.ListTasks(ctx, &pb.ListTasksRequest{
		AgentId:  "agent-1",
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Tasks) != 2 {
		t.Errorf("agent-1 tasks = %d, want 2", len(resp.Tasks))
	}
}

func TestHandler_UpdateTaskStatus_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	created := createTestTask(t, h)

	resp, err := h.UpdateTaskStatus(ctx, &pb.UpdateTaskStatusRequest{
		TaskId: created.TaskId,
		Status: pb.TaskStatus_TASK_STATUS_RUNNING,
		Reason: "starting execution",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Task.Status != pb.TaskStatus_TASK_STATUS_RUNNING {
		t.Errorf("status = %v, want RUNNING", resp.Task.Status)
	}
}

func TestHandler_UpdateTaskStatus_InvalidTransition(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	created := createTestTask(t, h)

	// pending → completed is not allowed
	_, err := h.UpdateTaskStatus(ctx, &pb.UpdateTaskStatusRequest{
		TaskId: created.TaskId,
		Status: pb.TaskStatus_TASK_STATUS_COMPLETED,
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", st.Code())
	}
}

func TestHandler_UpdateTaskStatus_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.UpdateTaskStatus(context.Background(), &pb.UpdateTaskStatusRequest{
		TaskId: "nonexistent",
		Status: pb.TaskStatus_TASK_STATUS_RUNNING,
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_CancelTask_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	created := createTestTask(t, h)
	_, err := h.CancelTask(ctx, &pb.CancelTaskRequest{
		TaskId: created.TaskId,
		Reason: "no longer needed",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := h.GetTask(ctx, &pb.GetTaskRequest{TaskId: created.TaskId})
	if got.Task.Status != pb.TaskStatus_TASK_STATUS_CANCELLED {
		t.Errorf("status = %v, want CANCELLED", got.Task.Status)
	}
}

func TestHandler_CancelTask_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.CancelTask(context.Background(), &pb.CancelTaskRequest{
		TaskId: "nonexistent",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_CancelTask_InvalidState(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	created := createTestTask(t, h)
	// Move to completed
	h.UpdateTaskStatus(ctx, &pb.UpdateTaskStatusRequest{
		TaskId: created.TaskId, Status: pb.TaskStatus_TASK_STATUS_RUNNING,
	})
	h.UpdateTaskStatus(ctx, &pb.UpdateTaskStatusRequest{
		TaskId: created.TaskId, Status: pb.TaskStatus_TASK_STATUS_COMPLETED,
	})

	_, err := h.CancelTask(ctx, &pb.CancelTaskRequest{
		TaskId: created.TaskId, Reason: "too late",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", st.Code())
	}
}

func TestHandler_TaskStatusConversion_RoundTrip(t *testing.T) {
	tests := []struct {
		proto pb.TaskStatus
		model models.TaskStatus
	}{
		{pb.TaskStatus_TASK_STATUS_PENDING, models.TaskStatusPending},
		{pb.TaskStatus_TASK_STATUS_RUNNING, models.TaskStatusRunning},
		{pb.TaskStatus_TASK_STATUS_WAITING_ON_HUMAN, models.TaskStatusWaitingOnHuman},
		{pb.TaskStatus_TASK_STATUS_COMPLETED, models.TaskStatusCompleted},
		{pb.TaskStatus_TASK_STATUS_FAILED, models.TaskStatusFailed},
		{pb.TaskStatus_TASK_STATUS_CANCELLED, models.TaskStatusCancelled},
	}
	for _, tt := range tests {
		got := protoTaskStatusToModel(tt.proto)
		if got != tt.model {
			t.Errorf("protoTaskStatusToModel(%v) = %q, want %q", tt.proto, got, tt.model)
		}
		back := modelTaskStatusToProto(got)
		if back != tt.proto {
			t.Errorf("modelTaskStatusToProto(%q) = %v, want %v", got, back, tt.proto)
		}
	}
}

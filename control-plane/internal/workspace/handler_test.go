package workspace

import (
	"context"
	"testing"
	"time"

	pb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/workspace/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

func newTestHandler() *Handler {
	return NewHandler(newTestService(newMockRepo()))
}

func TestHandler_CreateWorkspace_Success(t *testing.T) {
	h := newTestHandler()
	resp, err := h.CreateWorkspace(context.Background(), &pb.CreateWorkspaceRequest{
		AgentId: "agent-1",
		TaskId:  "task-1",
		Spec: &pb.WorkspaceSpec{
			MemoryMb:      2048,
			CpuMillicores: 1000,
			DiskMb:        4096,
			MaxDuration:   durationpb.New(2 * time.Hour),
			AllowedTools:  []string{"shell", "http"},
			EnvVars:       map[string]string{"K": "V"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ws := resp.Workspace
	if ws == nil {
		t.Fatal("expected workspace in response")
	}
	if ws.WorkspaceId == "" {
		t.Error("expected workspace ID")
	}
	if ws.AgentId != "agent-1" {
		t.Errorf("agent_id = %q, want 'agent-1'", ws.AgentId)
	}
	if ws.TaskId != "task-1" {
		t.Errorf("task_id = %q, want 'task-1'", ws.TaskId)
	}
	if ws.Status != pb.WorkspaceStatus_WORKSPACE_STATUS_PENDING {
		t.Errorf("status = %v, want PENDING", ws.Status)
	}
	if ws.Spec.MemoryMb != 2048 {
		t.Errorf("memory_mb = %d, want 2048", ws.Spec.MemoryMb)
	}
	if ws.Spec.CpuMillicores != 1000 {
		t.Errorf("cpu_millicores = %d, want 1000", ws.Spec.CpuMillicores)
	}
	if ws.Spec.AllowedTools[0] != "shell" {
		t.Errorf("allowed_tools[0] = %q, want 'shell'", ws.Spec.AllowedTools[0])
	}
	if ws.Spec.EnvVars["K"] != "V" {
		t.Errorf("env_vars[K] = %q, want 'V'", ws.Spec.EnvVars["K"])
	}
	if ws.ExpiresAt == nil {
		t.Error("expected expires_at timestamp")
	}
	if ws.CreatedAt == nil {
		t.Error("expected created_at timestamp")
	}
}

func TestHandler_CreateWorkspace_DefaultSpec(t *testing.T) {
	h := newTestHandler()
	resp, err := h.CreateWorkspace(context.Background(), &pb.CreateWorkspaceRequest{
		AgentId: "agent-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ws := resp.Workspace
	if ws.Spec.MemoryMb != 512 {
		t.Errorf("default memory_mb = %d, want 512", ws.Spec.MemoryMb)
	}
	if ws.Spec.CpuMillicores != 500 {
		t.Errorf("default cpu_millicores = %d, want 500", ws.Spec.CpuMillicores)
	}
	if ws.Spec.DiskMb != 1024 {
		t.Errorf("default disk_mb = %d, want 1024", ws.Spec.DiskMb)
	}
}

func TestHandler_CreateWorkspace_InvalidInput(t *testing.T) {
	h := newTestHandler()
	_, err := h.CreateWorkspace(context.Background(), &pb.CreateWorkspaceRequest{
		AgentId: "",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_GetWorkspace_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()
	created, _ := h.CreateWorkspace(ctx, &pb.CreateWorkspaceRequest{AgentId: "a"})

	resp, err := h.GetWorkspace(ctx, &pb.GetWorkspaceRequest{
		WorkspaceId: created.Workspace.WorkspaceId,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Workspace.WorkspaceId != created.Workspace.WorkspaceId {
		t.Errorf("ID mismatch")
	}
}

func TestHandler_GetWorkspace_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.GetWorkspace(context.Background(), &pb.GetWorkspaceRequest{
		WorkspaceId: "nonexistent",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_ListWorkspaces_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	h.CreateWorkspace(ctx, &pb.CreateWorkspaceRequest{AgentId: "a"})
	h.CreateWorkspace(ctx, &pb.CreateWorkspaceRequest{AgentId: "a"})
	h.CreateWorkspace(ctx, &pb.CreateWorkspaceRequest{AgentId: "b"})

	resp, err := h.ListWorkspaces(ctx, &pb.ListWorkspacesRequest{
		AgentId:  "a",
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Workspaces) != 2 {
		t.Errorf("count = %d, want 2", len(resp.Workspaces))
	}
}

func TestHandler_TerminateWorkspace_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	created, _ := h.CreateWorkspace(ctx, &pb.CreateWorkspaceRequest{AgentId: "a"})

	_, err := h.TerminateWorkspace(ctx, &pb.TerminateWorkspaceRequest{
		WorkspaceId: created.Workspace.WorkspaceId,
		Reason:      "test shutdown",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := h.GetWorkspace(ctx, &pb.GetWorkspaceRequest{
		WorkspaceId: created.Workspace.WorkspaceId,
	})
	if got.Workspace.Status != pb.WorkspaceStatus_WORKSPACE_STATUS_TERMINATED {
		t.Errorf("status = %v, want TERMINATED", got.Workspace.Status)
	}
}

func TestHandler_TerminateWorkspace_AlreadyTerminal(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	created, _ := h.CreateWorkspace(ctx, &pb.CreateWorkspaceRequest{AgentId: "a"})
	h.TerminateWorkspace(ctx, &pb.TerminateWorkspaceRequest{
		WorkspaceId: created.Workspace.WorkspaceId,
		Reason:      "first",
	})

	_, err := h.TerminateWorkspace(ctx, &pb.TerminateWorkspaceRequest{
		WorkspaceId: created.Workspace.WorkspaceId,
		Reason:      "second",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", st.Code())
	}
}

func TestHandler_TerminateWorkspace_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.TerminateWorkspace(context.Background(), &pb.TerminateWorkspaceRequest{
		WorkspaceId: "nonexistent",
		Reason:      "test",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_WorkspaceStatusConversion_RoundTrip(t *testing.T) {
	tests := []struct {
		proto pb.WorkspaceStatus
		model string
	}{
		{pb.WorkspaceStatus_WORKSPACE_STATUS_PENDING, "pending"},
		{pb.WorkspaceStatus_WORKSPACE_STATUS_CREATING, "creating"},
		{pb.WorkspaceStatus_WORKSPACE_STATUS_RUNNING, "running"},
		{pb.WorkspaceStatus_WORKSPACE_STATUS_PAUSED, "paused"},
		{pb.WorkspaceStatus_WORKSPACE_STATUS_TERMINATING, "terminating"},
		{pb.WorkspaceStatus_WORKSPACE_STATUS_TERMINATED, "terminated"},
		{pb.WorkspaceStatus_WORKSPACE_STATUS_FAILED, "failed"},
	}
	for _, tt := range tests {
		model := protoWorkspaceStatusToModel(tt.proto)
		if string(model) != tt.model {
			t.Errorf("protoWorkspaceStatusToModel(%v) = %q, want %q", tt.proto, model, tt.model)
		}
		back := modelWorkspaceStatusToProto(model)
		if back != tt.proto {
			t.Errorf("modelWorkspaceStatusToProto(%q) = %v, want %v", model, back, tt.proto)
		}
	}
}

func TestHandler_SpecConversion_DurationRoundTrip(t *testing.T) {
	spec := &pb.WorkspaceSpec{
		MemoryMb:      1024,
		CpuMillicores: 500,
		DiskMb:        2048,
		MaxDuration:   durationpb.New(90 * time.Minute),
		AllowedTools:  []string{"shell"},
		EnvVars:       map[string]string{"A": "B"},
	}
	model := protoSpecToModel(spec)
	if model.MaxDurationSecs != 5400 {
		t.Errorf("MaxDurationSecs = %d, want 5400", model.MaxDurationSecs)
	}
	if model.MemoryMb != 1024 {
		t.Errorf("MemoryMb = %d, want 1024", model.MemoryMb)
	}

	back := modelSpecToProto(model)
	if back.MaxDuration.AsDuration() != 90*time.Minute {
		t.Errorf("MaxDuration = %v, want 90m", back.MaxDuration.AsDuration())
	}
}

package human

import (
	"context"
	"testing"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
	pb "github.com/baselyne/agent-sandbox-platform/control-plane/pkg/gen/human/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestHandler() *Handler {
	return NewHandler(NewService(newMockRepo()))
}

func createTestRequest(t *testing.T, h *Handler) *pb.HumanRequest {
	t.Helper()
	resp, err := h.CreateRequest(context.Background(), &pb.CreateHumanRequestRequest{
		WorkspaceId:    "ws-1",
		AgentId:        "agent-1",
		Question:       "Should I proceed?",
		Options:        []string{"yes", "no"},
		Context:        "Processing invoice #1234",
		TimeoutSeconds: 3600,
	})
	if err != nil {
		t.Fatalf("CreateRequest: %v", err)
	}
	return resp.Request
}

func TestHandler_CreateRequest_Success(t *testing.T) {
	h := newTestHandler()
	req := createTestRequest(t, h)

	if req.RequestId == "" {
		t.Error("expected request ID")
	}
	if req.WorkspaceId != "ws-1" {
		t.Errorf("workspace_id = %q, want 'ws-1'", req.WorkspaceId)
	}
	if req.AgentId != "agent-1" {
		t.Errorf("agent_id = %q, want 'agent-1'", req.AgentId)
	}
	if req.Question != "Should I proceed?" {
		t.Errorf("question = %q, want 'Should I proceed?'", req.Question)
	}
	if len(req.Options) != 2 {
		t.Errorf("options len = %d, want 2", len(req.Options))
	}
	if req.Context != "Processing invoice #1234" {
		t.Errorf("context = %q, want 'Processing invoice #1234'", req.Context)
	}
	if req.Status != pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_PENDING {
		t.Errorf("status = %v, want PENDING", req.Status)
	}
	if req.ExpiresAt == nil {
		t.Error("expected expires_at timestamp")
	}
	if req.CreatedAt == nil {
		t.Error("expected created_at timestamp")
	}
}

func TestHandler_CreateRequest_InvalidInput(t *testing.T) {
	h := newTestHandler()
	tests := []struct {
		name string
		req  *pb.CreateHumanRequestRequest
	}{
		{"empty workspace", &pb.CreateHumanRequestRequest{WorkspaceId: "", AgentId: "a", Question: "q"}},
		{"empty agent", &pb.CreateHumanRequestRequest{WorkspaceId: "ws", AgentId: "", Question: "q"}},
		{"empty question", &pb.CreateHumanRequestRequest{WorkspaceId: "ws", AgentId: "a", Question: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := h.CreateRequest(context.Background(), tt.req)
			st, _ := status.FromError(err)
			if st.Code() != codes.InvalidArgument {
				t.Errorf("code = %v, want InvalidArgument", st.Code())
			}
		})
	}
}

func TestHandler_GetRequest_Success(t *testing.T) {
	h := newTestHandler()
	created := createTestRequest(t, h)

	resp, err := h.GetRequest(context.Background(), &pb.GetHumanRequestRequest{
		RequestId: created.RequestId,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Request.RequestId != created.RequestId {
		t.Errorf("ID mismatch: got %q, want %q", resp.Request.RequestId, created.RequestId)
	}
	if resp.Request.Question != "Should I proceed?" {
		t.Errorf("question = %q", resp.Request.Question)
	}
}

func TestHandler_GetRequest_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.GetRequest(context.Background(), &pb.GetHumanRequestRequest{
		RequestId: "nonexistent",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_RespondToRequest_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()
	created := createTestRequest(t, h)

	_, err := h.RespondToRequest(ctx, &pb.RespondToHumanRequestRequest{
		RequestId:   created.RequestId,
		Response:    "yes",
		ResponderId: "human-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := h.GetRequest(ctx, &pb.GetHumanRequestRequest{RequestId: created.RequestId})
	if got.Request.Status != pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_RESPONDED {
		t.Errorf("status = %v, want RESPONDED", got.Request.Status)
	}
	if got.Request.Response != "yes" {
		t.Errorf("response = %q, want 'yes'", got.Request.Response)
	}
	if got.Request.ResponderId != "human-1" {
		t.Errorf("responder_id = %q, want 'human-1'", got.Request.ResponderId)
	}
	if got.Request.RespondedAt == nil {
		t.Error("expected responded_at timestamp")
	}
}

func TestHandler_RespondToRequest_InvalidInput(t *testing.T) {
	h := newTestHandler()
	_, err := h.RespondToRequest(context.Background(), &pb.RespondToHumanRequestRequest{
		RequestId:   "",
		Response:    "yes",
		ResponderId: "human-1",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_RespondToRequest_NotPending(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()
	created := createTestRequest(t, h)

	// Respond first time
	h.RespondToRequest(ctx, &pb.RespondToHumanRequestRequest{
		RequestId: created.RequestId, Response: "yes", ResponderId: "h1",
	})

	// Second respond should fail
	_, err := h.RespondToRequest(ctx, &pb.RespondToHumanRequestRequest{
		RequestId: created.RequestId, Response: "no", ResponderId: "h2",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", st.Code())
	}
}

func TestHandler_ListRequests_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	createTestRequest(t, h)
	createTestRequest(t, h)

	resp, err := h.ListRequests(ctx, &pb.ListHumanRequestsRequest{
		WorkspaceId: "ws-1",
		PageSize:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Requests) != 2 {
		t.Errorf("requests = %d, want 2", len(resp.Requests))
	}
}

func TestHandler_ListRequests_StatusFilter(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	created := createTestRequest(t, h)
	createTestRequest(t, h)

	h.RespondToRequest(ctx, &pb.RespondToHumanRequestRequest{
		RequestId: created.RequestId, Response: "yes", ResponderId: "h1",
	})

	resp, err := h.ListRequests(ctx, &pb.ListHumanRequestsRequest{
		Status:   pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_PENDING,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Requests) != 1 {
		t.Errorf("pending requests = %d, want 1", len(resp.Requests))
	}
}

func TestHandler_StatusConversion_RoundTrip(t *testing.T) {
	tests := []struct {
		proto pb.HumanRequestStatus
		model models.HumanRequestStatus
	}{
		{pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_PENDING, models.HumanRequestStatusPending},
		{pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_RESPONDED, models.HumanRequestStatusResponded},
		{pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_EXPIRED, models.HumanRequestStatusExpired},
		{pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_CANCELLED, models.HumanRequestStatusCancelled},
	}
	for _, tt := range tests {
		got := protoStatusToModel(tt.proto)
		if got != tt.model {
			t.Errorf("protoStatusToModel(%v) = %q, want %q", tt.proto, got, tt.model)
		}
		back := modelStatusToProto(got)
		if back != tt.proto {
			t.Errorf("modelStatusToProto(%q) = %v, want %v", got, back, tt.proto)
		}
	}
}

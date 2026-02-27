package activity

import (
	"context"
	"testing"

	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/activity/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func newTestHandler() *Handler {
	return NewHandler(NewService(newMockRepo()))
}

func validProtoRecord() *pb.ActionRecord {
	params, _ := structpb.NewStruct(map[string]interface{}{"cmd": "ls"})
	return &pb.ActionRecord{
		WorkspaceId: "ws-1",
		AgentId:     "agent-1",
		ToolName:    "shell",
		Outcome:     pb.ActionOutcome_ACTION_OUTCOME_ALLOWED,
		Parameters:  params,
	}
}

func TestHandler_RecordAction_Success(t *testing.T) {
	h := newTestHandler()
	resp, err := h.RecordAction(context.Background(), &pb.RecordActionRequest{
		Record: validProtoRecord(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RecordId == "" {
		t.Error("expected record ID")
	}
}

func TestHandler_RecordAction_NilRecord(t *testing.T) {
	h := newTestHandler()
	_, err := h.RecordAction(context.Background(), &pb.RecordActionRequest{
		Record: nil,
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_RecordAction_MissingWorkspaceID(t *testing.T) {
	h := newTestHandler()
	rec := validProtoRecord()
	rec.WorkspaceId = ""
	_, err := h.RecordAction(context.Background(), &pb.RecordActionRequest{Record: rec})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_RecordAction_WithLatencies(t *testing.T) {
	h := newTestHandler()
	rec := validProtoRecord()
	rec.EvaluationLatencyUs = 150
	rec.ExecutionLatencyUs = 5000

	resp, err := h.RecordAction(context.Background(), &pb.RecordActionRequest{Record: rec})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := h.GetAction(context.Background(), &pb.GetActionRequest{RecordId: resp.RecordId})
	if got.Record.EvaluationLatencyUs != 150 {
		t.Errorf("eval_latency = %d, want 150", got.Record.EvaluationLatencyUs)
	}
	if got.Record.ExecutionLatencyUs != 5000 {
		t.Errorf("exec_latency = %d, want 5000", got.Record.ExecutionLatencyUs)
	}
}

func TestHandler_GetAction_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	created, _ := h.RecordAction(ctx, &pb.RecordActionRequest{Record: validProtoRecord()})

	resp, err := h.GetAction(ctx, &pb.GetActionRequest{RecordId: created.RecordId})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Record.RecordId != created.RecordId {
		t.Errorf("ID = %q, want %q", resp.Record.RecordId, created.RecordId)
	}
	if resp.Record.WorkspaceId != "ws-1" {
		t.Errorf("workspace_id = %q, want 'ws-1'", resp.Record.WorkspaceId)
	}
	if resp.Record.ToolName != "shell" {
		t.Errorf("tool_name = %q, want 'shell'", resp.Record.ToolName)
	}
	if resp.Record.Outcome != pb.ActionOutcome_ACTION_OUTCOME_ALLOWED {
		t.Errorf("outcome = %v, want ALLOWED", resp.Record.Outcome)
	}
	if resp.Record.RecordedAt == nil {
		t.Error("expected recorded_at timestamp")
	}
}

func TestHandler_GetAction_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.GetAction(context.Background(), &pb.GetActionRequest{
		RecordId: "nonexistent",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_QueryActions_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		h.RecordAction(ctx, &pb.RecordActionRequest{Record: validProtoRecord()})
	}

	resp, err := h.QueryActions(ctx, &pb.QueryActionsRequest{
		WorkspaceId: "ws-1",
		PageSize:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Records) != 5 {
		t.Errorf("records = %d, want 5", len(resp.Records))
	}
	if resp.TotalCount != 5 {
		t.Errorf("total_count = %d, want 5", resp.TotalCount)
	}
}

func TestHandler_QueryActions_Pagination(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		h.RecordAction(ctx, &pb.RecordActionRequest{Record: validProtoRecord()})
	}

	resp, err := h.QueryActions(ctx, &pb.QueryActionsRequest{PageSize: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Records) != 3 {
		t.Fatalf("page 1: %d records, want 3", len(resp.Records))
	}
	if resp.NextPageToken == "" {
		t.Error("expected next page token")
	}

	resp2, err := h.QueryActions(ctx, &pb.QueryActionsRequest{
		PageSize:  3,
		PageToken: resp.NextPageToken,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp2.Records) != 2 {
		t.Fatalf("page 2: %d records, want 2", len(resp2.Records))
	}
	if resp2.NextPageToken != "" {
		t.Error("expected no next page token on last page")
	}
}

func TestHandler_OutcomeConversion_RoundTrip(t *testing.T) {
	tests := []struct {
		proto pb.ActionOutcome
		model string
	}{
		{pb.ActionOutcome_ACTION_OUTCOME_ALLOWED, "allowed"},
		{pb.ActionOutcome_ACTION_OUTCOME_DENIED, "denied"},
		{pb.ActionOutcome_ACTION_OUTCOME_ESCALATED, "escalated"},
		{pb.ActionOutcome_ACTION_OUTCOME_ERROR, "error"},
	}
	for _, tt := range tests {
		model := protoOutcomeToModel(tt.proto)
		if string(model) != tt.model {
			t.Errorf("protoOutcomeToModel(%v) = %q, want %q", tt.proto, model, tt.model)
		}
		back := modelOutcomeToProto(model)
		if back != tt.proto {
			t.Errorf("modelOutcomeToProto(%q) = %v, want %v", model, back, tt.proto)
		}
	}
}

func TestHandler_StructJSON_Conversions(t *testing.T) {
	// Test structToJSON
	s, _ := structpb.NewStruct(map[string]interface{}{"key": "value"})
	j := structToJSON(s)
	if j == nil {
		t.Fatal("expected non-nil JSON")
	}

	// Test jsonToStruct
	back, err := jsonToStruct(j)
	if err != nil {
		t.Fatalf("jsonToStruct: %v", err)
	}
	if back.Fields["key"].GetStringValue() != "value" {
		t.Errorf("round-trip failed: key = %v", back.Fields["key"])
	}

	// Test nil handling
	if structToJSON(nil) != nil {
		t.Error("structToJSON(nil) should return nil")
	}
	if result, err := jsonToStruct(nil); result != nil || err != nil {
		t.Error("jsonToStruct(nil) should return nil, nil")
	}
}

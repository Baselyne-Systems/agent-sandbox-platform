package guardrails

import (
	"context"
	"testing"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
	pb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/guardrails/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestHandler() *Handler {
	return NewHandler(NewService(newMockRepo()))
}

func createTestRule(t *testing.T, h *Handler) *pb.GuardrailRule {
	t.Helper()
	resp, err := h.CreateRule(context.Background(), &pb.CreateRuleRequest{
		Name:      "deny-shell",
		Type:      pb.RuleType_RULE_TYPE_TOOL_FILTER,
		Condition: "shell,bash",
		Action:    pb.RuleAction_RULE_ACTION_DENY,
		Priority:  10,
		Labels:    map[string]string{"env": "prod"},
	})
	if err != nil {
		t.Fatalf("CreateRule: %v", err)
	}
	return resp.Rule
}

func TestHandler_CreateRule_Success(t *testing.T) {
	h := newTestHandler()
	rule := createTestRule(t, h)

	if rule.RuleId == "" {
		t.Error("expected rule ID")
	}
	if rule.Name != "deny-shell" {
		t.Errorf("name = %q, want 'deny-shell'", rule.Name)
	}
	if rule.Type != pb.RuleType_RULE_TYPE_TOOL_FILTER {
		t.Errorf("type = %v, want TOOL_FILTER", rule.Type)
	}
	if rule.Action != pb.RuleAction_RULE_ACTION_DENY {
		t.Errorf("action = %v, want DENY", rule.Action)
	}
	if rule.Priority != 10 {
		t.Errorf("priority = %d, want 10", rule.Priority)
	}
	if !rule.Enabled {
		t.Error("expected rule to be enabled by default")
	}
	if rule.Labels["env"] != "prod" {
		t.Errorf("labels[env] = %q, want 'prod'", rule.Labels["env"])
	}
	if rule.CreatedAt == nil {
		t.Error("expected created_at timestamp")
	}
}

func TestHandler_CreateRule_InvalidInput(t *testing.T) {
	h := newTestHandler()

	_, err := h.CreateRule(context.Background(), &pb.CreateRuleRequest{
		Name:      "",
		Type:      pb.RuleType_RULE_TYPE_TOOL_FILTER,
		Condition: "shell",
		Action:    pb.RuleAction_RULE_ACTION_DENY,
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_GetRule_Success(t *testing.T) {
	h := newTestHandler()
	created := createTestRule(t, h)

	resp, err := h.GetRule(context.Background(), &pb.GetRuleRequest{RuleId: created.RuleId})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Rule.RuleId != created.RuleId {
		t.Errorf("ID = %q, want %q", resp.Rule.RuleId, created.RuleId)
	}
	if resp.Rule.Condition != "shell,bash" {
		t.Errorf("condition = %q, want 'shell,bash'", resp.Rule.Condition)
	}
}

func TestHandler_GetRule_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.GetRule(context.Background(), &pb.GetRuleRequest{RuleId: "nonexistent"})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_UpdateRule_Success(t *testing.T) {
	h := newTestHandler()
	created := createTestRule(t, h)

	resp, err := h.UpdateRule(context.Background(), &pb.UpdateRuleRequest{
		Rule: &pb.GuardrailRule{
			RuleId:    created.RuleId,
			Name:      "allow-shell",
			Type:      pb.RuleType_RULE_TYPE_TOOL_FILTER,
			Condition: "shell",
			Action:    pb.RuleAction_RULE_ACTION_ALLOW,
			Priority:  5,
			Enabled:   true,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Rule.Name != "allow-shell" {
		t.Errorf("name = %q, want 'allow-shell'", resp.Rule.Name)
	}
	if resp.Rule.Action != pb.RuleAction_RULE_ACTION_ALLOW {
		t.Errorf("action = %v, want ALLOW", resp.Rule.Action)
	}
}

func TestHandler_UpdateRule_NilRule(t *testing.T) {
	h := newTestHandler()
	_, err := h.UpdateRule(context.Background(), &pb.UpdateRuleRequest{Rule: nil})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_DeleteRule_Success(t *testing.T) {
	h := newTestHandler()
	created := createTestRule(t, h)

	_, err := h.DeleteRule(context.Background(), &pb.DeleteRuleRequest{RuleId: created.RuleId})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify deleted
	_, err = h.GetRule(context.Background(), &pb.GetRuleRequest{RuleId: created.RuleId})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound after delete, got %v", st.Code())
	}
}

func TestHandler_DeleteRule_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.DeleteRule(context.Background(), &pb.DeleteRuleRequest{RuleId: "nonexistent"})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_ListRules_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	createTestRule(t, h)
	createTestRule(t, h)

	resp, err := h.ListRules(ctx, &pb.ListRulesRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Rules) != 2 {
		t.Errorf("rules count = %d, want 2", len(resp.Rules))
	}
}

func TestHandler_CompilePolicy_Success(t *testing.T) {
	h := newTestHandler()
	rule := createTestRule(t, h)

	resp, err := h.CompilePolicy(context.Background(), &pb.CompilePolicyRequest{
		RuleIds: []string{rule.RuleId},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.CompiledPolicy) == 0 {
		t.Error("expected non-empty compiled policy")
	}
	if resp.RuleCount != 1 {
		t.Errorf("rule_count = %d, want 1", resp.RuleCount)
	}
}

func TestHandler_CompilePolicy_EmptyRuleIDs(t *testing.T) {
	h := newTestHandler()
	_, err := h.CompilePolicy(context.Background(), &pb.CompilePolicyRequest{
		RuleIds: []string{},
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_RuleTypeConversion_RoundTrip(t *testing.T) {
	tests := []struct {
		proto pb.RuleType
		model models.RuleType
	}{
		{pb.RuleType_RULE_TYPE_TOOL_FILTER, models.RuleTypeToolFilter},
		{pb.RuleType_RULE_TYPE_PARAMETER_CHECK, models.RuleTypeParameterCheck},
		{pb.RuleType_RULE_TYPE_RATE_LIMIT, models.RuleTypeRateLimit},
		{pb.RuleType_RULE_TYPE_BUDGET_LIMIT, models.RuleTypeBudgetLimit},
	}
	for _, tt := range tests {
		got := protoRuleTypeToModel(tt.proto)
		if got != tt.model {
			t.Errorf("protoRuleTypeToModel(%v) = %q, want %q", tt.proto, got, tt.model)
		}
		back := modelRuleTypeToProto(got)
		if back != tt.proto {
			t.Errorf("modelRuleTypeToProto(%q) = %v, want %v", got, back, tt.proto)
		}
	}
}

func TestHandler_RuleActionConversion_RoundTrip(t *testing.T) {
	tests := []struct {
		proto pb.RuleAction
		model models.RuleAction
	}{
		{pb.RuleAction_RULE_ACTION_ALLOW, models.RuleActionAllow},
		{pb.RuleAction_RULE_ACTION_DENY, models.RuleActionDeny},
		{pb.RuleAction_RULE_ACTION_ESCALATE, models.RuleActionEscalate},
		{pb.RuleAction_RULE_ACTION_LOG, models.RuleActionLog},
	}
	for _, tt := range tests {
		got := protoRuleActionToModel(tt.proto)
		if got != tt.model {
			t.Errorf("protoRuleActionToModel(%v) = %q, want %q", tt.proto, got, tt.model)
		}
		back := modelRuleActionToProto(got)
		if back != tt.proto {
			t.Errorf("modelRuleActionToProto(%q) = %v, want %v", got, back, tt.proto)
		}
	}
}

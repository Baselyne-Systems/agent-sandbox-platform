package guardrails

import (
	"context"
	"errors"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/guardrails/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler implements the GuardrailsServiceServer gRPC interface.
type Handler struct {
	pb.UnimplementedGuardrailsServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleResponse, error) {
	rule, err := h.svc.CreateRule(ctx,
		req.GetName(),
		req.GetDescription(),
		protoRuleTypeToModel(req.GetType()),
		req.GetCondition(),
		protoRuleActionToModel(req.GetAction()),
		int(req.GetPriority()),
		req.GetLabels(),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CreateRuleResponse{Rule: ruleToProto(rule)}, nil
}

func (h *Handler) GetRule(ctx context.Context, req *pb.GetRuleRequest) (*pb.GetRuleResponse, error) {
	rule, err := h.svc.GetRule(ctx, req.GetRuleId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetRuleResponse{Rule: ruleToProto(rule)}, nil
}

func (h *Handler) ListRules(ctx context.Context, req *pb.ListRulesRequest) (*pb.ListRulesResponse, error) {
	typeFilter := protoRuleTypeToModel(req.GetType())
	rules, nextToken, err := h.svc.ListRules(ctx, typeFilter, req.GetEnabledOnly(), int(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbRules := make([]*pb.GuardrailRule, len(rules))
	for i := range rules {
		pbRules[i] = ruleToProto(&rules[i])
	}
	return &pb.ListRulesResponse{
		Rules:         pbRules,
		NextPageToken: nextToken,
	}, nil
}

func (h *Handler) UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.UpdateRuleResponse, error) {
	pbRule := req.GetRule()
	if pbRule == nil {
		return nil, status.Error(codes.InvalidArgument, "rule is required")
	}
	rule := protoToRule(pbRule)
	updated, err := h.svc.UpdateRule(ctx, rule)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.UpdateRuleResponse{Rule: ruleToProto(updated)}, nil
}

func (h *Handler) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleResponse, error) {
	if err := h.svc.DeleteRule(ctx, req.GetRuleId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeleteRuleResponse{}, nil
}

func (h *Handler) CompilePolicy(ctx context.Context, req *pb.CompilePolicyRequest) (*pb.CompilePolicyResponse, error) {
	compiled, count, err := h.svc.CompilePolicy(ctx, req.GetRuleIds())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CompilePolicyResponse{
		CompiledPolicy: compiled,
		RuleCount:      int32(count),
	}, nil
}

func (h *Handler) SimulatePolicy(ctx context.Context, req *pb.SimulatePolicyRequest) (*pb.SimulatePolicyResponse, error) {
	result, err := h.svc.SimulatePolicy(ctx, req.GetRuleIds(), req.GetToolName(), req.GetParameters(), req.GetAgentId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.SimulatePolicyResponse{
		Verdict:         result.Verdict,
		MatchedRuleId:   result.MatchedRuleID,
		MatchedRuleName: result.MatchedRuleName,
		Reason:          result.Reason,
	}, nil
}

// --- converters ---

func ruleToProto(r *models.GuardrailRule) *pb.GuardrailRule {
	return &pb.GuardrailRule{
		RuleId:      r.ID,
		Name:        r.Name,
		Description: r.Description,
		Type:        modelRuleTypeToProto(r.Type),
		Condition:   r.Condition,
		Action:      modelRuleActionToProto(r.Action),
		Priority:    int32(r.Priority),
		Enabled:     r.Enabled,
		Labels:      r.Labels,
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
	}
}

func protoToRule(p *pb.GuardrailRule) *models.GuardrailRule {
	return &models.GuardrailRule{
		ID:          p.GetRuleId(),
		Name:        p.GetName(),
		Description: p.GetDescription(),
		Type:        protoRuleTypeToModel(p.GetType()),
		Condition:   p.GetCondition(),
		Action:      protoRuleActionToModel(p.GetAction()),
		Priority:    int(p.GetPriority()),
		Enabled:     p.GetEnabled(),
		Labels:      p.GetLabels(),
	}
}

func modelRuleTypeToProto(t models.RuleType) pb.RuleType {
	switch t {
	case models.RuleTypeToolFilter:
		return pb.RuleType_RULE_TYPE_TOOL_FILTER
	case models.RuleTypeParameterCheck:
		return pb.RuleType_RULE_TYPE_PARAMETER_CHECK
	case models.RuleTypeRateLimit:
		return pb.RuleType_RULE_TYPE_RATE_LIMIT
	case models.RuleTypeBudgetLimit:
		return pb.RuleType_RULE_TYPE_BUDGET_LIMIT
	default:
		return pb.RuleType_RULE_TYPE_UNSPECIFIED
	}
}

func protoRuleTypeToModel(t pb.RuleType) models.RuleType {
	switch t {
	case pb.RuleType_RULE_TYPE_TOOL_FILTER:
		return models.RuleTypeToolFilter
	case pb.RuleType_RULE_TYPE_PARAMETER_CHECK:
		return models.RuleTypeParameterCheck
	case pb.RuleType_RULE_TYPE_RATE_LIMIT:
		return models.RuleTypeRateLimit
	case pb.RuleType_RULE_TYPE_BUDGET_LIMIT:
		return models.RuleTypeBudgetLimit
	default:
		return ""
	}
}

func modelRuleActionToProto(a models.RuleAction) pb.RuleAction {
	switch a {
	case models.RuleActionAllow:
		return pb.RuleAction_RULE_ACTION_ALLOW
	case models.RuleActionDeny:
		return pb.RuleAction_RULE_ACTION_DENY
	case models.RuleActionEscalate:
		return pb.RuleAction_RULE_ACTION_ESCALATE
	case models.RuleActionLog:
		return pb.RuleAction_RULE_ACTION_LOG
	default:
		return pb.RuleAction_RULE_ACTION_UNSPECIFIED
	}
}

func protoRuleActionToModel(a pb.RuleAction) models.RuleAction {
	switch a {
	case pb.RuleAction_RULE_ACTION_ALLOW:
		return models.RuleActionAllow
	case pb.RuleAction_RULE_ACTION_DENY:
		return models.RuleActionDeny
	case pb.RuleAction_RULE_ACTION_ESCALATE:
		return models.RuleActionEscalate
	case pb.RuleAction_RULE_ACTION_LOG:
		return models.RuleActionLog
	default:
		return ""
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrRuleNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

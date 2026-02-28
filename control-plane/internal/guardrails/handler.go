package guardrails

import (
	"context"
	"errors"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/middleware"
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
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	rule, err := h.svc.CreateRule(ctx,
		tenantID,
		req.GetName(),
		req.GetDescription(),
		protoRuleTypeToModel(req.GetType()),
		req.GetCondition(),
		protoRuleActionToModel(req.GetAction()),
		int(req.GetPriority()),
		req.GetLabels(),
		protoScopeToModel(req.GetScope()),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CreateRuleResponse{Rule: ruleToProto(rule)}, nil
}

func (h *Handler) GetRule(ctx context.Context, req *pb.GetRuleRequest) (*pb.GetRuleResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	rule, err := h.svc.GetRule(ctx, tenantID, req.GetRuleId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetRuleResponse{Rule: ruleToProto(rule)}, nil
}

func (h *Handler) ListRules(ctx context.Context, req *pb.ListRulesRequest) (*pb.ListRulesResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	typeFilter := protoRuleTypeToModel(req.GetType())
	rules, nextToken, err := h.svc.ListRules(ctx, tenantID, typeFilter, req.GetEnabledOnly(), int(req.GetPageSize()), req.GetPageToken())
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
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	pbRule := req.GetRule()
	if pbRule == nil {
		return nil, status.Error(codes.InvalidArgument, "rule is required")
	}
	rule := protoToRule(pbRule)
	updated, err := h.svc.UpdateRule(ctx, tenantID, rule)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.UpdateRuleResponse{Rule: ruleToProto(updated)}, nil
}

func (h *Handler) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	if err := h.svc.DeleteRule(ctx, tenantID, req.GetRuleId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeleteRuleResponse{}, nil
}

func (h *Handler) CompilePolicy(ctx context.Context, req *pb.CompilePolicyRequest) (*pb.CompilePolicyResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	compiled, count, err := h.svc.CompilePolicy(ctx, tenantID, req.GetRuleIds())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CompilePolicyResponse{
		CompiledPolicy: compiled,
		RuleCount:      int32(count),
	}, nil
}

func (h *Handler) SimulatePolicy(ctx context.Context, req *pb.SimulatePolicyRequest) (*pb.SimulatePolicyResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	result, err := h.svc.SimulatePolicy(ctx, tenantID, req.GetRuleIds(), req.GetToolName(), req.GetParameters(), req.GetAgentId())
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

func (h *Handler) GetBehaviorReport(ctx context.Context, req *pb.GetBehaviorReportRequest) (*pb.GetBehaviorReportResponse, error) {
	windowStart := req.GetWindowStart().AsTime()
	windowEnd := req.GetWindowEnd().AsTime()

	report, err := h.svc.GetBehaviorReport(ctx, req.GetAgentId(), windowStart, windowEnd)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetBehaviorReportResponse{
		Report: &pb.BehaviorReport{
			AgentId:        report.AgentID,
			WindowStart:    timestamppb.New(report.WindowStart),
			WindowEnd:      timestamppb.New(report.WindowEnd),
			ActionCount:    report.ActionCount,
			DenialRate:     report.DenialRate,
			ErrorRate:      report.ErrorRate,
			Flags:          report.Flags,
			Recommendation: report.Recommendation,
		},
	}, nil
}

// --- GuardrailSet RPCs ---

func (h *Handler) CreateGuardrailSet(ctx context.Context, req *pb.CreateGuardrailSetRequest) (*pb.CreateGuardrailSetResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	set, err := h.svc.CreateSet(ctx, tenantID, req.GetName(), req.GetDescription(), req.GetRuleIds(), req.GetLabels())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CreateGuardrailSetResponse{Set: setToProto(set)}, nil
}

func (h *Handler) GetGuardrailSet(ctx context.Context, req *pb.GetGuardrailSetRequest) (*pb.GetGuardrailSetResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	set, err := h.svc.GetSet(ctx, tenantID, req.GetSetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetGuardrailSetResponse{Set: setToProto(set)}, nil
}

func (h *Handler) ListGuardrailSets(ctx context.Context, req *pb.ListGuardrailSetsRequest) (*pb.ListGuardrailSetsResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	sets, nextToken, err := h.svc.ListSets(ctx, tenantID, int(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return nil, toGRPCError(err)
	}
	pbSets := make([]*pb.GuardrailSet, len(sets))
	for i := range sets {
		pbSets[i] = setToProto(&sets[i])
	}
	return &pb.ListGuardrailSetsResponse{
		Sets:          pbSets,
		NextPageToken: nextToken,
	}, nil
}

func (h *Handler) UpdateGuardrailSet(ctx context.Context, req *pb.UpdateGuardrailSetRequest) (*pb.UpdateGuardrailSetResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	pbSet := req.GetSet()
	if pbSet == nil {
		return nil, status.Error(codes.InvalidArgument, "set is required")
	}
	set := protoToSet(pbSet)
	updated, err := h.svc.UpdateSet(ctx, tenantID, set)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.UpdateGuardrailSetResponse{Set: setToProto(updated)}, nil
}

func (h *Handler) DeleteGuardrailSet(ctx context.Context, req *pb.DeleteGuardrailSetRequest) (*pb.DeleteGuardrailSetResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	if err := h.svc.DeleteSet(ctx, tenantID, req.GetSetId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeleteGuardrailSetResponse{}, nil
}

// --- converters ---

func ruleToProto(r *models.GuardrailRule) *pb.GuardrailRule {
	return &pb.GuardrailRule{
		RuleId:      r.ID,
		TenantId:    r.TenantID,
		Name:        r.Name,
		Description: r.Description,
		Type:        modelRuleTypeToProto(r.Type),
		Condition:   r.Condition,
		Action:      modelRuleActionToProto(r.Action),
		Priority:    int32(r.Priority),
		Enabled:     r.Enabled,
		Labels:      r.Labels,
		Scope:       modelScopeToProto(r.Scope),
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
		Scope:       protoScopeToModel(p.GetScope()),
	}
}

func modelScopeToProto(s models.RuleScope) *pb.RuleScope {
	if len(s.AgentIDs) == 0 && len(s.ToolNames) == 0 && len(s.TrustLevels) == 0 && len(s.DataClassifications) == 0 {
		return nil
	}
	return &pb.RuleScope{
		AgentIds:            s.AgentIDs,
		ToolNames:           s.ToolNames,
		TrustLevels:         s.TrustLevels,
		DataClassifications: s.DataClassifications,
	}
}

func protoScopeToModel(p *pb.RuleScope) models.RuleScope {
	if p == nil {
		return models.RuleScope{}
	}
	return models.RuleScope{
		AgentIDs:            p.GetAgentIds(),
		ToolNames:           p.GetToolNames(),
		TrustLevels:         p.GetTrustLevels(),
		DataClassifications: p.GetDataClassifications(),
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

func setToProto(s *models.GuardrailSet) *pb.GuardrailSet {
	return &pb.GuardrailSet{
		SetId:       s.ID,
		TenantId:    s.TenantID,
		Name:        s.Name,
		Description: s.Description,
		RuleIds:     s.RuleIDs,
		Labels:      s.Labels,
		CreatedAt:   timestamppb.New(s.CreatedAt),
		UpdatedAt:   timestamppb.New(s.UpdatedAt),
	}
}

func protoToSet(p *pb.GuardrailSet) *models.GuardrailSet {
	return &models.GuardrailSet{
		ID:          p.GetSetId(),
		Name:        p.GetName(),
		Description: p.GetDescription(),
		RuleIDs:     p.GetRuleIds(),
		Labels:      p.GetLabels(),
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrRuleNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrSetNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

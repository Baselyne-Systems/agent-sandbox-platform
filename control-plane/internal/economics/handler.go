package economics

import (
	"context"
	"errors"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/middleware"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/economics/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler implements the EconomicsServiceServer gRPC interface.
type Handler struct {
	pb.UnimplementedEconomicsServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RecordUsage(ctx context.Context, req *pb.RecordUsageRequest) (*pb.RecordUsageResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	r := req.GetRecord()
	if r == nil {
		return nil, status.Error(codes.InvalidArgument, "record is required")
	}
	record, err := h.svc.RecordUsage(ctx,
		tenantID,
		r.GetAgentId(),
		r.GetWorkspaceId(),
		r.GetResourceType(),
		r.GetUnit(),
		r.GetQuantity(),
		r.GetCost(),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RecordUsageResponse{RecordId: record.ID}, nil
}

func (h *Handler) GetBudget(ctx context.Context, req *pb.GetBudgetRequest) (*pb.GetBudgetResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	budget, err := h.svc.GetBudget(ctx, tenantID, req.GetAgentId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetBudgetResponse{Budget: budgetToProto(budget)}, nil
}

func (h *Handler) SetBudget(ctx context.Context, req *pb.SetBudgetRequest) (*pb.SetBudgetResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	onExceeded := onExceededProtoToString(req.GetOnExceeded())
	budget, err := h.svc.SetBudget(ctx, tenantID, req.GetAgentId(), req.GetLimit(), req.GetCurrency(), onExceeded, req.GetWarningThreshold())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.SetBudgetResponse{Budget: budgetToProto(budget)}, nil
}

func (h *Handler) CheckBudget(ctx context.Context, req *pb.CheckBudgetRequest) (*pb.CheckBudgetResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	result, err := h.svc.CheckBudget(ctx, tenantID, req.GetAgentId(), req.GetEstimatedCost())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CheckBudgetResponse{
		Allowed:           result.Allowed,
		Remaining:         result.Remaining,
		EnforcementAction: result.EnforcementAction,
		Warning:           result.Warning,
	}, nil
}

func (h *Handler) GetCostReport(ctx context.Context, req *pb.GetCostReportRequest) (*pb.GetCostReportResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	start := req.GetStartTime().AsTime()
	end := req.GetEndTime().AsTime()

	report, err := h.svc.GetCostReport(ctx, tenantID, req.GetAgentId(), start, end)
	if err != nil {
		return nil, toGRPCError(err)
	}

	breakdowns := make([]*pb.CostBreakdown, len(report.ByResourceType))
	for i, c := range report.ByResourceType {
		breakdowns[i] = &pb.CostBreakdown{
			ResourceType: c.ResourceType,
			TotalCost:    c.TotalCost,
			RecordCount:  int32(c.RecordCount),
		}
	}

	return &pb.GetCostReportResponse{
		TotalCost:      report.TotalCost,
		RecordCount:    int32(report.RecordCount),
		ByResourceType: breakdowns,
	}, nil
}

// --- converters ---

func budgetToProto(b *models.Budget) *pb.Budget {
	return &pb.Budget{
		BudgetId:         b.ID,
		TenantId:         b.TenantID,
		AgentId:          b.AgentID,
		Limit:            b.Limit,
		Used:             b.Used,
		Currency:         b.Currency,
		PeriodStart:      timestamppb.New(b.PeriodStart),
		PeriodEnd:        timestamppb.New(b.PeriodEnd),
		OnExceeded:       onExceededStringToProto(b.OnExceeded),
		WarningThreshold: b.WarningThreshold,
	}
}

func onExceededStringToProto(s string) pb.OnExceededAction {
	switch s {
	case "halt":
		return pb.OnExceededAction_ON_EXCEEDED_ACTION_HALT
	case "request_increase":
		return pb.OnExceededAction_ON_EXCEEDED_ACTION_REQUEST_INCREASE
	case "warn":
		return pb.OnExceededAction_ON_EXCEEDED_ACTION_WARN
	default:
		return pb.OnExceededAction_ON_EXCEEDED_ACTION_HALT
	}
}

func onExceededProtoToString(a pb.OnExceededAction) string {
	switch a {
	case pb.OnExceededAction_ON_EXCEEDED_ACTION_HALT:
		return "halt"
	case pb.OnExceededAction_ON_EXCEEDED_ACTION_REQUEST_INCREASE:
		return "request_increase"
	case pb.OnExceededAction_ON_EXCEEDED_ACTION_WARN:
		return "warn"
	default:
		return ""
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrBudgetNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

package economics

import (
	"context"
	"errors"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
	pb "github.com/baselyne/agent-sandbox-platform/control-plane/pkg/gen/economics/v1"
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
	r := req.GetRecord()
	if r == nil {
		return nil, status.Error(codes.InvalidArgument, "record is required")
	}
	record, err := h.svc.RecordUsage(ctx,
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
	budget, err := h.svc.GetBudget(ctx, req.GetAgentId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetBudgetResponse{Budget: budgetToProto(budget)}, nil
}

func (h *Handler) SetBudget(ctx context.Context, req *pb.SetBudgetRequest) (*pb.SetBudgetResponse, error) {
	budget, err := h.svc.SetBudget(ctx, req.GetAgentId(), req.GetLimit(), req.GetCurrency())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.SetBudgetResponse{Budget: budgetToProto(budget)}, nil
}

func (h *Handler) CheckBudget(ctx context.Context, req *pb.CheckBudgetRequest) (*pb.CheckBudgetResponse, error) {
	allowed, remaining, err := h.svc.CheckBudget(ctx, req.GetAgentId(), req.GetEstimatedCost())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CheckBudgetResponse{
		Allowed:   allowed,
		Remaining: remaining,
	}, nil
}

// --- converters ---

func budgetToProto(b *models.Budget) *pb.Budget {
	return &pb.Budget{
		BudgetId:    b.ID,
		AgentId:     b.AgentID,
		Limit:       b.Limit,
		Used:        b.Used,
		Currency:    b.Currency,
		PeriodStart: timestamppb.New(b.PeriodStart),
		PeriodEnd:   timestamppb.New(b.PeriodEnd),
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

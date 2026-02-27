package human

import (
	"context"
	"errors"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
	pb "github.com/baselyne/agent-sandbox-platform/control-plane/pkg/gen/human/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler implements the HumanInteractionServiceServer gRPC interface.
type Handler struct {
	pb.UnimplementedHumanInteractionServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateRequest(ctx context.Context, req *pb.CreateHumanRequestRequest) (*pb.CreateHumanRequestResponse, error) {
	result, err := h.svc.CreateRequest(ctx,
		req.GetWorkspaceId(),
		req.GetAgentId(),
		req.GetQuestion(),
		req.GetOptions(),
		req.GetContext(),
		req.GetTimeoutSeconds(),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CreateHumanRequestResponse{Request: requestToProto(result)}, nil
}

func (h *Handler) GetRequest(ctx context.Context, req *pb.GetHumanRequestRequest) (*pb.GetHumanRequestResponse, error) {
	result, err := h.svc.GetRequest(ctx, req.GetRequestId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetHumanRequestResponse{Request: requestToProto(result)}, nil
}

func (h *Handler) RespondToRequest(ctx context.Context, req *pb.RespondToHumanRequestRequest) (*pb.RespondToHumanRequestResponse, error) {
	if err := h.svc.RespondToRequest(ctx, req.GetRequestId(), req.GetResponse(), req.GetResponderId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RespondToHumanRequestResponse{}, nil
}

func (h *Handler) ListRequests(ctx context.Context, req *pb.ListHumanRequestsRequest) (*pb.ListHumanRequestsResponse, error) {
	statusFilter := protoStatusToModel(req.GetStatus())
	requests, nextToken, err := h.svc.ListRequests(ctx, req.GetWorkspaceId(), statusFilter, int(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbRequests := make([]*pb.HumanRequest, len(requests))
	for i := range requests {
		pbRequests[i] = requestToProto(&requests[i])
	}
	return &pb.ListHumanRequestsResponse{
		Requests:      pbRequests,
		NextPageToken: nextToken,
	}, nil
}

// --- converters ---

func requestToProto(r *models.HumanRequest) *pb.HumanRequest {
	p := &pb.HumanRequest{
		RequestId:   r.ID,
		WorkspaceId: r.WorkspaceID,
		AgentId:     r.AgentID,
		Question:    r.Question,
		Options:     r.Options,
		Context:     r.Context,
		Status:      modelStatusToProto(r.Status),
		Response:    r.Response,
		ResponderId: r.ResponderID,
		CreatedAt:   timestamppb.New(r.CreatedAt),
	}
	if r.RespondedAt != nil {
		p.RespondedAt = timestamppb.New(*r.RespondedAt)
	}
	if r.ExpiresAt != nil {
		p.ExpiresAt = timestamppb.New(*r.ExpiresAt)
	}
	return p
}

func modelStatusToProto(s models.HumanRequestStatus) pb.HumanRequestStatus {
	switch s {
	case models.HumanRequestStatusPending:
		return pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_PENDING
	case models.HumanRequestStatusResponded:
		return pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_RESPONDED
	case models.HumanRequestStatusExpired:
		return pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_EXPIRED
	case models.HumanRequestStatusCancelled:
		return pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_CANCELLED
	default:
		return pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_UNSPECIFIED
	}
}

func protoStatusToModel(s pb.HumanRequestStatus) models.HumanRequestStatus {
	switch s {
	case pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_PENDING:
		return models.HumanRequestStatusPending
	case pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_RESPONDED:
		return models.HumanRequestStatusResponded
	case pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_EXPIRED:
		return models.HumanRequestStatusExpired
	case pb.HumanRequestStatus_HUMAN_REQUEST_STATUS_CANCELLED:
		return models.HumanRequestStatusCancelled
	default:
		return ""
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrRequestNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrRequestNotPending):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

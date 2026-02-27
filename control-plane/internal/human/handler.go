package human

import (
	"context"
	"errors"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/human/v1"
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
		protoRequestTypeToModel(req.GetType()),
		protoUrgencyToModel(req.GetUrgency()),
		req.GetTaskId(),
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
		Type:        modelRequestTypeToProto(r.Type),
		Urgency:     modelUrgencyToProto(r.Urgency),
		TaskId:      r.TaskID,
	}
	if r.RespondedAt != nil {
		p.RespondedAt = timestamppb.New(*r.RespondedAt)
	}
	if r.ExpiresAt != nil {
		p.ExpiresAt = timestamppb.New(*r.ExpiresAt)
	}
	return p
}

func modelRequestTypeToProto(t models.HumanRequestType) pb.HumanRequestType {
	switch t {
	case models.HumanRequestTypeApproval:
		return pb.HumanRequestType_HUMAN_REQUEST_TYPE_APPROVAL
	case models.HumanRequestTypeQuestion:
		return pb.HumanRequestType_HUMAN_REQUEST_TYPE_QUESTION
	case models.HumanRequestTypeEscalation:
		return pb.HumanRequestType_HUMAN_REQUEST_TYPE_ESCALATION
	default:
		return pb.HumanRequestType_HUMAN_REQUEST_TYPE_UNSPECIFIED
	}
}

func protoRequestTypeToModel(t pb.HumanRequestType) models.HumanRequestType {
	switch t {
	case pb.HumanRequestType_HUMAN_REQUEST_TYPE_APPROVAL:
		return models.HumanRequestTypeApproval
	case pb.HumanRequestType_HUMAN_REQUEST_TYPE_QUESTION:
		return models.HumanRequestTypeQuestion
	case pb.HumanRequestType_HUMAN_REQUEST_TYPE_ESCALATION:
		return models.HumanRequestTypeEscalation
	default:
		return ""
	}
}

func modelUrgencyToProto(u models.HumanRequestUrgency) pb.HumanRequestUrgency {
	switch u {
	case models.HumanRequestUrgencyLow:
		return pb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_LOW
	case models.HumanRequestUrgencyNormal:
		return pb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_NORMAL
	case models.HumanRequestUrgencyHigh:
		return pb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_HIGH
	case models.HumanRequestUrgencyCritical:
		return pb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_CRITICAL
	default:
		return pb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_UNSPECIFIED
	}
}

func protoUrgencyToModel(u pb.HumanRequestUrgency) models.HumanRequestUrgency {
	switch u {
	case pb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_LOW:
		return models.HumanRequestUrgencyLow
	case pb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_NORMAL:
		return models.HumanRequestUrgencyNormal
	case pb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_HIGH:
		return models.HumanRequestUrgencyHigh
	case pb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_CRITICAL:
		return models.HumanRequestUrgencyCritical
	default:
		return ""
	}
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

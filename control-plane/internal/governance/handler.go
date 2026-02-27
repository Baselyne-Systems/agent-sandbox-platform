package governance

import (
	"context"
	"errors"

	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/governance/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler implements the DataGovernanceServiceServer gRPC interface.
type Handler struct {
	pb.UnimplementedDataGovernanceServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ClassifyData(_ context.Context, req *pb.ClassifyDataRequest) (*pb.ClassifyDataResponse, error) {
	classification, patterns, err := h.svc.ClassifyData(req.GetContent(), req.GetContentType())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.ClassifyDataResponse{
		Classification:   classificationToProto(classification),
		DetectedPatterns: patterns,
	}, nil
}

func (h *Handler) CheckPolicy(_ context.Context, req *pb.CheckPolicyRequest) (*pb.CheckPolicyResponse, error) {
	classification := protoToClassification(req.GetDataClassification())
	allowed, reason, err := h.svc.CheckPolicy(req.GetAgentId(), req.GetDestination(), classification)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CheckPolicyResponse{
		Allowed: allowed,
		Reason:  reason,
	}, nil
}

func (h *Handler) InspectEgress(_ context.Context, req *pb.InspectEgressRequest) (*pb.InspectEgressResponse, error) {
	allowed, reason, classification, patterns, err := h.svc.InspectEgress(
		req.GetAgentId(),
		req.GetDestination(),
		req.GetContent(),
		req.GetContentType(),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.InspectEgressResponse{
		Allowed:          allowed,
		Reason:           reason,
		Classification:   classificationToProto(classification),
		DetectedPatterns: patterns,
	}, nil
}

// --- converters ---

func classificationToProto(c DataClassification) pb.DataClassification {
	switch c {
	case ClassificationPublic:
		return pb.DataClassification_DATA_CLASSIFICATION_PUBLIC
	case ClassificationInternal:
		return pb.DataClassification_DATA_CLASSIFICATION_INTERNAL
	case ClassificationConfidential:
		return pb.DataClassification_DATA_CLASSIFICATION_CONFIDENTIAL
	case ClassificationRestricted:
		return pb.DataClassification_DATA_CLASSIFICATION_RESTRICTED
	default:
		return pb.DataClassification_DATA_CLASSIFICATION_UNSPECIFIED
	}
}

func protoToClassification(c pb.DataClassification) DataClassification {
	switch c {
	case pb.DataClassification_DATA_CLASSIFICATION_PUBLIC:
		return ClassificationPublic
	case pb.DataClassification_DATA_CLASSIFICATION_INTERNAL:
		return ClassificationInternal
	case pb.DataClassification_DATA_CLASSIFICATION_CONFIDENTIAL:
		return ClassificationConfidential
	case pb.DataClassification_DATA_CLASSIFICATION_RESTRICTED:
		return ClassificationRestricted
	default:
		return ClassificationPublic
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

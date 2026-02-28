package identity

import (
	"context"
	"errors"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/middleware"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/identity/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler implements the IdentityServiceServer gRPC interface.
type Handler struct {
	pb.UnimplementedIdentityServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterAgent(ctx context.Context, req *pb.RegisterAgentRequest) (*pb.RegisterAgentResponse, error) {
	tenantID := req.GetTenantId()
	if tenantID == "" {
		tenantID, _ = middleware.TenantIDFromContext(ctx)
	}
	agent, err := h.svc.RegisterAgent(ctx,
		tenantID,
		req.GetName(),
		req.GetDescription(),
		req.GetOwnerId(),
		req.GetLabels(),
		req.GetPurpose(),
		protoTrustLevelToModel(req.GetTrustLevel()),
		req.GetCapabilities(),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RegisterAgentResponse{Agent: agentToProto(agent)}, nil
}

func (h *Handler) GetAgent(ctx context.Context, req *pb.GetAgentRequest) (*pb.GetAgentResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	agent, err := h.svc.GetAgent(ctx, tenantID, req.GetAgentId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetAgentResponse{Agent: agentToProto(agent)}, nil
}

func (h *Handler) ListAgents(ctx context.Context, req *pb.ListAgentsRequest) (*pb.ListAgentsResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	statusFilter := protoStatusToModel(req.GetStatus())
	agents, nextToken, err := h.svc.ListAgents(ctx, tenantID, req.GetOwnerId(), statusFilter, int(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbAgents := make([]*pb.Agent, len(agents))
	for i := range agents {
		pbAgents[i] = agentToProto(&agents[i])
	}
	return &pb.ListAgentsResponse{
		Agents:        pbAgents,
		NextPageToken: nextToken,
	}, nil
}

func (h *Handler) DeactivateAgent(ctx context.Context, req *pb.DeactivateAgentRequest) (*pb.DeactivateAgentResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	if err := h.svc.DeactivateAgent(ctx, tenantID, req.GetAgentId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.DeactivateAgentResponse{}, nil
}

func (h *Handler) MintCredential(ctx context.Context, req *pb.MintCredentialRequest) (*pb.MintCredentialResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	cred, rawToken, err := h.svc.MintCredential(ctx, tenantID, req.GetAgentId(), req.GetScopes(), req.GetTtlSeconds())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.MintCredentialResponse{
		Credential: credToProto(cred),
		Token:      rawToken,
	}, nil
}

func (h *Handler) RevokeCredential(ctx context.Context, req *pb.RevokeCredentialRequest) (*pb.RevokeCredentialResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	if err := h.svc.RevokeCredential(ctx, tenantID, req.GetCredentialId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RevokeCredentialResponse{}, nil
}

func (h *Handler) UpdateTrustLevel(ctx context.Context, req *pb.UpdateTrustLevelRequest) (*pb.UpdateTrustLevelResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	agent, err := h.svc.UpdateTrustLevel(ctx, tenantID, req.GetAgentId(), protoTrustLevelToModel(req.GetTrustLevel()), req.GetJustification())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.UpdateTrustLevelResponse{Agent: agentToProto(agent)}, nil
}

func (h *Handler) SuspendAgent(ctx context.Context, req *pb.SuspendAgentRequest) (*pb.SuspendAgentResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	agent, err := h.svc.SuspendAgent(ctx, tenantID, req.GetAgentId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.SuspendAgentResponse{Agent: agentToProto(agent)}, nil
}

func (h *Handler) ReactivateAgent(ctx context.Context, req *pb.ReactivateAgentRequest) (*pb.ReactivateAgentResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	agent, err := h.svc.ReactivateAgent(ctx, tenantID, req.GetAgentId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.ReactivateAgentResponse{Agent: agentToProto(agent)}, nil
}

// --- converters ---

func agentToProto(a *models.Agent) *pb.Agent {
	return &pb.Agent{
		AgentId:      a.ID,
		TenantId:     a.TenantID,
		Name:         a.Name,
		Description:  a.Description,
		OwnerId:      a.OwnerID,
		Status:       modelStatusToProto(a.Status),
		Labels:       a.Labels,
		Purpose:      a.Purpose,
		TrustLevel:   modelTrustLevelToProto(a.TrustLevel),
		Capabilities: a.Capabilities,
		CreatedAt:    timestamppb.New(a.CreatedAt),
		UpdatedAt:    timestamppb.New(a.UpdatedAt),
	}
}

func modelTrustLevelToProto(t models.AgentTrustLevel) pb.AgentTrustLevel {
	switch t {
	case models.AgentTrustLevelNew:
		return pb.AgentTrustLevel_AGENT_TRUST_LEVEL_NEW
	case models.AgentTrustLevelEstablished:
		return pb.AgentTrustLevel_AGENT_TRUST_LEVEL_ESTABLISHED
	case models.AgentTrustLevelTrusted:
		return pb.AgentTrustLevel_AGENT_TRUST_LEVEL_TRUSTED
	default:
		return pb.AgentTrustLevel_AGENT_TRUST_LEVEL_UNSPECIFIED
	}
}

func protoTrustLevelToModel(t pb.AgentTrustLevel) models.AgentTrustLevel {
	switch t {
	case pb.AgentTrustLevel_AGENT_TRUST_LEVEL_NEW:
		return models.AgentTrustLevelNew
	case pb.AgentTrustLevel_AGENT_TRUST_LEVEL_ESTABLISHED:
		return models.AgentTrustLevelEstablished
	case pb.AgentTrustLevel_AGENT_TRUST_LEVEL_TRUSTED:
		return models.AgentTrustLevelTrusted
	default:
		return ""
	}
}

func credToProto(c *models.ScopedCredential) *pb.ScopedCredential {
	return &pb.ScopedCredential{
		CredentialId: c.ID,
		TenantId:     c.TenantID,
		AgentId:      c.AgentID,
		Scopes:       c.Scopes,
		ExpiresAt:    timestamppb.New(c.ExpiresAt),
		CreatedAt:    timestamppb.New(c.CreatedAt),
	}
}

func modelStatusToProto(s models.AgentStatus) pb.AgentStatus {
	switch s {
	case models.AgentStatusActive:
		return pb.AgentStatus_AGENT_STATUS_ACTIVE
	case models.AgentStatusInactive:
		return pb.AgentStatus_AGENT_STATUS_INACTIVE
	case models.AgentStatusSuspended:
		return pb.AgentStatus_AGENT_STATUS_SUSPENDED
	default:
		return pb.AgentStatus_AGENT_STATUS_UNSPECIFIED
	}
}

func protoStatusToModel(s pb.AgentStatus) models.AgentStatus {
	switch s {
	case pb.AgentStatus_AGENT_STATUS_ACTIVE:
		return models.AgentStatusActive
	case pb.AgentStatus_AGENT_STATUS_INACTIVE:
		return models.AgentStatusInactive
	case pb.AgentStatus_AGENT_STATUS_SUSPENDED:
		return models.AgentStatusSuspended
	default:
		return "" // unspecified = no filter
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrAgentNotFound), errors.Is(err, ErrCredentialNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrAgentInactive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, ErrInvalidStatusTransition):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, ErrInvalidTrustLevel):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

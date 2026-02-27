package workspace

import (
	"context"
	"errors"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/workspace/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler implements the WorkspaceServiceServer gRPC interface.
type Handler struct {
	pb.UnimplementedWorkspaceServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateWorkspace(ctx context.Context, req *pb.CreateWorkspaceRequest) (*pb.CreateWorkspaceResponse, error) {
	var spec *models.WorkspaceSpec
	if req.GetSpec() != nil {
		spec = protoSpecToModel(req.GetSpec())
	}
	ws, err := h.svc.CreateWorkspace(ctx, req.GetAgentId(), req.GetTaskId(), spec)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CreateWorkspaceResponse{Workspace: workspaceToProto(ws)}, nil
}

func (h *Handler) GetWorkspace(ctx context.Context, req *pb.GetWorkspaceRequest) (*pb.GetWorkspaceResponse, error) {
	ws, err := h.svc.GetWorkspace(ctx, req.GetWorkspaceId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetWorkspaceResponse{Workspace: workspaceToProto(ws)}, nil
}

func (h *Handler) ListWorkspaces(ctx context.Context, req *pb.ListWorkspacesRequest) (*pb.ListWorkspacesResponse, error) {
	statusFilter := protoWorkspaceStatusToModel(req.GetStatus())
	workspaces, nextToken, err := h.svc.ListWorkspaces(ctx, req.GetAgentId(), statusFilter, int(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbWorkspaces := make([]*pb.Workspace, len(workspaces))
	for i := range workspaces {
		pbWorkspaces[i] = workspaceToProto(&workspaces[i])
	}
	return &pb.ListWorkspacesResponse{
		Workspaces:    pbWorkspaces,
		NextPageToken: nextToken,
	}, nil
}

func (h *Handler) TerminateWorkspace(ctx context.Context, req *pb.TerminateWorkspaceRequest) (*pb.TerminateWorkspaceResponse, error) {
	if err := h.svc.TerminateWorkspace(ctx, req.GetWorkspaceId(), req.GetReason()); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.TerminateWorkspaceResponse{}, nil
}

func (h *Handler) SnapshotWorkspace(ctx context.Context, req *pb.SnapshotWorkspaceRequest) (*pb.SnapshotWorkspaceResponse, error) {
	snapshot, err := h.svc.SnapshotWorkspace(ctx, req.GetWorkspaceId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.SnapshotWorkspaceResponse{
		SnapshotId: snapshot.ID,
		CreatedAt:  timestamppb.New(snapshot.CreatedAt),
	}, nil
}

func (h *Handler) RestoreWorkspace(ctx context.Context, req *pb.RestoreWorkspaceRequest) (*pb.RestoreWorkspaceResponse, error) {
	ws, err := h.svc.RestoreWorkspace(ctx, req.GetSnapshotId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.RestoreWorkspaceResponse{Workspace: workspaceToProto(ws)}, nil
}

// --- converters ---

func workspaceToProto(ws *models.Workspace) *pb.Workspace {
	p := &pb.Workspace{
		WorkspaceId: ws.ID,
		AgentId:     ws.AgentID,
		TaskId:      ws.TaskID,
		Status:      modelWorkspaceStatusToProto(ws.Status),
		Spec:        modelSpecToProto(&ws.Spec),
		HostId:      ws.HostID,
		SnapshotId:  ws.SnapshotID,
		CreatedAt:   timestamppb.New(ws.CreatedAt),
		UpdatedAt:   timestamppb.New(ws.UpdatedAt),
	}
	if ws.ExpiresAt != nil {
		p.ExpiresAt = timestamppb.New(*ws.ExpiresAt)
	}
	return p
}

func modelSpecToProto(s *models.WorkspaceSpec) *pb.WorkspaceSpec {
	return &pb.WorkspaceSpec{
		MemoryMb:          s.MemoryMb,
		CpuMillicores:     s.CpuMillicores,
		DiskMb:            s.DiskMb,
		MaxDuration:       durationpb.New(time.Duration(s.MaxDurationSecs) * time.Second),
		AllowedTools:      s.AllowedTools,
		GuardrailPolicyId: s.GuardrailPolicyID,
		EnvVars:           s.EnvVars,
	}
}

func protoSpecToModel(s *pb.WorkspaceSpec) *models.WorkspaceSpec {
	spec := &models.WorkspaceSpec{
		MemoryMb:          s.GetMemoryMb(),
		CpuMillicores:     s.GetCpuMillicores(),
		DiskMb:            s.GetDiskMb(),
		AllowedTools:      s.GetAllowedTools(),
		GuardrailPolicyID: s.GetGuardrailPolicyId(),
		EnvVars:           s.GetEnvVars(),
	}
	if s.GetMaxDuration() != nil {
		spec.MaxDurationSecs = int64(s.GetMaxDuration().AsDuration().Seconds())
	}
	return spec
}

func modelWorkspaceStatusToProto(s models.WorkspaceStatus) pb.WorkspaceStatus {
	switch s {
	case models.WorkspaceStatusPending:
		return pb.WorkspaceStatus_WORKSPACE_STATUS_PENDING
	case models.WorkspaceStatusCreating:
		return pb.WorkspaceStatus_WORKSPACE_STATUS_CREATING
	case models.WorkspaceStatusRunning:
		return pb.WorkspaceStatus_WORKSPACE_STATUS_RUNNING
	case models.WorkspaceStatusPaused:
		return pb.WorkspaceStatus_WORKSPACE_STATUS_PAUSED
	case models.WorkspaceStatusTerminating:
		return pb.WorkspaceStatus_WORKSPACE_STATUS_TERMINATING
	case models.WorkspaceStatusTerminated:
		return pb.WorkspaceStatus_WORKSPACE_STATUS_TERMINATED
	case models.WorkspaceStatusFailed:
		return pb.WorkspaceStatus_WORKSPACE_STATUS_FAILED
	default:
		return pb.WorkspaceStatus_WORKSPACE_STATUS_UNSPECIFIED
	}
}

func protoWorkspaceStatusToModel(s pb.WorkspaceStatus) models.WorkspaceStatus {
	switch s {
	case pb.WorkspaceStatus_WORKSPACE_STATUS_PENDING:
		return models.WorkspaceStatusPending
	case pb.WorkspaceStatus_WORKSPACE_STATUS_CREATING:
		return models.WorkspaceStatusCreating
	case pb.WorkspaceStatus_WORKSPACE_STATUS_RUNNING:
		return models.WorkspaceStatusRunning
	case pb.WorkspaceStatus_WORKSPACE_STATUS_PAUSED:
		return models.WorkspaceStatusPaused
	case pb.WorkspaceStatus_WORKSPACE_STATUS_TERMINATING:
		return models.WorkspaceStatusTerminating
	case pb.WorkspaceStatus_WORKSPACE_STATUS_TERMINATED:
		return models.WorkspaceStatusTerminated
	case pb.WorkspaceStatus_WORKSPACE_STATUS_FAILED:
		return models.WorkspaceStatusFailed
	default:
		return ""
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrWorkspaceNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrWorkspaceAlreadyTerminal):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, ErrWorkspaceNotRunning):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, ErrSnapshotNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

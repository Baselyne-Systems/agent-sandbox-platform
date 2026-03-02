package task

import (
	"context"
	"errors"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/middleware"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
	pb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/task/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler implements the TaskServiceServer gRPC interface.
type Handler struct {
	pb.UnimplementedTaskServiceServer
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	var wsConfig *models.TaskWorkspaceConfig
	if req.GetWorkspaceConfig() != nil {
		wc := protoWSConfigToModel(req.GetWorkspaceConfig())
		wsConfig = &wc
	}
	var hiConfig *models.TaskHumanInteractionConfig
	if req.GetHumanInteraction() != nil {
		hi := protoHIConfigToModel(req.GetHumanInteraction())
		hiConfig = &hi
	}
	var budgetConfig *models.TaskBudgetConfig
	if req.GetBudget() != nil {
		bc := protoBudgetConfigToModel(req.GetBudget())
		budgetConfig = &bc
	}

	task, err := h.svc.CreateTask(ctx,
		tenantID,
		req.GetAgentId(),
		req.GetGoal(),
		wsConfig,
		req.GetGuardrailPolicyId(),
		hiConfig,
		budgetConfig,
		req.GetMaxDurationWithoutCheckinSecs(),
		req.GetInput(),
		req.GetLabels(),
	)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CreateTaskResponse{Task: taskToProto(task)}, nil
}

func (h *Handler) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	task, err := h.svc.GetTask(ctx, tenantID, req.GetTaskId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.GetTaskResponse{Task: taskToProto(task)}, nil
}

func (h *Handler) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	statusFilter := protoTaskStatusToModel(req.GetStatus())
	tasks, nextToken, err := h.svc.ListTasks(ctx, tenantID, req.GetAgentId(), statusFilter, int(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return nil, toGRPCError(err)
	}

	pbTasks := make([]*pb.Task, len(tasks))
	for i := range tasks {
		pbTasks[i] = taskToProto(&tasks[i])
	}
	return &pb.ListTasksResponse{
		Tasks:         pbTasks,
		NextPageToken: nextToken,
	}, nil
}

func (h *Handler) UpdateTaskStatus(ctx context.Context, req *pb.UpdateTaskStatusRequest) (*pb.UpdateTaskStatusResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	newStatus := protoTaskStatusToModel(req.GetStatus())
	task, err := h.svc.UpdateTaskStatus(ctx, tenantID, req.GetTaskId(), newStatus, req.GetReason())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.UpdateTaskStatusResponse{Task: taskToProto(task)}, nil
}

func (h *Handler) CancelTask(ctx context.Context, req *pb.CancelTaskRequest) (*pb.CancelTaskResponse, error) {
	tenantID, _ := middleware.TenantIDFromContext(ctx)
	if err := h.svc.CancelTask(ctx, tenantID, req.GetTaskId(), req.GetReason()); err != nil {
		return nil, toGRPCError(err)
	}
	return &pb.CancelTaskResponse{}, nil
}

// --- converters ---

func taskToProto(t *models.Task) *pb.Task {
	p := &pb.Task{
		TaskId:                        t.ID,
		TenantId:                      t.TenantID,
		AgentId:                       t.AgentID,
		Goal:                          t.Goal,
		Status:                        modelTaskStatusToProto(t.Status),
		WorkspaceId:                   t.WorkspaceID,
		GuardrailPolicyId:             t.GuardrailPolicyID,
		WorkspaceConfig:               modelWSConfigToProto(&t.WorkspaceConfig),
		HumanInteraction:              modelHIConfigToProto(&t.HumanInteractionConfig),
		Budget:                        modelBudgetConfigToProto(&t.BudgetConfig),
		MaxDurationWithoutCheckinSecs: t.MaxDurationWithoutCheckinSecs,
		Input:                         t.Input,
		Labels:                        t.Labels,
		CreatedAt:                     timestamppb.New(t.CreatedAt),
		UpdatedAt:                     timestamppb.New(t.UpdatedAt),
	}
	if t.CompletedAt != nil {
		p.CompletedAt = timestamppb.New(*t.CompletedAt)
	}
	return p
}

func modelTaskStatusToProto(s models.TaskStatus) pb.TaskStatus {
	switch s {
	case models.TaskStatusPending:
		return pb.TaskStatus_TASK_STATUS_PENDING
	case models.TaskStatusRunning:
		return pb.TaskStatus_TASK_STATUS_RUNNING
	case models.TaskStatusWaitingOnHuman:
		return pb.TaskStatus_TASK_STATUS_WAITING_ON_HUMAN
	case models.TaskStatusCompleted:
		return pb.TaskStatus_TASK_STATUS_COMPLETED
	case models.TaskStatusFailed:
		return pb.TaskStatus_TASK_STATUS_FAILED
	case models.TaskStatusCancelled:
		return pb.TaskStatus_TASK_STATUS_CANCELLED
	default:
		return pb.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

func protoTaskStatusToModel(s pb.TaskStatus) models.TaskStatus {
	switch s {
	case pb.TaskStatus_TASK_STATUS_PENDING:
		return models.TaskStatusPending
	case pb.TaskStatus_TASK_STATUS_RUNNING:
		return models.TaskStatusRunning
	case pb.TaskStatus_TASK_STATUS_WAITING_ON_HUMAN:
		return models.TaskStatusWaitingOnHuman
	case pb.TaskStatus_TASK_STATUS_COMPLETED:
		return models.TaskStatusCompleted
	case pb.TaskStatus_TASK_STATUS_FAILED:
		return models.TaskStatusFailed
	case pb.TaskStatus_TASK_STATUS_CANCELLED:
		return models.TaskStatusCancelled
	default:
		return ""
	}
}

func modelWSConfigToProto(c *models.TaskWorkspaceConfig) *pb.TaskWorkspaceConfig {
	return &pb.TaskWorkspaceConfig{
		IsolationTier:   c.IsolationTier,
		Persistent:      c.Persistent,
		MemoryMb:        c.MemoryMb,
		CpuMillicores:   c.CpuMillicores,
		DiskMb:          c.DiskMb,
		MaxDurationSecs: c.MaxDurationSecs,
		AllowedTools:    c.AllowedTools,
		EnvVars:         c.EnvVars,
		ContainerImage:  c.ContainerImage,
		EgressAllowlist: c.EgressAllowlist,
	}
}

func protoWSConfigToModel(c *pb.TaskWorkspaceConfig) models.TaskWorkspaceConfig {
	return models.TaskWorkspaceConfig{
		IsolationTier:   c.GetIsolationTier(),
		Persistent:      c.GetPersistent(),
		MemoryMb:        c.GetMemoryMb(),
		CpuMillicores:   c.GetCpuMillicores(),
		DiskMb:          c.GetDiskMb(),
		MaxDurationSecs: c.GetMaxDurationSecs(),
		AllowedTools:    c.GetAllowedTools(),
		EnvVars:         c.GetEnvVars(),
		ContainerImage:  c.GetContainerImage(),
		EgressAllowlist: c.GetEgressAllowlist(),
	}
}

func modelHIConfigToProto(c *models.TaskHumanInteractionConfig) *pb.HumanInteractionConfig {
	return &pb.HumanInteractionConfig{
		EscalationTargets: c.EscalationTargets,
		ApprovalTargets:   c.ApprovalTargets,
		TimeoutSecs:       c.TimeoutSecs,
		TimeoutAction:     c.TimeoutAction,
	}
}

func protoHIConfigToModel(c *pb.HumanInteractionConfig) models.TaskHumanInteractionConfig {
	return models.TaskHumanInteractionConfig{
		EscalationTargets: c.GetEscalationTargets(),
		ApprovalTargets:   c.GetApprovalTargets(),
		TimeoutSecs:       c.GetTimeoutSecs(),
		TimeoutAction:     c.GetTimeoutAction(),
	}
}

func modelBudgetConfigToProto(c *models.TaskBudgetConfig) *pb.BudgetConfig {
	return &pb.BudgetConfig{
		MaxCost:          c.MaxCost,
		WarningThreshold: c.WarningThreshold,
		OnExceeded:       c.OnExceeded,
		Currency:         c.Currency,
	}
}

func protoBudgetConfigToModel(c *pb.BudgetConfig) models.TaskBudgetConfig {
	return models.TaskBudgetConfig{
		MaxCost:          c.GetMaxCost(),
		WarningThreshold: c.GetWarningThreshold(),
		OnExceeded:       c.GetOnExceeded(),
		Currency:         c.GetCurrency(),
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, ErrTaskNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrInvalidTransition):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

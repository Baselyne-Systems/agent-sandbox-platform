package task

import (
	"context"
	"errors"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidTransition  = errors.New("invalid status transition")
)

// WorkspaceProvisioner provisions and terminates workspaces for tasks.
// Implemented as an adapter over the Workspace Service gRPC client.
type WorkspaceProvisioner interface {
	ProvisionWorkspace(ctx context.Context, task *models.Task) (workspaceID string, err error)
	TerminateWorkspace(ctx context.Context, workspaceID string, reason string) error
}

// ServiceConfig holds dependencies for the task Service.
type ServiceConfig struct {
	Repo        Repository
	Provisioner WorkspaceProvisioner // optional; nil disables workspace orchestration
}

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

// validTransitions defines allowed status transitions.
var validTransitions = map[models.TaskStatus][]models.TaskStatus{
	models.TaskStatusPending:        {models.TaskStatusRunning, models.TaskStatusCancelled, models.TaskStatusFailed},
	models.TaskStatusRunning:        {models.TaskStatusWaitingOnHuman, models.TaskStatusCompleted, models.TaskStatusFailed, models.TaskStatusCancelled},
	models.TaskStatusWaitingOnHuman: {models.TaskStatusRunning, models.TaskStatusFailed, models.TaskStatusCancelled},
}

// Service implements task business logic.
type Service struct {
	repo        Repository
	provisioner WorkspaceProvisioner
}

func NewService(cfg ServiceConfig) *Service {
	return &Service{
		repo:        cfg.Repo,
		provisioner: cfg.Provisioner,
	}
}

func (s *Service) CreateTask(ctx context.Context, tenantID, agentID, goal string, wsConfig *models.TaskWorkspaceConfig, guardrailPolicyID string, hiConfig *models.TaskHumanInteractionConfig, budgetConfig *models.TaskBudgetConfig, maxDurationWithoutCheckin int64, input, labels map[string]string) (*models.Task, error) {
	if agentID == "" {
		return nil, ErrInvalidInput
	}
	if goal == "" {
		return nil, ErrInvalidInput
	}

	task := &models.Task{
		TenantID:                      tenantID,
		AgentID:                       agentID,
		Goal:                          goal,
		Status:                        models.TaskStatusPending,
		GuardrailPolicyID:             guardrailPolicyID,
		MaxDurationWithoutCheckinSecs: maxDurationWithoutCheckin,
	}

	if wsConfig != nil {
		task.WorkspaceConfig = *wsConfig
	}
	if hiConfig != nil {
		task.HumanInteractionConfig = *hiConfig
	}
	if budgetConfig != nil {
		task.BudgetConfig = *budgetConfig
	}
	if input == nil {
		input = map[string]string{}
	}
	task.Input = input
	if labels == nil {
		labels = map[string]string{}
	}
	task.Labels = labels

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) GetTask(ctx context.Context, tenantID, id string) (*models.Task, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	task, err := s.repo.GetTask(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *Service) ListTasks(ctx context.Context, tenantID, agentID string, status models.TaskStatus, pageSize int, pageToken string) ([]models.Task, string, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	tasks, err := s.repo.ListTasks(ctx, tenantID, agentID, status, pageToken, pageSize+1)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(tasks) > pageSize {
		tasks = tasks[:pageSize]
		nextToken = tasks[pageSize-1].ID
	}

	return tasks, nextToken, nil
}

func isTerminalStatus(status models.TaskStatus) bool {
	switch status {
	case models.TaskStatusCompleted, models.TaskStatusFailed, models.TaskStatusCancelled:
		return true
	}
	return false
}

func (s *Service) UpdateTaskStatus(ctx context.Context, tenantID, id string, newStatus models.TaskStatus, reason string) (*models.Task, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}

	task, err := s.repo.GetTask(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	// Validate status transition.
	allowed, ok := validTransitions[task.Status]
	if !ok {
		return nil, ErrInvalidTransition
	}
	valid := false
	for _, s := range allowed {
		if s == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		return nil, ErrInvalidTransition
	}

	// Provision workspace on Pending → Running.
	if task.Status == models.TaskStatusPending && newStatus == models.TaskStatusRunning && s.provisioner != nil {
		wsID, provErr := s.provisioner.ProvisionWorkspace(ctx, task)
		if provErr != nil {
			// Workspace provisioning failed — transition task to Failed instead.
			_ = s.repo.UpdateTaskStatus(ctx, tenantID, id, models.TaskStatusFailed)
			return s.repo.GetTask(ctx, tenantID, id)
		}
		if err := s.repo.SetWorkspaceID(ctx, tenantID, id, wsID); err != nil {
			return nil, err
		}
		task.WorkspaceID = wsID
	}

	if err := s.repo.UpdateTaskStatus(ctx, tenantID, id, newStatus); err != nil {
		return nil, err
	}

	// Terminate workspace on terminal states (best-effort).
	if isTerminalStatus(newStatus) && task.WorkspaceID != "" && s.provisioner != nil {
		_ = s.provisioner.TerminateWorkspace(ctx, task.WorkspaceID, reason)
	}

	// Re-fetch to return updated state.
	return s.repo.GetTask(ctx, tenantID, id)
}

func (s *Service) CancelTask(ctx context.Context, tenantID, id, reason string) error {
	if id == "" {
		return ErrInvalidInput
	}

	task, err := s.repo.GetTask(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	// Can only cancel pending, running, or waiting_on_human tasks.
	switch task.Status {
	case models.TaskStatusPending, models.TaskStatusRunning, models.TaskStatusWaitingOnHuman:
		if err := s.repo.UpdateTaskStatus(ctx, tenantID, id, models.TaskStatusCancelled); err != nil {
			return err
		}
		// Terminate workspace if one exists (best-effort).
		if task.WorkspaceID != "" && s.provisioner != nil {
			_ = s.provisioner.TerminateWorkspace(ctx, task.WorkspaceID, reason)
		}
		return nil
	default:
		return ErrInvalidTransition
	}
}

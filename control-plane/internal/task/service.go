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
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTask(ctx context.Context, agentID, goal string, wsConfig *models.TaskWorkspaceConfig, guardrailPolicyID string, hiConfig *models.TaskHumanInteractionConfig, budgetConfig *models.TaskBudgetConfig, maxDurationWithoutCheckin int64, input, labels map[string]string) (*models.Task, error) {
	if agentID == "" {
		return nil, ErrInvalidInput
	}
	if goal == "" {
		return nil, ErrInvalidInput
	}

	task := &models.Task{
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

func (s *Service) GetTask(ctx context.Context, id string) (*models.Task, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *Service) ListTasks(ctx context.Context, agentID string, status models.TaskStatus, pageSize int, pageToken string) ([]models.Task, string, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	tasks, err := s.repo.ListTasks(ctx, agentID, status, pageToken, pageSize+1)
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

func (s *Service) UpdateTaskStatus(ctx context.Context, id string, newStatus models.TaskStatus, reason string) (*models.Task, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}

	task, err := s.repo.GetTask(ctx, id)
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

	if err := s.repo.UpdateTaskStatus(ctx, id, newStatus); err != nil {
		return nil, err
	}

	// Re-fetch to return updated state.
	return s.repo.GetTask(ctx, id)
}

func (s *Service) CancelTask(ctx context.Context, id, reason string) error {
	if id == "" {
		return ErrInvalidInput
	}

	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	// Can only cancel pending, running, or waiting_on_human tasks.
	switch task.Status {
	case models.TaskStatusPending, models.TaskStatusRunning, models.TaskStatusWaitingOnHuman:
		return s.repo.UpdateTaskStatus(ctx, id, models.TaskStatusCancelled)
	default:
		return ErrInvalidTransition
	}
}

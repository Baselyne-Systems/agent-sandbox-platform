package workspace

import (
	"context"
	"errors"
	"time"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
)

var (
	ErrWorkspaceNotFound       = errors.New("workspace not found")
	ErrWorkspaceAlreadyTerminal = errors.New("workspace is already in a terminal state")
	ErrInvalidInput            = errors.New("invalid input")
)

const (
	defaultPageSize       = 50
	maxPageSize           = 100
	defaultMemoryMb       = 512
	defaultCpuMillicores  = 500
	defaultDiskMb         = 1024
	defaultMaxDurationSec = 3600
)

// Service implements workspace business logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateWorkspace(ctx context.Context, agentID, taskID string, spec *models.WorkspaceSpec) (*models.Workspace, error) {
	if agentID == "" {
		return nil, ErrInvalidInput
	}

	ws := &models.Workspace{
		AgentID: agentID,
		TaskID:  taskID,
		Status:  models.WorkspaceStatusPending,
	}

	// Apply spec with defaults.
	if spec != nil {
		ws.Spec = *spec
	}
	if ws.Spec.MemoryMb <= 0 {
		ws.Spec.MemoryMb = defaultMemoryMb
	}
	if ws.Spec.CpuMillicores <= 0 {
		ws.Spec.CpuMillicores = defaultCpuMillicores
	}
	if ws.Spec.DiskMb <= 0 {
		ws.Spec.DiskMb = defaultDiskMb
	}
	if ws.Spec.MaxDurationSecs <= 0 {
		ws.Spec.MaxDurationSecs = defaultMaxDurationSec
	}
	if ws.Spec.AllowedTools == nil {
		ws.Spec.AllowedTools = []string{}
	}
	if ws.Spec.EnvVars == nil {
		ws.Spec.EnvVars = map[string]string{}
	}

	expiresAt := time.Now().Add(time.Duration(ws.Spec.MaxDurationSecs) * time.Second)
	ws.ExpiresAt = &expiresAt

	if err := s.repo.CreateWorkspace(ctx, ws); err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *Service) GetWorkspace(ctx context.Context, id string) (*models.Workspace, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	ws, err := s.repo.GetWorkspace(ctx, id)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}
	return ws, nil
}

func (s *Service) ListWorkspaces(ctx context.Context, agentID string, status models.WorkspaceStatus, pageSize int, pageToken string) ([]models.Workspace, string, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	afterID, err := decodePageToken(pageToken)
	if err != nil {
		return nil, "", ErrInvalidInput
	}

	workspaces, err := s.repo.ListWorkspaces(ctx, agentID, status, afterID, pageSize+1)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(workspaces) > pageSize {
		workspaces = workspaces[:pageSize]
		nextToken = encodePageToken(workspaces[pageSize-1].ID)
	}

	return workspaces, nextToken, nil
}

func (s *Service) TerminateWorkspace(ctx context.Context, id string, reason string) error {
	if id == "" {
		return ErrInvalidInput
	}
	return s.repo.TerminateWorkspace(ctx, id, reason)
}

func encodePageToken(id string) string {
	if id == "" {
		return ""
	}
	return id
}

func decodePageToken(token string) (string, error) {
	return token, nil
}

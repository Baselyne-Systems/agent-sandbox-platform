package workspace

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
)

// mockRepo is a hand-written in-memory Repository for testing.
type mockRepo struct {
	workspaces map[string]*models.Workspace
	nextID     int
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		workspaces: make(map[string]*models.Workspace),
	}
}

func (m *mockRepo) nextUUID() string {
	m.nextID++
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", m.nextID)
}

func (m *mockRepo) CreateWorkspace(_ context.Context, ws *models.Workspace) error {
	ws.ID = m.nextUUID()
	ws.CreatedAt = time.Now()
	ws.UpdatedAt = ws.CreatedAt
	cp := *ws
	cp.Spec.AllowedTools = copyStrings(ws.Spec.AllowedTools)
	cp.Spec.EnvVars = copyMap(ws.Spec.EnvVars)
	if ws.ExpiresAt != nil {
		t := *ws.ExpiresAt
		cp.ExpiresAt = &t
	}
	m.workspaces[ws.ID] = &cp
	return nil
}

func (m *mockRepo) GetWorkspace(_ context.Context, id string) (*models.Workspace, error) {
	ws, ok := m.workspaces[id]
	if !ok {
		return nil, nil
	}
	cp := *ws
	cp.Spec.AllowedTools = copyStrings(ws.Spec.AllowedTools)
	cp.Spec.EnvVars = copyMap(ws.Spec.EnvVars)
	if ws.ExpiresAt != nil {
		t := *ws.ExpiresAt
		cp.ExpiresAt = &t
	}
	return &cp, nil
}

func (m *mockRepo) ListWorkspaces(_ context.Context, agentID string, status models.WorkspaceStatus, afterID string, limit int) ([]models.Workspace, error) {
	var result []models.Workspace
	for _, ws := range m.workspaces {
		if agentID != "" && ws.AgentID != agentID {
			continue
		}
		if status != "" && ws.Status != status {
			continue
		}
		if afterID != "" && ws.ID <= afterID {
			continue
		}
		cp := *ws
		cp.Spec.AllowedTools = copyStrings(ws.Spec.AllowedTools)
		cp.Spec.EnvVars = copyMap(ws.Spec.EnvVars)
		result = append(result, cp)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockRepo) TerminateWorkspace(_ context.Context, id string, reason string) error {
	ws, ok := m.workspaces[id]
	if !ok {
		return ErrWorkspaceNotFound
	}
	if ws.Status == models.WorkspaceStatusTerminated || ws.Status == models.WorkspaceStatusFailed {
		return ErrWorkspaceAlreadyTerminal
	}
	ws.Status = models.WorkspaceStatusTerminated
	ws.UpdatedAt = time.Now()
	return nil
}

func copyStrings(s []string) []string {
	if s == nil {
		return nil
	}
	cp := make([]string, len(s))
	copy(cp, s)
	return cp
}

func copyMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

// --- CreateWorkspace tests ---

func TestCreateWorkspace_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	spec := &models.WorkspaceSpec{
		MemoryMb:      1024,
		CpuMillicores: 1000,
		DiskMb:        2048,
	}
	ws, err := svc.CreateWorkspace(context.Background(), "agent-1", "task-1", spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.ID == "" {
		t.Error("expected workspace ID to be set")
	}
	if ws.Status != models.WorkspaceStatusPending {
		t.Errorf("expected status pending, got %q", ws.Status)
	}
	if ws.AgentID != "agent-1" {
		t.Errorf("expected agentID 'agent-1', got %q", ws.AgentID)
	}
	if ws.ExpiresAt == nil {
		t.Error("expected expires_at to be set")
	}
}

func TestCreateWorkspace_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.CreateWorkspace(context.Background(), "", "task-1", nil)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty agentID, got: %v", err)
	}
}

func TestCreateWorkspace_DefaultSpec(t *testing.T) {
	svc := NewService(newMockRepo())
	ws, err := svc.CreateWorkspace(context.Background(), "agent-1", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Spec.MemoryMb != defaultMemoryMb {
		t.Errorf("expected default memory %d, got %d", defaultMemoryMb, ws.Spec.MemoryMb)
	}
	if ws.Spec.CpuMillicores != defaultCpuMillicores {
		t.Errorf("expected default cpu %d, got %d", defaultCpuMillicores, ws.Spec.CpuMillicores)
	}
	if ws.Spec.DiskMb != defaultDiskMb {
		t.Errorf("expected default disk %d, got %d", defaultDiskMb, ws.Spec.DiskMb)
	}
	if ws.Spec.MaxDurationSecs != defaultMaxDurationSec {
		t.Errorf("expected default max_duration %d, got %d", defaultMaxDurationSec, ws.Spec.MaxDurationSecs)
	}
}

func TestCreateWorkspace_NilCollections(t *testing.T) {
	svc := NewService(newMockRepo())
	ws, err := svc.CreateWorkspace(context.Background(), "agent-1", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Spec.AllowedTools == nil {
		t.Error("expected AllowedTools to be non-nil empty slice")
	}
	if ws.Spec.EnvVars == nil {
		t.Error("expected EnvVars to be non-nil empty map")
	}
}

// --- GetWorkspace tests ---

func TestGetWorkspace_Found(t *testing.T) {
	svc := NewService(newMockRepo())
	created, _ := svc.CreateWorkspace(context.Background(), "agent-1", "task-1", nil)
	got, err := svc.GetWorkspace(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}
}

func TestGetWorkspace_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetWorkspace(context.Background(), "nonexistent-id")
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got: %v", err)
	}
}

func TestGetWorkspace_EmptyID(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetWorkspace(context.Background(), "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty ID, got: %v", err)
	}
}

// --- ListWorkspaces tests ---

func TestListWorkspaces_WithFilters(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	svc.CreateWorkspace(ctx, "agent-1", "t1", nil)
	svc.CreateWorkspace(ctx, "agent-2", "t2", nil)
	svc.CreateWorkspace(ctx, "agent-1", "t3", nil)

	// Filter by agent
	ws, _, err := svc.ListWorkspaces(ctx, "agent-1", "", 50, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ws) != 2 {
		t.Errorf("expected 2 workspaces for agent-1, got %d", len(ws))
	}

	// All workspaces
	ws, _, err = svc.ListWorkspaces(ctx, "", "", 50, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ws) != 3 {
		t.Errorf("expected 3 workspaces total, got %d", len(ws))
	}
}

func TestListWorkspaces_Pagination(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		svc.CreateWorkspace(ctx, "agent-1", fmt.Sprintf("t%d", i), nil)
	}

	ws, nextToken, err := svc.ListWorkspaces(ctx, "", "", 3, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ws) != 3 {
		t.Fatalf("expected 3 workspaces, got %d", len(ws))
	}
	if nextToken == "" {
		t.Error("expected a next page token")
	}

	ws2, nextToken2, err := svc.ListWorkspaces(ctx, "", "", 3, nextToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ws2) != 2 {
		t.Fatalf("expected 2 workspaces on second page, got %d", len(ws2))
	}
	if nextToken2 != "" {
		t.Error("expected no next page token on last page")
	}
}

func TestListWorkspaces_DefaultPageSize(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	svc.CreateWorkspace(ctx, "agent-1", "t1", nil)

	// page_size=0 should use default
	ws, _, err := svc.ListWorkspaces(ctx, "", "", 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ws) != 1 {
		t.Errorf("expected 1 workspace, got %d", len(ws))
	}
}

// --- TerminateWorkspace tests ---

func TestTerminateWorkspace_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	ws, _ := svc.CreateWorkspace(ctx, "agent-1", "t1", nil)

	if err := svc.TerminateWorkspace(ctx, ws.ID, "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetWorkspace(ctx, ws.ID)
	if got.Status != models.WorkspaceStatusTerminated {
		t.Errorf("expected status terminated, got %q", got.Status)
	}
}

func TestTerminateWorkspace_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	err := svc.TerminateWorkspace(context.Background(), "nonexistent", "reason")
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got: %v", err)
	}
}

func TestTerminateWorkspace_AlreadyTerminal(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	ws, _ := svc.CreateWorkspace(ctx, "agent-1", "t1", nil)

	svc.TerminateWorkspace(ctx, ws.ID, "first")

	err := svc.TerminateWorkspace(ctx, ws.ID, "second")
	if !errors.Is(err, ErrWorkspaceAlreadyTerminal) {
		t.Errorf("expected ErrWorkspaceAlreadyTerminal, got: %v", err)
	}
}

func TestTerminateWorkspace_EmptyID(t *testing.T) {
	svc := NewService(newMockRepo())
	err := svc.TerminateWorkspace(context.Background(), "", "reason")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty ID, got: %v", err)
	}
}

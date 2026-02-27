package task

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// mockRepo is a hand-written in-memory Repository for testing.
type mockRepo struct {
	tasks  map[string]*models.Task
	nextID int
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		tasks: make(map[string]*models.Task),
	}
}

func (m *mockRepo) nextUUID() string {
	m.nextID++
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", m.nextID)
}

func (m *mockRepo) CreateTask(_ context.Context, task *models.Task) error {
	task.ID = m.nextUUID()
	task.CreatedAt = time.Now()
	task.UpdatedAt = task.CreatedAt
	cp := *task
	cp.Input = copyMap(task.Input)
	cp.Labels = copyMap(task.Labels)
	m.tasks[task.ID] = &cp
	return nil
}

func (m *mockRepo) GetTask(_ context.Context, id string) (*models.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return nil, nil
	}
	cp := *t
	cp.Input = copyMap(t.Input)
	cp.Labels = copyMap(t.Labels)
	if t.CompletedAt != nil {
		ca := *t.CompletedAt
		cp.CompletedAt = &ca
	}
	return &cp, nil
}

func (m *mockRepo) ListTasks(_ context.Context, agentID string, status models.TaskStatus, afterID string, limit int) ([]models.Task, error) {
	var result []models.Task
	for _, t := range m.tasks {
		if agentID != "" && t.AgentID != agentID {
			continue
		}
		if status != "" && t.Status != status {
			continue
		}
		if afterID != "" && t.ID <= afterID {
			continue
		}
		cp := *t
		cp.Input = copyMap(t.Input)
		cp.Labels = copyMap(t.Labels)
		result = append(result, cp)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockRepo) UpdateTaskStatus(_ context.Context, id string, status models.TaskStatus) error {
	t, ok := m.tasks[id]
	if !ok {
		return ErrTaskNotFound
	}
	t.Status = status
	t.UpdatedAt = time.Now()
	if status == models.TaskStatusCompleted || status == models.TaskStatusFailed || status == models.TaskStatusCancelled {
		now := time.Now()
		t.CompletedAt = &now
	}
	return nil
}

func (m *mockRepo) SetWorkspaceID(_ context.Context, taskID, workspaceID string) error {
	t, ok := m.tasks[taskID]
	if !ok {
		return ErrTaskNotFound
	}
	t.WorkspaceID = workspaceID
	t.UpdatedAt = time.Now()
	return nil
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

func TestCreateTask_Success(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	task, err := svc.CreateTask(context.Background(), "agent-1", "Process invoices", nil, "policy-1", nil, nil, 300, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID == "" {
		t.Error("expected task ID to be set")
	}
	if task.AgentID != "agent-1" {
		t.Errorf("expected agent_id 'agent-1', got %q", task.AgentID)
	}
	if task.Goal != "Process invoices" {
		t.Errorf("expected goal 'Process invoices', got %q", task.Goal)
	}
	if task.Status != models.TaskStatusPending {
		t.Errorf("expected status pending, got %q", task.Status)
	}
	if task.GuardrailPolicyID != "policy-1" {
		t.Errorf("expected guardrail_policy_id 'policy-1', got %q", task.GuardrailPolicyID)
	}
	if task.Input == nil {
		t.Error("expected input to be initialized")
	}
	if task.Labels == nil {
		t.Error("expected labels to be initialized")
	}
}

func TestCreateTask_WithConfigs(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	wsConfig := &models.TaskWorkspaceConfig{
		IsolationTier: "standard",
		MemoryMb:      512,
		CpuMillicores: 1000,
		DiskMb:        2048,
		AllowedTools:  []string{"bash", "curl"},
		EnvVars:       map[string]string{"ENV": "prod"},
	}
	hiConfig := &models.TaskHumanInteractionConfig{
		EscalationTargets: []string{"admin@example.com"},
		TimeoutSecs:       600,
		TimeoutAction:     "escalate",
	}
	budgetConfig := &models.TaskBudgetConfig{
		MaxCost:          10.0,
		WarningThreshold: 8.0,
		OnExceeded:       "halt",
		Currency:         "USD",
	}

	task, err := svc.CreateTask(context.Background(), "agent-1", "goal",
		wsConfig, "policy-1", hiConfig, budgetConfig, 300,
		map[string]string{"key": "value"}, map[string]string{"env": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.WorkspaceConfig.IsolationTier != "standard" {
		t.Errorf("expected isolation_tier 'standard', got %q", task.WorkspaceConfig.IsolationTier)
	}
	if task.WorkspaceConfig.MemoryMb != 512 {
		t.Errorf("expected memory_mb 512, got %d", task.WorkspaceConfig.MemoryMb)
	}
	if task.HumanInteractionConfig.TimeoutSecs != 600 {
		t.Errorf("expected timeout_secs 600, got %d", task.HumanInteractionConfig.TimeoutSecs)
	}
	if task.BudgetConfig.MaxCost != 10.0 {
		t.Errorf("expected max_cost 10.0, got %f", task.BudgetConfig.MaxCost)
	}
	if task.Input["key"] != "value" {
		t.Errorf("expected input[key]='value', got %q", task.Input["key"])
	}
	if task.Labels["env"] != "test" {
		t.Errorf("expected labels[env]='test', got %q", task.Labels["env"])
	}
}

func TestCreateTask_Validation(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	if _, err := svc.CreateTask(ctx, "", "goal", nil, "", nil, nil, 0, nil, nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty agent_id, got: %v", err)
	}
	if _, err := svc.CreateTask(ctx, "agent", "", nil, "", nil, nil, 0, nil, nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty goal, got: %v", err)
	}
}

func TestGetTask_Found(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	created, _ := svc.CreateTask(context.Background(), "a", "g", nil, "", nil, nil, 0, nil, nil)
	got, err := svc.GetTask(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	_, err := svc.GetTask(context.Background(), "nonexistent-id")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got: %v", err)
	}
}

func TestListTasks_Pagination(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		svc.CreateTask(ctx, "agent-1", fmt.Sprintf("goal-%d", i), nil, "", nil, nil, 0, nil, nil)
	}

	tasks, nextToken, err := svc.ListTasks(ctx, "", "", 3, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}
	if nextToken == "" {
		t.Error("expected a next page token")
	}

	tasks2, nextToken2, err := svc.ListTasks(ctx, "", "", 3, nextToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks2) != 2 {
		t.Fatalf("expected 2 tasks on second page, got %d", len(tasks2))
	}
	if nextToken2 != "" {
		t.Error("expected no next page token on last page")
	}
}

func TestListTasks_Filters(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	svc.CreateTask(ctx, "agent-1", "g1", nil, "", nil, nil, 0, nil, nil)
	svc.CreateTask(ctx, "agent-2", "g2", nil, "", nil, nil, 0, nil, nil)
	svc.CreateTask(ctx, "agent-1", "g3", nil, "", nil, nil, 0, nil, nil)

	// Filter by agent
	tasks, _, err := svc.ListTasks(ctx, "agent-1", "", 50, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks for agent-1, got %d", len(tasks))
	}
}

func TestUpdateTaskStatus_ValidTransitions(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	// pending → running
	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	updated, err := svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")
	if err != nil {
		t.Fatalf("pending→running: %v", err)
	}
	if updated.Status != models.TaskStatusRunning {
		t.Errorf("expected running, got %q", updated.Status)
	}

	// running → waiting_on_human
	updated, err = svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusWaitingOnHuman, "")
	if err != nil {
		t.Fatalf("running→waiting_on_human: %v", err)
	}
	if updated.Status != models.TaskStatusWaitingOnHuman {
		t.Errorf("expected waiting_on_human, got %q", updated.Status)
	}

	// waiting_on_human → running
	updated, err = svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")
	if err != nil {
		t.Fatalf("waiting_on_human→running: %v", err)
	}
	if updated.Status != models.TaskStatusRunning {
		t.Errorf("expected running, got %q", updated.Status)
	}

	// running → completed
	updated, err = svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusCompleted, "")
	if err != nil {
		t.Fatalf("running→completed: %v", err)
	}
	if updated.Status != models.TaskStatusCompleted {
		t.Errorf("expected completed, got %q", updated.Status)
	}
	if updated.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestUpdateTaskStatus_InvalidTransitions(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	// pending → completed (not allowed)
	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	_, err := svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusCompleted, "")
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition for pending→completed, got: %v", err)
	}

	// Move to completed, then try to transition again
	svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")
	svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusCompleted, "")

	_, err = svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition for completed→running, got: %v", err)
	}
}

func TestUpdateTaskStatus_NotFound(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	_, err := svc.UpdateTaskStatus(context.Background(), "nonexistent", models.TaskStatusRunning, "")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got: %v", err)
	}
}

func TestCancelTask_Success(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	// Cancel from pending
	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	if err := svc.CancelTask(ctx, task.ID, "no longer needed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetTask(ctx, task.ID)
	if got.Status != models.TaskStatusCancelled {
		t.Errorf("expected cancelled, got %q", got.Status)
	}

	// Cancel from running
	task2, _ := svc.CreateTask(ctx, "a", "g2", nil, "", nil, nil, 0, nil, nil)
	svc.UpdateTaskStatus(ctx, task2.ID, models.TaskStatusRunning, "")
	if err := svc.CancelTask(ctx, task2.ID, "timeout"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got2, _ := svc.GetTask(ctx, task2.ID)
	if got2.Status != models.TaskStatusCancelled {
		t.Errorf("expected cancelled, got %q", got2.Status)
	}
}

func TestCancelTask_InvalidState(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")
	svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusCompleted, "")

	err := svc.CancelTask(ctx, task.ID, "too late")
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition for cancelling completed task, got: %v", err)
	}
}

func TestCancelTask_NotFound(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	err := svc.CancelTask(context.Background(), "no-such-id", "")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got: %v", err)
	}
}

// --- Workspace orchestration tests ---

type mockProvisioner struct {
	provisionCalls   int
	terminateCalls   int
	provisionErr     error
	terminateErr     error
	returnWSID       string
	lastTerminateID  string
	lastTerminateReason string
}

func (m *mockProvisioner) ProvisionWorkspace(_ context.Context, _ *models.Task) (string, error) {
	m.provisionCalls++
	if m.provisionErr != nil {
		return "", m.provisionErr
	}
	return m.returnWSID, nil
}

func (m *mockProvisioner) TerminateWorkspace(_ context.Context, workspaceID, reason string) error {
	m.terminateCalls++
	m.lastTerminateID = workspaceID
	m.lastTerminateReason = reason
	return m.terminateErr
}

func TestPendingToRunning_ProvisionWorkspace(t *testing.T) {
	prov := &mockProvisioner{returnWSID: "ws-123"}
	svc := NewService(ServiceConfig{Repo: newMockRepo(), Provisioner: prov})
	ctx := context.Background()

	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	updated, err := svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != models.TaskStatusRunning {
		t.Errorf("expected running, got %q", updated.Status)
	}
	if updated.WorkspaceID != "ws-123" {
		t.Errorf("expected workspace_id 'ws-123', got %q", updated.WorkspaceID)
	}
	if prov.provisionCalls != 1 {
		t.Errorf("expected 1 provision call, got %d", prov.provisionCalls)
	}
}

func TestPendingToRunning_ProvisionFailure(t *testing.T) {
	prov := &mockProvisioner{provisionErr: errors.New("no capacity")}
	svc := NewService(ServiceConfig{Repo: newMockRepo(), Provisioner: prov})
	ctx := context.Background()

	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	updated, err := svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != models.TaskStatusFailed {
		t.Errorf("expected failed after provision error, got %q", updated.Status)
	}
	if updated.WorkspaceID != "" {
		t.Errorf("expected empty workspace_id, got %q", updated.WorkspaceID)
	}
}

func TestRunningToCompleted_TerminateWorkspace(t *testing.T) {
	prov := &mockProvisioner{returnWSID: "ws-456"}
	svc := NewService(ServiceConfig{Repo: newMockRepo(), Provisioner: prov})
	ctx := context.Background()

	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")

	updated, err := svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusCompleted, "done")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != models.TaskStatusCompleted {
		t.Errorf("expected completed, got %q", updated.Status)
	}
	if prov.terminateCalls != 1 {
		t.Errorf("expected 1 terminate call, got %d", prov.terminateCalls)
	}
	if prov.lastTerminateID != "ws-456" {
		t.Errorf("expected terminate workspace_id 'ws-456', got %q", prov.lastTerminateID)
	}
}

func TestRunningToFailed_TerminateWorkspace(t *testing.T) {
	prov := &mockProvisioner{returnWSID: "ws-789"}
	svc := NewService(ServiceConfig{Repo: newMockRepo(), Provisioner: prov})
	ctx := context.Background()

	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")

	_, err := svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusFailed, "error occurred")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov.terminateCalls != 1 {
		t.Errorf("expected 1 terminate call, got %d", prov.terminateCalls)
	}
	if prov.lastTerminateReason != "error occurred" {
		t.Errorf("expected reason 'error occurred', got %q", prov.lastTerminateReason)
	}
}

func TestWaitingToRunning_NoProvision(t *testing.T) {
	prov := &mockProvisioner{returnWSID: "ws-aaa"}
	svc := NewService(ServiceConfig{Repo: newMockRepo(), Provisioner: prov})
	ctx := context.Background()

	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")
	svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusWaitingOnHuman, "")

	// Reset counters after initial provision.
	prov.provisionCalls = 0

	_, err := svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "human responded")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov.provisionCalls != 0 {
		t.Errorf("expected 0 provision calls for WaitingOnHuman→Running, got %d", prov.provisionCalls)
	}
}

func TestNilProvisioner_NoOrchestration(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	updated, err := svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != models.TaskStatusRunning {
		t.Errorf("expected running, got %q", updated.Status)
	}
	if updated.WorkspaceID != "" {
		t.Errorf("expected empty workspace_id with nil provisioner, got %q", updated.WorkspaceID)
	}
}

func TestTerminateError_StillTransitions(t *testing.T) {
	prov := &mockProvisioner{returnWSID: "ws-bbb", terminateErr: errors.New("terminate failed")}
	svc := NewService(ServiceConfig{Repo: newMockRepo(), Provisioner: prov})
	ctx := context.Background()

	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")

	updated, err := svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusCompleted, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != models.TaskStatusCompleted {
		t.Errorf("expected completed despite terminate error, got %q", updated.Status)
	}
}

func TestCancelRunning_TerminatesWorkspace(t *testing.T) {
	prov := &mockProvisioner{returnWSID: "ws-ccc"}
	svc := NewService(ServiceConfig{Repo: newMockRepo(), Provisioner: prov})
	ctx := context.Background()

	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")

	// Reset after provision.
	prov.terminateCalls = 0

	err := svc.CancelTask(ctx, task.ID, "user cancelled")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov.terminateCalls != 1 {
		t.Errorf("expected 1 terminate call on cancel, got %d", prov.terminateCalls)
	}
	if prov.lastTerminateReason != "user cancelled" {
		t.Errorf("expected reason 'user cancelled', got %q", prov.lastTerminateReason)
	}
}

func TestCancelPending_NoTerminate(t *testing.T) {
	prov := &mockProvisioner{}
	svc := NewService(ServiceConfig{Repo: newMockRepo(), Provisioner: prov})
	ctx := context.Background()

	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	err := svc.CancelTask(ctx, task.ID, "changed mind")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prov.terminateCalls != 0 {
		t.Errorf("expected 0 terminate calls for pending cancel, got %d", prov.terminateCalls)
	}
}

func TestCancelNilProvisioner_StillCancels(t *testing.T) {
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	task, _ := svc.CreateTask(ctx, "a", "g", nil, "", nil, nil, 0, nil, nil)
	svc.UpdateTaskStatus(ctx, task.ID, models.TaskStatusRunning, "")

	err := svc.CancelTask(ctx, task.ID, "no provisioner")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetTask(ctx, task.ID)
	if got.Status != models.TaskStatusCancelled {
		t.Errorf("expected cancelled, got %q", got.Status)
	}
}

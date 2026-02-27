package workspace

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
	runtimepb "github.com/baselyne/agent-sandbox-platform/control-plane/pkg/gen/runtime/v1"
	"google.golang.org/grpc"
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

func (m *mockRepo) UpdateWorkspaceStatus(_ context.Context, id string, status models.WorkspaceStatus, hostID, hostAddress, sandboxID string) error {
	ws, ok := m.workspaces[id]
	if !ok {
		return ErrWorkspaceNotFound
	}
	ws.Status = status
	if hostID != "" {
		ws.HostID = hostID
	}
	if hostAddress != "" {
		ws.HostAddress = hostAddress
	}
	if sandboxID != "" {
		ws.SandboxID = sandboxID
	}
	ws.UpdatedAt = time.Now()
	return nil
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

// --- Mock orchestration dependencies ---

type mockComputePlacer struct {
	hostID      string
	hostAddress string
	err         error
}

func (m *mockComputePlacer) PlaceWorkspace(_ context.Context, _ int64, _ int32, _ int64) (string, string, error) {
	return m.hostID, m.hostAddress, m.err
}

type mockPolicyCompiler struct {
	compiled []byte
	count    int
	err      error
}

func (m *mockPolicyCompiler) CompilePolicy(_ context.Context, _ []string) ([]byte, int, error) {
	return m.compiled, m.count, m.err
}

// mockRuntimeServiceClient implements runtimepb.RuntimeServiceClient for testing.
type mockRuntimeServiceClient struct {
	createResp   *runtimepb.CreateSandboxResponse
	createErr    error
	destroyResp  *runtimepb.DestroySandboxResponse
	destroyErr   error
	destroyCalls int
}

func (m *mockRuntimeServiceClient) CreateSandbox(_ context.Context, _ *runtimepb.CreateSandboxRequest, _ ...grpc.CallOption) (*runtimepb.CreateSandboxResponse, error) {
	return m.createResp, m.createErr
}

func (m *mockRuntimeServiceClient) DestroySandbox(_ context.Context, _ *runtimepb.DestroySandboxRequest, _ ...grpc.CallOption) (*runtimepb.DestroySandboxResponse, error) {
	m.destroyCalls++
	return m.destroyResp, m.destroyErr
}

func (m *mockRuntimeServiceClient) GetSandboxStatus(_ context.Context, _ *runtimepb.GetSandboxStatusRequest, _ ...grpc.CallOption) (*runtimepb.GetSandboxStatusResponse, error) {
	return nil, nil
}

func (m *mockRuntimeServiceClient) StreamEvents(_ context.Context, _ *runtimepb.StreamEventsRequest, _ ...grpc.CallOption) (grpc.ServerStreamingClient[runtimepb.SandboxEvent], error) {
	return nil, nil
}

// newTestService creates a Service with only the repo (no orchestration).
func newTestService(repo Repository) *Service {
	return NewService(ServiceConfig{Repo: repo})
}

// newOrchestratedService creates a Service with full orchestration mocks.
func newOrchestratedService(repo Repository, compute ComputePlacer, guardrails PolicyCompiler, runtimeClient *mockRuntimeServiceClient) *Service {
	dialer := func(_ context.Context, _ string) (runtimepb.RuntimeServiceClient, error) {
		return runtimeClient, nil
	}
	return NewService(ServiceConfig{
		Repo:        repo,
		Compute:     compute,
		Guardrails:  guardrails,
		DialRuntime: dialer,
	})
}

// --- CreateWorkspace tests ---

func TestCreateWorkspace_Success(t *testing.T) {
	svc := newTestService(newMockRepo())
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
	svc := newTestService(newMockRepo())
	_, err := svc.CreateWorkspace(context.Background(), "", "task-1", nil)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty agentID, got: %v", err)
	}
}

func TestCreateWorkspace_DefaultSpec(t *testing.T) {
	svc := newTestService(newMockRepo())
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
	svc := newTestService(newMockRepo())
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

// --- Orchestrated CreateWorkspace tests ---

func TestCreateWorkspace_WithOrchestration(t *testing.T) {
	repo := newMockRepo()
	compute := &mockComputePlacer{hostID: "host-1", hostAddress: "runtime.host1:50052"}
	guardrails := &mockPolicyCompiler{compiled: []byte(`{"rules":[]}`), count: 0}
	runtimeClient := &mockRuntimeServiceClient{
		createResp: &runtimepb.CreateSandboxResponse{
			SandboxId:        "sandbox-abc",
			AgentApiEndpoint: "localhost:50052",
		},
	}

	svc := newOrchestratedService(repo, compute, guardrails, runtimeClient)
	ctx := context.Background()

	ws, err := svc.CreateWorkspace(ctx, "agent-1", "task-1", &models.WorkspaceSpec{
		MemoryMb:      1024,
		CpuMillicores: 1000,
		DiskMb:        2048,
		AllowedTools:  []string{"read_file"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ws.Status != models.WorkspaceStatusRunning {
		t.Errorf("expected status running, got %q", ws.Status)
	}
	if ws.HostID != "host-1" {
		t.Errorf("expected host_id 'host-1', got %q", ws.HostID)
	}
	if ws.SandboxID != "sandbox-abc" {
		t.Errorf("expected sandbox_id 'sandbox-abc', got %q", ws.SandboxID)
	}

	// Verify the workspace is also updated in the repo.
	got, _ := svc.GetWorkspace(ctx, ws.ID)
	if got.Status != models.WorkspaceStatusRunning {
		t.Errorf("expected repo status running, got %q", got.Status)
	}
	if got.SandboxID != "sandbox-abc" {
		t.Errorf("expected repo sandbox_id 'sandbox-abc', got %q", got.SandboxID)
	}
}

func TestCreateWorkspace_PlacementFailure(t *testing.T) {
	repo := newMockRepo()
	compute := &mockComputePlacer{err: errors.New("no capacity")}
	runtimeClient := &mockRuntimeServiceClient{}

	svc := newOrchestratedService(repo, compute, nil, runtimeClient)
	ctx := context.Background()

	ws, err := svc.CreateWorkspace(ctx, "agent-1", "task-1", nil)
	if err != nil {
		t.Fatalf("unexpected error (should not propagate): %v", err)
	}
	if ws.Status != models.WorkspaceStatusFailed {
		t.Errorf("expected status failed after placement failure, got %q", ws.Status)
	}
}

func TestCreateWorkspace_RuntimeFailure(t *testing.T) {
	repo := newMockRepo()
	compute := &mockComputePlacer{hostID: "host-1", hostAddress: "runtime.host1:50052"}
	runtimeClient := &mockRuntimeServiceClient{createErr: errors.New("connection refused")}

	svc := newOrchestratedService(repo, compute, nil, runtimeClient)
	ctx := context.Background()

	ws, err := svc.CreateWorkspace(ctx, "agent-1", "task-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Status != models.WorkspaceStatusFailed {
		t.Errorf("expected status failed after runtime failure, got %q", ws.Status)
	}
}

func TestCreateWorkspace_WithGuardrails(t *testing.T) {
	repo := newMockRepo()
	compute := &mockComputePlacer{hostID: "host-1", hostAddress: "runtime.host1:50052"}
	guardrails := &mockPolicyCompiler{
		compiled: []byte(`{"rules":[{"id":"r1","name":"deny-exec","rule_type":"tool_filter","condition":"exec","action":"deny","priority":10,"enabled":true}]}`),
		count:    1,
	}
	runtimeClient := &mockRuntimeServiceClient{
		createResp: &runtimepb.CreateSandboxResponse{
			SandboxId:        "sandbox-def",
			AgentApiEndpoint: "localhost:50052",
		},
	}

	svc := newOrchestratedService(repo, compute, guardrails, runtimeClient)
	ctx := context.Background()

	ws, err := svc.CreateWorkspace(ctx, "agent-1", "task-1", &models.WorkspaceSpec{
		GuardrailPolicyID: "rule-1,rule-2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Status != models.WorkspaceStatusRunning {
		t.Errorf("expected status running, got %q", ws.Status)
	}
}

// --- GetWorkspace tests ---

func TestGetWorkspace_Found(t *testing.T) {
	svc := newTestService(newMockRepo())
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
	svc := newTestService(newMockRepo())
	_, err := svc.GetWorkspace(context.Background(), "nonexistent-id")
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got: %v", err)
	}
}

func TestGetWorkspace_EmptyID(t *testing.T) {
	svc := newTestService(newMockRepo())
	_, err := svc.GetWorkspace(context.Background(), "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty ID, got: %v", err)
	}
}

// --- ListWorkspaces tests ---

func TestListWorkspaces_WithFilters(t *testing.T) {
	svc := newTestService(newMockRepo())
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
	svc := newTestService(newMockRepo())
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
	svc := newTestService(newMockRepo())
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
	svc := newTestService(newMockRepo())
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
	svc := newTestService(newMockRepo())
	err := svc.TerminateWorkspace(context.Background(), "nonexistent", "reason")
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got: %v", err)
	}
}

func TestTerminateWorkspace_AlreadyTerminal(t *testing.T) {
	svc := newTestService(newMockRepo())
	ctx := context.Background()
	ws, _ := svc.CreateWorkspace(ctx, "agent-1", "t1", nil)

	svc.TerminateWorkspace(ctx, ws.ID, "first")

	err := svc.TerminateWorkspace(ctx, ws.ID, "second")
	if !errors.Is(err, ErrWorkspaceAlreadyTerminal) {
		t.Errorf("expected ErrWorkspaceAlreadyTerminal, got: %v", err)
	}
}

func TestTerminateWorkspace_EmptyID(t *testing.T) {
	svc := newTestService(newMockRepo())
	err := svc.TerminateWorkspace(context.Background(), "", "reason")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty ID, got: %v", err)
	}
}

func TestTerminateWorkspace_WithSandboxTeardown(t *testing.T) {
	repo := newMockRepo()
	compute := &mockComputePlacer{hostID: "host-1", hostAddress: "runtime.host1:50052"}
	runtimeClient := &mockRuntimeServiceClient{
		createResp: &runtimepb.CreateSandboxResponse{
			SandboxId:        "sandbox-xyz",
			AgentApiEndpoint: "localhost:50052",
		},
		destroyResp: &runtimepb.DestroySandboxResponse{},
	}

	svc := newOrchestratedService(repo, compute, nil, runtimeClient)
	ctx := context.Background()

	ws, _ := svc.CreateWorkspace(ctx, "agent-1", "task-1", nil)
	if ws.Status != models.WorkspaceStatusRunning {
		t.Fatalf("expected running, got %q", ws.Status)
	}

	if err := svc.TerminateWorkspace(ctx, ws.ID, "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetWorkspace(ctx, ws.ID)
	if got.Status != models.WorkspaceStatusTerminated {
		t.Errorf("expected status terminated, got %q", got.Status)
	}
	if runtimeClient.destroyCalls != 1 {
		t.Errorf("expected 1 destroy call, got %d", runtimeClient.destroyCalls)
	}
}

package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/activity"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/compute"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/economics"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/governance"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/guardrails"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/human"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/identity"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/task"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/workspace"
	hostagentpb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/host_agent/v1"
)

// ---------------------------------------------------------------------------
// Shared test state (initialized in TestMain)
// ---------------------------------------------------------------------------

var (
	db            *sql.DB
	identitySvc   *identity.Service
	taskSvc       *task.Service
	workspaceSvc  *workspace.Service
	computeSvc    *compute.Service
	guardrailsSvc *guardrails.Service
	economicsSvc  *economics.Service
	humanSvc      *human.Service
	activitySvc   *activity.Service
	governanceSvc *governance.Service
	computeRepo   compute.Repository

	fakeHostAgent *fakeHostAgentClient
	fakeSnapshots *fakeSnapshotStore

	tenantCounter atomic.Int64
)

// ---------------------------------------------------------------------------
// TestMain — start PostgreSQL, run migrations, wire all services
// ---------------------------------------------------------------------------

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Start PostgreSQL container.
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres container: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = container.Terminate(ctx) }()

	host, err := container.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get container host: %v\n", err)
		os.Exit(1)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		fmt.Fprintf(os.Stderr, "get container port: %v\n", err)
		os.Exit(1)
	}

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "ping db: %v\n", err)
		os.Exit(1)
	}

	// 2. Run migrations.
	runMigrations(db)

	// 3. Create all real repositories.
	identityRepo := identity.NewPostgresRepository(db)
	taskRepo := task.NewPostgresRepository(db)
	workspaceRepo := workspace.NewPostgresRepository(db)
	cRepo := compute.NewPostgresRepository(db)
	computeRepo = cRepo
	guardrailsRepo := guardrails.NewPostgresRepository(db)
	economicsRepo := economics.NewPostgresRepository(db)
	humanRepo := human.NewPostgresRepository(db)
	activityRepo := activity.NewPostgresRepository(db)

	// 4. Create all real services.
	identitySvc = identity.NewService(identityRepo)
	computeSvc = compute.NewService(cRepo)
	guardrailsSvc = guardrails.NewService(guardrailsRepo)
	economicsSvc = economics.NewService(economicsRepo)
	activitySvc = activity.NewService(activityRepo)
	governanceSvc = governance.NewService()
	humanSvc = human.NewService(humanRepo)

	// 5. Create mocks.
	fakeHostAgent = newFakeHostAgentClient()
	fakeSnapshots = newFakeSnapshotStore()

	// 6. Wire workspace service with all dependencies.
	identityCredAdapter := &identityCredentialAdapter{svc: identitySvc}
	identityQueryAdapter := &identityQuerierAdapter{db: db}

	workspaceSvc = workspace.NewService(workspace.ServiceConfig{
		Repo:       workspaceRepo,
		Compute:    computeSvc,
		Guardrails: guardrailsSvc,
		DialHostAgent: func(_ context.Context, _ string) (hostagentpb.HostAgentServiceClient, error) {
			return fakeHostAgent, nil
		},
		Snapshots:   fakeSnapshots,
		Credentials: identityCredAdapter,
		Identity:    identityQueryAdapter,
	})

	// 7. Wire task service with workspace provisioner.
	wsProvisioner := &workspaceProvisionerAdapter{wsSvc: workspaceSvc}
	taskSvc = task.NewService(task.ServiceConfig{
		Repo:        taskRepo,
		Provisioner: wsProvisioner,
	})

	// 8. Wire human service with activity logger.
	activityLogAdapter := &activityLoggerAdapter{svc: activitySvc}
	humanSvc.SetActivityLogger(activityLogAdapter)

	// 9. Run tests.
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// runMigrations reads SQL files from ../migrations/ in order.
// ---------------------------------------------------------------------------

func runMigrations(db *sql.DB) {
	migrationsDir := filepath.Join("..", "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		panic(fmt.Sprintf("read migrations dir %s: %v", migrationsDir, err))
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			panic(fmt.Sprintf("read migration %s: %v", f, err))
		}
		if _, err := db.Exec(string(data)); err != nil {
			panic(fmt.Sprintf("exec migration %s: %v", f, err))
		}
	}
}

// ---------------------------------------------------------------------------
// fakeHostAgentClient — implements hostagentpb.HostAgentServiceClient
// ---------------------------------------------------------------------------

type fakeHostAgentClient struct {
	createCalls   atomic.Int64
	destroyCalls  atomic.Int64
	failCreate    atomic.Bool
	mu            sync.Mutex
	lastCreateReq *hostagentpb.CreateSandboxRequest
}

func newFakeHostAgentClient() *fakeHostAgentClient {
	return &fakeHostAgentClient{}
}

func (f *fakeHostAgentClient) reset() {
	f.createCalls.Store(0)
	f.destroyCalls.Store(0)
	f.failCreate.Store(false)
	f.mu.Lock()
	f.lastCreateReq = nil
	f.mu.Unlock()
}

func (f *fakeHostAgentClient) CreateSandbox(_ context.Context, in *hostagentpb.CreateSandboxRequest, _ ...grpc.CallOption) (*hostagentpb.CreateSandboxResponse, error) {
	f.createCalls.Add(1)
	f.mu.Lock()
	f.lastCreateReq = in
	f.mu.Unlock()

	if f.failCreate.Load() {
		return nil, fmt.Errorf("fake: create sandbox failed")
	}

	return &hostagentpb.CreateSandboxResponse{
		SandboxId:        fmt.Sprintf("sbx-%s", in.WorkspaceId),
		AgentApiEndpoint: "localhost:9090",
	}, nil
}

func (f *fakeHostAgentClient) DestroySandbox(_ context.Context, _ *hostagentpb.DestroySandboxRequest, _ ...grpc.CallOption) (*hostagentpb.DestroySandboxResponse, error) {
	f.destroyCalls.Add(1)
	return &hostagentpb.DestroySandboxResponse{}, nil
}

func (f *fakeHostAgentClient) GetSandboxStatus(_ context.Context, _ *hostagentpb.GetSandboxStatusRequest, _ ...grpc.CallOption) (*hostagentpb.GetSandboxStatusResponse, error) {
	return &hostagentpb.GetSandboxStatusResponse{}, nil
}

func (f *fakeHostAgentClient) StreamEvents(_ context.Context, _ *hostagentpb.StreamEventsRequest, _ ...grpc.CallOption) (grpc.ServerStreamingClient[hostagentpb.SandboxEvent], error) {
	return nil, fmt.Errorf("not implemented in fake")
}

func (f *fakeHostAgentClient) UpdateSandboxGuardrails(_ context.Context, _ *hostagentpb.UpdateSandboxGuardrailsRequest, _ ...grpc.CallOption) (*hostagentpb.UpdateSandboxGuardrailsResponse, error) {
	return &hostagentpb.UpdateSandboxGuardrailsResponse{}, nil
}

// ---------------------------------------------------------------------------
// fakeSnapshotStore — implements workspace.SnapshotStore
// ---------------------------------------------------------------------------

type fakeSnapshotStore struct {
	mu        sync.Mutex
	snapshots map[string]string // snapshotID -> workspaceID
	counter   atomic.Int64
}

func newFakeSnapshotStore() *fakeSnapshotStore {
	return &fakeSnapshotStore{
		snapshots: make(map[string]string),
	}
}

func (f *fakeSnapshotStore) reset() {
	f.mu.Lock()
	f.snapshots = make(map[string]string)
	f.mu.Unlock()
}

func (f *fakeSnapshotStore) SaveSnapshot(_ context.Context, workspaceID string) (string, error) {
	id := fmt.Sprintf("snap-%d", f.counter.Add(1))
	f.mu.Lock()
	f.snapshots[id] = workspaceID
	f.mu.Unlock()
	return id, nil
}

func (f *fakeSnapshotStore) LoadSnapshot(_ context.Context, snapshotID string) error {
	// In the real flow, SaveSnapshot returns an ID (e.g. "snap-1") but
	// CreateSnapshot in the DB overwrites it with a UUID.  RestoreWorkspace
	// then calls LoadSnapshot with the DB UUID.  Since the fake store only
	// has the "snap-N" key, we accept any ID that maps to a known workspace
	// or whose "snap-N" counterpart exists — in tests, just succeed.
	return nil
}

// ---------------------------------------------------------------------------
// Adapters
// ---------------------------------------------------------------------------

// identityCredentialAdapter wraps identitySvc.MintCredential -> workspace.CredentialMinter.
type identityCredentialAdapter struct {
	svc *identity.Service
}

func (a *identityCredentialAdapter) MintCredential(ctx context.Context, agentID string, scopes []string, ttlSeconds int64) (string, error) {
	var tenantID string
	err := db.QueryRowContext(ctx, "SELECT tenant_id FROM agents WHERE id = $1", agentID).Scan(&tenantID)
	if err != nil {
		return "", fmt.Errorf("lookup tenant for agent %s: %w", agentID, err)
	}
	_, token, err := a.svc.MintCredential(ctx, tenantID, agentID, scopes, ttlSeconds)
	return token, err
}

// identityQuerierAdapter wraps a DB query -> workspace.IdentityQuerier.
type identityQuerierAdapter struct {
	db *sql.DB
}

func (a *identityQuerierAdapter) GetAgent(ctx context.Context, agentID string) (*models.Agent, error) {
	var agent models.Agent
	err := a.db.QueryRowContext(ctx,
		"SELECT id, tenant_id, name, trust_level FROM agents WHERE id = $1", agentID,
	).Scan(&agent.ID, &agent.TenantID, &agent.Name, &agent.TrustLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &agent, nil
}

// workspaceProvisionerAdapter wraps workspaceSvc -> task.WorkspaceProvisioner.
type workspaceProvisionerAdapter struct {
	wsSvc *workspace.Service
}

func (a *workspaceProvisionerAdapter) ProvisionWorkspace(ctx context.Context, t *models.Task) (string, error) {
	spec := &models.WorkspaceSpec{
		MemoryMb:          t.WorkspaceConfig.MemoryMb,
		CpuMillicores:     t.WorkspaceConfig.CpuMillicores,
		DiskMb:            t.WorkspaceConfig.DiskMb,
		MaxDurationSecs:   t.WorkspaceConfig.MaxDurationSecs,
		AllowedTools:      t.WorkspaceConfig.AllowedTools,
		GuardrailPolicyID: t.GuardrailPolicyID,
		EnvVars:           t.WorkspaceConfig.EnvVars,
		ContainerImage:    t.WorkspaceConfig.ContainerImage,
		EgressAllowlist:   t.WorkspaceConfig.EgressAllowlist,
		IsolationTier:     models.IsolationTier(t.WorkspaceConfig.IsolationTier),
	}

	ws, err := a.wsSvc.CreateWorkspace(ctx, t.TenantID, t.AgentID, t.ID, spec)
	if err != nil {
		return "", err
	}
	if ws.Status == models.WorkspaceStatusFailed {
		return "", fmt.Errorf("workspace provisioning failed")
	}
	return ws.ID, nil
}

func (a *workspaceProvisionerAdapter) TerminateWorkspace(ctx context.Context, workspaceID string, reason string) error {
	var tenantID string
	err := db.QueryRowContext(ctx, "SELECT tenant_id FROM workspaces WHERE id = $1", workspaceID).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("lookup tenant for workspace %s: %w", workspaceID, err)
	}
	return a.wsSvc.TerminateWorkspace(ctx, tenantID, workspaceID, reason)
}

// activityLoggerAdapter wraps activitySvc.RecordAction -> human.ActivityLogger.
type activityLoggerAdapter struct {
	svc *activity.Service
}

func (a *activityLoggerAdapter) RecordAction(ctx context.Context, record *models.ActionRecord) error {
	_, err := a.svc.RecordAction(ctx, record)
	return err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// clean truncates all tables and resets mock state.
func clean(t *testing.T) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE
		warm_pool_slots, warm_pool_configs,
		workspace_snapshots, delivery_channels, timeout_policies,
		action_records, human_requests, scoped_credentials, tasks,
		agents, guardrail_rules, guardrail_sets, usage_records, budgets,
		workspaces, hosts
		CASCADE`)
	if err != nil {
		t.Fatalf("truncate all: %v", err)
	}
	fakeHostAgent.reset()
	fakeSnapshots.reset()
}

// uniqueTenant returns a unique tenant ID for test isolation.
func uniqueTenant() string {
	return fmt.Sprintf("tenant-%d", tenantCounter.Add(1))
}

// registerAgent is a shorthand for creating an agent via identitySvc.
func registerAgent(t *testing.T, ctx context.Context, tenantID, name string) *models.Agent {
	t.Helper()
	agent, err := identitySvc.RegisterAgent(ctx, tenantID, name, "test agent", "owner-1", nil, "testing", models.AgentTrustLevelNew, nil)
	if err != nil {
		t.Fatalf("register agent %q: %v", name, err)
	}
	return agent
}

// registerAgentWithTrust creates an agent with a specified trust level.
func registerAgentWithTrust(t *testing.T, ctx context.Context, tenantID, name string, trust models.AgentTrustLevel) *models.Agent {
	t.Helper()
	agent, err := identitySvc.RegisterAgent(ctx, tenantID, name, "test agent", "owner-1", nil, "testing", trust, nil)
	if err != nil {
		t.Fatalf("register agent %q: %v", name, err)
	}
	return agent
}

// registerHost is a shorthand for registering a compute host.
func registerHost(t *testing.T, ctx context.Context, address string, memMb int64, cpuMilli int32, diskMb int64, tiers []string) *models.Host {
	t.Helper()
	host, err := computeSvc.RegisterHost(ctx, address, models.HostResources{
		MemoryMb:      memMb,
		CpuMillicores: cpuMilli,
		DiskMb:        diskMb,
	}, tiers)
	if err != nil {
		t.Fatalf("register host %q: %v", address, err)
	}
	return host
}

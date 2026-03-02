package workspace

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

// syncMockRepo wraps mockRepo with a mutex to support concurrent benchmark access.
type syncMockRepo struct {
	mu   sync.Mutex
	mock *mockRepo
}

func newSyncMockRepo() *syncMockRepo {
	return &syncMockRepo{mock: newMockRepo()}
}

func (s *syncMockRepo) CreateWorkspace(ctx context.Context, ws *models.Workspace) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.CreateWorkspace(ctx, ws)
}

func (s *syncMockRepo) GetWorkspace(ctx context.Context, tenantID, id string) (*models.Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetWorkspace(ctx, tenantID, id)
}

func (s *syncMockRepo) ListWorkspaces(ctx context.Context, tenantID string, agentID string, status models.WorkspaceStatus, afterID string, limit int) ([]models.Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.ListWorkspaces(ctx, tenantID, agentID, status, afterID, limit)
}

func (s *syncMockRepo) UpdateWorkspaceStatus(ctx context.Context, tenantID, id string, status models.WorkspaceStatus, hostID, hostAddress, sandboxID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.UpdateWorkspaceStatus(ctx, tenantID, id, status, hostID, hostAddress, sandboxID)
}

func (s *syncMockRepo) TerminateWorkspace(ctx context.Context, tenantID, id string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.TerminateWorkspace(ctx, tenantID, id, reason)
}

func (s *syncMockRepo) SetSnapshotID(ctx context.Context, tenantID, workspaceID, snapshotID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.SetSnapshotID(ctx, tenantID, workspaceID, snapshotID)
}

func (s *syncMockRepo) CreateSnapshot(ctx context.Context, snapshot *models.WorkspaceSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.CreateSnapshot(ctx, snapshot)
}

func (s *syncMockRepo) GetSnapshot(ctx context.Context, tenantID, snapshotID string) (*models.WorkspaceSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetSnapshot(ctx, tenantID, snapshotID)
}

// ---------------------------------------------------------------------------
// 1. BenchmarkCreateWorkspace_Parallel
// ---------------------------------------------------------------------------

func BenchmarkCreateWorkspace_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := newTestService(newSyncMockRepo())
	ctx := context.Background()
	var counter atomic.Int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := counter.Add(1)
			_, err := svc.CreateWorkspace(ctx, "tenant-1", fmt.Sprintf("agent-%d", n), fmt.Sprintf("task-%d", n), nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 2. BenchmarkGetWorkspace_Parallel
// ---------------------------------------------------------------------------

func BenchmarkGetWorkspace_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := newTestService(newMockRepo())
	ctx := context.Background()

	const wsCount = 1000
	wsIDs := make([]string, wsCount)
	for i := 0; i < wsCount; i++ {
		ws, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", fmt.Sprintf("task-%d", i), nil)
		if err != nil {
			b.Fatal(err)
		}
		wsIDs[i] = ws.ID
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, err := svc.GetWorkspace(ctx, "tenant-1", wsIDs[i%wsCount])
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// ---------------------------------------------------------------------------
// 3. BenchmarkCreateWorkspace_FullSpec
// ---------------------------------------------------------------------------

func BenchmarkCreateWorkspace_FullSpec(b *testing.B) {
	b.ReportAllocs()
	svc := newTestService(newMockRepo())
	ctx := context.Background()

	spec := &models.WorkspaceSpec{
		MemoryMb:          4096,
		CpuMillicores:     4000,
		DiskMb:            8192,
		MaxDurationSecs:   7200,
		AllowedTools:      []string{"bash", "curl", "python", "node", "git"},
		GuardrailPolicyID: "policy-full",
		EnvVars: map[string]string{
			"ENV": "prod", "DEBUG": "false", "LOG_LEVEL": "info",
			"SERVICE_URL": "https://api.example.com", "REGION": "us-east-1",
		},
		ContainerImage:     "registry.example.com/agent-runtime:v2.1",
		EgressAllowlist:    []string{"api.example.com", "cdn.example.com", "auth.example.com", "logs.example.com"},
		IsolationTier:      models.IsolationTierHardened,
		DataClassification: "confidential",
	}

	for b.Loop() {
		_, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", "task-1", spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 4. BenchmarkCreateWorkspace_LargeEnvVars
// ---------------------------------------------------------------------------

func BenchmarkCreateWorkspace_LargeEnvVars(b *testing.B) {
	b.ReportAllocs()
	svc := newTestService(newMockRepo())
	ctx := context.Background()

	envVars := make(map[string]string, 50)
	for i := 0; i < 50; i++ {
		envVars[fmt.Sprintf("ENV_VAR_%d", i)] = fmt.Sprintf("value-%d-with-some-realistic-length-data", i)
	}
	spec := &models.WorkspaceSpec{EnvVars: envVars}

	for b.Loop() {
		_, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", "task-1", spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. BenchmarkCreateWorkspace_LargeEgressAllowlist
// ---------------------------------------------------------------------------

func BenchmarkCreateWorkspace_LargeEgressAllowlist(b *testing.B) {
	b.ReportAllocs()
	svc := newTestService(newMockRepo())
	ctx := context.Background()

	allowlist := make([]string, 100)
	for i := 0; i < 100; i++ {
		allowlist[i] = fmt.Sprintf("service-%d.example.com", i)
	}
	spec := &models.WorkspaceSpec{EgressAllowlist: allowlist}

	for b.Loop() {
		_, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", "task-1", spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. BenchmarkCreateWorkspace_LargeAllowedTools
// ---------------------------------------------------------------------------

func BenchmarkCreateWorkspace_LargeAllowedTools(b *testing.B) {
	b.ReportAllocs()
	svc := newTestService(newMockRepo())
	ctx := context.Background()

	tools := make([]string, 50)
	for i := 0; i < 50; i++ {
		tools[i] = fmt.Sprintf("tool-%d", i)
	}
	spec := &models.WorkspaceSpec{AllowedTools: tools}

	for b.Loop() {
		_, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", "task-1", spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 7. BenchmarkListWorkspaces_50K
// ---------------------------------------------------------------------------

func BenchmarkListWorkspaces_50K(b *testing.B) {
	b.ReportAllocs()
	svc := newTestService(newMockRepo())
	ctx := context.Background()

	const total = 50_000
	for i := 0; i < total; i++ {
		_, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", fmt.Sprintf("task-%d", i), nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.ListWorkspaces(ctx, "tenant-1", "", "", 50, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 8. BenchmarkListWorkspaces_WithAgentFilter
// ---------------------------------------------------------------------------

func BenchmarkListWorkspaces_WithAgentFilter(b *testing.B) {
	b.ReportAllocs()
	svc := newTestService(newMockRepo())
	ctx := context.Background()

	// Create workspaces across multiple agents.
	for i := 0; i < 1000; i++ {
		agentID := fmt.Sprintf("agent-%d", i%10)
		_, err := svc.CreateWorkspace(ctx, "tenant-1", agentID, fmt.Sprintf("task-%d", i), nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.ListWorkspaces(ctx, "tenant-1", "agent-5", "", 50, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 9. BenchmarkListWorkspaces_WithStatusFilter
// ---------------------------------------------------------------------------

func BenchmarkListWorkspaces_WithStatusFilter(b *testing.B) {
	b.ReportAllocs()
	repo := newMockRepo()
	svc := newTestService(repo)
	ctx := context.Background()

	// Create workspaces; some will be terminated.
	for i := 0; i < 1000; i++ {
		ws, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", fmt.Sprintf("task-%d", i), nil)
		if err != nil {
			b.Fatal(err)
		}
		if i%3 == 0 {
			svc.TerminateWorkspace(ctx, "tenant-1", ws.ID, "done")
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.ListWorkspaces(ctx, "tenant-1", "", models.WorkspaceStatusPending, 50, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 10. BenchmarkWorkspaceLifecycle_CreateTerminate
// ---------------------------------------------------------------------------

func BenchmarkWorkspaceLifecycle_CreateTerminate(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	for b.Loop() {
		b.StopTimer()
		svc := newTestService(newMockRepo())
		b.StartTimer()

		ws, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", "task-1", nil)
		if err != nil {
			b.Fatal(err)
		}
		err = svc.TerminateWorkspace(ctx, "tenant-1", ws.ID, "lifecycle complete")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 11. BenchmarkTerminateWorkspace_Parallel
// ---------------------------------------------------------------------------

func BenchmarkTerminateWorkspace_Parallel(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	// Pre-create workspaces. Each iteration consumes one.
	const poolSize = 1_000_000
	svc := newTestService(newSyncMockRepo())
	wsIDs := make([]string, poolSize)
	for i := 0; i < poolSize; i++ {
		ws, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", fmt.Sprintf("task-%d", i), nil)
		if err != nil {
			b.Fatal(err)
		}
		wsIDs[i] = ws.ID
	}

	var idx atomic.Int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := int(idx.Add(1) - 1)
			if i >= poolSize {
				return // pool exhausted
			}
			err := svc.TerminateWorkspace(ctx, "tenant-1", wsIDs[i], "parallel terminate")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 12. BenchmarkMultiTenantWorkspaceIsolation
// ---------------------------------------------------------------------------

func BenchmarkMultiTenantWorkspaceIsolation(b *testing.B) {
	b.ReportAllocs()
	svc := newTestService(newMockRepo())
	ctx := context.Background()

	const numTenants = 5
	const wsPerTenant = 200

	tenants := make([]string, numTenants)
	wsIDs := make([][]string, numTenants)

	for t := 0; t < numTenants; t++ {
		tenants[t] = fmt.Sprintf("tenant-%d", t)
		wsIDs[t] = make([]string, wsPerTenant)
		for i := 0; i < wsPerTenant; i++ {
			ws, err := svc.CreateWorkspace(ctx, tenants[t], "agent-1", fmt.Sprintf("task-%d", i), nil)
			if err != nil {
				b.Fatal(err)
			}
			wsIDs[t][i] = ws.ID
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := rand.IntN(numTenants * wsPerTenant)
		for pb.Next() {
			t := i % numTenants
			idx := i % wsPerTenant
			_, err := svc.GetWorkspace(ctx, tenants[t], wsIDs[t][idx])
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// ---------------------------------------------------------------------------
// 13. BenchmarkMixedWorkspaceWorkload
// ---------------------------------------------------------------------------

func BenchmarkMixedWorkspaceWorkload(b *testing.B) {
	b.ReportAllocs()
	svc := newTestService(newSyncMockRepo())
	ctx := context.Background()

	// Pre-populate workspaces for reads and terminates.
	const seedCount = 500
	wsIDs := make([]string, seedCount)
	for i := 0; i < seedCount; i++ {
		ws, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", fmt.Sprintf("task-%d", i), nil)
		if err != nil {
			b.Fatal(err)
		}
		wsIDs[i] = ws.ID
	}

	var createCounter atomic.Int64
	// For terminates, we need fresh workspaces since each can only be terminated once.
	// Pre-create a pool of terminable workspaces.
	const terminatePoolSize = 500_000
	terminateIDs := make([]string, terminatePoolSize)
	for i := 0; i < terminatePoolSize; i++ {
		ws, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", fmt.Sprintf("term-task-%d", i), nil)
		if err != nil {
			b.Fatal(err)
		}
		terminateIDs[i] = ws.ID
	}
	var terminateIdx atomic.Int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			r := i % 100
			switch {
			case r < 80: // 80% reads
				_, err := svc.GetWorkspace(ctx, "tenant-1", wsIDs[i%seedCount])
				if err != nil {
					b.Fatal(err)
				}
			case r < 95: // 15% creates
				n := createCounter.Add(1)
				_, err := svc.CreateWorkspace(ctx, "tenant-1", "agent-1", fmt.Sprintf("mixed-task-%d", n), nil)
				if err != nil {
					b.Fatal(err)
				}
			default: // 5% terminates
				ti := int(terminateIdx.Add(1) - 1)
				if ti < terminatePoolSize {
					// Ignore errors on double-terminate from concurrent access.
					svc.TerminateWorkspace(ctx, "tenant-1", terminateIDs[ti], "mixed workload")
				}
			}
			i++
		}
	})
}

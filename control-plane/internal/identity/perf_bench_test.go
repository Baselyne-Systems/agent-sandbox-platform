package identity

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// syncMockRepo wraps mockRepo with a mutex to support concurrent benchmark access.
type syncMockRepo struct {
	mu   sync.Mutex
	mock *mockRepo
}

func newSyncMockRepo() *syncMockRepo {
	return &syncMockRepo{mock: newMockRepo()}
}

func (s *syncMockRepo) CreateAgent(ctx context.Context, agent *models.Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.CreateAgent(ctx, agent)
}

func (s *syncMockRepo) GetAgent(ctx context.Context, tenantID, id string) (*models.Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetAgent(ctx, tenantID, id)
}

func (s *syncMockRepo) ListAgents(ctx context.Context, tenantID, ownerID string, status models.AgentStatus, afterID string, limit int) ([]models.Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.ListAgents(ctx, tenantID, ownerID, status, afterID, limit)
}

func (s *syncMockRepo) DeactivateAgent(ctx context.Context, tenantID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.DeactivateAgent(ctx, tenantID, id)
}

func (s *syncMockRepo) UpdateTrustLevel(ctx context.Context, tenantID, agentID string, level models.AgentTrustLevel) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.UpdateTrustLevel(ctx, tenantID, agentID, level)
}

func (s *syncMockRepo) UpdateAgentStatus(ctx context.Context, tenantID, agentID string, from []models.AgentStatus, to models.AgentStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.UpdateAgentStatus(ctx, tenantID, agentID, from, to)
}

func (s *syncMockRepo) CreateCredential(ctx context.Context, cred *models.ScopedCredential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.CreateCredential(ctx, cred)
}

func (s *syncMockRepo) RevokeCredential(ctx context.Context, tenantID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.RevokeCredential(ctx, tenantID, id)
}

// ---------------------------------------------------------------------------
// 1. BenchmarkRegisterAgent_Parallel
// ---------------------------------------------------------------------------

func BenchmarkRegisterAgent_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()
	var counter atomic.Int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := counter.Add(1)
			name := fmt.Sprintf("agent-%d", n)
			_, err := svc.RegisterAgent(ctx, testTenant, name, "benchmark agent", "owner-1",
				map[string]string{"env": "bench"}, "benchmark", models.AgentTrustLevelNew, []string{"bash"})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 2. BenchmarkGetAgent_Parallel
// ---------------------------------------------------------------------------

func BenchmarkGetAgent_Parallel(b *testing.B) {
	b.ReportAllocs()
	repo := newSyncMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	const agentCount = 1000
	agentIDs := make([]string, agentCount)
	for i := 0; i < agentCount; i++ {
		a, err := svc.RegisterAgent(ctx, testTenant, fmt.Sprintf("agent-%d", i), "", "owner-1", nil, "", "", nil)
		if err != nil {
			b.Fatal(err)
		}
		agentIDs[i] = a.ID
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, err := svc.GetAgent(ctx, testTenant, agentIDs[i%agentCount])
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// ---------------------------------------------------------------------------
// 3. BenchmarkMintCredential_Parallel
// ---------------------------------------------------------------------------

func BenchmarkMintCredential_Parallel(b *testing.B) {
	b.ReportAllocs()
	repo := newSyncMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	const agentCount = 100
	agentIDs := make([]string, agentCount)
	for i := 0; i < agentCount; i++ {
		a, err := svc.RegisterAgent(ctx, testTenant, fmt.Sprintf("agent-%d", i), "", "owner-1", nil, "", "", nil)
		if err != nil {
			b.Fatal(err)
		}
		agentIDs[i] = a.ID
	}
	scopes := []string{"read", "write"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _, err := svc.MintCredential(ctx, testTenant, agentIDs[i%agentCount], scopes, 3600)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// ---------------------------------------------------------------------------
// 4. BenchmarkMixedReadWrite
// ---------------------------------------------------------------------------

func BenchmarkMixedReadWrite(b *testing.B) {
	b.ReportAllocs()
	repo := newSyncMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Pre-populate agents for reads and mints.
	const seedCount = 200
	agentIDs := make([]string, seedCount)
	for i := 0; i < seedCount; i++ {
		a, err := svc.RegisterAgent(ctx, testTenant, fmt.Sprintf("seed-%d", i), "", "owner-1", nil, "", "", nil)
		if err != nil {
			b.Fatal(err)
		}
		agentIDs[i] = a.ID
	}

	var regCounter atomic.Int64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			r := i % 100
			switch {
			case r < 80: // 80% reads
				_, err := svc.GetAgent(ctx, testTenant, agentIDs[i%seedCount])
				if err != nil {
					b.Fatal(err)
				}
			case r < 90: // 10% registers
				n := regCounter.Add(1)
				_, err := svc.RegisterAgent(ctx, testTenant, fmt.Sprintf("mixed-%d", n), "", "owner-1", nil, "", "", nil)
				if err != nil {
					b.Fatal(err)
				}
			default: // 10% mints
				_, _, err := svc.MintCredential(ctx, testTenant, agentIDs[i%seedCount], []string{"read"}, 3600)
				if err != nil {
					b.Fatal(err)
				}
			}
			i++
		}
	})
}

// ---------------------------------------------------------------------------
// 5. BenchmarkRegisterAgent_MaxLabels
// ---------------------------------------------------------------------------

func BenchmarkRegisterAgent_MaxLabels(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	labels := make(map[string]string, 50)
	for i := 0; i < 50; i++ {
		labels[fmt.Sprintf("key-%d", i)] = fmt.Sprintf("value-%d", i)
	}

	for b.Loop() {
		_, err := svc.RegisterAgent(ctx, testTenant, "label-agent", "many labels", "owner-1",
			labels, "labelled", models.AgentTrustLevelNew, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. BenchmarkRegisterAgent_MaxCapabilities
// ---------------------------------------------------------------------------

func BenchmarkRegisterAgent_MaxCapabilities(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	caps := make([]string, 100)
	for i := 0; i < 100; i++ {
		caps[i] = fmt.Sprintf("capability-%d", i)
	}

	for b.Loop() {
		_, err := svc.RegisterAgent(ctx, testTenant, "cap-agent", "many capabilities", "owner-1",
			nil, "capable", models.AgentTrustLevelNew, caps)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 7. BenchmarkRegisterAgent_LongStrings
// ---------------------------------------------------------------------------

func BenchmarkRegisterAgent_LongStrings(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	longName := strings.Repeat("n", 255)
	longDesc := strings.Repeat("d", 4096)

	for b.Loop() {
		_, err := svc.RegisterAgent(ctx, testTenant, longName, longDesc, "owner-1",
			nil, "long strings test", models.AgentTrustLevelNew, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 8. BenchmarkMultiTenantConcurrentReads
// ---------------------------------------------------------------------------

func BenchmarkMultiTenantConcurrentReads(b *testing.B) {
	b.ReportAllocs()
	repo := newSyncMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	const numTenants = 5
	const agentsPerTenant = 200

	tenants := make([]string, numTenants)
	agentIDs := make([][]string, numTenants)
	for t := 0; t < numTenants; t++ {
		tenants[t] = fmt.Sprintf("tenant-%d", t)
		agentIDs[t] = make([]string, agentsPerTenant)
		for i := 0; i < agentsPerTenant; i++ {
			a, err := svc.RegisterAgent(ctx, tenants[t], fmt.Sprintf("agent-%d", i), "", "owner-1", nil, "", "", nil)
			if err != nil {
				b.Fatal(err)
			}
			agentIDs[t][i] = a.ID
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			t := i % numTenants
			a := i % agentsPerTenant
			_, err := svc.GetAgent(ctx, tenants[t], agentIDs[t][a])
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// ---------------------------------------------------------------------------
// 9. BenchmarkMultiTenantListIsolation
// ---------------------------------------------------------------------------

func BenchmarkMultiTenantListIsolation(b *testing.B) {
	b.ReportAllocs()
	repo := newSyncMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	const numTenants = 3
	const agentsPerTenant = 500

	tenants := make([]string, numTenants)
	for t := 0; t < numTenants; t++ {
		tenants[t] = fmt.Sprintf("list-tenant-%d", t)
		for i := 0; i < agentsPerTenant; i++ {
			_, err := svc.RegisterAgent(ctx, tenants[t], fmt.Sprintf("agent-%d", i), "", "owner-1", nil, "", "", nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			t := i % numTenants
			_, _, err := svc.ListAgents(ctx, tenants[t], "", "", 50, "")
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// ---------------------------------------------------------------------------
// 10. BenchmarkAgentLifecycle_FullCycle
// ---------------------------------------------------------------------------

func BenchmarkAgentLifecycle_FullCycle(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	for b.Loop() {
		b.StopTimer()
		svc := NewService(newMockRepo())
		b.StartTimer()

		// Register
		agent, err := svc.RegisterAgent(ctx, testTenant, "lifecycle-agent", "", "owner-1",
			nil, "", models.AgentTrustLevelNew, nil)
		if err != nil {
			b.Fatal(err)
		}

		// Mint credential
		_, _, err = svc.MintCredential(ctx, testTenant, agent.ID, []string{"read"}, 3600)
		if err != nil {
			b.Fatal(err)
		}

		// Update trust level
		_, err = svc.UpdateTrustLevel(ctx, testTenant, agent.ID, models.AgentTrustLevelEstablished, "passed checks")
		if err != nil {
			b.Fatal(err)
		}

		// Suspend
		_, err = svc.SuspendAgent(ctx, testTenant, agent.ID)
		if err != nil {
			b.Fatal(err)
		}

		// Reactivate
		_, err = svc.ReactivateAgent(ctx, testTenant, agent.ID)
		if err != nil {
			b.Fatal(err)
		}

		// Deactivate
		err = svc.DeactivateAgent(ctx, testTenant, agent.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 11. BenchmarkCredentialChurn
// ---------------------------------------------------------------------------

func BenchmarkCredentialChurn(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	agent, err := svc.RegisterAgent(ctx, testTenant, "churn-agent", "", "owner-1", nil, "", "", nil)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		cred, _, err := svc.MintCredential(ctx, testTenant, agent.ID, []string{"read", "write"}, 3600)
		if err != nil {
			b.Fatal(err)
		}
		err = svc.RevokeCredential(ctx, testTenant, cred.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 12. BenchmarkGenerateToken_Parallel
// ---------------------------------------------------------------------------

func BenchmarkGenerateToken_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, err := GenerateToken()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 13. BenchmarkHashToken_Parallel
// ---------------------------------------------------------------------------

func BenchmarkHashToken_Parallel(b *testing.B) {
	b.ReportAllocs()
	token := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			HashToken(token)
		}
	})
}

// ---------------------------------------------------------------------------
// 14. BenchmarkDeactivateAgent_ConcurrentCredentialAccess
// ---------------------------------------------------------------------------

func BenchmarkDeactivateAgent_ConcurrentCredentialAccess(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	for b.Loop() {
		b.StopTimer()
		svc := NewService(newMockRepo())
		agent, err := svc.RegisterAgent(ctx, testTenant, "deact-agent", "", "owner-1", nil, "", "", nil)
		if err != nil {
			b.Fatal(err)
		}
		for j := 0; j < 50; j++ {
			_, _, err := svc.MintCredential(ctx, testTenant, agent.ID, []string{"read"}, 3600)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StartTimer()

		err = svc.DeactivateAgent(ctx, testTenant, agent.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 15. BenchmarkListAgents_100K
// ---------------------------------------------------------------------------

func BenchmarkListAgents_100K(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	const total = 100_000
	for i := 0; i < total; i++ {
		_, err := svc.RegisterAgent(ctx, testTenant, fmt.Sprintf("agent-%d", i), "", "owner-1", nil, "", "", nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.ListAgents(ctx, testTenant, "", "", 50, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 16. BenchmarkGetAgent_HighCardinalityTenants
// ---------------------------------------------------------------------------

func BenchmarkGetAgent_HighCardinalityTenants(b *testing.B) {
	b.ReportAllocs()
	repo := newSyncMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	const numTenants = 1000
	const agentsPerTenant = 10

	type tenantAgent struct {
		tenantID string
		agentID  string
	}
	all := make([]tenantAgent, 0, numTenants*agentsPerTenant)

	for t := 0; t < numTenants; t++ {
		tid := fmt.Sprintf("tenant-%d", t)
		for i := 0; i < agentsPerTenant; i++ {
			a, err := svc.RegisterAgent(ctx, tid, fmt.Sprintf("agent-%d", i), "", "owner-1", nil, "", "", nil)
			if err != nil {
				b.Fatal(err)
			}
			all = append(all, tenantAgent{tenantID: tid, agentID: a.ID})
		}
	}

	total := len(all)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := rand.IntN(total) // stagger starting position
		for pb.Next() {
			ta := all[i%total]
			_, err := svc.GetAgent(ctx, ta.tenantID, ta.agentID)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

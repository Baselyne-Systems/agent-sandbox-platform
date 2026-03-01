package economics

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func (s *syncMockRepo) InsertUsage(ctx context.Context, record *models.UsageRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.InsertUsage(ctx, record)
}

func (s *syncMockRepo) GetBudget(ctx context.Context, tenantID, agentID string) (*models.Budget, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetBudget(ctx, tenantID, agentID)
}

func (s *syncMockRepo) UpsertBudget(ctx context.Context, budget *models.Budget) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.UpsertBudget(ctx, budget)
}

func (s *syncMockRepo) AddUsedAmount(ctx context.Context, tenantID, agentID string, amount float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.AddUsedAmount(ctx, tenantID, agentID, amount)
}

func (s *syncMockRepo) GetCostReport(ctx context.Context, tenantID, agentID string, start, end time.Time) ([]ResourceCost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetCostReport(ctx, tenantID, agentID, start, end)
}

// ---------------------------------------------------------------------------
// 1. BenchmarkRecordUsage_Parallel
// ---------------------------------------------------------------------------

func BenchmarkRecordUsage_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := svc.RecordUsage(ctx, "tenant-1", "agent-1", "ws-1", "compute", "seconds", 100, 0.50)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 2. BenchmarkCheckBudget_Parallel
// ---------------------------------------------------------------------------

func BenchmarkCheckBudget_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()
	svc.SetBudget(ctx, "tenant-1", "agent-1", 1000000, "USD", "halt", 0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 1.0)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 3. BenchmarkCheckBudget_NearLimit
// ---------------------------------------------------------------------------

func BenchmarkCheckBudget_NearLimit(b *testing.B) {
	b.ReportAllocs()
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()
	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "halt", 0.2)
	// Set Used to 99, leaving only 1.0 remaining.
	repo.budgets["agent-1"].Used = 99.0

	b.ResetTimer()
	for b.Loop() {
		result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 0.5)
		if err != nil {
			b.Fatal(err)
		}
		if !result.Allowed {
			b.Fatal("expected allowed")
		}
	}
}

// ---------------------------------------------------------------------------
// 4. BenchmarkCheckBudget_OverLimit_Halt
// ---------------------------------------------------------------------------

func BenchmarkCheckBudget_OverLimit_Halt(b *testing.B) {
	b.ReportAllocs()
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()
	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "halt", 0)
	repo.budgets["agent-1"].Used = 100.0

	b.ResetTimer()
	for b.Loop() {
		result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 10.0)
		if err != nil {
			b.Fatal(err)
		}
		if result.Allowed {
			b.Fatal("expected denied with halt")
		}
		if result.EnforcementAction != "halt" {
			b.Fatalf("expected enforcement=halt, got %q", result.EnforcementAction)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. BenchmarkCheckBudget_OverLimit_Warn
// ---------------------------------------------------------------------------

func BenchmarkCheckBudget_OverLimit_Warn(b *testing.B) {
	b.ReportAllocs()
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()
	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "warn", 0)
	repo.budgets["agent-1"].Used = 100.0

	b.ResetTimer()
	for b.Loop() {
		result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 10.0)
		if err != nil {
			b.Fatal(err)
		}
		if !result.Allowed {
			b.Fatal("expected allowed in warn mode")
		}
		if result.EnforcementAction != "warn" {
			b.Fatalf("expected enforcement=warn, got %q", result.EnforcementAction)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. BenchmarkCheckBudget_NoBudget
// ---------------------------------------------------------------------------

func BenchmarkCheckBudget_NoBudget(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()
	// No budget set -- should always allow.

	for b.Loop() {
		result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 1000.0)
		if err != nil {
			b.Fatal(err)
		}
		if !result.Allowed {
			b.Fatal("expected allowed when no budget exists")
		}
	}
}

// ---------------------------------------------------------------------------
// 7. BenchmarkSetBudget_Upsert
// ---------------------------------------------------------------------------

func BenchmarkSetBudget_Upsert(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()
	// Initial set.
	svc.SetBudget(ctx, "tenant-1", "agent-1", 1000, "USD", "halt", 0)

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.SetBudget(ctx, "tenant-1", "agent-1", 2000, "USD", "halt", 0.2)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 8. BenchmarkRecordUsage_WithBudgetUpdate_Parallel
// ---------------------------------------------------------------------------

func BenchmarkRecordUsage_WithBudgetUpdate_Parallel(b *testing.B) {
	b.ReportAllocs()
	repo := newSyncMockRepo()
	svc := NewService(repo)
	ctx := context.Background()
	svc.SetBudget(ctx, "tenant-1", "agent-1", 1e12, "USD", "halt", 0) // very high limit

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := svc.RecordUsage(ctx, "tenant-1", "agent-1", "ws-1", "compute", "seconds", 10, 0.01)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 9. BenchmarkGetCostReport_ManyResourceTypes
// ---------------------------------------------------------------------------

func BenchmarkGetCostReport_ManyResourceTypes(b *testing.B) {
	b.ReportAllocs()
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	resourceTypes := []string{
		"compute", "storage", "network", "memory", "gpu",
		"api_calls", "tokens_in", "tokens_out", "bandwidth", "disk_io",
	}
	for i := 0; i < 1000; i++ {
		svc.RecordUsage(ctx, "tenant-1", "agent-1", "ws-1", resourceTypes[i%len(resourceTypes)], "units", 10, 0.50)
	}

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)

	b.ResetTimer()
	for b.Loop() {
		report, err := svc.GetCostReport(ctx, "tenant-1", "agent-1", start, end)
		if err != nil {
			b.Fatal(err)
		}
		if len(report.ByResourceType) == 0 {
			b.Fatal("expected non-empty report")
		}
	}
}

// ---------------------------------------------------------------------------
// 10. BenchmarkGetCostReport_LargeDataset
// ---------------------------------------------------------------------------

func BenchmarkGetCostReport_LargeDataset(b *testing.B) {
	b.ReportAllocs()
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	for i := 0; i < 10000; i++ {
		svc.RecordUsage(ctx, "tenant-1", fmt.Sprintf("agent-%d", i%5), "ws-1",
			fmt.Sprintf("resource-%d", i%3), "units", float64(i%100+1), float64(i%50)*0.1)
	}

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.GetCostReport(ctx, "tenant-1", "", start, end)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 11. BenchmarkBudgetLifecycle_SetCheckRecord
// ---------------------------------------------------------------------------

func BenchmarkBudgetLifecycle_SetCheckRecord(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	for b.Loop() {
		svc := NewService(newMockRepo())

		// Set budget.
		_, err := svc.SetBudget(ctx, "tenant-1", "agent-1", 10000, "USD", "halt", 0.8)
		if err != nil {
			b.Fatal(err)
		}

		// Check budget (should be allowed).
		result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 5.0)
		if err != nil {
			b.Fatal(err)
		}
		if !result.Allowed {
			b.Fatal("expected allowed")
		}

		// Record usage.
		_, err = svc.RecordUsage(ctx, "tenant-1", "agent-1", "ws-1", "compute", "seconds", 100, 5.0)
		if err != nil {
			b.Fatal(err)
		}

		// Check again.
		_, err = svc.CheckBudget(ctx, "tenant-1", "agent-1", 5.0)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 12. BenchmarkMultiTenantBudgetIsolation
// ---------------------------------------------------------------------------

func BenchmarkMultiTenantBudgetIsolation(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	tenants := []string{"tenant-1", "tenant-2", "tenant-3", "tenant-4", "tenant-5"}
	for _, t := range tenants {
		svc.SetBudget(ctx, t, "agent-1", 100000, "USD", "halt", 0.8)
	}

	var idx atomic.Int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := idx.Add(1) % int64(len(tenants))
			_, err := svc.CheckBudget(ctx, tenants[i], "agent-1", 1.0)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 13. BenchmarkRecordUsage_HighFrequency
// ---------------------------------------------------------------------------

func BenchmarkRecordUsage_HighFrequency(b *testing.B) {
	b.ReportAllocs()
	repo := newSyncMockRepo()
	svc := NewService(repo)
	ctx := context.Background()
	svc.SetBudget(ctx, "tenant-1", "agent-1", 1e15, "USD", "halt", 0) // very high limit

	var ops atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := ops.Add(1)
			_, err := svc.RecordUsage(ctx, "tenant-1", "agent-1", fmt.Sprintf("ws-%d", n%10),
				"compute", "seconds", float64(n%100+1), 0.01)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 14. BenchmarkCheckBudget_WarningThreshold
// ---------------------------------------------------------------------------

func BenchmarkCheckBudget_WarningThreshold(b *testing.B) {
	b.ReportAllocs()
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Set warningThreshold=0.8 (warn when remaining < 80% of limit).
	svc.SetBudget(ctx, "tenant-1", "agent-1", 100, "USD", "halt", 0.8)
	// Used = 25, remaining = 75 which is < 0.8*100 = 80, so warning should fire.
	repo.budgets["agent-1"].Used = 25.0

	b.ResetTimer()
	for b.Loop() {
		result, err := svc.CheckBudget(ctx, "tenant-1", "agent-1", 5.0)
		if err != nil {
			b.Fatal(err)
		}
		if !result.Allowed {
			b.Fatal("expected allowed")
		}
		if !result.Warning {
			b.Fatal("expected warning to be set")
		}
	}
}

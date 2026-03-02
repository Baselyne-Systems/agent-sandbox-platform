package compute

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func (s *syncMockRepo) CreateHost(ctx context.Context, host *models.Host) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.CreateHost(ctx, host)
}

func (s *syncMockRepo) GetHost(ctx context.Context, id string) (*models.Host, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetHost(ctx, id)
}

func (s *syncMockRepo) ListHosts(ctx context.Context, status models.HostStatus) ([]models.Host, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.ListHosts(ctx, status)
}

func (s *syncMockRepo) SetHostStatus(ctx context.Context, id string, status models.HostStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.SetHostStatus(ctx, id, status)
}

func (s *syncMockRepo) PlaceAndDecrement(ctx context.Context, memoryMb int64, cpuMillicores int32, diskMb int64, isolationTier string) (*models.Host, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.PlaceAndDecrement(ctx, memoryMb, cpuMillicores, diskMb, isolationTier)
}

func (s *syncMockRepo) UpdateHeartbeat(ctx context.Context, hostID string, resources models.HostResources, activeSandboxes int32, supportedTiers []string) (*models.Host, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.UpdateHeartbeat(ctx, hostID, resources, activeSandboxes, supportedTiers)
}

func (s *syncMockRepo) MarkStaleHostsOffline(ctx context.Context, timeout time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.MarkStaleHostsOffline(ctx, timeout)
}

func (s *syncMockRepo) UpsertWarmPoolConfig(ctx context.Context, cfg *models.WarmPoolConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.UpsertWarmPoolConfig(ctx, cfg)
}

func (s *syncMockRepo) ListWarmPoolConfigs(ctx context.Context) ([]models.WarmPoolConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.ListWarmPoolConfigs(ctx)
}

func (s *syncMockRepo) ClaimWarmSlot(ctx context.Context, tier string) (*models.WarmPoolSlot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.ClaimWarmSlot(ctx, tier)
}

func (s *syncMockRepo) CreateWarmSlot(ctx context.Context, slot *models.WarmPoolSlot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.CreateWarmSlot(ctx, slot)
}

func (s *syncMockRepo) CountReadySlots(ctx context.Context, tier string) (int32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.CountReadySlots(ctx, tier)
}

func (s *syncMockRepo) CleanExpiredSlots(ctx context.Context) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.CleanExpiredSlots(ctx)
}

func (s *syncMockRepo) GetCapacity(ctx context.Context) ([]models.TierCapacity, int32, int32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mock.GetCapacity(ctx)
}

// ---------------------------------------------------------------------------
// 1. BenchmarkPlaceWorkspace_Parallel
// ---------------------------------------------------------------------------

func BenchmarkPlaceWorkspace_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		_, err := svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", i), models.HostResources{
			MemoryMb: 1 << 40, CpuMillicores: 1 << 30, DiskMb: 1 << 40,
		}, []string{"standard", "hardened"})
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "standard")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 2. BenchmarkRegisterHost_Parallel
// ---------------------------------------------------------------------------

func BenchmarkRegisterHost_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()
	var counter atomic.Int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := counter.Add(1)
			_, err := svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", n), models.HostResources{
				MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400,
			}, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 3. BenchmarkHeartbeat_Parallel
// ---------------------------------------------------------------------------

func BenchmarkHeartbeat_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()

	hostIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		host, err := svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", i), models.HostResources{
			MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400,
		}, nil)
		if err != nil {
			b.Fatal(err)
		}
		hostIDs[i] = host.ID
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			id := hostIDs[idx%len(hostIDs)]
			idx++
			_, err := svc.Heartbeat(ctx, id, models.HostResources{
				MemoryMb: 14000, CpuMillicores: 7000, DiskMb: 90000,
			}, 3, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 4. BenchmarkPlaceWorkspace_MinResources
// ---------------------------------------------------------------------------

func BenchmarkPlaceWorkspace_MinResources(b *testing.B) {
	b.ReportAllocs()
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	largeRes := models.HostResources{MemoryMb: math.MaxInt64, CpuMillicores: math.MaxInt32, DiskMb: math.MaxInt64}
	host, err := svc.RegisterHost(ctx, "host:9090", largeRes, nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		repo.hosts[host.ID].AvailableResources = largeRes
		_, _, err := svc.PlaceWorkspace(ctx, 1, 1, 1, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. BenchmarkPlaceWorkspace_MaxResources
// ---------------------------------------------------------------------------

func BenchmarkPlaceWorkspace_MaxResources(b *testing.B) {
	b.ReportAllocs()
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	largeRes := models.HostResources{MemoryMb: math.MaxInt64, CpuMillicores: math.MaxInt32, DiskMb: math.MaxInt64}
	host, err := svc.RegisterHost(ctx, "host:9090", largeRes, nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		repo.hosts[host.ID].AvailableResources = largeRes
		_, _, err := svc.PlaceWorkspace(ctx, 1<<20, 1<<15, 1<<20, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. BenchmarkPlaceWorkspace_TierFiltered
// ---------------------------------------------------------------------------

func BenchmarkPlaceWorkspace_TierFiltered(b *testing.B) {
	tiers := []struct {
		name string
		tier string
	}{
		{"standard", "standard"},
		{"hardened", "hardened"},
		{"isolated", "isolated"},
	}

	for _, tc := range tiers {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			repo := newMockRepo()
			svc := NewService(repo)
			ctx := context.Background()

			largeRes := models.HostResources{MemoryMb: math.MaxInt64, CpuMillicores: math.MaxInt32, DiskMb: math.MaxInt64}
			// 50 hosts with mixed tier support.
			var hostIDs []string
			for i := 0; i < 50; i++ {
				var supported []string
				switch i % 3 {
				case 0:
					supported = []string{"standard"}
				case 1:
					supported = []string{"standard", "hardened"}
				case 2:
					supported = []string{"standard", "hardened", "isolated"}
				}
				host, err := svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", i), largeRes, supported)
				if err != nil {
					b.Fatal(err)
				}
				hostIDs = append(hostIDs, host.ID)
			}

			b.ResetTimer()
			for b.Loop() {
				// Reset all host resources to prevent capacity exhaustion.
				for _, id := range hostIDs {
					repo.hosts[id].AvailableResources = largeRes
				}
				_, _, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, tc.tier)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 7. BenchmarkPlaceWorkspace_NearCapacity
// ---------------------------------------------------------------------------

func BenchmarkPlaceWorkspace_NearCapacity(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// 1000 hosts with just-enough capacity for a single 512/500/1024 placement.
	for i := 0; i < 1000; i++ {
		_, err := svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", i), models.HostResources{
			MemoryMb: 512, CpuMillicores: 500, DiskMb: 1024,
		}, nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		// Some placements may fail once capacity is exhausted; we just measure throughput.
		svc.PlaceWorkspace(ctx, 512, 500, 1024, "")
	}
}

// ---------------------------------------------------------------------------
// 8. BenchmarkConfigureWarmPool
// ---------------------------------------------------------------------------

func BenchmarkConfigureWarmPool(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		_, err := svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", i), models.HostResources{
			MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400,
		}, []string{"standard"})
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := svc.ConfigureWarmPool(ctx, &models.WarmPoolConfig{
			IsolationTier: "standard",
			TargetCount:   5,
			MemoryMb:      512,
			CpuMillicores: 500,
			DiskMb:        1024,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 9. BenchmarkPlaceWorkspace_WarmPoolHit
// ---------------------------------------------------------------------------

func BenchmarkPlaceWorkspace_WarmPoolHit(b *testing.B) {
	b.ReportAllocs()
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	host, err := svc.RegisterHost(ctx, "warm-host:9090", models.HostResources{
		MemoryMb: 1 << 30, CpuMillicores: 1 << 20, DiskMb: 1 << 30,
	}, []string{"standard"})
	if err != nil {
		b.Fatal(err)
	}

	_, err = svc.ConfigureWarmPool(ctx, &models.WarmPoolConfig{
		IsolationTier: "standard",
		TargetCount:   10,
		MemoryMb:      512,
		CpuMillicores: 500,
		DiskMb:        1024,
	})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		// Replenish a warm slot so the next iteration can claim it.
		repo.CreateWarmSlot(ctx, &models.WarmPoolSlot{
			HostID:        host.ID,
			IsolationTier: "standard",
			MemoryMb:      512,
			CpuMillicores: 500,
			DiskMb:        1024,
		})
		b.StartTimer()

		_, _, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "standard")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 10. BenchmarkListHosts_LargeFleet
// ---------------------------------------------------------------------------

func BenchmarkListHosts_LargeFleet(b *testing.B) {
	for _, n := range []int{100, 500, 2000, 5000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			ctx := context.Background()

			for i := 0; i < n; i++ {
				_, err := svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", i), models.HostResources{
					MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400,
				}, nil)
				if err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for b.Loop() {
				_, err := svc.ListHosts(ctx, "")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 11. BenchmarkGetCapacity_LargeFleet
// ---------------------------------------------------------------------------

func BenchmarkGetCapacity_LargeFleet(b *testing.B) {
	for _, n := range []int{100, 500, 2000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(newMockRepo())
			ctx := context.Background()

			for i := 0; i < n; i++ {
				var supported []string
				switch i % 3 {
				case 0:
					supported = []string{"standard"}
				case 1:
					supported = []string{"standard", "hardened"}
				case 2:
					supported = []string{"standard", "hardened", "isolated"}
				}
				_, err := svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", i), models.HostResources{
					MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400,
				}, supported)
				if err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for b.Loop() {
				_, _, _, err := svc.GetCapacity(ctx)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 12. BenchmarkMixedHeartbeatAndPlacement
// ---------------------------------------------------------------------------

func BenchmarkMixedHeartbeatAndPlacement(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()

	hostIDs := make([]string, 200)
	for i := 0; i < 200; i++ {
		host, err := svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", i), models.HostResources{
			MemoryMb: 1 << 40, CpuMillicores: 1 << 30, DiskMb: 1 << 40,
		}, []string{"standard"})
		if err != nil {
			b.Fatal(err)
		}
		hostIDs[i] = host.ID
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		idx := 0
		for pb.Next() {
			idx++
			if rng.Intn(100) < 80 {
				// 80% heartbeats
				id := hostIDs[idx%len(hostIDs)]
				_, err := svc.Heartbeat(ctx, id, models.HostResources{
					MemoryMb: 14000, CpuMillicores: 7000, DiskMb: 90000,
				}, 2, nil)
				if err != nil {
					b.Fatal(err)
				}
			} else {
				// 20% placements
				_, _, err := svc.PlaceWorkspace(ctx, 256, 250, 512, "standard")
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 13. BenchmarkPlaceWorkspace_HighContention
// ---------------------------------------------------------------------------

func BenchmarkPlaceWorkspace_HighContention(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(newSyncMockRepo())
	ctx := context.Background()

	// Only 10 hosts to create high contention.
	for i := 0; i < 10; i++ {
		_, err := svc.RegisterHost(ctx, fmt.Sprintf("host%d:9090", i), models.HostResources{
			MemoryMb: 1 << 40, CpuMillicores: 1 << 30, DiskMb: 1 << 40,
		}, []string{"standard"})
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "standard")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 14. BenchmarkHostLifecycle_RegisterDeregister
// ---------------------------------------------------------------------------

func BenchmarkHostLifecycle_RegisterDeregister(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	for b.Loop() {
		svc := NewService(newMockRepo())

		host, err := svc.RegisterHost(ctx, "lifecycle-host:9090", models.HostResources{
			MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400,
		}, []string{"standard"})
		if err != nil {
			b.Fatal(err)
		}

		_, err = svc.Heartbeat(ctx, host.ID, models.HostResources{
			MemoryMb: 14000, CpuMillicores: 7000, DiskMb: 90000,
		}, 1, nil)
		if err != nil {
			b.Fatal(err)
		}

		err = svc.DeregisterHost(ctx, host.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

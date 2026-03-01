//go:build integration

package compute

import (
	"context"
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	"github.com/lib/pq"
)

func benchSetup(b *testing.B) *PostgresRepository {
	b.Helper()
	_, err := testDB.DB.Exec(`TRUNCATE
		warm_pool_slots, warm_pool_configs,
		workspace_snapshots, delivery_channels, timeout_policies,
		action_records, human_requests, scoped_credentials, tasks,
		agents, guardrail_rules, guardrail_sets, usage_records, budgets,
		workspaces, hosts
		CASCADE`)
	if err != nil {
		b.Fatalf("truncate: %v", err)
	}
	return NewPostgresRepository(testDB.DB)
}

func seedHost(b *testing.B, db *sql.DB, addr string, mem int64, cpu int32, disk int64, tiers []string) string {
	b.Helper()
	var id string
	err := db.QueryRow(
		`INSERT INTO hosts (address, status, total_memory_mb, total_cpu_millicores, total_disk_mb,
		   available_memory_mb, available_cpu_millicores, available_disk_mb, active_sandboxes, last_heartbeat, supported_tiers)
		 VALUES ($1, 'ready', $2, $3, $4, $2, $3, $4, 0, $5, $6)
		 RETURNING id`,
		addr, mem, cpu, disk, time.Now(), pq.Array(tiers),
	).Scan(&id)
	if err != nil {
		b.Fatalf("seedHost %s: %v", addr, err)
	}
	return id
}

func BenchmarkDB_CreateHost(b *testing.B) {
	benchSetup(b)
	repo := NewPostgresRepository(testDB.DB)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		host := &models.Host{
			Address: fmt.Sprintf("10.0.%d.%d:8080", b.N/256, b.N%256),
			Status:  models.HostStatusReady,
			TotalResources: models.HostResources{
				MemoryMb:      32768,
				CpuMillicores: 16000,
				DiskMb:        102400,
			},
			AvailableResources: models.HostResources{
				MemoryMb:      32768,
				CpuMillicores: 16000,
				DiskMb:        102400,
			},
			ActiveSandboxes: 0,
			LastHeartbeat:   time.Now(),
			SupportedTiers:  []string{"standard", "hardened"},
		}
		if err := repo.CreateHost(ctx, host); err != nil {
			b.Fatalf("CreateHost: %v", err)
		}
	}
}

func BenchmarkDB_PlaceAndDecrement(b *testing.B) {
	benchSetup(b)
	repo := NewPostgresRepository(testDB.DB)
	ctx := context.Background()

	// Seed hosts with large capacity so placement never runs out.
	for i := 0; i < 5; i++ {
		seedHost(b, testDB.DB, fmt.Sprintf("place-%d:8080", i),
			1<<30, // ~1TB memory to avoid exhaustion
			1<<20, // large CPU
			1<<30, // large disk
			[]string{"standard"},
		)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := repo.PlaceAndDecrement(ctx, 512, 1000, 1024, "standard")
		if err != nil {
			b.Fatalf("PlaceAndDecrement: %v", err)
		}
	}
}

func BenchmarkDB_PlaceAndDecrement_Concurrent(b *testing.B) {
	benchSetup(b)
	repo := NewPostgresRepository(testDB.DB)
	ctx := context.Background()

	// Seed 10 large-capacity hosts.
	for i := 0; i < 10; i++ {
		seedHost(b, testDB.DB, fmt.Sprintf("concurrent-%d:8080", i),
			1<<30, 1<<20, 1<<30,
			[]string{"standard"},
		)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := repo.PlaceAndDecrement(ctx, 256, 500, 512, "standard")
			if err != nil {
				b.Fatalf("PlaceAndDecrement: %v", err)
			}
		}
	})
}

func BenchmarkDB_Heartbeat(b *testing.B) {
	benchSetup(b)
	repo := NewPostgresRepository(testDB.DB)
	ctx := context.Background()

	hostID := seedHost(b, testDB.DB, "heartbeat-host:8080", 32768, 16000, 102400, []string{"standard"})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := repo.UpdateHeartbeat(ctx, hostID, models.HostResources{
			MemoryMb:      28000,
			CpuMillicores: 14000,
			DiskMb:        90000,
		}, 5, []string{"standard"})
		if err != nil {
			b.Fatalf("UpdateHeartbeat: %v", err)
		}
	}
}

func BenchmarkDB_ListHosts_500(b *testing.B) {
	benchSetup(b)
	repo := NewPostgresRepository(testDB.DB)
	ctx := context.Background()

	for i := 0; i < 500; i++ {
		seedHost(b, testDB.DB, fmt.Sprintf("list-%d:8080", i), 8192, 4000, 50000, []string{"standard"})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		hosts, err := repo.ListHosts(ctx, "")
		if err != nil {
			b.Fatalf("ListHosts: %v", err)
		}
		if len(hosts) != 500 {
			b.Fatalf("expected 500 hosts, got %d", len(hosts))
		}
	}
}

func BenchmarkDB_ClaimWarmSlot(b *testing.B) {
	benchSetup(b)
	repo := NewPostgresRepository(testDB.DB)
	ctx := context.Background()

	// Configure warm pool.
	if err := repo.UpsertWarmPoolConfig(ctx, &models.WarmPoolConfig{
		IsolationTier: "hardened",
		TargetCount:   1000,
		MemoryMb:      512,
		CpuMillicores: 1000,
		DiskMb:        10240,
	}); err != nil {
		b.Fatalf("UpsertWarmPoolConfig: %v", err)
	}

	hostID := seedHost(b, testDB.DB, "warm-host:8080", 1<<20, 1<<20, 1<<20, []string{"hardened"})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		// Seed a ready slot to claim.
		if err := repo.CreateWarmSlot(ctx, &models.WarmPoolSlot{
			HostID:        hostID,
			IsolationTier: "hardened",
			MemoryMb:      512,
			CpuMillicores: 1000,
			DiskMb:        10240,
			Status:        "ready",
		}); err != nil {
			b.Fatalf("CreateWarmSlot: %v", err)
		}
		b.StartTimer()

		slot, err := repo.ClaimWarmSlot(ctx, "hardened")
		if err != nil {
			b.Fatalf("ClaimWarmSlot: %v", err)
		}
		if slot == nil {
			b.Fatal("expected slot, got nil")
		}
	}
}

func BenchmarkDB_GetCapacity(b *testing.B) {
	benchSetup(b)
	repo := NewPostgresRepository(testDB.DB)
	ctx := context.Background()

	// Seed 100 hosts with mixed tiers.
	tiers := [][]string{
		{"standard"},
		{"standard", "hardened"},
		{"hardened"},
		{"standard", "hardened", "isolated"},
	}
	var counter atomic.Int64
	for i := 0; i < 100; i++ {
		_ = counter.Add(1)
		seedHost(b, testDB.DB, fmt.Sprintf("cap-%d:8080", i),
			8192, 4000, 50000, tiers[i%len(tiers)])
	}

	// Configure warm pools for each tier.
	for _, tier := range []string{"standard", "hardened", "isolated"} {
		if err := repo.UpsertWarmPoolConfig(ctx, &models.WarmPoolConfig{
			IsolationTier: tier,
			TargetCount:   10,
			MemoryMb:      512,
			CpuMillicores: 1000,
			DiskMb:        10240,
		}); err != nil {
			b.Fatalf("UpsertWarmPoolConfig %s: %v", tier, err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		tierCaps, total, ready, err := repo.GetCapacity(ctx)
		if err != nil {
			b.Fatalf("GetCapacity: %v", err)
		}
		if total != 100 {
			b.Fatalf("totalHosts = %d, want 100", total)
		}
		if ready != 100 {
			b.Fatalf("readyHosts = %d, want 100", ready)
		}
		if len(tierCaps) == 0 {
			b.Fatal("expected tier capacities")
		}
	}
}

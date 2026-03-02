//go:build integration

package compute

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/testutil"
)

var testDB *testutil.TestDB

func TestMain(m *testing.M) {
	testDB = testutil.MustSetupTestDB()
	code := m.Run()
	testDB.Cleanup()
	os.Exit(code)
}

func setup(t *testing.T) (*PostgresRepository, *sql.DB) {
	t.Helper()
	testutil.TruncateAll(t, testDB.DB)
	return NewPostgresRepository(testDB.DB), testDB.DB
}

func createHost(t *testing.T, repo *PostgresRepository, address string, status models.HostStatus, memMb int64, cpuMilli int32, diskMb int64) *models.Host {
	t.Helper()
	h := &models.Host{
		Address: address,
		Status:  status,
		TotalResources: models.HostResources{
			MemoryMb:      memMb,
			CpuMillicores: cpuMilli,
			DiskMb:        diskMb,
		},
		AvailableResources: models.HostResources{
			MemoryMb:      memMb,
			CpuMillicores: cpuMilli,
			DiskMb:        diskMb,
		},
		ActiveSandboxes: 0,
		LastHeartbeat:   time.Now(),
	}
	if err := repo.CreateHost(context.Background(), h); err != nil {
		t.Fatalf("CreateHost %s: %v", address, err)
	}
	return h
}

func TestInteg_CreateAndGetHost_Resources(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	host := &models.Host{
		Address: "10.0.0.1:8080",
		Status:  models.HostStatusReady,
		TotalResources: models.HostResources{
			MemoryMb:      32768,
			CpuMillicores: 16000,
			DiskMb:        102400,
		},
		AvailableResources: models.HostResources{
			MemoryMb:      28000,
			CpuMillicores: 14000,
			DiskMb:        90000,
		},
		ActiveSandboxes: 3,
		LastHeartbeat:   time.Now().Truncate(time.Microsecond),
	}
	if err := repo.CreateHost(ctx, host); err != nil {
		t.Fatalf("CreateHost: %v", err)
	}

	if host.ID == "" {
		t.Fatal("expected server-generated ID")
	}

	got, err := repo.GetHost(ctx, host.ID)
	if err != nil {
		t.Fatalf("GetHost: %v", err)
	}
	if got == nil {
		t.Fatal("expected host, got nil")
	}
	if got.TotalResources.MemoryMb != 32768 {
		t.Errorf("TotalResources.MemoryMb = %d, want 32768", got.TotalResources.MemoryMb)
	}
	if got.AvailableResources.CpuMillicores != 14000 {
		t.Errorf("AvailableResources.CpuMillicores = %d, want 14000", got.AvailableResources.CpuMillicores)
	}
	if got.ActiveSandboxes != 3 {
		t.Errorf("ActiveSandboxes = %d, want 3", got.ActiveSandboxes)
	}
	if got.Address != "10.0.0.1:8080" {
		t.Errorf("Address = %q, want %q", got.Address, "10.0.0.1:8080")
	}
}

func TestInteg_GetHost_NotFound(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	got, err := repo.GetHost(ctx, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestInteg_ListHosts_StatusFilter(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	createHost(t, repo, "10.0.0.1:8080", models.HostStatusReady, 8192, 4000, 50000)
	createHost(t, repo, "10.0.0.2:8080", models.HostStatusDraining, 8192, 4000, 50000)
	createHost(t, repo, "10.0.0.3:8080", models.HostStatusReady, 8192, 4000, 50000)

	// All hosts
	all, err := repo.ListHosts(ctx, "")
	if err != nil {
		t.Fatalf("ListHosts all: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("all count = %d, want 3", len(all))
	}

	// Ready only
	ready, err := repo.ListHosts(ctx, models.HostStatusReady)
	if err != nil {
		t.Fatalf("ListHosts ready: %v", err)
	}
	if len(ready) != 2 {
		t.Errorf("ready count = %d, want 2", len(ready))
	}
}

func TestInteg_SetHostStatus_RowsAffected(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	host := createHost(t, repo, "10.0.0.1:8080", models.HostStatusReady, 8192, 4000, 50000)

	if err := repo.SetHostStatus(ctx, host.ID, models.HostStatusDraining); err != nil {
		t.Fatalf("SetHostStatus: %v", err)
	}

	got, err := repo.GetHost(ctx, host.ID)
	if err != nil {
		t.Fatalf("GetHost: %v", err)
	}
	if got.Status != models.HostStatusDraining {
		t.Errorf("status = %q, want %q", got.Status, models.HostStatusDraining)
	}

	// Non-existent host
	err = repo.SetHostStatus(ctx, "00000000-0000-0000-0000-000000000000", models.HostStatusOffline)
	if err != ErrHostNotFound {
		t.Errorf("error = %v, want ErrHostNotFound", err)
	}
}

func TestInteg_PlaceAndDecrement_FirstFit(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	// Create hosts with increasing resources — placement should pick smallest sufficient
	createHost(t, repo, "small", models.HostStatusReady, 2048, 2000, 10000)
	createHost(t, repo, "medium", models.HostStatusReady, 4096, 4000, 20000)
	createHost(t, repo, "large", models.HostStatusReady, 8192, 8000, 40000)

	// Request that fits medium+ — atomically selects and decrements
	got, err := repo.PlaceAndDecrement(ctx, 3000, 3000, 15000, "")
	if err != nil {
		t.Fatalf("PlaceAndDecrement: %v", err)
	}
	if got.Address != "medium" {
		t.Errorf("placed on %q, want %q (smallest sufficient)", got.Address, "medium")
	}
	// Verify resources were decremented
	if got.AvailableResources.MemoryMb != 1096 {
		t.Errorf("AvailableMemory = %d, want 1096", got.AvailableResources.MemoryMb)
	}
	if got.ActiveSandboxes != 1 {
		t.Errorf("ActiveSandboxes = %d, want 1", got.ActiveSandboxes)
	}
}

func TestInteg_PlaceAndDecrement_NoCapacity(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	createHost(t, repo, "tiny", models.HostStatusReady, 512, 500, 1024)

	_, err := repo.PlaceAndDecrement(ctx, 8192, 4000, 50000, "")
	if err != ErrNoCapacity {
		t.Errorf("error = %v, want ErrNoCapacity", err)
	}
}

func TestInteg_PlaceAndDecrement_SkipsDraining(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	createHost(t, repo, "draining-host", models.HostStatusDraining, 16384, 16000, 100000)
	createHost(t, repo, "ready-host", models.HostStatusReady, 4096, 4000, 20000)

	got, err := repo.PlaceAndDecrement(ctx, 2048, 2000, 10000, "")
	if err != nil {
		t.Fatalf("PlaceAndDecrement: %v", err)
	}
	if got.Address != "ready-host" {
		t.Errorf("placed on %q, want %q (should skip draining)", got.Address, "ready-host")
	}
}

func createHostWithTiers(t *testing.T, repo *PostgresRepository, address string, status models.HostStatus, memMb int64, cpuMilli int32, diskMb int64, tiers []string) *models.Host {
	t.Helper()
	h := &models.Host{
		Address: address,
		Status:  status,
		TotalResources: models.HostResources{
			MemoryMb:      memMb,
			CpuMillicores: cpuMilli,
			DiskMb:        diskMb,
		},
		AvailableResources: models.HostResources{
			MemoryMb:      memMb,
			CpuMillicores: cpuMilli,
			DiskMb:        diskMb,
		},
		ActiveSandboxes: 0,
		LastHeartbeat:   time.Now(),
		SupportedTiers:  tiers,
	}
	if err := repo.CreateHost(context.Background(), h); err != nil {
		t.Fatalf("CreateHost %s: %v", address, err)
	}
	return h
}

// --- Warm Pool Integration Tests ---

func TestInteg_UpsertWarmPoolConfig_InsertAndUpdate(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	cfg := &models.WarmPoolConfig{
		IsolationTier: "hardened",
		TargetCount:   5,
		MemoryMb:      512,
		CpuMillicores: 1000,
		DiskMb:        10240,
	}

	// Insert
	if err := repo.UpsertWarmPoolConfig(ctx, cfg); err != nil {
		t.Fatalf("UpsertWarmPoolConfig insert: %v", err)
	}

	configs, err := repo.ListWarmPoolConfigs(ctx)
	if err != nil {
		t.Fatalf("ListWarmPoolConfigs: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("configs count = %d, want 1", len(configs))
	}
	if configs[0].TargetCount != 5 {
		t.Errorf("TargetCount = %d, want 5", configs[0].TargetCount)
	}
	if configs[0].MemoryMb != 512 {
		t.Errorf("MemoryMb = %d, want 512", configs[0].MemoryMb)
	}

	// Update (upsert on same tier)
	cfg.TargetCount = 10
	cfg.MemoryMb = 1024
	if err := repo.UpsertWarmPoolConfig(ctx, cfg); err != nil {
		t.Fatalf("UpsertWarmPoolConfig update: %v", err)
	}

	configs, err = repo.ListWarmPoolConfigs(ctx)
	if err != nil {
		t.Fatalf("ListWarmPoolConfigs after update: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("configs count = %d after upsert, want 1", len(configs))
	}
	if configs[0].TargetCount != 10 {
		t.Errorf("TargetCount after update = %d, want 10", configs[0].TargetCount)
	}
	if configs[0].MemoryMb != 1024 {
		t.Errorf("MemoryMb after update = %d, want 1024", configs[0].MemoryMb)
	}
}

func TestInteg_ListWarmPoolConfigs_MultipleTiers(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	for _, tier := range []string{"hardened", "isolated", "standard"} {
		if err := repo.UpsertWarmPoolConfig(ctx, &models.WarmPoolConfig{
			IsolationTier: tier,
			TargetCount:   3,
			MemoryMb:      512,
			CpuMillicores: 1000,
			DiskMb:        10240,
		}); err != nil {
			t.Fatalf("UpsertWarmPoolConfig %s: %v", tier, err)
		}
	}

	configs, err := repo.ListWarmPoolConfigs(ctx)
	if err != nil {
		t.Fatalf("ListWarmPoolConfigs: %v", err)
	}
	if len(configs) != 3 {
		t.Fatalf("configs count = %d, want 3", len(configs))
	}
	// Sorted by isolation_tier
	if configs[0].IsolationTier != "hardened" {
		t.Errorf("first tier = %q, want %q", configs[0].IsolationTier, "hardened")
	}
	if configs[2].IsolationTier != "standard" {
		t.Errorf("last tier = %q, want %q", configs[2].IsolationTier, "standard")
	}
}

func TestInteg_CreateAndClaimWarmSlot(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	host := createHostWithTiers(t, repo, "warm-host", models.HostStatusReady, 8192, 4000, 50000, []string{"hardened"})

	slot := &models.WarmPoolSlot{
		HostID:        host.ID,
		IsolationTier: "hardened",
		MemoryMb:      512,
		CpuMillicores: 1000,
		DiskMb:        10240,
		Status:        "ready",
	}
	if err := repo.CreateWarmSlot(ctx, slot); err != nil {
		t.Fatalf("CreateWarmSlot: %v", err)
	}
	if slot.ID == "" {
		t.Fatal("expected slot ID to be set")
	}

	// Count ready slots
	count, err := repo.CountReadySlots(ctx, "hardened")
	if err != nil {
		t.Fatalf("CountReadySlots: %v", err)
	}
	if count != 1 {
		t.Errorf("ready count = %d, want 1", count)
	}

	// Claim the slot
	claimed, err := repo.ClaimWarmSlot(ctx, "hardened")
	if err != nil {
		t.Fatalf("ClaimWarmSlot: %v", err)
	}
	if claimed == nil {
		t.Fatal("expected claimed slot, got nil")
	}
	if claimed.Status != "claimed" {
		t.Errorf("claimed status = %q, want %q", claimed.Status, "claimed")
	}
	if claimed.HostID != host.ID {
		t.Errorf("claimed host_id = %q, want %q", claimed.HostID, host.ID)
	}

	// Count should now be 0
	count, err = repo.CountReadySlots(ctx, "hardened")
	if err != nil {
		t.Fatalf("CountReadySlots after claim: %v", err)
	}
	if count != 0 {
		t.Errorf("ready count after claim = %d, want 0", count)
	}

	// Second claim should return nil (no ready slots)
	second, err := repo.ClaimWarmSlot(ctx, "hardened")
	if err != nil {
		t.Fatalf("ClaimWarmSlot second: %v", err)
	}
	if second != nil {
		t.Errorf("expected nil for second claim, got %+v", second)
	}
}

func TestInteg_ClaimWarmSlot_WrongTier(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	host := createHostWithTiers(t, repo, "warm-host", models.HostStatusReady, 8192, 4000, 50000, []string{"standard"})

	// Create a standard slot
	if err := repo.CreateWarmSlot(ctx, &models.WarmPoolSlot{
		HostID:        host.ID,
		IsolationTier: "standard",
		MemoryMb:      512,
		CpuMillicores: 1000,
		DiskMb:        10240,
		Status:        "ready",
	}); err != nil {
		t.Fatalf("CreateWarmSlot: %v", err)
	}

	// Claim for a different tier should return nil
	got, err := repo.ClaimWarmSlot(ctx, "hardened")
	if err != nil {
		t.Fatalf("ClaimWarmSlot wrong tier: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for wrong tier, got %+v", got)
	}

	// Original standard slot should still be ready
	count, err := repo.CountReadySlots(ctx, "standard")
	if err != nil {
		t.Fatalf("CountReadySlots: %v", err)
	}
	if count != 1 {
		t.Errorf("standard ready count = %d, want 1", count)
	}
}

func TestInteg_ClaimWarmSlot_MultipleSlots_FIFO(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	host := createHostWithTiers(t, repo, "warm-host", models.HostStatusReady, 16384, 8000, 100000, []string{"hardened"})

	// Create 3 slots
	slotIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		slot := &models.WarmPoolSlot{
			HostID:        host.ID,
			IsolationTier: "hardened",
			MemoryMb:      512,
			CpuMillicores: 1000,
			DiskMb:        10240,
			Status:        "ready",
		}
		if err := repo.CreateWarmSlot(ctx, slot); err != nil {
			t.Fatalf("CreateWarmSlot %d: %v", i, err)
		}
		slotIDs[i] = slot.ID
	}

	count, err := repo.CountReadySlots(ctx, "hardened")
	if err != nil {
		t.Fatalf("CountReadySlots: %v", err)
	}
	if count != 3 {
		t.Errorf("ready count = %d, want 3", count)
	}

	// Claim all 3 sequentially
	for i := 0; i < 3; i++ {
		claimed, err := repo.ClaimWarmSlot(ctx, "hardened")
		if err != nil {
			t.Fatalf("ClaimWarmSlot %d: %v", i, err)
		}
		if claimed == nil {
			t.Fatalf("claim %d: expected slot, got nil", i)
		}
		if claimed.Status != "claimed" {
			t.Errorf("claim %d: status = %q, want %q", i, claimed.Status, "claimed")
		}
	}

	// 4th claim should return nil
	got, err := repo.ClaimWarmSlot(ctx, "hardened")
	if err != nil {
		t.Fatalf("ClaimWarmSlot exhausted: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil when pool exhausted, got %+v", got)
	}
}

func TestInteg_CleanExpiredSlots_RemovesOfflineHostSlots(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	readyHost := createHostWithTiers(t, repo, "alive-host", models.HostStatusReady, 8192, 4000, 50000, []string{"hardened"})
	offlineHost := createHostWithTiers(t, repo, "dead-host", models.HostStatusOffline, 8192, 4000, 50000, []string{"hardened"})

	// Create slots on both hosts
	for _, hid := range []string{readyHost.ID, offlineHost.ID} {
		if err := repo.CreateWarmSlot(ctx, &models.WarmPoolSlot{
			HostID:        hid,
			IsolationTier: "hardened",
			MemoryMb:      512,
			CpuMillicores: 1000,
			DiskMb:        10240,
			Status:        "ready",
		}); err != nil {
			t.Fatalf("CreateWarmSlot on %s: %v", hid, err)
		}
	}

	count, err := repo.CountReadySlots(ctx, "hardened")
	if err != nil {
		t.Fatalf("CountReadySlots before cleanup: %v", err)
	}
	if count != 2 {
		t.Errorf("ready count before cleanup = %d, want 2", count)
	}

	// Clean expired slots
	cleaned, err := repo.CleanExpiredSlots(ctx)
	if err != nil {
		t.Fatalf("CleanExpiredSlots: %v", err)
	}
	if cleaned != 1 {
		t.Errorf("cleaned = %d, want 1", cleaned)
	}

	// Only the alive host's slot should remain
	count, err = repo.CountReadySlots(ctx, "hardened")
	if err != nil {
		t.Fatalf("CountReadySlots after cleanup: %v", err)
	}
	if count != 1 {
		t.Errorf("ready count after cleanup = %d, want 1", count)
	}
}

func TestInteg_GetCapacity_AggregatesAll(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	// Create hosts with different tiers
	createHostWithTiers(t, repo, "host-a", models.HostStatusReady, 8192, 4000, 50000, []string{"standard", "hardened"})
	createHostWithTiers(t, repo, "host-b", models.HostStatusReady, 16384, 8000, 100000, []string{"standard"})
	createHostWithTiers(t, repo, "host-c", models.HostStatusOffline, 8192, 4000, 50000, []string{"hardened"})

	// Configure warm pools
	if err := repo.UpsertWarmPoolConfig(ctx, &models.WarmPoolConfig{
		IsolationTier: "hardened",
		TargetCount:   5,
		MemoryMb:      512,
		CpuMillicores: 1000,
		DiskMb:        10240,
	}); err != nil {
		t.Fatalf("UpsertWarmPoolConfig: %v", err)
	}

	tiers, totalHosts, readyHosts, err := repo.GetCapacity(ctx)
	if err != nil {
		t.Fatalf("GetCapacity: %v", err)
	}

	if totalHosts != 3 {
		t.Errorf("totalHosts = %d, want 3", totalHosts)
	}
	if readyHosts != 2 {
		t.Errorf("readyHosts = %d, want 2", readyHosts)
	}

	// Should have standard + hardened tiers from ready hosts
	tierMap := make(map[string]models.TierCapacity)
	for _, tc := range tiers {
		tierMap[tc.IsolationTier] = tc
	}

	// Standard: host-a (8192) + host-b (16384) = 24576 MB
	std, ok := tierMap["standard"]
	if !ok {
		t.Fatal("missing 'standard' tier in capacity")
	}
	if std.HostsSupporting != 2 {
		t.Errorf("standard hosts = %d, want 2", std.HostsSupporting)
	}
	if std.AvailableMemoryMb != 24576 {
		t.Errorf("standard memory = %d, want 24576", std.AvailableMemoryMb)
	}

	// Hardened: only host-a is ready (host-c is offline)
	hard, ok := tierMap["hardened"]
	if !ok {
		t.Fatal("missing 'hardened' tier in capacity")
	}
	if hard.HostsSupporting != 1 {
		t.Errorf("hardened hosts = %d, want 1", hard.HostsSupporting)
	}
	if hard.WarmSlotsTarget != 5 {
		t.Errorf("hardened warm target = %d, want 5", hard.WarmSlotsTarget)
	}
	if hard.WarmSlotsReady != 0 {
		t.Errorf("hardened warm ready = %d, want 0", hard.WarmSlotsReady)
	}
}

func TestInteg_PlaceAndDecrement_TierFiltering(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	createHostWithTiers(t, repo, "std-only", models.HostStatusReady, 8192, 4000, 50000, []string{"standard"})
	createHostWithTiers(t, repo, "hardened-host", models.HostStatusReady, 4096, 4000, 20000, []string{"standard", "hardened"})

	// Request hardened tier — should skip std-only even though it has more resources
	got, err := repo.PlaceAndDecrement(ctx, 2048, 2000, 10000, "hardened")
	if err != nil {
		t.Fatalf("PlaceAndDecrement hardened: %v", err)
	}
	if got.Address != "hardened-host" {
		t.Errorf("placed on %q, want %q", got.Address, "hardened-host")
	}
}

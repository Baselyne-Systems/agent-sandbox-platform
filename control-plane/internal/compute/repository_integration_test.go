//go:build integration

package compute

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/testutil"
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

func TestInteg_FindHostForPlacement_FirstFit(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	// Create hosts with increasing resources — placement should pick smallest sufficient
	createHost(t, repo, "small", models.HostStatusReady, 2048, 2000, 10000)
	createHost(t, repo, "medium", models.HostStatusReady, 4096, 4000, 20000)
	createHost(t, repo, "large", models.HostStatusReady, 8192, 8000, 40000)

	// Request that fits medium+
	got, err := repo.FindHostForPlacement(ctx, 3000, 3000, 15000)
	if err != nil {
		t.Fatalf("FindHostForPlacement: %v", err)
	}
	if got.Address != "medium" {
		t.Errorf("placed on %q, want %q (smallest sufficient)", got.Address, "medium")
	}
}

func TestInteg_FindHostForPlacement_NoCapacity(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	createHost(t, repo, "tiny", models.HostStatusReady, 512, 500, 1024)

	_, err := repo.FindHostForPlacement(ctx, 8192, 4000, 50000)
	if err != ErrNoCapacity {
		t.Errorf("error = %v, want ErrNoCapacity", err)
	}
}

func TestInteg_FindHostForPlacement_SkipsDraining(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	createHost(t, repo, "draining-host", models.HostStatusDraining, 16384, 16000, 100000)
	createHost(t, repo, "ready-host", models.HostStatusReady, 4096, 4000, 20000)

	got, err := repo.FindHostForPlacement(ctx, 2048, 2000, 10000)
	if err != nil {
		t.Fatalf("FindHostForPlacement: %v", err)
	}
	if got.Address != "ready-host" {
		t.Errorf("placed on %q, want %q (should skip draining)", got.Address, "ready-host")
	}
}

func TestInteg_DecrementResources(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	host := createHost(t, repo, "decrement-host", models.HostStatusReady, 8192, 8000, 40000)

	if err := repo.DecrementAvailableResources(ctx, host.ID, 1024, 1000, 5000); err != nil {
		t.Fatalf("DecrementAvailableResources: %v", err)
	}

	got, err := repo.GetHost(ctx, host.ID)
	if err != nil {
		t.Fatalf("GetHost: %v", err)
	}
	if got.AvailableResources.MemoryMb != 7168 {
		t.Errorf("AvailableMemory = %d, want 7168", got.AvailableResources.MemoryMb)
	}
	if got.AvailableResources.CpuMillicores != 7000 {
		t.Errorf("AvailableCpu = %d, want 7000", got.AvailableResources.CpuMillicores)
	}
	if got.AvailableResources.DiskMb != 35000 {
		t.Errorf("AvailableDisk = %d, want 35000", got.AvailableResources.DiskMb)
	}
	if got.ActiveSandboxes != 1 {
		t.Errorf("ActiveSandboxes = %d, want 1", got.ActiveSandboxes)
	}
}

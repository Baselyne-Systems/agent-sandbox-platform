package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/compute"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

func TestHostRegistrationAndHeartbeat(t *testing.T) {
	clean(t)
	ctx := context.Background()

	host, err := computeSvc.RegisterHost(ctx, "fleet-host.local:9090", models.HostResources{
		MemoryMb:      8192,
		CpuMillicores: 8000,
		DiskMb:        20480,
	}, []string{"standard", "hardened"})
	if err != nil {
		t.Fatalf("register host: %v", err)
	}
	if host.Status != models.HostStatusReady {
		t.Fatalf("expected ready, got %s", host.Status)
	}
	if host.TotalResources.MemoryMb != 8192 {
		t.Fatalf("expected 8192 MB memory, got %d", host.TotalResources.MemoryMb)
	}

	// Heartbeat updates AvailableResources (not TotalResources).
	updated, err := computeSvc.Heartbeat(ctx, host.ID, models.HostResources{
		MemoryMb:      7000,
		CpuMillicores: 6000,
		DiskMb:        18000,
	}, 2, []string{"standard", "hardened"})
	if err != nil {
		t.Fatalf("heartbeat: %v", err)
	}
	if updated.AvailableResources.MemoryMb != 7000 {
		t.Fatalf("expected 7000 MB available after heartbeat, got %d", updated.AvailableResources.MemoryMb)
	}
	if updated.ActiveSandboxes != 2 {
		t.Fatalf("expected 2 active sandboxes, got %d", updated.ActiveSandboxes)
	}
}

func TestWorkspacePlacement(t *testing.T) {
	clean(t)
	ctx := context.Background()

	registerHost(t, ctx, "place-host.local:9090", 4096, 4000, 10240, []string{"standard", "hardened"})

	hostID, address, err := computeSvc.PlaceWorkspace(ctx, 512, 500, 1024, "standard")
	if err != nil {
		t.Fatalf("place workspace: %v", err)
	}
	if hostID == "" {
		t.Fatal("expected non-empty host ID")
	}
	if address == "" {
		t.Fatal("expected non-empty address")
	}
}

func TestPlacementExhaustsCapacity(t *testing.T) {
	clean(t)
	ctx := context.Background()

	registerHost(t, ctx, "exhaust-host.local:9090", 1024, 2000, 5120, []string{"standard"})

	_, _, err := computeSvc.PlaceWorkspace(ctx, 512, 500, 1024, "standard")
	if err != nil {
		t.Fatalf("place 1: %v", err)
	}

	_, _, err = computeSvc.PlaceWorkspace(ctx, 512, 500, 1024, "standard")
	if err != nil {
		t.Fatalf("place 2: %v", err)
	}

	_, _, err = computeSvc.PlaceWorkspace(ctx, 512, 500, 1024, "standard")
	if err != compute.ErrNoCapacity {
		t.Fatalf("expected ErrNoCapacity, got %v", err)
	}
}

func TestTierAwarePlacement(t *testing.T) {
	clean(t)
	ctx := context.Background()

	registerHost(t, ctx, "tier-a.local:9090", 4096, 4000, 10240, []string{"standard"})
	registerHost(t, ctx, "tier-b.local:9090", 4096, 4000, 10240, []string{"standard", "hardened"})

	hostID, _, err := computeSvc.PlaceWorkspace(ctx, 512, 500, 1024, "hardened")
	if err != nil {
		t.Fatalf("place hardened: %v", err)
	}

	hosts, err := computeSvc.ListHosts(ctx, models.HostStatusReady)
	if err != nil {
		t.Fatalf("list hosts: %v", err)
	}
	var hostB *models.Host
	for i := range hosts {
		if hosts[i].Address == "tier-b.local:9090" {
			hostB = &hosts[i]
			break
		}
	}
	if hostB == nil {
		t.Fatal("host B not found")
	}
	if hostID != hostB.ID {
		t.Fatalf("expected placement on host B (%s), got %s", hostB.ID, hostID)
	}

	_, _, err = computeSvc.PlaceWorkspace(ctx, 512, 500, 1024, "isolated")
	if err != compute.ErrNoCapacity {
		t.Fatalf("expected ErrNoCapacity for isolated tier, got %v", err)
	}
}

func TestWarmPoolWorkflow(t *testing.T) {
	clean(t)
	ctx := context.Background()

	registerHost(t, ctx, "warm-host.local:9090", 8192, 8000, 20480, []string{"standard", "hardened"})

	cfg, err := computeSvc.ConfigureWarmPool(ctx, &models.WarmPoolConfig{
		IsolationTier: "standard",
		TargetCount:   3,
		MemoryMb:      512,
		CpuMillicores: 500,
		DiskMb:        1024,
	})
	if err != nil {
		t.Fatalf("configure warm pool: %v", err)
	}
	if cfg.TargetCount != 3 {
		t.Fatalf("expected target 3, got %d", cfg.TargetCount)
	}

	worker := compute.NewWarmPoolWorker(computeRepo, compute.WarmPoolWorkerConfig{
		Interval: 100 * time.Millisecond,
	})
	workerCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	worker.Run(workerCtx)

	count, err := computeRepo.CountReadySlots(ctx, "standard")
	if err != nil {
		t.Fatalf("count ready slots: %v", err)
	}
	if count < 1 {
		t.Fatalf("expected at least 1 warm slot, got %d", count)
	}
}

func TestHostLivenessDetection(t *testing.T) {
	clean(t)
	ctx := context.Background()

	host := registerHost(t, ctx, "liveness-host.local:9090", 4096, 4000, 10240, []string{"standard"})

	_, err := db.Exec("UPDATE hosts SET last_heartbeat = $1 WHERE id = $2",
		time.Now().Add(-10*time.Minute), host.ID)
	if err != nil {
		t.Fatalf("backdate heartbeat: %v", err)
	}

	worker := compute.NewLivenessWorker(computeRepo, compute.LivenessWorkerConfig{
		Interval:         100 * time.Millisecond,
		HeartbeatTimeout: 3 * time.Minute,
	})
	workerCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	worker.Run(workerCtx)

	hosts, err := computeSvc.ListHosts(ctx, models.HostStatusOffline)
	if err != nil {
		t.Fatalf("list offline hosts: %v", err)
	}
	found := false
	for _, h := range hosts {
		if h.ID == host.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected host to be marked offline after stale heartbeat")
	}
}

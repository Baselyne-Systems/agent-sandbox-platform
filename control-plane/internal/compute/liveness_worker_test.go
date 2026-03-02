package compute

import (
	"context"
	"testing"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
	"go.uber.org/zap"
)

func TestLivenessWorker_MarksStaleHostsOffline(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Register two hosts.
	host1, _ := svc.RegisterHost(ctx, "host1:9090", models.HostResources{
		MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240,
	}, nil)
	svc.RegisterHost(ctx, "host2:9090", models.HostResources{
		MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240,
	}, nil)

	// Backdate host1's heartbeat to make it stale.
	repo.hosts[host1.ID].LastHeartbeat = time.Now().Add(-5 * time.Minute)

	worker := NewLivenessWorker(repo, LivenessWorkerConfig{
		Interval:         time.Second,
		HeartbeatTimeout: 3 * time.Minute,
		Logger:           zap.NewNop(),
	})

	// Run a single sweep.
	worker.sweep(ctx)

	// host1 should be offline, host2 should still be ready.
	hosts, _ := svc.ListHosts(ctx, models.HostStatusOffline)
	if len(hosts) != 1 {
		t.Errorf("expected 1 offline host, got %d", len(hosts))
	}
	if hosts[0].ID != host1.ID {
		t.Errorf("expected host1 to be offline, got %s", hosts[0].ID)
	}

	readyHosts, _ := svc.ListHosts(ctx, models.HostStatusReady)
	if len(readyHosts) != 1 {
		t.Errorf("expected 1 ready host, got %d", len(readyHosts))
	}
}

func TestLivenessWorker_NoStaleHosts(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	svc.RegisterHost(ctx, "host1:9090", models.HostResources{
		MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240,
	}, nil)

	worker := NewLivenessWorker(repo, LivenessWorkerConfig{
		Interval:         time.Second,
		HeartbeatTimeout: 3 * time.Minute,
		Logger:           zap.NewNop(),
	})

	// All hosts are fresh — sweep should not change anything.
	worker.sweep(ctx)

	hosts, _ := svc.ListHosts(ctx, models.HostStatusReady)
	if len(hosts) != 1 {
		t.Errorf("expected 1 ready host, got %d", len(hosts))
	}
}

func TestLivenessWorker_RunStopsOnCancel(t *testing.T) {
	repo := newMockRepo()
	ctx, cancel := context.WithCancel(context.Background())

	worker := NewLivenessWorker(repo, LivenessWorkerConfig{
		Interval:         50 * time.Millisecond,
		HeartbeatTimeout: time.Minute,
		Logger:           zap.NewNop(),
	})

	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()

	// Let it tick once.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not stop after context cancellation")
	}
}

func TestLivenessWorker_SkipsDrainingHosts(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Register and then drain a host.
	host1, _ := svc.RegisterHost(ctx, "host1:9090", models.HostResources{
		MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240,
	}, nil)

	// Set to draining and backdate heartbeat.
	repo.hosts[host1.ID].Status = models.HostStatusDraining
	repo.hosts[host1.ID].LastHeartbeat = time.Now().Add(-5 * time.Minute)

	worker := NewLivenessWorker(repo, LivenessWorkerConfig{
		Interval:         time.Second,
		HeartbeatTimeout: 3 * time.Minute,
		Logger:           zap.NewNop(),
	})

	worker.sweep(ctx)

	// Should still be draining (liveness worker only marks ready→offline).
	if repo.hosts[host1.ID].Status != models.HostStatusDraining {
		t.Errorf("expected draining status preserved, got %q", repo.hosts[host1.ID].Status)
	}
}

func TestLivenessWorker_DefaultConfig(t *testing.T) {
	worker := NewLivenessWorker(newMockRepo(), LivenessWorkerConfig{})

	if worker.interval != 60*time.Second {
		t.Errorf("expected default interval 60s, got %v", worker.interval)
	}
	if worker.heartbeatTimeout != 180*time.Second {
		t.Errorf("expected default timeout 180s, got %v", worker.heartbeatTimeout)
	}
}

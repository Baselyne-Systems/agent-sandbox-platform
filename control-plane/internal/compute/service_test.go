package compute

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// mockRepo is a hand-written in-memory Repository for testing.
type mockRepo struct {
	hosts  map[string]*models.Host
	nextID int
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		hosts: make(map[string]*models.Host),
	}
}

func (m *mockRepo) nextUUID() string {
	m.nextID++
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", m.nextID)
}

func (m *mockRepo) CreateHost(_ context.Context, host *models.Host) error {
	host.ID = m.nextUUID()
	cp := *host
	m.hosts[host.ID] = &cp
	return nil
}

func (m *mockRepo) GetHost(_ context.Context, id string) (*models.Host, error) {
	h, ok := m.hosts[id]
	if !ok {
		return nil, nil
	}
	cp := *h
	return &cp, nil
}

func (m *mockRepo) ListHosts(_ context.Context, status models.HostStatus) ([]models.Host, error) {
	var result []models.Host
	for _, h := range m.hosts {
		if status != "" && h.Status != status {
			continue
		}
		cp := *h
		result = append(result, cp)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (m *mockRepo) SetHostStatus(_ context.Context, id string, status models.HostStatus) error {
	h, ok := m.hosts[id]
	if !ok {
		return ErrHostNotFound
	}
	h.Status = status
	return nil
}

func (m *mockRepo) PlaceAndDecrement(_ context.Context, memoryMb int64, cpuMillicores int32, diskMb int64, isolationTier string) (*models.Host, error) {
	var candidates []*models.Host
	for _, h := range m.hosts {
		if h.Status != models.HostStatusReady {
			continue
		}
		if h.AvailableResources.MemoryMb >= memoryMb &&
			h.AvailableResources.CpuMillicores >= cpuMillicores &&
			h.AvailableResources.DiskMb >= diskMb {
			// Filter by supported tiers if isolation tier is specified.
			if isolationTier != "" {
				found := false
				for _, tier := range h.SupportedTiers {
					if tier == isolationTier {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			candidates = append(candidates, h)
		}
	}
	if len(candidates) == 0 {
		return nil, ErrNoCapacity
	}
	// Tier-aware best-fit: prefer hosts with fewer tier capabilities (preserve
	// isolated-capable hosts for workloads that need them), then smallest memory.
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].AvailableResources.MemoryMb == candidates[j].AvailableResources.MemoryMb {
			return len(candidates[i].SupportedTiers) < len(candidates[j].SupportedTiers)
		}
		return candidates[i].AvailableResources.MemoryMb < candidates[j].AvailableResources.MemoryMb
	})
	h := candidates[0]
	h.AvailableResources.MemoryMb -= memoryMb
	h.AvailableResources.CpuMillicores -= cpuMillicores
	h.AvailableResources.DiskMb -= diskMb
	h.ActiveSandboxes++
	h.LastHeartbeat = time.Now()
	cp := *h
	return &cp, nil
}

func (m *mockRepo) MarkStaleHostsOffline(_ context.Context, timeout time.Duration) (int64, error) {
	cutoff := time.Now().Add(-timeout)
	var count int64
	for _, h := range m.hosts {
		if h.Status == models.HostStatusReady && h.LastHeartbeat.Before(cutoff) {
			h.Status = models.HostStatusOffline
			count++
		}
	}
	return count, nil
}

func (m *mockRepo) UpdateHeartbeat(_ context.Context, hostID string, resources models.HostResources, activeSandboxes int32, supportedTiers []string) (*models.Host, error) {
	h, ok := m.hosts[hostID]
	if !ok {
		return nil, ErrHostNotFound
	}
	h.AvailableResources = resources
	h.ActiveSandboxes = activeSandboxes
	h.LastHeartbeat = time.Now()
	if supportedTiers != nil {
		h.SupportedTiers = supportedTiers
	}
	cp := *h
	return &cp, nil
}

// --- RegisterHost tests ---

func TestRegisterHost_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	host, err := svc.RegisterHost(context.Background(), "host1.example.com:9090", models.HostResources{
		MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if host.ID == "" {
		t.Error("expected host ID to be set")
	}
	if host.Status != models.HostStatusReady {
		t.Errorf("expected status ready, got %q", host.Status)
	}
	if host.AvailableResources.MemoryMb != 16384 {
		t.Errorf("expected available memory 16384, got %d", host.AvailableResources.MemoryMb)
	}
}

func TestRegisterHost_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.RegisterHost(ctx, "", models.HostResources{MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 1024}, nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty address, got: %v", err)
	}
}

func TestRegisterHost_InvalidResources(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 0, CpuMillicores: 1000, DiskMb: 1024}, nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero memory, got: %v", err)
	}
	if _, err := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 1024, CpuMillicores: 0, DiskMb: 1024}, nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero cpu, got: %v", err)
	}
	if _, err := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 0}, nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero disk, got: %v", err)
	}
}

// --- DeregisterHost tests ---

func TestDeregisterHost_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	host, _ := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 1024}, nil)

	if err := svc.DeregisterHost(ctx, host.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hosts, _ := svc.ListHosts(ctx, models.HostStatusOffline)
	if len(hosts) != 1 {
		t.Errorf("expected 1 offline host, got %d", len(hosts))
	}
}

func TestDeregisterHost_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	err := svc.DeregisterHost(context.Background(), "nonexistent")
	if !errors.Is(err, ErrHostNotFound) {
		t.Errorf("expected ErrHostNotFound, got: %v", err)
	}
}

func TestDeregisterHost_EmptyID(t *testing.T) {
	svc := NewService(newMockRepo())
	err := svc.DeregisterHost(context.Background(), "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty ID, got: %v", err)
	}
}

// --- ListHosts tests ---

func TestListHosts_All(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 1024}, nil)
	svc.RegisterHost(ctx, "host2:9090", models.HostResources{MemoryMb: 2048, CpuMillicores: 2000, DiskMb: 2048}, nil)

	hosts, err := svc.ListHosts(ctx, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(hosts))
	}
}

func TestListHosts_FilterByStatus(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	host1, _ := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 1024}, nil)
	svc.RegisterHost(ctx, "host2:9090", models.HostResources{MemoryMb: 2048, CpuMillicores: 2000, DiskMb: 2048}, nil)

	svc.DeregisterHost(ctx, host1.ID)

	hosts, err := svc.ListHosts(ctx, models.HostStatusReady)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 1 {
		t.Errorf("expected 1 ready host, got %d", len(hosts))
	}
}

// --- PlaceWorkspace tests ---

func TestPlaceWorkspace_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240}, nil)

	hostID, address, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostID == "" {
		t.Error("expected hostID to be set")
	}
	if address != "host1:9090" {
		t.Errorf("expected address 'host1:9090', got %q", address)
	}
}

func TestPlaceWorkspace_NoCapacity(t *testing.T) {
	svc := NewService(newMockRepo())
	_, _, err := svc.PlaceWorkspace(context.Background(), 512, 500, 1024, "")
	if !errors.Is(err, ErrNoCapacity) {
		t.Errorf("expected ErrNoCapacity, got: %v", err)
	}
}

func TestPlaceWorkspace_InsufficientResources(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 256, CpuMillicores: 200, DiskMb: 512}, nil)

	_, _, err := svc.PlaceWorkspace(ctx, 1024, 500, 1024, "")
	if !errors.Is(err, ErrNoCapacity) {
		t.Errorf("expected ErrNoCapacity for insufficient resources, got: %v", err)
	}
}

func TestPlaceWorkspace_SelectsSmallestFit(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	// Register a large host and a small host.
	svc.RegisterHost(ctx, "large:9090", models.HostResources{MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400}, nil)
	svc.RegisterHost(ctx, "small:9090", models.HostResources{MemoryMb: 2048, CpuMillicores: 2000, DiskMb: 10240}, nil)

	_, address, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if address != "small:9090" {
		t.Errorf("expected smallest fitting host 'small:9090', got %q", address)
	}
}

func TestPlaceWorkspace_SkipsDraining(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	host1, _ := svc.RegisterHost(ctx, "draining:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240}, nil)
	svc.RegisterHost(ctx, "ready:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240}, nil)

	// Set host1 to draining (simulate via repo).
	svc.repo.SetHostStatus(ctx, host1.ID, models.HostStatusDraining)

	_, address, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if address != "ready:9090" {
		t.Errorf("expected to skip draining host, got %q", address)
	}
}

// --- Heartbeat tests ---

func TestHeartbeat_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	host, _ := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240}, nil)

	updated, err := svc.Heartbeat(ctx, host.ID, models.HostResources{
		MemoryMb: 3000, CpuMillicores: 3500, DiskMb: 9000,
	}, 2, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.AvailableResources.MemoryMb != 3000 {
		t.Errorf("expected memory 3000, got %d", updated.AvailableResources.MemoryMb)
	}
	if updated.ActiveSandboxes != 2 {
		t.Errorf("expected active sandboxes 2, got %d", updated.ActiveSandboxes)
	}
}

func TestHeartbeat_UnknownHost(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.Heartbeat(context.Background(), "nonexistent", models.HostResources{
		MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 1024,
	}, 0, nil)
	if !errors.Is(err, ErrHostNotFound) {
		t.Errorf("expected ErrHostNotFound, got: %v", err)
	}
}

func TestHeartbeat_EmptyID(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.Heartbeat(context.Background(), "", models.HostResources{
		MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 1024,
	}, 0, nil)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestHeartbeat_ReturnsCurrentStatus(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	host, _ := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240}, nil)

	// Deregister sets status to offline.
	svc.DeregisterHost(ctx, host.ID)

	updated, err := svc.Heartbeat(ctx, host.ID, models.HostResources{
		MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240,
	}, 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != models.HostStatusOffline {
		t.Errorf("expected status offline, got %q", updated.Status)
	}
}

// --- Isolation tier placement tests ---

func TestPlaceWorkspace_FiltersByTier(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// Register two hosts: one supports standard only, the other supports standard+hardened.
	svc.RegisterHost(ctx, "standard-only:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240}, []string{"standard"})
	svc.RegisterHost(ctx, "standard-hardened:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240}, []string{"standard", "hardened"})

	// Request hardened tier — should only match the second host.
	_, address, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "hardened")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if address != "standard-hardened:9090" {
		t.Errorf("expected 'standard-hardened:9090', got %q", address)
	}
}

func TestPlaceWorkspace_NoHostSupportsTier(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	svc.RegisterHost(ctx, "standard-only:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240}, []string{"standard"})

	// Request isolated tier — no host supports it.
	_, _, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "isolated")
	if !errors.Is(err, ErrNoCapacity) {
		t.Errorf("expected ErrNoCapacity for unsupported tier, got: %v", err)
	}
}

func TestPlaceWorkspace_EmptyTierMatchesAll(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240}, []string{"standard"})

	// Empty isolation tier should match any host.
	_, address, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if address != "host1:9090" {
		t.Errorf("expected 'host1:9090', got %q", address)
	}
}

func TestRegisterHost_DefaultTiers(t *testing.T) {
	svc := NewService(newMockRepo())
	host, err := svc.RegisterHost(context.Background(), "host1:9090", models.HostResources{
		MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(host.SupportedTiers) != 1 || host.SupportedTiers[0] != "standard" {
		t.Errorf("expected default tiers [standard], got %v", host.SupportedTiers)
	}
}

func TestHeartbeat_UpdatesSupportedTiers(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// Register a host with default (standard) tiers.
	host, _ := svc.RegisterHost(ctx, "host1:9090", models.HostResources{
		MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240,
	}, nil)
	if len(host.SupportedTiers) != 1 || host.SupportedTiers[0] != "standard" {
		t.Fatalf("expected initial tiers [standard], got %v", host.SupportedTiers)
	}

	// Heartbeat with updated tiers — host now supports hardened too.
	updated, err := svc.Heartbeat(ctx, host.ID, models.HostResources{
		MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240,
	}, 0, []string{"standard", "hardened"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updated.SupportedTiers) != 2 {
		t.Fatalf("expected 2 tiers after heartbeat, got %v", updated.SupportedTiers)
	}

	// Verify the host can now accept hardened placement requests.
	_, address, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "hardened")
	if err != nil {
		t.Fatalf("expected placement to succeed after tier update: %v", err)
	}
	if address != "host1:9090" {
		t.Errorf("expected 'host1:9090', got %q", address)
	}
}

func TestHeartbeat_NilTiersPreservesExisting(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	host, _ := svc.RegisterHost(ctx, "host1:9090", models.HostResources{
		MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240,
	}, []string{"standard", "hardened"})

	// Heartbeat with nil tiers should not clear them.
	updated, err := svc.Heartbeat(ctx, host.ID, models.HostResources{
		MemoryMb: 3000, CpuMillicores: 3000, DiskMb: 9000,
	}, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updated.SupportedTiers) != 2 {
		t.Errorf("expected tiers preserved (2), got %v", updated.SupportedTiers)
	}
}

func TestPlaceWorkspace_TierAwareBestFit(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// Register two hosts with identical resources but different tier capabilities.
	// "simple" supports only standard; "versatile" supports standard + hardened + isolated.
	svc.RegisterHost(ctx, "versatile:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240}, []string{"standard", "hardened", "isolated"})
	svc.RegisterHost(ctx, "simple:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240}, []string{"standard"})

	// A standard request (or empty tier) should prefer the simpler host,
	// preserving the versatile host for workloads that actually need it.
	_, address, err := svc.PlaceWorkspace(ctx, 512, 500, 1024, "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if address != "simple:9090" {
		t.Errorf("expected tier-aware best-fit to pick 'simple:9090', got %q", address)
	}

	// A hardened request must go to the versatile host (the only one supporting it).
	_, address, err = svc.PlaceWorkspace(ctx, 512, 500, 1024, "hardened")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if address != "versatile:9090" {
		t.Errorf("expected hardened to go to 'versatile:9090', got %q", address)
	}
}

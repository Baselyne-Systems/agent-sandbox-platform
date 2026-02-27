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

func (m *mockRepo) FindHostForPlacement(_ context.Context, memoryMb int64, cpuMillicores int32, diskMb int64) (*models.Host, error) {
	var candidates []*models.Host
	for _, h := range m.hosts {
		if h.Status != models.HostStatusReady {
			continue
		}
		if h.AvailableResources.MemoryMb >= memoryMb &&
			h.AvailableResources.CpuMillicores >= cpuMillicores &&
			h.AvailableResources.DiskMb >= diskMb {
			cp := *h
			candidates = append(candidates, &cp)
		}
	}
	if len(candidates) == 0 {
		return nil, ErrNoCapacity
	}
	// First-fit: sort by available memory ascending, pick smallest.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].AvailableResources.MemoryMb < candidates[j].AvailableResources.MemoryMb
	})
	return candidates[0], nil
}

func (m *mockRepo) DecrementAvailableResources(_ context.Context, hostID string, memoryMb int64, cpuMillicores int32, diskMb int64) error {
	h, ok := m.hosts[hostID]
	if !ok {
		return ErrHostNotFound
	}
	h.AvailableResources.MemoryMb -= memoryMb
	h.AvailableResources.CpuMillicores -= cpuMillicores
	h.AvailableResources.DiskMb -= diskMb
	h.ActiveSandboxes++
	h.LastHeartbeat = time.Now()
	return nil
}

// --- RegisterHost tests ---

func TestRegisterHost_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	host, err := svc.RegisterHost(context.Background(), "host1.example.com:9090", models.HostResources{
		MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400,
	})
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

	if _, err := svc.RegisterHost(ctx, "", models.HostResources{MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 1024}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty address, got: %v", err)
	}
}

func TestRegisterHost_InvalidResources(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 0, CpuMillicores: 1000, DiskMb: 1024}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero memory, got: %v", err)
	}
	if _, err := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 1024, CpuMillicores: 0, DiskMb: 1024}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero cpu, got: %v", err)
	}
	if _, err := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 0}); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero disk, got: %v", err)
	}
}

// --- DeregisterHost tests ---

func TestDeregisterHost_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	host, _ := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 1024})

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
	svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 1024})
	svc.RegisterHost(ctx, "host2:9090", models.HostResources{MemoryMb: 2048, CpuMillicores: 2000, DiskMb: 2048})

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
	host1, _ := svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 1024, CpuMillicores: 1000, DiskMb: 1024})
	svc.RegisterHost(ctx, "host2:9090", models.HostResources{MemoryMb: 2048, CpuMillicores: 2000, DiskMb: 2048})

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
	svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240})

	hostID, address, err := svc.PlaceWorkspace(ctx, 512, 500, 1024)
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
	_, _, err := svc.PlaceWorkspace(context.Background(), 512, 500, 1024)
	if !errors.Is(err, ErrNoCapacity) {
		t.Errorf("expected ErrNoCapacity, got: %v", err)
	}
}

func TestPlaceWorkspace_InsufficientResources(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	svc.RegisterHost(ctx, "host1:9090", models.HostResources{MemoryMb: 256, CpuMillicores: 200, DiskMb: 512})

	_, _, err := svc.PlaceWorkspace(ctx, 1024, 500, 1024)
	if !errors.Is(err, ErrNoCapacity) {
		t.Errorf("expected ErrNoCapacity for insufficient resources, got: %v", err)
	}
}

func TestPlaceWorkspace_SelectsSmallestFit(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	// Register a large host and a small host.
	svc.RegisterHost(ctx, "large:9090", models.HostResources{MemoryMb: 16384, CpuMillicores: 8000, DiskMb: 102400})
	svc.RegisterHost(ctx, "small:9090", models.HostResources{MemoryMb: 2048, CpuMillicores: 2000, DiskMb: 10240})

	_, address, err := svc.PlaceWorkspace(ctx, 512, 500, 1024)
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
	host1, _ := svc.RegisterHost(ctx, "draining:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240})
	svc.RegisterHost(ctx, "ready:9090", models.HostResources{MemoryMb: 4096, CpuMillicores: 4000, DiskMb: 10240})

	// Set host1 to draining (simulate via repo).
	svc.repo.SetHostStatus(ctx, host1.ID, models.HostStatusDraining)

	_, address, err := svc.PlaceWorkspace(ctx, 512, 500, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if address != "ready:9090" {
		t.Errorf("expected to skip draining host, got %q", address)
	}
}

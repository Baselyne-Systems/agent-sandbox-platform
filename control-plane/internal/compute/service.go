package compute

import (
	"context"
	"errors"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

var (
	ErrHostNotFound = errors.New("host not found")
	ErrNoCapacity   = errors.New("no host with sufficient capacity")
	ErrInvalidInput = errors.New("invalid input")
)

// Service implements compute plane business logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterHost(ctx context.Context, address string, resources models.HostResources) (*models.Host, error) {
	if address == "" {
		return nil, ErrInvalidInput
	}
	if resources.MemoryMb <= 0 || resources.CpuMillicores <= 0 || resources.DiskMb <= 0 {
		return nil, ErrInvalidInput
	}

	host := &models.Host{
		Address:            address,
		Status:             models.HostStatusReady,
		TotalResources:     resources,
		AvailableResources: resources,
		ActiveSandboxes:    0,
		LastHeartbeat:      time.Now().UTC(),
	}

	if err := s.repo.CreateHost(ctx, host); err != nil {
		return nil, err
	}
	return host, nil
}

func (s *Service) DeregisterHost(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}
	return s.repo.SetHostStatus(ctx, id, models.HostStatusOffline)
}

func (s *Service) ListHosts(ctx context.Context, status models.HostStatus) ([]models.Host, error) {
	return s.repo.ListHosts(ctx, status)
}

func (s *Service) PlaceWorkspace(ctx context.Context, memoryMb int64, cpuMillicores int32, diskMb int64) (string, string, error) {
	if memoryMb <= 0 || cpuMillicores <= 0 || diskMb <= 0 {
		return "", "", ErrInvalidInput
	}

	host, err := s.repo.FindHostForPlacement(ctx, memoryMb, cpuMillicores, diskMb)
	if err != nil {
		return "", "", err
	}

	if err := s.repo.DecrementAvailableResources(ctx, host.ID, memoryMb, cpuMillicores, diskMb); err != nil {
		return "", "", err
	}

	return host.ID, host.Address, nil
}

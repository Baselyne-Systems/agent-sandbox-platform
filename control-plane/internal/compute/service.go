package compute

import (
	"context"
	"errors"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
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

func (s *Service) RegisterHost(ctx context.Context, address string, resources models.HostResources, supportedTiers []string) (*models.Host, error) {
	if address == "" {
		return nil, ErrInvalidInput
	}
	if resources.MemoryMb <= 0 || resources.CpuMillicores <= 0 || resources.DiskMb <= 0 {
		return nil, ErrInvalidInput
	}

	if len(supportedTiers) == 0 {
		supportedTiers = []string{"standard"}
	}

	host := &models.Host{
		Address:            address,
		Status:             models.HostStatusReady,
		TotalResources:     resources,
		AvailableResources: resources,
		ActiveSandboxes:    0,
		LastHeartbeat:      time.Now().UTC(),
		SupportedTiers:     supportedTiers,
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

func (s *Service) PlaceWorkspace(ctx context.Context, memoryMb int64, cpuMillicores int32, diskMb int64, isolationTier string) (string, string, error) {
	if memoryMb <= 0 || cpuMillicores <= 0 || diskMb <= 0 {
		return "", "", ErrInvalidInput
	}

	// Try claiming a pre-warmed slot first.
	if isolationTier != "" {
		slot, err := s.repo.ClaimWarmSlot(ctx, isolationTier)
		if err == nil && slot != nil {
			host, err := s.repo.GetHost(ctx, slot.HostID)
			if err == nil && host != nil {
				return host.ID, host.Address, nil
			}
		}
	}

	// Fall back to cold placement.
	host, err := s.repo.PlaceAndDecrement(ctx, memoryMb, cpuMillicores, diskMb, isolationTier)
	if err != nil {
		return "", "", err
	}

	return host.ID, host.Address, nil
}

func (s *Service) ConfigureWarmPool(ctx context.Context, cfg *models.WarmPoolConfig) (*models.WarmPoolConfig, error) {
	if cfg.IsolationTier == "" {
		return nil, ErrInvalidInput
	}
	if cfg.TargetCount < 0 {
		return nil, ErrInvalidInput
	}
	if cfg.MemoryMb <= 0 || cfg.CpuMillicores <= 0 || cfg.DiskMb <= 0 {
		return nil, ErrInvalidInput
	}
	if err := s.repo.UpsertWarmPoolConfig(ctx, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Service) GetCapacity(ctx context.Context) ([]models.TierCapacity, int32, int32, error) {
	return s.repo.GetCapacity(ctx)
}

func (s *Service) Heartbeat(ctx context.Context, hostID string, resources models.HostResources, activeSandboxes int32, supportedTiers []string) (*models.Host, error) {
	if hostID == "" {
		return nil, ErrInvalidInput
	}
	if resources.MemoryMb < 0 || resources.CpuMillicores < 0 || resources.DiskMb < 0 {
		return nil, ErrInvalidInput
	}
	if activeSandboxes < 0 {
		return nil, ErrInvalidInput
	}
	return s.repo.UpdateHeartbeat(ctx, hostID, resources, activeSandboxes, supportedTiers)
}

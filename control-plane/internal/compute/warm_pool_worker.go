package compute

import (
	"context"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
	"go.uber.org/zap"
)

// WarmPoolWorkerConfig holds configuration for the background warm pool replenishment worker.
type WarmPoolWorkerConfig struct {
	Interval time.Duration
	Logger   *zap.Logger
}

// WarmPoolWorker periodically ensures warm pools stay at their configured target size.
type WarmPoolWorker struct {
	repo     Repository
	interval time.Duration
	logger   *zap.Logger
}

// NewWarmPoolWorker creates a worker that replenishes warm pool slots at the configured interval.
func NewWarmPoolWorker(repo Repository, cfg WarmPoolWorkerConfig) *WarmPoolWorker {
	interval := cfg.Interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	logger := cfg.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &WarmPoolWorker{
		repo:     repo,
		interval: interval,
		logger:   logger,
	}
}

// Run starts the warm pool replenishment loop. It blocks until ctx is cancelled.
func (w *WarmPoolWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("warm pool worker started", zap.Duration("interval", w.interval))

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("warm pool worker stopped")
			return
		case <-ticker.C:
			w.replenish(ctx)
		}
	}
}

func (w *WarmPoolWorker) replenish(ctx context.Context) {
	// Clean up slots on offline hosts.
	cleaned, err := w.repo.CleanExpiredSlots(ctx)
	if err != nil {
		w.logger.Error("warm pool cleanup failed", zap.Error(err))
	} else if cleaned > 0 {
		w.logger.Info("cleaned expired warm pool slots", zap.Int64("count", cleaned))
	}

	// List all warm pool configs.
	configs, err := w.repo.ListWarmPoolConfigs(ctx)
	if err != nil {
		w.logger.Error("failed to list warm pool configs", zap.Error(err))
		return
	}

	for _, cfg := range configs {
		if cfg.TargetCount <= 0 {
			continue
		}

		ready, err := w.repo.CountReadySlots(ctx, cfg.IsolationTier)
		if err != nil {
			w.logger.Error("failed to count ready slots",
				zap.String("tier", cfg.IsolationTier), zap.Error(err))
			continue
		}

		deficit := cfg.TargetCount - ready
		if deficit <= 0 {
			continue
		}

		w.logger.Info("replenishing warm pool",
			zap.String("tier", cfg.IsolationTier),
			zap.Int32("ready", ready),
			zap.Int32("target", cfg.TargetCount),
			zap.Int32("deficit", deficit),
		)

		for i := int32(0); i < deficit; i++ {
			// Allocate a host slot using the existing placement logic.
			host, err := w.repo.PlaceAndDecrement(ctx, cfg.MemoryMb, cfg.CpuMillicores, cfg.DiskMb, cfg.IsolationTier)
			if err != nil {
				w.logger.Warn("warm pool replenishment: no capacity",
					zap.String("tier", cfg.IsolationTier), zap.Error(err))
				break // no more capacity for this tier
			}

			slot := &models.WarmPoolSlot{
				HostID:        host.ID,
				IsolationTier: cfg.IsolationTier,
				MemoryMb:      cfg.MemoryMb,
				CpuMillicores: cfg.CpuMillicores,
				DiskMb:        cfg.DiskMb,
				Status:        "ready",
			}
			if err := w.repo.CreateWarmSlot(ctx, slot); err != nil {
				w.logger.Error("failed to create warm slot",
					zap.String("tier", cfg.IsolationTier), zap.Error(err))
				break
			}
		}
	}
}

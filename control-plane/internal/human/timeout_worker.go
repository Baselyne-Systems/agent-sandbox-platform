package human

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// TimeoutWorkerConfig holds configuration for the background timeout enforcement worker.
type TimeoutWorkerConfig struct {
	Interval time.Duration
	Logger   *zap.Logger
}

// TimeoutWorker periodically scans for expired pending requests and marks them expired.
type TimeoutWorker struct {
	repo     Repository
	interval time.Duration
	logger   *zap.Logger
}

// NewTimeoutWorker creates a worker that sweeps expired requests at the configured interval.
func NewTimeoutWorker(repo Repository, cfg TimeoutWorkerConfig) *TimeoutWorker {
	interval := cfg.Interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	logger := cfg.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &TimeoutWorker{
		repo:     repo,
		interval: interval,
		logger:   logger,
	}
}

// Run starts the timeout enforcement loop. It blocks until ctx is cancelled.
func (w *TimeoutWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("timeout enforcement worker started", zap.Duration("interval", w.interval))

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("timeout enforcement worker stopped")
			return
		case <-ticker.C:
			w.sweep(ctx)
		}
	}
}

// sweep finds and expires all pending requests past their deadline.
func (w *TimeoutWorker) sweep(ctx context.Context) {
	count, err := w.repo.ExpirePendingRequests(ctx)
	if err != nil {
		w.logger.Error("timeout sweep failed", zap.Error(err))
		return
	}
	if count > 0 {
		w.logger.Info("expired pending requests", zap.Int("count", count))
	}
}

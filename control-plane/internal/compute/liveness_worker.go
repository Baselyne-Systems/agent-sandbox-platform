package compute

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// LivenessWorkerConfig holds configuration for the background host liveness detection worker.
type LivenessWorkerConfig struct {
	// Interval between sweeps (how often to check for stale hosts).
	Interval time.Duration
	// HeartbeatTimeout is how long since the last heartbeat before a host is marked offline.
	HeartbeatTimeout time.Duration
	Logger           *zap.Logger
}

// LivenessWorker periodically scans for hosts that have missed heartbeats and marks them offline.
type LivenessWorker struct {
	repo             Repository
	interval         time.Duration
	heartbeatTimeout time.Duration
	logger           *zap.Logger
}

// NewLivenessWorker creates a worker that sweeps stale hosts at the configured interval.
func NewLivenessWorker(repo Repository, cfg LivenessWorkerConfig) *LivenessWorker {
	interval := cfg.Interval
	if interval <= 0 {
		interval = 60 * time.Second
	}
	heartbeatTimeout := cfg.HeartbeatTimeout
	if heartbeatTimeout <= 0 {
		heartbeatTimeout = 180 * time.Second
	}
	logger := cfg.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &LivenessWorker{
		repo:             repo,
		interval:         interval,
		heartbeatTimeout: heartbeatTimeout,
		logger:           logger,
	}
}

// Run starts the liveness detection loop. It blocks until ctx is cancelled.
func (w *LivenessWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("host liveness worker started",
		zap.Duration("interval", w.interval),
		zap.Duration("heartbeat_timeout", w.heartbeatTimeout),
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("host liveness worker stopped")
			return
		case <-ticker.C:
			w.sweep(ctx)
		}
	}
}

// sweep finds hosts with stale heartbeats and marks them offline.
func (w *LivenessWorker) sweep(ctx context.Context) {
	count, err := w.repo.MarkStaleHostsOffline(ctx, w.heartbeatTimeout)
	if err != nil {
		w.logger.Error("liveness sweep failed", zap.Error(err))
		return
	}
	if count > 0 {
		w.logger.Warn("marked stale hosts offline", zap.Int64("count", count))
	}
}

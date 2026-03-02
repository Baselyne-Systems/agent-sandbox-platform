// Binary observability consolidates Activity Store, Economics, and Human
// Interaction services into a single gRPC server. The Human→Activity
// dependency is wired in-process via an adapter.
package main

import (
	"context"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/activity"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/adapters"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/boot"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/economics"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/human"
	activitypb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/activity/v1"
	economicspb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/economics/v1"
	humanpb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/human/v1"
)

func main() {
	f := boot.New(boot.Options{ServiceName: "bulkhead-observability"})

	// -----------------------------------------------------------------------
	// 1. Activity Store
	// -----------------------------------------------------------------------
	activityRepo := activity.NewPostgresRepository(f.DB)
	activitySvc := activity.NewService(activityRepo)
	activityHandler := activity.NewHandler(activitySvc)
	activitypb.RegisterActivityServiceServer(f.Server, activityHandler)

	// -----------------------------------------------------------------------
	// 2. Economics
	// -----------------------------------------------------------------------
	economicsRepo := economics.NewPostgresRepository(f.DB)
	economicsSvc := economics.NewService(economicsRepo)
	economicsHandler := economics.NewHandler(economicsSvc)
	economicspb.RegisterEconomicsServiceServer(f.Server, economicsHandler)

	// -----------------------------------------------------------------------
	// 3. Human Interaction
	// -----------------------------------------------------------------------
	humanRepo := human.NewPostgresRepository(f.DB)
	humanSvc := human.NewService(humanRepo)

	// In-process: wire Activity logger directly (no gRPC hop).
	humanSvc.SetActivityLogger(&adapters.ActivityLoggerAdapter{Svc: activitySvc})

	humanHandler := human.NewHandler(humanSvc)
	humanpb.RegisterHumanInteractionServiceServer(f.Server, humanHandler)

	// Start human timeout enforcement worker.
	workerCtx, workerCancel := context.WithCancel(context.Background())
	worker := human.NewTimeoutWorker(humanRepo, human.TimeoutWorkerConfig{
		Interval: 30 * time.Second,
		Logger:   f.Logger,
	})
	go worker.Run(workerCtx)

	// -----------------------------------------------------------------------
	// Serve
	// -----------------------------------------------------------------------
	f.Logger.Info("registered services: Activity, Economics, Human")
	f.Serve()
	workerCancel()
}

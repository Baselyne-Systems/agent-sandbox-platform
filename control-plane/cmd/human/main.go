package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/config"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/database"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/human"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/identity"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/middleware"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/telemetry"
	activitypb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/activity/v1"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/human/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadFromEnv()

	logger, err := newLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	tp, err := telemetry.InitTracer("bulkhead-human", cfg.OTelEndpoint)
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer telemetry.Shutdown(context.Background(), tp)

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	repo := human.NewPostgresRepository(db)
	svc := human.NewService(repo)

	// Optional: wire Activity Store logger if ACTIVITY_ENDPOINT is set.
	if ep := os.Getenv("ACTIVITY_ENDPOINT"); ep != "" {
		logger.Info("connecting to Activity Store", zap.String("endpoint", ep))
		activityConn, err := grpc.NewClient(ep, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("failed to connect to Activity Store: %v", err)
		}
		defer activityConn.Close()
		svc.SetActivityLogger(&activityAdapter{client: activitypb.NewActivityServiceClient(activityConn)})
	}

	handler := human.NewHandler(svc)

	authCfg := middleware.AuthConfig{
		Credentials:   &middleware.PostgresCredentialLookup{DB: db},
		TokenHashFunc: identity.HashToken,
	}
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			middleware.UnaryLoggingInterceptor(logger),
			middleware.UnaryAuthInterceptor(authCfg),
			middleware.UnarySpanEnrichInterceptor(),
		),
	)
	pb.RegisterHumanInteractionServiceServer(srv, handler)
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(srv)

	// Start timeout enforcement worker.
	workerCtx, workerCancel := context.WithCancel(context.Background())
	worker := human.NewTimeoutWorker(repo, human.TimeoutWorkerConfig{
		Interval: 30 * time.Second,
		Logger:   logger,
	})
	go worker.Run(workerCtx)

	logger.Info("Human Interaction Service starting", zap.String("port", cfg.GRPCPort))

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	workerCancel()
	logger.Info("shutting down Human Interaction Service")
	srv.GracefulStop()
}

// activityAdapter adapts ActivityServiceClient to the human.ActivityLogger interface.
type activityAdapter struct {
	client activitypb.ActivityServiceClient
}

func (a *activityAdapter) RecordAction(ctx context.Context, record *models.ActionRecord) error {
	_, err := a.client.RecordAction(ctx, &activitypb.RecordActionRequest{
		Record: &activitypb.ActionRecord{
			TenantId:    record.TenantID,
			WorkspaceId: record.WorkspaceID,
			AgentId:     record.AgentID,
			TaskId:      record.TaskID,
			ToolName:    record.ToolName,
			Outcome:     activitypb.ActionOutcome(activitypb.ActionOutcome_value["ACTION_OUTCOME_"+strings.ToUpper(string(record.Outcome))]),
		},
	})
	return err
}

func newLogger(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	cfg.Level = lvl
	return cfg.Build()
}

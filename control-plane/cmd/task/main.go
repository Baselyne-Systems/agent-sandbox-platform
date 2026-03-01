package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/config"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/database"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/identity"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/middleware"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/telemetry"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/task"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/task/v1"
	workspacepb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/workspace/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/durationpb"
)

func main() {
	cfg := config.LoadFromEnv()

	logger, err := newLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	tp, err := telemetry.InitTracer("bulkhead-task", cfg.OTelEndpoint)
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer telemetry.Shutdown(context.Background(), tp)

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := task.NewPostgresRepository(db)

	// Optional: wire workspace provisioner if WORKSPACE_ENDPOINT is set.
	var provisioner task.WorkspaceProvisioner
	if ep := os.Getenv("WORKSPACE_ENDPOINT"); ep != "" {
		logger.Info("connecting to workspace service", zap.String("endpoint", ep))
		wsConn, err := grpc.NewClient(ep, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("failed to connect to workspace service: %v", err)
		}
		defer wsConn.Close()
		provisioner = &workspaceAdapter{client: workspacepb.NewWorkspaceServiceClient(wsConn)}
	}

	svc := task.NewService(task.ServiceConfig{
		Repo:        repo,
		Provisioner: provisioner,
	})
	handler := task.NewHandler(svc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

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
	pb.RegisterTaskServiceServer(srv, handler)
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(srv)

	logger.Info("Task Service starting", zap.String("port", cfg.GRPCPort))

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down Task Service")
	srv.GracefulStop()
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

// --- Adapter wrapping Workspace Service gRPC client ---

// workspaceAdapter implements task.WorkspaceProvisioner via the Workspace Service.
type workspaceAdapter struct {
	client workspacepb.WorkspaceServiceClient
}

func (a *workspaceAdapter) ProvisionWorkspace(ctx context.Context, t *models.Task) (string, error) {
	spec := &workspacepb.WorkspaceSpec{
		MemoryMb:          t.WorkspaceConfig.MemoryMb,
		CpuMillicores:     t.WorkspaceConfig.CpuMillicores,
		DiskMb:            t.WorkspaceConfig.DiskMb,
		AllowedTools:      t.WorkspaceConfig.AllowedTools,
		GuardrailPolicyId: t.GuardrailPolicyID,
		EnvVars:           t.WorkspaceConfig.EnvVars,
	}
	if t.WorkspaceConfig.MaxDurationSecs > 0 {
		spec.MaxDuration = durationpb.New(
			time.Duration(t.WorkspaceConfig.MaxDurationSecs) * time.Second,
		)
	}

	resp, err := a.client.CreateWorkspace(ctx, &workspacepb.CreateWorkspaceRequest{
		AgentId: t.AgentID,
		TaskId:  t.ID,
		Spec:    spec,
	})
	if err != nil {
		return "", fmt.Errorf("create workspace: %w", err)
	}
	return resp.GetWorkspace().GetWorkspaceId(), nil
}

func (a *workspaceAdapter) TerminateWorkspace(ctx context.Context, workspaceID, reason string) error {
	_, err := a.client.TerminateWorkspace(ctx, &workspacepb.TerminateWorkspaceRequest{
		WorkspaceId: workspaceID,
		Reason:      reason,
	})
	return err
}

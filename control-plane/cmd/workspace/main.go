package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/config"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/database"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/identity"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/middleware"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/workspace"
	computepb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/compute/v1"
	guardrailspb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/guardrails/v1"
	hostagentpb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/host_agent/v1"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/workspace/v1"
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

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	repo := workspace.NewPostgresRepository(db)

	// Optional: wire compute placer if COMPUTE_ENDPOINT is set.
	var compute workspace.ComputePlacer
	if ep := os.Getenv("COMPUTE_ENDPOINT"); ep != "" {
		logger.Info("connecting to compute plane", zap.String("endpoint", ep))
		conn, err := grpc.NewClient(ep, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("failed to connect to compute plane: %v", err)
		}
		defer conn.Close()
		compute = &computeAdapter{client: computepb.NewComputePlaneServiceClient(conn)}
	}

	// Optional: wire guardrails policy compiler if GUARDRAILS_ENDPOINT is set.
	var guardrails workspace.PolicyCompiler
	if ep := os.Getenv("GUARDRAILS_ENDPOINT"); ep != "" {
		logger.Info("connecting to guardrails service", zap.String("endpoint", ep))
		conn, err := grpc.NewClient(ep, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("failed to connect to guardrails service: %v", err)
		}
		defer conn.Close()
		guardrails = &guardrailsAdapter{client: guardrailspb.NewGuardrailsServiceClient(conn)}
	}

	// Host Agent dialer: creates a gRPC connection to a Host Agent on demand.
	var dialHostAgent workspace.HostAgentDialer
	dialHostAgent = func(_ context.Context, address string) (hostagentpb.HostAgentServiceClient, error) {
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("dial host agent at %s: %w", address, err)
		}
		return hostagentpb.NewHostAgentServiceClient(conn), nil
	}

	svc := workspace.NewService(workspace.ServiceConfig{
		Repo:        repo,
		Compute:     compute,
		Guardrails:  guardrails,
		DialHostAgent: dialHostAgent,
		Logger:      logger,
	})
	handler := workspace.NewHandler(svc)

	authCfg := middleware.AuthConfig{
		Credentials:   &middleware.PostgresCredentialLookup{DB: db},
		TokenHashFunc: identity.HashToken,
	}
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryLoggingInterceptor(logger),
			middleware.UnaryAuthInterceptor(authCfg),
		),
	)
	pb.RegisterWorkspaceServiceServer(srv, handler)
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(srv)

	logger.Info("Workspace Service starting", zap.String("port", cfg.GRPCPort))

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down Workspace Service")
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

// --- Adapter clients wrapping generated gRPC clients ---

// computeAdapter adapts ComputePlaneServiceClient to the ComputePlacer interface.
type computeAdapter struct {
	client computepb.ComputePlaneServiceClient
}

func (a *computeAdapter) PlaceWorkspace(ctx context.Context, memoryMb int64, cpuMillicores int32, diskMb int64) (string, string, error) {
	resp, err := a.client.PlaceWorkspace(ctx, &computepb.PlaceWorkspaceRequest{
		MemoryMb:      memoryMb,
		CpuMillicores: cpuMillicores,
		DiskMb:        diskMb,
	})
	if err != nil {
		return "", "", err
	}
	return resp.GetHostId(), resp.GetRuntimeEndpoint(), nil
}

// guardrailsAdapter adapts GuardrailsServiceClient to the PolicyCompiler interface.
type guardrailsAdapter struct {
	client guardrailspb.GuardrailsServiceClient
}

func (a *guardrailsAdapter) CompilePolicy(ctx context.Context, ruleIDs []string) ([]byte, int, error) {
	resp, err := a.client.CompilePolicy(ctx, &guardrailspb.CompilePolicyRequest{
		RuleIds: ruleIDs,
	})
	if err != nil {
		return nil, 0, err
	}
	return resp.GetCompiledPolicy(), int(resp.GetRuleCount()), nil
}

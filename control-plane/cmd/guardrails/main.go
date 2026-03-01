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
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/guardrails"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/identity"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/middleware"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/telemetry"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/guardrails/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	tp, err := telemetry.InitTracer("bulkhead-guardrails", cfg.OTelEndpoint)
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

	repo := guardrails.NewPostgresRepository(db)
	svc := guardrails.NewService(repo)
	handler := guardrails.NewHandler(svc)

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
	pb.RegisterGuardrailsServiceServer(srv, handler)
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(srv)

	logger.Info("Guardrails Service starting", zap.String("port", cfg.GRPCPort))

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down Guardrails Service")
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

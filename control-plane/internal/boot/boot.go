// Package boot extracts the common startup sequence shared by all control-plane
// binaries: config, logger, tracer, database, gRPC server (with auth + tracing
// interceptors), health check, and signal-based graceful shutdown.
package boot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/config"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/database"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/identity"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/middleware"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Foundation holds the shared infrastructure for a control-plane binary.
// Callers create services/handlers, register them on Server, then call Serve.
type Foundation struct {
	Cfg    *config.Config
	Logger *zap.Logger
	DB     *sql.DB
	Server *grpc.Server
	Health *health.Server

	tp  *sdktrace.TracerProvider
	lis net.Listener
}

// Options configures Foundation creation.
type Options struct {
	// ServiceName is used for the OTel tracer (e.g. "bulkhead-control-plane").
	ServiceName string

	// NoDB skips database connection (for stateless services like governance).
	NoDB bool

	// NoAuth skips auth interceptor (empty AuthConfig).
	NoAuth bool
}

// New creates a Foundation: config, logger, tracer, optional DB, gRPC server
// with standard interceptors, and health + reflection services pre-registered.
func New(opts Options) *Foundation {
	cfg := config.LoadFromEnv()

	logger, err := newLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	tp, err := telemetry.InitTracer(opts.ServiceName, cfg.OTelEndpoint)
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}

	var db *sql.DB
	if !opts.NoDB {
		db, err = database.NewConnection(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Build auth config. Most binaries use DB-backed credential lookup.
	var authCfg middleware.AuthConfig
	if !opts.NoAuth && db != nil {
		authCfg = middleware.AuthConfig{
			Credentials:   &middleware.PostgresCredentialLookup{DB: db},
			TokenHashFunc: identity.HashToken,
		}
	}

	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			middleware.UnaryLoggingInterceptor(logger),
			middleware.UnaryAuthInterceptor(authCfg),
			middleware.UnarySpanEnrichInterceptor(),
		),
	)

	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	reflection.Register(srv)

	return &Foundation{
		Cfg:    cfg,
		Logger: logger,
		DB:     db,
		Server: srv,
		Health: healthSrv,
		tp:     tp,
		lis:    lis,
	}
}

// Serve starts the gRPC server and blocks until SIGINT/SIGTERM. It then
// performs graceful shutdown of the server, database, and tracer.
func (f *Foundation) Serve() {
	f.Logger.Info("starting", zap.String("port", f.Cfg.GRPCPort))

	go func() {
		if err := f.Server.Serve(f.lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	f.Logger.Info("shutting down")
	f.Server.GracefulStop()

	if f.DB != nil {
		f.DB.Close()
	}
	telemetry.Shutdown(context.Background(), f.tp)
	f.Logger.Sync()
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

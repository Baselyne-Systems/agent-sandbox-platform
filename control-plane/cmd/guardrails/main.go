package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/config"
	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/database"
	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/guardrails"
	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/middleware"
	pb "github.com/baselyne/agent-sandbox-platform/control-plane/pkg/gen/guardrails/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	repo := guardrails.NewPostgresRepository(db)
	svc := guardrails.NewService(repo)
	handler := guardrails.NewHandler(svc)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor(logger)),
	)
	pb.RegisterGuardrailsServiceServer(srv, handler)
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

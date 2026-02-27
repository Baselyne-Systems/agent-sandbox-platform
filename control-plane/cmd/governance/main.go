package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/config"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/governance"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/middleware"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/governance/v1"
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	svc := governance.NewService()
	handler := governance.NewHandler(svc)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor(logger)),
	)
	pb.RegisterDataGovernanceServiceServer(srv, handler)
	reflection.Register(srv)

	logger.Info("Data Governance Service starting", zap.String("port", cfg.GRPCPort))

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down Data Governance Service")
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

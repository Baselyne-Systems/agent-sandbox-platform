package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/config"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/database"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/middleware"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/task"
	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/task/v1"
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

	repo := task.NewPostgresRepository(db)
	svc := task.NewService(repo)
	handler := task.NewHandler(svc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryLoggingInterceptor(logger)),
	)
	pb.RegisterTaskServiceServer(srv, handler)
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

package middleware

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
		}

		if err != nil {
			st, _ := status.FromError(err)
			fields = append(fields, zap.String("code", st.Code().String()))
			fields = append(fields, zap.Error(err))
			logger.Error("grpc call failed", fields...)
		} else {
			logger.Info("grpc call", fields...)
		}

		return resp, err
	}
}

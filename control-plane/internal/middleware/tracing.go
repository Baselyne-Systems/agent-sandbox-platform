package middleware

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// UnarySpanEnrichInterceptor adds tenant_id and agent_id attributes to the
// current trace span. It must run after UnaryAuthInterceptor so that the
// auth context values are available.
func UnarySpanEnrichInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			if tenantID, ok := TenantIDFromContext(ctx); ok {
				span.SetAttributes(attribute.String("bulkhead.tenant_id", tenantID))
			}
			if agentID, ok := AgentIDFromContext(ctx); ok {
				span.SetAttributes(attribute.String("bulkhead.agent_id", agentID))
			}
		}
		return handler(ctx, req)
	}
}

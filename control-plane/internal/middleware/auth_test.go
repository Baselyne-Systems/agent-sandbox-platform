package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// --- mock credential lookup ---

type mockCredentialLookup struct {
	agentID   string
	tenantID  string
	scopes    []string
	expiresAt time.Time
	err       error
}

func (m *mockCredentialLookup) LookupByTokenHash(_ context.Context, _ string) (string, string, []string, time.Time, error) {
	return m.agentID, m.tenantID, m.scopes, m.expiresAt, m.err
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// dummyHandler is a grpc.UnaryHandler that records whether it was called.
func dummyHandler(ctx context.Context, req any) (any, error) {
	return "ok", nil
}

func makeInfo(method string) *grpc.UnaryServerInfo {
	return &grpc.UnaryServerInfo{FullMethod: method}
}

func ctxWithAuth(token string) context.Context {
	md := metadata.Pairs("authorization", token)
	return metadata.NewIncomingContext(context.Background(), md)
}

// --- tests ---

func TestAuthInterceptor_NilCredentials_Passthrough(t *testing.T) {
	interceptor := UnaryAuthInterceptor(AuthConfig{Credentials: nil})
	resp, err := interceptor(context.Background(), nil, makeInfo("/test.Service/Method"), dummyHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}
}

func TestAuthInterceptor_SkipsHealthCheck(t *testing.T) {
	lookup := &mockCredentialLookup{err: ErrCredentialNotFound}
	interceptor := UnaryAuthInterceptor(AuthConfig{
		Credentials:   lookup,
		TokenHashFunc: hashToken,
	})
	// No auth metadata, but health check should still pass.
	resp, err := interceptor(context.Background(), nil, makeInfo("/grpc.health.v1.Health/Check"), dummyHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}
}

func TestAuthInterceptor_SkipsReflection(t *testing.T) {
	lookup := &mockCredentialLookup{err: ErrCredentialNotFound}
	interceptor := UnaryAuthInterceptor(AuthConfig{
		Credentials:   lookup,
		TokenHashFunc: hashToken,
	})
	resp, err := interceptor(context.Background(), nil, makeInfo("/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"), dummyHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}
}

func TestAuthInterceptor_MissingMetadata(t *testing.T) {
	lookup := &mockCredentialLookup{}
	interceptor := UnaryAuthInterceptor(AuthConfig{
		Credentials:   lookup,
		TokenHashFunc: hashToken,
	})
	_, err := interceptor(context.Background(), nil, makeInfo("/test.Service/Method"), dummyHandler)
	assertGRPCCode(t, err, codes.Unauthenticated)
}

func TestAuthInterceptor_MissingAuthHeader(t *testing.T) {
	lookup := &mockCredentialLookup{}
	interceptor := UnaryAuthInterceptor(AuthConfig{
		Credentials:   lookup,
		TokenHashFunc: hashToken,
	})
	md := metadata.Pairs("other-key", "value")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, err := interceptor(ctx, nil, makeInfo("/test.Service/Method"), dummyHandler)
	assertGRPCCode(t, err, codes.Unauthenticated)
}

func TestAuthInterceptor_InvalidFormat(t *testing.T) {
	lookup := &mockCredentialLookup{}
	interceptor := UnaryAuthInterceptor(AuthConfig{
		Credentials:   lookup,
		TokenHashFunc: hashToken,
	})
	_, err := interceptor(ctxWithAuth("Basic abc123"), nil, makeInfo("/test.Service/Method"), dummyHandler)
	assertGRPCCode(t, err, codes.Unauthenticated)
}

func TestAuthInterceptor_EmptyBearerToken(t *testing.T) {
	lookup := &mockCredentialLookup{}
	interceptor := UnaryAuthInterceptor(AuthConfig{
		Credentials:   lookup,
		TokenHashFunc: hashToken,
	})
	_, err := interceptor(ctxWithAuth("Bearer "), nil, makeInfo("/test.Service/Method"), dummyHandler)
	assertGRPCCode(t, err, codes.Unauthenticated)
}

func TestAuthInterceptor_TokenNotFound(t *testing.T) {
	lookup := &mockCredentialLookup{err: ErrCredentialNotFound}
	interceptor := UnaryAuthInterceptor(AuthConfig{
		Credentials:   lookup,
		TokenHashFunc: hashToken,
	})
	_, err := interceptor(ctxWithAuth("Bearer sometoken"), nil, makeInfo("/test.Service/Method"), dummyHandler)
	assertGRPCCode(t, err, codes.Unauthenticated)
}

func TestAuthInterceptor_TokenExpired(t *testing.T) {
	lookup := &mockCredentialLookup{
		agentID:   "agent-1",
		scopes:    []string{"read"},
		expiresAt: time.Now().Add(-1 * time.Hour),
	}
	interceptor := UnaryAuthInterceptor(AuthConfig{
		Credentials:   lookup,
		TokenHashFunc: hashToken,
	})
	_, err := interceptor(ctxWithAuth("Bearer sometoken"), nil, makeInfo("/test.Service/Method"), dummyHandler)
	assertGRPCCode(t, err, codes.Unauthenticated)
}

func TestAuthInterceptor_LookupInternalError(t *testing.T) {
	lookup := &mockCredentialLookup{err: errors.New("db connection failed")}
	interceptor := UnaryAuthInterceptor(AuthConfig{
		Credentials:   lookup,
		TokenHashFunc: hashToken,
	})
	_, err := interceptor(ctxWithAuth("Bearer sometoken"), nil, makeInfo("/test.Service/Method"), dummyHandler)
	assertGRPCCode(t, err, codes.Internal)
}

func TestAuthInterceptor_ValidToken(t *testing.T) {
	lookup := &mockCredentialLookup{
		agentID:   "agent-123",
		tenantID:  "tenant-abc",
		scopes:    []string{"read", "write"},
		expiresAt: time.Now().Add(1 * time.Hour),
	}
	interceptor := UnaryAuthInterceptor(AuthConfig{
		Credentials:   lookup,
		TokenHashFunc: hashToken,
	})

	var capturedCtx context.Context
	handler := func(ctx context.Context, req any) (any, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	resp, err := interceptor(ctxWithAuth("Bearer validtoken"), nil, makeInfo("/test.Service/Method"), handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}

	// Verify context enrichment.
	agentID, ok := AgentIDFromContext(capturedCtx)
	if !ok || agentID != "agent-123" {
		t.Errorf("expected agent_id=agent-123, got %q (ok=%v)", agentID, ok)
	}
	tenantID, ok := TenantIDFromContext(capturedCtx)
	if !ok || tenantID != "tenant-abc" {
		t.Errorf("expected tenant_id=tenant-abc, got %q (ok=%v)", tenantID, ok)
	}
	scopes, ok := ScopesFromContext(capturedCtx)
	if !ok || len(scopes) != 2 || scopes[0] != "read" || scopes[1] != "write" {
		t.Errorf("expected scopes=[read write], got %v (ok=%v)", scopes, ok)
	}
}

func assertGRPCCode(t *testing.T, err error, code codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %v, got nil", code)
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != code {
		t.Errorf("expected code %v, got %v: %s", code, st.Code(), st.Message())
	}
}

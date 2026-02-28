package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// contextKey is an unexported type used for context keys in this package.
type contextKey string

const (
	agentIDKey  contextKey = "agent_id"
	tenantIDKey contextKey = "tenant_id"
	scopesKey   contextKey = "scopes"
)

// AgentIDFromContext extracts the authenticated agent ID from the context.
func AgentIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(agentIDKey).(string)
	return v, ok
}

// TenantIDFromContext extracts the authenticated tenant ID from the context.
func TenantIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(tenantIDKey).(string)
	return v, ok
}

// ContextWithTenantID returns a new context with the given tenant ID set.
// This is useful for tests and internal service-to-service calls.
func ContextWithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// ScopesFromContext extracts the authenticated scopes from the context.
func ScopesFromContext(ctx context.Context) ([]string, bool) {
	v, ok := ctx.Value(scopesKey).([]string)
	return v, ok
}

// ErrCredentialNotFound is returned when no matching credential exists.
var ErrCredentialNotFound = errors.New("credential not found")

// CredentialLookup abstracts credential verification for the auth interceptor.
type CredentialLookup interface {
	LookupByTokenHash(ctx context.Context, tokenHash string) (agentID string, tenantID string, scopes []string, expiresAt time.Time, err error)
}

// PostgresCredentialLookup implements CredentialLookup using direct DB queries.
type PostgresCredentialLookup struct {
	DB *sql.DB
}

func (p *PostgresCredentialLookup) LookupByTokenHash(ctx context.Context, tokenHash string) (string, string, []string, time.Time, error) {
	var agentID, tenantID string
	var scopesJSON []byte
	var expiresAt time.Time
	err := p.DB.QueryRowContext(ctx,
		`SELECT agent_id, tenant_id, scopes, expires_at
		 FROM scoped_credentials
		 WHERE token_hash = $1 AND revoked = false`,
		tokenHash,
	).Scan(&agentID, &tenantID, &scopesJSON, &expiresAt)

	if err == sql.ErrNoRows {
		return "", "", nil, time.Time{}, ErrCredentialNotFound
	}
	if err != nil {
		return "", "", nil, time.Time{}, err
	}

	var scopes []string
	if err := json.Unmarshal(scopesJSON, &scopes); err != nil {
		return "", "", nil, time.Time{}, fmt.Errorf("unmarshal scopes: %w", err)
	}
	return agentID, tenantID, scopes, expiresAt, nil
}

// AuthConfig holds configuration for the auth interceptor.
type AuthConfig struct {
	Credentials   CredentialLookup
	TokenHashFunc func(string) string
}

// UnaryAuthInterceptor returns a gRPC unary interceptor that validates Bearer tokens.
// If cfg.Credentials is nil, all requests are passed through without auth.
func UnaryAuthInterceptor(cfg AuthConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// If no credential lookup is configured, skip auth.
		if cfg.Credentials == nil {
			return handler(ctx, req)
		}

		// Skip auth for gRPC reflection and health checks.
		if strings.HasPrefix(info.FullMethod, "/grpc.reflection.") ||
			strings.HasPrefix(info.FullMethod, "/grpc.health.") {
			return handler(ctx, req)
		}

		// Extract authorization metadata.
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authValues := md.Get("authorization")
		if len(authValues) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		token := authValues[0]
		if !strings.HasPrefix(token, "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}
		rawToken := strings.TrimPrefix(token, "Bearer ")
		if rawToken == "" {
			return nil, status.Error(codes.Unauthenticated, "empty bearer token")
		}

		// Hash the token and look up credentials.
		tokenHash := cfg.TokenHashFunc(rawToken)
		agentID, tenantID, scopes, expiresAt, err := cfg.Credentials.LookupByTokenHash(ctx, tokenHash)
		if errors.Is(err, ErrCredentialNotFound) {
			return nil, status.Error(codes.Unauthenticated, "invalid or revoked token")
		}
		if err != nil {
			return nil, status.Error(codes.Internal, "credential lookup failed")
		}

		// Check expiration.
		if time.Now().After(expiresAt) {
			return nil, status.Error(codes.Unauthenticated, "token expired")
		}

		// Enrich context with agent and tenant identity.
		ctx = context.WithValue(ctx, agentIDKey, agentID)
		ctx = context.WithValue(ctx, tenantIDKey, tenantID)
		ctx = context.WithValue(ctx, scopesKey, scopes)

		return handler(ctx, req)
	}
}

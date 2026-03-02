// Package adapters provides in-process wrappers that let services call each
// other directly when they live in the same binary, eliminating gRPC hops.
//
// Each adapter implements an interface defined in the consuming service package
// (e.g. task.WorkspaceProvisioner, human.ActivityLogger) by delegating to the
// producing service's concrete *Service type.
package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/activity"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/identity"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/workspace"
)

// ---------------------------------------------------------------------------
// WorkspaceProvisionerAdapter: workspace.Service → task.WorkspaceProvisioner
// ---------------------------------------------------------------------------

// WorkspaceProvisionerAdapter wraps a workspace.Service to satisfy the
// task.WorkspaceProvisioner interface for in-process calls.
type WorkspaceProvisionerAdapter struct {
	WsSvc *workspace.Service
	DB    *sql.DB
}

func (a *WorkspaceProvisionerAdapter) ProvisionWorkspace(ctx context.Context, t *models.Task) (string, error) {
	spec := &models.WorkspaceSpec{
		MemoryMb:          t.WorkspaceConfig.MemoryMb,
		CpuMillicores:     t.WorkspaceConfig.CpuMillicores,
		DiskMb:            t.WorkspaceConfig.DiskMb,
		MaxDurationSecs:   t.WorkspaceConfig.MaxDurationSecs,
		AllowedTools:      t.WorkspaceConfig.AllowedTools,
		GuardrailPolicyID: t.GuardrailPolicyID,
		EnvVars:           t.WorkspaceConfig.EnvVars,
		ContainerImage:    t.WorkspaceConfig.ContainerImage,
		EgressAllowlist:   t.WorkspaceConfig.EgressAllowlist,
		IsolationTier:     models.IsolationTier(t.WorkspaceConfig.IsolationTier),
	}

	ws, err := a.WsSvc.CreateWorkspace(ctx, t.TenantID, t.AgentID, t.ID, spec)
	if err != nil {
		return "", err
	}
	if ws.Status == models.WorkspaceStatusFailed {
		return "", fmt.Errorf("workspace provisioning failed")
	}
	return ws.ID, nil
}

func (a *WorkspaceProvisionerAdapter) TerminateWorkspace(ctx context.Context, workspaceID string, reason string) error {
	var tenantID string
	err := a.DB.QueryRowContext(ctx, "SELECT tenant_id FROM workspaces WHERE id = $1", workspaceID).Scan(&tenantID)
	if err != nil {
		return fmt.Errorf("lookup tenant for workspace %s: %w", workspaceID, err)
	}
	return a.WsSvc.TerminateWorkspace(ctx, tenantID, workspaceID, reason)
}

// ---------------------------------------------------------------------------
// IdentityCredentialAdapter: identity.Service → workspace.CredentialMinter
// ---------------------------------------------------------------------------

// IdentityCredentialAdapter wraps an identity.Service to satisfy the
// workspace.CredentialMinter interface for in-process calls.
type IdentityCredentialAdapter struct {
	Svc *identity.Service
	DB  *sql.DB
}

func (a *IdentityCredentialAdapter) MintCredential(ctx context.Context, agentID string, scopes []string, ttlSeconds int64) (string, error) {
	var tenantID string
	err := a.DB.QueryRowContext(ctx, "SELECT tenant_id FROM agents WHERE id = $1", agentID).Scan(&tenantID)
	if err != nil {
		return "", fmt.Errorf("lookup tenant for agent %s: %w", agentID, err)
	}
	_, token, err := a.Svc.MintCredential(ctx, tenantID, agentID, scopes, ttlSeconds)
	return token, err
}

// ---------------------------------------------------------------------------
// IdentityQuerierAdapter: direct DB query → workspace.IdentityQuerier
// ---------------------------------------------------------------------------

// IdentityQuerierAdapter satisfies workspace.IdentityQuerier using a direct
// DB query against the agents table (shared DB, no service call needed).
type IdentityQuerierAdapter struct {
	DB *sql.DB
}

func (a *IdentityQuerierAdapter) GetAgent(ctx context.Context, agentID string) (*models.Agent, error) {
	var agent models.Agent
	err := a.DB.QueryRowContext(ctx,
		"SELECT id, tenant_id, name, trust_level FROM agents WHERE id = $1", agentID,
	).Scan(&agent.ID, &agent.TenantID, &agent.Name, &agent.TrustLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &agent, nil
}

// ---------------------------------------------------------------------------
// ActivityLoggerAdapter: activity.Service → human.ActivityLogger
// ---------------------------------------------------------------------------

// ActivityLoggerAdapter wraps an activity.Service to satisfy the
// human.ActivityLogger interface for in-process calls.
type ActivityLoggerAdapter struct {
	Svc *activity.Service
}

func (a *ActivityLoggerAdapter) RecordAction(ctx context.Context, record *models.ActionRecord) error {
	_, err := a.Svc.RecordAction(ctx, record)
	return err
}

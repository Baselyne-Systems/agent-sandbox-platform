package e2e

import (
	"context"
	"testing"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/identity"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

func TestAgentFullLifecycle(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	// Register agent.
	agent, err := identitySvc.RegisterAgent(ctx, tenant, "lifecycle-agent", "test", "owner-1", map[string]string{"env": "test"}, "ci runner", models.AgentTrustLevelNew, []string{"shell", "web"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if agent.ID == "" {
		t.Fatal("expected non-empty agent ID")
	}
	if agent.Status != models.AgentStatusActive {
		t.Fatalf("expected active, got %s", agent.Status)
	}

	// Get agent.
	got, err := identitySvc.GetAgent(ctx, tenant, agent.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "lifecycle-agent" {
		t.Fatalf("expected name 'lifecycle-agent', got %q", got.Name)
	}

	// Mint credential.
	cred, token, err := identitySvc.MintCredential(ctx, tenant, agent.ID, []string{"shell"}, 3600)
	if err != nil {
		t.Fatalf("mint: %v", err)
	}
	if token == "" || cred.ID == "" {
		t.Fatal("expected non-empty token and credential ID")
	}

	// Update trust level.
	updated, err := identitySvc.UpdateTrustLevel(ctx, tenant, agent.ID, models.AgentTrustLevelEstablished, "good behavior")
	if err != nil {
		t.Fatalf("update trust: %v", err)
	}
	if updated.TrustLevel != models.AgentTrustLevelEstablished {
		t.Fatalf("expected established, got %s", updated.TrustLevel)
	}

	// Suspend agent.
	suspended, err := identitySvc.SuspendAgent(ctx, tenant, agent.ID)
	if err != nil {
		t.Fatalf("suspend: %v", err)
	}
	if suspended.Status != models.AgentStatusSuspended {
		t.Fatalf("expected suspended, got %s", suspended.Status)
	}

	// Mint credential should fail while suspended.
	_, _, err = identitySvc.MintCredential(ctx, tenant, agent.ID, []string{"shell"}, 3600)
	if err != identity.ErrAgentInactive {
		t.Fatalf("expected ErrAgentInactive, got %v", err)
	}

	// Reactivate.
	reactivated, err := identitySvc.ReactivateAgent(ctx, tenant, agent.ID)
	if err != nil {
		t.Fatalf("reactivate: %v", err)
	}
	if reactivated.Status != models.AgentStatusActive {
		t.Fatalf("expected active, got %s", reactivated.Status)
	}

	// Can mint again after reactivation.
	_, token2, err := identitySvc.MintCredential(ctx, tenant, agent.ID, []string{"web"}, 3600)
	if err != nil {
		t.Fatalf("mint after reactivate: %v", err)
	}
	if token2 == "" {
		t.Fatal("expected non-empty token after reactivation")
	}

	// Deactivate.
	if err := identitySvc.DeactivateAgent(ctx, tenant, agent.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	deactivated, err := identitySvc.GetAgent(ctx, tenant, agent.ID)
	if err != nil {
		t.Fatalf("get deactivated: %v", err)
	}
	if deactivated.Status != models.AgentStatusInactive {
		t.Fatalf("expected inactive, got %s", deactivated.Status)
	}
}

func TestAgentRegistrationValidation(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	// Empty name.
	_, err := identitySvc.RegisterAgent(ctx, tenant, "", "desc", "owner-1", nil, "", models.AgentTrustLevelNew, nil)
	if err != identity.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput for empty name, got %v", err)
	}

	// Empty owner.
	_, err = identitySvc.RegisterAgent(ctx, tenant, "test", "desc", "", nil, "", models.AgentTrustLevelNew, nil)
	if err != identity.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput for empty owner, got %v", err)
	}
}

func TestAgentTrustLevelProgression(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "trust-agent")
	if agent.TrustLevel != models.AgentTrustLevelNew {
		t.Fatalf("expected new, got %s", agent.TrustLevel)
	}

	// new -> established.
	updated, err := identitySvc.UpdateTrustLevel(ctx, tenant, agent.ID, models.AgentTrustLevelEstablished, "test")
	if err != nil {
		t.Fatalf("new -> established: %v", err)
	}
	if updated.TrustLevel != models.AgentTrustLevelEstablished {
		t.Fatalf("expected established, got %s", updated.TrustLevel)
	}

	// established -> trusted.
	updated, err = identitySvc.UpdateTrustLevel(ctx, tenant, agent.ID, models.AgentTrustLevelTrusted, "proven")
	if err != nil {
		t.Fatalf("established -> trusted: %v", err)
	}
	if updated.TrustLevel != models.AgentTrustLevelTrusted {
		t.Fatalf("expected trusted, got %s", updated.TrustLevel)
	}

	// Invalid trust level.
	_, err = identitySvc.UpdateTrustLevel(ctx, tenant, agent.ID, "invalid_level", "bad")
	if err != identity.ErrInvalidTrustLevel {
		t.Fatalf("expected ErrInvalidTrustLevel, got %v", err)
	}
}

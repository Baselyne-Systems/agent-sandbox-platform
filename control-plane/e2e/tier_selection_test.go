package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

func TestIsolationTierAutoSelection(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	registerHost(t, ctx, "tier-host-a.local:9090", 16384, 16000, 40960, []string{"standard", "hardened", "isolated"})

	tests := []struct {
		trustLevel         models.AgentTrustLevel
		dataClassification string
		expectedTier       models.IsolationTier
	}{
		{models.AgentTrustLevelTrusted, "public", models.IsolationTierStandard},
		{models.AgentTrustLevelTrusted, "internal", models.IsolationTierStandard},
		{models.AgentTrustLevelTrusted, "confidential", models.IsolationTierStandard},
		{models.AgentTrustLevelTrusted, "restricted", models.IsolationTierIsolated},

		{models.AgentTrustLevelEstablished, "public", models.IsolationTierStandard},
		{models.AgentTrustLevelEstablished, "internal", models.IsolationTierStandard},
		{models.AgentTrustLevelEstablished, "confidential", models.IsolationTierHardened},
		{models.AgentTrustLevelEstablished, "restricted", models.IsolationTierIsolated},

		{models.AgentTrustLevelNew, "public", models.IsolationTierHardened},
		{models.AgentTrustLevelNew, "internal", models.IsolationTierHardened},
		{models.AgentTrustLevelNew, "confidential", models.IsolationTierIsolated},
		{models.AgentTrustLevelNew, "restricted", models.IsolationTierIsolated},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s_%s", tt.trustLevel, tt.dataClassification)
		t.Run(name, func(t *testing.T) {
			clean(t)
			registerHost(t, ctx, "tier-host.local:9090", 16384, 16000, 40960, []string{"standard", "hardened", "isolated"})

			agent := registerAgentWithTrust(t, ctx, tenant, "tier-"+name, tt.trustLevel)

			spec := &models.WorkspaceSpec{
				DataClassification: tt.dataClassification,
			}

			ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", spec)
			if err != nil {
				t.Fatalf("create workspace: %v", err)
			}
			if ws.IsolationTier != tt.expectedTier {
				t.Fatalf("expected tier %s, got %s", tt.expectedTier, ws.IsolationTier)
			}
		})
	}
}

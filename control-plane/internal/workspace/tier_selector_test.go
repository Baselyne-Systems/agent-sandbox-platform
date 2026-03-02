package workspace

import (
	"testing"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

func TestTierSelector_DecisionMatrix(t *testing.T) {
	ts := NewTierSelector()

	tests := []struct {
		name           string
		trustLevel     models.AgentTrustLevel
		classification string
		want           models.IsolationTier
	}{
		// Trusted row
		{"trusted/public", models.AgentTrustLevelTrusted, "public", models.IsolationTierStandard},
		{"trusted/internal", models.AgentTrustLevelTrusted, "internal", models.IsolationTierStandard},
		{"trusted/confidential", models.AgentTrustLevelTrusted, "confidential", models.IsolationTierStandard},
		{"trusted/restricted", models.AgentTrustLevelTrusted, "restricted", models.IsolationTierIsolated},

		// Established row
		{"established/public", models.AgentTrustLevelEstablished, "public", models.IsolationTierStandard},
		{"established/internal", models.AgentTrustLevelEstablished, "internal", models.IsolationTierStandard},
		{"established/confidential", models.AgentTrustLevelEstablished, "confidential", models.IsolationTierHardened},
		{"established/restricted", models.AgentTrustLevelEstablished, "restricted", models.IsolationTierIsolated},

		// New row
		{"new/public", models.AgentTrustLevelNew, "public", models.IsolationTierHardened},
		{"new/internal", models.AgentTrustLevelNew, "internal", models.IsolationTierHardened},
		{"new/confidential", models.AgentTrustLevelNew, "confidential", models.IsolationTierIsolated},
		{"new/restricted", models.AgentTrustLevelNew, "restricted", models.IsolationTierIsolated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ts.SelectTier(tt.trustLevel, tt.classification)
			if got != tt.want {
				t.Errorf("SelectTier(%q, %q) = %q, want %q", tt.trustLevel, tt.classification, got, tt.want)
			}
		})
	}
}

func TestTierSelector_EmptyClassificationDefaultsToInternal(t *testing.T) {
	ts := NewTierSelector()

	// Empty classification should behave like "internal"
	got := ts.SelectTier(models.AgentTrustLevelNew, "")
	want := ts.SelectTier(models.AgentTrustLevelNew, "internal")
	if got != want {
		t.Errorf("empty classification: got %q, want %q (same as internal)", got, want)
	}
}

func TestTierSelector_UnknownClassificationDefaultsToInternal(t *testing.T) {
	ts := NewTierSelector()

	got := ts.SelectTier(models.AgentTrustLevelEstablished, "unknown-level")
	want := ts.SelectTier(models.AgentTrustLevelEstablished, "internal")
	if got != want {
		t.Errorf("unknown classification: got %q, want %q (same as internal)", got, want)
	}
}

func TestTierSelector_UnknownTrustLevelTreatedAsNew(t *testing.T) {
	ts := NewTierSelector()

	// Unknown trust level should use the default (new) path
	got := ts.SelectTier("unknown-trust", "public")
	want := ts.SelectTier(models.AgentTrustLevelNew, "public")
	if got != want {
		t.Errorf("unknown trust level: got %q, want %q (same as new)", got, want)
	}
}

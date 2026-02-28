package workspace

import "github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"

// TierSelector auto-selects an isolation tier based on agent trust level
// and data classification. The decision matrix escalates isolation as
// trust decreases and data sensitivity increases.
type TierSelector struct{}

// NewTierSelector returns a new TierSelector with the default decision matrix.
func NewTierSelector() *TierSelector {
	return &TierSelector{}
}

// SelectTier returns the appropriate isolation tier for the given trust level
// and data classification. If classification is empty, "internal" is assumed.
//
// Decision matrix:
//
//	                 public    internal    confidential   restricted
//	trusted        standard   standard      standard      isolated
//	established    standard   standard      hardened      isolated
//	new            hardened   hardened      isolated      isolated
func (ts *TierSelector) SelectTier(trustLevel models.AgentTrustLevel, dataClassification string) models.IsolationTier {
	if dataClassification == "" {
		dataClassification = "internal"
	}

	classLevel := classificationLevel(dataClassification)

	switch trustLevel {
	case models.AgentTrustLevelTrusted:
		if classLevel >= 4 {
			return models.IsolationTierIsolated
		}
		return models.IsolationTierStandard

	case models.AgentTrustLevelEstablished:
		if classLevel >= 4 {
			return models.IsolationTierIsolated
		}
		if classLevel >= 3 {
			return models.IsolationTierHardened
		}
		return models.IsolationTierStandard

	default: // new or unknown
		if classLevel >= 3 {
			return models.IsolationTierIsolated
		}
		return models.IsolationTierHardened
	}
}

// classificationLevel maps a data classification string to a numeric level
// for comparison. Higher = more sensitive.
func classificationLevel(classification string) int {
	switch classification {
	case "public":
		return 1
	case "internal":
		return 2
	case "confidential":
		return 3
	case "restricted":
		return 4
	default:
		return 2 // default to internal
	}
}

package models

import (
	"encoding/json"
	"time"
)

// AgentStatus represents the lifecycle state of an agent.
type AgentStatus string

const (
	AgentStatusActive    AgentStatus = "active"
	AgentStatusInactive  AgentStatus = "inactive"
	AgentStatusSuspended AgentStatus = "suspended"
)

// Agent represents an autonomous agent registered in the platform.
type Agent struct {
	ID          string
	Name        string
	Description string
	OwnerID     string
	Status      AgentStatus
	Labels      map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ScopedCredential is a time-limited credential with explicit permission scopes.
type ScopedCredential struct {
	ID        string
	AgentID   string
	Scopes    []string
	TokenHash string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

// ActionOutcome represents the result of a guardrail evaluation.
type ActionOutcome string

const (
	ActionOutcomeAllowed   ActionOutcome = "allowed"
	ActionOutcomeDenied    ActionOutcome = "denied"
	ActionOutcomeEscalated ActionOutcome = "escalated"
	ActionOutcomeError     ActionOutcome = "error"
)

// ActionRecord is an immutable record of a single agent action.
type ActionRecord struct {
	ID                  string
	WorkspaceID         string
	AgentID             string
	TaskID              string
	ToolName            string
	Parameters          json.RawMessage
	Result              json.RawMessage
	Outcome             ActionOutcome
	GuardrailRuleID     string
	DenialReason        string
	EvaluationLatencyUs *int64
	ExecutionLatencyUs  *int64
	RecordedAt          time.Time
}

type GuardrailRule struct {
	ID          string
	Name        string
	Description string
	Type        string
	Condition   string
	Action      string
	Priority    int
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

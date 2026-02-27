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

// RuleType classifies guardrail rules.
type RuleType string

const (
	RuleTypeToolFilter      RuleType = "tool_filter"
	RuleTypeParameterCheck  RuleType = "parameter_check"
	RuleTypeRateLimit       RuleType = "rate_limit"
	RuleTypeBudgetLimit     RuleType = "budget_limit"
)

// RuleAction determines what happens when a guardrail rule matches.
type RuleAction string

const (
	RuleActionAllow    RuleAction = "allow"
	RuleActionDeny     RuleAction = "deny"
	RuleActionEscalate RuleAction = "escalate"
	RuleActionLog      RuleAction = "log"
)

type GuardrailRule struct {
	ID          string
	Name        string
	Description string
	Type        RuleType
	Condition   string
	Action      RuleAction
	Priority    int
	Enabled     bool
	Labels      map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UsageRecord represents a single unit of resource consumption.
type UsageRecord struct {
	ID           string
	AgentID      string
	WorkspaceID  string
	ResourceType string
	Unit         string
	Quantity     float64
	Cost         float64
	RecordedAt   time.Time
}

// Budget represents a spending limit for an agent.
type Budget struct {
	ID          string
	AgentID     string
	Currency    string
	Limit       float64
	Used        float64
	PeriodStart time.Time
	PeriodEnd   time.Time
}

// WorkspaceStatus represents the lifecycle state of a workspace.
type WorkspaceStatus string

const (
	WorkspaceStatusPending     WorkspaceStatus = "pending"
	WorkspaceStatusCreating    WorkspaceStatus = "creating"
	WorkspaceStatusRunning     WorkspaceStatus = "running"
	WorkspaceStatusPaused      WorkspaceStatus = "paused"
	WorkspaceStatusTerminating WorkspaceStatus = "terminating"
	WorkspaceStatusTerminated  WorkspaceStatus = "terminated"
	WorkspaceStatusFailed      WorkspaceStatus = "failed"
)

// WorkspaceSpec defines the desired resource configuration for a workspace.
type WorkspaceSpec struct {
	MemoryMb          int64
	CpuMillicores     int32
	DiskMb            int64
	MaxDurationSecs   int64
	AllowedTools      []string
	GuardrailPolicyID string
	EnvVars           map[string]string
}

// Workspace represents an isolated execution environment for an agent.
type Workspace struct {
	ID        string
	AgentID   string
	TaskID    string
	Status    WorkspaceStatus
	Spec      WorkspaceSpec
	HostID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt *time.Time
}

// HostStatus represents the lifecycle state of a compute host.
type HostStatus string

const (
	HostStatusReady    HostStatus = "ready"
	HostStatusDraining HostStatus = "draining"
	HostStatusOffline  HostStatus = "offline"
)

// HostResources describes the resource capacity of a host.
type HostResources struct {
	MemoryMb      int64
	CpuMillicores int32
	DiskMb        int64
}

// Host represents a compute host in the fleet.
type Host struct {
	ID                 string
	Address            string
	Status             HostStatus
	TotalResources     HostResources
	AvailableResources HostResources
	ActiveSandboxes    int32
	LastHeartbeat      time.Time
}

// HumanRequestStatus represents the lifecycle state of a human interaction request.
type HumanRequestStatus string

const (
	HumanRequestStatusPending   HumanRequestStatus = "pending"
	HumanRequestStatusResponded HumanRequestStatus = "responded"
	HumanRequestStatusExpired   HumanRequestStatus = "expired"
	HumanRequestStatusCancelled HumanRequestStatus = "cancelled"
)

// HumanRequest represents an agent's request for human input.
type HumanRequest struct {
	ID          string
	WorkspaceID string
	AgentID     string
	Question    string
	Options     []string
	Context     string
	Status      HumanRequestStatus
	Response    string
	ResponderID string
	CreatedAt   time.Time
	RespondedAt *time.Time
	ExpiresAt   *time.Time
}

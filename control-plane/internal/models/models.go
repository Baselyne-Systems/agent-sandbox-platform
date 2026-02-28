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

// AgentTrustLevel represents the level of trust assigned to an agent.
type AgentTrustLevel string

const (
	AgentTrustLevelNew         AgentTrustLevel = "new"
	AgentTrustLevelEstablished AgentTrustLevel = "established"
	AgentTrustLevelTrusted     AgentTrustLevel = "trusted"
)

// Agent represents an autonomous agent registered in the platform.
type Agent struct {
	ID           string
	Name         string
	Description  string
	OwnerID      string
	Status       AgentStatus
	Labels       map[string]string
	Purpose      string           // what this agent does — input to guardrail scoping
	TrustLevel   AgentTrustLevel  // affects guardrail strictness and isolation tier
	Capabilities []string         // tools/systems this agent can use
	CreatedAt    time.Time
	UpdatedAt    time.Time
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

// RuleScope restricts which agents/tools/trust-levels/data-classifications a rule applies to.
// Empty fields mean "match all".
type RuleScope struct {
	AgentIDs            []string `json:"agent_ids,omitempty"`
	ToolNames           []string `json:"tool_names,omitempty"`
	TrustLevels         []string `json:"trust_levels,omitempty"`
	DataClassifications []string `json:"data_classifications,omitempty"`
}

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
	Scope       RuleScope
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
	ID               string
	AgentID          string
	Currency         string
	Limit            float64
	Used             float64
	PeriodStart      time.Time
	PeriodEnd        time.Time
	OnExceeded       string  // "halt", "request_increase", "warn", or "" (default: halt)
	WarningThreshold float64 // 0.0–1.0: fraction of limit that triggers a warning
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
	ContainerImage    string
	EgressAllowlist   []string
}

// Workspace represents an isolated execution environment for an agent.
type Workspace struct {
	ID        string
	AgentID   string
	TaskID    string
	Status    WorkspaceStatus
	Spec      WorkspaceSpec
	HostID      string
	HostAddress string
	SandboxID   string
	SnapshotID  string
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt *time.Time
}

// WorkspaceSnapshot records metadata for a workspace snapshot.
type WorkspaceSnapshot struct {
	ID          string
	WorkspaceID string
	AgentID     string
	TaskID      string
	SizeBytes   int64
	CreatedAt   time.Time
}

// DeliveryChannelConfig stores a user's notification channel preference.
type DeliveryChannelConfig struct {
	ID          string
	UserID      string
	ChannelType string // slack, email, teams
	Endpoint    string
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TimeoutPolicy configures default timeout behavior for human interaction requests.
type TimeoutPolicy struct {
	ID                string
	Scope             string // global, agent, workspace
	ScopeID           string
	TimeoutSecs       int64
	Action            string // escalate, continue, halt
	EscalationTargets []string
	CreatedAt         time.Time
	UpdatedAt         time.Time
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

// HumanRequestType classifies the kind of human interaction.
type HumanRequestType string

const (
	HumanRequestTypeApproval   HumanRequestType = "approval"
	HumanRequestTypeQuestion   HumanRequestType = "question"
	HumanRequestTypeEscalation HumanRequestType = "escalation"
)

// HumanRequestUrgency indicates how quickly a response is needed.
type HumanRequestUrgency string

const (
	HumanRequestUrgencyLow      HumanRequestUrgency = "low"
	HumanRequestUrgencyNormal   HumanRequestUrgency = "normal"
	HumanRequestUrgencyHigh     HumanRequestUrgency = "high"
	HumanRequestUrgencyCritical HumanRequestUrgency = "critical"
)

// HumanRequest represents an agent's request for human input.
type HumanRequest struct {
	ID          string
	WorkspaceID string
	AgentID     string
	TaskID      string
	Question    string
	Options     []string
	Context     string
	Type        HumanRequestType
	Urgency     HumanRequestUrgency
	Status      HumanRequestStatus
	Response    string
	ResponderID string
	CreatedAt   time.Time
	RespondedAt *time.Time
	ExpiresAt   *time.Time
}

// AlertConditionType classifies what condition triggers an alert.
type AlertConditionType string

const (
	AlertConditionDenialRate     AlertConditionType = "denial_rate"
	AlertConditionErrorRate      AlertConditionType = "error_rate"
	AlertConditionActionVelocity AlertConditionType = "action_velocity"
	AlertConditionBudgetBreach   AlertConditionType = "budget_breach"
	AlertConditionStuckAgent     AlertConditionType = "stuck_agent"
)

// AlertConfig stores an alert rule definition.
type AlertConfig struct {
	ID            string
	Name          string
	ConditionType AlertConditionType
	Threshold     float64
	AgentID       string // optional scope — empty means all agents
	Enabled       bool
	WebhookURL    string
	CreatedAt     time.Time
}

// Alert represents a triggered alert instance.
type Alert struct {
	ID            string
	ConfigID      string
	AgentID       string
	ConditionType AlertConditionType
	Message       string
	TriggeredAt   time.Time
	Resolved      bool
}

// BehaviorReport summarizes an agent's behavior over a time window
// (produced by the considered evaluation tier).
type BehaviorReport struct {
	AgentID        string
	WindowStart    time.Time
	WindowEnd      time.Time
	ActionCount    int64
	DenialRate     float64
	ErrorRate      float64
	Flags          []string
	Recommendation string
}

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	TaskStatusPending        TaskStatus = "pending"
	TaskStatusRunning        TaskStatus = "running"
	TaskStatusWaitingOnHuman TaskStatus = "waiting_on_human"
	TaskStatusCompleted      TaskStatus = "completed"
	TaskStatusFailed         TaskStatus = "failed"
	TaskStatusCancelled      TaskStatus = "cancelled"
)

// TaskWorkspaceConfig defines workspace requirements for a task.
type TaskWorkspaceConfig struct {
	IsolationTier string            `json:"isolation_tier,omitempty"`
	Persistent    bool              `json:"persistent,omitempty"`
	MemoryMb      int64             `json:"memory_mb,omitempty"`
	CpuMillicores int32            `json:"cpu_millicores,omitempty"`
	DiskMb        int64             `json:"disk_mb,omitempty"`
	MaxDurationSecs int64          `json:"max_duration_secs,omitempty"`
	AllowedTools    []string          `json:"allowed_tools,omitempty"`
	EnvVars         map[string]string `json:"env_vars,omitempty"`
	ContainerImage  string            `json:"container_image,omitempty"`
	EgressAllowlist []string          `json:"egress_allowlist,omitempty"`
}

// TaskHumanInteractionConfig defines how a task interacts with humans.
type TaskHumanInteractionConfig struct {
	EscalationTargets []string `json:"escalation_targets,omitempty"`
	ApprovalTargets   []string `json:"approval_targets,omitempty"`
	TimeoutSecs       int64    `json:"timeout_secs,omitempty"`
	TimeoutAction     string   `json:"timeout_action,omitempty"` // escalate, continue, halt
}

// TaskBudgetConfig defines cost constraints for a task.
type TaskBudgetConfig struct {
	MaxCost          float64 `json:"max_cost,omitempty"`
	WarningThreshold float64 `json:"warning_threshold,omitempty"`
	OnExceeded       string  `json:"on_exceeded,omitempty"` // halt, request_increase
	Currency         string  `json:"currency,omitempty"`
}

// Task represents a goal an agent pursues autonomously.
type Task struct {
	ID                          string
	AgentID                     string
	Goal                        string
	Status                      TaskStatus
	WorkspaceID                 string
	GuardrailPolicyID           string
	WorkspaceConfig             TaskWorkspaceConfig
	HumanInteractionConfig      TaskHumanInteractionConfig
	BudgetConfig                TaskBudgetConfig
	MaxDurationWithoutCheckinSecs int64
	Input                       map[string]string
	Labels                      map[string]string
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	CompletedAt                 *time.Time
}

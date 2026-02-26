package models

import "time"

type Agent struct {
	ID          string
	Name        string
	Description string
	OwnerID     string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ActionRecord struct {
	ID          string
	WorkspaceID string
	AgentID     string
	TaskID      string
	ToolName    string
	Outcome     string
	RecordedAt  time.Time
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

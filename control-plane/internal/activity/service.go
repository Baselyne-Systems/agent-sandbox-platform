package activity

import (
	"context"
	"errors"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

var (
	ErrRecordNotFound = errors.New("action record not found")
	ErrInvalidInput   = errors.New("invalid input")
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

// Service implements activity store business logic.
type Service struct {
	repo      Repository
	alertRepo AlertRepository
	broker    *Broker
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, broker: NewBroker()}
}

// SetAlertRepository attaches an alert repository, enabling alert features.
func (s *Service) SetAlertRepository(alertRepo AlertRepository) {
	s.alertRepo = alertRepo
}

// Broker returns the service's action record broker for streaming.
func (s *Service) Broker() *Broker {
	return s.broker
}

func (s *Service) RecordAction(ctx context.Context, record *models.ActionRecord) (string, error) {
	if record.WorkspaceID == "" {
		return "", ErrInvalidInput
	}
	if record.AgentID == "" {
		return "", ErrInvalidInput
	}
	if record.ToolName == "" {
		return "", ErrInvalidInput
	}
	if record.Outcome == "" {
		return "", ErrInvalidInput
	}
	if err := s.repo.InsertAction(ctx, record); err != nil {
		return "", err
	}

	// Publish to streaming subscribers (fire-and-forget, non-blocking).
	s.broker.Publish(record)

	return record.ID, nil
}

func (s *Service) GetAction(ctx context.Context, tenantID, id string) (*models.ActionRecord, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	record, err := s.repo.GetAction(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrRecordNotFound
	}
	return record, nil
}

func (s *Service) QueryActions(ctx context.Context, tenantID string, filter QueryFilter) ([]models.ActionRecord, string, error) {
	if filter.Limit <= 0 {
		filter.Limit = defaultPageSize
	}
	if filter.Limit > maxPageSize {
		filter.Limit = maxPageSize
	}

	// Fetch one extra to determine if there's a next page
	filter.Limit++
	records, err := s.repo.QueryActions(ctx, tenantID, filter)
	if err != nil {
		return nil, "", err
	}
	filter.Limit-- // restore original

	var nextToken string
	if len(records) > filter.Limit {
		records = records[:filter.Limit]
		nextToken = records[filter.Limit-1].ID
	}

	return records, nextToken, nil
}

// ParseQueryFilter builds a QueryFilter from request parameters.
func ParseQueryFilter(workspaceID, agentID, taskID, toolName string, outcome models.ActionOutcome, startTime, endTime *time.Time, pageSize int, pageToken string) QueryFilter {
	return QueryFilter{
		WorkspaceID: workspaceID,
		AgentID:     agentID,
		TaskID:      taskID,
		ToolName:    toolName,
		Outcome:     outcome,
		StartTime:   startTime,
		EndTime:     endTime,
		AfterID:     pageToken, // page token = last seen ID
		Limit:       pageSize,
	}
}

var (
	ErrAlertNotFound       = errors.New("alert not found")
	ErrAlertConfigNotFound = errors.New("alert config not found")
	ErrAlertsNotEnabled    = errors.New("alert repository not configured")
)

var validConditionTypes = map[models.AlertConditionType]bool{
	models.AlertConditionDenialRate:     true,
	models.AlertConditionErrorRate:      true,
	models.AlertConditionActionVelocity: true,
	models.AlertConditionBudgetBreach:   true,
	models.AlertConditionStuckAgent:     true,
}

func (s *Service) ConfigureAlert(ctx context.Context, name string, conditionType models.AlertConditionType, threshold float64, agentID, webhookURL string) (*models.AlertConfig, error) {
	if s.alertRepo == nil {
		return nil, ErrAlertsNotEnabled
	}
	if name == "" {
		return nil, ErrInvalidInput
	}
	if !validConditionTypes[conditionType] {
		return nil, ErrInvalidInput
	}
	if threshold <= 0 {
		return nil, ErrInvalidInput
	}

	config := &models.AlertConfig{
		Name:          name,
		ConditionType: conditionType,
		Threshold:     threshold,
		AgentID:       agentID,
		Enabled:       true,
		WebhookURL:    webhookURL,
	}
	if err := s.alertRepo.UpsertAlertConfig(ctx, config); err != nil {
		return nil, err
	}
	return config, nil
}

func (s *Service) ListAlerts(ctx context.Context, agentID string, activeOnly bool, pageSize int, pageToken string) ([]models.Alert, string, error) {
	if s.alertRepo == nil {
		return nil, "", ErrAlertsNotEnabled
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	alerts, err := s.alertRepo.ListAlerts(ctx, agentID, activeOnly, pageToken, pageSize+1)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(alerts) > pageSize {
		alerts = alerts[:pageSize]
		nextToken = alerts[pageSize-1].ID
	}
	return alerts, nextToken, nil
}

func (s *Service) GetAlert(ctx context.Context, id string) (*models.Alert, error) {
	if s.alertRepo == nil {
		return nil, ErrAlertsNotEnabled
	}
	if id == "" {
		return nil, ErrInvalidInput
	}
	alert, err := s.alertRepo.GetAlert(ctx, id)
	if err != nil {
		return nil, err
	}
	if alert == nil {
		return nil, ErrAlertNotFound
	}
	return alert, nil
}

func (s *Service) ResolveAlert(ctx context.Context, id string) error {
	if s.alertRepo == nil {
		return ErrAlertsNotEnabled
	}
	if id == "" {
		return ErrInvalidInput
	}
	return s.alertRepo.ResolveAlert(ctx, id)
}

package activity

import (
	"context"
	"errors"
	"time"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
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
	repo   Repository
	broker *Broker
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, broker: NewBroker()}
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

func (s *Service) GetAction(ctx context.Context, id string) (*models.ActionRecord, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	record, err := s.repo.GetAction(ctx, id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrRecordNotFound
	}
	return record, nil
}

func (s *Service) QueryActions(ctx context.Context, filter QueryFilter) ([]models.ActionRecord, string, error) {
	if filter.Limit <= 0 {
		filter.Limit = defaultPageSize
	}
	if filter.Limit > maxPageSize {
		filter.Limit = maxPageSize
	}

	// Fetch one extra to determine if there's a next page
	filter.Limit++
	records, err := s.repo.QueryActions(ctx, filter)
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

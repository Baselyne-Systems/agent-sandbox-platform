package human

import (
	"context"
	"errors"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

var (
	ErrRequestNotFound  = errors.New("request not found")
	ErrRequestNotPending = errors.New("request is not pending")
	ErrInvalidInput     = errors.New("invalid input")
)

const (
	defaultPageSize    = 50
	maxPageSize        = 100
	defaultTimeoutSecs = 3600 // 1 hour
)

// Service implements human interaction business logic on top of a Repository.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateRequest(ctx context.Context, workspaceID, agentID, question string, options []string, requestContext string, timeoutSeconds int64, requestType models.HumanRequestType, urgency models.HumanRequestUrgency, taskID string) (*models.HumanRequest, error) {
	if workspaceID == "" {
		return nil, ErrInvalidInput
	}
	if agentID == "" {
		return nil, ErrInvalidInput
	}
	if question == "" {
		return nil, ErrInvalidInput
	}
	if options == nil {
		options = []string{}
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultTimeoutSecs
	}
	if requestType == "" {
		requestType = models.HumanRequestTypeQuestion
	}
	if urgency == "" {
		urgency = models.HumanRequestUrgencyNormal
	}

	expiresAt := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	req := &models.HumanRequest{
		WorkspaceID: workspaceID,
		AgentID:     agentID,
		Question:    question,
		Options:     options,
		Context:     requestContext,
		Status:      models.HumanRequestStatusPending,
		ExpiresAt:   &expiresAt,
		Type:        requestType,
		Urgency:     urgency,
		TaskID:      taskID,
	}
	if err := s.repo.CreateRequest(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Service) GetRequest(ctx context.Context, id string) (*models.HumanRequest, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	req, err := s.repo.GetRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, ErrRequestNotFound
	}

	// Check expiration in the service layer
	if req.Status == models.HumanRequestStatusPending && req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		req.Status = models.HumanRequestStatusExpired
	}

	return req, nil
}

func (s *Service) RespondToRequest(ctx context.Context, id, response, responderID string) error {
	if id == "" {
		return ErrInvalidInput
	}
	if response == "" {
		return ErrInvalidInput
	}
	if responderID == "" {
		return ErrInvalidInput
	}
	return s.repo.RespondToRequest(ctx, id, response, responderID)
}

func (s *Service) ListRequests(ctx context.Context, workspaceID string, status models.HumanRequestStatus, pageSize int, pageToken string) ([]models.HumanRequest, string, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	afterID, err := decodePageToken(pageToken)
	if err != nil {
		return nil, "", ErrInvalidInput
	}

	requests, err := s.repo.ListRequests(ctx, workspaceID, status, afterID, pageSize+1)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(requests) > pageSize {
		requests = requests[:pageSize]
		nextToken = encodePageToken(requests[pageSize-1].ID)
	}

	return requests, nextToken, nil
}

func encodePageToken(id string) string {
	if id == "" {
		return ""
	}
	return id
}

func decodePageToken(token string) (string, error) {
	return token, nil
}

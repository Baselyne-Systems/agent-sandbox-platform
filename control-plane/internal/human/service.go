package human

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

var (
	ErrRequestNotFound       = errors.New("request not found")
	ErrRequestNotPending     = errors.New("request is not pending")
	ErrInvalidInput          = errors.New("invalid input")
	ErrChannelNotFound       = errors.New("delivery channel not found")
	ErrTimeoutPolicyNotFound = errors.New("timeout policy not found")
)

const (
	defaultPageSize    = 50
	maxPageSize        = 100
	defaultTimeoutSecs = 3600 // 1 hour
)

// ActivityLogger records HIS events to the Activity Store.
type ActivityLogger interface {
	RecordAction(ctx context.Context, record *models.ActionRecord) error
}

// Service implements human interaction business logic on top of a Repository.
type Service struct {
	repo      Repository
	deliverer Deliverer
	activity  ActivityLogger
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// SetActivityLogger attaches an Activity Store logger for audit trail.
func (s *Service) SetActivityLogger(a ActivityLogger) {
	s.activity = a
}

// SetDeliverer attaches a delivery adapter for sending notifications.
func (s *Service) SetDeliverer(d Deliverer) {
	s.deliverer = d
}

func (s *Service) CreateRequest(ctx context.Context, tenantID, workspaceID, agentID, question string, options []string, requestContext string, timeoutSeconds int64, requestType models.HumanRequestType, urgency models.HumanRequestUrgency, taskID string) (*models.HumanRequest, error) {
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
		TenantID:    tenantID,
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

	// Best-effort notification delivery.
	s.notifyChannels(ctx, tenantID, req)

	// Fire-and-forget: log to Activity Store.
	s.logHISEvent(ctx, tenantID, req.WorkspaceID, req.AgentID, req.TaskID, "his.create_request",
		map[string]string{"request_id": req.ID, "type": string(req.Type), "urgency": string(req.Urgency)}, nil)

	return req, nil
}

func (s *Service) GetRequest(ctx context.Context, tenantID, id string) (*models.HumanRequest, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	req, err := s.repo.GetRequest(ctx, tenantID, id)
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

func (s *Service) RespondToRequest(ctx context.Context, tenantID, id, response, responderID string) error {
	if id == "" {
		return ErrInvalidInput
	}
	if response == "" {
		return ErrInvalidInput
	}
	if responderID == "" {
		return ErrInvalidInput
	}

	// Fetch the request first so we have workspace/agent context for the audit log.
	req, err := s.repo.GetRequest(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if req == nil {
		return ErrRequestNotPending
	}

	if err := s.repo.RespondToRequest(ctx, tenantID, id, response, responderID); err != nil {
		return err
	}

	// Fire-and-forget: log to Activity Store.
	s.logHISEvent(ctx, tenantID, req.WorkspaceID, req.AgentID, req.TaskID, "his.respond_to_request",
		map[string]string{"request_id": id, "responder_id": responderID},
		map[string]string{"response": response})

	return nil
}

func (s *Service) ListRequests(ctx context.Context, tenantID, workspaceID string, status models.HumanRequestStatus, pageSize int, pageToken string) ([]models.HumanRequest, string, error) {
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

	requests, err := s.repo.ListRequests(ctx, tenantID, workspaceID, status, afterID, pageSize+1)
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

var validChannelTypes = map[string]bool{"slack": true, "email": true, "teams": true}
var validTimeoutActions = map[string]bool{"escalate": true, "continue": true, "halt": true}
var validScopes = map[string]bool{"global": true, "agent": true, "workspace": true}

func (s *Service) ConfigureDeliveryChannel(ctx context.Context, tenantID, userID, channelType, endpoint string) (*models.DeliveryChannelConfig, error) {
	if userID == "" || channelType == "" || endpoint == "" {
		return nil, ErrInvalidInput
	}
	if !validChannelTypes[channelType] {
		return nil, ErrInvalidInput
	}
	cfg := &models.DeliveryChannelConfig{
		TenantID:    tenantID,
		UserID:      userID,
		ChannelType: channelType,
		Endpoint:    endpoint,
		Enabled:     true,
	}
	if err := s.repo.UpsertDeliveryChannel(ctx, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Service) GetDeliveryChannel(ctx context.Context, tenantID, userID, channelType string) (*models.DeliveryChannelConfig, error) {
	if userID == "" || channelType == "" {
		return nil, ErrInvalidInput
	}
	cfg, err := s.repo.GetDeliveryChannel(ctx, tenantID, userID, channelType)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, ErrChannelNotFound
	}
	return cfg, nil
}

func (s *Service) SetTimeoutPolicy(ctx context.Context, tenantID, scope, scopeID string, timeoutSecs int64, action string, escalationTargets []string) (*models.TimeoutPolicy, error) {
	if scope == "" || action == "" {
		return nil, ErrInvalidInput
	}
	if !validScopes[scope] {
		return nil, ErrInvalidInput
	}
	if !validTimeoutActions[action] {
		return nil, ErrInvalidInput
	}
	if timeoutSecs <= 0 {
		return nil, ErrInvalidInput
	}
	if scope != "global" && scopeID == "" {
		return nil, ErrInvalidInput
	}
	if escalationTargets == nil {
		escalationTargets = []string{}
	}
	policy := &models.TimeoutPolicy{
		TenantID:          tenantID,
		Scope:             scope,
		ScopeID:           scopeID,
		TimeoutSecs:       timeoutSecs,
		Action:            action,
		EscalationTargets: escalationTargets,
	}
	if err := s.repo.UpsertTimeoutPolicy(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *Service) GetTimeoutPolicy(ctx context.Context, tenantID, scope, scopeID string) (*models.TimeoutPolicy, error) {
	if scope == "" {
		return nil, ErrInvalidInput
	}
	policy, err := s.repo.GetTimeoutPolicy(ctx, tenantID, scope, scopeID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, ErrTimeoutPolicyNotFound
	}
	return policy, nil
}

// logHISEvent records a human interaction event to the Activity Store.
// It is best-effort: errors are silently ignored.
func (s *Service) logHISEvent(ctx context.Context, tenantID, workspaceID, agentID, taskID, toolName string, params, result map[string]string) {
	if s.activity == nil {
		return
	}
	paramsJSON, _ := json.Marshal(params)
	var resultJSON json.RawMessage
	if result != nil {
		resultJSON, _ = json.Marshal(result)
	}
	record := &models.ActionRecord{
		TenantID:    tenantID,
		WorkspaceID: workspaceID,
		AgentID:     agentID,
		TaskID:      taskID,
		ToolName:    toolName,
		Parameters:  paramsJSON,
		Result:      resultJSON,
		Outcome:     models.ActionOutcomeAllowed,
	}
	// Best-effort: ignore errors.
	_ = s.activity.RecordAction(ctx, record)
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

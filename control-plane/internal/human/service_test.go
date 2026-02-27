package human

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// mockRepo is a hand-written in-memory Repository for testing.
type mockRepo struct {
	requests        map[string]*models.HumanRequest
	channels        map[string]*models.DeliveryChannelConfig
	timeoutPolicies map[string]*models.TimeoutPolicy
	nextID          int
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		requests:        make(map[string]*models.HumanRequest),
		channels:        make(map[string]*models.DeliveryChannelConfig),
		timeoutPolicies: make(map[string]*models.TimeoutPolicy),
	}
}

func (m *mockRepo) nextUUID() string {
	m.nextID++
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", m.nextID)
}

func (m *mockRepo) CreateRequest(_ context.Context, req *models.HumanRequest) error {
	req.ID = m.nextUUID()
	req.CreatedAt = time.Now()
	cp := *req
	cp.Options = copyOptions(req.Options)
	if req.ExpiresAt != nil {
		t := *req.ExpiresAt
		cp.ExpiresAt = &t
	}
	m.requests[req.ID] = &cp
	return nil
}

func (m *mockRepo) GetRequest(_ context.Context, id string) (*models.HumanRequest, error) {
	r, ok := m.requests[id]
	if !ok {
		return nil, nil
	}
	cp := *r
	cp.Options = copyOptions(r.Options)
	if r.ExpiresAt != nil {
		t := *r.ExpiresAt
		cp.ExpiresAt = &t
	}
	if r.RespondedAt != nil {
		t := *r.RespondedAt
		cp.RespondedAt = &t
	}
	return &cp, nil
}

func (m *mockRepo) RespondToRequest(_ context.Context, id, response, responderID string) error {
	r, ok := m.requests[id]
	if !ok {
		return ErrRequestNotPending
	}
	if r.Status != models.HumanRequestStatusPending {
		return ErrRequestNotPending
	}
	r.Status = models.HumanRequestStatusResponded
	r.Response = response
	r.ResponderID = responderID
	now := time.Now()
	r.RespondedAt = &now
	return nil
}

func (m *mockRepo) ListRequests(_ context.Context, workspaceID string, status models.HumanRequestStatus, afterID string, limit int) ([]models.HumanRequest, error) {
	var result []models.HumanRequest
	for _, r := range m.requests {
		if workspaceID != "" && r.WorkspaceID != workspaceID {
			continue
		}
		if status != "" && r.Status != status {
			continue
		}
		if afterID != "" && r.ID <= afterID {
			continue
		}
		cp := *r
		cp.Options = copyOptions(r.Options)
		result = append(result, cp)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockRepo) UpsertDeliveryChannel(_ context.Context, cfg *models.DeliveryChannelConfig) error {
	key := cfg.UserID + ":" + cfg.ChannelType
	existing, ok := m.channels[key]
	if ok {
		existing.Endpoint = cfg.Endpoint
		existing.Enabled = cfg.Enabled
		existing.UpdatedAt = time.Now()
		cfg.ID = existing.ID
		cfg.CreatedAt = existing.CreatedAt
		cfg.UpdatedAt = existing.UpdatedAt
	} else {
		cfg.ID = m.nextUUID()
		cfg.CreatedAt = time.Now()
		cfg.UpdatedAt = cfg.CreatedAt
		cp := *cfg
		m.channels[key] = &cp
	}
	return nil
}

func (m *mockRepo) GetDeliveryChannel(_ context.Context, userID, channelType string) (*models.DeliveryChannelConfig, error) {
	key := userID + ":" + channelType
	cfg, ok := m.channels[key]
	if !ok {
		return nil, nil
	}
	cp := *cfg
	return &cp, nil
}

func (m *mockRepo) UpsertTimeoutPolicy(_ context.Context, policy *models.TimeoutPolicy) error {
	key := policy.Scope + ":" + policy.ScopeID
	existing, ok := m.timeoutPolicies[key]
	if ok {
		existing.TimeoutSecs = policy.TimeoutSecs
		existing.Action = policy.Action
		existing.EscalationTargets = policy.EscalationTargets
		existing.UpdatedAt = time.Now()
		policy.ID = existing.ID
		policy.CreatedAt = existing.CreatedAt
		policy.UpdatedAt = existing.UpdatedAt
	} else {
		policy.ID = m.nextUUID()
		policy.CreatedAt = time.Now()
		policy.UpdatedAt = policy.CreatedAt
		cp := *policy
		cp.EscalationTargets = copyOptions(policy.EscalationTargets)
		m.timeoutPolicies[key] = &cp
	}
	return nil
}

func (m *mockRepo) GetTimeoutPolicy(_ context.Context, scope, scopeID string) (*models.TimeoutPolicy, error) {
	key := scope + ":" + scopeID
	policy, ok := m.timeoutPolicies[key]
	if !ok {
		return nil, nil
	}
	cp := *policy
	cp.EscalationTargets = copyOptions(policy.EscalationTargets)
	return &cp, nil
}

func (m *mockRepo) ExpirePendingRequests(_ context.Context) (int, error) {
	count := 0
	now := time.Now()
	for _, r := range m.requests {
		if r.Status == models.HumanRequestStatusPending && r.ExpiresAt != nil && r.ExpiresAt.Before(now) {
			r.Status = models.HumanRequestStatusExpired
			count++
		}
	}
	return count, nil
}

func copyOptions(s []string) []string {
	if s == nil {
		return nil
	}
	cp := make([]string, len(s))
	copy(cp, s)
	return cp
}

func TestCreateRequest_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	req, err := svc.CreateRequest(context.Background(), "ws-1", "agent-1", "Approve invoice?", []string{"yes", "no"}, "Invoice #123", 300, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.ID == "" {
		t.Error("expected request ID to be set")
	}
	if req.Status != models.HumanRequestStatusPending {
		t.Errorf("expected status pending, got %q", req.Status)
	}
	if req.ExpiresAt == nil {
		t.Error("expected expires_at to be set")
	}
	if req.ExpiresAt.Before(time.Now()) {
		t.Error("expected expires_at to be in the future")
	}
}

func TestCreateRequest_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.CreateRequest(ctx, "", "a", "q", nil, "", 0, "", "", ""); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty workspace_id, got: %v", err)
	}
	if _, err := svc.CreateRequest(ctx, "ws", "", "q", nil, "", 0, "", "", ""); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty agent_id, got: %v", err)
	}
	if _, err := svc.CreateRequest(ctx, "ws", "a", "", nil, "", 0, "", "", ""); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty question, got: %v", err)
	}
}

func TestCreateRequest_DefaultTimeout(t *testing.T) {
	svc := NewService(newMockRepo())
	req, err := svc.CreateRequest(context.Background(), "ws-1", "agent-1", "q", nil, "", 0, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.ExpiresAt == nil {
		t.Fatal("expected expires_at to be set")
	}
	// Should be ~1 hour from now (default)
	diff := time.Until(*req.ExpiresAt)
	if diff < 59*time.Minute || diff > 61*time.Minute {
		t.Errorf("expected ~1h timeout, got %v", diff)
	}
}

func TestGetRequest_Found(t *testing.T) {
	svc := NewService(newMockRepo())
	created, _ := svc.CreateRequest(context.Background(), "ws", "a", "q", nil, "", 300, "", "", "")
	got, err := svc.GetRequest(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}
}

func TestGetRequest_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetRequest(context.Background(), "nonexistent-id")
	if !errors.Is(err, ErrRequestNotFound) {
		t.Errorf("expected ErrRequestNotFound, got: %v", err)
	}
}

func TestGetRequest_Expired(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	req, _ := svc.CreateRequest(ctx, "ws", "a", "q", nil, "", 1, "", "", "")
	// Manually set expires_at to the past
	pastTime := time.Now().Add(-1 * time.Minute)
	repo.requests[req.ID].ExpiresAt = &pastTime

	got, err := svc.GetRequest(ctx, req.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != models.HumanRequestStatusExpired {
		t.Errorf("expected status expired, got %q", got.Status)
	}
}

func TestRespondToRequest_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	req, _ := svc.CreateRequest(ctx, "ws", "a", "q", nil, "", 300, "", "", "")

	if err := svc.RespondToRequest(ctx, req.ID, "approved", "human-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetRequest(ctx, req.ID)
	if got.Status != models.HumanRequestStatusResponded {
		t.Errorf("expected status responded, got %q", got.Status)
	}
	if got.Response != "approved" {
		t.Errorf("expected response 'approved', got %q", got.Response)
	}
	if got.ResponderID != "human-1" {
		t.Errorf("expected responder 'human-1', got %q", got.ResponderID)
	}
}

func TestRespondToRequest_NotPending(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	req, _ := svc.CreateRequest(ctx, "ws", "a", "q", nil, "", 300, "", "", "")

	svc.RespondToRequest(ctx, req.ID, "first", "h1")

	err := svc.RespondToRequest(ctx, req.ID, "second", "h2")
	if !errors.Is(err, ErrRequestNotPending) {
		t.Errorf("expected ErrRequestNotPending on double-respond, got: %v", err)
	}
}

func TestRespondToRequest_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	err := svc.RespondToRequest(context.Background(), "no-such-id", "resp", "h1")
	if !errors.Is(err, ErrRequestNotPending) {
		t.Errorf("expected ErrRequestNotPending, got: %v", err)
	}
}

func TestRespondToRequest_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	if err := svc.RespondToRequest(ctx, "", "r", "h"); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty id, got: %v", err)
	}
	if err := svc.RespondToRequest(ctx, "id", "", "h"); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty response, got: %v", err)
	}
	if err := svc.RespondToRequest(ctx, "id", "r", ""); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty responder_id, got: %v", err)
	}
}

func TestListRequests_WithFilters(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	svc.CreateRequest(ctx, "ws-1", "a", "q1", nil, "", 300, "", "", "")
	svc.CreateRequest(ctx, "ws-2", "a", "q2", nil, "", 300, "", "", "")
	svc.CreateRequest(ctx, "ws-1", "a", "q3", nil, "", 300, "", "", "")

	// Filter by workspace
	reqs, _, err := svc.ListRequests(ctx, "ws-1", "", 50, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reqs) != 2 {
		t.Errorf("expected 2 requests for ws-1, got %d", len(reqs))
	}

	// All requests
	reqs, _, err = svc.ListRequests(ctx, "", "", 50, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reqs) != 3 {
		t.Errorf("expected 3 requests total, got %d", len(reqs))
	}
}

func TestListRequests_Pagination(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		svc.CreateRequest(ctx, "ws", "a", fmt.Sprintf("q%d", i), nil, "", 300, "", "", "")
	}

	reqs, nextToken, err := svc.ListRequests(ctx, "", "", 3, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reqs) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(reqs))
	}
	if nextToken == "" {
		t.Error("expected a next page token")
	}

	reqs2, nextToken2, err := svc.ListRequests(ctx, "", "", 3, nextToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reqs2) != 2 {
		t.Fatalf("expected 2 requests on second page, got %d", len(reqs2))
	}
	if nextToken2 != "" {
		t.Error("expected no next page token on last page")
	}
}

func TestConfigureDeliveryChannel_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	cfg, err := svc.ConfigureDeliveryChannel(ctx, "user-1", "slack", "https://hooks.slack.com/services/T00/B00/xxx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ID == "" {
		t.Error("expected config ID")
	}
	if cfg.ChannelType != "slack" {
		t.Errorf("expected channel_type 'slack', got %q", cfg.ChannelType)
	}
	if !cfg.Enabled {
		t.Error("expected enabled to be true")
	}
}

func TestConfigureDeliveryChannel_InvalidType(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.ConfigureDeliveryChannel(context.Background(), "user-1", "sms", "1234567890")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for invalid channel type, got: %v", err)
	}
}

func TestConfigureDeliveryChannel_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	if _, err := svc.ConfigureDeliveryChannel(ctx, "", "slack", "ep"); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty user_id, got: %v", err)
	}
	if _, err := svc.ConfigureDeliveryChannel(ctx, "u", "", "ep"); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty channel_type, got: %v", err)
	}
	if _, err := svc.ConfigureDeliveryChannel(ctx, "u", "slack", ""); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty endpoint, got: %v", err)
	}
}

func TestGetDeliveryChannel_Found(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	svc.ConfigureDeliveryChannel(ctx, "user-1", "email", "user@example.com")
	cfg, err := svc.GetDeliveryChannel(ctx, "user-1", "email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Endpoint != "user@example.com" {
		t.Errorf("expected endpoint 'user@example.com', got %q", cfg.Endpoint)
	}
}

func TestGetDeliveryChannel_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetDeliveryChannel(context.Background(), "user-1", "slack")
	if !errors.Is(err, ErrChannelNotFound) {
		t.Errorf("expected ErrChannelNotFound, got: %v", err)
	}
}

func TestSetTimeoutPolicy_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	policy, err := svc.SetTimeoutPolicy(ctx, "agent", "agent-1", 600, "escalate", []string{"admin@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID == "" {
		t.Error("expected policy ID")
	}
	if policy.TimeoutSecs != 600 {
		t.Errorf("expected timeout_secs 600, got %d", policy.TimeoutSecs)
	}
	if policy.Action != "escalate" {
		t.Errorf("expected action 'escalate', got %q", policy.Action)
	}
}

func TestSetTimeoutPolicy_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// Invalid scope
	if _, err := svc.SetTimeoutPolicy(ctx, "invalid", "", 600, "escalate", nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for invalid scope, got: %v", err)
	}
	// Invalid action
	if _, err := svc.SetTimeoutPolicy(ctx, "global", "", 600, "invalid", nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for invalid action, got: %v", err)
	}
	// Zero timeout
	if _, err := svc.SetTimeoutPolicy(ctx, "global", "", 0, "escalate", nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero timeout, got: %v", err)
	}
	// Non-global scope without scope_id
	if _, err := svc.SetTimeoutPolicy(ctx, "agent", "", 600, "escalate", nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for agent scope without scope_id, got: %v", err)
	}
}

func TestSetTimeoutPolicy_GlobalScope(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// Global scope allows empty scope_id
	policy, err := svc.SetTimeoutPolicy(ctx, "global", "", 300, "halt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Scope != "global" {
		t.Errorf("expected scope 'global', got %q", policy.Scope)
	}
}

func TestGetTimeoutPolicy_Found(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	svc.SetTimeoutPolicy(ctx, "workspace", "ws-1", 900, "continue", nil)
	policy, err := svc.GetTimeoutPolicy(ctx, "workspace", "ws-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.TimeoutSecs != 900 {
		t.Errorf("expected timeout_secs 900, got %d", policy.TimeoutSecs)
	}
}

func TestGetTimeoutPolicy_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetTimeoutPolicy(context.Background(), "agent", "no-such-id")
	if !errors.Is(err, ErrTimeoutPolicyNotFound) {
		t.Errorf("expected ErrTimeoutPolicyNotFound, got: %v", err)
	}
}

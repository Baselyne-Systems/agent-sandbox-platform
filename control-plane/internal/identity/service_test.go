package identity

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
	agents      map[string]*models.Agent
	credentials map[string]*models.ScopedCredential
	nextID      int
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		agents:      make(map[string]*models.Agent),
		credentials: make(map[string]*models.ScopedCredential),
	}
}

func (m *mockRepo) nextUUID() string {
	m.nextID++
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", m.nextID)
}

func (m *mockRepo) CreateAgent(_ context.Context, agent *models.Agent) error {
	agent.ID = m.nextUUID()
	agent.CreatedAt = time.Now()
	agent.UpdatedAt = agent.CreatedAt
	cp := *agent
	cp.Capabilities = copyStringSlice(agent.Capabilities)
	m.agents[agent.ID] = &cp
	return nil
}

func copyStringSlice(s []string) []string {
	if s == nil {
		return nil
	}
	cp := make([]string, len(s))
	copy(cp, s)
	return cp
}

func (m *mockRepo) GetAgent(_ context.Context, id string) (*models.Agent, error) {
	a, ok := m.agents[id]
	if !ok {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (m *mockRepo) ListAgents(_ context.Context, ownerID string, status models.AgentStatus, afterID string, limit int) ([]models.Agent, error) {
	var result []models.Agent
	for _, a := range m.agents {
		if ownerID != "" && a.OwnerID != ownerID {
			continue
		}
		if status != "" && a.Status != status {
			continue
		}
		if afterID != "" && a.ID <= afterID {
			continue
		}
		result = append(result, *a)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockRepo) DeactivateAgent(_ context.Context, id string) error {
	a, ok := m.agents[id]
	if !ok {
		return ErrAgentNotFound
	}
	a.Status = models.AgentStatusInactive
	for _, c := range m.credentials {
		if c.AgentID == id {
			c.Revoked = true
		}
	}
	return nil
}

func (m *mockRepo) CreateCredential(_ context.Context, cred *models.ScopedCredential) error {
	cred.ID = m.nextUUID()
	cred.CreatedAt = time.Now()
	cp := *cred
	m.credentials[cred.ID] = &cp
	return nil
}

func (m *mockRepo) RevokeCredential(_ context.Context, id string) error {
	c, ok := m.credentials[id]
	if !ok || c.Revoked {
		return ErrCredentialNotFound
	}
	c.Revoked = true
	return nil
}

func TestRegisterAgent_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	agent, err := svc.RegisterAgent(context.Background(), "test-agent", "A test agent", "owner-1", nil, "invoice processing", models.AgentTrustLevelNew, []string{"bash", "curl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.ID == "" {
		t.Error("expected agent ID to be set")
	}
	if agent.Name != "test-agent" {
		t.Errorf("expected name 'test-agent', got %q", agent.Name)
	}
	if agent.Status != models.AgentStatusActive {
		t.Errorf("expected status active, got %q", agent.Status)
	}
	if agent.Labels == nil {
		t.Error("expected labels to be initialized")
	}
	if agent.Purpose != "invoice processing" {
		t.Errorf("expected purpose 'invoice processing', got %q", agent.Purpose)
	}
	if agent.TrustLevel != models.AgentTrustLevelNew {
		t.Errorf("expected trust_level 'new', got %q", agent.TrustLevel)
	}
	if len(agent.Capabilities) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(agent.Capabilities))
	}
}

func TestRegisterAgent_Validation(t *testing.T) {
	svc := NewService(newMockRepo())

	if _, err := svc.RegisterAgent(context.Background(), "", "desc", "owner", nil, "", "", nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty name, got: %v", err)
	}
	if _, err := svc.RegisterAgent(context.Background(), "name", "desc", "", nil, "", "", nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty owner, got: %v", err)
	}
}

func TestRegisterAgent_Defaults(t *testing.T) {
	svc := NewService(newMockRepo())
	agent, err := svc.RegisterAgent(context.Background(), "a", "", "o", nil, "", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.TrustLevel != models.AgentTrustLevelNew {
		t.Errorf("expected default trust_level 'new', got %q", agent.TrustLevel)
	}
	if agent.Capabilities == nil {
		t.Error("expected capabilities to be initialized")
	}
}

func TestGetAgent_Found(t *testing.T) {
	svc := NewService(newMockRepo())
	created, _ := svc.RegisterAgent(context.Background(), "a", "", "o", nil, "", "", nil)
	got, err := svc.GetAgent(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}
}

func TestGetAgent_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetAgent(context.Background(), "nonexistent-id")
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound, got: %v", err)
	}
}

func TestListAgents_Pagination(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		svc.RegisterAgent(ctx, fmt.Sprintf("agent-%d", i), "", "owner", nil, "", "", nil)
	}

	agents, nextToken, err := svc.ListAgents(ctx, "", "", 3, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != 3 {
		t.Fatalf("expected 3 agents, got %d", len(agents))
	}
	if nextToken == "" {
		t.Error("expected a next page token")
	}

	agents2, nextToken2, err := svc.ListAgents(ctx, "", "", 3, nextToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents2) != 2 {
		t.Fatalf("expected 2 agents on second page, got %d", len(agents2))
	}
	if nextToken2 != "" {
		t.Error("expected no next page token on last page")
	}
}

func TestDeactivateAgent(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, "a", "", "o", nil, "", "", nil)
	svc.MintCredential(ctx, agent.ID, []string{"read"}, 3600)

	if err := svc.DeactivateAgent(ctx, agent.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetAgent(ctx, agent.ID)
	if got.Status != models.AgentStatusInactive {
		t.Errorf("expected inactive, got %q", got.Status)
	}

	// Credentials should be revoked
	for _, c := range repo.credentials {
		if c.AgentID == agent.ID && !c.Revoked {
			t.Error("expected credential to be revoked")
		}
	}
}

func TestDeactivateAgent_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	err := svc.DeactivateAgent(context.Background(), "no-such-id")
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound, got: %v", err)
	}
}

func TestMintCredential_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, "a", "", "o", nil, "", "", nil)
	cred, rawToken, err := svc.MintCredential(ctx, agent.ID, []string{"read", "write"}, 3600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred.ID == "" {
		t.Error("expected credential ID")
	}
	if rawToken == "" {
		t.Error("expected raw token")
	}
	if len(rawToken) != 64 {
		t.Errorf("expected 64-char token, got %d", len(rawToken))
	}
	if cred.ExpiresAt.Before(time.Now()) {
		t.Error("credential should not be expired")
	}
}

func TestMintCredential_InactiveAgent(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, "a", "", "o", nil, "", "", nil)
	svc.DeactivateAgent(ctx, agent.ID)

	_, _, err := svc.MintCredential(ctx, agent.ID, []string{"read"}, 3600)
	if !errors.Is(err, ErrAgentInactive) {
		t.Errorf("expected ErrAgentInactive, got: %v", err)
	}
}

func TestMintCredential_TTLBounds(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, "a", "", "o", nil, "", "", nil)

	// Zero TTL
	_, _, err := svc.MintCredential(ctx, agent.ID, []string{"read"}, 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero TTL, got: %v", err)
	}

	// Negative TTL
	_, _, err = svc.MintCredential(ctx, agent.ID, []string{"read"}, -1)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for negative TTL, got: %v", err)
	}

	// Over 24h
	_, _, err = svc.MintCredential(ctx, agent.ID, []string{"read"}, 86401)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for TTL > 24h, got: %v", err)
	}
}

func TestMintCredential_AgentNotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, _, err := svc.MintCredential(context.Background(), "no-such-agent", []string{"read"}, 3600)
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound, got: %v", err)
	}
}

func TestRevokeCredential(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, "a", "", "o", nil, "", "", nil)
	cred, _, _ := svc.MintCredential(ctx, agent.ID, []string{"read"}, 3600)

	if err := svc.RevokeCredential(ctx, cred.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Revoking again should fail
	err := svc.RevokeCredential(ctx, cred.ID)
	if !errors.Is(err, ErrCredentialNotFound) {
		t.Errorf("expected ErrCredentialNotFound on double-revoke, got: %v", err)
	}
}

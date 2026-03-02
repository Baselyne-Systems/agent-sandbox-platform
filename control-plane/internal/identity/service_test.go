package identity

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
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
	m.agents[m.tenantKey(agent.TenantID, agent.ID)] = &cp
	return nil
}

// tenantKey returns a composite key for tenant-scoped lookups.
func (m *mockRepo) tenantKey(tenantID, id string) string {
	return tenantID + "/" + id
}

func copyStringSlice(s []string) []string {
	if s == nil {
		return nil
	}
	cp := make([]string, len(s))
	copy(cp, s)
	return cp
}

func (m *mockRepo) GetAgent(_ context.Context, tenantID, id string) (*models.Agent, error) {
	a, ok := m.agents[m.tenantKey(tenantID, id)]
	if !ok {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (m *mockRepo) ListAgents(_ context.Context, tenantID, ownerID string, status models.AgentStatus, afterID string, limit int) ([]models.Agent, error) {
	var result []models.Agent
	for _, a := range m.agents {
		if a.TenantID != tenantID {
			continue
		}
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

func (m *mockRepo) DeactivateAgent(_ context.Context, tenantID, id string) error {
	a, ok := m.agents[m.tenantKey(tenantID, id)]
	if !ok {
		return ErrAgentNotFound
	}
	a.Status = models.AgentStatusInactive
	for _, c := range m.credentials {
		if c.AgentID == id && c.TenantID == tenantID {
			c.Revoked = true
		}
	}
	return nil
}

func (m *mockRepo) UpdateTrustLevel(_ context.Context, tenantID, agentID string, level models.AgentTrustLevel) error {
	a, ok := m.agents[m.tenantKey(tenantID, agentID)]
	if !ok {
		return ErrAgentNotFound
	}
	if a.Status != models.AgentStatusActive {
		return ErrAgentInactive
	}
	a.TrustLevel = level
	a.UpdatedAt = time.Now()
	return nil
}

func (m *mockRepo) UpdateAgentStatus(_ context.Context, tenantID, agentID string, from []models.AgentStatus, to models.AgentStatus) error {
	a, ok := m.agents[m.tenantKey(tenantID, agentID)]
	if !ok {
		return ErrAgentNotFound
	}
	allowed := false
	for _, s := range from {
		if a.Status == s {
			allowed = true
			break
		}
	}
	if !allowed {
		return ErrInvalidStatusTransition
	}
	a.Status = to
	a.UpdatedAt = time.Now()
	return nil
}

func (m *mockRepo) CreateCredential(_ context.Context, cred *models.ScopedCredential) error {
	cred.ID = m.nextUUID()
	cred.CreatedAt = time.Now()
	cp := *cred
	m.credentials[m.tenantKey(cred.TenantID, cred.ID)] = &cp
	return nil
}

func (m *mockRepo) RevokeCredential(_ context.Context, tenantID, id string) error {
	c, ok := m.credentials[m.tenantKey(tenantID, id)]
	if !ok || c.Revoked {
		return ErrCredentialNotFound
	}
	c.Revoked = true
	return nil
}

const testTenant = "test-tenant"

func TestRegisterAgent_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	agent, err := svc.RegisterAgent(context.Background(), testTenant, "test-agent", "A test agent", "owner-1", nil, "invoice processing", models.AgentTrustLevelNew, []string{"bash", "curl"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.ID == "" {
		t.Error("expected agent ID to be set")
	}
	if agent.TenantID != testTenant {
		t.Errorf("expected tenant_id %q, got %q", testTenant, agent.TenantID)
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

	if _, err := svc.RegisterAgent(context.Background(), "", "name", "desc", "owner", nil, "", "", nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty tenant, got: %v", err)
	}
	if _, err := svc.RegisterAgent(context.Background(), testTenant, "", "desc", "owner", nil, "", "", nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty name, got: %v", err)
	}
	if _, err := svc.RegisterAgent(context.Background(), testTenant, "name", "desc", "", nil, "", "", nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty owner, got: %v", err)
	}
}

func TestRegisterAgent_Defaults(t *testing.T) {
	svc := NewService(newMockRepo())
	agent, err := svc.RegisterAgent(context.Background(), testTenant, "a", "", "o", nil, "", "", nil)
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
	created, _ := svc.RegisterAgent(context.Background(), testTenant, "a", "", "o", nil, "", "", nil)
	got, err := svc.GetAgent(context.Background(), testTenant, created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, got.ID)
	}
}

func TestGetAgent_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.GetAgent(context.Background(), testTenant, "nonexistent-id")
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound, got: %v", err)
	}
}

func TestListAgents_Pagination(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		svc.RegisterAgent(ctx, testTenant, fmt.Sprintf("agent-%d", i), "", "owner", nil, "", "", nil)
	}

	agents, nextToken, err := svc.ListAgents(ctx, testTenant, "", "", 3, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agents) != 3 {
		t.Fatalf("expected 3 agents, got %d", len(agents))
	}
	if nextToken == "" {
		t.Error("expected a next page token")
	}

	agents2, nextToken2, err := svc.ListAgents(ctx, testTenant, "", "", 3, nextToken)
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

	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", "", nil)
	svc.MintCredential(ctx, testTenant, agent.ID, []string{"read"}, 3600)

	if err := svc.DeactivateAgent(ctx, testTenant, agent.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetAgent(ctx, testTenant, agent.ID)
	if got.Status != models.AgentStatusInactive {
		t.Errorf("expected inactive, got %q", got.Status)
	}

	// Credentials should be revoked
	for _, c := range repo.credentials {
		if c.AgentID == agent.ID && c.TenantID == testTenant && !c.Revoked {
			t.Error("expected credential to be revoked")
		}
	}
}

func TestDeactivateAgent_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	err := svc.DeactivateAgent(context.Background(), testTenant, "no-such-id")
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound, got: %v", err)
	}
}

func TestMintCredential_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", "", nil)
	cred, rawToken, err := svc.MintCredential(ctx, testTenant, agent.ID, []string{"read", "write"}, 3600)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cred.ID == "" {
		t.Error("expected credential ID")
	}
	if cred.TenantID != testTenant {
		t.Errorf("expected tenant_id %q, got %q", testTenant, cred.TenantID)
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

	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", "", nil)
	svc.DeactivateAgent(ctx, testTenant, agent.ID)

	_, _, err := svc.MintCredential(ctx, testTenant, agent.ID, []string{"read"}, 3600)
	if !errors.Is(err, ErrAgentInactive) {
		t.Errorf("expected ErrAgentInactive, got: %v", err)
	}
}

func TestMintCredential_TTLBounds(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", "", nil)

	// Zero TTL
	_, _, err := svc.MintCredential(ctx, testTenant, agent.ID, []string{"read"}, 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for zero TTL, got: %v", err)
	}

	// Negative TTL
	_, _, err = svc.MintCredential(ctx, testTenant, agent.ID, []string{"read"}, -1)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for negative TTL, got: %v", err)
	}

	// Over 24h
	_, _, err = svc.MintCredential(ctx, testTenant, agent.ID, []string{"read"}, 86401)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for TTL > 24h, got: %v", err)
	}
}

func TestMintCredential_AgentNotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, _, err := svc.MintCredential(context.Background(), testTenant, "no-such-agent", []string{"read"}, 3600)
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound, got: %v", err)
	}
}

func TestRevokeCredential(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", "", nil)
	cred, _, _ := svc.MintCredential(ctx, testTenant, agent.ID, []string{"read"}, 3600)

	if err := svc.RevokeCredential(ctx, testTenant, cred.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Revoking again should fail
	err := svc.RevokeCredential(ctx, testTenant, cred.ID)
	if !errors.Is(err, ErrCredentialNotFound) {
		t.Errorf("expected ErrCredentialNotFound on double-revoke, got: %v", err)
	}
}

func TestUpdateTrustLevel_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", models.AgentTrustLevelNew, nil)

	updated, err := svc.UpdateTrustLevel(ctx, testTenant, agent.ID, models.AgentTrustLevelEstablished, "passed validation checks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.TrustLevel != models.AgentTrustLevelEstablished {
		t.Errorf("expected trust_level 'established', got %q", updated.TrustLevel)
	}

	// Upgrade again to trusted
	updated, err = svc.UpdateTrustLevel(ctx, testTenant, agent.ID, models.AgentTrustLevelTrusted, "fully verified")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.TrustLevel != models.AgentTrustLevelTrusted {
		t.Errorf("expected trust_level 'trusted', got %q", updated.TrustLevel)
	}
}

func TestUpdateTrustLevel_InvalidLevel(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", models.AgentTrustLevelNew, nil)

	_, err := svc.UpdateTrustLevel(ctx, testTenant, agent.ID, "bogus", "no reason")
	if !errors.Is(err, ErrInvalidTrustLevel) {
		t.Errorf("expected ErrInvalidTrustLevel, got: %v", err)
	}
}

func TestUpdateTrustLevel_AgentNotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.UpdateTrustLevel(context.Background(), testTenant, "no-such-agent", models.AgentTrustLevelEstablished, "test")
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound, got: %v", err)
	}
}

func TestUpdateTrustLevel_DeactivatedAgent(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", models.AgentTrustLevelNew, nil)
	svc.DeactivateAgent(ctx, testTenant, agent.ID)

	_, err := svc.UpdateTrustLevel(ctx, testTenant, agent.ID, models.AgentTrustLevelEstablished, "test")
	if !errors.Is(err, ErrAgentInactive) {
		t.Errorf("expected ErrAgentInactive, got: %v", err)
	}
}

func TestUpdateTrustLevel_EmptyID(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.UpdateTrustLevel(context.Background(), testTenant, "", models.AgentTrustLevelEstablished, "test")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

// --- SuspendAgent tests ---

func TestSuspendAgent_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", "", nil)

	suspended, err := svc.SuspendAgent(ctx, testTenant, agent.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if suspended.Status != models.AgentStatusSuspended {
		t.Errorf("expected suspended, got %q", suspended.Status)
	}
}

func TestSuspendAgent_AlreadySuspended(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", "", nil)

	svc.SuspendAgent(ctx, testTenant, agent.ID)
	// Suspending again should be idempotent.
	suspended, err := svc.SuspendAgent(ctx, testTenant, agent.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if suspended.Status != models.AgentStatusSuspended {
		t.Errorf("expected suspended, got %q", suspended.Status)
	}
}

func TestSuspendAgent_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.SuspendAgent(context.Background(), testTenant, "nonexistent")
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound, got: %v", err)
	}
}

func TestSuspendAgent_EmptyID(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.SuspendAgent(context.Background(), testTenant, "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got: %v", err)
	}
}

func TestSuspendAgent_InactiveAgent(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", "", nil)
	svc.DeactivateAgent(ctx, testTenant, agent.ID)

	_, err := svc.SuspendAgent(ctx, testTenant, agent.ID)
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Errorf("expected ErrInvalidStatusTransition, got: %v", err)
	}
}

// --- ReactivateAgent tests ---

func TestReactivateAgent_FromSuspended(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", "", nil)
	svc.SuspendAgent(ctx, testTenant, agent.ID)

	reactivated, err := svc.ReactivateAgent(ctx, testTenant, agent.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reactivated.Status != models.AgentStatusActive {
		t.Errorf("expected active, got %q", reactivated.Status)
	}
}

func TestReactivateAgent_FromInactive(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", "", nil)
	svc.DeactivateAgent(ctx, testTenant, agent.ID)

	reactivated, err := svc.ReactivateAgent(ctx, testTenant, agent.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reactivated.Status != models.AgentStatusActive {
		t.Errorf("expected active, got %q", reactivated.Status)
	}
}

func TestReactivateAgent_AlreadyActive(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()
	agent, _ := svc.RegisterAgent(ctx, testTenant, "a", "", "o", nil, "", "", nil)

	reactivated, err := svc.ReactivateAgent(ctx, testTenant, agent.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reactivated.Status != models.AgentStatusActive {
		t.Errorf("expected active, got %q", reactivated.Status)
	}
}

func TestReactivateAgent_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.ReactivateAgent(context.Background(), testTenant, "nonexistent")
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound, got: %v", err)
	}
}

// --- Cross-tenant isolation tests ---

func TestCrossTenantIsolation_GetAgent(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// Create agent in tenant-A
	agent, err := svc.RegisterAgent(ctx, "tenant-A", "isolated-agent", "", "o", nil, "", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Lookup from tenant-A should succeed
	got, err := svc.GetAgent(ctx, "tenant-A", agent.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching from tenant-A: %v", err)
	}
	if got.ID != agent.ID {
		t.Errorf("expected agent ID %q, got %q", agent.ID, got.ID)
	}

	// Lookup from tenant-B with the same agent ID should return not found
	_, err = svc.GetAgent(ctx, "tenant-B", agent.ID)
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound for cross-tenant lookup, got: %v", err)
	}
}

func TestCrossTenantIsolation_ListAgents(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// Create agents in two different tenants
	svc.RegisterAgent(ctx, "tenant-A", "agent-A1", "", "o", nil, "", "", nil)
	svc.RegisterAgent(ctx, "tenant-A", "agent-A2", "", "o", nil, "", "", nil)
	svc.RegisterAgent(ctx, "tenant-B", "agent-B1", "", "o", nil, "", "", nil)

	// List from tenant-A should only see 2 agents
	agentsA, _, err := svc.ListAgents(ctx, "tenant-A", "", "", 100, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agentsA) != 2 {
		t.Errorf("expected 2 agents for tenant-A, got %d", len(agentsA))
	}

	// List from tenant-B should only see 1 agent
	agentsB, _, err := svc.ListAgents(ctx, "tenant-B", "", "", 100, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agentsB) != 1 {
		t.Errorf("expected 1 agent for tenant-B, got %d", len(agentsB))
	}
}

func TestCrossTenantIsolation_DeactivateAgent(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	// Create agent in tenant-A
	agent, _ := svc.RegisterAgent(ctx, "tenant-A", "a", "", "o", nil, "", "", nil)

	// Attempt to deactivate from tenant-B should fail
	err := svc.DeactivateAgent(ctx, "tenant-B", agent.ID)
	if !errors.Is(err, ErrAgentNotFound) {
		t.Errorf("expected ErrAgentNotFound for cross-tenant deactivation, got: %v", err)
	}

	// Original tenant-A agent should still be active
	got, _ := svc.GetAgent(ctx, "tenant-A", agent.ID)
	if got.Status != models.AgentStatusActive {
		t.Errorf("expected agent to still be active, got %q", got.Status)
	}
}

func TestCrossTenantIsolation_RevokeCredential(t *testing.T) {
	svc := NewService(newMockRepo())
	ctx := context.Background()

	agent, _ := svc.RegisterAgent(ctx, "tenant-A", "a", "", "o", nil, "", "", nil)
	cred, _, _ := svc.MintCredential(ctx, "tenant-A", agent.ID, []string{"read"}, 3600)

	// Attempt to revoke from tenant-B should fail
	err := svc.RevokeCredential(ctx, "tenant-B", cred.ID)
	if !errors.Is(err, ErrCredentialNotFound) {
		t.Errorf("expected ErrCredentialNotFound for cross-tenant revocation, got: %v", err)
	}

	// Revoking from tenant-A should succeed
	if err := svc.RevokeCredential(ctx, "tenant-A", cred.ID); err != nil {
		t.Fatalf("unexpected error revoking from correct tenant: %v", err)
	}
}

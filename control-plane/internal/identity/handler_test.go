package identity

import (
	"context"
	"testing"

	pb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/identity/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestHandler() *Handler {
	return NewHandler(NewService(newMockRepo()))
}

func TestHandler_RegisterAgent_Success(t *testing.T) {
	h := newTestHandler()
	resp, err := h.RegisterAgent(context.Background(), &pb.RegisterAgentRequest{
		Name:        "test-agent",
		Description: "A test agent",
		OwnerId:     "owner-1",
		Labels:      map[string]string{"env": "test"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Agent == nil {
		t.Fatal("expected agent in response")
	}
	if resp.Agent.AgentId == "" {
		t.Error("expected agent ID")
	}
	if resp.Agent.Name != "test-agent" {
		t.Errorf("name = %q, want 'test-agent'", resp.Agent.Name)
	}
	if resp.Agent.OwnerId != "owner-1" {
		t.Errorf("owner_id = %q, want 'owner-1'", resp.Agent.OwnerId)
	}
	if resp.Agent.Status != pb.AgentStatus_AGENT_STATUS_ACTIVE {
		t.Errorf("status = %v, want ACTIVE", resp.Agent.Status)
	}
	if resp.Agent.Labels["env"] != "test" {
		t.Errorf("labels[env] = %q, want 'test'", resp.Agent.Labels["env"])
	}
	if resp.Agent.CreatedAt == nil {
		t.Error("expected created_at timestamp")
	}
}

func TestHandler_RegisterAgent_InvalidInput(t *testing.T) {
	h := newTestHandler()
	_, err := h.RegisterAgent(context.Background(), &pb.RegisterAgentRequest{
		Name:    "",
		OwnerId: "owner-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestHandler_GetAgent_Success(t *testing.T) {
	h := newTestHandler()
	created, _ := h.RegisterAgent(context.Background(), &pb.RegisterAgentRequest{
		Name: "a", OwnerId: "o",
	})

	resp, err := h.GetAgent(context.Background(), &pb.GetAgentRequest{
		AgentId: created.Agent.AgentId,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Agent.AgentId != created.Agent.AgentId {
		t.Errorf("ID mismatch: got %q, want %q", resp.Agent.AgentId, created.Agent.AgentId)
	}
}

func TestHandler_GetAgent_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.GetAgent(context.Background(), &pb.GetAgentRequest{
		AgentId: "nonexistent",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_ListAgents_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		h.RegisterAgent(ctx, &pb.RegisterAgentRequest{
			Name: "agent", OwnerId: "owner",
		})
	}

	resp, err := h.ListAgents(ctx, &pb.ListAgentsRequest{PageSize: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Agents) != 3 {
		t.Errorf("agents count = %d, want 3", len(resp.Agents))
	}
}

func TestHandler_ListAgents_StatusFilter(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	h.RegisterAgent(ctx, &pb.RegisterAgentRequest{Name: "a", OwnerId: "o"})
	created, _ := h.RegisterAgent(ctx, &pb.RegisterAgentRequest{Name: "b", OwnerId: "o"})
	h.DeactivateAgent(ctx, &pb.DeactivateAgentRequest{AgentId: created.Agent.AgentId})

	resp, err := h.ListAgents(ctx, &pb.ListAgentsRequest{
		Status:   pb.AgentStatus_AGENT_STATUS_ACTIVE,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Agents) != 1 {
		t.Errorf("active agents = %d, want 1", len(resp.Agents))
	}
}

func TestHandler_DeactivateAgent_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	created, _ := h.RegisterAgent(ctx, &pb.RegisterAgentRequest{Name: "a", OwnerId: "o"})
	_, err := h.DeactivateAgent(ctx, &pb.DeactivateAgentRequest{AgentId: created.Agent.AgentId})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := h.GetAgent(ctx, &pb.GetAgentRequest{AgentId: created.Agent.AgentId})
	if got.Agent.Status != pb.AgentStatus_AGENT_STATUS_INACTIVE {
		t.Errorf("status = %v, want INACTIVE", got.Agent.Status)
	}
}

func TestHandler_DeactivateAgent_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.DeactivateAgent(context.Background(), &pb.DeactivateAgentRequest{
		AgentId: "nonexistent",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_MintCredential_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	agent, _ := h.RegisterAgent(ctx, &pb.RegisterAgentRequest{Name: "a", OwnerId: "o"})

	resp, err := h.MintCredential(ctx, &pb.MintCredentialRequest{
		AgentId:    agent.Agent.AgentId,
		Scopes:     []string{"read", "write"},
		TtlSeconds: 3600,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.Credential == nil {
		t.Fatal("expected credential in response")
	}
	if resp.Credential.CredentialId == "" {
		t.Error("expected credential ID")
	}
	if resp.Credential.AgentId != agent.Agent.AgentId {
		t.Errorf("agent_id = %q, want %q", resp.Credential.AgentId, agent.Agent.AgentId)
	}
	if len(resp.Credential.Scopes) != 2 {
		t.Errorf("scopes len = %d, want 2", len(resp.Credential.Scopes))
	}
	if resp.Credential.ExpiresAt == nil {
		t.Error("expected expires_at timestamp")
	}
}

func TestHandler_MintCredential_InactiveAgent(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	agent, _ := h.RegisterAgent(ctx, &pb.RegisterAgentRequest{Name: "a", OwnerId: "o"})
	h.DeactivateAgent(ctx, &pb.DeactivateAgentRequest{AgentId: agent.Agent.AgentId})

	_, err := h.MintCredential(ctx, &pb.MintCredentialRequest{
		AgentId:    agent.Agent.AgentId,
		Scopes:     []string{"read"},
		TtlSeconds: 3600,
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.FailedPrecondition {
		t.Errorf("code = %v, want FailedPrecondition", st.Code())
	}
}

func TestHandler_RevokeCredential_Success(t *testing.T) {
	h := newTestHandler()
	ctx := context.Background()

	agent, _ := h.RegisterAgent(ctx, &pb.RegisterAgentRequest{Name: "a", OwnerId: "o"})
	cred, _ := h.MintCredential(ctx, &pb.MintCredentialRequest{
		AgentId: agent.Agent.AgentId, Scopes: []string{"read"}, TtlSeconds: 3600,
	})

	_, err := h.RevokeCredential(ctx, &pb.RevokeCredentialRequest{
		CredentialId: cred.Credential.CredentialId,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandler_RevokeCredential_NotFound(t *testing.T) {
	h := newTestHandler()
	_, err := h.RevokeCredential(context.Background(), &pb.RevokeCredentialRequest{
		CredentialId: "nonexistent",
	})
	st, _ := status.FromError(err)
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestHandler_StatusConversion_RoundTrip(t *testing.T) {
	tests := []struct {
		proto pb.AgentStatus
		model string
	}{
		{pb.AgentStatus_AGENT_STATUS_ACTIVE, "active"},
		{pb.AgentStatus_AGENT_STATUS_INACTIVE, "inactive"},
		{pb.AgentStatus_AGENT_STATUS_SUSPENDED, "suspended"},
		{pb.AgentStatus_AGENT_STATUS_UNSPECIFIED, ""},
	}
	for _, tt := range tests {
		model := protoStatusToModel(tt.proto)
		if string(model) != tt.model {
			t.Errorf("protoStatusToModel(%v) = %q, want %q", tt.proto, model, tt.model)
		}
		if tt.model != "" {
			back := modelStatusToProto(model)
			if back != tt.proto {
				t.Errorf("modelStatusToProto(%q) = %v, want %v", model, back, tt.proto)
			}
		}
	}
}

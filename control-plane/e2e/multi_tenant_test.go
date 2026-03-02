package e2e

import (
	"context"
	"testing"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/guardrails"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/identity"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/task"
)

func TestCrossTenantIsolation(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenantA := uniqueTenant()
	tenantB := uniqueTenant()

	agentA := registerAgent(t, ctx, tenantA, "agent-A")
	agentB := registerAgent(t, ctx, tenantB, "agent-B")

	_, err := taskSvc.CreateTask(ctx, tenantA, agentA.ID, "task A", nil, "", nil, nil, 0, nil, nil)
	if err != nil {
		t.Fatalf("create task A: %v", err)
	}
	taskB, err := taskSvc.CreateTask(ctx, tenantB, agentB.ID, "task B", nil, "", nil, nil, 0, nil, nil)
	if err != nil {
		t.Fatalf("create task B: %v", err)
	}

	ruleA, err := guardrailsSvc.CreateRule(ctx, tenantA, "rule-A", "desc", models.RuleTypeToolFilter, "shell", models.RuleActionDeny, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create rule A: %v", err)
	}

	if _, err := identitySvc.GetAgent(ctx, tenantA, agentA.ID); err != nil {
		t.Fatalf("tenantA get own agent: %v", err)
	}

	_, err = identitySvc.GetAgent(ctx, tenantA, agentB.ID)
	if err != identity.ErrAgentNotFound {
		t.Fatalf("expected ErrAgentNotFound for cross-tenant agent, got %v", err)
	}
	_, err = taskSvc.GetTask(ctx, tenantA, taskB.ID)
	if err != task.ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound for cross-tenant task, got %v", err)
	}

	_, err = guardrailsSvc.GetRule(ctx, tenantB, ruleA.ID)
	if err != guardrails.ErrRuleNotFound {
		t.Fatalf("expected ErrRuleNotFound for cross-tenant rule, got %v", err)
	}

	agentsA, _, err := identitySvc.ListAgents(ctx, tenantA, "", "", 50, "")
	if err != nil {
		t.Fatalf("list agents A: %v", err)
	}
	if len(agentsA) != 1 || agentsA[0].ID != agentA.ID {
		t.Fatalf("tenantA should see exactly 1 agent, got %d", len(agentsA))
	}

	agentsB, _, err := identitySvc.ListAgents(ctx, tenantB, "", "", 50, "")
	if err != nil {
		t.Fatalf("list agents B: %v", err)
	}
	if len(agentsB) != 1 || agentsB[0].ID != agentB.ID {
		t.Fatalf("tenantB should see exactly 1 agent, got %d", len(agentsB))
	}
}

func TestSharedComputeHosts(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenantA := uniqueTenant()
	tenantB := uniqueTenant()

	registerHost(t, ctx, "shared-host.local:9090", 8192, 8000, 20480, []string{"standard", "hardened"})

	agentA := registerAgent(t, ctx, tenantA, "shared-A")
	agentB := registerAgent(t, ctx, tenantB, "shared-B")

	wsA, err := workspaceSvc.CreateWorkspace(ctx, tenantA, agentA.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace A: %v", err)
	}
	if wsA.Status != models.WorkspaceStatusRunning {
		t.Fatalf("expected running for A, got %s", wsA.Status)
	}

	wsB, err := workspaceSvc.CreateWorkspace(ctx, tenantB, agentB.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace B: %v", err)
	}
	if wsB.Status != models.WorkspaceStatusRunning {
		t.Fatalf("expected running for B, got %s", wsB.Status)
	}

	if wsA.HostID != wsB.HostID {
		t.Fatalf("expected same host, got A=%s B=%s", wsA.HostID, wsB.HostID)
	}
}

func TestTenantImplicitCreation(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent, err := identitySvc.RegisterAgent(ctx, tenant, "first-agent", "desc", "owner", nil, "", models.AgentTrustLevelNew, nil)
	if err != nil {
		t.Fatalf("register with new tenant: %v", err)
	}
	if agent.TenantID != tenant {
		t.Fatalf("expected tenant %s, got %s", tenant, agent.TenantID)
	}
}

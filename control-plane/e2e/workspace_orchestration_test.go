package e2e

import (
	"context"
	"testing"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

func TestWorkspaceFullOrchestration(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "orch-agent")
	registerHost(t, ctx, "orch-host.local:9090", 4096, 4000, 10240, []string{"standard", "hardened"})

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if ws.Status != models.WorkspaceStatusRunning {
		t.Fatalf("expected running, got %s", ws.Status)
	}
	if ws.HostID == "" {
		t.Fatal("expected host ID to be set")
	}
	if ws.SandboxID == "" {
		t.Fatal("expected sandbox ID to be set")
	}
	if fakeHostAgent.createCalls.Load() != 1 {
		t.Fatalf("expected 1 create sandbox call, got %d", fakeHostAgent.createCalls.Load())
	}
}

func TestWorkspaceSnapshotRestore(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "snapshot-agent")
	registerHost(t, ctx, "snap-host.local:9090", 4096, 4000, 10240, []string{"standard", "hardened"})

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	snap, err := workspaceSvc.SnapshotWorkspace(ctx, tenant, ws.ID)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snap.WorkspaceID != ws.ID {
		t.Fatalf("expected workspace ID %s in snapshot, got %s", ws.ID, snap.WorkspaceID)
	}

	ws, err = workspaceSvc.GetWorkspace(ctx, tenant, ws.ID)
	if err != nil {
		t.Fatalf("get after snapshot: %v", err)
	}
	if ws.Status != models.WorkspaceStatusPaused {
		t.Fatalf("expected paused after snapshot, got %s", ws.Status)
	}

	restored, err := workspaceSvc.RestoreWorkspace(ctx, tenant, snap.ID)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if restored.Status != models.WorkspaceStatusRunning {
		t.Fatalf("expected running after restore, got %s", restored.Status)
	}
}

func TestWorkspaceTermination(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "term-agent")
	registerHost(t, ctx, "term-host.local:9090", 4096, 4000, 10240, []string{"standard", "hardened"})

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	err = workspaceSvc.TerminateWorkspace(ctx, tenant, ws.ID, "test termination")
	if err != nil {
		t.Fatalf("terminate: %v", err)
	}

	ws, err = workspaceSvc.GetWorkspace(ctx, tenant, ws.ID)
	if err != nil {
		t.Fatalf("get after terminate: %v", err)
	}
	if ws.Status != models.WorkspaceStatusTerminated {
		t.Fatalf("expected terminated, got %s", ws.Status)
	}
	if fakeHostAgent.destroyCalls.Load() != 1 {
		t.Fatalf("expected 1 destroy sandbox call, got %d", fakeHostAgent.destroyCalls.Load())
	}
}

func TestWorkspaceCredentialInjection(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "cred-agent")
	registerHost(t, ctx, "cred-host.local:9090", 4096, 4000, 10240, []string{"standard", "hardened"})

	// MintCredential requires non-empty scopes and positive TTL.
	spec := &models.WorkspaceSpec{
		AllowedTools:    []string{"web_search", "file_read"},
		MaxDurationSecs: 3600,
	}

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", spec)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if ws.Status != models.WorkspaceStatusRunning {
		t.Fatalf("expected running, got %s", ws.Status)
	}

	fakeHostAgent.mu.Lock()
	lastReq := fakeHostAgent.lastCreateReq
	fakeHostAgent.mu.Unlock()

	if lastReq == nil {
		t.Fatal("expected a CreateSandboxRequest to have been captured")
	}

	envVars := lastReq.GetSpec().GetEnvVars()
	if envVars == nil {
		t.Fatal("expected env vars in create sandbox request")
	}

	if envVars["BULKHEAD_AGENT_ID"] == "" {
		t.Fatal("expected BULKHEAD_AGENT_ID in env vars")
	}
	if envVars["BULKHEAD_AGENT_TOKEN"] == "" {
		t.Fatal("expected BULKHEAD_AGENT_TOKEN in env vars")
	}
}

func TestWorkspaceFailsGracefully(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "fail-agent")
	registerHost(t, ctx, "fail-host.local:9090", 4096, 4000, 10240, []string{"standard", "hardened"})

	fakeHostAgent.failCreate.Store(true)
	defer fakeHostAgent.failCreate.Store(false)

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace should not return error, got: %v", err)
	}
	if ws.Status != models.WorkspaceStatusFailed {
		t.Fatalf("expected failed status, got %s", ws.Status)
	}
}

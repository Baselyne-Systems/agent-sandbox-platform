package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

func TestHumanRequestLifecycle(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "human-agent")
	registerHost(t, ctx, "human-host.local:9090", 4096, 4000, 10240, []string{"standard"})

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	req, err := humanSvc.CreateRequest(ctx, tenant, ws.ID, agent.ID, "approve deploy?",
		[]string{"yes", "no"}, "deploy context", 300, models.HumanRequestTypeApproval,
		models.HumanRequestUrgencyNormal, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if req.Status != models.HumanRequestStatusPending {
		t.Fatalf("expected pending, got %s", req.Status)
	}
	if req.Question != "approve deploy?" {
		t.Fatalf("expected question 'approve deploy?', got %q", req.Question)
	}

	got, err := humanSvc.GetRequest(ctx, tenant, req.ID)
	if err != nil {
		t.Fatalf("get request: %v", err)
	}
	if got.ID != req.ID {
		t.Fatalf("expected ID %s, got %s", req.ID, got.ID)
	}

	err = humanSvc.RespondToRequest(ctx, tenant, req.ID, "yes", "responder-1")
	if err != nil {
		t.Fatalf("respond: %v", err)
	}

	got, err = humanSvc.GetRequest(ctx, tenant, req.ID)
	if err != nil {
		t.Fatalf("get after respond: %v", err)
	}
	if got.Status != models.HumanRequestStatusResponded {
		t.Fatalf("expected responded, got %s", got.Status)
	}
	if got.Response != "yes" {
		t.Fatalf("expected response 'yes', got %q", got.Response)
	}
	if got.ResponderID != "responder-1" {
		t.Fatalf("expected responder 'responder-1', got %q", got.ResponderID)
	}
}

func TestHumanRequestExpiration(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "expire-agent")
	registerHost(t, ctx, "expire-host.local:9090", 4096, 4000, 10240, []string{"standard"})

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	req, err := humanSvc.CreateRequest(ctx, tenant, ws.ID, agent.ID, "quick question",
		nil, "", 1, models.HumanRequestTypeQuestion,
		models.HumanRequestUrgencyLow, "")
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	time.Sleep(2 * time.Second)

	got, err := humanSvc.GetRequest(ctx, tenant, req.ID)
	if err != nil {
		t.Fatalf("get expired request: %v", err)
	}
	if got.Status != models.HumanRequestStatusExpired {
		t.Fatalf("expected expired, got %s", got.Status)
	}
}

func TestDeliveryChannelConfiguration(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	channel, err := humanSvc.ConfigureDeliveryChannel(ctx, tenant, "user-1", "slack", "https://hooks.slack.com/test")
	if err != nil {
		t.Fatalf("configure channel: %v", err)
	}
	if channel.ChannelType != "slack" {
		t.Fatalf("expected slack, got %s", channel.ChannelType)
	}
	if channel.Endpoint != "https://hooks.slack.com/test" {
		t.Fatalf("expected slack endpoint, got %s", channel.Endpoint)
	}

	got, err := humanSvc.GetDeliveryChannel(ctx, tenant, "user-1", "slack")
	if err != nil {
		t.Fatalf("get channel: %v", err)
	}
	if got.ID != channel.ID {
		t.Fatalf("expected ID %s, got %s", channel.ID, got.ID)
	}
	if !got.Enabled {
		t.Fatal("expected channel enabled")
	}
}

func TestTimeoutPolicies(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "timeout-agent")

	globalPolicy, err := humanSvc.SetTimeoutPolicy(ctx, tenant, "global", "", 300, "escalate", []string{"admin@example.com"})
	if err != nil {
		t.Fatalf("set global policy: %v", err)
	}
	if globalPolicy.Scope != "global" {
		t.Fatalf("expected global scope, got %s", globalPolicy.Scope)
	}
	if globalPolicy.TimeoutSecs != 300 {
		t.Fatalf("expected 300s timeout, got %d", globalPolicy.TimeoutSecs)
	}
	if globalPolicy.Action != "escalate" {
		t.Fatalf("expected escalate action, got %s", globalPolicy.Action)
	}

	agentPolicy, err := humanSvc.SetTimeoutPolicy(ctx, tenant, "agent", agent.ID, 60, "halt", nil)
	if err != nil {
		t.Fatalf("set agent policy: %v", err)
	}
	if agentPolicy.Scope != "agent" {
		t.Fatalf("expected agent scope, got %s", agentPolicy.Scope)
	}
	if agentPolicy.ScopeID != agent.ID {
		t.Fatalf("expected scope ID %s, got %s", agent.ID, agentPolicy.ScopeID)
	}

	gotGlobal, err := humanSvc.GetTimeoutPolicy(ctx, tenant, "global", "")
	if err != nil {
		t.Fatalf("get global policy: %v", err)
	}
	if gotGlobal.TimeoutSecs != 300 {
		t.Fatalf("expected 300s, got %d", gotGlobal.TimeoutSecs)
	}

	gotAgent, err := humanSvc.GetTimeoutPolicy(ctx, tenant, "agent", agent.ID)
	if err != nil {
		t.Fatalf("get agent policy: %v", err)
	}
	if gotAgent.TimeoutSecs != 60 {
		t.Fatalf("expected 60s, got %d", gotAgent.TimeoutSecs)
	}
}

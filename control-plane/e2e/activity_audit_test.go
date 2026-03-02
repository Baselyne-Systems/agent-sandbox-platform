package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/activity"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

func TestActionRecordingAndQuery(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "audit-agent")
	registerHost(t, ctx, "audit-host.local:9090", 8192, 8000, 20480, []string{"standard"})

	ws1, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace 1: %v", err)
	}
	ws2, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace 2: %v", err)
	}

	emptyParams := json.RawMessage(`{}`)
	actions := []models.ActionRecord{
		{TenantID: tenant, WorkspaceID: ws1.ID, AgentID: agent.ID, ToolName: "shell", Parameters: emptyParams, Outcome: models.ActionOutcomeDenied, DenialReason: "blocked by policy"},
		{TenantID: tenant, WorkspaceID: ws1.ID, AgentID: agent.ID, ToolName: "web_search", Parameters: emptyParams, Outcome: models.ActionOutcomeAllowed},
		{TenantID: tenant, WorkspaceID: ws1.ID, AgentID: agent.ID, ToolName: "shell", Parameters: emptyParams, Outcome: models.ActionOutcomeDenied, DenialReason: "blocked by policy"},
		{TenantID: tenant, WorkspaceID: ws2.ID, AgentID: agent.ID, ToolName: "file_read", Parameters: emptyParams, Outcome: models.ActionOutcomeAllowed},
		{TenantID: tenant, WorkspaceID: ws1.ID, AgentID: agent.ID, ToolName: "web_search", Parameters: emptyParams, Outcome: models.ActionOutcomeError},
	}
	for i, a := range actions {
		_, err := activitySvc.RecordAction(ctx, &a)
		if err != nil {
			t.Fatalf("record action %d: %v", i, err)
		}
	}

	all, _, err := activitySvc.QueryActions(ctx, tenant, activity.QueryFilter{WorkspaceID: ws1.ID, Limit: 50})
	if err != nil {
		t.Fatalf("query all ws-1: %v", err)
	}
	if len(all) != 4 {
		t.Fatalf("expected 4 actions for ws-1, got %d", len(all))
	}

	shells, _, err := activitySvc.QueryActions(ctx, tenant, activity.QueryFilter{ToolName: "shell", Limit: 50})
	if err != nil {
		t.Fatalf("query shell: %v", err)
	}
	if len(shells) != 2 {
		t.Fatalf("expected 2 shell actions, got %d", len(shells))
	}

	denied, _, err := activitySvc.QueryActions(ctx, tenant, activity.QueryFilter{Outcome: models.ActionOutcomeDenied, Limit: 50})
	if err != nil {
		t.Fatalf("query denied: %v", err)
	}
	if len(denied) != 2 {
		t.Fatalf("expected 2 denied actions, got %d", len(denied))
	}
}

func TestActionExportJSON(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "export-json-agent")
	registerHost(t, ctx, "export-json-host.local:9090", 4096, 4000, 10240, []string{"standard"})

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	for i := 0; i < 3; i++ {
		_, err := activitySvc.RecordAction(ctx, &models.ActionRecord{
			TenantID:    tenant,
			WorkspaceID: ws.ID,
			AgentID:     agent.ID,
			ToolName:    "web_search",
			Parameters:  json.RawMessage(`{}`),
			Outcome:     models.ActionOutcomeAllowed,
		})
		if err != nil {
			t.Fatalf("record action %d: %v", i, err)
		}
	}

	var buf bytes.Buffer
	err = activitySvc.ExportActions(ctx, tenant,
		activity.QueryFilter{WorkspaceID: ws.ID, Limit: 50},
		activity.ExportFormatJSON,
		func(data []byte, count int, isLast bool) error {
			buf.Write(data)
			return nil
		})
	if err != nil {
		t.Fatalf("export JSON: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 NDJSON lines, got %d", len(lines))
	}

	for i, line := range lines {
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("unmarshal line %d: %v", i, err)
		}
		if record["tool_name"] != "web_search" {
			t.Fatalf("expected tool_name=web_search in line %d, got %v", i, record["tool_name"])
		}
	}
}

func TestActionExportCSV(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "export-csv-agent")
	registerHost(t, ctx, "export-csv-host.local:9090", 4096, 4000, 10240, []string{"standard"})

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	for i := 0; i < 3; i++ {
		_, err := activitySvc.RecordAction(ctx, &models.ActionRecord{
			TenantID:    tenant,
			WorkspaceID: ws.ID,
			AgentID:     agent.ID,
			ToolName:    "file_read",
			Parameters:  json.RawMessage(`{}`),
			Outcome:     models.ActionOutcomeAllowed,
		})
		if err != nil {
			t.Fatalf("record action %d: %v", i, err)
		}
	}

	var buf bytes.Buffer
	err = activitySvc.ExportActions(ctx, tenant,
		activity.QueryFilter{WorkspaceID: ws.ID, Limit: 50},
		activity.ExportFormatCSV,
		func(data []byte, count int, isLast bool) error {
			buf.Write(data)
			return nil
		})
	if err != nil {
		t.Fatalf("export CSV: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 4 {
		t.Fatalf("expected header + 3 rows, got %d lines", len(lines))
	}
	header := lines[0]
	if !strings.Contains(header, "tool_name") {
		t.Fatalf("expected CSV header to contain 'tool_name', got %q", header)
	}
}

func TestAlertConfigurationWithoutRepo(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	_, err := activitySvc.ConfigureAlert(ctx, tenant, "high-denial-rate",
		models.AlertConditionDenialRate, 0.5, "", "https://hooks.example.com/alerts")
	if err != activity.ErrAlertsNotEnabled {
		t.Fatalf("expected ErrAlertsNotEnabled (no alert repo configured), got %v", err)
	}
}

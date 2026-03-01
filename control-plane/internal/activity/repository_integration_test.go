//go:build integration

package activity

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/testutil"
)

// wsUUID generates a deterministic valid UUID for workspace IDs in tests.
func wsUUID(n int) string {
	return fmt.Sprintf("a0000000-0000-0000-0000-%012d", n)
}

var testDB *testutil.TestDB

func TestMain(m *testing.M) {
	testDB = testutil.MustSetupTestDB()
	code := m.Run()
	testDB.Cleanup()
	os.Exit(code)
}

func setup(t *testing.T) (*PostgresRepository, *sql.DB) {
	t.Helper()
	testutil.TruncateAll(t, testDB.DB)
	return NewPostgresRepository(testDB.DB), testDB.DB
}

func TestInteg_InsertAndGetAction_JSONRoundTrip(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "json-agent")

	params := json.RawMessage(`{"key":"value","nested":{"a":1}}`)
	result := json.RawMessage(`{"status":"ok"}`)
	evalLatency := int64(150)
	execLatency := int64(5000)

	rec := &models.ActionRecord{
		WorkspaceID:         wsUUID(1),
		AgentID:             agentID,
		ToolName:            "http_fetch",
		Parameters:          params,
		Result:              result,
		Outcome:             models.ActionOutcomeAllowed,
		EvaluationLatencyUs: &evalLatency,
		ExecutionLatencyUs:  &execLatency,
	}
	if err := repo.InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	if rec.ID == "" {
		t.Fatal("expected server-generated ID")
	}
	if rec.RecordedAt.IsZero() {
		t.Fatal("expected server-generated recorded_at")
	}

	got, err := repo.GetAction(ctx, rec.ID)
	if err != nil {
		t.Fatalf("GetAction: %v", err)
	}
	if got == nil {
		t.Fatal("expected record, got nil")
	}

	// Verify JSON round-trip
	var gotParams map[string]any
	if err := json.Unmarshal(got.Parameters, &gotParams); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	if gotParams["key"] != "value" {
		t.Errorf("params[key] = %v, want value", gotParams["key"])
	}

	var gotResult map[string]any
	if err := json.Unmarshal(got.Result, &gotResult); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if gotResult["status"] != "ok" {
		t.Errorf("result[status] = %v, want ok", gotResult["status"])
	}

	if got.ToolName != "http_fetch" {
		t.Errorf("tool_name = %q, want %q", got.ToolName, "http_fetch")
	}
	if got.Outcome != models.ActionOutcomeAllowed {
		t.Errorf("outcome = %q, want %q", got.Outcome, models.ActionOutcomeAllowed)
	}
	if got.EvaluationLatencyUs == nil || *got.EvaluationLatencyUs != 150 {
		t.Errorf("eval latency = %v, want 150", got.EvaluationLatencyUs)
	}
}

func TestInteg_InsertAction_NullableFields(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "nullable-agent")

	rec := &models.ActionRecord{
		WorkspaceID:         wsUUID(2),
		AgentID:             agentID,
		TaskID:              "", // empty → NULL
		ToolName:            "bash",
		Parameters:          json.RawMessage(`{}`), // NOT NULL column — use empty object
		Result:              nil,                   // nullable column — test NULL round-trip
		Outcome:             models.ActionOutcomeDenied,
		GuardrailRuleID:     "",
		DenialReason:        "",
		EvaluationLatencyUs: nil,
		ExecutionLatencyUs:  nil,
	}
	if err := repo.InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	got, err := repo.GetAction(ctx, rec.ID)
	if err != nil {
		t.Fatalf("GetAction: %v", err)
	}
	if got.TaskID != "" {
		t.Errorf("TaskID = %q, want empty", got.TaskID)
	}
	if got.EvaluationLatencyUs != nil {
		t.Errorf("EvaluationLatencyUs = %v, want nil", got.EvaluationLatencyUs)
	}
	if got.ExecutionLatencyUs != nil {
		t.Errorf("ExecutionLatencyUs = %v, want nil", got.ExecutionLatencyUs)
	}
}

func TestInteg_GetAction_NotFound(t *testing.T) {
	repo, _ := setup(t)
	ctx := context.Background()

	got, err := repo.GetAction(ctx, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestInteg_QueryActions_WorkspaceAndAgentFilter(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agent1 := testutil.SeedAgent(t, db, "filter-agent-1")
	agent2 := testutil.SeedAgent(t, db, "filter-agent-2")

	wsA := wsUUID(10)
	wsB := wsUUID(11)
	records := []struct {
		wsID    string
		agentID string
	}{
		{wsA, agent1},
		{wsA, agent2},
		{wsB, agent1},
	}
	for _, r := range records {
		rec := &models.ActionRecord{
			WorkspaceID: r.wsID,
			AgentID:     r.agentID,
			ToolName:    "test",
			Outcome:     models.ActionOutcomeAllowed,
			Parameters:  json.RawMessage(`{}`),
		}
		if err := repo.InsertAction(ctx, rec); err != nil {
			t.Fatalf("InsertAction: %v", err)
		}
	}

	// Filter by workspace
	byWs, err := repo.QueryActions(ctx, QueryFilter{WorkspaceID: wsA, Limit: 10})
	if err != nil {
		t.Fatalf("QueryActions workspace: %v", err)
	}
	if len(byWs) != 2 {
		t.Errorf("workspace filter count = %d, want 2", len(byWs))
	}

	// Filter by agent
	byAgent, err := repo.QueryActions(ctx, QueryFilter{AgentID: agent1, Limit: 10})
	if err != nil {
		t.Fatalf("QueryActions agent: %v", err)
	}
	if len(byAgent) != 2 {
		t.Errorf("agent filter count = %d, want 2", len(byAgent))
	}
}

func TestInteg_QueryActions_TimeRange(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "time-agent")

	// Insert a record — its recorded_at is set by the DB to now()
	rec := &models.ActionRecord{
		WorkspaceID: wsUUID(20),
		AgentID:     agentID,
		ToolName:    "test",
		Outcome:     models.ActionOutcomeAllowed,
		Parameters:  json.RawMessage(`{}`),
	}
	if err := repo.InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	now := time.Now()
	past := now.Add(-1 * time.Minute)
	future := now.Add(1 * time.Minute)

	// Should find within range
	inRange, err := repo.QueryActions(ctx, QueryFilter{StartTime: &past, EndTime: &future, Limit: 10})
	if err != nil {
		t.Fatalf("QueryActions in range: %v", err)
	}
	if len(inRange) != 1 {
		t.Errorf("in-range count = %d, want 1", len(inRange))
	}

	// Should not find outside range
	farPast := now.Add(-2 * time.Hour)
	farPastEnd := now.Add(-1 * time.Hour)
	outRange, err := repo.QueryActions(ctx, QueryFilter{StartTime: &farPast, EndTime: &farPastEnd, Limit: 10})
	if err != nil {
		t.Fatalf("QueryActions out range: %v", err)
	}
	if len(outRange) != 0 {
		t.Errorf("out-range count = %d, want 0", len(outRange))
	}
}

func TestInteg_QueryActions_Pagination(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "page-agent")

	var ids []string
	for i := 0; i < 5; i++ {
		rec := &models.ActionRecord{
			WorkspaceID: wsUUID(30),
			AgentID:     agentID,
			ToolName:    "test",
			Outcome:     models.ActionOutcomeAllowed,
			Parameters:  json.RawMessage(`{}`),
		}
		if err := repo.InsertAction(ctx, rec); err != nil {
			t.Fatalf("InsertAction[%d]: %v", i, err)
		}
		ids = append(ids, rec.ID)
	}
	sort.Strings(ids)

	page1, err := repo.QueryActions(ctx, QueryFilter{Limit: 2})
	if err != nil {
		t.Fatalf("QueryActions page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("page1 len = %d, want 2", len(page1))
	}

	page2, err := repo.QueryActions(ctx, QueryFilter{AfterID: page1[1].ID, Limit: 2})
	if err != nil {
		t.Fatalf("QueryActions page2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("page2 len = %d, want 2", len(page2))
	}

	page3, err := repo.QueryActions(ctx, QueryFilter{AfterID: page2[1].ID, Limit: 2})
	if err != nil {
		t.Fatalf("QueryActions page3: %v", err)
	}
	if len(page3) != 1 {
		t.Fatalf("page3 len = %d, want 1", len(page3))
	}
}

func TestInteg_ImmutableActions_UpdateRejected(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "immutable-agent")

	rec := &models.ActionRecord{
		WorkspaceID: wsUUID(50),
		AgentID:     agentID,
		ToolName:    "test",
		Outcome:     models.ActionOutcomeAllowed,
		Parameters:  json.RawMessage(`{}`),
	}
	if err := repo.InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	// Attempt to UPDATE — should be rejected by the immutability trigger.
	_, err := db.ExecContext(ctx, `UPDATE action_records SET tool_name = 'hacked' WHERE id = $1`, rec.ID)
	if err == nil {
		t.Fatal("expected UPDATE to be rejected by immutability trigger, but it succeeded")
	}
	if !strings.Contains(err.Error(), "immutable") {
		t.Errorf("expected error to mention 'immutable', got: %v", err)
	}
}

func TestInteg_ImmutableActions_DeleteRejected(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "immutable-agent-2")

	rec := &models.ActionRecord{
		WorkspaceID: wsUUID(51),
		AgentID:     agentID,
		ToolName:    "test",
		Outcome:     models.ActionOutcomeAllowed,
		Parameters:  json.RawMessage(`{}`),
	}
	if err := repo.InsertAction(ctx, rec); err != nil {
		t.Fatalf("InsertAction: %v", err)
	}

	// Attempt to DELETE — should be rejected by the immutability trigger.
	_, err := db.ExecContext(ctx, `DELETE FROM action_records WHERE id = $1`, rec.ID)
	if err == nil {
		t.Fatal("expected DELETE to be rejected by immutability trigger, but it succeeded")
	}
	if !strings.Contains(err.Error(), "immutable") {
		t.Errorf("expected error to mention 'immutable', got: %v", err)
	}
}

func TestInteg_QueryActions_OutcomeFilter(t *testing.T) {
	repo, db := setup(t)
	ctx := context.Background()
	agentID := testutil.SeedAgent(t, db, "outcome-agent")

	outcomes := []models.ActionOutcome{
		models.ActionOutcomeAllowed,
		models.ActionOutcomeDenied,
		models.ActionOutcomeAllowed,
	}
	for _, o := range outcomes {
		rec := &models.ActionRecord{
			WorkspaceID: wsUUID(40),
			AgentID:     agentID,
			ToolName:    "test",
			Outcome:     o,
			Parameters:  json.RawMessage(`{}`),
		}
		if err := repo.InsertAction(ctx, rec); err != nil {
			t.Fatalf("InsertAction: %v", err)
		}
	}

	denied, err := repo.QueryActions(ctx, QueryFilter{Outcome: models.ActionOutcomeDenied, Limit: 10})
	if err != nil {
		t.Fatalf("QueryActions denied: %v", err)
	}
	if len(denied) != 1 {
		t.Errorf("denied count = %d, want 1", len(denied))
	}
}

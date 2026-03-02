package guardrails

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

// mockActivityQuerier implements ActivityQuerier for tests.
type mockActivityQuerier struct {
	records []models.ActionRecord
	err     error
}

func (m *mockActivityQuerier) QueryActions(_ context.Context, filter ActivityQueryFilter) ([]models.ActionRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []models.ActionRecord
	for _, r := range m.records {
		if filter.AgentID != "" && r.AgentID != filter.AgentID {
			continue
		}
		if filter.StartTime != nil && r.RecordedAt.Before(*filter.StartTime) {
			continue
		}
		if filter.EndTime != nil && r.RecordedAt.After(*filter.EndTime) {
			continue
		}
		result = append(result, r)
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result, nil
}

func makeRecord(agentID, tool string, outcome models.ActionOutcome, at time.Time) models.ActionRecord {
	return models.ActionRecord{
		AgentID:     agentID,
		WorkspaceID: "ws-1",
		ToolName:    tool,
		Outcome:     outcome,
		RecordedAt:  at,
	}
}

func TestConsideredEvaluator_NormalBehavior(t *testing.T) {
	now := time.Now()
	querier := &mockActivityQuerier{
		records: []models.ActionRecord{
			makeRecord("agent-1", "read_file", models.ActionOutcomeAllowed, now.Add(-2*time.Minute)),
			makeRecord("agent-1", "write_file", models.ActionOutcomeAllowed, now.Add(-1*time.Minute)),
			makeRecord("agent-1", "list_dir", models.ActionOutcomeAllowed, now),
		},
	}

	eval := NewConsideredEvaluator(querier, 5*time.Minute, 10*time.Minute)
	report, err := eval.GenerateReport(context.Background(), "agent-1", now.Add(-10*time.Minute), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ActionCount != 3 {
		t.Errorf("expected 3 actions, got %d", report.ActionCount)
	}
	if report.DenialRate != 0 {
		t.Errorf("expected 0 denial rate, got %f", report.DenialRate)
	}
	if report.ErrorRate != 0 {
		t.Errorf("expected 0 error rate, got %f", report.ErrorRate)
	}
	if len(report.Flags) != 0 {
		t.Errorf("expected no flags, got %v", report.Flags)
	}
	if report.Recommendation != "behavior normal" {
		t.Errorf("expected 'behavior normal', got %q", report.Recommendation)
	}
}

func TestConsideredEvaluator_HighDenialRate(t *testing.T) {
	now := time.Now()
	records := make([]models.ActionRecord, 10)
	for i := range records {
		outcome := models.ActionOutcomeDenied
		if i < 3 {
			outcome = models.ActionOutcomeAllowed
		}
		records[i] = makeRecord("agent-1", "exec", outcome, now.Add(-time.Duration(10-i)*time.Minute))
	}

	querier := &mockActivityQuerier{records: records}
	eval := NewConsideredEvaluator(querier, 5*time.Minute, 15*time.Minute)

	report, err := eval.GenerateReport(context.Background(), "agent-1", now.Add(-15*time.Minute), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.DenialRate < 0.5 {
		t.Errorf("expected denial rate > 50%%, got %f", report.DenialRate)
	}

	found := false
	for _, f := range report.Flags {
		if f == "high_denial_rate:70%" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected high_denial_rate flag, got %v", report.Flags)
	}
}

func TestConsideredEvaluator_HighErrorRate(t *testing.T) {
	now := time.Now()
	records := make([]models.ActionRecord, 10)
	for i := range records {
		outcome := models.ActionOutcomeError
		if i < 5 {
			outcome = models.ActionOutcomeAllowed
		}
		records[i] = makeRecord("agent-1", "api_call", outcome, now.Add(-time.Duration(10-i)*time.Minute))
	}

	querier := &mockActivityQuerier{records: records}
	eval := NewConsideredEvaluator(querier, 5*time.Minute, 15*time.Minute)

	report, err := eval.GenerateReport(context.Background(), "agent-1", now.Add(-15*time.Minute), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ErrorRate <= 0.3 {
		t.Errorf("expected error rate > 30%%, got %f", report.ErrorRate)
	}

	found := false
	for _, f := range report.Flags {
		if f == "high_error_rate:50%" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected high_error_rate flag, got %v", report.Flags)
	}
	if report.Recommendation != "high error rate — agent may need assistance or tool configuration fix" {
		t.Errorf("unexpected recommendation: %q", report.Recommendation)
	}
}

func TestConsideredEvaluator_HighVelocity(t *testing.T) {
	now := time.Now()
	records := make([]models.ActionRecord, 150)
	for i := range records {
		records[i] = makeRecord("agent-1", "read_file", models.ActionOutcomeAllowed, now.Add(-time.Duration(150-i)*time.Second))
	}

	querier := &mockActivityQuerier{records: records}
	eval := NewConsideredEvaluator(querier, 5*time.Minute, 10*time.Minute)

	report, err := eval.GenerateReport(context.Background(), "agent-1", now.Add(-10*time.Minute), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ActionCount != 150 {
		t.Errorf("expected 150 actions, got %d", report.ActionCount)
	}

	found := false
	for _, f := range report.Flags {
		if f == "high_velocity:150_actions" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected high_velocity flag, got %v", report.Flags)
	}
}

func TestConsideredEvaluator_StuckAgent(t *testing.T) {
	now := time.Now()
	records := make([]models.ActionRecord, 7)
	// 2 allowed, then 5 consecutive errors on same tool
	records[0] = makeRecord("agent-1", "read_file", models.ActionOutcomeAllowed, now.Add(-7*time.Minute))
	records[1] = makeRecord("agent-1", "read_file", models.ActionOutcomeAllowed, now.Add(-6*time.Minute))
	for i := 2; i < 7; i++ {
		records[i] = makeRecord("agent-1", "api_call", models.ActionOutcomeError, now.Add(-time.Duration(7-i)*time.Minute))
	}

	querier := &mockActivityQuerier{records: records}
	eval := NewConsideredEvaluator(querier, 5*time.Minute, 10*time.Minute)

	report, err := eval.GenerateReport(context.Background(), "agent-1", now.Add(-10*time.Minute), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, f := range report.Flags {
		if f == "stuck_agent:repeated_errors_on_api_call" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected stuck_agent flag, got %v", report.Flags)
	}
}

func TestConsideredEvaluator_NoActivity(t *testing.T) {
	querier := &mockActivityQuerier{records: nil}
	eval := NewConsideredEvaluator(querier, 5*time.Minute, 10*time.Minute)

	now := time.Now()
	report, err := eval.GenerateReport(context.Background(), "agent-1", now.Add(-10*time.Minute), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ActionCount != 0 {
		t.Errorf("expected 0 actions, got %d", report.ActionCount)
	}
	if report.Recommendation != "no activity in window" {
		t.Errorf("expected 'no activity in window', got %q", report.Recommendation)
	}
}

func TestConsideredEvaluator_Validation(t *testing.T) {
	querier := &mockActivityQuerier{}
	eval := NewConsideredEvaluator(querier, 5*time.Minute, 10*time.Minute)
	ctx := context.Background()
	now := time.Now()

	// Empty agent ID
	_, err := eval.GenerateReport(ctx, "", now.Add(-time.Hour), now)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty agent_id, got: %v", err)
	}

	// End before start
	_, err = eval.GenerateReport(ctx, "agent-1", now, now.Add(-time.Hour))
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for end before start, got: %v", err)
	}
}

func TestConsideredEvaluator_QueryError(t *testing.T) {
	querier := &mockActivityQuerier{err: errors.New("db connection failed")}
	eval := NewConsideredEvaluator(querier, 5*time.Minute, 10*time.Minute)

	now := time.Now()
	_, err := eval.GenerateReport(context.Background(), "agent-1", now.Add(-time.Hour), now)
	if err == nil {
		t.Error("expected error from query failure")
	}
}

func TestGetBehaviorReport_NotConfigured(t *testing.T) {
	svc := NewService(newMockRepo())
	now := time.Now()
	_, err := svc.GetBehaviorReport(context.Background(), "agent-1", now.Add(-time.Hour), now)
	if err == nil {
		t.Error("expected error when considered evaluator not configured")
	}
}

func TestGetBehaviorReport_ViaService(t *testing.T) {
	querier := &mockActivityQuerier{
		records: []models.ActionRecord{
			makeRecord("agent-1", "read_file", models.ActionOutcomeAllowed, time.Now().Add(-5*time.Minute)),
		},
	}
	eval := NewConsideredEvaluator(querier, 5*time.Minute, 10*time.Minute)
	svc := NewService(newMockRepo())
	svc.SetConsideredEvaluator(eval)

	now := time.Now()
	report, err := svc.GetBehaviorReport(context.Background(), "agent-1", now.Add(-10*time.Minute), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.AgentID != "agent-1" {
		t.Errorf("expected agent_id 'agent-1', got %q", report.AgentID)
	}
	if report.ActionCount != 1 {
		t.Errorf("expected 1 action, got %d", report.ActionCount)
	}
}

func TestDetectStuckAgent_NotEnoughActions(t *testing.T) {
	// Less than threshold actions — should not detect stuck
	records := []models.ActionRecord{
		makeRecord("a", "tool", models.ActionOutcomeError, time.Now()),
	}
	if tool := detectStuckAgent(records); tool != "" {
		t.Errorf("expected no stuck agent, got %q", tool)
	}
}

func TestDetectStuckAgent_DifferentTools(t *testing.T) {
	now := time.Now()
	records := make([]models.ActionRecord, 5)
	for i := range records {
		// Different tools for each error — shouldn't be "stuck"
		records[i] = makeRecord("a", "tool-"+string(rune('a'+i)), models.ActionOutcomeError, now.Add(time.Duration(i)*time.Second))
	}
	if tool := detectStuckAgent(records); tool != "" {
		t.Errorf("expected no stuck agent (different tools), got %q", tool)
	}
}

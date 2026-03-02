package guardrails

import (
	"context"
	"fmt"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

// ActivityQuerier provides read access to the activity store for behavior analysis.
type ActivityQuerier interface {
	QueryActions(ctx context.Context, filter ActivityQueryFilter) ([]models.ActionRecord, error)
}

// ActivityQueryFilter mirrors the activity store's query filter.
type ActivityQueryFilter struct {
	AgentID   string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
}

// ErrAlertEmissionNotImplemented is returned by Run() because the background
// evaluation loop detects anomalies but has no way to emit alerts yet.
// A real implementation would call Activity Store's ConfigureAlert/trigger
// mechanism or escalate via HIS.
var ErrAlertEmissionNotImplemented = fmt.Errorf("considered evaluator: background alert emission not yet implemented (use GetBehaviorReport RPC for on-demand reports)")

// ConsideredEvaluator performs periodic behavior analysis on agent activity,
// implementing the "considered" evaluation tier from the spec.
type ConsideredEvaluator struct {
	activity ActivityQuerier
	interval time.Duration // evaluation loop interval
	window   time.Duration // lookback window for analysis
}

// NewConsideredEvaluator creates a new considered evaluator.
func NewConsideredEvaluator(activity ActivityQuerier, interval, window time.Duration) *ConsideredEvaluator {
	return &ConsideredEvaluator{
		activity: activity,
		interval: interval,
		window:   window,
	}
}

// Thresholds for anomaly detection.
const (
	thresholdDenialRate     = 0.5 // 50%+ denial rate is suspicious
	thresholdErrorRate      = 0.3 // 30%+ error rate is suspicious
	thresholdActionVelocity = 100 // 100+ actions per window is suspicious
	thresholdStuckRepeat    = 5   // 5+ consecutive same-tool errors = stuck
)

// Run would start the background considered evaluation loop, periodically
// analyzing agent behavior and emitting alerts for anomalies. This is not
// yet implemented — the analysis logic works (see GenerateReport) but there
// is no alert emission path wired up. Use the GetBehaviorReport RPC for
// on-demand reports instead.
func (e *ConsideredEvaluator) Run(_ context.Context) error {
	return ErrAlertEmissionNotImplemented
}

// GenerateReport produces a BehaviorReport for a specific agent over a time window.
func (e *ConsideredEvaluator) GenerateReport(ctx context.Context, agentID string, windowStart, windowEnd time.Time) (*models.BehaviorReport, error) {
	if agentID == "" {
		return nil, ErrInvalidInput
	}
	if windowEnd.Before(windowStart) {
		return nil, ErrInvalidInput
	}

	records, err := e.activity.QueryActions(ctx, ActivityQueryFilter{
		AgentID:   agentID,
		StartTime: &windowStart,
		EndTime:   &windowEnd,
		Limit:     10000,
	})
	if err != nil {
		return nil, err
	}

	return e.analyzeAgent(agentID, records, windowStart, windowEnd), nil
}

// analyzeAgent computes behavior metrics and flags anomalies.
func (e *ConsideredEvaluator) analyzeAgent(agentID string, actions []models.ActionRecord, windowStart, windowEnd time.Time) *models.BehaviorReport {
	report := &models.BehaviorReport{
		AgentID:     agentID,
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
		ActionCount: int64(len(actions)),
	}

	if len(actions) == 0 {
		report.Recommendation = "no activity in window"
		return report
	}

	// Count outcomes.
	var denied, errored int
	for _, a := range actions {
		switch a.Outcome {
		case models.ActionOutcomeDenied:
			denied++
		case models.ActionOutcomeError:
			errored++
		}
	}

	report.DenialRate = float64(denied) / float64(len(actions))
	report.ErrorRate = float64(errored) / float64(len(actions))

	// Flag: high denial rate.
	if report.DenialRate > thresholdDenialRate {
		report.Flags = append(report.Flags, fmt.Sprintf("high_denial_rate:%.0f%%", report.DenialRate*100))
	}

	// Flag: high error rate.
	if report.ErrorRate > thresholdErrorRate {
		report.Flags = append(report.Flags, fmt.Sprintf("high_error_rate:%.0f%%", report.ErrorRate*100))
	}

	// Flag: high action velocity.
	if report.ActionCount > thresholdActionVelocity {
		report.Flags = append(report.Flags, fmt.Sprintf("high_velocity:%d_actions", report.ActionCount))
	}

	// Flag: stuck agent (repeated errors on same tool).
	if stuckTool := detectStuckAgent(actions); stuckTool != "" {
		report.Flags = append(report.Flags, fmt.Sprintf("stuck_agent:repeated_errors_on_%s", stuckTool))
	}

	// Generate recommendation.
	switch {
	case len(report.Flags) == 0:
		report.Recommendation = "behavior normal"
	case report.DenialRate > thresholdDenialRate && report.ActionCount > thresholdActionVelocity:
		report.Recommendation = "agent may be probing boundaries — consider restricting permissions"
	case report.ErrorRate > thresholdErrorRate:
		report.Recommendation = "high error rate — agent may need assistance or tool configuration fix"
	default:
		report.Recommendation = "anomalous behavior detected — review flags"
	}

	return report
}

// detectStuckAgent checks for consecutive errors on the same tool.
func detectStuckAgent(actions []models.ActionRecord) string {
	if len(actions) < thresholdStuckRepeat {
		return ""
	}

	var consecutiveErrors int
	var lastErrorTool string

	for _, a := range actions {
		if a.Outcome == models.ActionOutcomeError {
			if a.ToolName == lastErrorTool || lastErrorTool == "" {
				lastErrorTool = a.ToolName
				consecutiveErrors++
				if consecutiveErrors >= thresholdStuckRepeat {
					return lastErrorTool
				}
			} else {
				lastErrorTool = a.ToolName
				consecutiveErrors = 1
			}
		} else {
			consecutiveErrors = 0
			lastErrorTool = ""
		}
	}

	return ""
}

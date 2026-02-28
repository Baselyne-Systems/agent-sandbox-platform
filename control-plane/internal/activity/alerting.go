package activity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

const (
	defaultAlertWindow   = 5 * time.Minute
	defaultAlertInterval = 30 * time.Second
)

// AlertEngine periodically evaluates alert conditions against recent activity.
type AlertEngine struct {
	activity  Repository
	alerts    AlertRepository
	interval  time.Duration
	window    time.Duration
	httpClient *http.Client
}

// NewAlertEngine creates an alert engine that checks conditions at the given interval.
func NewAlertEngine(activity Repository, alerts AlertRepository) *AlertEngine {
	return &AlertEngine{
		activity:  activity,
		alerts:    alerts,
		interval:  defaultAlertInterval,
		window:    defaultAlertWindow,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Run starts the background evaluation loop. Blocks until ctx is cancelled.
func (e *AlertEngine) Run(ctx context.Context) {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.evaluate(ctx)
		}
	}
}

// evaluate checks all enabled alert configs against recent activity.
func (e *AlertEngine) evaluate(ctx context.Context) {
	configs, err := e.alerts.ListAlertConfigs(ctx, true)
	if err != nil {
		log.Printf("alerting: failed to list configs: %v", err)
		return
	}

	now := time.Now()
	windowStart := now.Add(-e.window)

	for _, cfg := range configs {
		e.evaluateConfig(ctx, cfg, windowStart, now)
	}
}

func (e *AlertEngine) evaluateConfig(ctx context.Context, cfg models.AlertConfig, windowStart, windowEnd time.Time) {
	// Determine which agents to evaluate. If scoped, just one; otherwise query distinct agents.
	agentIDs := []string{cfg.AgentID}
	if cfg.AgentID == "" {
		// Unscoped — evaluate a global aggregate with empty agent_id filter.
		agentIDs = []string{""}
	}

	for _, agentID := range agentIDs {
		triggered, message := e.checkCondition(ctx, cfg, agentID, windowStart, windowEnd)

		existing, _ := e.alerts.ActiveAlertForConfig(ctx, cfg.ID, agentID)

		if triggered && existing == nil {
			// Fire new alert.
			alert := &models.Alert{
				ConfigID:      cfg.ID,
				AgentID:       agentID,
				ConditionType: cfg.ConditionType,
				Message:       message,
				TriggeredAt:   time.Now(),
				Resolved:      false,
			}
			if err := e.alerts.CreateAlert(ctx, alert); err != nil {
				log.Printf("alerting: failed to create alert: %v", err)
				continue
			}
			e.sendWebhook(cfg.WebhookURL, alert)
		} else if !triggered && existing != nil {
			// Auto-resolve.
			_ = e.alerts.ResolveAlert(ctx, existing.ID)
		}
	}
}

func (e *AlertEngine) checkCondition(ctx context.Context, cfg models.AlertConfig, agentID string, start, end time.Time) (bool, string) {
	filter := QueryFilter{
		AgentID:   agentID,
		StartTime: &start,
		EndTime:   &end,
		Limit:     10000, // large enough window
	}
	records, err := e.activity.QueryActions(ctx, filter)
	if err != nil || len(records) == 0 {
		return false, ""
	}

	switch cfg.ConditionType {
	case models.AlertConditionDenialRate:
		denials := 0
		for _, r := range records {
			if r.Outcome == models.ActionOutcomeDenied {
				denials++
			}
		}
		rate := float64(denials) / float64(len(records))
		if rate > cfg.Threshold {
			return true, fmt.Sprintf("denial rate %.1f%% exceeds threshold %.1f%%", rate*100, cfg.Threshold*100)
		}

	case models.AlertConditionErrorRate:
		errors := 0
		for _, r := range records {
			if r.Outcome == models.ActionOutcomeError {
				errors++
			}
		}
		rate := float64(errors) / float64(len(records))
		if rate > cfg.Threshold {
			return true, fmt.Sprintf("error rate %.1f%% exceeds threshold %.1f%%", rate*100, cfg.Threshold*100)
		}

	case models.AlertConditionActionVelocity:
		// Threshold = max actions per minute.
		windowMins := end.Sub(start).Minutes()
		if windowMins <= 0 {
			windowMins = 1
		}
		velocity := float64(len(records)) / windowMins
		if velocity > cfg.Threshold {
			return true, fmt.Sprintf("action velocity %.1f/min exceeds threshold %.1f/min", velocity, cfg.Threshold)
		}

	case models.AlertConditionStuckAgent:
		// Check if all recent actions are errors on the same tool.
		if len(records) < 3 {
			return false, ""
		}
		toolName := records[0].ToolName
		allSameTool := true
		allErrors := true
		for _, r := range records {
			if r.ToolName != toolName {
				allSameTool = false
				break
			}
			if r.Outcome != models.ActionOutcomeError {
				allErrors = false
			}
		}
		if allSameTool && allErrors {
			return true, fmt.Sprintf("agent appears stuck: %d consecutive errors on tool %q", len(records), toolName)
		}

	case models.AlertConditionBudgetBreach:
		// Budget breach alerts are triggered by the economics service via RecordAction.
		// Check for denied actions with budget-related denial reasons.
		budgetDenials := 0
		for _, r := range records {
			if r.Outcome == models.ActionOutcomeDenied && r.DenialReason != "" {
				budgetDenials++
			}
		}
		if float64(budgetDenials) > cfg.Threshold {
			return true, fmt.Sprintf("%d budget-related denials exceed threshold %.0f", budgetDenials, cfg.Threshold)
		}
	}

	return false, ""
}

func (e *AlertEngine) sendWebhook(url string, alert *models.Alert) {
	if url == "" {
		return
	}
	payload, err := json.Marshal(map[string]any{
		"alert_id":       alert.ID,
		"config_id":      alert.ConfigID,
		"agent_id":       alert.AgentID,
		"condition_type": string(alert.ConditionType),
		"message":        alert.Message,
		"triggered_at":   alert.TriggeredAt.Format(time.RFC3339),
	})
	if err != nil {
		log.Printf("alerting: failed to marshal webhook payload: %v", err)
		return
	}

	resp, err := e.httpClient.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Printf("alerting: webhook POST to %s failed: %v", url, err)
		return
	}
	resp.Body.Close()
}

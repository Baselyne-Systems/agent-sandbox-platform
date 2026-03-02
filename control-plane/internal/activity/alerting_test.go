package activity

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

// mockAlertRepo is a hand-written in-memory AlertRepository for testing.
type mockAlertRepo struct {
	mu      sync.Mutex
	configs map[string]*models.AlertConfig
	alerts  map[string]*models.Alert
	nextID  int
}

func newMockAlertRepo() *mockAlertRepo {
	return &mockAlertRepo{
		configs: make(map[string]*models.AlertConfig),
		alerts:  make(map[string]*models.Alert),
	}
}

func (m *mockAlertRepo) nextUUID() string {
	m.nextID++
	return fmt.Sprintf("alert-%012d", m.nextID)
}

func (m *mockAlertRepo) UpsertAlertConfig(_ context.Context, config *models.AlertConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if config.ID == "" {
		config.ID = m.nextUUID()
		config.CreatedAt = time.Now()
	}
	cp := *config
	m.configs[config.ID] = &cp
	return nil
}

func (m *mockAlertRepo) GetAlertConfig(_ context.Context, id string) (*models.AlertConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.configs[id]
	if !ok {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}

func (m *mockAlertRepo) ListAlertConfigs(_ context.Context, tenantID string, enabledOnly bool) ([]models.AlertConfig, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []models.AlertConfig
	for _, c := range m.configs {
		if tenantID != "" && c.TenantID != tenantID {
			continue
		}
		if enabledOnly && !c.Enabled {
			continue
		}
		result = append(result, *c)
	}
	return result, nil
}

func (m *mockAlertRepo) CreateAlert(_ context.Context, alert *models.Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	alert.ID = m.nextUUID()
	cp := *alert
	m.alerts[alert.ID] = &cp
	return nil
}

func (m *mockAlertRepo) GetAlert(_ context.Context, id string) (*models.Alert, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alerts[id]
	if !ok {
		return nil, nil
	}
	cp := *a
	return &cp, nil
}

func (m *mockAlertRepo) ListAlerts(_ context.Context, agentID string, activeOnly bool, afterID string, limit int) ([]models.Alert, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []models.Alert
	for _, a := range m.alerts {
		if agentID != "" && a.AgentID != agentID {
			continue
		}
		if activeOnly && a.Resolved {
			continue
		}
		if afterID != "" && a.ID <= afterID {
			continue
		}
		result = append(result, *a)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockAlertRepo) ResolveAlert(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.alerts[id]
	if !ok {
		return errors.New("alert not found")
	}
	a.Resolved = true
	return nil
}

func (m *mockAlertRepo) ActiveAlertForConfig(_ context.Context, configID, agentID string) (*models.Alert, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.alerts {
		if a.ConfigID == configID && a.AgentID == agentID && !a.Resolved {
			cp := *a
			return &cp, nil
		}
	}
	return nil, nil
}

// --- Service-level alert tests ---

func TestConfigureAlert_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	svc.SetAlertRepository(newMockAlertRepo())

	config, err := svc.ConfigureAlert(context.Background(), "tenant-1", "high-denial", models.AlertConditionDenialRate, 0.5, "agent-1", "https://hooks.example.com/alert")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.ID == "" {
		t.Error("expected non-empty config ID")
	}
	if config.Name != "high-denial" {
		t.Errorf("expected name 'high-denial', got %q", config.Name)
	}
	if !config.Enabled {
		t.Error("expected config to be enabled")
	}
}

func TestConfigureAlert_Validation(t *testing.T) {
	svc := NewService(newMockRepo())
	svc.SetAlertRepository(newMockAlertRepo())
	ctx := context.Background()

	tests := []struct {
		name          string
		alertName     string
		conditionType models.AlertConditionType
		threshold     float64
	}{
		{"empty name", "", models.AlertConditionDenialRate, 0.5},
		{"invalid condition type", "test", "invalid", 0.5},
		{"zero threshold", "test", models.AlertConditionDenialRate, 0},
		{"negative threshold", "test", models.AlertConditionDenialRate, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ConfigureAlert(ctx, "tenant-1", tt.alertName, tt.conditionType, tt.threshold, "", "")
			if !errors.Is(err, ErrInvalidInput) {
				t.Errorf("expected ErrInvalidInput, got: %v", err)
			}
		})
	}
}

func TestConfigureAlert_NotEnabled(t *testing.T) {
	svc := NewService(newMockRepo())
	// No alert repo set.
	_, err := svc.ConfigureAlert(context.Background(), "tenant-1", "test", models.AlertConditionDenialRate, 0.5, "", "")
	if !errors.Is(err, ErrAlertsNotEnabled) {
		t.Errorf("expected ErrAlertsNotEnabled, got: %v", err)
	}
}

func TestGetAlert_Found(t *testing.T) {
	svc := NewService(newMockRepo())
	alertRepo := newMockAlertRepo()
	svc.SetAlertRepository(alertRepo)
	ctx := context.Background()

	// Create an alert directly.
	alert := &models.Alert{
		ConfigID:      "cfg-1",
		AgentID:       "agent-1",
		ConditionType: models.AlertConditionDenialRate,
		Message:       "test alert",
		TriggeredAt:   time.Now(),
	}
	alertRepo.CreateAlert(ctx, alert)

	got, err := svc.GetAlert(ctx, alert.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Message != "test alert" {
		t.Errorf("expected message 'test alert', got %q", got.Message)
	}
}

func TestGetAlert_NotFound(t *testing.T) {
	svc := NewService(newMockRepo())
	svc.SetAlertRepository(newMockAlertRepo())

	_, err := svc.GetAlert(context.Background(), "nonexistent")
	if !errors.Is(err, ErrAlertNotFound) {
		t.Errorf("expected ErrAlertNotFound, got: %v", err)
	}
}

func TestListAlerts_WithFilter(t *testing.T) {
	svc := NewService(newMockRepo())
	alertRepo := newMockAlertRepo()
	svc.SetAlertRepository(alertRepo)
	ctx := context.Background()

	// Create alerts for different agents.
	for i := 0; i < 5; i++ {
		alertRepo.CreateAlert(ctx, &models.Alert{
			ConfigID:      "cfg-1",
			AgentID:       fmt.Sprintf("agent-%d", i%2),
			ConditionType: models.AlertConditionDenialRate,
			Message:       fmt.Sprintf("alert %d", i),
			TriggeredAt:   time.Now(),
		})
	}

	alerts, _, err := svc.ListAlerts(ctx, "agent-0", false, 50, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 3 {
		t.Errorf("expected 3 alerts for agent-0, got %d", len(alerts))
	}
}

func TestListAlerts_ActiveOnly(t *testing.T) {
	svc := NewService(newMockRepo())
	alertRepo := newMockAlertRepo()
	svc.SetAlertRepository(alertRepo)
	ctx := context.Background()

	// Create 2 active + 1 resolved.
	alertRepo.CreateAlert(ctx, &models.Alert{ConfigID: "c1", AgentID: "a1", ConditionType: models.AlertConditionErrorRate, TriggeredAt: time.Now()})
	alertRepo.CreateAlert(ctx, &models.Alert{ConfigID: "c2", AgentID: "a1", ConditionType: models.AlertConditionErrorRate, TriggeredAt: time.Now()})
	resolved := &models.Alert{ConfigID: "c3", AgentID: "a1", ConditionType: models.AlertConditionErrorRate, TriggeredAt: time.Now()}
	alertRepo.CreateAlert(ctx, resolved)
	alertRepo.ResolveAlert(ctx, resolved.ID)

	alerts, _, err := svc.ListAlerts(ctx, "", true, 50, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 2 {
		t.Errorf("expected 2 active alerts, got %d", len(alerts))
	}
}

func TestResolveAlert_Success(t *testing.T) {
	svc := NewService(newMockRepo())
	alertRepo := newMockAlertRepo()
	svc.SetAlertRepository(alertRepo)
	ctx := context.Background()

	alert := &models.Alert{ConfigID: "c1", AgentID: "a1", ConditionType: models.AlertConditionDenialRate, TriggeredAt: time.Now()}
	alertRepo.CreateAlert(ctx, alert)

	if err := svc.ResolveAlert(ctx, alert.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := alertRepo.GetAlert(ctx, alert.ID)
	if !got.Resolved {
		t.Error("expected alert to be resolved")
	}
}

// --- AlertEngine tests ---

func TestAlertEngine_DenialRate(t *testing.T) {
	activityRepo := newMockRepo()
	alertRepo := newMockAlertRepo()
	ctx := context.Background()

	// Insert 10 actions: 6 denied, 4 allowed → 60% denial rate.
	for i := 0; i < 10; i++ {
		outcome := models.ActionOutcomeAllowed
		if i < 6 {
			outcome = models.ActionOutcomeDenied
		}
		activityRepo.InsertAction(ctx, &models.ActionRecord{
			WorkspaceID: "ws-1",
			AgentID:     "agent-1",
			ToolName:    "shell",
			Outcome:     outcome,
		})
	}

	// Configure alert: threshold 50%.
	alertRepo.UpsertAlertConfig(ctx, &models.AlertConfig{
		TenantID:      "tenant-1",
		Name:          "high-denial",
		ConditionType: models.AlertConditionDenialRate,
		Threshold:     0.5,
		AgentID:       "agent-1",
		Enabled:       true,
	})

	engine := NewAlertEngine(activityRepo, alertRepo, "tenant-1")
	engine.window = 1 * time.Hour // cover all test records
	engine.evaluate(ctx)

	alerts, _ := alertRepo.ListAlerts(ctx, "agent-1", true, "", 100)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].ConditionType != models.AlertConditionDenialRate {
		t.Errorf("expected denial_rate condition, got %q", alerts[0].ConditionType)
	}
}

func TestAlertEngine_NoTrigger(t *testing.T) {
	activityRepo := newMockRepo()
	alertRepo := newMockAlertRepo()
	ctx := context.Background()

	// Insert 10 actions: all allowed → 0% denial rate.
	for i := 0; i < 10; i++ {
		activityRepo.InsertAction(ctx, &models.ActionRecord{
			WorkspaceID: "ws-1",
			AgentID:     "agent-1",
			ToolName:    "shell",
			Outcome:     models.ActionOutcomeAllowed,
		})
	}

	alertRepo.UpsertAlertConfig(ctx, &models.AlertConfig{
		TenantID:      "tenant-1",
		Name:          "high-denial",
		ConditionType: models.AlertConditionDenialRate,
		Threshold:     0.5,
		AgentID:       "agent-1",
		Enabled:       true,
	})

	engine := NewAlertEngine(activityRepo, alertRepo, "tenant-1")
	engine.window = 1 * time.Hour
	engine.evaluate(ctx)

	alerts, _ := alertRepo.ListAlerts(ctx, "agent-1", true, "", 100)
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestAlertEngine_AutoResolve(t *testing.T) {
	activityRepo := newMockRepo()
	alertRepo := newMockAlertRepo()
	ctx := context.Background()

	// Initially insert denied actions to trigger alert.
	for i := 0; i < 10; i++ {
		activityRepo.InsertAction(ctx, &models.ActionRecord{
			WorkspaceID: "ws-1",
			AgentID:     "agent-1",
			ToolName:    "shell",
			Outcome:     models.ActionOutcomeDenied,
		})
	}

	cfg := &models.AlertConfig{
		TenantID:      "tenant-1",
		Name:          "high-denial",
		ConditionType: models.AlertConditionDenialRate,
		Threshold:     0.5,
		AgentID:       "agent-1",
		Enabled:       true,
	}
	alertRepo.UpsertAlertConfig(ctx, cfg)

	engine := NewAlertEngine(activityRepo, alertRepo, "tenant-1")
	engine.window = 1 * time.Hour
	engine.evaluate(ctx)

	alerts, _ := alertRepo.ListAlerts(ctx, "agent-1", true, "", 100)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert triggered, got %d", len(alerts))
	}
	alertID := alerts[0].ID

	// Now replace all records with allowed (simulate clearing by using new repo).
	activityRepo2 := newMockRepo()
	for i := 0; i < 10; i++ {
		activityRepo2.InsertAction(ctx, &models.ActionRecord{
			WorkspaceID: "ws-1",
			AgentID:     "agent-1",
			ToolName:    "shell",
			Outcome:     models.ActionOutcomeAllowed,
		})
	}

	engine2 := NewAlertEngine(activityRepo2, alertRepo, "tenant-1")
	engine2.window = 1 * time.Hour
	engine2.evaluate(ctx)

	got, _ := alertRepo.GetAlert(ctx, alertID)
	if !got.Resolved {
		t.Error("expected alert to be auto-resolved")
	}
}

func TestAlertEngine_StuckAgent(t *testing.T) {
	activityRepo := newMockRepo()
	alertRepo := newMockAlertRepo()
	ctx := context.Background()

	// Insert 5 consecutive errors on the same tool.
	for i := 0; i < 5; i++ {
		activityRepo.InsertAction(ctx, &models.ActionRecord{
			WorkspaceID: "ws-1",
			AgentID:     "agent-1",
			ToolName:    "database_query",
			Outcome:     models.ActionOutcomeError,
		})
	}

	alertRepo.UpsertAlertConfig(ctx, &models.AlertConfig{
		TenantID:      "tenant-1",
		Name:          "stuck-agent",
		ConditionType: models.AlertConditionStuckAgent,
		Threshold:     1, // threshold not used for stuck agent, but required > 0
		AgentID:       "agent-1",
		Enabled:       true,
	})

	engine := NewAlertEngine(activityRepo, alertRepo, "tenant-1")
	engine.window = 1 * time.Hour
	engine.evaluate(ctx)

	alerts, _ := alertRepo.ListAlerts(ctx, "agent-1", true, "", 100)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 stuck-agent alert, got %d", len(alerts))
	}
}

func TestAlertEngine_ErrorRate(t *testing.T) {
	activityRepo := newMockRepo()
	alertRepo := newMockAlertRepo()
	ctx := context.Background()

	// 7 errors out of 10 actions = 70%.
	for i := 0; i < 10; i++ {
		outcome := models.ActionOutcomeAllowed
		if i < 7 {
			outcome = models.ActionOutcomeError
		}
		activityRepo.InsertAction(ctx, &models.ActionRecord{
			WorkspaceID: "ws-1",
			AgentID:     "agent-1",
			ToolName:    "api_call",
			Outcome:     outcome,
		})
	}

	alertRepo.UpsertAlertConfig(ctx, &models.AlertConfig{
		TenantID:      "tenant-1",
		Name:          "high-error",
		ConditionType: models.AlertConditionErrorRate,
		Threshold:     0.5,
		AgentID:       "agent-1",
		Enabled:       true,
	})

	engine := NewAlertEngine(activityRepo, alertRepo, "tenant-1")
	engine.window = 1 * time.Hour
	engine.evaluate(ctx)

	alerts, _ := alertRepo.ListAlerts(ctx, "agent-1", true, "", 100)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 error-rate alert, got %d", len(alerts))
	}
}

func TestAlertEngine_DeduplicatesAlerts(t *testing.T) {
	activityRepo := newMockRepo()
	alertRepo := newMockAlertRepo()
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		activityRepo.InsertAction(ctx, &models.ActionRecord{
			WorkspaceID: "ws-1",
			AgentID:     "agent-1",
			ToolName:    "shell",
			Outcome:     models.ActionOutcomeDenied,
		})
	}

	alertRepo.UpsertAlertConfig(ctx, &models.AlertConfig{
		TenantID:      "tenant-1",
		Name:          "high-denial",
		ConditionType: models.AlertConditionDenialRate,
		Threshold:     0.5,
		AgentID:       "agent-1",
		Enabled:       true,
	})

	engine := NewAlertEngine(activityRepo, alertRepo, "tenant-1")
	engine.window = 1 * time.Hour

	// Evaluate multiple times — should still only have 1 alert.
	engine.evaluate(ctx)
	engine.evaluate(ctx)
	engine.evaluate(ctx)

	alerts, _ := alertRepo.ListAlerts(ctx, "agent-1", true, "", 100)
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert (deduplicated), got %d", len(alerts))
	}
}

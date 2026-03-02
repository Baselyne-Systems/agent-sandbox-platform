package activity

import (
	"context"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

// AlertRepository defines persistence operations for alert configs and alerts.
type AlertRepository interface {
	UpsertAlertConfig(ctx context.Context, config *models.AlertConfig) error
	GetAlertConfig(ctx context.Context, id string) (*models.AlertConfig, error)
	ListAlertConfigs(ctx context.Context, tenantID string, enabledOnly bool) ([]models.AlertConfig, error)
	CreateAlert(ctx context.Context, alert *models.Alert) error
	GetAlert(ctx context.Context, id string) (*models.Alert, error)
	ListAlerts(ctx context.Context, agentID string, activeOnly bool, afterID string, limit int) ([]models.Alert, error)
	ResolveAlert(ctx context.Context, id string) error
	// ActiveAlertForConfig returns the active (unresolved) alert for a config+agent, if any.
	ActiveAlertForConfig(ctx context.Context, configID, agentID string) (*models.Alert, error)
}

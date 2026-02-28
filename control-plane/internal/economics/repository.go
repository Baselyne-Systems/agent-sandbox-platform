package economics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// ResourceCost holds aggregated cost data for a single resource type.
type ResourceCost struct {
	ResourceType string
	TotalCost    float64
	RecordCount  int
}

// Repository defines data access for the economics service.
type Repository interface {
	InsertUsage(ctx context.Context, record *models.UsageRecord) error
	GetBudget(ctx context.Context, tenantID, agentID string) (*models.Budget, error)
	UpsertBudget(ctx context.Context, budget *models.Budget) error
	AddUsedAmount(ctx context.Context, tenantID, agentID string, amount float64) error
	GetCostReport(ctx context.Context, tenantID, agentID string, start, end time.Time) ([]ResourceCost, error)
}

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) InsertUsage(ctx context.Context, record *models.UsageRecord) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO usage_records (tenant_id, agent_id, workspace_id, resource_type, unit, quantity, cost)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, recorded_at`,
		record.TenantID, record.AgentID, record.WorkspaceID, record.ResourceType, record.Unit, record.Quantity, record.Cost,
	).Scan(&record.ID, &record.RecordedAt)
}

func (r *PostgresRepository) GetBudget(ctx context.Context, tenantID, agentID string) (*models.Budget, error) {
	b := &models.Budget{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, agent_id, currency, "limit", used, period_start, period_end,
		        COALESCE(on_exceeded, 'halt'), COALESCE(warning_threshold, 0)
		 FROM budgets WHERE agent_id = $1 AND tenant_id = $2`,
		agentID, tenantID,
	).Scan(&b.ID, &b.TenantID, &b.AgentID, &b.Currency, &b.Limit, &b.Used, &b.PeriodStart, &b.PeriodEnd,
		&b.OnExceeded, &b.WarningThreshold)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *PostgresRepository) UpsertBudget(ctx context.Context, budget *models.Budget) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO budgets (tenant_id, agent_id, currency, "limit", used, period_start, period_end, on_exceeded, warning_threshold)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (tenant_id, agent_id) DO UPDATE SET
		   currency = EXCLUDED.currency,
		   "limit" = EXCLUDED."limit",
		   used = EXCLUDED.used,
		   period_start = EXCLUDED.period_start,
		   period_end = EXCLUDED.period_end,
		   on_exceeded = EXCLUDED.on_exceeded,
		   warning_threshold = EXCLUDED.warning_threshold
		 RETURNING id, period_start, period_end`,
		budget.TenantID, budget.AgentID, budget.Currency, budget.Limit, budget.Used, budget.PeriodStart, budget.PeriodEnd,
		budget.OnExceeded, budget.WarningThreshold,
	).Scan(&budget.ID, &budget.PeriodStart, &budget.PeriodEnd)
}

func (r *PostgresRepository) AddUsedAmount(ctx context.Context, tenantID, agentID string, amount float64) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE budgets SET used = used + $1 WHERE agent_id = $2 AND tenant_id = $3`,
		amount, agentID, tenantID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBudgetNotFound
	}
	return nil
}

func (r *PostgresRepository) GetCostReport(ctx context.Context, tenantID, agentID string, start, end time.Time) ([]ResourceCost, error) {
	query := `SELECT resource_type, SUM(cost), COUNT(*)
		FROM usage_records
		WHERE tenant_id = $1 AND recorded_at >= $2 AND recorded_at <= $3`
	args := []any{tenantID, start, end}

	if agentID != "" {
		query += fmt.Sprintf(" AND agent_id = $%d", len(args)+1)
		args = append(args, agentID)
	}
	query += " GROUP BY resource_type ORDER BY resource_type"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var costs []ResourceCost
	for rows.Next() {
		var c ResourceCost
		if err := rows.Scan(&c.ResourceType, &c.TotalCost, &c.RecordCount); err != nil {
			return nil, err
		}
		costs = append(costs, c)
	}
	return costs, rows.Err()
}

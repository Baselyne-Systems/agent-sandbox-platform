package economics

import (
	"context"
	"database/sql"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
)

// Repository defines data access for the economics service.
type Repository interface {
	InsertUsage(ctx context.Context, record *models.UsageRecord) error
	GetBudget(ctx context.Context, agentID string) (*models.Budget, error)
	UpsertBudget(ctx context.Context, budget *models.Budget) error
	AddUsedAmount(ctx context.Context, agentID string, amount float64) error
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
		`INSERT INTO usage_records (agent_id, workspace_id, resource_type, unit, quantity, cost)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, recorded_at`,
		record.AgentID, record.WorkspaceID, record.ResourceType, record.Unit, record.Quantity, record.Cost,
	).Scan(&record.ID, &record.RecordedAt)
}

func (r *PostgresRepository) GetBudget(ctx context.Context, agentID string) (*models.Budget, error) {
	b := &models.Budget{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, currency, "limit", used, period_start, period_end
		 FROM budgets WHERE agent_id = $1`,
		agentID,
	).Scan(&b.ID, &b.AgentID, &b.Currency, &b.Limit, &b.Used, &b.PeriodStart, &b.PeriodEnd)
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
		`INSERT INTO budgets (agent_id, currency, "limit", used, period_start, period_end)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (agent_id) DO UPDATE SET
		   currency = EXCLUDED.currency,
		   "limit" = EXCLUDED."limit",
		   used = EXCLUDED.used,
		   period_start = EXCLUDED.period_start,
		   period_end = EXCLUDED.period_end
		 RETURNING id, period_start, period_end`,
		budget.AgentID, budget.Currency, budget.Limit, budget.Used, budget.PeriodStart, budget.PeriodEnd,
	).Scan(&budget.ID, &budget.PeriodStart, &budget.PeriodEnd)
}

func (r *PostgresRepository) AddUsedAmount(ctx context.Context, agentID string, amount float64) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE budgets SET used = used + $1 WHERE agent_id = $2`,
		amount, agentID,
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

package guardrails

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
)

// Repository defines persistence operations for guardrail rules.
type Repository interface {
	CreateRule(ctx context.Context, rule *models.GuardrailRule) error
	GetRule(ctx context.Context, id string) (*models.GuardrailRule, error)
	UpdateRule(ctx context.Context, rule *models.GuardrailRule) error
	DeleteRule(ctx context.Context, id string) error
	ListRules(ctx context.Context, ruleType models.RuleType, enabledOnly bool, afterID string, limit int) ([]models.GuardrailRule, error)
}

// PostgresRepository implements Repository backed by PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateRule(ctx context.Context, rule *models.GuardrailRule) error {
	labelsJSON, err := json.Marshal(rule.Labels)
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}
	return r.db.QueryRowContext(ctx,
		`INSERT INTO guardrail_rules (name, description, type, condition, action, priority, enabled, labels)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at, updated_at`,
		rule.Name, rule.Description, string(rule.Type), rule.Condition,
		string(rule.Action), rule.Priority, rule.Enabled, labelsJSON,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)
}

func (r *PostgresRepository) GetRule(ctx context.Context, id string) (*models.GuardrailRule, error) {
	var rule models.GuardrailRule
	var labelsJSON []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, type, condition, action, priority, enabled, labels, created_at, updated_at
		 FROM guardrail_rules WHERE id = $1`, id,
	).Scan(&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Condition,
		&rule.Action, &rule.Priority, &rule.Enabled, &labelsJSON, &rule.CreatedAt, &rule.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(labelsJSON, &rule.Labels); err != nil {
		return nil, fmt.Errorf("unmarshal labels: %w", err)
	}
	return &rule, nil
}

func (r *PostgresRepository) UpdateRule(ctx context.Context, rule *models.GuardrailRule) error {
	labelsJSON, err := json.Marshal(rule.Labels)
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}
	return r.db.QueryRowContext(ctx,
		`UPDATE guardrail_rules
		 SET name = $1, description = $2, type = $3, condition = $4, action = $5,
		     priority = $6, enabled = $7, labels = $8, updated_at = now()
		 WHERE id = $9
		 RETURNING updated_at`,
		rule.Name, rule.Description, string(rule.Type), rule.Condition,
		string(rule.Action), rule.Priority, rule.Enabled, labelsJSON, rule.ID,
	).Scan(&rule.UpdatedAt)
}

func (r *PostgresRepository) DeleteRule(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM guardrail_rules WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrRuleNotFound
	}
	return nil
}

func (r *PostgresRepository) ListRules(ctx context.Context, ruleType models.RuleType, enabledOnly bool, afterID string, limit int) ([]models.GuardrailRule, error) {
	query := `SELECT id, name, description, type, condition, action, priority, enabled, labels, created_at, updated_at
		FROM guardrail_rules WHERE 1=1`
	args := []any{}
	argIdx := 1

	if ruleType != "" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, string(ruleType))
		argIdx++
	}
	if enabledOnly {
		query += " AND enabled = true"
	}
	if afterID != "" {
		query += fmt.Sprintf(" AND id > $%d", argIdx)
		args = append(args, afterID)
		argIdx++
	}
	query += " ORDER BY id ASC"
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.GuardrailRule
	for rows.Next() {
		var rule models.GuardrailRule
		var labelsJSON []byte
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Description, &rule.Type, &rule.Condition,
			&rule.Action, &rule.Priority, &rule.Enabled, &labelsJSON, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(labelsJSON, &rule.Labels); err != nil {
			return nil, fmt.Errorf("unmarshal labels: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

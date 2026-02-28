package guardrails

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// Repository defines persistence operations for guardrail rules and sets.
type Repository interface {
	CreateRule(ctx context.Context, rule *models.GuardrailRule) error
	GetRule(ctx context.Context, tenantID, id string) (*models.GuardrailRule, error)
	UpdateRule(ctx context.Context, tenantID string, rule *models.GuardrailRule) error
	DeleteRule(ctx context.Context, tenantID, id string) error
	ListRules(ctx context.Context, tenantID string, ruleType models.RuleType, enabledOnly bool, afterID string, limit int) ([]models.GuardrailRule, error)

	CreateSet(ctx context.Context, set *models.GuardrailSet) error
	GetSet(ctx context.Context, tenantID, id string) (*models.GuardrailSet, error)
	GetSetByName(ctx context.Context, tenantID, name string) (*models.GuardrailSet, error)
	UpdateSet(ctx context.Context, tenantID string, set *models.GuardrailSet) error
	DeleteSet(ctx context.Context, tenantID, id string) error
	ListSets(ctx context.Context, tenantID string, afterID string, limit int) ([]models.GuardrailSet, error)
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
	scopeJSON, err := json.Marshal(rule.Scope)
	if err != nil {
		return fmt.Errorf("marshal scope: %w", err)
	}
	return r.db.QueryRowContext(ctx,
		`INSERT INTO guardrail_rules (tenant_id, name, description, type, condition, action, priority, enabled, labels, scope)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, created_at, updated_at`,
		rule.TenantID, rule.Name, rule.Description, string(rule.Type), rule.Condition,
		string(rule.Action), rule.Priority, rule.Enabled, labelsJSON, scopeJSON,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)
}

func (r *PostgresRepository) GetRule(ctx context.Context, tenantID, id string) (*models.GuardrailRule, error) {
	var rule models.GuardrailRule
	var labelsJSON, scopeJSON []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, name, description, type, condition, action, priority, enabled, labels, COALESCE(scope, '{}'), created_at, updated_at
		 FROM guardrail_rules WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	).Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.Description, &rule.Type, &rule.Condition,
		&rule.Action, &rule.Priority, &rule.Enabled, &labelsJSON, &scopeJSON, &rule.CreatedAt, &rule.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(labelsJSON, &rule.Labels); err != nil {
		return nil, fmt.Errorf("unmarshal labels: %w", err)
	}
	if err := json.Unmarshal(scopeJSON, &rule.Scope); err != nil {
		return nil, fmt.Errorf("unmarshal scope: %w", err)
	}
	return &rule, nil
}

func (r *PostgresRepository) UpdateRule(ctx context.Context, tenantID string, rule *models.GuardrailRule) error {
	labelsJSON, err := json.Marshal(rule.Labels)
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}
	scopeJSON, err := json.Marshal(rule.Scope)
	if err != nil {
		return fmt.Errorf("marshal scope: %w", err)
	}
	return r.db.QueryRowContext(ctx,
		`UPDATE guardrail_rules
		 SET name = $1, description = $2, type = $3, condition = $4, action = $5,
		     priority = $6, enabled = $7, labels = $8, scope = $9, updated_at = now()
		 WHERE id = $10 AND tenant_id = $11
		 RETURNING updated_at`,
		rule.Name, rule.Description, string(rule.Type), rule.Condition,
		string(rule.Action), rule.Priority, rule.Enabled, labelsJSON, scopeJSON, rule.ID, tenantID,
	).Scan(&rule.UpdatedAt)
}

func (r *PostgresRepository) DeleteRule(ctx context.Context, tenantID, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM guardrail_rules WHERE id = $1 AND tenant_id = $2`, id, tenantID)
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

func (r *PostgresRepository) ListRules(ctx context.Context, tenantID string, ruleType models.RuleType, enabledOnly bool, afterID string, limit int) ([]models.GuardrailRule, error) {
	query := `SELECT id, tenant_id, name, description, type, condition, action, priority, enabled, labels, COALESCE(scope, '{}'), created_at, updated_at
		FROM guardrail_rules WHERE tenant_id = $1`
	args := []any{tenantID}
	argIdx := 2

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
		var labelsJSON, scopeJSON []byte
		if err := rows.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.Description, &rule.Type, &rule.Condition,
			&rule.Action, &rule.Priority, &rule.Enabled, &labelsJSON, &scopeJSON, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(labelsJSON, &rule.Labels); err != nil {
			return nil, fmt.Errorf("unmarshal labels: %w", err)
		}
		if err := json.Unmarshal(scopeJSON, &rule.Scope); err != nil {
			return nil, fmt.Errorf("unmarshal scope: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *PostgresRepository) CreateSet(ctx context.Context, set *models.GuardrailSet) error {
	ruleIDsJSON, err := json.Marshal(set.RuleIDs)
	if err != nil {
		return fmt.Errorf("marshal rule_ids: %w", err)
	}
	labelsJSON, err := json.Marshal(set.Labels)
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}
	return r.db.QueryRowContext(ctx,
		`INSERT INTO guardrail_sets (tenant_id, name, description, rule_ids, labels)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		set.TenantID, set.Name, set.Description, ruleIDsJSON, labelsJSON,
	).Scan(&set.ID, &set.CreatedAt, &set.UpdatedAt)
}

func (r *PostgresRepository) GetSet(ctx context.Context, tenantID, id string) (*models.GuardrailSet, error) {
	return r.scanSet(ctx, `SELECT id, tenant_id, name, description, rule_ids, labels, created_at, updated_at FROM guardrail_sets WHERE id = $1 AND tenant_id = $2`, id, tenantID)
}

func (r *PostgresRepository) GetSetByName(ctx context.Context, tenantID, name string) (*models.GuardrailSet, error) {
	return r.scanSet(ctx, `SELECT id, tenant_id, name, description, rule_ids, labels, created_at, updated_at FROM guardrail_sets WHERE name = $1 AND tenant_id = $2`, name, tenantID)
}

func (r *PostgresRepository) scanSet(ctx context.Context, query string, args ...any) (*models.GuardrailSet, error) {
	var set models.GuardrailSet
	var ruleIDsJSON, labelsJSON []byte
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&set.ID, &set.TenantID, &set.Name, &set.Description, &ruleIDsJSON, &labelsJSON, &set.CreatedAt, &set.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(ruleIDsJSON, &set.RuleIDs); err != nil {
		return nil, fmt.Errorf("unmarshal rule_ids: %w", err)
	}
	if err := json.Unmarshal(labelsJSON, &set.Labels); err != nil {
		return nil, fmt.Errorf("unmarshal labels: %w", err)
	}
	return &set, nil
}

func (r *PostgresRepository) UpdateSet(ctx context.Context, tenantID string, set *models.GuardrailSet) error {
	ruleIDsJSON, err := json.Marshal(set.RuleIDs)
	if err != nil {
		return fmt.Errorf("marshal rule_ids: %w", err)
	}
	labelsJSON, err := json.Marshal(set.Labels)
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}
	return r.db.QueryRowContext(ctx,
		`UPDATE guardrail_sets
		 SET name = $1, description = $2, rule_ids = $3, labels = $4, updated_at = now()
		 WHERE id = $5 AND tenant_id = $6
		 RETURNING updated_at`,
		set.Name, set.Description, ruleIDsJSON, labelsJSON, set.ID, tenantID,
	).Scan(&set.UpdatedAt)
}

func (r *PostgresRepository) DeleteSet(ctx context.Context, tenantID, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM guardrail_sets WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrSetNotFound
	}
	return nil
}

func (r *PostgresRepository) ListSets(ctx context.Context, tenantID string, afterID string, limit int) ([]models.GuardrailSet, error) {
	query := `SELECT id, tenant_id, name, description, rule_ids, labels, created_at, updated_at FROM guardrail_sets WHERE tenant_id = $1`
	args := []any{tenantID}
	argIdx := 2

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

	var sets []models.GuardrailSet
	for rows.Next() {
		var set models.GuardrailSet
		var ruleIDsJSON, labelsJSON []byte
		if err := rows.Scan(&set.ID, &set.TenantID, &set.Name, &set.Description, &ruleIDsJSON, &labelsJSON, &set.CreatedAt, &set.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(ruleIDsJSON, &set.RuleIDs); err != nil {
			return nil, fmt.Errorf("unmarshal rule_ids: %w", err)
		}
		if err := json.Unmarshal(labelsJSON, &set.Labels); err != nil {
			return nil, fmt.Errorf("unmarshal labels: %w", err)
		}
		sets = append(sets, set)
	}
	return sets, rows.Err()
}

package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
)

// Repository defines persistence operations for tasks.
type Repository interface {
	CreateTask(ctx context.Context, task *models.Task) error
	GetTask(ctx context.Context, tenantID, id string) (*models.Task, error)
	ListTasks(ctx context.Context, tenantID string, agentID string, status models.TaskStatus, afterID string, limit int) ([]models.Task, error)
	UpdateTaskStatus(ctx context.Context, tenantID, id string, status models.TaskStatus) error
	SetWorkspaceID(ctx context.Context, tenantID, taskID, workspaceID string) error
}

// PostgresRepository implements Repository backed by PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateTask(ctx context.Context, task *models.Task) error {
	wsConfig, err := json.Marshal(task.WorkspaceConfig)
	if err != nil {
		return fmt.Errorf("marshal workspace_config: %w", err)
	}
	hiConfig, err := json.Marshal(task.HumanInteractionConfig)
	if err != nil {
		return fmt.Errorf("marshal human_interaction_config: %w", err)
	}
	budgetConfig, err := json.Marshal(task.BudgetConfig)
	if err != nil {
		return fmt.Errorf("marshal budget_config: %w", err)
	}
	inputJSON, err := json.Marshal(task.Input)
	if err != nil {
		return fmt.Errorf("marshal input: %w", err)
	}
	labelsJSON, err := json.Marshal(task.Labels)
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}

	return r.db.QueryRowContext(ctx,
		`INSERT INTO tasks (tenant_id, agent_id, goal, status, guardrail_policy_id, workspace_config,
		   human_interaction_config, budget_config, max_duration_without_checkin_secs,
		   input, labels)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, created_at, updated_at`,
		task.TenantID, task.AgentID, task.Goal, string(task.Status), task.GuardrailPolicyID,
		wsConfig, hiConfig, budgetConfig, task.MaxDurationWithoutCheckinSecs,
		inputJSON, labelsJSON,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *PostgresRepository) GetTask(ctx context.Context, tenantID, id string) (*models.Task, error) {
	var t models.Task
	var wsConfigJSON, hiConfigJSON, budgetConfigJSON, inputJSON, labelsJSON []byte
	var workspaceID sql.NullString
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, agent_id, goal, status, workspace_id, guardrail_policy_id,
			workspace_config, human_interaction_config, budget_config,
			max_duration_without_checkin_secs, input, labels,
			created_at, updated_at, completed_at
		 FROM tasks WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	).Scan(&t.ID, &t.TenantID, &t.AgentID, &t.Goal, &t.Status, &workspaceID, &t.GuardrailPolicyID,
		&wsConfigJSON, &hiConfigJSON, &budgetConfigJSON,
		&t.MaxDurationWithoutCheckinSecs, &inputJSON, &labelsJSON,
		&t.CreatedAt, &t.UpdatedAt, &completedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if workspaceID.Valid {
		t.WorkspaceID = workspaceID.String
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}

	if err := json.Unmarshal(wsConfigJSON, &t.WorkspaceConfig); err != nil {
		return nil, fmt.Errorf("unmarshal workspace_config: %w", err)
	}
	if err := json.Unmarshal(hiConfigJSON, &t.HumanInteractionConfig); err != nil {
		return nil, fmt.Errorf("unmarshal human_interaction_config: %w", err)
	}
	if err := json.Unmarshal(budgetConfigJSON, &t.BudgetConfig); err != nil {
		return nil, fmt.Errorf("unmarshal budget_config: %w", err)
	}
	if err := json.Unmarshal(inputJSON, &t.Input); err != nil {
		return nil, fmt.Errorf("unmarshal input: %w", err)
	}
	if err := json.Unmarshal(labelsJSON, &t.Labels); err != nil {
		return nil, fmt.Errorf("unmarshal labels: %w", err)
	}

	return &t, nil
}

func (r *PostgresRepository) ListTasks(ctx context.Context, tenantID string, agentID string, status models.TaskStatus, afterID string, limit int) ([]models.Task, error) {
	query := `SELECT id, tenant_id, agent_id, goal, status, workspace_id, guardrail_policy_id,
		workspace_config, human_interaction_config, budget_config,
		max_duration_without_checkin_secs, input, labels,
		created_at, updated_at, completed_at
		FROM tasks WHERE tenant_id = $1`
	args := []any{tenantID}
	argIdx := 2

	if agentID != "" {
		query += fmt.Sprintf(" AND agent_id = $%d", argIdx)
		args = append(args, agentID)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, string(status))
		argIdx++
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

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var wsConfigJSON, hiConfigJSON, budgetConfigJSON, inputJSON, labelsJSON []byte
		var workspaceID sql.NullString
		var completedAt sql.NullTime

		if err := rows.Scan(&t.ID, &t.TenantID, &t.AgentID, &t.Goal, &t.Status, &workspaceID, &t.GuardrailPolicyID,
			&wsConfigJSON, &hiConfigJSON, &budgetConfigJSON,
			&t.MaxDurationWithoutCheckinSecs, &inputJSON, &labelsJSON,
			&t.CreatedAt, &t.UpdatedAt, &completedAt); err != nil {
			return nil, err
		}

		if workspaceID.Valid {
			t.WorkspaceID = workspaceID.String
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}

		json.Unmarshal(wsConfigJSON, &t.WorkspaceConfig)
		json.Unmarshal(hiConfigJSON, &t.HumanInteractionConfig)
		json.Unmarshal(budgetConfigJSON, &t.BudgetConfig)
		json.Unmarshal(inputJSON, &t.Input)
		json.Unmarshal(labelsJSON, &t.Labels)

		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *PostgresRepository) UpdateTaskStatus(ctx context.Context, tenantID, id string, status models.TaskStatus) error {
	var completedAt interface{}
	if status == models.TaskStatusCompleted || status == models.TaskStatusFailed || status == models.TaskStatusCancelled {
		now := time.Now()
		completedAt = now
	}

	res, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET status = $1, updated_at = now(), completed_at = COALESCE($2, completed_at)
		 WHERE id = $3 AND tenant_id = $4`,
		string(status), completedAt, id, tenantID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *PostgresRepository) SetWorkspaceID(ctx context.Context, tenantID, taskID, workspaceID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET workspace_id = $1, updated_at = now() WHERE id = $2 AND tenant_id = $3`,
		workspaceID, taskID, tenantID)
	return err
}

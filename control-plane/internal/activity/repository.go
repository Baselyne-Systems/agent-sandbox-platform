package activity

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// Repository defines persistence operations for action records.
type Repository interface {
	InsertAction(ctx context.Context, record *models.ActionRecord) error
	GetAction(ctx context.Context, tenantID, id string) (*models.ActionRecord, error)
	QueryActions(ctx context.Context, tenantID string, filter QueryFilter) ([]models.ActionRecord, error)
}

// QueryFilter holds the optional filter fields for querying action records.
type QueryFilter struct {
	WorkspaceID string
	AgentID     string
	TaskID      string
	ToolName    string
	Outcome     models.ActionOutcome
	StartTime   *time.Time
	EndTime     *time.Time
	AfterID     string
	Limit       int
}

// PostgresRepository implements Repository backed by PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) InsertAction(ctx context.Context, record *models.ActionRecord) error {
	params := nullableJSON(record.Parameters)
	result := nullableJSON(record.Result)

	return r.db.QueryRowContext(ctx,
		`INSERT INTO action_records (tenant_id, workspace_id, agent_id, task_id, tool_name, parameters, result,
			outcome, guardrail_rule_id, denial_reason, evaluation_latency_us, execution_latency_us)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 RETURNING id, recorded_at`,
		record.TenantID, record.WorkspaceID, record.AgentID, nullIfEmpty(record.TaskID),
		record.ToolName, params, result,
		string(record.Outcome), nullIfEmpty(record.GuardrailRuleID), nullIfEmpty(record.DenialReason),
		record.EvaluationLatencyUs, record.ExecutionLatencyUs,
	).Scan(&record.ID, &record.RecordedAt)
}

func (r *PostgresRepository) GetAction(ctx context.Context, tenantID, id string) (*models.ActionRecord, error) {
	var rec models.ActionRecord
	var params, result []byte
	var taskID, guardrailRuleID, denialReason sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, workspace_id, agent_id, task_id, tool_name, parameters, result,
			outcome, guardrail_rule_id, denial_reason, evaluation_latency_us, execution_latency_us, recorded_at
		 FROM action_records WHERE id = $1 AND tenant_id = $2`, id, tenantID,
	).Scan(&rec.ID, &rec.TenantID, &rec.WorkspaceID, &rec.AgentID, &taskID,
		&rec.ToolName, &params, &result,
		&rec.Outcome, &guardrailRuleID, &denialReason,
		&rec.EvaluationLatencyUs, &rec.ExecutionLatencyUs, &rec.RecordedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rec.TaskID = taskID.String
	rec.GuardrailRuleID = guardrailRuleID.String
	rec.DenialReason = denialReason.String
	rec.Parameters = json.RawMessage(params)
	rec.Result = json.RawMessage(result)
	return &rec, nil
}

func (r *PostgresRepository) QueryActions(ctx context.Context, tenantID string, filter QueryFilter) ([]models.ActionRecord, error) {
	query := `SELECT id, tenant_id, workspace_id, agent_id, task_id, tool_name, parameters, result,
		outcome, guardrail_rule_id, denial_reason, evaluation_latency_us, execution_latency_us, recorded_at
		FROM action_records WHERE tenant_id = $1`
	args := []any{tenantID}
	argIdx := 2

	if filter.WorkspaceID != "" {
		query += fmt.Sprintf(" AND workspace_id = $%d", argIdx)
		args = append(args, filter.WorkspaceID)
		argIdx++
	}
	if filter.AgentID != "" {
		query += fmt.Sprintf(" AND agent_id = $%d", argIdx)
		args = append(args, filter.AgentID)
		argIdx++
	}
	if filter.TaskID != "" {
		query += fmt.Sprintf(" AND task_id = $%d", argIdx)
		args = append(args, filter.TaskID)
		argIdx++
	}
	if filter.ToolName != "" {
		query += fmt.Sprintf(" AND tool_name = $%d", argIdx)
		args = append(args, filter.ToolName)
		argIdx++
	}
	if filter.Outcome != "" {
		query += fmt.Sprintf(" AND outcome = $%d", argIdx)
		args = append(args, string(filter.Outcome))
		argIdx++
	}
	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND recorded_at >= $%d", argIdx)
		args = append(args, *filter.StartTime)
		argIdx++
	}
	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND recorded_at <= $%d", argIdx)
		args = append(args, *filter.EndTime)
		argIdx++
	}
	if filter.AfterID != "" {
		query += fmt.Sprintf(" AND id > $%d", argIdx)
		args = append(args, filter.AfterID)
		argIdx++
	}

	query += " ORDER BY id ASC"
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, filter.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.ActionRecord
	for rows.Next() {
		var rec models.ActionRecord
		var params, result []byte
		var taskID, guardrailRuleID, denialReason sql.NullString
		if err := rows.Scan(&rec.ID, &rec.TenantID, &rec.WorkspaceID, &rec.AgentID, &taskID,
			&rec.ToolName, &params, &result,
			&rec.Outcome, &guardrailRuleID, &denialReason,
			&rec.EvaluationLatencyUs, &rec.ExecutionLatencyUs, &rec.RecordedAt); err != nil {
			return nil, err
		}
		rec.TaskID = taskID.String
		rec.GuardrailRuleID = guardrailRuleID.String
		rec.DenialReason = denialReason.String
		rec.Parameters = json.RawMessage(params)
		rec.Result = json.RawMessage(result)
		records = append(records, rec)
	}
	return records, rows.Err()
}

// nullIfEmpty converts an empty string to sql.NullString{Valid: false}.
func nullIfEmpty(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullableJSON returns nil if the raw message is nil or empty, otherwise the raw bytes.
func nullableJSON(data json.RawMessage) any {
	if len(data) == 0 {
		return nil
	}
	return []byte(data)
}

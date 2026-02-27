package human

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// Repository defines persistence operations for human interaction requests.
type Repository interface {
	CreateRequest(ctx context.Context, req *models.HumanRequest) error
	GetRequest(ctx context.Context, id string) (*models.HumanRequest, error)
	RespondToRequest(ctx context.Context, id, response, responderID string) error
	ListRequests(ctx context.Context, workspaceID string, status models.HumanRequestStatus, afterID string, limit int) ([]models.HumanRequest, error)

	UpsertDeliveryChannel(ctx context.Context, cfg *models.DeliveryChannelConfig) error
	GetDeliveryChannel(ctx context.Context, userID, channelType string) (*models.DeliveryChannelConfig, error)
	UpsertTimeoutPolicy(ctx context.Context, policy *models.TimeoutPolicy) error
	GetTimeoutPolicy(ctx context.Context, scope, scopeID string) (*models.TimeoutPolicy, error)

	// ExpirePendingRequests marks all pending requests past their deadline as expired.
	ExpirePendingRequests(ctx context.Context) (int, error)
}

// PostgresRepository implements Repository backed by PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateRequest(ctx context.Context, req *models.HumanRequest) error {
	optionsJSON, err := json.Marshal(req.Options)
	if err != nil {
		return fmt.Errorf("marshal options: %w", err)
	}
	var taskID interface{}
	if req.TaskID != "" {
		taskID = req.TaskID
	}
	return r.db.QueryRowContext(ctx,
		`INSERT INTO human_requests (workspace_id, agent_id, question, options, context, status, expires_at, type, urgency, task_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, created_at`,
		req.WorkspaceID, req.AgentID, req.Question, optionsJSON,
		req.Context, string(req.Status), req.ExpiresAt,
		string(req.Type), string(req.Urgency), taskID,
	).Scan(&req.ID, &req.CreatedAt)
}

func (r *PostgresRepository) GetRequest(ctx context.Context, id string) (*models.HumanRequest, error) {
	var req models.HumanRequest
	var optionsJSON []byte
	var response, responderID sql.NullString
	var respondedAt sql.NullTime
	var expiresAt sql.NullTime

	var taskID sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, agent_id, question, options, context, status,
			response, responder_id, created_at, responded_at, expires_at,
			type, urgency, task_id
		 FROM human_requests WHERE id = $1`, id,
	).Scan(&req.ID, &req.WorkspaceID, &req.AgentID, &req.Question, &optionsJSON,
		&req.Context, &req.Status, &response, &responderID,
		&req.CreatedAt, &respondedAt, &expiresAt,
		&req.Type, &req.Urgency, &taskID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	req.Response = response.String
	req.ResponderID = responderID.String
	if respondedAt.Valid {
		req.RespondedAt = &respondedAt.Time
	}
	if expiresAt.Valid {
		req.ExpiresAt = &expiresAt.Time
	}
	if taskID.Valid {
		req.TaskID = taskID.String
	}

	if err := json.Unmarshal(optionsJSON, &req.Options); err != nil {
		return nil, fmt.Errorf("unmarshal options: %w", err)
	}
	return &req, nil
}

func (r *PostgresRepository) RespondToRequest(ctx context.Context, id, response, responderID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE human_requests
		 SET status = $1, response = $2, responder_id = $3, responded_at = $4
		 WHERE id = $5 AND status = 'pending'`,
		string(models.HumanRequestStatusResponded), response, responderID, time.Now(), id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrRequestNotPending
	}
	return nil
}

func (r *PostgresRepository) ListRequests(ctx context.Context, workspaceID string, status models.HumanRequestStatus, afterID string, limit int) ([]models.HumanRequest, error) {
	query := `SELECT id, workspace_id, agent_id, question, options, context, status,
		response, responder_id, created_at, responded_at, expires_at,
		type, urgency, task_id
		FROM human_requests WHERE 1=1`
	args := []any{}
	argIdx := 1

	if workspaceID != "" {
		query += fmt.Sprintf(" AND workspace_id = $%d", argIdx)
		args = append(args, workspaceID)
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

	var requests []models.HumanRequest
	for rows.Next() {
		var req models.HumanRequest
		var optionsJSON []byte
		var response, responderID, taskID sql.NullString
		var respondedAt, expiresAt sql.NullTime

		if err := rows.Scan(&req.ID, &req.WorkspaceID, &req.AgentID, &req.Question,
			&optionsJSON, &req.Context, &req.Status, &response, &responderID,
			&req.CreatedAt, &respondedAt, &expiresAt,
			&req.Type, &req.Urgency, &taskID); err != nil {
			return nil, err
		}
		req.Response = response.String
		req.ResponderID = responderID.String
		if respondedAt.Valid {
			req.RespondedAt = &respondedAt.Time
		}
		if expiresAt.Valid {
			req.ExpiresAt = &expiresAt.Time
		}
		if taskID.Valid {
			req.TaskID = taskID.String
		}
		if err := json.Unmarshal(optionsJSON, &req.Options); err != nil {
			return nil, fmt.Errorf("unmarshal options: %w", err)
		}
		requests = append(requests, req)
	}
	return requests, rows.Err()
}

func (r *PostgresRepository) UpsertDeliveryChannel(ctx context.Context, cfg *models.DeliveryChannelConfig) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO delivery_channels (user_id, channel_type, endpoint, enabled)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, channel_type) DO UPDATE SET endpoint = $3, enabled = $4, updated_at = now()
		 RETURNING id, created_at, updated_at`,
		cfg.UserID, cfg.ChannelType, cfg.Endpoint, cfg.Enabled,
	).Scan(&cfg.ID, &cfg.CreatedAt, &cfg.UpdatedAt)
}

func (r *PostgresRepository) GetDeliveryChannel(ctx context.Context, userID, channelType string) (*models.DeliveryChannelConfig, error) {
	var cfg models.DeliveryChannelConfig
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, channel_type, endpoint, enabled, created_at, updated_at
		 FROM delivery_channels WHERE user_id = $1 AND channel_type = $2`,
		userID, channelType,
	).Scan(&cfg.ID, &cfg.UserID, &cfg.ChannelType, &cfg.Endpoint, &cfg.Enabled, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *PostgresRepository) UpsertTimeoutPolicy(ctx context.Context, policy *models.TimeoutPolicy) error {
	targetsJSON, err := json.Marshal(policy.EscalationTargets)
	if err != nil {
		return fmt.Errorf("marshal escalation_targets: %w", err)
	}
	return r.db.QueryRowContext(ctx,
		`INSERT INTO timeout_policies (scope, scope_id, timeout_secs, action, escalation_targets)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (scope, scope_id) DO UPDATE SET timeout_secs = $3, action = $4, escalation_targets = $5, updated_at = now()
		 RETURNING id, created_at, updated_at`,
		policy.Scope, policy.ScopeID, policy.TimeoutSecs, policy.Action, targetsJSON,
	).Scan(&policy.ID, &policy.CreatedAt, &policy.UpdatedAt)
}

func (r *PostgresRepository) GetTimeoutPolicy(ctx context.Context, scope, scopeID string) (*models.TimeoutPolicy, error) {
	var policy models.TimeoutPolicy
	var targetsJSON []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT id, scope, scope_id, timeout_secs, action, escalation_targets, created_at, updated_at
		 FROM timeout_policies WHERE scope = $1 AND scope_id = $2`,
		scope, scopeID,
	).Scan(&policy.ID, &policy.Scope, &policy.ScopeID, &policy.TimeoutSecs, &policy.Action, &targetsJSON, &policy.CreatedAt, &policy.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(targetsJSON, &policy.EscalationTargets); err != nil {
		return nil, fmt.Errorf("unmarshal escalation_targets: %w", err)
	}
	return &policy, nil
}

func (r *PostgresRepository) ExpirePendingRequests(ctx context.Context) (int, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE human_requests SET status = 'expired'
		 WHERE status = 'pending' AND expires_at IS NOT NULL AND expires_at < now()`)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

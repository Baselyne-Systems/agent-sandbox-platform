package workspace

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// Repository defines data access for the workspace service.
type Repository interface {
	CreateWorkspace(ctx context.Context, ws *models.Workspace) error
	GetWorkspace(ctx context.Context, id string) (*models.Workspace, error)
	ListWorkspaces(ctx context.Context, agentID string, status models.WorkspaceStatus, afterID string, limit int) ([]models.Workspace, error)
	UpdateWorkspaceStatus(ctx context.Context, id string, status models.WorkspaceStatus, hostID, hostAddress, sandboxID string) error
	TerminateWorkspace(ctx context.Context, id string, reason string) error
	SetSnapshotID(ctx context.Context, workspaceID, snapshotID string) error
	CreateSnapshot(ctx context.Context, snapshot *models.WorkspaceSnapshot) error
	GetSnapshot(ctx context.Context, snapshotID string) (*models.WorkspaceSnapshot, error)
}

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateWorkspace(ctx context.Context, ws *models.Workspace) error {
	allowedToolsJSON, err := json.Marshal(ws.Spec.AllowedTools)
	if err != nil {
		return fmt.Errorf("marshal allowed_tools: %w", err)
	}
	envVarsJSON, err := json.Marshal(ws.Spec.EnvVars)
	if err != nil {
		return fmt.Errorf("marshal env_vars: %w", err)
	}
	return r.db.QueryRowContext(ctx,
		`INSERT INTO workspaces (agent_id, task_id, status, memory_mb, cpu_millicores, disk_mb,
		   max_duration_secs, allowed_tools, guardrail_policy_id, env_vars, host_id, host_address, sandbox_id, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		 RETURNING id, created_at, updated_at`,
		ws.AgentID, ws.TaskID, string(ws.Status),
		ws.Spec.MemoryMb, ws.Spec.CpuMillicores, ws.Spec.DiskMb,
		ws.Spec.MaxDurationSecs, allowedToolsJSON, ws.Spec.GuardrailPolicyID, envVarsJSON,
		ws.HostID, ws.HostAddress, ws.SandboxID, ws.ExpiresAt,
	).Scan(&ws.ID, &ws.CreatedAt, &ws.UpdatedAt)
}

func (r *PostgresRepository) GetWorkspace(ctx context.Context, id string) (*models.Workspace, error) {
	var ws models.Workspace
	var allowedToolsJSON, envVarsJSON []byte
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, agent_id, task_id, status, memory_mb, cpu_millicores, disk_mb,
			max_duration_secs, allowed_tools, guardrail_policy_id, env_vars,
			host_id, host_address, sandbox_id, created_at, updated_at, expires_at
		 FROM workspaces WHERE id = $1`, id,
	).Scan(&ws.ID, &ws.AgentID, &ws.TaskID, &ws.Status,
		&ws.Spec.MemoryMb, &ws.Spec.CpuMillicores, &ws.Spec.DiskMb,
		&ws.Spec.MaxDurationSecs, &allowedToolsJSON, &ws.Spec.GuardrailPolicyID, &envVarsJSON,
		&ws.HostID, &ws.HostAddress, &ws.SandboxID, &ws.CreatedAt, &ws.UpdatedAt, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		ws.ExpiresAt = &expiresAt.Time
	}
	if err := json.Unmarshal(allowedToolsJSON, &ws.Spec.AllowedTools); err != nil {
		return nil, fmt.Errorf("unmarshal allowed_tools: %w", err)
	}
	if err := json.Unmarshal(envVarsJSON, &ws.Spec.EnvVars); err != nil {
		return nil, fmt.Errorf("unmarshal env_vars: %w", err)
	}
	return &ws, nil
}

func (r *PostgresRepository) ListWorkspaces(ctx context.Context, agentID string, status models.WorkspaceStatus, afterID string, limit int) ([]models.Workspace, error) {
	query := `SELECT id, agent_id, task_id, status, memory_mb, cpu_millicores, disk_mb,
		max_duration_secs, allowed_tools, guardrail_policy_id, env_vars,
		host_id, host_address, sandbox_id, created_at, updated_at, expires_at
		FROM workspaces WHERE 1=1`
	args := []any{}
	argIdx := 1

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

	var workspaces []models.Workspace
	for rows.Next() {
		var ws models.Workspace
		var allowedToolsJSON, envVarsJSON []byte
		var expiresAt sql.NullTime

		if err := rows.Scan(&ws.ID, &ws.AgentID, &ws.TaskID, &ws.Status,
			&ws.Spec.MemoryMb, &ws.Spec.CpuMillicores, &ws.Spec.DiskMb,
			&ws.Spec.MaxDurationSecs, &allowedToolsJSON, &ws.Spec.GuardrailPolicyID, &envVarsJSON,
			&ws.HostID, &ws.HostAddress, &ws.SandboxID, &ws.CreatedAt, &ws.UpdatedAt, &expiresAt); err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			ws.ExpiresAt = &expiresAt.Time
		}
		if err := json.Unmarshal(allowedToolsJSON, &ws.Spec.AllowedTools); err != nil {
			return nil, fmt.Errorf("unmarshal allowed_tools: %w", err)
		}
		if err := json.Unmarshal(envVarsJSON, &ws.Spec.EnvVars); err != nil {
			return nil, fmt.Errorf("unmarshal env_vars: %w", err)
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces, rows.Err()
}

func (r *PostgresRepository) UpdateWorkspaceStatus(ctx context.Context, id string, status models.WorkspaceStatus, hostID, hostAddress, sandboxID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE workspaces SET status = $1, host_id = $2, host_address = $3, sandbox_id = $4, updated_at = $5
		 WHERE id = $6`,
		string(status), hostID, hostAddress, sandboxID, time.Now(), id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrWorkspaceNotFound
	}
	return nil
}

func (r *PostgresRepository) SetSnapshotID(ctx context.Context, workspaceID, snapshotID string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE workspaces SET snapshot_id = $1, updated_at = $2 WHERE id = $3`,
		snapshotID, time.Now(), workspaceID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrWorkspaceNotFound
	}
	return nil
}

func (r *PostgresRepository) CreateSnapshot(ctx context.Context, snapshot *models.WorkspaceSnapshot) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO workspace_snapshots (workspace_id, agent_id, task_id, size_bytes)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		snapshot.WorkspaceID, snapshot.AgentID, snapshot.TaskID, snapshot.SizeBytes,
	).Scan(&snapshot.ID, &snapshot.CreatedAt)
}

func (r *PostgresRepository) GetSnapshot(ctx context.Context, snapshotID string) (*models.WorkspaceSnapshot, error) {
	var s models.WorkspaceSnapshot
	err := r.db.QueryRowContext(ctx,
		`SELECT id, workspace_id, agent_id, task_id, size_bytes, created_at
		 FROM workspace_snapshots WHERE id = $1`, snapshotID,
	).Scan(&s.ID, &s.WorkspaceID, &s.AgentID, &s.TaskID, &s.SizeBytes, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepository) TerminateWorkspace(ctx context.Context, id string, reason string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE workspaces SET status = 'terminated', updated_at = $1
		 WHERE id = $2 AND status NOT IN ('terminated', 'failed')`,
		time.Now(), id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		// Check if the workspace exists at all.
		var exists bool
		err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = $1)`, id).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return ErrWorkspaceNotFound
		}
		return ErrWorkspaceAlreadyTerminal
	}
	return nil
}

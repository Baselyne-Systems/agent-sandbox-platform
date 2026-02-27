package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// Repository defines persistence operations for agent identities and credentials.
type Repository interface {
	CreateAgent(ctx context.Context, agent *models.Agent) error
	GetAgent(ctx context.Context, id string) (*models.Agent, error)
	ListAgents(ctx context.Context, ownerID string, status models.AgentStatus, afterID string, limit int) ([]models.Agent, error)
	DeactivateAgent(ctx context.Context, id string) error

	CreateCredential(ctx context.Context, cred *models.ScopedCredential) error
	RevokeCredential(ctx context.Context, id string) error
}

// PostgresRepository implements Repository backed by PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateAgent(ctx context.Context, agent *models.Agent) error {
	labelsJSON, err := json.Marshal(agent.Labels)
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}
	capsJSON, err := json.Marshal(agent.Capabilities)
	if err != nil {
		return fmt.Errorf("marshal capabilities: %w", err)
	}
	return r.db.QueryRowContext(ctx,
		`INSERT INTO agents (name, description, owner_id, status, labels, purpose, trust_level, capabilities)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at, updated_at`,
		agent.Name, agent.Description, agent.OwnerID, agent.Status, labelsJSON,
		agent.Purpose, string(agent.TrustLevel), capsJSON,
	).Scan(&agent.ID, &agent.CreatedAt, &agent.UpdatedAt)
}

func (r *PostgresRepository) GetAgent(ctx context.Context, id string) (*models.Agent, error) {
	var a models.Agent
	var labelsJSON, capsJSON []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, owner_id, status, labels, purpose, trust_level, capabilities, created_at, updated_at
		 FROM agents WHERE id = $1`, id,
	).Scan(&a.ID, &a.Name, &a.Description, &a.OwnerID, &a.Status, &labelsJSON,
		&a.Purpose, &a.TrustLevel, &capsJSON, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(labelsJSON, &a.Labels); err != nil {
		return nil, fmt.Errorf("unmarshal labels: %w", err)
	}
	if err := json.Unmarshal(capsJSON, &a.Capabilities); err != nil {
		return nil, fmt.Errorf("unmarshal capabilities: %w", err)
	}
	return &a, nil
}

func (r *PostgresRepository) ListAgents(ctx context.Context, ownerID string, status models.AgentStatus, afterID string, limit int) ([]models.Agent, error) {
	query := `SELECT id, name, description, owner_id, status, labels, purpose, trust_level, capabilities, created_at, updated_at FROM agents WHERE 1=1`
	args := []any{}
	argIdx := 1

	if ownerID != "" {
		query += fmt.Sprintf(" AND owner_id = $%d", argIdx)
		args = append(args, ownerID)
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

	var agents []models.Agent
	for rows.Next() {
		var a models.Agent
		var labelsJSON, capsJSON []byte
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.OwnerID, &a.Status, &labelsJSON,
			&a.Purpose, &a.TrustLevel, &capsJSON, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(labelsJSON, &a.Labels); err != nil {
			return nil, fmt.Errorf("unmarshal labels: %w", err)
		}
		if err := json.Unmarshal(capsJSON, &a.Capabilities); err != nil {
			return nil, fmt.Errorf("unmarshal capabilities: %w", err)
		}
		agents = append(agents, a)
	}
	return agents, rows.Err()
}

func (r *PostgresRepository) DeactivateAgent(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE agents SET status = 'inactive', updated_at = now() WHERE id = $1 AND status != 'inactive'`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		// Check if agent exists at all
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM agents WHERE id = $1)`, id).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return ErrAgentNotFound
		}
		// Already inactive — commit is still fine
	}

	// Revoke all active credentials atomically
	if _, err := tx.ExecContext(ctx,
		`UPDATE scoped_credentials SET revoked = true WHERE agent_id = $1 AND revoked = false`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) CreateCredential(ctx context.Context, cred *models.ScopedCredential) error {
	scopesJSON, err := json.Marshal(cred.Scopes)
	if err != nil {
		return fmt.Errorf("marshal scopes: %w", err)
	}
	return r.db.QueryRowContext(ctx,
		`INSERT INTO scoped_credentials (agent_id, scopes, token_hash, expires_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		cred.AgentID, scopesJSON, cred.TokenHash, cred.ExpiresAt,
	).Scan(&cred.ID, &cred.CreatedAt)
}

func (r *PostgresRepository) RevokeCredential(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE scoped_credentials SET revoked = true WHERE id = $1 AND revoked = false`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrCredentialNotFound
	}
	return nil
}

package compute

import (
	"context"
	"database/sql"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// Repository defines data access for the compute plane service.
type Repository interface {
	CreateHost(ctx context.Context, host *models.Host) error
	GetHost(ctx context.Context, id string) (*models.Host, error)
	ListHosts(ctx context.Context, status models.HostStatus) ([]models.Host, error)
	SetHostStatus(ctx context.Context, id string, status models.HostStatus) error
	FindHostForPlacement(ctx context.Context, memoryMb int64, cpuMillicores int32, diskMb int64) (*models.Host, error)
	DecrementAvailableResources(ctx context.Context, hostID string, memoryMb int64, cpuMillicores int32, diskMb int64) error
}

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateHost(ctx context.Context, host *models.Host) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO hosts (address, status, total_memory_mb, total_cpu_millicores, total_disk_mb,
		   available_memory_mb, available_cpu_millicores, available_disk_mb, active_sandboxes, last_heartbeat)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id`,
		host.Address, string(host.Status),
		host.TotalResources.MemoryMb, host.TotalResources.CpuMillicores, host.TotalResources.DiskMb,
		host.AvailableResources.MemoryMb, host.AvailableResources.CpuMillicores, host.AvailableResources.DiskMb,
		host.ActiveSandboxes, host.LastHeartbeat,
	).Scan(&host.ID)
}

func (r *PostgresRepository) GetHost(ctx context.Context, id string) (*models.Host, error) {
	var h models.Host
	err := r.db.QueryRowContext(ctx,
		`SELECT id, address, status, total_memory_mb, total_cpu_millicores, total_disk_mb,
			available_memory_mb, available_cpu_millicores, available_disk_mb,
			active_sandboxes, last_heartbeat
		 FROM hosts WHERE id = $1`, id,
	).Scan(&h.ID, &h.Address, &h.Status,
		&h.TotalResources.MemoryMb, &h.TotalResources.CpuMillicores, &h.TotalResources.DiskMb,
		&h.AvailableResources.MemoryMb, &h.AvailableResources.CpuMillicores, &h.AvailableResources.DiskMb,
		&h.ActiveSandboxes, &h.LastHeartbeat)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *PostgresRepository) ListHosts(ctx context.Context, status models.HostStatus) ([]models.Host, error) {
	query := `SELECT id, address, status, total_memory_mb, total_cpu_millicores, total_disk_mb,
		available_memory_mb, available_cpu_millicores, available_disk_mb,
		active_sandboxes, last_heartbeat
		FROM hosts`
	args := []any{}

	if status != "" {
		query += " WHERE status = $1"
		args = append(args, string(status))
	}
	query += " ORDER BY id ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []models.Host
	for rows.Next() {
		var h models.Host
		if err := rows.Scan(&h.ID, &h.Address, &h.Status,
			&h.TotalResources.MemoryMb, &h.TotalResources.CpuMillicores, &h.TotalResources.DiskMb,
			&h.AvailableResources.MemoryMb, &h.AvailableResources.CpuMillicores, &h.AvailableResources.DiskMb,
			&h.ActiveSandboxes, &h.LastHeartbeat); err != nil {
			return nil, err
		}
		hosts = append(hosts, h)
	}
	return hosts, rows.Err()
}

func (r *PostgresRepository) SetHostStatus(ctx context.Context, id string, status models.HostStatus) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE hosts SET status = $1 WHERE id = $2`,
		string(status), id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrHostNotFound
	}
	return nil
}

func (r *PostgresRepository) FindHostForPlacement(ctx context.Context, memoryMb int64, cpuMillicores int32, diskMb int64) (*models.Host, error) {
	var h models.Host
	err := r.db.QueryRowContext(ctx,
		`SELECT id, address, status, total_memory_mb, total_cpu_millicores, total_disk_mb,
			available_memory_mb, available_cpu_millicores, available_disk_mb,
			active_sandboxes, last_heartbeat
		 FROM hosts
		 WHERE status = 'ready'
		   AND available_memory_mb >= $1
		   AND available_cpu_millicores >= $2
		   AND available_disk_mb >= $3
		 ORDER BY available_memory_mb ASC
		 LIMIT 1`,
		memoryMb, cpuMillicores, diskMb,
	).Scan(&h.ID, &h.Address, &h.Status,
		&h.TotalResources.MemoryMb, &h.TotalResources.CpuMillicores, &h.TotalResources.DiskMb,
		&h.AvailableResources.MemoryMb, &h.AvailableResources.CpuMillicores, &h.AvailableResources.DiskMb,
		&h.ActiveSandboxes, &h.LastHeartbeat)
	if err == sql.ErrNoRows {
		return nil, ErrNoCapacity
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *PostgresRepository) DecrementAvailableResources(ctx context.Context, hostID string, memoryMb int64, cpuMillicores int32, diskMb int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE hosts SET
		   available_memory_mb = available_memory_mb - $1,
		   available_cpu_millicores = available_cpu_millicores - $2,
		   available_disk_mb = available_disk_mb - $3,
		   active_sandboxes = active_sandboxes + 1,
		   last_heartbeat = $4
		 WHERE id = $5`,
		memoryMb, cpuMillicores, diskMb, time.Now(), hostID)
	return err
}

package compute

import (
	"context"
	"database/sql"
	"time"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
	"github.com/lib/pq"
)

// Repository defines data access for the compute plane service.
type Repository interface {
	CreateHost(ctx context.Context, host *models.Host) error
	GetHost(ctx context.Context, id string) (*models.Host, error)
	ListHosts(ctx context.Context, status models.HostStatus) ([]models.Host, error)
	SetHostStatus(ctx context.Context, id string, status models.HostStatus) error
	PlaceAndDecrement(ctx context.Context, memoryMb int64, cpuMillicores int32, diskMb int64, isolationTier string) (*models.Host, error)
	UpdateHeartbeat(ctx context.Context, hostID string, resources models.HostResources, activeSandboxes int32, supportedTiers []string) (*models.Host, error)
	MarkStaleHostsOffline(ctx context.Context, timeout time.Duration) (int64, error)

	// Warm pool operations.
	UpsertWarmPoolConfig(ctx context.Context, cfg *models.WarmPoolConfig) error
	ListWarmPoolConfigs(ctx context.Context) ([]models.WarmPoolConfig, error)
	ClaimWarmSlot(ctx context.Context, tier string) (*models.WarmPoolSlot, error)
	CreateWarmSlot(ctx context.Context, slot *models.WarmPoolSlot) error
	CountReadySlots(ctx context.Context, tier string) (int32, error)
	CleanExpiredSlots(ctx context.Context) (int64, error)
	GetCapacity(ctx context.Context) ([]models.TierCapacity, int32, int32, error)
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
		   available_memory_mb, available_cpu_millicores, available_disk_mb, active_sandboxes, last_heartbeat, supported_tiers)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id`,
		host.Address, string(host.Status),
		host.TotalResources.MemoryMb, host.TotalResources.CpuMillicores, host.TotalResources.DiskMb,
		host.AvailableResources.MemoryMb, host.AvailableResources.CpuMillicores, host.AvailableResources.DiskMb,
		host.ActiveSandboxes, host.LastHeartbeat, pq.Array(host.SupportedTiers),
	).Scan(&host.ID)
}

func (r *PostgresRepository) GetHost(ctx context.Context, id string) (*models.Host, error) {
	var h models.Host
	err := r.db.QueryRowContext(ctx,
		`SELECT id, address, status, total_memory_mb, total_cpu_millicores, total_disk_mb,
			available_memory_mb, available_cpu_millicores, available_disk_mb,
			active_sandboxes, last_heartbeat, supported_tiers
		 FROM hosts WHERE id = $1`, id,
	).Scan(&h.ID, &h.Address, &h.Status,
		&h.TotalResources.MemoryMb, &h.TotalResources.CpuMillicores, &h.TotalResources.DiskMb,
		&h.AvailableResources.MemoryMb, &h.AvailableResources.CpuMillicores, &h.AvailableResources.DiskMb,
		&h.ActiveSandboxes, &h.LastHeartbeat, pq.Array(&h.SupportedTiers))
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
		active_sandboxes, last_heartbeat, supported_tiers
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
			&h.ActiveSandboxes, &h.LastHeartbeat, pq.Array(&h.SupportedTiers)); err != nil {
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

func (r *PostgresRepository) PlaceAndDecrement(ctx context.Context, memoryMb int64, cpuMillicores int32, diskMb int64, isolationTier string) (*models.Host, error) {
	var h models.Host
	err := r.db.QueryRowContext(ctx,
		`UPDATE hosts SET
		   available_memory_mb = available_memory_mb - $1,
		   available_cpu_millicores = available_cpu_millicores - $2,
		   available_disk_mb = available_disk_mb - $3,
		   active_sandboxes = active_sandboxes + 1,
		   last_heartbeat = $4
		 WHERE id = (
		   SELECT id FROM hosts
		   WHERE status = 'ready'
		     AND available_memory_mb >= $1
		     AND available_cpu_millicores >= $2
		     AND available_disk_mb >= $3
		     AND ($5 = '' OR supported_tiers @> ARRAY[$5]::text[])
		   ORDER BY array_length(supported_tiers, 1) ASC, available_memory_mb ASC
		   LIMIT 1
		   FOR UPDATE SKIP LOCKED
		 )
		 RETURNING id, address, status, total_memory_mb, total_cpu_millicores, total_disk_mb,
		           available_memory_mb, available_cpu_millicores, available_disk_mb,
		           active_sandboxes, last_heartbeat, supported_tiers`,
		memoryMb, cpuMillicores, diskMb, time.Now(), isolationTier,
	).Scan(&h.ID, &h.Address, &h.Status,
		&h.TotalResources.MemoryMb, &h.TotalResources.CpuMillicores, &h.TotalResources.DiskMb,
		&h.AvailableResources.MemoryMb, &h.AvailableResources.CpuMillicores, &h.AvailableResources.DiskMb,
		&h.ActiveSandboxes, &h.LastHeartbeat, pq.Array(&h.SupportedTiers))
	if err == sql.ErrNoRows {
		return nil, ErrNoCapacity
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *PostgresRepository) UpdateHeartbeat(ctx context.Context, hostID string, resources models.HostResources, activeSandboxes int32, supportedTiers []string) (*models.Host, error) {
	var h models.Host
	err := r.db.QueryRowContext(ctx,
		`UPDATE hosts SET
		   available_memory_mb = $1,
		   available_cpu_millicores = $2,
		   available_disk_mb = $3,
		   active_sandboxes = $4,
		   supported_tiers = $5,
		   last_heartbeat = now()
		 WHERE id = $6
		 RETURNING id, address, status, total_memory_mb, total_cpu_millicores, total_disk_mb,
		           available_memory_mb, available_cpu_millicores, available_disk_mb,
		           active_sandboxes, last_heartbeat, supported_tiers`,
		resources.MemoryMb, resources.CpuMillicores, resources.DiskMb, activeSandboxes,
		pq.Array(supportedTiers), hostID,
	).Scan(&h.ID, &h.Address, &h.Status,
		&h.TotalResources.MemoryMb, &h.TotalResources.CpuMillicores, &h.TotalResources.DiskMb,
		&h.AvailableResources.MemoryMb, &h.AvailableResources.CpuMillicores, &h.AvailableResources.DiskMb,
		&h.ActiveSandboxes, &h.LastHeartbeat, pq.Array(&h.SupportedTiers))
	if err == sql.ErrNoRows {
		return nil, ErrHostNotFound
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *PostgresRepository) MarkStaleHostsOffline(ctx context.Context, timeout time.Duration) (int64, error) {
	cutoff := time.Now().Add(-timeout)
	res, err := r.db.ExecContext(ctx,
		`UPDATE hosts SET status = 'offline'
		 WHERE status = 'ready'
		   AND last_heartbeat < $1`,
		cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// --- Warm Pool ---

func (r *PostgresRepository) UpsertWarmPoolConfig(ctx context.Context, cfg *models.WarmPoolConfig) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO warm_pool_configs (isolation_tier, target_count, memory_mb, cpu_millicores, disk_mb, updated_at)
		 VALUES ($1, $2, $3, $4, $5, now())
		 ON CONFLICT (isolation_tier) DO UPDATE SET
		   target_count = EXCLUDED.target_count,
		   memory_mb = EXCLUDED.memory_mb,
		   cpu_millicores = EXCLUDED.cpu_millicores,
		   disk_mb = EXCLUDED.disk_mb,
		   updated_at = now()`,
		cfg.IsolationTier, cfg.TargetCount, cfg.MemoryMb, cfg.CpuMillicores, cfg.DiskMb)
	return err
}

func (r *PostgresRepository) ListWarmPoolConfigs(ctx context.Context) ([]models.WarmPoolConfig, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT isolation_tier, target_count, memory_mb, cpu_millicores, disk_mb
		 FROM warm_pool_configs ORDER BY isolation_tier`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []models.WarmPoolConfig
	for rows.Next() {
		var c models.WarmPoolConfig
		if err := rows.Scan(&c.IsolationTier, &c.TargetCount, &c.MemoryMb, &c.CpuMillicores, &c.DiskMb); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

func (r *PostgresRepository) ClaimWarmSlot(ctx context.Context, tier string) (*models.WarmPoolSlot, error) {
	var s models.WarmPoolSlot
	err := r.db.QueryRowContext(ctx,
		`UPDATE warm_pool_slots SET status = 'claimed', claimed_at = now()
		 WHERE id = (
		   SELECT id FROM warm_pool_slots
		   WHERE isolation_tier = $1 AND status = 'ready'
		   LIMIT 1
		   FOR UPDATE SKIP LOCKED
		 )
		 RETURNING id, host_id, isolation_tier, memory_mb, cpu_millicores, disk_mb, status`,
		tier,
	).Scan(&s.ID, &s.HostID, &s.IsolationTier, &s.MemoryMb, &s.CpuMillicores, &s.DiskMb, &s.Status)
	if err == sql.ErrNoRows {
		return nil, nil // no warm slot available
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepository) CreateWarmSlot(ctx context.Context, slot *models.WarmPoolSlot) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO warm_pool_slots (host_id, isolation_tier, memory_mb, cpu_millicores, disk_mb)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		slot.HostID, slot.IsolationTier, slot.MemoryMb, slot.CpuMillicores, slot.DiskMb,
	).Scan(&slot.ID)
}

func (r *PostgresRepository) CountReadySlots(ctx context.Context, tier string) (int32, error) {
	var count int32
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM warm_pool_slots WHERE isolation_tier = $1 AND status = 'ready'`,
		tier,
	).Scan(&count)
	return count, err
}

func (r *PostgresRepository) CleanExpiredSlots(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM warm_pool_slots
		 WHERE status = 'ready'
		   AND host_id IN (SELECT id FROM hosts WHERE status = 'offline')`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *PostgresRepository) GetCapacity(ctx context.Context) ([]models.TierCapacity, int32, int32, error) {
	// Total and ready host counts.
	var totalHosts, readyHosts int32
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM hosts`).Scan(&totalHosts)
	if err != nil {
		return nil, 0, 0, err
	}
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM hosts WHERE status = 'ready'`).Scan(&readyHosts)
	if err != nil {
		return nil, 0, 0, err
	}

	// Per-tier capacity from hosts.
	rows, err := r.db.QueryContext(ctx,
		`SELECT tier,
			COUNT(DISTINCT h.id) AS hosts_supporting,
			COALESCE(SUM(h.available_memory_mb), 0),
			COALESCE(SUM(h.available_cpu_millicores), 0),
			COALESCE(SUM(h.available_disk_mb), 0)
		 FROM hosts h, unnest(h.supported_tiers) AS tier
		 WHERE h.status = 'ready'
		 GROUP BY tier
		 ORDER BY tier`)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	tierMap := make(map[string]*models.TierCapacity)
	var tiers []models.TierCapacity
	for rows.Next() {
		var tc models.TierCapacity
		if err := rows.Scan(&tc.IsolationTier, &tc.HostsSupporting,
			&tc.AvailableMemoryMb, &tc.AvailableCpuMilli, &tc.AvailableDiskMb); err != nil {
			return nil, 0, 0, err
		}
		tiers = append(tiers, tc)
		tierMap[tc.IsolationTier] = &tiers[len(tiers)-1]
	}
	if err := rows.Err(); err != nil {
		return nil, 0, 0, err
	}

	// Warm pool targets and ready counts.
	wpRows, err := r.db.QueryContext(ctx,
		`SELECT c.isolation_tier, c.target_count,
			COALESCE((SELECT COUNT(*) FROM warm_pool_slots s
				WHERE s.isolation_tier = c.isolation_tier AND s.status = 'ready'), 0)
		 FROM warm_pool_configs c`)
	if err != nil {
		return nil, 0, 0, err
	}
	defer wpRows.Close()

	for wpRows.Next() {
		var tier string
		var target, ready int32
		if err := wpRows.Scan(&tier, &target, &ready); err != nil {
			return nil, 0, 0, err
		}
		if tc, ok := tierMap[tier]; ok {
			tc.WarmSlotsTarget = target
			tc.WarmSlotsReady = ready
		} else {
			tiers = append(tiers, models.TierCapacity{
				IsolationTier:   tier,
				WarmSlotsTarget: target,
				WarmSlotsReady:  ready,
			})
		}
	}
	if err := wpRows.Err(); err != nil {
		return nil, 0, 0, err
	}

	return tiers, totalHosts, readyHosts, nil
}

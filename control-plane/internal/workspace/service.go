package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	hostagentpb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/host_agent/v1"
)

var (
	ErrWorkspaceNotFound        = errors.New("workspace not found")
	ErrWorkspaceAlreadyTerminal = errors.New("workspace is already in a terminal state")
	ErrWorkspaceNotRunning      = errors.New("workspace is not running")
	ErrSnapshotNotFound         = errors.New("snapshot not found")
	ErrInvalidInput             = errors.New("invalid input")
	ErrPlacementFailed          = errors.New("workspace placement failed")
	ErrHostAgentUnavailable     = errors.New("host agent unavailable")
)

const (
	defaultPageSize       = 50
	maxPageSize           = 100
	defaultMemoryMb       = 512
	defaultCpuMillicores  = 500
	defaultDiskMb         = 1024
	defaultMaxDurationSec = 3600
)

// ComputePlacer selects a host for workspace placement.
type ComputePlacer interface {
	PlaceWorkspace(ctx context.Context, memoryMb int64, cpuMillicores int32, diskMb int64) (hostID, hostAddress string, err error)
}

// PolicyCompiler compiles guardrail rules into bytes for the Host Agent evaluator.
type PolicyCompiler interface {
	CompilePolicy(ctx context.Context, ruleIDs []string) ([]byte, int, error)
}

// HostAgentDialer creates a gRPC connection to a Host Agent and returns a client.
type HostAgentDialer func(ctx context.Context, address string) (hostagentpb.HostAgentServiceClient, error)

// SnapshotStore persists and retrieves workspace snapshots.
type SnapshotStore interface {
	SaveSnapshot(ctx context.Context, workspaceID string) (snapshotID string, err error)
	LoadSnapshot(ctx context.Context, snapshotID string) error
}

// Service implements workspace business logic with Host Agent orchestration.
type Service struct {
	repo           Repository
	compute        ComputePlacer
	guardrails     PolicyCompiler
	dialHostAgent    HostAgentDialer
	snapshots      SnapshotStore
	logger         *zap.Logger
}

// ServiceConfig holds optional dependencies for the workspace service.
// If compute/guardrails/dialHostAgent are nil, the service operates in
// "DB-only" mode (no Host Agent orchestration).
type ServiceConfig struct {
	Repo        Repository
	Compute     ComputePlacer
	Guardrails  PolicyCompiler
	DialHostAgent HostAgentDialer
	Snapshots   SnapshotStore
	Logger      *zap.Logger
}

func NewService(cfg ServiceConfig) *Service {
	logger := cfg.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		repo:        cfg.Repo,
		compute:     cfg.Compute,
		guardrails:  cfg.Guardrails,
		dialHostAgent: cfg.DialHostAgent,
		snapshots:   cfg.Snapshots,
		logger:      logger,
	}
}

func (s *Service) CreateWorkspace(ctx context.Context, agentID, taskID string, spec *models.WorkspaceSpec) (*models.Workspace, error) {
	if agentID == "" {
		return nil, ErrInvalidInput
	}

	ws := &models.Workspace{
		AgentID: agentID,
		TaskID:  taskID,
		Status:  models.WorkspaceStatusPending,
	}

	// Apply spec with defaults.
	if spec != nil {
		ws.Spec = *spec
	}
	if ws.Spec.MemoryMb <= 0 {
		ws.Spec.MemoryMb = defaultMemoryMb
	}
	if ws.Spec.CpuMillicores <= 0 {
		ws.Spec.CpuMillicores = defaultCpuMillicores
	}
	if ws.Spec.DiskMb <= 0 {
		ws.Spec.DiskMb = defaultDiskMb
	}
	if ws.Spec.MaxDurationSecs <= 0 {
		ws.Spec.MaxDurationSecs = defaultMaxDurationSec
	}
	if ws.Spec.AllowedTools == nil {
		ws.Spec.AllowedTools = []string{}
	}
	if ws.Spec.EnvVars == nil {
		ws.Spec.EnvVars = map[string]string{}
	}
	if ws.Spec.EgressAllowlist == nil {
		ws.Spec.EgressAllowlist = []string{}
	}

	expiresAt := time.Now().Add(time.Duration(ws.Spec.MaxDurationSecs) * time.Second)
	ws.ExpiresAt = &expiresAt

	// Persist workspace in pending state.
	if err := s.repo.CreateWorkspace(ctx, ws); err != nil {
		return nil, err
	}

	// If orchestration dependencies are available, provision the sandbox.
	if s.compute != nil && s.dialHostAgent != nil {
		if err := s.provisionSandbox(ctx, ws); err != nil {
			s.logger.Error("sandbox provisioning failed",
				zap.String("workspace_id", ws.ID),
				zap.Error(err),
			)
			// Mark workspace as failed.
			_ = s.repo.UpdateWorkspaceStatus(ctx, ws.ID, models.WorkspaceStatusFailed, ws.HostID, "", "")
			ws.Status = models.WorkspaceStatusFailed
			return ws, nil
		}
	}

	return ws, nil
}

// provisionSandbox orchestrates: place → compile guardrails → create sandbox → update status.
func (s *Service) provisionSandbox(ctx context.Context, ws *models.Workspace) error {
	// 1. Transition to "creating".
	if err := s.repo.UpdateWorkspaceStatus(ctx, ws.ID, models.WorkspaceStatusCreating, "", "", ""); err != nil {
		return fmt.Errorf("update status to creating: %w", err)
	}
	ws.Status = models.WorkspaceStatusCreating

	// 2. Place workspace on a host via compute plane.
	hostID, hostAddress, err := s.compute.PlaceWorkspace(ctx, ws.Spec.MemoryMb, ws.Spec.CpuMillicores, ws.Spec.DiskMb)
	if err != nil {
		return fmt.Errorf("place workspace: %w", err)
	}
	ws.HostID = hostID

	s.logger.Info("workspace placed",
		zap.String("workspace_id", ws.ID),
		zap.String("host_id", hostID),
		zap.String("host_address", hostAddress),
	)

	// 3. Compile guardrail policy if policy ID is set.
	var compiledGuardrails []byte
	if ws.Spec.GuardrailPolicyID != "" && s.guardrails != nil {
		// GuardrailPolicyID is treated as a comma-separated list of rule IDs.
		ruleIDs := strings.Split(ws.Spec.GuardrailPolicyID, ",")
		for i := range ruleIDs {
			ruleIDs[i] = strings.TrimSpace(ruleIDs[i])
		}
		compiled, _, err := s.guardrails.CompilePolicy(ctx, ruleIDs)
		if err != nil {
			return fmt.Errorf("compile guardrails policy: %w", err)
		}
		compiledGuardrails = compiled
	} else {
		// Empty policy — all actions allowed.
		compiledGuardrails = []byte(`{"rules":[]}`)
	}

	// 4. Dial the Host Agent at the host address.
	hostAgentClient, err := s.dialHostAgent(ctx, hostAddress)
	if err != nil {
		return fmt.Errorf("dial host agent at %s: %w", hostAddress, err)
	}

	// 5. Create sandbox on the Host Agent.
	createResp, err := hostAgentClient.CreateSandbox(ctx, &hostagentpb.CreateSandboxRequest{
		WorkspaceId: ws.ID,
		AgentId:     ws.AgentID,
		Spec: &hostagentpb.SandboxSpec{
			MemoryMb:       ws.Spec.MemoryMb,
			CpuMillicores:  ws.Spec.CpuMillicores,
			DiskMb:         ws.Spec.DiskMb,
			AllowedTools:   ws.Spec.AllowedTools,
			EnvVars:        ws.Spec.EnvVars,
			ContainerImage:  ws.Spec.ContainerImage,
			EgressAllowlist: ws.Spec.EgressAllowlist,
		},
		CompiledGuardrails: compiledGuardrails,
	})
	if err != nil {
		return fmt.Errorf("create sandbox: %w", err)
	}

	ws.SandboxID = createResp.SandboxId

	s.logger.Info("sandbox created",
		zap.String("workspace_id", ws.ID),
		zap.String("sandbox_id", ws.SandboxID),
		zap.String("agent_api_endpoint", createResp.AgentApiEndpoint),
	)

	// 6. Transition to "running".
	if err := s.repo.UpdateWorkspaceStatus(ctx, ws.ID, models.WorkspaceStatusRunning, hostID, hostAddress, ws.SandboxID); err != nil {
		return fmt.Errorf("update status to running: %w", err)
	}
	ws.Status = models.WorkspaceStatusRunning

	return nil
}

func (s *Service) GetWorkspace(ctx context.Context, id string) (*models.Workspace, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	ws, err := s.repo.GetWorkspace(ctx, id)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}
	return ws, nil
}

func (s *Service) ListWorkspaces(ctx context.Context, agentID string, status models.WorkspaceStatus, pageSize int, pageToken string) ([]models.Workspace, string, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	afterID, err := decodePageToken(pageToken)
	if err != nil {
		return nil, "", ErrInvalidInput
	}

	workspaces, err := s.repo.ListWorkspaces(ctx, agentID, status, afterID, pageSize+1)
	if err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(workspaces) > pageSize {
		workspaces = workspaces[:pageSize]
		nextToken = encodePageToken(workspaces[pageSize-1].ID)
	}

	return workspaces, nextToken, nil
}

func (s *Service) TerminateWorkspace(ctx context.Context, id string, reason string) error {
	if id == "" {
		return ErrInvalidInput
	}

	// If orchestration is available, destroy the sandbox first.
	if s.dialHostAgent != nil {
		ws, err := s.repo.GetWorkspace(ctx, id)
		if err != nil {
			return err
		}
		if ws == nil {
			return ErrWorkspaceNotFound
		}

		if ws.SandboxID != "" && ws.HostID != "" {
			if err := s.destroySandbox(ctx, ws, reason); err != nil {
				s.logger.Warn("failed to destroy sandbox, proceeding with termination",
					zap.String("workspace_id", id),
					zap.String("sandbox_id", ws.SandboxID),
					zap.Error(err),
				)
			}
		}
	}

	return s.repo.TerminateWorkspace(ctx, id, reason)
}

func (s *Service) SnapshotWorkspace(ctx context.Context, id string) (*models.WorkspaceSnapshot, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	if s.snapshots == nil {
		return nil, errors.New("snapshot store not configured")
	}

	ws, err := s.repo.GetWorkspace(ctx, id)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}
	if ws.Status != models.WorkspaceStatusRunning {
		return nil, ErrWorkspaceNotRunning
	}

	// Save snapshot via snapshot store.
	snapshotID, err := s.snapshots.SaveSnapshot(ctx, ws.ID)
	if err != nil {
		return nil, fmt.Errorf("save snapshot: %w", err)
	}

	// Destroy the sandbox to free resources (best-effort).
	if s.dialHostAgent != nil && ws.SandboxID != "" && ws.HostAddress != "" {
		if err := s.destroySandbox(ctx, ws, "snapshot"); err != nil {
			s.logger.Warn("failed to destroy sandbox during snapshot",
				zap.String("workspace_id", id),
				zap.Error(err),
			)
		}
	}

	// Transition workspace to Paused with snapshot reference.
	if err := s.repo.UpdateWorkspaceStatus(ctx, ws.ID, models.WorkspaceStatusPaused, ws.HostID, ws.HostAddress, ""); err != nil {
		return nil, fmt.Errorf("update workspace status to paused: %w", err)
	}
	if err := s.repo.SetSnapshotID(ctx, ws.ID, snapshotID); err != nil {
		return nil, fmt.Errorf("set snapshot ID: %w", err)
	}

	snapshot := &models.WorkspaceSnapshot{
		ID:          snapshotID,
		WorkspaceID: ws.ID,
		AgentID:     ws.AgentID,
		TaskID:      ws.TaskID,
		CreatedAt:   time.Now(),
	}
	if err := s.repo.CreateSnapshot(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("create snapshot record: %w", err)
	}

	return snapshot, nil
}

func (s *Service) RestoreWorkspace(ctx context.Context, snapshotID string) (*models.Workspace, error) {
	if snapshotID == "" {
		return nil, ErrInvalidInput
	}
	if s.snapshots == nil {
		return nil, errors.New("snapshot store not configured")
	}

	snapshot, err := s.repo.GetSnapshot(ctx, snapshotID)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, ErrSnapshotNotFound
	}

	ws, err := s.repo.GetWorkspace(ctx, snapshot.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}
	if ws.Status != models.WorkspaceStatusPaused {
		return nil, fmt.Errorf("workspace must be paused to restore, current status: %s", ws.Status)
	}

	// Load snapshot data via snapshot store.
	if err := s.snapshots.LoadSnapshot(ctx, snapshotID); err != nil {
		return nil, fmt.Errorf("load snapshot: %w", err)
	}

	// Re-provision sandbox using existing orchestration pipeline.
	if s.compute != nil && s.dialHostAgent != nil {
		if err := s.provisionSandbox(ctx, ws); err != nil {
			s.logger.Error("sandbox re-provisioning failed during restore",
				zap.String("workspace_id", ws.ID),
				zap.Error(err),
			)
			_ = s.repo.UpdateWorkspaceStatus(ctx, ws.ID, models.WorkspaceStatusFailed, ws.HostID, "", "")
			ws.Status = models.WorkspaceStatusFailed
			return ws, nil
		}
	}

	// Clear snapshot reference.
	if err := s.repo.SetSnapshotID(ctx, ws.ID, ""); err != nil {
		return nil, fmt.Errorf("clear snapshot ID: %w", err)
	}

	return ws, nil
}

// destroySandbox dials the Host Agent and destroys the sandbox.
func (s *Service) destroySandbox(ctx context.Context, ws *models.Workspace, reason string) error {
	if ws.HostAddress == "" {
		s.logger.Warn("no host address stored, cannot dial host agent for teardown",
			zap.String("workspace_id", ws.ID),
			zap.String("sandbox_id", ws.SandboxID),
		)
		return nil
	}

	s.logger.Info("destroying sandbox",
		zap.String("workspace_id", ws.ID),
		zap.String("sandbox_id", ws.SandboxID),
		zap.String("host_address", ws.HostAddress),
	)

	hostAgentClient, err := s.dialHostAgent(ctx, ws.HostAddress)
	if err != nil {
		return fmt.Errorf("dial host agent at %s: %w", ws.HostAddress, err)
	}

	_, err = hostAgentClient.DestroySandbox(ctx, &hostagentpb.DestroySandboxRequest{
		SandboxId: ws.SandboxID,
		Reason:    reason,
	})
	if err != nil {
		return fmt.Errorf("destroy sandbox: %w", err)
	}

	return nil
}

func encodePageToken(id string) string {
	if id == "" {
		return ""
	}
	return id
}

func decodePageToken(token string) (string, error) {
	return token, nil
}

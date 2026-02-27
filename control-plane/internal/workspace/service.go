package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
	runtimepb "github.com/baselyne/agent-sandbox-platform/control-plane/pkg/gen/runtime/v1"
)

var (
	ErrWorkspaceNotFound        = errors.New("workspace not found")
	ErrWorkspaceAlreadyTerminal = errors.New("workspace is already in a terminal state")
	ErrInvalidInput             = errors.New("invalid input")
	ErrPlacementFailed          = errors.New("workspace placement failed")
	ErrRuntimeUnavailable       = errors.New("runtime unavailable")
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

// PolicyCompiler compiles guardrail rules into bytes for the runtime evaluator.
type PolicyCompiler interface {
	CompilePolicy(ctx context.Context, ruleIDs []string) ([]byte, int, error)
}

// RuntimeDialer creates a gRPC connection to a runtime host and returns a client.
type RuntimeDialer func(ctx context.Context, address string) (runtimepb.RuntimeServiceClient, error)

// Service implements workspace business logic with runtime orchestration.
type Service struct {
	repo           Repository
	compute        ComputePlacer
	guardrails     PolicyCompiler
	dialRuntime    RuntimeDialer
	logger         *zap.Logger
}

// ServiceConfig holds optional dependencies for the workspace service.
// If compute/guardrails/dialRuntime are nil, the service operates in
// "DB-only" mode (no runtime orchestration).
type ServiceConfig struct {
	Repo        Repository
	Compute     ComputePlacer
	Guardrails  PolicyCompiler
	DialRuntime RuntimeDialer
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
		dialRuntime: cfg.DialRuntime,
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

	expiresAt := time.Now().Add(time.Duration(ws.Spec.MaxDurationSecs) * time.Second)
	ws.ExpiresAt = &expiresAt

	// Persist workspace in pending state.
	if err := s.repo.CreateWorkspace(ctx, ws); err != nil {
		return nil, err
	}

	// If orchestration dependencies are available, provision the sandbox.
	if s.compute != nil && s.dialRuntime != nil {
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

	// 4. Dial the runtime at the host address.
	runtimeClient, err := s.dialRuntime(ctx, hostAddress)
	if err != nil {
		return fmt.Errorf("dial runtime at %s: %w", hostAddress, err)
	}

	// 5. Create sandbox on the runtime.
	createResp, err := runtimeClient.CreateSandbox(ctx, &runtimepb.CreateSandboxRequest{
		WorkspaceId: ws.ID,
		AgentId:     ws.AgentID,
		Spec: &runtimepb.SandboxSpec{
			MemoryMb:      ws.Spec.MemoryMb,
			CpuMillicores: ws.Spec.CpuMillicores,
			DiskMb:        ws.Spec.DiskMb,
			AllowedTools:  ws.Spec.AllowedTools,
			EnvVars:       ws.Spec.EnvVars,
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
	if s.dialRuntime != nil {
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

// destroySandbox dials the runtime and destroys the sandbox.
func (s *Service) destroySandbox(ctx context.Context, ws *models.Workspace, reason string) error {
	if ws.HostAddress == "" {
		s.logger.Warn("no host address stored, cannot dial runtime for teardown",
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

	runtimeClient, err := s.dialRuntime(ctx, ws.HostAddress)
	if err != nil {
		return fmt.Errorf("dial runtime at %s: %w", ws.HostAddress, err)
	}

	_, err = runtimeClient.DestroySandbox(ctx, &runtimepb.DestroySandboxRequest{
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

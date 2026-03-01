// Binary control-plane consolidates Identity, Task, Workspace, and Compute
// services into a single gRPC server. Inter-service calls that previously
// crossed the network (Task→Workspace, Workspace→Compute) are now direct
// function calls via in-process adapters. Cross-binary calls (Workspace→
// Guardrails, Workspace→HostAgent) remain gRPC.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/adapters"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/boot"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/compute"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/identity"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/task"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/workspace"
	computepb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/compute/v1"
	guardrailspb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/guardrails/v1"
	hostagentpb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/host_agent/v1"
	identitypb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/identity/v1"
	taskpb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/task/v1"
	workspacepb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/workspace/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	f := boot.New(boot.Options{ServiceName: "bulkhead-control-plane"})

	// -----------------------------------------------------------------------
	// 1. Identity
	// -----------------------------------------------------------------------
	identityRepo := identity.NewPostgresRepository(f.DB)
	identitySvc := identity.NewService(identityRepo)
	identityHandler := identity.NewHandler(identitySvc)
	identitypb.RegisterIdentityServiceServer(f.Server, identityHandler)

	// -----------------------------------------------------------------------
	// 2. Compute
	// -----------------------------------------------------------------------
	computeRepo := compute.NewPostgresRepository(f.DB)
	computeSvc := compute.NewService(computeRepo)
	computeHandler := compute.NewHandler(computeSvc)
	computepb.RegisterComputePlaneServiceServer(f.Server, computeHandler)

	// Start compute background workers.
	heartbeatTimeoutSecs := 180
	if v := os.Getenv("HEARTBEAT_TIMEOUT_SECS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			heartbeatTimeoutSecs = parsed
		}
	}
	workerCtx, workerCancel := context.WithCancel(context.Background())
	livenessWorker := compute.NewLivenessWorker(computeRepo, compute.LivenessWorkerConfig{
		Interval:         60 * time.Second,
		HeartbeatTimeout: time.Duration(heartbeatTimeoutSecs) * time.Second,
		Logger:           f.Logger,
	})
	go livenessWorker.Run(workerCtx)

	warmPoolWorker := compute.NewWarmPoolWorker(computeRepo, compute.WarmPoolWorkerConfig{
		Interval: 30 * time.Second,
		Logger:   f.Logger,
	})
	go warmPoolWorker.Run(workerCtx)

	// -----------------------------------------------------------------------
	// 3. Workspace
	// -----------------------------------------------------------------------
	wsRepo := workspace.NewPostgresRepository(f.DB)

	// In-process: Compute service directly satisfies workspace.ComputePlacer.
	var computePlacer workspace.ComputePlacer = computeSvc

	// Cross-binary: Guardrails policy compiler (gRPC to the policy binary).
	var guardrails workspace.PolicyCompiler
	if ep := os.Getenv("GUARDRAILS_ENDPOINT"); ep != "" {
		f.Logger.Info("connecting to guardrails service", zap.String("endpoint", ep))
		conn, err := grpc.NewClient(ep, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			f.Logger.Fatal("failed to connect to guardrails service", zap.Error(err))
		}
		defer conn.Close()
		guardrails = &guardrailsAdapter{client: guardrailspb.NewGuardrailsServiceClient(conn)}
	}

	// Cross-binary: Host Agent dialer (gRPC to individual hosts).
	var dialHostAgent workspace.HostAgentDialer
	dialHostAgent = func(_ context.Context, address string) (hostagentpb.HostAgentServiceClient, error) {
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("dial host agent at %s: %w", address, err)
		}
		return hostagentpb.NewHostAgentServiceClient(conn), nil
	}

	// In-process: Identity adapters for credential minting and agent queries.
	credentialMinter := &adapters.IdentityCredentialAdapter{Svc: identitySvc, DB: f.DB}
	identityQuerier := &adapters.IdentityQuerierAdapter{DB: f.DB}

	wsSvc := workspace.NewService(workspace.ServiceConfig{
		Repo:          wsRepo,
		Compute:       computePlacer,
		Guardrails:    guardrails,
		DialHostAgent: dialHostAgent,
		Credentials:   credentialMinter,
		Identity:      identityQuerier,
		Logger:        f.Logger,
	})
	wsHandler := workspace.NewHandler(wsSvc)
	workspacepb.RegisterWorkspaceServiceServer(f.Server, wsHandler)

	// -----------------------------------------------------------------------
	// 4. Task
	// -----------------------------------------------------------------------
	taskRepo := task.NewPostgresRepository(f.DB)

	// In-process: Workspace service as provisioner (no gRPC hop).
	provisioner := &adapters.WorkspaceProvisionerAdapter{WsSvc: wsSvc, DB: f.DB}

	taskSvc := task.NewService(task.ServiceConfig{
		Repo:        taskRepo,
		Provisioner: provisioner,
	})
	taskHandler := task.NewHandler(taskSvc)
	taskpb.RegisterTaskServiceServer(f.Server, taskHandler)

	// -----------------------------------------------------------------------
	// Serve
	// -----------------------------------------------------------------------
	f.Logger.Info("registered services: Identity, Compute, Workspace, Task")
	f.Serve()
	workerCancel()
}

// guardrailsAdapter wraps the GuardrailsServiceClient for cross-binary calls.
type guardrailsAdapter struct {
	client guardrailspb.GuardrailsServiceClient
}

func (a *guardrailsAdapter) CompilePolicy(ctx context.Context, _ string, ruleIDs []string) ([]byte, int, error) {
	resp, err := a.client.CompilePolicy(ctx, &guardrailspb.CompilePolicyRequest{
		RuleIds: ruleIDs,
	})
	if err != nil {
		return nil, 0, err
	}
	return resp.GetCompiledPolicy(), int(resp.GetRuleCount()), nil
}

func (a *guardrailsAdapter) ResolveRuleIDs(ctx context.Context, tenantID, policyID string) ([]string, error) {
	if strings.HasPrefix(policyID, "set:") {
		setName := strings.TrimPrefix(policyID, "set:")
		listResp, err := a.client.ListGuardrailSets(ctx, &guardrailspb.ListGuardrailSetsRequest{PageSize: 100})
		if err != nil {
			return nil, fmt.Errorf("list guardrail sets: %w", err)
		}
		for _, s := range listResp.GetSets() {
			if s.GetName() == setName {
				return s.GetRuleIds(), nil
			}
		}
		return nil, fmt.Errorf("guardrail set %q not found", setName)
	}
	ids := strings.Split(policyID, ",")
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}
	return ids, nil
}

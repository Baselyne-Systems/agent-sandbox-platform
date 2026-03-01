package e2e

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/activity"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/compute"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/economics"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/governance"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/guardrails"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/human"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/identity"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/workspace"

	activitypb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/activity/v1"
	computepb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/compute/v1"
	economicspb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/economics/v1"
	governancepb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/governance/v1"
	guardrailspb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/guardrails/v1"
	hostagentpb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/host_agent/v1"
	humanpb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/human/v1"
	identitypb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/identity/v1"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/middleware"
)

// ---------------------------------------------------------------------------
// Full-stack test helpers
// ---------------------------------------------------------------------------

// fullStackPorts holds the dynamic ports allocated for each consolidated binary.
type fullStackPorts struct {
	controlPlane  int // Identity, Compute, Workspace, Task
	policy        int // Guardrails, Governance
	observability int // Activity, Economics, Human
}

// tenantResolverInterceptor is a gRPC unary server interceptor that resolves
// tenant_id from the request's agent_id field via DB lookup. This simulates
// what the auth middleware does in production (extract tenant from JWT).
func tenantResolverInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if msg, ok := req.(proto.Message); ok {
			md := msg.ProtoReflect()
			// Try to resolve tenant from various ID fields in the request.
			type lookup struct {
				field string
				query string
			}
			lookups := []lookup{
				{"agent_id", "SELECT tenant_id FROM agents WHERE id = $1"},
				{"workspace_id", "SELECT tenant_id FROM workspaces WHERE id = $1"},
				{"request_id", "SELECT tenant_id FROM human_requests WHERE id = $1"},
			}
			for _, l := range lookups {
				fd := md.Descriptor().Fields().ByName(protoreflect.Name(l.field))
				if fd == nil {
					continue
				}
				id := md.Get(fd).String()
				if id == "" {
					continue
				}
				var tenantID string
				if err := db.QueryRowContext(ctx, l.query, id).Scan(&tenantID); err == nil && tenantID != "" {
					ctx = middleware.ContextWithTenantID(ctx, tenantID)
					break
				}
			}
		}
		return handler(ctx, req)
	}
}

// startGRPCServers launches the 3 consolidated control-plane gRPC services
// on dynamic ports. Returns servers and the ports they're listening on.
func startGRPCServers(t *testing.T) ([]*grpc.Server, fullStackPorts) {
	t.Helper()

	ports := fullStackPorts{}
	var servers []*grpc.Server
	interceptor := tenantResolverInterceptor()

	start := func(name string, register func(s *grpc.Server)) int {
		t.Helper()
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("listen for %s: %v", name, err)
		}
		port := lis.Addr().(*net.TCPAddr).Port

		srv := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
		register(srv)
		servers = append(servers, srv)

		go func() {
			if err := srv.Serve(lis); err != nil {
				// Swallow error from GracefulStop during cleanup.
			}
		}()
		return port
	}

	// Control-plane binary: Identity + Compute + Workspace
	ports.controlPlane = start("control-plane", func(s *grpc.Server) {
		identitypb.RegisterIdentityServiceServer(s, identity.NewHandler(identitySvc))
		computepb.RegisterComputePlaneServiceServer(s, compute.NewHandler(computeSvc))
		// Workspace handler registered for completeness; runtime doesn't call it directly.
		_ = workspace.NewHandler(workspaceSvc)
	})

	// Policy binary: Guardrails + Governance
	ports.policy = start("policy", func(s *grpc.Server) {
		guardrailspb.RegisterGuardrailsServiceServer(s, guardrails.NewHandler(guardrailsSvc))
		governancepb.RegisterDataGovernanceServiceServer(s, governance.NewHandler(governanceSvc))
	})

	// Observability binary: Activity + Economics + Human
	ports.observability = start("observability", func(s *grpc.Server) {
		activitypb.RegisterActivityServiceServer(s, activity.NewHandler(activitySvc))
		economicspb.RegisterEconomicsServiceServer(s, economics.NewHandler(economicsSvc))
		humanpb.RegisterHumanInteractionServiceServer(s, human.NewHandler(humanSvc))
	})

	return servers, ports
}

// runtimeBinaryPath returns the path to the runtime binary, or empty if not available.
func runtimeBinaryPath() string {
	if p := os.Getenv("RUNTIME_BINARY"); p != "" {
		return p
	}
	return ""
}

// startRuntime starts the Rust runtime binary as a subprocess.
// It connects to the control-plane gRPC services and self-registers with the compute plane.
func startRuntime(t *testing.T, ports fullStackPorts) (*exec.Cmd, int) {
	t.Helper()

	binary := runtimeBinaryPath()
	if binary == "" {
		t.Skip("RUNTIME_BINARY not set — skipping full-stack test")
	}
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		t.Skipf("runtime binary not found at %s — skipping full-stack test", binary)
	}

	// Find a free port for the runtime's gRPC server.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port for runtime: %v", err)
	}
	runtimePort := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	cmd := exec.Command(binary)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GRPC_PORT=%d", runtimePort),
		"ADVERTISE_ADDRESS=127.0.0.1",
		"ENABLE_DOCKER=false",
		fmt.Sprintf("TOTAL_MEMORY_MB=%d", 8192),
		fmt.Sprintf("TOTAL_CPU_MILLICORES=%d", 4000),
		fmt.Sprintf("TOTAL_DISK_MB=%d", 10240),
		"SUPPORTED_TIERS=standard,hardened",
		fmt.Sprintf("COMPUTE_ENDPOINT=http://127.0.0.1:%d", ports.controlPlane),
		fmt.Sprintf("HIS_ENDPOINT=http://127.0.0.1:%d", ports.observability),
		fmt.Sprintf("ACTIVITY_ENDPOINT=http://127.0.0.1:%d", ports.observability),
		fmt.Sprintf("ECONOMICS_ENDPOINT=http://127.0.0.1:%d", ports.observability),
		fmt.Sprintf("GOVERNANCE_ENDPOINT=http://127.0.0.1:%d", ports.policy),
		"RUST_LOG=info",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start runtime: %v", err)
	}

	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})

	return cmd, runtimePort
}

// waitForRuntime polls until the runtime gRPC port is accepting connections.
func waitForRuntime(t *testing.T, port int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("runtime did not start within %v", timeout)
}

// dialAgentAPI connects to the runtime's Agent API.
func dialAgentAPI(t *testing.T, port int) hostagentpb.HostAgentAPIServiceClient {
	t.Helper()
	conn, err := grpc.NewClient(
		fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial agent API: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return hostagentpb.NewHostAgentAPIServiceClient(conn)
}

// makeParams creates a structpb.Struct from a map.
func makeParams(t *testing.T, m map[string]interface{}) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	if err != nil {
		t.Fatalf("make params: %v", err)
	}
	return s
}

// dialHostAgentService connects to the runtime's HostAgentService (control-plane facing).
func dialHostAgentService(t *testing.T, port int) hostagentpb.HostAgentServiceClient {
	t.Helper()
	conn, err := grpc.NewClient(
		fmt.Sprintf("127.0.0.1:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial host agent service: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return hostagentpb.NewHostAgentServiceClient(conn)
}

// withSandboxID adds x-sandbox-id gRPC metadata to the context.
func withSandboxID(ctx context.Context, sandboxID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-sandbox-id", sandboxID)
}

// emptyPolicy is a valid empty guardrails policy the runtime can deserialize.
var emptyPolicy = []byte(`{"rules":[]}`)

// createSandboxOnRuntime creates a sandbox directly on the runtime via HostAgentService.
func createSandboxOnRuntime(t *testing.T, ctx context.Context, hostClient hostagentpb.HostAgentServiceClient,
	workspaceID, agentID string, compiledGuardrails []byte) string {
	t.Helper()
	if compiledGuardrails == nil {
		compiledGuardrails = emptyPolicy
	}
	resp, err := hostClient.CreateSandbox(ctx, &hostagentpb.CreateSandboxRequest{
		WorkspaceId: workspaceID,
		AgentId:     agentID,
		Spec: &hostagentpb.SandboxSpec{
			MemoryMb:      512,
			CpuMillicores: 500,
			DiskMb:        1024,
		},
		CompiledGuardrails: compiledGuardrails,
	})
	if err != nil {
		t.Fatalf("create sandbox on runtime: %v", err)
	}
	return resp.SandboxId
}

// ---------------------------------------------------------------------------
// Full-stack tests
// ---------------------------------------------------------------------------

func TestFullStackRuntimeRegistration(t *testing.T) {
	clean(t)

	servers, ports := startGRPCServers(t)
	defer func() {
		for _, s := range servers {
			s.GracefulStop()
		}
	}()

	_, runtimePort := startRuntime(t, ports)
	waitForRuntime(t, runtimePort, 30*time.Second)

	// Give the runtime time to register and send first heartbeat.
	time.Sleep(2 * time.Second)

	hosts, err := computeSvc.ListHosts(context.Background(), models.HostStatusReady)
	if err != nil {
		t.Fatalf("list hosts: %v", err)
	}
	if len(hosts) == 0 {
		t.Fatal("expected runtime to self-register as a host")
	}

	found := false
	for _, h := range hosts {
		if h.TotalResources.MemoryMb == 8192 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("did not find a host with 8192 MB memory (expected from runtime)")
	}
}

func TestFullStackWorkspaceCreation(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	servers, ports := startGRPCServers(t)
	defer func() {
		for _, s := range servers {
			s.GracefulStop()
		}
	}()

	_, runtimePort := startRuntime(t, ports)
	waitForRuntime(t, runtimePort, 30*time.Second)
	time.Sleep(2 * time.Second)

	agent := registerAgent(t, ctx, tenant, "fullstack-agent")

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if ws.Status != models.WorkspaceStatusRunning {
		t.Fatalf("expected workspace running, got %s", ws.Status)
	}
	if ws.SandboxID == "" {
		t.Fatal("expected sandbox ID to be set")
	}
}

func TestFullStackExecuteToolAllow(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	servers, ports := startGRPCServers(t)
	defer func() {
		for _, s := range servers {
			s.GracefulStop()
		}
	}()

	_, runtimePort := startRuntime(t, ports)
	waitForRuntime(t, runtimePort, 30*time.Second)
	time.Sleep(2 * time.Second)

	agent := registerAgent(t, ctx, tenant, "tool-allow-agent")

	// Create workspace for DB records (budget checks, activity recording).
	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	// Create sandbox on the real runtime (no guardrails = allow all).
	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, nil)

	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	resp, err := apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "web_search",
		Parameters: makeParams(t, map[string]interface{}{"query": "test"}),
	})
	if err != nil {
		t.Fatalf("execute tool: %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ALLOW {
		t.Fatalf("expected ALLOW verdict, got %s", resp.Verdict)
	}
}

func TestFullStackExecuteToolDeny(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	servers, ports := startGRPCServers(t)
	defer func() {
		for _, s := range servers {
			s.GracefulStop()
		}
	}()

	_, runtimePort := startRuntime(t, ports)
	waitForRuntime(t, runtimePort, 30*time.Second)
	time.Sleep(2 * time.Second)

	agent := registerAgent(t, ctx, tenant, "tool-deny-agent")

	// Create a deny rule for "shell".
	denyRule, err := guardrailsSvc.CreateRule(ctx, tenant, "deny-shell", "block shell",
		models.RuleTypeToolFilter, "shell", models.RuleActionDeny, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create deny rule: %v", err)
	}

	// Compile guardrails to pass to the runtime.
	compiled, _, err := guardrailsSvc.CompilePolicy(ctx, tenant, []string{denyRule.ID})
	if err != nil {
		t.Fatalf("compile policy: %v", err)
	}

	// Create workspace for DB records.
	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	// Create sandbox on the real runtime with compiled guardrails.
	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, compiled)

	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	resp, err := apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "shell",
		Parameters: makeParams(t, map[string]interface{}{"command": "rm -rf /"}),
	})
	if err != nil {
		t.Fatalf("execute tool: %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_DENY {
		t.Fatalf("expected DENY verdict, got %s", resp.Verdict)
	}
	if resp.DenialReason == "" {
		t.Fatal("expected denial reason to be set")
	}
}

func TestFullStackBudgetEnforcement(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	servers, ports := startGRPCServers(t)
	defer func() {
		for _, s := range servers {
			s.GracefulStop()
		}
	}()

	_, runtimePort := startRuntime(t, ports)
	waitForRuntime(t, runtimePort, 30*time.Second)
	time.Sleep(2 * time.Second)

	agent := registerAgent(t, ctx, tenant, "budget-agent")

	// Set a tight budget and over-exhaust it.
	_, err := economicsSvc.SetBudget(ctx, tenant, agent.ID, 10.0, "USD", "halt", 0.0)
	if err != nil {
		t.Fatalf("set budget: %v", err)
	}

	// Create workspace first (for valid workspace_id in usage records).
	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	// Exhaust the budget (record more than the limit so remaining goes negative).
	_, err = economicsSvc.RecordUsage(ctx, tenant, agent.ID, ws.ID, "compute", "hour", 1.0, 15.0)
	if err != nil {
		t.Fatalf("record usage: %v", err)
	}

	// Create sandbox on the real runtime.
	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, nil)

	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	resp, err := apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "web_search",
		Parameters: makeParams(t, map[string]interface{}{"query": "test"}),
	})
	if err != nil {
		t.Fatalf("execute tool: %v", err)
	}
	// Budget is exhausted, so the runtime should deny the tool execution.
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_DENY {
		t.Fatalf("expected DENY (budget exhausted), got %s", resp.Verdict)
	}
}

func TestFullStackHumanInteraction(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	servers, ports := startGRPCServers(t)
	defer func() {
		for _, s := range servers {
			s.GracefulStop()
		}
	}()

	_, runtimePort := startRuntime(t, ports)
	waitForRuntime(t, runtimePort, 30*time.Second)
	time.Sleep(2 * time.Second)

	agent := registerAgent(t, ctx, tenant, "human-agent")

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	// Create sandbox on the real runtime.
	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, nil)

	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	// Request human input via the runtime.
	hiResp, err := apiClient.RequestHumanInput(sandboxCtx, &hostagentpb.RequestHumanInputRequest{
		Question:       "approve deployment?",
		Options:        []string{"yes", "no"},
		Context:        "deploying to production",
		TimeoutSeconds: 300,
	})
	if err != nil {
		t.Fatalf("request human input: %v", err)
	}
	requestID := hiResp.GetRequestId()
	if requestID == "" {
		t.Fatal("expected request ID from human input request")
	}

	// Respond via the control-plane human service.
	err = humanSvc.RespondToRequest(ctx, tenant, requestID, "yes", "admin")
	if err != nil {
		t.Fatalf("respond to request: %v", err)
	}

	// Check via the runtime API that the response is available.
	checkResp, err := apiClient.CheckHumanRequest(sandboxCtx, &hostagentpb.CheckHumanRequestRequest{
		RequestId: requestID,
	})
	if err != nil {
		t.Fatalf("check human request: %v", err)
	}
	if checkResp.Status != "responded" {
		t.Fatalf("expected 'responded' status, got %q", checkResp.Status)
	}
	if checkResp.Response != "yes" {
		t.Fatalf("expected response 'yes', got %q", checkResp.Response)
	}
}

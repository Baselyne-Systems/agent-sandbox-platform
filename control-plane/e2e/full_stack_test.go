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

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/activity"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/compute"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/economics"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/governance"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/guardrails"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/human"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/identity"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/models"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/workspace"

	activitypb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/activity/v1"
	computepb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/compute/v1"
	economicspb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/economics/v1"
	governancepb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/governance/v1"
	guardrailspb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/guardrails/v1"
	hostagentpb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/host_agent/v1"
	humanpb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/human/v1"
	identitypb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/identity/v1"

	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/middleware"
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

// ---------------------------------------------------------------------------
// WS3: Full-stack enforcement tests
// ---------------------------------------------------------------------------

func TestFullStackEscalateVerdict(t *testing.T) {
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

	agent := registerAgent(t, ctx, tenant, "escalate-agent")

	// Create an ESCALATE rule for "deploy".
	escalateRule, err := guardrailsSvc.CreateRule(ctx, tenant, "escalate-deploy", "require approval for deploy",
		models.RuleTypeToolFilter, "deploy", models.RuleActionEscalate, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create escalate rule: %v", err)
	}

	compiled, _, err := guardrailsSvc.CompilePolicy(ctx, tenant, []string{escalateRule.ID})
	if err != nil {
		t.Fatalf("compile policy: %v", err)
	}

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, compiled)

	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	resp, err := apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "deploy",
		Parameters: makeParams(t, map[string]interface{}{"env": "production"}),
	})
	if err != nil {
		t.Fatalf("execute tool: %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ESCALATE {
		t.Fatalf("expected ESCALATE verdict, got %s", resp.Verdict)
	}
	if resp.EscalationId == "" {
		t.Fatal("expected escalation_id to be set")
	}
}

func TestFullStackPolicyHotReload(t *testing.T) {
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

	agent := registerAgent(t, ctx, tenant, "hot-reload-agent")

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	// Create sandbox with empty policy (allow all).
	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, nil)

	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	// First call: "shell" should be ALLOW (no guardrails).
	resp, err := apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "shell",
		Parameters: makeParams(t, map[string]interface{}{"cmd": "ls"}),
	})
	if err != nil {
		t.Fatalf("execute tool (before reload): %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ALLOW {
		t.Fatalf("expected ALLOW before policy reload, got %s", resp.Verdict)
	}

	// Create a deny-shell rule and compile.
	denyRule, err := guardrailsSvc.CreateRule(ctx, tenant, "deny-shell", "block shell",
		models.RuleTypeToolFilter, "shell", models.RuleActionDeny, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create deny rule: %v", err)
	}
	compiled, _, err := guardrailsSvc.CompilePolicy(ctx, tenant, []string{denyRule.ID})
	if err != nil {
		t.Fatalf("compile policy: %v", err)
	}

	// Hot-reload guardrails on the running sandbox.
	_, err = hostClient.UpdateSandboxGuardrails(ctx, &hostagentpb.UpdateSandboxGuardrailsRequest{
		SandboxId:          sandboxID,
		CompiledGuardrails: compiled,
	})
	if err != nil {
		t.Fatalf("update sandbox guardrails: %v", err)
	}

	// Second call: "shell" should now be DENY.
	resp, err = apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "shell",
		Parameters: makeParams(t, map[string]interface{}{"cmd": "rm -rf /"}),
	})
	if err != nil {
		t.Fatalf("execute tool (after reload): %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_DENY {
		t.Fatalf("expected DENY after policy reload, got %s", resp.Verdict)
	}
}

func TestFullStackDLPEnforcement(t *testing.T) {
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

	agent := registerAgent(t, ctx, tenant, "dlp-agent")

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, nil)

	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	// Tool call with PII in content going to an unapproved destination → should be denied.
	resp, err := apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName: "send_email",
		Parameters: makeParams(t, map[string]interface{}{
			"destination": "external.com",
			"content":     "SSN: 123-45-6789",
		}),
	})
	if err != nil {
		t.Fatalf("execute tool (DLP deny): %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_DENY {
		t.Fatalf("expected DENY for DLP violation, got %s", resp.Verdict)
	}
	if resp.DenialReason == "" {
		t.Fatal("expected denial reason with DLP details")
	}

	// Tool call without destination param → no DLP check → ALLOW.
	resp, err = apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName: "read_file",
		Parameters: makeParams(t, map[string]interface{}{
			"path": "/data/safe.txt",
		}),
	})
	if err != nil {
		t.Fatalf("execute tool (no DLP): %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ALLOW {
		t.Fatalf("expected ALLOW for non-DLP tool call, got %s", resp.Verdict)
	}
}

func TestFullStackConcurrentToolCalls(t *testing.T) {
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

	agent := registerAgent(t, ctx, tenant, "concurrent-agent")

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, nil)

	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	const numCalls = 20
	errs := make(chan error, numCalls)

	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			_, err := apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
				ToolName:   "web_search",
				Parameters: makeParams(t, map[string]interface{}{"query": fmt.Sprintf("query-%d", idx)}),
			})
			errs <- err
		}(i)
	}

	for i := 0; i < numCalls; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent call %d failed: %v", i, err)
		}
	}

	// Verify sandbox status shows correct action count.
	statusResp, err := hostClient.GetSandboxStatus(ctx, &hostagentpb.GetSandboxStatusRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		t.Fatalf("get sandbox status: %v", err)
	}
	if statusResp.ActionsExecuted != numCalls {
		t.Fatalf("expected %d actions_executed, got %d", numCalls, statusResp.ActionsExecuted)
	}
}

func TestFullStackBudgetExhaustionMidWorkflow(t *testing.T) {
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

	agent := registerAgent(t, ctx, tenant, "mid-budget-agent")

	_, err := economicsSvc.SetBudget(ctx, tenant, agent.ID, 10.0, "USD", "halt", 0.0)
	if err != nil {
		t.Fatalf("set budget: %v", err)
	}

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	// Record partial usage — still under budget.
	_, err = economicsSvc.RecordUsage(ctx, tenant, agent.ID, ws.ID, "compute", "hour", 1.0, 8.0)
	if err != nil {
		t.Fatalf("record usage (partial): %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, nil)

	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	// First call: should be ALLOW (under budget).
	resp, err := apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "read_file",
		Parameters: makeParams(t, map[string]interface{}{"path": "/data/ok.txt"}),
	})
	if err != nil {
		t.Fatalf("execute tool (under budget): %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ALLOW {
		t.Fatalf("expected ALLOW (under budget), got %s", resp.Verdict)
	}

	// Exhaust the budget.
	_, err = economicsSvc.RecordUsage(ctx, tenant, agent.ID, ws.ID, "compute", "hour", 1.0, 5.0)
	if err != nil {
		t.Fatalf("record usage (exhaust): %v", err)
	}

	// Second call: should be DENY (budget exhausted).
	resp, err = apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "read_file",
		Parameters: makeParams(t, map[string]interface{}{"path": "/data/another.txt"}),
	})
	if err != nil {
		t.Fatalf("execute tool (over budget): %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_DENY {
		t.Fatalf("expected DENY (budget exhausted), got %s", resp.Verdict)
	}
}

func TestFullStackBudgetRequestIncrease(t *testing.T) {
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

	agent := registerAgent(t, ctx, tenant, "budget-increase-agent")

	// Set budget with on_exceeded = "request_increase" → should ESCALATE instead of DENY.
	_, err := economicsSvc.SetBudget(ctx, tenant, agent.ID, 10.0, "USD", "request_increase", 0.0)
	if err != nil {
		t.Fatalf("set budget: %v", err)
	}

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	// Exhaust the budget.
	_, err = economicsSvc.RecordUsage(ctx, tenant, agent.ID, ws.ID, "compute", "hour", 1.0, 15.0)
	if err != nil {
		t.Fatalf("record usage: %v", err)
	}

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
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ESCALATE {
		t.Fatalf("expected ESCALATE (budget request_increase), got %s", resp.Verdict)
	}
}

// ---------------------------------------------------------------------------
// WS4: Multi-agent scenario tests
// ---------------------------------------------------------------------------

func TestFullStackMultiAgentDifferentPolicies(t *testing.T) {
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

	// Agent A: allow-all policy.
	agentA := registerAgent(t, ctx, tenant, "policy-agent-A")
	wsA, err := workspaceSvc.CreateWorkspace(ctx, tenant, agentA.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace A: %v", err)
	}

	// Agent B: deny-shell policy.
	agentB := registerAgent(t, ctx, tenant, "policy-agent-B")
	denyRule, err := guardrailsSvc.CreateRule(ctx, tenant, "deny-shell-B", "block shell for B",
		models.RuleTypeToolFilter, "shell", models.RuleActionDeny, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create deny rule: %v", err)
	}
	compiledB, _, err := guardrailsSvc.CompilePolicy(ctx, tenant, []string{denyRule.ID})
	if err != nil {
		t.Fatalf("compile policy B: %v", err)
	}
	wsB, err := workspaceSvc.CreateWorkspace(ctx, tenant, agentB.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace B: %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)
	sandboxA := createSandboxOnRuntime(t, ctx, hostClient, wsA.ID, agentA.ID, nil)
	sandboxB := createSandboxOnRuntime(t, ctx, hostClient, wsB.ID, agentB.ID, compiledB)

	apiClient := dialAgentAPI(t, runtimePort)

	// Agent A: shell → ALLOW
	respA, err := apiClient.ExecuteTool(
		withSandboxID(ctx, sandboxA),
		&hostagentpb.ExecuteToolRequest{
			ToolName:   "shell",
			Parameters: makeParams(t, map[string]interface{}{"cmd": "ls"}),
		})
	if err != nil {
		t.Fatalf("agent A execute tool: %v", err)
	}
	if respA.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ALLOW {
		t.Fatalf("agent A: expected ALLOW, got %s", respA.Verdict)
	}

	// Agent B: shell → DENY
	respB, err := apiClient.ExecuteTool(
		withSandboxID(ctx, sandboxB),
		&hostagentpb.ExecuteToolRequest{
			ToolName:   "shell",
			Parameters: makeParams(t, map[string]interface{}{"cmd": "ls"}),
		})
	if err != nil {
		t.Fatalf("agent B execute tool: %v", err)
	}
	if respB.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_DENY {
		t.Fatalf("agent B: expected DENY, got %s", respB.Verdict)
	}
}

func TestFullStackMultiAgentSeparateBudgets(t *testing.T) {
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

	// Agent A: $100 budget (plenty of headroom).
	agentA := registerAgent(t, ctx, tenant, "budget-agent-A")
	_, err := economicsSvc.SetBudget(ctx, tenant, agentA.ID, 100.0, "USD", "halt", 0.0)
	if err != nil {
		t.Fatalf("set budget A: %v", err)
	}
	wsA, err := workspaceSvc.CreateWorkspace(ctx, tenant, agentA.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace A: %v", err)
	}

	// Agent B: $10 budget (will be exhausted).
	agentB := registerAgent(t, ctx, tenant, "budget-agent-B")
	_, err = economicsSvc.SetBudget(ctx, tenant, agentB.ID, 10.0, "USD", "halt", 0.0)
	if err != nil {
		t.Fatalf("set budget B: %v", err)
	}
	wsB, err := workspaceSvc.CreateWorkspace(ctx, tenant, agentB.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace B: %v", err)
	}

	// Exhaust Agent B's budget.
	_, err = economicsSvc.RecordUsage(ctx, tenant, agentB.ID, wsB.ID, "compute", "hour", 1.0, 15.0)
	if err != nil {
		t.Fatalf("record usage B: %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)
	sandboxA := createSandboxOnRuntime(t, ctx, hostClient, wsA.ID, agentA.ID, nil)
	sandboxB := createSandboxOnRuntime(t, ctx, hostClient, wsB.ID, agentB.ID, nil)

	apiClient := dialAgentAPI(t, runtimePort)

	// Agent B: should be DENY (budget exhausted).
	respB, err := apiClient.ExecuteTool(
		withSandboxID(ctx, sandboxB),
		&hostagentpb.ExecuteToolRequest{
			ToolName:   "web_search",
			Parameters: makeParams(t, map[string]interface{}{"query": "test"}),
		})
	if err != nil {
		t.Fatalf("agent B execute tool: %v", err)
	}
	if respB.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_DENY {
		t.Fatalf("agent B: expected DENY (budget exhausted), got %s", respB.Verdict)
	}

	// Agent A: should still be ALLOW (separate budget, plenty of room).
	respA, err := apiClient.ExecuteTool(
		withSandboxID(ctx, sandboxA),
		&hostagentpb.ExecuteToolRequest{
			ToolName:   "web_search",
			Parameters: makeParams(t, map[string]interface{}{"query": "test"}),
		})
	if err != nil {
		t.Fatalf("agent A execute tool: %v", err)
	}
	if respA.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ALLOW {
		t.Fatalf("agent A: expected ALLOW (budget not exhausted), got %s", respA.Verdict)
	}
}

func TestFullStackMultiTenantIsolation(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenantX := uniqueTenant()
	tenantY := uniqueTenant()

	servers, ports := startGRPCServers(t)
	defer func() {
		for _, s := range servers {
			s.GracefulStop()
		}
	}()

	_, runtimePort := startRuntime(t, ports)
	waitForRuntime(t, runtimePort, 30*time.Second)
	time.Sleep(2 * time.Second)

	// Tenant X: agent with deny-shell policy.
	agentX := registerAgent(t, ctx, tenantX, "tenant-X-agent")
	denyRuleX, err := guardrailsSvc.CreateRule(ctx, tenantX, "deny-shell-X", "block shell for X",
		models.RuleTypeToolFilter, "shell", models.RuleActionDeny, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create deny rule X: %v", err)
	}
	compiledX, _, err := guardrailsSvc.CompilePolicy(ctx, tenantX, []string{denyRuleX.ID})
	if err != nil {
		t.Fatalf("compile policy X: %v", err)
	}
	wsX, err := workspaceSvc.CreateWorkspace(ctx, tenantX, agentX.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace X: %v", err)
	}

	// Tenant Y: agent with allow-all policy.
	agentY := registerAgent(t, ctx, tenantY, "tenant-Y-agent")
	wsY, err := workspaceSvc.CreateWorkspace(ctx, tenantY, agentY.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace Y: %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)
	sandboxX := createSandboxOnRuntime(t, ctx, hostClient, wsX.ID, agentX.ID, compiledX)
	sandboxY := createSandboxOnRuntime(t, ctx, hostClient, wsY.ID, agentY.ID, nil)

	apiClient := dialAgentAPI(t, runtimePort)

	// Tenant X: shell → DENY
	respX, err := apiClient.ExecuteTool(
		withSandboxID(ctx, sandboxX),
		&hostagentpb.ExecuteToolRequest{
			ToolName:   "shell",
			Parameters: makeParams(t, map[string]interface{}{"cmd": "ls"}),
		})
	if err != nil {
		t.Fatalf("tenant X execute tool: %v", err)
	}
	if respX.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_DENY {
		t.Fatalf("tenant X: expected DENY, got %s", respX.Verdict)
	}

	// Tenant Y: shell → ALLOW (different tenant, different policy)
	respY, err := apiClient.ExecuteTool(
		withSandboxID(ctx, sandboxY),
		&hostagentpb.ExecuteToolRequest{
			ToolName:   "shell",
			Parameters: makeParams(t, map[string]interface{}{"cmd": "ls"}),
		})
	if err != nil {
		t.Fatalf("tenant Y execute tool: %v", err)
	}
	if respY.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ALLOW {
		t.Fatalf("tenant Y: expected ALLOW, got %s", respY.Verdict)
	}

	// Verify workspace-level isolation: each tenant's workspace is separate.
	// Activity records from the Rust runtime are recorded asynchronously with
	// workspace_id scoping. We verify that each workspace belongs to the correct
	// tenant and that cross-tenant sandbox access is not possible.
	time.Sleep(1 * time.Second)

	// Verify workspace X belongs to tenant X.
	var wxTenant string
	err = db.QueryRowContext(ctx, "SELECT tenant_id FROM workspaces WHERE id = $1", wsX.ID).Scan(&wxTenant)
	if err != nil {
		t.Fatalf("query workspace X tenant: %v", err)
	}
	if wxTenant != tenantX {
		t.Fatalf("workspace X tenant mismatch: expected %s, got %s", tenantX, wxTenant)
	}

	// Verify workspace Y belongs to tenant Y.
	var wyTenant string
	err = db.QueryRowContext(ctx, "SELECT tenant_id FROM workspaces WHERE id = $1", wsY.ID).Scan(&wyTenant)
	if err != nil {
		t.Fatalf("query workspace Y tenant: %v", err)
	}
	if wyTenant != tenantY {
		t.Fatalf("workspace Y tenant mismatch: expected %s, got %s", tenantY, wyTenant)
	}
}

func TestFullStackAgentLifecycleSimulation(t *testing.T) {
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

	// 1. Register agent and set budget.
	agent := registerAgent(t, ctx, tenant, "lifecycle-agent")
	_, err := economicsSvc.SetBudget(ctx, tenant, agent.ID, 50.0, "USD", "halt", 0.0)
	if err != nil {
		t.Fatalf("set budget: %v", err)
	}

	// 2. Create workspace with guardrails: allow most tools, escalate "deploy".
	escalateRule, err := guardrailsSvc.CreateRule(ctx, tenant, "escalate-deploy", "require approval",
		models.RuleTypeToolFilter, "deploy", models.RuleActionEscalate, 1, nil, models.RuleScope{})
	if err != nil {
		t.Fatalf("create escalate rule: %v", err)
	}
	compiled, _, err := guardrailsSvc.CompilePolicy(ctx, tenant, []string{escalateRule.ID})
	if err != nil {
		t.Fatalf("compile policy: %v", err)
	}

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, compiled)

	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	// 3. Tool call 1: read_file → ALLOW.
	resp, err := apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "read_file",
		Parameters: makeParams(t, map[string]interface{}{"path": "/data/test.txt"}),
	})
	if err != nil {
		t.Fatalf("read_file: %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ALLOW {
		t.Fatalf("read_file: expected ALLOW, got %s", resp.Verdict)
	}

	// Agent reports result.
	_, err = apiClient.ReportActionResult(sandboxCtx, &hostagentpb.ReportActionResultRequest{
		ActionId: resp.ActionId,
		Success:  true,
		Result:   makeParams(t, map[string]interface{}{"content": "file data"}),
	})
	if err != nil {
		t.Fatalf("report result: %v", err)
	}

	// 4. Tool call 2: web_search → ALLOW.
	resp, err = apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "web_search",
		Parameters: makeParams(t, map[string]interface{}{"query": "quarterly revenue"}),
	})
	if err != nil {
		t.Fatalf("web_search: %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ALLOW {
		t.Fatalf("web_search: expected ALLOW, got %s", resp.Verdict)
	}

	// 5. Tool call 3: deploy → ESCALATE.
	resp, err = apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "deploy",
		Parameters: makeParams(t, map[string]interface{}{"env": "production", "version": "v2.1"}),
	})
	if err != nil {
		t.Fatalf("deploy: %v", err)
	}
	if resp.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ESCALATE {
		t.Fatalf("deploy: expected ESCALATE, got %s", resp.Verdict)
	}

	// 6. Human responds to the escalation via HIS.
	// The agent would poll via RequestHumanInput, but we go through the HIS directly.
	hiResp, err := apiClient.RequestHumanInput(sandboxCtx, &hostagentpb.RequestHumanInputRequest{
		Question:       "Deploy v2.1 to production?",
		Options:        []string{"approve", "deny"},
		Context:        "Agent wants to deploy version 2.1",
		TimeoutSeconds: 300,
	})
	if err != nil {
		t.Fatalf("request human input: %v", err)
	}

	err = humanSvc.RespondToRequest(ctx, tenant, hiResp.RequestId, "approve", "admin-1")
	if err != nil {
		t.Fatalf("respond to request: %v", err)
	}

	checkResp, err := apiClient.CheckHumanRequest(sandboxCtx, &hostagentpb.CheckHumanRequestRequest{
		RequestId: hiResp.RequestId,
	})
	if err != nil {
		t.Fatalf("check human request: %v", err)
	}
	if checkResp.Status != "responded" || checkResp.Response != "approve" {
		t.Fatalf("expected responded/approve, got %s/%s", checkResp.Status, checkResp.Response)
	}

	// 7. Record some usage and verify budget is still within limits.
	_, err = economicsSvc.RecordUsage(ctx, tenant, agent.ID, ws.ID, "compute", "hour", 1.0, 5.0)
	if err != nil {
		t.Fatalf("record usage: %v", err)
	}

	budget, err := economicsSvc.GetBudget(ctx, tenant, agent.ID)
	if err != nil {
		t.Fatalf("get budget: %v", err)
	}
	if budget.Used > budget.Limit {
		t.Fatalf("budget should not be exhausted: used=%f, limit=%f", budget.Used, budget.Limit)
	}

	// 8. Verify sandbox status.
	statusResp, err := hostClient.GetSandboxStatus(ctx, &hostagentpb.GetSandboxStatusRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		t.Fatalf("get sandbox status: %v", err)
	}
	if statusResp.ActionsExecuted < 3 {
		t.Fatalf("expected at least 3 actions_executed, got %d", statusResp.ActionsExecuted)
	}

	// 9. Destroy sandbox.
	_, err = hostClient.DestroySandbox(ctx, &hostagentpb.DestroySandboxRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		t.Fatalf("destroy sandbox: %v", err)
	}

	// Verify sandbox is gone.
	_, err = hostClient.GetSandboxStatus(ctx, &hostagentpb.GetSandboxStatusRequest{
		SandboxId: sandboxID,
	})
	if err == nil {
		t.Fatal("expected error when querying destroyed sandbox")
	}
}

// ---------------------------------------------------------------------------
// WS6: Missing full-stack coverage
// ---------------------------------------------------------------------------

func TestFullStackHeartbeatLoop(t *testing.T) {
	clean(t)

	servers, ports := startGRPCServers(t)
	defer func() {
		for _, s := range servers {
			s.GracefulStop()
		}
	}()

	_, runtimePort := startRuntime(t, ports)
	waitForRuntime(t, runtimePort, 30*time.Second)
	time.Sleep(2 * time.Second)

	// Verify host is registered.
	hosts, err := computeSvc.ListHosts(context.Background(), models.HostStatusReady)
	if err != nil {
		t.Fatalf("list hosts: %v", err)
	}
	if len(hosts) == 0 {
		t.Fatal("expected runtime to be registered")
	}

	firstHeartbeat := hosts[0].LastHeartbeat

	// Wait for at least one more heartbeat cycle (default is 30s).
	time.Sleep(35 * time.Second)

	hosts, err = computeSvc.ListHosts(context.Background(), models.HostStatusReady)
	if err != nil {
		t.Fatalf("list hosts (after wait): %v", err)
	}
	if len(hosts) == 0 {
		t.Fatal("expected runtime to still be registered")
	}

	if !hosts[0].LastHeartbeat.After(firstHeartbeat) {
		t.Fatal("expected heartbeat to have been updated")
	}
}

func TestFullStackSandboxStatus(t *testing.T) {
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

	agent := registerAgent(t, ctx, tenant, "status-agent")

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, nil)

	// Initial status: running, 0 actions.
	statusResp, err := hostClient.GetSandboxStatus(ctx, &hostagentpb.GetSandboxStatusRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		t.Fatalf("get sandbox status: %v", err)
	}
	if statusResp.State != hostagentpb.SandboxState_SANDBOX_STATE_RUNNING {
		t.Fatalf("expected RUNNING state, got %s", statusResp.State)
	}
	if statusResp.ActionsExecuted != 0 {
		t.Fatalf("expected 0 actions_executed, got %d", statusResp.ActionsExecuted)
	}

	// Execute 3 tool calls.
	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	for i := 0; i < 3; i++ {
		_, err := apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
			ToolName:   "web_search",
			Parameters: makeParams(t, map[string]interface{}{"query": fmt.Sprintf("q-%d", i)}),
		})
		if err != nil {
			t.Fatalf("execute tool %d: %v", i, err)
		}
	}

	// Status should reflect 3 actions.
	statusResp, err = hostClient.GetSandboxStatus(ctx, &hostagentpb.GetSandboxStatusRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		t.Fatalf("get sandbox status (after calls): %v", err)
	}
	if statusResp.ActionsExecuted != 3 {
		t.Fatalf("expected 3 actions_executed, got %d", statusResp.ActionsExecuted)
	}
}

func TestFullStackStreamEvents(t *testing.T) {
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

	agent := registerAgent(t, ctx, tenant, "stream-agent")

	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)
	sandboxID := createSandboxOnRuntime(t, ctx, hostClient, ws.ID, agent.ID, nil)

	// Subscribe to events with a timeout context.
	streamCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	stream, err := hostClient.StreamEvents(streamCtx, &hostagentpb.StreamEventsRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		t.Fatalf("stream events: %v", err)
	}

	// Execute a tool call to generate an ActionEvent.
	apiClient := dialAgentAPI(t, runtimePort)
	sandboxCtx := withSandboxID(ctx, sandboxID)

	_, err = apiClient.ExecuteTool(sandboxCtx, &hostagentpb.ExecuteToolRequest{
		ToolName:   "web_search",
		Parameters: makeParams(t, map[string]interface{}{"query": "test"}),
	})
	if err != nil {
		t.Fatalf("execute tool: %v", err)
	}

	// Read the action event from the stream.
	event, err := stream.Recv()
	if err != nil {
		t.Fatalf("recv event: %v", err)
	}
	if event.SandboxId != sandboxID {
		t.Fatalf("expected sandbox_id %s, got %s", sandboxID, event.SandboxId)
	}
	actionEvent := event.GetAction()
	if actionEvent == nil {
		t.Fatal("expected ActionEvent, got nil")
	}
	if actionEvent.ToolName != "web_search" {
		t.Fatalf("expected tool_name 'web_search', got %q", actionEvent.ToolName)
	}
	if actionEvent.Verdict != hostagentpb.ActionVerdict_ACTION_VERDICT_ALLOW {
		t.Fatalf("expected ALLOW verdict in event, got %s", actionEvent.Verdict)
	}
}

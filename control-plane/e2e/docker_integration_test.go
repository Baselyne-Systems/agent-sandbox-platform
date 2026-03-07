//go:build docker_integration

// Docker integration tests — require Docker daemon and ENABLE_DOCKER=true.
//
// Run with:
//
//	RUNTIME_BINARY=../runtime/target/release/host-agent \
//	ENABLE_DOCKER=true \
//	go test -v -tags docker_integration -run TestDocker ./e2e/
package e2e

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	hostagentpb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/host_agent/v1"
)

// requireDocker skips the test if Docker is not available or ENABLE_DOCKER != true.
func requireDocker(t *testing.T) {
	t.Helper()
	if os.Getenv("ENABLE_DOCKER") != "true" {
		t.Skip("ENABLE_DOCKER not set to true — skipping Docker integration test")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker binary not found in PATH — skipping")
	}
	// Quick connectivity check.
	out, err := exec.Command("docker", "info").CombinedOutput()
	if err != nil {
		t.Skipf("docker not available: %s", string(out))
	}
}

// dockerContainerExists checks if a container with the given name or ID exists.
func dockerContainerExists(nameOrID string) bool {
	out, err := exec.Command("docker", "inspect", nameOrID).CombinedOutput()
	return err == nil && len(out) > 2 // non-empty JSON array
}

// dockerContainerEnv returns the environment variables of a running container.
func dockerContainerEnv(t *testing.T, nameOrID string) map[string]string {
	t.Helper()
	out, err := exec.Command("docker", "inspect", "--format", "{{range .Config.Env}}{{println .}}{{end}}", nameOrID).CombinedOutput()
	if err != nil {
		t.Fatalf("docker inspect env: %v — %s", err, string(out))
	}
	env := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}

// startRuntimeWithDocker starts the runtime with ENABLE_DOCKER=true.
func startRuntimeWithDocker(t *testing.T, ports fullStackPorts) (*exec.Cmd, int) {
	t.Helper()
	requireDocker(t)

	binary := runtimeBinaryPath()
	if binary == "" {
		t.Skip("RUNTIME_BINARY not set")
	}

	// Find free port.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	runtimePort := lis.Addr().(*net.TCPAddr).Port
	lis.Close()

	cmd := exec.Command(binary)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GRPC_PORT=%d", runtimePort),
		"ADVERTISE_ADDRESS=127.0.0.1",
		"ENABLE_DOCKER=true",
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

func TestDockerContainerLifecycle(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	servers, ports := startGRPCServers(t)
	defer func() {
		for _, s := range servers {
			s.GracefulStop()
		}
	}()

	_, runtimePort := startRuntimeWithDocker(t, ports)
	waitForRuntime(t, runtimePort, 30*time.Second)
	time.Sleep(2 * time.Second)

	agent := registerAgent(t, ctx, tenant, "docker-lifecycle-agent")
	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)

	// Create sandbox with a container image.
	resp, err := hostClient.CreateSandbox(ctx, &hostagentpb.CreateSandboxRequest{
		WorkspaceId: ws.ID,
		AgentId:     agent.ID,
		Spec: &hostagentpb.SandboxSpec{
			MemoryMb:       512,
			CpuMillicores:  500,
			DiskMb:         1024,
			ContainerImage: "alpine:latest",
		},
		CompiledGuardrails: emptyPolicy,
	})
	if err != nil {
		t.Fatalf("create sandbox: %v", err)
	}
	sandboxID := resp.SandboxId

	// Give container time to start.
	time.Sleep(3 * time.Second)

	// Verify container exists and has expected env vars.
	env := dockerContainerEnv(t, sandboxID)
	if env["BULKHEAD_SANDBOX_ID"] != sandboxID {
		t.Fatalf("expected BULKHEAD_SANDBOX_ID=%s in container env, got %q", sandboxID, env["BULKHEAD_SANDBOX_ID"])
	}
	if env["BULKHEAD_ENDPOINT"] == "" {
		t.Fatal("expected BULKHEAD_ENDPOINT to be set in container env")
	}

	// Destroy sandbox → container should be removed.
	_, err = hostClient.DestroySandbox(ctx, &hostagentpb.DestroySandboxRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		t.Fatalf("destroy sandbox: %v", err)
	}

	time.Sleep(2 * time.Second)

	if dockerContainerExists(sandboxID) {
		t.Fatal("container should have been removed after sandbox destroy")
	}
}

func TestDockerIsolationTierSecurity(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	servers, ports := startGRPCServers(t)
	defer func() {
		for _, s := range servers {
			s.GracefulStop()
		}
	}()

	_, runtimePort := startRuntimeWithDocker(t, ports)
	waitForRuntime(t, runtimePort, 30*time.Second)
	time.Sleep(2 * time.Second)

	hostClient := dialHostAgentService(t, runtimePort)

	// Hardened sandbox.
	agentH := registerAgent(t, ctx, tenant, "hardened-agent")
	wsH, err := workspaceSvc.CreateWorkspace(ctx, tenant, agentH.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace (hardened): %v", err)
	}

	respH, err := hostClient.CreateSandbox(ctx, &hostagentpb.CreateSandboxRequest{
		WorkspaceId: wsH.ID,
		AgentId:     agentH.ID,
		Spec: &hostagentpb.SandboxSpec{
			MemoryMb:       512,
			CpuMillicores:  500,
			DiskMb:         1024,
			ContainerImage: "alpine:latest",
			IsolationTier:  "hardened",
		},
		CompiledGuardrails: emptyPolicy,
	})
	if err != nil {
		t.Fatalf("create hardened sandbox: %v", err)
	}
	time.Sleep(3 * time.Second)

	// Inspect Docker container for security settings.
	out, err := exec.Command("docker", "inspect", "--format",
		"ReadonlyRootfs={{.HostConfig.ReadonlyRootfs}} SecurityOpt={{.HostConfig.SecurityOpt}}",
		respH.SandboxId).CombinedOutput()
	if err != nil {
		t.Fatalf("docker inspect hardened: %v — %s", err, string(out))
	}
	inspectStr := string(out)
	if !strings.Contains(inspectStr, "ReadonlyRootfs=true") {
		t.Fatalf("hardened sandbox should have ReadonlyRootfs=true, got: %s", inspectStr)
	}
	if !strings.Contains(inspectStr, "no-new-privileges") {
		t.Fatalf("hardened sandbox should have no-new-privileges, got: %s", inspectStr)
	}

	// Cleanup.
	_, _ = hostClient.DestroySandbox(ctx, &hostagentpb.DestroySandboxRequest{SandboxId: respH.SandboxId})

	// Standard sandbox.
	agentS := registerAgent(t, ctx, tenant, "standard-agent")
	wsS, err := workspaceSvc.CreateWorkspace(ctx, tenant, agentS.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace (standard): %v", err)
	}

	respS, err := hostClient.CreateSandbox(ctx, &hostagentpb.CreateSandboxRequest{
		WorkspaceId: wsS.ID,
		AgentId:     agentS.ID,
		Spec: &hostagentpb.SandboxSpec{
			MemoryMb:       512,
			CpuMillicores:  500,
			DiskMb:         1024,
			ContainerImage: "alpine:latest",
			IsolationTier:  "standard",
		},
		CompiledGuardrails: emptyPolicy,
	})
	if err != nil {
		t.Fatalf("create standard sandbox: %v", err)
	}
	time.Sleep(3 * time.Second)

	out, err = exec.Command("docker", "inspect", "--format",
		"ReadonlyRootfs={{.HostConfig.ReadonlyRootfs}}",
		respS.SandboxId).CombinedOutput()
	if err != nil {
		t.Fatalf("docker inspect standard: %v — %s", err, string(out))
	}
	if strings.Contains(string(out), "ReadonlyRootfs=true") {
		t.Fatalf("standard sandbox should NOT have ReadonlyRootfs=true, got: %s", string(out))
	}

	_, _ = hostClient.DestroySandbox(ctx, &hostagentpb.DestroySandboxRequest{SandboxId: respS.SandboxId})
}

func TestDockerEgressEnforcement(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	servers, ports := startGRPCServers(t)
	defer func() {
		for _, s := range servers {
			s.GracefulStop()
		}
	}()

	_, runtimePort := startRuntimeWithDocker(t, ports)
	waitForRuntime(t, runtimePort, 30*time.Second)
	time.Sleep(2 * time.Second)

	agent := registerAgent(t, ctx, tenant, "egress-agent")
	ws, err := workspaceSvc.CreateWorkspace(ctx, tenant, agent.ID, "", nil)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	hostClient := dialHostAgentService(t, runtimePort)

	resp, err := hostClient.CreateSandbox(ctx, &hostagentpb.CreateSandboxRequest{
		WorkspaceId: ws.ID,
		AgentId:     agent.ID,
		Spec: &hostagentpb.SandboxSpec{
			MemoryMb:        512,
			CpuMillicores:   500,
			DiskMb:          1024,
			ContainerImage:  "alpine:latest",
			EgressAllowlist: []string{"1.1.1.1"},
		},
		CompiledGuardrails: emptyPolicy,
	})
	if err != nil {
		t.Fatalf("create sandbox: %v", err)
	}
	sandboxID := resp.SandboxId
	time.Sleep(3 * time.Second)

	// Verify iptables chain exists.
	// Rust truncates sandbox_id to first 12 chars: chain_name = "BH-{id[..12]}"
	idPrefix := sandboxID
	if len(idPrefix) > 12 {
		idPrefix = idPrefix[:12]
	}
	chainName := fmt.Sprintf("BH-%s", idPrefix)
	out, err := exec.Command("iptables", "-L", chainName, "-n").CombinedOutput()
	if err != nil {
		t.Fatalf("iptables chain %q not found for sandbox %s: %v — %s", chainName, sandboxID, err, string(out))
	}

	iptablesOutput := string(out)
	if !strings.Contains(iptablesOutput, "1.1.1.1") {
		t.Fatalf("expected ACCEPT rule for 1.1.1.1 in iptables chain, got:\n%s", iptablesOutput)
	}
	if !strings.Contains(iptablesOutput, "DROP") {
		t.Fatalf("expected default DROP rule in iptables chain, got:\n%s", iptablesOutput)
	}

	// Destroy sandbox → chain should be cleaned up.
	_, err = hostClient.DestroySandbox(ctx, &hostagentpb.DestroySandboxRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		t.Fatalf("destroy sandbox: %v", err)
	}
	time.Sleep(2 * time.Second)

	_, err = exec.Command("iptables", "-L", chainName, "-n").CombinedOutput()
	if err == nil {
		t.Fatal("iptables chain should have been cleaned up after sandbox destroy")
	}
}

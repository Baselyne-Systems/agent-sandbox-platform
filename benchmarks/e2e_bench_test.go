//go:build e2e

// Package benchmarks contains end-to-end gRPC benchmarks that run against a
// live Bulkhead cluster (typically deployed via Kind). Build tag "e2e" gates
// these tests so they are excluded from normal CI runs.
//
// Run: go test -tags e2e -bench=. -benchtime=5s -count=3 ./...
package benchmarks

import (
	"context"
	"fmt"
	"os"
	"testing"

	activitypb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/activity/v1"
	computepb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/compute/v1"
	economicspb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/economics/v1"
	governancepb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/governance/v1"
	guardrailspb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/guardrails/v1"
	humanpb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/human/v1"
	identitypb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/identity/v1"
	taskpb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/task/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ---------------------------------------------------------------------------
// Service addresses (override via environment variables)
// ---------------------------------------------------------------------------

// Consolidated binary addresses (services that share a binary share a port).
// control-plane binary (Identity, Task, Workspace, Compute): 50060
// policy binary (Guardrails, Governance):                     50062
// observability binary (Activity, Economics, Human):           50065
var (
	identityAddr   = envOrDefault("IDENTITY_ADDR", "localhost:50060")
	workspaceAddr  = envOrDefault("WORKSPACE_ADDR", "localhost:50060")
	guardrailsAddr = envOrDefault("GUARDRAILS_ADDR", "localhost:50062")
	humanAddr      = envOrDefault("HUMAN_ADDR", "localhost:50065")
	governanceAddr = envOrDefault("GOVERNANCE_ADDR", "localhost:50062")
	activityAddr   = envOrDefault("ACTIVITY_ADDR", "localhost:50065")
	economicsAddr  = envOrDefault("ECONOMICS_ADDR", "localhost:50065")
	computeAddr    = envOrDefault("COMPUTE_ADDR", "localhost:50060")
	taskAddr       = envOrDefault("TASK_ADDR", "localhost:50060")
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func dialGRPC(b *testing.B, addr string) *grpc.ClientConn {
	b.Helper()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		b.Fatalf("dial %s: %v", addr, err)
	}
	b.Cleanup(func() { conn.Close() })
	return conn
}

// ---------------------------------------------------------------------------
// 1. BenchmarkE2E_Identity_RegisterAgent
// ---------------------------------------------------------------------------

func BenchmarkE2E_Identity_RegisterAgent(b *testing.B) {
	b.ReportAllocs()
	conn := dialGRPC(b, identityAddr)
	client := identitypb.NewIdentityServiceClient(conn)
	ctx := context.Background()

	for b.Loop() {
		_, err := client.RegisterAgent(ctx, &identitypb.RegisterAgentRequest{
			Name:    "bench-agent",
			OwnerId: "bench-owner",
			Purpose: "e2e benchmark",
		})
		if err != nil {
			b.Fatalf("RegisterAgent: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// 2. BenchmarkE2E_Identity_GetAgent
// ---------------------------------------------------------------------------

func BenchmarkE2E_Identity_GetAgent(b *testing.B) {
	b.ReportAllocs()
	conn := dialGRPC(b, identityAddr)
	client := identitypb.NewIdentityServiceClient(conn)
	ctx := context.Background()

	// Setup: register an agent to retrieve in the hot loop.
	resp, err := client.RegisterAgent(ctx, &identitypb.RegisterAgentRequest{
		Name:    "bench-get-agent",
		OwnerId: "bench-owner",
		Purpose: "e2e benchmark get-agent",
	})
	if err != nil {
		b.Fatalf("setup RegisterAgent: %v", err)
	}
	agentID := resp.GetAgent().GetAgentId()

	b.ResetTimer()
	for b.Loop() {
		_, err := client.GetAgent(ctx, &identitypb.GetAgentRequest{
			AgentId: agentID,
		})
		if err != nil {
			b.Fatalf("GetAgent: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// 3. BenchmarkE2E_Task_CreateTask
// ---------------------------------------------------------------------------

func BenchmarkE2E_Task_CreateTask(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	// Setup: register an agent first.
	idConn := dialGRPC(b, identityAddr)
	idClient := identitypb.NewIdentityServiceClient(idConn)
	agentResp, err := idClient.RegisterAgent(ctx, &identitypb.RegisterAgentRequest{
		Name:    "bench-task-agent",
		OwnerId: "bench-owner",
		Purpose: "e2e benchmark create-task",
	})
	if err != nil {
		b.Fatalf("setup RegisterAgent: %v", err)
	}
	agentID := agentResp.GetAgent().GetAgentId()

	taskConn := dialGRPC(b, taskAddr)
	taskClient := taskpb.NewTaskServiceClient(taskConn)

	b.ResetTimer()
	for b.Loop() {
		_, err := taskClient.CreateTask(ctx, &taskpb.CreateTaskRequest{
			AgentId: agentID,
			Goal:    "benchmark task goal",
		})
		if err != nil {
			b.Fatalf("CreateTask: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// 4. BenchmarkE2E_Compute_RegisterHost
// ---------------------------------------------------------------------------

func BenchmarkE2E_Compute_RegisterHost(b *testing.B) {
	b.ReportAllocs()
	conn := dialGRPC(b, computeAddr)
	client := computepb.NewComputePlaneServiceClient(conn)
	ctx := context.Background()

	for b.Loop() {
		_, err := client.RegisterHost(ctx, &computepb.RegisterHostRequest{
			Address: "10.0.0.1:8080",
			TotalResources: &computepb.HostResources{
				MemoryMb:      8192,
				CpuMillicores: 4000,
				DiskMb:        102400,
			},
			SupportedTiers: []string{"standard"},
		})
		if err != nil {
			b.Fatalf("RegisterHost: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. BenchmarkE2E_Guardrails_CreateRule
// ---------------------------------------------------------------------------

func BenchmarkE2E_Guardrails_CreateRule(b *testing.B) {
	b.ReportAllocs()
	conn := dialGRPC(b, guardrailsAddr)
	client := guardrailspb.NewGuardrailsServiceClient(conn)
	ctx := context.Background()

	for b.Loop() {
		_, err := client.CreateRule(ctx, &guardrailspb.CreateRuleRequest{
			Name:        "bench-rule",
			Description: "benchmark guardrail rule",
			Type:        guardrailspb.RuleType_RULE_TYPE_TOOL_FILTER,
			Condition:   "tool_name == 'exec'",
			Action:      guardrailspb.RuleAction_RULE_ACTION_DENY,
			Priority:    100,
		})
		if err != nil {
			b.Fatalf("CreateRule: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. BenchmarkE2E_Activity_RecordAction
// ---------------------------------------------------------------------------

func BenchmarkE2E_Activity_RecordAction(b *testing.B) {
	b.ReportAllocs()
	conn := dialGRPC(b, activityAddr)
	client := activitypb.NewActivityServiceClient(conn)
	ctx := context.Background()

	for b.Loop() {
		_, err := client.RecordAction(ctx, &activitypb.RecordActionRequest{
			Record: &activitypb.ActionRecord{
				WorkspaceId: "ws-bench-001",
				AgentId:     "agent-bench-001",
				TaskId:      "task-bench-001",
				ToolName:    "file_read",
				Outcome:     activitypb.ActionOutcome_ACTION_OUTCOME_ALLOWED,
			},
		})
		if err != nil {
			b.Fatalf("RecordAction: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// 7. BenchmarkE2E_Economics_CheckBudget
// ---------------------------------------------------------------------------

func BenchmarkE2E_Economics_CheckBudget(b *testing.B) {
	b.ReportAllocs()
	conn := dialGRPC(b, economicsAddr)
	client := economicspb.NewEconomicsServiceClient(conn)
	ctx := context.Background()

	// Setup: set a budget so CheckBudget has something to evaluate.
	_, err := client.SetBudget(ctx, &economicspb.SetBudgetRequest{
		AgentId:          "agent-budget-bench",
		Limit:            1000.0,
		Currency:         "USD",
		OnExceeded:       economicspb.OnExceededAction_ON_EXCEEDED_ACTION_WARN,
		WarningThreshold: 0.8,
	})
	if err != nil {
		b.Fatalf("setup SetBudget: %v", err)
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := client.CheckBudget(ctx, &economicspb.CheckBudgetRequest{
			AgentId:       "agent-budget-bench",
			EstimatedCost: 1.50,
		})
		if err != nil {
			b.Fatalf("CheckBudget: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// 8. BenchmarkE2E_Governance_ClassifyData
// ---------------------------------------------------------------------------

func BenchmarkE2E_Governance_ClassifyData(b *testing.B) {
	b.ReportAllocs()
	conn := dialGRPC(b, governanceAddr)
	client := governancepb.NewDataGovernanceServiceClient(conn)
	ctx := context.Background()

	payload := []byte("SSN: 123-45-6789, credit card: 4111-1111-1111-1111")

	for b.Loop() {
		_, err := client.ClassifyData(ctx, &governancepb.ClassifyDataRequest{
			Content:     payload,
			ContentType: "text/plain",
		})
		if err != nil {
			b.Fatalf("ClassifyData: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// 9. BenchmarkE2E_Human_CreateRequest
// ---------------------------------------------------------------------------

func BenchmarkE2E_Human_CreateRequest(b *testing.B) {
	b.ReportAllocs()
	conn := dialGRPC(b, humanAddr)
	client := humanpb.NewHumanInteractionServiceClient(conn)
	ctx := context.Background()

	for b.Loop() {
		_, err := client.CreateRequest(ctx, &humanpb.CreateHumanRequestRequest{
			WorkspaceId:    "ws-bench-001",
			AgentId:        "agent-bench-001",
			Question:       "May I proceed with file deletion?",
			Options:        []string{"yes", "no"},
			Context:        "benchmark test",
			TimeoutSeconds: 300,
			Type:           humanpb.HumanRequestType_HUMAN_REQUEST_TYPE_APPROVAL,
			Urgency:        humanpb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_NORMAL,
		})
		if err != nil {
			b.Fatalf("CreateRequest: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// 10. BenchmarkE2E_FullWorkflow
//
// End-to-end workflow exercising the critical path:
//   1. Register agent          (identity)
//   2. Create guardrail rule   (guardrails)
//   3. Compile policy          (guardrails)
//   4. Register host           (compute)
//   5. Create task             (task)
//   6. Record action           (activity)
//   7. Check budget            (economics)
// ---------------------------------------------------------------------------

func BenchmarkE2E_FullWorkflow(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	idConn := dialGRPC(b, identityAddr)
	grConn := dialGRPC(b, guardrailsAddr)
	cpConn := dialGRPC(b, computeAddr)
	tkConn := dialGRPC(b, taskAddr)
	acConn := dialGRPC(b, activityAddr)
	ecConn := dialGRPC(b, economicsAddr)

	idClient := identitypb.NewIdentityServiceClient(idConn)
	grClient := guardrailspb.NewGuardrailsServiceClient(grConn)
	cpClient := computepb.NewComputePlaneServiceClient(cpConn)
	tkClient := taskpb.NewTaskServiceClient(tkConn)
	acClient := activitypb.NewActivityServiceClient(acConn)
	ecClient := economicspb.NewEconomicsServiceClient(ecConn)

	for b.Loop() {
		// Step 1: Register a new agent identity.
		agentResp, err := idClient.RegisterAgent(ctx, &identitypb.RegisterAgentRequest{
			Name:    "workflow-agent",
			OwnerId: "workflow-owner",
			Purpose: "full workflow benchmark",
		})
		if err != nil {
			b.Fatalf("step1 RegisterAgent: %v", err)
		}
		agentID := agentResp.GetAgent().GetAgentId()

		// Step 2: Create a guardrail rule.
		ruleResp, err := grClient.CreateRule(ctx, &guardrailspb.CreateRuleRequest{
			Name:        "workflow-rule",
			Description: "benchmark workflow rule",
			Type:        guardrailspb.RuleType_RULE_TYPE_TOOL_FILTER,
			Condition:   "tool_name == 'exec'",
			Action:      guardrailspb.RuleAction_RULE_ACTION_DENY,
			Priority:    100,
		})
		if err != nil {
			b.Fatalf("step2 CreateRule: %v", err)
		}
		ruleID := ruleResp.GetRule().GetRuleId()

		// Step 3: Compile the guardrail rule into a policy.
		_, err = grClient.CompilePolicy(ctx, &guardrailspb.CompilePolicyRequest{
			RuleIds: []string{ruleID},
		})
		if err != nil {
			b.Fatalf("step3 CompilePolicy: %v", err)
		}

		// Step 4: Register a compute host.
		_, err = cpClient.RegisterHost(ctx, &computepb.RegisterHostRequest{
			Address: "10.0.0.1:8080",
			TotalResources: &computepb.HostResources{
				MemoryMb:      8192,
				CpuMillicores: 4000,
				DiskMb:        102400,
			},
			SupportedTiers: []string{"standard"},
		})
		if err != nil {
			b.Fatalf("step4 RegisterHost: %v", err)
		}

		// Step 5: Create a task for the agent.
		taskResp, err := tkClient.CreateTask(ctx, &taskpb.CreateTaskRequest{
			AgentId: agentID,
			Goal:    "workflow benchmark task",
		})
		if err != nil {
			b.Fatalf("step5 CreateTask: %v", err)
		}
		taskID := taskResp.GetTask().GetTaskId()

		// Step 6: Record an action in the activity log.
		_, err = acClient.RecordAction(ctx, &activitypb.RecordActionRequest{
			Record: &activitypb.ActionRecord{
				AgentId:  agentID,
				TaskId:   taskID,
				ToolName: "file_read",
				Outcome:  activitypb.ActionOutcome_ACTION_OUTCOME_ALLOWED,
			},
		})
		if err != nil {
			b.Fatalf("step6 RecordAction: %v", err)
		}

		// Step 7: Check the budget for the agent.
		_, err = ecClient.CheckBudget(ctx, &economicspb.CheckBudgetRequest{
			AgentId:       agentID,
			EstimatedCost: 0.50,
		})
		if err != nil {
			b.Fatalf("step7 CheckBudget: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// 11. BenchmarkE2E_Identity_RegisterAgent_Parallel
// ---------------------------------------------------------------------------

func BenchmarkE2E_Identity_RegisterAgent_Parallel(b *testing.B) {
	b.ReportAllocs()
	conn := dialGRPC(b, identityAddr)
	client := identitypb.NewIdentityServiceClient(conn)

	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		i := 0
		for pb.Next() {
			_, err := client.RegisterAgent(ctx, &identitypb.RegisterAgentRequest{
				Name:    fmt.Sprintf("parallel-agent-%d", i),
				OwnerId: "bench-owner",
				Purpose: "parallel benchmark",
			})
			if err != nil {
				b.Errorf("RegisterAgent: %v", err)
				return
			}
			i++
		}
	})
}

// ---------------------------------------------------------------------------
// 12. BenchmarkE2E_Activity_RecordAction_Parallel
// ---------------------------------------------------------------------------

func BenchmarkE2E_Activity_RecordAction_Parallel(b *testing.B) {
	b.ReportAllocs()
	conn := dialGRPC(b, activityAddr)
	client := activitypb.NewActivityServiceClient(conn)

	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		i := 0
		for pb.Next() {
			_, err := client.RecordAction(ctx, &activitypb.RecordActionRequest{
				Record: &activitypb.ActionRecord{
					WorkspaceId: "ws-parallel-001",
					AgentId:     fmt.Sprintf("agent-par-%d", i),
					TaskId:      "task-parallel-001",
					ToolName:    "file_read",
					Outcome:     activitypb.ActionOutcome_ACTION_OUTCOME_ALLOWED,
				},
			})
			if err != nil {
				b.Errorf("RecordAction: %v", err)
				return
			}
			i++
		}
	})
}

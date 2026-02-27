package workspace

import (
	"context"
	"fmt"
	"testing"

	"github.com/baselyne/agent-sandbox-platform/control-plane/internal/models"
)

func BenchmarkCreateWorkspace(b *testing.B) {
	svc := newTestService(newMockRepo())
	ctx := context.Background()

	for b.Loop() {
		_, err := svc.CreateWorkspace(ctx, "agent-1", "task-1", nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetWorkspace(b *testing.B) {
	svc := newTestService(newMockRepo())
	ctx := context.Background()
	ws, _ := svc.CreateWorkspace(ctx, "agent-1", "task-1", nil)

	for b.Loop() {
		_, err := svc.GetWorkspace(ctx, ws.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListWorkspaces_Scaling(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			svc := newTestService(newMockRepo())
			ctx := context.Background()

			for i := 0; i < n; i++ {
				svc.CreateWorkspace(ctx, "agent-1", fmt.Sprintf("t%d", i), nil)
			}

			b.ResetTimer()
			for b.Loop() {
				_, _, err := svc.ListWorkspaces(ctx, "", "", 50, "")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkTerminateWorkspace(b *testing.B) {
	ctx := context.Background()

	for b.Loop() {
		b.StopTimer()
		svc := newTestService(newMockRepo())
		ws, _ := svc.CreateWorkspace(ctx, "agent-1", "task-1", nil)
		b.StartTimer()

		err := svc.TerminateWorkspace(ctx, ws.ID, "done")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateWorkspace_WithSpec(b *testing.B) {
	svc := newTestService(newMockRepo())
	ctx := context.Background()
	spec := &models.WorkspaceSpec{
		MemoryMb:          2048,
		CpuMillicores:     2000,
		DiskMb:            4096,
		MaxDurationSecs:   7200,
		AllowedTools:      []string{"shell", "http", "file"},
		GuardrailPolicyID: "policy-1",
		EnvVars:           map[string]string{"ENV": "prod", "DEBUG": "false"},
	}

	for b.Loop() {
		_, err := svc.CreateWorkspace(ctx, "agent-1", "task-1", spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

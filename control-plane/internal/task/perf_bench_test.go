package task

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync/atomic"
	"testing"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
)

// ---------------------------------------------------------------------------
// 1. BenchmarkCreateTask_Parallel
// ---------------------------------------------------------------------------

func BenchmarkCreateTask_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()
	var counter atomic.Int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := counter.Add(1)
			_, err := svc.CreateTask(ctx, "tenant-1", fmt.Sprintf("agent-%d", n), "parallel goal", nil, "", nil, nil, 0, nil, nil)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 2. BenchmarkGetTask_Parallel
// ---------------------------------------------------------------------------

func BenchmarkGetTask_Parallel(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	const taskCount = 1000
	taskIDs := make([]string, taskCount)
	for i := 0; i < taskCount; i++ {
		t, err := svc.CreateTask(ctx, "tenant-1", "agent-1", fmt.Sprintf("goal-%d", i), nil, "", nil, nil, 0, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
		taskIDs[i] = t.ID
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, err := svc.GetTask(ctx, "tenant-1", taskIDs[i%taskCount])
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// ---------------------------------------------------------------------------
// 3. BenchmarkCreateTask_WithFullConfig
// ---------------------------------------------------------------------------

func BenchmarkCreateTask_WithFullConfig(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	wsConfig := &models.TaskWorkspaceConfig{
		IsolationTier:   "standard",
		MemoryMb:        1024,
		CpuMillicores:   2000,
		DiskMb:          4096,
		AllowedTools:    []string{"bash", "curl", "python", "node"},
		EnvVars:         map[string]string{"ENV": "prod", "DEBUG": "false", "LOG_LEVEL": "info"},
		EgressAllowlist: []string{"api.example.com", "cdn.example.com"},
	}
	hiConfig := &models.TaskHumanInteractionConfig{
		EscalationTargets: []string{"admin@example.com", "ops@example.com"},
		TimeoutSecs:       600,
		TimeoutAction:     "escalate",
	}
	budgetConfig := &models.TaskBudgetConfig{
		MaxCost:          50.0,
		WarningThreshold: 40.0,
		OnExceeded:       "halt",
		Currency:         "USD",
	}
	input := map[string]string{"source": "s3://bucket/input", "format": "csv"}
	labels := map[string]string{"env": "prod", "priority": "high", "team": "platform"}

	for b.Loop() {
		_, err := svc.CreateTask(ctx, "tenant-1", "agent-1", "full config goal",
			wsConfig, "policy-1", hiConfig, budgetConfig, 300, input, labels)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 4. BenchmarkCreateTask_LargeInput
// ---------------------------------------------------------------------------

func BenchmarkCreateTask_LargeInput(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	input := make(map[string]string, 50)
	for i := 0; i < 50; i++ {
		input[fmt.Sprintf("key-%d", i)] = fmt.Sprintf("value-%d-with-some-payload-data", i)
	}

	for b.Loop() {
		_, err := svc.CreateTask(ctx, "tenant-1", "agent-1", "large input goal",
			nil, "", nil, nil, 0, input, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. BenchmarkCreateTask_LargeLabels
// ---------------------------------------------------------------------------

func BenchmarkCreateTask_LargeLabels(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	labels := make(map[string]string, 50)
	for i := 0; i < 50; i++ {
		labels[fmt.Sprintf("label-key-%d", i)] = fmt.Sprintf("label-value-%d", i)
	}

	for b.Loop() {
		_, err := svc.CreateTask(ctx, "tenant-1", "agent-1", "large labels goal",
			nil, "", nil, nil, 0, nil, labels)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. BenchmarkListTasks_Scaling
// ---------------------------------------------------------------------------

func BenchmarkListTasks_Scaling(b *testing.B) {
	for _, n := range []int{100, 1000, 10000, 50000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			svc := NewService(ServiceConfig{Repo: newMockRepo()})
			ctx := context.Background()

			for i := 0; i < n; i++ {
				_, err := svc.CreateTask(ctx, "tenant-1", "agent-1", fmt.Sprintf("goal-%d", i), nil, "", nil, nil, 0, nil, nil)
				if err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for b.Loop() {
				_, _, err := svc.ListTasks(ctx, "tenant-1", "", "", 50, "")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 7. BenchmarkListTasks_WithAgentFilter
// ---------------------------------------------------------------------------

func BenchmarkListTasks_WithAgentFilter(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	// Create tasks across multiple agents.
	for i := 0; i < 1000; i++ {
		agentID := fmt.Sprintf("agent-%d", i%10)
		_, err := svc.CreateTask(ctx, "tenant-1", agentID, fmt.Sprintf("goal-%d", i), nil, "", nil, nil, 0, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.ListTasks(ctx, "tenant-1", "agent-3", "", 50, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 8. BenchmarkListTasks_WithStatusFilter
// ---------------------------------------------------------------------------

func BenchmarkListTasks_WithStatusFilter(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	// Create tasks and transition some to running.
	for i := 0; i < 1000; i++ {
		t, err := svc.CreateTask(ctx, "tenant-1", "agent-1", fmt.Sprintf("goal-%d", i), nil, "", nil, nil, 0, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
		if i%3 == 0 {
			svc.UpdateTaskStatus(ctx, "tenant-1", t.ID, models.TaskStatusRunning, "")
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_, _, err := svc.ListTasks(ctx, "tenant-1", "", models.TaskStatusRunning, 50, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 9. BenchmarkStatusTransition_AllPaths
// ---------------------------------------------------------------------------

func BenchmarkStatusTransition_AllPaths(b *testing.B) {
	type transition struct {
		name string
		from models.TaskStatus
		to   models.TaskStatus
	}

	transitions := []transition{
		{"Pending_Running", models.TaskStatusPending, models.TaskStatusRunning},
		{"Running_Completed", models.TaskStatusRunning, models.TaskStatusCompleted},
		{"Running_Failed", models.TaskStatusRunning, models.TaskStatusFailed},
		{"Running_WaitingOnHuman", models.TaskStatusRunning, models.TaskStatusWaitingOnHuman},
		{"WaitingOnHuman_Running", models.TaskStatusWaitingOnHuman, models.TaskStatusRunning},
		{"Pending_Cancelled", models.TaskStatusPending, models.TaskStatusCancelled},
		{"Running_Cancelled", models.TaskStatusRunning, models.TaskStatusCancelled},
	}

	for _, tr := range transitions {
		b.Run(tr.name, func(b *testing.B) {
			b.ReportAllocs()
			ctx := context.Background()

			for b.Loop() {
				b.StopTimer()
				svc := NewService(ServiceConfig{Repo: newMockRepo()})
				t, err := svc.CreateTask(ctx, "tenant-1", "agent-1", "goal", nil, "", nil, nil, 0, nil, nil)
				if err != nil {
					b.Fatal(err)
				}

				// Walk from Pending to the `from` state.
				switch tr.from {
				case models.TaskStatusRunning:
					svc.UpdateTaskStatus(ctx, "tenant-1", t.ID, models.TaskStatusRunning, "")
				case models.TaskStatusWaitingOnHuman:
					svc.UpdateTaskStatus(ctx, "tenant-1", t.ID, models.TaskStatusRunning, "")
					svc.UpdateTaskStatus(ctx, "tenant-1", t.ID, models.TaskStatusWaitingOnHuman, "")
				}
				b.StartTimer()

				if tr.to == models.TaskStatusCancelled {
					err = svc.CancelTask(ctx, "tenant-1", t.ID, "bench cancel")
				} else {
					_, err = svc.UpdateTaskStatus(ctx, "tenant-1", t.ID, tr.to, "bench reason")
				}
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 10. BenchmarkTaskLifecycle_HappyPath
// ---------------------------------------------------------------------------

func BenchmarkTaskLifecycle_HappyPath(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	for b.Loop() {
		b.StopTimer()
		svc := NewService(ServiceConfig{Repo: newMockRepo()})
		b.StartTimer()

		// Create
		t, err := svc.CreateTask(ctx, "tenant-1", "agent-1", "lifecycle goal", nil, "", nil, nil, 0, nil, nil)
		if err != nil {
			b.Fatal(err)
		}

		// Pending → Running
		_, err = svc.UpdateTaskStatus(ctx, "tenant-1", t.ID, models.TaskStatusRunning, "")
		if err != nil {
			b.Fatal(err)
		}

		// Running → Completed
		_, err = svc.UpdateTaskStatus(ctx, "tenant-1", t.ID, models.TaskStatusCompleted, "done")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ---------------------------------------------------------------------------
// 11. BenchmarkCancelTask_Parallel
// ---------------------------------------------------------------------------

func BenchmarkCancelTask_Parallel(b *testing.B) {
	b.ReportAllocs()
	ctx := context.Background()

	// Pre-create tasks. Each iteration consumes one task, so create enough.
	// We use a shared atomic index to hand out tasks to goroutines.
	const poolSize = 1_000_000
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	taskIDs := make([]string, poolSize)
	for i := 0; i < poolSize; i++ {
		t, err := svc.CreateTask(ctx, "tenant-1", "agent-1", fmt.Sprintf("cancel-%d", i), nil, "", nil, nil, 0, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
		taskIDs[i] = t.ID
	}

	var idx atomic.Int64
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := int(idx.Add(1) - 1)
			if i >= poolSize {
				return // pool exhausted
			}
			err := svc.CancelTask(ctx, "tenant-1", taskIDs[i], "parallel cancel")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// 12. BenchmarkMultiTenantTaskIsolation
// ---------------------------------------------------------------------------

func BenchmarkMultiTenantTaskIsolation(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	const numTenants = 5
	const tasksPerTenant = 500

	tenants := make([]string, numTenants)
	taskIDs := make([][]string, numTenants)

	for t := 0; t < numTenants; t++ {
		tenants[t] = fmt.Sprintf("tenant-%d", t)
		taskIDs[t] = make([]string, tasksPerTenant)
		for i := 0; i < tasksPerTenant; i++ {
			tk, err := svc.CreateTask(ctx, tenants[t], "agent-1", fmt.Sprintf("goal-%d", i), nil, "", nil, nil, 0, nil, nil)
			if err != nil {
				b.Fatal(err)
			}
			taskIDs[t][i] = tk.ID
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := rand.IntN(numTenants * tasksPerTenant)
		for pb.Next() {
			t := i % numTenants
			idx := i % tasksPerTenant
			_, err := svc.GetTask(ctx, tenants[t], taskIDs[t][idx])
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// ---------------------------------------------------------------------------
// 13. BenchmarkMixedTaskWorkload
// ---------------------------------------------------------------------------

func BenchmarkMixedTaskWorkload(b *testing.B) {
	b.ReportAllocs()
	svc := NewService(ServiceConfig{Repo: newMockRepo()})
	ctx := context.Background()

	// Pre-populate tasks for reads and status updates.
	const seedCount = 500
	taskIDs := make([]string, seedCount)
	runningIDs := make([]string, 0, seedCount/2)

	for i := 0; i < seedCount; i++ {
		t, err := svc.CreateTask(ctx, "tenant-1", "agent-1", fmt.Sprintf("seed-%d", i), nil, "", nil, nil, 0, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
		taskIDs[i] = t.ID
		// Move some to running for status update benchmarking.
		if i%2 == 0 {
			_, err := svc.UpdateTaskStatus(ctx, "tenant-1", t.ID, models.TaskStatusRunning, "")
			if err != nil {
				b.Fatal(err)
			}
			runningIDs = append(runningIDs, t.ID)
		}
	}

	var createCounter atomic.Int64
	runningCount := len(runningIDs)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			r := i % 100
			switch {
			case r < 70: // 70% reads
				_, err := svc.GetTask(ctx, "tenant-1", taskIDs[i%seedCount])
				if err != nil {
					b.Fatal(err)
				}
			case r < 90: // 20% creates
				n := createCounter.Add(1)
				_, err := svc.CreateTask(ctx, "tenant-1", "agent-1", fmt.Sprintf("mixed-%d", n), nil, "", nil, nil, 0, nil, nil)
				if err != nil {
					b.Fatal(err)
				}
			default: // 10% status updates (Running → WaitingOnHuman is reversible)
				if runningCount > 0 {
					id := runningIDs[i%runningCount]
					svc.UpdateTaskStatus(ctx, "tenant-1", id, models.TaskStatusWaitingOnHuman, "")
					svc.UpdateTaskStatus(ctx, "tenant-1", id, models.TaskStatusRunning, "")
				}
			}
			i++
		}
	})
}

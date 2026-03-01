package e2e

import (
	"context"
	"testing"

	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/models"
	"github.com/Baselyne-Systems/bulkhead/control-plane/internal/task"
)

func TestTaskCreationToCompletion(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "task-agent")
	registerHost(t, ctx, "host-1.local:9090", 4096, 4000, 10240, []string{"standard", "hardened"})

	tk, err := taskSvc.CreateTask(ctx, tenant, agent.ID, "run tests", nil, "", nil, nil, 0, nil, nil)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if tk.Status != models.TaskStatusPending {
		t.Fatalf("expected pending, got %s", tk.Status)
	}

	tk, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk.ID, models.TaskStatusRunning, "")
	if err != nil {
		t.Fatalf("start task: %v", err)
	}
	if tk.Status != models.TaskStatusRunning {
		t.Fatalf("expected running, got %s", tk.Status)
	}
	if tk.WorkspaceID == "" {
		t.Fatal("expected workspace to be provisioned")
	}
	if fakeHostAgent.createCalls.Load() != 1 {
		t.Fatalf("expected 1 create sandbox call, got %d", fakeHostAgent.createCalls.Load())
	}

	ws, err := workspaceSvc.GetWorkspace(ctx, tenant, tk.WorkspaceID)
	if err != nil {
		t.Fatalf("get workspace: %v", err)
	}
	if ws.Status != models.WorkspaceStatusRunning {
		t.Fatalf("expected workspace running, got %s", ws.Status)
	}

	tk, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk.ID, models.TaskStatusCompleted, "done")
	if err != nil {
		t.Fatalf("complete task: %v", err)
	}
	if tk.Status != models.TaskStatusCompleted {
		t.Fatalf("expected completed, got %s", tk.Status)
	}
	if fakeHostAgent.destroyCalls.Load() != 1 {
		t.Fatalf("expected 1 destroy sandbox call, got %d", fakeHostAgent.destroyCalls.Load())
	}
}

func TestTaskCancellation(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "cancel-agent")
	registerHost(t, ctx, "host-cancel.local:9090", 4096, 4000, 10240, []string{"standard", "hardened"})

	tk, err := taskSvc.CreateTask(ctx, tenant, agent.ID, "will be cancelled", nil, "", nil, nil, 0, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	tk, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk.ID, models.TaskStatusRunning, "")
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	if err := taskSvc.CancelTask(ctx, tenant, tk.ID, "user requested"); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	tk, err = taskSvc.GetTask(ctx, tenant, tk.ID)
	if err != nil {
		t.Fatalf("get after cancel: %v", err)
	}
	if tk.Status != models.TaskStatusCancelled {
		t.Fatalf("expected cancelled, got %s", tk.Status)
	}
	if fakeHostAgent.destroyCalls.Load() != 1 {
		t.Fatalf("expected 1 destroy call, got %d", fakeHostAgent.destroyCalls.Load())
	}
}

func TestTaskFailureOnProvisioningError(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "fail-agent")

	tk, err := taskSvc.CreateTask(ctx, tenant, agent.ID, "doomed", nil, "", nil, nil, 0, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	tk, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk.ID, models.TaskStatusRunning, "")
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if tk.Status != models.TaskStatusFailed {
		t.Fatalf("expected failed, got %s", tk.Status)
	}
}

func TestTaskStatusTransitions(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "transition-agent")
	registerHost(t, ctx, "host-transitions.local:9090", 4096, 4000, 10240, []string{"standard", "hardened"})

	tk, err := taskSvc.CreateTask(ctx, tenant, agent.ID, "transition test", nil, "", nil, nil, 0, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	tk, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk.ID, models.TaskStatusRunning, "")
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	tk, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk.ID, models.TaskStatusCompleted, "done")
	if err != nil {
		t.Fatalf("complete: %v", err)
	}

	_, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk.ID, models.TaskStatusRunning, "")
	if err != task.ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition for completed->running, got %v", err)
	}

	tk2, err := taskSvc.CreateTask(ctx, tenant, agent.ID, "fail then try", nil, "", nil, nil, 0, nil, nil)
	if err != nil {
		t.Fatalf("create tk2: %v", err)
	}
	tk2, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk2.ID, models.TaskStatusFailed, "error")
	if err != nil {
		t.Fatalf("fail: %v", err)
	}

	_, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk2.ID, models.TaskStatusRunning, "")
	if err != task.ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition for failed->running, got %v", err)
	}
}

func TestTaskWaitingOnHuman(t *testing.T) {
	clean(t)
	ctx := context.Background()
	tenant := uniqueTenant()

	agent := registerAgent(t, ctx, tenant, "human-wait-agent")
	registerHost(t, ctx, "host-human.local:9090", 4096, 4000, 10240, []string{"standard", "hardened"})

	tk, err := taskSvc.CreateTask(ctx, tenant, agent.ID, "needs human", nil, "", nil, nil, 0, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	tk, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk.ID, models.TaskStatusRunning, "")
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	tk, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk.ID, models.TaskStatusWaitingOnHuman, "need approval")
	if err != nil {
		t.Fatalf("waiting: %v", err)
	}
	if tk.Status != models.TaskStatusWaitingOnHuman {
		t.Fatalf("expected waiting_on_human, got %s", tk.Status)
	}

	tk, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk.ID, models.TaskStatusRunning, "human approved")
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if tk.Status != models.TaskStatusRunning {
		t.Fatalf("expected running, got %s", tk.Status)
	}

	tk, err = taskSvc.UpdateTaskStatus(ctx, tenant, tk.ID, models.TaskStatusCompleted, "all done")
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if tk.Status != models.TaskStatusCompleted {
		t.Fatalf("expected completed, got %s", tk.Status)
	}
}

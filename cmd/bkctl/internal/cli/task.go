package cli

import (
	"fmt"

	taskpb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/task/v1"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks (create, list, cancel)",
}

func init() {
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskGetCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskCancelCmd)
	taskCmd.AddCommand(taskUpdateStatusCmd)
}

var taskHeaders = []string{"TASK ID", "AGENT ID", "GOAL", "STATUS", "WORKSPACE ID", "CREATED"}

func taskRow(t *taskpb.Task) []string {
	goal := t.GetGoal()
	if len(goal) > 40 {
		goal = goal[:37] + "..."
	}
	return []string{
		t.GetTaskId(),
		t.GetAgentId(),
		goal,
		formatTaskStatus(t.GetStatus()),
		t.GetWorkspaceId(),
		formatTimestamp(t.GetCreatedAt()),
	}
}

func formatTaskStatus(s taskpb.TaskStatus) string {
	switch s {
	case taskpb.TaskStatus_TASK_STATUS_PENDING:
		return "pending"
	case taskpb.TaskStatus_TASK_STATUS_RUNNING:
		return "running"
	case taskpb.TaskStatus_TASK_STATUS_WAITING_ON_HUMAN:
		return "waiting_on_human"
	case taskpb.TaskStatus_TASK_STATUS_COMPLETED:
		return "completed"
	case taskpb.TaskStatus_TASK_STATUS_FAILED:
		return "failed"
	case taskpb.TaskStatus_TASK_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "unspecified"
	}
}

func parseTaskStatus(s string) taskpb.TaskStatus {
	switch s {
	case "pending":
		return taskpb.TaskStatus_TASK_STATUS_PENDING
	case "running":
		return taskpb.TaskStatus_TASK_STATUS_RUNNING
	case "waiting_on_human":
		return taskpb.TaskStatus_TASK_STATUS_WAITING_ON_HUMAN
	case "completed":
		return taskpb.TaskStatus_TASK_STATUS_COMPLETED
	case "failed":
		return taskpb.TaskStatus_TASK_STATUS_FAILED
	case "cancelled":
		return taskpb.TaskStatus_TASK_STATUS_CANCELLED
	default:
		return taskpb.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

// --- Create ---

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, _ := cmd.Flags().GetString("agent-id")
		if agentID == "" {
			return fmt.Errorf("--agent-id is required")
		}
		goal, _ := cmd.Flags().GetString("goal")
		if goal == "" {
			return fmt.Errorf("--goal is required")
		}

		conn, err := dialService(cmd, "task")
		if err != nil {
			return err
		}
		defer conn.Close()

		memory, _ := cmd.Flags().GetInt64("memory")
		cpu, _ := cmd.Flags().GetInt32("cpu")
		disk, _ := cmd.Flags().GetInt64("disk")
		image, _ := cmd.Flags().GetString("image")
		isolationTier, _ := cmd.Flags().GetString("isolation-tier")
		egress, _ := cmd.Flags().GetStringSlice("egress-allowlist")
		tools, _ := cmd.Flags().GetStringSlice("allowed-tools")
		guardrailSetID, _ := cmd.Flags().GetString("guardrail-set-id")
		budget, _ := cmd.Flags().GetFloat64("budget")
		onExceeded, _ := cmd.Flags().GetString("on-exceeded")
		labelsRaw, _ := cmd.Flags().GetStringToString("labels")

		req := &taskpb.CreateTaskRequest{
			AgentId: agentID,
			Goal:    goal,
			WorkspaceConfig: &taskpb.TaskWorkspaceConfig{
				MemoryMb:        memory,
				CpuMillicores:   cpu,
				DiskMb:          disk,
				ContainerImage:  image,
				IsolationTier:   isolationTier,
				EgressAllowlist: egress,
				AllowedTools:    tools,
			},
			GuardrailPolicyId: guardrailSetID,
			Labels:            labelsRaw,
		}
		if budget > 0 {
			req.Budget = &taskpb.BudgetConfig{
				MaxCost:    budget,
				OnExceeded: onExceeded,
			}
		}

		client := taskpb.NewTaskServiceClient(conn)
		resp, err := client.CreateTask(cmd.Context(), req)
		if err != nil {
			return grpcError(err)
		}

		t := resp.GetTask()
		return outputResult(cmd, taskHeaders, [][]string{taskRow(t)}, resp)
	},
}

func init() {
	taskCreateCmd.Flags().String("agent-id", "", "Agent ID (required)")
	taskCreateCmd.Flags().String("goal", "", "Task goal (required)")
	taskCreateCmd.Flags().Int64("memory", 512, "Memory limit in MB")
	taskCreateCmd.Flags().Int32("cpu", 1000, "CPU limit in millicores")
	taskCreateCmd.Flags().Int64("disk", 1024, "Disk limit in MB")
	taskCreateCmd.Flags().String("image", "", "Container image to run")
	taskCreateCmd.Flags().String("isolation-tier", "", "Isolation tier: standard, hardened, isolated")
	taskCreateCmd.Flags().StringSlice("egress-allowlist", nil, "Allowed egress destinations")
	taskCreateCmd.Flags().StringSlice("allowed-tools", nil, "Allowed tool names")
	taskCreateCmd.Flags().String("guardrail-set-id", "", "Guardrail set ID to apply")
	taskCreateCmd.Flags().Float64("budget", 0, "Budget limit (e.g. 100.00)")
	taskCreateCmd.Flags().String("on-exceeded", "halt", "Action on budget exceeded: halt, request_increase")
	taskCreateCmd.Flags().StringToString("labels", nil, "Labels as key=value pairs")
}

// --- Get ---

var taskGetCmd = &cobra.Command{
	Use:   "get [task-id]",
	Short: "Get task details",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "task-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "task")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := taskpb.NewTaskServiceClient(conn)
		resp, err := client.GetTask(cmd.Context(), &taskpb.GetTaskRequest{TaskId: id})
		if err != nil {
			return grpcError(err)
		}

		t := resp.GetTask()
		return outputResult(cmd, taskHeaders, [][]string{taskRow(t)}, resp)
	},
}

func init() {
	taskGetCmd.Flags().String("task-id", "", "Task ID")
}

// --- List ---

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "task")
		if err != nil {
			return err
		}
		defer conn.Close()

		agentID, _ := cmd.Flags().GetString("agent-id")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt32("limit")

		client := taskpb.NewTaskServiceClient(conn)
		resp, err := client.ListTasks(cmd.Context(), &taskpb.ListTasksRequest{
			AgentId:  agentID,
			Status:   parseTaskStatus(status),
			PageSize: limit,
		})
		if err != nil {
			return grpcError(err)
		}

		rows := make([][]string, len(resp.GetTasks()))
		for i, t := range resp.GetTasks() {
			rows[i] = taskRow(t)
		}
		return outputResult(cmd, taskHeaders, rows, resp)
	},
}

func init() {
	taskListCmd.Flags().String("agent-id", "", "Filter by agent ID")
	taskListCmd.Flags().String("status", "", "Filter by status: pending, running, waiting_on_human, completed, failed, cancelled")
	taskListCmd.Flags().Int32("limit", 50, "Page size")
}

// --- Cancel ---

var taskCancelCmd = &cobra.Command{
	Use:   "cancel [task-id]",
	Short: "Cancel a running task",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "task-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "task")
		if err != nil {
			return err
		}
		defer conn.Close()

		reason, _ := cmd.Flags().GetString("reason")

		client := taskpb.NewTaskServiceClient(conn)
		_, err = client.CancelTask(cmd.Context(), &taskpb.CancelTaskRequest{
			TaskId: id,
			Reason: reason,
		})
		if err != nil {
			return grpcError(err)
		}

		fmt.Printf("Task %s cancelled.\n", id)
		return nil
	},
}

func init() {
	taskCancelCmd.Flags().String("task-id", "", "Task ID")
	taskCancelCmd.Flags().String("reason", "", "Cancellation reason")
}

// --- Update Status ---

var taskUpdateStatusCmd = &cobra.Command{
	Use:   "update-status [task-id]",
	Short: "Update a task's status",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "task-id")
		if err != nil {
			return err
		}
		status, _ := cmd.Flags().GetString("status")
		if status == "" {
			return fmt.Errorf("--status is required")
		}
		reason, _ := cmd.Flags().GetString("reason")

		conn, err := dialService(cmd, "task")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := taskpb.NewTaskServiceClient(conn)
		resp, err := client.UpdateTaskStatus(cmd.Context(), &taskpb.UpdateTaskStatusRequest{
			TaskId: id,
			Status: parseTaskStatus(status),
			Reason: reason,
		})
		if err != nil {
			return grpcError(err)
		}

		t := resp.GetTask()
		return outputResult(cmd, taskHeaders, [][]string{taskRow(t)}, resp)
	},
}

func init() {
	taskUpdateStatusCmd.Flags().String("task-id", "", "Task ID")
	taskUpdateStatusCmd.Flags().String("status", "", "New status: pending, running, waiting_on_human, completed, failed, cancelled (required)")
	taskUpdateStatusCmd.Flags().String("reason", "", "Reason for the status change")
}

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	workspacepb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/workspace/v1"
)

var workspaceCmd = &cobra.Command{
	Use:     "workspace",
	Aliases: []string{"ws"},
	Short:   "Manage workspaces (create, list, terminate)",
}

func init() {
	workspaceCmd.AddCommand(wsCreateCmd)
	workspaceCmd.AddCommand(wsGetCmd)
	workspaceCmd.AddCommand(wsListCmd)
	workspaceCmd.AddCommand(wsTerminateCmd)
}

var wsHeaders = []string{"WORKSPACE ID", "AGENT ID", "STATUS", "ISOLATION", "MEMORY MB", "CREATED"}

func wsRow(w *workspacepb.Workspace) []string {
	memMB := int64(0)
	isolation := "-"
	if w.GetSpec() != nil {
		memMB = w.GetSpec().GetMemoryMb()
	}
	if w.GetIsolationTier() != workspacepb.IsolationTier_ISOLATION_TIER_UNSPECIFIED {
		isolation = formatIsolationTier(w.GetIsolationTier())
	}
	return []string{
		w.GetWorkspaceId(),
		w.GetAgentId(),
		formatWsStatus(w.GetStatus()),
		isolation,
		fmt.Sprintf("%d", memMB),
		formatTimestamp(w.GetCreatedAt()),
	}
}

func formatWsStatus(s workspacepb.WorkspaceStatus) string {
	switch s {
	case workspacepb.WorkspaceStatus_WORKSPACE_STATUS_PENDING:
		return "pending"
	case workspacepb.WorkspaceStatus_WORKSPACE_STATUS_CREATING:
		return "creating"
	case workspacepb.WorkspaceStatus_WORKSPACE_STATUS_RUNNING:
		return "running"
	case workspacepb.WorkspaceStatus_WORKSPACE_STATUS_PAUSED:
		return "paused"
	case workspacepb.WorkspaceStatus_WORKSPACE_STATUS_TERMINATING:
		return "terminating"
	case workspacepb.WorkspaceStatus_WORKSPACE_STATUS_TERMINATED:
		return "terminated"
	case workspacepb.WorkspaceStatus_WORKSPACE_STATUS_FAILED:
		return "failed"
	default:
		return "unknown"
	}
}

func formatIsolationTier(t workspacepb.IsolationTier) string {
	switch t {
	case workspacepb.IsolationTier_ISOLATION_TIER_STANDARD:
		return "standard"
	case workspacepb.IsolationTier_ISOLATION_TIER_HARDENED:
		return "hardened"
	case workspacepb.IsolationTier_ISOLATION_TIER_ISOLATED:
		return "isolated"
	default:
		return "unspecified"
	}
}

func parseIsolationTier(s string) workspacepb.IsolationTier {
	switch s {
	case "standard":
		return workspacepb.IsolationTier_ISOLATION_TIER_STANDARD
	case "hardened":
		return workspacepb.IsolationTier_ISOLATION_TIER_HARDENED
	case "isolated":
		return workspacepb.IsolationTier_ISOLATION_TIER_ISOLATED
	default:
		return workspacepb.IsolationTier_ISOLATION_TIER_UNSPECIFIED
	}
}

// --- Create ---

var wsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, _ := cmd.Flags().GetString("agent-id")
		if agentID == "" {
			return fmt.Errorf("--agent-id is required")
		}

		conn, err := dialService(cmd, "workspace")
		if err != nil {
			return err
		}
		defer conn.Close()

		taskID, _ := cmd.Flags().GetString("task-id")
		memMB, _ := cmd.Flags().GetInt64("memory-mb")
		cpuMilli, _ := cmd.Flags().GetInt32("cpu-millicores")
		diskMB, _ := cmd.Flags().GetInt64("disk-mb")
		image, _ := cmd.Flags().GetString("image")
		tier, _ := cmd.Flags().GetString("isolation-tier")
		dataClass, _ := cmd.Flags().GetString("data-classification")
		egress, _ := cmd.Flags().GetStringSlice("egress-allowlist")
		tools, _ := cmd.Flags().GetStringSlice("allowed-tools")

		client := workspacepb.NewWorkspaceServiceClient(conn)
		resp, err := client.CreateWorkspace(cmd.Context(), &workspacepb.CreateWorkspaceRequest{
			AgentId: agentID,
			TaskId:  taskID,
			Spec: &workspacepb.WorkspaceSpec{
				MemoryMb:           memMB,
				CpuMillicores:      cpuMilli,
				DiskMb:             diskMB,
				ContainerImage:     image,
				IsolationTier:      parseIsolationTier(tier),
				DataClassification: dataClass,
				EgressAllowlist:    egress,
				AllowedTools:       tools,
			},
		})
		if err != nil {
			return grpcError(err)
		}

		w := resp.GetWorkspace()
		return outputResult(cmd, wsHeaders, [][]string{wsRow(w)}, resp)
	},
}

func init() {
	wsCreateCmd.Flags().String("agent-id", "", "Agent ID (required)")
	wsCreateCmd.Flags().String("task-id", "", "Task ID")
	wsCreateCmd.Flags().Int64("memory-mb", 512, "Memory limit in MB")
	wsCreateCmd.Flags().Int32("cpu-millicores", 1000, "CPU limit in millicores")
	wsCreateCmd.Flags().Int64("disk-mb", 1024, "Disk limit in MB")
	wsCreateCmd.Flags().String("image", "", "Container image")
	wsCreateCmd.Flags().String("isolation-tier", "", "Isolation tier: standard, hardened, isolated")
	wsCreateCmd.Flags().String("data-classification", "", "Data classification: public, internal, confidential, restricted")
	wsCreateCmd.Flags().StringSlice("egress-allowlist", nil, "Allowed egress destinations")
	wsCreateCmd.Flags().StringSlice("allowed-tools", nil, "Allowed tool names")
}

// --- Get ---

var wsGetCmd = &cobra.Command{
	Use:   "get [workspace-id]",
	Short: "Get workspace details",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "workspace-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "workspace")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := workspacepb.NewWorkspaceServiceClient(conn)
		resp, err := client.GetWorkspace(cmd.Context(), &workspacepb.GetWorkspaceRequest{WorkspaceId: id})
		if err != nil {
			return grpcError(err)
		}

		w := resp.GetWorkspace()
		return outputResult(cmd, wsHeaders, [][]string{wsRow(w)}, resp)
	},
}

func init() {
	wsGetCmd.Flags().String("workspace-id", "", "Workspace ID")
}

// --- List ---

var wsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "workspace")
		if err != nil {
			return err
		}
		defer conn.Close()

		agentID, _ := cmd.Flags().GetString("agent-id")
		limit, _ := cmd.Flags().GetInt32("limit")

		client := workspacepb.NewWorkspaceServiceClient(conn)
		resp, err := client.ListWorkspaces(cmd.Context(), &workspacepb.ListWorkspacesRequest{
			AgentId:  agentID,
			PageSize: limit,
		})
		if err != nil {
			return grpcError(err)
		}

		rows := make([][]string, len(resp.GetWorkspaces()))
		for i, w := range resp.GetWorkspaces() {
			rows[i] = wsRow(w)
		}
		return outputResult(cmd, wsHeaders, rows, resp)
	},
}

func init() {
	wsListCmd.Flags().String("agent-id", "", "Filter by agent ID")
	wsListCmd.Flags().Int32("limit", 50, "Page size")
}

// --- Terminate ---

var wsTerminateCmd = &cobra.Command{
	Use:   "terminate [workspace-id]",
	Short: "Terminate a workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "workspace-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "workspace")
		if err != nil {
			return err
		}
		defer conn.Close()

		reason, _ := cmd.Flags().GetString("reason")

		client := workspacepb.NewWorkspaceServiceClient(conn)
		_, err = client.TerminateWorkspace(cmd.Context(), &workspacepb.TerminateWorkspaceRequest{
			WorkspaceId: id,
			Reason:      reason,
		})
		if err != nil {
			return grpcError(err)
		}

		fmt.Printf("Workspace %s terminated.\n", id)
		return nil
	},
}

func init() {
	wsTerminateCmd.Flags().String("workspace-id", "", "Workspace ID")
	wsTerminateCmd.Flags().String("reason", "", "Termination reason")
}

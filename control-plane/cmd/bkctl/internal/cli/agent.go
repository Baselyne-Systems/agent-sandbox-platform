package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	identitypb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/identity/v1"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage agents (register, list, suspend, etc.)",
}

func init() {
	agentCmd.AddCommand(agentRegisterCmd)
	agentCmd.AddCommand(agentGetCmd)
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentSuspendCmd)
	agentCmd.AddCommand(agentReactivateCmd)
	agentCmd.AddCommand(agentUpdateTrustCmd)
}

var agentHeaders = []string{"AGENT ID", "NAME", "STATUS", "TRUST LEVEL", "OWNER", "CREATED"}

func agentRow(a *identitypb.Agent) []string {
	return []string{
		a.GetAgentId(),
		a.GetName(),
		formatAgentStatus(a.GetStatus()),
		formatTrustLevel(a.GetTrustLevel()),
		a.GetOwnerId(),
		formatTimestamp(a.GetCreatedAt()),
	}
}

func formatAgentStatus(s identitypb.AgentStatus) string {
	switch s {
	case identitypb.AgentStatus_AGENT_STATUS_ACTIVE:
		return "active"
	case identitypb.AgentStatus_AGENT_STATUS_INACTIVE:
		return "inactive"
	case identitypb.AgentStatus_AGENT_STATUS_SUSPENDED:
		return "suspended"
	default:
		return "unknown"
	}
}

func formatTrustLevel(t identitypb.AgentTrustLevel) string {
	switch t {
	case identitypb.AgentTrustLevel_AGENT_TRUST_LEVEL_NEW:
		return "new"
	case identitypb.AgentTrustLevel_AGENT_TRUST_LEVEL_ESTABLISHED:
		return "established"
	case identitypb.AgentTrustLevel_AGENT_TRUST_LEVEL_TRUSTED:
		return "trusted"
	default:
		return "unspecified"
	}
}

func parseTrustLevel(s string) identitypb.AgentTrustLevel {
	switch s {
	case "new":
		return identitypb.AgentTrustLevel_AGENT_TRUST_LEVEL_NEW
	case "established":
		return identitypb.AgentTrustLevel_AGENT_TRUST_LEVEL_ESTABLISHED
	case "trusted":
		return identitypb.AgentTrustLevel_AGENT_TRUST_LEVEL_TRUSTED
	default:
		return identitypb.AgentTrustLevel_AGENT_TRUST_LEVEL_NEW
	}
}

// --- Register ---

var agentRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "identity")
		if err != nil {
			return err
		}
		defer conn.Close()

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		desc, _ := cmd.Flags().GetString("description")
		ownerID, _ := cmd.Flags().GetString("owner-id")
		purpose, _ := cmd.Flags().GetString("purpose")
		trustLevel, _ := cmd.Flags().GetString("trust-level")
		capabilities, _ := cmd.Flags().GetStringSlice("capabilities")

		client := identitypb.NewIdentityServiceClient(conn)
		resp, err := client.RegisterAgent(cmd.Context(), &identitypb.RegisterAgentRequest{
			Name:         name,
			Description:  desc,
			OwnerId:      ownerID,
			Purpose:      purpose,
			TrustLevel:   parseTrustLevel(trustLevel),
			Capabilities: capabilities,
		})
		if err != nil {
			return grpcError(err)
		}

		a := resp.GetAgent()
		return outputResult(cmd, agentHeaders, [][]string{agentRow(a)}, resp)
	},
}

func init() {
	agentRegisterCmd.Flags().String("name", "", "Agent name (required)")
	agentRegisterCmd.Flags().String("description", "", "Agent description")
	agentRegisterCmd.Flags().String("owner-id", "", "Owner organization ID")
	agentRegisterCmd.Flags().String("purpose", "", "Agent purpose")
	agentRegisterCmd.Flags().String("trust-level", "new", "Trust level: new, established, trusted")
	agentRegisterCmd.Flags().StringSlice("capabilities", nil, "Agent capabilities")
}

// --- Get ---

var agentGetCmd = &cobra.Command{
	Use:   "get [agent-id]",
	Short: "Get agent details",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "agent-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "identity")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := identitypb.NewIdentityServiceClient(conn)
		resp, err := client.GetAgent(cmd.Context(), &identitypb.GetAgentRequest{AgentId: id})
		if err != nil {
			return grpcError(err)
		}

		a := resp.GetAgent()
		return outputResult(cmd, agentHeaders, [][]string{agentRow(a)}, resp)
	},
}

func init() {
	agentGetCmd.Flags().String("agent-id", "", "Agent ID")
}

// --- List ---

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "identity")
		if err != nil {
			return err
		}
		defer conn.Close()

		ownerID, _ := cmd.Flags().GetString("owner-id")
		limit, _ := cmd.Flags().GetInt32("limit")

		client := identitypb.NewIdentityServiceClient(conn)
		resp, err := client.ListAgents(cmd.Context(), &identitypb.ListAgentsRequest{
			OwnerId:  ownerID,
			PageSize: limit,
		})
		if err != nil {
			return grpcError(err)
		}

		rows := make([][]string, len(resp.GetAgents()))
		for i, a := range resp.GetAgents() {
			rows[i] = agentRow(a)
		}
		return outputResult(cmd, agentHeaders, rows, resp)
	},
}

func init() {
	agentListCmd.Flags().String("owner-id", "", "Filter by owner ID")
	agentListCmd.Flags().Int32("limit", 50, "Page size")
}

// --- Suspend ---

var agentSuspendCmd = &cobra.Command{
	Use:   "suspend [agent-id]",
	Short: "Suspend an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "agent-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "identity")
		if err != nil {
			return err
		}
		defer conn.Close()

		reason, _ := cmd.Flags().GetString("reason")

		client := identitypb.NewIdentityServiceClient(conn)
		_, err = client.SuspendAgent(cmd.Context(), &identitypb.SuspendAgentRequest{
			AgentId: id,
			Reason:  reason,
		})
		if err != nil {
			return grpcError(err)
		}

		fmt.Printf("Agent %s suspended.\n", id)
		return nil
	},
}

func init() {
	agentSuspendCmd.Flags().String("agent-id", "", "Agent ID")
	agentSuspendCmd.Flags().String("reason", "", "Suspension reason")
}

// --- Reactivate ---

var agentReactivateCmd = &cobra.Command{
	Use:   "reactivate [agent-id]",
	Short: "Reactivate a suspended agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "agent-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "identity")
		if err != nil {
			return err
		}
		defer conn.Close()

		reason, _ := cmd.Flags().GetString("reason")

		client := identitypb.NewIdentityServiceClient(conn)
		_, err = client.ReactivateAgent(cmd.Context(), &identitypb.ReactivateAgentRequest{
			AgentId: id,
			Reason:  reason,
		})
		if err != nil {
			return grpcError(err)
		}

		fmt.Printf("Agent %s reactivated.\n", id)
		return nil
	},
}

func init() {
	agentReactivateCmd.Flags().String("agent-id", "", "Agent ID")
	agentReactivateCmd.Flags().String("reason", "", "Reactivation reason")
}

// --- Update Trust Level ---

var agentUpdateTrustCmd = &cobra.Command{
	Use:   "update-trust-level",
	Short: "Update an agent's trust level",
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, _ := cmd.Flags().GetString("agent-id")
		if agentID == "" {
			return fmt.Errorf("--agent-id is required")
		}
		trustLevel, _ := cmd.Flags().GetString("trust-level")
		if trustLevel == "" {
			return fmt.Errorf("--trust-level is required")
		}
		justification, _ := cmd.Flags().GetString("justification")

		conn, err := dialService(cmd, "identity")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := identitypb.NewIdentityServiceClient(conn)
		_, err = client.UpdateTrustLevel(cmd.Context(), &identitypb.UpdateTrustLevelRequest{
			AgentId:       agentID,
			TrustLevel:    parseTrustLevel(trustLevel),
			Justification: justification,
		})
		if err != nil {
			return grpcError(err)
		}

		fmt.Printf("Agent %s trust level updated to %s.\n", agentID, trustLevel)
		return nil
	},
}

func init() {
	agentUpdateTrustCmd.Flags().String("agent-id", "", "Agent ID (required)")
	agentUpdateTrustCmd.Flags().String("trust-level", "", "Trust level: new, established, trusted (required)")
	agentUpdateTrustCmd.Flags().String("justification", "", "Justification for the change")
}

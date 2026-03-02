package cli

import (
	"fmt"
	"strings"

	humanpb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/human/v1"
	"github.com/spf13/cobra"
)

var humanCmd = &cobra.Command{
	Use:   "human",
	Short: "Manage human interaction requests and delivery",
}

func init() {
	humanCmd.AddCommand(humanListCmd)
	humanCmd.AddCommand(humanGetCmd)
	humanCmd.AddCommand(humanRespondCmd)
	humanCmd.AddCommand(humanConfigureDeliveryCmd)
	humanCmd.AddCommand(humanSetTimeoutCmd)
}

var humanHeaders = []string{"REQUEST ID", "AGENT ID", "TYPE", "QUESTION", "STATUS", "URGENCY", "CREATED"}

func humanRow(r *humanpb.HumanRequest) []string {
	question := r.GetQuestion()
	if len(question) > 40 {
		question = question[:37] + "..."
	}
	question = strings.ReplaceAll(question, "\n", " ")
	return []string{
		r.GetRequestId(),
		r.GetAgentId(),
		formatRequestType(r.GetType()),
		question,
		formatRequestStatus(r.GetStatus()),
		formatUrgency(r.GetUrgency()),
		formatTimestamp(r.GetCreatedAt()),
	}
}

func formatRequestStatus(s humanpb.HumanRequestStatus) string {
	switch s {
	case humanpb.HumanRequestStatus_HUMAN_REQUEST_STATUS_PENDING:
		return "pending"
	case humanpb.HumanRequestStatus_HUMAN_REQUEST_STATUS_RESPONDED:
		return "responded"
	case humanpb.HumanRequestStatus_HUMAN_REQUEST_STATUS_EXPIRED:
		return "expired"
	case humanpb.HumanRequestStatus_HUMAN_REQUEST_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "unspecified"
	}
}

func parseRequestStatus(s string) humanpb.HumanRequestStatus {
	switch s {
	case "pending":
		return humanpb.HumanRequestStatus_HUMAN_REQUEST_STATUS_PENDING
	case "responded":
		return humanpb.HumanRequestStatus_HUMAN_REQUEST_STATUS_RESPONDED
	case "expired":
		return humanpb.HumanRequestStatus_HUMAN_REQUEST_STATUS_EXPIRED
	case "cancelled":
		return humanpb.HumanRequestStatus_HUMAN_REQUEST_STATUS_CANCELLED
	default:
		return humanpb.HumanRequestStatus_HUMAN_REQUEST_STATUS_UNSPECIFIED
	}
}

func formatRequestType(t humanpb.HumanRequestType) string {
	switch t {
	case humanpb.HumanRequestType_HUMAN_REQUEST_TYPE_APPROVAL:
		return "approval"
	case humanpb.HumanRequestType_HUMAN_REQUEST_TYPE_QUESTION:
		return "question"
	case humanpb.HumanRequestType_HUMAN_REQUEST_TYPE_ESCALATION:
		return "escalation"
	default:
		return "unspecified"
	}
}

func formatUrgency(u humanpb.HumanRequestUrgency) string {
	switch u {
	case humanpb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_LOW:
		return "low"
	case humanpb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_NORMAL:
		return "normal"
	case humanpb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_HIGH:
		return "high"
	case humanpb.HumanRequestUrgency_HUMAN_REQUEST_URGENCY_CRITICAL:
		return "critical"
	default:
		return "unspecified"
	}
}

// --- List ---

var humanListCmd = &cobra.Command{
	Use:   "list",
	Short: "List human interaction requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "human")
		if err != nil {
			return err
		}
		defer conn.Close()

		wsID, _ := cmd.Flags().GetString("workspace-id")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt32("limit")

		client := humanpb.NewHumanInteractionServiceClient(conn)
		resp, err := client.ListRequests(cmd.Context(), &humanpb.ListHumanRequestsRequest{
			WorkspaceId: wsID,
			Status:      parseRequestStatus(status),
			PageSize:    limit,
		})
		if err != nil {
			return grpcError(err)
		}

		rows := make([][]string, len(resp.GetRequests()))
		for i, r := range resp.GetRequests() {
			rows[i] = humanRow(r)
		}
		return outputResult(cmd, humanHeaders, rows, resp)
	},
}

func init() {
	humanListCmd.Flags().String("workspace-id", "", "Filter by workspace ID")
	humanListCmd.Flags().String("status", "", "Filter by status: pending, responded, expired, cancelled")
	humanListCmd.Flags().Int32("limit", 50, "Page size")
}

// --- Get ---

var humanGetCmd = &cobra.Command{
	Use:   "get [request-id]",
	Short: "Get human request details",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "request-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "human")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := humanpb.NewHumanInteractionServiceClient(conn)
		resp, err := client.GetRequest(cmd.Context(), &humanpb.GetHumanRequestRequest{RequestId: id})
		if err != nil {
			return grpcError(err)
		}

		r := resp.GetRequest()
		return outputResult(cmd, humanHeaders, [][]string{humanRow(r)}, resp)
	},
}

func init() {
	humanGetCmd.Flags().String("request-id", "", "Request ID")
}

// --- Respond ---

var humanRespondCmd = &cobra.Command{
	Use:   "respond [request-id]",
	Short: "Respond to a human interaction request",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "request-id")
		if err != nil {
			return err
		}
		response, _ := cmd.Flags().GetString("response")
		if response == "" {
			return fmt.Errorf("--response is required")
		}
		responderID, _ := cmd.Flags().GetString("responder-id")
		if responderID == "" {
			return fmt.Errorf("--responder-id is required")
		}

		conn, err := dialService(cmd, "human")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := humanpb.NewHumanInteractionServiceClient(conn)
		_, err = client.RespondToRequest(cmd.Context(), &humanpb.RespondToHumanRequestRequest{
			RequestId:   id,
			Response:    response,
			ResponderId: responderID,
		})
		if err != nil {
			return grpcError(err)
		}

		fmt.Printf("Responded to request %s.\n", id)
		return nil
	},
}

func init() {
	humanRespondCmd.Flags().String("request-id", "", "Request ID")
	humanRespondCmd.Flags().String("response", "", "Response text (required)")
	humanRespondCmd.Flags().String("responder-id", "", "Responder ID (required)")
}

// --- Configure Delivery ---

var humanConfigureDeliveryCmd = &cobra.Command{
	Use:   "configure-delivery",
	Short: "Configure a delivery channel for notifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		if userID == "" {
			return fmt.Errorf("--user-id is required")
		}
		channelType, _ := cmd.Flags().GetString("channel-type")
		if channelType == "" {
			return fmt.Errorf("--channel-type is required")
		}
		endpoint, _ := cmd.Flags().GetString("endpoint")
		if endpoint == "" {
			return fmt.Errorf("--endpoint is required")
		}

		conn, err := dialService(cmd, "human")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := humanpb.NewHumanInteractionServiceClient(conn)
		resp, err := client.ConfigureDeliveryChannel(cmd.Context(), &humanpb.ConfigureDeliveryChannelRequest{
			UserId:      userID,
			ChannelType: channelType,
			Endpoint:    endpoint,
		})
		if err != nil {
			return grpcError(err)
		}

		cfg := resp.GetConfig()
		cfgHeaders := []string{"CONFIG ID", "USER ID", "CHANNEL", "ENDPOINT", "ENABLED"}
		rows := [][]string{{
			cfg.GetConfigId(),
			cfg.GetUserId(),
			cfg.GetChannelType(),
			cfg.GetEndpoint(),
			fmt.Sprintf("%v", cfg.GetEnabled()),
		}}
		return outputResult(cmd, cfgHeaders, rows, resp)
	},
}

func init() {
	humanConfigureDeliveryCmd.Flags().String("user-id", "", "User ID (required)")
	humanConfigureDeliveryCmd.Flags().String("channel-type", "", "Channel type: slack, email, teams (required)")
	humanConfigureDeliveryCmd.Flags().String("endpoint", "", "Channel endpoint (required)")
}

// --- Set Timeout ---

var humanSetTimeoutCmd = &cobra.Command{
	Use:   "set-timeout",
	Short: "Set timeout policy for human interaction requests",
	RunE: func(cmd *cobra.Command, args []string) error {
		scope, _ := cmd.Flags().GetString("scope")
		if scope == "" {
			return fmt.Errorf("--scope is required (global, agent, workspace)")
		}
		timeoutSecs, _ := cmd.Flags().GetInt64("timeout-seconds")
		if timeoutSecs <= 0 {
			return fmt.Errorf("--timeout-seconds must be > 0")
		}
		action, _ := cmd.Flags().GetString("action")
		if action == "" {
			return fmt.Errorf("--action is required (escalate, continue, halt)")
		}

		conn, err := dialService(cmd, "human")
		if err != nil {
			return err
		}
		defer conn.Close()

		scopeID, _ := cmd.Flags().GetString("scope-id")
		escalationTargets, _ := cmd.Flags().GetStringSlice("escalation-targets")

		client := humanpb.NewHumanInteractionServiceClient(conn)
		resp, err := client.SetTimeoutPolicy(cmd.Context(), &humanpb.SetTimeoutPolicyRequest{
			Scope:             scope,
			ScopeId:           scopeID,
			TimeoutSeconds:    timeoutSecs,
			Action:            action,
			EscalationTargets: escalationTargets,
		})
		if err != nil {
			return grpcError(err)
		}

		p := resp.GetPolicy()
		pHeaders := []string{"POLICY ID", "SCOPE", "SCOPE ID", "TIMEOUT (s)", "ACTION"}
		rows := [][]string{{
			p.GetPolicyId(),
			p.GetScope(),
			p.GetScopeId(),
			fmt.Sprintf("%d", p.GetTimeoutSeconds()),
			p.GetAction(),
		}}
		return outputResult(cmd, pHeaders, rows, resp)
	},
}

func init() {
	humanSetTimeoutCmd.Flags().String("scope", "", "Scope: global, agent, workspace (required)")
	humanSetTimeoutCmd.Flags().String("scope-id", "", "Scope ID (agent or workspace ID)")
	humanSetTimeoutCmd.Flags().Int64("timeout-seconds", 0, "Timeout in seconds (required)")
	humanSetTimeoutCmd.Flags().String("action", "", "Timeout action: escalate, continue, halt (required)")
	humanSetTimeoutCmd.Flags().StringSlice("escalation-targets", nil, "Escalation target user IDs")
}

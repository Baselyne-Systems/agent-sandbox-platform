package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	activitypb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/activity/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var activityCmd = &cobra.Command{
	Use:     "activity",
	Aliases: []string{"act"},
	Short:   "Query activity, manage alerts",
}

func init() {
	activityCmd.AddCommand(actQueryCmd)
	activityCmd.AddCommand(actGetCmd)
	activityCmd.AddCommand(actStreamCmd)
	activityCmd.AddCommand(actExportCmd)
	activityCmd.AddCommand(actConfigureAlertCmd)
	activityCmd.AddCommand(actListAlertsCmd)
	activityCmd.AddCommand(actResolveAlertCmd)
}

var actionHeaders = []string{"RECORD ID", "AGENT ID", "TOOL", "OUTCOME", "RECORDED AT"}
var alertHeaders = []string{"ALERT ID", "AGENT ID", "CONDITION", "MESSAGE", "RESOLVED", "TRIGGERED AT"}

func actionRow(r *activitypb.ActionRecord) []string {
	return []string{
		r.GetRecordId(),
		r.GetAgentId(),
		r.GetToolName(),
		formatOutcome(r.GetOutcome()),
		formatTimestamp(r.GetRecordedAt()),
	}
}

func alertRow(a *activitypb.Alert) []string {
	return []string{
		a.GetAlertId(),
		a.GetAgentId(),
		formatConditionType(a.GetConditionType()),
		a.GetMessage(),
		fmt.Sprintf("%v", a.GetResolved()),
		formatTimestamp(a.GetTriggeredAt()),
	}
}

func formatOutcome(o activitypb.ActionOutcome) string {
	switch o {
	case activitypb.ActionOutcome_ACTION_OUTCOME_ALLOWED:
		return "allowed"
	case activitypb.ActionOutcome_ACTION_OUTCOME_DENIED:
		return "denied"
	case activitypb.ActionOutcome_ACTION_OUTCOME_ESCALATED:
		return "escalated"
	case activitypb.ActionOutcome_ACTION_OUTCOME_ERROR:
		return "error"
	default:
		return "unspecified"
	}
}

func parseOutcome(s string) activitypb.ActionOutcome {
	switch s {
	case "allowed":
		return activitypb.ActionOutcome_ACTION_OUTCOME_ALLOWED
	case "denied":
		return activitypb.ActionOutcome_ACTION_OUTCOME_DENIED
	case "escalated":
		return activitypb.ActionOutcome_ACTION_OUTCOME_ESCALATED
	case "error":
		return activitypb.ActionOutcome_ACTION_OUTCOME_ERROR
	default:
		return activitypb.ActionOutcome_ACTION_OUTCOME_UNSPECIFIED
	}
}

func formatConditionType(ct activitypb.AlertConditionType) string {
	switch ct {
	case activitypb.AlertConditionType_ALERT_CONDITION_TYPE_DENIAL_RATE:
		return "denial_rate"
	case activitypb.AlertConditionType_ALERT_CONDITION_TYPE_ERROR_RATE:
		return "error_rate"
	case activitypb.AlertConditionType_ALERT_CONDITION_TYPE_ACTION_VELOCITY:
		return "action_velocity"
	case activitypb.AlertConditionType_ALERT_CONDITION_TYPE_BUDGET_BREACH:
		return "budget_breach"
	case activitypb.AlertConditionType_ALERT_CONDITION_TYPE_STUCK_AGENT:
		return "stuck_agent"
	default:
		return "unspecified"
	}
}

func parseConditionType(s string) activitypb.AlertConditionType {
	switch s {
	case "denial_rate":
		return activitypb.AlertConditionType_ALERT_CONDITION_TYPE_DENIAL_RATE
	case "error_rate":
		return activitypb.AlertConditionType_ALERT_CONDITION_TYPE_ERROR_RATE
	case "action_velocity":
		return activitypb.AlertConditionType_ALERT_CONDITION_TYPE_ACTION_VELOCITY
	case "budget_breach":
		return activitypb.AlertConditionType_ALERT_CONDITION_TYPE_BUDGET_BREACH
	case "stuck_agent":
		return activitypb.AlertConditionType_ALERT_CONDITION_TYPE_STUCK_AGENT
	default:
		return activitypb.AlertConditionType_ALERT_CONDITION_TYPE_UNSPECIFIED
	}
}

func parseExportFormat(s string) activitypb.ExportFormat {
	switch s {
	case "json":
		return activitypb.ExportFormat_EXPORT_FORMAT_JSON
	case "csv":
		return activitypb.ExportFormat_EXPORT_FORMAT_CSV
	default:
		return activitypb.ExportFormat_EXPORT_FORMAT_UNSPECIFIED
	}
}

func parseTime(s string) (*timestamppb.Timestamp, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, fmt.Errorf("invalid time %q (use RFC3339 format): %w", s, err)
	}
	return timestamppb.New(t), nil
}

// --- Query ---

var actQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query action records",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "activity")
		if err != nil {
			return err
		}
		defer conn.Close()

		agentID, _ := cmd.Flags().GetString("agent-id")
		wsID, _ := cmd.Flags().GetString("workspace-id")
		taskID, _ := cmd.Flags().GetString("task-id")
		toolName, _ := cmd.Flags().GetString("tool-name")
		outcome, _ := cmd.Flags().GetString("outcome")
		startStr, _ := cmd.Flags().GetString("start")
		endStr, _ := cmd.Flags().GetString("end")
		limit, _ := cmd.Flags().GetInt32("limit")

		startTime, err := parseTime(startStr)
		if err != nil {
			return err
		}
		endTime, err := parseTime(endStr)
		if err != nil {
			return err
		}

		client := activitypb.NewActivityServiceClient(conn)
		resp, err := client.QueryActions(cmd.Context(), &activitypb.QueryActionsRequest{
			AgentId:     agentID,
			WorkspaceId: wsID,
			TaskId:      taskID,
			ToolName:    toolName,
			Outcome:     parseOutcome(outcome),
			StartTime:   startTime,
			EndTime:     endTime,
			PageSize:    limit,
		})
		if err != nil {
			return grpcError(err)
		}

		rows := make([][]string, len(resp.GetRecords()))
		for i, r := range resp.GetRecords() {
			rows[i] = actionRow(r)
		}
		return outputResult(cmd, actionHeaders, rows, resp)
	},
}

func init() {
	actQueryCmd.Flags().String("agent-id", "", "Filter by agent ID")
	actQueryCmd.Flags().String("workspace-id", "", "Filter by workspace ID")
	actQueryCmd.Flags().String("task-id", "", "Filter by task ID")
	actQueryCmd.Flags().String("tool-name", "", "Filter by tool name")
	actQueryCmd.Flags().String("outcome", "", "Filter by outcome: allowed, denied, escalated, error")
	actQueryCmd.Flags().String("start", "", "Start time (RFC3339)")
	actQueryCmd.Flags().String("end", "", "End time (RFC3339)")
	actQueryCmd.Flags().Int32("limit", 50, "Page size")
}

// --- Get ---

var actGetCmd = &cobra.Command{
	Use:   "get [record-id]",
	Short: "Get action record details",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "record-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "activity")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := activitypb.NewActivityServiceClient(conn)
		resp, err := client.GetAction(cmd.Context(), &activitypb.GetActionRequest{RecordId: id})
		if err != nil {
			return grpcError(err)
		}

		r := resp.GetRecord()
		return outputResult(cmd, actionHeaders, [][]string{actionRow(r)}, resp)
	},
}

func init() {
	actGetCmd.Flags().String("record-id", "", "Record ID")
}

// --- Stream ---

var actStreamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Stream action records in real-time",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "activity")
		if err != nil {
			return err
		}
		defer conn.Close()

		agentID, _ := cmd.Flags().GetString("agent-id")
		wsID, _ := cmd.Flags().GetString("workspace-id")

		client := activitypb.NewActivityServiceClient(conn)
		stream, err := client.StreamActions(cmd.Context(), &activitypb.StreamActionsRequest{
			AgentId:     agentID,
			WorkspaceId: wsID,
		})
		if err != nil {
			return grpcError(err)
		}

		isJSON := getOutputFormat(cmd) == "json"
		if !isJSON {
			printTable(os.Stdout, actionHeaders, nil)
		}

		for {
			record, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return grpcError(err)
			}

			if isJSON {
				if err := printJSON(os.Stdout, record); err != nil {
					return err
				}
			} else {
				printTable(os.Stdout, nil, [][]string{actionRow(record)})
			}
		}
	},
}

func init() {
	actStreamCmd.Flags().String("agent-id", "", "Filter by agent ID")
	actStreamCmd.Flags().String("workspace-id", "", "Filter by workspace ID")
}

// --- Export ---

var actExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export action records to file",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "activity")
		if err != nil {
			return err
		}
		defer conn.Close()

		agentID, _ := cmd.Flags().GetString("agent-id")
		wsID, _ := cmd.Flags().GetString("workspace-id")
		format, _ := cmd.Flags().GetString("format")
		startStr, _ := cmd.Flags().GetString("start")
		endStr, _ := cmd.Flags().GetString("end")
		outputFile, _ := cmd.Flags().GetString("output-file")

		startTime, err := parseTime(startStr)
		if err != nil {
			return err
		}
		endTime, err := parseTime(endStr)
		if err != nil {
			return err
		}

		client := activitypb.NewActivityServiceClient(conn)
		stream, err := client.ExportActions(cmd.Context(), &activitypb.ExportActionsRequest{
			AgentId:     agentID,
			WorkspaceId: wsID,
			StartTime:   startTime,
			EndTime:     endTime,
			Format:      parseExportFormat(format),
		})
		if err != nil {
			return grpcError(err)
		}

		var w io.Writer = os.Stdout
		if outputFile != "" {
			f, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("create output file: %w", err)
			}
			defer f.Close()
			w = f
		}

		totalRecords := 0
		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return grpcError(err)
			}

			if len(chunk.GetData()) > 0 {
				if _, err := w.Write(chunk.GetData()); err != nil {
					return err
				}
			}
			totalRecords += int(chunk.GetRecordCount())

			if chunk.GetIsLast() {
				break
			}
		}

		if outputFile != "" {
			fmt.Fprintf(os.Stderr, "Exported %d records to %s\n", totalRecords, outputFile)
		}
		return nil
	},
}

func init() {
	actExportCmd.Flags().String("agent-id", "", "Filter by agent ID")
	actExportCmd.Flags().String("workspace-id", "", "Filter by workspace ID")
	actExportCmd.Flags().String("format", "json", "Export format: json, csv")
	actExportCmd.Flags().String("start", "", "Start time (RFC3339)")
	actExportCmd.Flags().String("end", "", "End time (RFC3339)")
	actExportCmd.Flags().String("output-file", "", "Output file path (default: stdout)")
}

// --- Configure Alert ---

var actConfigureAlertCmd = &cobra.Command{
	Use:   "configure-alert",
	Short: "Configure an activity alert",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		condType, _ := cmd.Flags().GetString("condition-type")
		if condType == "" {
			return fmt.Errorf("--condition-type is required")
		}
		threshold, _ := cmd.Flags().GetFloat64("threshold")
		if threshold <= 0 {
			return fmt.Errorf("--threshold must be > 0")
		}

		conn, err := dialService(cmd, "activity")
		if err != nil {
			return err
		}
		defer conn.Close()

		agentID, _ := cmd.Flags().GetString("agent-id")
		webhookURL, _ := cmd.Flags().GetString("webhook-url")

		client := activitypb.NewActivityServiceClient(conn)
		resp, err := client.ConfigureAlert(cmd.Context(), &activitypb.ConfigureAlertRequest{
			Name:          name,
			ConditionType: parseConditionType(condType),
			Threshold:     threshold,
			AgentId:       agentID,
			WebhookUrl:    webhookURL,
		})
		if err != nil {
			return grpcError(err)
		}

		cfg := resp.GetConfig()
		cfgHeaders := []string{"CONFIG ID", "NAME", "CONDITION", "THRESHOLD", "AGENT", "ENABLED"}
		rows := [][]string{{
			cfg.GetConfigId(),
			cfg.GetName(),
			formatConditionType(cfg.GetConditionType()),
			fmt.Sprintf("%.2f", cfg.GetThreshold()),
			cfg.GetAgentId(),
			fmt.Sprintf("%v", cfg.GetEnabled()),
		}}
		return outputResult(cmd, cfgHeaders, rows, resp)
	},
}

func init() {
	actConfigureAlertCmd.Flags().String("name", "", "Alert name (required)")
	actConfigureAlertCmd.Flags().String("condition-type", "", "Condition: denial_rate, error_rate, action_velocity, budget_breach, stuck_agent (required)")
	actConfigureAlertCmd.Flags().Float64("threshold", 0, "Threshold value (required)")
	actConfigureAlertCmd.Flags().String("agent-id", "", "Scope to specific agent")
	actConfigureAlertCmd.Flags().String("webhook-url", "", "Webhook URL for notifications")
}

// --- List Alerts ---

var actListAlertsCmd = &cobra.Command{
	Use:   "list-alerts",
	Short: "List triggered alerts",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "activity")
		if err != nil {
			return err
		}
		defer conn.Close()

		agentID, _ := cmd.Flags().GetString("agent-id")
		activeOnly, _ := cmd.Flags().GetBool("active-only")
		limit, _ := cmd.Flags().GetInt32("limit")

		client := activitypb.NewActivityServiceClient(conn)
		resp, err := client.ListAlerts(cmd.Context(), &activitypb.ListAlertsRequest{
			AgentId:    agentID,
			ActiveOnly: activeOnly,
			PageSize:   limit,
		})
		if err != nil {
			return grpcError(err)
		}

		rows := make([][]string, len(resp.GetAlerts()))
		for i, a := range resp.GetAlerts() {
			rows[i] = alertRow(a)
		}
		return outputResult(cmd, alertHeaders, rows, resp)
	},
}

func init() {
	actListAlertsCmd.Flags().String("agent-id", "", "Filter by agent ID")
	actListAlertsCmd.Flags().Bool("active-only", false, "Only show unresolved alerts")
	actListAlertsCmd.Flags().Int32("limit", 50, "Page size")
}

// --- Resolve Alert ---

var actResolveAlertCmd = &cobra.Command{
	Use:   "resolve-alert [alert-id]",
	Short: "Resolve an alert",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "alert-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "activity")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := activitypb.NewActivityServiceClient(conn)
		_, err = client.ResolveAlert(cmd.Context(), &activitypb.ResolveAlertRequest{AlertId: id})
		if err != nil {
			return grpcError(err)
		}

		fmt.Printf("Alert %s resolved.\n", id)
		return nil
	},
}

func init() {
	actResolveAlertCmd.Flags().String("alert-id", "", "Alert ID")
}

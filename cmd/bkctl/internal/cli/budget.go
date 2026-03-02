package cli

import (
	"fmt"
	"time"

	economicspb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/economics/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var budgetCmd = &cobra.Command{
	Use:   "budget",
	Short: "Manage budgets and view cost reports",
}

func init() {
	budgetCmd.AddCommand(budgetSetCmd)
	budgetCmd.AddCommand(budgetGetCmd)
	budgetCmd.AddCommand(budgetCheckCmd)
	budgetCmd.AddCommand(budgetCostReportCmd)
}

var budgetHeaders = []string{"BUDGET ID", "AGENT ID", "LIMIT", "USED", "REMAINING", "CURRENCY", "ON EXCEEDED"}

func budgetRow(b *economicspb.Budget) []string {
	remaining := b.GetLimit() - b.GetUsed()
	return []string{
		b.GetBudgetId(),
		b.GetAgentId(),
		fmt.Sprintf("%.2f", b.GetLimit()),
		fmt.Sprintf("%.2f", b.GetUsed()),
		fmt.Sprintf("%.2f", remaining),
		b.GetCurrency(),
		formatOnExceeded(b.GetOnExceeded()),
	}
}

func formatOnExceeded(a economicspb.OnExceededAction) string {
	switch a {
	case economicspb.OnExceededAction_ON_EXCEEDED_ACTION_HALT:
		return "halt"
	case economicspb.OnExceededAction_ON_EXCEEDED_ACTION_REQUEST_INCREASE:
		return "request_increase"
	case economicspb.OnExceededAction_ON_EXCEEDED_ACTION_WARN:
		return "warn"
	default:
		return "unspecified"
	}
}

func parseOnExceeded(s string) economicspb.OnExceededAction {
	switch s {
	case "halt":
		return economicspb.OnExceededAction_ON_EXCEEDED_ACTION_HALT
	case "request_increase":
		return economicspb.OnExceededAction_ON_EXCEEDED_ACTION_REQUEST_INCREASE
	case "warn":
		return economicspb.OnExceededAction_ON_EXCEEDED_ACTION_WARN
	default:
		return economicspb.OnExceededAction_ON_EXCEEDED_ACTION_HALT
	}
}

// --- Set ---

var budgetSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a budget for an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, _ := cmd.Flags().GetString("agent-id")
		if agentID == "" {
			return fmt.Errorf("--agent-id is required")
		}
		maxCost, _ := cmd.Flags().GetFloat64("max-cost")
		if maxCost <= 0 {
			return fmt.Errorf("--max-cost must be > 0")
		}

		conn, err := dialService(cmd, "economics")
		if err != nil {
			return err
		}
		defer conn.Close()

		currency, _ := cmd.Flags().GetString("currency")
		onExceeded, _ := cmd.Flags().GetString("on-exceeded")
		warningThreshold, _ := cmd.Flags().GetFloat64("warning-threshold")

		client := economicspb.NewEconomicsServiceClient(conn)
		resp, err := client.SetBudget(cmd.Context(), &economicspb.SetBudgetRequest{
			AgentId:          agentID,
			Limit:            maxCost,
			Currency:         currency,
			OnExceeded:       parseOnExceeded(onExceeded),
			WarningThreshold: warningThreshold,
		})
		if err != nil {
			return grpcError(err)
		}

		b := resp.GetBudget()
		return outputResult(cmd, budgetHeaders, [][]string{budgetRow(b)}, resp)
	},
}

func init() {
	budgetSetCmd.Flags().String("agent-id", "", "Agent ID (required)")
	budgetSetCmd.Flags().Float64("max-cost", 0, "Maximum cost limit (required)")
	budgetSetCmd.Flags().String("currency", "USD", "Currency code")
	budgetSetCmd.Flags().String("on-exceeded", "halt", "Action on budget exceeded: halt, request_increase, warn")
	budgetSetCmd.Flags().Float64("warning-threshold", 0.8, "Warning threshold (0.0-1.0)")
}

// --- Get ---

var budgetGetCmd = &cobra.Command{
	Use:   "get [agent-id]",
	Short: "Get an agent's budget",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "agent-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "economics")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := economicspb.NewEconomicsServiceClient(conn)
		resp, err := client.GetBudget(cmd.Context(), &economicspb.GetBudgetRequest{AgentId: id})
		if err != nil {
			return grpcError(err)
		}

		b := resp.GetBudget()
		return outputResult(cmd, budgetHeaders, [][]string{budgetRow(b)}, resp)
	},
}

func init() {
	budgetGetCmd.Flags().String("agent-id", "", "Agent ID")
}

// --- Check ---

var budgetCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if an action is within budget",
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID, _ := cmd.Flags().GetString("agent-id")
		if agentID == "" {
			return fmt.Errorf("--agent-id is required")
		}
		cost, _ := cmd.Flags().GetFloat64("estimated-cost")
		if cost <= 0 {
			return fmt.Errorf("--estimated-cost must be > 0")
		}

		conn, err := dialService(cmd, "economics")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := economicspb.NewEconomicsServiceClient(conn)
		resp, err := client.CheckBudget(cmd.Context(), &economicspb.CheckBudgetRequest{
			AgentId:       agentID,
			EstimatedCost: cost,
		})
		if err != nil {
			return grpcError(err)
		}

		checkHeaders := []string{"ALLOWED", "REMAINING", "ACTION", "WARNING"}
		rows := [][]string{{
			fmt.Sprintf("%v", resp.GetAllowed()),
			fmt.Sprintf("%.2f", resp.GetRemaining()),
			resp.GetEnforcementAction(),
			fmt.Sprintf("%v", resp.GetWarning()),
		}}
		return outputResult(cmd, checkHeaders, rows, resp)
	},
}

func init() {
	budgetCheckCmd.Flags().String("agent-id", "", "Agent ID (required)")
	budgetCheckCmd.Flags().Float64("estimated-cost", 0, "Estimated cost of the action (required)")
}

// --- Cost Report ---

var budgetCostReportCmd = &cobra.Command{
	Use:   "cost-report",
	Short: "Get a cost breakdown report",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "economics")
		if err != nil {
			return err
		}
		defer conn.Close()

		agentID, _ := cmd.Flags().GetString("agent-id")
		startStr, _ := cmd.Flags().GetString("start")
		endStr, _ := cmd.Flags().GetString("end")

		var startTime, endTime *timestamppb.Timestamp
		if startStr != "" {
			t, err := time.Parse(time.RFC3339, startStr)
			if err != nil {
				return fmt.Errorf("invalid --start time: %w", err)
			}
			startTime = timestamppb.New(t)
		}
		if endStr != "" {
			t, err := time.Parse(time.RFC3339, endStr)
			if err != nil {
				return fmt.Errorf("invalid --end time: %w", err)
			}
			endTime = timestamppb.New(t)
		}

		client := economicspb.NewEconomicsServiceClient(conn)
		resp, err := client.GetCostReport(cmd.Context(), &economicspb.GetCostReportRequest{
			AgentId:   agentID,
			StartTime: startTime,
			EndTime:   endTime,
		})
		if err != nil {
			return grpcError(err)
		}

		costHeaders := []string{"RESOURCE TYPE", "TOTAL COST", "RECORDS"}
		rows := make([][]string, len(resp.GetByResourceType()))
		for i, b := range resp.GetByResourceType() {
			rows[i] = []string{
				b.GetResourceType(),
				fmt.Sprintf("%.2f", b.GetTotalCost()),
				fmt.Sprintf("%d", b.GetRecordCount()),
			}
		}
		return outputResult(cmd, costHeaders, rows, resp)
	},
}

func init() {
	budgetCostReportCmd.Flags().String("agent-id", "", "Filter by agent ID")
	budgetCostReportCmd.Flags().String("start", "", "Start time (RFC3339)")
	budgetCostReportCmd.Flags().String("end", "", "End time (RFC3339)")
}

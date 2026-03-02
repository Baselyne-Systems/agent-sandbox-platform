package cli

import (
	"encoding/json"
	"fmt"

	hostagentpb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/host_agent/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

var sandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Sandbox operations (status, evaluate-tool, report-result, human input)",
}

func init() {
	sandboxCmd.AddCommand(sandboxStatusCmd)
	sandboxCmd.AddCommand(sandboxEvaluateToolCmd)
	sandboxCmd.AddCommand(sandboxReportResultCmd)
	sandboxCmd.AddCommand(sandboxRequestHumanInputCmd)
	sandboxCmd.AddCommand(sandboxCheckHumanRequestCmd)
}

func formatSandboxState(s hostagentpb.SandboxState) string {
	switch s {
	case hostagentpb.SandboxState_SANDBOX_STATE_STARTING:
		return "starting"
	case hostagentpb.SandboxState_SANDBOX_STATE_RUNNING:
		return "running"
	case hostagentpb.SandboxState_SANDBOX_STATE_PAUSED:
		return "paused"
	case hostagentpb.SandboxState_SANDBOX_STATE_STOPPED:
		return "stopped"
	case hostagentpb.SandboxState_SANDBOX_STATE_FAILED:
		return "failed"
	default:
		return "unspecified"
	}
}

func formatVerdict(v hostagentpb.ActionVerdict) string {
	switch v {
	case hostagentpb.ActionVerdict_ACTION_VERDICT_ALLOW:
		return "ALLOW"
	case hostagentpb.ActionVerdict_ACTION_VERDICT_DENY:
		return "DENY"
	case hostagentpb.ActionVerdict_ACTION_VERDICT_ESCALATE:
		return "ESCALATE"
	default:
		return "UNSPECIFIED"
	}
}

// --- Status ---

var sandboxStatusCmd = &cobra.Command{
	Use:   "status [sandbox-id]",
	Short: "Get sandbox status",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "sandbox-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "host-agent")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := hostagentpb.NewHostAgentServiceClient(conn)
		resp, err := client.GetSandboxStatus(cmd.Context(), &hostagentpb.GetSandboxStatusRequest{
			SandboxId: id,
		})
		if err != nil {
			return grpcError(err)
		}

		headers := []string{"SANDBOX ID", "STATE", "MEMORY USED MB", "CPU USED MILLI", "ACTIONS", "STARTED"}
		rows := [][]string{{
			resp.GetSandboxId(),
			formatSandboxState(resp.GetState()),
			fmt.Sprintf("%d", resp.GetMemoryUsedMb()),
			fmt.Sprintf("%d", resp.GetCpuUsedMillicores()),
			fmt.Sprintf("%d", resp.GetActionsExecuted()),
			formatTimestamp(resp.GetStartedAt()),
		}}
		return outputResult(cmd, headers, rows, resp)
	},
}

func init() {
	sandboxStatusCmd.Flags().String("sandbox-id", "", "Sandbox ID")
}

// --- Evaluate Tool ---

var sandboxEvaluateToolCmd = &cobra.Command{
	Use:   "evaluate-tool",
	Short: "Evaluate a tool call against guardrails (debug)",
	RunE: func(cmd *cobra.Command, args []string) error {
		sandboxID, _ := cmd.Flags().GetString("sandbox-id")
		if sandboxID == "" {
			return fmt.Errorf("--sandbox-id is required")
		}
		toolName, _ := cmd.Flags().GetString("tool-name")
		if toolName == "" {
			return fmt.Errorf("--tool-name is required")
		}

		conn, err := dialService(cmd, "host-agent")
		if err != nil {
			return err
		}
		defer conn.Close()

		justification, _ := cmd.Flags().GetString("justification")
		paramsJSON, _ := cmd.Flags().GetString("parameters")

		var params *structpb.Struct
		if paramsJSON != "" {
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(paramsJSON), &m); err != nil {
				return fmt.Errorf("invalid --parameters JSON: %w", err)
			}
			params, err = structpb.NewStruct(m)
			if err != nil {
				return fmt.Errorf("building parameters struct: %w", err)
			}
		}

		ctx := metadata.AppendToOutgoingContext(cmd.Context(), "x-sandbox-id", sandboxID)

		client := hostagentpb.NewHostAgentAPIServiceClient(conn)
		resp, err := client.ExecuteTool(ctx, &hostagentpb.ExecuteToolRequest{
			ToolName:      toolName,
			Parameters:    params,
			Justification: justification,
		})
		if err != nil {
			return grpcError(err)
		}

		headers := []string{"VERDICT", "ACTION ID", "DENIAL REASON", "ESCALATION ID"}
		rows := [][]string{{
			formatVerdict(resp.GetVerdict()),
			resp.GetActionId(),
			resp.GetDenialReason(),
			resp.GetEscalationId(),
		}}
		return outputResult(cmd, headers, rows, resp)
	},
}

func init() {
	sandboxEvaluateToolCmd.Flags().String("sandbox-id", "", "Sandbox ID (required)")
	sandboxEvaluateToolCmd.Flags().String("tool-name", "", "Tool name (required)")
	sandboxEvaluateToolCmd.Flags().String("parameters", "", "Tool parameters as JSON string")
	sandboxEvaluateToolCmd.Flags().String("justification", "", "Justification for the tool call")
}

// --- Report Result ---

var sandboxReportResultCmd = &cobra.Command{
	Use:   "report-result",
	Short: "Report a tool execution result (debug)",
	RunE: func(cmd *cobra.Command, args []string) error {
		sandboxID, _ := cmd.Flags().GetString("sandbox-id")
		if sandboxID == "" {
			return fmt.Errorf("--sandbox-id is required")
		}
		actionID, _ := cmd.Flags().GetString("action-id")
		if actionID == "" {
			return fmt.Errorf("--action-id is required")
		}

		conn, err := dialService(cmd, "host-agent")
		if err != nil {
			return err
		}
		defer conn.Close()

		success, _ := cmd.Flags().GetBool("success")
		errorMsg, _ := cmd.Flags().GetString("error-message")
		resultJSON, _ := cmd.Flags().GetString("result")

		var result *structpb.Struct
		if resultJSON != "" {
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(resultJSON), &m); err != nil {
				return fmt.Errorf("invalid --result JSON: %w", err)
			}
			result, err = structpb.NewStruct(m)
			if err != nil {
				return fmt.Errorf("building result struct: %w", err)
			}
		}

		ctx := metadata.AppendToOutgoingContext(cmd.Context(), "x-sandbox-id", sandboxID)

		client := hostagentpb.NewHostAgentAPIServiceClient(conn)
		_, err = client.ReportActionResult(ctx, &hostagentpb.ReportActionResultRequest{
			ActionId:     actionID,
			Success:      success,
			Result:       result,
			ErrorMessage: errorMsg,
		})
		if err != nil {
			return grpcError(err)
		}

		fmt.Printf("Result reported for action %s.\n", actionID)
		return nil
	},
}

func init() {
	sandboxReportResultCmd.Flags().String("sandbox-id", "", "Sandbox ID (required)")
	sandboxReportResultCmd.Flags().String("action-id", "", "Action ID from evaluate-tool (required)")
	sandboxReportResultCmd.Flags().Bool("success", true, "Whether the tool execution succeeded")
	sandboxReportResultCmd.Flags().String("result", "", "Tool result as JSON string")
	sandboxReportResultCmd.Flags().String("error-message", "", "Error message if execution failed")
}

// --- Request Human Input ---

var sandboxRequestHumanInputCmd = &cobra.Command{
	Use:   "request-human-input",
	Short: "Request human input from a sandbox (debug)",
	RunE: func(cmd *cobra.Command, args []string) error {
		sandboxID, _ := cmd.Flags().GetString("sandbox-id")
		if sandboxID == "" {
			return fmt.Errorf("--sandbox-id is required")
		}
		question, _ := cmd.Flags().GetString("question")
		if question == "" {
			return fmt.Errorf("--question is required")
		}

		conn, err := dialService(cmd, "host-agent")
		if err != nil {
			return err
		}
		defer conn.Close()

		options, _ := cmd.Flags().GetStringSlice("options")
		context, _ := cmd.Flags().GetString("context")
		timeout, _ := cmd.Flags().GetInt64("timeout-seconds")

		ctx := metadata.AppendToOutgoingContext(cmd.Context(), "x-sandbox-id", sandboxID)

		client := hostagentpb.NewHostAgentAPIServiceClient(conn)
		resp, err := client.RequestHumanInput(ctx, &hostagentpb.RequestHumanInputRequest{
			Question:       question,
			Options:        options,
			Context:        context,
			TimeoutSeconds: timeout,
		})
		if err != nil {
			return grpcError(err)
		}

		headers := []string{"REQUEST ID", "RESPONSE", "RESPONDER", "TIMED OUT"}
		rows := [][]string{{
			resp.GetRequestId(),
			resp.GetResponse(),
			resp.GetResponderId(),
			fmt.Sprintf("%v", resp.GetTimedOut()),
		}}
		return outputResult(cmd, headers, rows, resp)
	},
}

func init() {
	sandboxRequestHumanInputCmd.Flags().String("sandbox-id", "", "Sandbox ID (required)")
	sandboxRequestHumanInputCmd.Flags().String("question", "", "Question for the human (required)")
	sandboxRequestHumanInputCmd.Flags().StringSlice("options", nil, "Response options")
	sandboxRequestHumanInputCmd.Flags().String("context", "", "Additional context")
	sandboxRequestHumanInputCmd.Flags().Int64("timeout-seconds", 300, "Timeout in seconds")
}

// --- Check Human Request ---

var sandboxCheckHumanRequestCmd = &cobra.Command{
	Use:   "check-human-request [request-id]",
	Short: "Check status of a human interaction request (debug)",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "request-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "host-agent")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := hostagentpb.NewHostAgentAPIServiceClient(conn)
		resp, err := client.CheckHumanRequest(cmd.Context(), &hostagentpb.CheckHumanRequestRequest{
			RequestId: id,
		})
		if err != nil {
			return grpcError(err)
		}

		headers := []string{"STATUS", "RESPONSE", "RESPONDER"}
		rows := [][]string{{
			resp.GetStatus(),
			resp.GetResponse(),
			resp.GetResponderId(),
		}}
		return outputResult(cmd, headers, rows, resp)
	},
}

func init() {
	sandboxCheckHumanRequestCmd.Flags().String("request-id", "", "Request ID")
}

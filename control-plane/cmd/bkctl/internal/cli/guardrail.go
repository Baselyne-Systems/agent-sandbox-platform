package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	guardrailspb "github.com/Baselyne-Systems/bulkhead/control-plane/pkg/gen/guardrails/v1"
)

var guardrailCmd = &cobra.Command{
	Use:     "guardrail",
	Aliases: []string{"gr"},
	Short:   "Manage guardrail rules and sets",
}

func init() {
	guardrailCmd.AddCommand(grCreateRuleCmd)
	guardrailCmd.AddCommand(grGetRuleCmd)
	guardrailCmd.AddCommand(grListRulesCmd)
	guardrailCmd.AddCommand(grUpdateRuleCmd)
	guardrailCmd.AddCommand(grDeleteRuleCmd)
	guardrailCmd.AddCommand(grSimulateCmd)
	guardrailCmd.AddCommand(grCreateSetCmd)
	guardrailCmd.AddCommand(grListSetsCmd)
}

var ruleHeaders = []string{"RULE ID", "NAME", "TYPE", "ACTION", "PRIORITY", "ENABLED"}
var setHeaders = []string{"SET ID", "NAME", "RULES", "CREATED"}

func ruleRow(r *guardrailspb.GuardrailRule) []string {
	return []string{
		r.GetRuleId(),
		r.GetName(),
		formatRuleType(r.GetType()),
		formatRuleAction(r.GetAction()),
		fmt.Sprintf("%d", r.GetPriority()),
		fmt.Sprintf("%v", r.GetEnabled()),
	}
}

func setRow(s *guardrailspb.GuardrailSet) []string {
	return []string{
		s.GetSetId(),
		s.GetName(),
		fmt.Sprintf("%d", len(s.GetRuleIds())),
		formatTimestamp(s.GetCreatedAt()),
	}
}

func formatRuleType(t guardrailspb.RuleType) string {
	switch t {
	case guardrailspb.RuleType_RULE_TYPE_TOOL_FILTER:
		return "tool_filter"
	case guardrailspb.RuleType_RULE_TYPE_PARAMETER_CHECK:
		return "parameter_check"
	case guardrailspb.RuleType_RULE_TYPE_RATE_LIMIT:
		return "rate_limit"
	case guardrailspb.RuleType_RULE_TYPE_BUDGET_LIMIT:
		return "budget_limit"
	default:
		return "unspecified"
	}
}

func parseRuleType(s string) guardrailspb.RuleType {
	switch s {
	case "tool_filter":
		return guardrailspb.RuleType_RULE_TYPE_TOOL_FILTER
	case "parameter_check":
		return guardrailspb.RuleType_RULE_TYPE_PARAMETER_CHECK
	case "rate_limit":
		return guardrailspb.RuleType_RULE_TYPE_RATE_LIMIT
	case "budget_limit":
		return guardrailspb.RuleType_RULE_TYPE_BUDGET_LIMIT
	default:
		return guardrailspb.RuleType_RULE_TYPE_UNSPECIFIED
	}
}

func formatRuleAction(a guardrailspb.RuleAction) string {
	switch a {
	case guardrailspb.RuleAction_RULE_ACTION_ALLOW:
		return "allow"
	case guardrailspb.RuleAction_RULE_ACTION_DENY:
		return "deny"
	case guardrailspb.RuleAction_RULE_ACTION_ESCALATE:
		return "escalate"
	case guardrailspb.RuleAction_RULE_ACTION_LOG:
		return "log"
	default:
		return "unspecified"
	}
}

func parseRuleAction(s string) guardrailspb.RuleAction {
	switch s {
	case "allow":
		return guardrailspb.RuleAction_RULE_ACTION_ALLOW
	case "deny":
		return guardrailspb.RuleAction_RULE_ACTION_DENY
	case "escalate":
		return guardrailspb.RuleAction_RULE_ACTION_ESCALATE
	case "log":
		return guardrailspb.RuleAction_RULE_ACTION_LOG
	default:
		return guardrailspb.RuleAction_RULE_ACTION_UNSPECIFIED
	}
}

// --- Create Rule ---

var grCreateRuleCmd = &cobra.Command{
	Use:   "create-rule",
	Short: "Create a new guardrail rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		ruleType, _ := cmd.Flags().GetString("type")
		if ruleType == "" {
			return fmt.Errorf("--type is required")
		}
		condition, _ := cmd.Flags().GetString("condition")
		if condition == "" {
			return fmt.Errorf("--condition is required")
		}
		action, _ := cmd.Flags().GetString("action")
		if action == "" {
			return fmt.Errorf("--action is required")
		}

		conn, err := dialService(cmd, "guardrails")
		if err != nil {
			return err
		}
		defer conn.Close()

		desc, _ := cmd.Flags().GetString("description")
		priority, _ := cmd.Flags().GetInt32("priority")

		client := guardrailspb.NewGuardrailsServiceClient(conn)
		resp, err := client.CreateRule(cmd.Context(), &guardrailspb.CreateRuleRequest{
			Name:        name,
			Description: desc,
			Type:        parseRuleType(ruleType),
			Condition:   condition,
			Action:      parseRuleAction(action),
			Priority:    priority,
		})
		if err != nil {
			return grpcError(err)
		}

		r := resp.GetRule()
		return outputResult(cmd, ruleHeaders, [][]string{ruleRow(r)}, resp)
	},
}

func init() {
	grCreateRuleCmd.Flags().String("name", "", "Rule name (required)")
	grCreateRuleCmd.Flags().String("description", "", "Rule description")
	grCreateRuleCmd.Flags().String("type", "", "Rule type: tool_filter, parameter_check, rate_limit, budget_limit (required)")
	grCreateRuleCmd.Flags().String("condition", "", "Rule condition expression (required)")
	grCreateRuleCmd.Flags().String("action", "", "Rule action: allow, deny, escalate, log (required)")
	grCreateRuleCmd.Flags().Int32("priority", 0, "Rule priority (higher = evaluated first)")
}

// --- Get Rule ---

var grGetRuleCmd = &cobra.Command{
	Use:   "get-rule [rule-id]",
	Short: "Get rule details",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "rule-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "guardrails")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := guardrailspb.NewGuardrailsServiceClient(conn)
		resp, err := client.GetRule(cmd.Context(), &guardrailspb.GetRuleRequest{RuleId: id})
		if err != nil {
			return grpcError(err)
		}

		r := resp.GetRule()
		return outputResult(cmd, ruleHeaders, [][]string{ruleRow(r)}, resp)
	},
}

func init() {
	grGetRuleCmd.Flags().String("rule-id", "", "Rule ID")
}

// --- List Rules ---

var grListRulesCmd = &cobra.Command{
	Use:   "list-rules",
	Short: "List guardrail rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "guardrails")
		if err != nil {
			return err
		}
		defer conn.Close()

		ruleType, _ := cmd.Flags().GetString("type")
		enabledOnly, _ := cmd.Flags().GetBool("enabled-only")
		limit, _ := cmd.Flags().GetInt32("limit")

		client := guardrailspb.NewGuardrailsServiceClient(conn)
		resp, err := client.ListRules(cmd.Context(), &guardrailspb.ListRulesRequest{
			Type:        parseRuleType(ruleType),
			EnabledOnly: enabledOnly,
			PageSize:    limit,
		})
		if err != nil {
			return grpcError(err)
		}

		rows := make([][]string, len(resp.GetRules()))
		for i, r := range resp.GetRules() {
			rows[i] = ruleRow(r)
		}
		return outputResult(cmd, ruleHeaders, rows, resp)
	},
}

func init() {
	grListRulesCmd.Flags().String("type", "", "Filter by rule type")
	grListRulesCmd.Flags().Bool("enabled-only", false, "Only show enabled rules")
	grListRulesCmd.Flags().Int32("limit", 50, "Page size")
}

// --- Update Rule ---

var grUpdateRuleCmd = &cobra.Command{
	Use:   "update-rule",
	Short: "Update a guardrail rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		ruleID, _ := cmd.Flags().GetString("rule-id")
		if ruleID == "" {
			return fmt.Errorf("--rule-id is required")
		}

		conn, err := dialService(cmd, "guardrails")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := guardrailspb.NewGuardrailsServiceClient(conn)

		// Fetch current rule.
		getResp, err := client.GetRule(cmd.Context(), &guardrailspb.GetRuleRequest{RuleId: ruleID})
		if err != nil {
			return grpcError(err)
		}
		rule := getResp.GetRule()

		// Apply flag overrides.
		if cmd.Flags().Changed("name") {
			rule.Name, _ = cmd.Flags().GetString("name")
		}
		if cmd.Flags().Changed("description") {
			rule.Description, _ = cmd.Flags().GetString("description")
		}
		if cmd.Flags().Changed("condition") {
			rule.Condition, _ = cmd.Flags().GetString("condition")
		}
		if cmd.Flags().Changed("action") {
			a, _ := cmd.Flags().GetString("action")
			rule.Action = parseRuleAction(a)
		}
		if cmd.Flags().Changed("priority") {
			rule.Priority, _ = cmd.Flags().GetInt32("priority")
		}
		if cmd.Flags().Changed("enabled") {
			rule.Enabled, _ = cmd.Flags().GetBool("enabled")
		}

		resp, err := client.UpdateRule(cmd.Context(), &guardrailspb.UpdateRuleRequest{Rule: rule})
		if err != nil {
			return grpcError(err)
		}

		r := resp.GetRule()
		return outputResult(cmd, ruleHeaders, [][]string{ruleRow(r)}, resp)
	},
}

func init() {
	grUpdateRuleCmd.Flags().String("rule-id", "", "Rule ID (required)")
	grUpdateRuleCmd.Flags().String("name", "", "New name")
	grUpdateRuleCmd.Flags().String("description", "", "New description")
	grUpdateRuleCmd.Flags().String("condition", "", "New condition")
	grUpdateRuleCmd.Flags().String("action", "", "New action: allow, deny, escalate, log")
	grUpdateRuleCmd.Flags().Int32("priority", 0, "New priority")
	grUpdateRuleCmd.Flags().Bool("enabled", true, "Enable/disable the rule")
}

// --- Delete Rule ---

var grDeleteRuleCmd = &cobra.Command{
	Use:   "delete-rule [rule-id]",
	Short: "Delete a guardrail rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := resolveID(cmd, args, "rule-id")
		if err != nil {
			return err
		}

		conn, err := dialService(cmd, "guardrails")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := guardrailspb.NewGuardrailsServiceClient(conn)
		_, err = client.DeleteRule(cmd.Context(), &guardrailspb.DeleteRuleRequest{RuleId: id})
		if err != nil {
			return grpcError(err)
		}

		fmt.Printf("Rule %s deleted.\n", id)
		return nil
	},
}

func init() {
	grDeleteRuleCmd.Flags().String("rule-id", "", "Rule ID")
}

// --- Simulate ---

var grSimulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Simulate a policy against a tool call",
	RunE: func(cmd *cobra.Command, args []string) error {
		ruleIDs, _ := cmd.Flags().GetStringSlice("rule-ids")
		if len(ruleIDs) == 0 {
			return fmt.Errorf("--rule-ids is required")
		}
		toolName, _ := cmd.Flags().GetString("tool-name")
		if toolName == "" {
			return fmt.Errorf("--tool-name is required")
		}

		conn, err := dialService(cmd, "guardrails")
		if err != nil {
			return err
		}
		defer conn.Close()

		agentID, _ := cmd.Flags().GetString("agent-id")
		params, _ := cmd.Flags().GetStringToString("parameters")

		client := guardrailspb.NewGuardrailsServiceClient(conn)
		resp, err := client.SimulatePolicy(cmd.Context(), &guardrailspb.SimulatePolicyRequest{
			RuleIds:    ruleIDs,
			ToolName:   toolName,
			Parameters: params,
			AgentId:    agentID,
		})
		if err != nil {
			return grpcError(err)
		}

		simHeaders := []string{"VERDICT", "MATCHED RULE", "REASON"}
		rows := [][]string{{
			resp.GetVerdict(),
			fmt.Sprintf("%s (%s)", resp.GetMatchedRuleName(), resp.GetMatchedRuleId()),
			resp.GetReason(),
		}}
		return outputResult(cmd, simHeaders, rows, resp)
	},
}

func init() {
	grSimulateCmd.Flags().StringSlice("rule-ids", nil, "Rule IDs to evaluate (required)")
	grSimulateCmd.Flags().String("tool-name", "", "Tool name to test (required)")
	grSimulateCmd.Flags().StringToString("parameters", nil, "Tool parameters as key=value pairs")
	grSimulateCmd.Flags().String("agent-id", "", "Agent ID for context")
}

// --- Create Set ---

var grCreateSetCmd = &cobra.Command{
	Use:   "create-set",
	Short: "Create a named guardrail set",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}
		ruleIDs, _ := cmd.Flags().GetStringSlice("rule-ids")
		if len(ruleIDs) == 0 {
			return fmt.Errorf("--rule-ids is required")
		}

		conn, err := dialService(cmd, "guardrails")
		if err != nil {
			return err
		}
		defer conn.Close()

		desc, _ := cmd.Flags().GetString("description")

		client := guardrailspb.NewGuardrailsServiceClient(conn)
		resp, err := client.CreateGuardrailSet(cmd.Context(), &guardrailspb.CreateGuardrailSetRequest{
			Name:        name,
			Description: desc,
			RuleIds:     ruleIDs,
		})
		if err != nil {
			return grpcError(err)
		}

		s := resp.GetSet()
		return outputResult(cmd, setHeaders, [][]string{setRow(s)}, resp)
	},
}

func init() {
	grCreateSetCmd.Flags().String("name", "", "Set name (required)")
	grCreateSetCmd.Flags().String("description", "", "Set description")
	grCreateSetCmd.Flags().StringSlice("rule-ids", nil, "Rule IDs to include (required)")
}

// --- List Sets ---

var grListSetsCmd = &cobra.Command{
	Use:   "list-sets",
	Short: "List guardrail sets",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "guardrails")
		if err != nil {
			return err
		}
		defer conn.Close()

		limit, _ := cmd.Flags().GetInt32("limit")

		client := guardrailspb.NewGuardrailsServiceClient(conn)
		resp, err := client.ListGuardrailSets(cmd.Context(), &guardrailspb.ListGuardrailSetsRequest{
			PageSize: limit,
		})
		if err != nil {
			return grpcError(err)
		}

		rows := make([][]string, len(resp.GetSets()))
		for i, s := range resp.GetSets() {
			rows[i] = setRow(s)
		}
		return outputResult(cmd, setHeaders, rows, resp)
	},
}

func init() {
	grListSetsCmd.Flags().Int32("limit", 50, "Page size")
}

// parseRuleIDsCSV splits a comma-separated string into a slice.
func parseRuleIDsCSV(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Default ports matching docker-compose.yml.
var defaultPorts = map[string]string{
	"identity":   "50060",
	"workspace":  "50061",
	"guardrails": "50062",
	"human":      "50063",
	"governance": "50064",
	"activity":   "50065",
	"economics":  "50066",
	"compute":    "50067",
	"task":       "50068",
}

var rootCmd = &cobra.Command{
	Use:   "bkctl",
	Short: "Bulkhead CLI — manage agents, workspaces, guardrails, and more",
	Long: `bkctl is the operator CLI for the Bulkhead agent sandbox platform.

It connects to control plane services via gRPC to manage agents, workspaces,
guardrail policies, activity monitoring, and budgets.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().String("control-plane", "", "Control plane hostname (default: localhost, env: BKCTL_CONTROL_PLANE)")
	rootCmd.PersistentFlags().String("token", "", "Bearer token for authentication (env: BKCTL_TOKEN)")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format: table or json")

	// Per-service endpoint overrides.
	for svc := range defaultPorts {
		rootCmd.PersistentFlags().String(svc+"-endpoint", "", fmt.Sprintf("Override %s service endpoint (env: BKCTL_%s_ENDPOINT)", svc, envKey(svc)))
	}

	rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(guardrailCmd)
	rootCmd.AddCommand(activityCmd)
	rootCmd.AddCommand(budgetCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// resolveEndpoint returns the gRPC endpoint for a service.
func resolveEndpoint(cmd *cobra.Command, service string) string {
	// 1. Per-service flag.
	if ep, _ := cmd.Flags().GetString(service + "-endpoint"); ep != "" {
		return ep
	}

	// 2. Per-service env var.
	if ep := os.Getenv("BKCTL_" + envKey(service) + "_ENDPOINT"); ep != "" {
		return ep
	}

	// 3. Control plane host + default port.
	host := "localhost"
	if h, _ := cmd.Flags().GetString("control-plane"); h != "" {
		host = h
	} else if h := os.Getenv("BKCTL_CONTROL_PLANE"); h != "" {
		host = h
	}

	port := defaultPorts[service]
	return host + ":" + port
}

// getToken returns the bearer token from flag or env.
func getToken(cmd *cobra.Command) string {
	if t, _ := cmd.Flags().GetString("token"); t != "" {
		return t
	}
	return os.Getenv("BKCTL_TOKEN")
}

// getOutputFormat returns "table" or "json".
func getOutputFormat(cmd *cobra.Command) string {
	f, _ := cmd.Flags().GetString("output")
	if f == "json" {
		return "json"
	}
	return "table"
}

func envKey(service string) string {
	switch service {
	case "identity":
		return "IDENTITY"
	case "workspace":
		return "WORKSPACE"
	case "guardrails":
		return "GUARDRAILS"
	case "human":
		return "HUMAN"
	case "governance":
		return "GOVERNANCE"
	case "activity":
		return "ACTIVITY"
	case "economics":
		return "ECONOMICS"
	case "compute":
		return "COMPUTE"
	case "task":
		return "TASK"
	default:
		return service
	}
}

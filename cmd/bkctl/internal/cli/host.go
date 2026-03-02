package cli

import (
	"fmt"
	"strings"

	computepb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/compute/v1"
	"github.com/spf13/cobra"
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "View hosts and fleet capacity",
}

func init() {
	hostCmd.AddCommand(hostListCmd)
	hostCmd.AddCommand(hostCapacityCmd)
	hostCmd.AddCommand(hostConfigureWarmPoolCmd)
}

var hostHeaders = []string{"HOST ID", "ADDRESS", "STATUS", "MEMORY (AVAIL/TOTAL)", "CPU (AVAIL/TOTAL)", "SANDBOXES", "TIERS", "LAST HEARTBEAT"}

func hostRow(h *computepb.Host) []string {
	avail := h.GetAvailableResources()
	total := h.GetTotalResources()
	memStr := fmt.Sprintf("%d/%d MB", avail.GetMemoryMb(), total.GetMemoryMb())
	cpuStr := fmt.Sprintf("%d/%d m", avail.GetCpuMillicores(), total.GetCpuMillicores())
	return []string{
		h.GetHostId(),
		h.GetAddress(),
		formatHostStatus(h.GetStatus()),
		memStr,
		cpuStr,
		fmt.Sprintf("%d", h.GetActiveSandboxes()),
		strings.Join(h.GetSupportedTiers(), ","),
		formatTimestamp(h.GetLastHeartbeat()),
	}
}

func formatHostStatus(s computepb.HostStatus) string {
	switch s {
	case computepb.HostStatus_HOST_STATUS_READY:
		return "ready"
	case computepb.HostStatus_HOST_STATUS_DRAINING:
		return "draining"
	case computepb.HostStatus_HOST_STATUS_OFFLINE:
		return "offline"
	default:
		return "unspecified"
	}
}

func parseHostStatus(s string) computepb.HostStatus {
	switch s {
	case "ready":
		return computepb.HostStatus_HOST_STATUS_READY
	case "draining":
		return computepb.HostStatus_HOST_STATUS_DRAINING
	case "offline":
		return computepb.HostStatus_HOST_STATUS_OFFLINE
	default:
		return computepb.HostStatus_HOST_STATUS_UNSPECIFIED
	}
}

// --- List ---

var hostListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered hosts",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "compute")
		if err != nil {
			return err
		}
		defer conn.Close()

		status, _ := cmd.Flags().GetString("status")

		client := computepb.NewComputePlaneServiceClient(conn)
		resp, err := client.ListHosts(cmd.Context(), &computepb.ListHostsRequest{
			Status: parseHostStatus(status),
		})
		if err != nil {
			return grpcError(err)
		}

		rows := make([][]string, len(resp.GetHosts()))
		for i, h := range resp.GetHosts() {
			rows[i] = hostRow(h)
		}
		return outputResult(cmd, hostHeaders, rows, resp)
	},
}

func init() {
	hostListCmd.Flags().String("status", "", "Filter by status: ready, draining, offline")
}

// --- Capacity ---

var hostCapacityCmd = &cobra.Command{
	Use:   "capacity",
	Short: "Show fleet capacity by isolation tier",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := dialService(cmd, "compute")
		if err != nil {
			return err
		}
		defer conn.Close()

		client := computepb.NewComputePlaneServiceClient(conn)
		resp, err := client.GetCapacity(cmd.Context(), &computepb.GetCapacityRequest{})
		if err != nil {
			return grpcError(err)
		}

		capHeaders := []string{"TIER", "HOSTS", "AVAIL MEMORY MB", "AVAIL CPU MILLI", "WARM TARGET", "WARM READY"}
		rows := make([][]string, len(resp.GetTiers()))
		for i, t := range resp.GetTiers() {
			rows[i] = []string{
				t.GetIsolationTier(),
				fmt.Sprintf("%d", t.GetHostsSupporting()),
				fmt.Sprintf("%d", t.GetAvailableMemoryMb()),
				fmt.Sprintf("%d", t.GetAvailableCpuMillicores()),
				fmt.Sprintf("%d", t.GetWarmSlotsTarget()),
				fmt.Sprintf("%d", t.GetWarmSlotsReady()),
			}
		}

		// Append summary row.
		if len(rows) > 0 {
			rows = append(rows, []string{
				fmt.Sprintf("TOTAL (%d/%d hosts ready)", resp.GetReadyHosts(), resp.GetTotalHosts()),
				"", "", "", "", "",
			})
		}

		return outputResult(cmd, capHeaders, rows, resp)
	},
}

// --- Configure Warm Pool ---

var hostConfigureWarmPoolCmd = &cobra.Command{
	Use:   "configure-warm-pool",
	Short: "Configure warm pool for an isolation tier",
	RunE: func(cmd *cobra.Command, args []string) error {
		tier, _ := cmd.Flags().GetString("isolation-tier")
		if tier == "" {
			return fmt.Errorf("--isolation-tier is required")
		}
		targetCount, _ := cmd.Flags().GetInt32("target-count")
		if targetCount <= 0 {
			return fmt.Errorf("--target-count must be > 0")
		}

		conn, err := dialService(cmd, "compute")
		if err != nil {
			return err
		}
		defer conn.Close()

		memory, _ := cmd.Flags().GetInt64("memory")
		cpu, _ := cmd.Flags().GetInt32("cpu")
		disk, _ := cmd.Flags().GetInt64("disk")

		client := computepb.NewComputePlaneServiceClient(conn)
		resp, err := client.ConfigureWarmPool(cmd.Context(), &computepb.ConfigureWarmPoolRequest{
			Config: &computepb.WarmPoolConfig{
				IsolationTier: tier,
				TargetCount:   targetCount,
				MemoryMb:      memory,
				CpuMillicores: cpu,
				DiskMb:        disk,
			},
		})
		if err != nil {
			return grpcError(err)
		}

		cfg := resp.GetConfig()
		wpHeaders := []string{"TIER", "TARGET COUNT", "MEMORY MB", "CPU MILLI", "DISK MB"}
		rows := [][]string{{
			cfg.GetIsolationTier(),
			fmt.Sprintf("%d", cfg.GetTargetCount()),
			fmt.Sprintf("%d", cfg.GetMemoryMb()),
			fmt.Sprintf("%d", cfg.GetCpuMillicores()),
			fmt.Sprintf("%d", cfg.GetDiskMb()),
		}}
		return outputResult(cmd, wpHeaders, rows, resp)
	},
}

func init() {
	hostConfigureWarmPoolCmd.Flags().String("isolation-tier", "", "Isolation tier: standard, hardened, isolated (required)")
	hostConfigureWarmPoolCmd.Flags().Int32("target-count", 0, "Number of pre-warmed slots (required)")
	hostConfigureWarmPoolCmd.Flags().Int64("memory", 512, "Memory per warm slot in MB")
	hostConfigureWarmPoolCmd.Flags().Int32("cpu", 1000, "CPU per warm slot in millicores")
	hostConfigureWarmPoolCmd.Flags().Int64("disk", 1024, "Disk per warm slot in MB")
}

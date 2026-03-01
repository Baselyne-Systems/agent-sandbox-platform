package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print bkctl version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("bkctl %s (commit: %s, built: %s, %s/%s)\n",
			version, commit, date, runtime.GOOS, runtime.GOARCH)
	},
}

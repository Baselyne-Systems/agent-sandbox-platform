package main

import (
	"os"

	"github.com/Baselyne-Systems/bulkhead/control-plane/cmd/bkctl/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

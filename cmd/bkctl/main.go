package main

import (
	"os"

	"github.com/achyuthnsamudrala/bulkhead/cmd/bkctl/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

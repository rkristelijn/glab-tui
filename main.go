package main

import (
	"os"

	"github.com/rkristelijn/glab-tui/cmd/cli"
)

func main() {
	// Check if we should run TUI or CLI mode
	if len(os.Args) > 1 {
		// CLI mode - traditional glab-style commands
		cli.Run(os.Args[1:])
	} else {
		// Default: auto mode (try local repo first, then selector)
		cli.Run([]string{"auto"})
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/rkristelijn/glab-tui/cmd/cli"
	"github.com/rkristelijn/glab-tui/cmd/tui"
)

func main() {
	// Check if we should run TUI or CLI mode
	if len(os.Args) > 1 {
		// CLI mode - traditional glab-style commands
		cli.Run(os.Args[1:])
	} else {
		// TUI mode - interactive interface (default)
		if err := tui.Run(); err != nil {
			fmt.Printf("TUI error: %v\n", err)
			os.Exit(1)
		}
	}
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show help information",
	Long:  `Show help information for sinkzone commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Sinkzone - DNS-based productivity tool")
		fmt.Println("=====================================")
		fmt.Println()
		fmt.Println("Available commands:")
		fmt.Println("  sinkzone          - Start the TUI interface")
		fmt.Println("  sinkzone help     - Show this help")
		fmt.Println("  sinkzone resolver - Start the DNS resolver (requires root)")
		fmt.Println("  sinkzone focus    - Switch to focus mode with duration")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  sinkzone                    # Start TUI")
		fmt.Println("  sudo sinkzone resolver      # Start DNS server")
		fmt.Println("  sinkzone focus 1h           # Focus for 1 hour")
		fmt.Println("  sinkzone focus 30m          # Focus for 30 minutes")
	},
}

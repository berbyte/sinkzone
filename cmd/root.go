package cmd

import (
	"github.com/berbyte/sinkzone/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sinkzone",
	Short: "A productivity tool for DNS-based focus mode",
	Long: `Sinkzone is a lightweight DNS resolver that helps you stay focused
by blocking distracting websites during focus sessions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand is provided, start the TUI
		return tui.Start()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(resolverCmd)
	rootCmd.AddCommand(focusCmd)
	rootCmd.AddCommand(helpCmd)
}

package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sinkzone",
	Short: "DNS-based productivity tool",
	Long: `Sinkzone is a DNS-based productivity tool that helps you stay focused by blocking distracting websites in real time.

It works by intercepting DNS requests and enforcing a focus mode, where only allowed domains are accessible.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand is provided, show help
		return cmd.Help()
	},
}

func Execute() error {
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(focusCmd)
	rootCmd.AddCommand(resolverCmd)
	rootCmd.AddCommand(allowlistCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(manCmd)
	return rootCmd.Execute()
}

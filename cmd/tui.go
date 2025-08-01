package cmd

import (
	"github.com/berbyte/sinkzone/internal/tui"
	"github.com/spf13/cobra"
)

var tuiAPIURL string

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the interactive user interface",
	Long:  `The TUI provides a more visual way to manage your resolver, monitor traffic, update the allowlist, and control focus mode — all in one place.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.StartWithAPIURL(tuiAPIURL)
	},
}

func init() {
	tuiCmd.Flags().StringVarP(&tuiAPIURL, "api-url", "u", "http://localhost:8080", "URL of the resolver API")
}

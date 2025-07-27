package cmd

import (
	"fmt"
	"time"

	"github.com/berbyte/sinkzone/internal/config"
	"github.com/berbyte/sinkzone/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sinkzone",
	Short: "DNS-based productivity tool",
	Long:  `Sinkzone is a DNS-based productivity tool that helps you stay focused by blocking distracting websites during focus sessions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand is provided, start the TUI
		return tui.Start()
	},
}

var disableFocusCmd = &cobra.Command{
	Use:   "disable-focus",
	Short: "Disable focus mode",
	Long:  `Disable focus mode immediately.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize state manager
		stateMgr, err := config.NewStateManager()
		if err != nil {
			return fmt.Errorf("failed to initialize state manager: %w", err)
		}

		// Disable focus mode
		if err := stateMgr.SetFocusMode(false, 0); err != nil {
			return fmt.Errorf("failed to disable focus mode: %w", err)
		}

		fmt.Printf("Focus mode disabled. All domains will be allowed.\n")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show focus mode status",
	Long:  `Show the current focus mode status and remaining time.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize state manager
		stateMgr, err := config.NewStateManager()
		if err != nil {
			return fmt.Errorf("failed to initialize state manager: %w", err)
		}

		state := stateMgr.GetState()

		if state.FocusMode {
			if state.FocusEndTime != nil {
				remaining := time.Until(*state.FocusEndTime)
				if remaining > 0 {
					fmt.Printf("Focus mode: ENABLED\n")
					fmt.Printf("Remaining time: %s\n", remaining.Round(time.Minute))
					fmt.Printf("Ends at: %s\n", state.FocusEndTime.Format("15:04:05"))
				} else {
					fmt.Printf("Focus mode: EXPIRED\n")
					fmt.Printf("Ended at: %s\n", state.FocusEndTime.Format("15:04:05"))
				}
			} else {
				fmt.Printf("Focus mode: ENABLED (no expiration)\n")
			}
		} else {
			fmt.Printf("Focus mode: DISABLED\n")
		}

		fmt.Printf("Last updated: %s\n", state.LastUpdated.Format("15:04:05"))
		return nil
	},
}

func Execute() error {
	rootCmd.AddCommand(resolverCmd)
	rootCmd.AddCommand(focusCmd)
	rootCmd.AddCommand(disableFocusCmd)
	rootCmd.AddCommand(statusCmd)
	return rootCmd.Execute()
}

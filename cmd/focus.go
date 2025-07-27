package cmd

import (
	"fmt"
	"time"

	"github.com/berbyte/sinkzone/internal/config"
	"github.com/spf13/cobra"
)

var focusCmd = &cobra.Command{
	Use:   "focus [duration]",
	Short: "Switch to focus mode",
	Long:  `Switch to focus mode for the specified duration. Duration can be specified as "1h", "30m", etc.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		duration, err := time.ParseDuration(args[0])
		if err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}

		// Initialize state manager
		stateMgr, err := config.NewStateManager()
		if err != nil {
			return fmt.Errorf("failed to initialize state manager: %w", err)
		}

		// Enable focus mode with duration
		if err := stateMgr.SetFocusMode(true, duration); err != nil {
			return fmt.Errorf("failed to enable focus mode: %w", err)
		}

		endTime := time.Now().Add(duration)
		fmt.Printf("Focus mode activated for %s (until %s)\n", duration, endTime.Format("15:04:05"))
		fmt.Printf("DNS resolver will block non-allowlisted domains immediately.\n")
		return nil
	},
}

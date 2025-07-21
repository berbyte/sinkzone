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

		// Load current config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Set focus mode
		cfg.Mode = "focus"
		endTime := time.Now().Add(duration)
		cfg.FocusEndTime = &endTime

		// Save config
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Focus mode activated for %s (until %s)\n", duration, cfg.FocusEndTime.Format("15:04:05"))
		return nil
	},
}

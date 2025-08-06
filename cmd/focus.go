package cmd

import (
	"fmt"
	"time"

	"github.com/berbyte/sinkzone/internal/api"
	"github.com/berbyte/sinkzone/internal/config"
	"github.com/spf13/cobra"
)

var (
	focusEnable   bool
	focusDisable  bool
	focusDuration string
	focusAPIURL   string
)

var focusCmd = &cobra.Command{
	Use:   "focus [command]",
	Short: "Manage focus mode",
	Long: `Enables or disables focus mode, which blocks all non-allowlisted domains.

Focus mode is the core productivity feature in Sinkzone. When enabled, only DNS requests to domains on your allowlist will be resolved — everything else is silently blocked.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle subcommands
		if len(args) > 0 {
			switch args[0] {
			case "start":
				return enableFocusMode(1 * time.Hour)
			default:
				return fmt.Errorf("unknown command: %s", args[0])
			}
		}

		// Handle flags
		if focusDisable {
			return disableFocusMode()
		}

		if focusEnable {
			duration := 1 * time.Hour // Default 1 hour
			if focusDuration != "" {
				var err error
				duration, err = time.ParseDuration(focusDuration)
				if err != nil {
					return fmt.Errorf("invalid duration format: %w", err)
				}
			}
			return enableFocusMode(duration)
		}

		// If no args or flags, show help
		return cmd.Help()
	},
}

func init() {
	focusCmd.Flags().BoolVar(&focusEnable, "enable", false, "Enable focus mode")
	focusCmd.Flags().BoolVar(&focusDisable, "disable", false, "Disable focus mode")
	focusCmd.Flags().StringVar(&focusDuration, "duration", "", "Duration for focus mode (e.g., '1h', '30m')")
	focusCmd.Flags().StringVar(&focusAPIURL, "api-url", "http://127.0.0.1:8080", "URL of the resolver API")
}

func enableFocusMode(duration time.Duration) error {
	// Create API client
	client := api.NewClient(focusAPIURL)

	// Try to connect to API
	if err := client.HealthCheck(); err != nil {
		return config.AdminError(err, "failed to connect to resolver API")
	}

	// Set focus mode via API
	if err := client.SetFocusMode(true, duration.String()); err != nil {
		return fmt.Errorf("failed to enable focus mode: %w", err)
	}

	endTime := time.Now().Add(duration)
	fmt.Printf("Focus mode activated for %s (until %s)\n", duration, endTime.Format("15:04:05"))
	fmt.Printf("DNS resolver will block non-allowlisted domains immediately.\n")
	return nil
}

func disableFocusMode() error {
	// Create API client
	client := api.NewClient(focusAPIURL)

	// Try to connect to API
	if err := client.HealthCheck(); err != nil {
		return config.AdminError(err, "failed to connect to resolver API")
	}

	// Set focus mode via API
	if err := client.SetFocusMode(false, ""); err != nil {
		return fmt.Errorf("failed to disable focus mode: %w", err)
	}

	fmt.Printf("Focus mode disabled. All domains will be allowed.\n")
	return nil
}

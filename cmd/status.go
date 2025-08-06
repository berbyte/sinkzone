package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/berbyte/sinkzone/internal/api"
	"github.com/berbyte/sinkzone/internal/config"
	"github.com/spf13/cobra"
)

// getPIDFilePath returns the platform-specific path for the PID file
func getPIDFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if runtime.GOOS == "windows" {
		// On Windows, use AppData for better compatibility
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "sinkzone", "resolver.pid"), nil
		}
		// Fallback to user home directory
		return filepath.Join(homeDir, "sinkzone", "resolver.pid"), nil
	}

	// Unix-like systems use ~/.sinkzone/
	return filepath.Join(homeDir, ".sinkzone", "resolver.pid"), nil
}

var statusAPIURL string

var statusCmd = &cobra.Command{
	Use:   "status [type]",
	Short: "Show system status",
	Long: `Displays the current state of the Sinkzone system, including:

- Whether the resolver is running
- If focus mode is active

Use this to get a quick overview of what Sinkzone is doing.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return showGeneralStatus()
		}

		switch args[0] {
		case "resolver":
			return showResolverStatus()
		case "focus":
			return showFocusStatus()
		default:
			return fmt.Errorf("unknown status type: %s. Use 'resolver' or 'focus'", args[0])
		}
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusAPIURL, "api-url", "http://127.0.0.1:8080", "URL of the resolver API")
}

func showGeneralStatus() error {
	fmt.Println("=== Sinkzone Status ===")

	// Show focus status
	if err := showFocusStatus(); err != nil {
		return err
	}

	fmt.Println()

	// Show resolver status
	if err := showResolverStatus(); err != nil {
		return err
	}

	return nil
}

func showResolverStatus() error {
	pidFile, err := getPIDFilePath()
	if err != nil {
		return fmt.Errorf("failed to get PID file path: %w", err)
	}

	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		fmt.Println("Resolver: NOT RUNNING")
		return nil
	}

	// #nosec G304 -- pidFile is a hardcoded path from user home directory
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Println("Resolver: UNKNOWN (cannot read PID file)")
		return nil
	}

	fmt.Printf("Resolver: RUNNING (PID: %s)\n", string(pidData))
	return nil
}

func showFocusStatus() error {
	// Try to get focus mode state from API first
	client := api.NewClient(statusAPIURL)
	if err := client.HealthCheck(); err == nil {
		focusState, err := client.GetFocusMode()
		if err != nil {
			return fmt.Errorf("failed to get focus mode state: %w", err)
		}

		if focusState.Enabled {
			if focusState.EndTime != nil {
				remaining := time.Until(*focusState.EndTime)
				if remaining > 0 {
					fmt.Printf("Focus mode: ENABLED\n")
					fmt.Printf("Remaining time: %s\n", remaining.Round(time.Minute))
					fmt.Printf("Ends at: %s\n", focusState.EndTime.Format("15:04:05"))
				} else {
					fmt.Printf("Focus mode: EXPIRED\n")
					fmt.Printf("Ended at: %s\n", focusState.EndTime.Format("15:04:05"))
				}
			} else {
				fmt.Printf("Focus mode: ENABLED (no expiration)\n")
			}
		} else {
			fmt.Printf("Focus mode: DISABLED\n")
		}

		fmt.Printf("Last updated: %s\n", time.Now().Format("15:04:05"))
		return nil
	}

	// Fallback to state manager if API is not available
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
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/berbyte/sinkzone/config"

	"github.com/spf13/cobra"
)

var modeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Manage filtering modes",
	Long:  "Switch between different filtering modes: monitor, focus, lockdown, and off",
}

var modeFocusCmd = &cobra.Command{
	Use:   "focus",
	Short: "Switch to Focus Mode",
	Long:  "Enable Focus Mode - blocks social media and entertainment sites",
	RunE: func(cmd *cobra.Command, args []string) error {
		db := getDB()
		if err := db.SetMode("focus"); err != nil {
			return fmt.Errorf("failed to set mode: %v", err)
		}
		fmt.Println("Switched to Focus Mode")
		fmt.Println("Blocking: Social media, entertainment, and distracting sites")
		return nil
	},
}

var modeLockdownCmd = &cobra.Command{
	Use:   "lockdown",
	Short: "Switch to Lockdown Mode",
	Long:  "Enable Lockdown Mode - blocks most sites except essential ones (requires PIN)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()

		// Prompt for confirmation
		fmt.Print("Are you sure you want to enter Lockdown Mode? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Lockdown Mode cancelled")
			return nil
		}

		// Prompt for PIN
		fmt.Print("Enter PIN: ")
		pin, _ := reader.ReadString('\n')
		pin = strings.TrimSpace(pin)

		if !config.VerifyPIN(pin, cfg.PIN) {
			return fmt.Errorf("incorrect PIN")
		}

		db := getDB()
		if err := db.SetMode("lockdown"); err != nil {
			return fmt.Errorf("failed to set mode: %v", err)
		}

		fmt.Println("Switched to Lockdown Mode")
		fmt.Println("Blocking: All sites except essential ones")
		return nil
	},
}

var modeMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Switch to Monitor Mode",
	Long:  "Enable Monitor Mode - logs all queries but doesn't block anything",
	RunE: func(cmd *cobra.Command, args []string) error {
		db := getDB()
		if err := db.SetMode("monitor"); err != nil {
			return fmt.Errorf("failed to set mode: %v", err)
		}
		fmt.Println("Switched to Monitor Mode")
		fmt.Println("Logging: All DNS queries (no blocking)")
		return nil
	},
}

var modeOffCmd = &cobra.Command{
	Use:   "off",
	Short: "Disable filtering",
	Long:  "Disable all filtering - acts as a simple DNS forwarder",
	RunE: func(cmd *cobra.Command, args []string) error {
		db := getDB()
		if err := db.SetMode("off"); err != nil {
			return fmt.Errorf("failed to set mode: %v", err)
		}
		fmt.Println("Filtering disabled")
		fmt.Println("Mode: Simple DNS forwarder")
		return nil
	},
}

var modeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current mode",
	Long:  "Display the current filtering mode and status",
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		mode, err := db.GetMode()
		if err != nil {
			fmt.Printf("Error getting mode: %v\n", err)
			return
		}

		fmt.Printf("Current mode: %s\n", mode)

		switch mode {
		case "off":
			fmt.Println("Status: No filtering - simple DNS forwarder")
		case "monitor":
			fmt.Println("Status: Logging all queries (no blocking)")
		case "focus":
			fmt.Println("Status: Blocking distracting sites")
		case "lockdown":
			fmt.Println("Status: Blocking most sites (PIN protected)")
		default:
			fmt.Println("Status: Unknown mode")
		}
	},
}

var modeUnlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Exit Lockdown Mode",
	Long:  "Exit Lockdown Mode by entering your PIN",
	RunE: func(cmd *cobra.Command, args []string) error {
		db := getDB()
		currentMode, err := db.GetMode()
		if err != nil {
			return fmt.Errorf("failed to get current mode: %v", err)
		}

		if currentMode != "lockdown" {
			return fmt.Errorf("not currently in Lockdown Mode")
		}

		cfg := getConfig()

		// Prompt for PIN
		fmt.Print("Enter PIN to exit Lockdown Mode: ")
		reader := bufio.NewReader(os.Stdin)
		pin, _ := reader.ReadString('\n')
		pin = strings.TrimSpace(pin)

		if !config.VerifyPIN(pin, cfg.PIN) {
			return fmt.Errorf("incorrect PIN")
		}

		if err := db.SetMode("off"); err != nil {
			return fmt.Errorf("failed to set mode: %v", err)
		}

		fmt.Println("Exited Lockdown Mode")
		fmt.Println("Switched to: No filtering")
		return nil
	},
}

func init() {
	modeCmd.AddCommand(modeFocusCmd, modeLockdownCmd, modeMonitorCmd, modeOffCmd, modeStatusCmd, modeUnlockCmd)
	rootCmd.AddCommand(modeCmd)
}

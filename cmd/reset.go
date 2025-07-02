package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all application state",
	Long:  "Clear all domain rules, DNS query history, and reset to default settings (requires confirmation)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("WARNING: This will reset all sinkzone data!")
		fmt.Println("This includes:")
		fmt.Println("  - All domain rules (allowlist and blocklist)")
		fmt.Println("  - All DNS query history")
		fmt.Println("  - Current mode (will be set to 'off')")
		fmt.Println("")
		fmt.Print("Are you sure you want to continue? (type 'yes' to confirm): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		if response != "yes" {
			fmt.Println("Reset cancelled")
			return nil
		}

		db := getDB()
		if err := db.Reset(); err != nil {
			return fmt.Errorf("failed to reset database: %v", err)
		}

		fmt.Println("All sinkzone data has been reset")
		fmt.Println("Mode set to: off")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}

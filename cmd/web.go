package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start web dashboard",
	Long:  "Start the web dashboard for managing sinkzone (coming soon)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Web dashboard is not yet implemented")
		fmt.Println("This feature is planned for a future release")
		fmt.Println("For now, use the CLI commands to manage sinkzone")
	},
}

func init() {
	rootCmd.AddCommand(webCmd)
}

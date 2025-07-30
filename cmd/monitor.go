package cmd

import (
	"fmt"
	"time"

	"github.com/berbyte/sinkzone/internal/socket"
	"github.com/spf13/cobra"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "View recent DNS requests",
	Long: `Displays a live feed of DNS queries captured by the Sinkzone resolver.

Use this to observe which domains your system is accessing in real time. It's especially useful when configuring your allowlist — you'll see which domains need to be permitted for tools or websites you want to use during focus sessions.

Make sure the resolver is running before using this command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create socket client
		client := socket.NewClient()

		// Try to connect to socket
		if err := client.Connect(); err != nil {
			return fmt.Errorf("failed to connect to resolver socket: %w\nMake sure the resolver is running with 'sudo sinkzone resolver'", err)
		}
		fmt.Printf("Connected successfully!\n")
		defer client.Disconnect()

		// Wait a moment for data to be received
		time.Sleep(100 * time.Millisecond)

		// Get recent queries
		queries := client.GetQueries()

		if len(queries) == 0 {
			fmt.Println("No DNS queries recorded yet.")
			fmt.Println("Try making some web requests to see DNS activity.")
			return nil
		}

		// Show last 20 queries (or all if less than 20)
		start := 0
		if len(queries) > 20 {
			start = len(queries) - 20
		}

		fmt.Printf("Last %d DNS requests:\n\n", len(queries[start:]))
		fmt.Printf("%-40s %-10s %-20s %s\n", "Domain", "Status", "Time", "Blocked")
		fmt.Println(string(make([]byte, 80, 80)))

		for _, query := range queries[start:] {
			status := "ALLOWED"
			if query.Blocked {
				status = "BLOCKED"
			}

			timeStr := query.Timestamp.Format("15:04:05")
			blockedStr := "No"
			if query.Blocked {
				blockedStr = "Yes"
			}

			// Truncate domain if too long
			domain := query.Domain
			if len(domain) > 38 {
				domain = domain[:35] + "..."
			}

			fmt.Printf("%-40s %-10s %-20s %s\n", domain, status, timeStr, blockedStr)
		}

		fmt.Printf("\nTotal queries: %d\n", len(queries))
		return nil
	},
}

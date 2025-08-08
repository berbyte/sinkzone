package cmd

import (
	"fmt"

	"github.com/berbyte/sinkzone/internal/api"
	"github.com/berbyte/sinkzone/internal/config"
	"github.com/spf13/cobra"
)

var apiURL string

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "View recent DNS requests",
	Long: `Displays a live feed of DNS queries captured by the Sinkzone resolver.

Use this to observe which domains your system is accessing in real time. It's especially useful when configuring your allowlist — you'll see which domains need to be permitted for tools or websites you want to use during focus sessions.

Make sure the resolver is running before using this command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create API client
		client := api.NewClient(apiURL)

		// Try to connect to API
		if err := client.HealthCheck(); err != nil {
			return config.AdminError(err, "failed to connect to resolver API")
		}
		fmt.Printf("Connected successfully!\n")

		// Get recent queries
		queries, err := client.GetQueries()
		if err != nil {
			return fmt.Errorf("failed to get queries: %w", err)
		}

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
		fmt.Printf("%-40s %-27s %-10s %-20s %s\n", "Domain", "Client", "Status", "Time", "Blocked")
		fmt.Println(string(make([]byte, 80)))

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

			// Truncate hostname if too long
			dnsClient := query.Client
			if len(dnsClient) > 25 {
				dnsClient = dnsClient[:22] + "..."
			}

			fmt.Printf("%-40s %-27s %-10s %-20s %s\n", domain, dnsClient, status, timeStr, blockedStr)
		}

		fmt.Printf("\nTotal queries: %d\n", len(queries))
		return nil
	},
}

func init() {
	monitorCmd.Flags().StringVarP(&apiURL, "api-url", "u", "http://127.0.0.1:8080", "URL of the resolver API")
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show DNS query statistics",
	Long:  "Display statistics about DNS queries, blocked requests, and client activity",
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		stats, err := db.GetStats()
		if err != nil {
			fmt.Printf("Error getting stats: %v\n", err)
			return
		}

		fmt.Println("DNS Query Statistics")
		fmt.Println("====================")
		fmt.Printf("Total Queries: %d\n", stats["total_queries"])
		fmt.Printf("Blocked Queries: %d\n", stats["blocked_queries"])
		fmt.Printf("Unique Clients: %d\n", stats["unique_clients"])
		fmt.Printf("Unique Domains: %d\n", stats["unique_domains"])
		fmt.Printf("Current Mode: %s\n", stats["current_mode"])

		// Calculate block rate
		total := stats["total_queries"].(int)
		blocked := stats["blocked_queries"].(int)
		if total > 0 {
			blockRate := float64(blocked) / float64(total) * 100
			fmt.Printf("Block Rate: %.1f%%\n", blockRate)
		}
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

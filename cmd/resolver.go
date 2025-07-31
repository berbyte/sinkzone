package cmd

import (
	"fmt"
	"log"

	"github.com/berbyte/sinkzone/internal/config"
	"github.com/berbyte/sinkzone/internal/dns"
	"github.com/spf13/cobra"
)

var port string

var resolverCmd = &cobra.Command{
	Use:   "resolver",
	Short: "Start the local DNS resolver (required first step)",
	Long: `Starts Sinkzone's local DNS resolver, which intercepts DNS queries made by your system.

This must be the first command you run. The resolver captures all outgoing DNS requests and enables Sinkzone to monitor and control domain access.

Once running, other features like monitoring, allowlisting, and focus mode become active.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create DNS server with specified port
		server := dns.NewServerWithPort(cfg, port)

		log.Printf("Starting sinkzone DNS resolver on :%s", port)
		return server.Start()
	},
}

func init() {
	resolverCmd.Flags().StringVarP(&port, "port", "p", "53", "Port to bind the DNS server to")
	rootCmd.AddCommand(resolverCmd)
}

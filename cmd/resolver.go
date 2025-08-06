package cmd

import (
	"fmt"
	"log"
	"sync"

	"github.com/berbyte/sinkzone/internal/api"
	"github.com/berbyte/sinkzone/internal/config"
	"github.com/berbyte/sinkzone/internal/dns"
	"github.com/spf13/cobra"
)

var port string
var apiPort string

var resolverCmd = &cobra.Command{
	Use:   "resolver",
	Short: "Start the local DNS resolver with HTTP API (required first step)",
	Long: `Starts Sinkzone's local DNS resolver with HTTP API, which intercepts DNS queries made by your system.

This must be the first command you run. The resolver captures all outgoing DNS requests and enables Sinkzone to monitor and control domain access.

The HTTP API provides endpoints for:
- GET /api/queries - Get last 100 DNS queries
- GET /api/focus - Get focus mode state
- POST /api/focus - Set focus mode
- GET /api/state - Get complete resolver state

Once running, other features like monitoring, allowlisting, and focus mode become active.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check admin privileges for privileged ports
		if err := config.CheckPortPrivileges(port); err != nil {
			return err
		}

		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create API server
		apiServer := api.NewServer(apiPort)

		// Create DNS server with API server reference
		dnsServer := dns.NewServerWithPort(cfg, apiServer, port)

		log.Printf("Starting sinkzone DNS resolver on :%s with API on :%s", port, apiPort)

		// Start both servers in goroutines
		var wg sync.WaitGroup
		var dnsErr, apiErr error

		wg.Add(2)

		// Start DNS server
		go func() {
			defer wg.Done()
			dnsErr = dnsServer.Start()
		}()

		// Start API server
		go func() {
			defer wg.Done()
			apiErr = apiServer.Start()
		}()

		// Wait for both servers to finish (or error)
		wg.Wait()

		// Return the first error that occurred
		if dnsErr != nil {
			return fmt.Errorf("DNS server error: %w", dnsErr)
		}
		if apiErr != nil {
			return fmt.Errorf("API server error: %w", apiErr)
		}

		return nil
	},
}

func init() {
	resolverCmd.Flags().StringVarP(&port, "port", "p", "53", "Port to bind the DNS server to")
	resolverCmd.Flags().StringVarP(&apiPort, "api-port", "a", "8080", "Port to bind the HTTP API server to")
}

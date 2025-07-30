package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/berbyte/sinkzone/internal/config"
	"github.com/berbyte/sinkzone/internal/dns"
	"github.com/spf13/cobra"
)

var resolverCmd = &cobra.Command{
	Use:   "resolver",
	Short: "Start the local DNS resolver (required first step)",
	Long: `Starts Sinkzone's local DNS resolver, which intercepts DNS queries made by your system.

This must be the first command you run. The resolver captures all outgoing DNS requests and enables Sinkzone to monitor and control domain access.

Once running, other features like monitoring, allowlisting, and focus mode become active.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running with administrator/root privileges (required for port 53)
		if os.Geteuid() != 0 {
			return fmt.Errorf("resolver command must be run as root (sudo) to bind to port 53")
		}

		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create DNS server
		server := dns.NewServer(cfg)

		log.Println("Starting sinkzone DNS resolver on :53")
		return server.Start()
	},
}

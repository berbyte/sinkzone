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
	Short: "Start the DNS resolver service",
	Long:  `Start the DNS resolver service that handles DNS requests and applies focus mode rules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running as root (required for port 53)
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

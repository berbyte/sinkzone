package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/berbyte/sinkzone/internal/config"
	"github.com/berbyte/sinkzone/internal/dns"
	"github.com/spf13/cobra"
)

// isAdmin checks if the current process is running with administrator/root privileges
func isAdmin() bool {
	if runtime.GOOS == "windows" {
		// On Windows, check if we can access a privileged resource
		// This is a simple heuristic - try to open a system directory
		_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		return err == nil
	}
	// On Unix systems, check if effective user ID is 0 (root)
	return os.Geteuid() == 0
}

var resolverCmd = &cobra.Command{
	Use:   "resolver",
	Short: "Start the DNS resolver service",
	Long:  `Start the DNS resolver service that handles DNS requests and applies focus mode rules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running with administrator/root privileges (required for port 53)
		if !isAdmin() {
			if runtime.GOOS == "windows" {
				return fmt.Errorf("resolver command must be run as Administrator to bind to port 53")
			}
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

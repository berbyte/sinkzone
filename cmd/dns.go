package cmd

import (
	"fmt"

	"github.com/berbyte/sinkzone/dns"

	"github.com/spf13/cobra"
)

var dnsServer *dns.Server

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Manage DNS server",
	Long:  "Start, stop, and check status of the DNS filtering server",
}

var dnsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start DNS filtering server",
	Long:  "Start the DNS server on port 53 (requires root privileges)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dnsServer != nil && dnsServer.IsRunning() {
			return fmt.Errorf("DNS server is already running")
		}

		cfg := getConfig()
		db := getDB()

		dnsServer = dns.NewServer(cfg.UpstreamDNS, db)
		return dnsServer.Start()
	},
}

var dnsStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop DNS filtering server",
	Long:  "Stop the running DNS server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dnsServer == nil || !dnsServer.IsRunning() {
			return fmt.Errorf("DNS server is not running")
		}

		return dnsServer.Stop()
	},
}

var dnsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show DNS server status",
	Long:  "Display the current status of the DNS server",
	Run: func(cmd *cobra.Command, args []string) {
		if dnsServer == nil {
			fmt.Println("DNS server: Not initialized")
			return
		}

		if dnsServer.IsRunning() {
			fmt.Println("DNS server: Running")
			fmt.Printf("Forwarding to: %s\n", getConfig().UpstreamDNS)
		} else {
			fmt.Println("DNS server: Stopped")
		}
	},
}

func init() {
	dnsCmd.AddCommand(dnsStartCmd, dnsStopCmd, dnsStatusCmd)
	rootCmd.AddCommand(dnsCmd)
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/berbyte/sinkzone/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [get/set] [key] [value]",
	Short: "Manage configuration",
	Long:  `Manage sinkzone configuration. Currently supports setting resolver IP addresses.`,
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := args[0]
		key := args[1]
		value := args[2]

		switch command {
		case "set":
			return setConfig(key, value)
		case "get":
			return getConfig(key)
		default:
			return fmt.Errorf("unknown command: %s. Use 'set'", command)
		}
	},
}

func setConfig(key, value string) error {
	// Load existing config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch key {
	case "resolver":
		// Validate IP address
		if !isValidIP(value) {
			return fmt.Errorf("invalid IP address: %s", value)
		}

		// Update resolver in config
		if len(cfg.UpstreamNameservers) == 0 {
			cfg.UpstreamNameservers = []string{value}
		} else {
			cfg.UpstreamNameservers[0] = value
		}

		// Save config
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Primary resolver set to: %s\n", value)
		return nil

	default:
		return fmt.Errorf("unknown config key: %s. Use 'resolver'", key)
	}
}

func getConfig(key string) error {
	// Load existing config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch key {
	case "resolver":
		if len(cfg.UpstreamNameservers) > 0 {
			fmt.Printf("Primary resolver: %s\n", cfg.UpstreamNameservers[0])
		} else {
			fmt.Println("No resolver configured")
		}
		return nil

	default:
		return fmt.Errorf("unknown config key: %s. Use 'resolver'", key)
	}
}

func isValidIP(ip string) bool {
	// Basic IP validation
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
		num := 0
		for _, char := range part {
			num = num*10 + int(char-'0')
		}
		if num > 255 {
			return false
		}
	}

	return true
}

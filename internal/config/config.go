package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	UpstreamNameservers []string `yaml:"upstream_nameservers"`
}

func Load() (*Config, error) {
	configPath := getConfigPath()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config or create default
	cfg := &Config{}
	if _, err := os.Stat(configPath); err == nil {
		// #nosec G304 -- configPath is a hardcoded path from user home directory
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	} else {
		// Create default config
		cfg = &Config{
			UpstreamNameservers: []string{"8.8.8.8", "1.1.1.1"},
		}

		// Save default config
		if err := Save(cfg); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	// Check for environment variable override
	if envNameservers := os.Getenv("SINKZONE_UPSTREAM_NAMESERVERS"); envNameservers != "" {
		// Split by comma if multiple nameservers are provided
		nameservers := strings.Split(envNameservers, ",")
		// Trim whitespace from each nameserver
		for i, ns := range nameservers {
			nameservers[i] = strings.TrimSpace(ns)
		}
		cfg.UpstreamNameservers = nameservers
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	configPath := getConfigPath()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Use different paths for Windows vs Unix-like systems
	if runtime.GOOS == "windows" {
		// On Windows, use AppData for better compatibility
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "sinkzone", "sinkzone.yaml")
		}
		// Fallback to user home directory
		return filepath.Join(homeDir, "sinkzone", "sinkzone.yaml")
	}

	// Unix-like systems use ~/.sinkzone/
	return filepath.Join(homeDir, ".sinkzone", "sinkzone.yaml")
}

// GetUpstreamAddresses returns the upstream nameservers with port 53 appended
func (c *Config) GetUpstreamAddresses() []string {
	addresses := make([]string, len(c.UpstreamNameservers))
	for i, addr := range c.UpstreamNameservers {
		// If the address doesn't already have a port, append :53
		if !strings.Contains(addr, ":") {
			addresses[i] = addr + ":53"
		} else {
			addresses[i] = addr
		}
	}
	return addresses
}

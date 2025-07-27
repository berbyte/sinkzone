package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Mode                string     `yaml:"mode"`
	FocusEndTime        *time.Time `yaml:"focus_end_time,omitempty"`
	UpstreamNameservers []string   `yaml:"upstream_nameservers"`
}

func Load() (*Config, error) {
	configPath := getConfigPath()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config or create default
	cfg := &Config{}
	if _, err := os.Stat(configPath); err == nil {
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
			Mode:                "normal",
			UpstreamNameservers: []string{"8.8.8.8", "1.1.1.1"},
		}

		// Save default config
		if err := Save(cfg); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	// Check if focus mode has expired
	if cfg.Mode == "focus" && cfg.FocusEndTime != nil && time.Now().After(*cfg.FocusEndTime) {
		cfg.Mode = "normal"
		cfg.FocusEndTime = nil
		if err := Save(cfg); err != nil {
			return nil, fmt.Errorf("failed to update expired focus mode: %w", err)
		}
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	configPath := getConfigPath()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
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

// ValidateUpstreamAddress validates if an address is a valid IP address
func ValidateUpstreamAddress(addr string) bool {
	// Remove port if present for validation
	ipAddr := addr
	if strings.Contains(addr, ":") {
		ipAddr = strings.Split(addr, ":")[0]
	}

	// Basic IP validation (IPv4)
	parts := strings.Split(ipAddr, ".")
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

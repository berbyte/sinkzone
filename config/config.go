package config

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	UpstreamDNS string `toml:"upstream_dns"`
	PIN         string `toml:"pin"`
}

const (
	DefaultUpstreamDNS = "8.8.8.8:53"
	DefaultPIN         = "1234"
)

func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".sinkzone"
	}
	return filepath.Join(home, ".sinkzone", "config.toml")
}

func GetConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".sinkzone"
	}
	return filepath.Join(home, ".sinkzone")
}

func LoadConfig() (*Config, error) {
	configDir := GetConfigDir()
	configPath := GetConfigPath()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config := &Config{
			UpstreamDNS: DefaultUpstreamDNS,
			PIN:         hashPIN(DefaultPIN),
		}
		if err := SaveConfig(config); err != nil {
			return nil, err
		}
		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Set defaults if missing
	if config.UpstreamDNS == "" {
		config.UpstreamDNS = DefaultUpstreamDNS
	}
	if config.PIN == "" {
		config.PIN = hashPIN(DefaultPIN)
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	configPath := GetConfigPath()

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %v", err)
	}

	return nil
}

func hashPIN(pin string) string {
	hash := sha1.Sum([]byte(pin))
	return fmt.Sprintf("%x", hash)
}

func VerifyPIN(pin string, hashedPIN string) bool {
	return hashPIN(pin) == hashedPIN
}

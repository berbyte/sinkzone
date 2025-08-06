package allowlist

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Manager handles allowlist operations
type Manager struct {
	allowlistPath string
}

// NewManager creates a new allowlist manager
func NewManager() (*Manager, error) {
	allowlistPath, err := getAllowlistPath()
	if err != nil {
		return nil, err
	}
	return &Manager{allowlistPath: allowlistPath}, nil
}

// getAllowlistPath returns the platform-specific path for the allowlist file
func getAllowlistPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if runtime.GOOS == "windows" {
		// On Windows, use AppData for better compatibility
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "sinkzone", "allowlist.txt"), nil
		}
		// Fallback to user home directory
		return filepath.Join(homeDir, "sinkzone", "allowlist.txt"), nil
	}

	// Unix-like systems use ~/.sinkzone/
	return filepath.Join(homeDir, ".sinkzone", "allowlist.txt"), nil
}

// Add adds a domain to the allowlist
func (m *Manager) Add(domain string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(m.allowlistPath), 0750); err != nil {
		return fmt.Errorf("failed to create allowlist directory: %w", err)
	}

	// Read existing allowlist
	existingDomains := make(map[string]bool)
	if _, err := os.Stat(m.allowlistPath); err == nil {
		// #nosec G304 -- m.allowlistPath is a hardcoded path from user home directory
		file, err := os.Open(m.allowlistPath)
		if err != nil {
			return fmt.Errorf("failed to open allowlist file: %w", err)
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				fmt.Printf("Warning: failed to close allowlist file: %v\n", closeErr)
			}
		}()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			existingDomain := strings.TrimSpace(scanner.Text())
			if existingDomain != "" && !strings.HasPrefix(existingDomain, "#") {
				existingDomains[existingDomain] = true
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read allowlist file: %w", err)
		}
	}

	// Check if domain is already in allowlist
	if existingDomains[domain] {
		return fmt.Errorf("domain '%s' is already in the allowlist", domain)
	}

	// Add domain to allowlist
	// #nosec G304 -- m.allowlistPath is a hardcoded path from user home directory
	file, err := os.OpenFile(m.allowlistPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open allowlist file for writing: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close allowlist file: %v\n", closeErr)
		}
	}()

	if _, err := file.WriteString(domain + "\n"); err != nil {
		return fmt.Errorf("failed to write to allowlist file: %w", err)
	}

	return nil
}

// Remove removes a domain from the allowlist
func (m *Manager) Remove(domain string) error {
	// Check if allowlist file exists
	if _, err := os.Stat(m.allowlistPath); os.IsNotExist(err) {
		return fmt.Errorf("domain '%s' is not in the allowlist", domain)
	}

	// Read existing allowlist
	// #nosec G304 -- m.allowlistPath is a hardcoded path from user home directory
	file, err := os.Open(m.allowlistPath)
	if err != nil {
		return fmt.Errorf("failed to open allowlist file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close allowlist file: %v\n", closeErr)
		}
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	found := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == domain {
			found = true
			// Skip this line (remove it)
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read allowlist file: %w", err)
	}

	if !found {
		return fmt.Errorf("domain '%s' is not in the allowlist", domain)
	}

	// Write updated allowlist
	if err := os.WriteFile(m.allowlistPath, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		return fmt.Errorf("failed to write allowlist file: %w", err)
	}

	return nil
}

// List returns all domains in the allowlist
func (m *Manager) List() ([]string, error) {
	// Check if allowlist file exists
	if _, err := os.Stat(m.allowlistPath); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Read and return allowlist
	// #nosec G304 -- m.allowlistPath is a hardcoded path from user home directory
	file, err := os.Open(m.allowlistPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open allowlist file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close allowlist file: %v\n", closeErr)
		}
	}()

	var domains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain != "" && !strings.HasPrefix(domain, "#") {
			domains = append(domains, domain)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read allowlist file: %w", err)
	}

	return domains, nil
}

// GetPath returns the allowlist file path
func (m *Manager) GetPath() string {
	return m.allowlistPath
}

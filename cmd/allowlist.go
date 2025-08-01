package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

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

var allowlistCmd = &cobra.Command{
	Use:   "allowlist [add/remove/list] [domain]",
	Short: "Manage the allowlist",
	Long: `Add, remove, or list domains from the allowlist — the list of domains permitted during focus mode.

During focus sessions, all DNS requests are blocked except for domains in your allowlist. You can use 'sinkzone allowlist add <domain>' to permit access, 'remove <domain>' to revoke it, or 'list' to see all allowed domains.

Wildcard patterns are supported:
  * "*github*" matches any domain containing "github"
  * "*.example.com" matches all subdomains of example.com
  * "api.*.com" matches api.anydomain.com

Monitor DNS requests first to discover which domains are needed for your work.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := args[0]

		switch command {
		case "add":
			if len(args) < 2 {
				return fmt.Errorf("domain required for 'add' command")
			}
			return addToAllowlist(args[1])
		case "remove":
			if len(args) < 2 {
				return fmt.Errorf("domain required for 'remove' command")
			}
			return removeFromAllowlist(args[1])
		case "list":
			return listAllowlist()
		default:
			return fmt.Errorf("unknown command: %s. Use 'add', 'remove', or 'list'", command)
		}
	},
}

func addToAllowlist(domain string) error {
	allowlistPath, err := getAllowlistPath()
	if err != nil {
		return fmt.Errorf("failed to get allowlist path: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(allowlistPath), 0750); err != nil {
		return fmt.Errorf("failed to create allowlist directory: %w", err)
	}

	// Read existing allowlist
	existingDomains := make(map[string]bool)
	if _, err := os.Stat(allowlistPath); err == nil {
		// #nosec G304 -- allowlistPath is a hardcoded path from user home directory
		file, err := os.Open(allowlistPath)
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
		fmt.Printf("Domain '%s' is already in the allowlist.\n", domain)
		return nil
	}

	// Add domain to allowlist
	// #nosec G304 -- allowlistPath is a hardcoded path from user home directory
	file, err := os.OpenFile(allowlistPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
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

	fmt.Printf("Domain '%s' added to allowlist.\n", domain)
	fmt.Printf("Note: Allowlist changes take effect when you start a new focus session.\n")
	return nil
}

func removeFromAllowlist(domain string) error {
	allowlistPath, err := getAllowlistPath()
	if err != nil {
		return fmt.Errorf("failed to get allowlist path: %w", err)
	}

	// Check if allowlist file exists
	if _, err := os.Stat(allowlistPath); os.IsNotExist(err) {
		fmt.Printf("Domain '%s' is not in the allowlist.\n", domain)
		return nil
	}

	// Read existing allowlist
	// #nosec G304 -- allowlistPath is a hardcoded path from user home directory
	file, err := os.Open(allowlistPath)
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
		fmt.Printf("Domain '%s' is not in the allowlist.\n", domain)
		return nil
	}

	// Write updated allowlist
	if err := os.WriteFile(allowlistPath, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		return fmt.Errorf("failed to write allowlist file: %w", err)
	}

	fmt.Printf("Domain '%s' removed from allowlist.\n", domain)
	fmt.Printf("Note: Allowlist changes take effect when you start a new focus session.\n")
	return nil
}

func listAllowlist() error {
	allowlistPath, err := getAllowlistPath()
	if err != nil {
		return fmt.Errorf("failed to get allowlist path: %w", err)
	}

	// Check if allowlist file exists
	if _, err := os.Stat(allowlistPath); os.IsNotExist(err) {
		fmt.Println("Allowlist is empty.")
		return nil
	}

	// Read and display allowlist
	// #nosec G304 -- allowlistPath is a hardcoded path from user home directory
	file, err := os.Open(allowlistPath)
	if err != nil {
		return fmt.Errorf("failed to open allowlist file: %w", err)
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
		return fmt.Errorf("failed to read allowlist file: %w", err)
	}

	if len(domains) == 0 {
		fmt.Println("Allowlist is empty.")
		return nil
	}

	fmt.Printf("Allowlist (%d domains):\n", len(domains))
	for i, domain := range domains {
		fmt.Printf("  %d. %s\n", i+1, domain)
	}

	return nil
}

package cmd

import (
	"fmt"

	"github.com/berbyte/sinkzone/internal/allowlist"
	"github.com/spf13/cobra"
)

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
	manager, err := allowlist.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create allowlist manager: %w", err)
	}

	if err := manager.Add(domain); err != nil {
		return err
	}

	fmt.Printf("Domain '%s' added to allowlist.\n", domain)
	fmt.Printf("Note: Allowlist changes take effect when you start a new focus session.\n")
	return nil
}

func removeFromAllowlist(domain string) error {
	manager, err := allowlist.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create allowlist manager: %w", err)
	}

	if err := manager.Remove(domain); err != nil {
		return err
	}

	fmt.Printf("Domain '%s' removed from allowlist.\n", domain)
	fmt.Printf("Note: Allowlist changes take effect when you start a new focus session.\n")
	return nil
}

func listAllowlist() error {
	manager, err := allowlist.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create allowlist manager: %w", err)
	}

	domains, err := manager.List()
	if err != nil {
		return fmt.Errorf("failed to list allowlist: %w", err)
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

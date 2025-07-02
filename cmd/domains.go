package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var allowCmd = &cobra.Command{
	Use:   "allow [domain]",
	Short: "Add domain to allowlist",
	Long:  "Add a domain to the allowlist (will always be allowed regardless of mode)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := strings.ToLower(strings.TrimSpace(args[0]))
		if domain == "" {
			return fmt.Errorf("domain cannot be empty")
		}

		db := getDB()
		if err := db.AddDomainRule(domain, "allow"); err != nil {
			return fmt.Errorf("failed to add domain rule: %v", err)
		}

		fmt.Printf("Added %s to allowlist\n", domain)
		return nil
	},
}

var blockCmd = &cobra.Command{
	Use:   "block [domain]",
	Short: "Add domain to blocklist",
	Long:  "Add a domain to the blocklist (will always be blocked regardless of mode)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := strings.ToLower(strings.TrimSpace(args[0]))
		if domain == "" {
			return fmt.Errorf("domain cannot be empty")
		}

		db := getDB()
		if err := db.AddDomainRule(domain, "block"); err != nil {
			return fmt.Errorf("failed to add domain rule: %v", err)
		}

		fmt.Printf("Added %s to blocklist\n", domain)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current domain rules",
	Long:  "Display all current domain rules (allowlist and blocklist)",
	Run: func(cmd *cobra.Command, args []string) {
		db := getDB()
		rules, err := db.GetDomainRules()
		if err != nil {
			fmt.Printf("Error getting domain rules: %v\n", err)
			return
		}

		if len(rules) == 0 {
			fmt.Println("No domain rules configured")
			return
		}

		fmt.Println("Domain Rules:")
		fmt.Println("=============")

		allowlist := []string{}
		blocklist := []string{}

		for _, rule := range rules {
			if rule.Action == "allow" {
				allowlist = append(allowlist, rule.Domain)
			} else {
				blocklist = append(blocklist, rule.Domain)
			}
		}

		if len(allowlist) > 0 {
			fmt.Println("\nAllowlist:")
			for _, domain := range allowlist {
				fmt.Printf("  ✓ %s\n", domain)
			}
		}

		if len(blocklist) > 0 {
			fmt.Println("\nBlocklist:")
			for _, domain := range blocklist {
				fmt.Printf("  ✗ %s\n", domain)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(allowCmd, blockCmd, listCmd)
}

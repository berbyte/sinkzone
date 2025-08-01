package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var manCmd = &cobra.Command{
	Use:   "man",
	Short: "Show the manual page",
	Long:  `Display the manual page for sinkzone. This shows detailed documentation and usage examples.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Try to find the man page file
		manPagePath := "docs/sinkzone.1"

		// Check if the file exists
		if _, err := os.Stat(manPagePath); os.IsNotExist(err) {
			// Try to find it relative to the executable
			execPath, err := os.Executable()
			if err == nil {
				execDir := filepath.Dir(execPath)
				manPagePath = filepath.Join(execDir, "docs", "sinkzone.1")
			}
		}

		// First try to use the system's man command
		// Note: manPagePath is a hardcoded path, so this is safe from command injection
		if manPath, err := exec.LookPath("man"); err == nil {
			// #nosec G204 -- manPagePath is a hardcoded path, safe from command injection
			cmd := exec.Command(manPath, manPagePath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err == nil {
				return nil
			}
		}

		// Fallback: read and display formatted content
		content, err := os.ReadFile(manPagePath)
		if err != nil {
			return fmt.Errorf("failed to read man page: %w\nMan page not found at: %s", err, manPagePath)
		}

		// Display formatted content
		fmt.Print(formatManPage(string(content)))
		return nil
	},
}

// formatManPage converts troff markup to readable text
func formatManPage(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			result = append(result, "")
			continue
		}

		// Handle different troff commands
		switch {
		case strings.HasPrefix(line, ".TH"):
			// Title header - extract name and description
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				name := strings.Trim(parts[1], "\"")
				desc := strings.Trim(parts[2], "\"")
				result = append(result, "")
				result = append(result, fmt.Sprintf("%s(1) - %s", name, desc))
				result = append(result, "")
			}

		case strings.HasPrefix(line, ".SH"):
			// Section header
			section := strings.TrimSpace(strings.TrimPrefix(line, ".SH"))
			result = append(result, "")
			result = append(result, strings.ToUpper(section))
			result = append(result, strings.Repeat("=", len(section)))
			result = append(result, "")

		case strings.HasPrefix(line, ".TP"):
			// Tagged paragraph - skip, will handle in next line
			continue

		case strings.HasPrefix(line, ".B"):
			// Bold text
			bold := strings.TrimSpace(strings.TrimPrefix(line, ".B"))
			bold = strings.Trim(bold, "\"")
			result = append(result, fmt.Sprintf("  %s", bold))

		case strings.HasPrefix(line, ".IP"):
			// Indented paragraph
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				indent := parts[1]
				text := strings.Join(parts[2:], " ")
				result = append(result, fmt.Sprintf("  %s. %s", indent, text))
			}

		case strings.HasPrefix(line, ".br"):
			// Line break
			result = append(result, "")

		case strings.HasPrefix(line, ".BR"):
			// Bold and roman text
			text := strings.TrimSpace(strings.TrimPrefix(line, ".BR"))
			text = strings.Trim(text, "\"")
			result = append(result, fmt.Sprintf("  %s", text))

		case strings.HasPrefix(line, "\\fB"):
			// Bold text in line
			text := strings.TrimPrefix(line, "\\fB")
			text = strings.TrimSuffix(text, "\\fR")
			result = append(result, fmt.Sprintf("  %s", text))

		case strings.HasPrefix(line, "\\fI"):
			// Italic text in line
			text := strings.TrimPrefix(line, "\\fI")
			text = strings.TrimSuffix(text, "\\fR")
			result = append(result, fmt.Sprintf("  %s", text))

		case strings.HasPrefix(line, "."):
			// Other troff commands - skip
			continue

		default:
			// Regular text
			if line != "" {
				result = append(result, line)
			}
		}
	}

	return strings.Join(result, "\n")
}

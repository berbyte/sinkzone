package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/berbyte/sinkzone/internal/database"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ASCII Banner
const sinkzoneBanner = `
  ██████  ██▓ ███▄    █  ██ ▄█▀▒███████▒ ▒█████   ███▄    █ ▓█████ 
▒██    ▒ ▓██▒ ██ ▀█   █  ██▄█▒ ▒ ▒ ▒ ▒ ▄▀░▒██▒  ██▒ ██ ▀█   █ ▓█   ▀ 
░ ▓██▄   ▒██▒▓██  ▀█ ██▒▓███▄░ ░ ▒ ▄▀▒░ ▒██░  ██▒▓██  ▀█ ██▒▒███   
  ▒   ██▒░██░▓██▒  ▐▌██▒▓██ █▄   ▄▀▒   ░▒██   ██░▓██▒  ▐▌██▒▒▓█  ▄ 
▒██████▒▒░██░▒██░   ▓██░▒██▒ █▄▒███████▒░ ████▓▒░▒██░   ▓██░░▒████▒
▒ ▒▓▒ ▒ ░░▓  ░ ▒░   ▒ ▒ ▒ ▒▒ ▓▒░▒▒ ▓░▒░▒░ ▒░▒░▒░ ░ ▒░   ▒ ▒ ░░ ▒░ ░
░ ░▒  ░ ░ ▒ ░░ ░░   ░ ▒░░ ░▒ ▒░░░▒ ▒ ░ ▒  ░ ▒ ▒░ ░ ░░   ░ ▒░ ░ ░  ░
░  ░  ░   ▒ ░   ░   ░ ░ ░ ░░ ░ ░ ░ ░ ░░ ░ ░ ▒     ░   ░ ░    ░   
      ░   ░           ░ ░  ░     ░ ░        ░ ░           ░    ░  ░
                               ░                                   
`

type Model struct {
	width      int
	height     int
	activeTab  int
	quitting   bool
	tabs       []string
	tabContent []string

	// Animation state
	bannerLines   []string
	currentLine   int
	animationDone bool

	// Database and monitoring
	db          *database.Database
	dnsQueries  []database.DNSQuery
	lastUpdate  time.Time
	lastRefresh time.Time // Track when we last refreshed to only show new records

	// Table scrolling
	tableCursor int
	allowlist   []string

	// Help state
	showHelp bool
}

// Cleanup function to restore terminal
func (m Model) cleanup() {
	// Restore terminal state
	fmt.Print("\033[?25h") // Show cursor
	fmt.Print("\033[2J")   // Clear screen
	fmt.Print("\033[H")    // Move cursor to top
}

// Style definitions
var (
	// Colors
	background = lipgloss.Color("#000000")
	textColor  = lipgloss.Color("#FFFFFF")
	accent1    = lipgloss.Color("#00FFFF") // Cyan
	accent2    = lipgloss.Color("#FF69B4") // Pink
	accent3    = lipgloss.Color("#90EE90") // Light Green
	accent4    = lipgloss.Color("#87CEEB") // Sky Blue
	muted      = lipgloss.Color("#808080") // Grey

	// Header style - blue bar like in screenshot
	headerStyle = lipgloss.NewStyle().
			Foreground(accent2). // Pink color for banner
			Background(background).
			Bold(true).
			Align(lipgloss.Center).
			Margin(1, 0).
			Width(0) // Full width

	// Simple tab style - just text, no borders
	tabStyle = lipgloss.NewStyle().
			Foreground(muted).
			Padding(0, 2).
			Background(background)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true).
			Padding(0, 2).
			Background(background)

	// Content area style
	contentStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(accent4).
			Padding(1, 2).
			Background(background).
			Foreground(textColor)

	// Footer style - pink bar like in screenshot
	footerStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(accent2). // Pink background
			Padding(0, 1).
			Width(0) // Full width

	// Document style
	docStyle = lipgloss.NewStyle().
			Background(background).
			Foreground(textColor).
			Width(0).
			Height(0)
)

// Tick message for animation
type tickMsg time.Time

func Start() error {
	// Restore terminal state before starting
	checkAndRestoreTerminal()

	// Ensure terminal is restored on exit
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic: %v\n", r)
		}
		// Force terminal restoration
		fmt.Print("\033[?25h") // Show cursor
		fmt.Print("\033[2J")   // Clear screen
		fmt.Print("\033[H")    // Move cursor to top
	}()

	// Split banner into lines for animation
	bannerLines := strings.Split(strings.TrimSpace(sinkzoneBanner), "\n")

	// Initialize database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	dbPath := filepath.Join(homeDir, ".sinkzone", "sinkzone.db")
	db, err := database.New(dbPath)
	if err != nil {
		// Continue without database if it fails to initialize
		fmt.Printf("Warning: failed to initialize database: %v\n", err)
		db = nil
	}

	m := Model{
		tabs: []string{"Monitoring", "Settings", "About"},
		tabContent: []string{
			"Main Content Area\n\nThis is where the main content will go.\n\nPress 'q' to quit.",
			"Second Tab Content\n\nThis is the second tab content.",
			"Third Tab Content\n\nThis is the third tab content.",
		},
		bannerLines:   bannerLines,
		currentLine:   0,
		animationDone: false,
		db:            db,
		dnsQueries:    []database.DNSQuery{},
		lastUpdate:    time.Now(),
		lastRefresh:   time.Now(), // Initialize lastRefresh
		tableCursor:   0,
		allowlist:     []string{},
		showHelp:      false,
	}

	// Create program with improved terminal handling
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stderr),
	)

	// Run the program with error handling
	if _, err := p.Run(); err != nil {
		// Ensure terminal is restored even on error
		fmt.Print("\033[?25h\033[2J\033[H")
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

// checkAndRestoreTerminal ensures the terminal is in a proper state
func checkAndRestoreTerminal() {
	// Check if terminal is in a bad state and restore it
	fmt.Print("\033[?25h") // Show cursor
	fmt.Print("\033[2J")   // Clear screen
	fmt.Print("\033[H")    // Move cursor to top
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		if !m.animationDone {
			m.currentLine++
			if m.currentLine >= len(m.bannerLines) {
				m.animationDone = true
				// Start monitoring tick after animation is done
				return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return tickMsg(t)
				})
			}
			return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
				return tickMsg(t)
			})
		} else {
			// Update DNS data every second
			if m.db != nil {
				// Get new records since last refresh
				newQueries, err := m.db.GetNewDNSRecords(m.lastRefresh)
				if err == nil && len(newQueries) > 0 {
					// Append new queries to existing ones
					m.dnsQueries = append(newQueries, m.dnsQueries...)

					// Limit to last 100 records to prevent memory issues
					if len(m.dnsQueries) > 100 {
						m.dnsQueries = m.dnsQueries[:100]
					}

					m.lastUpdate = time.Now()
				}

				// Update last refresh time
				m.lastRefresh = time.Now()

				// Load allowlist
				allowlist, err := m.db.GetAllowlist()
				if err == nil {
					m.allowlist = allowlist
				}
			}
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return tickMsg(t)
			})
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			// Cleanup terminal before quitting
			m.cleanup()
			return m, tea.Quit
		case "1":
			if len(m.tabs) > 0 {
				m.activeTab = 0
				// Reset cursor when switching to monitoring tab
				m.tableCursor = 0
			}
		case "2":
			if len(m.tabs) > 1 {
				m.activeTab = 1
			}
		case "3":
			if len(m.tabs) > 2 {
				m.activeTab = 2
			}
		case "right", "l", "n", "tab":
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
		case "left", "p", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
		case "up", "k":
			if m.activeTab == 0 && len(m.dnsQueries) > 0 {
				// Calculate visible range
				availableHeight := m.height - 12
				if availableHeight < 1 {
					availableHeight = 1
				}
				startIndex := 0
				endIndex := len(m.dnsQueries)
				if len(m.dnsQueries) > availableHeight {
					startIndex = len(m.dnsQueries) - availableHeight
					endIndex = len(m.dnsQueries)
				}

				// Calculate visible position (0-based index within visible items)
				visibleIndex := m.tableCursor - startIndex
				visibleCount := endIndex - startIndex

				// Wrap around within visible range
				if visibleIndex <= 0 {
					visibleIndex = visibleCount - 1
				} else {
					visibleIndex--
				}

				// Map back to dataset index
				m.tableCursor = startIndex + visibleIndex
			}
		case "down", "j":
			if m.activeTab == 0 && len(m.dnsQueries) > 0 {
				// Calculate visible range
				availableHeight := m.height - 12
				if availableHeight < 1 {
					availableHeight = 1
				}
				startIndex := 0
				endIndex := len(m.dnsQueries)
				if len(m.dnsQueries) > availableHeight {
					startIndex = len(m.dnsQueries) - availableHeight
					endIndex = len(m.dnsQueries)
				}

				// Calculate visible position (0-based index within visible items)
				visibleIndex := m.tableCursor - startIndex
				visibleCount := endIndex - startIndex

				// Wrap around within visible range
				if visibleIndex >= visibleCount-1 {
					visibleIndex = 0
				} else {
					visibleIndex++
				}

				// Map back to dataset index
				m.tableCursor = startIndex + visibleIndex
			}
		case " ", "enter":
			if m.activeTab == 0 && len(m.dnsQueries) > 0 && m.tableCursor < len(m.dnsQueries) {
				// Toggle allow/block status for selected domain
				selectedDomain := m.dnsQueries[m.tableCursor].Domain
				if m.isInAllowlist(selectedDomain) {
					// Remove from allowlist
					m.removeFromAllowlist(selectedDomain)
					// Remove from database
					if m.db != nil {
						m.db.RemoveFromAllowlist(selectedDomain)
					}
				} else {
					// Add to allowlist
					if !m.isInAllowlist(selectedDomain) {
						m.allowlist = append(m.allowlist, selectedDomain)
						// Save to database
						if m.db != nil {
							m.db.AddToAllowlist(selectedDomain)
						}
					}
				}
			}
		case "h", "?":
			if m.activeTab == 0 {
				// Toggle help display
				m.showHelp = !m.showHelp
			}
		case "esc":
			if m.activeTab == 0 && m.showHelp {
				// Exit help and return to monitoring
				m.showHelp = false
			}
		}
	}
	return m, nil
}

func (m Model) renderTabs() string {
	var renderedTabs []string

	for i, t := range m.tabs {
		if i == m.activeTab {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, tabStyle.Render(t))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Safety check to ensure activeTab is within bounds
	if m.activeTab >= len(m.tabContent) {
		m.activeTab = 0
	}

	// Render header with banner animation
	bannerText := ""
	if m.animationDone {
		bannerText = sinkzoneBanner
	} else {
		// Show animated banner
		for i := 0; i <= m.currentLine && i < len(m.bannerLines); i++ {
			bannerText += m.bannerLines[i] + "\n"
		}
		// Add empty lines to maintain height during animation
		for i := len(m.bannerLines) - m.currentLine - 1; i > 0; i-- {
			bannerText += "\n"
		}
	}

	// Calculate consistent heights to prevent jiggling
	headerHeight := lipgloss.Height(headerStyle.Render(sinkzoneBanner)) // Use full banner height
	tabHeight := 1
	footerHeight := 1

	// Calculate content height with consistent padding
	contentHeight := m.height - headerHeight - tabHeight - footerHeight - 6 // Extra padding for stability

	// Ensure minimum content height
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Always render header with full height to prevent jiggling
	header := headerStyle.Width(m.width).Height(headerHeight).Align(lipgloss.Center).Render(bannerText)

	// Render tabs
	tabs := m.renderTabs()

	// Content area with safety check
	contentText := "No content available"
	if m.activeTab < len(m.tabContent) {
		if m.activeTab == 0 { // Monitoring tab
			if m.showHelp {
				contentText = m.renderHelp()
			} else {
				contentText = m.renderDNSMonitoring()
			}
		} else {
			contentText = m.tabContent[m.activeTab]
		}
	}

	// Truncate content if it's too long to prevent layout shifts
	contentLines := strings.Split(contentText, "\n")
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
		contentText = strings.Join(contentLines, "\n")
	}

	content := contentStyle.Width(m.width - 2).Height(contentHeight).Render(contentText)

	// Footer - pink bar like in screenshot
	footerText := "q: Quit • h: Help • t: Toggle Focus • 1-3: Switch Tabs • ←/→: Navigate"
	if m.activeTab == 0 {
		if m.showHelp {
			footerText = "esc: Exit Help • h: Hide Help"
		} else {
			footerText = "↑/↓: Navigate • Space/Enter: Toggle Allow/Block • h: Help • q: Quit"
		}
	}
	footer := footerStyle.Width(m.width).Render(footerText)

	// Build the layout vertically with consistent spacing
	doc := strings.Builder{}
	doc.WriteString(header)
	doc.WriteString("\n")
	doc.WriteString(tabs)
	doc.WriteString("\n")
	doc.WriteString(content)
	doc.WriteString("\n")
	doc.WriteString(footer)

	return docStyle.Width(0).Height(0).Render(doc.String())
}

func (m Model) renderDNSMonitoring() string {
	if len(m.dnsQueries) == 0 {
		return "No DNS queries recorded yet.\n\nStart the DNS resolver to see real-time data."
	}

	// Calculate available space for table
	// Reserve space for: header (1) + separator (1) + summary (1) + some padding
	availableHeight := m.height - 12 // More conservative space reservation

	// Limit the number of rows to display
	maxRows := availableHeight
	if maxRows < 1 {
		maxRows = 1
	}

	// Just show the most recent items that fit on screen
	startIndex := 0
	endIndex := len(m.dnsQueries)
	if len(m.dnsQueries) > maxRows {
		// Show the most recent items
		startIndex = len(m.dnsQueries) - maxRows
		endIndex = len(m.dnsQueries)
	}

	// Adjust cursor to be within the visible range
	if m.tableCursor < startIndex {
		m.tableCursor = startIndex
	}
	if m.tableCursor >= endIndex {
		m.tableCursor = endIndex - 1
	}

	// Create table header with proper alignment
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(accent4).
		Render("Domain                         │  Count │ Last Seen  │ Action")

	separator := lipgloss.NewStyle().
		Foreground(muted).
		Render("──────────────────────────────────────────────────────────────")

	var rows []string
	for i := startIndex; i < endIndex; i++ {
		query := m.dnsQueries[i]
		status := "Block"                        // Default to blocked
		statusColor := lipgloss.Color("#FF6B6B") // Red
		if m.isInAllowlist(query.Domain) {
			status = "Allow"
			statusColor = accent3 // Green
		}

		// Add indicator for new records
		domainText := query.Domain

		// Format status with color
		statusText := lipgloss.NewStyle().Foreground(statusColor).Render(status)

		// Use helper function for consistent formatting
		// The cursor should be highlighted if it matches the current row in the dataset
		isSelected := (i == m.tableCursor)
		rowText := formatTableRow(domainText, query.Count, query.Timestamp, statusText, isSelected)
		rows = append(rows, rowText)
	}

	// Ensure we always have the same number of lines to prevent layout shifts
	for len(rows) < maxRows {
		rows = append(rows, strings.Repeat(" ", 60)) // Empty line with consistent width
	}

	// Join all rows
	table := header + "\n" + separator + "\n" + strings.Join(rows, "\n")

	// Add summary
	summary := lipgloss.NewStyle().
		Foreground(muted).
		Render(fmt.Sprintf("\nTotal queries: %d | Last update: %s",
			len(m.dnsQueries),
			m.lastUpdate.Format("15:04:05")))

	return table + summary
}

func (m Model) renderHelp() string {
	// Create help table with nicely formatted information
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(accent4).
		Render("Command                        │ Description")

	separator := lipgloss.NewStyle().
		Foreground(muted).
		Render("──────────────────────────────────────────────────────────────")

	helpData := []struct {
		command     string
		description string
	}{
		{"↑/↓ (Arrow Keys)", "Navigate through DNS query list"},
		{"Space/Enter", "Toggle Allow/Block status for selected domain"},
		{"h", "Show/Hide this help screen"},
		{"esc", "Exit help and return to monitoring"},
		{"1-3", "Switch between tabs (Monitoring, Settings, About)"},
		{"q", "Quit the application"},
		{"←/→ (Arrow Keys)", "Navigate between tabs"},
	}

	var rows []string
	for _, item := range helpData {
		row := lipgloss.NewStyle().
			Foreground(textColor).
			Render(fmt.Sprintf("%-30s │ %s", item.command, item.description))
		rows = append(rows, row)
	}

	// Join all rows
	table := "\n" + header + "\n" + separator + "\n" + strings.Join(rows, "\n")

	// Add additional information
	info := lipgloss.NewStyle().
		Foreground(muted).
		Render(fmt.Sprintf("\n\nSinkzone is the world's most effective productivity tool. It will block all your DNS queries on your machine when you are switching it to focus mode except the ones that are explicitly allowed.\n\n" +
			"In the monitoring menu you can see all the DNS queries you are doing and you can allow the ones you want to use during your focus mode.\n\n" +
			"DNS Monitoring Information:\n" +
			"• The table shows the most recent DNS queries\n" +
			"• Domains are blocked by default unless added to allowlist\n" +
			"• Use Space/Enter to toggle between Allow/Block status\n" +
			"• The list automatically truncates to fit your screen\n" +
			"• Navigation wraps around from bottom to top\n\n"))

	return info + table
}

func formatTableRow(domain string, count int, timestamp time.Time, status string, isSelected bool) string {
	// Format each column with proper width
	domainCol := fmt.Sprintf("%-30s", truncateString(domain, 25))
	countCol := fmt.Sprintf("%6d", count)
	timeCol := fmt.Sprintf("%-10s", timestamp.Format("15:04:05"))
	statusCol := status

	// Combine columns
	rowText := fmt.Sprintf("%s │ %s │ %s │ %s", domainCol, countCol, timeCol, statusCol)

	// Apply styling based on selection
	if isSelected {
		rowStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#2E3440")). // Dark gray background
			Foreground(lipgloss.Color("#ECEFF4")). // Light text
			Bold(true)
		return rowStyle.Render(rowText)
	} else {
		rowStyle := lipgloss.NewStyle().Foreground(textColor)
		return rowStyle.Render(rowText)
	}
}

func (m Model) isInAllowlist(domain string) bool {
	for _, allowed := range m.allowlist {
		if allowed == domain {
			return true
		}
	}
	return false
}

func (m Model) removeFromAllowlist(domain string) {
	newAllowlist := make([]string, 0, len(m.allowlist))
	for _, allowed := range m.allowlist {
		if allowed != domain {
			newAllowlist = append(newAllowlist, allowed)
		}
	}
	m.allowlist = newAllowlist
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return b
	}
	return a
}

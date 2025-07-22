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
	tableCursor  int
	scrollOffset int // Track the start of the visible window
	allowlist    []string
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
		scrollOffset:  0, // Initialize scrollOffset
		allowlist:     []string{},
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
				// Reset scroll offset when switching to monitoring tab
				m.scrollOffset = 0
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
		case "left", "h", "p", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
		case "up", "k":
			if m.activeTab == 0 && len(m.dnsQueries) > 0 {
				m.tableCursor = max(m.tableCursor-1, 0)
				// Adjust scroll offset if cursor moved above visible window
				if m.tableCursor < m.scrollOffset {
					m.scrollOffset = m.tableCursor
				}
			}
		case "down", "j":
			if m.activeTab == 0 && len(m.dnsQueries) > 0 {
				m.tableCursor = min(m.tableCursor+1, len(m.dnsQueries)-1)
				// Calculate available height for table
				availableHeight := m.height - 10 // Reserve space for header, footer, and padding
				if availableHeight < 1 {
					availableHeight = 1
				}
				// Adjust scroll offset if cursor moved below visible window
				if m.tableCursor >= m.scrollOffset+availableHeight {
					m.scrollOffset = m.tableCursor - availableHeight + 1
				}
			}
		case "a":
			if m.activeTab == 0 && len(m.dnsQueries) > 0 && m.tableCursor < len(m.dnsQueries) {
				// Add selected domain to allowlist
				selectedDomain := m.dnsQueries[m.tableCursor].Domain
				if !m.isInAllowlist(selectedDomain) {
					m.allowlist = append(m.allowlist, selectedDomain)
					// Save to database
					if m.db != nil {
						m.db.AddToAllowlist(selectedDomain)
					}
				}
			}
		case "r":
			if m.activeTab == 0 && len(m.dnsQueries) > 0 && m.tableCursor < len(m.dnsQueries) {
				// Remove selected domain from allowlist
				selectedDomain := m.dnsQueries[m.tableCursor].Domain
				m.removeFromAllowlist(selectedDomain)
				// Remove from database
				if m.db != nil {
					m.db.RemoveFromAllowlist(selectedDomain)
				}
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

	// Calculate heights
	headerHeight := lipgloss.Height(headerStyle.Render(bannerText))
	tabHeight := 1
	footerHeight := 1

	// Calculate content height to prevent footer cutoff
	contentHeight := m.height - headerHeight - tabHeight - footerHeight - 4 // Extra padding

	// Ensure minimum content height
	if contentHeight < 5 {
		contentHeight = 5
	}

	header := headerStyle.Width(m.width).Height(headerHeight).Align(lipgloss.Center).Render(bannerText)

	// Render tabs
	tabs := m.renderTabs()

	// Content area with safety check
	contentText := "No content available"
	if m.activeTab < len(m.tabContent) {
		if m.activeTab == 0 { // Monitoring tab
			contentText = m.renderDNSMonitoring()
		} else {
			contentText = m.tabContent[m.activeTab]
		}
	}

	// Truncate content if it's too long
	contentLines := strings.Split(contentText, "\n")
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
		contentText = strings.Join(contentLines, "\n")
	}

	content := contentStyle.Width(m.width - 2).Height(contentHeight).Render(contentText)

	// Footer - pink bar like in screenshot
	footerText := "q: Quit • h: Help • t: Toggle Focus • 1-3: Switch Tabs • ←/→: Navigate"
	if m.activeTab == 0 {
		footerText = "↑/↓: Navigate • a: Add to allowlist • r: Remove from allowlist • q: Quit"
	}
	footer := footerStyle.Width(m.width).Render(footerText)

	// Build the layout vertically
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
	availableHeight := m.height - 10 // Reserve space for header, footer, and padding

	// Limit the number of rows to display
	maxRows := availableHeight
	if maxRows < 1 {
		maxRows = 1
	}

	// Use scroll offset to determine which rows to show
	startIndex := m.scrollOffset
	endIndex := startIndex + maxRows
	if endIndex > len(m.dnsQueries) {
		endIndex = len(m.dnsQueries)
		// Adjust start index if we're at the end
		if endIndex-startIndex < maxRows {
			startIndex = max(0, endIndex-maxRows)
		}
	}

	// Create table header with proper alignment
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(accent4).
		Render("Domain                         │  Count │ Last Seen  │ Status")

	separator := lipgloss.NewStyle().
		Foreground(muted).
		Render("──────────────────────────────────────────────────────────────")

	var rows []string
	for i := startIndex; i < endIndex; i++ {
		query := m.dnsQueries[i]
		status := "Blocked"                      // Default to blocked
		statusColor := lipgloss.Color("#FF6B6B") // Red
		if m.isInAllowlist(query.Domain) {
			status = "Allowed"
			statusColor = accent3 // Green
		}

		// Add indicator for new records
		domainText := query.Domain

		// Format status with color
		statusText := lipgloss.NewStyle().Foreground(statusColor).Render(status)

		// Use helper function for consistent formatting
		// Adjust the cursor index for the window
		isSelected := (i == m.tableCursor)
		rowText := formatTableRow(domainText, query.Count, query.Timestamp, statusText, isSelected)
		rows = append(rows, rowText)
	}

	// Join all rows
	table := header + "\n" + separator + "\n" + strings.Join(rows, "\n")

	// Add summary with scroll indicator
	scrollInfo := ""
	if len(m.dnsQueries) > maxRows {
		scrollInfo = fmt.Sprintf(" | Showing %d-%d of %d", startIndex+1, endIndex, len(m.dnsQueries))
	}

	summary := lipgloss.NewStyle().
		Foreground(muted).
		Render(fmt.Sprintf("\nTotal queries: %d | Last update: %s%s",
			len(m.dnsQueries),
			m.lastUpdate.Format("15:04:05"),
			scrollInfo))

	return table + summary
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

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
		return fmt.Errorf("failed to initialize database: %w", err)
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
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
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
			return m, tea.Quit
		case "1":
			if len(m.tabs) > 0 {
				m.activeTab = 0
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
			}
		case "down", "j":
			if m.activeTab == 0 && len(m.dnsQueries) > 0 {
				m.tableCursor = min(m.tableCursor+1, len(m.dnsQueries)-1)
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

	// Calculate layout dimensions - adjust for proper spacing
	headerHeight := lipgloss.Height(headerStyle.Render(sinkzoneBanner))
	footerHeight := 1                                                       // Simple footer bar
	tabHeight := 1                                                          // Simple text tabs
	contentHeight := m.height - headerHeight - footerHeight - tabHeight - 4 // More space for footer

	// Render animated banner
	var bannerText string
	if m.animationDone {
		// Show full banner when animation is complete
		bannerText = sinkzoneBanner
	} else {
		// Show partial banner during animation
		visibleLines := m.bannerLines[:m.currentLine+1]
		bannerText = strings.Join(visibleLines, "\n")
	}

	// Header - animated banner with fixed height container
	header := headerStyle.Width(m.width).Height(headerHeight).Align(lipgloss.Center).Render(bannerText)

	// Simple tabs - just text, no borders
	tabs := m.renderTabs()

	// Content area with safety check
	contentText := "No content available"
	if m.activeTab < len(m.tabContent) {
		if m.activeTab == 0 {
			// Monitoring tab - show DNS table
			contentText = m.renderDNSMonitoring()
		} else {
			contentText = m.tabContent[m.activeTab]
		}
	}

	content := contentStyle.
		Width(m.width - 4).
		Height(contentHeight).
		Render(contentText)

	// Footer - pink bar like in screenshot
	footerText := "q: Quit • h: Help • t: Toggle Focus • 1-3: Switch Tabs • ←/→: Navigate"
	if m.activeTab == 0 {
		footerText = "↑/↓: Navigate • a: Add to allowlist • r: Remove from allowlist • q: Quit"
	}
	footer := footerStyle.Width(m.width).Render(footerText)

	// Build the layout vertically
	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		footer,
	)

	// Apply full terminal styling
	return docStyle.MaxWidth(m.width).MaxHeight(m.height).Render(layout)
}

func (m Model) renderDNSMonitoring() string {
	if len(m.dnsQueries) == 0 {
		return "No DNS queries recorded yet.\n\nStart the DNS resolver to see real-time data."
	}

	// Create table header with proper alignment
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(accent4).
		Render("Domain                    │ Count │ Last Seen          │ Status")

	separator := lipgloss.NewStyle().
		Foreground(muted).
		Render("────────────────────────────┼───────┼────────────────────┼────────")

	var rows []string
	for i, query := range m.dnsQueries {
		status := "Blocked"                      // Default to blocked
		statusColor := lipgloss.Color("#FF6B6B") // Red
		if m.isInAllowlist(query.Domain) {
			status = "Allowed"
			statusColor = accent3 // Green
		}

		// Check if this is a new record (seen in last 5 seconds)
		isNew := time.Since(query.Timestamp) < 5*time.Second

		// Add indicator for new records
		domainText := query.Domain
		if isNew {
			domainText = "🆕 " + query.Domain
		}

		// Format status with color
		statusText := lipgloss.NewStyle().Foreground(statusColor).Render(status)

		// Use helper function for consistent formatting
		rowText := formatTableRow(domainText, query.Count, query.Timestamp, statusText, i == m.tableCursor)
		rows = append(rows, rowText)
	}

	// Join all rows
	table := header + "\n" + separator + "\n" + strings.Join(rows, "\n")

	// Add summary
	summary := lipgloss.NewStyle().
		Foreground(muted).
		Render(fmt.Sprintf("\nTotal queries: %d | Last update: %s | 🆕 = New in last 5s",
			len(m.dnsQueries),
			m.lastUpdate.Format("15:04:05")))

	return table + summary
}

func formatTableRow(domain string, count int, timestamp time.Time, status string, isSelected bool) string {
	// Format each column with proper width
	domainCol := fmt.Sprintf("%-25s", truncateString(domain, 25))
	countCol := fmt.Sprintf("%5d", count)
	timeCol := fmt.Sprintf("%-17s", timestamp.Format("15:04:05"))
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

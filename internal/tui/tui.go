package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/berbyte/sinkzone/internal/config"
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

// Tab-specific state structures
type MonitoringState struct {
	dnsQueries  []database.DNSQuery
	lastUpdate  time.Time
	lastRefresh time.Time
	tableCursor int
	allowlist   []string
}

type SettingsState struct {
	cursor         int // Which field is selected (0 = resolver1, 1 = resolver2, 2 = save)
	resolver1Input string
	resolver2Input string
	editingField   int // -1 = not editing, 0 = editing resolver1, 1 = editing resolver2
}

type AboutState struct {
	// No specific state needed for now
}

type AllowedDomainsState struct {
	cursor int // Which domain is currently selected
}

type Model struct {
	width     int
	height    int
	activeTab int
	quitting  bool
	tabs      []string

	// Animation state
	bannerLines   []string
	currentLine   int
	animationDone bool

	// Database and config
	db     *database.Database
	config *config.Config

	// Tab-specific states
	monitoring     MonitoringState
	settings       SettingsState
	about          AboutState
	allowedDomains AllowedDomainsState
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

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: failed to load config: %v\n", err)
		cfg = &config.Config{
			Mode:                "normal",
			UpstreamNameservers: []string{"8.8.8.8:53", "1.1.1.1:53"},
		}
	}

	// Initialize resolver inputs from config
	resolver1Input := "8.8.8.8:53"
	resolver2Input := "1.1.1.1:53"
	if len(cfg.UpstreamNameservers) > 0 {
		resolver1Input = cfg.UpstreamNameservers[0]
	}
	if len(cfg.UpstreamNameservers) > 1 {
		resolver2Input = cfg.UpstreamNameservers[1]
	}

	m := Model{
		tabs:          []string{"Monitoring", "Allowed Domains", "Settings", "About"},
		bannerLines:   bannerLines,
		currentLine:   0,
		animationDone: false,
		db:            db,
		config:        cfg,
		monitoring: MonitoringState{
			dnsQueries:  []database.DNSQuery{},
			lastUpdate:  time.Now(),
			lastRefresh: time.Now(),
			tableCursor: 0,
			allowlist:   []string{},
		},
		settings: SettingsState{
			cursor:         0,
			resolver1Input: resolver1Input,
			resolver2Input: resolver2Input,
			editingField:   -1,
		},
		about: AboutState{},
		allowedDomains: AllowedDomainsState{
			cursor: 0,
		},
	}

	// Create program with improved terminal handling
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program with error handling
	if _, err := p.Run(); err != nil {
		// Ensure terminal is restored even on error
		fmt.Print("\033[?25h\033[2J\033[H")
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

func (m Model) saveSettings() {
	// Update config with new resolver values
	m.config.UpstreamNameservers = []string{m.settings.resolver1Input, m.settings.resolver2Input}

	// Save config to file
	if err := config.Save(m.config); err != nil {
		// Could add error handling here, but for now just continue
		fmt.Printf("Warning: failed to save config: %v\n", err)
	}
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
				newQueries, err := m.db.GetNewDNSRecords(m.monitoring.lastRefresh)
				if err == nil && len(newQueries) > 0 {
					// Append new queries to existing ones
					m.monitoring.dnsQueries = append(newQueries, m.monitoring.dnsQueries...)

					// Limit to last 100 records to prevent memory issues
					if len(m.monitoring.dnsQueries) > 100 {
						m.monitoring.dnsQueries = m.monitoring.dnsQueries[:100]
					}

					m.monitoring.lastUpdate = time.Now()
				}

				// Update last refresh time
				m.monitoring.lastRefresh = time.Now()

				// Load allowlist
				allowlist, err := m.db.GetAllowlist()
				if err == nil {
					m.monitoring.allowlist = allowlist
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
		case "left", "h":
			// Navigate to previous tab
			m.activeTab = max(m.activeTab-1, 0)
		case "right", "l":
			// Navigate to next tab
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
		default:
			// Delegate tab-specific updates to their respective functions
			switch m.activeTab {
			case 0: // Monitoring tab
				m, _ = m.updateMonitoring(msg)
			case 1: // Allowed Domains tab
				m, _ = m.updateAllowedDomains(msg)
			case 2: // Settings tab
				m, _ = m.updateSettings(msg)
			case 3: // About tab
				m, _ = m.updateAbout(msg)
			}
		}
	}
	return m, nil
}

// Tab-specific update functions
func (m Model) updateMonitoring(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if len(m.monitoring.dnsQueries) > 0 {
			// Calculate visible range
			availableHeight := m.height - 12
			if availableHeight < 1 {
				availableHeight = 1
			}
			startIndex := 0
			endIndex := len(m.monitoring.dnsQueries)
			if len(m.monitoring.dnsQueries) > availableHeight {
				startIndex = len(m.monitoring.dnsQueries) - availableHeight
				endIndex = len(m.monitoring.dnsQueries)
			}

			// Calculate visible position (0-based index within visible items)
			visibleIndex := m.monitoring.tableCursor - startIndex
			visibleCount := endIndex - startIndex

			// Wrap around within visible range
			if visibleIndex <= 0 {
				visibleIndex = visibleCount - 1
			} else {
				visibleIndex--
			}

			// Map back to dataset index
			m.monitoring.tableCursor = startIndex + visibleIndex
		}
	case "down", "j":
		if len(m.monitoring.dnsQueries) > 0 {
			// Calculate visible range
			availableHeight := m.height - 12
			if availableHeight < 1 {
				availableHeight = 1
			}
			startIndex := 0
			endIndex := len(m.monitoring.dnsQueries)
			if len(m.monitoring.dnsQueries) > availableHeight {
				startIndex = len(m.monitoring.dnsQueries) - availableHeight
				endIndex = len(m.monitoring.dnsQueries)
			}

			// Calculate visible position (0-based index within visible items)
			visibleIndex := m.monitoring.tableCursor - startIndex
			visibleCount := endIndex - startIndex

			// Wrap around within visible range
			if visibleIndex >= visibleCount-1 {
				visibleIndex = 0
			} else {
				visibleIndex++
			}

			// Map back to dataset index
			m.monitoring.tableCursor = startIndex + visibleIndex
		}
	case "enter", " ":
		if len(m.monitoring.dnsQueries) > 0 && m.monitoring.tableCursor < len(m.monitoring.dnsQueries) {
			// Toggle allow/block status for selected domain
			selectedDomain := m.monitoring.dnsQueries[m.monitoring.tableCursor].Domain
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
					m.monitoring.allowlist = append(m.monitoring.allowlist, selectedDomain)
					// Save to database
					if m.db != nil {
						m.db.AddToAllowlist(selectedDomain)
					}
				}
			}
		}
	}
	return m, nil
}

func (m Model) updateSettings(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.settings.editingField == -1 {
			// Not editing, navigate between fields
			m.settings.cursor = max(m.settings.cursor-1, 0)
		}
	case "down", "j":
		if m.settings.editingField == -1 {
			// Not editing, navigate between fields
			m.settings.cursor = min(m.settings.cursor+1, 2) // 0=resolver1, 1=resolver2, 2=save
		}
	case "enter":
		if m.settings.editingField == -1 {
			// Start editing the selected field
			if m.settings.cursor == 0 {
				m.settings.editingField = 0
			} else if m.settings.cursor == 1 {
				m.settings.editingField = 1
			} else if m.settings.cursor == 2 {
				// Save button pressed
				m.saveSettings()
			}
		} else {
			// Finish editing
			m.settings.editingField = -1
		}
	case "escape":
		if m.settings.editingField != -1 {
			// Cancel editing
			m.settings.editingField = -1
		}
	case "backspace":
		if m.settings.editingField != -1 {
			if m.settings.editingField == 0 && len(m.settings.resolver1Input) > 0 {
				m.settings.resolver1Input = m.settings.resolver1Input[:len(m.settings.resolver1Input)-1]
			} else if m.settings.editingField == 1 && len(m.settings.resolver2Input) > 0 {
				m.settings.resolver2Input = m.settings.resolver2Input[:len(m.settings.resolver2Input)-1]
			}
		}
	default:
		// Handle text input for settings form
		if m.settings.editingField != -1 {
			if len(msg.Runes) > 0 {
				r := msg.Runes[0]
				if r >= 32 && r <= 126 { // Printable ASCII characters
					if m.settings.editingField == 0 {
						m.settings.resolver1Input += string(r)
					} else if m.settings.editingField == 1 {
						m.settings.resolver2Input += string(r)
					}
				}
			}
		}
	}
	return m, nil
}

func (m Model) updateAbout(msg tea.KeyMsg) (Model, tea.Cmd) {
	// About tab has no interactive elements
	return m, nil
}

func (m Model) updateAllowedDomains(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if len(m.monitoring.allowlist) > 0 {
			// Calculate visible range
			availableHeight := m.height - 12
			if availableHeight < 1 {
				availableHeight = 1
			}
			startIndex := 0
			endIndex := len(m.monitoring.allowlist)
			if len(m.monitoring.allowlist) > availableHeight {
				startIndex = len(m.monitoring.allowlist) - availableHeight
				endIndex = len(m.monitoring.allowlist)
			}

			// Calculate visible position (0-based index within visible items)
			visibleIndex := m.allowedDomains.cursor - startIndex
			visibleCount := endIndex - startIndex

			// Wrap around within visible range
			if visibleIndex <= 0 {
				visibleIndex = visibleCount - 1
			} else {
				visibleIndex--
			}

			// Map back to dataset index
			m.allowedDomains.cursor = startIndex + visibleIndex
		}
	case "down", "j":
		if len(m.monitoring.allowlist) > 0 {
			// Calculate visible range
			availableHeight := m.height - 12
			if availableHeight < 1 {
				availableHeight = 1
			}
			startIndex := 0
			endIndex := len(m.monitoring.allowlist)
			if len(m.monitoring.allowlist) > availableHeight {
				startIndex = len(m.monitoring.allowlist) - availableHeight
				endIndex = len(m.monitoring.allowlist)
			}

			// Calculate visible position (0-based index within visible items)
			visibleIndex := m.allowedDomains.cursor - startIndex
			visibleCount := endIndex - startIndex

			// Wrap around within visible range
			if visibleIndex >= visibleCount-1 {
				visibleIndex = 0
			} else {
				visibleIndex++
			}

			// Map back to dataset index
			m.allowedDomains.cursor = startIndex + visibleIndex
		}
	case "enter", " ":
		if len(m.monitoring.allowlist) > 0 && m.allowedDomains.cursor < len(m.monitoring.allowlist) {
			// Remove selected domain from allowlist
			selectedDomain := m.monitoring.allowlist[m.allowedDomains.cursor]
			m.removeFromAllowlist(selectedDomain)

			// Remove from database
			if m.db != nil {
				m.db.RemoveFromAllowlist(selectedDomain)
			}

			// Adjust cursor if we removed the last item
			if m.allowedDomains.cursor >= len(m.monitoring.allowlist) {
				m.allowedDomains.cursor = max(len(m.monitoring.allowlist)-1, 0)
			}
		}
	}
	return m, nil
}

func (m Model) renderTabs() string {
	// Safety check to ensure activeTab is within bounds
	if m.activeTab >= len(m.tabs) {
		m.activeTab = 0
	}

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
	if m.activeTab >= len(m.tabs) {
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
	if m.activeTab < len(m.tabs) {
		if m.activeTab == 0 { // Monitoring tab
			contentText = m.renderDNSMonitoring()
		} else if m.activeTab == 1 { // Allowed Domains tab
			contentText = m.renderAllowedDomains()
		} else if m.activeTab == 2 { // Settings tab
			contentText = m.renderSettings()
		} else if m.activeTab == 3 { // About tab
			contentText = m.renderHelp()
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
	footerText := "q: Quit • h: Help • t: Toggle Focus • ←/→: Switch Tabs"
	if m.activeTab == 0 {
		footerText = "↑/↓: Navigate • Space/Enter: Toggle Allow/Block • q: Quit • ←/→: Switch Tabs"
	} else if m.activeTab == 1 {
		footerText = "↑/↓: Navigate • Space/Enter: Remove Domain • q: Quit • ←/→: Switch Tabs"
	} else if m.activeTab == 2 {
		footerText = "↑/↓: Navigate • Enter: Edit/Save • Escape: Cancel • q: Quit • ←/→: Switch Tabs"
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
	if len(m.monitoring.dnsQueries) == 0 {
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
	endIndex := len(m.monitoring.dnsQueries)
	if len(m.monitoring.dnsQueries) > maxRows {
		// Show the most recent items
		startIndex = len(m.monitoring.dnsQueries) - maxRows
		endIndex = len(m.monitoring.dnsQueries)
	}

	// Adjust cursor to be within the visible range
	if m.monitoring.tableCursor < startIndex {
		m.monitoring.tableCursor = startIndex
	}
	if m.monitoring.tableCursor >= endIndex {
		m.monitoring.tableCursor = endIndex - 1
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
		query := m.monitoring.dnsQueries[i]
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
		isSelected := (i == m.monitoring.tableCursor)
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
			len(m.monitoring.dnsQueries),
			m.monitoring.lastUpdate.Format("15:04:05")))

	return table + summary
}

func (m Model) renderHelp() string {
	helpText := `
Sinkzone - DNS-based Productivity Tool

This tool helps you stay focused by blocking distracting websites during focus sessions.

Features:
• Real-time DNS monitoring
• Configurable upstream resolvers
• Allowlist management
• Focus mode with automatic expiration

Usage:
• Press ←/→ to switch between tabs
• Use ↑/↓ to navigate the monitoring table
• Press Space/Enter to toggle allow/block status
• Press q to quit

For more information, visit: https://github.com/berbyte/sinkzone
`
	return helpText
}

func (m Model) renderSettings() string {
	// Form styles
	formStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E5E7EB")).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6B7280"))

	selectedInputStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Background(lipgloss.Color("#1F2937"))

	buttonStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6B7280")).
		Foreground(lipgloss.Color("#E5E7EB"))

	selectedButtonStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Background(lipgloss.Color("#1F2937")).
		Foreground(lipgloss.Color("#FFFFFF"))

	// Render form
	var form strings.Builder
	form.WriteString("Upstream DNS Resolvers\n\n")

	// Resolver 1
	form.WriteString(labelStyle.Render("Primary Resolver:"))
	form.WriteString("\n")
	if m.settings.cursor == 0 {
		form.WriteString(selectedInputStyle.Render(m.settings.resolver1Input))
	} else {
		form.WriteString(inputStyle.Render(m.settings.resolver1Input))
	}
	form.WriteString("\n\n")

	// Resolver 2
	form.WriteString(labelStyle.Render("Secondary Resolver:"))
	form.WriteString("\n")
	if m.settings.cursor == 1 {
		form.WriteString(selectedInputStyle.Render(m.settings.resolver2Input))
	} else {
		form.WriteString(inputStyle.Render(m.settings.resolver2Input))
	}
	form.WriteString("\n\n")

	// Save button
	if m.settings.cursor == 2 {
		form.WriteString(selectedButtonStyle.Render("Save"))
	} else {
		form.WriteString(buttonStyle.Render("Save"))
	}

	return formStyle.Render(form.String())
}

func (m Model) renderAllowedDomains() string {
	if len(m.monitoring.allowlist) == 0 {
		return "No domains in allowlist.\n\nAdd domains from the Monitoring tab by pressing Space/Enter on blocked domains."
	}

	// Calculate available height for the list
	availableHeight := m.height - 12
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Determine which domains to show
	startIndex := 0
	endIndex := len(m.monitoring.allowlist)
	if len(m.monitoring.allowlist) > availableHeight {
		startIndex = len(m.monitoring.allowlist) - availableHeight
		endIndex = len(m.monitoring.allowlist)
	}

	// Adjust cursor to be within the visible range
	if m.allowedDomains.cursor < startIndex {
		m.allowedDomains.cursor = startIndex
	}
	if m.allowedDomains.cursor >= endIndex {
		m.allowedDomains.cursor = endIndex - 1
	}

	// Use the same styles as monitoring tab
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(accent4).
		Padding(0, 1)

	// Create table header
	header := headerStyle.Render("Domain                          │ Status")
	separator := lipgloss.NewStyle().
		Foreground(muted).
		Render("──────────────────────────────────────────────────────────────")

	// Render domain list with consistent table format
	var rows []string
	for i := startIndex; i < endIndex; i++ {
		domain := m.monitoring.allowlist[i]
		isSelected := (i == m.allowedDomains.cursor)

		// Use the same row formatting as monitoring tab
		rowText := formatAllowedDomainRow(domain, isSelected)
		rows = append(rows, rowText)
	}

	// Join all rows
	table := "\n" + header + "\n" + separator + "\n" + strings.Join(rows, "\n")

	// Add summary with consistent styling
	summary := lipgloss.NewStyle().
		Foreground(muted).
		Render(fmt.Sprintf("\nTotal allowed domains: %d", len(m.monitoring.allowlist)))

	return table + summary
}

func formatAllowedDomainRow(domain string, isSelected bool) string {
	// Use the same styling as the monitoring tab
	statusText := "Allowed"
	statusColor := lipgloss.Color("#4ADE80") // Green for allowed

	// Format the row with consistent column widths
	domainText := truncateString(domain, 30)
	statusTextFormatted := lipgloss.NewStyle().
		Foreground(statusColor).
		Render(statusText)

	// Create the row with consistent formatting
	row := fmt.Sprintf("%-30s │ %s", domainText, statusTextFormatted)

	if isSelected {
		// Use the same selected row styling as monitoring tab
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#2E3440")).
			Foreground(lipgloss.Color("#ECEFF4")).
			Render(row)
	} else {
		return lipgloss.NewStyle().
			Foreground(textColor).
			Render(row)
	}
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
	for _, allowed := range m.monitoring.allowlist {
		if allowed == domain {
			return true
		}
	}
	return false
}

func (m Model) removeFromAllowlist(domain string) {
	newAllowlist := make([]string, 0, len(m.monitoring.allowlist))
	for _, allowed := range m.monitoring.allowlist {
		if allowed != domain {
			newAllowlist = append(newAllowlist, allowed)
		}
	}
	m.monitoring.allowlist = newAllowlist
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

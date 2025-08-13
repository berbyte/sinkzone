package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/berbyte/sinkzone/internal/allowlist"
	"github.com/berbyte/sinkzone/internal/api"
	"github.com/berbyte/sinkzone/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ASCII Banner
const sinkzoneBanner = `
  ██████  ██▓ ███▄    █  ██ ▄█▀▒███████▒ ▒█████   ███▄    █ ▓█████ 
▒██    ▒ ▓██▒ ██ ▀█   █  ██▄█▒ ▒ ▒ ▒ ▒ ▄▀░▒██▒  ██▒ ██ ▀█   █ ▓█   ▀ 
░ ▓██▄   ▒██▒▓██  ▀█ ██▒▓███▄░ ░ ▒ ▄▀▒░ ▒██░  ██▒▓██▒  ▐▌██▒▒███   
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
	dnsQueries  []api.DNSQuery
	lastUpdate  time.Time
	lastRefresh time.Time
	tableCursor int
}

type AllowedDomainsState struct {
	cursor  int // Which domain is currently selected
	domains []string
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

	// API client and config
	apiClient *api.Client
	config    *config.Config

	// Focus mode state
	focusModeActive  bool
	focusEndTime     *time.Time
	focusMessage     string // Temporary message when focus mode is activated
	focusMessageTime time.Time

	// Tab-specific states
	monitoring     MonitoringState
	allowedDomains AllowedDomainsState

	// Update tracking
	lastChangedDomain   string    // Track the last domain that was changed
	lastChangeTime      time.Time // When the last change occurred
	lastAllowlistReload time.Time // When the allowlist was last reloaded
	lastUserActivity    time.Time // When the user last pressed a key

	// Easter egg state
	rainbowMode   bool   // Whether rainbow mode is active
	rainbowOffset int    // Current rainbow color offset
	keyBuffer     string // Buffer for detecting key sequences
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
	accent2    = lipgloss.Color("#FF69B4") // Pink
	accent4    = lipgloss.Color("#87CEEB") // Sky Blue
	muted      = lipgloss.Color("#808080") // Grey

	// Rainbow colors for easter egg
	rainbowColors = []lipgloss.Color{
		lipgloss.Color("#FF0000"), // Red
		lipgloss.Color("#FF7F00"), // Orange
		lipgloss.Color("#FFFF00"), // Yellow
		lipgloss.Color("#00FF00"), // Green
		lipgloss.Color("#0000FF"), // Blue
		lipgloss.Color("#4B0082"), // Indigo
		lipgloss.Color("#9400D3"), // Violet
	}

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
	return StartWithAPIURL("http://127.0.0.1:8080")
}

func StartWithAPIURL(apiURL string) error {
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

	// Initialize API client
	apiClient := api.NewClient(apiURL)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: failed to load config: %v\n", err)
		cfg = &config.Config{
			UpstreamNameservers: []string{"8.8.8.8", "1.1.1.1"},
		}
	}

	m := Model{
		tabs:          []string{"Monitoring", "Allowlist"},
		bannerLines:   bannerLines,
		currentLine:   0,
		animationDone: false,
		apiClient:     apiClient,
		config:        cfg,
		monitoring: MonitoringState{
			dnsQueries:  []api.DNSQuery{},
			lastUpdate:  time.Now(),
			lastRefresh: time.Now(),
			tableCursor: 0,
		},
		allowedDomains: AllowedDomainsState{
			cursor:  0,
			domains: []string{},
		},
		lastAllowlistReload: time.Now(),
		lastUserActivity:    time.Now(),
		rainbowMode:         false,
		rainbowOffset:       0,
		keyBuffer:           "",
	}

	// Initialize focus mode status
	m.updateFocusModeStatus()

	// Load initial data
	m.loadInitialData()

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

func (m Model) loadInitialData() {
	// Load initial DNS queries
	if queries, err := m.apiClient.GetQueries(); err == nil {
		m.monitoring.dnsQueries = queries
		m.monitoring.lastUpdate = time.Now()
	}

	// Load initial allowlist
	m.loadAllowlistData()

	// Initialize cursor bounds
	if len(m.monitoring.dnsQueries) > 0 {
		m.monitoring.tableCursor = 0 // Start at the top (newest entries)
	}
}

// validatePath ensures the path is within the user's home directory and doesn't contain path traversal
func validatePath(path string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Resolve any symlinks and get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check if the path is within the home directory
	if !strings.HasPrefix(absPath, homeDir) {
		return fmt.Errorf("path is outside home directory: %s", path)
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains traversal attempt: %s", path)
	}

	return nil
}

func (m *Model) loadAllowlistData() {
	manager, err := allowlist.NewManager()
	if err != nil {
		// If we can't create the manager, set empty domains
		m.allowedDomains.domains = []string{}
		m.allowedDomains.cursor = 0
		return
	}

	domains, err := manager.List()
	if err != nil {
		// If we can't list domains, set empty domains
		m.allowedDomains.domains = []string{}
		m.allowedDomains.cursor = 0
		return
	}

	m.allowedDomains.domains = domains

	// Adjust cursor if needed
	if len(domains) > 0 {
		if m.allowedDomains.cursor >= len(domains) {
			m.allowedDomains.cursor = len(domains) - 1
		}
	} else {
		m.allowedDomains.cursor = 0
	}
}

func (m Model) enableFocusMode() error {
	// Enable focus mode for 1 hour via API
	if err := m.apiClient.SetFocusMode(true, "1h"); err != nil {
		return fmt.Errorf("failed to enable focus mode: %w", err)
	}

	// Update focus mode status immediately
	m.updateFocusModeStatus()
	return nil
}

func (m Model) updateFocusModeStatus() {
	// Get focus mode state from API
	if focusState, err := m.apiClient.GetFocusMode(); err == nil {
		// Update focus mode state from API response
		//nolint:staticcheck // SA4005: These assignments are necessary for state synchronization
		m.focusModeActive = focusState.Enabled
		//nolint:staticcheck // SA4005: These assignments are necessary for state synchronization
		m.focusEndTime = focusState.EndTime
		return
	}

	// Fallback to state manager if API is not available
	stateMgr, err := config.NewStateManager()
	if err != nil {
		return
	}

	state := stateMgr.GetState()
	// Update focus mode state from state manager
	//nolint:staticcheck // SA4005: These assignments are necessary for state synchronization
	m.focusModeActive = state.FocusMode
	//nolint:staticcheck // SA4005: These assignments are necessary for state synchronization
	m.focusEndTime = state.FocusEndTime
}

func checkAndRestoreTerminal() {
	// Check if terminal is in raw mode and restore if needed
	fmt.Print("\033[?25h") // Show cursor
}

func (m Model) Init() tea.Cmd {
	// Start animation tick
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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

			// Check focus mode status during animation too
			m.updateFocusModeStatus()

			return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
				return tickMsg(t)
			})
		} else {
			// Update DNS data every 3 seconds, but pause if user is actively navigating
			if time.Since(m.lastUserActivity) > 2*time.Second {
				if queries, err := m.apiClient.GetQueries(); err == nil {
					if len(queries) > 0 {
						// Calculate how many entries we can display
						headerHeight := lipgloss.Height(headerStyle.Render(sinkzoneBanner)) + 2
						tabHeight := 1
						footerHeight := 1
						contentHeight := m.height - headerHeight - tabHeight - footerHeight - 2
						maxVisibleEntries := contentHeight - 4 // Account for header, footer, and padding
						if maxVisibleEntries < 5 {
							maxVisibleEntries = 5 // Minimum entries
						}

						// Truncate to only keep the most recent entries that fit
						if len(queries) > maxVisibleEntries {
							queries = queries[len(queries)-maxVisibleEntries:]
						}

						// Store the currently selected domain before updating data
						var selectedDomain string
						if len(m.monitoring.dnsQueries) > 0 && m.monitoring.tableCursor < len(m.monitoring.dnsQueries) {
							selectedDomain = m.monitoring.dnsQueries[m.monitoring.tableCursor].Domain
						}

						// Update the data
						m.monitoring.dnsQueries = queries
						m.monitoring.lastUpdate = time.Now()

						// Try to restore cursor position to the same domain
						if selectedDomain != "" {
							for i, query := range queries {
								if query.Domain == selectedDomain {
									m.monitoring.tableCursor = i
									break
								}
							}
						} else if len(queries) > 0 {
							// If no domain was selected, default to the newest entry (first in array since we display newest first)
							m.monitoring.tableCursor = 0
						}
					}
				}
			}

			// Update last refresh time
			m.monitoring.lastRefresh = time.Now()

			// Check focus mode status
			m.updateFocusModeStatus()

			// Clear focus message after 3 seconds
			if m.focusMessage != "" && time.Since(m.focusMessageTime) > 3*time.Second {
				m.focusMessage = ""
			}

			// Clear last changed domain after 2 seconds
			if m.lastChangedDomain != "" && time.Since(m.lastChangeTime) > 2*time.Second {
				m.lastChangedDomain = ""
			}

			// Reload allowlist data periodically (every 5 seconds)
			if time.Since(m.lastAllowlistReload) >= 5*time.Second {
				m.loadAllowlistData()
				m.lastAllowlistReload = time.Now()
			}

			// Update rainbow animation if active
			if m.rainbowMode {
				m.rainbowOffset = (m.rainbowOffset + 1) % len(rainbowColors)
			}

			return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
				return tickMsg(t)
			})
		}
	case tea.KeyMsg:
		// Handle easter egg key sequence detection
		if !m.rainbowMode {
			// Only add to buffer if it's a single character (not special keys like arrows, etc.)
			if len(msg.String()) == 1 {
				m.keyBuffer += msg.String()
				// Keep only last 5 characters to detect "iddqd"
				if len(m.keyBuffer) > 5 {
					m.keyBuffer = m.keyBuffer[len(m.keyBuffer)-5:]
				}
				// Check for "iddqd" easter egg
				if strings.Contains(m.keyBuffer, "iddqd") {
					m.rainbowMode = true
					m.rainbowOffset = 0
					m.keyBuffer = "" // Clear buffer after activation
				}
			} else {
				// Reset buffer on special keys to prevent false triggers
				m.keyBuffer = ""
			}
		}

		switch msg.String() {
		case "esc", "ctrl+c":
			m.quitting = true
			// Cleanup terminal before quitting
			m.cleanup()
			return m, tea.Quit
		case "f":
			// Enable focus mode for 1 hour
			if err := m.enableFocusMode(); err != nil {
				// Could add error handling here, but for now just continue
				fmt.Printf("Warning: failed to enable focus mode: %v\n", err)
			} else {
				// If we're on monitoring tab, switch to allowlist tab
				if m.activeTab == 0 {
					m.activeTab = 1
				}
				// Show temporary success message
				m.focusMessage = "🔒 Focus mode activated for 1 hour!"
				m.focusMessageTime = time.Now()
			}
		case "left", "h":
			// Navigate to previous tab
			if m.activeTab > 0 {
				m.activeTab--
			} else {
				m.activeTab = len(m.tabs) - 1
			}
			// Reload allowlist data when switching to allowlist tab
			if m.activeTab == 1 {
				m.loadAllowlistData()
			}
		case "right", "l":
			// Navigate to next tab
			if m.activeTab < len(m.tabs)-1 {
				m.activeTab++
			} else {
				m.activeTab = 0
			}
			// Reload allowlist data when switching to allowlist tab
			if m.activeTab == 1 {
				m.loadAllowlistData()
			}
		case "1":
			m.activeTab = 0
			// Reload allowlist data when switching to allowlist tab
			if m.activeTab == 1 {
				m.loadAllowlistData()
			}
		case "2":
			m.activeTab = 1
			// Reload allowlist data when switching to allowlist tab
			m.loadAllowlistData()
		default:
			// Handle tab-specific key events
			switch m.activeTab {
			case 0:
				return m.updateMonitoring(msg)
			case 1:
				return m.updateAllowedDomains(msg)
			}
		}
	}
	return m, nil
}

func (m *Model) updateMonitoring(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Track user activity
	m.lastUserActivity = time.Now()

	// Since we're now keeping only the visible entries, we can simplify this
	visibleCount := len(m.monitoring.dnsQueries)

	switch msg.String() {
	case "up", "k":
		if m.monitoring.tableCursor > 0 {
			m.monitoring.tableCursor--
		}
	case "down", "j":
		if m.monitoring.tableCursor < visibleCount-1 {
			m.monitoring.tableCursor++
		}
	case " ", "enter":
		if len(m.monitoring.dnsQueries) > 0 && m.monitoring.tableCursor < len(m.monitoring.dnsQueries) {
			// Map cursor position to the original data order (since we reversed for display)
			originalIndex := len(m.monitoring.dnsQueries) - 1 - m.monitoring.tableCursor
			selectedQuery := m.monitoring.dnsQueries[originalIndex]
			selectedDomain := selectedQuery.Domain

			// Check if domain is already in allowlist
			isInAllowlist := m.isInAllowlist(selectedDomain)

			if isInAllowlist {
				// Remove from allowlist if already present
				if err := m.removeFromAllowlist(selectedDomain); err == nil {
					m.loadAllowlistData()
					m.lastChangedDomain = selectedDomain
					m.lastChangeTime = time.Now()
				}
			} else {
				// Add to allowlist if not present
				if err := m.addToAllowlist(selectedDomain); err == nil {
					m.loadAllowlistData()
					m.lastChangedDomain = selectedDomain
					m.lastChangeTime = time.Now()
				}
			}
		}
	}
	return *m, nil
}

func (m *Model) updateAllowedDomains(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Track user activity
	m.lastUserActivity = time.Now()

	switch msg.String() {
	case "up", "k":
		if m.allowedDomains.cursor > 0 {
			m.allowedDomains.cursor--
		}
	case "down", "j":
		if m.allowedDomains.cursor < len(m.allowedDomains.domains)-1 {
			m.allowedDomains.cursor++
		}
	case " ", "enter":
		if len(m.allowedDomains.domains) > 0 && m.allowedDomains.cursor < len(m.allowedDomains.domains) {
			selectedDomain := m.allowedDomains.domains[m.allowedDomains.cursor]

			// Remove from allowlist
			if err := m.removeFromAllowlist(selectedDomain); err == nil {
				m.loadAllowlistData()
				m.lastChangedDomain = selectedDomain
				m.lastChangeTime = time.Now()
			}
		}
	}
	return *m, nil
}

func (m Model) renderTabs() string {
	var renderedTabs []string
	for i, tab := range m.tabs {
		if i == m.activeTab {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, tabStyle.Render(tab))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, renderedTabs...)
}

func (m Model) renderBanner() string {
	if m.rainbowMode {
		// Render banner with rainbow colors
		var rainbowBanner strings.Builder
		bannerLines := strings.Split(strings.TrimSpace(sinkzoneBanner), "\n")

		for i, line := range bannerLines {
			if line == "" {
				rainbowBanner.WriteString("\n")
				continue
			}

			// Calculate color index for this line
			colorIndex := (m.rainbowOffset + i) % len(rainbowColors)
			color := rainbowColors[colorIndex]

			// Create rainbow style for this line
			rainbowStyle := lipgloss.NewStyle().
				Foreground(color).
				Background(background).
				Bold(true)

			rainbowBanner.WriteString(rainbowStyle.Render(line) + "\n")
		}

		return rainbowBanner.String()
	} else {
		// Render normal banner
		return sinkzoneBanner
	}
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
		bannerText = "\n" + m.renderBanner() // Add newline to start from 2nd line
	} else {
		// Show animated banner starting from 2nd line
		bannerText = "\n" // Start from 2nd line
		for i := 0; i <= m.currentLine && i < len(m.bannerLines); i++ {
			bannerText += m.bannerLines[i] + "\n"
		}
		// Add empty lines to maintain height during animation
		for i := len(m.bannerLines) - m.currentLine - 1; i > 0; i-- {
			bannerText += "\n"
		}
	}

	// Calculate consistent heights to prevent jiggling
	headerHeight := lipgloss.Height(headerStyle.Render(m.renderBanner())) + 2 // Add padding for banner
	tabHeight := 1
	footerHeight := 1

	// Calculate content height to fill remaining space
	contentHeight := m.height - headerHeight - tabHeight - footerHeight - 2 // Minimal padding

	// Ensure minimum content height
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Add focus mode indicator to header if active
	var header string
	if m.focusModeActive {
		focusIndicator := lipgloss.NewStyle().
			Background(lipgloss.Color("#FF6B6B")). // Red background
			Foreground(lipgloss.Color("#FFFFFF")). // White text
			Bold(true).
			Padding(0, 1).
			Render("🔒 FOCUS MODE ACTIVE")

		// Combine banner with focus indicator
		headerContent := bannerText + "\n" + focusIndicator

		// Use red-tinted header style for focus mode
		focusHeaderStyle := headerStyle.
			Background(lipgloss.Color("#2D1B1B")). // Dark red background
			Foreground(lipgloss.Color("#FF6B6B"))  // Red text
		header = focusHeaderStyle.Width(m.width).Height(headerHeight).Align(lipgloss.Center).Padding(1, 0).Render(headerContent)
	} else {
		// Always render header with full height to prevent jiggling
		header = headerStyle.Width(m.width).Height(headerHeight).Align(lipgloss.Center).Padding(1, 0).Render(bannerText)
	}

	// Render tabs
	tabs := m.renderTabs()

	// Content area with safety check
	contentText := "No content available"
	if m.activeTab < len(m.tabs) {
		switch m.activeTab {
		case 0: // Monitoring tab
			if m.focusModeActive {
				contentText = `
🔒 FOCUS MODE ACTIVE

Monitoring is disabled during focus mode.

DNS monitoring is temporarily disabled to help you stay focused.

You can still manage your allowlist.

Press ←/→ to switch to other tabs.`
			} else {
				contentText = m.renderDNSMonitoring()
			}
		case 1: // Allowlist tab
			contentText = m.renderAllowedDomains()
		}
	}

	// Show temporary focus message if present
	if m.focusMessage != "" {
		messageStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#4ADE80")). // Green background
			Foreground(lipgloss.Color("#FFFFFF")). // White text
			Bold(true).
			Padding(1, 2).
			Align(lipgloss.Center)

		contentText = messageStyle.Render(m.focusMessage) + "\n\n" + contentText
	}

	// Apply content style with conditional height
	content := contentStyle.Width(m.width - 4).Height(contentHeight).Render(contentText)

	// Footer with full width
	footer := footerStyle.Width(m.width).Render("Navigation: ←/→ Switch tabs | ↑/↓ Navigate | Space/Enter Add/Remove | F Focus mode | ESC Quit")

	// Combine all elements
	return docStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			tabs,
			content,
			footer,
		),
	)
}

func (m Model) renderDNSMonitoring() string {
	if len(m.monitoring.dnsQueries) == 0 {
		return `
No DNS queries recorded yet.

Try making some web requests to see DNS activity.

Make sure the resolver is running with 'sinkzone resolver'`
	}

	// Reverse the data to show newest entries first (at the top)
	queries := make([]api.DNSQuery, len(m.monitoring.dnsQueries))
	copy(queries, m.monitoring.dnsQueries)
	for i, j := 0, len(queries)-1; i < j; i, j = i+1, j-1 {
		queries[i], queries[j] = queries[j], queries[i]
	}

	// Header
	header := fmt.Sprintf("%-40s %-27s %-20s %-10s\n", "Domain", "Client", "Time", "Status")
	header += strings.Repeat("-", 97) + "\n"

	// Table rows
	var rows []string
	for i, query := range queries {
		// Check if domain is in allowlist
		isInAllowlist := m.isInAllowlist(query.Domain)
		status := "BLOCK"
		if isInAllowlist {
			status = "ALLOW"
		}

		// Truncate domain if too long
		domain := query.Domain
		if len(domain) > 38 {
			domain = domain[:35] + "..."
		}

		// Truncate hostname if too long
		dnsClient := query.Client
		if len(dnsClient) > 25 {
			dnsClient = dnsClient[:22] + "..."
		}

		// Check if this row is selected
		// Since we display newest first (reversed), map cursor position
		isSelected := i == m.monitoring.tableCursor
		recentlyChanged := query.Domain == m.lastChangedDomain && time.Since(m.lastChangeTime) < 2*time.Second

		row := formatTableRow(domain, dnsClient, query.Timestamp, status, isSelected, recentlyChanged)
		rows = append(rows, row)
	}

	// Footer
	footer := fmt.Sprintf("\nLast updated: %s | Press Space/Enter to add domains to allowlist", m.monitoring.lastUpdate.Format("15:04:05"))

	return header + strings.Join(rows, "\n") + footer
}

func (m Model) renderAllowedDomains() string {
	if len(m.allowedDomains.domains) == 0 {
		return `
Allowlist is empty.

Add domains to your allowlist to permit them during focus mode.

Use the Monitoring tab to see which domains are being accessed.`
	}

	// Header - use same format as monitoring tab
	header := fmt.Sprintf("%-40s %-20s %-10s\n", "Domain", "Type", "Status")
	header += strings.Repeat("-", 70) + "\n"

	// Table rows
	var rows []string
	for i, domain := range m.allowedDomains.domains {
		// Determine domain type
		domainType := "EXACT"
		if strings.Contains(domain, "*") {
			domainType = "WILDCARD"
		}

		// Status is always ALLOWED for allowlist
		status := "ALLOWED"

		// Truncate domain if too long
		displayDomain := domain
		if len(displayDomain) > 38 {
			displayDomain = displayDomain[:35] + "..."
		}

		// Check if this row is selected
		isSelected := i == m.allowedDomains.cursor
		recentlyChanged := domain == m.lastChangedDomain && time.Since(m.lastChangeTime) < 2*time.Second

		// Use a custom format function for allowlist rows
		row := formatAllowlistRow(displayDomain, domainType, status, isSelected, recentlyChanged)
		rows = append(rows, row)
	}

	// Footer
	footer := fmt.Sprintf("\nAllowlist (%d domains) | Press Space/Enter to remove domains", len(m.allowedDomains.domains))

	return header + strings.Join(rows, "\n") + footer
}

func formatAllowlistRow(domain string, domainType string, status string, isSelected bool, recentlyChanged bool) string {
	row := fmt.Sprintf("%-40s %-20s %-10s", domain, domainType, status)

	if isSelected && recentlyChanged {
		// Combined state: selected and recently changed - use a distinct color
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#059669")). // Green background for selected + recently changed
			Foreground(lipgloss.Color("#FFFFFF")). // White text
			Padding(0, 1).
			Render(row)
	} else if isSelected {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#3B82F6")). // Blue background for selected
			Foreground(lipgloss.Color("#FFFFFF")). // White text
			Padding(0, 1).
			Render(row)
	} else if recentlyChanged {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#8B5CF6")). // Purple background for recently changed
			Foreground(lipgloss.Color("#FFFFFF")). // White text
			Padding(0, 1).
			Render(row)
	}

	return row
}

func formatTableRow(domain string, dnsClient string, timestamp time.Time, status string, isSelected bool, recentlyChanged bool) string {
	row := fmt.Sprintf("%-40s %-27s %-20s %-10s", domain, dnsClient, timestamp.Format("15:04:05"), status)

	if isSelected && recentlyChanged {
		// Combined state: selected and recently changed - use a distinct color
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#059669")). // Green background for selected + recently changed
			Foreground(lipgloss.Color("#FFFFFF")). // White text
			Padding(0, 1).
			Render(row)
	} else if isSelected {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#3B82F6")). // Blue background for selected
			Foreground(lipgloss.Color("#FFFFFF")). // White text
			Padding(0, 1).
			Render(row)
	} else if recentlyChanged {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#8B5CF6")). // Purple background for recently changed
			Foreground(lipgloss.Color("#FFFFFF")). // White text
			Padding(0, 1).
			Render(row)
	}

	return row
}

func (m *Model) addToAllowlist(domain string) error {
	manager, err := allowlist.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create allowlist manager: %w", err)
	}

	return manager.Add(domain)
}

func (m *Model) removeFromAllowlist(domain string) error {
	manager, err := allowlist.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create allowlist manager: %w", err)
	}

	return manager.Remove(domain)
}

func (m Model) isInAllowlist(domain string) bool {
	for _, allowedDomain := range m.allowedDomains.domains {
		if allowedDomain == domain {
			return true
		}
	}
	return false
}

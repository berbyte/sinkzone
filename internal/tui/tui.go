package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/berbyte/sinkzone/internal/config"
	"github.com/berbyte/sinkzone/internal/database"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Tab int

const (
	OverviewTab Tab = iota
	MonitorTab
	ConfigTab
)

type Model struct {
	width     int
	height    int
	activeTab Tab
	quitting  bool

	// Overview data
	resolverRunning bool
	currentMode     string
	focusEndTime    *time.Time

	// Monitor data
	domains    map[string]*database.DNSQuery
	allowlist  []string
	table      table.Model
	lastUpdate time.Time

	// Database
	db *database.Database
}

// Style definitions inspired by the screenshot aesthetics
var (
	// Colors - dark theme with neon accents
	background = lipgloss.Color("#000000")
	textColor  = lipgloss.Color("#FFFFFF")
	accent1    = lipgloss.Color("#00FFFF") // Cyan
	accent2    = lipgloss.Color("#FF69B4") // Pink
	accent3    = lipgloss.Color("#90EE90") // Light Green
	accent4    = lipgloss.Color("#87CEEB") // Sky Blue
	muted      = lipgloss.Color("#808080") // Grey

	// Banner
	bannerStyle = lipgloss.NewStyle().
			Foreground(accent1).
			Bold(true).
			Margin(1, 0)

	// Tabs - always visible
	tabStyle = lipgloss.NewStyle().
			Foreground(muted).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(accent2).
			Padding(0, 1).
			Bold(true)

	// Content boxes
	contentBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent2).
			Padding(1, 2).
			Background(background)

	// Status
	statusStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(accent1).
			Padding(0, 1).
			Bold(true)

	// Table
	tableStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(accent4).
			Padding(0, 1).
			Background(background)

	// Help
	helpStyle = lipgloss.NewStyle().
			Foreground(muted).
			Italic(true)

	// Document - full terminal usage
	docStyle = lipgloss.NewStyle().
			Background(background).
			Foreground(textColor)
)

// ASCII Banner
const sinkzoneBanner = `
  ██████  ██▓ ███▄    █  ██ ▄█▀▒███████▒ ▒█████   ███▄    █ ▓█████ 
▒██    ▒ ▓██▒ ██ ▀█   █  ██▄█▒ ▒ ▒ ▒ ▄▀░▒██▒  ██▒ ██ ▀█   █ ▓█   ▀ 
░ ▓██▄   ▒██▒▓██  ▀█ ██▒▓███▄░ ░ ▒ ▄▀▒░ ▒██░  ██▒▓██  ▀█ ██▒▒███   
  ▒   ██▒░██░▓██▒  ▐▌██▒▓██ █▄   ▄▀▒   ░▒██   ██░▓██▒  ▐▌██▒▒▓█  ▄ 
▒██████▒▒░██░▒██░   ▓██░▒██▒ █▄▒███████▒░ ████▓▒░▒██░   ▓██░░▒████▒
▒ ▒▓▒ ▒ ░░▓  ░ ▒░   ▒ ▒ ▒ ▒▒ ▓▒░▒▒ ▓░▒░▒░ ▒░▒░▒░ ░ ▒░   ▒ ▒ ░░ ▒░ ░
░ ░▒  ░ ░ ▒ ░░ ░░   ░ ▒░░ ░▒ ▒░░░▒ ▒ ░ ▒  ░ ▒ ▒░ ░ ░░   ░ ▒░ ░ ░  ░
░  ░  ░   ▒ ░   ░   ░ ░ ░ ░░ ░ ░ ░ ░ ░ ░░ ░ ░ ▒     ░   ░ ░    ░   
      ░   ░           ░ ░  ░     ░ ░        ░ ░           ░    ░  ░
                               ░                                   
`

func Start() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database
	configPath := getConfigPath()
	dbPath := filepath.Join(filepath.Dir(configPath), "sinkzone.db")
	db, err := database.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Load allowlist from database
	allowlist, err := db.GetAllowlist()
	if err != nil {
		return fmt.Errorf("failed to load allowlist: %w", err)
	}

	// Initialize domain stats
	domains := make(map[string]*database.DNSQuery)

	// Create table
	columns := []table.Column{
		{Title: "Domain", Width: 30},
		{Title: "Count", Width: 8},
		{Title: "Last Seen", Width: 20},
		{Title: "Status", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	// Enhanced table styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(accent4).
		BorderBottom(true).
		Bold(true).
		Foreground(textColor).
		Background(background)
	s.Selected = s.Selected.
		Foreground(textColor).
		Background(accent2).
		Bold(false)
	t.SetStyles(s)

	m := Model{
		table:        t,
		domains:      domains,
		allowlist:    allowlist,
		activeTab:    OverviewTab,
		db:           db,
		currentMode:  cfg.Mode,
		focusEndTime: cfg.FocusEndTime,
	}

	// Check if resolver is running
	m.resolverRunning = checkResolverRunning()

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust table height to use available space
		tableHeight := m.height - 20 // Account for banner, tabs, and help
		if tableHeight > 0 {
			m.table.SetHeight(tableHeight)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			// Switch between tabs
			m.activeTab = (m.activeTab + 1) % 3
		case "1":
			m.activeTab = OverviewTab
		case "2":
			m.activeTab = MonitorTab
		case "3":
			m.activeTab = ConfigTab
		case "t":
			// Toggle focus/normal mode
			m.toggleFocusMode()
		case "up", "down":
			if m.activeTab == MonitorTab {
				m.table, cmd = m.table.Update(msg)
				return m, cmd
			}
		case "a":
			// Add selected domain to allowlist (only in monitor tab)
			if m.activeTab == MonitorTab && len(m.table.Rows()) > 0 && m.table.Cursor() < len(m.table.Rows()) {
				domain := m.table.Rows()[m.table.Cursor()][0]
				m.addToAllowlist(domain)
			}
		case "r":
			// Remove selected domain from allowlist (only in monitor tab)
			if m.activeTab == MonitorTab && len(m.table.Rows()) > 0 && m.table.Cursor() < len(m.table.Rows()) {
				domain := m.table.Rows()[m.table.Cursor()][0]
				m.removeFromAllowlist(domain)
			}
		}
	case tickMsg:
		// Update stats from database
		m.updateStats()
		m.updateTable()
		// Check resolver status
		m.resolverRunning = checkResolverRunning()
		return m, m.tick()
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	doc := strings.Builder{}

	// Banner
	{
		banner := bannerStyle.Render(sinkzoneBanner)
		doc.WriteString(banner + "\n")
	}

	// Tabs - always visible
	{
		tabs := []string{"1 Overview", "2 Monitor", "3 Config"}
		var tabViews []string

		for i, tabName := range tabs {
			if Tab(i) == m.activeTab {
				tabViews = append(tabViews, activeTabStyle.Render(tabName))
			} else {
				tabViews = append(tabViews, tabStyle.Render(tabName))
			}
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)
		doc.WriteString(row + "\n\n")
	}

	// Content based on active tab
	switch m.activeTab {
	case OverviewTab:
		doc.WriteString(m.renderOverviewContent())
	case MonitorTab:
		doc.WriteString(m.renderMonitorContent())
	case ConfigTab:
		doc.WriteString(m.renderConfigContent())
	}

	// Help
	{
		help := helpStyle.Render("Tab: Switch tabs • 1/2/3: Quick tab switch • t: Toggle focus mode • q: Quit")
		doc.WriteString("\n" + help)
	}

	// Apply full terminal styling
	if m.width > 0 {
		docStyle = docStyle.MaxWidth(m.width).MaxHeight(m.height)
	}

	return docStyle.Render(doc.String())
}

func (m Model) renderOverviewContent() string {
	// Resolver status
	resolverStatus := "● Running"
	resolverColor := accent3
	if !m.resolverRunning {
		resolverStatus = "● Stopped"
		resolverColor = lipgloss.Color("#FF6B6B") // Red
	}

	resolverBox := contentBox.Copy().
		BorderForeground(resolverColor).
		Render(fmt.Sprintf("DNS Resolver\n%s", resolverStatus))

	// Current mode
	modeStatus := fmt.Sprintf("Mode: %s", m.currentMode)
	if m.currentMode == "focus" && m.focusEndTime != nil {
		modeStatus += fmt.Sprintf(" (ends at %s)", m.focusEndTime.Format("15:04:05"))
	}

	modeBox := contentBox.Copy().
		BorderForeground(accent2).
		Render(fmt.Sprintf("Current Status\n%s", modeStatus))

	// Quick stats
	statsBox := contentBox.Copy().
		BorderForeground(accent1).
		Render(fmt.Sprintf("Quick Stats\nDomains tracked: %d\nAllowlist: %d domains",
			len(m.domains), len(m.allowlist)))

	// Layout sections horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		resolverBox,
		modeBox,
		statsBox,
	)
}

func (m Model) renderMonitorContent() string {
	// Status bar
	status := statusStyle.Render(
		fmt.Sprintf("Mode: %s | Allowlist: %d domains | Last Update: %s",
			m.currentMode, len(m.allowlist), m.lastUpdate.Format("15:04:05")),
	)

	// Table
	tableContent := tableStyle.Render(m.table.View())

	// Monitor help
	monitorHelp := helpStyle.Render(
		"↑/↓: Navigate • a: Add to allowlist • r: Remove from allowlist",
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		status,
		tableContent,
		monitorHelp,
	)
}

func (m Model) renderConfigContent() string {
	// Upstream nameservers
	nameserversContent := "Upstream Nameservers:\n"
	for _, ns := range []string{"8.8.8.8:53", "1.1.1.1:53"} {
		nameserversContent += fmt.Sprintf("  • %s\n", ns)
	}

	nameserversBox := contentBox.Copy().
		BorderForeground(accent1).
		Render(nameserversContent)

	// Allowlist
	allowlistContent := fmt.Sprintf("Allowlist (%d domains):\n", len(m.allowlist))
	if len(m.allowlist) == 0 {
		allowlistContent += "  (empty)"
	} else {
		for _, domain := range m.allowlist {
			allowlistContent += fmt.Sprintf("  • %s\n", domain)
		}
	}

	allowlistBox := contentBox.Copy().
		BorderForeground(accent2).
		Render(allowlistContent)

	// Configuration help
	configHelp := helpStyle.Render(
		"Configuration is managed via ~/.sinkzone/sinkzone.yaml",
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, nameserversBox, allowlistBox),
		configHelp,
	)
}

func (m *Model) toggleFocusMode() {
	// Load current config
	cfg, err := config.Load()
	if err != nil {
		return
	}

	if cfg.Mode == "normal" {
		// Switch to focus mode for 1 hour
		cfg.Mode = "focus"
		endTime := time.Now().Add(1 * time.Hour)
		cfg.FocusEndTime = &endTime
		m.currentMode = "focus"
		m.focusEndTime = &endTime
	} else {
		// Switch to normal mode
		cfg.Mode = "normal"
		cfg.FocusEndTime = nil
		m.currentMode = "normal"
		m.focusEndTime = nil
	}

	// Save config
	if err := config.Save(cfg); err != nil {
		return
	}
}

func (m *Model) updateStats() {
	// Get real DNS stats from database
	stats, err := m.db.GetDNSStats()
	if err != nil {
		return
	}

	m.domains = stats
	m.lastUpdate = time.Now()

	// Refresh allowlist
	allowlist, err := m.db.GetAllowlist()
	if err != nil {
		return
	}
	m.allowlist = allowlist
}

func (m *Model) updateTable() {
	var rows []table.Row

	for _, stats := range m.domains {
		status := "Allowed"
		if stats.Blocked {
			status = "Blocked"
		}

		rows = append(rows, table.Row{
			stats.Domain,
			fmt.Sprintf("%d", stats.Count),
			stats.Timestamp.Format("15:04:05"),
			status,
		})
	}

	// Sort by count (descending)
	sort.Slice(rows, func(i, j int) bool {
		countI := 0
		fmt.Sscanf(rows[i][1], "%d", &countI)
		countJ := 0
		fmt.Sscanf(rows[j][1], "%d", &countJ)
		return countI > countJ
	})

	m.table.SetRows(rows)
}

func (m *Model) addToAllowlist(domain string) {
	if !m.isInAllowlist(domain) {
		if err := m.db.AddToAllowlist(domain); err != nil {
			return
		}

		// Refresh allowlist
		allowlist, err := m.db.GetAllowlist()
		if err != nil {
			return
		}
		m.allowlist = allowlist
	}
}

func (m *Model) removeFromAllowlist(domain string) {
	if err := m.db.RemoveFromAllowlist(domain); err != nil {
		return
	}

	// Refresh allowlist
	allowlist, err := m.db.GetAllowlist()
	if err != nil {
		return
	}
	m.allowlist = allowlist
}

func (m *Model) isInAllowlist(domain string) bool {
	for _, allowed := range m.allowlist {
		if domain == allowed || strings.HasSuffix(domain, "."+allowed) {
			return true
		}
	}
	return false
}

func checkResolverRunning() bool {
	// Check if resolver process is running by looking for PID file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	pidFile := filepath.Join(homeDir, ".sinkzone", "resolver.pid")
	if _, err := os.Stat(pidFile); err != nil {
		return false
	}

	// Read PID and check if process exists
	// This is a simplified check - in production you'd want to verify the process is actually running
	return true
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".sinkzone", "sinkzone.yaml")
}

type tickMsg time.Time

func (m Model) tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

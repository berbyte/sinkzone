package monitor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/berbyte/sinkzone/internal/config"
	"github.com/berbyte/sinkzone/internal/database"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type Tab int

const (
	MonitorTab Tab = iota
	ConfigTab
)

type MonitorModel struct {
	table      table.Model
	domains    map[string]*database.DNSQuery
	allowlist  []string
	config     *config.Config
	db         *database.Database
	selected   int
	quitting   bool
	lastUpdate time.Time
	activeTab  Tab
	width      int
	height     int
	form       *huh.Form
	showForm   bool
	bannerTick int // For banner animation
}

// ANSI Banner - SINKZONE
const sinkzoneBanner = `
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║    ███████╗██╗███╗   ██╗██╗  ██╗███████╗ ██████╗ ██████╗ ███████╗███████╗    ║
║    ██╔════╝██║████╗  ██║██║ ██╔╝╚══███╔╝██╔═══██╗██╔══██╗██╔════╝██╔════╝    ║
║    ███████╗██║██╔██╗ ██║█████╔╝   ███╔╝ ██║   ██║██████╔╝█████╗  ███████╗    ║
║    ╚════██║██║██║╚██╗██║██╔═██╗  ███╔╝  ██║   ██║██╔══██╗██╔══╝  ╚════██║    ║
║    ███████║██║██║ ╚████║██║  ██╗███████╗╚██████╔╝██║  ██║███████╗███████║    ║
║    ╚══════╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚══════╝╚══════╝    ║
║                                                                              ║
║                        DNS-based Productivity Tool                           ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
`

// Aesthetic Charmbracelet-style colors for banner
var bannerColors = []string{
	"#FF6B6B", // Coral
	"#4ECDC4", // Turquoise
	"#45B7D1", // Sky Blue
	"#96CEB4", // Mint
	"#FFEAA7", // Cream
	"#DDA0DD", // Plum
	"#98D8C8", // Seafoam
}

// Styles
var (
	bannerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4")).
			Bold(true)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2C3E50")).
			Background(lipgloss.Color("#4ECDC4")).
			Padding(0, 1).
			Bold(true)

	tabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#95A5A6")).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2C3E50")).
			Background(lipgloss.Color("#4ECDC4")).
			Padding(0, 1).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#45B7D1")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7F8C8D")).
			Italic(true)

	containerStyle = lipgloss.NewStyle().
			Padding(1, 2)

	tableStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4ECDC4")).
			Padding(0, 1)
)

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
		BorderForeground(lipgloss.Color("#7D56F4")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(false)
	t.SetStyles(s)

	m := MonitorModel{
		table:     t,
		domains:   domains,
		allowlist: allowlist,
		config:    cfg,
		db:        db,
		activeTab: MonitorTab,
	}

	// Start the program
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run monitor: %w", err)
	}

	return nil
}

func (m MonitorModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	)
}

func (m MonitorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust table height to use available space
		tableHeight := m.height - 18 // Account for banner, tabs, and help
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
			if m.activeTab == MonitorTab {
				m.activeTab = ConfigTab
			} else {
				m.activeTab = MonitorTab
			}
		case "1":
			m.activeTab = MonitorTab
		case "2":
			m.activeTab = ConfigTab
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
		case "f":
			// Show focus mode form
			if m.activeTab == ConfigTab {
				m.showForm = true
				m.form = m.createFocusForm()
			}
		case "n":
			// Switch to normal mode
			if m.activeTab == ConfigTab {
				m.config.Mode = "normal"
				m.config.FocusEndTime = nil
				if err := config.Save(m.config); err != nil {
					log.Printf("Failed to save config: %v", err)
				}
			}
		case "up", "down":
			if m.activeTab == MonitorTab && !m.showForm {
				m.table, cmd = m.table.Update(msg)
				return m, cmd
			}
		case "enter":
			if m.showForm && m.form != nil {
				if err := m.form.Run(); err != nil {
					log.Printf("Form error: %v", err)
				}
				m.showForm = false
				m.form = nil
			}
		case "esc":
			if m.showForm {
				m.showForm = false
				m.form = nil
			}
		}
	case tickMsg:
		// Update stats from database
		m.updateStats()
		m.updateTable()
		// Animate banner
		m.bannerTick = (m.bannerTick + 1) % len(bannerColors)
		return m, m.tick()
	}

	if m.showForm && m.form != nil {
		// Handle form updates
		if _, cmd := m.form.Update(msg); cmd != nil {
			return m, cmd
		}
	} else {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m MonitorModel) renderRainbowBanner() string {
	lines := strings.Split(strings.TrimSpace(sinkzoneBanner), "\n")
	var rainbowLines []string

	for i, line := range lines {
		if line == "" {
			continue
		}

		// Different styling for different parts of the banner
		var style lipgloss.Style
		if strings.Contains(line, "╔") || strings.Contains(line, "╗") ||
			strings.Contains(line, "╚") || strings.Contains(line, "╝") ||
			strings.Contains(line, "║") {
			// Border lines - use accent color
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4ECDC4")).
				Bold(true)
		} else if strings.Contains(line, "DNS-based Productivity Tool") {
			// Subtitle - use muted color
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#95A5A6")).
				Italic(true)
		} else {
			// Main logo - use animated colors
			colorIndex := (i + m.bannerTick) % len(bannerColors)
			color := bannerColors[colorIndex]
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color(color)).
				Bold(true)
		}

		rainbowLines = append(rainbowLines, style.Render(line))
	}

	return strings.Join(rainbowLines, "\n")
}

func (m MonitorModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// Create rainbow banner
	banner := m.renderRainbowBanner()

	// Create tab bar
	monitorTab := "1 Monitor"
	configTab := "2 Config"

	if m.activeTab == MonitorTab {
		monitorTab = activeTabStyle.Render(monitorTab)
		configTab = tabStyle.Render(configTab)
	} else {
		monitorTab = tabStyle.Render(monitorTab)
		configTab = activeTabStyle.Render(configTab)
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Left, monitorTab, configTab)

	// Create main content based on active tab
	var content string
	if m.showForm && m.form != nil {
		content = m.renderForm()
	} else if m.activeTab == MonitorTab {
		content = m.renderMonitorTab()
	} else {
		content = m.renderConfigTab()
	}

	// Create help text
	help := helpStyle.Render("Tab: Switch tabs • 1/2: Quick tab switch • q: Quit")

	// Combine everything with container styling
	return containerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			banner,
			tabBar,
			content,
			help,
		),
	)
}

func (m MonitorModel) renderMonitorTab() string {
	status := statusStyle.Render(
		fmt.Sprintf("Mode: %s | Allowlist: %d domains | Last Update: %s",
			m.config.Mode, len(m.allowlist), m.lastUpdate.Format("15:04:05")),
	)

	monitorHelp := helpStyle.Render(
		"↑/↓: Navigate • a: Add to allowlist • r: Remove from allowlist",
	)

	tableContent := tableStyle.Render(m.table.View())

	return lipgloss.JoinVertical(
		lipgloss.Left,
		status,
		tableContent,
		monitorHelp,
	)
}

func (m MonitorModel) renderConfigTab() string {
	// Configuration view with enhanced styling
	var configContent []string

	// Current mode and focus status
	modeStatus := statusStyle.Render(fmt.Sprintf("Current Mode: %s", m.config.Mode))
	if m.config.Mode == "focus" && m.config.FocusEndTime != nil {
		modeStatus += fmt.Sprintf(" (ends at %s)", m.config.FocusEndTime.Format("15:04:05"))
	}
	configContent = append(configContent, modeStatus)

	// Upstream nameservers
	configContent = append(configContent, "")
	configContent = append(configContent, statusStyle.Render("Upstream Nameservers:"))
	for _, ns := range m.config.UpstreamNameservers {
		configContent = append(configContent, fmt.Sprintf("  • %s", ns))
	}

	// Allowlist
	configContent = append(configContent, "")
	configContent = append(configContent, statusStyle.Render(fmt.Sprintf("Allowlist (%d domains):", len(m.allowlist))))
	if len(m.allowlist) == 0 {
		configContent = append(configContent, "  (empty)")
	} else {
		for _, domain := range m.allowlist {
			configContent = append(configContent, fmt.Sprintf("  • %s", domain))
		}
	}

	// Focus mode controls
	configContent = append(configContent, "")
	configContent = append(configContent, statusStyle.Render("Focus Mode Controls:"))
	configContent = append(configContent, "  Press 'f' to start focus mode")
	configContent = append(configContent, "  Press 'n' to switch to normal mode")

	configHelp := helpStyle.Render(
		"f: Start focus mode • n: Switch to normal mode",
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		configContent...,
	) + "\n" + configHelp
}

func (m MonitorModel) renderForm() string {
	if m.form == nil {
		return "Loading form..."
	}
	return m.form.View()
}

func (m *MonitorModel) createFocusForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Focus Duration").
				Description("Enter duration (e.g., 1h, 30m, 2h30m)").
				Placeholder("1h").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("duration is required")
					}
					_, err := time.ParseDuration(s)
					if err != nil {
						return fmt.Errorf("invalid duration format")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm())
}

type tickMsg time.Time

func (m MonitorModel) tick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *MonitorModel) updateStats() {
	// Get real DNS stats from database
	stats, err := m.db.GetDNSStats()
	if err != nil {
		log.Printf("Failed to get DNS stats: %v", err)
		return
	}

	m.domains = stats
	m.lastUpdate = time.Now()

	// Refresh allowlist
	allowlist, err := m.db.GetAllowlist()
	if err != nil {
		log.Printf("Failed to refresh allowlist: %v", err)
	} else {
		m.allowlist = allowlist
	}
}

func (m *MonitorModel) updateTable() {
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

func (m *MonitorModel) isInAllowlist(domain string) bool {
	for _, allowed := range m.allowlist {
		if domain == allowed || strings.HasSuffix(domain, "."+allowed) {
			return true
		}
	}
	return false
}

func (m *MonitorModel) addToAllowlist(domain string) {
	if !m.isInAllowlist(domain) {
		if err := m.db.AddToAllowlist(domain); err != nil {
			log.Printf("Failed to add domain to allowlist: %v", err)
			return
		}

		// Refresh allowlist
		allowlist, err := m.db.GetAllowlist()
		if err != nil {
			log.Printf("Failed to refresh allowlist: %v", err)
		} else {
			m.allowlist = allowlist
		}

		// Update domain status in table
		if stats, exists := m.domains[domain]; exists {
			stats.Blocked = false
		}
	}
}

func (m *MonitorModel) removeFromAllowlist(domain string) {
	if err := m.db.RemoveFromAllowlist(domain); err != nil {
		log.Printf("Failed to remove domain from allowlist: %v", err)
		return
	}

	// Refresh allowlist
	allowlist, err := m.db.GetAllowlist()
	if err != nil {
		log.Printf("Failed to refresh allowlist: %v", err)
	} else {
		m.allowlist = allowlist
	}

	// Update domain status in table
	if stats, exists := m.domains[domain]; exists {
		stats.Blocked = true
	}
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".sinkzone", "sinkzone.yaml")
}

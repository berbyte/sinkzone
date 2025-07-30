package socket

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type DNSQuery struct {
	Domain    string    `json:"domain"`
	Timestamp time.Time `json:"timestamp"`
	Blocked   bool      `json:"blocked"`
}

type Client struct {
	socketPath string
	conn       net.Conn
	queries    []DNSQuery
	allowlist  []string
	connected  bool

	// Focus mode state
	focusMode    bool
	focusEndTime *time.Time

	// Callbacks for updates
	onAllowlistUpdate func([]string)
	onQueryUpdate     func([]DNSQuery)
	onFocusModeUpdate func(bool, *time.Time)
}

func NewClient() *Client {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	socketPath := filepath.Join(homeDir, ".sinkzone", "sinkzone.sock")

	return &Client{
		socketPath: socketPath,
		queries:    make([]DNSQuery, 0),
		allowlist:  make([]string, 0),
	}
}

func (c *Client) Connect() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to socket: %w", err)
	}

	c.conn = conn
	c.connected = true

	// Start reading from socket
	go c.readSocket()

	return nil
}

func (c *Client) Disconnect() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connected = false
}

func (c *Client) readSocket() {
	scanner := bufio.NewScanner(c.conn)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		command := parts[0]
		data := parts[1]

		switch command {
		case "ALLOWLIST":
			c.handleAllowlist(data)
		case "QUERY":
			c.handleQuery(data)
		case "FOCUS_MODE":
			c.handleFocusMode(data)
		case "HEARTBEAT":
			// Ignore heartbeat for now
		}
	}

	c.connected = false
}

func (c *Client) handleAllowlist(data string) {
	if data == "" {
		c.allowlist = make([]string, 0)
	} else {
		c.allowlist = strings.Split(data, ",")
	}

	// Notify callback if set
	if c.onAllowlistUpdate != nil {
		c.onAllowlistUpdate(c.allowlist)
	}
}

func (c *Client) handleQuery(data string) {
	parts := strings.SplitN(data, ":", 3)
	if len(parts) != 3 {
		return
	}

	domain := parts[0]
	blockedStr := parts[1]
	timestampStr := parts[2]

	blocked, err := strconv.ParseBool(blockedStr)
	if err != nil {
		return
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return
	}

	query := DNSQuery{
		Domain:    domain,
		Timestamp: timestamp,
		Blocked:   blocked,
	}

	// Add to queries (keep last 100)
	c.queries = append(c.queries, query)
	if len(c.queries) > 100 {
		c.queries = c.queries[1:]
	}

	// Notify callback if set
	if c.onQueryUpdate != nil {
		c.onQueryUpdate(c.queries)
	}
}

func (c *Client) handleFocusMode(data string) {
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		return
	}

	focusState := parts[0]
	focusEndTimeStr := parts[1]

	c.focusMode = focusState == "true"

	if focusEndTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, focusEndTimeStr); err == nil {
			c.focusEndTime = &endTime
		} else {
			c.focusEndTime = nil
		}
	} else {
		c.focusEndTime = nil
	}

	// Notify callback if set
	if c.onFocusModeUpdate != nil {
		c.onFocusModeUpdate(c.focusMode, c.focusEndTime)
	}
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) GetQueries() []DNSQuery {
	// Return a copy to avoid race conditions
	queries := make([]DNSQuery, len(c.queries))
	copy(queries, c.queries)
	return queries
}

func (c *Client) GetAllowlist() []string {
	// Return a copy to avoid race conditions
	allowlist := make([]string, len(c.allowlist))
	copy(allowlist, c.allowlist)
	return allowlist
}

func (c *Client) AddToAllowlist(domain string) error {
	if !c.connected {
		return fmt.Errorf("not connected to socket")
	}

	// Send command to add domain to allowlist
	_, err := fmt.Fprintf(c.conn, "ADD_ALLOWLIST:%s\n", domain)
	return err
}

func (c *Client) RemoveFromAllowlist(domain string) error {
	if !c.connected {
		return fmt.Errorf("not connected to socket")
	}

	// Send command to remove domain from allowlist
	_, err := fmt.Fprintf(c.conn, "REMOVE_ALLOWLIST:%s\n", domain)
	return err
}

func (c *Client) SetFocusMode(enabled bool, duration time.Duration) error {
	if !c.connected {
		return fmt.Errorf("not connected to socket")
	}

	// Format: "true:1h" or "false:0"
	durationStr := "0"
	if enabled && duration > 0 {
		durationStr = duration.String()
	}

	enabledStr := "false"
	if enabled {
		enabledStr = "true"
	}

	// Send command to set focus mode
	_, err := fmt.Fprintf(c.conn, "SET_FOCUS_MODE:%s:%s\n", enabledStr, durationStr)
	return err
}

// SetAllowlistUpdateCallback sets a callback function that will be called when the allowlist is updated
func (c *Client) SetAllowlistUpdateCallback(callback func([]string)) {
	c.onAllowlistUpdate = callback
}

// SetQueryUpdateCallback sets a callback function that will be called when queries are updated
func (c *Client) SetQueryUpdateCallback(callback func([]DNSQuery)) {
	c.onQueryUpdate = callback
}

// SetFocusModeUpdateCallback sets a callback function that will be called when focus mode is updated
func (c *Client) SetFocusModeUpdateCallback(callback func(bool, *time.Time)) {
	c.onFocusModeUpdate = callback
}

// GetFocusModeState returns the current focus mode state
func (c *Client) GetFocusModeState() (bool, *time.Time) {
	return c.focusMode, c.focusEndTime
}

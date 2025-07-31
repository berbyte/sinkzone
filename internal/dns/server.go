package dns

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/berbyte/sinkzone/internal/config"
	"github.com/miekg/dns"
)

type DNSQuery struct {
	Domain    string    `json:"domain"`
	Timestamp time.Time `json:"timestamp"`
	Blocked   bool      `json:"blocked"`
}

type Server struct {
	config       *config.Config
	server       *dns.Server
	stateMgr     *config.StateManager
	stateChan    chan config.State
	currentState config.State

	// Socket communication
	socketPath    string
	socket        net.Listener
	recentQueries []DNSQuery
	queriesMutex  sync.RWMutex

	// Allowlist management
	allowlistPath  string
	allowlist      map[string]bool
	allowlistMutex sync.RWMutex

	// Focus mode state (in-memory)
	focusMode    bool
	focusEndTime *time.Time
	focusMutex   sync.RWMutex

	// DNS server configuration
	port string
}

func NewServer(cfg *config.Config) *Server {
	return NewServerWithPort(cfg, "53")
}

func NewServerWithPort(cfg *config.Config, port string) *Server {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	socketPath := filepath.Join(homeDir, ".sinkzone", "sinkzone.sock")
	allowlistPath := filepath.Join(homeDir, ".sinkzone", "allowlist.txt")

	return &Server{
		config:        cfg,
		socketPath:    socketPath,
		allowlistPath: allowlistPath,
		allowlist:     make(map[string]bool),
		recentQueries: make([]DNSQuery, 0, 100),
		port:          port,
	}
}

func (s *Server) Start() error {
	// Load allowlist
	if err := s.loadAllowlist(); err != nil {
		return fmt.Errorf("failed to load allowlist: %w", err)
	}

	// Initialize state manager
	stateMgr, err := config.NewStateManager()
	if err != nil {
		return fmt.Errorf("failed to initialize state manager: %w", err)
	}
	s.stateMgr = stateMgr

	// Initialize state channel
	s.stateChan = make(chan config.State, 10)

	// Start state watching
	s.stateMgr.WatchState(s.stateChan)

	// Start state update goroutine
	go s.handleStateUpdates()

	// Start socket server
	if err := s.startSocketServer(); err != nil {
		return fmt.Errorf("failed to start socket server: %w", err)
	}
	defer s.stopSocketServer()

	// Create PID file
	if err := s.createPIDFile(); err != nil {
		return fmt.Errorf("failed to create PID file: %w", err)
	}
	defer s.cleanupPIDFile()

	dns.HandleFunc(".", s.handleRequest)

	s.server = &dns.Server{
		Addr: ":" + s.port,
		Net:  "udp",
	}

	log.Printf("Starting DNS server on :%s", s.port)
	return s.server.ListenAndServe()
}

func (s *Server) startSocketServer() error {
	// Remove existing socket file if it exists
	os.Remove(s.socketPath)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(s.socketPath), 0755); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Create Unix domain socket
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket: %w", err)
	}

	s.socket = listener

	// Change socket permissions to allow non-root users to connect
	if err := os.Chmod(s.socketPath, 0666); err != nil {
		log.Printf("Warning: failed to change socket permissions: %v", err)
	}

	// Start accepting connections
	go s.acceptConnections()

	log.Printf("Socket server started on %s", s.socketPath)
	return nil
}

func (s *Server) stopSocketServer() {
	if s.socket != nil {
		s.socket.Close()
		os.Remove(s.socketPath)
	}
}

func (s *Server) acceptConnections() {
	for {
		conn, err := s.socket.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Printf("Socket accept error: %v", err)
			}
			return
		}

		go s.handleSocketConnection(conn)
	}
}

func (s *Server) handleSocketConnection(conn net.Conn) {
	defer conn.Close()

	// Send current allowlist
	s.allowlistMutex.RLock()
	allowlist := make([]string, 0, len(s.allowlist))
	for domain := range s.allowlist {
		allowlist = append(allowlist, domain)
	}
	s.allowlistMutex.RUnlock()

	// Send allowlist
	fmt.Fprintf(conn, "ALLOWLIST:%s\n", strings.Join(allowlist, ","))

	// Send focus mode state
	focusMode, focusEndTime := s.getFocusModeState()
	focusState := "false"
	if focusMode {
		focusState = "true"
	}
	focusEndTimeStr := ""
	if focusEndTime != nil {
		focusEndTimeStr = focusEndTime.Format(time.RFC3339)
	}
	fmt.Fprintf(conn, "FOCUS_MODE:%s:%s\n", focusState, focusEndTimeStr)

	// Send recent queries
	s.queriesMutex.RLock()
	for _, query := range s.recentQueries {
		fmt.Fprintf(conn, "QUERY:%s:%t:%s\n",
			query.Domain, query.Blocked, query.Timestamp.Format(time.RFC3339))
	}
	s.queriesMutex.RUnlock()

	// Start reading commands from client
	scanner := bufio.NewScanner(conn)
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
		case "ADD_ALLOWLIST":
			if err := s.addToAllowlist(data); err != nil {
				log.Printf("Failed to add domain to allowlist: %v", err)
			}
		case "REMOVE_ALLOWLIST":
			if err := s.removeFromAllowlist(data); err != nil {
				log.Printf("Failed to remove domain from allowlist: %v", err)
			}
		case "SET_FOCUS_MODE":
			if err := s.setFocusMode(data); err != nil {
				log.Printf("Failed to set focus mode: %v", err)
			}
		}
	}
}

func (s *Server) loadAllowlist() error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(s.allowlistPath), 0755); err != nil {
		return fmt.Errorf("failed to create allowlist directory: %w", err)
	}

	// Load allowlist from file
	if _, err := os.Stat(s.allowlistPath); err == nil {
		file, err := os.Open(s.allowlistPath)
		if err != nil {
			return fmt.Errorf("failed to open allowlist file: %w", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		s.allowlistMutex.Lock()
		s.allowlist = make(map[string]bool)
		for scanner.Scan() {
			domain := strings.TrimSpace(scanner.Text())
			if domain != "" && !strings.HasPrefix(domain, "#") {
				s.allowlist[domain] = true
			}
		}
		s.allowlistMutex.Unlock()

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read allowlist file: %w", err)
		}
	}

	return nil
}

func (s *Server) saveAllowlist() error {
	s.allowlistMutex.RLock()
	domains := make([]string, 0, len(s.allowlist))
	for domain := range s.allowlist {
		domains = append(domains, domain)
	}
	s.allowlistMutex.RUnlock()

	data := strings.Join(domains, "\n") + "\n"
	return os.WriteFile(s.allowlistPath, []byte(data), 0644)
}

func (s *Server) addToAllowlist(domain string) error {
	s.allowlistMutex.Lock()
	s.allowlist[domain] = true
	s.allowlistMutex.Unlock()

	return s.saveAllowlist()
}

func (s *Server) removeFromAllowlist(domain string) error {
	s.allowlistMutex.Lock()
	delete(s.allowlist, domain)
	s.allowlistMutex.Unlock()

	return s.saveAllowlist()
}

func (s *Server) getAllowlist() []string {
	s.allowlistMutex.RLock()
	defer s.allowlistMutex.RUnlock()

	domains := make([]string, 0, len(s.allowlist))
	for domain := range s.allowlist {
		domains = append(domains, domain)
	}
	return domains
}

func (s *Server) getFocusModeState() (bool, *time.Time) {
	s.focusMutex.RLock()
	defer s.focusMutex.RUnlock()
	return s.focusMode, s.focusEndTime
}

func (s *Server) setFocusMode(data string) error {
	// Parse the focus mode command
	// Format: "true:1h" or "false:0"
	parts := strings.Split(data, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid focus mode format: %s", data)
	}

	enabled := parts[0] == "true"
	durationStr := parts[1]

	var duration time.Duration
	if enabled && durationStr != "0" {
		var err error
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}
	}

	// Set focus mode in memory
	s.focusMutex.Lock()
	s.focusMode = enabled
	if enabled && duration > 0 {
		endTime := time.Now().Add(duration)
		s.focusEndTime = &endTime
	} else {
		s.focusEndTime = nil
	}
	s.focusMutex.Unlock()

	log.Printf("Focus mode set to: %t, duration: %v", enabled, duration)
	return nil
}

func (s *Server) addQuery(query DNSQuery) {
	s.queriesMutex.Lock()
	defer s.queriesMutex.Unlock()

	// Add to recent queries (keep last 100)
	s.recentQueries = append(s.recentQueries, query)
	if len(s.recentQueries) > 100 {
		s.recentQueries = s.recentQueries[1:]
	}
}

// handleStateUpdates processes state updates from the state manager
func (s *Server) handleStateUpdates() {
	for state := range s.stateChan {
		s.currentState = state

		if state.FocusMode {
			log.Printf("Focus mode ENABLED (until %s)", state.FocusEndTime.Format("15:04:05"))
		} else {
			log.Printf("Focus mode DISABLED")
		}
	}
}

func (s *Server) createPIDFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pidDir := filepath.Join(homeDir, ".sinkzone")
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	pidFile := filepath.Join(pidDir, "resolver.pid")
	pid := os.Getpid()

	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

func (s *Server) cleanupPIDFile() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	pidFile := filepath.Join(homeDir, ".sinkzone", "resolver.pid")
	os.Remove(pidFile)
}

func (s *Server) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	// Get the domain being requested
	domain := ""
	if len(r.Question) > 0 {
		domain = strings.TrimSuffix(r.Question[0].Name, ".")
	}

	// Check if we're in focus mode
	s.focusMutex.RLock()
	focusMode := s.focusMode
	focusEndTime := s.focusEndTime
	s.focusMutex.RUnlock()

	// Check for expiration
	if focusMode && focusEndTime != nil && time.Now().After(*focusEndTime) {
		// Focus mode has expired, disable it
		s.focusMutex.Lock()
		s.focusMode = false
		s.focusEndTime = nil
		s.focusMutex.Unlock()
		focusMode = false
		log.Printf("Focus mode expired and disabled")
	}

	// Log the request and record query
	if domain != "" {
		blocked := focusMode && !s.isAllowed(domain)

		// Add to recent queries
		query := DNSQuery{
			Domain:    domain,
			Timestamp: time.Now(),
			Blocked:   blocked,
		}
		s.addQuery(query)

		// Check if domain is in allowlist for logging purposes
		isAllowed := s.isAllowed(domain)

		if focusMode {
			if blocked {
				log.Printf("BLOCKED: %s (focus mode active)", domain)
			} else {
				log.Printf("ALLOWED: %s (in allowlist)", domain)
			}
		} else {
			// In normal mode, show what would happen if focus mode were active
			if isAllowed {
				log.Printf("DNS request: %s (normal mode) - would be ALLOWED in focus mode", domain)
			} else {
				log.Printf("DNS request: %s (normal mode) - would be BLOCKED in focus mode", domain)
			}
		}
	}

	// If in focus mode, check allowlist
	if focusMode {
		if !s.isAllowed(domain) {
			// Return NXDOMAIN for blocked domains
			msg.SetRcode(r, dns.RcodeNameError)

			// Add SOA record for negative response with 5-minute TTL
			soa := &dns.SOA{
				Hdr: dns.RR_Header{
					Name:   r.Question[0].Name,
					Rrtype: dns.TypeSOA,
					Class:  dns.ClassINET,
					Ttl:    300, // 5 minutes
				},
				Ns:      "sinkzone.local.",
				Mbox:    "admin.sinkzone.local.",
				Serial:  uint32(time.Now().Unix()),
				Refresh: 300,
				Retry:   300,
				Expire:  300,
				Minttl:  300,
			}
			msg.Ns = append(msg.Ns, soa)

			w.WriteMsg(&msg)
			return
		}
	}

	// Forward to upstream nameservers
	response, err := s.forward(r)
	if err != nil {
		log.Printf("Forward error: %v", err)
		msg.SetRcode(r, dns.RcodeServerFailure)
		w.WriteMsg(&msg)
		return
	}

	w.WriteMsg(response)
}

func (s *Server) forward(r *dns.Msg) (*dns.Msg, error) {
	client := &dns.Client{
		Timeout: 5 * time.Second,
	}

	for _, upstream := range s.config.GetUpstreamAddresses() {
		response, _, err := client.Exchange(r, upstream)
		if err == nil {
			return response, nil
		}
		log.Printf("Upstream %s failed: %v", upstream, err)
	}

	return nil, fmt.Errorf("all upstream nameservers failed")
}

func (s *Server) isAllowed(domain string) bool {
	s.allowlistMutex.RLock()
	defer s.allowlistMutex.RUnlock()
	return s.allowlist[domain]
}

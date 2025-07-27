package dns

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/berbyte/sinkzone/internal/config"
	"github.com/berbyte/sinkzone/internal/database"
	"github.com/miekg/dns"
)

type Server struct {
	config       *config.Config
	server       *dns.Server
	db           *database.Database
	stateMgr     *config.StateManager
	stateChan    chan config.State
	currentState config.State
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
	}
}

func (s *Server) Start() error {
	// Initialize database
	configPath := getConfigPath()
	dbPath := filepath.Join(filepath.Dir(configPath), "sinkzone.db")
	db, err := database.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	s.db = db
	defer s.db.Close()

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

	// Create PID file
	if err := s.createPIDFile(); err != nil {
		return fmt.Errorf("failed to create PID file: %w", err)
	}
	defer s.cleanupPIDFile()

	dns.HandleFunc(".", s.handleRequest)

	s.server = &dns.Server{
		Addr: ":53",
		Net:  "udp",
	}

	log.Printf("Starting DNS server on :53")
	return s.server.ListenAndServe()
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
	focusMode := s.currentState.FocusMode

	// Check for expiration
	if focusMode && s.currentState.FocusEndTime != nil && time.Now().After(*s.currentState.FocusEndTime) {
		// Focus mode has expired, disable it
		if err := s.stateMgr.SetFocusMode(false, 0); err != nil {
			log.Printf("Failed to disable expired focus mode: %v", err)
		}
		focusMode = false
	}

	// Log the request and record in database
	if domain != "" {
		blocked := focusMode && !s.isAllowed(domain)
		if err := s.db.RecordDNSQuery(domain, blocked); err != nil {
			log.Printf("Failed to record DNS query: %v", err)
		}

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
	// Check if domain is in allowlist using database
	allowed, err := s.db.IsInAllowlist(domain)
	if err != nil {
		log.Printf("Failed to check allowlist: %v", err)
		return false
	}
	return allowed
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".sinkzone", "sinkzone.yaml")
}

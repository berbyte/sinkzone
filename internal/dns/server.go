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
	config *config.Config
	server *dns.Server
	db     *database.Database
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

	// Check if we're in focus mode
	if s.config.Mode == "focus" {
		// Check if focus mode has expired
		if s.config.FocusEndTime != nil && time.Now().After(*s.config.FocusEndTime) {
			s.config.Mode = "normal"
			s.config.FocusEndTime = nil
			if err := config.Save(s.config); err != nil {
				log.Printf("Failed to update expired focus mode: %v", err)
			}
		}
	}

	// Get the domain being requested
	domain := ""
	if len(r.Question) > 0 {
		domain = strings.TrimSuffix(r.Question[0].Name, ".")
	}

	// Log the request and record in database
	if domain != "" {
		blocked := s.config.Mode == "focus" && !s.isAllowed(domain)
		if err := s.db.RecordDNSQuery(domain, blocked); err != nil {
			log.Printf("Failed to record DNS query: %v", err)
		}
		log.Printf("DNS request: %s (mode: %s)", domain, s.config.Mode)
	}

	// If in focus mode, check allowlist
	if s.config.Mode == "focus" {
		if !s.isAllowed(domain) {
			log.Printf("Blocked: %s", domain)
			// Return NXDOMAIN for blocked domains
			msg.SetRcode(r, dns.RcodeNameError)
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

	for _, upstream := range s.config.UpstreamNameservers {
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

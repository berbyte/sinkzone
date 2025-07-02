package dns

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/berbyte/sinkzone/database"
	"github.com/berbyte/sinkzone/pkg/filter"
	"github.com/miekg/dns"
)

type Server struct {
	upstream string
	db       *database.DB
	server   *dns.Server
	handler  *DNSHandler
	mu       sync.RWMutex
	running  bool
}

type DNSHandler struct {
	upstream string
	db       *database.DB
	filter   *filter.Filter
}

func NewServer(upstream string, db *database.DB) *Server {
	handler := &DNSHandler{
		upstream: upstream,
		db:       db,
		filter:   filter.New(),
	}

	return &Server{
		upstream: upstream,
		db:       db,
		handler:  handler,
	}
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("DNS server is already running")
	}

	// Check if running as root (required for port 53)
	if os.Geteuid() != 0 {
		return fmt.Errorf("DNS server must be run as root to bind to port 53")
	}

	// Create DNS server for UDP
	s.server = &dns.Server{
		Addr:    ":53",
		Net:     "udp",
		Handler: s.handler,
	}

	fmt.Printf("Starting DNS server on port 53\n")
	fmt.Printf("Forwarding requests to: %s\n", s.upstream)

	s.running = true

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Printf("DNS server error: %v", err)
		}
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.server == nil {
		return fmt.Errorf("DNS server is not running")
	}

	if err := s.server.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown DNS server: %v", err)
	}

	s.running = false
	fmt.Println("DNS server stopped")
	return nil
}

func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (h *DNSHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	clientIP := w.RemoteAddr().String()
	if strings.Contains(clientIP, ":") {
		clientIP = strings.Split(clientIP, ":")[0]
	}

	// Get current mode
	mode, err := h.db.GetMode()
	if err != nil {
		log.Printf("Error getting mode: %v", err)
		mode = "off"
	}

	// Process each question
	for _, question := range r.Question {
		domain := strings.TrimSuffix(question.Name, ".")
		queryType := dns.TypeToString[question.Qtype]

		fmt.Printf("DNS Query: %s (Type: %s) from %s\n", domain, queryType, clientIP)

		// Check if domain should be blocked based on mode and rules
		blocked := h.shouldBlock(domain, mode)

		// Log the query
		if err := h.db.LogDNSQuery(domain, clientIP, queryType, blocked); err != nil {
			log.Printf("Error logging DNS query: %v", err)
		}

		if blocked {
			fmt.Printf("BLOCKED: %s (Mode: %s)\n", domain, mode)
			// Send NXDOMAIN response for blocked domains
			response := new(dns.Msg)
			response.SetReply(r)
			response.Rcode = dns.RcodeNameError
			w.WriteMsg(response)
			return
		}
	}

	// Forward the request to upstream DNS
	client := &dns.Client{
		Timeout: 10 * time.Second,
	}

	resp, _, err := client.Exchange(r, h.upstream)
	if err != nil {
		log.Printf("Error forwarding DNS request: %v", err)
		response := new(dns.Msg)
		response.SetReply(r)
		response.Rcode = dns.RcodeServerFailure
		w.WriteMsg(response)
		return
	}

	w.WriteMsg(resp)
}

func (h *DNSHandler) shouldBlock(domain, mode string) bool {
	// Get custom domain rules
	rules := make(map[string]string)
	action, err := h.db.GetDomainAction(domain)
	if err != nil {
		log.Printf("Error checking domain action: %v", err)
	} else if action != "" {
		rules[domain] = action
	}

	// Use the filter package to determine if domain should be blocked
	return h.filter.ShouldBlock(domain, mode, rules)
}

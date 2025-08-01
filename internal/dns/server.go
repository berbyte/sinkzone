package dns

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/berbyte/sinkzone/internal/api"
	"github.com/berbyte/sinkzone/internal/config"
	"github.com/miekg/dns"
)

type Server struct {
	config *config.Config
	server *dns.Server
	port   string

	// API server reference
	apiServer *api.Server

	// Allowlist management
	allowlistPath    string
	allowlist        map[string]bool  // Exact domain matches
	wildcardPatterns []*regexp.Regexp // Compiled wildcard patterns
	allowlistMutex   sync.RWMutex

	// Focus mode state (in-memory)
	focusMode    bool
	focusEndTime *time.Time
	focusMutex   sync.RWMutex
}

func NewServer(cfg *config.Config, apiServer *api.Server) *Server {
	return NewServerWithPort(cfg, apiServer, "53")
}

func NewServerWithPort(cfg *config.Config, apiServer *api.Server, port string) *Server {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	allowlistPath := filepath.Join(homeDir, ".sinkzone", "allowlist.txt")

	return &Server{
		config:        cfg,
		apiServer:     apiServer,
		allowlistPath: allowlistPath,
		allowlist:     make(map[string]bool),
		port:          port,
	}
}

// wildcardToRegex converts a wildcard pattern to a regex pattern
// Examples:
//
//	"*github*" -> ".*github.*"
//	"*.example.com" -> ".*\\.example\\.com"
//	"api.*.com" -> "api\\..*\\.com"
func wildcardToRegex(pattern string) (*regexp.Regexp, error) {
	// Escape regex special characters except *
	escaped := regexp.QuoteMeta(pattern)

	// Replace escaped asterisks with .*
	// We need to handle the case where * was escaped by QuoteMeta
	escaped = strings.ReplaceAll(escaped, "\\*", ".*")

	// Add anchors to ensure full match
	regexPattern := "^" + escaped + "$"

	return regexp.Compile(regexPattern)
}

// isWildcardPattern checks if a pattern contains wildcards
func isWildcardPattern(pattern string) bool {
	return strings.Contains(pattern, "*")
}

func (s *Server) Start() error {
	// Load allowlist
	if err := s.loadAllowlist(); err != nil {
		return fmt.Errorf("failed to load allowlist: %w", err)
	}

	// Set up API server callback for focus mode changes
	if s.apiServer != nil {
		s.apiServer.SetFocusModeCallback(s.setFocusMode)
	}

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

func (s *Server) loadAllowlist() error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(s.allowlistPath), 0750); err != nil {
		return fmt.Errorf("failed to create allowlist directory: %w", err)
	}

	// Load allowlist from file
	if _, err := os.Stat(s.allowlistPath); err == nil {
		// #nosec G304 -- s.allowlistPath is a hardcoded path from user home directory
		file, err := os.Open(s.allowlistPath)
		if err != nil {
			return fmt.Errorf("failed to open allowlist file: %w", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("Warning: failed to close allowlist file: %v", err)
			}
		}()

		scanner := bufio.NewScanner(file)
		s.allowlistMutex.Lock()
		s.allowlist = make(map[string]bool)
		s.wildcardPatterns = nil // Reset wildcard patterns

		for scanner.Scan() {
			pattern := strings.TrimSpace(scanner.Text())
			if pattern != "" && !strings.HasPrefix(pattern, "#") {
				if isWildcardPattern(pattern) {
					// Compile wildcard pattern
					if regex, err := wildcardToRegex(pattern); err == nil {
						s.wildcardPatterns = append(s.wildcardPatterns, regex)
						log.Printf("Loaded wildcard pattern: %s", pattern)
					} else {
						log.Printf("Warning: invalid wildcard pattern '%s': %v", pattern, err)
					}
				} else {
					// Exact domain match
					s.allowlist[pattern] = true
				}
			}
		}
		s.allowlistMutex.Unlock()

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read allowlist file: %w", err)
		}
	}

	return nil
}

func (s *Server) setFocusMode(enabled bool, duration time.Duration) error {
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

	// Reload allowlist when enabling focus mode to pick up any changes
	if enabled {
		if err := s.loadAllowlist(); err != nil {
			log.Printf("Warning: failed to reload allowlist: %v", err)
		} else {
			log.Printf("Allowlist reloaded for focus session")
		}
	}

	log.Printf("Focus mode set to: %t, duration: %v", enabled, duration)
	return nil
}

func (s *Server) createPIDFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	pidDir := filepath.Join(homeDir, ".sinkzone")
	if err := os.MkdirAll(pidDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	pidFile := filepath.Join(pidDir, "resolver.pid")
	pid := os.Getpid()

	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0600); err != nil {
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
	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: failed to remove PID file: %v", err)
	}
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

		// Add to API server if available
		if s.apiServer != nil {
			query := api.DNSQuery{
				Domain:    domain,
				Timestamp: time.Now(),
				Blocked:   blocked,
			}
			s.apiServer.AddQuery(query)
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
				Serial:  getDNSSerial(),
				Refresh: 300,
				Retry:   300,
				Expire:  300,
				Minttl:  300,
			}
			msg.Ns = append(msg.Ns, soa)

			if err := w.WriteMsg(&msg); err != nil {
				log.Printf("Warning: failed to write DNS response: %v", err)
			}
			return
		}
	}

	// Forward to upstream nameservers
	response, err := s.forward(r)
	if err != nil {
		log.Printf("Forward error: %v", err)
		msg.SetRcode(r, dns.RcodeServerFailure)
		if err := w.WriteMsg(&msg); err != nil {
			log.Printf("Warning: failed to write DNS error response: %v", err)
		}
		return
	}

	if err := w.WriteMsg(response); err != nil {
		log.Printf("Warning: failed to write DNS response: %v", err)
	}
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

// getDNSSerial returns a safe DNS serial number
func getDNSSerial() uint32 {
	// Use current time as serial, but ensure it fits in uint32
	// DNS serial numbers are not security-critical, so overflow is acceptable
	unixTime := time.Now().Unix()
	if unixTime < 0 {
		return 0
	}
	if unixTime > 0x7FFFFFFF {
		return 0x7FFFFFFF
	}
	return uint32(unixTime)
}

func (s *Server) isAllowed(domain string) bool {
	s.allowlistMutex.RLock()
	defer s.allowlistMutex.RUnlock()

	// Check exact match first
	if s.allowlist[domain] {
		return true
	}

	// Check wildcard patterns
	for _, pattern := range s.wildcardPatterns {
		if pattern.MatchString(domain) {
			return true
		}
	}

	return false
}

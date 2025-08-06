package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type DNSQuery struct {
	Domain    string    `json:"domain"`
	Timestamp time.Time `json:"timestamp"`
	Blocked   bool      `json:"blocked"`
}

type FocusModeState struct {
	Enabled  bool       `json:"enabled"`
	EndTime  *time.Time `json:"end_time,omitempty"`
	Duration string     `json:"duration,omitempty"`
}

type ResolverState struct {
	FocusMode FocusModeState `json:"focus_mode"`
	Queries   []DNSQuery     `json:"queries"`
}

type Server struct {
	port string
	addr string

	// State management - using map for unique hostnames with timestamps and blocked status
	queryMap      map[string]DNSQuery // hostname -> DNSQuery (with timestamp and blocked status)
	queryMapMutex sync.RWMutex

	focusMode    bool
	focusEndTime *time.Time
	focusMutex   sync.RWMutex

	// Callbacks for DNS server communication
	onFocusModeChange func(enabled bool, duration time.Duration) error
}

func NewServer(port string) *Server {
	return &Server{
		port:     port,
		addr:     ":" + port,
		queryMap: make(map[string]DNSQuery),
	}
}

func (s *Server) SetFocusModeCallback(callback func(enabled bool, duration time.Duration) error) {
	s.onFocusModeChange = callback
}

// loggingMiddleware logs all HTTP requests with method, path, and response status
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		responseWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log the incoming request
		log.Printf("API Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Call the next handler
		next.ServeHTTP(responseWriter, r)

		// Log the response
		duration := time.Since(start)
		log.Printf("API Response: %s %s - %d (%v)", r.Method, r.URL.Path, responseWriter.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) Start() error {
	r := mux.NewRouter()

	// Add logging middleware
	r.Use(s.loggingMiddleware)

	// API routes
	r.HandleFunc("/api/queries", s.handleGetQueries).Methods("GET")
	r.HandleFunc("/api/focus", s.handleGetFocusMode).Methods("GET")
	r.HandleFunc("/api/focus", s.handleSetFocusMode).Methods("POST")
	r.HandleFunc("/api/state", s.handleGetState).Methods("GET")

	// Health check
	r.HandleFunc("/health", s.handleHealth).Methods("GET")

	server := &http.Server{
		Addr:              s.addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	log.Printf("API server starting on %s", s.addr)
	return server.ListenAndServe()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	log.Printf("Health check request from %s", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		// Log error but don't return it since we can't change the response now
		log.Printf("Warning: failed to write health response: %v", err)
	}
}

func (s *Server) handleGetQueries(w http.ResponseWriter, r *http.Request) {
	log.Printf("Get queries request from %s", r.RemoteAddr)

	s.queryMapMutex.RLock()
	defer s.queryMapMutex.RUnlock()

	// Convert map to sorted slice of DNSQuery
	queries := s.getSortedQueries()

	// Return last 100 queries
	if len(queries) > 100 {
		queries = queries[len(queries)-100:]
	}

	log.Printf("Returning %d unique queries", len(queries))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(queries); err != nil {
		log.Printf("Error encoding queries response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleGetFocusMode(w http.ResponseWriter, r *http.Request) {
	log.Printf("Get focus mode request from %s", r.RemoteAddr)

	s.focusMutex.RLock()
	defer s.focusMutex.RUnlock()

	state := FocusModeState{
		Enabled: s.focusMode,
		EndTime: s.focusEndTime,
	}

	log.Printf("Focus mode state: enabled=%v, endTime=%v", s.focusMode, s.focusEndTime)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(state); err != nil {
		log.Printf("Error encoding focus mode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleSetFocusMode(w http.ResponseWriter, r *http.Request) {
	log.Printf("Set focus mode request from %s", r.RemoteAddr)

	var req struct {
		Enabled  bool   `json:"enabled"`
		Duration string `json:"duration,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding focus mode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Focus mode request: enabled=%v, duration=%s", req.Enabled, req.Duration)

	var duration time.Duration
	var err error
	if req.Enabled && req.Duration != "" {
		duration, err = time.ParseDuration(req.Duration)
		if err != nil {
			log.Printf("Invalid duration format: %s", req.Duration)
			http.Error(w, "Invalid duration format", http.StatusBadRequest)
			return
		}
	}

	// Update focus mode
	s.focusMutex.Lock()
	s.focusMode = req.Enabled
	if req.Enabled && duration > 0 {
		endTime := time.Now().Add(duration)
		s.focusEndTime = &endTime
		log.Printf("Focus mode enabled until %v", endTime)
	} else {
		s.focusEndTime = nil
		if req.Enabled {
			log.Printf("Focus mode enabled indefinitely")
		} else {
			log.Printf("Focus mode disabled")
		}
	}
	s.focusMutex.Unlock()

	// Call DNS server callback if set
	if s.onFocusModeChange != nil {
		if err := s.onFocusModeChange(req.Enabled, duration); err != nil {
			log.Printf("Error updating focus mode in DNS server: %v", err)
			http.Error(w, fmt.Sprintf("Failed to update focus mode: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Focus mode updated successfully")
}

func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	log.Printf("Get state request from %s", r.RemoteAddr)

	s.focusMutex.RLock()
	s.queryMapMutex.RLock()

	// Convert map to sorted slice of DNSQuery
	queries := s.getSortedQueries()

	state := ResolverState{
		FocusMode: FocusModeState{
			Enabled: s.focusMode,
			EndTime: s.focusEndTime,
		},
		Queries: queries,
	}

	// Limit to last 100 queries
	if len(state.Queries) > 100 {
		state.Queries = state.Queries[len(state.Queries)-100:]
	}

	s.focusMutex.RUnlock()
	s.queryMapMutex.RUnlock()

	log.Printf("Returning state with %d unique queries, focus mode: %v", len(state.Queries), s.focusMode)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(state); err != nil {
		log.Printf("Error encoding state response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getSortedQueries converts the query map to a sorted slice of DNSQuery
// This method assumes the caller holds the appropriate read lock
func (s *Server) getSortedQueries() []DNSQuery {
	// Create a slice to hold the queries for sorting
	var entries []DNSQuery
	for _, query := range s.queryMap {
		entries = append(entries, query)
	}

	// Sort by timestamp (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	return entries
}

// AddQuery adds a new DNS query to the server's query history
// Now updates the timestamp for existing domains or adds new ones
func (s *Server) AddQuery(query DNSQuery) {
	s.queryMapMutex.Lock()
	defer s.queryMapMutex.Unlock()

	// Update or add the domain with the current timestamp and blocked status
	s.queryMap[query.Domain] = query

	// Keep only the last 100 unique domains
	if len(s.queryMap) > 100 {
		// Convert to slice for sorting
		var entries []DNSQuery
		for _, query := range s.queryMap {
			entries = append(entries, query)
		}

		// Sort by timestamp (oldest first)
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp.Before(entries[j].Timestamp)
		})

		// Remove oldest entries to keep only 100
		entriesToRemove := len(entries) - 100
		for i := 0; i < entriesToRemove; i++ {
			delete(s.queryMap, entries[i].Domain)
		}
	}

	log.Printf("DNS Query: %s (blocked: %v) - Updated timestamp", query.Domain, query.Blocked)
}

// GetFocusMode returns the current focus mode state
func (s *Server) GetFocusMode() (bool, *time.Time) {
	s.focusMutex.RLock()
	defer s.focusMutex.RUnlock()
	return s.focusMode, s.focusEndTime
}

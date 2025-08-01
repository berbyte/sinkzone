package api

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	// State management
	queries      []DNSQuery
	queriesMutex sync.RWMutex

	focusMode    bool
	focusEndTime *time.Time
	focusMutex   sync.RWMutex

	// Callbacks for DNS server communication
	onFocusModeChange func(enabled bool, duration time.Duration) error
}

func NewServer(port string) *Server {
	return &Server{
		port: port,
		addr: ":" + port,
	}
}

func (s *Server) SetFocusModeCallback(callback func(enabled bool, duration time.Duration) error) {
	s.onFocusModeChange = callback
}

func (s *Server) Start() error {
	r := mux.NewRouter()

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

	fmt.Printf("API server starting on %s\n", s.addr)
	return server.ListenAndServe()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		// Log error but don't return it since we can't change the response now
		fmt.Printf("Warning: failed to write health response: %v", err)
	}
}

func (s *Server) handleGetQueries(w http.ResponseWriter, r *http.Request) {
	s.queriesMutex.RLock()
	defer s.queriesMutex.RUnlock()

	// Return last 100 queries
	queries := s.queries
	if len(queries) > 100 {
		queries = queries[len(queries)-100:]
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(queries); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleGetFocusMode(w http.ResponseWriter, r *http.Request) {
	s.focusMutex.RLock()
	defer s.focusMutex.RUnlock()

	state := FocusModeState{
		Enabled: s.focusMode,
		EndTime: s.focusEndTime,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleSetFocusMode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled  bool   `json:"enabled"`
		Duration string `json:"duration,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var duration time.Duration
	var err error
	if req.Enabled && req.Duration != "" {
		duration, err = time.ParseDuration(req.Duration)
		if err != nil {
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
	} else {
		s.focusEndTime = nil
	}
	s.focusMutex.Unlock()

	// Call DNS server callback if set
	if s.onFocusModeChange != nil {
		if err := s.onFocusModeChange(req.Enabled, duration); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update focus mode: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	s.focusMutex.RLock()
	s.queriesMutex.RLock()

	state := ResolverState{
		FocusMode: FocusModeState{
			Enabled: s.focusMode,
			EndTime: s.focusEndTime,
		},
		Queries: s.queries,
	}

	// Limit to last 100 queries
	if len(state.Queries) > 100 {
		state.Queries = state.Queries[len(state.Queries)-100:]
	}

	s.focusMutex.RUnlock()
	s.queriesMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// AddQuery adds a new DNS query to the server's query history
func (s *Server) AddQuery(query DNSQuery) {
	s.queriesMutex.Lock()
	defer s.queriesMutex.Unlock()

	s.queries = append(s.queries, query)
	if len(s.queries) > 100 {
		s.queries = s.queries[1:]
	}
}

// GetFocusMode returns the current focus mode state
func (s *Server) GetFocusMode() (bool, *time.Time) {
	s.focusMutex.RLock()
	defer s.focusMutex.RUnlock()
	return s.focusMode, s.focusEndTime
}

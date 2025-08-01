package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// State represents the real-time state that can be shared between processes
type State struct {
	FocusMode    bool       `json:"focus_mode"`
	FocusEndTime *time.Time `json:"focus_end_time,omitempty"`
	LastUpdated  time.Time  `json:"last_updated"`
}

// StateManager handles real-time state updates
type StateManager struct {
	statePath string
	mu        sync.RWMutex
	state     State
	listeners []chan State
}

// NewStateManager creates a new state manager
func NewStateManager() (*StateManager, error) {
	statePath, err := getStatePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get state path: %w", err)
	}

	sm := &StateManager{
		statePath: statePath,
		listeners: make([]chan State, 0),
	}

	// Load initial state
	if err := sm.loadState(); err != nil {
		// Create default state if file doesn't exist
		sm.state = State{
			FocusMode:   false,
			LastUpdated: time.Now(),
		}
		if err := sm.saveState(); err != nil {
			return nil, fmt.Errorf("failed to create default state: %w", err)
		}
	}

	return sm, nil
}

// GetState returns the current state (thread-safe)
func (sm *StateManager) GetState() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state
}

// SetFocusMode updates the focus mode state
func (sm *StateManager) SetFocusMode(enabled bool, duration time.Duration) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.state.FocusMode = enabled
	sm.state.LastUpdated = time.Now()

	if enabled && duration > 0 {
		endTime := time.Now().Add(duration)
		sm.state.FocusEndTime = &endTime
	} else {
		sm.state.FocusEndTime = nil
	}

	// Save to file
	if err := sm.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Notify listeners
	sm.notifyListeners()

	return nil
}

// CheckFocusMode checks if focus mode is active and handles expiration
func (sm *StateManager) CheckFocusMode() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if focus mode has expired
	if sm.state.FocusMode && sm.state.FocusEndTime != nil && time.Now().After(*sm.state.FocusEndTime) {
		sm.state.FocusMode = false
		sm.state.FocusEndTime = nil
		sm.state.LastUpdated = time.Now()

		// Save updated state
		if err := sm.saveState(); err != nil {
			// Log error but don't fail
			fmt.Printf("Warning: failed to save expired focus state: %v\n", err)
		}

		// Notify listeners
		sm.notifyListeners()
	}

	return sm.state.FocusMode
}

// AddListener adds a channel to receive state updates
func (sm *StateManager) AddListener(ch chan State) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.listeners = append(sm.listeners, ch)
}

// RemoveListener removes a channel from listeners
func (sm *StateManager) RemoveListener(ch chan State) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, listener := range sm.listeners {
		if listener == ch {
			sm.listeners = append(sm.listeners[:i], sm.listeners[i+1:]...)
			break
		}
	}
}

// notifyListeners sends state updates to all listeners
func (sm *StateManager) notifyListeners() {
	for _, ch := range sm.listeners {
		select {
		case ch <- sm.state:
		default:
			// Channel is full or closed, skip
		}
	}
}

// loadState loads state from file
func (sm *StateManager) loadState() error {
	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	if err := json.Unmarshal(data, &sm.state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	return nil
}

// saveState saves state to file
func (sm *StateManager) saveState() error {
	// Ensure directory exists with proper permissions
	dir := filepath.Dir(sm.statePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Try to create the file with user permissions first
	data, err := json.MarshalIndent(sm.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Try to write the file
	if err := os.WriteFile(sm.statePath, data, 0600); err != nil {
		// If we can't write to the file, try to create it in a user-writable location
		if os.IsPermission(err) {
			// Try to create the file in a temporary location first
			tempFile := sm.statePath + ".tmp"
			if writeErr := os.WriteFile(tempFile, data, 0600); writeErr == nil {
				// Try to move it to the final location
				if moveErr := os.Rename(tempFile, sm.statePath); moveErr == nil {
					return nil
				}
				// Clean up temp file
				if removeErr := os.Remove(tempFile); removeErr != nil {
					// Log but don't fail - this is cleanup
					fmt.Printf("Warning: failed to remove temp file %s: %v\n", tempFile, removeErr)
				}
			}
		}
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// WatchState starts watching for state changes (for resolver)
func (sm *StateManager) WatchState(updateChan chan State) {
	// Send initial state
	updateChan <- sm.GetState()

	// Start file watcher
	go func() {
		lastMod := time.Time{}

		for {
			// Check file modification time
			if info, err := os.Stat(sm.statePath); err == nil {
				if info.ModTime().After(lastMod) {
					// File was modified, reload state
					if err := sm.loadState(); err == nil {
						sm.mu.RLock()
						state := sm.state
						sm.mu.RUnlock()

						// Check for expiration
						sm.CheckFocusMode()

						// Send updated state
						select {
						case updateChan <- state:
						default:
							// Channel is full, skip
						}

						lastMod = info.ModTime()
					}
				}
			}

			// Check every 100ms for changes
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// getStatePath returns the platform-specific path for the state file
func getStatePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if runtime.GOOS == "windows" {
		// On Windows, use AppData for better compatibility
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "sinkzone", "state.json"), nil
		}
		// Fallback to user home directory
		return filepath.Join(homeDir, "sinkzone", "state.json"), nil
	}

	// Unix-like systems use ~/.sinkzone/
	return filepath.Join(homeDir, ".sinkzone", "state.json"), nil
}

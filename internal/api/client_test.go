package api

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.baseURL != "http://127.0.0.1:8080" {
		t.Errorf("Expected baseURL to be 'http://127.0.0.1:8080', got '%s'", client.baseURL)
	}
}

func TestClientTimeout(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080")
	if client.client.Timeout != 10*time.Second {
		t.Errorf("Expected timeout to be 10 seconds, got %v", client.client.Timeout)
	}
}

// Note: These tests require a running resolver to pass
// They are commented out to avoid failing in CI/CD
/*
func TestHealthCheck(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080")
	err := client.HealthCheck()
	if err != nil {
		t.Skipf("Health check failed (resolver not running): %v", err)
	}
}

func TestGetQueries(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080")
	queries, err := client.GetQueries()
	if err != nil {
		t.Skipf("Get queries failed (resolver not running): %v", err)
	}
	if queries == nil {
		t.Error("Expected queries to be non-nil")
	}
}

func TestGetFocusMode(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080")
	focusState, err := client.GetFocusMode()
	if err != nil {
		t.Skipf("Get focus mode failed (resolver not running): %v", err)
	}
	if focusState == nil {
		t.Error("Expected focus state to be non-nil")
	}
}

func TestSetFocusMode(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080")
	err := client.SetFocusMode(true, "5m")
	if err != nil {
		t.Skipf("Set focus mode failed (resolver not running): %v", err)
	}
}

func TestGetState(t *testing.T) {
	client := NewClient("http://127.0.0.1:8080")
	state, err := client.GetState()
	if err != nil {
		t.Skipf("Get state failed (resolver not running): %v", err)
	}
	if state == nil {
		t.Error("Expected state to be non-nil")
	}
}
*/

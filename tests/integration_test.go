// Package tests contains integration tests for Sinkzone.
package tests

import (
	"testing"

	"github.com/berbyte/sinkzone/config"
	"github.com/berbyte/sinkzone/database"
	"github.com/berbyte/sinkzone/pkg/filter"
)

func TestIntegration_FilterWithDatabase(t *testing.T) {
	// This is an example integration test
	// In a real implementation, you would use a test database

	// Create filter
	f := filter.New()

	// Test distracting sites
	distractingSites := []string{
		"facebook.com",
		"youtube.com",
		"twitter.com",
	}

	for _, site := range distractingSites {
		if !f.IsDistractingSite(site) {
			t.Errorf("Expected %s to be distracting", site)
		}
	}

	// Test essential sites
	essentialSites := []string{
		"google.com",
		"github.com",
		"stackoverflow.com",
	}

	for _, site := range essentialSites {
		if !f.IsEssentialSite(site) {
			t.Errorf("Expected %s to be essential", site)
		}
	}
}

func TestIntegration_Configuration(t *testing.T) {
	// Test configuration loading
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.UpstreamDNS == "" {
		t.Error("Expected upstream DNS to be set")
	}

	if cfg.PIN == "" {
		t.Error("Expected PIN to be set")
	}
}

func TestIntegration_DatabaseOperations(t *testing.T) {
	// Test database operations
	// In a real test, you would use a temporary database file
	dbPath := ":memory:" // Use in-memory database for testing

	db, err := database.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test adding domain rule
	err = db.AddDomainRule("example.com", "block")
	if err != nil {
		t.Errorf("Failed to add domain rule: %v", err)
	}

	// Test getting domain action
	action, err := db.GetDomainAction("example.com")
	if err != nil {
		t.Errorf("Failed to get domain action: %v", err)
	}

	if action != "block" {
		t.Errorf("Expected action 'block', got '%s'", action)
	}

	// Test getting statistics
	stats, err := db.GetStats()
	if err != nil {
		t.Errorf("Failed to get stats: %v", err)
	}

	if stats["current_mode"] == nil {
		t.Error("Expected current_mode in stats")
	}
}

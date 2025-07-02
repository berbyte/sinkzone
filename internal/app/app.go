// Package app provides the core application logic for Sinkzone.
package app

import (
	"fmt"
	"os"

	"github.com/berbyte/sinkzone/config"
	"github.com/berbyte/sinkzone/database"
	"github.com/berbyte/sinkzone/dns"
)

// App represents the main application instance.
type App struct {
	Config *config.Config
	DB     *database.DB
	Server *dns.Server
}

// New creates a new application instance.
func New() (*App, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Initialize database
	dbPath := config.GetConfigDir() + "/sinkzone.db"
	db, err := database.OpenDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create DNS server
	server := dns.NewServer(cfg.UpstreamDNS, db)

	return &App{
		Config: cfg,
		DB:     db,
		Server: server,
	}, nil
}

// Close cleans up application resources.
func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}

// Run starts the application.
func (a *App) Run() error {
	defer a.Close()

	// Start DNS server
	if err := a.Server.Start(); err != nil {
		return fmt.Errorf("failed to start DNS server: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	<-sigChan

	// Graceful shutdown
	if err := a.Server.Stop(); err != nil {
		return fmt.Errorf("failed to stop DNS server: %v", err)
	}

	return nil
}

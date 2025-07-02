package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/berbyte/sinkzone/config"
	"github.com/berbyte/sinkzone/database"

	"github.com/spf13/cobra"
)

var (
	cfg *config.Config
	db  *database.DB
)

var rootCmd = &cobra.Command{
	Use:   "sinkzone",
	Short: "Sinkzone - DNS filtering and focus tool",
	Long: `Sinkzone is a DNS filtering tool that helps you stay focused by blocking distracting websites.
It supports multiple modes: monitor, focus, lockdown, and off.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Load configuration
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	dbPath := filepath.Join(config.GetConfigDir(), "sinkzone.db")
	db, err = database.OpenDB(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
}

func getConfig() *config.Config {
	return cfg
}

func getDB() *database.DB {
	return db
}

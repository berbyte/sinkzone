package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

type DNSQuery struct {
	ID        int64
	Domain    string
	Timestamp time.Time
	Blocked   bool
	Count     int
}

type AllowlistEntry struct {
	ID       int64
	Domain   string
	AddedAt  time.Time
	IsActive bool
}

func New(dbPath string) (*Database, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Database{db: db}, nil
}

func createTables(db *sql.DB) error {
	// DNS queries table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS dns_queries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			blocked BOOLEAN NOT NULL,
			UNIQUE(domain, timestamp)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create dns_queries table: %w", err)
	}

	// Allowlist table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS allowlist (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT UNIQUE NOT NULL,
			added_at DATETIME NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT 1
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create allowlist table: %w", err)
	}

	// Create indexes for better performance
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_dns_queries_domain ON dns_queries(domain)`)
	if err != nil {
		return fmt.Errorf("failed to create dns_queries index: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_dns_queries_timestamp ON dns_queries(timestamp)`)
	if err != nil {
		return fmt.Errorf("failed to create timestamp index: %w", err)
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// RecordDNSQuery records a DNS query in the database
func (d *Database) RecordDNSQuery(domain string, blocked bool) error {
	_, err := d.db.Exec(`
		INSERT OR IGNORE INTO dns_queries (domain, timestamp, blocked)
		VALUES (?, ?, ?)
	`, domain, time.Now(), blocked)

	if err != nil {
		return fmt.Errorf("failed to record DNS query: %w", err)
	}

	return nil
}

// GetDNSStats returns DNS statistics grouped by domain
func (d *Database) GetDNSStats() (map[string]*DNSQuery, error) {
	rows, err := d.db.Query(`
		SELECT 
			domain,
			COUNT(*) as count,
			MAX(timestamp) as last_seen,
			MAX(CASE WHEN blocked THEN 1 ELSE 0 END) as blocked
		FROM dns_queries 
		WHERE timestamp > datetime('now', '-1 hour')
		GROUP BY domain
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query DNS stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]*DNSQuery)
	for rows.Next() {
		var domain string
		var count int
		var lastSeenStr string
		var blocked bool

		if err := rows.Scan(&domain, &count, &lastSeenStr, &blocked); err != nil {
			log.Printf("Failed to scan DNS stats row: %v", err)
			continue
		}

		// Parse the timestamp string
		lastSeen, err := time.Parse("2006-01-02 15:04:05.999999-07:00", lastSeenStr)
		if err != nil {
			// Try without timezone
			lastSeen, err = time.Parse("2006-01-02 15:04:05.999999", lastSeenStr)
			if err != nil {
				// Try basic format
				lastSeen, err = time.Parse("2006-01-02 15:04:05", lastSeenStr)
				if err != nil {
					log.Printf("Failed to parse timestamp '%s': %v", lastSeenStr, err)
					continue
				}
			}
		}

		stats[domain] = &DNSQuery{
			Domain:    domain,
			Count:     count,
			Timestamp: lastSeen,
			Blocked:   blocked,
		}
	}

	return stats, nil
}

// AddToAllowlist adds a domain to the allowlist
func (d *Database) AddToAllowlist(domain string) error {
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO allowlist (domain, added_at, is_active)
		VALUES (?, ?, 1)
	`, domain, time.Now())

	if err != nil {
		return fmt.Errorf("failed to add domain to allowlist: %w", err)
	}

	return nil
}

// RemoveFromAllowlist removes a domain from the allowlist
func (d *Database) RemoveFromAllowlist(domain string) error {
	_, err := d.db.Exec(`
		UPDATE allowlist SET is_active = 0 WHERE domain = ?
	`, domain)

	if err != nil {
		return fmt.Errorf("failed to remove domain from allowlist: %w", err)
	}

	return nil
}

// GetAllowlist returns all active domains in the allowlist
func (d *Database) GetAllowlist() ([]string, error) {
	rows, err := d.db.Query(`
		SELECT domain FROM allowlist 
		WHERE is_active = 1 
		ORDER BY domain
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query allowlist: %w", err)
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var domain string
		if err := rows.Scan(&domain); err != nil {
			log.Printf("Failed to scan allowlist row: %v", err)
			continue
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

// IsInAllowlist checks if a domain is in the allowlist
func (d *Database) IsInAllowlist(domain string) (bool, error) {
	var count int
	err := d.db.QueryRow(`
		SELECT COUNT(*) FROM allowlist 
		WHERE domain = ? AND is_active = 1
	`, domain).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check allowlist: %w", err)
	}

	return count > 0, nil
}

// CleanupOldQueries removes DNS queries older than 24 hours
func (d *Database) CleanupOldQueries() error {
	_, err := d.db.Exec(`
		DELETE FROM dns_queries 
		WHERE timestamp < datetime('now', '-1 hours')
	`)

	if err != nil {
		return fmt.Errorf("failed to cleanup old queries: %w", err)
	}

	return nil
}

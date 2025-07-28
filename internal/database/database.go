package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func init() {
	// Ensure the SQLite driver is registered
	_ = sql.Drivers()
}

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

	db, err := sql.Open("sqlite", dbPath)
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
	// DNS queries table - domain is unique, count tracks occurrences
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS dns_queries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT UNIQUE NOT NULL,
			timestamp DATETIME NOT NULL,
			blocked BOOLEAN NOT NULL,
			count INTEGER NOT NULL DEFAULT 1
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

	// Migrate existing data if needed
	if err := migrateDatabase(db); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

func migrateDatabase(db *sql.DB) error {
	// Check if count column exists
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('dns_queries') 
		WHERE name = 'count'
	`).Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to check for count column: %w", err)
	}

	// If count column doesn't exist, add it
	if count == 0 {
		// Add count column (SQLite doesn't support NOT NULL with DEFAULT in ALTER TABLE)
		_, err = db.Exec(`ALTER TABLE dns_queries ADD COLUMN count INTEGER`)
		if err != nil {
			return fmt.Errorf("failed to add count column: %w", err)
		}

		// Update existing records to have count = 1
		_, err = db.Exec(`UPDATE dns_queries SET count = 1 WHERE count IS NULL`)
		if err != nil {
			return fmt.Errorf("failed to update existing records: %w", err)
		}

		// Now make the column NOT NULL
		_, err = db.Exec(`CREATE TABLE dns_queries_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT UNIQUE NOT NULL,
			timestamp DATETIME NOT NULL,
			blocked BOOLEAN NOT NULL,
			count INTEGER NOT NULL DEFAULT 1
		)`)
		if err != nil {
			return fmt.Errorf("failed to create new table: %w", err)
		}

		// Copy data to new table
		_, err = db.Exec(`INSERT INTO dns_queries_new SELECT id, domain, timestamp, blocked, count FROM dns_queries`)
		if err != nil {
			return fmt.Errorf("failed to copy data: %w", err)
		}

		// Drop old table and rename new one
		_, err = db.Exec(`DROP TABLE dns_queries`)
		if err != nil {
			return fmt.Errorf("failed to drop old table: %w", err)
		}

		_, err = db.Exec(`ALTER TABLE dns_queries_new RENAME TO dns_queries`)
		if err != nil {
			return fmt.Errorf("failed to rename table: %w", err)
		}

		// Recreate indexes
		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_dns_queries_domain ON dns_queries(domain)`)
		if err != nil {
			return fmt.Errorf("failed to recreate domain index: %w", err)
		}

		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_dns_queries_timestamp ON dns_queries(timestamp)`)
		if err != nil {
			return fmt.Errorf("failed to recreate timestamp index: %w", err)
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// RecordDNSQuery records a DNS query in the database
func (d *Database) RecordDNSQuery(domain string, blocked bool) error {
	now := time.Now().Format(time.RFC3339Nano)
	_, err := d.db.Exec(`
		INSERT INTO dns_queries (domain, timestamp, blocked, count)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(domain) DO UPDATE SET
			timestamp = ?,
			blocked = ?,
			count = count + 1
	`, domain, now, blocked, now, blocked)

	if err != nil {
		return fmt.Errorf("failed to record DNS query: %w", err)
	}

	return nil
}

// GetDNSStats returns DNS statistics as a slice
func (d *Database) GetDNSStats() ([]DNSQuery, error) {
	rows, err := d.db.Query(`
		SELECT 
			domain,
			count,
			timestamp,
			blocked
		FROM dns_queries 
		WHERE timestamp > datetime('now', '-1 hour')
		ORDER BY timestamp DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query DNS stats: %w", err)
	}
	defer rows.Close()

	var queries []DNSQuery
	for rows.Next() {
		var domain string
		var count int
		var timestampStr string
		var blocked bool

		if err := rows.Scan(&domain, &count, &timestampStr, &blocked); err != nil {
			log.Printf("Failed to scan DNS stats row: %v", err)
			continue
		}

		// Parse the timestamp string
		timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
		if err != nil {
			// Try without timezone
			timestamp, err = time.Parse("2006-01-02T15:04:05.999999999", timestampStr)
			if err != nil {
				// Try basic format
				timestamp, err = time.Parse("2006-01-02 15:04:05", timestampStr)
				if err != nil {
					log.Printf("Failed to parse timestamp '%s': %v", timestampStr, err)
					continue
				}
			}
		}

		queries = append(queries, DNSQuery{
			Domain:    domain,
			Count:     count,
			Timestamp: timestamp,
			Blocked:   blocked,
		})
	}

	return queries, nil
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
	// Try a more explicit query that should work with modernc.org/sqlite
	rows, err := d.db.Query(`
		SELECT domain FROM allowlist 
		WHERE is_active IS NOT NULL AND is_active != 0
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
		WHERE domain = ? AND is_active IS NOT NULL AND is_active != 0
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

// GetNewDNSRecords returns DNS records seen since the given time
func (d *Database) GetNewDNSRecords(since time.Time) ([]DNSQuery, error) {
	rows, err := d.db.Query(`
		SELECT 
			domain,
			count,
			timestamp,
			blocked
		FROM dns_queries 
		WHERE timestamp > ?
		ORDER BY timestamp DESC
	`, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query new DNS records: %w", err)
	}
	defer rows.Close()

	var queries []DNSQuery
	for rows.Next() {
		var domain string
		var count int
		var timestampStr string
		var blocked bool

		if err := rows.Scan(&domain, &count, &timestampStr, &blocked); err != nil {
			log.Printf("Failed to scan DNS record row: %v", err)
			continue
		}

		// Parse the timestamp string
		timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
		if err != nil {
			// Try without timezone
			timestamp, err = time.Parse("2006-01-02T15:04:05.999999999", timestampStr)
			if err != nil {
				// Try basic format
				timestamp, err = time.Parse("2006-01-02 15:04:05", timestampStr)
				if err != nil {
					log.Printf("Failed to parse timestamp '%s': %v", timestampStr, err)
					continue
				}
			}
		}

		queries = append(queries, DNSQuery{
			Domain:    domain,
			Count:     count,
			Timestamp: timestamp,
			Blocked:   blocked,
		})
	}

	return queries, nil
}

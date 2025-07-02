package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

type DomainRule struct {
	ID        int64
	Domain    string
	Action    string // "allow" or "block"
	CreatedAt time.Time
}

type DNSQuery struct {
	ID        int64
	Domain    string
	ClientIP  string
	QueryType string
	Timestamp time.Time
	Blocked   bool
}

type AppState struct {
	Mode      string // "focus", "lockdown", "monitor", "off"
	StartedAt time.Time
	UpdatedAt time.Time
}

func OpenDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	if err := initTables(db); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %v", err)
	}

	return &DB{db}, nil
}

func initTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS domain_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT NOT NULL UNIQUE,
			action TEXT NOT NULL CHECK(action IN ('allow', 'block')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS dns_queries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT NOT NULL,
			client_ip TEXT NOT NULL,
			query_type TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			blocked BOOLEAN DEFAULT FALSE
		)`,
		`CREATE TABLE IF NOT EXISTS app_state (
			id INTEGER PRIMARY KEY CHECK(id = 1),
			mode TEXT NOT NULL DEFAULT 'off',
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_domain_rules_domain ON domain_rules(domain)`,
		`CREATE INDEX IF NOT EXISTS idx_dns_queries_domain ON dns_queries(domain)`,
		`CREATE INDEX IF NOT EXISTS idx_dns_queries_timestamp ON dns_queries(timestamp)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %v", query, err)
		}
	}

	// Initialize app_state if empty
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM app_state").Scan(&count); err != nil {
		return fmt.Errorf("failed to check app_state: %v", err)
	}

	if count == 0 {
		if _, err := db.Exec("INSERT INTO app_state (id, mode) VALUES (1, 'off')"); err != nil {
			return fmt.Errorf("failed to initialize app_state: %v", err)
		}
	}

	return nil
}

func (db *DB) AddDomainRule(domain, action string) error {
	_, err := db.Exec("INSERT OR REPLACE INTO domain_rules (domain, action) VALUES (?, ?)", domain, action)
	return err
}

func (db *DB) RemoveDomainRule(domain string) error {
	_, err := db.Exec("DELETE FROM domain_rules WHERE domain = ?", domain)
	return err
}

func (db *DB) GetDomainRules() ([]DomainRule, error) {
	rows, err := db.Query("SELECT id, domain, action, created_at FROM domain_rules ORDER BY domain")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []DomainRule
	for rows.Next() {
		var rule DomainRule
		if err := rows.Scan(&rule.ID, &rule.Domain, &rule.Action, &rule.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (db *DB) GetDomainAction(domain string) (string, error) {
	var action string
	err := db.QueryRow("SELECT action FROM domain_rules WHERE domain = ?", domain).Scan(&action)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return action, err
}

func (db *DB) LogDNSQuery(domain, clientIP, queryType string, blocked bool) error {
	_, err := db.Exec("INSERT INTO dns_queries (domain, client_ip, query_type, blocked) VALUES (?, ?, ?, ?)",
		domain, clientIP, queryType, blocked)
	return err
}

func (db *DB) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total queries
	var totalQueries int
	if err := db.QueryRow("SELECT COUNT(*) FROM dns_queries").Scan(&totalQueries); err != nil {
		return nil, err
	}
	stats["total_queries"] = totalQueries

	// Blocked queries
	var blockedQueries int
	if err := db.QueryRow("SELECT COUNT(*) FROM dns_queries WHERE blocked = 1").Scan(&blockedQueries); err != nil {
		return nil, err
	}
	stats["blocked_queries"] = blockedQueries

	// Unique clients
	var uniqueClients int
	if err := db.QueryRow("SELECT COUNT(DISTINCT client_ip) FROM dns_queries").Scan(&uniqueClients); err != nil {
		return nil, err
	}
	stats["unique_clients"] = uniqueClients

	// Unique domains
	var uniqueDomains int
	if err := db.QueryRow("SELECT COUNT(DISTINCT domain) FROM dns_queries").Scan(&uniqueDomains); err != nil {
		return nil, err
	}
	stats["unique_domains"] = uniqueDomains

	// Current mode
	var mode string
	if err := db.QueryRow("SELECT mode FROM app_state WHERE id = 1").Scan(&mode); err != nil {
		return nil, err
	}
	stats["current_mode"] = mode

	return stats, nil
}

func (db *DB) SetMode(mode string) error {
	_, err := db.Exec("UPDATE app_state SET mode = ?, updated_at = CURRENT_TIMESTAMP WHERE id = 1", mode)
	return err
}

func (db *DB) GetMode() (string, error) {
	var mode string
	err := db.QueryRow("SELECT mode FROM app_state WHERE id = 1").Scan(&mode)
	return mode, err
}

func (db *DB) Reset() error {
	queries := []string{
		"DELETE FROM domain_rules",
		"DELETE FROM dns_queries",
		"UPDATE app_state SET mode = 'off', updated_at = CURRENT_TIMESTAMP WHERE id = 1",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute reset query: %v", err)
		}
	}
	return nil
}

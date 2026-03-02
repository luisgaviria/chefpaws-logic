package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Connect opens a connection pool to the MySQL database.
// DATABASE_URL may be a raw DSN (user:pass@tcp(host:port)/db) or include a scheme.
// parseTime=true is injected automatically if not already present.
func Connect(dsn string) (*sql.DB, error) {
	dsn = ensureParseTime(dsn)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open failed: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	return db, nil
}

// CreateLeadsTable creates the leads table if it does not already exist.
func CreateLeadsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS leads (
			id             INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			owner_name     TEXT NOT NULL,
			email          TEXT NOT NULL,
			phone          TEXT,
			zip            TEXT,
			dog_name       TEXT,
			daily_calories FLOAT,
			portion_grams  FLOAT,
			special_reqs   TEXT,
			source         TEXT,
			created_at     DATETIME DEFAULT NOW()
		)
	`)
	return err
}

// ensureParseTime appends parseTime=true to the DSN if not already set.
func ensureParseTime(dsn string) string {
	if strings.Contains(dsn, "parseTime") {
		return dsn
	}
	if strings.Contains(dsn, "?") {
		return dsn + "&parseTime=true"
	}
	return dsn + "?parseTime=true"
}

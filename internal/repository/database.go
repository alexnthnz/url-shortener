package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings optimized for high load
	db.SetMaxOpenConns(100)                 // Increase from 25 to handle more concurrent requests
	db.SetMaxIdleConns(25)                  // Increase from 5 to reduce connection establishment overhead
	db.SetConnMaxLifetime(time.Hour)        // Prevent connection leaks and ensure fresh connections
	db.SetConnMaxIdleTime(30 * time.Minute) // Close idle connections after 30 minutes

	return db, nil
}

// RunMigrations executes database migrations
func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			short_code VARCHAR(10) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
			custom_alias BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at)`,
		`CREATE TABLE IF NOT EXISTS analytics (
			id SERIAL PRIMARY KEY,
			short_code VARCHAR(10) NOT NULL,
			clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ip_address INET,
			user_agent TEXT,
			FOREIGN KEY (short_code) REFERENCES urls(short_code) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_short_code ON analytics(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_clicked_at ON analytics(clicked_at)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_clicked_at_short_code ON analytics(clicked_at, short_code)`,
		// Create atomic sequence for URL ID generation to prevent race conditions
		`CREATE SEQUENCE IF NOT EXISTS url_id_sequence START WITH 1 INCREMENT BY 1`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}

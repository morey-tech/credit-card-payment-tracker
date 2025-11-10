package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// DB holds the database connection
var DB *sql.DB

// InitDB initializes the SQLite database and creates tables
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(1) // SQLite works best with single connection
	DB.SetMaxIdleConns(1)

	// Test the connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// createTables creates the necessary database tables
func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS credit_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		last_four TEXT NOT NULL,
		statement_day INTEGER NOT NULL,
		due_day INTEGER NOT NULL,
		credit_limit REAL,
		discord_webhook_url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS statements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		card_id INTEGER NOT NULL,
		statement_date TEXT NOT NULL,
		due_date TEXT NOT NULL,
		amount REAL NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		notified_statement BOOLEAN DEFAULT 0,
		notified_payment BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (card_id) REFERENCES credit_cards(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_statements_card_id ON statements(card_id);
	CREATE INDEX IF NOT EXISTS idx_statements_status ON statements(status);
	CREATE INDEX IF NOT EXISTS idx_statements_due_date ON statements(due_date);
	`

	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

// DB holds the database connection
var DB *sql.DB

// InitDB initializes the SQLite database and creates tables
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
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

	// Load sample data if environment variable is set
	if os.Getenv("LOAD_SAMPLE_DATA") == "true" {
		if err := LoadSampleData(DB); err != nil {
			return fmt.Errorf("failed to load sample data: %w", err)
		}
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
		days_until_due INTEGER NOT NULL,
		credit_limit REAL,
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
		reviewed_at DATETIME,
		scheduled_payment_date TEXT,
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

	// Run migrations to add new columns to existing databases
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// runMigrations applies database schema migrations
func runMigrations() error {
	// Check if reviewed_at column exists
	var reviewedAtExists bool
	row := DB.QueryRow(`
		SELECT COUNT(*)
		FROM pragma_table_info('statements')
		WHERE name='reviewed_at'
	`)
	var count int
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("failed to check for reviewed_at column: %w", err)
	}
	reviewedAtExists = count > 0

	// Check if scheduled_payment_date column exists
	var scheduledPaymentDateExists bool
	row = DB.QueryRow(`
		SELECT COUNT(*)
		FROM pragma_table_info('statements')
		WHERE name='scheduled_payment_date'
	`)
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("failed to check for scheduled_payment_date column: %w", err)
	}
	scheduledPaymentDateExists = count > 0

	// Add reviewed_at column if it doesn't exist
	if !reviewedAtExists {
		log.Println("Running migration: adding reviewed_at column to statements table")
		_, err := DB.Exec("ALTER TABLE statements ADD COLUMN reviewed_at DATETIME")
		if err != nil {
			return fmt.Errorf("failed to add reviewed_at column: %w", err)
		}
		log.Println("Migration completed: reviewed_at column added")
	}

	// Add scheduled_payment_date column if it doesn't exist
	if !scheduledPaymentDateExists {
		log.Println("Running migration: adding scheduled_payment_date column to statements table")
		_, err := DB.Exec("ALTER TABLE statements ADD COLUMN scheduled_payment_date TEXT")
		if err != nil {
			return fmt.Errorf("failed to add scheduled_payment_date column: %w", err)
		}
		log.Println("Migration completed: scheduled_payment_date column added")
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

package database

import (
	"database/sql"
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	// Create a temporary database file
	tmpDB := "./test_init.db"
	defer os.Remove(tmpDB)

	// Test database initialization
	err := InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify DB is not nil
	if DB == nil {
		t.Fatal("DB should not be nil after initialization")
	}

	// Test connection with ping
	err = DB.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Close the database
	err = Close()
	if err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}
}

func TestCreateTables(t *testing.T) {
	// Create a temporary database file
	tmpDB := "./test_tables.db"
	defer os.Remove(tmpDB)

	// Initialize database
	err := InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer Close()

	// Verify credit_cards table exists
	var tableName string
	err = DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='credit_cards'").Scan(&tableName)
	if err != nil {
		t.Fatalf("credit_cards table not found: %v", err)
	}
	if tableName != "credit_cards" {
		t.Errorf("Expected table name 'credit_cards', got '%s'", tableName)
	}

	// Verify statements table exists
	err = DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='statements'").Scan(&tableName)
	if err != nil {
		t.Fatalf("statements table not found: %v", err)
	}
	if tableName != "statements" {
		t.Errorf("Expected table name 'statements', got '%s'", tableName)
	}

	// Verify indexes exist
	indexes := []string{
		"idx_statements_card_id",
		"idx_statements_status",
		"idx_statements_due_date",
	}

	for _, indexName := range indexes {
		var name string
		err = DB.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", indexName).Scan(&name)
		if err != nil {
			t.Errorf("Index %s not found: %v", indexName, err)
		}
	}
}

func TestDatabaseSchema(t *testing.T) {
	// Create a temporary database file
	tmpDB := "./test_schema.db"
	defer os.Remove(tmpDB)

	// Initialize database
	err := InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer Close()

	// Test credit_cards table schema
	rows, err := DB.Query("PRAGMA table_info(credit_cards)")
	if err != nil {
		t.Fatalf("Failed to get credit_cards schema: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]bool{
		"id":                  false,
		"name":                false,
		"last_four":           false,
		"statement_day":       false,
		"due_day":             false,
		"credit_limit":        false,
		"discord_webhook_url": false,
		"created_at":          false,
		"updated_at":          false,
	}

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString

		err = rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		if err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}

		if _, exists := expectedColumns[name]; exists {
			expectedColumns[name] = true
		}
	}

	// Verify all expected columns were found
	for col, found := range expectedColumns {
		if !found {
			t.Errorf("Expected column '%s' not found in credit_cards table", col)
		}
	}
}

func TestDatabaseReopening(t *testing.T) {
	// Create a temporary database file
	tmpDB := "./test_reopen.db"
	defer os.Remove(tmpDB)

	// Initialize database first time
	err := InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize database first time: %v", err)
	}

	// Close the database
	err = Close()
	if err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}

	// Reopen the database
	err = InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer Close()

	// Verify tables still exist
	var tableName string
	err = DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='credit_cards'").Scan(&tableName)
	if err != nil {
		t.Fatalf("credit_cards table not found after reopening: %v", err)
	}
}

func TestLoadSampleDataEnvironmentVariable(t *testing.T) {
	// Create a temporary database file
	tmpDB := "./test_sample_data_env.db"
	defer os.Remove(tmpDB)

	// Set environment variable
	os.Setenv("LOAD_SAMPLE_DATA", "true")
	defer os.Unsetenv("LOAD_SAMPLE_DATA")

	// Initialize database
	err := InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer Close()

	// Verify sample data was loaded
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM credit_cards").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count credit cards: %v", err)
	}

	if count < 2 {
		t.Errorf("Expected at least 2 credit cards, got %d", count)
	}
}

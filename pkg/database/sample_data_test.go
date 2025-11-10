package database

import (
	"os"
	"testing"
)

func TestLoadSampleData(t *testing.T) {
	// Create a temporary database file
	tmpDB := "./test_sample_data.db"
	defer os.Remove(tmpDB)

	// Initialize database
	err := InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer Close()

	// Load sample data
	err = LoadSampleData(DB)
	if err != nil {
		t.Fatalf("Failed to load sample data: %v", err)
	}

	// Verify credit cards were inserted
	var cardCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM credit_cards WHERE name IN ('TD Aeroplan Visa', 'Amex Cobalt')").Scan(&cardCount)
	if err != nil {
		t.Fatalf("Failed to count credit cards: %v", err)
	}

	if cardCount != 2 {
		t.Errorf("Expected 2 sample credit cards, got %d", cardCount)
	}

	// Verify TD Aeroplan Visa card details
	var name, lastFour string
	var statementDay, dueDay int
	var creditLimit float64

	err = DB.QueryRow(`
		SELECT name, last_four, statement_day, due_day, credit_limit
		FROM credit_cards
		WHERE name = 'TD Aeroplan Visa'
	`).Scan(&name, &lastFour, &statementDay, &dueDay, &creditLimit)

	if err != nil {
		t.Fatalf("Failed to query TD Aeroplan Visa: %v", err)
	}

	if lastFour != "9876" {
		t.Errorf("Expected last_four '9876', got '%s'", lastFour)
	}
	if statementDay != 15 {
		t.Errorf("Expected statement_day 15, got %d", statementDay)
	}
	if dueDay != 10 {
		t.Errorf("Expected due_day 10, got %d", dueDay)
	}
	if creditLimit != 5000.00 {
		t.Errorf("Expected credit_limit 5000.00, got %.2f", creditLimit)
	}

	// Verify Amex Cobalt card details
	err = DB.QueryRow(`
		SELECT name, last_four, statement_day, due_day, credit_limit
		FROM credit_cards
		WHERE name = 'Amex Cobalt'
	`).Scan(&name, &lastFour, &statementDay, &dueDay, &creditLimit)

	if err != nil {
		t.Fatalf("Failed to query Amex Cobalt: %v", err)
	}

	if lastFour != "1234" {
		t.Errorf("Expected last_four '1234', got '%s'", lastFour)
	}
	if statementDay != 28 {
		t.Errorf("Expected statement_day 28, got %d", statementDay)
	}
	if dueDay != 23 {
		t.Errorf("Expected due_day 23, got %d", dueDay)
	}
	if creditLimit != 10000.00 {
		t.Errorf("Expected credit_limit 10000.00, got %.2f", creditLimit)
	}
}

func TestLoadSampleDataStatements(t *testing.T) {
	// Create a temporary database file
	tmpDB := "./test_sample_statements.db"
	defer os.Remove(tmpDB)

	// Initialize database
	err := InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer Close()

	// Load sample data
	err = LoadSampleData(DB)
	if err != nil {
		t.Fatalf("Failed to load sample data: %v", err)
	}

	// Verify statements were inserted
	var stmtCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM statements").Scan(&stmtCount)
	if err != nil {
		t.Fatalf("Failed to count statements: %v", err)
	}

	if stmtCount != 4 {
		t.Errorf("Expected 4 sample statements, got %d", stmtCount)
	}

	// Verify at least one pending statement exists
	var pendingCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM statements WHERE status = 'pending'").Scan(&pendingCount)
	if err != nil {
		t.Fatalf("Failed to count pending statements: %v", err)
	}

	if pendingCount < 1 {
		t.Errorf("Expected at least 1 pending statement, got %d", pendingCount)
	}

	// Verify at least one paid statement exists
	var paidCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM statements WHERE status = 'paid'").Scan(&paidCount)
	if err != nil {
		t.Fatalf("Failed to count paid statements: %v", err)
	}

	if paidCount < 1 {
		t.Errorf("Expected at least 1 paid statement, got %d", paidCount)
	}
}

func TestLoadSampleDataIdempotency(t *testing.T) {
	// Create a temporary database file
	tmpDB := "./test_idempotency.db"
	defer os.Remove(tmpDB)

	// Initialize database
	err := InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer Close()

	// Load sample data first time
	err = LoadSampleData(DB)
	if err != nil {
		t.Fatalf("Failed to load sample data first time: %v", err)
	}

	// Count cards after first load
	var firstCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM credit_cards").Scan(&firstCount)
	if err != nil {
		t.Fatalf("Failed to count credit cards: %v", err)
	}

	// Load sample data second time
	err = LoadSampleData(DB)
	if err != nil {
		t.Fatalf("Failed to load sample data second time: %v", err)
	}

	// Count cards after second load
	var secondCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM credit_cards").Scan(&secondCount)
	if err != nil {
		t.Fatalf("Failed to count credit cards: %v", err)
	}

	// Verify count didn't increase (idempotent)
	if firstCount != secondCount {
		t.Errorf("Sample data loading is not idempotent: first count %d, second count %d", firstCount, secondCount)
	}
}

func TestLoadSampleDataForeignKeys(t *testing.T) {
	// Create a temporary database file
	tmpDB := "./test_foreign_keys.db"
	defer os.Remove(tmpDB)

	// Initialize database
	err := InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer Close()

	// Load sample data
	err = LoadSampleData(DB)
	if err != nil {
		t.Fatalf("Failed to load sample data: %v", err)
	}

	// Verify all statements have valid card_id references
	var invalidCount int
	err = DB.QueryRow(`
		SELECT COUNT(*)
		FROM statements s
		LEFT JOIN credit_cards c ON s.card_id = c.id
		WHERE c.id IS NULL
	`).Scan(&invalidCount)

	if err != nil {
		t.Fatalf("Failed to check foreign keys: %v", err)
	}

	if invalidCount > 0 {
		t.Errorf("Found %d statements with invalid card_id references", invalidCount)
	}
}

func TestLoadSampleDataAmounts(t *testing.T) {
	// Create a temporary database file
	tmpDB := "./test_amounts.db"
	defer os.Remove(tmpDB)

	// Initialize database
	err := InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer Close()

	// Load sample data
	err = LoadSampleData(DB)
	if err != nil {
		t.Fatalf("Failed to load sample data: %v", err)
	}

	// Verify all statement amounts are positive
	var negativeCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM statements WHERE amount <= 0").Scan(&negativeCount)
	if err != nil {
		t.Fatalf("Failed to check statement amounts: %v", err)
	}

	if negativeCount > 0 {
		t.Errorf("Found %d statements with non-positive amounts", negativeCount)
	}

	// Verify statement amounts are realistic (between $1 and $100,000)
	var unrealisticCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM statements WHERE amount < 1 OR amount > 100000").Scan(&unrealisticCount)
	if err != nil {
		t.Fatalf("Failed to check realistic amounts: %v", err)
	}

	if unrealisticCount > 0 {
		t.Errorf("Found %d statements with unrealistic amounts", unrealisticCount)
	}
}

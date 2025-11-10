package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// LoadSampleData inserts sample credit cards and statements into the database
// This is intended for development and testing purposes only
func LoadSampleData(db *sql.DB) error {
	log.Println("Loading sample data into database...")

	// Check if sample data already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM credit_cards WHERE name IN ('TD Aeroplan Visa', 'Amex Cobalt')").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for existing sample data: %w", err)
	}

	if count > 0 {
		log.Printf("Sample data already exists (%d cards found), skipping load", count)
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Sample Card 1: TD Aeroplan Visa
	result, err := tx.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, due_day, credit_limit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "TD Aeroplan Visa", "9876", 15, 10, 5000.00, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert TD Aeroplan Visa: %w", err)
	}
	tdCardID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get TD card ID: %w", err)
	}

	// Sample Card 2: Amex Cobalt
	result, err = tx.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, due_day, credit_limit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "Amex Cobalt", "1234", 28, 23, 10000.00, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert Amex Cobalt: %w", err)
	}
	amexCardID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get Amex card ID: %w", err)
	}

	// Sample Statements for TD Aeroplan Visa
	now := time.Now()

	// Past statement (paid)
	pastStatementDate := time.Date(now.Year(), now.Month()-1, 15, 0, 0, 0, 0, time.UTC)
	pastDueDate := time.Date(now.Year(), now.Month(), 10, 0, 0, 0, 0, time.UTC)
	_, err = tx.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status, notified_statement, notified_payment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tdCardID, pastStatementDate.Format("2006-01-02"), pastDueDate.Format("2006-01-02"), 1250.75, "paid", true, true, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert past TD statement: %w", err)
	}

	// Current statement (pending)
	currentStatementDate := time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, time.UTC)
	currentDueDate := time.Date(now.Year(), now.Month()+1, 10, 0, 0, 0, 0, time.UTC)
	_, err = tx.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status, notified_statement, notified_payment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tdCardID, currentStatementDate.Format("2006-01-02"), currentDueDate.Format("2006-01-02"), 892.50, "pending", false, false, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert current TD statement: %w", err)
	}

	// Sample Statements for Amex Cobalt
	// Past statement (paid)
	amexPastStatementDate := time.Date(now.Year(), now.Month()-1, 28, 0, 0, 0, 0, time.UTC)
	amexPastDueDate := time.Date(now.Year(), now.Month(), 23, 0, 0, 0, 0, time.UTC)
	_, err = tx.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status, notified_statement, notified_payment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, amexCardID, amexPastStatementDate.Format("2006-01-02"), amexPastDueDate.Format("2006-01-02"), 2150.00, "paid", true, true, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert past Amex statement: %w", err)
	}

	// Current statement (pending)
	amexCurrentStatementDate := time.Date(now.Year(), now.Month(), 28, 0, 0, 0, 0, time.UTC)
	amexCurrentDueDate := time.Date(now.Year(), now.Month()+1, 23, 0, 0, 0, 0, time.UTC)
	_, err = tx.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status, notified_statement, notified_payment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, amexCardID, amexCurrentStatementDate.Format("2006-01-02"), amexCurrentDueDate.Format("2006-01-02"), 3421.89, "pending", false, false, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert current Amex statement: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Sample data loaded successfully: 2 cards, 4 statements")
	return nil
}

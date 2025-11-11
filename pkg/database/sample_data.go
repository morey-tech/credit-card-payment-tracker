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
	err := db.QueryRow("SELECT COUNT(*) FROM credit_cards WHERE name IN ('TD Aeroplan Visa', 'Amex Cobalt', 'Chase Sapphire Reserve', 'Capital One Quicksilver', 'Discover It', 'Citi Double Cash')").Scan(&count)
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
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due, credit_limit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "TD Aeroplan Visa", "9876", 15, 25, 5000.00, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert TD Aeroplan Visa: %w", err)
	}
	tdCardID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get TD card ID: %w", err)
	}

	// Sample Card 2: Amex Cobalt
	result, err = tx.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due, credit_limit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "Amex Cobalt", "1234", 28, 25, 10000.00, time.Now(), time.Now())
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

	// Sample Card 3: Chase Sapphire Reserve (with no credit limit set - testing optional field)
	result, err = tx.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "Chase Sapphire Reserve", "5678", 1, 21, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert Chase Sapphire Reserve: %w", err)
	}
	chaseCardID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get Chase card ID: %w", err)
	}

	// Chase: Overdue statement (testing overdue scenario)
	overdueStatementDate := time.Date(now.Year(), now.Month()-2, 1, 0, 0, 0, 0, time.UTC)
	overdueDueDate := time.Date(now.Year(), now.Month()-1, 22, 0, 0, 0, 0, time.UTC)
	_, err = tx.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status, notified_statement, notified_payment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, chaseCardID, overdueStatementDate.Format("2006-01-02"), overdueDueDate.Format("2006-01-02"), 567.25, "pending", true, true, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert overdue Chase statement: %w", err)
	}

	// Sample Card 4: Capital One Quicksilver (different due date offset - short cycle)
	result, err = tx.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due, credit_limit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "Capital One Quicksilver", "4321", 5, 15, 3000.00, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert Capital One Quicksilver: %w", err)
	}
	capitalOneCardID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get Capital One card ID: %w", err)
	}

	// Capital One: Multiple statements with various statuses
	// Statement 1: Old paid statement
	oldPaidStatementDate := time.Date(now.Year(), now.Month()-3, 5, 0, 0, 0, 0, time.UTC)
	oldPaidDueDate := time.Date(now.Year(), now.Month()-3, 20, 0, 0, 0, 0, time.UTC)
	_, err = tx.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status, notified_statement, notified_payment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, capitalOneCardID, oldPaidStatementDate.Format("2006-01-02"), oldPaidDueDate.Format("2006-01-02"), 125.50, "paid", true, true, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert old paid Capital One statement: %w", err)
	}

	// Statement 2: Recent paid statement
	recentPaidStatementDate := time.Date(now.Year(), now.Month()-2, 5, 0, 0, 0, 0, time.UTC)
	recentPaidDueDate := time.Date(now.Year(), now.Month()-2, 20, 0, 0, 0, 0, time.UTC)
	_, err = tx.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status, notified_statement, notified_payment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, capitalOneCardID, recentPaidStatementDate.Format("2006-01-02"), recentPaidDueDate.Format("2006-01-02"), 435.99, "paid", true, true, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert recent paid Capital One statement: %w", err)
	}

	// Statement 3: Current pending with small amount
	currentCapitalOneStatementDate := time.Date(now.Year(), now.Month()-1, 5, 0, 0, 0, 0, time.UTC)
	currentCapitalOneDueDate := time.Date(now.Year(), now.Month()-1, 20, 0, 0, 0, 0, time.UTC)
	_, err = tx.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status, notified_statement, notified_payment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, capitalOneCardID, currentCapitalOneStatementDate.Format("2006-01-02"), currentCapitalOneDueDate.Format("2006-01-02"), 15.00, "pending", true, false, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert current Capital One statement: %w", err)
	}

	// Sample Card 5: Discover It (long due date cycle)
	result, err = tx.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due, credit_limit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "Discover It", "8888", 20, 30, 7500.00, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert Discover It: %w", err)
	}
	discoverCardID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get Discover card ID: %w", err)
	}

	// Discover: Large amount pending statement
	discoverStatementDate := time.Date(now.Year(), now.Month(), 20, 0, 0, 0, 0, time.UTC)
	discoverDueDate := time.Date(now.Year(), now.Month()+1, 20, 0, 0, 0, 0, time.UTC)
	_, err = tx.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status, notified_statement, notified_payment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, discoverCardID, discoverStatementDate.Format("2006-01-02"), discoverDueDate.Format("2006-01-02"), 4567.89, "pending", false, false, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert Discover statement: %w", err)
	}

	// Sample Card 6: Citi Double Cash (no statements - testing card without any statements)
	_, err = tx.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due, credit_limit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "Citi Double Cash", "2468", 10, 25, 8000.00, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to insert Citi Double Cash: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Sample data loaded successfully: 6 cards, 9 statements")
	return nil
}

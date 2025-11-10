package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/morey-tech/credit-card-payment-tracker/pkg/database"
	"github.com/morey-tech/credit-card-payment-tracker/pkg/models"
)

// HealthCheck returns the health status of the API
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
		"message": "Credit Card Payment Tracker API is running",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetCards returns all credit cards
func GetCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `
		SELECT id, name, last_four, statement_day, due_day,
		       credit_limit, discord_webhook_url, created_at, updated_at
		FROM credit_cards
		ORDER BY name
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		log.Printf("Error querying credit cards: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	cards := []models.CreditCard{}
	for rows.Next() {
		var card models.CreditCard
		var creditLimit sql.NullFloat64
		var webhookURL sql.NullString

		err := rows.Scan(
			&card.ID,
			&card.Name,
			&card.LastFour,
			&card.StatementDay,
			&card.DueDay,
			&creditLimit,
			&webhookURL,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning credit card: %v", err)
			continue
		}

		// Handle NULL values
		if creditLimit.Valid {
			card.CreditLimit = creditLimit.Float64
		}
		if webhookURL.Valid {
			card.DiscordWebhookURL = webhookURL.String
		}

		cards = append(cards, card)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cards)
}

// GetStatements returns all statements
func GetStatements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `
		SELECT id, card_id, statement_date, due_date, amount,
		       status, notified_statement, notified_payment,
		       created_at, updated_at
		FROM statements
		ORDER BY due_date DESC
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		log.Printf("Error querying statements: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	statements := []models.Statement{}
	for rows.Next() {
		var stmt models.Statement
		err := rows.Scan(
			&stmt.ID,
			&stmt.CardID,
			&stmt.StatementDate,
			&stmt.DueDate,
			&stmt.Amount,
			&stmt.Status,
			&stmt.NotifiedStatement,
			&stmt.NotifiedPayment,
			&stmt.CreatedAt,
			&stmt.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning statement: %v", err)
			continue
		}
		statements = append(statements, stmt)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(statements)
}

// GetCardByID returns a single credit card by ID
func GetCardByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path (e.g., /api/v1/cards/1)
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	idStr := pathParts[4]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, name, last_four, statement_day, due_day,
		       credit_limit, discord_webhook_url, created_at, updated_at
		FROM credit_cards
		WHERE id = ?
	`

	var card models.CreditCard
	var creditLimit sql.NullFloat64
	var webhookURL sql.NullString

	err = database.DB.QueryRow(query, id).Scan(
		&card.ID,
		&card.Name,
		&card.LastFour,
		&card.StatementDay,
		&card.DueDay,
		&creditLimit,
		&webhookURL,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error querying credit card by ID %d: %v", id, err)
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	// Handle NULL values
	if creditLimit.Valid {
		card.CreditLimit = creditLimit.Float64
	}
	if webhookURL.Valid {
		card.DiscordWebhookURL = webhookURL.String
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(card)
}

// CreateStatement creates a new statement
func CreateStatement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var stmt models.Statement
	if err := json.NewDecoder(r.Body).Decode(&stmt); err != nil {
		log.Printf("Error decoding statement: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if stmt.CardID == 0 {
		http.Error(w, "card_id is required", http.StatusBadRequest)
		return
	}
	if stmt.StatementDate == "" {
		http.Error(w, "statement_date is required", http.StatusBadRequest)
		return
	}
	if stmt.DueDate == "" {
		http.Error(w, "due_date is required", http.StatusBadRequest)
		return
	}
	if stmt.Amount <= 0 {
		http.Error(w, "amount must be greater than 0", http.StatusBadRequest)
		return
	}

	// Set defaults
	if stmt.Status == "" {
		stmt.Status = "pending"
	}
	stmt.CreatedAt = time.Now()
	stmt.UpdatedAt = time.Now()

	query := `
		INSERT INTO statements (card_id, statement_date, due_date, amount, status, notified_statement, notified_payment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := database.DB.Exec(query,
		stmt.CardID,
		stmt.StatementDate,
		stmt.DueDate,
		stmt.Amount,
		stmt.Status,
		stmt.NotifiedStatement,
		stmt.NotifiedPayment,
		stmt.CreatedAt,
		stmt.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error creating statement: %v", err)
		http.Error(w, "Failed to create statement", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		http.Error(w, "Failed to create statement", http.StatusInternalServerError)
		return
	}

	stmt.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(stmt)
}

// UpdateStatement updates a statement's status
func UpdateStatement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path (e.g., /api/v1/statements/1)
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	idStr := pathParts[4]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid statement ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("Error decoding updates: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For now, only support updating status
	status, ok := updates["status"].(string)
	if !ok {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE statements
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = database.DB.Exec(query, status, time.Now(), id)
	if err != nil {
		log.Printf("Error updating statement %d: %v", id, err)
		http.Error(w, "Failed to update statement", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/morey-tech/credit-card-payment-tracker/pkg/config"
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
		SELECT id, name, last_four, statement_day, days_until_due,
		       credit_limit, created_at, updated_at
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

		err := rows.Scan(
			&card.ID,
			&card.Name,
			&card.LastFour,
			&card.StatementDay,
			&card.DaysUntilDue,
			&creditLimit,
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
		       reviewed_at, scheduled_payment_date,
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
		var reviewedAt sql.NullTime
		var scheduledPaymentDate sql.NullString

		err := rows.Scan(
			&stmt.ID,
			&stmt.CardID,
			&stmt.StatementDate,
			&stmt.DueDate,
			&stmt.Amount,
			&stmt.Status,
			&stmt.NotifiedStatement,
			&stmt.NotifiedPayment,
			&reviewedAt,
			&scheduledPaymentDate,
			&stmt.CreatedAt,
			&stmt.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning statement: %v", err)
			continue
		}

		// Handle nullable fields
		if reviewedAt.Valid {
			stmt.ReviewedAt = &reviewedAt.Time
		}
		if scheduledPaymentDate.Valid {
			stmt.ScheduledPaymentDate = &scheduledPaymentDate.String
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
		SELECT id, name, last_four, statement_day, days_until_due,
		       credit_limit, created_at, updated_at
		FROM credit_cards
		WHERE id = ?
	`

	var card models.CreditCard
	var creditLimit sql.NullFloat64

	err = database.DB.QueryRow(query, id).Scan(
		&card.ID,
		&card.Name,
		&card.LastFour,
		&card.StatementDay,
		&card.DaysUntilDue,
		&creditLimit,
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

// SchedulePaymentRequest represents the request body for scheduling a payment
type SchedulePaymentRequest struct {
	ScheduledPaymentDate string `json:"scheduled_payment_date"`
}

// SchedulePayment schedules a payment for a statement
func SchedulePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path (e.g., /api/v1/statements/1/schedule)
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	idStr := pathParts[4]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid statement ID", http.StatusBadRequest)
		return
	}

	var req SchedulePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding schedule payment request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate scheduled_payment_date
	if req.ScheduledPaymentDate == "" {
		http.Error(w, "scheduled_payment_date is required", http.StatusBadRequest)
		return
	}

	// Validate date format (ISO 8601: YYYY-MM-DD)
	_, err = time.Parse("2006-01-02", req.ScheduledPaymentDate)
	if err != nil {
		http.Error(w, "scheduled_payment_date must be in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}

	// Update statement with reviewed_at (current time) and scheduled_payment_date
	query := `
		UPDATE statements
		SET reviewed_at = ?, scheduled_payment_date = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err = database.DB.Exec(query, now, req.ScheduledPaymentDate, now, id)
	if err != nil {
		log.Printf("Error scheduling payment for statement %d: %v", id, err)
		http.Error(w, "Failed to schedule payment", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":                 "scheduled",
		"reviewed_at":            now.Format(time.RFC3339),
		"scheduled_payment_date": req.ScheduledPaymentDate,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateCardRequest represents the request body for creating a credit card
type CreateCardRequest struct {
	Name          string  `json:"name"`
	LastFour      string  `json:"last_four"`
	StatementDate string  `json:"statement_date"`
	DueDate       string  `json:"due_date"`
	CreditLimit   float64 `json:"credit_limit,omitempty"`
}

// CreateCard creates a new credit card
func CreateCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding card: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if len(req.Name) < 2 || len(req.Name) > 255 {
		http.Error(w, "name must be between 2 and 255 characters", http.StatusBadRequest)
		return
	}
	if req.LastFour == "" {
		http.Error(w, "last_four is required", http.StatusBadRequest)
		return
	}
	if len(req.LastFour) != 4 {
		http.Error(w, "last_four must be exactly 4 digits", http.StatusBadRequest)
		return
	}
	// Validate that last_four is numeric
	if _, err := strconv.Atoi(req.LastFour); err != nil {
		http.Error(w, "last_four must be numeric", http.StatusBadRequest)
		return
	}
	if req.StatementDate == "" {
		http.Error(w, "statement_date is required", http.StatusBadRequest)
		return
	}
	if req.DueDate == "" {
		http.Error(w, "due_date is required", http.StatusBadRequest)
		return
	}
	if req.CreditLimit < 0 {
		http.Error(w, "credit_limit must be positive", http.StatusBadRequest)
		return
	}

	// Parse and validate dates
	statementDate, err := time.Parse("2006-01-02", req.StatementDate)
	if err != nil {
		http.Error(w, "statement_date must be a valid date (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		http.Error(w, "due_date must be a valid date (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Validate that due_date is after statement_date
	if !dueDate.After(statementDate) {
		http.Error(w, "due_date must be after statement_date", http.StatusBadRequest)
		return
	}

	// Calculate statement_day and days_until_due
	statementDay := statementDate.Day()
	daysUntilDue := int(dueDate.Sub(statementDate).Hours() / 24)

	// Set timestamps
	now := time.Now()

	// Insert into database
	query := `
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due, credit_limit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	var result sql.Result
	if req.CreditLimit > 0 {
		result, err = database.DB.Exec(query, req.Name, req.LastFour, statementDay, daysUntilDue, req.CreditLimit, now, now)
	} else {
		result, err = database.DB.Exec(query, req.Name, req.LastFour, statementDay, daysUntilDue, nil, now, now)
	}
	if err != nil {
		log.Printf("Error creating card: %v", err)
		http.Error(w, "Failed to create card", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		http.Error(w, "Failed to create card", http.StatusInternalServerError)
		return
	}

	// Return created card
	card := models.CreditCard{
		ID:           int(id),
		Name:         req.Name,
		LastFour:     req.LastFour,
		StatementDay: statementDay,
		DaysUntilDue: daysUntilDue,
		CreditLimit:  req.CreditLimit,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
}

// UpdateCard updates an existing credit card
func UpdateCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	// Check if card exists
	var exists int
	err = database.DB.QueryRow("SELECT 1 FROM credit_cards WHERE id = ?", id).Scan(&exists)
	if err == sql.ErrNoRows {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error checking card existence: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var req CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding card: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate fields if provided
	if req.Name != "" && (len(req.Name) < 2 || len(req.Name) > 255) {
		http.Error(w, "name must be between 2 and 255 characters", http.StatusBadRequest)
		return
	}
	if req.LastFour != "" && len(req.LastFour) != 4 {
		http.Error(w, "last_four must be exactly 4 digits", http.StatusBadRequest)
		return
	}
	if req.LastFour != "" {
		if _, err := strconv.Atoi(req.LastFour); err != nil {
			http.Error(w, "last_four must be numeric", http.StatusBadRequest)
			return
		}
	}
	if req.CreditLimit < 0 {
		http.Error(w, "credit_limit must be positive", http.StatusBadRequest)
		return
	}

	// Build update query dynamically based on provided fields
	updates := []string{}
	args := []interface{}{}
	hasUpdates := false

	if req.Name != "" {
		updates = append(updates, "name = ?")
		args = append(args, req.Name)
		hasUpdates = true
	}
	if req.LastFour != "" {
		updates = append(updates, "last_four = ?")
		args = append(args, req.LastFour)
		hasUpdates = true
	}

	// Handle date updates
	if req.StatementDate != "" && req.DueDate != "" {
		statementDate, err := time.Parse("2006-01-02", req.StatementDate)
		if err != nil {
			http.Error(w, "statement_date must be a valid date (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		dueDate, err := time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			http.Error(w, "due_date must be a valid date (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}

		if !dueDate.After(statementDate) {
			http.Error(w, "due_date must be after statement_date", http.StatusBadRequest)
			return
		}

		statementDay := statementDate.Day()
		daysUntilDue := int(dueDate.Sub(statementDate).Hours() / 24)

		updates = append(updates, "statement_day = ?", "days_until_due = ?")
		args = append(args, statementDay, daysUntilDue)
		hasUpdates = true
	} else if req.StatementDate != "" || req.DueDate != "" {
		http.Error(w, "both statement_date and due_date must be provided together", http.StatusBadRequest)
		return
	}

	if req.CreditLimit > 0 {
		updates = append(updates, "credit_limit = ?")
		args = append(args, req.CreditLimit)
		hasUpdates = true
	}

	if !hasUpdates {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	// Always update updated_at
	updates = append(updates, "updated_at = ?")
	args = append(args, time.Now())

	// Add ID to args
	args = append(args, id)

	query := "UPDATE credit_cards SET " + strings.Join(updates, ", ") + " WHERE id = ?"
	_, err = database.DB.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating card %d: %v", id, err)
		http.Error(w, "Failed to update card", http.StatusInternalServerError)
		return
	}

	// Fetch and return updated card
	querySelect := `
		SELECT id, name, last_four, statement_day, days_until_due,
		       credit_limit, created_at, updated_at
		FROM credit_cards
		WHERE id = ?
	`

	var card models.CreditCard
	var creditLimit sql.NullFloat64

	err = database.DB.QueryRow(querySelect, id).Scan(
		&card.ID,
		&card.Name,
		&card.LastFour,
		&card.StatementDay,
		&card.DaysUntilDue,
		&creditLimit,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error fetching updated card %d: %v", id, err)
		http.Error(w, "Failed to fetch updated card", http.StatusInternalServerError)
		return
	}

	// Handle NULL values
	if creditLimit.Valid {
		card.CreditLimit = creditLimit.Float64
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(card)
}

// DeleteCard deletes a credit card and its associated statements
func DeleteCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
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

	// Count associated statements
	var statementCount int
	err = database.DB.QueryRow("SELECT COUNT(*) FROM statements WHERE card_id = ?", id).Scan(&statementCount)
	if err != nil {
		log.Printf("Error counting statements for card %d: %v", id, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Delete card (CASCADE will delete statements)
	result, err := database.DB.Exec("DELETE FROM credit_cards WHERE id = ?", id)
	if err != nil {
		log.Printf("Error deleting card %d: %v", id, err)
		http.Error(w, "Failed to delete card", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		http.Error(w, "Failed to delete card", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"message":         "Card deleted successfully",
		"statements_deleted": statementCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetSettings returns the current application settings
func GetSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Load config from file
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Printf("Error loading config: %v", err)
		http.Error(w, "Failed to load settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cfg)
}

// UpdateSettings updates the application settings
func UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		log.Printf("Error decoding settings: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Printf("Invalid configuration: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save configuration
	if err := config.SaveConfig("", &cfg); err != nil {
		log.Printf("Error saving config: %v", err)
		http.Error(w, "Failed to save settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cfg)
}

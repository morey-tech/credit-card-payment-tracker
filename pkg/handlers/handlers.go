package handlers

import (
	"encoding/json"
	"log"
	"net/http"

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
		err := rows.Scan(
			&card.ID,
			&card.Name,
			&card.LastFour,
			&card.StatementDay,
			&card.DueDay,
			&card.CreditLimit,
			&card.DiscordWebhookURL,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning credit card: %v", err)
			continue
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

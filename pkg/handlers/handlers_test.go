package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/morey-tech/credit-card-payment-tracker/pkg/database"
	"github.com/morey-tech/credit-card-payment-tracker/pkg/models"
)

func setupTestDB(t *testing.T) string {
	tmpDB := "./test_handlers.db"
	err := database.InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	return tmpDB
}

func teardownTestDB(tmpDB string) {
	database.Close()
	os.Remove(tmpDB)
}

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	HealthCheck(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]string
	err := json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}

	if response["message"] == "" {
		t.Error("Expected non-empty message")
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}

func TestGetCards(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	// Insert test data
	_, err := database.DB.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due, credit_limit)
		VALUES ('Test Card', '1234', 15, 25, 5000.00)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards", nil)
	w := httptest.NewRecorder()

	GetCards(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var cards []models.CreditCard
	err = json.NewDecoder(resp.Body).Decode(&cards)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(cards) != 1 {
		t.Errorf("Expected 1 card, got %d", len(cards))
	}

	if cards[0].Name != "Test Card" {
		t.Errorf("Expected card name 'Test Card', got '%s'", cards[0].Name)
	}
}

func TestGetCardsMethodNotAllowed(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", nil)
	w := httptest.NewRecorder()

	GetCards(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestGetStatements(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	// Insert test card
	result, err := database.DB.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due)
		VALUES ('Test Card', '1234', 15, 25)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test card: %v", err)
	}

	cardID, _ := result.LastInsertId()

	// Insert test statement
	_, err = database.DB.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status)
		VALUES (?, '2024-11-01', '2024-11-15', 1250.75, 'pending')
	`, cardID)
	if err != nil {
		t.Fatalf("Failed to insert test statement: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/statements", nil)
	w := httptest.NewRecorder()

	GetStatements(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var statements []models.Statement
	err = json.NewDecoder(resp.Body).Decode(&statements)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(statements) != 1 {
		t.Errorf("Expected 1 statement, got %d", len(statements))
	}

	if statements[0].Amount != 1250.75 {
		t.Errorf("Expected amount 1250.75, got %.2f", statements[0].Amount)
	}
}

func TestGetCardByID(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	// Insert test data
	result, err := database.DB.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due, credit_limit)
		VALUES ('Test Card', '5678', 20, 25, 3000.00)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	cardID, _ := result.LastInsertId()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/"+string(rune(cardID+'0')), nil)
	w := httptest.NewRecorder()

	GetCardByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var card models.CreditCard
	err = json.NewDecoder(resp.Body).Decode(&card)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if card.Name != "Test Card" {
		t.Errorf("Expected card name 'Test Card', got '%s'", card.Name)
	}

	if card.LastFour != "5678" {
		t.Errorf("Expected last_four '5678', got '%s'", card.LastFour)
	}
}

func TestGetCardByIDNotFound(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/9999", nil)
	w := httptest.NewRecorder()

	GetCardByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestGetCardByIDInvalidID(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards/invalid", nil)
	w := httptest.NewRecorder()

	GetCardByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateStatement(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	// Insert test card
	result, err := database.DB.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due)
		VALUES ('Test Card', '1234', 15, 25)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test card: %v", err)
	}

	cardID, _ := result.LastInsertId()

	// Create statement request
	stmt := models.Statement{
		CardID:        int(cardID),
		StatementDate: "2024-11-01",
		DueDate:       "2024-11-15",
		Amount:        1500.50,
	}

	body, err := json.Marshal(stmt)
	if err != nil {
		t.Fatalf("Failed to marshal statement: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/statements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var createdStmt models.Statement
	err = json.NewDecoder(resp.Body).Decode(&createdStmt)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if createdStmt.ID == 0 {
		t.Error("Expected non-zero ID")
	}

	if createdStmt.Amount != 1500.50 {
		t.Errorf("Expected amount 1500.50, got %.2f", createdStmt.Amount)
	}

	if createdStmt.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", createdStmt.Status)
	}
}

func TestCreateStatementInvalidJSON(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/statements", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateStatementMissingCardID(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	stmt := models.Statement{
		StatementDate: "2024-11-01",
		DueDate:       "2024-11-15",
		Amount:        1500.50,
	}

	body, _ := json.Marshal(stmt)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/statements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateStatementInvalidAmount(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	stmt := models.Statement{
		CardID:        1,
		StatementDate: "2024-11-01",
		DueDate:       "2024-11-15",
		Amount:        -100.00,
	}

	body, _ := json.Marshal(stmt)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/statements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdateStatement(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	// Insert test card and statement
	result, err := database.DB.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due)
		VALUES ('Test Card', '1234', 15, 25)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test card: %v", err)
	}

	cardID, _ := result.LastInsertId()

	result, err = database.DB.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status)
		VALUES (?, '2024-11-01', '2024-11-15', 1250.75, 'pending')
	`, cardID)
	if err != nil {
		t.Fatalf("Failed to insert test statement: %v", err)
	}

	stmtID, _ := result.LastInsertId()

	// Update statement
	updates := map[string]string{
		"status": "paid",
	}

	body, _ := json.Marshal(updates)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/statements/"+string(rune(stmtID+'0')), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify update
	var status string
	err = database.DB.QueryRow("SELECT status FROM statements WHERE id = ?", stmtID).Scan(&status)
	if err != nil {
		t.Fatalf("Failed to query updated statement: %v", err)
	}

	if status != "paid" {
		t.Errorf("Expected status 'paid', got '%s'", status)
	}
}

func TestUpdateStatementInvalidID(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	updates := map[string]string{
		"status": "paid",
	}

	body, _ := json.Marshal(updates)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/statements/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdateStatementMissingStatus(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	updates := map[string]string{}

	body, _ := json.Marshal(updates)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/statements/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestGetStatementsMethodNotAllowed(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/statements", nil)
	w := httptest.NewRecorder()

	GetStatements(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestGetCardByIDMethodNotAllowed(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/1", nil)
	w := httptest.NewRecorder()

	GetCardByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestCreateStatementMethodNotAllowed(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/statements", nil)
	w := httptest.NewRecorder()

	CreateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestUpdateStatementMethodNotAllowed(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/statements/1", nil)
	w := httptest.NewRecorder()

	UpdateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestCreateStatementMissingStatementDate(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	stmt := models.Statement{
		CardID:  1,
		DueDate: "2024-11-15",
		Amount:  1500.50,
	}

	body, _ := json.Marshal(stmt)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/statements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateStatementMissingDueDate(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	stmt := models.Statement{
		CardID:        1,
		StatementDate: "2024-11-01",
		Amount:        1500.50,
	}

	body, _ := json.Marshal(stmt)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/statements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdateStatementInvalidJSON(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/statements/1", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateStatement(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

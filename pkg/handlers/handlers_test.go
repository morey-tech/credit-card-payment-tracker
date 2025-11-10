package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/morey-tech/credit-card-payment-tracker/pkg/config"
	"github.com/morey-tech/credit-card-payment-tracker/pkg/database"
	"github.com/morey-tech/credit-card-payment-tracker/pkg/models"
)

func setupTestDB(t *testing.T) string {
	tmpDB := "./test_handlers.db"
	err := database.InitDB(tmpDB)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	// Enable foreign keys for CASCADE deletes
	_, err = database.DB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
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

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/cards/%d", cardID), nil)
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
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/statements/%d", stmtID), bytes.NewReader(body))
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

// --- Tests for CreateCard ---

func TestCreateCard_Success(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	cardReq := CreateCardRequest{
		Name:          "Chase Sapphire",
		LastFour:      "1234",
		StatementDate: "2024-11-15",
		DueDate:       "2024-12-10",
		CreditLimit:   5000.00,
	}

	body, _ := json.Marshal(cardReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var card models.CreditCard
	err := json.NewDecoder(resp.Body).Decode(&card)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if card.ID == 0 {
		t.Error("Expected non-zero ID")
	}
	if card.Name != "Chase Sapphire" {
		t.Errorf("Expected name 'Chase Sapphire', got '%s'", card.Name)
	}
	if card.LastFour != "1234" {
		t.Errorf("Expected last_four '1234', got '%s'", card.LastFour)
	}
	if card.StatementDay != 15 {
		t.Errorf("Expected statement_day 15, got %d", card.StatementDay)
	}
	if card.DaysUntilDue != 25 {
		t.Errorf("Expected days_until_due 25, got %d", card.DaysUntilDue)
	}
	if card.CreditLimit != 5000.00 {
		t.Errorf("Expected credit_limit 5000.00, got %.2f", card.CreditLimit)
	}
}

func TestCreateCard_MissingName(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	cardReq := CreateCardRequest{
		LastFour:      "1234",
		StatementDate: "2024-11-15",
		DueDate:       "2024-12-10",
	}

	body, _ := json.Marshal(cardReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateCard_NameTooShort(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	cardReq := CreateCardRequest{
		Name:          "A",
		LastFour:      "1234",
		StatementDate: "2024-11-15",
		DueDate:       "2024-12-10",
	}

	body, _ := json.Marshal(cardReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateCard_LastFourNotFourDigits(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	cardReq := CreateCardRequest{
		Name:          "Test Card",
		LastFour:      "123",
		StatementDate: "2024-11-15",
		DueDate:       "2024-12-10",
	}

	body, _ := json.Marshal(cardReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateCard_LastFourNotNumeric(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	cardReq := CreateCardRequest{
		Name:          "Test Card",
		LastFour:      "abcd",
		StatementDate: "2024-11-15",
		DueDate:       "2024-12-10",
	}

	body, _ := json.Marshal(cardReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateCard_InvalidStatementDate(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	cardReq := CreateCardRequest{
		Name:          "Test Card",
		LastFour:      "1234",
		StatementDate: "invalid-date",
		DueDate:       "2024-12-10",
	}

	body, _ := json.Marshal(cardReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateCard_InvalidDueDate(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	cardReq := CreateCardRequest{
		Name:          "Test Card",
		LastFour:      "1234",
		StatementDate: "2024-11-15",
		DueDate:       "invalid-date",
	}

	body, _ := json.Marshal(cardReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateCard_DueDateNotAfterStatementDate(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	cardReq := CreateCardRequest{
		Name:          "Test Card",
		LastFour:      "1234",
		StatementDate: "2024-12-10",
		DueDate:       "2024-11-15",
	}

	body, _ := json.Marshal(cardReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateCard_DueDateSameAsStatementDate(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	cardReq := CreateCardRequest{
		Name:          "Test Card",
		LastFour:      "1234",
		StatementDate: "2024-11-15",
		DueDate:       "2024-11-15",
	}

	body, _ := json.Marshal(cardReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateCard_NegativeCreditLimit(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	cardReq := CreateCardRequest{
		Name:          "Test Card",
		LastFour:      "1234",
		StatementDate: "2024-11-15",
		DueDate:       "2024-12-10",
		CreditLimit:   -1000.00,
	}

	body, _ := json.Marshal(cardReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateCard_WithoutCreditLimit(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	cardReq := CreateCardRequest{
		Name:          "Test Card",
		LastFour:      "1234",
		StatementDate: "2024-11-15",
		DueDate:       "2024-12-10",
	}

	body, _ := json.Marshal(cardReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestCreateCard_InvalidJSON(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateCard_MethodNotAllowed(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cards", nil)
	w := httptest.NewRecorder()

	CreateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

// --- Tests for UpdateCard ---

func TestUpdateCard_Success(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	// Insert test card
	result, err := database.DB.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due, credit_limit)
		VALUES ('Original Name', '1234', 15, 25, 3000.00)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test card: %v", err)
	}

	cardID, _ := result.LastInsertId()

	// Update card
	updateReq := CreateCardRequest{
		Name:     "Updated Name",
		LastFour: "5678",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/cards/%d", cardID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateCard(w, req)

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

	if card.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", card.Name)
	}
	if card.LastFour != "5678" {
		t.Errorf("Expected last_four '5678', got '%s'", card.LastFour)
	}
}

func TestUpdateCard_UpdateDates(t *testing.T) {
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

	// Update dates
	updateReq := CreateCardRequest{
		StatementDate: "2024-11-20",
		DueDate:       "2024-12-15",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/cards/%d", cardID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateCard(w, req)

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

	if card.StatementDay != 20 {
		t.Errorf("Expected statement_day 20, got %d", card.StatementDay)
	}
	if card.DaysUntilDue != 25 {
		t.Errorf("Expected days_until_due 25, got %d", card.DaysUntilDue)
	}
}

func TestUpdateCard_CardNotFound(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	updateReq := CreateCardRequest{
		Name: "Updated Name",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/cards/9999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestUpdateCard_OnlyStatementDateProvided(t *testing.T) {
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

	// Try to update only statement_date
	updateReq := CreateCardRequest{
		StatementDate: "2024-11-20",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/cards/%d", cardID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdateCard_InvalidID(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	updateReq := CreateCardRequest{
		Name: "Updated Name",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/cards/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdateCard_NoFieldsToUpdate(t *testing.T) {
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

	// Empty update
	updateReq := CreateCardRequest{}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/cards/%d", cardID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdateCard_MethodNotAllowed(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/1", nil)
	w := httptest.NewRecorder()

	UpdateCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

// --- Tests for DeleteCard ---

func TestDeleteCard_Success(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	// Insert test card with statements
	result, err := database.DB.Exec(`
		INSERT INTO credit_cards (name, last_four, statement_day, days_until_due)
		VALUES ('Test Card', '1234', 15, 25)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test card: %v", err)
	}

	cardID, _ := result.LastInsertId()

	// Insert statements
	_, err = database.DB.Exec(`
		INSERT INTO statements (card_id, statement_date, due_date, amount, status)
		VALUES (?, '2024-11-01', '2024-11-15', 1000.00, 'pending')
	`, cardID)
	if err != nil {
		t.Fatalf("Failed to insert test statement: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/cards/%d", cardID), nil)
	w := httptest.NewRecorder()

	DeleteCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "Card deleted successfully" {
		t.Errorf("Expected success message, got '%v'", response["message"])
	}

	statementsDeleted := response["statements_deleted"].(float64)
	if statementsDeleted != 1 {
		t.Errorf("Expected 1 statement deleted, got %v", statementsDeleted)
	}

	// Verify card is deleted
	var count int
	err = database.DB.QueryRow("SELECT COUNT(*) FROM credit_cards WHERE id = ?", cardID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query cards: %v", err)
	}
	if count != 0 {
		t.Error("Card was not deleted")
	}

	// Verify statements are deleted (CASCADE)
	err = database.DB.QueryRow("SELECT COUNT(*) FROM statements WHERE card_id = ?", cardID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query statements: %v", err)
	}
	if count != 0 {
		t.Error("Statements were not cascaded deleted")
	}
}

func TestDeleteCard_NotFound(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/cards/9999", nil)
	w := httptest.NewRecorder()

	DeleteCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestDeleteCard_InvalidID(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/cards/invalid", nil)
	w := httptest.NewRecorder()

	DeleteCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestDeleteCard_MethodNotAllowed(t *testing.T) {
	tmpDB := setupTestDB(t)
	defer teardownTestDB(tmpDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cards/1", nil)
	w := httptest.NewRecorder()

	DeleteCard(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

// --- Tests for GetSettings and UpdateSettings ---

func TestGetSettings_Success(t *testing.T) {
	// Create a temporary config file
	tmpConfig := "./test_config.yaml"
	defer os.Remove(tmpConfig)

	os.Setenv("CONFIG_PATH", tmpConfig)
	defer os.Unsetenv("CONFIG_PATH")

	// Write test config
	testConfig := `discord_webhook_url: "https://discord.com/api/webhooks/123/abc"`
	err := os.WriteFile(tmpConfig, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()

	GetSettings(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var cfg config.Config
	err = json.NewDecoder(resp.Body).Decode(&cfg)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if cfg.DiscordWebhookURL != "https://discord.com/api/webhooks/123/abc" {
		t.Errorf("Expected webhook URL, got '%s'", cfg.DiscordWebhookURL)
	}
}

func TestGetSettings_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/settings", nil)
	w := httptest.NewRecorder()

	GetSettings(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

func TestUpdateSettings_Success(t *testing.T) {
	tmpConfig := "./test_config_update.yaml"
	defer os.Remove(tmpConfig)

	os.Setenv("CONFIG_PATH", tmpConfig)
	defer os.Unsetenv("CONFIG_PATH")

	settingsReq := config.Config{
		DiscordWebhookURL: "https://discord.com/api/webhooks/456/xyz",
	}

	body, _ := json.Marshal(settingsReq)
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateSettings(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify config was saved
	data, err := os.ReadFile(tmpConfig)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	if !bytes.Contains(data, []byte("https://discord.com/api/webhooks/456/xyz")) {
		t.Error("Config file does not contain expected webhook URL")
	}
}

func TestUpdateSettings_InvalidWebhookURL(t *testing.T) {
	tmpConfig := "./test_config_invalid.yaml"
	defer os.Remove(tmpConfig)

	os.Setenv("CONFIG_PATH", tmpConfig)
	defer os.Unsetenv("CONFIG_PATH")

	settingsReq := config.Config{
		DiscordWebhookURL: "https://example.com/invalid",
	}

	body, _ := json.Marshal(settingsReq)
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateSettings(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdateSettings_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdateSettings(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdateSettings_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/settings", nil)
	w := httptest.NewRecorder()

	UpdateSettings(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

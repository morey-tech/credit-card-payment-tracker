package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStatementJSONMarshaling(t *testing.T) {
	stmt := Statement{
		ID:                 1,
		CardID:             5,
		StatementDate:      "2024-11-01",
		DueDate:            "2024-11-15",
		Amount:             1250.75,
		Status:             "pending",
		NotifiedStatement:  false,
		NotifiedPayment:    false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Marshal to JSON
	data, err := json.Marshal(stmt)
	if err != nil {
		t.Fatalf("Failed to marshal statement: %v", err)
	}

	// Unmarshal back
	var unmarshaled Statement
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal statement: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != stmt.ID {
		t.Errorf("Expected ID %d, got %d", stmt.ID, unmarshaled.ID)
	}

	if unmarshaled.CardID != stmt.CardID {
		t.Errorf("Expected card_id %d, got %d", stmt.CardID, unmarshaled.CardID)
	}

	if unmarshaled.StatementDate != stmt.StatementDate {
		t.Errorf("Expected statement_date '%s', got '%s'", stmt.StatementDate, unmarshaled.StatementDate)
	}

	if unmarshaled.DueDate != stmt.DueDate {
		t.Errorf("Expected due_date '%s', got '%s'", stmt.DueDate, unmarshaled.DueDate)
	}

	if unmarshaled.Amount != stmt.Amount {
		t.Errorf("Expected amount %.2f, got %.2f", stmt.Amount, unmarshaled.Amount)
	}

	if unmarshaled.Status != stmt.Status {
		t.Errorf("Expected status '%s', got '%s'", stmt.Status, unmarshaled.Status)
	}

	if unmarshaled.NotifiedStatement != stmt.NotifiedStatement {
		t.Errorf("Expected notified_statement %v, got %v", stmt.NotifiedStatement, unmarshaled.NotifiedStatement)
	}

	if unmarshaled.NotifiedPayment != stmt.NotifiedPayment {
		t.Errorf("Expected notified_payment %v, got %v", stmt.NotifiedPayment, unmarshaled.NotifiedPayment)
	}
}

func TestStatementJSONTags(t *testing.T) {
	stmt := Statement{
		ID:                 1,
		CardID:             2,
		StatementDate:      "2024-10-01",
		DueDate:            "2024-10-15",
		Amount:             500.00,
		Status:             "paid",
		NotifiedStatement:  true,
		NotifiedPayment:    true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	data, err := json.Marshal(stmt)
	if err != nil {
		t.Fatalf("Failed to marshal statement: %v", err)
	}

	// Parse JSON to verify field names
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Verify snake_case field names
	expectedFields := []string{
		"id",
		"card_id",
		"statement_date",
		"due_date",
		"amount",
		"status",
		"notified_statement",
		"notified_payment",
		"created_at",
		"updated_at",
	}

	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Expected field '%s' not found in JSON", field)
		}
	}
}

func TestStatementUnmarshalPartial(t *testing.T) {
	jsonData := `{
		"card_id": 3,
		"statement_date": "2024-09-15",
		"due_date": "2024-09-30",
		"amount": 2500.00
	}`

	var stmt Statement
	err := json.Unmarshal([]byte(jsonData), &stmt)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if stmt.CardID != 3 {
		t.Errorf("Expected card_id 3, got %d", stmt.CardID)
	}

	if stmt.StatementDate != "2024-09-15" {
		t.Errorf("Expected statement_date '2024-09-15', got '%s'", stmt.StatementDate)
	}

	if stmt.DueDate != "2024-09-30" {
		t.Errorf("Expected due_date '2024-09-30', got '%s'", stmt.DueDate)
	}

	if stmt.Amount != 2500.00 {
		t.Errorf("Expected amount 2500.00, got %.2f", stmt.Amount)
	}

	// Fields not in JSON should have zero values
	if stmt.ID != 0 {
		t.Errorf("Expected ID 0, got %d", stmt.ID)
	}

	if stmt.Status != "" {
		t.Errorf("Expected empty status, got '%s'", stmt.Status)
	}

	if stmt.NotifiedStatement {
		t.Error("Expected notified_statement false")
	}

	if stmt.NotifiedPayment {
		t.Error("Expected notified_payment false")
	}
}

func TestStatementStatusValues(t *testing.T) {
	statuses := []string{"pending", "paid", "overdue"}

	for _, status := range statuses {
		stmt := Statement{
			ID:            1,
			CardID:        1,
			StatementDate: "2024-11-01",
			DueDate:       "2024-11-15",
			Amount:        1000.00,
			Status:        status,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		data, err := json.Marshal(stmt)
		if err != nil {
			t.Fatalf("Failed to marshal statement with status '%s': %v", status, err)
		}

		var unmarshaled Statement
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal statement with status '%s': %v", status, err)
		}

		if unmarshaled.Status != status {
			t.Errorf("Expected status '%s', got '%s'", status, unmarshaled.Status)
		}
	}
}

func TestStatementBooleanFields(t *testing.T) {
	testCases := []struct {
		name              string
		notifiedStatement bool
		notifiedPayment   bool
	}{
		{"both false", false, false},
		{"statement true", true, false},
		{"payment true", false, true},
		{"both true", true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stmt := Statement{
				ID:                 1,
				CardID:             1,
				StatementDate:      "2024-11-01",
				DueDate:            "2024-11-15",
				Amount:             750.00,
				Status:             "pending",
				NotifiedStatement:  tc.notifiedStatement,
				NotifiedPayment:    tc.notifiedPayment,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}

			data, err := json.Marshal(stmt)
			if err != nil {
				t.Fatalf("Failed to marshal statement: %v", err)
			}

			var unmarshaled Statement
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal statement: %v", err)
			}

			if unmarshaled.NotifiedStatement != tc.notifiedStatement {
				t.Errorf("Expected notified_statement %v, got %v", tc.notifiedStatement, unmarshaled.NotifiedStatement)
			}

			if unmarshaled.NotifiedPayment != tc.notifiedPayment {
				t.Errorf("Expected notified_payment %v, got %v", tc.notifiedPayment, unmarshaled.NotifiedPayment)
			}
		})
	}
}

func TestStatementAmountPrecision(t *testing.T) {
	amounts := []float64{
		0.01,
		1.99,
		100.50,
		1250.75,
		9999.99,
	}

	for _, amount := range amounts {
		stmt := Statement{
			ID:            1,
			CardID:        1,
			StatementDate: "2024-11-01",
			DueDate:       "2024-11-15",
			Amount:        amount,
			Status:        "pending",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		data, err := json.Marshal(stmt)
		if err != nil {
			t.Fatalf("Failed to marshal statement with amount %.2f: %v", amount, err)
		}

		var unmarshaled Statement
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal statement with amount %.2f: %v", amount, err)
		}

		if unmarshaled.Amount != amount {
			t.Errorf("Amount precision lost: expected %.2f, got %.2f", amount, unmarshaled.Amount)
		}
	}
}

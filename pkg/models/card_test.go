package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCreditCardJSONMarshaling(t *testing.T) {
	card := CreditCard{
		ID:           1,
		Name:         "Test Visa",
		LastFour:     "1234",
		StatementDay: 15,
		DueDay:       10,
		CreditLimit:  5000.00,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Marshal to JSON
	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("Failed to marshal card: %v", err)
	}

	// Unmarshal back
	var unmarshaled CreditCard
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal card: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != card.ID {
		t.Errorf("Expected ID %d, got %d", card.ID, unmarshaled.ID)
	}

	if unmarshaled.Name != card.Name {
		t.Errorf("Expected name '%s', got '%s'", card.Name, unmarshaled.Name)
	}

	if unmarshaled.LastFour != card.LastFour {
		t.Errorf("Expected last_four '%s', got '%s'", card.LastFour, unmarshaled.LastFour)
	}

	if unmarshaled.StatementDay != card.StatementDay {
		t.Errorf("Expected statement_day %d, got %d", card.StatementDay, unmarshaled.StatementDay)
	}

	if unmarshaled.DueDay != card.DueDay {
		t.Errorf("Expected due_day %d, got %d", card.DueDay, unmarshaled.DueDay)
	}

	if unmarshaled.CreditLimit != card.CreditLimit {
		t.Errorf("Expected credit_limit %.2f, got %.2f", card.CreditLimit, unmarshaled.CreditLimit)
	}
}

func TestCreditCardJSONTags(t *testing.T) {
	card := CreditCard{
		ID:           1,
		Name:         "Test Card",
		LastFour:     "5678",
		StatementDay: 20,
		DueDay:       15,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("Failed to marshal card: %v", err)
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
		"name",
		"last_four",
		"statement_day",
		"due_day",
		"created_at",
		"updated_at",
	}

	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Expected field '%s' not found in JSON", field)
		}
	}
}

func TestCreditCardOmitEmpty(t *testing.T) {
	// Card without optional fields
	card := CreditCard{
		ID:           1,
		Name:         "Test Card",
		LastFour:     "9999",
		StatementDay: 25,
		DueDay:       20,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("Failed to marshal card: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// CreditLimit should be omitted when zero
	if _, exists := result["credit_limit"]; exists {
		t.Error("Expected credit_limit to be omitted when zero")
	}

	// DiscordWebhookURL should be omitted when empty
	if _, exists := result["discord_webhook_url"]; exists {
		t.Error("Expected discord_webhook_url to be omitted when empty")
	}
}

func TestCreditCardWithOptionalFields(t *testing.T) {
	card := CreditCard{
		ID:                1,
		Name:              "Premium Card",
		LastFour:          "8888",
		StatementDay:      1,
		DueDay:            25,
		CreditLimit:       15000.00,
		DiscordWebhookURL: "https://discord.com/api/webhooks/123/abc",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	data, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("Failed to marshal card: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Optional fields should be present
	if _, exists := result["credit_limit"]; !exists {
		t.Error("Expected credit_limit to be present")
	}

	if _, exists := result["discord_webhook_url"]; !exists {
		t.Error("Expected discord_webhook_url to be present")
	}
}

func TestCreditCardUnmarshalPartial(t *testing.T) {
	jsonData := `{
		"id": 42,
		"name": "Partial Card",
		"last_four": "7777",
		"statement_day": 10,
		"due_day": 5
	}`

	var card CreditCard
	err := json.Unmarshal([]byte(jsonData), &card)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if card.ID != 42 {
		t.Errorf("Expected ID 42, got %d", card.ID)
	}

	if card.Name != "Partial Card" {
		t.Errorf("Expected name 'Partial Card', got '%s'", card.Name)
	}

	if card.LastFour != "7777" {
		t.Errorf("Expected last_four '7777', got '%s'", card.LastFour)
	}

	if card.StatementDay != 10 {
		t.Errorf("Expected statement_day 10, got %d", card.StatementDay)
	}

	if card.DueDay != 5 {
		t.Errorf("Expected due_day 5, got %d", card.DueDay)
	}

	// Optional fields should have zero values
	if card.CreditLimit != 0 {
		t.Errorf("Expected credit_limit 0, got %.2f", card.CreditLimit)
	}

	if card.DiscordWebhookURL != "" {
		t.Errorf("Expected empty discord_webhook_url, got '%s'", card.DiscordWebhookURL)
	}
}

package models

import "time"

// CreditCard represents a credit card in the system
type CreditCard struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	LastFour          string    `json:"last_four"`
	StatementDay      int       `json:"statement_day"`
	DueDay            int       `json:"due_day"`
	CreditLimit       float64   `json:"credit_limit,omitempty"`
	DiscordWebhookURL string    `json:"discord_webhook_url,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

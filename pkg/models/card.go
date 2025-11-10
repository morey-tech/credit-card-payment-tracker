package models

import "time"

// CreditCard represents a credit card in the system
type CreditCard struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	LastFour     string    `json:"last_four"`
	StatementDay int       `json:"statement_day"`
	DaysUntilDue int       `json:"days_until_due"`
	CreditLimit  float64   `json:"credit_limit,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

package models

import "time"

// Statement represents a credit card statement
type Statement struct {
	ID                   int        `json:"id"`
	CardID               int        `json:"card_id"`
	StatementDate        string     `json:"statement_date"`
	DueDate              string     `json:"due_date"`
	Amount               float64    `json:"amount"`
	Status               string     `json:"status"`
	NotifiedStatement    bool       `json:"notified_statement"`
	NotifiedPayment      bool       `json:"notified_payment"`
	ReviewedAt           *time.Time `json:"reviewed_at,omitempty"`
	ScheduledPaymentDate *string    `json:"scheduled_payment_date,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

package entity

import "time"

// Product represents a product in the POS system.
type Product struct {
	ID         string
	OrgID      string
	Name       string
	SKU        string
	PriceCents int64
	Active     bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

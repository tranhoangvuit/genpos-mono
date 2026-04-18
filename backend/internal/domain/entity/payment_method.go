package entity

import "time"

// PaymentMethod is a payment option an org accepts.
type PaymentMethod struct {
	ID        string
	OrgID     string
	Name      string
	Type      string
	IsActive  bool
	SortOrder int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

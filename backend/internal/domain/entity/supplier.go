package entity

import "time"

// Supplier represents an organization's supplier.
type Supplier struct {
	ID          string
	OrgID       string
	Name        string
	ContactName string
	Email       string
	Phone       string
	Address     string
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

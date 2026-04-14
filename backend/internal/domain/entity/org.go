package entity

import "time"

// Org represents a tenant organization (business).
type Org struct {
	ID        string
	Slug      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

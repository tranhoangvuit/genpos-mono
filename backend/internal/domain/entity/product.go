package entity

import "time"

// Product represents a product in the POS system.
type Product struct {
	ID          string
	OrgID       string
	CategoryID  string
	Name        string
	Description string
	ImageURL    string
	IsActive    bool
	SortOrder   int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

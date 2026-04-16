package entity

import "time"

// Role represents an org-scoped permission group.
type Role struct {
	ID          string
	OrgID       string
	Name        string
	Permissions map[string]string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

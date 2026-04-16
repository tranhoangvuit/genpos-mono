package entity

import "time"

// User represents a member of an org.
type User struct {
	ID           string
	OrgID        string
	RoleID       string
	Email        string
	PasswordHash string
	Name         string
	RoleName     string
	Permissions  map[string]string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

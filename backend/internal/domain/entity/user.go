package entity

import "time"

// User represents a member of an org. PasswordHash is the argon2id PHC string
// and must never leave the backend.
type User struct {
	ID           string
	OrgID        string
	Email        string
	PasswordHash string
	Name         string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

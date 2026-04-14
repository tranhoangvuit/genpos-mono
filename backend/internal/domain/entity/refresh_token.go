package entity

import "time"

// RefreshToken is a hashed record of an issued refresh cookie. The plaintext
// token value is never stored — only the sha256 hex digest.
type RefreshToken struct {
	ID        string
	UserID    string
	OrgID     string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	UserAgent string
	CreatedAt time.Time
}

// IsActive reports whether the token can still be used for rotation.
func (t *RefreshToken) IsActive(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	return now.Before(t.ExpiresAt)
}

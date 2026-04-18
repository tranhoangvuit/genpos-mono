package entity

import "time"

// Store represents a physical location or outlet under an organization.
type Store struct {
	ID        string
	OrgID     string
	Name      string
	Address   string
	Phone     string
	Email     string
	Timezone  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

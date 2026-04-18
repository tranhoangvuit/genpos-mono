package entity

import "time"

// Member is a user of an org, exposed via the settings/members UI.
type Member struct {
	ID        string
	OrgID     string
	Name      string
	Email     string
	Phone     string
	RoleID    string
	RoleName  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RoleOption is a role selection for the members dropdown.
type RoleOption struct {
	ID       string
	Name     string
	IsSystem bool
}

package auth

import "github.com/genpick/genpos-mono/backend/internal/domain/gateway"

// PermissionSet maps resource → action (e.g. "orders" → "*").
type PermissionSet map[string]string

// Allows returns true when the set grants the requested resource:action.
func (p PermissionSet) Allows(resource, action string) bool {
	if p["*"] == "*" {
		return true
	}
	v, ok := p[resource]
	if !ok {
		return false
	}
	return v == "*" || v == action
}

// RoleSeed describes a role to be created during org signup.
type RoleSeed struct {
	Name        string
	Permissions PermissionSet
	IsSystem    bool
}

// DefaultRoles returns the three system roles seeded on signup.
func DefaultRoles() []RoleSeed {
	return []RoleSeed{
		{
			Name:     "admin",
			IsSystem: true,
			Permissions: PermissionSet{
				"*": "*",
			},
		},
		{
			Name:     "manager",
			IsSystem: false,
			Permissions: PermissionSet{
				"orders":    "*",
				"inventory": "*",
				"reports":   "*",
				"users":     "read",
				"products":  "*",
				"customers": "*",
			},
		},
		{
			Name:     "cashier",
			IsSystem: false,
			Permissions: PermissionSet{
				"orders":    "create",
				"customers": "read",
				"products":  "read",
			},
		},
	}
}

// ToCreateRoleParams converts a RoleSeed into a gateway.CreateRoleParams.
func (s RoleSeed) ToCreateRoleParams(orgID string) gateway.CreateRoleParams {
	return gateway.CreateRoleParams{
		OrgID:       orgID,
		Name:        s.Name,
		Permissions: map[string]string(s.Permissions),
		IsSystem:    s.IsSystem,
	}
}

package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=role_gateway.go -destination=mock/mock_role_gateway.go -package=mock

// CreateRoleParams carries parameters for creating a role.
type CreateRoleParams struct {
	OrgID       string
	Name        string
	Permissions map[string]string
	IsSystem    bool
}

// RoleReader reads roles.
type RoleReader interface {
	GetByOrgAndName(ctx context.Context, orgID, name string) (*entity.Role, error)
}

// RoleWriter creates roles.
type RoleWriter interface {
	Create(ctx context.Context, params CreateRoleParams) (*entity.Role, error)
}

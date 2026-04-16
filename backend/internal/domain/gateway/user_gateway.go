package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=user_gateway.go -destination=mock/mock_user_gateway.go -package=mock

// CreateUserParams carries parameters for creating a user.
type CreateUserParams struct {
	OrgID        string
	RoleID       string
	Email        string
	PasswordHash string
	Name         string
}

// UserReader reads users without tenant context (used during auth).
type UserReader interface {
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
}

// UserWriter creates users without tenant context (used during signup).
type UserWriter interface {
	Create(ctx context.Context, params CreateUserParams) (*entity.User, error)
}

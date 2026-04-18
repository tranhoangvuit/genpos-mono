package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=member_gateway.go -destination=mock/mock_member_gateway.go -package=mock

type CreateMemberParams struct {
	OrgID        string
	RoleID       string
	Name         string
	Email        string
	Phone        string
	PasswordHash string
}

type UpdateMemberParams struct {
	ID     string
	Name   string
	Phone  string
	RoleID string
	Status string
}

// MemberReader reads members (users exposed through settings/members).
type MemberReader interface {
	List(ctx context.Context) ([]*entity.Member, error)
	GetByID(ctx context.Context, id string) (*entity.Member, error)
	ListRoleOptions(ctx context.Context, orgID string) ([]*entity.RoleOption, error)
}

// MemberWriter mutates members within a tenant-scoped tx.
type MemberWriter interface {
	Create(ctx context.Context, params CreateMemberParams) (string, error)
	Update(ctx context.Context, params UpdateMemberParams) error
	SoftDelete(ctx context.Context, id string) error
}

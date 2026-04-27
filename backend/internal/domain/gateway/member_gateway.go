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
	AllStores    bool
	StoreIDs     []string
}

type UpdateMemberParams struct {
	ID        string
	Name      string
	Phone     string
	RoleID    string
	Status    string
	AllStores bool
	StoreIDs  []string
}

// MemberReader reads members (users exposed through settings/members).
type MemberReader interface {
	List(ctx context.Context) ([]*entity.Member, error)
	GetByID(ctx context.Context, id string) (*entity.Member, error)
	ListRoleOptions(ctx context.Context, orgID string) ([]*entity.RoleOption, error)
	// HasStoreAccess returns true if the user can operate the given store
	// (either via all_stores or an explicit user_stores row).
	HasStoreAccess(ctx context.Context, userID, storeID string) (bool, error)
}

// MemberWriter mutates members within a tenant-scoped tx.
type MemberWriter interface {
	Create(ctx context.Context, params CreateMemberParams) (string, error)
	Update(ctx context.Context, params UpdateMemberParams) error
	SoftDelete(ctx context.Context, id string) error
	// ReplaceStores wipes user_stores for the user and inserts the given set
	// (deduped). Caller is responsible for supplying org_id-correct store IDs.
	ReplaceStores(ctx context.Context, orgID, userID string, storeIDs []string) error
}

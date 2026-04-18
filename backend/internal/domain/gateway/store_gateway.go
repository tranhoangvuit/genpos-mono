package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=store_gateway.go -destination=mock/mock_store_gateway.go -package=mock

// CreateStoreParams carries parameters for creating a store.
type CreateStoreParams struct {
	OrgID    string
	Name     string
	Address  string
	Phone    string
	Email    string
	Timezone string
	Status   string
}

// UpdateStoreParams carries parameters for updating a store.
type UpdateStoreParams struct {
	ID       string
	Name     string
	Address  string
	Phone    string
	Email    string
	Timezone string
	Status   string
}

// StoreReader reads stores. GetFirstForOrg is used outside tenant context
// during auth bootstrap; List runs inside tenant context.
type StoreReader interface {
	GetFirstForOrg(ctx context.Context, orgID string) (*entity.Store, error)
	List(ctx context.Context) ([]*entity.Store, error)
}

// StoreWriter mutates stores. Create may run outside a tenant tx during signup;
// Update/SoftDelete run inside a tenant-scoped transaction.
type StoreWriter interface {
	Create(ctx context.Context, params CreateStoreParams) (*entity.Store, error)
	Update(ctx context.Context, params UpdateStoreParams) (*entity.Store, error)
	SoftDelete(ctx context.Context, id string) error
}

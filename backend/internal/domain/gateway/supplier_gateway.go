package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=supplier_gateway.go -destination=mock/mock_supplier_gateway.go -package=mock

type CreateSupplierParams struct {
	OrgID       string
	Name        string
	ContactName string
	Email       string
	Phone       string
	Address     string
	Notes       string
}

type UpdateSupplierParams struct {
	ID          string
	Name        string
	ContactName string
	Email       string
	Phone       string
	Address     string
	Notes       string
}

// SupplierReader reads suppliers.
type SupplierReader interface {
	GetByID(ctx context.Context, id string) (*entity.Supplier, error)
	List(ctx context.Context) ([]*entity.Supplier, error)
}

// SupplierWriter mutates suppliers within a tenant-scoped transaction.
type SupplierWriter interface {
	Create(ctx context.Context, params CreateSupplierParams) (*entity.Supplier, error)
	Update(ctx context.Context, params UpdateSupplierParams) (*entity.Supplier, error)
	SoftDelete(ctx context.Context, id string) error
}

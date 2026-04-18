package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=tax_rate_gateway.go -destination=mock/mock_tax_rate_gateway.go -package=mock

type CreateTaxRateParams struct {
	OrgID       string
	Name        string
	Rate        string
	IsInclusive bool
	IsDefault   bool
}

type UpdateTaxRateParams struct {
	ID          string
	Name        string
	Rate        string
	IsInclusive bool
	IsDefault   bool
}

// TaxRateReader reads tax rates.
type TaxRateReader interface {
	List(ctx context.Context) ([]*entity.TaxRate, error)
}

// TaxRateWriter mutates tax rates within a tenant-scoped tx.
type TaxRateWriter interface {
	Create(ctx context.Context, params CreateTaxRateParams) (*entity.TaxRate, error)
	Update(ctx context.Context, params UpdateTaxRateParams) (*entity.TaxRate, error)
	SoftDelete(ctx context.Context, id string) error
	ClearDefaults(ctx context.Context) error
}

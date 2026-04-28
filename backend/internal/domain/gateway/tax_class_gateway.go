package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=tax_class_gateway.go -destination=mock/mock_tax_class_gateway.go -package=mock

type TaxClassRateParams struct {
	TaxRateID  string
	Sequence   int32
	IsCompound bool
}

type CreateTaxClassParams struct {
	OrgID       string
	Name        string
	Description string
	IsDefault   bool
	SortOrder   int32
	Rates       []TaxClassRateParams
}

type UpdateTaxClassParams struct {
	ID          string
	OrgID       string
	Name        string
	Description string
	IsDefault   bool
	SortOrder   int32
	Rates       []TaxClassRateParams
}

// TaxClassReader reads tax classes with their nested rate links.
type TaxClassReader interface {
	List(ctx context.Context) ([]*entity.TaxClass, error)
	Get(ctx context.Context, id string) (*entity.TaxClass, error)
}

// TaxClassWriter mutates tax classes within a tenant-scoped tx. Rate
// membership is rewritten on every Update (the engine soft-deletes the
// previous rates and inserts the new set in one transaction).
type TaxClassWriter interface {
	Create(ctx context.Context, params CreateTaxClassParams) (*entity.TaxClass, error)
	Update(ctx context.Context, params UpdateTaxClassParams) (*entity.TaxClass, error)
	SoftDelete(ctx context.Context, id string) error
	ClearDefaults(ctx context.Context) error
	ReplaceRates(ctx context.Context, orgID, classID string, rates []TaxClassRateParams) error
}

// VariantTaxRate is one snapshot rate the cart engine receives when
// resolving a variant's tax class. Mirrors entity.TaxRate fields plus the
// class-level sequence/compound flags.
type VariantTaxRate struct {
	VariantID    string
	TaxRateID    string
	NameSnapshot string
	Rate         string // decimal fraction string (e.g. "0.1000")
	IsInclusive  bool
	IsCompound   bool
	Sequence     int32
}

// VariantTaxResolver bridges product variants to the tax_class chain.
// Returned rows are the raw join output -- the usecase groups them by
// variant id for the resolver's per-line LineInput.TaxRates.
type VariantTaxResolver interface {
	RatesForVariants(ctx context.Context, variantIDs []string) ([]VariantTaxRate, error)
}

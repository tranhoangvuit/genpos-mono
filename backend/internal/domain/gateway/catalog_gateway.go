package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=catalog_gateway.go -destination=mock/mock_catalog_gateway.go -package=mock

// ----- Category --------------------------------------------------------------

type CreateCategoryParams struct {
	OrgID     string
	Name      string
	ParentID  string
	Color     string
	SortOrder int32
}

type UpdateCategoryParams struct {
	ID        string
	Name      string
	ParentID  string
	Color     string
	SortOrder int32
}

// CategoryReader lists and retrieves categories.
type CategoryReader interface {
	List(ctx context.Context) ([]*entity.Category, error)
	GetByID(ctx context.Context, id string) (*entity.Category, error)
	GetByName(ctx context.Context, name string) (*entity.Category, error)
}

// CategoryWriter mutates categories.
type CategoryWriter interface {
	Create(ctx context.Context, params CreateCategoryParams) (*entity.Category, error)
	Update(ctx context.Context, params UpdateCategoryParams) (*entity.Category, error)
	SoftDelete(ctx context.Context, id string) error
}

// ----- Product ---------------------------------------------------------------

type CreateProductBaseParams struct {
	OrgID       string
	Name        string
	Description string
	CategoryID  string
	IsActive    bool
	SortOrder   int32
}

type UpdateProductBaseParams struct {
	ID          string
	Name        string
	Description string
	CategoryID  string
	IsActive    bool
	SortOrder   int32
}

type CreateProductOptionParams struct {
	OrgID     string
	ProductID string
	Name      string
	SortOrder int32
}

type CreateProductOptionValueParams struct {
	OrgID     string
	OptionID  string
	Value     string
	SortOrder int32
}

type CreateProductVariantParams struct {
	OrgID      string
	ProductID  string
	Name       string
	SKU        string
	Barcode    string
	Price      string
	CostPrice  string
	TrackStock bool
	IsActive   bool
	SortOrder  int32
	TaxClassID string // optional -- empty = no automatic tax resolution
}

type CreateProductImageParams struct {
	OrgID     string
	ProductID string
	VariantID string
	URL       string
	SortOrder int32
}

// ProductWriter mutates products and their nested entities. All calls assume
// a tenant-scoped transaction is already active on the context.
type ProductWriter interface {
	CreateBase(ctx context.Context, params CreateProductBaseParams) (*entity.ProductDetail, error)
	UpdateBase(ctx context.Context, params UpdateProductBaseParams) (*entity.ProductDetail, error)
	SoftDelete(ctx context.Context, id string) error

	InsertOption(ctx context.Context, params CreateProductOptionParams) (*entity.ProductOption, error)
	InsertOptionValue(ctx context.Context, params CreateProductOptionValueParams) (*entity.ProductOptionValue, error)
	InsertVariant(ctx context.Context, params CreateProductVariantParams) (*entity.ProductVariant, error)
	InsertVariantOptionValue(ctx context.Context, orgID, variantID, optionValueID string) error
	InsertImage(ctx context.Context, params CreateProductImageParams) (*entity.ProductImage, error)

	DeleteOptionsByProduct(ctx context.Context, productID string) error
	SoftDeleteVariantsByProduct(ctx context.Context, productID string) error
	DeleteImagesByProduct(ctx context.Context, productID string) error
}

// ProductDetailReader loads product rows (full graph or summary list).
type ProductDetailReader interface {
	GetByID(ctx context.Context, id string) (*entity.ProductDetail, error)
	GetByName(ctx context.Context, name string) (*entity.ProductDetail, error)
	ListSummaries(ctx context.Context) ([]*entity.ProductListItem, error)
}

package entity

import "time"

// Category represents a product category.
type Category struct {
	ID        string
	OrgID     string
	ParentID  string
	Name      string
	SortOrder int32
	Color     string
	ImageURL  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ProductOption represents an option axis (e.g. "Size") for a product with variants.
type ProductOption struct {
	ID        string
	Name      string
	SortOrder int32
	Values    []*ProductOptionValue
}

// ProductOptionValue is a concrete value on an option (e.g. "Medium").
type ProductOptionValue struct {
	ID        string
	OptionID  string
	Value     string
	SortOrder int32
}

// ProductVariant represents a sellable SKU under a product.
type ProductVariant struct {
	ID             string
	ProductID      string
	Name           string
	SKU            string
	Barcode        string
	Price          string // decimal string for NUMERIC(12,4)
	CostPrice      string
	TrackStock     bool
	IsActive       bool
	SortOrder      int32
	OptionValueIDs []string
	// TaxClassID is the optional FK that the cart engine uses to resolve
	// per-line tax rows at sale time. Empty string = no automatic tax.
	TaxClassID string
}

// ProductImage represents an image attached to a product or a specific variant.
type ProductImage struct {
	ID        string
	VariantID string
	URL       string
	SortOrder int32
}

// ProductDetail is a product with all nested collections loaded.
type ProductDetail struct {
	ID          string
	OrgID       string
	CategoryID  string
	Name        string
	Description string
	IsActive    bool
	SortOrder   int32
	Options     []*ProductOption
	Variants    []*ProductVariant
	Images      []*ProductImage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

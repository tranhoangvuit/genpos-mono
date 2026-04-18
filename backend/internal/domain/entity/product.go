package entity

// ProductListItem is the summary row used on the product list page.
type ProductListItem struct {
	ID           string
	Name         string
	CategoryID   string
	CategoryName string
	Price        string
	VariantCount int32
	IsActive     bool
}

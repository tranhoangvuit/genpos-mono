package entity

import "time"

// PurchaseOrder represents a PO with nested items.
type PurchaseOrder struct {
	ID           string
	OrgID        string
	StoreID      string
	UserID       string
	PONumber     string
	SupplierName string
	Status       string // draft|submitted|partial|received|cancelled
	Notes        string
	ExpectedAt   time.Time
	ReceivedAt   time.Time
	Items        []*PurchaseOrderItem
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// PurchaseOrderListItem is the summary row for the PO list page.
type PurchaseOrderListItem struct {
	ID           string
	PONumber     string
	SupplierName string
	Status       string
	StoreName    string
	ExpectedAt   time.Time
	ItemCount    int32
	Total        string // decimal string
	CreatedAt    time.Time
}

// PurchaseOrderItem is one line on a PO.
type PurchaseOrderItem struct {
	ID               string
	VariantID        string
	VariantName      string // denormalized
	ProductName      string // denormalized
	QuantityOrdered  string // decimal string
	QuantityReceived string
	CostPrice        string
}

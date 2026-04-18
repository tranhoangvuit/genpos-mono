package entity

import "time"

// StockTake represents an inventory count session.
type StockTake struct {
	ID          string
	OrgID       string
	StoreID     string
	UserID      string
	Status      string // in_progress|completed|cancelled
	Notes       string
	CompletedAt time.Time
	Items       []*StockTakeItem
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// StockTakeListItem is the summary row for the stock takes list page.
type StockTakeListItem struct {
	ID            string
	StoreName     string
	Status        string
	ItemCount     int32
	VarianceLines int32
	CreatedAt     time.Time
	CompletedAt   time.Time
}

// VariantPickerItem is a lightweight variant shape for the PO item picker.
type VariantPickerItem struct {
	ID          string
	ProductName string
	VariantName string
	SKU         string
	Price       string
	CostPrice   string
}

// StoreRef is a store {id, name} tuple for selectors.
type StoreRef struct {
	ID   string
	Name string
}

// StockTakeItem is one counted line on a stock take.
type StockTakeItem struct {
	ID          string
	VariantID   string
	VariantName string
	ProductName string
	ExpectedQty string // decimal string
	CountedQty  string
}

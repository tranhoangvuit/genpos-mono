package input

import "time"

// ListDailySalesInput filters orders by [DateFrom, DateTo) and optional StoreID.
type ListDailySalesInput struct {
	OrgID    string
	StoreID  string // empty = all stores
	DateFrom time.Time
	DateTo   time.Time
}

// GetOrderInput loads one order with line items + payments.
type GetOrderInput struct {
	OrgID string
	ID    string
}

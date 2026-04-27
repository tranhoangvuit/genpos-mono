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

// CreateOrderLineItemInput is one line item on a CreateOrder request.
type CreateOrderLineItemInput struct {
	VariantID      string
	ProductName    string
	VariantName    string
	SKU            string
	Quantity       string
	UnitPrice      string
	DiscountAmount string
	TaxRate        string
	TaxAmount      string
	LineTotal      string
	Notes          string
}

// CreateOrderPaymentInput is one payment row on a CreateOrder request.
type CreateOrderPaymentInput struct {
	PaymentMethodID string
	Amount          string
	Tendered        string
	ChangeAmount    string
	Reference       string
}

// CreateOrderInput is the usecase input for OrderUsecase.CreateOrder. The
// (Source, ExternalID) pair acts as the idempotency key.
type CreateOrderInput struct {
	OrgID            string
	Source           string
	ExternalID       string
	ExternalSourceID string

	OrderNumber   string
	StoreID       string // optional — usecase falls back to the org's first store
	RegisterID    string
	CustomerID    string
	UserID        string
	Status        string
	Subtotal      string
	TaxTotal      string
	DiscountTotal string
	Total         string
	Notes         string
	CompletedAt   time.Time

	LineItems []CreateOrderLineItemInput
	Payments  []CreateOrderPaymentInput
}

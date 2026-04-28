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

// CreateOrderLineItemTaxInput is one snapshot tax row attached to a line.
type CreateOrderLineItemTaxInput struct {
	Sequence     int32
	TaxRateID    string
	NameSnapshot string
	RateSnapshot string
	IsInclusive  bool
	IsCompound   bool
	TaxableBase  string
	Amount       string
}

// CreateOrderLineAdjustmentInput is one line-level discount/fee/comp.
type CreateOrderLineAdjustmentInput struct {
	Sequence           int32
	Kind               string
	SourceType         string
	SourceID           string
	SourceCodeSnapshot string
	NameSnapshot       string
	Reason             string
	CalculationType    string
	CalculationValue   string
	Amount             string
	AppliesBeforeTax   bool
	AppliedBy          string
	ApprovedBy         string
}

// CreateOrderAdjustmentInput is one order-level adjustment.
type CreateOrderAdjustmentInput struct {
	Sequence           int32
	Kind               string
	SourceType         string
	SourceID           string
	SourceCodeSnapshot string
	NameSnapshot       string
	Reason             string
	CalculationType    string
	CalculationValue   string
	Amount             string
	AppliesBeforeTax   bool
	ProrateStrategy    string
	AppliedBy          string
	ApprovedBy         string
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
	Taxes          []CreateOrderLineItemTaxInput
	Adjustments    []CreateOrderLineAdjustmentInput
}

// CreateOrderPaymentInput is one payment row on a CreateOrder request.
type CreateOrderPaymentInput struct {
	PaymentMethodID string
	Amount          string
	Tendered        string
	ChangeAmount    string
	Reference       string
}

// ComputeOrderLineInput is one cart line for OrderUsecase.ComputeOrder.
// VariantID is required -- the engine resolves the tax_class through it.
// HasInclusiveOverride distinguishes "caller did not set the field" (false)
// from "caller forced exclusive treatment" (true with InclusiveOverride =
// false), since proto bool defaults to false.
type ComputeOrderLineInput struct {
	VariantID            string
	Quantity             string
	UnitPrice            string
	HasInclusiveOverride bool
	InclusiveOverride    bool
	Adjustments          []CreateOrderLineAdjustmentInput
}

// ComputeOrderInput is the usecase input for OrderUsecase.ComputeOrder.
// Stateless -- nothing is persisted.
type ComputeOrderInput struct {
	OrgID       string
	Lines       []ComputeOrderLineInput
	Adjustments []CreateOrderAdjustmentInput
	Round       string // "per_line" (default) or "per_order"
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

	LineItems   []CreateOrderLineItemInput
	Payments    []CreateOrderPaymentInput
	Adjustments []CreateOrderAdjustmentInput
}

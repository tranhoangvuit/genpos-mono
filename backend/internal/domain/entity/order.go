package entity

import "time"

// OrderSummary is the list row for the daily sales report.
type OrderSummary struct {
	ID            string
	OrderNumber   string
	Status        string // open|completed|voided|refunded|partially_refunded
	Subtotal      string // decimal string
	TaxTotal      string
	DiscountTotal string
	Total         string
	StoreID       string
	StoreName     string
	RegisterID    string
	UserID        string
	UserName      string
	CustomerID    string
	CustomerName  string
	CreatedAt     time.Time
	Source        string // pos|online_store|shopify|woocommerce|manual|import
	ExternalID    string
}

// Order is the detail view with nested items + payments.
type Order struct {
	ID            string
	OrderNumber   string
	Status        string
	Subtotal      string
	TaxTotal      string
	DiscountTotal string
	Total         string
	Notes         string
	StoreID       string
	StoreName     string
	RegisterID    string
	UserID        string
	UserName      string
	CustomerID    string
	CustomerName  string
	CreatedAt     time.Time
	CompletedAt   time.Time
	Source        string
	ExternalID    string
	LineItems     []*OrderLineItem
	Payments      []*OrderPayment
	Adjustments   []*OrderAdjustment
}

type OrderLineItem struct {
	ID             string
	VariantID      string
	ProductName    string
	VariantName    string
	SKU            string
	Quantity       string
	UnitPrice      string
	TaxRate        string
	TaxAmount      string
	DiscountAmount string
	LineTotal      string
	Notes          string
	Taxes          []*OrderLineItemTax
	Adjustments    []*OrderLineAdjustment
}

// OrderLineItemTax is the per-tax snapshot row for one line item. Snapshot
// fields freeze the rate's identity so editing the source rate later does
// not change historical orders.
type OrderLineItemTax struct {
	ID           string
	Sequence     int32
	TaxRateID    string
	NameSnapshot string
	RateSnapshot string
	IsInclusive  bool
	IsCompound   bool
	TaxableBase  string
	Amount       string
}

// OrderLineAdjustment captures a discount, fee, comp, or service charge
// applied to a single line item.
type OrderLineAdjustment struct {
	ID                 string
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
	AppliedAt          time.Time
	ApprovedBy         string
}

// OrderAdjustment is the order-level equivalent: order discount, delivery
// fee, tip, system rounding. ProrateStrategy controls how the engine
// distributes its impact across lines for tax purposes.
type OrderAdjustment struct {
	ID                 string
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
	AppliedAt          time.Time
	ApprovedBy         string
}

type OrderPayment struct {
	ID                string
	PaymentMethodID   string
	PaymentMethodName string
	Amount            string
	Tendered          string
	ChangeAmount      string
	Reference         string
	Status            string
	CreatedAt         time.Time
}

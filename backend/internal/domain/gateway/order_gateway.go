package gateway

import (
	"context"
	"time"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=order_gateway.go -destination=mock/mock_order_gateway.go -package=mock

type ListOrdersParams struct {
	DateFrom time.Time
	DateTo   time.Time
	StoreID  string // empty = all stores
}

// OrderReader loads orders for reporting.
type OrderReader interface {
	ListByDateRange(ctx context.Context, params ListOrdersParams) ([]*entity.OrderSummary, error)
	GetByID(ctx context.Context, id string) (*entity.Order, error)
	GetByExternalID(ctx context.Context, source, externalID string) (*entity.Order, error)
	ListLineItems(ctx context.Context, orderID string) ([]*entity.OrderLineItem, error)
	ListPayments(ctx context.Context, orderID string) ([]*entity.OrderPayment, error)
	ListOrderAdjustments(ctx context.Context, orderID string) ([]*entity.OrderAdjustment, error)
}

// OrderLineTaxParams is one snapshot tax row attached to a line on Create.
type OrderLineTaxParams struct {
	Sequence     int32
	TaxRateID    string // optional -- snapshot survives if rate is deleted
	NameSnapshot string
	RateSnapshot string // decimal fraction string (e.g. "0.1000")
	IsInclusive  bool
	IsCompound   bool
	TaxableBase  string
	Amount       string
}

// OrderLineAdjustmentParams is one line-level adjustment on Create.
type OrderLineAdjustmentParams struct {
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

// OrderAdjustmentParams is one order-level adjustment on Create.
type OrderAdjustmentParams struct {
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

// CreateOrderLineItemParams is the per-line-item write payload for Create.
// Taxes and Adjustments are optional: empty slices preserve the legacy
// aggregate-only contract (existing desk POS uploads); when populated the
// children are persisted alongside the aggregates the caller supplied.
type CreateOrderLineItemParams struct {
	VariantID      string // optional — empty when variant has been deleted upstream
	ProductName    string
	VariantName    string
	SKU            string
	Quantity       string // decimal string
	UnitPrice      string
	DiscountAmount string
	TaxRate        string
	TaxAmount      string
	LineTotal      string
	Notes          string
	Taxes          []OrderLineTaxParams
	Adjustments    []OrderLineAdjustmentParams
}

// CreateOrderPaymentParams is the per-payment write payload for Create.
type CreateOrderPaymentParams struct {
	PaymentMethodID string
	Amount          string
	Tendered        string
	ChangeAmount    string
	Reference       string
}

// CreateOrderParams is the Create payload.
type CreateOrderParams struct {
	OrgID            string
	Source           string // pos | online_store | shopify | woocommerce | manual | import
	ExternalID       string // client-provided idempotency key (e.g., desk's local order UUID)
	ExternalSourceID string

	OrderNumber   string
	StoreID       string
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

	LineItems   []CreateOrderLineItemParams
	Payments    []CreateOrderPaymentParams
	Adjustments []OrderAdjustmentParams
}

// OrderWriter persists orders coming from external channels (POS, online
// stores, importers). Implementations must enforce idempotency via
// (source, external_id): a re-submission returns the previously persisted row.
type OrderWriter interface {
	Create(ctx context.Context, params CreateOrderParams) (*entity.Order, error)
}

// OrgStoreReader resolves the default store for an org (used by CreateOrder
// when the client doesn't supply store_id — the desk POS today is single-store
// per device and doesn't track which cloud store row it represents).
type OrgStoreReader interface {
	FirstStoreID(ctx context.Context, orgID string) (string, error)
}

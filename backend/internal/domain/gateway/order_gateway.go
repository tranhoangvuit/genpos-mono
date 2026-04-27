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
}

// CreateOrderLineItemParams is the per-line-item write payload for Create.
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

	LineItems []CreateOrderLineItemParams
	Payments  []CreateOrderPaymentParams
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

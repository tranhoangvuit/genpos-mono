package gateway

import (
	"context"
	"time"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=purchase_order_gateway.go -destination=mock/mock_purchase_order_gateway.go -package=mock

type CreatePurchaseOrderParams struct {
	OrgID        string
	StoreID      string
	UserID       string // optional
	PONumber     string
	SupplierName string
	Notes        string
	ExpectedAt   time.Time // zero means null
}

type UpdatePurchaseOrderParams struct {
	ID           string
	StoreID      string
	SupplierName string
	Notes        string
	ExpectedAt   time.Time
}

type CreatePurchaseOrderItemParams struct {
	OrgID           string
	PurchaseOrderID string
	VariantID       string
	QuantityOrdered string
	CostPrice       string
}

// PurchaseOrderReader loads POs and their items.
type PurchaseOrderReader interface {
	GetByID(ctx context.Context, id string) (*entity.PurchaseOrder, error)
	CountForPrefix(ctx context.Context, orgID, prefix string) (int, error)
	ListItems(ctx context.Context, poID string) ([]*entity.PurchaseOrderItem, error)
	GetItemByID(ctx context.Context, id string) (*entity.PurchaseOrderItem, error)
	ListSummaries(ctx context.Context) ([]*entity.PurchaseOrderListItem, error)
	ListStoreRefs(ctx context.Context) ([]*entity.StoreRef, error)
	ListVariantPickerItems(ctx context.Context) ([]*entity.VariantPickerItem, error)
}

// PurchaseOrderWriter mutates POs and their items.
type PurchaseOrderWriter interface {
	Create(ctx context.Context, params CreatePurchaseOrderParams) (*entity.PurchaseOrder, error)
	Update(ctx context.Context, params UpdatePurchaseOrderParams) (*entity.PurchaseOrder, error)
	UpdateStatus(ctx context.Context, id, status string, receivedAt time.Time) error
	SoftDelete(ctx context.Context, id string) error

	InsertItem(ctx context.Context, params CreatePurchaseOrderItemParams) (*entity.PurchaseOrderItem, error)
	DeleteItemsByPO(ctx context.Context, poID string) error
	AddItemReceived(ctx context.Context, itemID, delta string) error
}

// StockMovementWriter records entries in the inventory ledger.
type StockMovementWriter interface {
	Insert(ctx context.Context, params CreateStockMovementParams) error
}

type CreateStockMovementParams struct {
	OrgID         string
	StoreID       string
	VariantID     string
	Direction     string // in|out
	Quantity      string
	MovementType  string // purchase|stock_in|sale|refund|adjustment|stock_take|transfer_in|transfer_out
	ReferenceType string
	ReferenceID   string
	UserID        string
	Notes         string
}

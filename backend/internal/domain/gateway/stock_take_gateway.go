package gateway

import (
	"context"
	"time"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=stock_take_gateway.go -destination=mock/mock_stock_take_gateway.go -package=mock

type CreateStockTakeParams struct {
	OrgID   string
	StoreID string
	UserID  string
	Notes   string
}

type CreateStockTakeItemParams struct {
	OrgID       string
	StockTakeID string
	VariantID   string
	ExpectedQty string
	CountedQty  string
}

// StockTakeReader loads stock takes and their items.
type StockTakeReader interface {
	GetByID(ctx context.Context, id string) (*entity.StockTake, error)
	ListItems(ctx context.Context, stockTakeID string) ([]*entity.StockTakeItem, error)
	ListSummaries(ctx context.Context) ([]*entity.StockTakeListItem, error)
}

// StockTakeWriter mutates stock takes and their items.
type StockTakeWriter interface {
	Create(ctx context.Context, params CreateStockTakeParams) (*entity.StockTake, error)
	SeedItemsFromOnHand(ctx context.Context, orgID, stockTakeID, storeID string) error
	InsertItem(ctx context.Context, params CreateStockTakeItemParams) (*entity.StockTakeItem, error)
	UpdateItemCount(ctx context.Context, itemID, countedQty string) error
	UpdateNotes(ctx context.Context, id, notes string) error
	UpdateStatus(ctx context.Context, id, status string, completedAt time.Time) error
	SoftDelete(ctx context.Context, id string) error
}

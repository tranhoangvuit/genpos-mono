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
	ListLineItems(ctx context.Context, orderID string) ([]*entity.OrderLineItem, error)
	ListPayments(ctx context.Context, orderID string) ([]*entity.OrderPayment, error)
}

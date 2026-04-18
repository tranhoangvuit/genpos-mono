package datastore

import (
	"context"
	stderrors "errors"

	"github.com/jackc/pgx/v5"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type orderStore struct{}

// NewOrderReader returns an OrderReader backed by sqlc.
func NewOrderReader() gateway.OrderReader { return &orderStore{} }

func (s *orderStore) ListByDateRange(ctx context.Context, params gateway.ListOrdersParams) ([]*entity.OrderSummary, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	storeID, err := uuidOrNull(params.StoreID)
	if err != nil {
		return nil, errors.BadRequest("invalid store id")
	}
	rows, err := sqlc.New(dbtx).ListOrdersByDateRange(ctx, sqlc.ListOrdersByDateRangeParams{
		DateFrom: timestampOrNull(params.DateFrom),
		DateTo:   timestampOrNull(params.DateTo),
		StoreID:  storeID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "list orders by date range")
	}
	out := make([]*entity.OrderSummary, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.OrderSummary{
			ID:            uuidString(r.ID),
			OrderNumber:   r.OrderNumber,
			Status:        r.Status,
			Subtotal:      r.Subtotal,
			TaxTotal:      r.TaxTotal,
			DiscountTotal: r.DiscountTotal,
			Total:         r.Total,
			StoreID:       uuidString(r.StoreID),
			StoreName:     r.StoreName,
			RegisterID:    uuidString(r.RegisterID),
			UserID:        uuidString(r.UserID),
			UserName:      r.UserName,
			CustomerID:    uuidString(r.CustomerID),
			CustomerName:  r.CustomerName,
			CreatedAt:     r.CreatedAt.Time,
		})
	}
	return out, nil
}

func (s *orderStore) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid order id")
	}
	r, err := sqlc.New(dbtx).GetOrderByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("order not found")
		}
		return nil, errors.Wrap(err, "get order by id")
	}
	return &entity.Order{
		ID:            uuidString(r.ID),
		OrderNumber:   r.OrderNumber,
		Status:        r.Status,
		Subtotal:      r.Subtotal,
		TaxTotal:      r.TaxTotal,
		DiscountTotal: r.DiscountTotal,
		Total:         r.Total,
		Notes:         r.Notes,
		StoreID:       uuidString(r.StoreID),
		StoreName:     r.StoreName,
		RegisterID:    uuidString(r.RegisterID),
		UserID:        uuidString(r.UserID),
		UserName:      r.UserName,
		CustomerID:    uuidString(r.CustomerID),
		CustomerName:  r.CustomerName,
		CreatedAt:     r.CreatedAt.Time,
		CompletedAt:   timestampTime(r.CompletedAt),
	}, nil
}

func (s *orderStore) ListLineItems(ctx context.Context, orderID string) ([]*entity.OrderLineItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(orderID)
	if err != nil {
		return nil, errors.BadRequest("invalid order id")
	}
	rows, err := sqlc.New(dbtx).ListOrderLineItems(ctx, uid)
	if err != nil {
		return nil, errors.Wrap(err, "list order line items")
	}
	out := make([]*entity.OrderLineItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.OrderLineItem{
			ID:             uuidString(r.ID),
			VariantID:      uuidString(r.VariantID),
			ProductName:    r.ProductName,
			VariantName:    r.VariantName,
			SKU:            r.Sku,
			Quantity:       r.Quantity,
			UnitPrice:      r.UnitPrice,
			TaxRate:        r.TaxRate,
			TaxAmount:      r.TaxAmount,
			DiscountAmount: r.DiscountAmount,
			LineTotal:      r.LineTotal,
			Notes:          r.Notes,
		})
	}
	return out, nil
}

func (s *orderStore) ListPayments(ctx context.Context, orderID string) ([]*entity.OrderPayment, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(orderID)
	if err != nil {
		return nil, errors.BadRequest("invalid order id")
	}
	rows, err := sqlc.New(dbtx).ListOrderPayments(ctx, uid)
	if err != nil {
		return nil, errors.Wrap(err, "list order payments")
	}
	out := make([]*entity.OrderPayment, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.OrderPayment{
			ID:                uuidString(r.ID),
			PaymentMethodID:   uuidString(r.PaymentMethodID),
			PaymentMethodName: r.PaymentMethodName,
			Amount:            r.Amount,
			Tendered:          r.Tendered,
			ChangeAmount:      r.ChangeAmount,
			Reference:         r.Reference,
			Status:            r.Status,
			CreatedAt:         r.CreatedAt.Time,
		})
	}
	return out, nil
}

package usecase

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type orderUsecase struct {
	tenantDB gateway.TenantDB
	reader   gateway.OrderReader
}

// NewOrderUsecase constructs an OrderUsecase.
func NewOrderUsecase(tenantDB gateway.TenantDB, reader gateway.OrderReader) OrderUsecase {
	return &orderUsecase{tenantDB: tenantDB, reader: reader}
}

func (u *orderUsecase) GetOrder(ctx context.Context, in input.GetOrderInput) (*entity.Order, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	var out *entity.Order
	if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		o, err := u.reader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		items, err := u.reader.ListLineItems(ctx, in.ID)
		if err != nil {
			return err
		}
		payments, err := u.reader.ListPayments(ctx, in.ID)
		if err != nil {
			return err
		}
		o.LineItems = items
		o.Payments = payments
		out = o
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "get order")
	}
	return out, nil
}

func (u *orderUsecase) ListOrders(ctx context.Context, in input.ListDailySalesInput) ([]*entity.OrderSummary, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if in.DateFrom.IsZero() || in.DateTo.IsZero() {
		return nil, errors.BadRequest("date range is required")
	}
	if !in.DateFrom.Before(in.DateTo) {
		return nil, errors.BadRequest("date_from must be before date_to")
	}
	var out []*entity.OrderSummary
	if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		v, err := u.reader.ListByDateRange(ctx, gateway.ListOrdersParams{
			DateFrom: in.DateFrom,
			DateTo:   in.DateTo,
			StoreID:  in.StoreID,
		})
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list orders")
	}
	return out, nil
}

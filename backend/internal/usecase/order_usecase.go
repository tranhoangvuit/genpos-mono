package usecase

import (
	"context"
	"time"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type orderUsecase struct {
	tenantDB    gateway.TenantDB
	reader      gateway.OrderReader
	writer      gateway.OrderWriter
	storeReader gateway.OrgStoreReader
}

// NewOrderUsecase constructs an OrderUsecase.
func NewOrderUsecase(
	tenantDB gateway.TenantDB,
	reader gateway.OrderReader,
	writer gateway.OrderWriter,
	storeReader gateway.OrgStoreReader,
) OrderUsecase {
	return &orderUsecase{
		tenantDB:    tenantDB,
		reader:      reader,
		writer:      writer,
		storeReader: storeReader,
	}
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

func (u *orderUsecase) CreateOrder(ctx context.Context, in input.CreateOrderInput) (*entity.Order, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if in.OrderNumber == "" {
		return nil, errors.BadRequest("order number is required")
	}
	if len(in.LineItems) == 0 {
		return nil, errors.BadRequest("at least one line item is required")
	}

	source := in.Source
	if source == "" {
		source = "pos"
	}
	if source == "pos" && in.UserID == "" {
		return nil, errors.BadRequest("user id is required for pos orders")
	}

	status := in.Status
	if status == "" {
		status = "completed"
	}
	completedAt := in.CompletedAt
	if status == "completed" && completedAt.IsZero() {
		completedAt = time.Now().UTC()
	}

	var out *entity.Order
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		// Idempotency: if a row already exists for (source, external_id), return
		// it without re-inserting. The desk client retries uploads on failure;
		// without this guard a flaky network would create duplicates.
		if in.ExternalID != "" {
			existing, gErr := u.reader.GetByExternalID(ctx, source, in.ExternalID)
			if gErr == nil {
				if hErr := u.hydrate(ctx, existing); hErr != nil {
					return hErr
				}
				out = existing
				return nil
			}
			if errors.GetCode(gErr) != errors.CodeNotFound {
				return gErr
			}
		}

		storeID := in.StoreID
		if storeID == "" {
			s, sErr := u.storeReader.FirstStoreID(ctx, in.OrgID)
			if sErr != nil {
				return sErr
			}
			storeID = s
		}

		params := gateway.CreateOrderParams{
			OrgID:            in.OrgID,
			Source:           source,
			ExternalID:       in.ExternalID,
			ExternalSourceID: in.ExternalSourceID,
			OrderNumber:      in.OrderNumber,
			StoreID:          storeID,
			RegisterID:       in.RegisterID,
			CustomerID:       in.CustomerID,
			UserID:           in.UserID,
			Status:           status,
			Subtotal:         in.Subtotal,
			TaxTotal:         in.TaxTotal,
			DiscountTotal:    in.DiscountTotal,
			Total:            in.Total,
			Notes:            in.Notes,
			CompletedAt:      completedAt,
			LineItems:        toGatewayLineItems(in.LineItems),
			Payments:         toGatewayPayments(in.Payments),
		}
		created, cErr := u.writer.Create(ctx, params)
		if cErr != nil {
			return cErr
		}
		if hErr := u.hydrate(ctx, created); hErr != nil {
			return hErr
		}
		out = created
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "create order")
	}
	return out, nil
}

func (u *orderUsecase) hydrate(ctx context.Context, o *entity.Order) error {
	items, err := u.reader.ListLineItems(ctx, o.ID)
	if err != nil {
		return err
	}
	payments, err := u.reader.ListPayments(ctx, o.ID)
	if err != nil {
		return err
	}
	o.LineItems = items
	o.Payments = payments
	return nil
}

func toGatewayLineItems(in []input.CreateOrderLineItemInput) []gateway.CreateOrderLineItemParams {
	out := make([]gateway.CreateOrderLineItemParams, len(in))
	for i, v := range in {
		out[i] = gateway.CreateOrderLineItemParams{
			VariantID:      v.VariantID,
			ProductName:    v.ProductName,
			VariantName:    v.VariantName,
			SKU:            v.SKU,
			Quantity:       v.Quantity,
			UnitPrice:      v.UnitPrice,
			DiscountAmount: v.DiscountAmount,
			TaxRate:        v.TaxRate,
			TaxAmount:      v.TaxAmount,
			LineTotal:      v.LineTotal,
			Notes:          v.Notes,
		}
	}
	return out
}

func toGatewayPayments(in []input.CreateOrderPaymentInput) []gateway.CreateOrderPaymentParams {
	out := make([]gateway.CreateOrderPaymentParams, len(in))
	for i, v := range in {
		out[i] = gateway.CreateOrderPaymentParams{
			PaymentMethodID: v.PaymentMethodID,
			Amount:          v.Amount,
			Tendered:        v.Tendered,
			ChangeAmount:    v.ChangeAmount,
			Reference:       v.Reference,
		}
	}
	return out
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

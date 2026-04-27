package grpc

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	genposv1 "github.com/genpick/genpos-mono/backend/gen/genpos/v1"
	"github.com/genpick/genpos-mono/backend/gen/genpos/v1/genposv1connect"
	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/handler/interceptor"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

// OrderHandler implements OrderServiceHandler.
type OrderHandler struct {
	genposv1connect.UnimplementedOrderServiceHandler
	logger  *slog.Logger
	usecase usecase.OrderUsecase
}

// NewOrderHandler constructs an OrderHandler.
func NewOrderHandler(logger *slog.Logger, uc usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{logger: logger, usecase: uc}
}

func (h *OrderHandler) ListOrders(
	ctx context.Context,
	req *connect.Request[genposv1.ListOrdersRequest],
) (*connect.Response[genposv1.ListOrdersResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	from := req.Msg.GetDateFrom()
	to := req.Msg.GetDateTo()
	if from == nil || to == nil {
		return nil, errors.ToConnectError(errors.BadRequest("date range is required"))
	}
	items, err := h.usecase.ListOrders(ctx, input.ListDailySalesInput{
		OrgID:    authCtx.OrgID,
		StoreID:  req.Msg.GetStoreId(),
		DateFrom: from.AsTime(),
		DateTo:   to.AsTime(),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "list orders", err)
	}
	pb := make([]*genposv1.OrderSummary, 0, len(items))
	for _, o := range items {
		pb = append(pb, toOrderSummaryProto(o))
	}
	return connect.NewResponse(&genposv1.ListOrdersResponse{Orders: pb}), nil
}

func (h *OrderHandler) CreateOrder(
	ctx context.Context,
	req *connect.Request[genposv1.CreateOrderRequest],
) (*connect.Response[genposv1.CreateOrderResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	msg := req.Msg

	completedAt := time.Time{}
	if ts := msg.GetCompletedAt(); ts != nil {
		completedAt = ts.AsTime()
	}

	items := make([]input.CreateOrderLineItemInput, 0, len(msg.GetLineItems()))
	for _, it := range msg.GetLineItems() {
		items = append(items, input.CreateOrderLineItemInput{
			VariantID:      it.GetVariantId(),
			ProductName:    it.GetProductName(),
			VariantName:    it.GetVariantName(),
			SKU:            it.GetSku(),
			Quantity:       it.GetQuantity(),
			UnitPrice:      it.GetUnitPrice(),
			DiscountAmount: it.GetDiscountAmount(),
			TaxRate:        it.GetTaxRate(),
			TaxAmount:      it.GetTaxAmount(),
			LineTotal:      it.GetLineTotal(),
			Notes:          it.GetNotes(),
		})
	}
	payments := make([]input.CreateOrderPaymentInput, 0, len(msg.GetPayments()))
	for _, p := range msg.GetPayments() {
		payments = append(payments, input.CreateOrderPaymentInput{
			PaymentMethodID: p.GetPaymentMethodId(),
			Amount:          p.GetAmount(),
			Tendered:        p.GetTendered(),
			ChangeAmount:    p.GetChangeAmount(),
			Reference:       p.GetReference(),
		})
	}

	o, err := h.usecase.CreateOrder(ctx, input.CreateOrderInput{
		OrgID:            authCtx.OrgID,
		Source:           msg.GetSource(),
		ExternalID:       msg.GetExternalId(),
		ExternalSourceID: msg.GetExternalSourceId(),
		OrderNumber:      msg.GetOrderNumber(),
		StoreID:          msg.GetStoreId(),
		RegisterID:       msg.GetRegisterId(),
		CustomerID:       msg.GetCustomerId(),
		UserID:           msg.GetUserId(),
		Status:           msg.GetStatus(),
		Subtotal:         msg.GetSubtotal(),
		TaxTotal:         msg.GetTaxTotal(),
		DiscountTotal:    msg.GetDiscountTotal(),
		Total:            msg.GetTotal(),
		Notes:            msg.GetNotes(),
		CompletedAt:      completedAt,
		LineItems:        items,
		Payments:         payments,
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "create order", err)
	}
	return connect.NewResponse(&genposv1.CreateOrderResponse{Order: toOrderProto(o)}), nil
}

func (h *OrderHandler) GetOrder(
	ctx context.Context,
	req *connect.Request[genposv1.GetOrderRequest],
) (*connect.Response[genposv1.GetOrderResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	o, err := h.usecase.GetOrder(ctx, input.GetOrderInput{
		OrgID: authCtx.OrgID,
		ID:    req.Msg.GetId(),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "get order", err)
	}
	return connect.NewResponse(&genposv1.GetOrderResponse{Order: toOrderProto(o)}), nil
}

func toOrderProto(o *entity.Order) *genposv1.Order {
	if o == nil {
		return nil
	}
	items := make([]*genposv1.OrderLineItem, 0, len(o.LineItems))
	for _, it := range o.LineItems {
		items = append(items, &genposv1.OrderLineItem{
			Id:             it.ID,
			VariantId:      it.VariantID,
			ProductName:    it.ProductName,
			VariantName:    it.VariantName,
			Sku:            it.SKU,
			Quantity:       it.Quantity,
			UnitPrice:      it.UnitPrice,
			TaxRate:        it.TaxRate,
			TaxAmount:      it.TaxAmount,
			DiscountAmount: it.DiscountAmount,
			LineTotal:      it.LineTotal,
			Notes:          it.Notes,
		})
	}
	payments := make([]*genposv1.OrderPayment, 0, len(o.Payments))
	for _, p := range o.Payments {
		payments = append(payments, &genposv1.OrderPayment{
			Id:                p.ID,
			PaymentMethodId:   p.PaymentMethodID,
			PaymentMethodName: p.PaymentMethodName,
			Amount:            p.Amount,
			Tendered:          p.Tendered,
			ChangeAmount:      p.ChangeAmount,
			Reference:         p.Reference,
			Status:            p.Status,
			CreatedAt:         timestamppb.New(p.CreatedAt),
		})
	}
	out := &genposv1.Order{
		Id:            o.ID,
		OrderNumber:   o.OrderNumber,
		Status:        o.Status,
		Subtotal:      o.Subtotal,
		TaxTotal:      o.TaxTotal,
		DiscountTotal: o.DiscountTotal,
		Total:         o.Total,
		Notes:         o.Notes,
		StoreId:       o.StoreID,
		StoreName:     o.StoreName,
		RegisterId:    o.RegisterID,
		UserId:        o.UserID,
		UserName:      o.UserName,
		CustomerId:    o.CustomerID,
		CustomerName:  o.CustomerName,
		CreatedAt:     timestamppb.New(o.CreatedAt),
		LineItems:     items,
		Payments:      payments,
		Source:        o.Source,
		ExternalId:    o.ExternalID,
	}
	if !o.CompletedAt.IsZero() {
		out.CompletedAt = timestamppb.New(o.CompletedAt)
	}
	return out
}

func toOrderSummaryProto(o *entity.OrderSummary) *genposv1.OrderSummary {
	if o == nil {
		return nil
	}
	return &genposv1.OrderSummary{
		Id:            o.ID,
		OrderNumber:   o.OrderNumber,
		Status:        o.Status,
		Subtotal:      o.Subtotal,
		TaxTotal:      o.TaxTotal,
		DiscountTotal: o.DiscountTotal,
		Total:         o.Total,
		StoreId:       o.StoreID,
		StoreName:     o.StoreName,
		RegisterId:    o.RegisterID,
		UserId:        o.UserID,
		UserName:      o.UserName,
		CustomerId:    o.CustomerID,
		CustomerName:  o.CustomerName,
		CreatedAt:     timestamppb.New(o.CreatedAt),
		Source:        o.Source,
		ExternalId:    o.ExternalID,
	}
}

func (h *OrderHandler) requireAuth(ctx context.Context) (*interceptor.AuthContext, error) {
	a := interceptor.FromContext(ctx)
	if a == nil {
		return nil, errors.ToConnectError(errors.Unauthorized("not signed in"))
	}
	return a, nil
}

func (h *OrderHandler) logAndConvert(ctx context.Context, op string, err error) error {
	if errors.ShouldLog(errors.GetCode(err)) {
		h.logger.ErrorContext(ctx, op+" failed", "error", err)
	}
	return errors.ToConnectError(err)
}

var _ genposv1connect.OrderServiceHandler = (*OrderHandler)(nil)

package grpc

import (
	"context"
	"log/slog"

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

// ----- SupplierHandler -----------------------------------------------------

// SupplierHandler implements SupplierServiceHandler.
type SupplierHandler struct {
	genposv1connect.UnimplementedSupplierServiceHandler
	logger  *slog.Logger
	usecase usecase.SupplierUsecase
}

// NewSupplierHandler constructs a SupplierHandler.
func NewSupplierHandler(logger *slog.Logger, uc usecase.SupplierUsecase) *SupplierHandler {
	return &SupplierHandler{logger: logger, usecase: uc}
}

func (h *SupplierHandler) ListSuppliers(
	ctx context.Context,
	_ *connect.Request[genposv1.ListSuppliersRequest],
) (*connect.Response[genposv1.ListSuppliersResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListSuppliers(ctx, authCtx.OrgID)
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "list suppliers", err)
	}
	pb := make([]*genposv1.Supplier, 0, len(items))
	for _, s := range items {
		pb = append(pb, toSupplierProto(s))
	}
	return connect.NewResponse(&genposv1.ListSuppliersResponse{Suppliers: pb}), nil
}

func (h *SupplierHandler) CreateSupplier(
	ctx context.Context,
	req *connect.Request[genposv1.CreateSupplierRequest],
) (*connect.Response[genposv1.CreateSupplierResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	s, err := h.usecase.CreateSupplier(ctx, input.CreateSupplierInput{
		OrgID:    authCtx.OrgID,
		Supplier: fromSupplierInputProto(req.Msg.GetSupplier()),
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "create supplier", err)
	}
	return connect.NewResponse(&genposv1.CreateSupplierResponse{Supplier: toSupplierProto(s)}), nil
}

func (h *SupplierHandler) UpdateSupplier(
	ctx context.Context,
	req *connect.Request[genposv1.UpdateSupplierRequest],
) (*connect.Response[genposv1.UpdateSupplierResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	s, err := h.usecase.UpdateSupplier(ctx, input.UpdateSupplierInput{
		ID:       req.Msg.GetId(),
		OrgID:    authCtx.OrgID,
		Supplier: fromSupplierInputProto(req.Msg.GetSupplier()),
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "update supplier", err)
	}
	return connect.NewResponse(&genposv1.UpdateSupplierResponse{Supplier: toSupplierProto(s)}), nil
}

func (h *SupplierHandler) DeleteSupplier(
	ctx context.Context,
	req *connect.Request[genposv1.DeleteSupplierRequest],
) (*connect.Response[genposv1.DeleteSupplierResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeleteSupplier(ctx, input.DeleteSupplierInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	}); err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "delete supplier", err)
	}
	return connect.NewResponse(&genposv1.DeleteSupplierResponse{}), nil
}

// ----- PurchaseOrderHandler ------------------------------------------------

// PurchaseOrderHandler implements PurchaseOrderServiceHandler.
type PurchaseOrderHandler struct {
	genposv1connect.UnimplementedPurchaseOrderServiceHandler
	logger  *slog.Logger
	usecase usecase.PurchaseOrderUsecase
}

// NewPurchaseOrderHandler constructs a PurchaseOrderHandler.
func NewPurchaseOrderHandler(logger *slog.Logger, uc usecase.PurchaseOrderUsecase) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{logger: logger, usecase: uc}
}

func (h *PurchaseOrderHandler) ListPurchaseOrders(
	ctx context.Context,
	_ *connect.Request[genposv1.ListPurchaseOrdersRequest],
) (*connect.Response[genposv1.ListPurchaseOrdersResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListPurchaseOrders(ctx, authCtx.OrgID)
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "list purchase orders", err)
	}
	pb := make([]*genposv1.PurchaseOrderListItem, 0, len(items))
	for _, p := range items {
		item := &genposv1.PurchaseOrderListItem{
			Id:           p.ID,
			PoNumber:     p.PONumber,
			SupplierName: p.SupplierName,
			Status:       p.Status,
			StoreName:    p.StoreName,
			ItemCount:    p.ItemCount,
			Total:        p.Total,
			CreatedAt:    timestamppb.New(p.CreatedAt),
		}
		if !p.ExpectedAt.IsZero() {
			item.ExpectedAt = timestamppb.New(p.ExpectedAt)
		}
		pb = append(pb, item)
	}
	return connect.NewResponse(&genposv1.ListPurchaseOrdersResponse{PurchaseOrders: pb}), nil
}

func (h *PurchaseOrderHandler) ListStores(
	ctx context.Context,
	_ *connect.Request[genposv1.ListStoresRequest],
) (*connect.Response[genposv1.ListStoresResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListStores(ctx, authCtx.OrgID)
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "list stores", err)
	}
	pb := make([]*genposv1.StoreRef, 0, len(items))
	for _, s := range items {
		pb = append(pb, &genposv1.StoreRef{Id: s.ID, Name: s.Name})
	}
	return connect.NewResponse(&genposv1.ListStoresResponse{Stores: pb}), nil
}

func (h *PurchaseOrderHandler) ListVariantsForPicker(
	ctx context.Context,
	_ *connect.Request[genposv1.ListVariantsForPickerRequest],
) (*connect.Response[genposv1.ListVariantsForPickerResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListVariantsForPicker(ctx, authCtx.OrgID)
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "list variants for picker", err)
	}
	pb := make([]*genposv1.VariantPickerItem, 0, len(items))
	for _, v := range items {
		pb = append(pb, &genposv1.VariantPickerItem{
			Id:          v.ID,
			ProductName: v.ProductName,
			VariantName: v.VariantName,
			Sku:         v.SKU,
			Price:       v.Price,
			CostPrice:   v.CostPrice,
		})
	}
	return connect.NewResponse(&genposv1.ListVariantsForPickerResponse{Variants: pb}), nil
}

func (h *PurchaseOrderHandler) GetPurchaseOrder(
	ctx context.Context,
	req *connect.Request[genposv1.GetPurchaseOrderRequest],
) (*connect.Response[genposv1.GetPurchaseOrderResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	po, err := h.usecase.GetPurchaseOrder(ctx, input.GetPurchaseOrderInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "get purchase order", err)
	}
	return connect.NewResponse(&genposv1.GetPurchaseOrderResponse{PurchaseOrder: toPurchaseOrderProto(po)}), nil
}

func (h *PurchaseOrderHandler) CreatePurchaseOrder(
	ctx context.Context,
	req *connect.Request[genposv1.CreatePurchaseOrderRequest],
) (*connect.Response[genposv1.CreatePurchaseOrderResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	po, err := h.usecase.CreatePurchaseOrder(ctx, input.CreatePurchaseOrderInput{
		OrgID:         authCtx.OrgID,
		UserID:        authCtx.UserID,
		PurchaseOrder: fromPurchaseOrderInputProto(req.Msg.GetPurchaseOrder()),
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "create purchase order", err)
	}
	return connect.NewResponse(&genposv1.CreatePurchaseOrderResponse{PurchaseOrder: toPurchaseOrderProto(po)}), nil
}

func (h *PurchaseOrderHandler) UpdatePurchaseOrder(
	ctx context.Context,
	req *connect.Request[genposv1.UpdatePurchaseOrderRequest],
) (*connect.Response[genposv1.UpdatePurchaseOrderResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	po, err := h.usecase.UpdatePurchaseOrder(ctx, input.UpdatePurchaseOrderInput{
		ID:            req.Msg.GetId(),
		OrgID:         authCtx.OrgID,
		PurchaseOrder: fromPurchaseOrderInputProto(req.Msg.GetPurchaseOrder()),
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "update purchase order", err)
	}
	return connect.NewResponse(&genposv1.UpdatePurchaseOrderResponse{PurchaseOrder: toPurchaseOrderProto(po)}), nil
}

func (h *PurchaseOrderHandler) SubmitPurchaseOrder(
	ctx context.Context,
	req *connect.Request[genposv1.SubmitPurchaseOrderRequest],
) (*connect.Response[genposv1.SubmitPurchaseOrderResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	po, err := h.usecase.SubmitPurchaseOrder(ctx, input.SubmitPurchaseOrderInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "submit purchase order", err)
	}
	return connect.NewResponse(&genposv1.SubmitPurchaseOrderResponse{PurchaseOrder: toPurchaseOrderProto(po)}), nil
}

func (h *PurchaseOrderHandler) ReceivePurchaseOrder(
	ctx context.Context,
	req *connect.Request[genposv1.ReceivePurchaseOrderRequest],
) (*connect.Response[genposv1.ReceivePurchaseOrderResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	lines := make([]input.ReceivePurchaseOrderLineInput, 0, len(req.Msg.GetLines()))
	for _, ln := range req.Msg.GetLines() {
		lines = append(lines, input.ReceivePurchaseOrderLineInput{
			ItemID:            ln.GetItemId(),
			QuantityToReceive: ln.GetQuantityToReceive(),
		})
	}
	po, err := h.usecase.ReceivePurchaseOrder(ctx, input.ReceivePurchaseOrderInput{
		ID:     req.Msg.GetId(),
		OrgID:  authCtx.OrgID,
		UserID: authCtx.UserID,
		Lines:  lines,
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "receive purchase order", err)
	}
	return connect.NewResponse(&genposv1.ReceivePurchaseOrderResponse{PurchaseOrder: toPurchaseOrderProto(po)}), nil
}

func (h *PurchaseOrderHandler) CancelPurchaseOrder(
	ctx context.Context,
	req *connect.Request[genposv1.CancelPurchaseOrderRequest],
) (*connect.Response[genposv1.CancelPurchaseOrderResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	po, err := h.usecase.CancelPurchaseOrder(ctx, input.CancelPurchaseOrderInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "cancel purchase order", err)
	}
	return connect.NewResponse(&genposv1.CancelPurchaseOrderResponse{PurchaseOrder: toPurchaseOrderProto(po)}), nil
}

func (h *PurchaseOrderHandler) DeletePurchaseOrder(
	ctx context.Context,
	req *connect.Request[genposv1.DeletePurchaseOrderRequest],
) (*connect.Response[genposv1.DeletePurchaseOrderResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeletePurchaseOrder(ctx, input.DeletePurchaseOrderInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	}); err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "delete purchase order", err)
	}
	return connect.NewResponse(&genposv1.DeletePurchaseOrderResponse{}), nil
}

// ----- helpers -------------------------------------------------------------

func requireInventoryAuth(ctx context.Context) (*interceptor.AuthContext, error) {
	a := interceptor.FromContext(ctx)
	if a == nil {
		return nil, errors.ToConnectError(errors.Unauthorized("not signed in"))
	}
	return a, nil
}

func logAndConvertInventory(logger *slog.Logger, ctx context.Context, op string, err error) error {
	if errors.ShouldLog(errors.GetCode(err)) {
		logger.ErrorContext(ctx, op+" failed", "error", err)
	}
	return errors.ToConnectError(err)
}

func toSupplierProto(s *entity.Supplier) *genposv1.Supplier {
	if s == nil {
		return nil
	}
	return &genposv1.Supplier{
		Id:          s.ID,
		OrgId:       s.OrgID,
		Name:        s.Name,
		ContactName: s.ContactName,
		Email:       s.Email,
		Phone:       s.Phone,
		Address:     s.Address,
		Notes:       s.Notes,
		CreatedAt:   timestamppb.New(s.CreatedAt),
		UpdatedAt:   timestamppb.New(s.UpdatedAt),
	}
}

func fromSupplierInputProto(s *genposv1.SupplierInput) input.SupplierInput {
	if s == nil {
		return input.SupplierInput{}
	}
	return input.SupplierInput{
		Name:        s.GetName(),
		ContactName: s.GetContactName(),
		Email:       s.GetEmail(),
		Phone:       s.GetPhone(),
		Address:     s.GetAddress(),
		Notes:       s.GetNotes(),
	}
}

func toPurchaseOrderProto(po *entity.PurchaseOrder) *genposv1.PurchaseOrder {
	if po == nil {
		return nil
	}
	items := make([]*genposv1.PurchaseOrderItem, 0, len(po.Items))
	for _, it := range po.Items {
		items = append(items, &genposv1.PurchaseOrderItem{
			Id:               it.ID,
			VariantId:        it.VariantID,
			VariantName:      it.VariantName,
			ProductName:      it.ProductName,
			QuantityOrdered:  it.QuantityOrdered,
			QuantityReceived: it.QuantityReceived,
			CostPrice:        it.CostPrice,
		})
	}
	out := &genposv1.PurchaseOrder{
		Id:           po.ID,
		OrgId:        po.OrgID,
		StoreId:      po.StoreID,
		PoNumber:     po.PONumber,
		SupplierName: po.SupplierName,
		Status:       po.Status,
		Notes:        po.Notes,
		Items:        items,
		CreatedAt:    timestamppb.New(po.CreatedAt),
		UpdatedAt:    timestamppb.New(po.UpdatedAt),
	}
	if !po.ExpectedAt.IsZero() {
		out.ExpectedAt = timestamppb.New(po.ExpectedAt)
	}
	if !po.ReceivedAt.IsZero() {
		out.ReceivedAt = timestamppb.New(po.ReceivedAt)
	}
	return out
}

func fromPurchaseOrderInputProto(p *genposv1.PurchaseOrderInput) input.PurchaseOrderInput {
	if p == nil {
		return input.PurchaseOrderInput{}
	}
	items := make([]input.PurchaseOrderItemInput, 0, len(p.GetItems()))
	for _, it := range p.GetItems() {
		items = append(items, input.PurchaseOrderItemInput{
			VariantID:       it.GetVariantId(),
			QuantityOrdered: it.GetQuantityOrdered(),
			CostPrice:       it.GetCostPrice(),
		})
	}
	out := input.PurchaseOrderInput{
		StoreID:      p.GetStoreId(),
		SupplierName: p.GetSupplierName(),
		Notes:        p.GetNotes(),
		Items:        items,
	}
	if p.ExpectedAt != nil {
		out.ExpectedAt = p.ExpectedAt.AsTime()
	}
	return out
}

var (
	_ genposv1connect.SupplierServiceHandler      = (*SupplierHandler)(nil)
	_ genposv1connect.PurchaseOrderServiceHandler = (*PurchaseOrderHandler)(nil)
)

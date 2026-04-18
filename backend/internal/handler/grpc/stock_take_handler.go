package grpc

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	genposv1 "github.com/genpick/genpos-mono/backend/gen/genpos/v1"
	"github.com/genpick/genpos-mono/backend/gen/genpos/v1/genposv1connect"
	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
)

// StockTakeHandler implements StockTakeServiceHandler.
type StockTakeHandler struct {
	genposv1connect.UnimplementedStockTakeServiceHandler
	logger  *slog.Logger
	usecase usecase.StockTakeUsecase
}

// NewStockTakeHandler constructs a StockTakeHandler.
func NewStockTakeHandler(logger *slog.Logger, uc usecase.StockTakeUsecase) *StockTakeHandler {
	return &StockTakeHandler{logger: logger, usecase: uc}
}

func (h *StockTakeHandler) ListStockTakes(
	ctx context.Context,
	_ *connect.Request[genposv1.ListStockTakesRequest],
) (*connect.Response[genposv1.ListStockTakesResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListStockTakes(ctx, authCtx.OrgID)
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "list stock takes", err)
	}
	pb := make([]*genposv1.StockTakeListItem, 0, len(items))
	for _, s := range items {
		item := &genposv1.StockTakeListItem{
			Id:            s.ID,
			StoreName:     s.StoreName,
			Status:        s.Status,
			ItemCount:     s.ItemCount,
			VarianceLines: s.VarianceLines,
			CreatedAt:     timestamppb.New(s.CreatedAt),
		}
		if !s.CompletedAt.IsZero() {
			item.CompletedAt = timestamppb.New(s.CompletedAt)
		}
		pb = append(pb, item)
	}
	return connect.NewResponse(&genposv1.ListStockTakesResponse{StockTakes: pb}), nil
}

func (h *StockTakeHandler) GetStockTake(
	ctx context.Context,
	req *connect.Request[genposv1.GetStockTakeRequest],
) (*connect.Response[genposv1.GetStockTakeResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	st, err := h.usecase.GetStockTake(ctx, input.GetStockTakeInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "get stock take", err)
	}
	return connect.NewResponse(&genposv1.GetStockTakeResponse{StockTake: toStockTakeProto(st)}), nil
}

func (h *StockTakeHandler) CreateStockTake(
	ctx context.Context,
	req *connect.Request[genposv1.CreateStockTakeRequest],
) (*connect.Response[genposv1.CreateStockTakeResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	st, err := h.usecase.CreateStockTake(ctx, input.CreateStockTakeInput{
		OrgID:   authCtx.OrgID,
		UserID:  authCtx.UserID,
		StoreID: req.Msg.GetStoreId(),
		Notes:   req.Msg.GetNotes(),
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "create stock take", err)
	}
	return connect.NewResponse(&genposv1.CreateStockTakeResponse{StockTake: toStockTakeProto(st)}), nil
}

func (h *StockTakeHandler) SaveStockTakeProgress(
	ctx context.Context,
	req *connect.Request[genposv1.SaveStockTakeProgressRequest],
) (*connect.Response[genposv1.SaveStockTakeProgressResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	lines := make([]input.StockTakeLineInput, 0, len(req.Msg.GetLines()))
	for _, ln := range req.Msg.GetLines() {
		lines = append(lines, input.StockTakeLineInput{
			ItemID:     ln.GetItemId(),
			CountedQty: ln.GetCountedQty(),
		})
	}
	st, err := h.usecase.SaveStockTakeProgress(ctx, input.SaveStockTakeProgressInput{
		ID:    req.Msg.GetId(),
		OrgID: authCtx.OrgID,
		Notes: req.Msg.GetNotes(),
		Lines: lines,
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "save stock take progress", err)
	}
	return connect.NewResponse(&genposv1.SaveStockTakeProgressResponse{StockTake: toStockTakeProto(st)}), nil
}

func (h *StockTakeHandler) FinalizeStockTake(
	ctx context.Context,
	req *connect.Request[genposv1.FinalizeStockTakeRequest],
) (*connect.Response[genposv1.FinalizeStockTakeResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	st, err := h.usecase.FinalizeStockTake(ctx, input.FinalizeStockTakeInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID, UserID: authCtx.UserID,
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "finalize stock take", err)
	}
	return connect.NewResponse(&genposv1.FinalizeStockTakeResponse{StockTake: toStockTakeProto(st)}), nil
}

func (h *StockTakeHandler) CancelStockTake(
	ctx context.Context,
	req *connect.Request[genposv1.CancelStockTakeRequest],
) (*connect.Response[genposv1.CancelStockTakeResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	st, err := h.usecase.CancelStockTake(ctx, input.CancelStockTakeInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	})
	if err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "cancel stock take", err)
	}
	return connect.NewResponse(&genposv1.CancelStockTakeResponse{StockTake: toStockTakeProto(st)}), nil
}

func (h *StockTakeHandler) DeleteStockTake(
	ctx context.Context,
	req *connect.Request[genposv1.DeleteStockTakeRequest],
) (*connect.Response[genposv1.DeleteStockTakeResponse], error) {
	authCtx, err := requireInventoryAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeleteStockTake(ctx, input.DeleteStockTakeInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	}); err != nil {
		return nil, logAndConvertInventory(h.logger, ctx, "delete stock take", err)
	}
	return connect.NewResponse(&genposv1.DeleteStockTakeResponse{}), nil
}

func toStockTakeProto(st *entity.StockTake) *genposv1.StockTake {
	if st == nil {
		return nil
	}
	items := make([]*genposv1.StockTakeItem, 0, len(st.Items))
	for _, it := range st.Items {
		items = append(items, &genposv1.StockTakeItem{
			Id:          it.ID,
			VariantId:   it.VariantID,
			VariantName: it.VariantName,
			ProductName: it.ProductName,
			ExpectedQty: it.ExpectedQty,
			CountedQty:  it.CountedQty,
		})
	}
	out := &genposv1.StockTake{
		Id:        st.ID,
		OrgId:     st.OrgID,
		StoreId:   st.StoreID,
		Status:    st.Status,
		Notes:     st.Notes,
		Items:     items,
		CreatedAt: timestamppb.New(st.CreatedAt),
		UpdatedAt: timestamppb.New(st.UpdatedAt),
	}
	if !st.CompletedAt.IsZero() {
		out.CompletedAt = timestamppb.New(st.CompletedAt)
	}
	return out
}

var _ genposv1connect.StockTakeServiceHandler = (*StockTakeHandler)(nil)

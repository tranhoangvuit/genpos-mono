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

// StoreHandler implements StoreServiceHandler.
type StoreHandler struct {
	genposv1connect.UnimplementedStoreServiceHandler
	logger  *slog.Logger
	usecase usecase.StoreUsecase
}

// NewStoreHandler constructs a StoreHandler.
func NewStoreHandler(logger *slog.Logger, uc usecase.StoreUsecase) *StoreHandler {
	return &StoreHandler{logger: logger, usecase: uc}
}

func (h *StoreHandler) ListStoreDetails(
	ctx context.Context,
	_ *connect.Request[genposv1.ListStoreDetailsRequest],
) (*connect.Response[genposv1.ListStoreDetailsResponse], error) {
	authCtx, err := h.requireStoreAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListStoreDetails(ctx, authCtx.OrgID)
	if err != nil {
		return nil, h.logAndConvertStore(ctx, "list stores", err)
	}
	pb := make([]*genposv1.Store, 0, len(items))
	for _, s := range items {
		pb = append(pb, toStoreProto(s))
	}
	return connect.NewResponse(&genposv1.ListStoreDetailsResponse{Stores: pb}), nil
}

func (h *StoreHandler) CreateStore(
	ctx context.Context,
	req *connect.Request[genposv1.CreateStoreRequest],
) (*connect.Response[genposv1.CreateStoreResponse], error) {
	authCtx, err := h.requireStoreAuth(ctx)
	if err != nil {
		return nil, err
	}
	s, err := h.usecase.CreateStore(ctx, input.CreateStoreInput{
		OrgID: authCtx.OrgID,
		Store: fromStoreInputProto(req.Msg.GetStore()),
	})
	if err != nil {
		return nil, h.logAndConvertStore(ctx, "create store", err)
	}
	return connect.NewResponse(&genposv1.CreateStoreResponse{Store: toStoreProto(s)}), nil
}

func (h *StoreHandler) UpdateStore(
	ctx context.Context,
	req *connect.Request[genposv1.UpdateStoreRequest],
) (*connect.Response[genposv1.UpdateStoreResponse], error) {
	authCtx, err := h.requireStoreAuth(ctx)
	if err != nil {
		return nil, err
	}
	s, err := h.usecase.UpdateStore(ctx, input.UpdateStoreInput{
		ID:    req.Msg.GetId(),
		OrgID: authCtx.OrgID,
		Store: fromStoreInputProto(req.Msg.GetStore()),
	})
	if err != nil {
		return nil, h.logAndConvertStore(ctx, "update store", err)
	}
	return connect.NewResponse(&genposv1.UpdateStoreResponse{Store: toStoreProto(s)}), nil
}

func (h *StoreHandler) DeleteStore(
	ctx context.Context,
	req *connect.Request[genposv1.DeleteStoreRequest],
) (*connect.Response[genposv1.DeleteStoreResponse], error) {
	authCtx, err := h.requireStoreAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeleteStore(ctx, input.DeleteStoreInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	}); err != nil {
		return nil, h.logAndConvertStore(ctx, "delete store", err)
	}
	return connect.NewResponse(&genposv1.DeleteStoreResponse{}), nil
}

func (h *StoreHandler) requireStoreAuth(ctx context.Context) (*interceptor.AuthContext, error) {
	a := interceptor.FromContext(ctx)
	if a == nil {
		return nil, errors.ToConnectError(errors.Unauthorized("not signed in"))
	}
	return a, nil
}

func (h *StoreHandler) logAndConvertStore(ctx context.Context, op string, err error) error {
	if errors.ShouldLog(errors.GetCode(err)) {
		h.logger.ErrorContext(ctx, op+" failed", "error", err)
	}
	return errors.ToConnectError(err)
}

func toStoreProto(s *entity.Store) *genposv1.Store {
	if s == nil {
		return nil
	}
	return &genposv1.Store{
		Id:        s.ID,
		OrgId:     s.OrgID,
		Name:      s.Name,
		Address:   s.Address,
		Phone:     s.Phone,
		Email:     s.Email,
		Timezone:  s.Timezone,
		Status:    s.Status,
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: timestamppb.New(s.UpdatedAt),
	}
}

func fromStoreInputProto(s *genposv1.StoreInput) input.StoreInput {
	if s == nil {
		return input.StoreInput{}
	}
	return input.StoreInput{
		Name:     s.GetName(),
		Address:  s.GetAddress(),
		Phone:    s.GetPhone(),
		Email:    s.GetEmail(),
		Timezone: s.GetTimezone(),
		Status:   s.GetStatus(),
	}
}

var _ genposv1connect.StoreServiceHandler = (*StoreHandler)(nil)

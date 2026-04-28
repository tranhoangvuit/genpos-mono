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

type TaxClassHandler struct {
	genposv1connect.UnimplementedTaxClassServiceHandler
	logger  *slog.Logger
	usecase usecase.TaxClassUsecase
}

func NewTaxClassHandler(logger *slog.Logger, uc usecase.TaxClassUsecase) *TaxClassHandler {
	return &TaxClassHandler{logger: logger, usecase: uc}
}

func (h *TaxClassHandler) ListTaxClasses(
	ctx context.Context,
	_ *connect.Request[genposv1.ListTaxClassesRequest],
) (*connect.Response[genposv1.ListTaxClassesResponse], error) {
	authCtx, err := requireTaxClassAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListTaxClasses(ctx, authCtx.OrgID)
	if err != nil {
		return nil, h.logAndConvert(ctx, "list tax classes", err)
	}
	pb := make([]*genposv1.TaxClass, 0, len(items))
	for _, c := range items {
		pb = append(pb, toTaxClassProto(c))
	}
	return connect.NewResponse(&genposv1.ListTaxClassesResponse{Classes: pb}), nil
}

func (h *TaxClassHandler) GetTaxClass(
	ctx context.Context,
	req *connect.Request[genposv1.GetTaxClassRequest],
) (*connect.Response[genposv1.GetTaxClassResponse], error) {
	authCtx, err := requireTaxClassAuth(ctx)
	if err != nil {
		return nil, err
	}
	c, err := h.usecase.GetTaxClass(ctx, input.GetTaxClassInput{
		OrgID: authCtx.OrgID, ID: req.Msg.GetId(),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "get tax class", err)
	}
	return connect.NewResponse(&genposv1.GetTaxClassResponse{Class: toTaxClassProto(c)}), nil
}

func (h *TaxClassHandler) CreateTaxClass(
	ctx context.Context,
	req *connect.Request[genposv1.CreateTaxClassRequest],
) (*connect.Response[genposv1.CreateTaxClassResponse], error) {
	authCtx, err := requireTaxClassAuth(ctx)
	if err != nil {
		return nil, err
	}
	c, err := h.usecase.CreateTaxClass(ctx, input.CreateTaxClassInput{
		OrgID: authCtx.OrgID,
		Class: fromTaxClassInputProto(req.Msg.GetClass()),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "create tax class", err)
	}
	return connect.NewResponse(&genposv1.CreateTaxClassResponse{Class: toTaxClassProto(c)}), nil
}

func (h *TaxClassHandler) UpdateTaxClass(
	ctx context.Context,
	req *connect.Request[genposv1.UpdateTaxClassRequest],
) (*connect.Response[genposv1.UpdateTaxClassResponse], error) {
	authCtx, err := requireTaxClassAuth(ctx)
	if err != nil {
		return nil, err
	}
	c, err := h.usecase.UpdateTaxClass(ctx, input.UpdateTaxClassInput{
		ID:    req.Msg.GetId(),
		OrgID: authCtx.OrgID,
		Class: fromTaxClassInputProto(req.Msg.GetClass()),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "update tax class", err)
	}
	return connect.NewResponse(&genposv1.UpdateTaxClassResponse{Class: toTaxClassProto(c)}), nil
}

func (h *TaxClassHandler) DeleteTaxClass(
	ctx context.Context,
	req *connect.Request[genposv1.DeleteTaxClassRequest],
) (*connect.Response[genposv1.DeleteTaxClassResponse], error) {
	authCtx, err := requireTaxClassAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeleteTaxClass(ctx, input.DeleteTaxClassInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	}); err != nil {
		return nil, h.logAndConvert(ctx, "delete tax class", err)
	}
	return connect.NewResponse(&genposv1.DeleteTaxClassResponse{}), nil
}

func requireTaxClassAuth(ctx context.Context) (*interceptor.AuthContext, error) {
	a := interceptor.FromContext(ctx)
	if a == nil {
		return nil, errors.ToConnectError(errors.Unauthorized("not signed in"))
	}
	return a, nil
}

func (h *TaxClassHandler) logAndConvert(ctx context.Context, op string, err error) error {
	if errors.ShouldLog(errors.GetCode(err)) {
		h.logger.ErrorContext(ctx, op+" failed", "error", err)
	}
	return errors.ToConnectError(err)
}

func toTaxClassProto(c *entity.TaxClass) *genposv1.TaxClass {
	if c == nil {
		return nil
	}
	rates := make([]*genposv1.TaxClassRate, 0, len(c.Rates))
	for _, r := range c.Rates {
		rates = append(rates, &genposv1.TaxClassRate{
			Id:         r.ID,
			TaxRateId:  r.TaxRateID,
			Sequence:   r.Sequence,
			IsCompound: r.IsCompound,
		})
	}
	return &genposv1.TaxClass{
		Id:          c.ID,
		OrgId:       c.OrgID,
		Name:        c.Name,
		Description: c.Description,
		IsDefault:   c.IsDefault,
		SortOrder:   c.SortOrder,
		Rates:       rates,
		CreatedAt:   timestamppb.New(c.CreatedAt),
		UpdatedAt:   timestamppb.New(c.UpdatedAt),
	}
}

func fromTaxClassInputProto(c *genposv1.TaxClassInput) input.TaxClassInput {
	if c == nil {
		return input.TaxClassInput{}
	}
	rates := make([]input.TaxClassRateInput, 0, len(c.GetRates()))
	for _, r := range c.GetRates() {
		rates = append(rates, input.TaxClassRateInput{
			TaxRateID:  r.GetTaxRateId(),
			Sequence:   r.GetSequence(),
			IsCompound: r.GetIsCompound(),
		})
	}
	return input.TaxClassInput{
		Name:        c.GetName(),
		Description: c.GetDescription(),
		IsDefault:   c.GetIsDefault(),
		SortOrder:   c.GetSortOrder(),
		Rates:       rates,
	}
}

var _ genposv1connect.TaxClassServiceHandler = (*TaxClassHandler)(nil)

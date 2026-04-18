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

type TaxRateHandler struct {
	genposv1connect.UnimplementedTaxRateServiceHandler
	logger  *slog.Logger
	usecase usecase.TaxRateUsecase
}

func NewTaxRateHandler(logger *slog.Logger, uc usecase.TaxRateUsecase) *TaxRateHandler {
	return &TaxRateHandler{logger: logger, usecase: uc}
}

func (h *TaxRateHandler) ListTaxRates(
	ctx context.Context,
	_ *connect.Request[genposv1.ListTaxRatesRequest],
) (*connect.Response[genposv1.ListTaxRatesResponse], error) {
	authCtx, err := requireTaxRateAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListTaxRates(ctx, authCtx.OrgID)
	if err != nil {
		return nil, h.logAndConvert(ctx, "list tax rates", err)
	}
	pb := make([]*genposv1.TaxRate, 0, len(items))
	for _, r := range items {
		pb = append(pb, toTaxRateProto(r))
	}
	return connect.NewResponse(&genposv1.ListTaxRatesResponse{Rates: pb}), nil
}

func (h *TaxRateHandler) CreateTaxRate(
	ctx context.Context,
	req *connect.Request[genposv1.CreateTaxRateRequest],
) (*connect.Response[genposv1.CreateTaxRateResponse], error) {
	authCtx, err := requireTaxRateAuth(ctx)
	if err != nil {
		return nil, err
	}
	r, err := h.usecase.CreateTaxRate(ctx, input.CreateTaxRateInput{
		OrgID: authCtx.OrgID,
		Rate:  fromTaxRateInputProto(req.Msg.GetRate()),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "create tax rate", err)
	}
	return connect.NewResponse(&genposv1.CreateTaxRateResponse{Rate: toTaxRateProto(r)}), nil
}

func (h *TaxRateHandler) UpdateTaxRate(
	ctx context.Context,
	req *connect.Request[genposv1.UpdateTaxRateRequest],
) (*connect.Response[genposv1.UpdateTaxRateResponse], error) {
	authCtx, err := requireTaxRateAuth(ctx)
	if err != nil {
		return nil, err
	}
	r, err := h.usecase.UpdateTaxRate(ctx, input.UpdateTaxRateInput{
		ID:    req.Msg.GetId(),
		OrgID: authCtx.OrgID,
		Rate:  fromTaxRateInputProto(req.Msg.GetRate()),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "update tax rate", err)
	}
	return connect.NewResponse(&genposv1.UpdateTaxRateResponse{Rate: toTaxRateProto(r)}), nil
}

func (h *TaxRateHandler) DeleteTaxRate(
	ctx context.Context,
	req *connect.Request[genposv1.DeleteTaxRateRequest],
) (*connect.Response[genposv1.DeleteTaxRateResponse], error) {
	authCtx, err := requireTaxRateAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeleteTaxRate(ctx, input.DeleteTaxRateInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	}); err != nil {
		return nil, h.logAndConvert(ctx, "delete tax rate", err)
	}
	return connect.NewResponse(&genposv1.DeleteTaxRateResponse{}), nil
}

func requireTaxRateAuth(ctx context.Context) (*interceptor.AuthContext, error) {
	a := interceptor.FromContext(ctx)
	if a == nil {
		return nil, errors.ToConnectError(errors.Unauthorized("not signed in"))
	}
	return a, nil
}

func (h *TaxRateHandler) logAndConvert(ctx context.Context, op string, err error) error {
	if errors.ShouldLog(errors.GetCode(err)) {
		h.logger.ErrorContext(ctx, op+" failed", "error", err)
	}
	return errors.ToConnectError(err)
}

func toTaxRateProto(r *entity.TaxRate) *genposv1.TaxRate {
	if r == nil {
		return nil
	}
	return &genposv1.TaxRate{
		Id:          r.ID,
		OrgId:       r.OrgID,
		Name:        r.Name,
		Rate:        r.Rate,
		IsInclusive: r.IsInclusive,
		IsDefault:   r.IsDefault,
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
	}
}

func fromTaxRateInputProto(r *genposv1.TaxRateInput) input.TaxRateInput {
	if r == nil {
		return input.TaxRateInput{}
	}
	return input.TaxRateInput{
		Name:        r.GetName(),
		Rate:        r.GetRate(),
		IsInclusive: r.GetIsInclusive(),
		IsDefault:   r.GetIsDefault(),
	}
}

var _ genposv1connect.TaxRateServiceHandler = (*TaxRateHandler)(nil)

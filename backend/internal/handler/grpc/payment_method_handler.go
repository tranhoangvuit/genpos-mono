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

type PaymentMethodHandler struct {
	genposv1connect.UnimplementedPaymentMethodServiceHandler
	logger  *slog.Logger
	usecase usecase.PaymentMethodUsecase
}

func NewPaymentMethodHandler(logger *slog.Logger, uc usecase.PaymentMethodUsecase) *PaymentMethodHandler {
	return &PaymentMethodHandler{logger: logger, usecase: uc}
}

func (h *PaymentMethodHandler) ListPaymentMethods(
	ctx context.Context,
	_ *connect.Request[genposv1.ListPaymentMethodsRequest],
) (*connect.Response[genposv1.ListPaymentMethodsResponse], error) {
	authCtx, err := requirePaymentMethodAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListPaymentMethods(ctx, authCtx.OrgID)
	if err != nil {
		return nil, h.logAndConvert(ctx, "list payment methods", err)
	}
	pb := make([]*genposv1.PaymentMethod, 0, len(items))
	for _, m := range items {
		pb = append(pb, toPaymentMethodProto(m))
	}
	return connect.NewResponse(&genposv1.ListPaymentMethodsResponse{Methods: pb}), nil
}

func (h *PaymentMethodHandler) CreatePaymentMethod(
	ctx context.Context,
	req *connect.Request[genposv1.CreatePaymentMethodRequest],
) (*connect.Response[genposv1.CreatePaymentMethodResponse], error) {
	authCtx, err := requirePaymentMethodAuth(ctx)
	if err != nil {
		return nil, err
	}
	m, err := h.usecase.CreatePaymentMethod(ctx, input.CreatePaymentMethodInput{
		OrgID:  authCtx.OrgID,
		Method: fromPaymentMethodInputProto(req.Msg.GetMethod()),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "create payment method", err)
	}
	return connect.NewResponse(&genposv1.CreatePaymentMethodResponse{Method: toPaymentMethodProto(m)}), nil
}

func (h *PaymentMethodHandler) UpdatePaymentMethod(
	ctx context.Context,
	req *connect.Request[genposv1.UpdatePaymentMethodRequest],
) (*connect.Response[genposv1.UpdatePaymentMethodResponse], error) {
	authCtx, err := requirePaymentMethodAuth(ctx)
	if err != nil {
		return nil, err
	}
	m, err := h.usecase.UpdatePaymentMethod(ctx, input.UpdatePaymentMethodInput{
		ID:     req.Msg.GetId(),
		OrgID:  authCtx.OrgID,
		Method: fromPaymentMethodInputProto(req.Msg.GetMethod()),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "update payment method", err)
	}
	return connect.NewResponse(&genposv1.UpdatePaymentMethodResponse{Method: toPaymentMethodProto(m)}), nil
}

func (h *PaymentMethodHandler) DeletePaymentMethod(
	ctx context.Context,
	req *connect.Request[genposv1.DeletePaymentMethodRequest],
) (*connect.Response[genposv1.DeletePaymentMethodResponse], error) {
	authCtx, err := requirePaymentMethodAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeletePaymentMethod(ctx, input.DeletePaymentMethodInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	}); err != nil {
		return nil, h.logAndConvert(ctx, "delete payment method", err)
	}
	return connect.NewResponse(&genposv1.DeletePaymentMethodResponse{}), nil
}

func requirePaymentMethodAuth(ctx context.Context) (*interceptor.AuthContext, error) {
	a := interceptor.FromContext(ctx)
	if a == nil {
		return nil, errors.ToConnectError(errors.Unauthorized("not signed in"))
	}
	return a, nil
}

func (h *PaymentMethodHandler) logAndConvert(ctx context.Context, op string, err error) error {
	if errors.ShouldLog(errors.GetCode(err)) {
		h.logger.ErrorContext(ctx, op+" failed", "error", err)
	}
	return errors.ToConnectError(err)
}

func toPaymentMethodProto(m *entity.PaymentMethod) *genposv1.PaymentMethod {
	if m == nil {
		return nil
	}
	return &genposv1.PaymentMethod{
		Id:        m.ID,
		OrgId:     m.OrgID,
		Name:      m.Name,
		Type:      m.Type,
		IsActive:  m.IsActive,
		SortOrder: m.SortOrder,
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
	}
}

func fromPaymentMethodInputProto(m *genposv1.PaymentMethodInput) input.PaymentMethodInput {
	if m == nil {
		return input.PaymentMethodInput{}
	}
	return input.PaymentMethodInput{
		Name:      m.GetName(),
		Type:      m.GetType(),
		IsActive:  m.GetIsActive(),
		SortOrder: m.GetSortOrder(),
	}
}

var _ genposv1connect.PaymentMethodServiceHandler = (*PaymentMethodHandler)(nil)

package handler

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	genposv1 "github.com/genpick/genpos-mono/backend/gen/genpos/v1"
	"github.com/genpick/genpos-mono/backend/gen/genpos/v1/genposv1connect"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements GenposServiceHandler.
type Server struct {
	genposv1connect.UnimplementedGenposServiceHandler
	logger         *slog.Logger
	productUsecase usecase.ProductUsecase
}

// NewServer creates a Server with all required dependencies.
func NewServer(logger *slog.Logger, productUsecase usecase.ProductUsecase) *Server {
	return &Server{
		logger:         logger,
		productUsecase: productUsecase,
	}
}

func (s *Server) Ping(
	_ context.Context,
	_ *connect.Request[genposv1.PingRequest],
) (*connect.Response[genposv1.PingResponse], error) {
	return connect.NewResponse(&genposv1.PingResponse{
		Message: "pong",
	}), nil
}

func (s *Server) ListProducts(
	ctx context.Context,
	req *connect.Request[genposv1.ListProductsRequest],
) (*connect.Response[genposv1.ListProductsResponse], error) {
	msg := req.Msg

	if msg.GetOrgId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("org_id is required"))
	}

	products, err := s.productUsecase.ListProducts(ctx, input.ListProductsInput{
		OrgID:    msg.GetOrgId(),
		PageSize: msg.GetPageSize(),
		Offset:   msg.GetOffset(),
	})
	if err != nil {
		if errors.ShouldLog(errors.GetCode(err)) {
			s.logger.ErrorContext(ctx, "list products failed", "error", err)
		}
		return nil, errors.ToConnectError(err)
	}

	pbProducts := make([]*genposv1.Product, 0, len(products))
	for _, p := range products {
		pbProducts = append(pbProducts, &genposv1.Product{
			Id:         p.ID,
			OrgId:      p.OrgID,
			Name:       p.Name,
			Sku:        p.SKU,
			PriceCents: p.PriceCents,
			Active:     p.Active,
			CreatedAt:  timestamppb.New(p.CreatedAt),
			UpdatedAt:  timestamppb.New(p.UpdatedAt),
		})
	}

	return connect.NewResponse(&genposv1.ListProductsResponse{
		Products: pbProducts,
	}), nil
}

var _ genposv1connect.GenposServiceHandler = (*Server)(nil)

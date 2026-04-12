package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=product_gateway.go -destination=mock/mock_product_gateway.go -package=mock

// ListProductsParams holds filtering and pagination parameters.
type ListProductsParams struct {
	Limit  int32
	Offset int32
}

// ProductReader defines the read-side data access contract for products.
type ProductReader interface {
	ListProducts(ctx context.Context, params ListProductsParams) ([]*entity.Product, error)
}

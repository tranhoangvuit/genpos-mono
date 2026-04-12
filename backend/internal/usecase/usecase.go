package usecase

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
)

//go:generate mockgen -source=usecase.go -destination=mock/mock_usecase.go -package=mock

// ProductUsecase is the service contract consumed by handlers.
type ProductUsecase interface {
	ListProducts(ctx context.Context, in input.ListProductsInput) ([]*entity.Product, error)
}

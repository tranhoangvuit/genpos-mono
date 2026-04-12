package usecase

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type productUsecase struct {
	tenantDB       gateway.TenantDB
	productReader  gateway.ProductReader
}

// NewProductUsecase constructs a ProductUsecase.
func NewProductUsecase(tenantDB gateway.TenantDB, pr gateway.ProductReader) ProductUsecase {
	return &productUsecase{
		tenantDB:      tenantDB,
		productReader: pr,
	}
}

func (u *productUsecase) ListProducts(ctx context.Context, in input.ListProductsInput) ([]*entity.Product, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org_id is required")
	}

	pageSize := in.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var products []*entity.Product

	err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		var err error
		products, err = u.productReader.ListProducts(ctx, gateway.ListProductsParams{
			Limit:  pageSize,
			Offset: in.Offset,
		})
		return err
	})
	if err != nil {
		return nil, errors.Wrap(err, "list products")
	}

	return products, nil
}

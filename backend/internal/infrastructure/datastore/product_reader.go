package datastore

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type productReader struct{}

// NewProductReader creates a ProductReader backed by sqlc.
func NewProductReader() gateway.ProductReader {
	return &productReader{}
}

func (r *productReader) ListProducts(ctx context.Context, params gateway.ListProductsParams) ([]*entity.Product, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}

	q := sqlc.New(dbtx)
	rows, err := q.ListProducts(ctx, sqlc.ListProductsParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, errors.Wrap(err, "list products")
	}

	products := make([]*entity.Product, 0, len(rows))
	for _, row := range rows {
		products = append(products, toProductEntity(row))
	}
	return products, nil
}

func toProductEntity(row sqlc.Product) *entity.Product {
	return &entity.Product{
		ID:         row.ID.String(),
		OrgID:      row.OrgID,
		Name:       row.Name,
		SKU:        row.Sku,
		PriceCents: row.PriceCents,
		Active:     row.Active,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
}

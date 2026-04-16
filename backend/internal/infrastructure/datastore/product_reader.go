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
		products = append(products, &entity.Product{
			ID:          uuidString(row.ID),
			OrgID:       uuidString(row.OrgID),
			CategoryID:  uuidString(row.CategoryID),
			Name:        row.Name,
			Description: row.Description.String,
			ImageURL:    row.ImageUrl.String,
			IsActive:    row.IsActive,
			SortOrder:   row.SortOrder,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return products, nil
}

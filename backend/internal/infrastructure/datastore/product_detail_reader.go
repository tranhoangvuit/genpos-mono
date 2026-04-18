package datastore

import (
	"context"
	stderrors "errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type productDetailReader struct{}

// NewProductDetailReader returns a ProductDetailReader backed by sqlc.
func NewProductDetailReader() gateway.ProductDetailReader { return &productDetailReader{} }

func (r *productDetailReader) GetByID(ctx context.Context, id string) (*entity.ProductDetail, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid product id")
	}
	q := sqlc.New(dbtx)
	p, err := q.GetProductByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("product not found")
		}
		return nil, errors.Wrap(err, "get product")
	}
	return loadDetail(ctx, q, p.ID, p.OrgID, p.CategoryID, p.Name, p.Description, p.IsActive, p.SortOrder, p.CreatedAt, p.UpdatedAt)
}

func (r *productDetailReader) ListSummaries(ctx context.Context) ([]*entity.ProductListItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListProductSummaries(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list product summaries")
	}
	out := make([]*entity.ProductListItem, 0, len(rows))
	for _, r := range rows {
		price, _ := r.Price.(string)
		if price == "" {
			price = "0"
		}
		out = append(out, &entity.ProductListItem{
			ID:           uuidString(r.ID),
			Name:         r.Name,
			CategoryID:   uuidString(r.CategoryID),
			CategoryName: textString(r.CategoryName),
			Price:        price,
			VariantCount: r.VariantCount,
			IsActive:     r.IsActive,
		})
	}
	return out, nil
}

func (r *productDetailReader) GetByName(ctx context.Context, name string) (*entity.ProductDetail, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	q := sqlc.New(dbtx)
	p, err := q.GetProductByName(ctx, name)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("product not found")
		}
		return nil, errors.Wrap(err, "get product by name")
	}
	return loadDetail(ctx, q, p.ID, p.OrgID, p.CategoryID, p.Name, p.Description, p.IsActive, p.SortOrder, p.CreatedAt, p.UpdatedAt)
}

func loadDetail(ctx context.Context, q *sqlc.Queries,
	id, orgID, categoryID pgtype.UUID, name string, description pgtype.Text,
	isActive bool, sortOrder int32, createdAt, updatedAt pgtype.Timestamptz,
) (*entity.ProductDetail, error) {
	detail := &entity.ProductDetail{
		ID:          uuidString(id),
		OrgID:       uuidString(orgID),
		CategoryID:  uuidString(categoryID),
		Name:        name,
		Description: textString(description),
		IsActive:    isActive,
		SortOrder:   sortOrder,
		CreatedAt:   createdAt.Time,
		UpdatedAt:   updatedAt.Time,
	}

	optionRows, err := q.ListProductOptions(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "list options")
	}
	valueRows, err := q.ListProductOptionValues(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "list option values")
	}
	variantRows, err := q.ListProductVariants(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "list variants")
	}
	vovRows, err := q.ListProductVariantOptionValues(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "list variant option values")
	}
	imageRows, err := q.ListProductImages(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "list images")
	}

	optionsByID := make(map[string]*entity.ProductOption, len(optionRows))
	for _, o := range optionRows {
		op := &entity.ProductOption{
			ID:        uuidString(o.ID),
			Name:      o.Name,
			SortOrder: o.SortOrder,
		}
		optionsByID[op.ID] = op
		detail.Options = append(detail.Options, op)
	}
	for _, v := range valueRows {
		if op, ok := optionsByID[uuidString(v.OptionID)]; ok {
			op.Values = append(op.Values, &entity.ProductOptionValue{
				ID:        uuidString(v.ID),
				OptionID:  uuidString(v.OptionID),
				Value:     v.Value,
				SortOrder: v.SortOrder,
			})
		}
	}

	variantsByID := make(map[string]*entity.ProductVariant, len(variantRows))
	for _, vr := range variantRows {
		vt := &entity.ProductVariant{
			ID:         uuidString(vr.ID),
			ProductID:  uuidString(vr.ProductID),
			Name:       vr.Name,
			SKU:        textString(vr.Sku),
			Barcode:    textString(vr.Barcode),
			Price:      numericToString(vr.Price),
			CostPrice:  numericToString(vr.CostPrice),
			TrackStock: vr.TrackStock,
			IsActive:   vr.IsActive,
			SortOrder:  vr.SortOrder,
		}
		variantsByID[vt.ID] = vt
		detail.Variants = append(detail.Variants, vt)
	}
	for _, vov := range vovRows {
		if vt, ok := variantsByID[uuidString(vov.VariantID)]; ok {
			vt.OptionValueIDs = append(vt.OptionValueIDs, uuidString(vov.OptionValueID))
		}
	}

	for _, img := range imageRows {
		detail.Images = append(detail.Images, &entity.ProductImage{
			ID:        uuidString(img.ID),
			VariantID: uuidString(img.VariantID),
			URL:       img.Url,
			SortOrder: img.SortOrder,
		})
	}

	return detail, nil
}

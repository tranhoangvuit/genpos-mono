package datastore

import (
	"context"
	stderrors "errors"

	"github.com/jackc/pgx/v5"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type productWriter struct{}

// NewProductWriter returns a ProductWriter backed by sqlc.
func NewProductWriter() gateway.ProductWriter { return &productWriter{} }

func (w *productWriter) CreateBase(ctx context.Context, p gateway.CreateProductBaseParams) (*entity.ProductDetail, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(p.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	categoryID, err := uuidOrNull(p.CategoryID)
	if err != nil {
		return nil, errors.BadRequest("invalid category id")
	}
	r, err := sqlc.New(dbtx).CreateProduct(ctx, sqlc.CreateProductParams{
		OrgID:       orgID,
		CategoryID:  categoryID,
		Name:        p.Name,
		Description: textOrNull(p.Description),
		IsActive:    p.IsActive,
		SortOrder:   p.SortOrder,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create product")
	}
	return &entity.ProductDetail{
		ID:          uuidString(r.ID),
		OrgID:       uuidString(r.OrgID),
		CategoryID:  uuidString(r.CategoryID),
		Name:        r.Name,
		Description: textString(r.Description),
		IsActive:    r.IsActive,
		SortOrder:   r.SortOrder,
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
	}, nil
}

func (w *productWriter) UpdateBase(ctx context.Context, p gateway.UpdateProductBaseParams) (*entity.ProductDetail, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseUUID(p.ID)
	if err != nil {
		return nil, errors.BadRequest("invalid product id")
	}
	categoryID, err := uuidOrNull(p.CategoryID)
	if err != nil {
		return nil, errors.BadRequest("invalid category id")
	}
	r, err := sqlc.New(dbtx).UpdateProduct(ctx, sqlc.UpdateProductParams{
		ID:          id,
		Name:        p.Name,
		Description: textOrNull(p.Description),
		CategoryID:  categoryID,
		IsActive:    p.IsActive,
		SortOrder:   p.SortOrder,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("product not found")
		}
		return nil, errors.Wrap(err, "update product")
	}
	return &entity.ProductDetail{
		ID:          uuidString(r.ID),
		OrgID:       uuidString(r.OrgID),
		CategoryID:  uuidString(r.CategoryID),
		Name:        r.Name,
		Description: textString(r.Description),
		IsActive:    r.IsActive,
		SortOrder:   r.SortOrder,
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
	}, nil
}

func (w *productWriter) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid product id")
	}
	if err := sqlc.New(dbtx).SoftDeleteProduct(ctx, uid); err != nil {
		return errors.Wrap(err, "soft delete product")
	}
	return nil
}

func (w *productWriter) InsertOption(ctx context.Context, p gateway.CreateProductOptionParams) (*entity.ProductOption, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(p.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	productID, err := parseUUID(p.ProductID)
	if err != nil {
		return nil, errors.BadRequest("invalid product id")
	}
	r, err := sqlc.New(dbtx).InsertProductOption(ctx, sqlc.InsertProductOptionParams{
		OrgID:     orgID,
		ProductID: productID,
		Name:      p.Name,
		SortOrder: p.SortOrder,
	})
	if err != nil {
		return nil, errors.Wrap(err, "insert option")
	}
	return &entity.ProductOption{
		ID:        uuidString(r.ID),
		Name:      r.Name,
		SortOrder: r.SortOrder,
	}, nil
}

func (w *productWriter) InsertOptionValue(ctx context.Context, p gateway.CreateProductOptionValueParams) (*entity.ProductOptionValue, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(p.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	optionID, err := parseUUID(p.OptionID)
	if err != nil {
		return nil, errors.BadRequest("invalid option id")
	}
	r, err := sqlc.New(dbtx).InsertProductOptionValue(ctx, sqlc.InsertProductOptionValueParams{
		OrgID:     orgID,
		OptionID:  optionID,
		Value:     p.Value,
		SortOrder: p.SortOrder,
	})
	if err != nil {
		return nil, errors.Wrap(err, "insert option value")
	}
	return &entity.ProductOptionValue{
		ID:        uuidString(r.ID),
		OptionID:  uuidString(r.OptionID),
		Value:     r.Value,
		SortOrder: r.SortOrder,
	}, nil
}

func (w *productWriter) InsertVariant(ctx context.Context, p gateway.CreateProductVariantParams) (*entity.ProductVariant, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(p.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	productID, err := parseUUID(p.ProductID)
	if err != nil {
		return nil, errors.BadRequest("invalid product id")
	}
	price, err := numericFromString(p.Price)
	if err != nil {
		return nil, errors.BadRequest("invalid price")
	}
	costPrice, err := numericFromString(p.CostPrice)
	if err != nil {
		return nil, errors.BadRequest("invalid cost price")
	}
	taxClassID, err := uuidOrNull(p.TaxClassID)
	if err != nil {
		return nil, errors.BadRequest("invalid tax class id")
	}
	r, err := sqlc.New(dbtx).InsertProductVariant(ctx, sqlc.InsertProductVariantParams{
		OrgID:      orgID,
		ProductID:  productID,
		Name:       p.Name,
		Sku:        textOrNull(p.SKU),
		Barcode:    textOrNull(p.Barcode),
		Price:      price,
		CostPrice:  costPrice,
		TrackStock: p.TrackStock,
		IsActive:   p.IsActive,
		SortOrder:  p.SortOrder,
		TaxClassID: taxClassID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "insert variant")
	}
	return &entity.ProductVariant{
		ID:         uuidString(r.ID),
		ProductID:  uuidString(r.ProductID),
		Name:       r.Name,
		SKU:        textString(r.Sku),
		Barcode:    textString(r.Barcode),
		Price:      numericToString(r.Price),
		CostPrice:  numericToString(r.CostPrice),
		TrackStock: r.TrackStock,
		IsActive:   r.IsActive,
		SortOrder:  r.SortOrder,
		TaxClassID: uuidString(r.TaxClassID),
	}, nil
}

func (w *productWriter) InsertVariantOptionValue(ctx context.Context, orgID, variantID, optionValueID string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	oid, err := parseUUID(orgID)
	if err != nil {
		return errors.BadRequest("invalid org id")
	}
	vid, err := parseUUID(variantID)
	if err != nil {
		return errors.BadRequest("invalid variant id")
	}
	ovid, err := parseUUID(optionValueID)
	if err != nil {
		return errors.BadRequest("invalid option value id")
	}
	if err := sqlc.New(dbtx).InsertProductVariantOptionValue(ctx, sqlc.InsertProductVariantOptionValueParams{
		OrgID:         oid,
		VariantID:     vid,
		OptionValueID: ovid,
	}); err != nil {
		return errors.Wrap(err, "insert variant option value")
	}
	return nil
}

func (w *productWriter) InsertImage(ctx context.Context, p gateway.CreateProductImageParams) (*entity.ProductImage, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(p.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	productID, err := parseUUID(p.ProductID)
	if err != nil {
		return nil, errors.BadRequest("invalid product id")
	}
	variantID, err := uuidOrNull(p.VariantID)
	if err != nil {
		return nil, errors.BadRequest("invalid variant id")
	}
	r, err := sqlc.New(dbtx).InsertProductImage(ctx, sqlc.InsertProductImageParams{
		OrgID:     orgID,
		ProductID: productID,
		VariantID: variantID,
		Url:       p.URL,
		SortOrder: p.SortOrder,
	})
	if err != nil {
		return nil, errors.Wrap(err, "insert image")
	}
	return &entity.ProductImage{
		ID:        uuidString(r.ID),
		VariantID: uuidString(r.VariantID),
		URL:       r.Url,
		SortOrder: r.SortOrder,
	}, nil
}

func (w *productWriter) DeleteOptionsByProduct(ctx context.Context, productID string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	pid, err := parseUUID(productID)
	if err != nil {
		return errors.BadRequest("invalid product id")
	}
	if err := sqlc.New(dbtx).DeleteProductOptionsByProduct(ctx, pid); err != nil {
		return errors.Wrap(err, "delete options")
	}
	return nil
}

func (w *productWriter) SoftDeleteVariantsByProduct(ctx context.Context, productID string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	pid, err := parseUUID(productID)
	if err != nil {
		return errors.BadRequest("invalid product id")
	}
	if err := sqlc.New(dbtx).SoftDeleteProductVariantsByProduct(ctx, pid); err != nil {
		return errors.Wrap(err, "soft delete variants")
	}
	return nil
}

func (w *productWriter) DeleteImagesByProduct(ctx context.Context, productID string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	pid, err := parseUUID(productID)
	if err != nil {
		return errors.BadRequest("invalid product id")
	}
	if err := sqlc.New(dbtx).DeleteProductImagesByProduct(ctx, pid); err != nil {
		return errors.Wrap(err, "delete images")
	}
	return nil
}

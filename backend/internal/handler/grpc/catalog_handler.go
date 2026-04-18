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

// CatalogHandler implements CatalogServiceHandler.
type CatalogHandler struct {
	genposv1connect.UnimplementedCatalogServiceHandler
	logger  *slog.Logger
	usecase usecase.CatalogUsecase
}

// NewCatalogHandler constructs a CatalogHandler.
func NewCatalogHandler(logger *slog.Logger, uc usecase.CatalogUsecase) *CatalogHandler {
	return &CatalogHandler{logger: logger, usecase: uc}
}

// ----- Categories ----------------------------------------------------------

func (h *CatalogHandler) ListCategories(
	ctx context.Context,
	_ *connect.Request[genposv1.ListCategoriesRequest],
) (*connect.Response[genposv1.ListCategoriesResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	categories, err := h.usecase.ListCategories(ctx, authCtx.OrgID)
	if err != nil {
		return nil, h.logAndConvert(ctx, "list categories", err)
	}
	pb := make([]*genposv1.Category, 0, len(categories))
	for _, c := range categories {
		pb = append(pb, toCategoryProto(c))
	}
	return connect.NewResponse(&genposv1.ListCategoriesResponse{Categories: pb}), nil
}

func (h *CatalogHandler) CreateCategory(
	ctx context.Context,
	req *connect.Request[genposv1.CreateCategoryRequest],
) (*connect.Response[genposv1.CreateCategoryResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	m := req.Msg
	cat, err := h.usecase.CreateCategory(ctx, input.CreateCategoryInput{
		OrgID:     authCtx.OrgID,
		Name:      m.GetName(),
		ParentID:  m.GetParentId(),
		Color:     m.GetColor(),
		SortOrder: m.GetSortOrder(),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "create category", err)
	}
	return connect.NewResponse(&genposv1.CreateCategoryResponse{Category: toCategoryProto(cat)}), nil
}

func (h *CatalogHandler) UpdateCategory(
	ctx context.Context,
	req *connect.Request[genposv1.UpdateCategoryRequest],
) (*connect.Response[genposv1.UpdateCategoryResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	m := req.Msg
	cat, err := h.usecase.UpdateCategory(ctx, input.UpdateCategoryInput{
		ID:        m.GetId(),
		OrgID:     authCtx.OrgID,
		Name:      m.GetName(),
		ParentID:  m.GetParentId(),
		Color:     m.GetColor(),
		SortOrder: m.GetSortOrder(),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "update category", err)
	}
	return connect.NewResponse(&genposv1.UpdateCategoryResponse{Category: toCategoryProto(cat)}), nil
}

func (h *CatalogHandler) DeleteCategory(
	ctx context.Context,
	req *connect.Request[genposv1.DeleteCategoryRequest],
) (*connect.Response[genposv1.DeleteCategoryResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeleteCategory(ctx, input.DeleteCategoryInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	}); err != nil {
		return nil, h.logAndConvert(ctx, "delete category", err)
	}
	return connect.NewResponse(&genposv1.DeleteCategoryResponse{}), nil
}

// ----- Products ------------------------------------------------------------

func (h *CatalogHandler) ListProducts(
	ctx context.Context,
	_ *connect.Request[genposv1.ListProductsRequest],
) (*connect.Response[genposv1.ListProductsResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListProducts(ctx, authCtx.OrgID)
	if err != nil {
		return nil, h.logAndConvert(ctx, "list products", err)
	}
	pb := make([]*genposv1.ProductListItem, 0, len(items))
	for _, p := range items {
		pb = append(pb, &genposv1.ProductListItem{
			Id:           p.ID,
			Name:         p.Name,
			CategoryId:   p.CategoryID,
			CategoryName: p.CategoryName,
			Price:        p.Price,
			VariantCount: p.VariantCount,
			IsActive:     p.IsActive,
		})
	}
	return connect.NewResponse(&genposv1.ListProductsResponse{Products: pb}), nil
}

func (h *CatalogHandler) GetProduct(
	ctx context.Context,
	req *connect.Request[genposv1.GetProductRequest],
) (*connect.Response[genposv1.GetProductResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	p, err := h.usecase.GetProduct(ctx, input.GetProductInput{ID: req.Msg.GetId(), OrgID: authCtx.OrgID})
	if err != nil {
		return nil, h.logAndConvert(ctx, "get product", err)
	}
	return connect.NewResponse(&genposv1.GetProductResponse{Product: toProductDetailProto(p)}), nil
}

func (h *CatalogHandler) CreateProduct(
	ctx context.Context,
	req *connect.Request[genposv1.CreateProductRequest],
) (*connect.Response[genposv1.CreateProductResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	p, err := h.usecase.CreateProduct(ctx, input.CreateProductInput{
		OrgID:   authCtx.OrgID,
		Product: fromProductInputProto(req.Msg.GetProduct()),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "create product", err)
	}
	return connect.NewResponse(&genposv1.CreateProductResponse{Product: toProductDetailProto(p)}), nil
}

func (h *CatalogHandler) UpdateProduct(
	ctx context.Context,
	req *connect.Request[genposv1.UpdateProductRequest],
) (*connect.Response[genposv1.UpdateProductResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	p, err := h.usecase.UpdateProduct(ctx, input.UpdateProductInput{
		ID:      req.Msg.GetId(),
		OrgID:   authCtx.OrgID,
		Product: fromProductInputProto(req.Msg.GetProduct()),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "update product", err)
	}
	return connect.NewResponse(&genposv1.UpdateProductResponse{Product: toProductDetailProto(p)}), nil
}

func (h *CatalogHandler) DeleteProduct(
	ctx context.Context,
	req *connect.Request[genposv1.DeleteProductRequest],
) (*connect.Response[genposv1.DeleteProductResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeleteProduct(ctx, input.DeleteProductInput{ID: req.Msg.GetId(), OrgID: authCtx.OrgID}); err != nil {
		return nil, h.logAndConvert(ctx, "delete product", err)
	}
	return connect.NewResponse(&genposv1.DeleteProductResponse{}), nil
}

// ----- CSV import ----------------------------------------------------------

func (h *CatalogHandler) ParseImportCsv(
	ctx context.Context,
	req *connect.Request[genposv1.ParseImportCsvRequest],
) (*connect.Response[genposv1.ParseImportCsvResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	res, err := h.usecase.ParseImportCsv(ctx, input.ParseImportCsvInput{
		OrgID:   authCtx.OrgID,
		CsvData: req.Msg.GetCsvData(),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "parse import csv", err)
	}
	rows := make([]*genposv1.CsvProductRow, 0, len(res.Rows))
	for _, r := range res.Rows {
		rows = append(rows, toCsvRowProto(r))
	}
	return connect.NewResponse(&genposv1.ParseImportCsvResponse{
		Rows:       rows,
		ValidCount: res.ValidCount,
		ErrorCount: res.ErrorCount,
		Warnings:   res.Warnings,
	}), nil
}

func (h *CatalogHandler) ImportProducts(
	ctx context.Context,
	req *connect.Request[genposv1.ImportProductsRequest],
) (*connect.Response[genposv1.ImportProductsResponse], error) {
	authCtx, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]input.ImportProductItem, 0, len(req.Msg.GetItems()))
	for _, it := range req.Msg.GetItems() {
		items = append(items, input.ImportProductItem{
			Row:              fromCsvRowProto(it.GetRow()),
			OverrideExisting: it.GetOverrideExisting(),
			ExistingID:       it.GetExistingId(),
		})
	}
	res, err := h.usecase.ImportProducts(ctx, input.ImportProductsInput{
		OrgID: authCtx.OrgID,
		Items: items,
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "import products", err)
	}
	return connect.NewResponse(&genposv1.ImportProductsResponse{
		Created: res.Created,
		Updated: res.Updated,
		Skipped: res.Skipped,
		Errors:  res.Errors,
	}), nil
}

// ----- helpers -------------------------------------------------------------

func (h *CatalogHandler) requireAuth(ctx context.Context) (*interceptor.AuthContext, error) {
	a := interceptor.FromContext(ctx)
	if a == nil {
		return nil, errors.ToConnectError(errors.Unauthorized("not signed in"))
	}
	return a, nil
}

func (h *CatalogHandler) logAndConvert(ctx context.Context, op string, err error) error {
	if errors.ShouldLog(errors.GetCode(err)) {
		h.logger.ErrorContext(ctx, op+" failed", "error", err)
	}
	return errors.ToConnectError(err)
}

func toCategoryProto(c *entity.Category) *genposv1.Category {
	return &genposv1.Category{
		Id:        c.ID,
		OrgId:     c.OrgID,
		ParentId:  c.ParentID,
		Name:      c.Name,
		SortOrder: c.SortOrder,
		Color:     c.Color,
		ImageUrl:  c.ImageURL,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

func toProductDetailProto(p *entity.ProductDetail) *genposv1.ProductDetail {
	if p == nil {
		return nil
	}
	out := &genposv1.ProductDetail{
		Id:          p.ID,
		OrgId:       p.OrgID,
		CategoryId:  p.CategoryID,
		Name:        p.Name,
		Description: p.Description,
		IsActive:    p.IsActive,
		SortOrder:   p.SortOrder,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
	for _, opt := range p.Options {
		pbOpt := &genposv1.ProductOption{
			Id: opt.ID, Name: opt.Name, SortOrder: opt.SortOrder,
		}
		for _, v := range opt.Values {
			pbOpt.Values = append(pbOpt.Values, &genposv1.ProductOptionValue{
				Id: v.ID, Value: v.Value, SortOrder: v.SortOrder,
			})
		}
		out.Options = append(out.Options, pbOpt)
	}
	for _, v := range p.Variants {
		out.Variants = append(out.Variants, &genposv1.ProductVariant{
			Id:             v.ID,
			Name:           v.Name,
			Sku:            v.SKU,
			Barcode:        v.Barcode,
			Price:          v.Price,
			CostPrice:      v.CostPrice,
			TrackStock:     v.TrackStock,
			IsActive:       v.IsActive,
			SortOrder:      v.SortOrder,
			OptionValueIds: v.OptionValueIDs,
		})
	}
	for _, img := range p.Images {
		out.Images = append(out.Images, &genposv1.ProductImage{
			Id: img.ID, VariantId: img.VariantID, Url: img.URL, SortOrder: img.SortOrder,
		})
	}
	return out
}

func fromProductInputProto(p *genposv1.ProductInput) input.ProductInput {
	if p == nil {
		return input.ProductInput{}
	}
	opts := make([]input.OptionInput, 0, len(p.GetOptions()))
	for _, o := range p.GetOptions() {
		opts = append(opts, input.OptionInput{Name: o.GetName(), Values: o.GetValues()})
	}
	vars := make([]input.VariantInput, 0, len(p.GetVariants()))
	for _, v := range p.GetVariants() {
		vars = append(vars, input.VariantInput{
			Name:         v.GetName(),
			SKU:          v.GetSku(),
			Barcode:      v.GetBarcode(),
			Price:        v.GetPrice(),
			CostPrice:    v.GetCostPrice(),
			TrackStock:   v.GetTrackStock(),
			IsActive:     v.GetIsActive(),
			SortOrder:    v.GetSortOrder(),
			OptionValues: v.GetOptionValues(),
		})
	}
	images := make([]input.ProductImageInput, 0, len(p.GetImages()))
	for _, img := range p.GetImages() {
		images = append(images, input.ProductImageInput{URL: img.GetUrl(), SortOrder: img.GetSortOrder()})
	}
	return input.ProductInput{
		Name:        p.GetName(),
		Description: p.GetDescription(),
		CategoryID:  p.GetCategoryId(),
		IsActive:    p.GetIsActive(),
		SortOrder:   p.GetSortOrder(),
		Options:     opts,
		Variants:    vars,
		Images:      images,
	}
}

func toCsvRowProto(r input.CsvProductRow) *genposv1.CsvProductRow {
	return &genposv1.CsvProductRow{
		Name:         r.Name,
		CategoryName: r.CategoryName,
		Description:  r.Description,
		Sku:          r.SKU,
		Barcode:      r.Barcode,
		Price:        r.Price,
		CostPrice:    r.CostPrice,
		IsActive:     r.IsActive,
		Errors:       r.Errors,
		Exists:       r.Exists,
		ExistingId:   r.ExistingID,
	}
}

func fromCsvRowProto(r *genposv1.CsvProductRow) input.CsvProductRow {
	if r == nil {
		return input.CsvProductRow{}
	}
	return input.CsvProductRow{
		Name:         r.GetName(),
		CategoryName: r.GetCategoryName(),
		Description:  r.GetDescription(),
		SKU:          r.GetSku(),
		Barcode:      r.GetBarcode(),
		Price:        r.GetPrice(),
		CostPrice:    r.GetCostPrice(),
		IsActive:     r.GetIsActive(),
		Errors:       r.GetErrors(),
		Exists:       r.GetExists(),
		ExistingID:   r.GetExistingId(),
	}
}

var _ genposv1connect.CatalogServiceHandler = (*CatalogHandler)(nil)

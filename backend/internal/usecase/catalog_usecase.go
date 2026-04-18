package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"strings"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type catalogUsecase struct {
	tenantDB       gateway.TenantDB
	categoryReader gateway.CategoryReader
	categoryWriter gateway.CategoryWriter
	productDetail  gateway.ProductDetailReader
	productWriter  gateway.ProductWriter
}

// NewCatalogUsecase constructs a CatalogUsecase.
func NewCatalogUsecase(
	tenantDB gateway.TenantDB,
	categoryReader gateway.CategoryReader,
	categoryWriter gateway.CategoryWriter,
	productDetail gateway.ProductDetailReader,
	productWriter gateway.ProductWriter,
) CatalogUsecase {
	return &catalogUsecase{
		tenantDB:       tenantDB,
		categoryReader: categoryReader,
		categoryWriter: categoryWriter,
		productDetail:  productDetail,
		productWriter:  productWriter,
	}
}

// ----- Categories ----------------------------------------------------------

func (u *catalogUsecase) ListCategories(ctx context.Context, orgID string) ([]*entity.Category, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.Category
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.categoryReader.List(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list categories")
	}
	return out, nil
}

func (u *catalogUsecase) CreateCategory(ctx context.Context, in input.CreateCategoryInput) (*entity.Category, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	var out *entity.Category
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		v, err := u.categoryWriter.Create(ctx, gateway.CreateCategoryParams{
			OrgID:     in.OrgID,
			Name:      strings.TrimSpace(in.Name),
			ParentID:  in.ParentID,
			Color:     in.Color,
			SortOrder: in.SortOrder,
		})
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "create category")
	}
	return out, nil
}

func (u *catalogUsecase) UpdateCategory(ctx context.Context, in input.UpdateCategoryInput) (*entity.Category, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	var out *entity.Category
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		v, err := u.categoryWriter.Update(ctx, gateway.UpdateCategoryParams{
			ID:        in.ID,
			Name:      strings.TrimSpace(in.Name),
			ParentID:  in.ParentID,
			Color:     in.Color,
			SortOrder: in.SortOrder,
		})
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "update category")
	}
	return out, nil
}

func (u *catalogUsecase) DeleteCategory(ctx context.Context, in input.DeleteCategoryInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		return u.categoryWriter.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete category")
	}
	return nil
}

// ----- Products ------------------------------------------------------------

func (u *catalogUsecase) ListProducts(ctx context.Context, orgID string) ([]*entity.ProductListItem, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.ProductListItem
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.productDetail.ListSummaries(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list products")
	}
	return out, nil
}

func (u *catalogUsecase) GetProduct(ctx context.Context, in input.GetProductInput) (*entity.ProductDetail, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	var out *entity.ProductDetail
	if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		v, err := u.productDetail.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "get product")
	}
	return out, nil
}

func (u *catalogUsecase) CreateProduct(ctx context.Context, in input.CreateProductInput) (*entity.ProductDetail, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if strings.TrimSpace(in.Product.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	if len(in.Product.Variants) == 0 {
		return nil, errors.BadRequest("at least one variant is required")
	}
	var productID string
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		base, err := u.productWriter.CreateBase(ctx, gateway.CreateProductBaseParams{
			OrgID:       in.OrgID,
			Name:        strings.TrimSpace(in.Product.Name),
			Description: in.Product.Description,
			CategoryID:  in.Product.CategoryID,
			IsActive:    in.Product.IsActive,
			SortOrder:   in.Product.SortOrder,
		})
		if err != nil {
			return err
		}
		productID = base.ID
		return u.writeProductGraph(ctx, in.OrgID, base.ID, in.Product)
	}); err != nil {
		return nil, errors.Wrap(err, "create product")
	}
	return u.GetProduct(ctx, input.GetProductInput{ID: productID, OrgID: in.OrgID})
}

func (u *catalogUsecase) UpdateProduct(ctx context.Context, in input.UpdateProductInput) (*entity.ProductDetail, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if strings.TrimSpace(in.Product.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	if len(in.Product.Variants) == 0 {
		return nil, errors.BadRequest("at least one variant is required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		if _, err := u.productWriter.UpdateBase(ctx, gateway.UpdateProductBaseParams{
			ID:          in.ID,
			Name:        strings.TrimSpace(in.Product.Name),
			Description: in.Product.Description,
			CategoryID:  in.Product.CategoryID,
			IsActive:    in.Product.IsActive,
			SortOrder:   in.Product.SortOrder,
		}); err != nil {
			return err
		}
		if err := u.productWriter.DeleteImagesByProduct(ctx, in.ID); err != nil {
			return err
		}
		if err := u.productWriter.SoftDeleteVariantsByProduct(ctx, in.ID); err != nil {
			return err
		}
		if err := u.productWriter.DeleteOptionsByProduct(ctx, in.ID); err != nil {
			return err
		}
		return u.writeProductGraph(ctx, in.OrgID, in.ID, in.Product)
	}); err != nil {
		return nil, errors.Wrap(err, "update product")
	}
	return u.GetProduct(ctx, input.GetProductInput{ID: in.ID, OrgID: in.OrgID})
}

func (u *catalogUsecase) DeleteProduct(ctx context.Context, in input.DeleteProductInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		return u.productWriter.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete product")
	}
	return nil
}

// writeProductGraph inserts options, option values, variants, variant-option-value
// joins and images for the product. Must run inside a tenant-scoped transaction.
func (u *catalogUsecase) writeProductGraph(ctx context.Context, orgID, productID string, p input.ProductInput) error {
	// Options + values
	valueIDByOptionIdxAndLabel := make(map[int]map[string]string, len(p.Options))
	for i, opt := range p.Options {
		name := strings.TrimSpace(opt.Name)
		if name == "" {
			continue
		}
		createdOpt, err := u.productWriter.InsertOption(ctx, gateway.CreateProductOptionParams{
			OrgID:     orgID,
			ProductID: productID,
			Name:      name,
			SortOrder: int32(i),
		})
		if err != nil {
			return err
		}
		valueIDByOptionIdxAndLabel[i] = make(map[string]string, len(opt.Values))
		for j, val := range opt.Values {
			val = strings.TrimSpace(val)
			if val == "" {
				continue
			}
			createdVal, err := u.productWriter.InsertOptionValue(ctx, gateway.CreateProductOptionValueParams{
				OrgID:     orgID,
				OptionID:  createdOpt.ID,
				Value:     val,
				SortOrder: int32(j),
			})
			if err != nil {
				return err
			}
			valueIDByOptionIdxAndLabel[i][val] = createdVal.ID
		}
	}

	// Variants + links
	for vi, v := range p.Variants {
		created, err := u.productWriter.InsertVariant(ctx, gateway.CreateProductVariantParams{
			OrgID:      orgID,
			ProductID:  productID,
			Name:       fallback(v.Name, "Default"),
			SKU:        v.SKU,
			Barcode:    v.Barcode,
			Price:      fallback(v.Price, "0"),
			CostPrice:  fallback(v.CostPrice, "0"),
			TrackStock: v.TrackStock,
			IsActive:   v.IsActive,
			SortOrder:  int32(vi),
		})
		if err != nil {
			return err
		}
		for oi, label := range v.OptionValues {
			labels := valueIDByOptionIdxAndLabel[oi]
			if labels == nil {
				continue
			}
			id, ok := labels[strings.TrimSpace(label)]
			if !ok {
				continue
			}
			if err := u.productWriter.InsertVariantOptionValue(ctx, orgID, created.ID, id); err != nil {
				return err
			}
		}
	}

	// Images (attached to the product; variant_id omitted for now)
	for i, img := range p.Images {
		url := strings.TrimSpace(img.URL)
		if url == "" {
			continue
		}
		if _, err := u.productWriter.InsertImage(ctx, gateway.CreateProductImageParams{
			OrgID:     orgID,
			ProductID: productID,
			URL:       url,
			SortOrder: int32(i),
		}); err != nil {
			return err
		}
	}
	return nil
}

// ----- CSV Import ----------------------------------------------------------

const csvExpectedHeader = "name,category,description,sku,barcode,price,cost_price,is_active"

func (u *catalogUsecase) ParseImportCsv(ctx context.Context, in input.ParseImportCsvInput) (*input.ParseImportCsvResult, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if len(in.CsvData) == 0 {
		return nil, errors.BadRequest("csv data is empty")
	}

	reader := csv.NewReader(bytes.NewReader(in.CsvData))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	rows := make([]input.CsvProductRow, 0)
	warnings := make([]string, 0)
	valid, invalid := int32(0), int32(0)

	var header []string
	lineNum := 0
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.BadRequest("csv parse error: " + err.Error())
		}
		lineNum++
		if lineNum == 1 {
			header = rec
			if !headerMatches(header) {
				warnings = append(warnings, "unexpected header; expected: "+csvExpectedHeader)
			}
			continue
		}
		row := rowFromRecord(header, rec)
		row.Errors = validateRow(row)
		if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
			existing, err := u.productDetail.GetByName(ctx, row.Name)
			if err == nil && existing != nil {
				row.Exists = true
				row.ExistingID = existing.ID
			}
			return nil
		}); err != nil {
			// swallow lookup errors; treat as not-exists
		}
		if len(row.Errors) == 0 {
			valid++
		} else {
			invalid++
		}
		rows = append(rows, row)
	}

	return &input.ParseImportCsvResult{
		Rows:       rows,
		ValidCount: valid,
		ErrorCount: invalid,
		Warnings:   warnings,
	}, nil
}

func (u *catalogUsecase) ImportProducts(ctx context.Context, in input.ImportProductsInput) (*input.ImportProductsResult, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if len(in.Items) == 0 {
		return nil, errors.BadRequest("no items to import")
	}

	result := &input.ImportProductsResult{}

	// Resolve category names to ids once up front
	categoryByName := make(map[string]string)
	if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		cats, err := u.categoryReader.List(ctx)
		if err != nil {
			return err
		}
		for _, c := range cats {
			categoryByName[strings.ToLower(c.Name)] = c.ID
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "load categories")
	}

	for _, item := range in.Items {
		if len(item.Row.Errors) > 0 {
			result.Skipped++
			continue
		}
		productIn := rowToProductInput(item.Row, categoryByName)

		if item.Row.Exists {
			if !item.OverrideExisting {
				result.Skipped++
				continue
			}
			id := item.ExistingID
			if id == "" {
				id = item.Row.ExistingID
			}
			if _, err := u.UpdateProduct(ctx, input.UpdateProductInput{
				ID: id, OrgID: in.OrgID, Product: productIn,
			}); err != nil {
				result.Errors = append(result.Errors, item.Row.Name+": "+errors.GetPublicMessage(err))
				result.Skipped++
				continue
			}
			result.Updated++
			continue
		}

		if _, err := u.CreateProduct(ctx, input.CreateProductInput{
			OrgID: in.OrgID, Product: productIn,
		}); err != nil {
			result.Errors = append(result.Errors, item.Row.Name+": "+errors.GetPublicMessage(err))
			result.Skipped++
			continue
		}
		result.Created++
	}
	return result, nil
}

func rowFromRecord(header, rec []string) input.CsvProductRow {
	get := func(name string) string {
		for i, h := range header {
			if strings.EqualFold(strings.TrimSpace(h), name) && i < len(rec) {
				return strings.TrimSpace(rec[i])
			}
		}
		return ""
	}
	return input.CsvProductRow{
		Name:         get("name"),
		CategoryName: get("category"),
		Description:  get("description"),
		SKU:          get("sku"),
		Barcode:      get("barcode"),
		Price:        get("price"),
		CostPrice:    get("cost_price"),
		IsActive:     get("is_active"),
	}
}

func validateRow(r input.CsvProductRow) []string {
	errs := make([]string, 0)
	if r.Name == "" {
		errs = append(errs, "name is required")
	}
	if r.Price == "" {
		errs = append(errs, "price is required")
	}
	return errs
}

func rowToProductInput(r input.CsvProductRow, categoryByName map[string]string) input.ProductInput {
	isActive := strings.EqualFold(r.IsActive, "true") || r.IsActive == "1" || r.IsActive == ""
	return input.ProductInput{
		Name:        r.Name,
		Description: r.Description,
		CategoryID:  categoryByName[strings.ToLower(r.CategoryName)],
		IsActive:    isActive,
		Variants: []input.VariantInput{
			{
				Name:       "Default",
				SKU:        r.SKU,
				Barcode:    r.Barcode,
				Price:      r.Price,
				CostPrice:  r.CostPrice,
				TrackStock: true,
				IsActive:   isActive,
			},
		},
	}
}

func headerMatches(header []string) bool {
	expected := strings.Split(csvExpectedHeader, ",")
	if len(header) < len(expected) {
		return false
	}
	for i, want := range expected {
		if !strings.EqualFold(strings.TrimSpace(header[i]), want) {
			return false
		}
	}
	return true
}

func fallback(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

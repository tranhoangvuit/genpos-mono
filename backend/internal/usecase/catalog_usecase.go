package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"strconv"
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
			TaxClassID: v.TaxClassID,
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

// Expected column headers. Rows with the same Title collapse into a single
// product with one variant per row.
var csvExpectedColumns = []string{
	"Title", "Description", "Status", "SKU", "Barcode",
	"Option1 name", "Option1 value",
	"Option2 name", "Option2 value",
	"Option3 name", "Option3 value",
	"Price", "Cost price", "Inventory quantity",
}

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

	warnings := make([]string, 0)

	var header []string
	// Preserve product insertion order.
	order := make([]string, 0)
	byTitle := make(map[string]*input.CsvProductRow)

	lineNum := 0
	inventoryMentioned := false
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
				warnings = append(warnings, "unexpected header; expected: "+strings.Join(csvExpectedColumns, ","))
			}
			continue
		}
		get := func(name string) string {
			for i, h := range header {
				if strings.EqualFold(strings.TrimSpace(h), name) && i < len(rec) {
					return strings.TrimSpace(rec[i])
				}
			}
			return ""
		}
		title := get("Title")
		if title == "" {
			// Rows with no title are skipped silently — likely blank lines.
			continue
		}
		variant := input.CsvVariantRow{
			SKU:               get("SKU"),
			Barcode:           get("Barcode"),
			Option1Name:       get("Option1 name"),
			Option1Value:      get("Option1 value"),
			Option2Name:       get("Option2 name"),
			Option2Value:      get("Option2 value"),
			Option3Name:       get("Option3 name"),
			Option3Value:      get("Option3 value"),
			Price:             get("Price"),
			CostPrice:         get("Cost price"),
			InventoryQuantity: get("Inventory quantity"),
		}
		if variant.InventoryQuantity != "" {
			inventoryMentioned = true
		}
		group, ok := byTitle[title]
		if !ok {
			group = &input.CsvProductRow{
				Name:        title,
				Description: get("Description"),
				Status:      get("Status"),
			}
			byTitle[title] = group
			order = append(order, title)
		}
		group.Variants = append(group.Variants, variant)
	}

	if inventoryMentioned {
		warnings = append(warnings, "Inventory quantity is ignored — stock levels must be set via a stock take in a store context")
	}

	rows := make([]input.CsvProductRow, 0, len(order))
	valid, invalid := int32(0), int32(0)
	for _, title := range order {
		row := *byTitle[title]
		row.Errors = validateGroup(row)
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

	for _, item := range in.Items {
		if len(item.Row.Errors) > 0 {
			result.Skipped++
			continue
		}
		productIn := groupToProductInput(item.Row)

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

func validateGroup(r input.CsvProductRow) []string {
	errs := make([]string, 0)
	if strings.TrimSpace(r.Name) == "" {
		errs = append(errs, "title is required")
	}
	if len(r.Variants) == 0 {
		errs = append(errs, "at least one variant row is required")
		return errs
	}
	for i, v := range r.Variants {
		if strings.TrimSpace(v.Price) == "" {
			errs = append(errs, "variant "+itoa(i+1)+": price is required")
		}
	}
	return errs
}

// groupToProductInput builds a ProductInput from a grouped CSV row.
// Options are inferred per-axis from the variants: each axis becomes a
// ProductOption (with the first non-empty name on that axis), and unique
// values across variants become its values. A product with no option names
// across any variant emits no options and a single "Default" variant name.
func groupToProductInput(r input.CsvProductRow) input.ProductInput {
	isActive := parseStatus(r.Status)

	// Per-axis accumulated option name + ordered unique values.
	type axisAcc struct {
		name   string
		values []string
		seen   map[string]bool
	}
	axes := [3]axisAcc{{seen: map[string]bool{}}, {seen: map[string]bool{}}, {seen: map[string]bool{}}}
	pickOption := func(axis int, v input.CsvVariantRow) (string, string) {
		switch axis {
		case 0:
			return v.Option1Name, v.Option1Value
		case 1:
			return v.Option2Name, v.Option2Value
		default:
			return v.Option3Name, v.Option3Value
		}
	}
	for _, v := range r.Variants {
		for axis := 0; axis < 3; axis++ {
			name, val := pickOption(axis, v)
			name = strings.TrimSpace(name)
			val = strings.TrimSpace(val)
			if name != "" && axes[axis].name == "" {
				axes[axis].name = name
			}
			if val != "" && !axes[axis].seen[val] {
				axes[axis].seen[val] = true
				axes[axis].values = append(axes[axis].values, val)
			}
		}
	}

	options := make([]input.OptionInput, 0, 3)
	axisIncluded := [3]bool{}
	for axis := 0; axis < 3; axis++ {
		if axes[axis].name == "" || len(axes[axis].values) == 0 {
			continue
		}
		axisIncluded[axis] = true
		options = append(options, input.OptionInput{
			Name:   axes[axis].name,
			Values: axes[axis].values,
		})
	}

	variants := make([]input.VariantInput, 0, len(r.Variants))
	for i, v := range r.Variants {
		optValues := make([]string, 0, len(options))
		name := ""
		for axis := 0; axis < 3; axis++ {
			if !axisIncluded[axis] {
				continue
			}
			_, val := pickOption(axis, v)
			val = strings.TrimSpace(val)
			optValues = append(optValues, val)
			if val != "" {
				if name != "" {
					name += " / "
				}
				name += val
			}
		}
		if name == "" {
			name = "Default"
		}
		variants = append(variants, input.VariantInput{
			Name:         name,
			SKU:          v.SKU,
			Barcode:      v.Barcode,
			Price:        v.Price,
			CostPrice:    v.CostPrice,
			TrackStock:   true,
			IsActive:     isActive,
			SortOrder:    int32(i),
			OptionValues: optValues,
		})
	}

	return input.ProductInput{
		Name:        r.Name,
		Description: r.Description,
		IsActive:    isActive,
		Options:     options,
		Variants:    variants,
	}
}

func parseStatus(s string) bool {
	t := strings.ToLower(strings.TrimSpace(s))
	if t == "" {
		return true
	}
	switch t {
	case "inactive", "false", "0", "no", "disabled", "draft":
		return false
	default:
		return true
	}
}

func headerMatches(header []string) bool {
	if len(header) < len(csvExpectedColumns) {
		return false
	}
	for i, want := range csvExpectedColumns {
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

func itoa(n int) string {
	return strconv.Itoa(n)
}

package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"strings"
	"time"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type customerUsecase struct {
	tenantDB            gateway.TenantDB
	customerReader      gateway.CustomerReader
	customerWriter      gateway.CustomerWriter
	customerGroupReader gateway.CustomerGroupReader
	customerGroupWriter gateway.CustomerGroupWriter
}

// NewCustomerUsecase constructs a CustomerUsecase.
func NewCustomerUsecase(
	tenantDB gateway.TenantDB,
	customerReader gateway.CustomerReader,
	customerWriter gateway.CustomerWriter,
	customerGroupReader gateway.CustomerGroupReader,
	customerGroupWriter gateway.CustomerGroupWriter,
) CustomerUsecase {
	return &customerUsecase{
		tenantDB:            tenantDB,
		customerReader:      customerReader,
		customerWriter:      customerWriter,
		customerGroupReader: customerGroupReader,
		customerGroupWriter: customerGroupWriter,
	}
}

// ----- Customers -----------------------------------------------------------

func (u *customerUsecase) ListCustomers(ctx context.Context, orgID string) ([]*entity.CustomerListItem, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.CustomerListItem
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.customerReader.ListSummaries(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list customers")
	}
	return out, nil
}

func (u *customerUsecase) GetCustomer(ctx context.Context, in input.GetCustomerInput) (*entity.Customer, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	var out *entity.Customer
	if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		c, err := u.customerReader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		ids, err := u.customerReader.ListGroupIDsByCustomer(ctx, in.ID)
		if err != nil {
			return err
		}
		c.GroupIDs = ids
		out = c
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "get customer")
	}
	return out, nil
}

func (u *customerUsecase) ListCustomerGroups(ctx context.Context, orgID string) ([]*entity.CustomerGroup, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.CustomerGroup
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.customerGroupReader.List(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list customer groups")
	}
	return out, nil
}

func (u *customerUsecase) CreateCustomer(ctx context.Context, in input.CreateCustomerInput) (*entity.Customer, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if strings.TrimSpace(in.Customer.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	var out *entity.Customer
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		c, err := u.customerWriter.Create(ctx, gateway.CreateCustomerParams{
			OrgID:       in.OrgID,
			Name:        strings.TrimSpace(in.Customer.Name),
			Email:       strings.TrimSpace(in.Customer.Email),
			Phone:       strings.TrimSpace(in.Customer.Phone),
			Notes:       in.Customer.Notes,
			Code:        strings.TrimSpace(in.Customer.Code),
			Address:     strings.TrimSpace(in.Customer.Address),
			Company:     strings.TrimSpace(in.Customer.Company),
			TaxCode:     strings.TrimSpace(in.Customer.TaxCode),
			DateOfBirth: in.Customer.DateOfBirth,
			Gender:      strings.TrimSpace(in.Customer.Gender),
			Facebook:    strings.TrimSpace(in.Customer.Facebook),
			IsActive:    in.Customer.IsActive,
		})
		if err != nil {
			return err
		}
		if err := u.customerWriter.ReplaceGroups(ctx, in.OrgID, c.ID, in.Customer.GroupIDs); err != nil {
			return err
		}
		c.GroupIDs = dedupe(in.Customer.GroupIDs)
		out = c
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "create customer")
	}
	return out, nil
}

func (u *customerUsecase) UpdateCustomer(ctx context.Context, in input.UpdateCustomerInput) (*entity.Customer, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if strings.TrimSpace(in.Customer.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	var out *entity.Customer
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		c, err := u.customerWriter.Update(ctx, gateway.UpdateCustomerParams{
			ID:          in.ID,
			Name:        strings.TrimSpace(in.Customer.Name),
			Email:       strings.TrimSpace(in.Customer.Email),
			Phone:       strings.TrimSpace(in.Customer.Phone),
			Notes:       in.Customer.Notes,
			Code:        strings.TrimSpace(in.Customer.Code),
			Address:     strings.TrimSpace(in.Customer.Address),
			Company:     strings.TrimSpace(in.Customer.Company),
			TaxCode:     strings.TrimSpace(in.Customer.TaxCode),
			DateOfBirth: in.Customer.DateOfBirth,
			Gender:      strings.TrimSpace(in.Customer.Gender),
			Facebook:    strings.TrimSpace(in.Customer.Facebook),
			IsActive:    in.Customer.IsActive,
		})
		if err != nil {
			return err
		}
		if err := u.customerWriter.ReplaceGroups(ctx, in.OrgID, c.ID, in.Customer.GroupIDs); err != nil {
			return err
		}
		c.GroupIDs = dedupe(in.Customer.GroupIDs)
		out = c
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "update customer")
	}
	return out, nil
}

func (u *customerUsecase) DeleteCustomer(ctx context.Context, in input.DeleteCustomerInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		return u.customerWriter.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete customer")
	}
	return nil
}

// ----- Customer groups -----------------------------------------------------

func (u *customerUsecase) CreateCustomerGroup(ctx context.Context, in input.CreateCustomerGroupInput) (*entity.CustomerGroup, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if strings.TrimSpace(in.Group.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	if err := validateDiscount(in.Group.DiscountType, in.Group.DiscountValue); err != nil {
		return nil, err
	}
	var out *entity.CustomerGroup
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		g, err := u.customerGroupWriter.Create(ctx, gateway.CreateCustomerGroupParams{
			OrgID:         in.OrgID,
			Name:          strings.TrimSpace(in.Group.Name),
			Description:   in.Group.Description,
			DiscountType:  in.Group.DiscountType,
			DiscountValue: in.Group.DiscountValue,
		})
		if err != nil {
			return err
		}
		out = g
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "create customer group")
	}
	return out, nil
}

func (u *customerUsecase) UpdateCustomerGroup(ctx context.Context, in input.UpdateCustomerGroupInput) (*entity.CustomerGroup, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if strings.TrimSpace(in.Group.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	if err := validateDiscount(in.Group.DiscountType, in.Group.DiscountValue); err != nil {
		return nil, err
	}
	var out *entity.CustomerGroup
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		g, err := u.customerGroupWriter.Update(ctx, gateway.UpdateCustomerGroupParams{
			ID:            in.ID,
			Name:          strings.TrimSpace(in.Group.Name),
			Description:   in.Group.Description,
			DiscountType:  in.Group.DiscountType,
			DiscountValue: in.Group.DiscountValue,
		})
		if err != nil {
			return err
		}
		out = g
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "update customer group")
	}
	return out, nil
}

func (u *customerUsecase) DeleteCustomerGroup(ctx context.Context, in input.DeleteCustomerGroupInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		return u.customerGroupWriter.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete customer group")
	}
	return nil
}

// ----- CSV import ----------------------------------------------------------

// customerColumnAliases maps canonical column keys to the synonyms that may
// appear in real-world exports (English + Vietnamese from KiotViet/etc.).
// Matching is case-insensitive and trimmed.
var customerColumnAliases = map[string][]string{
	"name":          {"name", "full name", "customer name", "tên khách hàng", "ten khach hang", "ho ten", "họ tên"},
	"email":         {"email", "e-mail"},
	"phone":         {"phone", "phone number", "mobile", "điện thoại", "dien thoai", "so dien thoai"},
	"code":          {"code", "customer code", "mã khách hàng", "ma khach hang"},
	"address":       {"address", "địa chỉ", "dia chi"},
	"company":       {"company", "công ty", "cong ty"},
	"tax_code":      {"tax code", "tax_code", "mã số thuế", "ma so thue", "mst"},
	"date_of_birth": {"date_of_birth", "date of birth", "dob", "birthday", "ngày sinh", "ngay sinh"},
	"gender":        {"gender", "giới tính", "gioi tinh"},
	"facebook":      {"facebook", "fb"},
	"groups":        {"groups", "customer group", "nhóm khách hàng", "nhom khach hang", "loại khách", "loai khach"},
	"notes":         {"notes", "note", "ghi chú", "ghi chu"},
	"status":        {"status", "active", "trạng thái", "trang thai"},
}

// customerCanonicalHeader is advisory — any subset is accepted.
const customerCsvExpectedHeader = "name,email,phone,code,address,company,tax_code,date_of_birth,gender,facebook,groups,notes,status"

func (u *customerUsecase) ParseImportCustomerCsv(ctx context.Context, in input.ParseImportCustomerCsvInput) (*input.ParseImportCustomerCsvResult, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if len(in.CsvData) == 0 {
		return nil, errors.BadRequest("csv data is empty")
	}

	reader := csv.NewReader(bytes.NewReader(in.CsvData))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	rows := make([]input.CsvCustomerRow, 0)
	warnings := make([]string, 0)
	valid, invalid := int32(0), int32(0)

	var colIdx map[string]int
	lineNum := 0
	unmatchedHeaders := make([]string, 0)
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
			colIdx, unmatchedHeaders = mapCustomerHeader(rec)
			if _, ok := colIdx["name"]; !ok {
				warnings = append(warnings, "missing required column: name")
			}
			if len(unmatchedHeaders) > 0 {
				warnings = append(warnings, "ignored columns: "+strings.Join(unmatchedHeaders, ", "))
			}
			continue
		}
		row := customerRowFromRecord(colIdx, rec)
		row.Errors = validateCustomerRow(row)
		if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
			// Dedupe priority: code → phone → email. Suppress not-found so the row stays fresh.
			if row.Code != "" {
				if existing, err := u.customerReader.GetByCode(ctx, in.OrgID, row.Code); err == nil && existing != nil {
					row.Exists = true
					row.ExistingID = existing.ID
					return nil
				} else if err != nil && !isNotFound(err) {
					return err
				}
			}
			if row.Phone != "" {
				if existing, err := u.customerReader.GetByPhone(ctx, row.Phone); err == nil && existing != nil {
					row.Exists = true
					row.ExistingID = existing.ID
					return nil
				} else if err != nil && !isNotFound(err) {
					return err
				}
			}
			if row.Email != "" {
				if existing, err := u.customerReader.GetByEmail(ctx, row.Email); err == nil && existing != nil {
					row.Exists = true
					row.ExistingID = existing.ID
					return nil
				} else if err != nil && !isNotFound(err) {
					return err
				}
			}
			return nil
		}); err != nil {
			// swallow lookup errors
		}
		if len(row.Errors) == 0 {
			valid++
		} else {
			invalid++
		}
		rows = append(rows, row)
	}

	return &input.ParseImportCustomerCsvResult{
		Rows:       rows,
		ValidCount: valid,
		ErrorCount: invalid,
		Warnings:   warnings,
	}, nil
}

func (u *customerUsecase) ImportCustomers(ctx context.Context, in input.ImportCustomersInput) (*input.ImportCustomersResult, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if len(in.Items) == 0 {
		return nil, errors.BadRequest("no items to import")
	}

	result := &input.ImportCustomersResult{}

	// Resolve existing group names → ids. Auto-create missing groups so the
	// import preserves segmentation coming from the source system.
	groupByName := make(map[string]string)
	if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		gs, err := u.customerGroupReader.List(ctx)
		if err != nil {
			return err
		}
		for _, g := range gs {
			groupByName[strings.ToLower(g.Name)] = g.ID
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "load customer groups")
	}
	for _, item := range in.Items {
		for _, name := range splitGroupNames(item.Row.Groups) {
			key := strings.ToLower(name)
			if _, ok := groupByName[key]; ok {
				continue
			}
			g, err := u.CreateCustomerGroup(ctx, input.CreateCustomerGroupInput{
				OrgID: in.OrgID,
				Group: input.CustomerGroupInput{Name: name},
			})
			if err != nil {
				return nil, errors.Wrap(err, "auto-create customer group")
			}
			groupByName[key] = g.ID
		}
	}

	for _, item := range in.Items {
		if len(item.Row.Errors) > 0 {
			result.Skipped++
			continue
		}
		custIn := customerRowToInput(item.Row, groupByName)

		if item.Row.Exists {
			if !item.OverrideExisting {
				result.Skipped++
				continue
			}
			id := item.ExistingID
			if id == "" {
				id = item.Row.ExistingID
			}
			if _, err := u.UpdateCustomer(ctx, input.UpdateCustomerInput{
				ID: id, OrgID: in.OrgID, Customer: custIn,
			}); err != nil {
				result.Errors = append(result.Errors, item.Row.Name+": "+errors.GetPublicMessage(err))
				result.Skipped++
				continue
			}
			result.Updated++
			continue
		}

		if _, err := u.CreateCustomer(ctx, input.CreateCustomerInput{
			OrgID: in.OrgID, Customer: custIn,
		}); err != nil {
			result.Errors = append(result.Errors, item.Row.Name+": "+errors.GetPublicMessage(err))
			result.Skipped++
			continue
		}
		result.Created++
	}
	return result, nil
}

// ----- helpers -------------------------------------------------------------

// mapCustomerHeader resolves each CSV header cell to a canonical key via
// customerColumnAliases. Returns the column index per canonical key plus a
// list of header cells that did not map to anything.
func mapCustomerHeader(header []string) (map[string]int, []string) {
	aliasToKey := make(map[string]string, 64)
	for key, aliases := range customerColumnAliases {
		for _, a := range aliases {
			aliasToKey[strings.ToLower(strings.TrimSpace(a))] = key
		}
	}
	idx := make(map[string]int, len(header))
	unmatched := make([]string, 0)
	for i, h := range header {
		cell := strings.ToLower(strings.TrimSpace(h))
		if cell == "" {
			continue
		}
		if key, ok := aliasToKey[cell]; ok {
			if _, exists := idx[key]; !exists {
				idx[key] = i
			}
		} else {
			unmatched = append(unmatched, h)
		}
	}
	return idx, unmatched
}

func customerRowFromRecord(colIdx map[string]int, rec []string) input.CsvCustomerRow {
	get := func(key string) string {
		if i, ok := colIdx[key]; ok && i < len(rec) {
			return strings.TrimSpace(rec[i])
		}
		return ""
	}
	return input.CsvCustomerRow{
		Name:        get("name"),
		Email:       get("email"),
		Phone:       get("phone"),
		Notes:       get("notes"),
		Groups:      get("groups"),
		Code:        get("code"),
		Address:     get("address"),
		Company:     get("company"),
		TaxCode:     get("tax_code"),
		DateOfBirth: normalizeDOB(get("date_of_birth")),
		Gender:      normalizeGender(get("gender")),
		Facebook:    get("facebook"),
		Status:      strings.ToLower(get("status")),
	}
}

func validateCustomerRow(r input.CsvCustomerRow) []string {
	errs := make([]string, 0)
	if r.Name == "" {
		errs = append(errs, "name is required")
	}
	if r.DateOfBirth != "" {
		if _, err := time.Parse("2006-01-02", r.DateOfBirth); err != nil {
			errs = append(errs, "date_of_birth must be YYYY-MM-DD")
		}
	}
	return errs
}

func customerRowToInput(r input.CsvCustomerRow, groupByName map[string]string) input.CustomerInput {
	groupIDs := make([]string, 0)
	for _, name := range splitGroupNames(r.Groups) {
		if id, ok := groupByName[strings.ToLower(name)]; ok {
			groupIDs = append(groupIDs, id)
		}
	}
	var dob time.Time
	if r.DateOfBirth != "" {
		if t, err := time.Parse("2006-01-02", r.DateOfBirth); err == nil {
			dob = t
		}
	}
	return input.CustomerInput{
		Name:        r.Name,
		Email:       r.Email,
		Phone:       r.Phone,
		Notes:       r.Notes,
		GroupIDs:    groupIDs,
		Code:        r.Code,
		Address:     r.Address,
		Company:     r.Company,
		TaxCode:     r.TaxCode,
		DateOfBirth: dob,
		Gender:      r.Gender,
		Facebook:    r.Facebook,
		IsActive:    parseCustomerActive(r.Status),
	}
}

func splitGroupNames(s string) []string {
	out := make([]string, 0)
	for _, name := range strings.Split(s, ",") {
		name = strings.TrimSpace(name)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

// normalizeDOB accepts YYYY-MM-DD, DD/MM/YYYY, or DD-MM-YYYY and returns
// the canonical YYYY-MM-DD form; invalid or empty input returns "".
func normalizeDOB(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	for _, layout := range []string{"2006-01-02", "02/01/2006", "02-01-2006", "2006/01/02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return s // leave as-is so validateCustomerRow can flag it
}

func normalizeGender(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "m", "male", "nam":
		return "male"
	case "f", "female", "nữ", "nu":
		return "female"
	default:
		return ""
	}
}

func parseCustomerActive(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "inactive", "false", "0", "no", "disabled", "ngưng hoạt động", "ngung hoat dong":
		return false
	default:
		return true
	}
}

func validateDiscount(discountType, discountValue string) error {
	if discountType == "" && discountValue == "" {
		return nil
	}
	if discountType != "percentage" && discountType != "fixed" {
		return errors.BadRequest("discount type must be percentage or fixed")
	}
	return nil
}

func dedupe(ss []string) []string {
	seen := make(map[string]struct{}, len(ss))
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func isNotFound(err error) bool {
	return err != nil && errors.GetCode(err) == errors.CodeNotFound
}

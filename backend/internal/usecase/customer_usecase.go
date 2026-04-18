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
			OrgID: in.OrgID,
			Name:  strings.TrimSpace(in.Customer.Name),
			Email: strings.TrimSpace(in.Customer.Email),
			Phone: strings.TrimSpace(in.Customer.Phone),
			Notes: in.Customer.Notes,
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
			ID:    in.ID,
			Name:  strings.TrimSpace(in.Customer.Name),
			Email: strings.TrimSpace(in.Customer.Email),
			Phone: strings.TrimSpace(in.Customer.Phone),
			Notes: in.Customer.Notes,
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

const customerCsvExpectedHeader = "name,email,phone,notes,groups"

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
			if !customerHeaderMatches(header) {
				warnings = append(warnings, "unexpected header; expected: "+customerCsvExpectedHeader)
			}
			continue
		}
		row := customerRowFromRecord(header, rec)
		row.Errors = validateCustomerRow(row)
		if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
			// Dedupe by email first, then phone. Suppress not-found so the row stays fresh.
			if row.Email != "" {
				if existing, err := u.customerReader.GetByEmail(ctx, row.Email); err == nil && existing != nil {
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

	// Resolve group names to ids once
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

func customerRowFromRecord(header, rec []string) input.CsvCustomerRow {
	get := func(name string) string {
		for i, h := range header {
			if strings.EqualFold(strings.TrimSpace(h), name) && i < len(rec) {
				return strings.TrimSpace(rec[i])
			}
		}
		return ""
	}
	return input.CsvCustomerRow{
		Name:   get("name"),
		Email:  get("email"),
		Phone:  get("phone"),
		Notes:  get("notes"),
		Groups: get("groups"),
	}
}

func validateCustomerRow(r input.CsvCustomerRow) []string {
	errs := make([]string, 0)
	if r.Name == "" {
		errs = append(errs, "name is required")
	}
	return errs
}

func customerRowToInput(r input.CsvCustomerRow, groupByName map[string]string) input.CustomerInput {
	groupIDs := make([]string, 0)
	for _, name := range strings.Split(r.Groups, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if id, ok := groupByName[strings.ToLower(name)]; ok {
			groupIDs = append(groupIDs, id)
		}
	}
	return input.CustomerInput{
		Name:     r.Name,
		Email:    r.Email,
		Phone:    r.Phone,
		Notes:    r.Notes,
		GroupIDs: groupIDs,
	}
}

func customerHeaderMatches(header []string) bool {
	expected := strings.Split(customerCsvExpectedHeader, ",")
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

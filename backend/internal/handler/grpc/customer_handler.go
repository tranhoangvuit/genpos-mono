package grpc

import (
	"context"
	"log/slog"
	"time"

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

// CustomerHandler implements CustomerServiceHandler.
type CustomerHandler struct {
	genposv1connect.UnimplementedCustomerServiceHandler
	logger  *slog.Logger
	usecase usecase.CustomerUsecase
}

// NewCustomerHandler constructs a CustomerHandler.
func NewCustomerHandler(logger *slog.Logger, uc usecase.CustomerUsecase) *CustomerHandler {
	return &CustomerHandler{logger: logger, usecase: uc}
}

// ----- Customers -----------------------------------------------------------

func (h *CustomerHandler) ListCustomers(
	ctx context.Context,
	_ *connect.Request[genposv1.ListCustomersRequest],
) (*connect.Response[genposv1.ListCustomersResponse], error) {
	authCtx, err := h.requireCustomerAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListCustomers(ctx, authCtx.OrgID)
	if err != nil {
		return nil, h.logAndConvertCustomer(ctx, "list customers", err)
	}
	pb := make([]*genposv1.CustomerListItem, 0, len(items))
	for _, c := range items {
		pb = append(pb, &genposv1.CustomerListItem{
			Id:         c.ID,
			Name:       c.Name,
			Email:      c.Email,
			Phone:      c.Phone,
			GroupNames: c.GroupNames,
			Code:       c.Code,
			Company:    c.Company,
			IsActive:   c.IsActive,
		})
	}
	return connect.NewResponse(&genposv1.ListCustomersResponse{Customers: pb}), nil
}

func (h *CustomerHandler) GetCustomer(
	ctx context.Context,
	req *connect.Request[genposv1.GetCustomerRequest],
) (*connect.Response[genposv1.GetCustomerResponse], error) {
	authCtx, err := h.requireCustomerAuth(ctx)
	if err != nil {
		return nil, err
	}
	c, err := h.usecase.GetCustomer(ctx, input.GetCustomerInput{ID: req.Msg.GetId(), OrgID: authCtx.OrgID})
	if err != nil {
		return nil, h.logAndConvertCustomer(ctx, "get customer", err)
	}
	return connect.NewResponse(&genposv1.GetCustomerResponse{Customer: toCustomerProto(c)}), nil
}

func (h *CustomerHandler) ListCustomerGroups(
	ctx context.Context,
	_ *connect.Request[genposv1.ListCustomerGroupsRequest],
) (*connect.Response[genposv1.ListCustomerGroupsResponse], error) {
	authCtx, err := h.requireCustomerAuth(ctx)
	if err != nil {
		return nil, err
	}
	groups, err := h.usecase.ListCustomerGroups(ctx, authCtx.OrgID)
	if err != nil {
		return nil, h.logAndConvertCustomer(ctx, "list customer groups", err)
	}
	pb := make([]*genposv1.CustomerGroup, 0, len(groups))
	for _, g := range groups {
		pb = append(pb, toCustomerGroupProto(g))
	}
	return connect.NewResponse(&genposv1.ListCustomerGroupsResponse{Groups: pb}), nil
}

func (h *CustomerHandler) CreateCustomer(
	ctx context.Context,
	req *connect.Request[genposv1.CreateCustomerRequest],
) (*connect.Response[genposv1.CreateCustomerResponse], error) {
	authCtx, err := h.requireCustomerAuth(ctx)
	if err != nil {
		return nil, err
	}
	c, err := h.usecase.CreateCustomer(ctx, input.CreateCustomerInput{
		OrgID:    authCtx.OrgID,
		Customer: fromCustomerInputProto(req.Msg.GetCustomer()),
	})
	if err != nil {
		return nil, h.logAndConvertCustomer(ctx, "create customer", err)
	}
	return connect.NewResponse(&genposv1.CreateCustomerResponse{Customer: toCustomerProto(c)}), nil
}

func (h *CustomerHandler) UpdateCustomer(
	ctx context.Context,
	req *connect.Request[genposv1.UpdateCustomerRequest],
) (*connect.Response[genposv1.UpdateCustomerResponse], error) {
	authCtx, err := h.requireCustomerAuth(ctx)
	if err != nil {
		return nil, err
	}
	c, err := h.usecase.UpdateCustomer(ctx, input.UpdateCustomerInput{
		ID:       req.Msg.GetId(),
		OrgID:    authCtx.OrgID,
		Customer: fromCustomerInputProto(req.Msg.GetCustomer()),
	})
	if err != nil {
		return nil, h.logAndConvertCustomer(ctx, "update customer", err)
	}
	return connect.NewResponse(&genposv1.UpdateCustomerResponse{Customer: toCustomerProto(c)}), nil
}

func (h *CustomerHandler) DeleteCustomer(
	ctx context.Context,
	req *connect.Request[genposv1.DeleteCustomerRequest],
) (*connect.Response[genposv1.DeleteCustomerResponse], error) {
	authCtx, err := h.requireCustomerAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeleteCustomer(ctx, input.DeleteCustomerInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	}); err != nil {
		return nil, h.logAndConvertCustomer(ctx, "delete customer", err)
	}
	return connect.NewResponse(&genposv1.DeleteCustomerResponse{}), nil
}

// ----- Customer groups -----------------------------------------------------

func (h *CustomerHandler) CreateCustomerGroup(
	ctx context.Context,
	req *connect.Request[genposv1.CreateCustomerGroupRequest],
) (*connect.Response[genposv1.CreateCustomerGroupResponse], error) {
	authCtx, err := h.requireCustomerAuth(ctx)
	if err != nil {
		return nil, err
	}
	g, err := h.usecase.CreateCustomerGroup(ctx, input.CreateCustomerGroupInput{
		OrgID: authCtx.OrgID,
		Group: fromCustomerGroupInputProto(req.Msg.GetGroup()),
	})
	if err != nil {
		return nil, h.logAndConvertCustomer(ctx, "create customer group", err)
	}
	return connect.NewResponse(&genposv1.CreateCustomerGroupResponse{Group: toCustomerGroupProto(g)}), nil
}

func (h *CustomerHandler) UpdateCustomerGroup(
	ctx context.Context,
	req *connect.Request[genposv1.UpdateCustomerGroupRequest],
) (*connect.Response[genposv1.UpdateCustomerGroupResponse], error) {
	authCtx, err := h.requireCustomerAuth(ctx)
	if err != nil {
		return nil, err
	}
	g, err := h.usecase.UpdateCustomerGroup(ctx, input.UpdateCustomerGroupInput{
		ID:    req.Msg.GetId(),
		OrgID: authCtx.OrgID,
		Group: fromCustomerGroupInputProto(req.Msg.GetGroup()),
	})
	if err != nil {
		return nil, h.logAndConvertCustomer(ctx, "update customer group", err)
	}
	return connect.NewResponse(&genposv1.UpdateCustomerGroupResponse{Group: toCustomerGroupProto(g)}), nil
}

func (h *CustomerHandler) DeleteCustomerGroup(
	ctx context.Context,
	req *connect.Request[genposv1.DeleteCustomerGroupRequest],
) (*connect.Response[genposv1.DeleteCustomerGroupResponse], error) {
	authCtx, err := h.requireCustomerAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeleteCustomerGroup(ctx, input.DeleteCustomerGroupInput{
		ID: req.Msg.GetId(), OrgID: authCtx.OrgID,
	}); err != nil {
		return nil, h.logAndConvertCustomer(ctx, "delete customer group", err)
	}
	return connect.NewResponse(&genposv1.DeleteCustomerGroupResponse{}), nil
}

// ----- CSV import ----------------------------------------------------------

func (h *CustomerHandler) ParseImportCustomerCsv(
	ctx context.Context,
	req *connect.Request[genposv1.ParseImportCustomerCsvRequest],
) (*connect.Response[genposv1.ParseImportCustomerCsvResponse], error) {
	authCtx, err := h.requireCustomerAuth(ctx)
	if err != nil {
		return nil, err
	}
	res, err := h.usecase.ParseImportCustomerCsv(ctx, input.ParseImportCustomerCsvInput{
		OrgID:   authCtx.OrgID,
		CsvData: req.Msg.GetCsvData(),
	})
	if err != nil {
		return nil, h.logAndConvertCustomer(ctx, "parse customer csv", err)
	}
	rows := make([]*genposv1.CsvCustomerRow, 0, len(res.Rows))
	for _, r := range res.Rows {
		rows = append(rows, toCustomerCsvRowProto(r))
	}
	return connect.NewResponse(&genposv1.ParseImportCustomerCsvResponse{
		Rows:       rows,
		ValidCount: res.ValidCount,
		ErrorCount: res.ErrorCount,
		Warnings:   res.Warnings,
	}), nil
}

func (h *CustomerHandler) ImportCustomers(
	ctx context.Context,
	req *connect.Request[genposv1.ImportCustomersRequest],
) (*connect.Response[genposv1.ImportCustomersResponse], error) {
	authCtx, err := h.requireCustomerAuth(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]input.ImportCustomerItem, 0, len(req.Msg.GetItems()))
	for _, it := range req.Msg.GetItems() {
		items = append(items, input.ImportCustomerItem{
			Row:              fromCustomerCsvRowProto(it.GetRow()),
			OverrideExisting: it.GetOverrideExisting(),
			ExistingID:       it.GetExistingId(),
		})
	}
	res, err := h.usecase.ImportCustomers(ctx, input.ImportCustomersInput{
		OrgID: authCtx.OrgID,
		Items: items,
	})
	if err != nil {
		return nil, h.logAndConvertCustomer(ctx, "import customers", err)
	}
	return connect.NewResponse(&genposv1.ImportCustomersResponse{
		Created: res.Created,
		Updated: res.Updated,
		Skipped: res.Skipped,
		Errors:  res.Errors,
	}), nil
}

// ----- helpers -------------------------------------------------------------

func (h *CustomerHandler) requireCustomerAuth(ctx context.Context) (*interceptor.AuthContext, error) {
	a := interceptor.FromContext(ctx)
	if a == nil {
		return nil, errors.ToConnectError(errors.Unauthorized("not signed in"))
	}
	return a, nil
}

func (h *CustomerHandler) logAndConvertCustomer(ctx context.Context, op string, err error) error {
	if errors.ShouldLog(errors.GetCode(err)) {
		h.logger.ErrorContext(ctx, op+" failed", "error", err)
	}
	return errors.ToConnectError(err)
}

func toCustomerProto(c *entity.Customer) *genposv1.Customer {
	if c == nil {
		return nil
	}
	dob := ""
	if !c.DateOfBirth.IsZero() {
		dob = c.DateOfBirth.Format("2006-01-02")
	}
	return &genposv1.Customer{
		Id:          c.ID,
		OrgId:       c.OrgID,
		Name:        c.Name,
		Email:       c.Email,
		Phone:       c.Phone,
		Notes:       c.Notes,
		GroupIds:    c.GroupIDs,
		CreatedAt:   timestamppb.New(c.CreatedAt),
		UpdatedAt:   timestamppb.New(c.UpdatedAt),
		Code:        c.Code,
		Address:     c.Address,
		Company:     c.Company,
		TaxCode:     c.TaxCode,
		DateOfBirth: dob,
		Gender:      c.Gender,
		Facebook:    c.Facebook,
		IsActive:    c.IsActive,
	}
}

func fromCustomerInputProto(c *genposv1.CustomerInput) input.CustomerInput {
	if c == nil {
		return input.CustomerInput{}
	}
	var dob time.Time
	if s := c.GetDateOfBirth(); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			dob = t
		}
	}
	return input.CustomerInput{
		Name:        c.GetName(),
		Email:       c.GetEmail(),
		Phone:       c.GetPhone(),
		Notes:       c.GetNotes(),
		GroupIDs:    c.GetGroupIds(),
		Code:        c.GetCode(),
		Address:     c.GetAddress(),
		Company:     c.GetCompany(),
		TaxCode:     c.GetTaxCode(),
		DateOfBirth: dob,
		Gender:      c.GetGender(),
		Facebook:    c.GetFacebook(),
		IsActive:    c.GetIsActive(),
	}
}

func toCustomerGroupProto(g *entity.CustomerGroup) *genposv1.CustomerGroup {
	if g == nil {
		return nil
	}
	return &genposv1.CustomerGroup{
		Id:            g.ID,
		OrgId:         g.OrgID,
		Name:          g.Name,
		Description:   g.Description,
		DiscountType:  g.DiscountType,
		DiscountValue: g.DiscountValue,
		CreatedAt:     timestamppb.New(g.CreatedAt),
		UpdatedAt:     timestamppb.New(g.UpdatedAt),
	}
}

func fromCustomerGroupInputProto(g *genposv1.CustomerGroupInput) input.CustomerGroupInput {
	if g == nil {
		return input.CustomerGroupInput{}
	}
	return input.CustomerGroupInput{
		Name:          g.GetName(),
		Description:   g.GetDescription(),
		DiscountType:  g.GetDiscountType(),
		DiscountValue: g.GetDiscountValue(),
	}
}

func toCustomerCsvRowProto(r input.CsvCustomerRow) *genposv1.CsvCustomerRow {
	return &genposv1.CsvCustomerRow{
		Name:        r.Name,
		Email:       r.Email,
		Phone:       r.Phone,
		Notes:       r.Notes,
		Groups:      r.Groups,
		Errors:      r.Errors,
		Exists:      r.Exists,
		ExistingId:  r.ExistingID,
		Code:        r.Code,
		Address:     r.Address,
		Company:     r.Company,
		TaxCode:     r.TaxCode,
		DateOfBirth: r.DateOfBirth,
		Gender:      r.Gender,
		Facebook:    r.Facebook,
		Status:      r.Status,
	}
}

func fromCustomerCsvRowProto(r *genposv1.CsvCustomerRow) input.CsvCustomerRow {
	if r == nil {
		return input.CsvCustomerRow{}
	}
	return input.CsvCustomerRow{
		Name:        r.GetName(),
		Email:       r.GetEmail(),
		Phone:       r.GetPhone(),
		Notes:       r.GetNotes(),
		Groups:      r.GetGroups(),
		Errors:      r.GetErrors(),
		Exists:      r.GetExists(),
		ExistingID:  r.GetExistingId(),
		Code:        r.GetCode(),
		Address:     r.GetAddress(),
		Company:     r.GetCompany(),
		TaxCode:     r.GetTaxCode(),
		DateOfBirth: r.GetDateOfBirth(),
		Gender:      r.GetGender(),
		Facebook:    r.GetFacebook(),
		Status:      r.GetStatus(),
	}
}

var _ genposv1connect.CustomerServiceHandler = (*CustomerHandler)(nil)

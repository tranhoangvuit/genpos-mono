package usecase

import (
	"context"
	"strings"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type supplierUsecase struct {
	tenantDB       gateway.TenantDB
	supplierReader gateway.SupplierReader
	supplierWriter gateway.SupplierWriter
}

// NewSupplierUsecase constructs a SupplierUsecase.
func NewSupplierUsecase(
	tenantDB gateway.TenantDB,
	supplierReader gateway.SupplierReader,
	supplierWriter gateway.SupplierWriter,
) SupplierUsecase {
	return &supplierUsecase{
		tenantDB:       tenantDB,
		supplierReader: supplierReader,
		supplierWriter: supplierWriter,
	}
}

func (u *supplierUsecase) ListSuppliers(ctx context.Context, orgID string) ([]*entity.Supplier, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.Supplier
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.supplierReader.List(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list suppliers")
	}
	return out, nil
}

func (u *supplierUsecase) CreateSupplier(ctx context.Context, in input.CreateSupplierInput) (*entity.Supplier, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if strings.TrimSpace(in.Supplier.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	var out *entity.Supplier
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		s, err := u.supplierWriter.Create(ctx, gateway.CreateSupplierParams{
			OrgID:       in.OrgID,
			Name:        strings.TrimSpace(in.Supplier.Name),
			ContactName: strings.TrimSpace(in.Supplier.ContactName),
			Email:       strings.TrimSpace(in.Supplier.Email),
			Phone:       strings.TrimSpace(in.Supplier.Phone),
			Address:     in.Supplier.Address,
			Notes:       in.Supplier.Notes,
		})
		if err != nil {
			return err
		}
		out = s
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "create supplier")
	}
	return out, nil
}

func (u *supplierUsecase) UpdateSupplier(ctx context.Context, in input.UpdateSupplierInput) (*entity.Supplier, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if strings.TrimSpace(in.Supplier.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	var out *entity.Supplier
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		s, err := u.supplierWriter.Update(ctx, gateway.UpdateSupplierParams{
			ID:          in.ID,
			Name:        strings.TrimSpace(in.Supplier.Name),
			ContactName: strings.TrimSpace(in.Supplier.ContactName),
			Email:       strings.TrimSpace(in.Supplier.Email),
			Phone:       strings.TrimSpace(in.Supplier.Phone),
			Address:     in.Supplier.Address,
			Notes:       in.Supplier.Notes,
		})
		if err != nil {
			return err
		}
		out = s
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "update supplier")
	}
	return out, nil
}

func (u *supplierUsecase) DeleteSupplier(ctx context.Context, in input.DeleteSupplierInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		return u.supplierWriter.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete supplier")
	}
	return nil
}

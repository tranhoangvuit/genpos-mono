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

type supplierStore struct{}

// NewSupplierReader returns a SupplierReader backed by sqlc.
func NewSupplierReader() gateway.SupplierReader { return &supplierStore{} }

// NewSupplierWriter returns a SupplierWriter backed by sqlc.
func NewSupplierWriter() gateway.SupplierWriter { return &supplierStore{} }

func (s *supplierStore) GetByID(ctx context.Context, id string) (*entity.Supplier, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid supplier id")
	}
	r, err := sqlc.New(dbtx).GetSupplierByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("supplier not found")
		}
		return nil, errors.Wrap(err, "get supplier by id")
	}
	return toSupplierEntity(r.ID, r.OrgID, r.Name, r.ContactName, r.Email, r.Phone, r.Address, r.Notes, r.CreatedAt, r.UpdatedAt), nil
}

func (s *supplierStore) List(ctx context.Context) ([]*entity.Supplier, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListSuppliers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list suppliers")
	}
	out := make([]*entity.Supplier, 0, len(rows))
	for _, r := range rows {
		out = append(out, toSupplierEntity(r.ID, r.OrgID, r.Name, r.ContactName, r.Email, r.Phone, r.Address, r.Notes, r.CreatedAt, r.UpdatedAt))
	}
	return out, nil
}

func (s *supplierStore) Create(ctx context.Context, params gateway.CreateSupplierParams) (*entity.Supplier, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	r, err := sqlc.New(dbtx).CreateSupplier(ctx, sqlc.CreateSupplierParams{
		OrgID:       orgID,
		Name:        params.Name,
		ContactName: textOrNull(params.ContactName),
		Email:       textOrNull(params.Email),
		Phone:       textOrNull(params.Phone),
		Address:     textOrNull(params.Address),
		Notes:       textOrNull(params.Notes),
	})
	if err != nil {
		return nil, errors.Wrap(err, "create supplier")
	}
	return toSupplierEntity(r.ID, r.OrgID, r.Name, r.ContactName, r.Email, r.Phone, r.Address, r.Notes, r.CreatedAt, r.UpdatedAt), nil
}

func (s *supplierStore) Update(ctx context.Context, params gateway.UpdateSupplierParams) (*entity.Supplier, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseUUID(params.ID)
	if err != nil {
		return nil, errors.BadRequest("invalid supplier id")
	}
	r, err := sqlc.New(dbtx).UpdateSupplier(ctx, sqlc.UpdateSupplierParams{
		ID:          id,
		Name:        params.Name,
		ContactName: textOrNull(params.ContactName),
		Email:       textOrNull(params.Email),
		Phone:       textOrNull(params.Phone),
		Address:     textOrNull(params.Address),
		Notes:       textOrNull(params.Notes),
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("supplier not found")
		}
		return nil, errors.Wrap(err, "update supplier")
	}
	return toSupplierEntity(r.ID, r.OrgID, r.Name, r.ContactName, r.Email, r.Phone, r.Address, r.Notes, r.CreatedAt, r.UpdatedAt), nil
}

func (s *supplierStore) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid supplier id")
	}
	if err := sqlc.New(dbtx).SoftDeleteSupplier(ctx, uid); err != nil {
		return errors.Wrap(err, "soft delete supplier")
	}
	return nil
}

func toSupplierEntity(id, orgID pgtype.UUID, name string,
	contactName, email, phone, address, notes pgtype.Text,
	createdAt, updatedAt pgtype.Timestamptz) *entity.Supplier {
	return &entity.Supplier{
		ID:          uuidString(id),
		OrgID:       uuidString(orgID),
		Name:        name,
		ContactName: textString(contactName),
		Email:       textString(email),
		Phone:       textString(phone),
		Address:     textString(address),
		Notes:       textString(notes),
		CreatedAt:   createdAt.Time,
		UpdatedAt:   updatedAt.Time,
	}
}

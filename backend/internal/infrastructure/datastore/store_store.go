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

type storeStore struct{}

// NewStoreReader returns a StoreReader backed by sqlc.
func NewStoreReader() gateway.StoreReader { return &storeStore{} }

// NewStoreWriter returns a StoreWriter backed by sqlc.
func NewStoreWriter() gateway.StoreWriter { return &storeStore{} }

func (s *storeStore) GetFirstForOrg(ctx context.Context, orgID string) (*entity.Store, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(orgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	r, err := sqlc.New(dbtx).GetFirstStoreForOrg(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("store not found")
		}
		return nil, errors.Wrap(err, "get first store")
	}
	return toStoreEntity(r.ID, r.OrgID, r.Name, r.Address, r.Phone, r.Email, r.Timezone, r.Status, r.CreatedAt, r.UpdatedAt), nil
}

func (s *storeStore) List(ctx context.Context) ([]*entity.Store, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListStores(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list stores")
	}
	out := make([]*entity.Store, 0, len(rows))
	for _, r := range rows {
		out = append(out, toStoreEntity(r.ID, r.OrgID, r.Name, r.Address, r.Phone, r.Email, r.Timezone, r.Status, r.CreatedAt, r.UpdatedAt))
	}
	return out, nil
}

func (s *storeStore) Create(ctx context.Context, params gateway.CreateStoreParams) (*entity.Store, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	status := params.Status
	if status == "" {
		status = "active"
	}
	r, err := sqlc.New(dbtx).CreateStore(ctx, sqlc.CreateStoreParams{
		OrgID:    orgID,
		Name:     params.Name,
		Address:  textOrNull(params.Address),
		Phone:    textOrNull(params.Phone),
		Email:    textOrNull(params.Email),
		Timezone: textOrNull(params.Timezone),
		Status:   status,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create store")
	}
	return toStoreEntity(r.ID, r.OrgID, r.Name, r.Address, r.Phone, r.Email, r.Timezone, r.Status, r.CreatedAt, r.UpdatedAt), nil
}

func (s *storeStore) Update(ctx context.Context, params gateway.UpdateStoreParams) (*entity.Store, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseUUID(params.ID)
	if err != nil {
		return nil, errors.BadRequest("invalid store id")
	}
	r, err := sqlc.New(dbtx).UpdateStore(ctx, sqlc.UpdateStoreParams{
		ID:       id,
		Name:     params.Name,
		Address:  textOrNull(params.Address),
		Phone:    textOrNull(params.Phone),
		Email:    textOrNull(params.Email),
		Timezone: textOrNull(params.Timezone),
		Status:   params.Status,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("store not found")
		}
		return nil, errors.Wrap(err, "update store")
	}
	return toStoreEntity(r.ID, r.OrgID, r.Name, r.Address, r.Phone, r.Email, r.Timezone, r.Status, r.CreatedAt, r.UpdatedAt), nil
}

func (s *storeStore) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid store id")
	}
	n, err := sqlc.New(dbtx).SoftDeleteStore(ctx, uid)
	if err != nil {
		return errors.Wrap(err, "soft delete store")
	}
	if n == 0 {
		return errors.NotFound("store not found")
	}
	return nil
}

func toStoreEntity(id, orgID pgtype.UUID, name string, address, phone, email, timezone pgtype.Text,
	status string, createdAt, updatedAt pgtype.Timestamptz) *entity.Store {
	return &entity.Store{
		ID:        uuidString(id),
		OrgID:     uuidString(orgID),
		Name:      name,
		Address:   textString(address),
		Phone:     textString(phone),
		Email:     textString(email),
		Timezone:  textString(timezone),
		Status:    status,
		CreatedAt: createdAt.Time,
		UpdatedAt: updatedAt.Time,
	}
}

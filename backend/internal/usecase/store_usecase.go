package usecase

import (
	"context"
	"strings"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

var validStoreStatuses = map[string]struct{}{
	"active":   {},
	"inactive": {},
	"closed":   {},
}

type storeUsecase struct {
	tenantDB    gateway.TenantDB
	storeReader gateway.StoreReader
	storeWriter gateway.StoreWriter
}

// NewStoreUsecase constructs a StoreUsecase.
func NewStoreUsecase(
	tenantDB gateway.TenantDB,
	storeReader gateway.StoreReader,
	storeWriter gateway.StoreWriter,
) StoreUsecase {
	return &storeUsecase{
		tenantDB:    tenantDB,
		storeReader: storeReader,
		storeWriter: storeWriter,
	}
}

func (u *storeUsecase) ListStoreDetails(ctx context.Context, orgID string) ([]*entity.Store, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.Store
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.storeReader.List(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list stores")
	}
	return out, nil
}

func (u *storeUsecase) CreateStore(ctx context.Context, in input.CreateStoreInput) (*entity.Store, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if strings.TrimSpace(in.Store.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	status := in.Store.Status
	if status == "" {
		status = "active"
	}
	if _, ok := validStoreStatuses[status]; !ok {
		return nil, errors.BadRequest("status must be active, inactive, or closed")
	}
	var out *entity.Store
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		s, err := u.storeWriter.Create(ctx, gateway.CreateStoreParams{
			OrgID:    in.OrgID,
			Name:     strings.TrimSpace(in.Store.Name),
			Address:  strings.TrimSpace(in.Store.Address),
			Phone:    strings.TrimSpace(in.Store.Phone),
			Email:    strings.TrimSpace(in.Store.Email),
			Timezone: strings.TrimSpace(in.Store.Timezone),
			Status:   status,
		})
		if err != nil {
			return err
		}
		out = s
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "create store")
	}
	return out, nil
}

func (u *storeUsecase) UpdateStore(ctx context.Context, in input.UpdateStoreInput) (*entity.Store, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if strings.TrimSpace(in.Store.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	status := in.Store.Status
	if status == "" {
		status = "active"
	}
	if _, ok := validStoreStatuses[status]; !ok {
		return nil, errors.BadRequest("status must be active, inactive, or closed")
	}
	var out *entity.Store
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		s, err := u.storeWriter.Update(ctx, gateway.UpdateStoreParams{
			ID:       in.ID,
			Name:     strings.TrimSpace(in.Store.Name),
			Address:  strings.TrimSpace(in.Store.Address),
			Phone:    strings.TrimSpace(in.Store.Phone),
			Email:    strings.TrimSpace(in.Store.Email),
			Timezone: strings.TrimSpace(in.Store.Timezone),
			Status:   status,
		})
		if err != nil {
			return err
		}
		out = s
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "update store")
	}
	return out, nil
}

func (u *storeUsecase) DeleteStore(ctx context.Context, in input.DeleteStoreInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		return u.storeWriter.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete store")
	}
	return nil
}

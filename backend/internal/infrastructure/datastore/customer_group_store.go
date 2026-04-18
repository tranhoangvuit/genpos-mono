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

type customerGroupStore struct{}

// NewCustomerGroupReader returns a CustomerGroupReader backed by sqlc.
func NewCustomerGroupReader() gateway.CustomerGroupReader { return &customerGroupStore{} }

// NewCustomerGroupWriter returns a CustomerGroupWriter backed by sqlc.
func NewCustomerGroupWriter() gateway.CustomerGroupWriter { return &customerGroupStore{} }

func (s *customerGroupStore) List(ctx context.Context) ([]*entity.CustomerGroup, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListCustomerGroups(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list customer groups")
	}
	out := make([]*entity.CustomerGroup, 0, len(rows))
	for _, r := range rows {
		out = append(out, toCustomerGroupEntity(r.ID, r.OrgID, r.Name, r.Description, r.DiscountType, r.DiscountValue, r.CreatedAt, r.UpdatedAt))
	}
	return out, nil
}

func (s *customerGroupStore) GetByID(ctx context.Context, id string) (*entity.CustomerGroup, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid customer group id")
	}
	r, err := sqlc.New(dbtx).GetCustomerGroupByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("customer group not found")
		}
		return nil, errors.Wrap(err, "get customer group by id")
	}
	return toCustomerGroupEntity(r.ID, r.OrgID, r.Name, r.Description, r.DiscountType, r.DiscountValue, r.CreatedAt, r.UpdatedAt), nil
}

func (s *customerGroupStore) GetByName(ctx context.Context, name string) (*entity.CustomerGroup, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	r, err := sqlc.New(dbtx).GetCustomerGroupByName(ctx, name)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("customer group not found")
		}
		return nil, errors.Wrap(err, "get customer group by name")
	}
	return toCustomerGroupEntity(r.ID, r.OrgID, r.Name, r.Description, r.DiscountType, r.DiscountValue, r.CreatedAt, r.UpdatedAt), nil
}

func (s *customerGroupStore) Create(ctx context.Context, params gateway.CreateCustomerGroupParams) (*entity.CustomerGroup, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	discountVal, err := numericOrNull(params.DiscountValue)
	if err != nil {
		return nil, errors.BadRequest("invalid discount value")
	}
	r, err := sqlc.New(dbtx).CreateCustomerGroup(ctx, sqlc.CreateCustomerGroupParams{
		OrgID:         orgID,
		Name:          params.Name,
		Description:   textOrNull(params.Description),
		DiscountType:  textOrNull(params.DiscountType),
		DiscountValue: discountVal,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create customer group")
	}
	return toCustomerGroupEntity(r.ID, r.OrgID, r.Name, r.Description, r.DiscountType, r.DiscountValue, r.CreatedAt, r.UpdatedAt), nil
}

func (s *customerGroupStore) Update(ctx context.Context, params gateway.UpdateCustomerGroupParams) (*entity.CustomerGroup, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseUUID(params.ID)
	if err != nil {
		return nil, errors.BadRequest("invalid customer group id")
	}
	discountVal, err := numericOrNull(params.DiscountValue)
	if err != nil {
		return nil, errors.BadRequest("invalid discount value")
	}
	r, err := sqlc.New(dbtx).UpdateCustomerGroup(ctx, sqlc.UpdateCustomerGroupParams{
		ID:            id,
		Name:          params.Name,
		Description:   textOrNull(params.Description),
		DiscountType:  textOrNull(params.DiscountType),
		DiscountValue: discountVal,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("customer group not found")
		}
		return nil, errors.Wrap(err, "update customer group")
	}
	return toCustomerGroupEntity(r.ID, r.OrgID, r.Name, r.Description, r.DiscountType, r.DiscountValue, r.CreatedAt, r.UpdatedAt), nil
}

func (s *customerGroupStore) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid customer group id")
	}
	if err := sqlc.New(dbtx).SoftDeleteCustomerGroup(ctx, uid); err != nil {
		return errors.Wrap(err, "soft delete customer group")
	}
	return nil
}

func toCustomerGroupEntity(id, orgID pgtype.UUID, name string,
	description, discountType pgtype.Text, discountValue pgtype.Numeric,
	createdAt, updatedAt pgtype.Timestamptz) *entity.CustomerGroup {
	return &entity.CustomerGroup{
		ID:            uuidString(id),
		OrgID:         uuidString(orgID),
		Name:          name,
		Description:   textString(description),
		DiscountType:  textString(discountType),
		DiscountValue: numericToString(discountValue),
		CreatedAt:     createdAt.Time,
		UpdatedAt:     updatedAt.Time,
	}
}

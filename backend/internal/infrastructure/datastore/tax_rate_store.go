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

type taxRateStore struct{}

func NewTaxRateReader() gateway.TaxRateReader { return &taxRateStore{} }
func NewTaxRateWriter() gateway.TaxRateWriter { return &taxRateStore{} }

func (s *taxRateStore) List(ctx context.Context) ([]*entity.TaxRate, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListTaxRates(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list tax rates")
	}
	out := make([]*entity.TaxRate, 0, len(rows))
	for _, r := range rows {
		out = append(out, toTaxRateEntity(r.ID, r.OrgID, r.Name, r.Rate, r.IsInclusive, r.IsDefault, r.CreatedAt, r.UpdatedAt))
	}
	return out, nil
}

func (s *taxRateStore) Create(ctx context.Context, params gateway.CreateTaxRateParams) (*entity.TaxRate, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	rate, err := numericFromString(params.Rate)
	if err != nil {
		return nil, errors.BadRequest("invalid rate")
	}
	r, err := sqlc.New(dbtx).CreateTaxRate(ctx, sqlc.CreateTaxRateParams{
		OrgID:       orgID,
		Name:        params.Name,
		Rate:        rate,
		IsInclusive: params.IsInclusive,
		IsDefault:   params.IsDefault,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create tax rate")
	}
	return toTaxRateEntity(r.ID, r.OrgID, r.Name, r.Rate, r.IsInclusive, r.IsDefault, r.CreatedAt, r.UpdatedAt), nil
}

func (s *taxRateStore) Update(ctx context.Context, params gateway.UpdateTaxRateParams) (*entity.TaxRate, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseUUID(params.ID)
	if err != nil {
		return nil, errors.BadRequest("invalid tax rate id")
	}
	rate, err := numericFromString(params.Rate)
	if err != nil {
		return nil, errors.BadRequest("invalid rate")
	}
	r, err := sqlc.New(dbtx).UpdateTaxRate(ctx, sqlc.UpdateTaxRateParams{
		ID:          id,
		Name:        params.Name,
		Rate:        rate,
		IsInclusive: params.IsInclusive,
		IsDefault:   params.IsDefault,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("tax rate not found")
		}
		return nil, errors.Wrap(err, "update tax rate")
	}
	return toTaxRateEntity(r.ID, r.OrgID, r.Name, r.Rate, r.IsInclusive, r.IsDefault, r.CreatedAt, r.UpdatedAt), nil
}

func (s *taxRateStore) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid tax rate id")
	}
	n, err := sqlc.New(dbtx).SoftDeleteTaxRate(ctx, uid)
	if err != nil {
		return errors.Wrap(err, "soft delete tax rate")
	}
	if n == 0 {
		return errors.NotFound("tax rate not found")
	}
	return nil
}

func (s *taxRateStore) ClearDefaults(ctx context.Context) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	if err := sqlc.New(dbtx).ClearDefaultTaxRates(ctx); err != nil {
		return errors.Wrap(err, "clear default tax rates")
	}
	return nil
}

func toTaxRateEntity(id, orgID pgtype.UUID, name string, rate pgtype.Numeric,
	isInclusive, isDefault bool, createdAt, updatedAt pgtype.Timestamptz) *entity.TaxRate {
	return &entity.TaxRate{
		ID:          uuidString(id),
		OrgID:       uuidString(orgID),
		Name:        name,
		Rate:        numericToString(rate),
		IsInclusive: isInclusive,
		IsDefault:   isDefault,
		CreatedAt:   createdAt.Time,
		UpdatedAt:   updatedAt.Time,
	}
}

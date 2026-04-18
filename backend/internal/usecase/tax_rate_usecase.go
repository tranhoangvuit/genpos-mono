package usecase

import (
	"context"
	"strconv"
	"strings"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type taxRateUsecase struct {
	tenantDB gateway.TenantDB
	reader   gateway.TaxRateReader
	writer   gateway.TaxRateWriter
}

// NewTaxRateUsecase constructs a TaxRateUsecase.
func NewTaxRateUsecase(
	tenantDB gateway.TenantDB,
	reader gateway.TaxRateReader,
	writer gateway.TaxRateWriter,
) TaxRateUsecase {
	return &taxRateUsecase{tenantDB: tenantDB, reader: reader, writer: writer}
}

func (u *taxRateUsecase) ListTaxRates(ctx context.Context, orgID string) ([]*entity.TaxRate, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.TaxRate
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.reader.List(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list tax rates")
	}
	return out, nil
}

func (u *taxRateUsecase) CreateTaxRate(ctx context.Context, in input.CreateTaxRateInput) (*entity.TaxRate, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if err := validateTaxRate(in.Rate); err != nil {
		return nil, err
	}
	var out *entity.TaxRate
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		if in.Rate.IsDefault {
			if err := u.writer.ClearDefaults(ctx); err != nil {
				return err
			}
		}
		r, err := u.writer.Create(ctx, gateway.CreateTaxRateParams{
			OrgID:       in.OrgID,
			Name:        strings.TrimSpace(in.Rate.Name),
			Rate:        strings.TrimSpace(in.Rate.Rate),
			IsInclusive: in.Rate.IsInclusive,
			IsDefault:   in.Rate.IsDefault,
		})
		if err != nil {
			return err
		}
		out = r
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "create tax rate")
	}
	return out, nil
}

func (u *taxRateUsecase) UpdateTaxRate(ctx context.Context, in input.UpdateTaxRateInput) (*entity.TaxRate, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if err := validateTaxRate(in.Rate); err != nil {
		return nil, err
	}
	var out *entity.TaxRate
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		if in.Rate.IsDefault {
			if err := u.writer.ClearDefaults(ctx); err != nil {
				return err
			}
		}
		r, err := u.writer.Update(ctx, gateway.UpdateTaxRateParams{
			ID:          in.ID,
			Name:        strings.TrimSpace(in.Rate.Name),
			Rate:        strings.TrimSpace(in.Rate.Rate),
			IsInclusive: in.Rate.IsInclusive,
			IsDefault:   in.Rate.IsDefault,
		})
		if err != nil {
			return err
		}
		out = r
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "update tax rate")
	}
	return out, nil
}

func (u *taxRateUsecase) DeleteTaxRate(ctx context.Context, in input.DeleteTaxRateInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		return u.writer.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete tax rate")
	}
	return nil
}

func validateTaxRate(r input.TaxRateInput) error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.BadRequest("name is required")
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(r.Rate), 64)
	if err != nil {
		return errors.BadRequest("rate must be a number")
	}
	if v < 0 || v > 1 {
		return errors.BadRequest("rate must be between 0 and 1 (e.g. 0.1 for 10%)")
	}
	return nil
}

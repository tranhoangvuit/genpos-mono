package usecase

import (
	"context"
	"strings"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type taxClassUsecase struct {
	tenantDB gateway.TenantDB
	reader   gateway.TaxClassReader
	writer   gateway.TaxClassWriter
}

// NewTaxClassUsecase constructs a TaxClassUsecase.
func NewTaxClassUsecase(
	tenantDB gateway.TenantDB,
	reader gateway.TaxClassReader,
	writer gateway.TaxClassWriter,
) TaxClassUsecase {
	return &taxClassUsecase{tenantDB: tenantDB, reader: reader, writer: writer}
}

func (u *taxClassUsecase) ListTaxClasses(ctx context.Context, orgID string) ([]*entity.TaxClass, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.TaxClass
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.reader.List(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list tax classes")
	}
	return out, nil
}

func (u *taxClassUsecase) GetTaxClass(ctx context.Context, in input.GetTaxClassInput) (*entity.TaxClass, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	var out *entity.TaxClass
	if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		c, err := u.reader.Get(ctx, in.ID)
		if err != nil {
			return err
		}
		out = c
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "get tax class")
	}
	return out, nil
}

func (u *taxClassUsecase) CreateTaxClass(ctx context.Context, in input.CreateTaxClassInput) (*entity.TaxClass, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if err := validateTaxClass(in.Class); err != nil {
		return nil, err
	}
	var out *entity.TaxClass
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		// The DB enforces a unique partial index on (org_id, is_default=TRUE).
		// Mirror the tax_rate flow: clear the existing default before setting
		// a new one so the user can switch defaults without a separate API.
		if in.Class.IsDefault {
			if err := u.writer.ClearDefaults(ctx); err != nil {
				return err
			}
		}
		c, err := u.writer.Create(ctx, gateway.CreateTaxClassParams{
			OrgID:       in.OrgID,
			Name:        strings.TrimSpace(in.Class.Name),
			Description: strings.TrimSpace(in.Class.Description),
			IsDefault:   in.Class.IsDefault,
			SortOrder:   in.Class.SortOrder,
			Rates:       toGatewayRateParams(in.Class.Rates),
		})
		if err != nil {
			return err
		}
		out = c
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "create tax class")
	}
	return out, nil
}

func (u *taxClassUsecase) UpdateTaxClass(ctx context.Context, in input.UpdateTaxClassInput) (*entity.TaxClass, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if err := validateTaxClass(in.Class); err != nil {
		return nil, err
	}
	var out *entity.TaxClass
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		if in.Class.IsDefault {
			if err := u.writer.ClearDefaults(ctx); err != nil {
				return err
			}
		}
		c, err := u.writer.Update(ctx, gateway.UpdateTaxClassParams{
			ID:          in.ID,
			OrgID:       in.OrgID,
			Name:        strings.TrimSpace(in.Class.Name),
			Description: strings.TrimSpace(in.Class.Description),
			IsDefault:   in.Class.IsDefault,
			SortOrder:   in.Class.SortOrder,
			Rates:       toGatewayRateParams(in.Class.Rates),
		})
		if err != nil {
			return err
		}
		out = c
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "update tax class")
	}
	return out, nil
}

func (u *taxClassUsecase) DeleteTaxClass(ctx context.Context, in input.DeleteTaxClassInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		return u.writer.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete tax class")
	}
	return nil
}

func validateTaxClass(c input.TaxClassInput) error {
	if strings.TrimSpace(c.Name) == "" {
		return errors.BadRequest("name is required")
	}
	// Reject duplicate (tax_rate_id) entries inside one class -- the DB
	// enforces this via a unique partial index on
	// (tax_class_id, tax_rate_id) WHERE deleted_at IS NULL, but catching it
	// here gives a friendlier error than a constraint violation.
	seen := make(map[string]struct{}, len(c.Rates))
	for _, r := range c.Rates {
		if strings.TrimSpace(r.TaxRateID) == "" {
			return errors.BadRequest("each rate requires a tax_rate_id")
		}
		if _, dup := seen[r.TaxRateID]; dup {
			return errors.BadRequest("each tax rate may appear at most once in a class")
		}
		seen[r.TaxRateID] = struct{}{}
	}
	return nil
}

func toGatewayRateParams(in []input.TaxClassRateInput) []gateway.TaxClassRateParams {
	out := make([]gateway.TaxClassRateParams, len(in))
	for i, r := range in {
		out[i] = gateway.TaxClassRateParams{
			TaxRateID:  r.TaxRateID,
			Sequence:   r.Sequence,
			IsCompound: r.IsCompound,
		}
	}
	return out
}

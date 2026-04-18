package usecase

import (
	"context"
	"strings"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

var validPaymentMethodTypes = map[string]struct{}{
	"cash":          {},
	"card":          {},
	"mobile":        {},
	"bank_transfer": {},
	"voucher":       {},
	"other":         {},
}

type paymentMethodUsecase struct {
	tenantDB gateway.TenantDB
	reader   gateway.PaymentMethodReader
	writer   gateway.PaymentMethodWriter
}

// NewPaymentMethodUsecase constructs a PaymentMethodUsecase.
func NewPaymentMethodUsecase(
	tenantDB gateway.TenantDB,
	reader gateway.PaymentMethodReader,
	writer gateway.PaymentMethodWriter,
) PaymentMethodUsecase {
	return &paymentMethodUsecase{tenantDB: tenantDB, reader: reader, writer: writer}
}

func (u *paymentMethodUsecase) ListPaymentMethods(ctx context.Context, orgID string) ([]*entity.PaymentMethod, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.PaymentMethod
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.reader.List(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list payment methods")
	}
	return out, nil
}

func (u *paymentMethodUsecase) CreatePaymentMethod(ctx context.Context, in input.CreatePaymentMethodInput) (*entity.PaymentMethod, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if err := validatePaymentMethod(in.Method); err != nil {
		return nil, err
	}
	var out *entity.PaymentMethod
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		m, err := u.writer.Create(ctx, gateway.CreatePaymentMethodParams{
			OrgID:     in.OrgID,
			Name:      strings.TrimSpace(in.Method.Name),
			Type:      in.Method.Type,
			IsActive:  in.Method.IsActive,
			SortOrder: in.Method.SortOrder,
		})
		if err != nil {
			return err
		}
		out = m
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "create payment method")
	}
	return out, nil
}

func (u *paymentMethodUsecase) UpdatePaymentMethod(ctx context.Context, in input.UpdatePaymentMethodInput) (*entity.PaymentMethod, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if err := validatePaymentMethod(in.Method); err != nil {
		return nil, err
	}
	var out *entity.PaymentMethod
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		m, err := u.writer.Update(ctx, gateway.UpdatePaymentMethodParams{
			ID:        in.ID,
			Name:      strings.TrimSpace(in.Method.Name),
			Type:      in.Method.Type,
			IsActive:  in.Method.IsActive,
			SortOrder: in.Method.SortOrder,
		})
		if err != nil {
			return err
		}
		out = m
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "update payment method")
	}
	return out, nil
}

func (u *paymentMethodUsecase) DeletePaymentMethod(ctx context.Context, in input.DeletePaymentMethodInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		return u.writer.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete payment method")
	}
	return nil
}

func validatePaymentMethod(m input.PaymentMethodInput) error {
	if strings.TrimSpace(m.Name) == "" {
		return errors.BadRequest("name is required")
	}
	if _, ok := validPaymentMethodTypes[m.Type]; !ok {
		return errors.BadRequest("type must be cash, card, mobile, bank_transfer, voucher, or other")
	}
	return nil
}

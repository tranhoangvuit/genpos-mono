package datastore

import (
	"context"
	stderrors "errors"

	"github.com/jackc/pgx/v5"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type paymentMethodStore struct{}

func NewPaymentMethodReader() gateway.PaymentMethodReader { return &paymentMethodStore{} }
func NewPaymentMethodWriter() gateway.PaymentMethodWriter { return &paymentMethodStore{} }

func (s *paymentMethodStore) List(ctx context.Context) ([]*entity.PaymentMethod, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListPaymentMethods(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list payment methods")
	}
	out := make([]*entity.PaymentMethod, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.PaymentMethod{
			ID:        uuidString(r.ID),
			OrgID:     uuidString(r.OrgID),
			Name:      r.Name,
			Type:      r.Type,
			IsActive:  r.IsActive,
			SortOrder: r.SortOrder,
			CreatedAt: r.CreatedAt.Time,
			UpdatedAt: r.UpdatedAt.Time,
		})
	}
	return out, nil
}

func (s *paymentMethodStore) Create(ctx context.Context, params gateway.CreatePaymentMethodParams) (*entity.PaymentMethod, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	r, err := sqlc.New(dbtx).CreatePaymentMethod(ctx, sqlc.CreatePaymentMethodParams{
		OrgID:     orgID,
		Name:      params.Name,
		Type:      params.Type,
		IsActive:  params.IsActive,
		SortOrder: params.SortOrder,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create payment method")
	}
	return &entity.PaymentMethod{
		ID:        uuidString(r.ID),
		OrgID:     uuidString(r.OrgID),
		Name:      r.Name,
		Type:      r.Type,
		IsActive:  r.IsActive,
		SortOrder: r.SortOrder,
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
	}, nil
}

func (s *paymentMethodStore) Update(ctx context.Context, params gateway.UpdatePaymentMethodParams) (*entity.PaymentMethod, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseUUID(params.ID)
	if err != nil {
		return nil, errors.BadRequest("invalid payment method id")
	}
	r, err := sqlc.New(dbtx).UpdatePaymentMethod(ctx, sqlc.UpdatePaymentMethodParams{
		ID:        id,
		Name:      params.Name,
		Type:      params.Type,
		IsActive:  params.IsActive,
		SortOrder: params.SortOrder,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("payment method not found")
		}
		return nil, errors.Wrap(err, "update payment method")
	}
	return &entity.PaymentMethod{
		ID:        uuidString(r.ID),
		OrgID:     uuidString(r.OrgID),
		Name:      r.Name,
		Type:      r.Type,
		IsActive:  r.IsActive,
		SortOrder: r.SortOrder,
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
	}, nil
}

func (s *paymentMethodStore) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid payment method id")
	}
	n, err := sqlc.New(dbtx).SoftDeletePaymentMethod(ctx, uid)
	if err != nil {
		return errors.Wrap(err, "soft delete payment method")
	}
	if n == 0 {
		return errors.NotFound("payment method not found")
	}
	return nil
}

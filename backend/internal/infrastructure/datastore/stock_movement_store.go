package datastore

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type stockMovementStore struct{}

// NewStockMovementWriter returns a StockMovementWriter backed by sqlc.
func NewStockMovementWriter() gateway.StockMovementWriter { return &stockMovementStore{} }

func (s *stockMovementStore) Insert(ctx context.Context, params gateway.CreateStockMovementParams) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return errors.BadRequest("invalid org id")
	}
	storeID, err := parseUUID(params.StoreID)
	if err != nil {
		return errors.BadRequest("invalid store id")
	}
	variantID, err := parseUUID(params.VariantID)
	if err != nil {
		return errors.BadRequest("invalid variant id")
	}
	refID, err := uuidOrNull(params.ReferenceID)
	if err != nil {
		return errors.BadRequest("invalid reference id")
	}
	userID, err := uuidOrNull(params.UserID)
	if err != nil {
		return errors.BadRequest("invalid user id")
	}
	qty, err := numericFromString(params.Quantity)
	if err != nil {
		return errors.BadRequest("invalid quantity")
	}
	if _, err := sqlc.New(dbtx).InsertStockMovement(ctx, sqlc.InsertStockMovementParams{
		OrgID:         orgID,
		StoreID:       storeID,
		RegisterID:    pgtype.UUID{Valid: false},
		VariantID:     variantID,
		Direction:     params.Direction,
		Quantity:      qty,
		MovementType:  params.MovementType,
		ReferenceType: textOrNull(params.ReferenceType),
		ReferenceID:   refID,
		UserID:        userID,
		Notes:         textOrNull(params.Notes),
	}); err != nil {
		return errors.Wrap(err, "insert stock movement")
	}
	return nil
}

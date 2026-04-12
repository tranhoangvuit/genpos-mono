package datastore

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

const defaultLimit int32 = 100

type dbtxKey struct{}

// WithDBTX attaches a database connection to the context.
func WithDBTX(ctx context.Context, dbtx sqlc.DBTX) context.Context {
	return context.WithValue(ctx, dbtxKey{}, dbtx)
}

// GetDBTX returns the database connection from context.
func GetDBTX(ctx context.Context) (sqlc.DBTX, error) {
	if dbtx, ok := ctx.Value(dbtxKey{}).(sqlc.DBTX); ok {
		return dbtx, nil
	}
	return nil, errors.Internal("database connection not found in context")
}

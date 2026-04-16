package datastore

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txManager struct {
	pool *pgxpool.Pool
}

// NewTxManager returns a TxManager backed by the pgx pool.
func NewTxManager(pool *pgxpool.Pool) gateway.TxManager {
	return &txManager{pool: pool}
}

func (m *txManager) Do(ctx context.Context, fn func(context.Context) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	txCtx := WithDBTX(ctx, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "commit tx")
	}
	return nil
}

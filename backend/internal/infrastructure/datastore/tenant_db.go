package datastore

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tenantDB struct {
	pool *pgxpool.Pool
}

// NewTenantDB creates a new TenantDB instance.
func NewTenantDB(pool *pgxpool.Pool) gateway.TenantDB {
	return &tenantDB{pool: pool}
}

// WithTenant executes fn within a tenant-scoped database transaction.
func (t *tenantDB) WithTenant(ctx context.Context, clientID string, fn func(ctx context.Context) error) error {
	conn, err := t.pool.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "acquire connection")
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "begin transaction")
	}

	// Transaction-local GUC — scoped to this tx only, resets on commit/rollback.
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org_id', $1, true)", clientID); err != nil {
		_ = tx.Rollback(ctx)
		return errors.Wrap(err, "set tenant context")
	}

	txCtx := WithDBTX(ctx, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "commit transaction")
	}

	return nil
}

// ReadWithTenant executes fn inside a READ ONLY transaction with the tenant
// GUC set transaction-locally. Using a tx (not a bare pooled connection)
// guarantees the GUC cannot leak to a later request that picks up the same
// pool connection — commit/rollback resets it.
func (t *tenantDB) ReadWithTenant(ctx context.Context, clientID string, fn func(ctx context.Context) error) error {
	conn, err := t.pool.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "acquire connection")
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return errors.Wrap(err, "begin read-only transaction")
	}

	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_org_id', $1, true)", clientID); err != nil {
		_ = tx.Rollback(ctx)
		return errors.Wrap(err, "set tenant context")
	}

	txCtx := WithDBTX(ctx, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "commit read-only transaction")
	}

	return nil
}

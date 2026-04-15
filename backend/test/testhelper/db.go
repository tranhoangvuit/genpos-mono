package testhelper

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupTestDB creates a pgxpool.Pool connected to the genpos_test database.
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	host := envOrDefault("TEST_DB_HOST", "localhost")
	port := envOrDefault("TEST_DB_PORT", "3033")
	user := envOrDefault("TEST_DB_USER", "postgres")
	pass := envOrDefault("TEST_DB_PASSWORD", "postgres")
	dbName := envOrDefault("TEST_DB_NAME", "genpos_test")

	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		host, port, dbName, user, pass)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("failed to create test pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

// ApplyMigrations runs the migration SQL and creates a non-superuser role
// so that RLS policies are enforced during tests.
func ApplyMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()

	// Create table schema
	migration := `
CREATE TABLE IF NOT EXISTS products (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    sku         TEXT        NOT NULL,
    price_cents BIGINT      NOT NULL DEFAULT 0,
    active      BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_products_org_id ON products (org_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_org_sku ON products (org_id, sku);

ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE products FORCE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies WHERE policyname = 'products_tenant_isolation'
    ) THEN
        CREATE POLICY products_tenant_isolation ON products
            USING (org_id = current_setting('app.current_org_id', true));
    END IF;
END
$$;
`
	if _, err := pool.Exec(ctx, migration); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	// Create a non-superuser role for RLS enforcement.
	// Superusers bypass RLS, so we need a regular role.
	roleSetup := `
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'genpos_app') THEN
        CREATE ROLE genpos_app LOGIN PASSWORD 'genpos_app';
    END IF;
END
$$;
GRANT ALL ON ALL TABLES IN SCHEMA public TO genpos_app;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO genpos_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO genpos_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO genpos_app;
`
	if _, err := pool.Exec(ctx, roleSetup); err != nil {
		t.Fatalf("failed to create test role: %v", err)
	}
}

// SetupTestDBWithRLS creates a pool connected as the non-superuser genpos_app
// so that RLS policies are enforced.
func SetupTestDBWithRLS(t *testing.T) *pgxpool.Pool {
	t.Helper()

	host := envOrDefault("TEST_DB_HOST", "localhost")
	port := envOrDefault("TEST_DB_PORT", "3033")
	dbName := envOrDefault("TEST_DB_NAME", "genpos_test")

	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=genpos_app password=genpos_app sslmode=disable",
		host, port, dbName)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("failed to create RLS test pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("failed to ping test database as genpos_app: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

// CleanTable truncates the given table.
func CleanTable(t *testing.T, pool *pgxpool.Pool, table string) {
	t.Helper()
	quoted := (&pgx.Identifier{table}).Sanitize()
	_, err := pool.Exec(context.Background(), "TRUNCATE "+quoted+" CASCADE")
	if err != nil {
		t.Fatalf("failed to clean table %s: %v", table, err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

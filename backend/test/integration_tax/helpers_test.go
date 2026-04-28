//go:build integration

package integration_tax_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var setupRLSRoleOnce sync.Once

// ensureRLSRole bootstraps the non-superuser `genpos_app` role on the test
// DB the first time it's invoked. Superusers bypass RLS, so a regular role
// is required for the tenant_db.go's set_config('app.current_org_id') trick
// to actually filter rows. The role survives across test runs once seeded.
func ensureRLSRole(t *testing.T, adminPool *pgxpool.Pool) {
	t.Helper()
	setupRLSRoleOnce.Do(func() {
		ctx := context.Background()
		stmts := []string{
			`DO $$
			BEGIN
				IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'genpos_app') THEN
					CREATE ROLE genpos_app LOGIN PASSWORD 'genpos_app';
				END IF;
			END $$`,
			`GRANT ALL ON ALL TABLES IN SCHEMA public TO genpos_app`,
			`GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO genpos_app`,
			`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO genpos_app`,
			`ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO genpos_app`,
		}
		for _, s := range stmts {
			if _, err := adminPool.Exec(ctx, s); err != nil {
				t.Fatalf("bootstrap genpos_app role: %v", err)
			}
		}
	})
}

// rlsPool returns a pool that connects as `genpos_app` so RLS policies are
// enforced. Mirrors testhelper.SetupTestDBWithRLS but lives here to avoid
// dragging in the broken ApplyMigrations dependency.
func rlsPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	host := envOrDefault("TEST_DB_HOST", "localhost")
	port := envOrDefault("TEST_DB_PORT", "3033")
	dbName := envOrDefault("TEST_DB_NAME", "genpos_test")
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=genpos_app password=genpos_app sslmode=disable",
		host, port, dbName,
	)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("create rls pool: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("ping rls pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// insertTaxRate seeds one tax_rates row via the admin pool and returns its
// UUID as a string. RLS is bypassed at insert time (postgres superuser);
// reads from the usecase under test still flow through the genpos_app role.
func insertTaxRate(t *testing.T, pool *pgxpool.Pool, orgID, name, rate string, inclusive bool) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(), `
		INSERT INTO tax_rates (org_id, name, rate, is_inclusive)
		VALUES ($1, $2, $3::NUMERIC, $4)
		RETURNING id::TEXT`, orgID, name, rate, inclusive).Scan(&id)
	if err != nil {
		t.Fatalf("insert tax_rate %s: %v", name, err)
	}
	return id
}

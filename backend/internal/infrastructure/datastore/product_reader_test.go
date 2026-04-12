package datastore_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore"
	"github.com/genpick/genpos-mono/backend/test/testhelper"
)

func Test_ProductReader_ListProducts(t *testing.T) {
	t.Parallel()

	// Use superuser to set up schema and seed data
	adminPool := testhelper.SetupTestDB(t)
	testhelper.ApplyMigrations(t, adminPool)
	testhelper.CleanTable(t, adminPool, "products")
	seedProducts(t, adminPool)

	// Use non-superuser pool so RLS is enforced
	rlsPool := testhelper.SetupTestDBWithRLS(t)

	reader := datastore.NewProductReader()

	cases := map[string]struct {
		orgID  string
		params gateway.ListProductsParams
		want   []*entity.Product
	}{
		"returns products for org": {
			orgID:  "test-org-1",
			params: gateway.ListProductsParams{Limit: 10, Offset: 0},
			want: []*entity.Product{
				{OrgID: "test-org-1", Name: "Widget B", SKU: "WGT-B", PriceCents: 2099, Active: true},
				{OrgID: "test-org-1", Name: "Widget A", SKU: "WGT-A", PriceCents: 1099, Active: true},
			},
		},
		"returns empty list for unknown org": {
			orgID:  "org-nonexistent",
			params: gateway.ListProductsParams{Limit: 10, Offset: 0},
			want:   []*entity.Product{},
		},
		"respects limit": {
			orgID:  "test-org-1",
			params: gateway.ListProductsParams{Limit: 1, Offset: 0},
			want: []*entity.Product{
				{OrgID: "test-org-1", Name: "Widget B", SKU: "WGT-B", PriceCents: 2099, Active: true},
			},
		},
		"respects offset": {
			orgID:  "test-org-1",
			params: gateway.ListProductsParams{Limit: 10, Offset: 1},
			want: []*entity.Product{
				{OrgID: "test-org-1", Name: "Widget A", SKU: "WGT-A", PriceCents: 1099, Active: true},
			},
		},
		"isolates tenants": {
			orgID:  "test-org-2",
			params: gateway.ListProductsParams{Limit: 10, Offset: 0},
			want: []*entity.Product{
				{OrgID: "test-org-2", Name: "Gadget X", SKU: "GDG-X", PriceCents: 999, Active: true},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tenantDB := datastore.NewTenantDB(rlsPool)
			var got []*entity.Product

			err := tenantDB.ReadWithTenant(context.Background(), tc.orgID, func(ctx context.Context) error {
				var err error
				got, err = reader.ListProducts(ctx, tc.params)
				return err
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			opts := []cmp.Option{
				cmpopts.IgnoreFields(entity.Product{}, "ID", "CreatedAt", "UpdatedAt"),
				cmpopts.EquateApproxTime(time.Second),
			}
			if diff := cmp.Diff(tc.want, got, opts...); diff != "" {
				t.Errorf("result mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func seedProducts(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	products := []struct {
		orgID      string
		name       string
		sku        string
		priceCents int64
		createdAt  string
	}{
		{"test-org-1", "Widget A", "WGT-A", 1099, "2024-01-01T00:00:00Z"},
		{"test-org-1", "Widget B", "WGT-B", 2099, "2024-01-02T00:00:00Z"},
		{"test-org-2", "Gadget X", "GDG-X", 999, "2024-01-01T00:00:00Z"},
	}

	for _, p := range products {
		_, err := pool.Exec(context.Background(),
			`INSERT INTO products (org_id, name, sku, price_cents, created_at) VALUES ($1, $2, $3, $4, $5)`,
			p.orgID, p.name, p.sku, p.priceCents, p.createdAt,
		)
		if err != nil {
			t.Fatalf("failed to seed product %s: %v", p.name, err)
		}
	}
}

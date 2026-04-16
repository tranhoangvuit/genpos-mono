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

	adminPool := testhelper.SetupTestDB(t)
	testhelper.ApplyMigrations(t, adminPool)
	testhelper.CleanTable(t, adminPool, "products")
	seedProducts(t, adminPool)

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
				{OrgID: "test-org-1", Name: "Widget A", IsActive: true, SortOrder: 0},
				{OrgID: "test-org-1", Name: "Widget B", IsActive: true, SortOrder: 1},
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
				{OrgID: "test-org-1", Name: "Widget A", IsActive: true, SortOrder: 0},
			},
		},
		"respects offset": {
			orgID:  "test-org-1",
			params: gateway.ListProductsParams{Limit: 10, Offset: 1},
			want: []*entity.Product{
				{OrgID: "test-org-1", Name: "Widget B", IsActive: true, SortOrder: 1},
			},
		},
		"isolates tenants": {
			orgID:  "test-org-2",
			params: gateway.ListProductsParams{Limit: 10, Offset: 0},
			want: []*entity.Product{
				{OrgID: "test-org-2", Name: "Gadget X", IsActive: true, SortOrder: 0},
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
				cmpopts.IgnoreFields(entity.Product{}, "ID", "CategoryID", "Description", "ImageURL", "CreatedAt", "UpdatedAt"),
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
		orgID     string
		name      string
		sortOrder int
		createdAt string
	}{
		{"test-org-1", "Widget A", 0, "2024-01-01T00:00:00Z"},
		{"test-org-1", "Widget B", 1, "2024-01-02T00:00:00Z"},
		{"test-org-2", "Gadget X", 0, "2024-01-01T00:00:00Z"},
	}

	for _, p := range products {
		_, err := pool.Exec(context.Background(),
			`INSERT INTO products (org_id, name, sort_order, created_at) VALUES ($1, $2, $3, $4)`,
			p.orgID, p.name, p.sortOrder, p.createdAt,
		)
		if err != nil {
			t.Fatalf("failed to seed product %s: %v", p.name, err)
		}
	}
}

//go:build integration

package integration_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/testing/protocmp"

	genposv1 "github.com/genpick/genpos-mono/backend/gen/genpos/v1"
	"github.com/genpick/genpos-mono/backend/gen/genpos/v1/genposv1connect"
	"github.com/genpick/genpos-mono/backend/internal/handler"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/test/testhelper"
)

func Test_Integration_ListProducts(t *testing.T) {
	t.Parallel()

	// Set up schema and seed data as superuser
	adminPool := testhelper.SetupTestDB(t)
	testhelper.ApplyMigrations(t, adminPool)
	testhelper.CleanTable(t, adminPool, "products")
	seedIntegrationProducts(t, adminPool)

	// Use non-superuser pool so RLS is enforced
	rlsPool := testhelper.SetupTestDBWithRLS(t)

	// Wire up the full stack
	tenantDB := datastore.NewTenantDB(rlsPool)
	productReader := datastore.NewProductReader()
	productUsecase := usecase.NewProductUsecase(tenantDB, productReader)
	logger := slog.Default()
	server := handler.NewServer(logger, productUsecase)

	mux := http.NewServeMux()
	path, h := genposv1connect.NewGenposServiceHandler(server)
	mux.Handle(path, h)

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	client := genposv1connect.NewGenposServiceClient(ts.Client(), ts.URL)

	cases := map[string]struct {
		req     *genposv1.ListProductsRequest
		want    *genposv1.ListProductsResponse
		wantErr connect.Code
	}{
		"returns products for org": {
			req: &genposv1.ListProductsRequest{OrgId: "integ-org-1", PageSize: 10},
			want: &genposv1.ListProductsResponse{
				Products: []*genposv1.Product{
					{OrgId: "integ-org-1", Name: "Item B", Sku: "ITEM-B", PriceCents: 2000, Active: true},
					{OrgId: "integ-org-1", Name: "Item A", Sku: "ITEM-A", PriceCents: 1000, Active: true},
				},
			},
		},
		"returns empty for unknown org": {
			req: &genposv1.ListProductsRequest{OrgId: "integ-org-none", PageSize: 10},
			want: &genposv1.ListProductsResponse{
				Products: []*genposv1.Product{},
			},
		},
		"error on missing org_id": {
			req:     &genposv1.ListProductsRequest{OrgId: ""},
			wantErr: connect.CodeInvalidArgument,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp, err := client.ListProducts(context.Background(), connect.NewRequest(tc.req))
			if tc.wantErr != 0 {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if connect.CodeOf(err) != tc.wantErr {
					t.Errorf("error code: want %v, got %v", tc.wantErr, connect.CodeOf(err))
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			opts := []cmp.Option{
				protocmp.Transform(),
				protocmp.IgnoreFields(&genposv1.Product{}, "id", "created_at", "updated_at"),
			}
			if diff := cmp.Diff(tc.want, resp.Msg, opts...); diff != "" {
				t.Errorf("response mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func seedIntegrationProducts(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	products := []struct {
		orgID      string
		name       string
		sku        string
		priceCents int64
		createdAt  string
	}{
		{"integ-org-1", "Item A", "ITEM-A", 1000, "2024-01-01T00:00:00Z"},
		{"integ-org-1", "Item B", "ITEM-B", 2000, "2024-01-02T00:00:00Z"},
		{"integ-org-2", "Other", "OTH-1", 500, "2024-01-01T00:00:00Z"},
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

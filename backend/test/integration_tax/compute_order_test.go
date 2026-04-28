//go:build integration

package integration_tax_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/test/testhelper"
)

const computeOrgA = "01900000-0000-0000-0000-00000000e001"

// Test_Integration_ComputeOrder proves ComputeOrder resolves real
// variant->tax_class->tax_rates rows from the DB and returns a breakdown
// matching the resolver's spec. RLS is on (genpos_app role).
func Test_Integration_ComputeOrder(t *testing.T) {
	t.Parallel()

	adminPool := testhelper.SetupTestDB(t)
	ensureRLSRole(t, adminPool)
	seedComputeOrg(t, adminPool)
	t.Cleanup(func() { cleanupComputeData(t, adminPool) })

	tenantDB := datastore.NewTenantDB(rlsPool(t))

	// Build the supporting graph through the real usecases so the tax-class
	// FK chain matches what production writes.
	taxUC := usecase.NewTaxClassUsecase(tenantDB,
		datastore.NewTaxClassReader(), datastore.NewTaxClassWriter())
	catalogUC := usecase.NewCatalogUsecase(tenantDB,
		datastore.NewCategoryReader(), datastore.NewCategoryWriter(),
		datastore.NewProductDetailReader(), datastore.NewProductWriter())
	orderUC := usecase.NewOrderUsecase(tenantDB,
		datastore.NewOrderReader(), datastore.NewOrderWriter(),
		datastore.NewOrgStoreReader(), datastore.NewMemberReader(),
		datastore.NewVariantTaxResolver())

	rateID := insertTaxRate(t, adminPool, computeOrgA, "VAT 10% / compute", "0.1000", false)
	taxClass, err := taxUC.CreateTaxClass(context.Background(), input.CreateTaxClassInput{
		OrgID: computeOrgA,
		Class: input.TaxClassInput{
			Name: "Standard / compute",
			Rates: []input.TaxClassRateInput{
				{TaxRateID: rateID, Sequence: 0},
			},
		},
	})
	if err != nil {
		t.Fatalf("seed tax class: %v", err)
	}

	taxedProduct, err := catalogUC.CreateProduct(context.Background(), input.CreateProductInput{
		OrgID: computeOrgA,
		Product: input.ProductInput{
			Name: "Coffee / compute",
			Variants: []input.VariantInput{{
				Name:       "Default",
				Price:      "100",
				IsActive:   true,
				TaxClassID: taxClass.ID,
			}},
		},
	})
	if err != nil {
		t.Fatalf("seed taxed product: %v", err)
	}
	taxedVariantID := taxedProduct.Variants[0].ID

	untaxedProduct, err := catalogUC.CreateProduct(context.Background(), input.CreateProductInput{
		OrgID: computeOrgA,
		Product: input.ProductInput{
			Name: "Bagel / compute (no tax)",
			Variants: []input.VariantInput{{
				Name:     "Default",
				Price:    "50",
				IsActive: true,
			}},
		},
	})
	if err != nil {
		t.Fatalf("seed untaxed product: %v", err)
	}
	untaxedVariantID := untaxedProduct.Variants[0].ID

	t.Run("two-line cart with one taxed and one untaxed variant", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		got, err := orderUC.ComputeOrder(ctx, input.ComputeOrderInput{
			OrgID: computeOrgA,
			Lines: []input.ComputeOrderLineInput{
				{VariantID: taxedVariantID, Quantity: "2", UnitPrice: "100"},
				{VariantID: untaxedVariantID, Quantity: "1", UnitPrice: "50"},
			},
		})
		if err != nil {
			t.Fatalf("compute: %v", err)
		}
		// Taxed line: 2 * 100 = 200 base, +20 tax. Untaxed: 50, no tax.
		// Subtotal = 200 + 50 = 250. Tax = 20. Total = 270.
		if got.Subtotal != "250.0000" {
			t.Errorf("subtotal: want 250.0000, got %s", got.Subtotal)
		}
		if got.TaxTotal != "20.0000" {
			t.Errorf("tax_total: want 20.0000, got %s", got.TaxTotal)
		}
		if got.Total != "270.0000" {
			t.Errorf("total: want 270.0000, got %s", got.Total)
		}
		if len(got.Lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(got.Lines))
		}
		if len(got.Lines[0].Taxes) != 1 {
			t.Errorf("taxed line: want 1 tax, got %d", len(got.Lines[0].Taxes))
		}
		if len(got.Lines[1].Taxes) != 0 {
			t.Errorf("untaxed line: want 0 taxes, got %d", len(got.Lines[1].Taxes))
		}
	})

	t.Run("inclusive override flips snapshot to inclusive math", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		got, err := orderUC.ComputeOrder(ctx, input.ComputeOrderInput{
			OrgID: computeOrgA,
			Lines: []input.ComputeOrderLineInput{{
				VariantID:            taxedVariantID,
				Quantity:             "1",
				UnitPrice:            "110", // gross
				HasInclusiveOverride: true,
				InclusiveOverride:    true,
			}},
		})
		if err != nil {
			t.Fatalf("compute: %v", err)
		}
		// Inclusive math: 110 contains 10% tax => base 100, tax 10. Total 110.
		if got.Total != "110.0000" {
			t.Errorf("total: want 110.0000, got %s", got.Total)
		}
		if got.TaxTotal != "10.0000" {
			t.Errorf("tax_total: want 10.0000, got %s", got.TaxTotal)
		}
	})
}

func seedComputeOrg(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	slug := "org-" + computeOrgA[len(computeOrgA)-12:]
	_, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, name, currency, timezone, status)
		VALUES ($1::UUID, $2, $2, 'VND', 'Asia/Ho_Chi_Minh', 'active')
		ON CONFLICT (id) DO NOTHING`, computeOrgA, slug)
	if err != nil {
		t.Fatalf("seed org: %v", err)
	}
}

func cleanupComputeData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	stmts := []string{
		`DELETE FROM product_variants WHERE org_id = $1`,
		`DELETE FROM products WHERE org_id = $1`,
		`DELETE FROM tax_class_rates WHERE org_id = $1`,
		`DELETE FROM tax_classes WHERE org_id = $1`,
		`DELETE FROM tax_rates WHERE org_id = $1`,
		`DELETE FROM organizations WHERE id = $1`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s, computeOrgA); err != nil {
			t.Logf("cleanup %q: %v", s, err)
		}
	}
}

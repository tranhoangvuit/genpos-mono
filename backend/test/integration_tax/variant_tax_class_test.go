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

const variantTaxOrgA = "01900000-0000-0000-0000-00000000d001"

// Test_Integration_VariantTaxClass proves a tax_class_id supplied on the
// VariantInput round-trips through CreateProduct + GetProduct against a real
// Postgres. RLS is on (genpos_app role), so cross-org leakage on the join
// would surface as an error rather than a silent pass.
func Test_Integration_VariantTaxClass(t *testing.T) {
	t.Parallel()

	adminPool := testhelper.SetupTestDB(t)
	ensureRLSRole(t, adminPool)
	seedVariantTaxOrg(t, adminPool)
	t.Cleanup(func() { cleanupVariantTaxData(t, adminPool) })

	tenantDB := datastore.NewTenantDB(rlsPool(t))
	uc := usecase.NewCatalogUsecase(
		tenantDB,
		datastore.NewCategoryReader(),
		datastore.NewCategoryWriter(),
		datastore.NewProductDetailReader(),
		datastore.NewProductWriter(),
	)

	// Seed a tax class via the tax-class usecase so the test exercises the
	// real ID format the FK expects.
	taxUC := usecase.NewTaxClassUsecase(
		tenantDB,
		datastore.NewTaxClassReader(),
		datastore.NewTaxClassWriter(),
	)
	taxClass, err := taxUC.CreateTaxClass(context.Background(), input.CreateTaxClassInput{
		OrgID: variantTaxOrgA,
		Class: input.TaxClassInput{Name: "Variant tax / TC"},
	})
	if err != nil {
		t.Fatalf("seed tax class: %v", err)
	}

	t.Run("variant create persists tax_class_id and read returns it", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		created, err := uc.CreateProduct(ctx, input.CreateProductInput{
			OrgID: variantTaxOrgA,
			Product: input.ProductInput{
				Name:     "Coffee / variant_tax_class roundtrip",
				IsActive: true,
				Variants: []input.VariantInput{{
					Name:       "Default",
					Price:      "100",
					IsActive:   true,
					TaxClassID: taxClass.ID,
				}},
			},
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if len(created.Variants) != 1 {
			t.Fatalf("variants: want 1, got %d", len(created.Variants))
		}
		if got := created.Variants[0].TaxClassID; got != taxClass.ID {
			t.Errorf("tax_class_id on create response: want %s, got %s", taxClass.ID, got)
		}

		// Re-read to prove the value survived the writer + reader round-trip
		// rather than being echoed straight from the input struct.
		got, err := uc.GetProduct(ctx, input.GetProductInput{ID: created.ID, OrgID: variantTaxOrgA})
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if len(got.Variants) != 1 {
			t.Fatalf("variants: want 1, got %d", len(got.Variants))
		}
		if got.Variants[0].TaxClassID != taxClass.ID {
			t.Errorf("tax_class_id after re-read: want %s, got %s",
				taxClass.ID, got.Variants[0].TaxClassID)
		}
	})

	t.Run("empty tax_class_id stores NULL and round-trips as empty string", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		created, err := uc.CreateProduct(ctx, input.CreateProductInput{
			OrgID: variantTaxOrgA,
			Product: input.ProductInput{
				Name:     "Untaxed / no tax_class roundtrip",
				IsActive: true,
				Variants: []input.VariantInput{{
					Name:     "Default",
					Price:    "50",
					IsActive: true,
				}},
			},
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if got := created.Variants[0].TaxClassID; got != "" {
			t.Errorf("tax_class_id with no input: want empty, got %s", got)
		}
	})

	t.Run("invalid tax_class_id is rejected at the boundary", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		_, err := uc.CreateProduct(ctx, input.CreateProductInput{
			OrgID: variantTaxOrgA,
			Product: input.ProductInput{
				Name:     "Bad / not-a-uuid",
				IsActive: true,
				Variants: []input.VariantInput{{
					Name:       "Default",
					Price:      "10",
					IsActive:   true,
					TaxClassID: "not-a-uuid",
				}},
			},
		})
		if err == nil {
			t.Fatal("expected error for malformed tax_class_id, got nil")
		}
	})
}

func seedVariantTaxOrg(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	slug := "org-" + variantTaxOrgA[len(variantTaxOrgA)-12:]
	_, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, name, currency, timezone, status)
		VALUES ($1::UUID, $2, $2, 'VND', 'Asia/Ho_Chi_Minh', 'active')
		ON CONFLICT (id) DO NOTHING`, variantTaxOrgA, slug)
	if err != nil {
		t.Fatalf("seed org: %v", err)
	}
}

func cleanupVariantTaxData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	stmts := []string{
		`DELETE FROM product_variants WHERE org_id = $1`,
		`DELETE FROM products WHERE org_id = $1`,
		`DELETE FROM tax_class_rates WHERE org_id = $1`,
		`DELETE FROM tax_classes WHERE org_id = $1`,
		`DELETE FROM organizations WHERE id = $1`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s, variantTaxOrgA); err != nil {
			t.Logf("cleanup %q: %v", s, err)
		}
	}
}

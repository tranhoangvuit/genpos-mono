//go:build integration

package integration_tax_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/test/testhelper"
)

const (
	taxClassOrgA = "01900000-0000-0000-0000-00000000b001"
	taxClassOrgB = "01900000-0000-0000-0000-00000000b002"
)

// Test_Integration_TaxClass exercises the TaxClass usecase end-to-end against
// a real Postgres backed by Atlas migrations. RLS is enforced because we
// connect as the non-superuser genpos_app role -- so cross-org leakage would
// fail visibly on read.
func Test_Integration_TaxClass(t *testing.T) {
	t.Parallel()

	adminPool := testhelper.SetupTestDB(t)
	// Atlas migrations are already applied to genpos_test by the docker
	// migrate service -- no test-time DDL needed. RLS bootstrap creates the
	// genpos_app role on first run so policies actually filter rows when the
	// usecase connects through the rls pool.
	ensureRLSRole(t, adminPool)
	seedTaxClassOrgs(t, adminPool)
	t.Cleanup(func() { cleanupTaxClassData(t, adminPool) })

	rateA := insertTaxRate(t, adminPool, taxClassOrgA, "VAT 10% / TC", "0.1000", false)
	rateB := insertTaxRate(t, adminPool, taxClassOrgA, "Service 5% / TC", "0.0500", false)
	otherOrgRate := insertTaxRate(t, adminPool, taxClassOrgB, "Other / TC", "0.0800", false)

	tenantDB := datastore.NewTenantDB(rlsPool(t))
	uc := usecase.NewTaxClassUsecase(
		tenantDB,
		datastore.NewTaxClassReader(),
		datastore.NewTaxClassWriter(),
	)

	t.Run("create then get round-trips rates in sequence order", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		created, err := uc.CreateTaxClass(ctx, input.CreateTaxClassInput{
			OrgID: taxClassOrgA,
			Class: input.TaxClassInput{
				Name:        "Standard / round-trip",
				Description: "VAT 10% then service 5% compound",
				SortOrder:   1,
				Rates: []input.TaxClassRateInput{
					{TaxRateID: rateA, Sequence: 0, IsCompound: false},
					{TaxRateID: rateB, Sequence: 1, IsCompound: true},
				},
			},
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		got, err := uc.GetTaxClass(ctx, input.GetTaxClassInput{OrgID: taxClassOrgA, ID: created.ID})
		if err != nil {
			t.Fatalf("get: %v", err)
		}

		want := &entity.TaxClass{
			OrgID:       taxClassOrgA,
			Name:        "Standard / round-trip",
			Description: "VAT 10% then service 5% compound",
			SortOrder:   1,
			Rates: []*entity.TaxClassRate{
				{TaxRateID: rateA, Sequence: 0, IsCompound: false},
				{TaxRateID: rateB, Sequence: 1, IsCompound: true},
			},
		}
		opts := []cmp.Option{
			cmpopts.IgnoreFields(entity.TaxClass{}, "ID", "CreatedAt", "UpdatedAt"),
			cmpopts.IgnoreFields(entity.TaxClassRate{}, "ID"),
		}
		if diff := cmp.Diff(want, got, opts...); diff != "" {
			t.Errorf("class mismatch (-want +got):\n%s", diff)
		}
	})

	// otherOrgRate exists; cross-org FK enforcement is a DB-level concern
	// covered by RLS isolation tests (Test_Integration_TaxSchema). FK
	// constraints in Postgres bypass RLS, so this layer cannot block it.
	_ = otherOrgRate

	t.Run("rejects empty name", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		_, err := uc.CreateTaxClass(ctx, input.CreateTaxClassInput{
			OrgID: taxClassOrgA,
			Class: input.TaxClassInput{Name: "  "},
		})
		if err == nil {
			t.Fatal("expected error for empty name, got nil")
		}
	})

	t.Run("rejects duplicate rate inside a class", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		_, err := uc.CreateTaxClass(ctx, input.CreateTaxClassInput{
			OrgID: taxClassOrgA,
			Class: input.TaxClassInput{
				Name: "Dup / TC",
				Rates: []input.TaxClassRateInput{
					{TaxRateID: rateA, Sequence: 0},
					{TaxRateID: rateA, Sequence: 1},
				},
			},
		})
		if err == nil {
			t.Fatal("expected duplicate rate error, got nil")
		}
	})

	t.Run("setting a new default clears the previous one", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		first, err := uc.CreateTaxClass(ctx, input.CreateTaxClassInput{
			OrgID: taxClassOrgA,
			Class: input.TaxClassInput{Name: "First default / TC", IsDefault: true},
		})
		if err != nil {
			t.Fatalf("create first default: %v", err)
		}
		second, err := uc.CreateTaxClass(ctx, input.CreateTaxClassInput{
			OrgID: taxClassOrgA,
			Class: input.TaxClassInput{Name: "Second default / TC", IsDefault: true},
		})
		if err != nil {
			t.Fatalf("create second default: %v", err)
		}

		// Re-read both to confirm the switch.
		gotFirst, err := uc.GetTaxClass(ctx, input.GetTaxClassInput{OrgID: taxClassOrgA, ID: first.ID})
		if err != nil {
			t.Fatalf("get first: %v", err)
		}
		gotSecond, err := uc.GetTaxClass(ctx, input.GetTaxClassInput{OrgID: taxClassOrgA, ID: second.ID})
		if err != nil {
			t.Fatalf("get second: %v", err)
		}
		if gotFirst.IsDefault {
			t.Error("first class should no longer be default")
		}
		if !gotSecond.IsDefault {
			t.Error("second class should be the new default")
		}
	})

	t.Run("list returns only own org and excludes soft-deleted", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		// Seed one class in the other org via the same usecase so RLS is the
		// only thing keeping them apart at read time.
		if _, err := uc.CreateTaxClass(ctx, input.CreateTaxClassInput{
			OrgID: taxClassOrgB,
			Class: input.TaxClassInput{Name: "Visible to B only / TC list"},
		}); err != nil {
			t.Fatalf("seed B: %v", err)
		}
		keeperA, err := uc.CreateTaxClass(ctx, input.CreateTaxClassInput{
			OrgID: taxClassOrgA,
			Class: input.TaxClassInput{Name: "Keeper / TC list"},
		})
		if err != nil {
			t.Fatalf("seed keeper: %v", err)
		}
		deletedA, err := uc.CreateTaxClass(ctx, input.CreateTaxClassInput{
			OrgID: taxClassOrgA,
			Class: input.TaxClassInput{Name: "Deleted / TC list"},
		})
		if err != nil {
			t.Fatalf("seed deleted: %v", err)
		}
		if err := uc.DeleteTaxClass(ctx, input.DeleteTaxClassInput{OrgID: taxClassOrgA, ID: deletedA.ID}); err != nil {
			t.Fatalf("soft delete: %v", err)
		}

		listA, err := uc.ListTaxClasses(ctx, taxClassOrgA)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		var names []string
		for _, c := range listA {
			if c.OrgID != taxClassOrgA {
				t.Errorf("RLS leak: list returned class for org %s", c.OrgID)
			}
			names = append(names, c.Name)
		}
		// keeperA must show, deletedA must not, OrgB's must not.
		foundKeeper := false
		for _, n := range names {
			if n == keeperA.Name {
				foundKeeper = true
			}
			if n == "Deleted / TC list" {
				t.Error("soft-deleted class should be hidden from list")
			}
			if n == "Visible to B only / TC list" {
				t.Error("RLS leak: OrgB's class returned in OrgA list")
			}
		}
		if !foundKeeper {
			t.Errorf("keeper not in list: %v", names)
		}
	})
}

func seedTaxClassOrgs(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	for _, orgID := range []string{taxClassOrgA, taxClassOrgB} {
		// The seed orgs share a leading prefix; use the trailing chunk so each
		// org gets a unique slug.
		slug := "org-" + orgID[len(orgID)-12:]
		_, err := pool.Exec(ctx, `
			INSERT INTO organizations (id, slug, name, currency, timezone, status)
			VALUES ($1::UUID, $2, $2, 'VND', 'Asia/Ho_Chi_Minh', 'active')
			ON CONFLICT (id) DO NOTHING`, orgID, slug)
		if err != nil {
			t.Fatalf("seed org %s: %v", orgID, err)
		}
	}
}

func cleanupTaxClassData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	stmts := []string{
		`DELETE FROM tax_class_rates WHERE org_id IN ($1, $2)`,
		`DELETE FROM tax_classes WHERE org_id IN ($1, $2)`,
		`DELETE FROM tax_rates WHERE org_id IN ($1, $2)`,
		`DELETE FROM organizations WHERE id IN ($1, $2)`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s, taxClassOrgA, taxClassOrgB); err != nil {
			t.Logf("cleanup %q: %v", s, err)
		}
	}
}

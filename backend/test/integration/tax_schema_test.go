//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/genpick/genpos-mono/backend/test/testhelper"
)

const (
	taxSchemaOrgA = "01900000-0000-0000-0000-00000000a001"
	taxSchemaOrgB = "01900000-0000-0000-0000-00000000a002"
)

// Test_Integration_TaxSchema exercises the migrations
// 20260428000001_create_tax_classes.sql and
// 20260428000002_create_order_adjustments.sql directly via SQL: FK chain,
// CHECK constraints, soft-delete uniqueness, RLS isolation. PR 1 has no
// usecase or handler code yet, so this test does not go through Connect.
func Test_Integration_TaxSchema(t *testing.T) {
	t.Parallel()

	adminPool := testhelper.SetupTestDB(t)
	// SetupTestDBWithRLS connects as the genpos_app role created by
	// ApplyMigrations. Bootstrap it here so the RLS subtest does not depend
	// on a sibling test having run first.
	testhelper.ApplyMigrations(t, adminPool)
	seedTwoOrgs(t, adminPool)
	t.Cleanup(func() { cleanupTaxSchema(t, adminPool) })

	cases := map[string]struct {
		run func(t *testing.T)
	}{
		"tax_class_rates FK chain accepts a valid graph": {
			run: func(t *testing.T) {
				ctx := context.Background()
				rateID := insertTaxRate(t, adminPool, taxSchemaOrgA, "VAT 10% / "+t.Name(), "0.1000", false)
				classID := insertTaxClass(t, adminPool, taxSchemaOrgA, "Standard / "+t.Name(), false)
				linkID := insertTaxClassRate(t, adminPool, taxSchemaOrgA, classID, rateID, 0, false)

				var got struct {
					ClassID  string
					RateID   string
					Sequence int
					Compound bool
				}
				err := adminPool.QueryRow(ctx,
					`SELECT tax_class_id::TEXT, tax_rate_id::TEXT, sequence, is_compound
					   FROM tax_class_rates WHERE id = $1`, linkID).
					Scan(&got.ClassID, &got.RateID, &got.Sequence, &got.Compound)
				if err != nil {
					t.Fatalf("read tax_class_rate: %v", err)
				}
				want := struct {
					ClassID  string
					RateID   string
					Sequence int
					Compound bool
				}{ClassID: classID, RateID: rateID, Sequence: 0, Compound: false}
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("link mismatch (-want +got):\n%s", diff)
				}
			},
		},
		"deleting a tax_class cascades to tax_class_rates": {
			run: func(t *testing.T) {
				ctx := context.Background()
				rateID := insertTaxRate(t, adminPool, taxSchemaOrgA, "VAT 8% / "+t.Name(), "0.0800", false)
				classID := insertTaxClass(t, adminPool, taxSchemaOrgA, "Holiday / "+t.Name(), false)
				linkID := insertTaxClassRate(t, adminPool, taxSchemaOrgA, classID, rateID, 0, false)

				if _, err := adminPool.Exec(ctx, `DELETE FROM tax_classes WHERE id = $1`, classID); err != nil {
					t.Fatalf("delete tax_class: %v", err)
				}

				var remaining int
				if err := adminPool.QueryRow(ctx,
					`SELECT COUNT(*) FROM tax_class_rates WHERE id = $1`, linkID).Scan(&remaining); err != nil {
					t.Fatalf("count remaining: %v", err)
				}
				if remaining != 0 {
					t.Errorf("expected cascade delete, got %d remaining rows", remaining)
				}
			},
		},
		"product_variants.tax_class_id references tax_classes": {
			run: func(t *testing.T) {
				ctx := context.Background()
				classID := insertTaxClass(t, adminPool, taxSchemaOrgA, "Variant Link / "+t.Name(), false)
				productID := insertProduct(t, adminPool, taxSchemaOrgA)
				variantID := insertVariantWithTaxClass(t, adminPool, taxSchemaOrgA, productID, classID)

				var gotClassID string
				if err := adminPool.QueryRow(ctx,
					`SELECT tax_class_id::TEXT FROM product_variants WHERE id = $1`, variantID).Scan(&gotClassID); err != nil {
					t.Fatalf("read variant: %v", err)
				}
				if gotClassID != classID {
					t.Errorf("variant tax_class_id: want %s, got %s", classID, gotClassID)
				}

				if _, err := adminPool.Exec(ctx,
					`UPDATE product_variants SET tax_class_id = NULL WHERE id = $1`, variantID); err != nil {
					t.Fatalf("nulling tax_class_id failed: %v", err)
				}
			},
		},
		"hard-deleting a tax_class sets variant tax_class_id to NULL": {
			run: func(t *testing.T) {
				ctx := context.Background()
				classID := insertTaxClass(t, adminPool, taxSchemaOrgA, "SetNull / "+t.Name(), false)
				productID := insertProduct(t, adminPool, taxSchemaOrgA)
				variantID := insertVariantWithTaxClass(t, adminPool, taxSchemaOrgA, productID, classID)

				if _, err := adminPool.Exec(ctx, `DELETE FROM tax_classes WHERE id = $1`, classID); err != nil {
					t.Fatalf("delete tax_class: %v", err)
				}

				var got *string
				if err := adminPool.QueryRow(ctx,
					`SELECT tax_class_id::TEXT FROM product_variants WHERE id = $1`, variantID).Scan(&got); err != nil {
					t.Fatalf("read variant: %v", err)
				}
				if got != nil {
					t.Errorf("expected tax_class_id to be NULL, got %q", *got)
				}
			},
		},
		"only one default tax_class per org": {
			run: func(t *testing.T) {
				ctx := context.Background()
				insertTaxClass(t, adminPool, taxSchemaOrgA, "First Default / "+t.Name(), true)
				_, err := adminPool.Exec(ctx,
					`INSERT INTO tax_classes (org_id, name, is_default)
					    VALUES ($1, $2, TRUE)`, taxSchemaOrgA, "Second Default / "+t.Name())
				if err == nil {
					t.Error("expected unique-default constraint violation, got nil")
				}

				if _, err := adminPool.Exec(ctx,
					`INSERT INTO tax_classes (org_id, name, is_default)
					    VALUES ($1, $2, TRUE)`, taxSchemaOrgB, "OrgB Default / "+t.Name()); err != nil {
					t.Errorf("orgB default rejected (different org should be allowed): %v", err)
				}
			},
		},
		"order_adjustments rejects tip with applies_before_tax = TRUE": {
			run: func(t *testing.T) {
				ctx := context.Background()
				orderID, _ := insertOrderAndLineItem(t, adminPool, taxSchemaOrgA)
				_, err := adminPool.Exec(ctx, `
					INSERT INTO order_adjustments
					    (org_id, order_id, kind, source_type, name_snapshot,
					     calculation_type, calculation_value, amount,
					     applies_before_tax, prorate_strategy)
					VALUES ($1, $2, 'tip', 'manual', 'Bad Tip',
					        'fixed_amount', 5000, 5000,
					        TRUE, 'no_prorate')`, taxSchemaOrgA, orderID)
				if err == nil {
					t.Error("expected chk_post_tax_kinds violation for tip with applies_before_tax=TRUE, got nil")
				}
			},
		},
		"order_adjustments rejects rounding with prorate to lines": {
			run: func(t *testing.T) {
				ctx := context.Background()
				orderID, _ := insertOrderAndLineItem(t, adminPool, taxSchemaOrgA)
				_, err := adminPool.Exec(ctx, `
					INSERT INTO order_adjustments
					    (org_id, order_id, kind, source_type, name_snapshot,
					     calculation_type, calculation_value, amount,
					     applies_before_tax, prorate_strategy)
					VALUES ($1, $2, 'rounding', 'system', 'Bad Rounding',
					        'fixed_amount', -3, -3,
					        FALSE, 'pro_rata_taxable_base')`, taxSchemaOrgA, orderID)
				if err == nil {
					t.Error("expected chk_no_prorate_kinds violation for rounding with prorate strategy, got nil")
				}
			},
		},
		"order_line_adjustments rejects line-level tip kind": {
			run: func(t *testing.T) {
				ctx := context.Background()
				_, lineItemID := insertOrderAndLineItem(t, adminPool, taxSchemaOrgA)
				_, err := adminPool.Exec(ctx, `
					INSERT INTO order_line_adjustments
					    (org_id, line_item_id, kind, source_type, name_snapshot,
					     calculation_type, calculation_value, amount)
					VALUES ($1, $2, 'tip', 'manual', 'Bad', 'percentage', 0, 0)`,
					taxSchemaOrgA, lineItemID)
				if err == nil {
					t.Error("expected kind CHECK violation, got nil")
				}
			},
		},
		"order_line_adjustments rejects unknown calculation_type": {
			run: func(t *testing.T) {
				ctx := context.Background()
				_, lineItemID := insertOrderAndLineItem(t, adminPool, taxSchemaOrgA)
				_, err := adminPool.Exec(ctx, `
					INSERT INTO order_line_adjustments
					    (org_id, line_item_id, kind, source_type, name_snapshot,
					     calculation_type, calculation_value, amount)
					VALUES ($1, $2, 'discount', 'manual', 'Bad', 'mystery', 0, 0)`,
					taxSchemaOrgA, lineItemID)
				if err == nil {
					t.Error("expected calculation_type CHECK violation, got nil")
				}
			},
		},
		"RLS isolates tax_classes between orgs": {
			run: func(t *testing.T) {
				ctx := context.Background()
				visibleA := "Visible to A only / " + t.Name()
				visibleB := "Visible to B only / " + t.Name()
				insertTaxClass(t, adminPool, taxSchemaOrgA, visibleA, false)
				insertTaxClass(t, adminPool, taxSchemaOrgB, visibleB, false)

				rlsPool := testhelper.SetupTestDBWithRLS(t)
				conn, err := rlsPool.Acquire(ctx)
				if err != nil {
					t.Fatalf("acquire rls conn: %v", err)
				}
				defer conn.Release()

				if _, err := conn.Exec(ctx,
					`SELECT set_config('app.current_org_id', $1, false)`, taxSchemaOrgA); err != nil {
					t.Fatalf("set guc: %v", err)
				}

				rows, err := conn.Query(ctx,
					`SELECT name FROM tax_classes WHERE name IN ($1, $2)`, visibleA, visibleB)
				if err != nil {
					t.Fatalf("query: %v", err)
				}
				var visible []string
				for rows.Next() {
					var name string
					if err := rows.Scan(&name); err != nil {
						t.Fatalf("scan: %v", err)
					}
					visible = append(visible, name)
				}
				rows.Close()

				want := []string{visibleA}
				if diff := cmp.Diff(want, visible, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
					t.Errorf("RLS visibility mismatch (-want +got):\n%s", diff)
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tc.run(t)
		})
	}
}

// ---------------------------------------------------------------------------
// Seeding helpers. These insert via the superuser pool (RLS bypassed) and
// return the generated UUIDs as strings for downstream FK use.
// ---------------------------------------------------------------------------

func seedTwoOrgs(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	for _, orgID := range []string{taxSchemaOrgA, taxSchemaOrgB} {
		_, err := pool.Exec(ctx, `
			INSERT INTO organizations (id, slug, name, currency, timezone, status)
			VALUES ($1, 'org-' || substr($1::TEXT, 1, 8), 'Test ' || substr($1::TEXT, 1, 8),
			        'VND', 'Asia/Ho_Chi_Minh', 'active')
			ON CONFLICT (id) DO NOTHING`, orgID)
		if err != nil {
			t.Fatalf("seed org %s: %v", orgID, err)
		}
	}
}

func cleanupTaxSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	// Children before parents. product_variants must precede tax_classes
	// because the FK is ON DELETE SET NULL, not CASCADE -- but the variants
	// themselves reference products which reference organizations, so we
	// drop variants first regardless to keep the order obvious.
	stmts := []string{
		`DELETE FROM order_adjustments WHERE org_id IN ($1, $2)`,
		`DELETE FROM order_line_adjustments WHERE org_id IN ($1, $2)`,
		`DELETE FROM order_line_taxes WHERE org_id IN ($1, $2)`,
		`DELETE FROM order_line_items WHERE org_id IN ($1, $2)`,
		`DELETE FROM orders WHERE org_id IN ($1, $2)`,
		`DELETE FROM product_variants WHERE org_id IN ($1, $2)`,
		`DELETE FROM products WHERE org_id IN ($1, $2)`,
		`DELETE FROM tax_class_rates WHERE org_id IN ($1, $2)`,
		`DELETE FROM tax_classes WHERE org_id IN ($1, $2)`,
		`DELETE FROM tax_rates WHERE org_id IN ($1, $2)`,
		`DELETE FROM stores WHERE org_id IN ($1, $2)`,
		`DELETE FROM organizations WHERE id IN ($1, $2)`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s, taxSchemaOrgA, taxSchemaOrgB); err != nil {
			t.Logf("cleanup %q: %v", s, err)
		}
	}
}

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

func insertTaxClass(t *testing.T, pool *pgxpool.Pool, orgID, name string, isDefault bool) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(), `
		INSERT INTO tax_classes (org_id, name, is_default) VALUES ($1, $2, $3) RETURNING id::TEXT`,
		orgID, name, isDefault).Scan(&id)
	if err != nil {
		t.Fatalf("insert tax_class %s: %v", name, err)
	}
	return id
}

func insertTaxClassRate(t *testing.T, pool *pgxpool.Pool, orgID, classID, rateID string, sequence int, compound bool) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(), `
		INSERT INTO tax_class_rates (org_id, tax_class_id, tax_rate_id, sequence, is_compound)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::TEXT`, orgID, classID, rateID, sequence, compound).Scan(&id)
	if err != nil {
		t.Fatalf("insert tax_class_rate: %v", err)
	}
	return id
}

func insertProduct(t *testing.T, pool *pgxpool.Pool, orgID string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(), `
		INSERT INTO products (org_id, name) VALUES ($1, 'Test Product') RETURNING id::TEXT`,
		orgID).Scan(&id)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	return id
}

func insertVariantWithTaxClass(t *testing.T, pool *pgxpool.Pool, orgID, productID, taxClassID string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(), `
		INSERT INTO product_variants (org_id, product_id, name, price, tax_class_id)
		VALUES ($1, $2, 'Default', 10000, $3)
		RETURNING id::TEXT`, orgID, productID, taxClassID).Scan(&id)
	if err != nil {
		t.Fatalf("insert variant: %v", err)
	}
	return id
}

func insertOrderAndLineItem(t *testing.T, pool *pgxpool.Pool, orgID string) (orderID, lineItemID string) {
	t.Helper()
	ctx := context.Background()

	var storeID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO stores (org_id, name, timezone, status)
		VALUES ($1, 'Test Store', 'Asia/Ho_Chi_Minh', 'active')
		RETURNING id::TEXT`, orgID).Scan(&storeID); err != nil {
		t.Fatalf("insert store: %v", err)
	}

	if err := pool.QueryRow(ctx, `
		INSERT INTO orders (org_id, store_id, order_number, status, subtotal, tax_total, discount_total, total)
		VALUES ($1, $2, 'TEST-' || substr(gen_random_uuid()::TEXT, 1, 8),
		        'completed', 100, 10, 0, 110)
		RETURNING id::TEXT`, orgID, storeID).Scan(&orderID); err != nil {
		t.Fatalf("insert order: %v", err)
	}

	if err := pool.QueryRow(ctx, `
		INSERT INTO order_line_items (org_id, order_id, product_name, quantity, unit_price, line_total)
		VALUES ($1, $2, 'Test Item', 1, 100, 100)
		RETURNING id::TEXT`, orgID, orderID).Scan(&lineItemID); err != nil {
		t.Fatalf("insert line item: %v", err)
	}
	return orderID, lineItemID
}

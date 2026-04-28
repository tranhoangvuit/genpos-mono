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
	orderTaxOrgA = "01900000-0000-0000-0000-00000000c001"
)

// Test_Integration_OrderWithTaxes drives the OrderUsecase through CreateOrder
// + GetOrder against a real Postgres. The point is to prove the new tax /
// adjustment children round-trip from input -> sqlc inserts -> RLS-enforced
// reads, and that the children-win-on-aggregates recomputation actually
// reaches the row that gets persisted.
func Test_Integration_OrderWithTaxes(t *testing.T) {
	t.Parallel()

	adminPool := testhelper.SetupTestDB(t)
	ensureRLSRole(t, adminPool)
	seedOrderTaxOrg(t, adminPool)
	t.Cleanup(func() { cleanupOrderTaxData(t, adminPool) })

	storeID := insertStoreForOrderTax(t, adminPool, orderTaxOrgA)
	roleID := insertRoleForOrderTax(t, adminPool, orderTaxOrgA)
	userID := insertUserForOrderTax(t, adminPool, orderTaxOrgA, roleID)
	insertUserStore(t, adminPool, orderTaxOrgA, userID, storeID)
	paymentMethodID := insertPaymentMethodForOrderTax(t, adminPool, orderTaxOrgA)
	taxRateID := insertTaxRate(t, adminPool, orderTaxOrgA, "VAT 10% / order", "0.1000", false)

	tenantDB := datastore.NewTenantDB(rlsPool(t))
	uc := usecase.NewOrderUsecase(
		tenantDB,
		datastore.NewOrderReader(),
		datastore.NewOrderWriter(),
		datastore.NewOrgStoreReader(),
		datastore.NewMemberReader(),
	)

	t.Run("rich children round-trip and recompute aggregates", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		created, err := uc.CreateOrder(ctx, input.CreateOrderInput{
			OrgID:       orderTaxOrgA,
			Source:      "pos",
			ExternalID:  "ext-rich-1",
			OrderNumber: "HD-INT-RICH-001",
			StoreID:     storeID,
			UserID:      userID,
			// Caller's aggregates are intentionally wrong -- recompute should
			// overwrite them.
			Subtotal:      "9999",
			TaxTotal:      "9999",
			DiscountTotal: "9999",
			Total:         "9999",
			LineItems: []input.CreateOrderLineItemInput{{
				ProductName: "Coffee",
				VariantName: "Default",
				Quantity:    "2",
				UnitPrice:   "100",
				LineTotal:   "190",
				Taxes: []input.CreateOrderLineItemTaxInput{{
					Sequence:     1,
					TaxRateID:    taxRateID,
					NameSnapshot: "VAT",
					RateSnapshot: "0.1000",
					IsInclusive:  false,
					IsCompound:   false,
					TaxableBase:  "170",
					Amount:       "17",
				}},
				Adjustments: []input.CreateOrderLineAdjustmentInput{{
					Sequence:         1,
					Kind:             "discount",
					SourceType:       "manual",
					NameSnapshot:     "Promo",
					CalculationType:  "fixed_amount",
					CalculationValue: "30",
					Amount:           "-30",
					AppliesBeforeTax: true,
				}},
			}},
			Payments: []input.CreateOrderPaymentInput{{
				PaymentMethodID: paymentMethodID,
				Amount:          "187",
			}},
			Adjustments: []input.CreateOrderAdjustmentInput{{
				Sequence:         1,
				Kind:             "tip",
				SourceType:       "manual",
				NameSnapshot:     "Tip",
				CalculationType:  "fixed_amount",
				CalculationValue: "10",
				Amount:           "10",
				AppliesBeforeTax: false,
				ProrateStrategy:  "no_prorate",
			}},
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		// Aggregates derived in the usecase: subtotal = 2*100 = 200, tax = 17,
		// discount = 30, order tip = +10, total = 200 + 17 + 10 - 30 = 197.
		if created.Subtotal != "200.0000" {
			t.Errorf("subtotal: want 200.0000, got %s", created.Subtotal)
		}
		if created.TaxTotal != "17.0000" {
			t.Errorf("tax_total: want 17.0000, got %s", created.TaxTotal)
		}
		if created.DiscountTotal != "30.0000" {
			t.Errorf("discount_total: want 30.0000, got %s", created.DiscountTotal)
		}
		if created.Total != "197.0000" {
			t.Errorf("total: want 197.0000, got %s", created.Total)
		}

		got, err := uc.GetOrder(ctx, input.GetOrderInput{OrgID: orderTaxOrgA, ID: created.ID})
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if len(got.LineItems) != 1 {
			t.Fatalf("line items: want 1, got %d", len(got.LineItems))
		}
		line := got.LineItems[0]
		if line.TaxAmount != "17.0000" {
			t.Errorf("line tax_amount: want 17.0000, got %s", line.TaxAmount)
		}
		if line.DiscountAmount != "30.0000" {
			t.Errorf("line discount_amount: want 30.0000, got %s", line.DiscountAmount)
		}

		wantTaxes := []*entity.OrderLineItemTax{{
			Sequence:     1,
			TaxRateID:    taxRateID,
			NameSnapshot: "VAT",
			RateSnapshot: "0.1000",
			IsInclusive:  false,
			IsCompound:   false,
			TaxableBase:  "170.0000",
			Amount:       "17.0000",
		}}
		taxOpts := []cmp.Option{cmpopts.IgnoreFields(entity.OrderLineItemTax{}, "ID")}
		if diff := cmp.Diff(wantTaxes, line.Taxes, taxOpts...); diff != "" {
			t.Errorf("line taxes mismatch (-want +got):\n%s", diff)
		}

		wantLineAdj := []*entity.OrderLineAdjustment{{
			Sequence:         1,
			Kind:             "discount",
			SourceType:       "manual",
			NameSnapshot:     "Promo",
			CalculationType:  "fixed_amount",
			CalculationValue: "30.0000",
			Amount:           "-30.0000",
			AppliesBeforeTax: true,
		}}
		adjOpts := []cmp.Option{cmpopts.IgnoreFields(entity.OrderLineAdjustment{}, "ID", "AppliedAt")}
		if diff := cmp.Diff(wantLineAdj, line.Adjustments, adjOpts...); diff != "" {
			t.Errorf("line adjustments mismatch (-want +got):\n%s", diff)
		}

		wantOrderAdj := []*entity.OrderAdjustment{{
			Sequence:         1,
			Kind:             "tip",
			SourceType:       "manual",
			NameSnapshot:     "Tip",
			CalculationType:  "fixed_amount",
			CalculationValue: "10.0000",
			Amount:           "10.0000",
			AppliesBeforeTax: false,
			ProrateStrategy:  "no_prorate",
		}}
		oaOpts := []cmp.Option{cmpopts.IgnoreFields(entity.OrderAdjustment{}, "ID", "AppliedAt")}
		if diff := cmp.Diff(wantOrderAdj, got.Adjustments, oaOpts...); diff != "" {
			t.Errorf("order adjustments mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("idempotent on (source, external_id)", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		in := input.CreateOrderInput{
			OrgID:       orderTaxOrgA,
			Source:      "pos",
			ExternalID:  "ext-idem-1",
			OrderNumber: "HD-INT-IDEM-001",
			StoreID:     storeID,
			UserID:      userID,
			Subtotal:    "100",
			TaxTotal:    "0",
			Total:       "100",
			LineItems: []input.CreateOrderLineItemInput{{
				ProductName: "Latte",
				VariantName: "Default",
				Quantity:    "1",
				UnitPrice:   "100",
				LineTotal:   "100",
			}},
			Payments: []input.CreateOrderPaymentInput{{PaymentMethodID: paymentMethodID, Amount: "100"}},
		}

		first, err := uc.CreateOrder(ctx, in)
		if err != nil {
			t.Fatalf("first create: %v", err)
		}
		second, err := uc.CreateOrder(ctx, in)
		if err != nil {
			t.Fatalf("second create: %v", err)
		}
		if first.ID != second.ID {
			t.Errorf("idempotency broken: first.ID=%s second.ID=%s", first.ID, second.ID)
		}
	})

	t.Run("legacy aggregates path persists when no children supplied", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		created, err := uc.CreateOrder(ctx, input.CreateOrderInput{
			OrgID:         orderTaxOrgA,
			Source:        "pos",
			ExternalID:    "ext-legacy-1",
			OrderNumber:   "HD-INT-LEG-001",
			StoreID:       storeID,
			UserID:        userID,
			Subtotal:      "111",
			TaxTotal:      "11",
			DiscountTotal: "5",
			Total:         "117",
			LineItems: []input.CreateOrderLineItemInput{{
				ProductName:    "Bagel",
				VariantName:    "Default",
				Quantity:       "1",
				UnitPrice:      "111",
				DiscountAmount: "5",
				TaxRate:        "0.1000",
				TaxAmount:      "11",
				LineTotal:      "117",
			}},
			Payments: []input.CreateOrderPaymentInput{{PaymentMethodID: paymentMethodID, Amount: "117"}},
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		// Caller-supplied aggregates must survive untouched (no children to
		// derive from). Postgres NUMERIC(12,4) formats values with 4 fractional
		// digits on read regardless of the input precision, so compare to the
		// formatted form rather than the raw string the caller sent.
		if created.Subtotal != "111.0000" || created.TaxTotal != "11.0000" ||
			created.DiscountTotal != "5.0000" || created.Total != "117.0000" {
			t.Errorf("legacy aggregates rewritten: subtotal=%s tax=%s discount=%s total=%s",
				created.Subtotal, created.TaxTotal, created.DiscountTotal, created.Total)
		}
	})
}

func seedOrderTaxOrg(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	slug := "org-" + orderTaxOrgA[len(orderTaxOrgA)-12:]
	_, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, name, currency, timezone, status)
		VALUES ($1::UUID, $2, $2, 'VND', 'Asia/Ho_Chi_Minh', 'active')
		ON CONFLICT (id) DO NOTHING`, orderTaxOrgA, slug)
	if err != nil {
		t.Fatalf("seed org: %v", err)
	}
}

func cleanupOrderTaxData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	stmts := []string{
		`DELETE FROM order_adjustments WHERE org_id = $1`,
		`DELETE FROM order_line_adjustments WHERE org_id = $1`,
		`DELETE FROM order_line_taxes WHERE org_id = $1`,
		`DELETE FROM payments WHERE org_id = $1`,
		`DELETE FROM order_line_items WHERE org_id = $1`,
		`DELETE FROM orders WHERE org_id = $1`,
		`DELETE FROM payment_methods WHERE org_id = $1`,
		`DELETE FROM tax_rates WHERE org_id = $1`,
		`DELETE FROM user_stores WHERE org_id = $1`,
		`DELETE FROM users WHERE org_id = $1`,
		`DELETE FROM roles WHERE org_id = $1`,
		`DELETE FROM stores WHERE org_id = $1`,
		`DELETE FROM organizations WHERE id = $1`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s, orderTaxOrgA); err != nil {
			t.Logf("cleanup %q: %v", s, err)
		}
	}
}

func insertStoreForOrderTax(t *testing.T, pool *pgxpool.Pool, orgID string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO stores (org_id, name, timezone, status)
		VALUES ($1, 'Order Tax Store', 'Asia/Ho_Chi_Minh', 'active')
		RETURNING id::TEXT`, orgID).Scan(&id); err != nil {
		t.Fatalf("insert store: %v", err)
	}
	return id
}

func insertRoleForOrderTax(t *testing.T, pool *pgxpool.Pool, orgID string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO roles (org_id, name, permissions, is_system)
		VALUES ($1, 'Cashier / order_tax', '{}'::JSONB, FALSE)
		RETURNING id::TEXT`, orgID).Scan(&id); err != nil {
		t.Fatalf("insert role: %v", err)
	}
	return id
}

func insertUserForOrderTax(t *testing.T, pool *pgxpool.Pool, orgID, roleID string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO users (org_id, role_id, name, email, status)
		VALUES ($1, $2, 'Cashier', 'cashier-' || substr(gen_random_uuid()::TEXT, 1, 8) || '@test.local',
		        'active')
		RETURNING id::TEXT`, orgID, roleID).Scan(&id); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertUserStore(t *testing.T, pool *pgxpool.Pool, orgID, userID, storeID string) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO user_stores (org_id, user_id, store_id)
		VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, orgID, userID, storeID); err != nil {
		t.Fatalf("insert user_stores: %v", err)
	}
}

func insertPaymentMethodForOrderTax(t *testing.T, pool *pgxpool.Pool, orgID string) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO payment_methods (org_id, name, type, is_active, sort_order)
		VALUES ($1, 'Cash', 'cash', TRUE, 0)
		RETURNING id::TEXT`, orgID).Scan(&id); err != nil {
		t.Fatalf("insert payment method: %v", err)
	}
	return id
}

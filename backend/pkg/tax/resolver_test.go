package tax_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/genpick/genpos-mono/backend/pkg/money"
	"github.com/genpick/genpos-mono/backend/pkg/tax"
)

func Test_TaxResolver_Resolve(t *testing.T) {
	t.Parallel()

	// Three rate refs reused across cases.
	vat10Excl := tax.RateRef{
		TaxRateID: "rate-vat", NameSnapshot: "VAT 10%",
		Rate: money.MustParse("0.10"), IsInclusive: false, IsCompound: false, Sequence: 0,
	}
	vat10Incl := tax.RateRef{
		TaxRateID: "rate-vat-i", NameSnapshot: "VAT 10% (inclusive)",
		Rate: money.MustParse("0.10"), IsInclusive: true, IsCompound: false, Sequence: 0,
	}
	gst5 := tax.RateRef{
		TaxRateID: "rate-gst", NameSnapshot: "GST 5%",
		Rate: money.MustParse("0.05"), IsInclusive: false, IsCompound: false, Sequence: 0,
	}
	pst8Compound := tax.RateRef{
		TaxRateID: "rate-pst", NameSnapshot: "PST 8% (compound)",
		Rate: money.MustParse("0.08"), IsInclusive: false, IsCompound: true, Sequence: 1,
	}
	servTax5Excl := tax.RateRef{
		TaxRateID: "rate-srv", NameSnapshot: "Service 5%",
		Rate: money.MustParse("0.05"), IsInclusive: false, IsCompound: false, Sequence: 1,
	}
	servTax5Incl := tax.RateRef{
		TaxRateID: "rate-srv-i", NameSnapshot: "Service 5% (inclusive)",
		Rate: money.MustParse("0.05"), IsInclusive: true, IsCompound: false, Sequence: 1,
	}

	cases := map[string]struct {
		in           tax.CartInput
		wantSubtotal string
		wantTax      string
		wantDiscount string
		wantTotal    string
		wantLines    []lineSnapshot
		wantErr      bool
		wantErrMsg   string
	}{
		"single line, no tax, no adjustments": {
			in: tax.CartInput{
				Lines: []tax.LineInput{{
					Quantity:  money.MustParse("1"),
					UnitPrice: money.MustParse("100000"),
				}},
				Round: "per_line",
			},
			wantSubtotal: "100000",
			wantTax:      "0",
			wantDiscount: "0",
			wantTotal:    "100000",
			wantLines: []lineSnapshot{{
				TaxableBase:    "100000",
				DiscountAmount: "0",
				TaxAmount:      "0",
				LineTotal:      "100000",
				Taxes:          nil,
			}},
		},

		"exclusive VAT 10% on a single line": {
			in: tax.CartInput{
				Lines: []tax.LineInput{{
					Quantity:  money.MustParse("1"),
					UnitPrice: money.MustParse("100000"),
					TaxRates:  []tax.RateRef{vat10Excl},
				}},
				Round: "per_line",
			},
			wantSubtotal: "100000",
			wantTax:      "10000",
			wantDiscount: "0",
			wantTotal:    "110000",
			wantLines: []lineSnapshot{{
				TaxableBase:    "100000",
				DiscountAmount: "0",
				TaxAmount:      "10000",
				LineTotal:      "110000",
				Taxes: []taxSnapshot{
					{Name: "VAT 10%", Rate: "0.1", TaxableBase: "100000", Amount: "10000", IsInclusive: false},
				},
			}},
		},

		"inclusive VAT 10% extracted from gross": {
			in: tax.CartInput{
				Lines: []tax.LineInput{{
					Quantity:       money.MustParse("1"),
					UnitPrice:      money.MustParse("110000"),
					IsTaxInclusive: true,
					TaxRates:       []tax.RateRef{vat10Incl},
				}},
				Round: "per_line",
			},
			// pre_tax = 110000 / 1.10 = 100000; vat = 10000; gross = 110000
			wantSubtotal: "100000",
			wantTax:      "10000",
			wantDiscount: "0",
			wantTotal:    "110000",
			wantLines: []lineSnapshot{{
				TaxableBase:    "100000",
				DiscountAmount: "0",
				TaxAmount:      "10000",
				LineTotal:      "110000",
				Taxes: []taxSnapshot{
					{Name: "VAT 10% (inclusive)", Rate: "0.1", TaxableBase: "100000", Amount: "10000", IsInclusive: true},
				},
			}},
		},

		"two parallel exclusive taxes": {
			// Vietnam-style food + service tax, both 10% and 5% on the same base.
			in: tax.CartInput{
				Lines: []tax.LineInput{{
					Quantity:  money.MustParse("1"),
					UnitPrice: money.MustParse("100000"),
					TaxRates:  []tax.RateRef{vat10Excl, servTax5Excl},
				}},
				Round: "per_line",
			},
			wantSubtotal: "100000",
			wantTax:      "15000",
			wantDiscount: "0",
			wantTotal:    "115000",
			wantLines: []lineSnapshot{{
				TaxableBase:    "100000",
				DiscountAmount: "0",
				TaxAmount:      "15000",
				LineTotal:      "115000",
				Taxes: []taxSnapshot{
					{Name: "VAT 10%", Rate: "0.1", TaxableBase: "100000", Amount: "10000", IsInclusive: false},
					{Name: "Service 5%", Rate: "0.05", TaxableBase: "100000", Amount: "5000", IsInclusive: false},
				},
			}},
		},

		"two parallel inclusive taxes (VND food example)": {
			in: tax.CartInput{
				Lines: []tax.LineInput{{
					Quantity:       money.MustParse("1"),
					UnitPrice:      money.MustParse("115000"),
					IsTaxInclusive: true,
					TaxRates:       []tax.RateRef{vat10Incl, servTax5Incl},
				}},
				Round: "per_line",
			},
			// Total inclusive rate = 15%; pre-tax = 115000 / 1.15 = 100000
			wantSubtotal: "100000",
			wantTax:      "15000",
			wantDiscount: "0",
			wantTotal:    "115000",
			wantLines: []lineSnapshot{{
				TaxableBase:    "100000",
				DiscountAmount: "0",
				TaxAmount:      "15000",
				LineTotal:      "115000",
				Taxes: []taxSnapshot{
					{Name: "VAT 10% (inclusive)", Rate: "0.1", TaxableBase: "100000", Amount: "10000", IsInclusive: true},
					{Name: "Service 5% (inclusive)", Rate: "0.05", TaxableBase: "100000", Amount: "5000", IsInclusive: true},
				},
			}},
		},

		"compound exclusive: GST 5% then PST 8% on (base + GST)": {
			// 100000 -> GST = 5000; PST sees (100000 + 5000) = 8400
			// total tax = 13400; line total = 113400
			in: tax.CartInput{
				Lines: []tax.LineInput{{
					Quantity:  money.MustParse("1"),
					UnitPrice: money.MustParse("100000"),
					TaxRates:  []tax.RateRef{gst5, pst8Compound},
				}},
				Round: "per_line",
			},
			wantSubtotal: "100000",
			wantTax:      "13400",
			wantDiscount: "0",
			wantTotal:    "113400",
			wantLines: []lineSnapshot{{
				TaxableBase:    "100000",
				DiscountAmount: "0",
				TaxAmount:      "13400",
				LineTotal:      "113400",
				Taxes: []taxSnapshot{
					{Name: "GST 5%", Rate: "0.05", TaxableBase: "100000", Amount: "5000", IsInclusive: false},
					{Name: "PST 8% (compound)", Rate: "0.08", TaxableBase: "105000", Amount: "8400", IsInclusive: false, IsCompound: true},
				},
			}},
		},

		"line-level percentage discount before tax": {
			// 100000 - 10% = 90000; VAT 10% = 9000; total = 99000
			in: tax.CartInput{
				Lines: []tax.LineInput{{
					Quantity:  money.MustParse("1"),
					UnitPrice: money.MustParse("100000"),
					TaxRates:  []tax.RateRef{vat10Excl},
					Adjustments: []tax.LineAdjustment{{
						Sequence:         0,
						Kind:             tax.KindDiscount,
						SourceType:       tax.SourceManual,
						NameSnapshot:     "10% off",
						CalculationType:  tax.CalcPercentage,
						CalculationValue: money.MustParse("10"),
						AppliesBeforeTax: true,
					}},
				}},
				Round: "per_line",
			},
			wantSubtotal: "90000",
			wantTax:      "9000",
			wantDiscount: "10000",
			wantTotal:    "99000",
			wantLines: []lineSnapshot{{
				TaxableBase:    "90000",
				DiscountAmount: "10000",
				TaxAmount:      "9000",
				LineTotal:      "99000",
				Taxes: []taxSnapshot{
					{Name: "VAT 10%", Rate: "0.1", TaxableBase: "90000", Amount: "9000", IsInclusive: false},
				},
			}},
		},

		"post-tax tip on a single line (no tax impact)": {
			// 100000 + VAT 10000 = 110000; tip 5000 added on top; total 115000
			// taxable_base unchanged; tax unchanged.
			in: tax.CartInput{
				Lines: []tax.LineInput{{
					Quantity:  money.MustParse("1"),
					UnitPrice: money.MustParse("100000"),
					TaxRates:  []tax.RateRef{vat10Excl},
					Adjustments: []tax.LineAdjustment{{
						Sequence:         0,
						Kind:             tax.KindServiceCharge,
						SourceType:       tax.SourceManual,
						NameSnapshot:     "Server tip",
						CalculationType:  tax.CalcFixedAmount,
						CalculationValue: money.MustParse("5000"),
						AppliesBeforeTax: false,
					}},
				}},
				Round: "per_line",
			},
			wantSubtotal: "100000",
			wantTax:      "10000",
			wantDiscount: "0",
			wantTotal:    "115000",
			wantLines: []lineSnapshot{{
				TaxableBase:    "100000",
				DiscountAmount: "0",
				TaxAmount:      "10000",
				LineTotal:      "115000",
				Taxes: []taxSnapshot{
					{Name: "VAT 10%", Rate: "0.1", TaxableBase: "100000", Amount: "10000", IsInclusive: false},
				},
			}},
		},

		"order-level percentage discount pro-rated by taxable base": {
			// Two lines: 50000 + 80000 = 130000 base
			// Order discount 10% = 13000 pro-rated:
			//   line 1 share = 13000 * 50/130 = 5000
			//   line 2 share = 13000 - 5000 = 8000 (last line absorbs remainder)
			// Line 1: base 45000, VAT 4500, total 49500
			// Line 2: base 72000, VAT 7200, total 79200
			// Order subtotal=117000, tax=11700, total=128700, discount=13000
			in: tax.CartInput{
				Lines: []tax.LineInput{
					{Quantity: money.MustParse("1"), UnitPrice: money.MustParse("50000"), TaxRates: []tax.RateRef{vat10Excl}},
					{Quantity: money.MustParse("1"), UnitPrice: money.MustParse("80000"), TaxRates: []tax.RateRef{vat10Excl}},
				},
				OrderAdjustments: []tax.OrderAdjustment{{
					Sequence:         0,
					Kind:             tax.KindDiscount,
					SourceType:       tax.SourceManual,
					NameSnapshot:     "10% off whole order",
					CalculationType:  tax.CalcPercentage,
					CalculationValue: money.MustParse("10"),
					AppliesBeforeTax: true,
					ProrateStrategy:  tax.ProrateTaxableBase,
				}},
				Round: "per_line",
			},
			wantSubtotal: "117000",
			wantTax:      "11700",
			wantDiscount: "13000",
			wantTotal:    "128700",
			wantLines: []lineSnapshot{
				{
					TaxableBase: "45000", DiscountAmount: "5000", TaxAmount: "4500", LineTotal: "49500",
					Taxes: []taxSnapshot{{Name: "VAT 10%", Rate: "0.1", TaxableBase: "45000", Amount: "4500"}},
				},
				{
					TaxableBase: "72000", DiscountAmount: "8000", TaxAmount: "7200", LineTotal: "79200",
					Taxes: []taxSnapshot{{Name: "VAT 10%", Rate: "0.1", TaxableBase: "72000", Amount: "7200"}},
				},
			},
		},

		"no-prorate delivery fee adds to order total only": {
			// Single line 100000 + VAT 10000; delivery 15000 sits at order level.
			// Lines unchanged; order_total = 110000 + 15000 = 125000.
			in: tax.CartInput{
				Lines: []tax.LineInput{
					{Quantity: money.MustParse("1"), UnitPrice: money.MustParse("100000"), TaxRates: []tax.RateRef{vat10Excl}},
				},
				OrderAdjustments: []tax.OrderAdjustment{{
					Sequence:         0,
					Kind:             tax.KindDelivery,
					SourceType:       tax.SourceManual,
					NameSnapshot:     "Bike courier",
					CalculationType:  tax.CalcFixedAmount,
					CalculationValue: money.MustParse("15000"),
					AppliesBeforeTax: false,
					ProrateStrategy:  tax.ProrateNone,
				}},
				Round: "per_line",
			},
			wantSubtotal: "100000",
			wantTax:      "10000",
			wantDiscount: "0",
			wantTotal:    "125000",
			wantLines: []lineSnapshot{
				{
					TaxableBase: "100000", DiscountAmount: "0", TaxAmount: "10000", LineTotal: "110000",
					Taxes: []taxSnapshot{{Name: "VAT 10%", Rate: "0.1", TaxableBase: "100000", Amount: "10000"}},
				},
			},
		},

		"empty cart is rejected": {
			in:         tax.CartInput{Lines: nil},
			wantErr:    true,
			wantErrMsg: "at least one line is required",
		},

		"compound + inclusive on the same rate is rejected": {
			in: tax.CartInput{
				Lines: []tax.LineInput{{
					Quantity:       money.MustParse("1"),
					UnitPrice:      money.MustParse("100000"),
					IsTaxInclusive: true,
					TaxRates: []tax.RateRef{{
						TaxRateID: "x", NameSnapshot: "Bad",
						Rate: money.MustParse("0.10"), IsInclusive: true, IsCompound: true,
					}},
				}},
			},
			wantErr:    true,
			wantErrMsg: "compound + inclusive on the same rate is not supported",
		},

		"mixing inclusive and exclusive on the same line is rejected": {
			in: tax.CartInput{
				Lines: []tax.LineInput{{
					Quantity:       money.MustParse("1"),
					UnitPrice:      money.MustParse("100000"),
					IsTaxInclusive: true,
					TaxRates:       []tax.RateRef{vat10Incl, vat10Excl},
				}},
			},
			wantErr:    true,
			wantErrMsg: "rate is_inclusive must match the line's is_tax_inclusive flag",
		},

		"unknown calculation_type at line level rejected": {
			in: tax.CartInput{
				Lines: []tax.LineInput{{
					Quantity:  money.MustParse("1"),
					UnitPrice: money.MustParse("100000"),
					Adjustments: []tax.LineAdjustment{{
						Sequence: 0, Kind: tax.KindDiscount, SourceType: tax.SourceManual,
						NameSnapshot: "Bad", CalculationType: "mystery",
						CalculationValue: money.MustParse("10"),
					}},
				}},
			},
			wantErr:    true,
			wantErrMsg: "unknown calculation_type: mystery",
		},

		"fixed_price at order level is rejected": {
			in: tax.CartInput{
				Lines: []tax.LineInput{
					{Quantity: money.MustParse("1"), UnitPrice: money.MustParse("100000")},
				},
				OrderAdjustments: []tax.OrderAdjustment{{
					Sequence: 0, Kind: tax.KindDiscount, SourceType: tax.SourceManual,
					NameSnapshot: "Bad", CalculationType: tax.CalcFixedPrice,
					CalculationValue: money.MustParse("80000"),
					ProrateStrategy:  tax.ProrateTaxableBase,
				}},
			},
			wantErr:    true,
			wantErrMsg: "fixed_price is not valid at order level",
		},

		"per_order rounding keeps fractional precision on lines": {
			// 33333 * 3 = 99999 face value; VAT 10% = 9999.9 (fractional).
			// per_order keeps lines at full precision, then rounds each
			// aggregate (Subtotal/TaxTotal/Total) once. Lines display
			// fractional values; the receipt totals are bankers-rounded.
			in: tax.CartInput{
				Lines: []tax.LineInput{
					{Quantity: money.MustParse("3"), UnitPrice: money.MustParse("33333"), TaxRates: []tax.RateRef{vat10Excl}},
				},
				Round: "per_order",
			},
			wantSubtotal: "99999",
			wantTax:      "10000", // 9999.9 bankers-rounded to 10000
			wantDiscount: "0",
			wantTotal:    "109999", // 109998.9 bankers-rounded to 109999
			// per_order skips line rounding so we expect fractional tax_amount per line
			wantLines: []lineSnapshot{{
				TaxableBase: "99999", DiscountAmount: "0",
				TaxAmount: "9999.9", LineTotal: "109998.9",
				Taxes: []taxSnapshot{{Name: "VAT 10%", Rate: "0.1", TaxableBase: "99999", Amount: "9999.9"}},
			}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := tax.Resolve(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error %q, got result %+v", tc.wantErrMsg, got)
				}
				if !contains(err.Error(), tc.wantErrMsg) {
					t.Errorf("error message: want substring %q, got %q", tc.wantErrMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Subtotal.String() != tc.wantSubtotal {
				t.Errorf("Subtotal: want %s, got %s", tc.wantSubtotal, got.Subtotal)
			}
			if got.TaxTotal.String() != tc.wantTax {
				t.Errorf("TaxTotal: want %s, got %s", tc.wantTax, got.TaxTotal)
			}
			if got.DiscountTotal.String() != tc.wantDiscount {
				t.Errorf("DiscountTotal: want %s, got %s", tc.wantDiscount, got.DiscountTotal)
			}
			if got.Total.String() != tc.wantTotal {
				t.Errorf("Total: want %s, got %s", tc.wantTotal, got.Total)
			}

			if len(got.Lines) != len(tc.wantLines) {
				t.Fatalf("line count: want %d, got %d", len(tc.wantLines), len(got.Lines))
			}
			for i, want := range tc.wantLines {
				snap := snapshotLine(got.Lines[i])
				if diff := cmp.Diff(want, snap); diff != "" {
					t.Errorf("line[%d] (-want +got):\n%s", i, diff)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Snapshot helpers: collapse decimal.Decimal into canonical strings so the
// table cases stay readable. cmp.Diff over the snapshot structs gives clean
// diffs without dragging the full decimal type into the want-set.
// ---------------------------------------------------------------------------

type lineSnapshot struct {
	TaxableBase    string
	DiscountAmount string
	TaxAmount      string
	LineTotal      string
	Taxes          []taxSnapshot
}

type taxSnapshot struct {
	Name        string
	Rate        string
	TaxableBase string
	Amount      string
	IsInclusive bool
	IsCompound  bool
}

func snapshotLine(lr tax.LineResult) lineSnapshot {
	out := lineSnapshot{
		TaxableBase:    lr.TaxableBase.String(),
		DiscountAmount: lr.DiscountAmount.String(),
		TaxAmount:      lr.TaxAmount.String(),
		LineTotal:      lr.LineTotal.String(),
	}
	for _, t := range lr.Taxes {
		out.Taxes = append(out.Taxes, taxSnapshot{
			Name:        t.NameSnapshot,
			Rate:        t.RateSnapshot.String(),
			TaxableBase: t.TaxableBase.String(),
			Amount:      t.Amount.String(),
			IsInclusive: t.IsInclusive,
			IsCompound:  t.IsCompound,
		})
	}
	return out
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

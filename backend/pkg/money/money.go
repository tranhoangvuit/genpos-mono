// Package money is the arithmetic surface for monetary amounts and tax
// rates, anchored on shopspring/decimal so calculations are exact (never
// float64). Storage and transport stay as decimal strings -- this package
// only operates inside computation paths (the tax resolver, future
// invoicing, etc.) and serialises back to string at the boundary.
//
// Schema anchors:
//   - Money columns are NUMERIC(12,4) -- 4 fractional digits.
//   - Rate columns are NUMERIC(6,4)   -- 4 fractional digits.
//
// VND has no fractional unit but the schema still stores 4 places to keep
// inclusive-tax math reversible (extracting tax from a gross amount can
// produce sub-unit intermediates).
package money

import (
	"strings"

	"github.com/shopspring/decimal"
)

// MoneyScale is the decimal places NUMERIC(12,4) carries.
const MoneyScale int32 = 4

// RateScale is the decimal places NUMERIC(6,4) carries.
const RateScale int32 = 4

// RoundStrategy controls when individual line totals are quantised to a
// whole monetary unit. The default is per-line: every line displays an
// integer that sums exactly to the printed total. Per-order keeps lines at
// full precision and rounds only the order total once -- minor display drift
// at the line level but penny-perfect at the bottom.
type RoundStrategy string

const (
	RoundPerLine  RoundStrategy = "per_line"
	RoundPerOrder RoundStrategy = "per_order"
)

// Zero is the canonical zero amount.
var Zero = decimal.Zero

// Parse turns a decimal string into decimal.Decimal. Empty string is treated
// as zero -- mirrors numericFromString in the datastore layer.
func Parse(s string) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Zero, nil
	}
	return decimal.NewFromString(s)
}

// MustParse panics on parse error. Use only in tests / hard-coded constants.
func MustParse(s string) decimal.Decimal {
	d, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return d
}

// String renders a decimal at MoneyScale (e.g. "100.0000") so writes land in
// NUMERIC(12,4) without surprises.
func String(d decimal.Decimal) string {
	return d.StringFixed(MoneyScale)
}

// RateString renders a rate at RateScale (e.g. "0.1000").
func RateString(d decimal.Decimal) string {
	return d.StringFixed(RateScale)
}

// Round applies the configured rounding to a monetary value. VND has zero
// fractional digits in practice, so per-line rounding snaps to whole units;
// per-order rounding keeps full precision until the order total stage.
func Round(d decimal.Decimal, strategy RoundStrategy) decimal.Decimal {
	if strategy == RoundPerOrder {
		return d
	}
	// Bankers' rounding mirrors shopspring's RoundBank: half-to-even is the
	// fairest choice when rounding many tax fractions in aggregate. For VND we
	// snap to 0 places; if the schema ever supports fractional currency this
	// will need a per-currency lookup.
	return d.RoundBank(0)
}

// RoundOrderTotal is invoked once on the final order total when the strategy
// is per_order. For per_line the line totals already sum to integers, so this
// is a no-op.
func RoundOrderTotal(d decimal.Decimal, strategy RoundStrategy) decimal.Decimal {
	if strategy == RoundPerLine {
		return d
	}
	return d.RoundBank(0)
}

// Sum totals a slice of decimals. Empty slice returns zero.
func Sum(values ...decimal.Decimal) decimal.Decimal {
	out := Zero
	for _, v := range values {
		out = out.Add(v)
	}
	return out
}

// IsValidRoundStrategy reports whether s is a known strategy. Unknown values
// fall back to per_line at read time, but this lets the engine flag bad
// store_config rows.
func IsValidRoundStrategy(s string) bool {
	switch RoundStrategy(s) {
	case RoundPerLine, RoundPerOrder:
		return true
	default:
		return false
	}
}

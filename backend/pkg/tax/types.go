// Package tax is the pure-functional resolver that turns a cart (lines +
// rate memberships + adjustments) into the snapshot rows that
// CreateOrder persists: order_line_taxes, order_line_adjustments,
// order_adjustments. It implements the canonical order of operations
// agreed in the PR 1 design:
//
//  1. raw line base = quantity * unit_price
//  2. line-level adjustments applied in sequence; pre-tax adjustments
//     reduce the taxable base, post-tax adjustments add to line total only
//  3. order-level adjustments distributed across lines per prorate_strategy
//  4. taxes computed per line on the resulting taxable base
//     (compound rates see base + previously-applied taxes; non-compound
//     always see the original taxable base)
//  5. summed back to order aggregates
//
// CreateOrder does not invoke this resolver in PR 2 -- the desk POS
// computes locally and uploads pre-resolved children. The resolver lives
// here for use by future server-side carts (admin preview, online store)
// and as the spec for the desk port in PR 5.
package tax

import "github.com/shopspring/decimal"

// Adjustment kinds, source types, calculation types, and prorate strategies
// duplicated from the SQL CHECK constraints to keep the engine validation
// in step with the schema.
const (
	KindDiscount       = "discount"
	KindPromotion      = "promotion"
	KindFee            = "fee"
	KindServiceCharge  = "service_charge"
	KindComp           = "comp"
	KindTip            = "tip"
	KindDelivery       = "delivery"
	KindRounding       = "rounding"

	SourceManual         = "manual"
	SourcePromotionRule  = "promotion_rule"
	SourceCoupon         = "coupon"
	SourceLoyalty        = "loyalty"
	SourceCustomerGroup  = "customer_group"
	SourceAuto           = "auto"
	SourceSystem         = "system"

	CalcPercentage  = "percentage"
	CalcFixedAmount = "fixed_amount"
	CalcFixedPrice  = "fixed_price"

	ProrateTaxableBase = "pro_rata_taxable_base"
	ProrateQty         = "pro_rata_qty"
	ProrateNone        = "no_prorate"
)

// RateRef is one tax rate inside a line's tax class, with the snapshot fields
// that survive onto order_line_taxes after the source row is edited or
// deleted.
type RateRef struct {
	TaxRateID    string
	NameSnapshot string
	Rate         decimal.Decimal // fraction (0.10 = 10%)
	IsInclusive  bool
	IsCompound   bool
	Sequence     int
}

// LineAdjustment is a per-line discount, fee, comp, etc. Sequence drives
// application order.
type LineAdjustment struct {
	Sequence           int
	Kind               string
	SourceType         string
	SourceID           string
	SourceCodeSnapshot string
	NameSnapshot       string
	Reason             string
	CalculationType    string
	CalculationValue   decimal.Decimal
	AppliesBeforeTax   bool
	AppliedBy          string
	ApprovedBy         string
}

// OrderAdjustment is a cart-wide adjustment (order discount, delivery fee,
// tip, system rounding). ProrateStrategy controls distribution onto lines.
type OrderAdjustment struct {
	Sequence           int
	Kind               string
	SourceType         string
	SourceID           string
	SourceCodeSnapshot string
	NameSnapshot       string
	Reason             string
	CalculationType    string
	CalculationValue   decimal.Decimal
	AppliesBeforeTax   bool
	ProrateStrategy    string
	AppliedBy          string
	ApprovedBy         string
}

// LineInput is one cart line.
//
// Quantity, UnitPrice are the as-rung values (unit price already inclusive
// of tax when IsTaxInclusive=true; pre-tax otherwise). TaxRates are the
// rates from the variant's tax_class, in the class's stored sequence order.
// Adjustments are line-level; order-wide adjustments come in via
// CartInput.OrderAdjustments and are pro-rated by the resolver.
type LineInput struct {
	Quantity       decimal.Decimal
	UnitPrice      decimal.Decimal
	IsTaxInclusive bool
	TaxRates       []RateRef
	Adjustments    []LineAdjustment
}

// CartInput is the full resolver input.
type CartInput struct {
	Lines            []LineInput
	OrderAdjustments []OrderAdjustment
	Round            string // money.RoundStrategy as string; falls back to per_line on unknown
}

// ResolvedTax is one snapshot tax row for a line.
type ResolvedTax struct {
	Sequence     int
	TaxRateID    string
	NameSnapshot string
	RateSnapshot decimal.Decimal
	IsInclusive  bool
	IsCompound   bool
	TaxableBase  decimal.Decimal
	Amount       decimal.Decimal
}

// ResolvedAdjustment is the resolver's view of an applied adjustment with
// its computed monetary amount (signed: negative for discount/comp,
// positive for fee/service_charge/tip/delivery).
type ResolvedAdjustment struct {
	Sequence           int
	Kind               string
	SourceType         string
	SourceID           string
	SourceCodeSnapshot string
	NameSnapshot       string
	Reason             string
	CalculationType    string
	CalculationValue   decimal.Decimal
	Amount             decimal.Decimal
	AppliesBeforeTax   bool
	AppliedBy          string
	ApprovedBy         string
}

// ResolvedOrderAdjustment is a ResolvedAdjustment plus the prorate strategy.
type ResolvedOrderAdjustment struct {
	ResolvedAdjustment
	ProrateStrategy string
}

// LineResult is the per-line output. Aggregates (DiscountAmount, TaxAmount,
// LineTotal, EffectiveRate) match the snapshot fields stored on
// order_line_items so the persistence layer can write both children and
// aggregates from a single source.
type LineResult struct {
	TaxableBase    decimal.Decimal
	DiscountAmount decimal.Decimal // absolute value of all negative adjustments allocated to the line
	TaxAmount      decimal.Decimal
	EffectiveRate  decimal.Decimal // sum of rate snapshots (matches existing OrderLineItem.tax_rate aggregate)
	LineTotal      decimal.Decimal
	Taxes          []ResolvedTax
	Adjustments    []ResolvedAdjustment // line-level only; allocated order shares are reflected in DiscountAmount/TaxableBase
}

// Result is the resolver output.
type Result struct {
	Lines            []LineResult
	OrderAdjustments []ResolvedOrderAdjustment

	Subtotal      decimal.Decimal // sum(line taxable bases, pre-tax-extraction for inclusive)
	TaxTotal      decimal.Decimal
	DiscountTotal decimal.Decimal // absolute value of all negative-amount adjustments (line + order)
	Total         decimal.Decimal
}

package tax

import (
	"sort"

	"github.com/shopspring/decimal"

	"github.com/genpick/genpos-mono/backend/pkg/errors"
	"github.com/genpick/genpos-mono/backend/pkg/money"
)

var (
	hundred = decimal.NewFromInt(100)
	one     = decimal.NewFromInt(1)
)

// Resolve runs the full tax + adjustment calculation. Returns a domain
// error (errors.BadRequest) when the input is malformed, or a wrapped
// error for unsupported configurations (compound + inclusive on the same
// line).
func Resolve(in CartInput) (*Result, error) {
	if len(in.Lines) == 0 {
		return nil, errors.BadRequest("at least one line is required")
	}

	round := money.RoundStrategy(in.Round)
	if !money.IsValidRoundStrategy(string(round)) {
		round = money.RoundPerLine
	}

	lines := make([]lineState, len(in.Lines))
	for i, raw := range in.Lines {
		s, err := buildLineState(raw)
		if err != nil {
			return nil, err
		}
		lines[i] = s
	}

	// Apply line-level adjustments in sequence. Pre-tax adjustments mutate
	// taxable_base; post-tax adjustments are recorded but applied to
	// line_total at the very end.
	for i := range lines {
		if err := applyLineAdjustments(&lines[i]); err != nil {
			return nil, err
		}
	}

	// Distribute order-level adjustments across lines per their prorate
	// strategy. no_prorate adjustments are stashed for the order-total step.
	resolvedOrderAdjustments, orderTotalAdjustments, err := allocateOrderAdjustments(in.OrderAdjustments, lines)
	if err != nil {
		return nil, err
	}

	// Compute taxes per line on the final taxable base.
	for i := range lines {
		if err := computeTaxes(&lines[i]); err != nil {
			return nil, err
		}
	}

	// Apply rounding per strategy and collect aggregates.
	result := &Result{
		Lines:            make([]LineResult, len(lines)),
		OrderAdjustments: resolvedOrderAdjustments,
	}
	subtotal := money.Zero
	taxTotal := money.Zero
	discountTotal := money.Zero
	total := money.Zero

	for i := range lines {
		lr := finaliseLine(&lines[i], round)
		result.Lines[i] = lr
		subtotal = subtotal.Add(lr.TaxableBase)
		taxTotal = taxTotal.Add(lr.TaxAmount)
		discountTotal = discountTotal.Add(lr.DiscountAmount)
		total = total.Add(lr.LineTotal)
	}

	// Order-level no_prorate adjustments add to / subtract from total.
	// They also count toward discount_total when negative.
	for _, oa := range orderTotalAdjustments {
		total = total.Add(oa.Amount)
		if oa.Amount.IsNegative() {
			discountTotal = discountTotal.Add(oa.Amount.Abs())
		}
	}

	// Pro-rated order discounts also count toward order discount_total
	// (their per-line shares already affected line.DiscountAmount, but
	// summing line.DiscountAmount above already covered that).
	// However for non-prorated negative adjustments not yet counted via
	// lines, we already added them in the loop above.

	result.Subtotal = money.RoundOrderTotal(subtotal, round)
	result.TaxTotal = money.RoundOrderTotal(taxTotal, round)
	result.DiscountTotal = money.RoundOrderTotal(discountTotal, round)
	result.Total = money.RoundOrderTotal(total, round)

	return result, nil
}

// ---------------------------------------------------------------------------
// Internal pipeline state. lineState carries mutable working values through
// the four pipeline stages; we copy out to LineResult only at the end.
// ---------------------------------------------------------------------------

type lineState struct {
	in              LineInput
	rawBase         decimal.Decimal // qty * unit_price (face value)
	taxableBase     decimal.Decimal // current working base; mutated by adjustments
	postTaxAdjAmt   decimal.Decimal // sum of resolved post-tax line adjustments (signed)
	resolvedAdjs    []ResolvedAdjustment
	allocatedShares decimal.Decimal // sum of allocated order-adjustment shares (signed)
	taxes           []ResolvedTax
	taxAmount       decimal.Decimal
}

func buildLineState(in LineInput) (lineState, error) {
	if in.Quantity.IsZero() {
		return lineState{}, errors.BadRequest("line quantity must be non-zero")
	}
	// Copy and sort tax rates by sequence so the engine is independent of
	// caller-side ordering.
	rates := make([]RateRef, len(in.TaxRates))
	copy(rates, in.TaxRates)
	sort.SliceStable(rates, func(i, j int) bool { return rates[i].Sequence < rates[j].Sequence })
	in.TaxRates = rates

	if err := validateRates(in.IsTaxInclusive, in.TaxRates); err != nil {
		return lineState{}, err
	}

	raw := in.Quantity.Mul(in.UnitPrice)
	return lineState{
		in:          in,
		rawBase:     raw,
		taxableBase: raw,
	}, nil
}

func validateRates(lineInclusive bool, rates []RateRef) error {
	for _, r := range rates {
		if r.IsCompound && r.IsInclusive {
			return errors.BadRequest("compound + inclusive on the same rate is not supported")
		}
		// Mixing inclusive and exclusive rates on the same line is an
		// accounting ambiguity (which base is the inclusive extracted from?).
		// Reject rather than guess.
		if r.IsInclusive != lineInclusive {
			return errors.BadRequest("rate is_inclusive must match the line's is_tax_inclusive flag")
		}
	}
	return nil
}

func applyLineAdjustments(ls *lineState) error {
	adjs := make([]LineAdjustment, len(ls.in.Adjustments))
	copy(adjs, ls.in.Adjustments)
	sort.SliceStable(adjs, func(i, j int) bool { return adjs[i].Sequence < adjs[j].Sequence })

	for _, a := range adjs {
		amt, err := computeAdjustmentAmount(a.CalculationType, a.CalculationValue, a.Kind, ls.taxableBase, ls.in.Quantity)
		if err != nil {
			return err
		}
		resolved := ResolvedAdjustment{
			Sequence:           a.Sequence,
			Kind:               a.Kind,
			SourceType:         a.SourceType,
			SourceID:           a.SourceID,
			SourceCodeSnapshot: a.SourceCodeSnapshot,
			NameSnapshot:       a.NameSnapshot,
			Reason:             a.Reason,
			CalculationType:    a.CalculationType,
			CalculationValue:   a.CalculationValue,
			Amount:             amt,
			AppliesBeforeTax:   a.AppliesBeforeTax,
			AppliedBy:          a.AppliedBy,
			ApprovedBy:         a.ApprovedBy,
		}
		ls.resolvedAdjs = append(ls.resolvedAdjs, resolved)

		if a.AppliesBeforeTax {
			ls.taxableBase = ls.taxableBase.Add(amt)
		} else {
			ls.postTaxAdjAmt = ls.postTaxAdjAmt.Add(amt)
		}
	}
	return nil
}

// computeAdjustmentAmount turns a calculation_type + calculation_value into
// the signed monetary amount the adjustment represents. Sign convention:
// discount and comp produce negatives; fee, service_charge, promotion,
// delivery, tip produce positives. promotion is assumed to be a discount-
// shaped reduction unless the caller pre-signed it via fixed_amount.
func computeAdjustmentAmount(
	calcType string,
	calcValue decimal.Decimal,
	kind string,
	currentBase decimal.Decimal,
	qty decimal.Decimal,
) (decimal.Decimal, error) {
	sign := signForKind(kind)
	switch calcType {
	case CalcPercentage:
		// Applied to the current working base, so a 10% discount on a
		// post-line-item-discount base shrinks the next adjustment too --
		// matches "sequence" semantics.
		share := currentBase.Mul(calcValue).Div(hundred)
		return share.Mul(sign), nil
	case CalcFixedAmount:
		return calcValue.Mul(sign), nil
	case CalcFixedPrice:
		// A fixed_price adjustment overrides the unit price for the line.
		// The amount is the delta from the current base. For order-level
		// adjustments we treat fixed_price as "apply this absolute amount"
		// (caller computes the per-line price elsewhere); since we only
		// compute this for line adjustments it's safe to assume per-unit.
		newBase := calcValue.Mul(qty)
		return newBase.Sub(currentBase), nil
	default:
		return money.Zero, errors.BadRequest("unknown calculation_type: " + calcType)
	}
}

func signForKind(kind string) decimal.Decimal {
	switch kind {
	case KindDiscount, KindComp, KindPromotion:
		return decimal.NewFromInt(-1)
	default:
		return one
	}
}

// allocateOrderAdjustments distributes each order-level adjustment per its
// prorate strategy. pro_rata_taxable_base weights by current line.taxableBase;
// pro_rata_qty weights by quantity. no_prorate adjustments are returned
// separately for the order-total computation.
func allocateOrderAdjustments(
	in []OrderAdjustment,
	lines []lineState,
) ([]ResolvedOrderAdjustment, []ResolvedAdjustment, error) {
	resolved := make([]ResolvedOrderAdjustment, 0, len(in))
	noProrate := make([]ResolvedAdjustment, 0)

	// Sort by sequence so allocations are deterministic.
	sorted := make([]OrderAdjustment, len(in))
	copy(sorted, in)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Sequence < sorted[j].Sequence })

	for _, a := range sorted {
		// Compute the absolute order-level amount on the current cart total.
		// percentage uses sum of taxable bases; fixed_amount uses value as-is;
		// fixed_price is invalid at order level.
		amt, err := computeOrderAdjustmentAmount(a, lines)
		if err != nil {
			return nil, nil, err
		}
		resolvedAdj := ResolvedAdjustment{
			Sequence:           a.Sequence,
			Kind:               a.Kind,
			SourceType:         a.SourceType,
			SourceID:           a.SourceID,
			SourceCodeSnapshot: a.SourceCodeSnapshot,
			NameSnapshot:       a.NameSnapshot,
			Reason:             a.Reason,
			CalculationType:    a.CalculationType,
			CalculationValue:   a.CalculationValue,
			Amount:             amt,
			AppliesBeforeTax:   a.AppliesBeforeTax,
			AppliedBy:          a.AppliedBy,
			ApprovedBy:         a.ApprovedBy,
		}
		resolved = append(resolved, ResolvedOrderAdjustment{
			ResolvedAdjustment: resolvedAdj,
			ProrateStrategy:    a.ProrateStrategy,
		})

		switch a.ProrateStrategy {
		case ProrateNone:
			noProrate = append(noProrate, resolvedAdj)
		case ProrateTaxableBase:
			distribute(amt, lines, weightsByTaxableBase(lines), a.AppliesBeforeTax)
		case ProrateQty:
			distribute(amt, lines, weightsByQty(lines), a.AppliesBeforeTax)
		default:
			return nil, nil, errors.BadRequest("unknown prorate_strategy: " + a.ProrateStrategy)
		}
	}
	return resolved, noProrate, nil
}

func computeOrderAdjustmentAmount(a OrderAdjustment, lines []lineState) (decimal.Decimal, error) {
	sign := signForKind(a.Kind)
	switch a.CalculationType {
	case CalcPercentage:
		base := money.Zero
		for _, l := range lines {
			base = base.Add(l.taxableBase)
		}
		return base.Mul(a.CalculationValue).Div(hundred).Mul(sign), nil
	case CalcFixedAmount:
		return a.CalculationValue.Mul(sign), nil
	case CalcFixedPrice:
		return money.Zero, errors.BadRequest("fixed_price is not valid at order level")
	default:
		return money.Zero, errors.BadRequest("unknown calculation_type: " + a.CalculationType)
	}
}

// distribute spreads the signed amount across lines per the supplied
// weights. The last line absorbs any rounding remainder so the per-line
// shares always sum to the original amount exactly.
func distribute(total decimal.Decimal, lines []lineState, weights []decimal.Decimal, appliesBeforeTax bool) {
	if total.IsZero() || len(lines) == 0 {
		return
	}
	weightSum := money.Zero
	for _, w := range weights {
		weightSum = weightSum.Add(w)
	}
	if weightSum.IsZero() {
		// Degenerate: cart is all-zero base (e.g. pre-tax discount that
		// already reduced base to zero). Nothing to distribute.
		return
	}

	allocated := money.Zero
	for i := range lines {
		var share decimal.Decimal
		if i == len(lines)-1 {
			share = total.Sub(allocated)
		} else {
			share = total.Mul(weights[i]).Div(weightSum)
			allocated = allocated.Add(share)
		}
		if appliesBeforeTax {
			lines[i].taxableBase = lines[i].taxableBase.Add(share)
		} else {
			lines[i].postTaxAdjAmt = lines[i].postTaxAdjAmt.Add(share)
		}
		lines[i].allocatedShares = lines[i].allocatedShares.Add(share)
	}
}

func weightsByTaxableBase(lines []lineState) []decimal.Decimal {
	out := make([]decimal.Decimal, len(lines))
	for i, l := range lines {
		out[i] = l.taxableBase
	}
	return out
}

func weightsByQty(lines []lineState) []decimal.Decimal {
	out := make([]decimal.Decimal, len(lines))
	for i, l := range lines {
		out[i] = l.in.Quantity
	}
	return out
}

// computeTaxes walks the line's tax rates in sequence, producing one
// ResolvedTax per rate. Inclusive rates are extracted from the gross base
// (taxableBase is the gross when IsTaxInclusive=true); exclusive rates are
// added on top, with compound rates seeing base + previously applied taxes.
func computeTaxes(ls *lineState) error {
	if len(ls.in.TaxRates) == 0 {
		return nil
	}
	if ls.in.IsTaxInclusive {
		return computeInclusiveTaxes(ls)
	}
	return computeExclusiveTaxes(ls)
}

func computeInclusiveTaxes(ls *lineState) error {
	// Gross = pre_tax * (1 + sum_of_rates) for non-compound parallel taxes.
	// Compound inclusive was rejected at validation time.
	gross := ls.taxableBase
	rateSum := money.Zero
	for _, r := range ls.in.TaxRates {
		rateSum = rateSum.Add(r.Rate)
	}
	if rateSum.IsZero() {
		return nil
	}
	preTax := gross.Div(one.Add(rateSum))
	// taxable_base reported on the line is the pre-tax amount (subtotal-style).
	ls.taxableBase = preTax

	for _, r := range ls.in.TaxRates {
		amt := preTax.Mul(r.Rate)
		ls.taxes = append(ls.taxes, ResolvedTax{
			Sequence:     r.Sequence,
			TaxRateID:    r.TaxRateID,
			NameSnapshot: r.NameSnapshot,
			RateSnapshot: r.Rate,
			IsInclusive:  true,
			IsCompound:   false,
			TaxableBase:  preTax,
			Amount:       amt,
		})
		ls.taxAmount = ls.taxAmount.Add(amt)
	}
	return nil
}

func computeExclusiveTaxes(ls *lineState) error {
	original := ls.taxableBase
	runningTaxSum := money.Zero
	for _, r := range ls.in.TaxRates {
		base := original
		if r.IsCompound {
			base = original.Add(runningTaxSum)
		}
		amt := base.Mul(r.Rate)
		ls.taxes = append(ls.taxes, ResolvedTax{
			Sequence:     r.Sequence,
			TaxRateID:    r.TaxRateID,
			NameSnapshot: r.NameSnapshot,
			RateSnapshot: r.Rate,
			IsInclusive:  false,
			IsCompound:   r.IsCompound,
			TaxableBase:  base,
			Amount:       amt,
		})
		ls.taxAmount = ls.taxAmount.Add(amt)
		runningTaxSum = runningTaxSum.Add(amt)
	}
	return nil
}

// finaliseLine applies the rounding strategy and copies out to a LineResult.
//
// LineTotal computation:
//   - exclusive: taxable_base + tax_amount + post_tax_adjustments
//   - inclusive: taxable_base + tax_amount + post_tax_adjustments
//     (taxable_base after extraction is pre-tax; adding tax_amount back
//     gives the gross, which equals the original unit_price*qty net of
//     pre-tax discounts)
//
// DiscountAmount reports the absolute value of all negative-amount
// resolved adjustments allocated to this line (line-level + order-level
// shares). This matches the existing aggregate semantics on
// order_line_items.discount_amount.
func finaliseLine(ls *lineState, round money.RoundStrategy) LineResult {
	taxableBase := money.Round(ls.taxableBase, round)
	taxAmount := money.Round(ls.taxAmount, round)
	postTax := money.Round(ls.postTaxAdjAmt, round)

	roundedTaxes := make([]ResolvedTax, len(ls.taxes))
	for i, t := range ls.taxes {
		roundedTaxes[i] = ResolvedTax{
			Sequence:     t.Sequence,
			TaxRateID:    t.TaxRateID,
			NameSnapshot: t.NameSnapshot,
			RateSnapshot: t.RateSnapshot,
			IsInclusive:  t.IsInclusive,
			IsCompound:   t.IsCompound,
			TaxableBase:  money.Round(t.TaxableBase, round),
			Amount:       money.Round(t.Amount, round),
		}
	}

	roundedAdjs := make([]ResolvedAdjustment, len(ls.resolvedAdjs))
	for i, a := range ls.resolvedAdjs {
		roundedAdjs[i] = a
		roundedAdjs[i].Amount = money.Round(a.Amount, round)
	}

	// Aggregate discount: sum the absolute value of every negative-amount
	// influence on this line -- line-level resolved adjustments plus the
	// allocated order-share. Positive influences (fees, tips) don't count.
	discountAmt := money.Zero
	for _, a := range roundedAdjs {
		if a.Amount.IsNegative() {
			discountAmt = discountAmt.Add(a.Amount.Abs())
		}
	}
	if ls.allocatedShares.IsNegative() {
		discountAmt = discountAmt.Add(money.Round(ls.allocatedShares.Abs(), round))
	}

	// EffectiveRate is the sum of rate snapshots, matching the pre-PR-1
	// aggregate semantics on order_line_items.tax_rate.
	effRate := money.Zero
	for _, r := range ls.in.TaxRates {
		effRate = effRate.Add(r.Rate)
	}

	lineTotal := taxableBase.Add(taxAmount).Add(postTax)

	return LineResult{
		TaxableBase:    taxableBase,
		DiscountAmount: discountAmt,
		TaxAmount:      taxAmount,
		EffectiveRate:  effRate,
		LineTotal:      lineTotal,
		Taxes:          roundedTaxes,
		Adjustments:    roundedAdjs,
	}
}

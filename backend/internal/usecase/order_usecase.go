package usecase

import (
	"context"
	"time"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
	"github.com/genpick/genpos-mono/backend/pkg/money"
)

type orderUsecase struct {
	tenantDB     gateway.TenantDB
	reader       gateway.OrderReader
	writer       gateway.OrderWriter
	storeReader  gateway.OrgStoreReader
	memberReader gateway.MemberReader
}

// NewOrderUsecase constructs an OrderUsecase.
func NewOrderUsecase(
	tenantDB gateway.TenantDB,
	reader gateway.OrderReader,
	writer gateway.OrderWriter,
	storeReader gateway.OrgStoreReader,
	memberReader gateway.MemberReader,
) OrderUsecase {
	return &orderUsecase{
		tenantDB:     tenantDB,
		reader:       reader,
		writer:       writer,
		storeReader:  storeReader,
		memberReader: memberReader,
	}
}

func (u *orderUsecase) GetOrder(ctx context.Context, in input.GetOrderInput) (*entity.Order, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	var out *entity.Order
	if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		o, err := u.reader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if hErr := u.hydrate(ctx, o); hErr != nil {
			return hErr
		}
		out = o
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "get order")
	}
	return out, nil
}

func (u *orderUsecase) CreateOrder(ctx context.Context, in input.CreateOrderInput) (*entity.Order, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if in.OrderNumber == "" {
		return nil, errors.BadRequest("order number is required")
	}
	if len(in.LineItems) == 0 {
		return nil, errors.BadRequest("at least one line item is required")
	}

	source := in.Source
	if source == "" {
		source = "pos"
	}
	if source == "pos" && in.UserID == "" {
		return nil, errors.BadRequest("user id is required for pos orders")
	}

	status := in.Status
	if status == "" {
		status = "completed"
	}
	completedAt := in.CompletedAt
	if status == "completed" && completedAt.IsZero() {
		completedAt = time.Now().UTC()
	}

	gatewayItems, err := buildGatewayLineItems(in.LineItems)
	if err != nil {
		return nil, err
	}
	gatewayAdj, err := buildGatewayOrderAdjustments(in.Adjustments)
	if err != nil {
		return nil, err
	}

	// Children-win-on-aggregates: when the caller sends a per-tax breakdown or
	// adjustment list, we treat those as the source of truth and recompute the
	// aggregate fields the legacy desk POS used to send. This keeps the wire
	// contract backward compatible (aggregates still flow when nothing richer
	// is provided) without letting the two diverge silently when both arrive.
	subtotal, taxTotal, discountTotal, total, recErr := recomputeAggregates(in, gatewayItems, gatewayAdj)
	if recErr != nil {
		return nil, recErr
	}

	var out *entity.Order
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		// Idempotency: if a row already exists for (source, external_id), return
		// it without re-inserting. The desk client retries uploads on failure;
		// without this guard a flaky network would create duplicates.
		if in.ExternalID != "" {
			existing, gErr := u.reader.GetByExternalID(ctx, source, in.ExternalID)
			if gErr == nil {
				if hErr := u.hydrate(ctx, existing); hErr != nil {
					return hErr
				}
				out = existing
				return nil
			}
			if errors.GetCode(gErr) != errors.CodeNotFound {
				return gErr
			}
		}

		// Membership is enforced when the client explicitly chose a store. If the
		// client omitted store_id we fall back to the org's first store and skip
		// the check — that path exists for legacy queued uploads (pre-store-picker
		// rows in the desk pending_order_uploads queue) and for online/marketplace
		// orders that don't carry a user_id.
		clientSuppliedStoreID := in.StoreID != ""
		storeID := in.StoreID
		if storeID == "" {
			s, sErr := u.storeReader.FirstStoreID(ctx, in.OrgID)
			if sErr != nil {
				return sErr
			}
			storeID = s
		}

		if source == "pos" && in.UserID != "" && clientSuppliedStoreID {
			ok, accErr := u.memberReader.HasStoreAccess(ctx, in.UserID, storeID)
			if accErr != nil {
				return accErr
			}
			if !ok {
				return errors.Forbidden("user is not assigned to this store")
			}
		}

		params := gateway.CreateOrderParams{
			OrgID:            in.OrgID,
			Source:           source,
			ExternalID:       in.ExternalID,
			ExternalSourceID: in.ExternalSourceID,
			OrderNumber:      in.OrderNumber,
			StoreID:          storeID,
			RegisterID:       in.RegisterID,
			CustomerID:       in.CustomerID,
			UserID:           in.UserID,
			Status:           status,
			Subtotal:         subtotal,
			TaxTotal:         taxTotal,
			DiscountTotal:    discountTotal,
			Total:            total,
			Notes:            in.Notes,
			CompletedAt:      completedAt,
			LineItems:        gatewayItems,
			Payments:         toGatewayPayments(in.Payments),
			Adjustments:      gatewayAdj,
		}
		created, cErr := u.writer.Create(ctx, params)
		if cErr != nil {
			return cErr
		}
		if hErr := u.hydrate(ctx, created); hErr != nil {
			return hErr
		}
		out = created
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "create order")
	}
	return out, nil
}

func (u *orderUsecase) hydrate(ctx context.Context, o *entity.Order) error {
	items, err := u.reader.ListLineItems(ctx, o.ID)
	if err != nil {
		return err
	}
	payments, err := u.reader.ListPayments(ctx, o.ID)
	if err != nil {
		return err
	}
	adjustments, err := u.reader.ListOrderAdjustments(ctx, o.ID)
	if err != nil {
		return err
	}
	o.LineItems = items
	o.Payments = payments
	o.Adjustments = adjustments
	return nil
}

func buildGatewayLineItems(in []input.CreateOrderLineItemInput) ([]gateway.CreateOrderLineItemParams, error) {
	// Children must be all-or-nothing across lines: either every line carries
	// the per-tax / per-adjustment breakdown, or none does. A mixed request
	// would silently let one line's caller-supplied aggregates leak into the
	// recomputed totals -- guard against it explicitly.
	var withChildren, withoutChildren int
	for _, v := range in {
		if len(v.Taxes) > 0 || len(v.Adjustments) > 0 {
			withChildren++
		} else {
			withoutChildren++
		}
	}
	if withChildren > 0 && withoutChildren > 0 {
		return nil, errors.BadRequest("line item taxes and adjustments must be supplied for all lines or none")
	}

	out := make([]gateway.CreateOrderLineItemParams, len(in))
	for i, v := range in {
		taxes, err := toGatewayLineTaxes(v.Taxes)
		if err != nil {
			return nil, err
		}
		adjs, err := toGatewayLineAdjustments(v.Adjustments)
		if err != nil {
			return nil, err
		}
		discount, taxAmount, err := resolveLineAggregates(v, taxes, adjs)
		if err != nil {
			return nil, err
		}
		out[i] = gateway.CreateOrderLineItemParams{
			VariantID:      v.VariantID,
			ProductName:    v.ProductName,
			VariantName:    v.VariantName,
			SKU:            v.SKU,
			Quantity:       v.Quantity,
			UnitPrice:      v.UnitPrice,
			DiscountAmount: discount,
			TaxRate:        v.TaxRate,
			TaxAmount:      taxAmount,
			LineTotal:      v.LineTotal,
			Notes:          v.Notes,
			Taxes:          taxes,
			Adjustments:    adjs,
		}
	}
	return out, nil
}

func toGatewayLineTaxes(in []input.CreateOrderLineItemTaxInput) ([]gateway.OrderLineTaxParams, error) {
	if len(in) == 0 {
		return nil, nil
	}
	out := make([]gateway.OrderLineTaxParams, len(in))
	for i, t := range in {
		if t.NameSnapshot == "" {
			return nil, errors.BadRequest("line tax name_snapshot is required")
		}
		if _, err := money.Parse(t.RateSnapshot); err != nil {
			return nil, errors.BadRequest("invalid line tax rate snapshot")
		}
		if _, err := money.Parse(t.Amount); err != nil {
			return nil, errors.BadRequest("invalid line tax amount")
		}
		if _, err := money.Parse(t.TaxableBase); err != nil {
			return nil, errors.BadRequest("invalid line tax taxable_base")
		}
		out[i] = gateway.OrderLineTaxParams{
			Sequence:     t.Sequence,
			TaxRateID:    t.TaxRateID,
			NameSnapshot: t.NameSnapshot,
			RateSnapshot: t.RateSnapshot,
			IsInclusive:  t.IsInclusive,
			IsCompound:   t.IsCompound,
			TaxableBase:  t.TaxableBase,
			Amount:       t.Amount,
		}
	}
	return out, nil
}

func toGatewayLineAdjustments(in []input.CreateOrderLineAdjustmentInput) ([]gateway.OrderLineAdjustmentParams, error) {
	if len(in) == 0 {
		return nil, nil
	}
	out := make([]gateway.OrderLineAdjustmentParams, len(in))
	for i, a := range in {
		if a.Kind == "" || a.SourceType == "" || a.NameSnapshot == "" || a.CalculationType == "" {
			return nil, errors.BadRequest("line adjustment kind, source_type, name_snapshot, calculation_type are required")
		}
		if _, err := money.Parse(a.Amount); err != nil {
			return nil, errors.BadRequest("invalid line adjustment amount")
		}
		out[i] = gateway.OrderLineAdjustmentParams{
			Sequence:           a.Sequence,
			Kind:               a.Kind,
			SourceType:         a.SourceType,
			SourceID:           a.SourceID,
			SourceCodeSnapshot: a.SourceCodeSnapshot,
			NameSnapshot:       a.NameSnapshot,
			Reason:             a.Reason,
			CalculationType:    a.CalculationType,
			CalculationValue:   a.CalculationValue,
			Amount:             a.Amount,
			AppliesBeforeTax:   a.AppliesBeforeTax,
			AppliedBy:          a.AppliedBy,
			ApprovedBy:         a.ApprovedBy,
		}
	}
	return out, nil
}

func buildGatewayOrderAdjustments(in []input.CreateOrderAdjustmentInput) ([]gateway.OrderAdjustmentParams, error) {
	if len(in) == 0 {
		return nil, nil
	}
	out := make([]gateway.OrderAdjustmentParams, len(in))
	for i, a := range in {
		if a.Kind == "" || a.SourceType == "" || a.NameSnapshot == "" || a.CalculationType == "" {
			return nil, errors.BadRequest("order adjustment kind, source_type, name_snapshot, calculation_type are required")
		}
		if a.ProrateStrategy == "" {
			return nil, errors.BadRequest("order adjustment prorate_strategy is required")
		}
		if _, err := money.Parse(a.Amount); err != nil {
			return nil, errors.BadRequest("invalid order adjustment amount")
		}
		out[i] = gateway.OrderAdjustmentParams{
			Sequence:           a.Sequence,
			Kind:               a.Kind,
			SourceType:         a.SourceType,
			SourceID:           a.SourceID,
			SourceCodeSnapshot: a.SourceCodeSnapshot,
			NameSnapshot:       a.NameSnapshot,
			Reason:             a.Reason,
			CalculationType:    a.CalculationType,
			CalculationValue:   a.CalculationValue,
			Amount:             a.Amount,
			AppliesBeforeTax:   a.AppliesBeforeTax,
			ProrateStrategy:    a.ProrateStrategy,
			AppliedBy:          a.AppliedBy,
			ApprovedBy:         a.ApprovedBy,
		}
	}
	return out, nil
}

// resolveLineAggregates derives the line's discount_amount and tax_amount
// from taxes[] / adjustments[] when the caller supplied them. When neither is
// present the caller's pre-computed aggregates are kept (legacy desk POS path).
func resolveLineAggregates(
	v input.CreateOrderLineItemInput,
	taxes []gateway.OrderLineTaxParams,
	adjs []gateway.OrderLineAdjustmentParams,
) (discount string, tax string, err error) {
	discount = v.DiscountAmount
	tax = v.TaxAmount

	if len(adjs) > 0 {
		sum := money.Zero
		for _, a := range adjs {
			d, pErr := money.Parse(a.Amount)
			if pErr != nil {
				return "", "", errors.BadRequest("invalid line adjustment amount")
			}
			// Discounts/comps arrive as negative amounts; discount_amount is the
			// absolute reduction, so flip the sign before summing. Positive
			// amounts (fees, service charges) are not part of the discount total.
			if d.IsNegative() {
				sum = sum.Add(d.Neg())
			}
		}
		discount = money.String(sum)
	}

	if len(taxes) > 0 {
		sum := money.Zero
		for _, t := range taxes {
			d, pErr := money.Parse(t.Amount)
			if pErr != nil {
				return "", "", errors.BadRequest("invalid line tax amount")
			}
			sum = sum.Add(d)
		}
		tax = money.String(sum)
	}
	return discount, tax, nil
}

// recomputeAggregates derives order-level totals when the caller sent a
// breakdown. With nothing richer to draw on we return whatever was on the
// input — the legacy contract.
func recomputeAggregates(
	in input.CreateOrderInput,
	items []gateway.CreateOrderLineItemParams,
	orderAdj []gateway.OrderAdjustmentParams,
) (subtotal, taxTotal, discountTotal, total string, err error) {
	hasLineChildren := false
	for _, it := range items {
		if len(it.Taxes) > 0 || len(it.Adjustments) > 0 {
			hasLineChildren = true
			break
		}
	}
	if !hasLineChildren && len(orderAdj) == 0 {
		return in.Subtotal, in.TaxTotal, in.DiscountTotal, in.Total, nil
	}

	sub := money.Zero
	tax := money.Zero
	discount := money.Zero
	for _, it := range items {
		qty, pErr := money.Parse(it.Quantity)
		if pErr != nil {
			return "", "", "", "", errors.BadRequest("invalid line quantity")
		}
		unit, pErr := money.Parse(it.UnitPrice)
		if pErr != nil {
			return "", "", "", "", errors.BadRequest("invalid line unit price")
		}
		// sub uses gross unit_price * qty. For inclusive-tax lines this still
		// matches the merchant's printed price; the caller-supplied tax
		// breakdown is what carries the inclusive split, and the resolver in
		// pkg/tax owns that math when it produces the children. The recompute
		// path here trusts the snapshot it was given.
		sub = sub.Add(qty.Mul(unit))

		t, pErr := money.Parse(it.TaxAmount)
		if pErr != nil {
			return "", "", "", "", errors.BadRequest("invalid line tax amount")
		}
		tax = tax.Add(t)

		d, pErr := money.Parse(it.DiscountAmount)
		if pErr != nil {
			return "", "", "", "", errors.BadRequest("invalid line discount amount")
		}
		discount = discount.Add(d)
	}

	// Order-level adjustments fold into total in a single pass: positive
	// (tip, fee) grows the bill, negative (order discount, comp) shrinks it.
	// discount_total separately tracks the absolute reductions for reporting,
	// but it is *not* re-applied to total -- doing so would double-subtract
	// the order discount.
	totalDec := sub.Add(tax).Sub(discount)
	for _, a := range orderAdj {
		amt, pErr := money.Parse(a.Amount)
		if pErr != nil {
			return "", "", "", "", errors.BadRequest("invalid order adjustment amount")
		}
		totalDec = totalDec.Add(amt)
		if amt.IsNegative() {
			discount = discount.Add(amt.Neg())
		}
	}

	return money.String(sub), money.String(tax), money.String(discount), money.String(totalDec), nil
}

func toGatewayPayments(in []input.CreateOrderPaymentInput) []gateway.CreateOrderPaymentParams {
	out := make([]gateway.CreateOrderPaymentParams, len(in))
	for i, v := range in {
		out[i] = gateway.CreateOrderPaymentParams{
			PaymentMethodID: v.PaymentMethodID,
			Amount:          v.Amount,
			Tendered:        v.Tendered,
			ChangeAmount:    v.ChangeAmount,
			Reference:       v.Reference,
		}
	}
	return out
}

func (u *orderUsecase) ListOrders(ctx context.Context, in input.ListDailySalesInput) ([]*entity.OrderSummary, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if in.DateFrom.IsZero() || in.DateTo.IsZero() {
		return nil, errors.BadRequest("date range is required")
	}
	if !in.DateFrom.Before(in.DateTo) {
		return nil, errors.BadRequest("date_from must be before date_to")
	}
	var out []*entity.OrderSummary
	if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		v, err := u.reader.ListByDateRange(ctx, gateway.ListOrdersParams{
			DateFrom: in.DateFrom,
			DateTo:   in.DateTo,
			StoreID:  in.StoreID,
		})
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list orders")
	}
	return out, nil
}

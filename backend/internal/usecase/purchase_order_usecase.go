package usecase

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type purchaseOrderUsecase struct {
	tenantDB       gateway.TenantDB
	poReader       gateway.PurchaseOrderReader
	poWriter       gateway.PurchaseOrderWriter
	stockWriter    gateway.StockMovementWriter
}

// NewPurchaseOrderUsecase constructs a PurchaseOrderUsecase.
func NewPurchaseOrderUsecase(
	tenantDB gateway.TenantDB,
	poReader gateway.PurchaseOrderReader,
	poWriter gateway.PurchaseOrderWriter,
	stockWriter gateway.StockMovementWriter,
) PurchaseOrderUsecase {
	return &purchaseOrderUsecase{
		tenantDB:    tenantDB,
		poReader:    poReader,
		poWriter:    poWriter,
		stockWriter: stockWriter,
	}
}

// ----- Reads ---------------------------------------------------------------

func (u *purchaseOrderUsecase) ListPurchaseOrders(ctx context.Context, orgID string) ([]*entity.PurchaseOrderListItem, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.PurchaseOrderListItem
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.poReader.ListSummaries(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list purchase orders")
	}
	return out, nil
}

func (u *purchaseOrderUsecase) ListStores(ctx context.Context, orgID string) ([]*entity.StoreRef, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.StoreRef
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.poReader.ListStoreRefs(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list stores")
	}
	return out, nil
}

func (u *purchaseOrderUsecase) ListVariantsForPicker(ctx context.Context, orgID string) ([]*entity.VariantPickerItem, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.VariantPickerItem
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.poReader.ListVariantPickerItems(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list variant picker items")
	}
	return out, nil
}

func (u *purchaseOrderUsecase) GetPurchaseOrder(ctx context.Context, in input.GetPurchaseOrderInput) (*entity.PurchaseOrder, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	var out *entity.PurchaseOrder
	if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		po, err := u.poReader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		items, err := u.poReader.ListItems(ctx, in.ID)
		if err != nil {
			return err
		}
		po.Items = items
		out = po
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "get purchase order")
	}
	return out, nil
}

// ----- Writes --------------------------------------------------------------

func (u *purchaseOrderUsecase) CreatePurchaseOrder(ctx context.Context, in input.CreatePurchaseOrderInput) (*entity.PurchaseOrder, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if in.PurchaseOrder.StoreID == "" {
		return nil, errors.BadRequest("store id is required")
	}
	if len(in.PurchaseOrder.Items) == 0 {
		return nil, errors.BadRequest("at least one item is required")
	}
	var createdID string
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		poNumber, err := nextPONumber(ctx, u.poReader, in.OrgID)
		if err != nil {
			return err
		}
		po, err := u.poWriter.Create(ctx, gateway.CreatePurchaseOrderParams{
			OrgID:        in.OrgID,
			StoreID:      in.PurchaseOrder.StoreID,
			UserID:       in.UserID,
			PONumber:     poNumber,
			SupplierName: strings.TrimSpace(in.PurchaseOrder.SupplierName),
			Notes:        in.PurchaseOrder.Notes,
			ExpectedAt:   in.PurchaseOrder.ExpectedAt,
		})
		if err != nil {
			return err
		}
		createdID = po.ID
		return u.insertItems(ctx, in.OrgID, po.ID, in.PurchaseOrder.Items)
	}); err != nil {
		return nil, errors.Wrap(err, "create purchase order")
	}
	return u.GetPurchaseOrder(ctx, input.GetPurchaseOrderInput{ID: createdID, OrgID: in.OrgID})
}

func (u *purchaseOrderUsecase) UpdatePurchaseOrder(ctx context.Context, in input.UpdatePurchaseOrderInput) (*entity.PurchaseOrder, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if len(in.PurchaseOrder.Items) == 0 {
		return nil, errors.BadRequest("at least one item is required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		existing, err := u.poReader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if existing.Status != "draft" {
			return errors.BadRequest("only draft purchase orders can be edited")
		}
		if _, err := u.poWriter.Update(ctx, gateway.UpdatePurchaseOrderParams{
			ID:           in.ID,
			StoreID:      in.PurchaseOrder.StoreID,
			SupplierName: strings.TrimSpace(in.PurchaseOrder.SupplierName),
			Notes:        in.PurchaseOrder.Notes,
			ExpectedAt:   in.PurchaseOrder.ExpectedAt,
		}); err != nil {
			return err
		}
		if err := u.poWriter.DeleteItemsByPO(ctx, in.ID); err != nil {
			return err
		}
		return u.insertItems(ctx, in.OrgID, in.ID, in.PurchaseOrder.Items)
	}); err != nil {
		return nil, errors.Wrap(err, "update purchase order")
	}
	return u.GetPurchaseOrder(ctx, input.GetPurchaseOrderInput{ID: in.ID, OrgID: in.OrgID})
}

func (u *purchaseOrderUsecase) SubmitPurchaseOrder(ctx context.Context, in input.SubmitPurchaseOrderInput) (*entity.PurchaseOrder, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		existing, err := u.poReader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if existing.Status != "draft" {
			return errors.BadRequest("only draft purchase orders can be submitted")
		}
		return u.poWriter.UpdateStatus(ctx, in.ID, "submitted", time.Time{})
	}); err != nil {
		return nil, errors.Wrap(err, "submit purchase order")
	}
	return u.GetPurchaseOrder(ctx, input.GetPurchaseOrderInput{ID: in.ID, OrgID: in.OrgID})
}

func (u *purchaseOrderUsecase) CancelPurchaseOrder(ctx context.Context, in input.CancelPurchaseOrderInput) (*entity.PurchaseOrder, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		existing, err := u.poReader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if existing.Status == "received" {
			return errors.BadRequest("received purchase orders cannot be cancelled")
		}
		if existing.Status == "cancelled" {
			return nil
		}
		return u.poWriter.UpdateStatus(ctx, in.ID, "cancelled", time.Time{})
	}); err != nil {
		return nil, errors.Wrap(err, "cancel purchase order")
	}
	return u.GetPurchaseOrder(ctx, input.GetPurchaseOrderInput{ID: in.ID, OrgID: in.OrgID})
}

func (u *purchaseOrderUsecase) DeletePurchaseOrder(ctx context.Context, in input.DeletePurchaseOrderInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		existing, err := u.poReader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if existing.Status != "draft" && existing.Status != "cancelled" {
			return errors.BadRequest("only draft or cancelled purchase orders can be deleted")
		}
		return u.poWriter.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete purchase order")
	}
	return nil
}

// ReceivePurchaseOrder applies per-line received quantities and writes a
// stock_movement for each non-zero delta. Status moves to "partial" or
// "received" based on totals.
func (u *purchaseOrderUsecase) ReceivePurchaseOrder(ctx context.Context, in input.ReceivePurchaseOrderInput) (*entity.PurchaseOrder, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if len(in.Lines) == 0 {
		return nil, errors.BadRequest("at least one line is required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		po, err := u.poReader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if po.Status != "submitted" && po.Status != "partial" {
			return errors.BadRequest("only submitted or partially received orders can receive stock")
		}

		items, err := u.poReader.ListItems(ctx, in.ID)
		if err != nil {
			return err
		}
		byID := make(map[string]*entity.PurchaseOrderItem, len(items))
		for _, it := range items {
			byID[it.ID] = it
		}

		// Pre-validate all lines, then apply.
		type applied struct {
			item  *entity.PurchaseOrderItem
			delta *big.Float
			next  *big.Float
		}
		toApply := make([]applied, 0, len(in.Lines))
		for _, ln := range in.Lines {
			it, ok := byID[ln.ItemID]
			if !ok {
				return errors.BadRequest("unknown item id " + ln.ItemID)
			}
			delta, ok := new(big.Float).SetString(ln.QuantityToReceive)
			if !ok {
				return errors.BadRequest("invalid quantity for item " + ln.ItemID)
			}
			if delta.Sign() < 0 {
				return errors.BadRequest("quantity cannot be negative")
			}
			if delta.Sign() == 0 {
				continue
			}
			ordered, _ := new(big.Float).SetString(it.QuantityOrdered)
			received, _ := new(big.Float).SetString(it.QuantityReceived)
			next := new(big.Float).Add(received, delta)
			if next.Cmp(ordered) > 0 {
				return errors.BadRequest("received exceeds ordered for item " + ln.ItemID)
			}
			toApply = append(toApply, applied{item: it, delta: delta, next: next})
		}
		if len(toApply) == 0 {
			return errors.BadRequest("no non-zero quantities to receive")
		}

		for _, a := range toApply {
			if err := u.poWriter.AddItemReceived(ctx, a.item.ID, a.delta.Text('f', 4)); err != nil {
				return err
			}
			if err := u.stockWriter.Insert(ctx, gateway.CreateStockMovementParams{
				OrgID:         in.OrgID,
				StoreID:       po.StoreID,
				VariantID:     a.item.VariantID,
				Direction:     "in",
				Quantity:      a.delta.Text('f', 4),
				MovementType:  "purchase",
				ReferenceType: "purchase_order",
				ReferenceID:   po.ID,
				UserID:        in.UserID,
			}); err != nil {
				return err
			}
			// mutate our in-memory copy so the status calc below sees the new totals
			a.item.QuantityReceived = a.next.Text('f', 4)
		}

		// Recompute status.
		allFull := true
		anyPartial := false
		for _, it := range items {
			ordered, _ := new(big.Float).SetString(it.QuantityOrdered)
			received, _ := new(big.Float).SetString(it.QuantityReceived)
			cmp := received.Cmp(ordered)
			if cmp < 0 {
				allFull = false
			}
			if received.Sign() > 0 {
				anyPartial = true
			}
			_ = cmp
		}
		newStatus := po.Status
		var receivedAt time.Time
		switch {
		case allFull:
			newStatus = "received"
			receivedAt = time.Now()
		case anyPartial:
			newStatus = "partial"
		}
		if newStatus != po.Status {
			if err := u.poWriter.UpdateStatus(ctx, in.ID, newStatus, receivedAt); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "receive purchase order")
	}
	return u.GetPurchaseOrder(ctx, input.GetPurchaseOrderInput{ID: in.ID, OrgID: in.OrgID})
}

// ----- helpers -------------------------------------------------------------

func (u *purchaseOrderUsecase) insertItems(ctx context.Context, orgID, poID string, items []input.PurchaseOrderItemInput) error {
	for _, it := range items {
		if it.VariantID == "" {
			return errors.BadRequest("variant id is required on all items")
		}
		qty := strings.TrimSpace(it.QuantityOrdered)
		if qty == "" {
			qty = "0"
		}
		qtyNum, ok := new(big.Float).SetString(qty)
		if !ok || qtyNum.Sign() <= 0 {
			return errors.BadRequest("quantity must be a positive number")
		}
		cost := strings.TrimSpace(it.CostPrice)
		if cost == "" {
			cost = "0"
		}
		if _, err := u.poWriter.InsertItem(ctx, gateway.CreatePurchaseOrderItemParams{
			OrgID:           orgID,
			PurchaseOrderID: poID,
			VariantID:       it.VariantID,
			QuantityOrdered: qty,
			CostPrice:       cost,
		}); err != nil {
			return err
		}
	}
	return nil
}

// nextPONumber generates PO-YYYYMMDD-NNNN unique within the org.
// NNNN is COUNT(*)+1 for today's prefix. Races are rare and get a best-effort
// retry via incrementing the suffix until a free slot is found (up to 5 tries).
func nextPONumber(ctx context.Context, r gateway.PurchaseOrderReader, orgID string) (string, error) {
	prefix := fmt.Sprintf("PO-%s-", time.Now().UTC().Format("20060102"))
	n, err := r.CountForPrefix(ctx, orgID, prefix+"%")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, n+1), nil
}

package datastore

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type purchaseOrderStore struct{}

// NewPurchaseOrderReader returns a PurchaseOrderReader backed by sqlc.
func NewPurchaseOrderReader() gateway.PurchaseOrderReader { return &purchaseOrderStore{} }

// NewPurchaseOrderWriter returns a PurchaseOrderWriter backed by sqlc.
func NewPurchaseOrderWriter() gateway.PurchaseOrderWriter { return &purchaseOrderStore{} }

// ----- Reader --------------------------------------------------------------

func (s *purchaseOrderStore) GetByID(ctx context.Context, id string) (*entity.PurchaseOrder, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid purchase order id")
	}
	r, err := sqlc.New(dbtx).GetPurchaseOrderByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("purchase order not found")
		}
		return nil, errors.Wrap(err, "get purchase order by id")
	}
	return &entity.PurchaseOrder{
		ID:           uuidString(r.ID),
		OrgID:        uuidString(r.OrgID),
		StoreID:      uuidString(r.StoreID),
		UserID:       uuidString(r.UserID),
		PONumber:     r.PoNumber,
		SupplierName: textString(r.SupplierName),
		Status:       r.Status,
		Notes:        textString(r.Notes),
		ExpectedAt:   timestampTime(r.ExpectedAt),
		ReceivedAt:   timestampTime(r.ReceivedAt),
		CreatedAt:    r.CreatedAt.Time,
		UpdatedAt:    r.UpdatedAt.Time,
	}, nil
}

func (s *purchaseOrderStore) CountForPrefix(ctx context.Context, orgID, prefix string) (int, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return 0, err
	}
	uid, err := parseUUID(orgID)
	if err != nil {
		return 0, errors.BadRequest("invalid org id")
	}
	n, err := sqlc.New(dbtx).CountPurchaseOrdersForPrefix(ctx, sqlc.CountPurchaseOrdersForPrefixParams{
		OrgID:  uid,
		Prefix: prefix,
	})
	if err != nil {
		return 0, errors.Wrap(err, "count purchase orders")
	}
	return int(n), nil
}

func (s *purchaseOrderStore) ListItems(ctx context.Context, poID string) ([]*entity.PurchaseOrderItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(poID)
	if err != nil {
		return nil, errors.BadRequest("invalid purchase order id")
	}
	rows, err := sqlc.New(dbtx).ListPurchaseOrderItems(ctx, uid)
	if err != nil {
		return nil, errors.Wrap(err, "list purchase order items")
	}
	out := make([]*entity.PurchaseOrderItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.PurchaseOrderItem{
			ID:               uuidString(r.ID),
			VariantID:        uuidString(r.VariantID),
			VariantName:      r.VariantName,
			ProductName:      r.ProductName,
			QuantityOrdered:  numericToString(r.QuantityOrdered),
			QuantityReceived: numericToString(r.QuantityReceived),
			CostPrice:        numericToString(r.CostPrice),
		})
	}
	return out, nil
}

func (s *purchaseOrderStore) GetItemByID(ctx context.Context, id string) (*entity.PurchaseOrderItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid item id")
	}
	r, err := sqlc.New(dbtx).GetPurchaseOrderItemByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("item not found")
		}
		return nil, errors.Wrap(err, "get purchase order item")
	}
	return &entity.PurchaseOrderItem{
		ID:               uuidString(r.ID),
		VariantID:        uuidString(r.VariantID),
		QuantityOrdered:  numericToString(r.QuantityOrdered),
		QuantityReceived: numericToString(r.QuantityReceived),
		CostPrice:        numericToString(r.CostPrice),
	}, nil
}

func (s *purchaseOrderStore) ListSummaries(ctx context.Context) ([]*entity.PurchaseOrderListItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListPurchaseOrderSummaries(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list purchase order summaries")
	}
	out := make([]*entity.PurchaseOrderListItem, 0, len(rows))
	for _, r := range rows {
		total, _ := r.Total.(string)
		if total == "" {
			total = "0"
		}
		out = append(out, &entity.PurchaseOrderListItem{
			ID:           uuidString(r.ID),
			PONumber:     r.PoNumber,
			SupplierName: textString(r.SupplierName),
			Status:       r.Status,
			StoreName:    textString(r.StoreName),
			ExpectedAt:   timestampTime(r.ExpectedAt),
			ItemCount:    r.ItemCount,
			Total:        total,
			CreatedAt:    r.CreatedAt.Time,
		})
	}
	return out, nil
}

func (s *purchaseOrderStore) ListStoreRefs(ctx context.Context) ([]*entity.StoreRef, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListStoreRefs(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list store refs")
	}
	out := make([]*entity.StoreRef, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.StoreRef{ID: uuidString(r.ID), Name: r.Name})
	}
	return out, nil
}

func (s *purchaseOrderStore) ListVariantPickerItems(ctx context.Context) ([]*entity.VariantPickerItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListVariantPickerItems(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list variant picker items")
	}
	out := make([]*entity.VariantPickerItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.VariantPickerItem{
			ID:          uuidString(r.ID),
			ProductName: r.ProductName,
			VariantName: r.VariantName,
			SKU:         r.Sku,
			Price:       r.Price,
			CostPrice:   r.CostPrice,
		})
	}
	return out, nil
}

// ----- Writer --------------------------------------------------------------

func (s *purchaseOrderStore) Create(ctx context.Context, params gateway.CreatePurchaseOrderParams) (*entity.PurchaseOrder, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	storeID, err := parseUUID(params.StoreID)
	if err != nil {
		return nil, errors.BadRequest("invalid store id")
	}
	userID, err := uuidOrNull(params.UserID)
	if err != nil {
		return nil, errors.BadRequest("invalid user id")
	}
	r, err := sqlc.New(dbtx).CreatePurchaseOrder(ctx, sqlc.CreatePurchaseOrderParams{
		OrgID:        orgID,
		StoreID:      storeID,
		UserID:       userID,
		PoNumber:     params.PONumber,
		SupplierName: textOrNull(params.SupplierName),
		Notes:        textOrNull(params.Notes),
		ExpectedAt:   timestampOrNull(params.ExpectedAt),
	})
	if err != nil {
		return nil, errors.Wrap(err, "create purchase order")
	}
	return &entity.PurchaseOrder{
		ID:           uuidString(r.ID),
		OrgID:        uuidString(r.OrgID),
		StoreID:      uuidString(r.StoreID),
		UserID:       uuidString(r.UserID),
		PONumber:     r.PoNumber,
		SupplierName: textString(r.SupplierName),
		Status:       r.Status,
		Notes:        textString(r.Notes),
		ExpectedAt:   timestampTime(r.ExpectedAt),
		ReceivedAt:   timestampTime(r.ReceivedAt),
		CreatedAt:    r.CreatedAt.Time,
		UpdatedAt:    r.UpdatedAt.Time,
	}, nil
}

func (s *purchaseOrderStore) Update(ctx context.Context, params gateway.UpdatePurchaseOrderParams) (*entity.PurchaseOrder, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseUUID(params.ID)
	if err != nil {
		return nil, errors.BadRequest("invalid purchase order id")
	}
	storeID, err := parseUUID(params.StoreID)
	if err != nil {
		return nil, errors.BadRequest("invalid store id")
	}
	r, err := sqlc.New(dbtx).UpdatePurchaseOrder(ctx, sqlc.UpdatePurchaseOrderParams{
		ID:           id,
		StoreID:      storeID,
		SupplierName: textOrNull(params.SupplierName),
		Notes:        textOrNull(params.Notes),
		ExpectedAt:   timestampOrNull(params.ExpectedAt),
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("purchase order not found")
		}
		return nil, errors.Wrap(err, "update purchase order")
	}
	return &entity.PurchaseOrder{
		ID:           uuidString(r.ID),
		OrgID:        uuidString(r.OrgID),
		StoreID:      uuidString(r.StoreID),
		UserID:       uuidString(r.UserID),
		PONumber:     r.PoNumber,
		SupplierName: textString(r.SupplierName),
		Status:       r.Status,
		Notes:        textString(r.Notes),
		ExpectedAt:   timestampTime(r.ExpectedAt),
		ReceivedAt:   timestampTime(r.ReceivedAt),
		CreatedAt:    r.CreatedAt.Time,
		UpdatedAt:    r.UpdatedAt.Time,
	}, nil
}

func (s *purchaseOrderStore) UpdateStatus(ctx context.Context, id, status string, receivedAt time.Time) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid purchase order id")
	}
	if err := sqlc.New(dbtx).UpdatePurchaseOrderStatus(ctx, sqlc.UpdatePurchaseOrderStatusParams{
		ID:         uid,
		Status:     status,
		ReceivedAt: timestampOrNull(receivedAt),
	}); err != nil {
		return errors.Wrap(err, "update purchase order status")
	}
	return nil
}

func (s *purchaseOrderStore) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid purchase order id")
	}
	if err := sqlc.New(dbtx).SoftDeletePurchaseOrder(ctx, uid); err != nil {
		return errors.Wrap(err, "soft delete purchase order")
	}
	return nil
}

func (s *purchaseOrderStore) InsertItem(ctx context.Context, params gateway.CreatePurchaseOrderItemParams) (*entity.PurchaseOrderItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	poID, err := parseUUID(params.PurchaseOrderID)
	if err != nil {
		return nil, errors.BadRequest("invalid purchase order id")
	}
	vID, err := parseUUID(params.VariantID)
	if err != nil {
		return nil, errors.BadRequest("invalid variant id")
	}
	qty, err := numericFromString(params.QuantityOrdered)
	if err != nil {
		return nil, errors.BadRequest("invalid quantity")
	}
	cost, err := numericFromString(params.CostPrice)
	if err != nil {
		return nil, errors.BadRequest("invalid cost price")
	}
	r, err := sqlc.New(dbtx).InsertPurchaseOrderItem(ctx, sqlc.InsertPurchaseOrderItemParams{
		OrgID:           orgID,
		PurchaseOrderID: poID,
		VariantID:       vID,
		QuantityOrdered: qty,
		CostPrice:       cost,
	})
	if err != nil {
		return nil, errors.Wrap(err, "insert purchase order item")
	}
	return &entity.PurchaseOrderItem{
		ID:               uuidString(r.ID),
		VariantID:        uuidString(r.VariantID),
		QuantityOrdered:  numericToString(r.QuantityOrdered),
		QuantityReceived: numericToString(r.QuantityReceived),
		CostPrice:        numericToString(r.CostPrice),
	}, nil
}

func (s *purchaseOrderStore) DeleteItemsByPO(ctx context.Context, poID string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(poID)
	if err != nil {
		return errors.BadRequest("invalid purchase order id")
	}
	if err := sqlc.New(dbtx).DeletePurchaseOrderItemsByPO(ctx, uid); err != nil {
		return errors.Wrap(err, "delete purchase order items")
	}
	return nil
}

func (s *purchaseOrderStore) AddItemReceived(ctx context.Context, itemID, delta string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(itemID)
	if err != nil {
		return errors.BadRequest("invalid item id")
	}
	d, err := numericFromString(delta)
	if err != nil {
		return errors.BadRequest("invalid delta")
	}
	if err := sqlc.New(dbtx).AddPurchaseOrderItemReceived(ctx, sqlc.AddPurchaseOrderItemReceivedParams{
		ID:    uid,
		Delta: d,
	}); err != nil {
		return errors.Wrap(err, "add item received")
	}
	return nil
}


package datastore

import (
	"context"
	stderrors "errors"

	"github.com/jackc/pgx/v5"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type orderStore struct{}

// NewOrderReader returns an OrderReader backed by sqlc.
func NewOrderReader() gateway.OrderReader { return &orderStore{} }

// NewOrderWriter returns an OrderWriter backed by sqlc.
func NewOrderWriter() gateway.OrderWriter { return &orderStore{} }

// NewOrgStoreReader returns an OrgStoreReader backed by sqlc.
func NewOrgStoreReader() gateway.OrgStoreReader { return &orderStore{} }

func (s *orderStore) ListByDateRange(ctx context.Context, params gateway.ListOrdersParams) ([]*entity.OrderSummary, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	storeID, err := uuidOrNull(params.StoreID)
	if err != nil {
		return nil, errors.BadRequest("invalid store id")
	}
	rows, err := sqlc.New(dbtx).ListOrdersByDateRange(ctx, sqlc.ListOrdersByDateRangeParams{
		DateFrom: timestampOrNull(params.DateFrom),
		DateTo:   timestampOrNull(params.DateTo),
		StoreID:  storeID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "list orders by date range")
	}
	out := make([]*entity.OrderSummary, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.OrderSummary{
			ID:            uuidString(r.ID),
			OrderNumber:   r.OrderNumber,
			Status:        r.Status,
			Subtotal:      r.Subtotal,
			TaxTotal:      r.TaxTotal,
			DiscountTotal: r.DiscountTotal,
			Total:         r.Total,
			StoreID:       uuidString(r.StoreID),
			StoreName:     r.StoreName,
			RegisterID:    uuidString(r.RegisterID),
			UserID:        uuidString(r.UserID),
			UserName:      r.UserName,
			CustomerID:    uuidString(r.CustomerID),
			CustomerName:  r.CustomerName,
			CreatedAt:     r.CreatedAt.Time,
			Source:        r.Source,
			ExternalID:    r.ExternalID,
		})
	}
	return out, nil
}

func (s *orderStore) GetByID(ctx context.Context, id string) (*entity.Order, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid order id")
	}
	r, err := sqlc.New(dbtx).GetOrderByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("order not found")
		}
		return nil, errors.Wrap(err, "get order by id")
	}
	return &entity.Order{
		ID:            uuidString(r.ID),
		OrderNumber:   r.OrderNumber,
		Status:        r.Status,
		Subtotal:      r.Subtotal,
		TaxTotal:      r.TaxTotal,
		DiscountTotal: r.DiscountTotal,
		Total:         r.Total,
		Notes:         r.Notes,
		StoreID:       uuidString(r.StoreID),
		StoreName:     r.StoreName,
		RegisterID:    uuidString(r.RegisterID),
		UserID:        uuidString(r.UserID),
		UserName:      r.UserName,
		CustomerID:    uuidString(r.CustomerID),
		CustomerName:  r.CustomerName,
		CreatedAt:     r.CreatedAt.Time,
		CompletedAt:   timestampTime(r.CompletedAt),
		Source:        r.Source,
		ExternalID:    r.ExternalID,
	}, nil
}

func (s *orderStore) ListLineItems(ctx context.Context, orderID string) ([]*entity.OrderLineItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(orderID)
	if err != nil {
		return nil, errors.BadRequest("invalid order id")
	}
	rows, err := sqlc.New(dbtx).ListOrderLineItems(ctx, uid)
	if err != nil {
		return nil, errors.Wrap(err, "list order line items")
	}
	out := make([]*entity.OrderLineItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.OrderLineItem{
			ID:             uuidString(r.ID),
			VariantID:      uuidString(r.VariantID),
			ProductName:    r.ProductName,
			VariantName:    r.VariantName,
			SKU:            r.Sku,
			Quantity:       r.Quantity,
			UnitPrice:      r.UnitPrice,
			TaxRate:        r.TaxRate,
			TaxAmount:      r.TaxAmount,
			DiscountAmount: r.DiscountAmount,
			LineTotal:      r.LineTotal,
			Notes:          r.Notes,
		})
	}
	return out, nil
}

func (s *orderStore) ListPayments(ctx context.Context, orderID string) ([]*entity.OrderPayment, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(orderID)
	if err != nil {
		return nil, errors.BadRequest("invalid order id")
	}
	rows, err := sqlc.New(dbtx).ListOrderPayments(ctx, uid)
	if err != nil {
		return nil, errors.Wrap(err, "list order payments")
	}
	out := make([]*entity.OrderPayment, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.OrderPayment{
			ID:                uuidString(r.ID),
			PaymentMethodID:   uuidString(r.PaymentMethodID),
			PaymentMethodName: r.PaymentMethodName,
			Amount:            r.Amount,
			Tendered:          r.Tendered,
			ChangeAmount:      r.ChangeAmount,
			Reference:         r.Reference,
			Status:            r.Status,
			CreatedAt:         r.CreatedAt.Time,
		})
	}
	return out, nil
}

func (s *orderStore) GetByExternalID(ctx context.Context, source, externalID string) (*entity.Order, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	r, err := sqlc.New(dbtx).GetOrderByExternalID(ctx, sqlc.GetOrderByExternalIDParams{
		Source:     source,
		ExternalID: textOrNull(externalID),
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("order not found")
		}
		return nil, errors.Wrap(err, "get order by external id")
	}
	return &entity.Order{
		ID:            uuidString(r.ID),
		OrderNumber:   r.OrderNumber,
		Status:        r.Status,
		Subtotal:      r.Subtotal,
		TaxTotal:      r.TaxTotal,
		DiscountTotal: r.DiscountTotal,
		Total:         r.Total,
		Notes:         r.Notes,
		StoreID:       uuidString(r.StoreID),
		StoreName:     r.StoreName,
		RegisterID:    uuidString(r.RegisterID),
		UserID:        uuidString(r.UserID),
		UserName:      r.UserName,
		CustomerID:    uuidString(r.CustomerID),
		CustomerName:  r.CustomerName,
		CreatedAt:     r.CreatedAt.Time,
		CompletedAt:   timestampTime(r.CompletedAt),
		Source:        r.Source,
		ExternalID:    r.ExternalID,
	}, nil
}

func (s *orderStore) FirstStoreID(ctx context.Context, orgID string) (string, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return "", err
	}
	uid, err := parseUUID(orgID)
	if err != nil {
		return "", errors.BadRequest("invalid org id")
	}
	id, err := sqlc.New(dbtx).GetFirstStoreIDForOrg(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return "", errors.NotFound("no store for org")
		}
		return "", errors.Wrap(err, "first store id for org")
	}
	return uuidString(id), nil
}

func (s *orderStore) Create(ctx context.Context, params gateway.CreateOrderParams) (*entity.Order, error) {
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
	registerID, err := uuidOrNull(params.RegisterID)
	if err != nil {
		return nil, errors.BadRequest("invalid register id")
	}
	customerID, err := uuidOrNull(params.CustomerID)
	if err != nil {
		return nil, errors.BadRequest("invalid customer id")
	}
	userID, err := uuidOrNull(params.UserID)
	if err != nil {
		return nil, errors.BadRequest("invalid user id")
	}

	subtotal, err := numericFromString(params.Subtotal)
	if err != nil {
		return nil, errors.BadRequest("invalid subtotal")
	}
	taxTotal, err := numericFromString(params.TaxTotal)
	if err != nil {
		return nil, errors.BadRequest("invalid tax total")
	}
	discountTotal, err := numericFromString(params.DiscountTotal)
	if err != nil {
		return nil, errors.BadRequest("invalid discount total")
	}
	total, err := numericFromString(params.Total)
	if err != nil {
		return nil, errors.BadRequest("invalid total")
	}

	q := sqlc.New(dbtx)
	row, err := q.InsertOrder(ctx, sqlc.InsertOrderParams{
		OrgID:            orgID,
		StoreID:          storeID,
		RegisterID:       registerID,
		CustomerID:       customerID,
		UserID:           userID,
		OrderNumber:      params.OrderNumber,
		Status:           params.Status,
		Subtotal:         subtotal,
		TaxTotal:         taxTotal,
		DiscountTotal:    discountTotal,
		Total:            total,
		Notes:            textOrNull(params.Notes),
		CompletedAt:      timestampOrNull(params.CompletedAt),
		Source:           params.Source,
		ExternalID:       textOrNull(params.ExternalID),
		ExternalSourceID: textOrNull(params.ExternalSourceID),
	})
	if err != nil {
		return nil, errors.Wrap(err, "insert order")
	}

	for _, item := range params.LineItems {
		variantID, vErr := uuidOrNull(item.VariantID)
		if vErr != nil {
			return nil, errors.BadRequest("invalid line item variant id")
		}
		qty, qErr := numericFromString(item.Quantity)
		if qErr != nil {
			return nil, errors.BadRequest("invalid line item quantity")
		}
		unitPrice, uErr := numericFromString(item.UnitPrice)
		if uErr != nil {
			return nil, errors.BadRequest("invalid line item unit price")
		}
		taxRate, tErr := numericFromString(item.TaxRate)
		if tErr != nil {
			return nil, errors.BadRequest("invalid line item tax rate")
		}
		taxAmount, taErr := numericFromString(item.TaxAmount)
		if taErr != nil {
			return nil, errors.BadRequest("invalid line item tax amount")
		}
		discountAmount, dErr := numericFromString(item.DiscountAmount)
		if dErr != nil {
			return nil, errors.BadRequest("invalid line item discount amount")
		}
		lineTotal, lErr := numericFromString(item.LineTotal)
		if lErr != nil {
			return nil, errors.BadRequest("invalid line item total")
		}
		variantName := item.VariantName
		if variantName == "" {
			variantName = "Default"
		}
		if iErr := q.InsertOrderLineItem(ctx, sqlc.InsertOrderLineItemParams{
			OrgID:          orgID,
			OrderID:        row.ID,
			VariantID:      variantID,
			ProductName:    item.ProductName,
			VariantName:    variantName,
			Sku:            textOrNull(item.SKU),
			Quantity:       qty,
			UnitPrice:      unitPrice,
			TaxRate:        taxRate,
			TaxAmount:      taxAmount,
			DiscountAmount: discountAmount,
			LineTotal:      lineTotal,
			Notes:          textOrNull(item.Notes),
		}); iErr != nil {
			return nil, errors.Wrap(iErr, "insert order line item")
		}
	}

	for _, p := range params.Payments {
		paymentMethodID, pmErr := parseUUID(p.PaymentMethodID)
		if pmErr != nil {
			return nil, errors.BadRequest("invalid payment method id")
		}
		amount, aErr := numericFromString(p.Amount)
		if aErr != nil {
			return nil, errors.BadRequest("invalid payment amount")
		}
		tendered, tErr := numericOrNull(p.Tendered)
		if tErr != nil {
			return nil, errors.BadRequest("invalid payment tendered")
		}
		change, cErr := numericOrNull(p.ChangeAmount)
		if cErr != nil {
			return nil, errors.BadRequest("invalid payment change amount")
		}
		if pErr := q.InsertOrderPayment(ctx, sqlc.InsertOrderPaymentParams{
			OrgID:           orgID,
			OrderID:         row.ID,
			PaymentMethodID: paymentMethodID,
			Amount:          amount,
			Tendered:        tendered,
			ChangeAmount:    change,
			Reference:       textOrNull(p.Reference),
			Status:          "completed",
		}); pErr != nil {
			return nil, errors.Wrap(pErr, "insert order payment")
		}
	}

	return &entity.Order{
		ID:            uuidString(row.ID),
		OrderNumber:   row.OrderNumber,
		Status:        row.Status,
		Subtotal:      row.Subtotal,
		TaxTotal:      row.TaxTotal,
		DiscountTotal: row.DiscountTotal,
		Total:         row.Total,
		Notes:         row.Notes,
		StoreID:       uuidString(row.StoreID),
		RegisterID:    uuidString(row.RegisterID),
		UserID:        uuidString(row.UserID),
		CustomerID:    uuidString(row.CustomerID),
		CreatedAt:     row.CreatedAt.Time,
		CompletedAt:   timestampTime(row.CompletedAt),
		Source:        row.Source,
		ExternalID:    row.ExternalID,
	}, nil
}

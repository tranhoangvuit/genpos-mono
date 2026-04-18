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

type stockTakeStore struct{}

// NewStockTakeReader returns a StockTakeReader backed by sqlc.
func NewStockTakeReader() gateway.StockTakeReader { return &stockTakeStore{} }

// NewStockTakeWriter returns a StockTakeWriter backed by sqlc.
func NewStockTakeWriter() gateway.StockTakeWriter { return &stockTakeStore{} }

func (s *stockTakeStore) ListSummaries(ctx context.Context) ([]*entity.StockTakeListItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListStockTakeSummaries(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list stock take summaries")
	}
	out := make([]*entity.StockTakeListItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.StockTakeListItem{
			ID:            uuidString(r.ID),
			StoreName:     textString(r.StoreName),
			Status:        r.Status,
			ItemCount:     r.ItemCount,
			VarianceLines: r.VarianceLines,
			CreatedAt:     r.CreatedAt.Time,
			CompletedAt:   timestampTime(r.CompletedAt),
		})
	}
	return out, nil
}

func (s *stockTakeStore) GetByID(ctx context.Context, id string) (*entity.StockTake, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid stock take id")
	}
	r, err := sqlc.New(dbtx).GetStockTakeByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("stock take not found")
		}
		return nil, errors.Wrap(err, "get stock take by id")
	}
	return &entity.StockTake{
		ID:          uuidString(r.ID),
		OrgID:       uuidString(r.OrgID),
		StoreID:     uuidString(r.StoreID),
		UserID:      uuidString(r.UserID),
		Status:      r.Status,
		Notes:       textString(r.Notes),
		CompletedAt: timestampTime(r.CompletedAt),
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
	}, nil
}

func (s *stockTakeStore) ListItems(ctx context.Context, stockTakeID string) ([]*entity.StockTakeItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(stockTakeID)
	if err != nil {
		return nil, errors.BadRequest("invalid stock take id")
	}
	rows, err := sqlc.New(dbtx).ListStockTakeItems(ctx, uid)
	if err != nil {
		return nil, errors.Wrap(err, "list stock take items")
	}
	out := make([]*entity.StockTakeItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.StockTakeItem{
			ID:          uuidString(r.ID),
			VariantID:   uuidString(r.VariantID),
			VariantName: r.VariantName,
			ProductName: r.ProductName,
			ExpectedQty: numericToString(r.ExpectedQty),
			CountedQty:  numericToString(r.CountedQty),
		})
	}
	return out, nil
}

func (s *stockTakeStore) Create(ctx context.Context, params gateway.CreateStockTakeParams) (*entity.StockTake, error) {
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
	r, err := sqlc.New(dbtx).CreateStockTake(ctx, sqlc.CreateStockTakeParams{
		OrgID:   orgID,
		StoreID: storeID,
		UserID:  userID,
		Notes:   textOrNull(params.Notes),
	})
	if err != nil {
		return nil, errors.Wrap(err, "create stock take")
	}
	return &entity.StockTake{
		ID:          uuidString(r.ID),
		OrgID:       uuidString(r.OrgID),
		StoreID:     uuidString(r.StoreID),
		UserID:      uuidString(r.UserID),
		Status:      r.Status,
		Notes:       textString(r.Notes),
		CompletedAt: timestampTime(r.CompletedAt),
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
	}, nil
}

func (s *stockTakeStore) SeedItemsFromOnHand(ctx context.Context, orgID, stockTakeID, storeID string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	orgUID, err := parseUUID(orgID)
	if err != nil {
		return errors.BadRequest("invalid org id")
	}
	takeUID, err := parseUUID(stockTakeID)
	if err != nil {
		return errors.BadRequest("invalid stock take id")
	}
	storeUID, err := parseUUID(storeID)
	if err != nil {
		return errors.BadRequest("invalid store id")
	}
	if err := sqlc.New(dbtx).SeedStockTakeItemsFromOnHand(ctx, sqlc.SeedStockTakeItemsFromOnHandParams{
		OrgID:       orgUID,
		StockTakeID: takeUID,
		StoreID:     storeUID,
	}); err != nil {
		return errors.Wrap(err, "seed stock take items")
	}
	return nil
}

func (s *stockTakeStore) InsertItem(ctx context.Context, params gateway.CreateStockTakeItemParams) (*entity.StockTakeItem, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	takeID, err := parseUUID(params.StockTakeID)
	if err != nil {
		return nil, errors.BadRequest("invalid stock take id")
	}
	vID, err := parseUUID(params.VariantID)
	if err != nil {
		return nil, errors.BadRequest("invalid variant id")
	}
	expected, err := numericFromString(params.ExpectedQty)
	if err != nil {
		return nil, errors.BadRequest("invalid expected qty")
	}
	counted, err := numericFromString(params.CountedQty)
	if err != nil {
		return nil, errors.BadRequest("invalid counted qty")
	}
	r, err := sqlc.New(dbtx).InsertStockTakeItem(ctx, sqlc.InsertStockTakeItemParams{
		OrgID:       orgID,
		StockTakeID: takeID,
		VariantID:   vID,
		ExpectedQty: expected,
		CountedQty:  counted,
	})
	if err != nil {
		return nil, errors.Wrap(err, "insert stock take item")
	}
	return &entity.StockTakeItem{
		ID:          uuidString(r.ID),
		VariantID:   uuidString(r.VariantID),
		ExpectedQty: numericToString(r.ExpectedQty),
		CountedQty:  numericToString(r.CountedQty),
	}, nil
}

func (s *stockTakeStore) UpdateItemCount(ctx context.Context, itemID, countedQty string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(itemID)
	if err != nil {
		return errors.BadRequest("invalid item id")
	}
	counted, err := numericFromString(countedQty)
	if err != nil {
		return errors.BadRequest("invalid counted qty")
	}
	if err := sqlc.New(dbtx).UpdateStockTakeItemCount(ctx, sqlc.UpdateStockTakeItemCountParams{
		ID:         uid,
		CountedQty: counted,
	}); err != nil {
		return errors.Wrap(err, "update stock take item count")
	}
	return nil
}

func (s *stockTakeStore) UpdateNotes(ctx context.Context, id, notes string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid stock take id")
	}
	if err := sqlc.New(dbtx).UpdateStockTakeNotes(ctx, sqlc.UpdateStockTakeNotesParams{
		ID:    uid,
		Notes: textOrNull(notes),
	}); err != nil {
		return errors.Wrap(err, "update stock take notes")
	}
	return nil
}

func (s *stockTakeStore) UpdateStatus(ctx context.Context, id, status string, completedAt time.Time) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid stock take id")
	}
	if err := sqlc.New(dbtx).UpdateStockTakeStatus(ctx, sqlc.UpdateStockTakeStatusParams{
		ID:          uid,
		Status:      status,
		CompletedAt: timestampOrNull(completedAt),
	}); err != nil {
		return errors.Wrap(err, "update stock take status")
	}
	return nil
}

func (s *stockTakeStore) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid stock take id")
	}
	if err := sqlc.New(dbtx).SoftDeleteStockTake(ctx, uid); err != nil {
		return errors.Wrap(err, "soft delete stock take")
	}
	return nil
}

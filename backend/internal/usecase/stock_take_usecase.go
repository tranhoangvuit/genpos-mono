package usecase

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type stockTakeUsecase struct {
	tenantDB    gateway.TenantDB
	reader      gateway.StockTakeReader
	writer      gateway.StockTakeWriter
	stockWriter gateway.StockMovementWriter
}

// NewStockTakeUsecase constructs a StockTakeUsecase.
func NewStockTakeUsecase(
	tenantDB gateway.TenantDB,
	reader gateway.StockTakeReader,
	writer gateway.StockTakeWriter,
	stockWriter gateway.StockMovementWriter,
) StockTakeUsecase {
	return &stockTakeUsecase{
		tenantDB:    tenantDB,
		reader:      reader,
		writer:      writer,
		stockWriter: stockWriter,
	}
}

func (u *stockTakeUsecase) ListStockTakes(ctx context.Context, orgID string) ([]*entity.StockTakeListItem, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.StockTakeListItem
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.reader.ListSummaries(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list stock takes")
	}
	return out, nil
}

func (u *stockTakeUsecase) GetStockTake(ctx context.Context, in input.GetStockTakeInput) (*entity.StockTake, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	var out *entity.StockTake
	if err := u.tenantDB.ReadWithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		st, err := u.reader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		items, err := u.reader.ListItems(ctx, in.ID)
		if err != nil {
			return err
		}
		st.Items = items
		out = st
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "get stock take")
	}
	return out, nil
}

func (u *stockTakeUsecase) CreateStockTake(ctx context.Context, in input.CreateStockTakeInput) (*entity.StockTake, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if in.StoreID == "" {
		return nil, errors.BadRequest("store id is required")
	}
	var createdID string
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		st, err := u.writer.Create(ctx, gateway.CreateStockTakeParams{
			OrgID:   in.OrgID,
			StoreID: in.StoreID,
			UserID:  in.UserID,
			Notes:   in.Notes,
		})
		if err != nil {
			return err
		}
		createdID = st.ID
		return u.writer.SeedItemsFromOnHand(ctx, in.OrgID, st.ID, in.StoreID)
	}); err != nil {
		return nil, errors.Wrap(err, "create stock take")
	}
	return u.GetStockTake(ctx, input.GetStockTakeInput{ID: createdID, OrgID: in.OrgID})
}

func (u *stockTakeUsecase) SaveStockTakeProgress(ctx context.Context, in input.SaveStockTakeProgressInput) (*entity.StockTake, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		existing, err := u.reader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if existing.Status != "in_progress" {
			return errors.BadRequest("only in-progress stock takes can be updated")
		}
		if err := u.writer.UpdateNotes(ctx, in.ID, in.Notes); err != nil {
			return err
		}
		for _, ln := range in.Lines {
			if ln.ItemID == "" {
				continue
			}
			qty := strings.TrimSpace(ln.CountedQty)
			if qty == "" {
				qty = "0"
			}
			if _, ok := new(big.Float).SetString(qty); !ok {
				return errors.BadRequest("invalid counted qty for item " + ln.ItemID)
			}
			if err := u.writer.UpdateItemCount(ctx, ln.ItemID, qty); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "save stock take progress")
	}
	return u.GetStockTake(ctx, input.GetStockTakeInput{ID: in.ID, OrgID: in.OrgID})
}

// FinalizeStockTake writes one stock_movements entry per non-zero variance
// and marks the take as completed. Variance = counted_qty - expected_qty.
func (u *stockTakeUsecase) FinalizeStockTake(ctx context.Context, in input.FinalizeStockTakeInput) (*entity.StockTake, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		existing, err := u.reader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if existing.Status != "in_progress" {
			return errors.BadRequest("only in-progress stock takes can be finalized")
		}
		items, err := u.reader.ListItems(ctx, in.ID)
		if err != nil {
			return err
		}
		for _, it := range items {
			expected, _ := new(big.Float).SetString(it.ExpectedQty)
			counted, _ := new(big.Float).SetString(it.CountedQty)
			if expected == nil || counted == nil {
				continue
			}
			variance := new(big.Float).Sub(counted, expected)
			if variance.Sign() == 0 {
				continue
			}
			direction := "in"
			qty := variance
			if variance.Sign() < 0 {
				direction = "out"
				qty = new(big.Float).Neg(variance)
			}
			if err := u.stockWriter.Insert(ctx, gateway.CreateStockMovementParams{
				OrgID:         in.OrgID,
				StoreID:       existing.StoreID,
				VariantID:     it.VariantID,
				Direction:     direction,
				Quantity:      qty.Text('f', 4),
				MovementType:  "stock_take",
				ReferenceType: "stock_take",
				ReferenceID:   existing.ID,
				UserID:        in.UserID,
			}); err != nil {
				return err
			}
		}
		return u.writer.UpdateStatus(ctx, in.ID, "completed", time.Now())
	}); err != nil {
		return nil, errors.Wrap(err, "finalize stock take")
	}
	return u.GetStockTake(ctx, input.GetStockTakeInput{ID: in.ID, OrgID: in.OrgID})
}

func (u *stockTakeUsecase) CancelStockTake(ctx context.Context, in input.CancelStockTakeInput) (*entity.StockTake, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		existing, err := u.reader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if existing.Status == "completed" {
			return errors.BadRequest("completed stock takes cannot be cancelled")
		}
		if existing.Status == "cancelled" {
			return nil
		}
		return u.writer.UpdateStatus(ctx, in.ID, "cancelled", time.Time{})
	}); err != nil {
		return nil, errors.Wrap(err, "cancel stock take")
	}
	return u.GetStockTake(ctx, input.GetStockTakeInput{ID: in.ID, OrgID: in.OrgID})
}

func (u *stockTakeUsecase) DeleteStockTake(ctx context.Context, in input.DeleteStockTakeInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		existing, err := u.reader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		if existing.Status == "completed" {
			return errors.BadRequest("completed stock takes cannot be deleted")
		}
		return u.writer.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete stock take")
	}
	return nil
}

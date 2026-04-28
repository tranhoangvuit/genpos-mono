package datastore

import (
	"context"
	stderrors "errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type taxClassStore struct{}

func NewTaxClassReader() gateway.TaxClassReader { return &taxClassStore{} }
func NewTaxClassWriter() gateway.TaxClassWriter { return &taxClassStore{} }

func (s *taxClassStore) List(ctx context.Context) ([]*entity.TaxClass, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	q := sqlc.New(dbtx)
	rows, err := q.ListTaxClasses(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list tax classes")
	}
	if len(rows) == 0 {
		return []*entity.TaxClass{}, nil
	}

	classIDs := make([]pgtype.UUID, 0, len(rows))
	for _, r := range rows {
		classIDs = append(classIDs, r.ID)
	}
	rateRows, err := q.ListTaxClassRatesByClassIDs(ctx, classIDs)
	if err != nil {
		return nil, errors.Wrap(err, "list tax class rates")
	}

	byClass := map[string][]*entity.TaxClassRate{}
	for _, rr := range rateRows {
		classID := uuidString(rr.TaxClassID)
		byClass[classID] = append(byClass[classID], &entity.TaxClassRate{
			ID:         uuidString(rr.ID),
			TaxRateID:  uuidString(rr.TaxRateID),
			Sequence:   rr.Sequence,
			IsCompound: rr.IsCompound,
		})
	}

	out := make([]*entity.TaxClass, 0, len(rows))
	for _, r := range rows {
		c := &entity.TaxClass{
			ID:          uuidString(r.ID),
			OrgID:       uuidString(r.OrgID),
			Name:        r.Name,
			Description: textString(r.Description),
			IsDefault:   r.IsDefault,
			SortOrder:   r.SortOrder,
			CreatedAt:   r.CreatedAt.Time,
			UpdatedAt:   r.UpdatedAt.Time,
			Rates:       byClass[uuidString(r.ID)],
		}
		out = append(out, c)
	}
	return out, nil
}

func (s *taxClassStore) Get(ctx context.Context, id string) (*entity.TaxClass, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid tax class id")
	}
	q := sqlc.New(dbtx)
	r, err := q.GetTaxClass(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("tax class not found")
		}
		return nil, errors.Wrap(err, "get tax class")
	}
	rateRows, err := q.ListTaxClassRatesByClassIDs(ctx, []pgtype.UUID{r.ID})
	if err != nil {
		return nil, errors.Wrap(err, "list tax class rates")
	}
	rates := make([]*entity.TaxClassRate, 0, len(rateRows))
	for _, rr := range rateRows {
		rates = append(rates, &entity.TaxClassRate{
			ID:         uuidString(rr.ID),
			TaxRateID:  uuidString(rr.TaxRateID),
			Sequence:   rr.Sequence,
			IsCompound: rr.IsCompound,
		})
	}
	return &entity.TaxClass{
		ID:          uuidString(r.ID),
		OrgID:       uuidString(r.OrgID),
		Name:        r.Name,
		Description: textString(r.Description),
		IsDefault:   r.IsDefault,
		SortOrder:   r.SortOrder,
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
		Rates:       rates,
	}, nil
}

func (s *taxClassStore) Create(ctx context.Context, params gateway.CreateTaxClassParams) (*entity.TaxClass, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	q := sqlc.New(dbtx)
	row, err := q.CreateTaxClass(ctx, sqlc.CreateTaxClassParams{
		OrgID:       orgID,
		Name:        params.Name,
		Description: textOrNull(params.Description),
		IsDefault:   params.IsDefault,
		SortOrder:   params.SortOrder,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create tax class")
	}

	rates, err := s.insertRates(ctx, q, params.OrgID, uuidString(row.ID), params.Rates)
	if err != nil {
		return nil, err
	}
	return &entity.TaxClass{
		ID:          uuidString(row.ID),
		OrgID:       uuidString(row.OrgID),
		Name:        row.Name,
		Description: textString(row.Description),
		IsDefault:   row.IsDefault,
		SortOrder:   row.SortOrder,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		Rates:       rates,
	}, nil
}

func (s *taxClassStore) Update(ctx context.Context, params gateway.UpdateTaxClassParams) (*entity.TaxClass, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseUUID(params.ID)
	if err != nil {
		return nil, errors.BadRequest("invalid tax class id")
	}
	q := sqlc.New(dbtx)
	row, err := q.UpdateTaxClass(ctx, sqlc.UpdateTaxClassParams{
		ID:          id,
		Name:        params.Name,
		Description: textOrNull(params.Description),
		IsDefault:   params.IsDefault,
		SortOrder:   params.SortOrder,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("tax class not found")
		}
		return nil, errors.Wrap(err, "update tax class")
	}

	if err := q.SoftDeleteTaxClassRatesByClassID(ctx, id); err != nil {
		return nil, errors.Wrap(err, "soft delete tax class rates")
	}
	rates, err := s.insertRates(ctx, q, params.OrgID, uuidString(row.ID), params.Rates)
	if err != nil {
		return nil, err
	}
	return &entity.TaxClass{
		ID:          uuidString(row.ID),
		OrgID:       uuidString(row.OrgID),
		Name:        row.Name,
		Description: textString(row.Description),
		IsDefault:   row.IsDefault,
		SortOrder:   row.SortOrder,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		Rates:       rates,
	}, nil
}

func (s *taxClassStore) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid tax class id")
	}
	q := sqlc.New(dbtx)
	// Soft-delete the rates first so the sync rule stops projecting them
	// the moment the class disappears from clients.
	if err := q.SoftDeleteTaxClassRatesByClassID(ctx, uid); err != nil {
		return errors.Wrap(err, "soft delete tax class rates")
	}
	n, err := q.SoftDeleteTaxClass(ctx, uid)
	if err != nil {
		return errors.Wrap(err, "soft delete tax class")
	}
	if n == 0 {
		return errors.NotFound("tax class not found")
	}
	return nil
}

func (s *taxClassStore) ClearDefaults(ctx context.Context) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	if err := sqlc.New(dbtx).ClearDefaultTaxClasses(ctx); err != nil {
		return errors.Wrap(err, "clear default tax classes")
	}
	return nil
}

func (s *taxClassStore) ReplaceRates(ctx context.Context, orgID, classID string, rates []gateway.TaxClassRateParams) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	cid, err := parseUUID(classID)
	if err != nil {
		return errors.BadRequest("invalid tax class id")
	}
	q := sqlc.New(dbtx)
	if err := q.SoftDeleteTaxClassRatesByClassID(ctx, cid); err != nil {
		return errors.Wrap(err, "soft delete tax class rates")
	}
	if _, err := s.insertRates(ctx, q, orgID, classID, rates); err != nil {
		return err
	}
	return nil
}

// insertRates is shared between Create / Update / ReplaceRates. Returns the
// freshly inserted entities so callers can build the response without an
// extra read.
func (s *taxClassStore) insertRates(
	ctx context.Context, q *sqlc.Queries, orgIDStr, classIDStr string, rates []gateway.TaxClassRateParams,
) ([]*entity.TaxClassRate, error) {
	if len(rates) == 0 {
		return []*entity.TaxClassRate{}, nil
	}
	orgID, err := parseUUID(orgIDStr)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	classID, err := parseUUID(classIDStr)
	if err != nil {
		return nil, errors.BadRequest("invalid tax class id")
	}
	out := make([]*entity.TaxClassRate, 0, len(rates))
	for _, r := range rates {
		rateID, rErr := parseUUID(r.TaxRateID)
		if rErr != nil {
			return nil, errors.BadRequest("invalid tax rate id")
		}
		row, iErr := q.InsertTaxClassRate(ctx, sqlc.InsertTaxClassRateParams{
			OrgID:       orgID,
			TaxClassID:  classID,
			TaxRateID:   rateID,
			Sequence:    r.Sequence,
			IsCompound:  r.IsCompound,
		})
		if iErr != nil {
			return nil, errors.Wrap(iErr, "insert tax class rate")
		}
		out = append(out, &entity.TaxClassRate{
			ID:         uuidString(row.ID),
			TaxRateID:  uuidString(row.TaxRateID),
			Sequence:   row.Sequence,
			IsCompound: row.IsCompound,
		})
	}
	return out, nil
}

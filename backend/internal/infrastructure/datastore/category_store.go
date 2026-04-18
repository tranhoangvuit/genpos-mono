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

type categoryStore struct{}

// NewCategoryReader returns a CategoryReader backed by sqlc.
func NewCategoryReader() gateway.CategoryReader { return &categoryStore{} }

// NewCategoryWriter returns a CategoryWriter backed by sqlc.
func NewCategoryWriter() gateway.CategoryWriter { return &categoryStore{} }

func (s *categoryStore) List(ctx context.Context) ([]*entity.Category, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := sqlc.New(dbtx).ListCategories(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list categories")
	}
	out := make([]*entity.Category, 0, len(rows))
	for _, r := range rows {
		out = append(out, toCategoryEntity(r.ID, r.OrgID, r.ParentID, r.Name, r.SortOrder, r.Color, r.ImageUrl, r.CreatedAt, r.UpdatedAt))
	}
	return out, nil
}

func (s *categoryStore) GetByID(ctx context.Context, id string) (*entity.Category, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid category id")
	}
	r, err := sqlc.New(dbtx).GetCategoryByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("category not found")
		}
		return nil, errors.Wrap(err, "get category by id")
	}
	return toCategoryEntity(r.ID, r.OrgID, r.ParentID, r.Name, r.SortOrder, r.Color, r.ImageUrl, r.CreatedAt, r.UpdatedAt), nil
}

func (s *categoryStore) GetByName(ctx context.Context, name string) (*entity.Category, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	r, err := sqlc.New(dbtx).GetCategoryByName(ctx, name)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("category not found")
		}
		return nil, errors.Wrap(err, "get category by name")
	}
	return toCategoryEntity(r.ID, r.OrgID, r.ParentID, r.Name, r.SortOrder, r.Color, r.ImageUrl, r.CreatedAt, r.UpdatedAt), nil
}

func (s *categoryStore) Create(ctx context.Context, params gateway.CreateCategoryParams) (*entity.Category, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	parentID, err := uuidOrNull(params.ParentID)
	if err != nil {
		return nil, errors.BadRequest("invalid parent id")
	}
	r, err := sqlc.New(dbtx).CreateCategory(ctx, sqlc.CreateCategoryParams{
		OrgID:     orgID,
		ParentID:  parentID,
		Name:      params.Name,
		SortOrder: params.SortOrder,
		Color:     textOrNull(params.Color),
	})
	if err != nil {
		return nil, errors.Wrap(err, "create category")
	}
	return toCategoryEntity(r.ID, r.OrgID, r.ParentID, r.Name, r.SortOrder, r.Color, r.ImageUrl, r.CreatedAt, r.UpdatedAt), nil
}

func (s *categoryStore) Update(ctx context.Context, params gateway.UpdateCategoryParams) (*entity.Category, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseUUID(params.ID)
	if err != nil {
		return nil, errors.BadRequest("invalid category id")
	}
	parentID, err := uuidOrNull(params.ParentID)
	if err != nil {
		return nil, errors.BadRequest("invalid parent id")
	}
	r, err := sqlc.New(dbtx).UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:        id,
		Name:      params.Name,
		ParentID:  parentID,
		Color:     textOrNull(params.Color),
		SortOrder: params.SortOrder,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("category not found")
		}
		return nil, errors.Wrap(err, "update category")
	}
	return toCategoryEntity(r.ID, r.OrgID, r.ParentID, r.Name, r.SortOrder, r.Color, r.ImageUrl, r.CreatedAt, r.UpdatedAt), nil
}

func (s *categoryStore) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid category id")
	}
	if err := sqlc.New(dbtx).SoftDeleteCategory(ctx, uid); err != nil {
		return errors.Wrap(err, "soft delete category")
	}
	return nil
}

func toCategoryEntity(id, orgID, parentID pgtype.UUID, name string, sortOrder int32,
	color, imageURL pgtype.Text, createdAt, updatedAt pgtype.Timestamptz) *entity.Category {
	return &entity.Category{
		ID:        uuidString(id),
		OrgID:     uuidString(orgID),
		ParentID:  uuidString(parentID),
		Name:      name,
		SortOrder: sortOrder,
		Color:     textString(color),
		ImageURL:  textString(imageURL),
		CreatedAt: createdAt.Time,
		UpdatedAt: updatedAt.Time,
	}
}

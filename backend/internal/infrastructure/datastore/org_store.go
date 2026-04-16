package datastore

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore/sqlc"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type orgStore struct{}

// NewOrgReader returns an OrgReader backed by sqlc.
func NewOrgReader() gateway.OrgReader { return &orgStore{} }

// NewOrgWriter returns an OrgWriter backed by sqlc.
func NewOrgWriter() gateway.OrgWriter { return &orgStore{} }

func (r *orgStore) GetBySlug(ctx context.Context, slug string) (*entity.Org, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	row, err := sqlc.New(dbtx).GetOrgBySlug(ctx, slug)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("org not found")
		}
		return nil, errors.Wrap(err, "get org by slug")
	}
	return toOrgEntity(row.ID, row.Slug, row.Name, row.CreatedAt, row.UpdatedAt), nil
}

func (r *orgStore) GetByID(ctx context.Context, id string) (*entity.Org, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	row, err := sqlc.New(dbtx).GetOrgByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("org not found")
		}
		return nil, errors.Wrap(err, "get org by id")
	}
	return toOrgEntity(row.ID, row.Slug, row.Name, row.CreatedAt, row.UpdatedAt), nil
}

func (r *orgStore) Create(ctx context.Context, params gateway.CreateOrgParams) (*entity.Org, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	row, err := sqlc.New(dbtx).CreateOrg(ctx, sqlc.CreateOrgParams{
		Slug: params.Slug,
		Name: params.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create org")
	}
	return toOrgEntity(row.ID, row.Slug, row.Name, row.CreatedAt, row.UpdatedAt), nil
}

func toOrgEntity(id pgtype.UUID, slug, name string, createdAt, updatedAt pgtype.Timestamptz) *entity.Org {
	return &entity.Org{
		ID:        uuidString(id),
		Slug:      slug,
		Name:      name,
		CreatedAt: createdAt.Time,
		UpdatedAt: updatedAt.Time,
	}
}

// parseUUID converts a hyphenated string UUID into a pgtype.UUID.
func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return u, err
	}
	return u, nil
}

func uuidString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	return u.String()
}

// unused helper so imports stay satisfied when we stub future methods.
var _ = time.Time{}

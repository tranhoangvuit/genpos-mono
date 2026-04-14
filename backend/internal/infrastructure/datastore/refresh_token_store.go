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

type refreshTokenStore struct{}

// NewRefreshTokenReader returns a RefreshTokenReader backed by sqlc.
func NewRefreshTokenReader() gateway.RefreshTokenReader { return &refreshTokenStore{} }

// NewRefreshTokenWriter returns a RefreshTokenWriter backed by sqlc.
func NewRefreshTokenWriter() gateway.RefreshTokenWriter { return &refreshTokenStore{} }

func (r *refreshTokenStore) GetByHash(ctx context.Context, hash string) (*entity.RefreshToken, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	row, err := sqlc.New(dbtx).GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("refresh token not found")
		}
		return nil, errors.Wrap(err, "get refresh token")
	}
	return toRefreshTokenEntity(row), nil
}

func (r *refreshTokenStore) Create(ctx context.Context, params gateway.CreateRefreshTokenParams) (*entity.RefreshToken, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := parseUUID(params.UserID)
	if err != nil {
		return nil, errors.BadRequest("invalid user id")
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	row, err := sqlc.New(dbtx).CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID:    userID,
		OrgID:     orgID,
		TokenHash: params.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: params.ExpiresAt, Valid: true},
		UserAgent: params.UserAgent,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create refresh token")
	}
	return toRefreshTokenEntity(row), nil
}

func (r *refreshTokenStore) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid refresh token id")
	}
	err = sqlc.New(dbtx).RevokeRefreshToken(ctx, sqlc.RevokeRefreshTokenParams{
		ID:        uid,
		RevokedAt: pgtype.Timestamptz{Time: revokedAt, Valid: true},
	})
	if err != nil {
		return errors.Wrap(err, "revoke refresh token")
	}
	return nil
}

func toRefreshTokenEntity(row sqlc.RefreshToken) *entity.RefreshToken {
	var revokedAt *time.Time
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		revokedAt = &t
	}
	return &entity.RefreshToken{
		ID:        uuidString(row.ID),
		UserID:    uuidString(row.UserID),
		OrgID:     uuidString(row.OrgID),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt.Time,
		RevokedAt: revokedAt,
		UserAgent: row.UserAgent,
		CreatedAt: row.CreatedAt.Time,
	}
}

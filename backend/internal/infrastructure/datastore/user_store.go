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

type userStore struct{}

// NewUserReader returns a UserReader backed by sqlc.
func NewUserReader() gateway.UserReader { return &userStore{} }

// NewUserWriter returns a UserWriter backed by sqlc.
func NewUserWriter() gateway.UserWriter { return &userStore{} }

func (r *userStore) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	row, err := sqlc.New(dbtx).GetUserByEmail(ctx, email)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Wrap(err, "get user by email")
	}
	return toUserEntity(row), nil
}

func (r *userStore) GetByID(ctx context.Context, id string) (*entity.User, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid user id")
	}
	row, err := sqlc.New(dbtx).GetUserByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Wrap(err, "get user by id")
	}
	return toUserEntity(row), nil
}

func (r *userStore) Create(ctx context.Context, params gateway.CreateUserParams) (*entity.User, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	row, err := sqlc.New(dbtx).CreateUser(ctx, sqlc.CreateUserParams{
		OrgID:        orgID,
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		Name:         params.Name,
		Role:         params.Role,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create user")
	}
	return toUserEntity(row), nil
}

func toUserEntity(row sqlc.User) *entity.User {
	return &entity.User{
		ID:           uuidString(row.ID),
		OrgID:        uuidString(row.OrgID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Name:         row.Name,
		Role:         row.Role,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}

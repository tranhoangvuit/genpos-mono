package datastore

import (
	"context"
	"encoding/json"
	stderrors "errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

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
	return toUserFromGetRow(row.ID, row.OrgID, row.RoleID, row.Email, row.PasswordHash,
		row.Name, row.RoleName, row.RolePermissions, row.CreatedAt, row.UpdatedAt), nil
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
	return toUserFromGetRow(row.ID, row.OrgID, row.RoleID, row.Email, row.PasswordHash,
		row.Name, row.RoleName, row.RolePermissions, row.CreatedAt, row.UpdatedAt), nil
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
	roleID, err := parseUUID(params.RoleID)
	if err != nil {
		return nil, errors.BadRequest("invalid role id")
	}
	row, err := sqlc.New(dbtx).CreateUser(ctx, sqlc.CreateUserParams{
		OrgID:        orgID,
		RoleID:       roleID,
		Email:        pgtype.Text{String: params.Email, Valid: params.Email != ""},
		PasswordHash: pgtype.Text{String: params.PasswordHash, Valid: params.PasswordHash != ""},
		Name:         params.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create user")
	}
	return &entity.User{
		ID:           uuidString(row.ID),
		OrgID:        uuidString(row.OrgID),
		RoleID:       uuidString(row.RoleID),
		Email:        row.Email.String,
		PasswordHash: row.PasswordHash.String,
		Name:         row.Name,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}, nil
}

func toUserFromGetRow(
	id, orgID, roleID pgtype.UUID,
	email, passwordHash pgtype.Text,
	name, roleName string,
	rolePerms []byte,
	createdAt, updatedAt pgtype.Timestamptz,
) *entity.User {
	perms := make(map[string]string)
	_ = json.Unmarshal(rolePerms, &perms)

	return &entity.User{
		ID:           uuidString(id),
		OrgID:        uuidString(orgID),
		RoleID:       uuidString(roleID),
		Email:        email.String,
		PasswordHash: passwordHash.String,
		Name:         name,
		RoleName:     roleName,
		Permissions:  perms,
		CreatedAt:    createdAt.Time,
		UpdatedAt:    updatedAt.Time,
	}
}

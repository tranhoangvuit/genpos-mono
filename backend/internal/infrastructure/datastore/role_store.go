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

type roleStore struct{}

// NewRoleReader returns a RoleReader backed by sqlc.
func NewRoleReader() gateway.RoleReader { return &roleStore{} }

// NewRoleWriter returns a RoleWriter backed by sqlc.
func NewRoleWriter() gateway.RoleWriter { return &roleStore{} }

func (r *roleStore) GetByOrgAndName(ctx context.Context, orgID, name string) (*entity.Role, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	oid, err := parseUUID(orgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	row, err := sqlc.New(dbtx).GetRoleByOrgAndName(ctx, sqlc.GetRoleByOrgAndNameParams{
		OrgID: oid,
		Name:  name,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("role not found")
		}
		return nil, errors.Wrap(err, "get role by name")
	}
	return toRoleEntity(row.ID, row.OrgID, row.Name, row.Permissions, row.IsSystem, row.CreatedAt, row.UpdatedAt), nil
}

func (r *roleStore) Create(ctx context.Context, params gateway.CreateRoleParams) (*entity.Role, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	permsJSON, err := json.Marshal(params.Permissions)
	if err != nil {
		return nil, errors.Wrap(err, "marshal permissions")
	}
	row, err := sqlc.New(dbtx).CreateRole(ctx, sqlc.CreateRoleParams{
		OrgID:       orgID,
		Name:        params.Name,
		Permissions: permsJSON,
		IsSystem:    params.IsSystem,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create role")
	}
	return toRoleEntity(row.ID, row.OrgID, row.Name, row.Permissions, row.IsSystem, row.CreatedAt, row.UpdatedAt), nil
}

func toRoleEntity(
	id, orgID pgtype.UUID,
	name string,
	permsJSON []byte,
	isSystem bool,
	createdAt, updatedAt pgtype.Timestamptz,
) *entity.Role {
	perms := make(map[string]string)
	_ = json.Unmarshal(permsJSON, &perms)

	return &entity.Role{
		ID:          uuidString(id),
		OrgID:       uuidString(orgID),
		Name:        name,
		Permissions: perms,
		IsSystem:    isSystem,
		CreatedAt:   createdAt.Time,
		UpdatedAt:   updatedAt.Time,
	}
}

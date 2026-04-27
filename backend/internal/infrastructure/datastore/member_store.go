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

type memberStore struct{}

func NewMemberReader() gateway.MemberReader { return &memberStore{} }
func NewMemberWriter() gateway.MemberWriter { return &memberStore{} }

type memberRow struct {
	ID        pgtype.UUID
	OrgID     pgtype.UUID
	Name      string
	Email     pgtype.Text
	Phone     pgtype.Text
	RoleID    pgtype.UUID
	RoleName  string
	Status    string
	AllStores bool
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func toMemberEntity(r memberRow, storeIDs []pgtype.UUID) *entity.Member {
	return &entity.Member{
		ID:        uuidString(r.ID),
		OrgID:     uuidString(r.OrgID),
		Name:      r.Name,
		Email:     textString(r.Email),
		Phone:     textString(r.Phone),
		RoleID:    uuidString(r.RoleID),
		RoleName:  r.RoleName,
		Status:    r.Status,
		AllStores: r.AllStores,
		StoreIDs:  uuidStrings(storeIDs),
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: r.UpdatedAt.Time,
	}
}

func (s *memberStore) List(ctx context.Context) ([]*entity.Member, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	q := sqlc.New(dbtx)
	rows, err := q.ListMembers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "list members")
	}
	out := make([]*entity.Member, 0, len(rows))
	for _, r := range rows {
		storeIDs, err := q.ListMemberStoreIDs(ctx, r.ID)
		if err != nil {
			return nil, errors.Wrap(err, "list member stores")
		}
		out = append(out, toMemberEntity(memberRow(r), storeIDs))
	}
	return out, nil
}

func (s *memberStore) GetByID(ctx context.Context, id string) (*entity.Member, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return nil, errors.BadRequest("invalid member id")
	}
	q := sqlc.New(dbtx)
	r, err := q.GetMemberByID(ctx, uid)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFound("member not found")
		}
		return nil, errors.Wrap(err, "get member by id")
	}
	storeIDs, err := q.ListMemberStoreIDs(ctx, r.ID)
	if err != nil {
		return nil, errors.Wrap(err, "list member stores")
	}
	return toMemberEntity(memberRow(r), storeIDs), nil
}

func (s *memberStore) ListRoleOptions(ctx context.Context, orgID string) ([]*entity.RoleOption, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(orgID)
	if err != nil {
		return nil, errors.BadRequest("invalid org id")
	}
	rows, err := sqlc.New(dbtx).ListRolesByOrg(ctx, uid)
	if err != nil {
		return nil, errors.Wrap(err, "list roles")
	}
	out := make([]*entity.RoleOption, 0, len(rows))
	for _, r := range rows {
		out = append(out, &entity.RoleOption{
			ID:       uuidString(r.ID),
			Name:     r.Name,
			IsSystem: r.IsSystem,
		})
	}
	return out, nil
}

func (s *memberStore) HasStoreAccess(ctx context.Context, userID, storeID string) (bool, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return false, err
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return false, errors.BadRequest("invalid user id")
	}
	sid, err := parseUUID(storeID)
	if err != nil {
		return false, errors.BadRequest("invalid store id")
	}
	ok, err := sqlc.New(dbtx).HasStoreAccess(ctx, sqlc.HasStoreAccessParams{
		UserID:  uid,
		StoreID: sid,
	})
	if err != nil {
		return false, errors.Wrap(err, "has store access")
	}
	return ok, nil
}

func (s *memberStore) Create(ctx context.Context, params gateway.CreateMemberParams) (string, error) {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return "", err
	}
	orgID, err := parseUUID(params.OrgID)
	if err != nil {
		return "", errors.BadRequest("invalid org id")
	}
	roleID, err := parseUUID(params.RoleID)
	if err != nil {
		return "", errors.BadRequest("invalid role id")
	}
	id, err := sqlc.New(dbtx).CreateMember(ctx, sqlc.CreateMemberParams{
		OrgID:        orgID,
		RoleID:       roleID,
		Name:         params.Name,
		Email:        textOrNull(params.Email),
		Phone:        textOrNull(params.Phone),
		PasswordHash: textOrNull(params.PasswordHash),
		AllStores:    params.AllStores,
	})
	if err != nil {
		return "", errors.Wrap(err, "create member")
	}
	return uuidString(id), nil
}

func (s *memberStore) Update(ctx context.Context, params gateway.UpdateMemberParams) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	id, err := parseUUID(params.ID)
	if err != nil {
		return errors.BadRequest("invalid member id")
	}
	roleID, err := parseUUID(params.RoleID)
	if err != nil {
		return errors.BadRequest("invalid role id")
	}
	n, err := sqlc.New(dbtx).UpdateMember(ctx, sqlc.UpdateMemberParams{
		ID:        id,
		Name:      params.Name,
		Phone:     textOrNull(params.Phone),
		RoleID:    roleID,
		Status:    params.Status,
		AllStores: params.AllStores,
	})
	if err != nil {
		return errors.Wrap(err, "update member")
	}
	if n == 0 {
		return errors.NotFound("member not found")
	}
	return nil
}

func (s *memberStore) SoftDelete(ctx context.Context, id string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	uid, err := parseUUID(id)
	if err != nil {
		return errors.BadRequest("invalid member id")
	}
	q := sqlc.New(dbtx)
	if err := q.DeleteMemberStores(ctx, uid); err != nil {
		return errors.Wrap(err, "delete member stores")
	}
	n, err := q.SoftDeleteMember(ctx, uid)
	if err != nil {
		return errors.Wrap(err, "soft delete member")
	}
	if n == 0 {
		return errors.NotFound("member not found")
	}
	return nil
}

func (s *memberStore) ReplaceStores(ctx context.Context, orgID, userID string, storeIDs []string) error {
	dbtx, err := GetDBTX(ctx)
	if err != nil {
		return err
	}
	oid, err := parseUUID(orgID)
	if err != nil {
		return errors.BadRequest("invalid org id")
	}
	uid, err := parseUUID(userID)
	if err != nil {
		return errors.BadRequest("invalid user id")
	}
	q := sqlc.New(dbtx)
	if err := q.DeleteMemberStores(ctx, uid); err != nil {
		return errors.Wrap(err, "delete member stores")
	}
	seen := make(map[string]struct{}, len(storeIDs))
	for _, raw := range storeIDs {
		if raw == "" {
			continue
		}
		if _, dup := seen[raw]; dup {
			continue
		}
		seen[raw] = struct{}{}
		sid, err := parseUUID(raw)
		if err != nil {
			return errors.BadRequest("invalid store id")
		}
		if err := q.InsertMemberStore(ctx, sqlc.InsertMemberStoreParams{
			OrgID:   oid,
			UserID:  uid,
			StoreID: sid,
		}); err != nil {
			return errors.Wrap(err, "insert member store")
		}
	}
	return nil
}

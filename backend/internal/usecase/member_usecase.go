package usecase

import (
	"context"
	"strings"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/auth"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

var validMemberStatuses = map[string]struct{}{
	"active":    {},
	"inactive":  {},
	"suspended": {},
}

type memberUsecase struct {
	tenantDB gateway.TenantDB
	reader   gateway.MemberReader
	writer   gateway.MemberWriter
}

// NewMemberUsecase constructs a MemberUsecase.
func NewMemberUsecase(
	tenantDB gateway.TenantDB,
	reader gateway.MemberReader,
	writer gateway.MemberWriter,
) MemberUsecase {
	return &memberUsecase{tenantDB: tenantDB, reader: reader, writer: writer}
}

func (u *memberUsecase) ListMembers(ctx context.Context, orgID string) ([]*entity.Member, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.Member
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.reader.List(ctx)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list members")
	}
	return out, nil
}

func (u *memberUsecase) ListRoleOptions(ctx context.Context, orgID string) ([]*entity.RoleOption, error) {
	if orgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	var out []*entity.RoleOption
	if err := u.tenantDB.ReadWithTenant(ctx, orgID, func(ctx context.Context) error {
		v, err := u.reader.ListRoleOptions(ctx, orgID)
		if err != nil {
			return err
		}
		out = v
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "list role options")
	}
	return out, nil
}

func (u *memberUsecase) CreateMember(ctx context.Context, in input.CreateMemberInput) (*entity.Member, error) {
	if in.OrgID == "" {
		return nil, errors.BadRequest("org id is required")
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	if strings.TrimSpace(in.Email) == "" {
		return nil, errors.BadRequest("email is required")
	}
	if strings.TrimSpace(in.RoleID) == "" {
		return nil, errors.BadRequest("role is required")
	}
	if len(in.Password) < 8 {
		return nil, errors.BadRequest("password must be at least 8 characters")
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, errors.Wrap(err, "hash password")
	}
	// all_stores is the single source of truth — wipe explicit assignments
	// when it's on, regardless of what the client sent.
	storeIDs := in.StoreIDs
	if in.AllStores {
		storeIDs = nil
	}
	var out *entity.Member
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		id, err := u.writer.Create(ctx, gateway.CreateMemberParams{
			OrgID:        in.OrgID,
			RoleID:       in.RoleID,
			Name:         strings.TrimSpace(in.Name),
			Email:        strings.TrimSpace(strings.ToLower(in.Email)),
			Phone:        strings.TrimSpace(in.Phone),
			PasswordHash: hash,
			AllStores:    in.AllStores,
		})
		if err != nil {
			return err
		}
		if err := u.writer.ReplaceStores(ctx, in.OrgID, id, storeIDs); err != nil {
			return err
		}
		m, err := u.reader.GetByID(ctx, id)
		if err != nil {
			return err
		}
		out = m
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "create member")
	}
	return out, nil
}

func (u *memberUsecase) UpdateMember(ctx context.Context, in input.UpdateMemberInput) (*entity.Member, error) {
	if in.OrgID == "" || in.ID == "" {
		return nil, errors.BadRequest("id and org id are required")
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.BadRequest("name is required")
	}
	if strings.TrimSpace(in.RoleID) == "" {
		return nil, errors.BadRequest("role is required")
	}
	if _, ok := validMemberStatuses[in.Status]; !ok {
		return nil, errors.BadRequest("status must be active, inactive, or suspended")
	}
	storeIDs := in.StoreIDs
	if in.AllStores {
		storeIDs = nil
	}
	var out *entity.Member
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		if err := u.writer.Update(ctx, gateway.UpdateMemberParams{
			ID:        in.ID,
			Name:      strings.TrimSpace(in.Name),
			Phone:     strings.TrimSpace(in.Phone),
			RoleID:    in.RoleID,
			Status:    in.Status,
			AllStores: in.AllStores,
		}); err != nil {
			return err
		}
		if err := u.writer.ReplaceStores(ctx, in.OrgID, in.ID, storeIDs); err != nil {
			return err
		}
		m, err := u.reader.GetByID(ctx, in.ID)
		if err != nil {
			return err
		}
		out = m
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "update member")
	}
	return out, nil
}


func (u *memberUsecase) DeleteMember(ctx context.Context, in input.DeleteMemberInput) error {
	if in.OrgID == "" || in.ID == "" {
		return errors.BadRequest("id and org id are required")
	}
	if in.CurrentUserID != "" && in.CurrentUserID == in.ID {
		return errors.BadRequest("cannot delete your own account")
	}
	if err := u.tenantDB.WithTenant(ctx, in.OrgID, func(ctx context.Context) error {
		return u.writer.SoftDelete(ctx, in.ID)
	}); err != nil {
		return errors.Wrap(err, "delete member")
	}
	return nil
}

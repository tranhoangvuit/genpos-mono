package grpc

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	genposv1 "github.com/genpick/genpos-mono/backend/gen/genpos/v1"
	"github.com/genpick/genpos-mono/backend/gen/genpos/v1/genposv1connect"
	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/handler/interceptor"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

type MemberHandler struct {
	genposv1connect.UnimplementedMemberServiceHandler
	logger  *slog.Logger
	usecase usecase.MemberUsecase
}

func NewMemberHandler(logger *slog.Logger, uc usecase.MemberUsecase) *MemberHandler {
	return &MemberHandler{logger: logger, usecase: uc}
}

func (h *MemberHandler) ListMembers(
	ctx context.Context,
	_ *connect.Request[genposv1.ListMembersRequest],
) (*connect.Response[genposv1.ListMembersResponse], error) {
	authCtx, err := requireMemberAuth(ctx)
	if err != nil {
		return nil, err
	}
	items, err := h.usecase.ListMembers(ctx, authCtx.OrgID)
	if err != nil {
		return nil, h.logAndConvert(ctx, "list members", err)
	}
	pb := make([]*genposv1.Member, 0, len(items))
	for _, m := range items {
		pb = append(pb, toMemberProto(m))
	}
	return connect.NewResponse(&genposv1.ListMembersResponse{Members: pb}), nil
}

func (h *MemberHandler) ListRoles(
	ctx context.Context,
	_ *connect.Request[genposv1.ListRolesRequest],
) (*connect.Response[genposv1.ListRolesResponse], error) {
	authCtx, err := requireMemberAuth(ctx)
	if err != nil {
		return nil, err
	}
	opts, err := h.usecase.ListRoleOptions(ctx, authCtx.OrgID)
	if err != nil {
		return nil, h.logAndConvert(ctx, "list roles", err)
	}
	pb := make([]*genposv1.RoleOption, 0, len(opts))
	for _, r := range opts {
		pb = append(pb, &genposv1.RoleOption{
			Id:       r.ID,
			Name:     r.Name,
			IsSystem: r.IsSystem,
		})
	}
	return connect.NewResponse(&genposv1.ListRolesResponse{Roles: pb}), nil
}

func (h *MemberHandler) CreateMember(
	ctx context.Context,
	req *connect.Request[genposv1.CreateMemberRequest],
) (*connect.Response[genposv1.CreateMemberResponse], error) {
	authCtx, err := requireMemberAuth(ctx)
	if err != nil {
		return nil, err
	}
	msg := req.Msg.GetMember()
	m, err := h.usecase.CreateMember(ctx, input.CreateMemberInput{
		OrgID:    authCtx.OrgID,
		Name:     msg.GetName(),
		Email:    msg.GetEmail(),
		Phone:    msg.GetPhone(),
		RoleID:   msg.GetRoleId(),
		Password: msg.GetPassword(),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "create member", err)
	}
	return connect.NewResponse(&genposv1.CreateMemberResponse{Member: toMemberProto(m)}), nil
}

func (h *MemberHandler) UpdateMember(
	ctx context.Context,
	req *connect.Request[genposv1.UpdateMemberRequest],
) (*connect.Response[genposv1.UpdateMemberResponse], error) {
	authCtx, err := requireMemberAuth(ctx)
	if err != nil {
		return nil, err
	}
	msg := req.Msg.GetMember()
	m, err := h.usecase.UpdateMember(ctx, input.UpdateMemberInput{
		ID:     req.Msg.GetId(),
		OrgID:  authCtx.OrgID,
		Name:   msg.GetName(),
		Phone:  msg.GetPhone(),
		RoleID: msg.GetRoleId(),
		Status: msg.GetStatus(),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "update member", err)
	}
	return connect.NewResponse(&genposv1.UpdateMemberResponse{Member: toMemberProto(m)}), nil
}

func (h *MemberHandler) DeleteMember(
	ctx context.Context,
	req *connect.Request[genposv1.DeleteMemberRequest],
) (*connect.Response[genposv1.DeleteMemberResponse], error) {
	authCtx, err := requireMemberAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.usecase.DeleteMember(ctx, input.DeleteMemberInput{
		ID:            req.Msg.GetId(),
		OrgID:         authCtx.OrgID,
		CurrentUserID: authCtx.UserID,
	}); err != nil {
		return nil, h.logAndConvert(ctx, "delete member", err)
	}
	return connect.NewResponse(&genposv1.DeleteMemberResponse{}), nil
}

func requireMemberAuth(ctx context.Context) (*interceptor.AuthContext, error) {
	a := interceptor.FromContext(ctx)
	if a == nil {
		return nil, errors.ToConnectError(errors.Unauthorized("not signed in"))
	}
	return a, nil
}

func (h *MemberHandler) logAndConvert(ctx context.Context, op string, err error) error {
	if errors.ShouldLog(errors.GetCode(err)) {
		h.logger.ErrorContext(ctx, op+" failed", "error", err)
	}
	return errors.ToConnectError(err)
}

func toMemberProto(m *entity.Member) *genposv1.Member {
	if m == nil {
		return nil
	}
	return &genposv1.Member{
		Id:        m.ID,
		OrgId:     m.OrgID,
		Name:      m.Name,
		Email:     m.Email,
		Phone:     m.Phone,
		RoleId:    m.RoleID,
		RoleName:  m.RoleName,
		Status:    m.Status,
		CreatedAt: timestamppb.New(m.CreatedAt),
		UpdatedAt: timestamppb.New(m.UpdatedAt),
	}
}

var _ genposv1connect.MemberServiceHandler = (*MemberHandler)(nil)

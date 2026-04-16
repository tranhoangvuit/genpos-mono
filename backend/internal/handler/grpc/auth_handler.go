// Package grpc contains ConnectRPC handler adapters.
package grpc

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"

	genposv1 "github.com/genpick/genpos-mono/backend/gen/genpos/v1"
	"github.com/genpick/genpos-mono/backend/gen/genpos/v1/genposv1connect"
	"github.com/genpick/genpos-mono/backend/internal/config"
	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/handler/interceptor"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/auth"
	"github.com/genpick/genpos-mono/backend/pkg/cookies"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

// AuthHandler implements AuthServiceHandler.
type AuthHandler struct {
	genposv1connect.UnimplementedAuthServiceHandler
	logger    *slog.Logger
	usecase   usecase.AuthUsecase
	cookieCfg cookies.Config
	syncCfg   config.PowerSyncConfig
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(cfg *config.Config, logger *slog.Logger, uc usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		logger:  logger,
		usecase: uc,
		cookieCfg: cookies.Config{
			Domain: cfg.Auth.CookieDomain,
			Secure: cfg.Auth.CookieSecure,
		},
		syncCfg: cfg.PowerSync,
	}
}

func (h *AuthHandler) SignUp(
	ctx context.Context,
	req *connect.Request[genposv1.SignUpRequest],
) (*connect.Response[genposv1.SignUpResponse], error) {
	session, err := h.usecase.SignUp(ctx, input.SignUpInput{
		Domain:    req.Msg.GetDomain(),
		Email:     req.Msg.GetEmail(),
		Password:  req.Msg.GetPassword(),
		UserAgent: userAgent(req.Header()),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "sign up", err)
	}

	resp := connect.NewResponse(&genposv1.SignUpResponse{
		User: toProtoUser(session.User, session.Org),
	})
	h.setSessionCookies(resp.Header(), session)
	return resp, nil
}

func (h *AuthHandler) SignIn(
	ctx context.Context,
	req *connect.Request[genposv1.SignInRequest],
) (*connect.Response[genposv1.SignInResponse], error) {
	session, err := h.usecase.SignIn(ctx, input.SignInInput{
		Email:      req.Msg.GetEmail(),
		Password:   req.Msg.GetPassword(),
		RememberMe: req.Msg.GetRememberMe(),
		UserAgent:  userAgent(req.Header()),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "sign in", err)
	}

	resp := connect.NewResponse(&genposv1.SignInResponse{
		User: toProtoUser(session.User, session.Org),
	})
	h.setSessionCookies(resp.Header(), session)
	return resp, nil
}

func (h *AuthHandler) SignOut(
	ctx context.Context,
	req *connect.Request[genposv1.SignOutRequest],
) (*connect.Response[genposv1.SignOutResponse], error) {
	refreshToken, _ := cookies.Get(req.Header(), cookies.RefreshName)
	if err := h.usecase.SignOut(ctx, input.SignOutInput{RefreshToken: refreshToken}); err != nil {
		return nil, h.logAndConvert(ctx, "sign out", err)
	}

	resp := connect.NewResponse(&genposv1.SignOutResponse{})
	resp.Header().Add("Set-Cookie", cookies.ClearAccess(h.cookieCfg).String())
	resp.Header().Add("Set-Cookie", cookies.ClearRefresh(h.cookieCfg).String())
	return resp, nil
}

func (h *AuthHandler) Refresh(
	ctx context.Context,
	req *connect.Request[genposv1.RefreshRequest],
) (*connect.Response[genposv1.RefreshResponse], error) {
	refreshToken, ok := cookies.Get(req.Header(), cookies.RefreshName)
	if !ok {
		return nil, errors.ToConnectError(errors.Unauthorized("missing refresh token"))
	}
	session, err := h.usecase.Refresh(ctx, input.RefreshInput{
		RefreshToken: refreshToken,
		UserAgent:    userAgent(req.Header()),
	})
	if err != nil {
		return nil, h.logAndConvert(ctx, "refresh", err)
	}

	resp := connect.NewResponse(&genposv1.RefreshResponse{
		User: toProtoUser(session.User, session.Org),
	})
	h.setSessionCookies(resp.Header(), session)
	return resp, nil
}

func (h *AuthHandler) Me(
	ctx context.Context,
	_ *connect.Request[genposv1.MeRequest],
) (*connect.Response[genposv1.MeResponse], error) {
	authCtx := interceptor.FromContext(ctx)
	if authCtx == nil {
		return nil, errors.ToConnectError(errors.Unauthorized("not signed in"))
	}
	user, org, err := h.usecase.Me(ctx, authCtx.UserID)
	if err != nil {
		return nil, h.logAndConvert(ctx, "me", err)
	}
	return connect.NewResponse(&genposv1.MeResponse{
		User: toProtoUser(user, org),
	}), nil
}

func (h *AuthHandler) setSessionCookies(header http.Header, session *usecase.AuthSession) {
	access := cookies.Access(session.AccessToken, session.AccessTokenTTL, h.cookieCfg)
	refresh := cookies.Refresh(session.RefreshToken, session.RefreshTokenTTL, session.RefreshTokenIsLong, h.cookieCfg)
	header.Add("Set-Cookie", access.String())
	header.Add("Set-Cookie", refresh.String())
}

func (h *AuthHandler) logAndConvert(ctx context.Context, op string, err error) error {
	if errors.ShouldLog(errors.GetCode(err)) {
		h.logger.ErrorContext(ctx, op+" failed", "error", err)
	}
	return errors.ToConnectError(err)
}

func (h *AuthHandler) GetSyncToken(
	ctx context.Context,
	_ *connect.Request[genposv1.GetSyncTokenRequest],
) (*connect.Response[genposv1.GetSyncTokenResponse], error) {
	authCtx := interceptor.FromContext(ctx)
	if authCtx == nil {
		return nil, errors.ToConnectError(errors.Unauthorized("not signed in"))
	}

	token, err := auth.SignSyncToken(
		[]byte(h.syncCfg.JWTSecret),
		h.syncCfg.Audience,
		authCtx.UserID,
		authCtx.OrgID,
		h.syncCfg.TokenTTL,
	)
	if err != nil {
		return nil, h.logAndConvert(ctx, "get sync token", err)
	}

	expiresAt := time.Now().UTC().Add(h.syncCfg.TokenTTL).Unix()
	return connect.NewResponse(&genposv1.GetSyncTokenResponse{
		Token:     token,
		Endpoint:  h.syncCfg.Endpoint,
		ExpiresAt: expiresAt,
	}), nil
}

func toProtoUser(user *entity.User, org *entity.Org) *genposv1.AuthUser {
	return &genposv1.AuthUser{
		Id:      user.ID,
		OrgId:   user.OrgID,
		OrgSlug: org.Slug,
		Email:   user.Email,
		Name:    user.Name,
		Role:    user.RoleName,
	}
}

func userAgent(h http.Header) string {
	return h.Get("User-Agent")
}

var _ genposv1connect.AuthServiceHandler = (*AuthHandler)(nil)

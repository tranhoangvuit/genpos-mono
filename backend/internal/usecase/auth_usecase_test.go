package usecase_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/genpick/genpos-mono/backend/internal/config"
	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	gatewaymock "github.com/genpick/genpos-mono/backend/internal/domain/gateway/mock"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/auth"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

func newAuthTestConfig() *config.Config {
	return &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:       "test-secret-change-me",
			AccessTTL:       15 * time.Minute,
			RefreshTTLLong:  30 * 24 * time.Hour,
			RefreshTTLShort: 24 * time.Hour,
		},
	}
}

type authMocks struct {
	ctrl           *gomock.Controller
	tx             *gatewaymock.MockTxManager
	users          *gatewaymock.MockUserReader
	usersW         *gatewaymock.MockUserWriter
	orgs           *gatewaymock.MockOrgReader
	orgsW          *gatewaymock.MockOrgWriter
	roles          *gatewaymock.MockRoleReader
	rolesW         *gatewaymock.MockRoleWriter
	refreshTokens  *gatewaymock.MockRefreshTokenReader
	refreshTokensW *gatewaymock.MockRefreshTokenWriter
}

func newAuthMocks(t *testing.T) *authMocks {
	t.Helper()
	ctrl := gomock.NewController(t)
	return &authMocks{
		ctrl:           ctrl,
		tx:             gatewaymock.NewMockTxManager(ctrl),
		users:          gatewaymock.NewMockUserReader(ctrl),
		usersW:         gatewaymock.NewMockUserWriter(ctrl),
		orgs:           gatewaymock.NewMockOrgReader(ctrl),
		orgsW:          gatewaymock.NewMockOrgWriter(ctrl),
		roles:          gatewaymock.NewMockRoleReader(ctrl),
		rolesW:         gatewaymock.NewMockRoleWriter(ctrl),
		refreshTokens:  gatewaymock.NewMockRefreshTokenReader(ctrl),
		refreshTokensW: gatewaymock.NewMockRefreshTokenWriter(ctrl),
	}
}

func (m *authMocks) newUsecase() usecase.AuthUsecase {
	return usecase.NewAuthUsecase(
		newAuthTestConfig(),
		m.tx,
		m.users, m.usersW,
		m.orgs, m.orgsW,
		m.roles, m.rolesW,
		m.refreshTokens, m.refreshTokensW,
	)
}

// stubPassthroughTx makes the mock TxManager just execute fn directly.
func (m *authMocks) stubPassthroughTx() {
	m.tx.EXPECT().Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).AnyTimes()
}

func Test_AuthUsecase_SignUp(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		in          input.SignUpInput
		setup       func(*authMocks)
		wantErr     bool
		wantErrCode string
	}{
		"creates org and user with seeded roles": {
			in: input.SignUpInput{Domain: "acme", Email: "owner@acme.test", Password: "hunter2hunter"},
			setup: func(m *authMocks) {
				m.orgs.EXPECT().GetBySlug(gomock.Any(), "acme").
					Return(nil, errors.NotFound("org not found"))
				m.users.EXPECT().GetByEmail(gomock.Any(), "owner@acme.test").
					Return(nil, errors.NotFound("user not found"))
				m.stubPassthroughTx()
				m.orgsW.EXPECT().Create(gomock.Any(), gateway.CreateOrgParams{Slug: "acme", Name: "acme"}).
					Return(&entity.Org{ID: "org-1", Slug: "acme", Name: "acme"}, nil)
				// 3 default roles seeded; admin first
				m.rolesW.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(gateway.CreateRoleParams{})).
					DoAndReturn(func(_ context.Context, p gateway.CreateRoleParams) (*entity.Role, error) {
						return &entity.Role{
							ID:          "role-" + p.Name,
							OrgID:       p.OrgID,
							Name:        p.Name,
							Permissions: p.Permissions,
							IsSystem:    p.IsSystem,
						}, nil
					}).Times(3)
				m.usersW.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(gateway.CreateUserParams{})).
					DoAndReturn(func(_ context.Context, p gateway.CreateUserParams) (*entity.User, error) {
						if p.OrgID != "org-1" || p.Email != "owner@acme.test" || p.RoleID != "role-admin" || p.Name != "owner" {
							t.Errorf("unexpected create user params: %+v", p)
						}
						ok, err := auth.VerifyPassword("hunter2hunter", p.PasswordHash)
						if err != nil || !ok {
							t.Errorf("password hash does not verify: ok=%v err=%v", ok, err)
						}
						return &entity.User{
							ID: "user-1", OrgID: p.OrgID, RoleID: p.RoleID,
							Email: p.Email, PasswordHash: p.PasswordHash, Name: p.Name,
						}, nil
					})
				m.refreshTokensW.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(gateway.CreateRefreshTokenParams{})).
					Return(&entity.RefreshToken{ID: "rt-1"}, nil)
			},
		},
		"rejects invalid domain": {
			in:          input.SignUpInput{Domain: "Bad Domain!", Email: "a@b.com", Password: "hunter2hunter"},
			setup:       func(_ *authMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
		"rejects invalid email": {
			in:          input.SignUpInput{Domain: "acme", Email: "not-an-email", Password: "hunter2hunter"},
			setup:       func(_ *authMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
		"rejects short password": {
			in:          input.SignUpInput{Domain: "acme", Email: "a@b.com", Password: "short"},
			setup:       func(_ *authMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeBadRequest,
		},
		"rejects duplicate domain": {
			in: input.SignUpInput{Domain: "acme", Email: "a@b.com", Password: "hunter2hunter"},
			setup: func(m *authMocks) {
				m.orgs.EXPECT().GetBySlug(gomock.Any(), "acme").
					Return(&entity.Org{ID: "o", Slug: "acme"}, nil)
			},
			wantErr:     true,
			wantErrCode: errors.CodeConflict,
		},
		"rejects duplicate email": {
			in: input.SignUpInput{Domain: "acme", Email: "a@b.com", Password: "hunter2hunter"},
			setup: func(m *authMocks) {
				m.orgs.EXPECT().GetBySlug(gomock.Any(), "acme").
					Return(nil, errors.NotFound("org not found"))
				m.users.EXPECT().GetByEmail(gomock.Any(), "a@b.com").
					Return(&entity.User{ID: "u", Email: "a@b.com"}, nil)
			},
			wantErr:     true,
			wantErrCode: errors.CodeConflict,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := newAuthMocks(t)
			tc.setup(m)

			session, err := m.newUsecase().SignUp(context.Background(), tc.in)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.wantErrCode != "" && errors.GetCode(err) != tc.wantErrCode {
					t.Errorf("error code: want %s, got %s", tc.wantErrCode, errors.GetCode(err))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if session == nil || session.AccessToken == "" || session.RefreshToken == "" {
				t.Errorf("session is incomplete: %+v", session)
			}
			if !session.RefreshTokenIsLong {
				t.Errorf("sign up should issue long-lived refresh token")
			}
		})
	}
}

func Test_AuthUsecase_SignIn(t *testing.T) {
	t.Parallel()

	hash, err := auth.HashPassword("hunter2hunter")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	cases := map[string]struct {
		in          input.SignInInput
		setup       func(*authMocks)
		wantErr     bool
		wantErrCode string
		wantLong    bool
	}{
		"valid credentials with remember me": {
			in: input.SignInInput{Email: "owner@acme.test", Password: "hunter2hunter", RememberMe: true},
			setup: func(m *authMocks) {
				m.users.EXPECT().GetByEmail(gomock.Any(), "owner@acme.test").
					Return(&entity.User{
						ID: "u-1", OrgID: "o-1", Email: "owner@acme.test",
						PasswordHash: hash, Name: "owner", RoleName: "admin",
						Permissions: map[string]string{"*": "*"},
					}, nil)
				m.orgs.EXPECT().GetByID(gomock.Any(), "o-1").
					Return(&entity.Org{ID: "o-1", Slug: "acme", Name: "acme"}, nil)
				m.refreshTokensW.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(gateway.CreateRefreshTokenParams{})).
					Return(&entity.RefreshToken{ID: "rt-1"}, nil)
			},
			wantLong: true,
		},
		"valid credentials without remember me": {
			in: input.SignInInput{Email: "owner@acme.test", Password: "hunter2hunter", RememberMe: false},
			setup: func(m *authMocks) {
				m.users.EXPECT().GetByEmail(gomock.Any(), "owner@acme.test").
					Return(&entity.User{
						ID: "u-1", OrgID: "o-1", Email: "owner@acme.test",
						PasswordHash: hash, RoleName: "admin",
						Permissions: map[string]string{"*": "*"},
					}, nil)
				m.orgs.EXPECT().GetByID(gomock.Any(), "o-1").
					Return(&entity.Org{ID: "o-1", Slug: "acme"}, nil)
				m.refreshTokensW.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(gateway.CreateRefreshTokenParams{})).
					Return(&entity.RefreshToken{ID: "rt-2"}, nil)
			},
			wantLong: false,
		},
		"wrong password": {
			in: input.SignInInput{Email: "owner@acme.test", Password: "wrong-password", RememberMe: true},
			setup: func(m *authMocks) {
				m.users.EXPECT().GetByEmail(gomock.Any(), "owner@acme.test").
					Return(&entity.User{ID: "u-1", OrgID: "o-1", PasswordHash: hash}, nil)
			},
			wantErr:     true,
			wantErrCode: errors.CodeUnauthorized,
		},
		"unknown email": {
			in: input.SignInInput{Email: "ghost@acme.test", Password: "hunter2hunter", RememberMe: true},
			setup: func(m *authMocks) {
				m.users.EXPECT().GetByEmail(gomock.Any(), "ghost@acme.test").
					Return(nil, errors.NotFound("user not found"))
			},
			wantErr:     true,
			wantErrCode: errors.CodeUnauthorized,
		},
		"empty credentials": {
			in:          input.SignInInput{Email: "", Password: "", RememberMe: true},
			setup:       func(_ *authMocks) {},
			wantErr:     true,
			wantErrCode: errors.CodeUnauthorized,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := newAuthMocks(t)
			tc.setup(m)

			session, err := m.newUsecase().SignIn(context.Background(), tc.in)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.wantErrCode != "" && errors.GetCode(err) != tc.wantErrCode {
					t.Errorf("error code: want %s, got %s", tc.wantErrCode, errors.GetCode(err))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if session.RefreshTokenIsLong != tc.wantLong {
				t.Errorf("RefreshTokenIsLong: want %v, got %v", tc.wantLong, session.RefreshTokenIsLong)
			}
		})
	}
}

func Test_AuthUsecase_Refresh(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	cases := map[string]struct {
		refreshToken string
		setup        func(*authMocks, string)
		wantErr      bool
		wantErrCode  string
	}{
		"rotates active token": {
			refreshToken: "valid-refresh-value",
			setup: func(m *authMocks, token string) {
				hashed := auth.HashRefreshToken(token)
				m.refreshTokens.EXPECT().GetByHash(gomock.Any(), hashed).
					Return(&entity.RefreshToken{
						ID: "rt-old", UserID: "u-1", OrgID: "o-1",
						TokenHash: hashed, ExpiresAt: now.Add(10 * 24 * time.Hour),
					}, nil)
				m.users.EXPECT().GetByID(gomock.Any(), "u-1").
					Return(&entity.User{
						ID: "u-1", OrgID: "o-1", RoleName: "admin",
						Permissions: map[string]string{"*": "*"},
					}, nil)
				m.orgs.EXPECT().GetByID(gomock.Any(), "o-1").
					Return(&entity.Org{ID: "o-1", Slug: "acme"}, nil)
				m.refreshTokensW.EXPECT().Revoke(gomock.Any(), "rt-old", gomock.Any()).
					Return(nil)
				m.refreshTokensW.EXPECT().
					Create(gomock.Any(), gomock.AssignableToTypeOf(gateway.CreateRefreshTokenParams{})).
					Return(&entity.RefreshToken{ID: "rt-new"}, nil)
			},
		},
		"rejects expired token": {
			refreshToken: "expired-refresh-value",
			setup: func(m *authMocks, token string) {
				m.refreshTokens.EXPECT().GetByHash(gomock.Any(), auth.HashRefreshToken(token)).
					Return(&entity.RefreshToken{
						ID: "rt-old", UserID: "u-1", OrgID: "o-1",
						ExpiresAt: now.Add(-1 * time.Hour),
					}, nil)
			},
			wantErr:     true,
			wantErrCode: errors.CodeUnauthorized,
		},
		"rejects revoked token": {
			refreshToken: "revoked-refresh-value",
			setup: func(m *authMocks, token string) {
				revokedAt := now.Add(-1 * time.Minute)
				m.refreshTokens.EXPECT().GetByHash(gomock.Any(), auth.HashRefreshToken(token)).
					Return(&entity.RefreshToken{
						ID: "rt-old", UserID: "u-1", OrgID: "o-1",
						ExpiresAt: now.Add(1 * time.Hour), RevokedAt: &revokedAt,
					}, nil)
			},
			wantErr:     true,
			wantErrCode: errors.CodeUnauthorized,
		},
		"rejects unknown token": {
			refreshToken: "unknown-refresh-value",
			setup: func(m *authMocks, token string) {
				m.refreshTokens.EXPECT().GetByHash(gomock.Any(), auth.HashRefreshToken(token)).
					Return(nil, errors.NotFound("refresh token not found"))
			},
			wantErr:     true,
			wantErrCode: errors.CodeUnauthorized,
		},
		"rejects empty token": {
			refreshToken: "",
			setup:        func(_ *authMocks, _ string) {},
			wantErr:      true,
			wantErrCode:  errors.CodeUnauthorized,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := newAuthMocks(t)
			tc.setup(m, tc.refreshToken)

			_, err := m.newUsecase().Refresh(context.Background(), input.RefreshInput{RefreshToken: tc.refreshToken})

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.wantErrCode != "" && errors.GetCode(err) != tc.wantErrCode {
					t.Errorf("error code: want %s, got %s", tc.wantErrCode, errors.GetCode(err))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_AuthUsecase_SignOut(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		token string
		setup func(*authMocks, string)
	}{
		"revokes known token": {
			token: "active-token",
			setup: func(m *authMocks, token string) {
				m.refreshTokens.EXPECT().GetByHash(gomock.Any(), auth.HashRefreshToken(token)).
					Return(&entity.RefreshToken{ID: "rt-1"}, nil)
				m.refreshTokensW.EXPECT().Revoke(gomock.Any(), "rt-1", gomock.Any()).
					Return(nil)
			},
		},
		"tolerates unknown token": {
			token: "ghost-token",
			setup: func(m *authMocks, token string) {
				m.refreshTokens.EXPECT().GetByHash(gomock.Any(), auth.HashRefreshToken(token)).
					Return(nil, errors.NotFound("refresh token not found"))
			},
		},
		"no-op on empty token": {
			token: "",
			setup: func(_ *authMocks, _ string) {},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := newAuthMocks(t)
			tc.setup(m, tc.token)

			if err := m.newUsecase().SignOut(context.Background(), input.SignOutInput{RefreshToken: tc.token}); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

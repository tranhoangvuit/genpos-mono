package usecase

import (
	"context"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/genpick/genpos-mono/backend/internal/config"
	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
	"github.com/genpick/genpos-mono/backend/pkg/auth"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

//go:generate mockgen -source=auth_usecase.go -destination=mock/mock_auth_usecase.go -package=mock

// AuthSession is the result of a successful SignIn/SignUp/Refresh operation.
// Handlers translate this into two Set-Cookie response headers.
type AuthSession struct {
	User               *entity.User
	Org                *entity.Org
	AccessToken        string
	AccessTokenTTL     time.Duration
	RefreshToken       string
	RefreshTokenTTL    time.Duration
	RefreshTokenIsLong bool
}

// AuthUsecase is the service contract consumed by the AuthService handler.
type AuthUsecase interface {
	SignUp(ctx context.Context, in input.SignUpInput) (*AuthSession, error)
	SignIn(ctx context.Context, in input.SignInInput) (*AuthSession, error)
	SignOut(ctx context.Context, in input.SignOutInput) error
	Refresh(ctx context.Context, in input.RefreshInput) (*AuthSession, error)
	Me(ctx context.Context, userID string) (*entity.User, *entity.Org, error)
}

type authUsecase struct {
	cfg            config.AuthConfig
	users          gateway.UserReader
	usersW         gateway.UserWriter
	orgs           gateway.OrgReader
	orgsW          gateway.OrgWriter
	refreshTokens  gateway.RefreshTokenReader
	refreshTokensW gateway.RefreshTokenWriter
}

// NewAuthUsecase constructs an AuthUsecase.
func NewAuthUsecase(
	cfg *config.Config,
	users gateway.UserReader,
	usersW gateway.UserWriter,
	orgs gateway.OrgReader,
	orgsW gateway.OrgWriter,
	refreshTokens gateway.RefreshTokenReader,
	refreshTokensW gateway.RefreshTokenWriter,
) AuthUsecase {
	return &authUsecase{
		cfg:            cfg.Auth,
		users:          users,
		usersW:         usersW,
		orgs:           orgs,
		orgsW:          orgsW,
		refreshTokens:  refreshTokens,
		refreshTokensW: refreshTokensW,
	}
}

var slugRe = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$`)

const minPasswordLen = 8

func (u *authUsecase) SignUp(ctx context.Context, in input.SignUpInput) (*AuthSession, error) {
	domain := strings.ToLower(strings.TrimSpace(in.Domain))
	email := strings.TrimSpace(in.Email)

	if !slugRe.MatchString(domain) {
		return nil, errors.BadRequest("domain must be 1-32 lowercase alphanumeric characters or hyphens")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, errors.BadRequest("invalid email address")
	}
	if len(in.Password) < minPasswordLen {
		return nil, errors.BadRequest("password must be at least 8 characters")
	}

	// Reject duplicate domains and duplicate emails up front so we can return
	// a specific message. A concurrent race still hits the DB unique index.
	if existing, err := u.orgs.GetBySlug(ctx, domain); err == nil && existing != nil {
		return nil, errors.Conflict("domain already taken")
	} else if err != nil && errors.GetCode(err) != errors.CodeNotFound {
		return nil, errors.Wrap(err, "check org slug")
	}

	if existing, err := u.users.GetByEmail(ctx, email); err == nil && existing != nil {
		return nil, errors.Conflict("email already registered")
	} else if err != nil && errors.GetCode(err) != errors.CodeNotFound {
		return nil, errors.Wrap(err, "check user email")
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, errors.Wrap(err, "hash password")
	}

	org, err := u.orgsW.Create(ctx, gateway.CreateOrgParams{
		Slug: domain,
		Name: domain,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create org")
	}

	user, err := u.usersW.Create(ctx, gateway.CreateUserParams{
		OrgID:        org.ID,
		Email:        email,
		PasswordHash: hash,
		Name:         deriveName(email),
		Role:         "admin",
	})
	if err != nil {
		return nil, errors.Wrap(err, "create user")
	}

	return u.issueSession(ctx, user, org, true, in.UserAgent)
}

func (u *authUsecase) SignIn(ctx context.Context, in input.SignInInput) (*AuthSession, error) {
	email := strings.TrimSpace(in.Email)
	if email == "" || in.Password == "" {
		return nil, errors.Unauthorized("invalid email or password")
	}

	user, err := u.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.GetCode(err) == errors.CodeNotFound {
			return nil, errors.Unauthorized("invalid email or password")
		}
		return nil, errors.Wrap(err, "lookup user")
	}

	ok, err := auth.VerifyPassword(in.Password, user.PasswordHash)
	if err != nil {
		return nil, errors.Wrap(err, "verify password")
	}
	if !ok {
		return nil, errors.Unauthorized("invalid email or password")
	}

	org, err := u.orgs.GetByID(ctx, user.OrgID)
	if err != nil {
		return nil, errors.Wrap(err, "load org")
	}

	return u.issueSession(ctx, user, org, in.RememberMe, in.UserAgent)
}

func (u *authUsecase) SignOut(ctx context.Context, in input.SignOutInput) error {
	if in.RefreshToken == "" {
		return nil
	}
	hash := auth.HashRefreshToken(in.RefreshToken)
	token, err := u.refreshTokens.GetByHash(ctx, hash)
	if err != nil {
		if errors.GetCode(err) == errors.CodeNotFound {
			return nil
		}
		return errors.Wrap(err, "lookup refresh token")
	}
	return u.refreshTokensW.Revoke(ctx, token.ID, time.Now().UTC())
}

func (u *authUsecase) Refresh(ctx context.Context, in input.RefreshInput) (*AuthSession, error) {
	if in.RefreshToken == "" {
		return nil, errors.Unauthorized("missing refresh token")
	}
	hash := auth.HashRefreshToken(in.RefreshToken)

	existing, err := u.refreshTokens.GetByHash(ctx, hash)
	if err != nil {
		if errors.GetCode(err) == errors.CodeNotFound {
			return nil, errors.Unauthorized("invalid refresh token")
		}
		return nil, errors.Wrap(err, "lookup refresh token")
	}
	now := time.Now().UTC()
	if !existing.IsActive(now) {
		return nil, errors.Unauthorized("refresh token expired or revoked")
	}

	user, err := u.users.GetByID(ctx, existing.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "load user")
	}
	org, err := u.orgs.GetByID(ctx, existing.OrgID)
	if err != nil {
		return nil, errors.Wrap(err, "load org")
	}

	// Rotate: revoke the current token before issuing a new one.
	if err := u.refreshTokensW.Revoke(ctx, existing.ID, now); err != nil {
		return nil, errors.Wrap(err, "revoke old refresh token")
	}

	longLived := time.Until(existing.ExpiresAt) > u.cfg.RefreshTTLShort
	return u.issueSession(ctx, user, org, longLived, in.UserAgent)
}

func (u *authUsecase) Me(ctx context.Context, userID string) (*entity.User, *entity.Org, error) {
	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		if errors.GetCode(err) == errors.CodeNotFound {
			return nil, nil, errors.Unauthorized("session user not found")
		}
		return nil, nil, errors.Wrap(err, "load user")
	}
	org, err := u.orgs.GetByID(ctx, user.OrgID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "load org")
	}
	return user, org, nil
}

func (u *authUsecase) issueSession(ctx context.Context, user *entity.User, org *entity.Org, longLived bool, userAgent string) (*AuthSession, error) {
	accessToken, err := auth.SignAccessToken(
		[]byte(u.cfg.JWTSecret),
		user.ID, org.ID, org.Slug, user.Role,
		u.cfg.AccessTTL,
	)
	if err != nil {
		return nil, errors.Wrap(err, "sign access token")
	}

	refreshTTL := u.cfg.RefreshTTLLong
	if !longLived {
		refreshTTL = u.cfg.RefreshTTLShort
	}

	refreshToken, refreshHash, err := auth.NewRefreshToken()
	if err != nil {
		return nil, errors.Wrap(err, "generate refresh token")
	}

	if _, err := u.refreshTokensW.Create(ctx, gateway.CreateRefreshTokenParams{
		UserID:    user.ID,
		OrgID:     org.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().UTC().Add(refreshTTL),
		UserAgent: userAgent,
	}); err != nil {
		return nil, errors.Wrap(err, "persist refresh token")
	}

	return &AuthSession{
		User:               user,
		Org:                org,
		AccessToken:        accessToken,
		AccessTokenTTL:     u.cfg.AccessTTL,
		RefreshToken:       refreshToken,
		RefreshTokenTTL:    refreshTTL,
		RefreshTokenIsLong: longLived,
	}, nil
}

// deriveName returns a best-effort display name from an email address.
func deriveName(email string) string {
	at := strings.IndexByte(email, '@')
	if at <= 0 {
		return email
	}
	return email[:at]
}


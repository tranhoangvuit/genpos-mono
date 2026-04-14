// Package interceptor contains ConnectRPC interceptors.
package interceptor

import (
	"context"

	"connectrpc.com/connect"

	"github.com/genpick/genpos-mono/backend/internal/config"
	"github.com/genpick/genpos-mono/backend/pkg/auth"
	"github.com/genpick/genpos-mono/backend/pkg/cookies"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

// AuthContext holds the authenticated user extracted from the access cookie.
type AuthContext struct {
	UserID  string
	OrgID   string
	OrgSlug string
	Role    string
}

type authCtxKey struct{}

// FromContext returns the AuthContext for the current request, or nil.
func FromContext(ctx context.Context) *AuthContext {
	if a, ok := ctx.Value(authCtxKey{}).(*AuthContext); ok {
		return a
	}
	return nil
}

// WithAuth attaches an AuthContext to ctx.
func WithAuth(ctx context.Context, a *AuthContext) context.Context {
	return context.WithValue(ctx, authCtxKey{}, a)
}

// publicProcedures are RPCs that must succeed without an access cookie.
// Refresh is included because it validates the refresh cookie itself.
var publicProcedures = map[string]struct{}{
	"/genpos.v1.AuthService/SignUp":  {},
	"/genpos.v1.AuthService/SignIn":  {},
	"/genpos.v1.AuthService/SignOut": {},
	"/genpos.v1.AuthService/Refresh": {},
	"/genpos.v1.GenposService/Ping":  {},
}

// NewAuthInterceptor parses the gp_access cookie on every unary request and
// injects an AuthContext into the handler ctx. Public procedures are passed
// through without validation — they read cookies themselves when needed.
func NewAuthInterceptor(cfg *config.Config) connect.UnaryInterceptorFunc {
	secret := []byte(cfg.Auth.JWTSecret)
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if _, public := publicProcedures[req.Spec().Procedure]; public {
				return next(ctx, req)
			}

			token, ok := cookies.Get(req.Header(), cookies.AccessName)
			if !ok || token == "" {
				return nil, errors.Unauthorized("not signed in")
			}

			claims, err := auth.ParseAccessToken(secret, token)
			if err != nil {
				return nil, errors.Unauthorized("invalid session")
			}

			ctx = WithAuth(ctx, &AuthContext{
				UserID:  claims.UserID,
				OrgID:   claims.OrgID,
				OrgSlug: claims.OrgSlug,
				Role:    claims.Role,
			})
			return next(ctx, req)
		})
	}
}

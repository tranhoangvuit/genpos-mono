package interceptor

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore"
)

// NewDBInterceptor injects the raw pgx pool as a datastore DBTX for
// AuthService procedures. Auth-table queries (orgs, users, refresh_tokens)
// run without RLS, so they only need any DB connection — not a tenant-
// scoped one. Product procedures continue to use TenantDB.WithTenant.
func NewDBInterceptor(pool *pgxpool.Pool) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if strings.HasPrefix(req.Spec().Procedure, "/genpos.v1.AuthService/") {
				ctx = datastore.WithDBTX(ctx, pool)
			}
			return next(ctx, req)
		})
	}
}

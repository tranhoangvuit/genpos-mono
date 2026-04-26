package interceptor

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore"
)

// NewDBInterceptor injects the auth pgx pool as the datastore DBTX for
// AuthService procedures. The auth pool connects as a BYPASSRLS role so
// cross-tenant lookups (find user by email, list orgs for user) succeed
// before any org context is established. Product procedures continue to
// use TenantDB.WithTenant on the NOBYPASSRLS app pool.
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

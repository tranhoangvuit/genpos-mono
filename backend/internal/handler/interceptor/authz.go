package interceptor

import (
	"context"

	"connectrpc.com/connect"

	"github.com/genpick/genpos-mono/backend/pkg/auth"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

// ProcedurePermission maps a procedure name to its required resource and action.
type ProcedurePermission struct {
	Resource string
	Action   string
}

// NewPermissionInterceptor enforces that the caller's JWT permissions grant
// the resource:action required by the procedure. Public and auth-only
// procedures must be excluded from the map — they are handled by the auth
// interceptor or explicitly skipped here.
func NewPermissionInterceptor(rules map[string]ProcedurePermission) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			proc := req.Spec().Procedure

			if IsPublicProcedure(proc) {
				return next(ctx, req)
			}

			rule, mapped := rules[proc]
			if !mapped {
				return next(ctx, req)
			}

			authCtx := FromContext(ctx)
			if authCtx == nil {
				return nil, errors.Unauthorized("not signed in")
			}

			perms := auth.PermissionSet(authCtx.Permissions)
			if !perms.Allows(rule.Resource, rule.Action) {
				return nil, errors.Forbidden("insufficient permissions")
			}

			return next(ctx, req)
		})
	}
}

// DefaultProcedurePermissions returns the initial procedure → permission map.
func DefaultProcedurePermissions() map[string]ProcedurePermission {
	return map[string]ProcedurePermission{
		"/genpos.v1.GenposService/ListProducts": {Resource: "products", Action: "read"},
	}
}

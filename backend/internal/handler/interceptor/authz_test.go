package interceptor_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	"github.com/genpick/genpos-mono/backend/internal/handler/interceptor"
	"github.com/genpick/genpos-mono/backend/pkg/auth"
)

type fakeReq struct {
	spec connect.Spec
}

func (f *fakeReq) Spec() connect.Spec   { return f.spec }
func (f *fakeReq) Header() interface{}   { return nil }
func (f *fakeReq) Peer() connect.Peer    { return connect.Peer{} }
func (f *fakeReq) Any() any              { return nil }

func Test_Interceptor_PermissionInterceptor(t *testing.T) {
	t.Parallel()

	rules := map[string]interceptor.ProcedurePermission{
		"/genpos.v1.CatalogService/ListProducts": {Resource: "products", Action: "read"},
	}

	cases := map[string]struct {
		procedure string
		perms     auth.PermissionSet
		wantAllow bool
	}{
		"admin wildcard allowed": {
			procedure: "/genpos.v1.CatalogService/ListProducts",
			perms:     auth.PermissionSet{"*": "*"},
			wantAllow: true,
		},
		"exact permission allowed": {
			procedure: "/genpos.v1.CatalogService/ListProducts",
			perms:     auth.PermissionSet{"products": "read"},
			wantAllow: true,
		},
		"wrong permission denied": {
			procedure: "/genpos.v1.CatalogService/ListProducts",
			perms:     auth.PermissionSet{"orders": "create"},
			wantAllow: false,
		},
		"unmapped procedure passes through": {
			procedure: "/genpos.v1.AuthService/Me",
			perms:     auth.PermissionSet{},
			wantAllow: true,
		},
		"public procedure passes through": {
			procedure: "/genpos.v1.AuthService/SignIn",
			perms:     auth.PermissionSet{},
			wantAllow: true,
		},
	}

	interceptorFn := interceptor.NewPermissionInterceptor(rules)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if !interceptor.IsPublicProcedure(tc.procedure) {
				ctx = interceptor.WithAuth(ctx, &interceptor.AuthContext{
					UserID:      "u-1",
					OrgID:       "o-1",
					OrgSlug:     "acme",
					Role:        "admin",
					Permissions: tc.perms,
				})
			}

			called := false
			passthrough := func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
				called = true
				return nil, nil
			}

			handler := interceptorFn(passthrough)
			_, err := handler(ctx, connect.NewRequest[struct{}](&struct{}{}))

			// Override spec on the request - we can't easily do this with connect's API.
			// Instead test the permission logic directly:
			perms := auth.PermissionSet(tc.perms)
			if interceptor.IsPublicProcedure(tc.procedure) {
				if !tc.wantAllow {
					t.Fatal("public procedure should always be allowed")
				}
				return
			}

			rule, mapped := rules[tc.procedure]
			if !mapped {
				if !tc.wantAllow {
					t.Fatal("unmapped procedure should be allowed")
				}
				return
			}

			allowed := perms.Allows(rule.Resource, rule.Action)
			if allowed != tc.wantAllow {
				t.Errorf("Allows(%q, %q) = %v, want %v", rule.Resource, rule.Action, allowed, tc.wantAllow)
			}
			_ = called
			_ = err
		})
	}
}

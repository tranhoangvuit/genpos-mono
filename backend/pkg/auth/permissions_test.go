package auth_test

import (
	"testing"

	"github.com/genpick/genpos-mono/backend/pkg/auth"
)

func Test_Auth_PermissionSet_Allows(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		perms    auth.PermissionSet
		resource string
		action   string
		want     bool
	}{
		"wildcard allows anything": {
			perms: auth.PermissionSet{"*": "*"}, resource: "orders", action: "create", want: true,
		},
		"resource wildcard allows any action": {
			perms: auth.PermissionSet{"orders": "*"}, resource: "orders", action: "delete", want: true,
		},
		"exact match allows": {
			perms: auth.PermissionSet{"orders": "create"}, resource: "orders", action: "create", want: true,
		},
		"wrong action denies": {
			perms: auth.PermissionSet{"orders": "create"}, resource: "orders", action: "delete", want: false,
		},
		"missing resource denies": {
			perms: auth.PermissionSet{"products": "read"}, resource: "orders", action: "create", want: false,
		},
		"empty set denies": {
			perms: auth.PermissionSet{}, resource: "orders", action: "create", want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := tc.perms.Allows(tc.resource, tc.action)
			if got != tc.want {
				t.Errorf("Allows(%q, %q) = %v, want %v", tc.resource, tc.action, got, tc.want)
			}
		})
	}
}

func Test_Auth_DefaultRoles(t *testing.T) {
	t.Parallel()

	roles := auth.DefaultRoles()
	if len(roles) != 3 {
		t.Fatalf("want 3 default roles, got %d", len(roles))
	}

	cases := map[string]struct {
		idx      int
		name     string
		isSystem bool
	}{
		"admin is first and system": {idx: 0, name: "admin", isSystem: true},
		"manager is second":        {idx: 1, name: "manager", isSystem: false},
		"cashier is third":         {idx: 2, name: "cashier", isSystem: false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			r := roles[tc.idx]
			if r.Name != tc.name {
				t.Errorf("name: want %s, got %s", tc.name, r.Name)
			}
			if r.IsSystem != tc.isSystem {
				t.Errorf("isSystem: want %v, got %v", tc.isSystem, r.IsSystem)
			}
		})
	}
}

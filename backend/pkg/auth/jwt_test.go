package auth_test

import (
	"testing"
	"time"

	"github.com/genpick/genpos-mono/backend/pkg/auth"
)

func Test_Auth_SignAndParseAccessToken(t *testing.T) {
	t.Parallel()

	secret := []byte("test-secret-do-not-use-in-production")
	perms := auth.PermissionSet{"*": "*"}

	cases := map[string]struct {
		userID  string
		orgID   string
		orgSlug string
		role    string
		ttl     time.Duration
	}{
		"admin claims round-trip": {
			userID: "u-1", orgID: "o-1", orgSlug: "acme", role: "admin", ttl: time.Hour,
		},
		"staff claims round-trip": {
			userID: "u-2", orgID: "o-2", orgSlug: "globex", role: "staff", ttl: time.Hour,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			token, err := auth.SignAccessToken(secret, tc.userID, tc.orgID, tc.orgSlug, tc.role, perms, tc.ttl)
			if err != nil {
				t.Fatalf("sign: %v", err)
			}

			claims, err := auth.ParseAccessToken(secret, token)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if claims.UserID != tc.userID ||
				claims.OrgID != tc.orgID ||
				claims.OrgSlug != tc.orgSlug ||
				claims.Role != tc.role {
				t.Errorf("claims mismatch: %+v", claims)
			}
			if !claims.Permissions.Allows("anything", "anything") {
				t.Errorf("expected wildcard permissions in token")
			}
		})
	}
}

func Test_Auth_ParseAccessToken_Rejects(t *testing.T) {
	t.Parallel()

	secret := []byte("test-secret")
	wrongSecret := []byte("other-secret")
	perms := auth.PermissionSet{"*": "*"}

	validToken, err := auth.SignAccessToken(secret, "u", "o", "s", "admin", perms, time.Hour)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	expiredToken, err := auth.SignAccessToken(secret, "u", "o", "s", "admin", perms, -time.Second)
	if err != nil {
		t.Fatalf("sign expired: %v", err)
	}

	cases := map[string]struct {
		token  string
		secret []byte
	}{
		"wrong signing secret": {validToken, wrongSecret},
		"expired token":        {expiredToken, secret},
		"garbage token":        {"not.a.jwt", secret},
		"empty token":          {"", secret},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := auth.ParseAccessToken(tc.secret, tc.token); err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

func Test_Auth_NewRefreshToken_IsUniqueAndHashable(t *testing.T) {
	t.Parallel()

	seen := map[string]struct{}{}
	for range 50 {
		token, hash, err := auth.NewRefreshToken()
		if err != nil {
			t.Fatalf("NewRefreshToken: %v", err)
		}
		if token == "" || hash == "" {
			t.Fatal("token or hash is empty")
		}
		if auth.HashRefreshToken(token) != hash {
			t.Errorf("HashRefreshToken produced a different hash than NewRefreshToken")
		}
		if _, dup := seen[token]; dup {
			t.Errorf("duplicate token generated")
		}
		seen[token] = struct{}{}
	}
}

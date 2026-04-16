package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/genpick/genpos-mono/backend/pkg/auth"
)

func Test_Auth_SignSyncToken_RoundTrip(t *testing.T) {
	t.Parallel()

	secret := []byte("test-sync-secret")

	cases := map[string]struct {
		audience string
		userID   string
		orgID    string
		ttl      time.Duration
	}{
		"standard claims": {
			audience: "powersync-dev", userID: "u-1", orgID: "o-1", ttl: 5 * time.Minute,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			token, err := auth.SignSyncToken(secret, tc.audience, tc.userID, tc.orgID, tc.ttl)
			if err != nil {
				t.Fatalf("sign: %v", err)
			}

			parsed, err := jwt.ParseWithClaims(token, &auth.SyncClaims{}, func(_ *jwt.Token) (any, error) {
				return secret, nil
			})
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			claims := parsed.Claims.(*auth.SyncClaims)
			if claims.Subject != tc.userID {
				t.Errorf("sub: want %s, got %s", tc.userID, claims.Subject)
			}
			if claims.OrgID != tc.orgID {
				t.Errorf("org: want %s, got %s", tc.orgID, claims.OrgID)
			}
			aud, _ := claims.GetAudience()
			if len(aud) != 1 || aud[0] != tc.audience {
				t.Errorf("aud: want [%s], got %v", tc.audience, aud)
			}
		})
	}
}

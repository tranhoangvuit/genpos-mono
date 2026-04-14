package auth_test

import (
	"testing"

	"github.com/genpick/genpos-mono/backend/pkg/auth"
)

func Test_Auth_HashAndVerifyPassword(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		password string
		verify   string
		want     bool
	}{
		"matches correct password":     {"correct horse battery", "correct horse battery", true},
		"rejects wrong password":       {"correct horse battery", "wrong horse battery", false},
		"rejects empty against hashed": {"correct horse battery", "", false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			hash, err := auth.HashPassword(tc.password)
			if err != nil {
				t.Fatalf("HashPassword: %v", err)
			}
			got, err := auth.VerifyPassword(tc.verify, hash)
			if err != nil {
				t.Fatalf("VerifyPassword: %v", err)
			}
			if got != tc.want {
				t.Errorf("VerifyPassword = %v, want %v", got, tc.want)
			}
		})
	}
}

func Test_Auth_VerifyPassword_RejectsMalformedHash(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"empty":            "",
		"not phc":          "not-a-real-hash",
		"wrong algorithm":  "$argon2i$v=19$m=65536,t=2,p=4$c2FsdA$aGFzaA",
		"too few segments": "$argon2id$v=19$m=65536",
	}

	for name, hash := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ok, err := auth.VerifyPassword("any-password", hash)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok {
				t.Errorf("expected malformed hash to be rejected")
			}
		})
	}
}

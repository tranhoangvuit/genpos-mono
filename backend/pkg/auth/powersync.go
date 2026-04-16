package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

// SyncClaims is the JWT payload for PowerSync client authentication.
type SyncClaims struct {
	OrgID string `json:"o"`
	jwt.RegisteredClaims
}

// SignSyncToken returns a signed JWT for PowerSync with user_id as sub and org_id as custom claim.
func SignSyncToken(secret []byte, audience, userID, orgID string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := SyncClaims{
		OrgID: orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(secret)
	if err != nil {
		return "", errors.Wrap(err, "sign sync token")
	}
	return signed, nil
}

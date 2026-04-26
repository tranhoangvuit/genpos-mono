package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

// SyncParameters holds the dynamic claims that PowerSync sync rules read via
// `token_parameters.<name>`. PowerSync looks for them under a top-level
// `parameters` JWT claim, NOT at the JWT root.
type SyncParameters struct {
	OrgID string `json:"o"`
}

// SyncClaims is the JWT payload for PowerSync client authentication.
type SyncClaims struct {
	Parameters SyncParameters `json:"parameters"`
	jwt.RegisteredClaims
}

// OrgID returns the org_id carried in the parameters claim. Kept for callers
// (and tests) that previously read it from the JWT root.
func (c *SyncClaims) OrgID() string { return c.Parameters.OrgID }

// SignSyncToken returns a signed JWT for PowerSync with user_id as sub and org_id as a token parameter.
func SignSyncToken(secret []byte, audience, userID, orgID string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := SyncClaims{
		Parameters: SyncParameters{OrgID: orgID},
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

package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

// AccessClaims is the JWT payload for the short-lived access token.
type AccessClaims struct {
	UserID  string `json:"uid"`
	OrgID   string `json:"oid"`
	OrgSlug string `json:"osl"`
	Role    string `json:"rol"`
	jwt.RegisteredClaims
}

// SignAccessToken returns a signed JWT for the given claims.
func SignAccessToken(secret []byte, userID, orgID, orgSlug, role string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := AccessClaims{
		UserID:  userID,
		OrgID:   orgID,
		OrgSlug: orgSlug,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(secret)
	if err != nil {
		return "", errors.Wrap(err, "sign access token")
	}
	return signed, nil
}

// ParseAccessToken verifies signature + expiry and returns the claims.
// Returns an Unauthorized error on any failure so callers don't leak details.
func ParseAccessToken(secret []byte, token string) (*AccessClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Unauthorized("invalid signing method")
		}
		return secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.Unauthorized("invalid or expired access token")
	}
	claims, ok := parsed.Claims.(*AccessClaims)
	if !ok {
		return nil, errors.Unauthorized("invalid access token claims")
	}
	return claims, nil
}

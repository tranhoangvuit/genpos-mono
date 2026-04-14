package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

// NewRefreshToken returns a cryptographically-random opaque token (base64url,
// no padding) along with its sha256 hash suitable for storage.
func NewRefreshToken() (token string, hash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", errors.Wrap(err, "generate refresh token")
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	hash = HashRefreshToken(token)
	return token, hash, nil
}

// HashRefreshToken returns the hex-encoded sha256 of the token.
// Storing only the hash means a leaked refresh_tokens table is useless to
// an attacker without also stealing the plaintext cookie value.
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

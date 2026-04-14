package gateway

import (
	"context"
	"time"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=refresh_token_gateway.go -destination=mock/mock_refresh_token_gateway.go -package=mock

// CreateRefreshTokenParams carries parameters for creating a refresh token row.
type CreateRefreshTokenParams struct {
	UserID    string
	OrgID     string
	TokenHash string
	ExpiresAt time.Time
	UserAgent string
}

// RefreshTokenReader reads refresh tokens without tenant context.
type RefreshTokenReader interface {
	GetByHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
}

// RefreshTokenWriter creates and revokes refresh tokens without tenant context.
type RefreshTokenWriter interface {
	Create(ctx context.Context, params CreateRefreshTokenParams) (*entity.RefreshToken, error)
	Revoke(ctx context.Context, id string, revokedAt time.Time) error
}

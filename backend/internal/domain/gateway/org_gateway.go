package gateway

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
)

//go:generate mockgen -source=org_gateway.go -destination=mock/mock_org_gateway.go -package=mock

// CreateOrgParams carries parameters for creating an org.
type CreateOrgParams struct {
	Slug string
	Name string
}

// OrgReader reads orgs without tenant context (used during auth).
type OrgReader interface {
	GetBySlug(ctx context.Context, slug string) (*entity.Org, error)
	GetByID(ctx context.Context, id string) (*entity.Org, error)
}

// OrgWriter creates orgs without tenant context (used during signup).
type OrgWriter interface {
	Create(ctx context.Context, params CreateOrgParams) (*entity.Org, error)
}

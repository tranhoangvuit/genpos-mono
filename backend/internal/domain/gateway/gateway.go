package gateway

import "context"

//go:generate mockgen -source=gateway.go -destination=mock/mock_gateway.go -package=mock

// TenantDB abstracts multi-tenant database scoping.
// Implementations set the PostgreSQL session variable app.current_org_id
// so that RLS policies filter rows automatically.
type TenantDB interface {
	// WithTenant executes fn inside a transaction scoped to orgID.
	WithTenant(ctx context.Context, orgID string, fn func(ctx context.Context) error) error

	// ReadWithTenant executes fn on a connection scoped to orgID (no transaction).
	ReadWithTenant(ctx context.Context, orgID string, fn func(ctx context.Context) error) error
}

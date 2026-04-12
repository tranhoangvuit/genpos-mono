//go:build wireinject

package app

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/config"
	"github.com/genpick/genpos-mono/backend/internal/domain/gateway"
	"github.com/genpick/genpos-mono/backend/internal/handler"
	"github.com/genpick/genpos-mono/backend/internal/infrastructure/datastore"
	"github.com/genpick/genpos-mono/backend/internal/usecase"
	"github.com/goforj/wire"
)

// InitializeApp documents the dependency graph.
// wire_gen.go is the hand-written equivalent until wire supports Go 1.25.
func InitializeApp(ctx context.Context, cfg *config.Config) (*App, error) {
	wire.Build(
		provideLogger,
		providePostgresDB,
		providePool,
		datastore.NewTenantDB,
		wire.Bind(new(gateway.TenantDB), new(*datastore.TenantDB)),
		datastore.NewProductReader,
		wire.Bind(new(gateway.ProductReader), new(*datastore.ProductReader)),
		usecase.NewProductUsecase,
		wire.Bind(new(usecase.ProductUsecase), new(*usecase.ProductUsecaseImpl)),
		handler.NewServer,
		wire.Struct(new(App), "*"),
	)
	return nil, nil
}

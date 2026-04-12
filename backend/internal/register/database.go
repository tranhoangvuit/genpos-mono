package register

import (
	"context"
	"log/slog"

	"github.com/genpick/genpos-mono/backend/internal/config"
	"github.com/genpick/genpos-mono/backend/pkg/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProvidePostgresDB creates a new PostgreSQL database connection.
func ProvidePostgresDB(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*database.PostgresDB, error) {
	db, err := database.NewPostgresDB(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// ProvidePool extracts the pgxpool.Pool from PostgresDB.
func ProvidePool(db *database.PostgresDB) *pgxpool.Pool {
	return db.Pool
}

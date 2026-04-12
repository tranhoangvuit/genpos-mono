package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds PostgreSQL connection configuration.
type Config struct {
	Host            string `required:"true" split_words:"true"`
	Port            int    `required:"true" split_words:"true" default:"5432"`
	Database        string `required:"true" split_words:"true"`
	User            string `required:"true" split_words:"true"`
	Password        string `required:"true" split_words:"true"`
	SSLMode         string `split_words:"true" default:"disable"`
	MaxConns        int32  `split_words:"true" default:"25"`
	MinConns        int32  `split_words:"true" default:"5"`
	MaxConnLifetime string `split_words:"true" default:"1h"`
	MaxConnIdleTime string `split_words:"true" default:"30m"`
}

// ConnectionString returns the PostgreSQL connection string.
func (c Config) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Database, c.User, c.Password, c.SSLMode,
	)
}

// PostgresDB wraps pgxpool.Pool for database operations.
type PostgresDB struct {
	Pool *pgxpool.Pool
	cfg  Config
}

// NewPostgresDB creates a new PostgreSQL database connection pool.
func NewPostgresDB(ctx context.Context, cfg Config) (*PostgresDB, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set pool configuration
	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns

	if cfg.MaxConnLifetime != "" {
		duration, err := time.ParseDuration(cfg.MaxConnLifetime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse max_conn_lifetime: %w", err)
		}
		poolCfg.MaxConnLifetime = duration
	}

	if cfg.MaxConnIdleTime != "" {
		duration, err := time.ParseDuration(cfg.MaxConnIdleTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse max_conn_idle_time: %w", err)
		}
		poolCfg.MaxConnIdleTime = duration
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{
		Pool: pool,
		cfg:  cfg,
	}, nil
}

// Close closes the database connection pool.
func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Health checks if the database connection is healthy.
func (db *PostgresDB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Stats returns the current pool statistics.
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

package config

import (
	"log/slog"
	"time"

	"github.com/genpick/genpos-mono/backend/pkg/database"
	"github.com/genpick/genpos-mono/backend/pkg/log"
	"github.com/kelseyhightower/envconfig"
)

// Env represents the application environment.
type Env string

const (
	EnvDev  Env = "dev"
	EnvStag Env = "stag"
	EnvProd Env = "prod"
)

func (e Env) IsProduction() bool  { return e == EnvProd }
func (e Env) IsDevelopment() bool { return e == EnvDev }

// AuthConfig holds session/auth settings.
type AuthConfig struct {
	JWTSecret       string        `envconfig:"JWT_SECRET" default:"dev-insecure-change-me"`
	AccessTTL       time.Duration `envconfig:"ACCESS_TTL" default:"15m"`
	RefreshTTLLong  time.Duration `envconfig:"REFRESH_TTL_LONG" default:"720h"` // 30 days
	RefreshTTLShort time.Duration `envconfig:"REFRESH_TTL_SHORT" default:"24h"`
	CookieDomain    string        `envconfig:"COOKIE_DOMAIN" default:""`
	CookieSecure    bool          `envconfig:"COOKIE_SECURE" default:"false"`
	FrontendOrigins []string      `envconfig:"FRONTEND_ORIGINS" default:"http://localhost:3032,http://localhost:1420,tauri://localhost,http://tauri.localhost,https://tauri.localhost"`
}

// PowerSyncConfig holds PowerSync JWT bridge settings.
type PowerSyncConfig struct {
	JWTSecret string        `envconfig:"JWT_SECRET" default:"my-dev-secret-key-for-jwt-signing"`
	Audience  string        `envconfig:"AUDIENCE" default:"powersync-dev"`
	Endpoint  string        `envconfig:"ENDPOINT" default:"http://localhost:3034"`
	TokenTTL  time.Duration `envconfig:"TOKEN_TTL" default:"5m"`
}

// Config holds all application configuration.
type Config struct {
	ServiceName string `envconfig:"SERVICE_NAME" default:"genpos"`
	ServerPort  int    `envconfig:"SERVER_PORT" default:"3031"`
	Env         Env    `envconfig:"ENV" default:"dev"`

	Database     database.Config `envconfig:"DATABASE"`
	DatabaseAuth AuthDBOverride  `envconfig:"DATABASE_AUTH"`
	Log          log.Config      `envconfig:"LOG"`
	Auth         AuthConfig      `envconfig:"AUTH"`
	PowerSync    PowerSyncConfig `envconfig:"POWERSYNC"`
}

// AuthDBOverride supplies the credentials for the auth-only DB pool. Host /
// port / database / SSL come from the main Database config — only the user
// and password differ (auth uses a BYPASSRLS role for cross-tenant lookups).
type AuthDBOverride struct {
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
}

// AuthDatabase returns a database.Config that reuses Database connection
// settings but with the auth role's credentials.
func (c *Config) AuthDatabase() database.Config {
	cfg := c.Database
	cfg.User = c.DatabaseAuth.User
	cfg.Password = c.DatabaseAuth.Password
	return cfg
}

// Load loads the configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	if cfg.Auth.JWTSecret == "dev-insecure-change-me" {
		if cfg.Env.IsProduction() {
			return nil, envconfig.ErrInvalidSpecification
		}
		slog.Warn("AUTH_JWT_SECRET is using the insecure dev default — set it in production")
	}
	return &cfg, nil
}

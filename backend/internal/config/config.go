package config

import (
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

// Config holds all application configuration.
type Config struct {
	ServiceName string `envconfig:"SERVICE_NAME" default:"genpos"`
	ServerPort  int    `envconfig:"SERVER_PORT" default:"8081"`
	Env         Env    `envconfig:"ENV" default:"dev"`

	Database database.Config `envconfig:"DATABASE"`
	Log      log.Config      `envconfig:"LOG"`
}

// Load loads the configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

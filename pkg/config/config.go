package config

import (
	"github.com/caarlos0/env/v11"
)

// AppConfig holds the global application configuration
type AppConfig struct {
	// Server configuration
	Port     int    `env:"PORT" envDefault:"8080"`
	Host     string `env:"HOST" envDefault:"localhost"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	// Hatchet configuration
	HatchetServerURL string `env:"HATCHET_CLIENT_SERVER_URL" envDefault:"http://localhost:8888"`
	HatchetHostPort  string `env:"HATCHET_CLIENT_HOST_PORT" envDefault:"localhost:7070"`
	HatchetToken     string `env:"HATCHET_CLIENT_TOKEN" envDefault:"test-token-for-integration"`

	// Database configuration (for future use)
	DatabaseURL string `env:"DATABASE_URL" envDefault:""`
}

// LoadAppConfig loads the global application configuration from environment variables
func LoadAppConfig() (*AppConfig, error) {
	cfg := &AppConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

package config

import (
	"github.com/Gen-Do/skeleton-service/internal/pkg/env"
)

// Config holds all configuration for the service
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logging  LoggingConfig
	Tracing  TracingConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port        string
	ServiceName string
	Environment string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	URL string
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level string
}

// TracingConfig holds tracing-related configuration
type TracingConfig struct {
	JaegerEndpoint string
	Enabled        bool
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:        env.GetString("PORT", "8080"),
			ServiceName: env.GetString("SERVICE_NAME", "service-skeleton"),
			Environment: env.GetString("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			URL: env.GetString("DATABASE_URL", ""),
		},
		Logging: LoggingConfig{
			Level: env.GetString("LOG_LEVEL", "info"),
		},
		Tracing: TracingConfig{
			JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
			Enabled:        env.GetBool("TRACING_ENABLED", true),
		},
	}
}

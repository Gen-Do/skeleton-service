package config

import (
	"os"
	"strconv"
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
			Port:        getEnv("PORT", "8080"),
			ServiceName: getEnv("SERVICE_NAME", "service-skeleton"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		Logging: LoggingConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Tracing: TracingConfig{
			JaegerEndpoint: getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
			Enabled:        getEnvBool("TRACING_ENABLED", true),
		},
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets boolean environment variable with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

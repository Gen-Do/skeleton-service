package logger

import (
	"time"

	"github.com/Gen-Do/skeleton-service/internal/pkg/env"
	"github.com/sirupsen/logrus"
)

// Config содержит настройки для логгера
type Config struct {
	Level  string
	Format string // "json" или "text"
}

// DefaultConfig возвращает конфигурацию логгера по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Level:  env.GetString("LOG_LEVEL", "info"),
		Format: env.GetString("LOG_FORMAT", "json"),
	}
}

// Setup настраивает и возвращает настроенный логгер
func Setup(config *Config) *logrus.Logger {
	logger := logrus.New()

	// Настройка форматтера
	if config.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Настройка уровня логирования
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		logger.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return logger
}

// SetupDefault настраивает логгер с конфигурацией по умолчанию
func SetupDefault() *logrus.Logger {
	return Setup(DefaultConfig())
}

// WithFields создает новый логгер с дополнительными полями
func WithFields(logger *logrus.Logger, fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}

// WithServiceContext добавляет контекст сервиса к логгеру
func WithServiceContext(logger *logrus.Logger, serviceName, version string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"service": serviceName,
		"version": version,
	})
}

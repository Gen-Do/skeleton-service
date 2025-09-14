package logger

import (
	"time"

	"github.com/Gen-Do/skeleton-service/internal/pkg/env"
	"github.com/sirupsen/logrus"
)

// New настраивает и возвращает настроенный логгер
func New() logrus.FieldLogger {
	logger := logrus.New()

	if env.GetString("LOG_FORMAT", "json") == "json" {
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
	level, err := logrus.ParseLevel(env.GetString("LOG_LEVEL", "info"))
	if err != nil {
		logger.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return withServiceContext(
		logger,
		env.GetString("SERVICE_NAME", "skeleton"),
		env.GetString("SERVICE_VERSION", "0"),
	)
}

// withServiceContext добавляет контекст сервиса к логгеру
func withServiceContext(logger *logrus.Logger, serviceName, version string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"service": serviceName,
		"version": version,
	})
}

package tracing

import (
	"context"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// Config содержит настройки для трассировки
type Config struct {
	ServiceName    string
	ServiceVersion string
	JaegerEndpoint string
	Enabled        bool
	SamplingRate   float64
}

// DefaultConfig возвращает конфигурацию трассировки по умолчанию
func DefaultConfig() *Config {
	return &Config{
		ServiceName:    getEnv("SERVICE_NAME", "service-skeleton"),
		ServiceVersion: getEnv("SERVICE_VERSION", "1.0.0"),
		JaegerEndpoint: getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		Enabled:        getEnvBool("TRACING_ENABLED", false),
		SamplingRate:   getEnvFloat("TRACING_SAMPLING_RATE", 1.0),
	}
}

// TracerProvider обертка для trace.TracerProvider с дополнительными методами
type TracerProvider struct {
	*trace.TracerProvider
	config *Config
}

// Setup настраивает и возвращает TracerProvider
func Setup(config *Config) (*TracerProvider, error) {
	if !config.Enabled {
		// Возвращаем no-op tracer provider
		tp := trace.NewTracerProvider()
		return &TracerProvider{
			TracerProvider: tp,
			config:         config,
		}, nil
	}

	// Создание Jaeger экспортера
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint(config.JaegerEndpoint),
	))
	if err != nil {
		return nil, err
	}

	// Создание ресурса с информацией о сервисе
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(config.ServiceName),
		semconv.ServiceVersion(config.ServiceVersion),
	)

	// Создание tracer provider с настройками
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(config.SamplingRate)),
	)

	// Установка глобального tracer provider
	otel.SetTracerProvider(tp)

	return &TracerProvider{
		TracerProvider: tp,
		config:         config,
	}, nil
}

// SetupDefault настраивает трассировку с конфигурацией по умолчанию
func SetupDefault() (*TracerProvider, error) {
	return Setup(DefaultConfig())
}

// Shutdown корректно завершает работу tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.TracerProvider != nil {
		return tp.TracerProvider.Shutdown(ctx)
	}
	return nil
}

// IsEnabled возвращает true, если трассировка включена
func (tp *TracerProvider) IsEnabled() bool {
	return tp.config.Enabled
}

// GetConfig возвращает конфигурацию трассировки
func (tp *TracerProvider) GetConfig() *Config {
	return tp.config
}

// Вспомогательные функции для работы с переменными окружения

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := parseFloat(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// parseFloat парсит строку в float64
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

package tracing

import (
	"context"

	"github.com/Gen-Do/skeleton-service/internal/pkg/env"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// Config содержит настройки для трассировки
type Config struct {
	ServiceName    string
	ServiceVersion string
	OTLPEndpoint   string
	Enabled        bool
	SamplingRate   float64
}

// defaultConfig возвращает конфигурацию трассировки по умолчанию
func defaultConfig() *Config {
	return &Config{
		ServiceName:    env.GetString("SERVICE_NAME", "service-skeleton"),
		ServiceVersion: env.GetString("SERVICE_VERSION", "0"),
		OTLPEndpoint:   env.GetString("TRACING_ENDPOINT", "http://localhost:4318/v1/traces"),
		Enabled:        env.GetBool("TRACING_ENABLED", false),
		SamplingRate:   env.GetFloat64("TRACING_SAMPLING_RATE", 1.0),
	}
}

// TracerProvider обертка для trace.TracerProvider с дополнительными методами
type TracerProvider struct {
	*trace.TracerProvider
	config *Config
}

// New настраивает и возвращает TracerProvider
func New() (*TracerProvider, error) {
	config := defaultConfig()
	if !config.Enabled {
		// Возвращаем no-op tracer provider
		tp := trace.NewTracerProvider()
		return &TracerProvider{
			TracerProvider: tp,
			config:         config,
		}, nil
	}

	// Создание OTLP HTTP экспортера
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(config.OTLPEndpoint),
		otlptracehttp.WithInsecure(), // для локальной разработки
	)
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

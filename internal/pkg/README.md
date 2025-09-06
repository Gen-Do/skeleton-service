# Пакеты инфраструктуры

Этот каталог содержит переиспользуемые пакеты для настройки основных компонентов сервиса.

## Структура

- `logger/` - настройка и конфигурация логгера (Logrus)
- `tracing/` - настройка трассировки (OpenTelemetry + Jaeger)  
- `metrics/` - настройка метрик (Prometheus)
- `server/` - настройка HTTP сервера (Chi router)

## Использование

### Logger

```go
import "github.com/gendo/service-skeleton/internal/pkg/logger"

// С конфигурацией по умолчанию
log := logger.SetupDefault()

// С кастомной конфигурацией
log := logger.Setup(&logger.Config{
    Level:  "debug",
    Format: "json",
})

// С контекстом сервиса
serviceLog := logger.WithServiceContext(log, "my-service", "1.0.0")
```

### Tracing

```go
import "github.com/gendo/service-skeleton/internal/pkg/tracing"

// Настройка трассировки
tracerProvider, err := tracing.Setup(&tracing.Config{
    ServiceName:    "my-service",
    ServiceVersion: "1.0.0",
    JaegerEndpoint: "http://localhost:14268/api/traces",
    Enabled:        true,
    SamplingRate:   1.0,
})
defer tracerProvider.Shutdown(context.Background())
```

### Metrics

```go
import "github.com/gendo/service-skeleton/internal/pkg/metrics"

// Настройка метрик
metricsCollector := metrics.Setup(&metrics.Config{
    ServiceName: "my-service",
    Namespace:   "service",
    Enabled:     true,
})

// Регистрация дополнительных метрик
counter := metricsCollector.CreateCounter(
    "operations_total",
    "Total operations",
    []string{"operation", "status"},
)

// Middleware для HTTP метрик
router.Use(metricsCollector.Middleware(logger))

// Handler для /metrics эндпоинта
router.Handle("/metrics", metricsCollector.Handler())
```

### Server

```go
import "github.com/gendo/service-skeleton/internal/pkg/server"

// Настройка сервера
srv := server.Setup(&server.Config{
    Port:         "8080",
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}, logger)

// Настройка middleware
srv.SetupMiddleware(true, true) // CORS, Tracing

// Регистрация маршрутов
srv.Get("/ping", pingHandler)
srv.Route("/api/v1", func(r chi.Router) {
    r.Get("/users", getUsersHandler)
})

// Запуск
srv.StartAsync()
```

## Принципы

1. **Конфигурируемость** - каждый пакет принимает структуру конфигурации
2. **Значения по умолчанию** - функции `SetupDefault()` для быстрого старта
3. **Переиспользование** - пакеты можно использовать в разных сервисах
4. **Расширяемость** - возможность добавления кастомных коллекторов, middleware, маршрутов
5. **Изоляция** - каждый пакет независим и может быть заменен

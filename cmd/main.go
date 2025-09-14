package main

import (
	"context"
	"net/http"

	"github.com/Gen-Do/skeleton-service/internal/pkg/env"
	"github.com/Gen-Do/skeleton-service/internal/pkg/logger"
	"github.com/Gen-Do/skeleton-service/internal/pkg/metrics"
	"github.com/Gen-Do/skeleton-service/internal/pkg/server"
	"github.com/Gen-Do/skeleton-service/internal/pkg/shutdown"
	"github.com/Gen-Do/skeleton-service/internal/pkg/tracing"
	"github.com/go-chi/chi/v5"
	// Импорты для работы с сгенерированными API
	// "github.com/Gen-Do/skeleton-service/internal/generated/api"
	// "github.com/Gen-Do/skeleton-service/internal/handlers"
	// "github.com/getkin/kin-openapi/openapi3filter"
	// oapimiddleware "github.com/deepmap/oapi-codegen/v2/pkg/chi-middleware"
)

func main() {
	// Загрузка переменных окружения из файлов .env.paas и .env.override
	env.LoadEnvFiles()

	log := logger.New()

	log.Info("Starting service")

	// Настройка трассировки
	tracerProvider, err := tracing.New()
	if err != nil {
		log.WithError(err).Fatal("Failed to setup tracing")
	}
	defer func() {
		if err = tracerProvider.Shutdown(context.Background()); err != nil {
			log.WithError(err).Error("Failed to shutdown tracer provider")
		}
	}()

	// Настройка метрик
	metricsCollector := metrics.New()

	// Пример регистрации дополнительных метрик
	// customCounter := metricsCollector.CreateCounter(
	//     "custom_operations_total",
	//     "Total number of custom operations",
	//     []string{"operation_type", "status"},
	// )

	// Настройка HTTP сервера
	httpServer := server.New(metricsCollector, log)
	httpServer.AddMiddleware(metricsCollector.Middleware())
	httpServer.AddHealthCheck("/health")
	httpServer.Handle("GET", "/metrics", metricsCollector.Handler())
	httpServer.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"pong"}`))
	})

	// Регистрация API маршрутов
	httpServer.Route("/api/v1", func(r chi.Router) {
		// Пример подключения сгенерированных API хендлеров с валидацией:
		//
		// 1. Загрузка OpenAPI спецификации для валидации
		// swagger, err := api.GetSwagger()
		// if err != nil {
		//     logger.WithError(err).Fatal("Failed to load OpenAPI spec")
		// }
		// swagger.Servers = nil // Убираем серверы для локальной валидации
		//
		// 2. Настройка middleware для валидации запросов
		// validationOptions := &oapimiddleware.Options{
		//     ErrorHandler: func(w http.ResponseWriter, message string, statusCode int) {
		//         w.Header().Set("Content-Type", "application/json")
		//         w.WriteHeader(statusCode)
		//         response := fmt.Sprintf(`{"error":"validation_error","message":"%s"}`, message)
		//         w.Write([]byte(response))
		//     },
		//     Options: openapi3filter.Options{
		//         AuthenticationFunc: func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		//             // Здесь можно добавить логику аутентификации
		//             return nil
		//         },
		//     },
		// }
		//
		// 3. Применение middleware валидации ко всем API маршрутам
		// r.Use(oapimiddleware.OapiRequestValidatorWithOptions(swagger, validationOptions))
		//
		// 4. Создание экземпляра обработчиков
		// apiHandlers := &handlers.APIHandlers{
		//     Logger: logger.Logger, // Преобразуем Entry в Logger
		//     // DB: db,
		//     // Services: services,
		// }
		//
		// 5. Регистрация сгенерированных маршрутов
		// api.HandlerFromMux(apiHandlers, r)
		//
		// Альтернативный способ - ручная регистрация конкретных маршрутов:
		// r.Get("/users", apiHandlers.GetUsers)
		// r.Post("/users", apiHandlers.CreateUser)
		// r.Get("/users/{userId}", apiHandlers.GetUserById)
		// r.Put("/users/{userId}", apiHandlers.UpdateUser)
		// r.Delete("/users/{userId}", apiHandlers.DeleteUser)
	})

	// Запуск сервера асинхронно
	httpServer.StartAsync()

	// Graceful shutdown
	shutdown.GracefulShutdown(httpServer, log)
}

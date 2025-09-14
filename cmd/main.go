package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/Gen-Do/skeleton-service/internal/config"
	"github.com/Gen-Do/skeleton-service/internal/pkg/logger"
	"github.com/Gen-Do/skeleton-service/internal/pkg/metrics"
	"github.com/Gen-Do/skeleton-service/internal/pkg/server"
	"github.com/Gen-Do/skeleton-service/internal/pkg/tracing"
	// Импорты для работы с сгенерированными API
	// "github.com/Gen-Do/skeleton-service/internal/generated/api"
	// "github.com/Gen-Do/skeleton-service/internal/handlers"
	// "github.com/getkin/kin-openapi/openapi3filter"
	// oapimiddleware "github.com/deepmap/oapi-codegen/v2/pkg/chi-middleware"
)

func main() {
	// Загрузка переменных окружения из .env файла
	if err := godotenv.Load(); err != nil {
		// Не фатальная ошибка - можем работать с системными переменными
		logrus.Info("No .env file found, using system environment variables")
	}

	// Загрузка конфигурации
	cfg := config.Load()

	// Настройка логгера
	log := logger.Setup(&logger.Config{
		Level:  cfg.Logging.Level,
		Format: "json", // Можно добавить в конфиг
	})

	// Создание контекстного логгера для сервиса
	serviceLogger := logger.WithServiceContext(log, cfg.Server.ServiceName, "1.0.0")

	serviceLogger.WithFields(logrus.Fields{
		"service": cfg.Server.ServiceName,
		"port":    cfg.Server.Port,
	}).Info("Starting service")

	// Настройка трассировки
	tracerProvider, err := tracing.Setup(&tracing.Config{
		ServiceName:    cfg.Server.ServiceName,
		ServiceVersion: "1.0.0",
		OTLPEndpoint:   cfg.Tracing.JaegerEndpoint,
		Enabled:        cfg.Tracing.Enabled,
		SamplingRate:   1.0,
	})
	if err != nil {
		serviceLogger.WithError(err).Fatal("Failed to setup tracing")
	}
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			serviceLogger.WithError(err).Error("Failed to shutdown tracer provider")
		}
	}()

	// Настройка метрик
	metricsCollector := metrics.Setup(&metrics.Config{
		ServiceName: cfg.Server.ServiceName,
		Namespace:   "service",
		Enabled:     true,
	})

	// Пример регистрации дополнительных метрик
	// customCounter := metricsCollector.CreateCounter(
	//     "custom_operations_total",
	//     "Total number of custom operations",
	//     []string{"operation_type", "status"},
	// )

	// Настройка HTTP сервера
	httpServer := server.Setup(&server.Config{
		Port:         cfg.Server.Port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}, log)

	// Настройка middleware
	httpServer.SetupMiddleware(
		true,                       // Enable CORS
		tracerProvider.IsEnabled(), // Enable tracing if configured
	)

	// Добавление middleware для метрик
	httpServer.AddMiddleware(metricsCollector.Middleware(log))

	// Регистрация базовых маршрутов
	setupBasicRoutes(httpServer, metricsCollector)

	// Регистрация API маршрутов
	setupAPIRoutes(httpServer, serviceLogger)

	// Запуск сервера асинхронно
	httpServer.StartAsync()

	// Graceful shutdown
	gracefulShutdown(httpServer, serviceLogger)
}

// setupBasicRoutes настраивает базовые маршруты (health, metrics, ping)
func setupBasicRoutes(srv *server.Server, metricsCollector *metrics.Metrics) {
	// Health check
	srv.AddHealthCheck("/health")

	// Metrics endpoint
	srv.Handle("GET", "/metrics", metricsCollector.Handler())

	// Ping endpoint
	srv.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"pong"}`))
	})
}

// setupAPIRoutes настраивает API маршруты
func setupAPIRoutes(srv *server.Server, logger *logrus.Entry) {
	srv.Route("/api/v1", func(r chi.Router) {
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

		// Пример дополнительного маршрута
		r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ready","version":"1.0.0"}`))
		})
	})
}

// gracefulShutdown обрабатывает корректное завершение работы сервиса
func gracefulShutdown(srv *server.Server, logger *logrus.Entry) {
	// Создание канала для получения сигналов ОС
	quit := make(chan os.Signal, 1)

	// Регистрация канала для получения определенных сигналов
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Блокируем выполнение до получения сигнала
	<-quit
	logger.Info("Shutting down server...")

	// Создание контекста с таймаутом для завершения работы
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Попытка корректного завершения работы сервера
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
		return
	}

	logger.Info("Server exited gracefully")
}

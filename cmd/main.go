package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	// Импорты для работы с сгенерированными API
	// "github.com/gendo/service-skeleton/internal/generated/api"
	// "github.com/gendo/service-skeleton/internal/api/handlers"
	// "github.com/getkin/kin-openapi/openapi3filter"
	// oapimiddleware "github.com/deepmap/oapi-codegen/v2/pkg/chi-middleware"
)

var (
	// Prometheus metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		logrus.Info("No .env file found, using system environment variables")
	}

	// Setup logger
	logger := setupLogger()
	
	// Setup OpenTelemetry tracing
	tracerProvider, err := setupTracing()
	if err != nil {
		logger.WithError(err).Fatal("Failed to setup tracing")
	}
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			logger.WithError(err).Error("Failed to shutdown tracer provider")
		}
	}()

	// Get configuration from environment
	port := getEnv("PORT", "8080")
	serviceName := getEnv("SERVICE_NAME", "service-skeleton")
	
	logger.WithFields(logrus.Fields{
		"service": serviceName,
		"port":    port,
	}).Info("Starting service")

	// Setup router with middlewares
	router := setupRouter(logger)
	
	// Setup HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.WithField("port", port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	gracefulShutdown(server, logger)
}

func setupLogger() *logrus.Logger {
	logger := logrus.New()
	
	// Set log format to JSON for structured logging
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	
	// Set log level
	level := getEnv("LOG_LEVEL", "info")
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logger.WithError(err).Warn("Invalid log level, using info")
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)
	
	return logger
}

func setupTracing() (*trace.TracerProvider, error) {
	serviceName := getEnv("SERVICE_NAME", "service-skeleton")
	jaegerEndpoint := getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces")
	
	// Create Jaeger exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		return nil, err
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create tracer provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)
	
	return tp, nil
}

func setupRouter(logger *logrus.Logger) *chi.Mux {
	router := chi.NewRouter()

	// Setup CORS
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Setup middlewares
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(loggingMiddleware(logger))
	router.Use(middleware.Recoverer)
	router.Use(metricsMiddleware)
	router.Use(otelhttp.NewMiddleware("http-server"))

	// Setup routes
	setupRoutes(router)
	
	return router
}

func setupRoutes(router *chi.Mux) {
	// Health check endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	// API routes group
	router.Route("/api/v1", func(r chi.Router) {
		// Пример подключения сгенерированных API хендлеров с валидацией:
		// 
		// 1. Загрузка OpenAPI спецификации для валидации
		// swagger, err := api.GetSwagger()
		// if err != nil {
		//     panic(err)
		// }
		// swagger.Servers = nil // Убираем серверы для локальной валидации
		//
		// 2. Настройка middleware для валидации запросов
		// validationOptions := &oapimiddleware.Options{
		//     ErrorHandler: func(w http.ResponseWriter, message string, statusCode int) {
		//         w.Header().Set("Content-Type", "application/json")
		//         w.WriteHeader(statusCode)
		//         w.Write([]byte(fmt.Sprintf(`{"error":"validation_error","message":"%s"}`, message)))
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
		//     // Здесь инициализируются зависимости для хендлеров
		//     Logger: logger,
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

		// Базовый тестовый эндпоинт (удалить после добавления реальных API)
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"pong"}`))
		})
	})
}

func loggingMiddleware(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Create a wrapped response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			
			// Get trace context
			span := oteltrace.SpanFromContext(r.Context())
			traceID := span.SpanContext().TraceID().String()
			
			// Call next handler
			next.ServeHTTP(ww, r)
			
			// Log request
			logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      ww.Status(),
				"duration_ms": time.Since(start).Milliseconds(),
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
				"trace_id":    traceID,
			}).Info("HTTP request processed")
		})
	}
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a wrapped response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		
		// Call next handler
		next.ServeHTTP(ww, r)
		
		// Record metrics
		duration := time.Since(start).Seconds()
		status := http.StatusText(ww.Status())
		
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

func gracefulShutdown(server *http.Server, logger *logrus.Logger) {
	// Create a channel to receive OS signals
	quit := make(chan os.Signal, 1)
	
	// Register the channel to receive specific signals
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// Block until a signal is received
	<-quit
	logger.Info("Shutting down server...")

	// Create a deadline for the shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
		return
	}

	logger.Info("Server exited gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

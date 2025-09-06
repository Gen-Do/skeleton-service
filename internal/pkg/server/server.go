package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// Config содержит настройки для HTTP сервера
type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DefaultConfig возвращает конфигурацию сервера по умолчанию
func DefaultConfig(port string) *Config {
	return &Config{
		Port:         port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// Server представляет HTTP сервер с настроенными middleware
type Server struct {
	config *Config
	router *chi.Mux
	server *http.Server
	logger *logrus.Logger
}

// RouteRegistrar интерфейс для регистрации маршрутов
type RouteRegistrar interface {
	RegisterRoutes(router chi.Router)
}

// Setup создает и настраивает HTTP сервер
func Setup(config *Config, logger *logrus.Logger) *Server {
	router := chi.NewRouter()

	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &Server{
		config: config,
		router: router,
		server: server,
		logger: logger,
	}
}

// SetupDefault создает сервер с конфигурацией по умолчанию
func SetupDefault(port string, logger *logrus.Logger) *Server {
	return Setup(DefaultConfig(port), logger)
}

// SetupMiddleware настраивает стандартные middleware
func (s *Server) SetupMiddleware(enableCORS bool, enableTracing bool) {
	// CORS middleware
	if enableCORS {
		s.router.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}))
	}

	// Базовые middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(s.loggingMiddleware())
	s.router.Use(middleware.Recoverer)

	// OpenTelemetry middleware
	if enableTracing {
		s.router.Use(otelhttp.NewMiddleware("http-server"))
	}
}

// AddMiddleware добавляет кастомное middleware
func (s *Server) AddMiddleware(middlewares ...func(http.Handler) http.Handler) {
	for _, mw := range middlewares {
		s.router.Use(mw)
	}
}

// RegisterRoutes регистрирует маршруты через RouteRegistrar
func (s *Server) RegisterRoutes(registrar RouteRegistrar) {
	registrar.RegisterRoutes(s.router)
}

// Route создает группу маршрутов
func (s *Server) Route(pattern string, fn func(r chi.Router)) {
	s.router.Route(pattern, fn)
}

// Get регистрирует GET маршрут
func (s *Server) Get(pattern string, handlerFn http.HandlerFunc) {
	s.router.Get(pattern, handlerFn)
}

// Post регистрирует POST маршрут
func (s *Server) Post(pattern string, handlerFn http.HandlerFunc) {
	s.router.Post(pattern, handlerFn)
}

// Put регистрирует PUT маршрут
func (s *Server) Put(pattern string, handlerFn http.HandlerFunc) {
	s.router.Put(pattern, handlerFn)
}

// Delete регистрирует DELETE маршрут
func (s *Server) Delete(pattern string, handlerFn http.HandlerFunc) {
	s.router.Delete(pattern, handlerFn)
}

// Handle регистрирует маршрут с HTTP методом
func (s *Server) Handle(method, pattern string, handler http.Handler) {
	s.router.Method(method, pattern, handler)
}

// Mount подключает sub-router
func (s *Server) Mount(pattern string, handler http.Handler) {
	s.router.Mount(pattern, handler)
}

// AddHealthCheck добавляет стандартный health check эндпоинт
func (s *Server) AddHealthCheck(path string) {
	s.router.Get(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{"status":"ok","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
		w.Write([]byte(response))
	})
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	s.logger.WithField("port", s.config.Port).Info("Starting HTTP server")
	return s.server.ListenAndServe()
}

// StartAsync запускает HTTP сервер асинхронно
func (s *Server) StartAsync() {
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("Failed to start server")
		}
	}()
}

// Shutdown корректно останавливает сервер
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}

// GetRouter возвращает Chi router для дополнительной настройки
func (s *Server) GetRouter() *chi.Mux {
	return s.router
}

// GetConfig возвращает конфигурацию сервера
func (s *Server) GetConfig() *Config {
	return s.config
}

// loggingMiddleware создает middleware для логирования HTTP запросов
func (s *Server) loggingMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем wrapped response writer для получения статус кода
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Получаем trace context
			span := trace.SpanFromContext(r.Context())
			traceID := span.SpanContext().TraceID().String()

			// Обрабатываем запрос
			next.ServeHTTP(ww, r)

			// Логируем запрос
			s.logger.WithFields(logrus.Fields{
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

// BasicRoutes структура для базовых маршрутов
type BasicRoutes struct {
	HealthPath  string
	MetricsPath string
	PingPath    string
}

// RegisterRoutes реализует RouteRegistrar для базовых маршрутов
func (br *BasicRoutes) RegisterRoutes(router chi.Router) {
	if br.HealthPath != "" {
		router.Get(br.HealthPath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := fmt.Sprintf(`{"status":"ok","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
			w.Write([]byte(response))
		})
	}

	if br.PingPath != "" {
		router.Get(br.PingPath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"pong"}`))
		})
	}
}

// DefaultBasicRoutes возвращает стандартные базовые маршруты
func DefaultBasicRoutes() *BasicRoutes {
	return &BasicRoutes{
		HealthPath: "/health",
		PingPath:   "/ping",
	}
}

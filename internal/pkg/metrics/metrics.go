package metrics

import (
	"github.com/Gen-Do/skeleton-service/internal/pkg/env"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config содержит настройки для метрик
type Config struct {
	ServiceName string
	Namespace   string
	Enabled     bool
}

// defaultConfig возвращает конфигурацию метрик по умолчанию
func defaultConfig() *Config {
	return &Config{
		ServiceName: env.GetString("SERVICE_NAME", "skeleton"),
		Namespace:   "service", // Можно настраивать через переменные окружения
		Enabled:     true,
	}
}

// Metrics содержит все метрики сервиса
type Metrics struct {
	config            *Config
	registry          *prometheus.Registry
	httpRequestsTotal *prometheus.CounterVec
	httpDuration      *prometheus.HistogramVec
	httpInFlight      prometheus.Gauge
}

// New настраивает и возвращает Metrics
func New() *Metrics {
	config := defaultConfig()
	if !config.Enabled {
		return &Metrics{config: config}
	}

	// Создаем собственный registry для изоляции метрик
	registry := prometheus.NewRegistry()

	// Базовые HTTP метрики
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Name:      "http_request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	httpInFlight := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Name:      "http_requests_in_flight",
			Help:      "Current number of HTTP requests being processed",
		},
	)

	// Регистрируем метрики
	registry.MustRegister(httpRequestsTotal)
	registry.MustRegister(httpDuration)
	registry.MustRegister(httpInFlight)

	// Регистрируем стандартные метрики Go
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	return &Metrics{
		config:            config,
		registry:          registry,
		httpRequestsTotal: httpRequestsTotal,
		httpDuration:      httpDuration,
		httpInFlight:      httpInFlight,
	}
}

// RegisterCollector регистрирует дополнительный коллектор метрик
func (m *Metrics) RegisterCollector(collector prometheus.Collector) error {
	if !m.config.Enabled || m.registry == nil {
		return nil
	}
	return m.registry.Register(collector)
}

// MustRegisterCollector регистрирует коллектор с паникой при ошибке
func (m *Metrics) MustRegisterCollector(collector prometheus.Collector) {
	if !m.config.Enabled || m.registry == nil {
		return
	}
	m.registry.MustRegister(collector)
}

// Handler возвращает HTTP handler для эндпоинта /metrics
func (m *Metrics) Handler() http.Handler {
	if !m.config.Enabled || m.registry == nil {
		// Возвращаем пустой handler, если метрики отключены
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
	}
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// Middleware возвращает middleware для сбора HTTP метрик
func (m *Metrics) Middleware() func(next http.Handler) http.Handler {
	if !m.config.Enabled {
		// Возвращаем no-op middleware, если метрики отключены
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Увеличиваем счетчик активных запросов
			m.httpInFlight.Inc()
			defer m.httpInFlight.Dec()

			// Создаем wrapped response writer для получения статус кода
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Обрабатываем запрос
			next.ServeHTTP(ww, r)

			// Записываем метрики
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(ww.Status())
			endpoint := r.URL.Path

			m.httpRequestsTotal.WithLabelValues(r.Method, endpoint, status).Inc()
			m.httpDuration.WithLabelValues(r.Method, endpoint).Observe(duration)
		})
	}
}

// GetRegistry возвращает Prometheus registry для регистрации дополнительных метрик
func (m *Metrics) GetRegistry() *prometheus.Registry {
	return m.registry
}

// IsEnabled возвращает true, если метрики включены
func (m *Metrics) IsEnabled() bool {
	return m.config.Enabled
}

// GetConfig возвращает конфигурацию метрик
func (m *Metrics) GetConfig() *Config {
	return m.config
}

// CreateCounter создает новую Counter метрику и регистрирует её
func (m *Metrics) CreateCounter(name, help string, labels []string) *prometheus.CounterVec {
	if !m.config.Enabled {
		return nil
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: m.config.Namespace,
			Name:      name,
			Help:      help,
		},
		labels,
	)

	m.registry.MustRegister(counter)
	return counter
}

// CreateHistogram создает новую Histogram метрику и регистрирует её
func (m *Metrics) CreateHistogram(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	if !m.config.Enabled {
		return nil
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: m.config.Namespace,
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labels,
	)

	m.registry.MustRegister(histogram)
	return histogram
}

// CreateGauge создает новую Gauge метрику и регистрирует её
func (m *Metrics) CreateGauge(name, help string, labels []string) *prometheus.GaugeVec {
	if !m.config.Enabled {
		return nil
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: m.config.Namespace,
			Name:      name,
			Help:      help,
		},
		labels,
	)

	m.registry.MustRegister(gauge)
	return gauge
}

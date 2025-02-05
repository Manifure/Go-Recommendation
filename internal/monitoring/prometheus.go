package monitoring

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Кастомные HTTP метрики.
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Кастомные метрики для Kafka.
	KafkaMessagesConsumedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_consumed_total",
			Help: "Total number of Kafka messages consumed",
		},
		[]string{"topic", "status"},
	)

	KafkaMessageProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_message_processing_duration_seconds",
			Help:    "Duration of Kafka message processing",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic"},
	)

	// Redis метрики.
	RedisRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_requests_total",
			Help: "Total number of Redis requests",
		},
		[]string{"operation", "status"},
	)
	RedisRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_request_duration_seconds",
			Help:    "Histogram of Redis request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

// Инициализация метрик.
func Init() {
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	prometheus.MustRegister(KafkaMessagesConsumedTotal)
	prometheus.MustRegister(KafkaMessageProcessingDuration)
	prometheus.MustRegister(RedisRequestsTotal)
	prometheus.MustRegister(RedisRequestDuration)
}

// Middleware для мониторинга HTTP запросов.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()

		// Обновляем метрики
		HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, http.StatusText(rec.statusCode)).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

// Используется для записи статуса ответа.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// Экспорт маршрута для метрик.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

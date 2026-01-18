package middleware

import (
	"net/http"
	"strconv"
	"time"

	"tutoring-platform/pkg/metrics"
)

// MetricsMiddleware собирает метрики для всех HTTP запросов
// Использует существующий responseWriter из logger.go
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем ResponseWriter для захвата status code
		wrapped := newResponseWriter(w)

		// Обрабатываем запрос
		next.ServeHTTP(wrapped, r)

		// Вычисляем время обработки
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.Status())

		// Записываем метрики
		metrics.HTTPRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			status,
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration)
	})
}

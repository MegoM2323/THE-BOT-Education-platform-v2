package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// DBPool интерфейс для абстракции работы с БД (для тестирования)
type DBPool interface {
	Ping(ctx context.Context) error
}

// HealthHandler обрабатывает health check запросы
type HealthHandler struct {
	db DBPool
}

// NewHealthHandler создаёт новый health handler
func NewHealthHandler(db DBPool) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthCheckResponse структура ответа health check
type HealthCheckResponse struct {
	Status   string `json:"status"`   // "healthy" или "unhealthy"
	Database string `json:"database"` // "connected" или "disconnected"
}

// HealthCheck проверяет доступность базы данных и возвращает статус сервера
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	const healthCheckTimeout = 5 * time.Second
	const slowHealthCheckMs = 1000 // Log if health check takes longer than 1 second

	// Контекст с таймаутом 5 секунд для DB ping (derived from request context for propagation)
	ctx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout)
	defer cancel()

	// Measure health check duration
	startTime := time.Now()

	// Проверка соединения с базой данных
	err := h.db.Ping(ctx)
	duration := time.Since(startTime)

	// Log slow health checks for monitoring
	if duration.Milliseconds() > int64(slowHealthCheckMs) {
		log.Warn().Int64("duration_ms", duration.Milliseconds()).Msg("Slow database health check")
	}

	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			log.Error().Dur("timeout", healthCheckTimeout).Msg("Database health check timed out")
		} else {
			log.Warn().Err(err).Msg("Database health check failed")
		}

		// База данных недоступна - возвращаем 503 Service Unavailable
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(HealthCheckResponse{
			Status:   "unhealthy",
			Database: "disconnected",
		}); err != nil {
			log.Error().Err(err).Msg("Failed to encode health check response")
		}
		return
	}

	// Всё хорошо - возвращаем 200 OK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(HealthCheckResponse{
		Status:   "healthy",
		Database: "connected",
	}); err != nil {
		log.Error().Err(err).Msg("Failed to encode health check response")
	}
}

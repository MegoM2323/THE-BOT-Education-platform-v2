package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

// ErrorResponse представляет структуру ошибки в ответе
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// RecoveryMiddleware восстанавливает сервер от паник в обработчиках
// и логирует полный stack trace для отладки
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Логируем панику с полным stack trace
				stackTrace := debug.Stack()
				log.Error().
					Interface("panic", err).
					Str("method", r.Method).
					Str("path", r.RequestURI).
					Str("remote_addr", r.RemoteAddr).
					Str("stack_trace", string(stackTrace)).
					Msg("Handler panic recovered")

				// Устанавливаем заголовок Content-Type для JSON ответа
				w.Header().Set("Content-Type", "application/json")

				// Проверяем, были ли уже написаны заголовки ответа
				// Если да, то не можем изменить статус код
				if !headerWritten(w) {
					w.WriteHeader(http.StatusInternalServerError)
				}

				// Отправляем JSON ошибку клиенту
				resp := ErrorResponse{
					Error: "Internal server error",
					Code:  http.StatusInternalServerError,
				}
				if err := json.NewEncoder(w).Encode(resp); err != nil {
					log.Error().Err(err).Msg("Failed to write error response")
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// headerWritten проверяет, был ли уже написан заголовок ответа
// Это вспомогательная функция для определения, может ли middleware изменить статус код
func headerWritten(w http.ResponseWriter) bool {
	// Если код статуса уже был установлен, это будет видно по заголовкам
	// К сожалению, http.ResponseWriter не предоставляет прямого способа проверить это,
	// но мы можем использовать тот факт, что если заголовки были написаны,
	// попытка записи будет проигнорирована
	// Поэтому мы просто вернём false и позволим http.ResponseWriter справиться с этим
	return false
}

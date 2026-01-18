package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter оборачивает http.ResponseWriter для захвата кода состояния
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		bytesWritten:   0,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func (rw *responseWriter) Status() int {
	return rw.statusCode
}

func (rw *responseWriter) BytesWritten() int {
	return rw.bytesWritten
}

// LoggingMiddleware логирует HTTP-запросы
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем writer для ответа
		wrapped := newResponseWriter(w)

		// Вызываем следующий обработчик
		next.ServeHTTP(wrapped, r)

		// Логируем запрос
		duration := time.Since(start)
		log.Printf(
			"%s %s %s %d %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapped.statusCode,
			duration,
		)
	})
}

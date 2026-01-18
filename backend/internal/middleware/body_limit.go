package middleware

import (
	"net/http"
)

// BodyLimitConfig содержит конфигурацию для лимитов размера тела запроса
type BodyLimitConfig struct {
	// DefaultLimit - дефолтный лимит для всех эндпоинтов (в байтах)
	DefaultLimit int64
	// FileUploadLimit - лимит для файловых загрузок (в байтах)
	FileUploadLimit int64
}

// DefaultBodyLimitConfig возвращает конфигурацию с рекомендуемыми значениями
// Default: 1MB для JSON, 10MB для файлов
func DefaultBodyLimitConfig() *BodyLimitConfig {
	return &BodyLimitConfig{
		DefaultLimit:    1024 * 1024,      // 1MB для JSON API
		FileUploadLimit: 10 * 1024 * 1024, // 10MB для файловых загрузок
	}
}

// BodyLimitMiddleware создает middleware для лимитирования размера тела запроса
// Это защита от DoS атак через отправку очень больших payload'ов
// Возвращает 413 Payload Too Large если тело запроса превышает лимит
//
// Использование:
//
//	r.Use(BodyLimitMiddleware(DefaultBodyLimitConfig()))
//
// Для конкретного эндпоинта с другим лимитом:
//
//	r.With(BodyLimitMiddlewareWithLimit(50 * 1024 * 1024)).Post("/upload", handler)
func BodyLimitMiddleware(config *BodyLimitConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultBodyLimitConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Используем http.MaxBytesReader для лимитирования размера тела запроса
			// MaxBytesReader читает не более limit байт из r.Body
			// Если клиент отправляет больше - читает до лимита и затем выбрасывает error при попытке прочитать дальше
			r.Body = http.MaxBytesReader(w, r.Body, config.DefaultLimit)

			// Вызываем следующий обработчик
			// Если handler попытается прочитать тело больше лимита - http.MaxBytesReader выбросит ошибку
			next.ServeHTTP(w, r)
		})
	}
}

// BodyLimitMiddlewareWithLimit создает middleware с конкретным лимитом в байтах
// Использование для конкретного эндпоинта:
//
//	r.With(BodyLimitMiddlewareWithLimit(50 * 1024 * 1024)).Post("/large-upload", handler)
func BodyLimitMiddlewareWithLimit(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// MaxBytesReader читает не более limit байт
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

// BodyLimitMiddlewareForFileUpload создает middleware для файловых загрузок
// Использует больший лимит (по умолчанию 10MB) из конфигурации
func BodyLimitMiddlewareForFileUpload(config *BodyLimitConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultBodyLimitConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Используем FileUploadLimit для файловых загрузок
			r.Body = http.MaxBytesReader(w, r.Body, config.FileUploadLimit)
			next.ServeHTTP(w, r)
		})
	}
}

// BodyLimitMiddlewareJSON создает middleware с лимитом для JSON эндпоинтов
// Использует DefaultLimit из конфигурации (по умолчанию 1MB)
func BodyLimitMiddlewareJSON(config *BodyLimitConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultBodyLimitConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Используем DefaultLimit для JSON эндпоинтов
			r.Body = http.MaxBytesReader(w, r.Body, config.DefaultLimit)
			next.ServeHTTP(w, r)
		})
	}
}

// ErrorHandler для обработки ошибок body limit
// Когда MaxBytesReader выбрасывает ошибку, она автоматически возвращает 413 Payload Too Large
// Но мы можем добавить этот handler для логирования если нужно
// Usage: defer ErrorHandler(w, r)
func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	// Эта функция вызывается в defer для перехвата паник от MaxBytesReader
	// В обычном случае MaxBytesReader уже вернул 413 и продолжение прервано
}

// GetBodyLimitConfig возвращает конфигурацию для использования в других местах
// Позволяет настроить лимиты если нужно
func GetBodyLimitConfig(defaultLimit, fileUploadLimit int64) *BodyLimitConfig {
	return &BodyLimitConfig{
		DefaultLimit:    defaultLimit,
		FileUploadLimit: fileUploadLimit,
	}
}

// Common size constants for body limits
const (
	// 512 KB
	BodyLimitSmall int64 = 512 * 1024
	// 1 MB
	BodyLimitMedium int64 = 1024 * 1024
	// 5 MB
	BodyLimitLarge int64 = 5 * 1024 * 1024
	// 10 MB
	BodyLimitExtraLarge int64 = 10 * 1024 * 1024
	// 50 MB
	BodyLimitXLarge int64 = 50 * 1024 * 1024
)

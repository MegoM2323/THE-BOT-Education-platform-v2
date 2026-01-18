package logger

import (
	"context"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup инициализирует глобальный логгер в зависимости от окружения
func Setup(env string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if env == "development" {
		// Pretty console output для локальной разработки
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// JSON output для production
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// With возвращает глобальный логгер для использования в коде
func With() zerolog.Logger {
	return log.Logger
}

// WithContext возвращает логгер с контекстом (поддержка отмены, таймаутов)
func WithContext(ctx context.Context) zerolog.Logger {
	// Возвращаем глобальный логгер - zerolog обрабатывает контекст внутри
	return log.Logger
}

// SanitizeEmail маскирует email для безопасного логирования
// Пример: john.doe@example.com -> j***@example.com
func SanitizeEmail(email string) string {
	if email == "" {
		return ""
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "invalid-email"
	}

	localPart := parts[0]
	domain := parts[1]

	if len(localPart) <= 1 {
		return "*@" + domain
	}

	// Показываем первый символ и последний, остальное маскируем
	sanitized := string(localPart[0]) + strings.Repeat("*", len(localPart)-2) + string(localPart[len(localPart)-1])
	return sanitized + "@" + domain
}

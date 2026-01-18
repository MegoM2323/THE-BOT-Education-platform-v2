package middleware

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// CORSConfig содержит конфигурацию CORS
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string // Заголовки, доступные JavaScript в браузере
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig возвращает конфигурацию CORS по умолчанию
// ВАЖНО: Не используется подстановочный символ (*), только явные домены
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:3002", "http://localhost:3003", "http://localhost:3004", "http://localhost:3005", "http://localhost:3006", "http://localhost:5173", "http://localhost:5174"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-CSRF-Token"},
		ExposedHeaders:   []string{"X-CSRF-Token"}, // CSRF токен должен быть доступен frontend
		AllowCredentials: true,
		MaxAge:           86400, // 24 часа
	}
}

// isOriginAllowed проверяет, разрешен ли origin, поддерживает подстановочные символы для портов
// Например: http://localhost:* разрешит любой localhost порт
// FIX: Не допускается использование * для всех доменов с credentials
func isOriginAllowed(origin string, allowedOrigins []string, allowCredentials bool) bool {
	// Проверяем, что origin не пустой и содержит валидный URL
	origin = strings.TrimSpace(origin)
	if origin == "" || origin == "null" {
		return false
	}

	// Парсим origin как URL для валидации
	parsedOrigin, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Проверяем, что есть схема и хост
	if parsedOrigin.Scheme == "" || parsedOrigin.Host == "" {
		return false
	}

	// Нормализуем origin к нижнему регистру для case-insensitive сравнения
	normalizedOrigin := strings.ToLower(origin)

	for _, allowed := range allowedOrigins {
		// Обрезаем пробелы в allowed
		allowed = strings.TrimSpace(allowed)

		// SECURITY: Не допускается подстановочный символ * с credentials
		if allowed == "*" {
			if !allowCredentials {
				// Wildcard разрешен только БЕЗ credentials
				return true
			}
			// Пропускаем этот вход - это небезопасная конфигурация
			continue
		}

		// Точное совпадение (case-insensitive)
		if strings.EqualFold(allowed, origin) {
			return true
		}

		// Поддержка подстановочного символа для портов: http://localhost:*
		if strings.Contains(allowed, ":*") {
			// Разбираем allowed pattern: http://localhost:*
			parts := strings.Split(allowed, ":*")
			if len(parts) == 2 {
				allowedPrefix := parts[0] // http://localhost
				// Проверяем, что origin начинается с того же хоста и схемы (case-insensitive)
				normalizedPrefix := strings.ToLower(allowedPrefix)
				if strings.HasPrefix(normalizedOrigin, normalizedPrefix+":") {
					// Дополнительная проверка: извлекаем полный хост (с портом) из allowed
					_, allowedHost, _ := parseOriginParts(allowedPrefix)
					_, originHost, _ := parseOriginParts(origin)

					// Сравниваем хост (без порта) с allowedHost (без порта)
					if compareHosts(originHost, allowedHost) {
						return true
					}
				}
			}
		}
	}

	return false
}

// parseOriginParts разбирает origin на схему, хост и порт
func parseOriginParts(origin string) (scheme, host string, port string) {
	parsedURL, err := url.Parse(origin)
	if err != nil {
		return "", "", ""
	}

	scheme = parsedURL.Scheme
	hostPort := parsedURL.Host

	// Разбираем хост и порт
	h, p, err := net.SplitHostPort(hostPort)
	_ = err // Игнорируем ошибку если порт не найден
	if h == "" {
		// Нет порта
		host = hostPort
		port = ""
	} else {
		host = h
		port = p
	}

	return scheme, host, port
}

// compareHosts сравнивает два хоста (домена)
func compareHosts(host1, host2 string) bool {
	// Приводим в нижний регистр для сравнения (домены case-insensitive)
	return strings.EqualFold(strings.TrimSpace(host1), strings.TrimSpace(host2))
}

// CORSMiddleware обрабатывает заголовки CORS
// FIX: T115 - Удален подстановочный символ * с credentials для безопасности
// Теперь используется явное сравнение origin с whitelist
func CORSMiddleware(config *CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// SECURITY: Проверяем, разрешен ли origin с использованием функции валидации
			// Не используется подстановочный символ * с credentials
			allowed := isOriginAllowed(origin, config.AllowedOrigins, config.AllowCredentials)

			// SECURITY: Только если origin разрешен, отражаем его обратно
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)

				// SECURITY: Credentials только если origin был явно разрешен
				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			// Остальные заголовки CORS (не привязаны к разрешению origin)
			w.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders, ", "))
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))

			// Expose-Headers позволяет JavaScript читать указанные заголовки из ответа
			// Без этого браузер блокирует доступ к X-CSRF-Token для frontend
			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", joinStrings(config.ExposedHeaders, ", "))
			}

			// Обрабатываем preflight-запросы (OPTIONS)
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func joinStrings(strings []string, separator string) string {
	result := ""
	for i, s := range strings {
		if i > 0 {
			result += separator
		}
		result += s
	}
	return result
}

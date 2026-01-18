package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIsOriginAllowed проверяет функцию валидации origin
// Таблица-driven тесты для различных сценариев
func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name             string
		origin           string
		allowedOrigins   []string
		allowCredentials bool
		expected         bool
		description      string
	}{
		// === Valid origins ===
		{
			name:             "exact_match",
			origin:           "http://localhost:5173",
			allowedOrigins:   []string{"http://localhost:5173"},
			allowCredentials: true,
			expected:         true,
			description:      "Origin точно совпадает с allowed",
		},
		{
			name:             "multiple_allowed_first",
			origin:           "http://localhost:3000",
			allowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
			allowCredentials: true,
			expected:         true,
			description:      "Origin в списке разрешенных",
		},
		{
			name:             "multiple_allowed_second",
			origin:           "http://localhost:5173",
			allowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
			allowCredentials: true,
			expected:         true,
			description:      "Origin второй в списке разрешенных",
		},
		{
			name:             "wildcard_port_exact",
			origin:           "http://localhost:5173",
			allowedOrigins:   []string{"http://localhost:*"},
			allowCredentials: false,
			expected:         true,
			description:      "Origin соответствует подстановке :* (без credentials)",
		},
		{
			name:             "wildcard_port_different",
			origin:           "http://localhost:3000",
			allowedOrigins:   []string{"http://localhost:*"},
			allowCredentials: false,
			expected:         true,
			description:      "Другой порт localhost соответствует :*",
		},
		{
			name:             "case_insensitive_host",
			origin:           "http://LOCALHOST:5173",
			allowedOrigins:   []string{"http://localhost:5173"},
			allowCredentials: true,
			expected:         true,
			description:      "Домены case-insensitive",
		},

		// === Invalid origins (no credentials) ===
		{
			name:             "not_allowed_origin",
			origin:           "http://attacker.com",
			allowedOrigins:   []string{"http://localhost:5173"},
			allowCredentials: true,
			expected:         false,
			description:      "Origin не в списке разрешенных",
		},
		{
			name:             "empty_origin",
			origin:           "",
			allowedOrigins:   []string{"http://localhost:5173"},
			allowCredentials: true,
			expected:         false,
			description:      "Пустой origin отклоняется",
		},
		{
			name:             "null_origin",
			origin:           "null",
			allowedOrigins:   []string{"http://localhost:5173"},
			allowCredentials: true,
			expected:         false,
			description:      "null origin отклоняется",
		},
		{
			name:             "invalid_url",
			origin:           "not a url",
			allowedOrigins:   []string{"http://localhost:5173"},
			allowCredentials: true,
			expected:         false,
			description:      "Невалидный URL отклоняется",
		},
		{
			name:             "no_scheme",
			origin:           "localhost:5173",
			allowedOrigins:   []string{"http://localhost:5173"},
			allowCredentials: true,
			expected:         false,
			description:      "Origin без схемы отклоняется",
		},
		{
			name:             "no_host",
			origin:           "http://",
			allowedOrigins:   []string{"http://localhost:5173"},
			allowCredentials: true,
			expected:         false,
			description:      "Origin без хоста отклоняется",
		},

		// === SECURITY: Wildcard with credentials ===
		{
			name:             "wildcard_with_credentials",
			origin:           "http://anything.com",
			allowedOrigins:   []string{"*"},
			allowCredentials: true,
			expected:         false,
			description:      "Подстановка * с credentials ОТКЛОНЯЕТСЯ (уязвимость!)",
		},
		{
			name:             "wildcard_without_credentials",
			origin:           "http://anything.com",
			allowedOrigins:   []string{"*"},
			allowCredentials: false,
			expected:         true,
			description:      "Подстановка * БЕЗ credentials РАЗРЕШЕНА",
		},

		// === Wildcard port with credentials ===
		{
			name:             "wildcard_port_with_credentials",
			origin:           "http://localhost:8000",
			allowedOrigins:   []string{"http://localhost:*"},
			allowCredentials: true,
			expected:         true,
			description:      "Подстановка :* с localhost и credentials разрешена",
		},

		// === Protocol mismatch ===
		{
			name:             "https_vs_http",
			origin:           "https://localhost:5173",
			allowedOrigins:   []string{"http://localhost:5173"},
			allowCredentials: true,
			expected:         false,
			description:      "HTTPS и HTTP - разные origins",
		},
		{
			name:             "http_vs_https",
			origin:           "http://localhost:5173",
			allowedOrigins:   []string{"https://localhost:5173"},
			allowCredentials: true,
			expected:         false,
			description:      "HTTP и HTTPS - разные origins",
		},

		// === Port mismatch ===
		{
			name:             "port_mismatch",
			origin:           "http://localhost:8000",
			allowedOrigins:   []string{"http://localhost:5173"},
			allowCredentials: true,
			expected:         false,
			description:      "Разные порты - разные origins",
		},

		// === Trailing whitespace ===
		{
			name:             "origin_with_whitespace",
			origin:           "http://localhost:5173",
			allowedOrigins:   []string{"  http://localhost:5173  "},
			allowCredentials: true,
			expected:         true,
			description:      "Origin с пробелами обрезается",
		},

		// === Subdomain ===
		{
			name:             "subdomain_mismatch",
			origin:           "http://api.example.com:5173",
			allowedOrigins:   []string{"http://example.com:5173"},
			allowCredentials: true,
			expected:         false,
			description:      "Поддомены - разные origins",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowedOrigins, tt.allowCredentials)
			if result != tt.expected {
				t.Errorf("%s: origin=%q, expected=%v, got=%v", tt.description, tt.origin, tt.expected, result)
			}
		})
	}
}

// TestCORSMiddlewareHeaders проверяет корректность CORS заголовков
// Таблица-driven тесты для различных конфигураций
func TestCORSMiddlewareHeaders(t *testing.T) {
	tests := []struct {
		name                     string
		origin                   string
		config                   *CORSConfig
		expectedOriginHeader     string
		expectedCredentialsValue string
		shouldHaveCredentials    bool
		description              string
	}{
		{
			name:                     "allowed_origin_with_credentials",
			origin:                   "http://localhost:5173",
			config:                   &CORSConfig{AllowedOrigins: []string{"http://localhost:5173"}, AllowCredentials: true, AllowedMethods: []string{"GET", "POST"}, AllowedHeaders: []string{"Content-Type"}, MaxAge: 86400},
			expectedOriginHeader:     "http://localhost:5173",
			expectedCredentialsValue: "true",
			shouldHaveCredentials:    true,
			description:              "Разрешенный origin - возвращает credentials",
		},
		{
			name:                     "denied_origin_no_credentials",
			origin:                   "http://attacker.com",
			config:                   &CORSConfig{AllowedOrigins: []string{"http://localhost:5173"}, AllowCredentials: true, AllowedMethods: []string{"GET", "POST"}, AllowedHeaders: []string{"Content-Type"}, MaxAge: 86400},
			expectedOriginHeader:     "",
			expectedCredentialsValue: "",
			shouldHaveCredentials:    false,
			description:              "Запрещенный origin - НЕ возвращает credentials",
		},
		{
			name:                     "empty_origin_no_response",
			origin:                   "",
			config:                   &CORSConfig{AllowedOrigins: []string{"http://localhost:5173"}, AllowCredentials: true, AllowedMethods: []string{"GET", "POST"}, AllowedHeaders: []string{"Content-Type"}, MaxAge: 86400},
			expectedOriginHeader:     "",
			expectedCredentialsValue: "",
			shouldHaveCredentials:    false,
			description:              "Пустой origin - НЕ возвращает CORS заголовки",
		},
		{
			name:                     "wildcard_without_credentials",
			origin:                   "http://any-origin.com",
			config:                   &CORSConfig{AllowedOrigins: []string{"*"}, AllowCredentials: false, AllowedMethods: []string{"GET", "POST"}, AllowedHeaders: []string{"Content-Type"}, MaxAge: 86400},
			expectedOriginHeader:     "http://any-origin.com",
			expectedCredentialsValue: "",
			shouldHaveCredentials:    false,
			description:              "Подстановка * без credentials - разрешает любой origin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем простой next handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Оборачиваем в CORS middleware
			handler := CORSMiddleware(tt.config)(nextHandler)

			// Создаем тестовый запрос
			req := httptest.NewRequest("OPTIONS", "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			// Выполняем запрос
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// Проверяем заголовок Access-Control-Allow-Origin
			actualOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if actualOrigin != tt.expectedOriginHeader {
				t.Errorf("%s: expected origin=%q, got=%q", tt.description, tt.expectedOriginHeader, actualOrigin)
			}

			// Проверяем наличие credentials заголовка
			credentialsValue := w.Header().Get("Access-Control-Allow-Credentials")
			if tt.shouldHaveCredentials {
				if credentialsValue != "true" {
					t.Errorf("%s: expected Credentials=true, got=%q", tt.description, credentialsValue)
				}
			} else {
				if credentialsValue != "" {
					t.Errorf("%s: expected NO Credentials header, got=%q", tt.description, credentialsValue)
				}
			}

			// Проверяем стандартные CORS заголовки всегда присутствуют
			if w.Header().Get("Access-Control-Allow-Methods") == "" {
				t.Errorf("%s: missing Access-Control-Allow-Methods header", tt.description)
			}
		})
	}
}

// TestCORSPreflight проверяет обработку preflight запросов
func TestCORSPreflight(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           86400,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := CORSMiddleware(config)(nextHandler)

	// OPTIONS preflight запрос
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Проверяем, что preflight возвращает 200 без тела
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "" {
		t.Errorf("expected empty body for preflight, got %q", w.Body.String())
	}

	// Проверяем заголовки
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Error("missing or wrong Access-Control-Allow-Origin header")
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("missing or wrong Access-Control-Allow-Credentials header")
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("missing Access-Control-Allow-Methods header")
	}

	if w.Header().Get("Access-Control-Max-Age") != "86400" {
		t.Error("wrong or missing Access-Control-Max-Age header")
	}
}

// TestCORSExposeHeaders проверяет что Access-Control-Expose-Headers устанавливается
func TestCORSExposeHeaders(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-CSRF-Token"},
		ExposedHeaders:   []string{"X-CSRF-Token"}, // Важно для CSRF
		AllowCredentials: true,
		MaxAge:           86400,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSMiddleware(config)(nextHandler)

	// Запрос с валидным origin
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Проверяем что Access-Control-Expose-Headers установлен
	exposeHeaders := w.Header().Get("Access-Control-Expose-Headers")
	if exposeHeaders != "X-CSRF-Token" {
		t.Errorf("expected Access-Control-Expose-Headers 'X-CSRF-Token', got %q", exposeHeaders)
	}
}

// TestCORSActualRequest проверяет обработку реальных (не preflight) запросов
func TestCORSActualRequest(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           86400,
	}

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := CORSMiddleware(config)(nextHandler)

	// Реальный GET запрос
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Проверяем, что next handler был вызван
	if !handlerCalled {
		t.Error("next handler was not called")
	}

	// Проверяем статус
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Проверяем, что тело сохранено
	if w.Body.String() != "OK" {
		t.Errorf("expected body 'OK', got %q", w.Body.String())
	}

	// Проверяем CORS заголовки
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Error("missing or wrong Access-Control-Allow-Origin header")
	}
}

// TestSecurityWildcardWithCredentials проверяет безопасность - запрещает * с credentials
func TestSecurityWildcardWithCredentials(t *testing.T) {
	// ОПАСНАЯ конфигурация: wildcard с credentials
	dangerousConfig := &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true, // DANGER!
		MaxAge:           86400,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORSMiddleware(dangerousConfig)(nextHandler)

	// Попытка запроса с произвольным origin
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://attacker.com")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Проверяем, что Origin НЕ возвращается (блокировка уязвимости)
	actualOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if actualOrigin != "" {
		t.Errorf("SECURITY FAIL: wildcard * with credentials should be blocked. Got origin=%q", actualOrigin)
	}

	// Проверяем, что credentials НЕ возвращаются
	credentials := w.Header().Get("Access-Control-Allow-Credentials")
	if credentials != "" {
		t.Errorf("SECURITY FAIL: credentials should not be returned for unauthorized origin")
	}
}

// TestParseOriginParts проверяет разбор parts из origin
func TestParseOriginParts(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		expectedScheme string
		expectedHost   string
		expectedPort   string
	}{
		{
			name:           "full_url",
			origin:         "http://localhost:5173",
			expectedScheme: "http",
			expectedHost:   "localhost",
			expectedPort:   "5173",
		},
		{
			name:           "https_url",
			origin:         "https://example.com:443",
			expectedScheme: "https",
			expectedHost:   "example.com",
			expectedPort:   "443",
		},
		{
			name:           "no_port",
			origin:         "http://localhost",
			expectedScheme: "http",
			expectedHost:   "localhost",
			expectedPort:   "",
		},
		{
			name:           "invalid_url",
			origin:         "not a url",
			expectedScheme: "",
			expectedHost:   "",
			expectedPort:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme, host, port := parseOriginParts(tt.origin)
			if scheme != tt.expectedScheme {
				t.Errorf("expected scheme=%q, got=%q", tt.expectedScheme, scheme)
			}
			if host != tt.expectedHost {
				t.Errorf("expected host=%q, got=%q", tt.expectedHost, host)
			}
			if port != tt.expectedPort {
				t.Errorf("expected port=%q, got=%q", tt.expectedPort, port)
			}
		})
	}
}

// TestCompareHosts проверяет сравнение хостов (case-insensitive)
func TestCompareHosts(t *testing.T) {
	tests := []struct {
		name     string
		host1    string
		host2    string
		expected bool
	}{
		{"same", "localhost", "localhost", true},
		{"case_insensitive", "LOCALHOST", "localhost", true},
		{"mixed_case", "LocalHost", "localhost", true},
		{"with_whitespace", "  localhost  ", "localhost", true},
		{"different", "localhost", "example.com", false},
		{"different_case", "LOCALHOST", "example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareHosts(tt.host1, tt.host2)
			if result != tt.expected {
				t.Errorf("compareHosts(%q, %q): expected=%v, got=%v", tt.host1, tt.host2, tt.expected, result)
			}
		})
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

// TestGetIPAddressSecure_TrustedProxyWithValidXForwardedFor проверяет безопасное извлечение от доверенного прокси
func TestGetIPAddressSecure_TrustedProxyWithValidXForwardedFor(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		trustedProxies map[string]bool
		expectedIP     string
	}{
		{
			name:           "trusted proxy with single client IP",
			remoteAddr:     "10.0.0.1:8080",
			xForwardedFor:  "203.0.113.100",
			trustedProxies: map[string]bool{"10.0.0.1": true},
			expectedIP:     "203.0.113.100",
		},
		{
			name:           "trusted proxy with multiple IPs in chain",
			remoteAddr:     "10.0.0.2:8080",
			xForwardedFor:  "203.0.113.100, 10.0.0.1",
			trustedProxies: map[string]bool{"10.0.0.1": true, "10.0.0.2": true},
			expectedIP:     "203.0.113.100",
		},
		{
			name:           "trusted proxy chain: client, untrusted, trusted",
			remoteAddr:     "10.0.0.1:8080",
			xForwardedFor:  "203.0.113.100, 203.0.113.200, 10.0.0.1",
			trustedProxies: map[string]bool{"10.0.0.1": true},
			expectedIP:     "203.0.113.200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("X-Forwarded-For", tt.xForwardedFor)

			ip := getIPAddressSecure(req, tt.trustedProxies)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

// TestGetIPAddressSecure_UntrustedProxyIgnoresXForwardedFor проверяет что untrusted proxy игнорируется
func TestGetIPAddressSecure_UntrustedProxyIgnoresXForwardedFor(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		trustedProxies map[string]bool
		expectedIP     string
	}{
		{
			name:           "untrusted proxy IP ignored",
			remoteAddr:     "203.0.113.50:12345",
			xForwardedFor:  "203.0.113.100",
			trustedProxies: map[string]bool{"10.0.0.1": true},
			expectedIP:     "203.0.113.50",
		},
		{
			name:           "no trusted proxies configured",
			remoteAddr:     "10.0.0.1:8080",
			xForwardedFor:  "203.0.113.100",
			trustedProxies: map[string]bool{},
			expectedIP:     "10.0.0.1",
		},
		{
			name:           "attacker tries to spoof via X-Forwarded-For",
			remoteAddr:     "192.168.1.50:12345",
			xForwardedFor:  "admin.internal.network.fake",
			trustedProxies: map[string]bool{"10.0.0.1": true},
			expectedIP:     "192.168.1.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("X-Forwarded-For", tt.xForwardedFor)

			ip := getIPAddressSecure(req, tt.trustedProxies)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

// TestGetIPAddressSecure_InvalidIPsInXForwardedFor проверяет обработку невалидных IP в заголовке
func TestGetIPAddressSecure_InvalidIPsInXForwardedFor(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		trustedProxies map[string]bool
		expectedIP     string
	}{
		{
			name:           "invalid IP in X-Forwarded-For is skipped",
			remoteAddr:     "10.0.0.1:8080",
			xForwardedFor:  "invalid.ip, 203.0.113.100",
			trustedProxies: map[string]bool{"10.0.0.1": true},
			expectedIP:     "203.0.113.100",
		},
		{
			name:           "all IPs invalid falls back to directIP",
			remoteAddr:     "10.0.0.1:8080",
			xForwardedFor:  "invalid.one, invalid.two",
			trustedProxies: map[string]bool{"10.0.0.1": true},
			expectedIP:     "10.0.0.1",
		},
		{
			name:           "SQL injection attempt blocked",
			remoteAddr:     "10.0.0.1:8080",
			xForwardedFor:  "'; DROP TABLE users; --",
			trustedProxies: map[string]bool{"10.0.0.1": true},
			expectedIP:     "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("X-Forwarded-For", tt.xForwardedFor)

			ip := getIPAddressSecure(req, tt.trustedProxies)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

// TestGetIPAddressSecure_XRealIPFallback проверяет fallback на X-Real-IP
func TestGetIPAddressSecure_XRealIPFallback(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		xRealIP        string
		trustedProxies map[string]bool
		expectedIP     string
	}{
		{
			name:           "X-Real-IP used when X-Forwarded-For not present",
			remoteAddr:     "10.0.0.1:8080",
			xRealIP:        "203.0.113.100",
			trustedProxies: map[string]bool{"10.0.0.1": true},
			expectedIP:     "203.0.113.100",
		},
		{
			name:           "X-Real-IP ignored when untrusted proxy",
			remoteAddr:     "203.0.113.50:12345",
			xRealIP:        "203.0.113.100",
			trustedProxies: map[string]bool{"10.0.0.1": true},
			expectedIP:     "203.0.113.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("X-Real-IP", tt.xRealIP)

			ip := getIPAddressSecure(req, tt.trustedProxies)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

// TestRateLimiterWithTrustedProxies_SpoofingProtection проверяет защиту от спуфинга в rate limiter
func TestRateLimiterWithTrustedProxies_SpoofingProtection(t *testing.T) {
	// Создаем rate limiter с доверенным прокси
	trustedProxies := []string{"10.0.0.1"}
	limiter := NewIPRateLimiterWithProxies(rate.Every(time.Second), 1, trustedProxies)

	t.Run("spoofed IP does not bypass rate limits", func(t *testing.T) {
		middleware := RateLimitMiddleware(limiter)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrappedHandler := middleware(handler)

		// Попытка 1: прямое соединение от 203.0.113.50 - проходит
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "203.0.113.50:12345"
		req1.Header.Set("X-Forwarded-For", "admin.internal")
		w1 := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Попытка 2: попытка спуфить другой IP в X-Forwarded-For - блокируется
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "203.0.113.50:12345"
		req2.Header.Set("X-Forwarded-For", "203.0.113.100")
		w2 := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	})

	t.Run("trusted proxy X-Forwarded-For is respected", func(t *testing.T) {
		middleware := RateLimitMiddleware(limiter)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrappedHandler := middleware(handler)

		// Запрос от доверенного прокси с реальным IP клиента - проходит
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "10.0.0.1:8080"
		req.Header.Set("X-Forwarded-For", "203.0.113.100")
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Второй запрос от ТОГО ЖЕ клиента (203.0.113.100) блокируется
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "10.0.0.1:8080"
		req2.Header.Set("X-Forwarded-For", "203.0.113.100")
		w2 := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	})
}

// TestNewIPRateLimiterWithProxies_ProxyNormalization проверяет нормализацию IP прокси
func TestNewIPRateLimiterWithProxies_ProxyNormalization(t *testing.T) {
	tests := []struct {
		name        string
		proxies     []string
		testIP      string
		shouldExist bool
	}{
		{
			name:        "plain IP address",
			proxies:     []string{"10.0.0.1"},
			testIP:      "10.0.0.1",
			shouldExist: true,
		},
		{
			name:        "IP with port",
			proxies:     []string{"10.0.0.1:8080"},
			testIP:      "10.0.0.1",
			shouldExist: true,
		},
		{
			name:        "invalid IP is skipped",
			proxies:     []string{"invalid.ip"},
			testIP:      "invalid.ip",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewIPRateLimiterWithProxies(rate.Every(time.Second), 1, tt.proxies)
			exists := limiter.trustedProxies[tt.testIP]
			if tt.shouldExist {
				assert.True(t, exists, "proxy should be in trusted list")
			} else {
				assert.False(t, exists, "invalid proxy should not be in trusted list")
			}
		})
	}
}

// TestIsValidIP проверяет валидацию IP адресов
func TestIsValidIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		isValid bool
	}{
		{"valid IPv4", "192.168.1.1", true},
		{"valid IPv6", "2001:db8::1", true},
		{"localhost", "127.0.0.1", true},
		{"empty string", "", false},
		{"invalid IP", "invalid.ip", false},
		{"SQL injection", "'; DROP TABLE users; --", false},
		{"path traversal", "../../../etc/passwd", false},
		{"CIDR notation", "192.168.1.0/24", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidIP(tt.ip)
			assert.Equal(t, tt.isValid, result)
		})
	}
}

// TestIsTrustedProxy проверяет определение trusted proxy
func TestIsTrustedProxy(t *testing.T) {
	trustedProxies := map[string]bool{
		"10.0.0.1": true,
		"10.0.0.2": true,
	}

	tests := []struct {
		name      string
		ip        string
		isTrusted bool
	}{
		{"trusted proxy", "10.0.0.1", true},
		{"untrusted proxy", "203.0.113.100", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTrustedProxy(tt.ip, trustedProxies)
			assert.Equal(t, tt.isTrusted, result)
		})
	}
}

// TestGetIPAddressBackwardCompatibility проверяет backward compatibility старой функции
func TestGetIPAddressBackwardCompatibility(t *testing.T) {
	t.Run("old function falls back to RemoteAddr for untrusted proxies", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "203.0.113.50:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.100")

		// Старая функция без trusted proxies возвращает RemoteAddr
		ip := getIPAddress(req)
		assert.Equal(t, "203.0.113.50", ip)
	})
}

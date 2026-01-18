package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

// TestIPRateLimiter_Basic проверяет базовую функциональность rate limiter
func TestIPRateLimiter_Basic(t *testing.T) {
	// 2 запроса в секунду, burst = 2
	limiter := NewIPRateLimiter(rate.Every(time.Second/2), 2)

	ip := "192.168.1.100"

	// Первые 2 запроса должны пройти (burst)
	for i := 0; i < 2; i++ {
		l := limiter.GetLimiter(ip)
		assert.True(t, l.Allow(), "Request %d should be allowed (burst)", i+1)
	}

	// 3-й запрос должен быть заблокирован
	l := limiter.GetLimiter(ip)
	assert.False(t, l.Allow(), "3rd request should be blocked")
}

// TestIPRateLimiter_MultipleIPs проверяет изоляцию лимитов между IP
func TestIPRateLimiter_MultipleIPs(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 1)

	ip1 := "192.168.1.100"
	ip2 := "192.168.1.200"

	// IP1: первый запрос проходит
	l1 := limiter.GetLimiter(ip1)
	assert.True(t, l1.Allow(), "IP1 first request should pass")

	// IP1: второй запрос блокируется
	l1 = limiter.GetLimiter(ip1)
	assert.False(t, l1.Allow(), "IP1 second request should be blocked")

	// IP2: первый запрос проходит (независимый лимит)
	l2 := limiter.GetLimiter(ip2)
	assert.True(t, l2.Allow(), "IP2 first request should pass")
}

// TestIPRateLimiter_CleanupExpired проверяет что cleanup удаляет только EXPIRED entries
// и сохраняет активные limiters
func TestIPRateLimiter_CleanupExpired(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 1)
	// Устанавливаем очень короткий TTL для тестирования
	limiter.ttl = 100 * time.Millisecond

	ipActive := "192.168.1.100"
	ipExpired := "192.168.1.200"

	// Создаем limiter для двух IP
	l1 := limiter.GetLimiter(ipActive)
	assert.NotNil(t, l1)

	l2 := limiter.GetLimiter(ipExpired)
	assert.NotNil(t, l2)

	// Проверяем что оба IP есть в map
	limiter.mu.Lock()
	assert.Equal(t, 2, len(limiter.ips), "Should have 2 entries")
	limiter.mu.Unlock()

	// Ждем пока TTL для ipExpired истечет
	time.Sleep(150 * time.Millisecond)

	// Обновляем lastAccessed для ipActive (сделав его "активным")
	limiter.GetLimiter(ipActive)

	// Теперь ipExpired должен быть expired, а ipActive - активным
	// Очищаем
	limiter.CleanupExpired()

	// Проверяем что только ipExpired удален
	limiter.mu.Lock()
	_, existsActive := limiter.ips[ipActive]
	_, existsExpired := limiter.ips[ipExpired]
	limiter.mu.Unlock()

	assert.True(t, existsActive, "Active IP should still exist")
	assert.False(t, existsExpired, "Expired IP should be removed")

	// Очищаем горутину
	limiter.Stop()
}

// TestRateLimitMiddleware_AllowsRequests проверяет middleware пропускает запросы
func TestRateLimitMiddleware_AllowsRequests(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 5)
	middleware := RateLimitMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := middleware(handler)

	// Первые 5 запросов должны пройти
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
		assert.Equal(t, "OK", w.Body.String())
	}
}

// TestRateLimitMiddleware_BlocksRequests проверяет middleware блокирует запросы
func TestRateLimitMiddleware_BlocksRequests(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 2)
	middleware := RateLimitMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrappedHandler := middleware(handler)

	// Первые 2 запроса проходят
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 3-й запрос блокируется
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "3rd request should be blocked")
	assert.Contains(t, w.Body.String(), "Rate limit exceeded")
}

// TestGetIPAddress_RemoteAddr проверяет извлечение IP из RemoteAddr
func TestGetIPAddress_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	ip := getIPAddress(req)
	assert.Equal(t, "192.168.1.100", ip)
}

// TestGetIPAddress_XForwardedFor проверяет что X-Forwarded-For игнорируется без доверенного прокси
func TestGetIPAddress_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:8080" // Untrusted proxy IP (not in trusted list)
	req.Header.Set("X-Forwarded-For", "203.0.113.100")

	// SECURITY: Без доверенного прокси getIPAddress должен вернуть RemoteAddr
	ip := getIPAddress(req)
	assert.Equal(t, "10.0.0.1", ip, "Should ignore X-Forwarded-For from untrusted proxy")
}

// TestGetIPAddress_XRealIP проверяет что X-Real-IP игнорируется без доверенного прокси
func TestGetIPAddress_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:8080" // Untrusted proxy IP
	req.Header.Set("X-Real-IP", "203.0.113.200")

	// SECURITY: Без доверенного прокси getIPAddress должен вернуть RemoteAddr
	ip := getIPAddress(req)
	assert.Equal(t, "10.0.0.1", ip, "Should ignore X-Real-IP from untrusted proxy")
}

// TestGetIPAddress_Priority проверяет что заголовки игнорируются без доверенного прокси
func TestGetIPAddress_Priority(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "10.0.0.1:8080" // Untrusted proxy IP
	req.Header.Set("X-Forwarded-For", "203.0.113.100")
	req.Header.Set("X-Real-IP", "203.0.113.200")

	// SECURITY: Без доверенного прокси оба заголовка игнорируются
	ip := getIPAddress(req)
	assert.Equal(t, "10.0.0.1", ip, "Should ignore both X-Forwarded-For and X-Real-IP from untrusted proxy")
}

// TestLoginRateLimiter_Config проверяет конфигурацию login rate limiter
func TestLoginRateLimiter_Config(t *testing.T) {
	limiter := LoginRateLimiter()
	assert.NotNil(t, limiter)

	// Проверяем что можно сделать 10 запросов (burst=10 для LoginRateLimiter)
	ip := "192.168.1.100"
	for i := 0; i < 10; i++ {
		l := limiter.GetLimiter(ip)
		assert.True(t, l.Allow(), "Login attempt %d should be allowed", i+1)
	}

	// 11-й запрос блокируется
	l := limiter.GetLimiter(ip)
	assert.False(t, l.Allow(), "11th login attempt should be blocked")
}

// TestTrialRequestRateLimiter_Config проверяет конфигурацию trial request rate limiter
func TestTrialRequestRateLimiter_Config(t *testing.T) {
	limiter := TrialRequestRateLimiter()
	assert.NotNil(t, limiter)

	// Проверяем что можно сделать 5 запросов (burst)
	ip := "192.168.1.100"
	for i := 0; i < 5; i++ {
		l := limiter.GetLimiter(ip)
		assert.True(t, l.Allow(), "Trial request %d should be allowed", i+1)
	}

	// 6-й запрос блокируется
	l := limiter.GetLimiter(ip)
	assert.False(t, l.Allow(), "6th trial request should be blocked")
}

// TestRateLimitMiddleware_DifferentIPs проверяет изоляцию между IP
func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 1)
	middleware := RateLimitMiddleware(limiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	// IP1: первый запрос проходит
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	w1 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code, "IP1 first request should pass")

	// IP1: второй запрос блокируется
	req1b := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1b.RemoteAddr = "192.168.1.100:12345"
	w1b := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w1b, req1b)
	assert.Equal(t, http.StatusTooManyRequests, w1b.Code, "IP1 second request should be blocked")

	// IP2: первый запрос проходит (независимый лимит)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.200:54321"
	w2 := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code, "IP2 first request should pass")

	// Очищаем горутину
	limiter.Stop()
}

// TestIPRateLimiter_MemoryLeakPrevention проверяет что cleanup горутина запущена
// и работает, предотвращая утечку памяти
func TestIPRateLimiter_MemoryLeakPrevention(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 1)
	limiter.ttl = 100 * time.Millisecond

	// Создаем много IP entries
	initialCount := 100
	for i := 0; i < initialCount; i++ {
		ip := "192.168.1." + string(rune(i%255))
		limiter.GetLimiter(ip)
	}

	// Проверяем что все созданы
	limiter.mu.Lock()
	assert.Equal(t, initialCount, len(limiter.ips), "Should have created all entries")
	limiter.mu.Unlock()

	// Ждем пока TTL истечет
	time.Sleep(150 * time.Millisecond)

	// Запускаем cleanup вручную (горутина тоже должна работать)
	limiter.CleanupExpired()

	// Проверяем что все старые entries удалены
	limiter.mu.Lock()
	finalCount := len(limiter.ips)
	limiter.mu.Unlock()

	assert.Equal(t, 0, finalCount, "All expired entries should be removed")

	limiter.Stop()
}

// TestIPRateLimiter_CleanupPreservesActiveEntries проверяет что cleanup сохраняет активные entries
func TestIPRateLimiter_CleanupPreservesActiveEntries(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 1)
	limiter.ttl = 100 * time.Millisecond

	oldIP := "192.168.1.100"
	newIP := "192.168.1.200"

	// Создаем старый entry
	limiter.GetLimiter(oldIP)
	time.Sleep(150 * time.Millisecond)

	// Создаем новый entry (свежий)
	limiter.GetLimiter(newIP)

	// Очищаем
	limiter.CleanupExpired()

	// Проверяем результаты
	limiter.mu.Lock()
	_, hasOld := limiter.ips[oldIP]
	_, hasNew := limiter.ips[newIP]
	limiter.mu.Unlock()

	assert.False(t, hasOld, "Old entry should be removed")
	assert.True(t, hasNew, "New entry should be preserved")

	limiter.Stop()
}

// TestIPRateLimiter_LastAccessedUpdated проверяет что lastAccessed обновляется при GetLimiter
func TestIPRateLimiter_LastAccessedUpdated(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 1)
	ip := "192.168.1.100"

	// Создаем entry
	limiter.GetLimiter(ip)

	limiter.mu.Lock()
	entry, exists := limiter.ips[ip]
	assert.True(t, exists, "Entry should exist")
	firstAccess := entry.lastAccessed
	limiter.mu.Unlock()

	// Ждем немного
	time.Sleep(50 * time.Millisecond)

	// Обращаемся к limiter снова
	limiter.GetLimiter(ip)

	limiter.mu.Lock()
	entry, _ = limiter.ips[ip]
	secondAccess := entry.lastAccessed
	limiter.mu.Unlock()

	// Проверяем что время обновилось
	assert.True(t, secondAccess.After(firstAccess), "lastAccessed should be updated")

	limiter.Stop()
}

// TestIPRateLimiter_CleanupGoroutineStarts проверяет что cleanup горутина запускается автоматически
func TestIPRateLimiter_CleanupGoroutineStarts(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 1)
	assert.NotNil(t, limiter.stopChan, "stopChan should be initialized")

	// Создаем entry
	ip := "192.168.1.100"
	limiter.GetLimiter(ip)

	limiter.mu.Lock()
	_, exists := limiter.ips[ip]
	limiter.mu.Unlock()
	assert.True(t, exists, "Entry should exist")

	// Останавливаем лимитер
	limiter.Stop()

	// Проверяем что Stop не вызовет паник
	assert.NotPanics(t, func() {
		limiter.Stop()
	}, "Multiple Stop() calls should not panic")
}

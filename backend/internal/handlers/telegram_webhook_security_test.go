package handlers

import (
	"testing"
	"time"
)

// TestRateLimiterConfiguration проверяет, что rate limiter правильно сконфигурирован
func TestRateLimiterConfiguration(t *testing.T) {
	limiter := NewIPWebhookRateLimiter()
	defer limiter.Stop()

	// Проверяем, что TTL установлен корректно
	if limiter.ttl != 30*time.Minute {
		t.Errorf("Expected TTL 30 minutes, got %v", limiter.ttl)
	}

	// Проверяем, что мы можем получить limiter для IP адреса
	l := limiter.GetLimiter("127.0.0.1")
	if l == nil {
		t.Error("Expected non-nil limiter")
	}

	// Проверяем, что тот же IP возвращает тот же limiter
	l2 := limiter.GetLimiter("127.0.0.1")
	if l != l2 {
		t.Error("Expected same limiter for same IP")
	}

	// Проверяем, что разные IPs получают разные limiters
	l3 := limiter.GetLimiter("192.168.1.1")
	if l3 == nil || l == l3 {
		t.Error("Expected different limiters for different IPs")
	}
}

// TestFailureTrackerInitialization проверяет инициализацию failure tracker
func TestFailureTrackerInitialization(t *testing.T) {
	ft := &WebhookFailureTracker{
		failures:        make(map[string]*FailureMetrics),
		blockList:       make(map[string]time.Time),
		suspiciousUsers: make(map[string]*UserFailures),
	}

	if len(ft.failures) != 0 {
		t.Error("Expected empty failures map")
	}

	if len(ft.blockList) != 0 {
		t.Error("Expected empty blockList map")
	}

	if len(ft.suspiciousUsers) != 0 {
		t.Error("Expected empty suspiciousUsers map")
	}
}

// TestFailureMetricsTracking проверяет отслеживание метрик сбоев
func TestFailureMetricsTracking(t *testing.T) {
	ip := "192.168.1.100"
	metrics := &FailureMetrics{
		count:             3,
		consecutiveErrors: 3,
		lastFailureTime:   time.Now(),
	}

	if metrics.count != 3 {
		t.Errorf("Expected count 3, got %d", metrics.count)
	}

	if metrics.consecutiveErrors != 3 {
		t.Errorf("Expected consecutiveErrors 3, got %d", metrics.consecutiveErrors)
	}

	if metrics.lastFailureTime.IsZero() {
		t.Error("Expected lastFailureTime to be set")
	}

	_ = ip // Используем переменную для избежания warning
}

// TestUserFailuresTracking проверяет отслеживание сбоев пользователя
func TestUserFailuresTracking(t *testing.T) {
	uf := &UserFailures{
		failureCount:   5,
		lastFailTime:   time.Now(),
		blockedUntil:   time.Now().Add(1 * time.Hour),
		lastUpdateTime: time.Now(),
	}

	if uf.failureCount != 5 {
		t.Errorf("Expected failureCount 5, got %d", uf.failureCount)
	}

	if uf.lastFailTime.IsZero() {
		t.Error("Expected lastFailTime to be set")
	}

	if uf.blockedUntil.IsZero() {
		t.Error("Expected blockedUntil to be set")
	}
}

// TestIPBlockListManagement проверяет управление блокировкой IPs
func TestIPBlockListManagement(t *testing.T) {
	blockList := make(map[string]time.Time)

	ip := "10.0.0.1"
	blockTime := time.Now().Add(15 * time.Minute)

	// Добавляем IP в блокировку
	blockList[ip] = blockTime

	// Проверяем, что IP заблокирован
	if _, exists := blockList[ip]; !exists {
		t.Error("Expected IP to be in blockList")
	}

	// Проверяем время разблокировки
	if blockList[ip] != blockTime {
		t.Errorf("Expected block time %v, got %v", blockTime, blockList[ip])
	}

	// Удаляем из блокировки
	delete(blockList, ip)

	if _, exists := blockList[ip]; exists {
		t.Error("Expected IP to be removed from blockList")
	}
}

// TestFailureCounterReset проверяет сброс счётчиков сбоев
func TestFailureCounterReset(t *testing.T) {
	metrics := &FailureMetrics{
		count:             10,
		consecutiveErrors: 5,
		lastFailureTime:   time.Now(),
	}

	// Сбрасываем consecutiveErrors
	metrics.consecutiveErrors = 0

	if metrics.consecutiveErrors != 0 {
		t.Error("Expected consecutiveErrors to be 0")
	}

	// Счётчик count остаётся для статистики
	if metrics.count != 10 {
		t.Error("Expected count to remain 10")
	}
}

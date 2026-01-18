package handlers

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"
	"tutoring-platform/pkg/telegram"

	"golang.org/x/time/rate"
)

// WebhookFailureTracker отслеживает неудачные попытки обработки webhook
// для обнаружения подозрительных паттернов
type WebhookFailureTracker struct {
	mu              sync.RWMutex
	failures        map[string]*FailureMetrics // Ключ: IP адрес
	blockList       map[string]time.Time       // Заблокированные IPs с временем разблокировки
	suspiciousUsers map[string]*UserFailures   // Ключ: user_id, отслеживание сбоев
}

// FailureMetrics содержит метрики сбоев для IP адреса
type FailureMetrics struct {
	count             int
	lastFailureTime   time.Time
	consecutiveErrors int // Количество подряд идущих ошибок
}

// UserFailures отслеживает неудачные попытки для пользователя
type UserFailures struct {
	failureCount   int
	lastFailTime   time.Time
	blockedUntil   time.Time
	lastUpdateTime time.Time
}

// TelegramHandler обрабатывает эндпоинты для работы с Telegram интеграцией
type TelegramHandler struct {
	telegramService   *service.TelegramService
	botUsername       string
	webhookSecret     string
	webhookLimiter    *IPWebhookRateLimiter // Per-IP rate limiting для webhook
	failureTracker    *WebhookFailureTracker
	trustedProxies    map[string]bool
	cleanupTickerDone chan struct{}
}

// IPWebhookRateLimiter специализированный rate limiter для webhook с более строгими лимитами
type IPWebhookRateLimiter struct {
	mu         sync.RWMutex
	limiters   map[string]*rate.Limiter
	lastAccess map[string]time.Time
	ttl        time.Duration
	stopChan   chan struct{}
}

// NewIPWebhookRateLimiter создает новый webhook rate limiter
// Лимит: 35 запросов в минуту (немного выше Telegram's ~30 req/min)
// Burst: 2 для защиты от sudden spikes
func NewIPWebhookRateLimiter() *IPWebhookRateLimiter {
	wrl := &IPWebhookRateLimiter{
		limiters:   make(map[string]*rate.Limiter),
		lastAccess: make(map[string]time.Time),
		ttl:        30 * time.Minute, // TTL для неиспользуемых entries
		stopChan:   make(chan struct{}),
	}

	// Запускаем cleanup goroutine
	go wrl.cleanupExpired()

	return wrl
}

// GetLimiter возвращает rate limiter для IP адреса
// Лимит: 35 req/min, burst: 2
func (wrl *IPWebhookRateLimiter) GetLimiter(ip string) *rate.Limiter {
	wrl.mu.Lock()
	defer wrl.mu.Unlock()

	if limiter, exists := wrl.limiters[ip]; exists {
		wrl.lastAccess[ip] = time.Now()
		return limiter
	}

	// Telegram max is ~30 req/sec, но через webhook приходит ~30 req/min
	// Лимит: 35 requests per minute (rate.Every(1.71 seconds) ≈ 35 req/min)
	// Burst: 2 для защиты от rate limit attacks
	limiter := rate.NewLimiter(rate.Every(1710*time.Millisecond), 2)
	wrl.limiters[ip] = limiter
	wrl.lastAccess[ip] = time.Now()

	return limiter
}

// Stop останавливает cleanup goroutine
func (wrl *IPWebhookRateLimiter) Stop() {
	select {
	case wrl.stopChan <- struct{}{}:
	default:
		// Канал может быть уже закрыт или full
	}
}

// cleanupExpired периодически удаляет неиспользуемые limiters
func (wrl *IPWebhookRateLimiter) cleanupExpired() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-wrl.stopChan:
			return
		case <-ticker.C:
			wrl.mu.Lock()
			now := time.Now()
			for ip, lastAccess := range wrl.lastAccess {
				if now.Sub(lastAccess) > wrl.ttl {
					delete(wrl.limiters, ip)
					delete(wrl.lastAccess, ip)
				}
			}
			wrl.mu.Unlock()
		}
	}
}

// NewTelegramHandler создает новый TelegramHandler
func NewTelegramHandler(telegramService *service.TelegramService, botUsername, webhookSecret string) *TelegramHandler {
	handler := &TelegramHandler{
		telegramService: telegramService,
		botUsername:     botUsername,
		webhookSecret:   webhookSecret,
		webhookLimiter:  NewIPWebhookRateLimiter(),
		failureTracker: &WebhookFailureTracker{
			failures:        make(map[string]*FailureMetrics),
			blockList:       make(map[string]time.Time),
			suspiciousUsers: make(map[string]*UserFailures),
		},
		trustedProxies:    make(map[string]bool),
		cleanupTickerDone: make(chan struct{}),
	}

	// Запускаем cleanup для failure tracker
	go handler.cleanupFailureMetrics()

	return handler
}

// GenerateLinkToken обрабатывает GET /api/v1/telegram/link-token
// Генерирует токен для привязки Telegram аккаунта
func (h *TelegramHandler) GenerateLinkToken(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Генерируем токен привязки
	token, err := h.telegramService.GenerateLinkToken(r.Context(), user.ID)
	if err != nil {
		log.Printf("Failed to generate link token for user %s: %v", user.ID, err)

		// Проверяем специфичные ошибки используя errors.Is()
		// Возвращаем 409 Conflict ТОЛЬКО если пользователь РЕАЛЬНО привязан (telegram_id > 0)
		if errors.Is(err, service.ErrTelegramAlreadyLinked) {
			response.Conflict(w, response.ErrCodeAlreadyExists, "Telegram account already linked")
			return
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			response.NotFound(w, "User not found")
			return
		}

		response.InternalError(w, "Failed to generate link token")
		return
	}

	log.Printf("Successfully generated link token for user %s", user.ID)
	response.OK(w, map[string]interface{}{
		"token":        token,
		"bot_username": h.botUsername,
	})
}

// GetMyTelegramLink обрабатывает GET /api/v1/telegram/me
// Возвращает статус привязки Telegram аккаунта текущего пользователя
func (h *TelegramHandler) GetMyTelegramLink(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Получаем статус привязки
	telegramUser, err := h.telegramService.GetUserLinkStatus(r.Context(), user.ID)
	if err != nil {
		// Определяем категорию ошибки для логирования и сообщения об ошибке
		errorCategory := "unknown_error"

		// Проверяем специфичные ошибки используя errors.Is()
		if errors.Is(err, context.Canceled) {
			errorCategory = "context_canceled"
			log.Printf("ERROR: Request canceled while getting telegram link status for user %s", user.ID)
			response.RequestTimeout(w, "Request timeout while fetching telegram link status")
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			errorCategory = "deadline_exceeded"
			log.Printf("ERROR: Deadline exceeded while getting telegram link status for user %s", user.ID)
			response.ServiceUnavailable(w, "Request timeout while fetching telegram link status")
			return
		}

		// Для прочих ошибок логируем с категорией и минималистичным сообщением
		log.Printf("ERROR: Failed to get telegram link status for user %s (category=%s): %v",
			user.ID, errorCategory, err)
		response.InternalError(w, "Failed to fetch telegram link status")
		return
	}

	// Если пользователь не привязан, GetUserLinkStatus возвращает nil, nil
	if telegramUser == nil {
		response.OK(w, map[string]interface{}{
			"linked": false,
		})
		return
	}

	// Пользователь привязан
	response.OK(w, map[string]interface{}{
		"linked":   true,
		"telegram": telegramUser,
	})
}

// UnlinkTelegram обрабатывает DELETE /api/v1/telegram/link
// Отвязывает Telegram аккаунт от пользователя
func (h *TelegramHandler) UnlinkTelegram(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Отвязываем Telegram
	if err := h.telegramService.UnlinkUser(r.Context(), user.ID); err != nil {
		log.Printf("ERROR: Failed to unlink telegram for user %s: %v", user.ID, err)

		if errors.Is(err, service.ErrUserNotLinked) {
			response.NotFound(w, "Telegram account not linked")
			return
		}

		response.InternalError(w, "Failed to unlink telegram account")
		return
	}

	response.OK(w, map[string]string{
		"message": "Telegram account unlinked successfully",
	})
}

// Subscribe обрабатывает POST /api/v1/telegram/subscribe
// Подписывает пользователя на Telegram уведомления
func (h *TelegramHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Подписываем пользователя
	if err := h.telegramService.SubscribeToNotifications(r.Context(), user.ID); err != nil {
		log.Printf("ERROR: Failed to subscribe to notifications for user %s: %v", user.ID, err)

		if errors.Is(err, service.ErrTelegramUserNotFound) {
			response.NotFound(w, "Telegram account not linked")
			return
		}

		response.InternalError(w, "Failed to subscribe to notifications")
		return
	}

	response.OK(w, map[string]string{
		"message": "Successfully subscribed to notifications",
	})
}

// Unsubscribe обрабатывает POST /api/v1/telegram/unsubscribe
// Отписывает пользователя от Telegram уведомлений
func (h *TelegramHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Отписываем пользователя
	if err := h.telegramService.UnsubscribeFromNotifications(r.Context(), user.ID); err != nil {
		log.Printf("ERROR: Failed to unsubscribe from notifications for user %s: %v", user.ID, err)

		if errors.Is(err, service.ErrTelegramUserNotFound) {
			response.NotFound(w, "Telegram account not linked")
			return
		}

		response.InternalError(w, "Failed to unsubscribe from notifications")
		return
	}

	response.OK(w, map[string]string{
		"message": "Successfully unsubscribed from notifications",
	})
}

// recordWebhookFailure записывает неудачную попытку обработки webhook
func (h *TelegramHandler) recordWebhookFailure(ip string, userID string) {
	h.failureTracker.mu.Lock()
	defer h.failureTracker.mu.Unlock()

	// Обновляем метрики для IP адреса
	if metrics, exists := h.failureTracker.failures[ip]; exists {
		metrics.count++
		metrics.consecutiveErrors++
		metrics.lastFailureTime = time.Now()
	} else {
		h.failureTracker.failures[ip] = &FailureMetrics{
			count:             1,
			consecutiveErrors: 1,
			lastFailureTime:   time.Now(),
		}
	}

	// Если 5+ подряд ошибок от одного IP - блокируем его на 15 минут
	if metrics := h.failureTracker.failures[ip]; metrics.consecutiveErrors >= 5 {
		h.failureTracker.blockList[ip] = time.Now().Add(15 * time.Minute)
		log.Printf("SECURITY: Blocking IP %s for 15 minutes due to %d consecutive webhook errors", ip, metrics.consecutiveErrors)
	}

	// Отслеживаем ошибки пользователя
	if userID != "" {
		if userFailures, exists := h.failureTracker.suspiciousUsers[userID]; exists {
			userFailures.failureCount++
			userFailures.lastFailTime = time.Now()
			// Если 10+ ошибок в течение часа - блокируем пользователя на 1 час
			if userFailures.failureCount >= 10 && time.Since(userFailures.lastUpdateTime) < 1*time.Hour {
				userFailures.blockedUntil = time.Now().Add(1 * time.Hour)
				log.Printf("SECURITY: Blocking Telegram user %s for 1 hour due to %d webhook errors", userID, userFailures.failureCount)
			}
		} else {
			h.failureTracker.suspiciousUsers[userID] = &UserFailures{
				failureCount:   1,
				lastFailTime:   time.Now(),
				lastUpdateTime: time.Now(),
			}
		}
	}
}

// resetWebhookFailures сбрасывает счётчик ошибок при успешной обработке
func (h *TelegramHandler) resetWebhookFailures(ip string, userID string) {
	h.failureTracker.mu.Lock()
	defer h.failureTracker.mu.Unlock()

	// Сбрасываем consecutiveErrors для IP
	if metrics, exists := h.failureTracker.failures[ip]; exists {
		metrics.consecutiveErrors = 0
	}

	// Обнуляем счётчик для пользователя если достаточно времени прошло
	if userID != "" {
		if userFailures, exists := h.failureTracker.suspiciousUsers[userID]; exists {
			if time.Since(userFailures.lastFailTime) > 30*time.Minute {
				userFailures.failureCount = 0
				userFailures.lastUpdateTime = time.Now()
			}
		}
	}
}

// isIPBlocked проверяет, заблокирован ли IP адрес
func (h *TelegramHandler) isIPBlocked(ip string) bool {
	h.failureTracker.mu.RLock()
	defer h.failureTracker.mu.RUnlock()

	blockTime, exists := h.failureTracker.blockList[ip]
	if !exists {
		return false
	}

	if time.Now().After(blockTime) {
		// Время разблокировки истекло
		return false
	}

	return true
}

// getClientIP безопасно извлекает IP адрес клиента
func (h *TelegramHandler) getClientIP(r *http.Request) string {
	// Извлекаем IP из RemoteAddr (прямое соединение)
	directIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		directIP = r.RemoteAddr
	}

	// Если прямое соединение от доверенного прокси, проверяем X-Forwarded-For
	if len(h.trustedProxies) > 0 && h.trustedProxies[directIP] {
		xForwardedFor := r.Header.Get("X-Forwarded-For")
		if xForwardedFor != "" {
			ips := strings.Split(xForwardedFor, ",")
			// Берем первый (leftmost) IP из X-Forwarded-For
			if len(ips) > 0 {
				candidate := strings.TrimSpace(ips[0])
				if net.ParseIP(candidate) != nil {
					return candidate
				}
			}
		}
	}

	return directIP
}

// HandleWebhook обрабатывает POST /api/v1/telegram/webhook
// Обрабатывает входящие обновления от Telegram бота (публичный эндпоинт)
// SECURITY: Использует строгие rate limits, IP-based tracking и failure detection
func (h *TelegramHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	clientIP := h.getClientIP(r)

	// Проверяем, заблокирован ли IP адрес
	if h.isIPBlocked(clientIP) {
		log.Printf("SECURITY: Rejected webhook from blocked IP %s", clientIP)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// КРИТИЧНО: Вебхук ВСЕГДА требует правильно настроенный секретный токен
	// Если TELEGRAM_WEBHOOK_SECRET не установлен, это ошибка конфигурации
	// и запрос ДОЛЖЕН быть отклонён для безопасности
	if h.webhookSecret == "" {
		log.Printf("SECURITY WARNING: Telegram webhook received but TELEGRAM_WEBHOOK_SECRET is not configured. Request rejected from %s", clientIP)
		h.recordWebhookFailure(clientIP, "")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Проверяем подпись для безопасности с использованием constant-time сравнения
	// для защиты от timing attack
	secretToken := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	// Используем subtle.ConstantTimeCompare для защиты от timing attack
	// ConstantTimeCompare возвращает 1 если равны, 0 если не равны
	if subtle.ConstantTimeCompare([]byte(secretToken), []byte(h.webhookSecret)) != 1 {
		log.Printf("SECURITY: Invalid Telegram webhook signature from %s. Request rejected", clientIP)
		h.recordWebhookFailure(clientIP, "")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Проверяем IP-based rate limit
	// Лимит: 35 req/min (более строгий чем глобальный)
	limiter := h.webhookLimiter.GetLimiter(clientIP)
	if !limiter.Allow() {
		log.Printf("SECURITY: Webhook rate limit exceeded for IP %s", clientIP)
		h.recordWebhookFailure(clientIP, "")
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	// Парсим тело запроса как Telegram Update
	var update telegram.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Printf("ERROR: Failed to parse Telegram webhook body from %s: %v", clientIP, err)
		h.recordWebhookFailure(clientIP, "")
		// Всегда возвращаем 200 OK для Telegram webhook
		w.WriteHeader(http.StatusOK)
		return
	}

	// Извлекаем user ID из update для отслеживания
	var userID string
	if update.Message != nil && update.Message.From != nil {
		userID = fmt.Sprintf("%d", update.Message.From.ID)
	}

	// Обрабатываем webhook асинхронно с отделённым контекстом
	// Используем WithoutCancel для отмены от request context, но сохраняем
	// связь с базовым контекстом и timeouts на операции
	go func() {
		// Создаём контекст, отделённый от request, но с timeout для долгих операций
		// Это предотвращает отмену при завершении HTTP request, но гарантирует
		// что операция не будет висеть бесконечно
		bgCtx := context.WithoutCancel(r.Context())
		ctx, cancel := context.WithTimeout(bgCtx, 30*time.Second)
		defer cancel()

		err := h.telegramService.HandleWebhook(ctx, &update)
		if err != nil {
			log.Printf("ERROR: Failed to handle Telegram webhook from %s: %v", clientIP, err)
			h.recordWebhookFailure(clientIP, userID)
		} else {
			// Успешная обработка - сбрасываем счётчик ошибок
			h.resetWebhookFailures(clientIP, userID)
		}
	}()

	// Telegram требует быстрого ответа 200 OK
	w.WriteHeader(http.StatusOK)
}

// Stop останавливает все cleanup goroutines в handler
func (h *TelegramHandler) Stop() {
	h.webhookLimiter.Stop()
	select {
	case h.cleanupTickerDone <- struct{}{}:
	default:
		// Канал может быть full или закрыт
	}
}

// cleanupFailureMetrics периодически очищает старые метрики сбоев
func (h *TelegramHandler) cleanupFailureMetrics() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-h.cleanupTickerDone:
			return
		case <-ticker.C:
			h.failureTracker.mu.Lock()

			now := time.Now()
			// Удаляем старые записи о сбоях (старше 2 часов)
			for ip, metrics := range h.failureTracker.failures {
				if now.Sub(metrics.lastFailureTime) > 2*time.Hour {
					delete(h.failureTracker.failures, ip)
				}
			}

			// Удаляем неактивные записи о пользователях (старше 3 часов)
			for userID, userFailures := range h.failureTracker.suspiciousUsers {
				if now.Sub(userFailures.lastFailTime) > 3*time.Hour {
					delete(h.failureTracker.suspiciousUsers, userID)
				}
			}

			// Удаляем разблокированные IPs из blacklist
			for ip, blockTime := range h.failureTracker.blockList {
				if now.After(blockTime) {
					delete(h.failureTracker.blockList, ip)
				}
			}

			h.failureTracker.mu.Unlock()
		}
	}
}

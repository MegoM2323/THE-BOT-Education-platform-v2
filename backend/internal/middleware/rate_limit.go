package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"tutoring-platform/pkg/response"

	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

// LimiterEntry содержит rate limiter и метаданные о последнем использовании
type LimiterEntry struct {
	limiter      *rate.Limiter
	lastAccessed time.Time
}

// IPRateLimiter управляет rate limiting на основе IP-адресов
type IPRateLimiter struct {
	ips            map[string]*LimiterEntry
	mu             *sync.RWMutex
	r              rate.Limit
	b              int
	trustedProxies map[string]bool // Карта доверенных прокси (для быстрого поиска)
	ttl            time.Duration   // Время жизни неиспользуемого limiter
	stopChan       chan struct{}   // Канал для остановки cleanup goroutine
}

// NewIPRateLimiter создает новый rate limiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		ips:            make(map[string]*LimiterEntry),
		mu:             &sync.RWMutex{},
		r:              r,
		b:              b,
		trustedProxies: make(map[string]bool),
		ttl:            1 * time.Hour, // TTL для неиспользуемых entries
		stopChan:       make(chan struct{}),
	}
	// Запускаем cleanup task при создании limiter
	limiter.startCleanupGoroutine()
	return limiter
}

// NewIPRateLimiterWithProxies создает новый rate limiter с доверенными прокси
// trustedProxies - список доверенных прокси-серверов (могут содержать порт)
func NewIPRateLimiterWithProxies(r rate.Limit, b int, trustedProxies []string) *IPRateLimiter {
	proxiesMap := make(map[string]bool)
	for _, proxy := range trustedProxies {
		// Нормализуем IP (убираем порт если есть)
		ip, _, err := net.SplitHostPort(proxy)
		if err != nil {
			ip = proxy
		}
		// Парсим IP для валидации
		if parsed := net.ParseIP(strings.TrimSpace(ip)); parsed != nil {
			proxiesMap[parsed.String()] = true
		}
	}

	limiter := &IPRateLimiter{
		ips:            make(map[string]*LimiterEntry),
		mu:             &sync.RWMutex{},
		r:              r,
		b:              b,
		trustedProxies: proxiesMap,
		ttl:            1 * time.Hour, // TTL для неиспользуемых entries
		stopChan:       make(chan struct{}),
	}
	// Запускаем cleanup task при создании limiter
	limiter.startCleanupGoroutine()
	return limiter
}

// AddIP создает новый limiter для IP-адреса если его нет
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = &LimiterEntry{
		limiter:      limiter,
		lastAccessed: time.Now(),
	}

	return limiter
}

// GetLimiter возвращает limiter для IP-адреса и обновляет lastAccessed
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	entry, exists := i.ips[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}

	// Обновляем время последнего доступа
	entry.lastAccessed = time.Now()
	i.mu.Unlock()
	return entry.limiter
}

// CleanupExpired удаляет только неактивные (expired) limiters на основе TTL
// Сохраняет активные limiter'ы, используемые в текущий момент
func (i *IPRateLimiter) CleanupExpired() {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()
	removed := 0

	// Проходим по всем entries и удаляем только те, которые не использовались > TTL
	for ip, entry := range i.ips {
		if now.Sub(entry.lastAccessed) > i.ttl {
			delete(i.ips, ip)
			removed++
		}
	}

	if removed > 0 {
		log.Debug().
			Int("removed_entries", removed).
			Int("remaining_entries", len(i.ips)).
			Msg("Rate limiter cleanup: expired entries removed")
	}
}

// isValidIP проверяет валидность IP адреса
func isValidIP(ipStr string) bool {
	if ipStr == "" {
		return false
	}
	parsed := net.ParseIP(ipStr)
	return parsed != nil
}

// isTrustedProxy проверяет является ли IP доверенным прокси-сервером
func isTrustedProxy(ip string, trustedProxies map[string]bool) bool {
	if len(trustedProxies) == 0 {
		return false
	}
	return trustedProxies[ip]
}

// getIPAddressSecure безопасно извлекает IP-адрес с защитой от спуфинга
// SECURITY: Следует RFC 7239 (Forwarded) и принципам безопасности для X-Forwarded-For
//
// Алгоритм:
// 1. Извлекаем IP из RemoteAddr (прямое соединение)
// 2. Если RemoteAddr НЕ из доверенного прокси - используем его и игнорируем заголовки
// 3. Если RemoteAddr из доверенного прокси - проверяем X-Forwarded-For
// 4. В X-Forwarded-For берем ПРАВЫЙ IP (ближайший к нам), пропуская известные доверенные прокси
// 5. Валидируем и очищаем IP перед использованием
func getIPAddressSecure(r *http.Request, trustedProxies map[string]bool) string {
	// Извлекаем IP из RemoteAddr (это исходящее соединение)
	directIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		directIP = r.RemoteAddr
	}

	// Если прямое соединение не от доверенного прокси,
	// не верим заголовкам и возвращаем directIP
	if !isTrustedProxy(directIP, trustedProxies) {
		return directIP
	}

	// RemoteAddr от доверенного прокси - можем проверить X-Forwarded-For
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For формат: client, proxy1, proxy2
		ips := strings.Split(xForwardedFor, ",")

		// Проходим справа налево, пропуская доверенные прокси
		for i := len(ips) - 1; i >= 0; i-- {
			candidate := strings.TrimSpace(ips[i])

			// Валидируем IP
			if !isValidIP(candidate) {
				log.Warn().
					Str("direct_ip", directIP).
					Str("invalid_candidate", candidate).
					Str("x_forwarded_for", xForwardedFor).
					Msg("Rate limiter: detected invalid IP in X-Forwarded-For header (potential spoofing)")
				continue
			}

			// Если это доверенный прокси, пропускаем и смотрим дальше влево
			if isTrustedProxy(candidate, trustedProxies) {
				continue
			}

			// Это клиентский IP (первый слева который не из доверенных прокси)
			return candidate
		}

		// Если все IP в X-Forwarded-For - доверенные прокси, используем directIP
		return directIP
	}

	// Проверяем X-Real-IP только если это от доверенного прокси
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" && isValidIP(xRealIP) {
		return xRealIP
	}

	return directIP
}

// getIPAddress извлекает IP-адрес из HTTP запроса (DEPRECATED)
// ВНИМАНИЕ: Эта функция УЯЗВИМА к спуфингу без доверенных прокси!
// Используйте getIPAddressSecure с параметром trustedProxies
func getIPAddress(r *http.Request) string {
	return getIPAddressSecure(r, make(map[string]bool))
}

// RateLimitMiddleware создает middleware для rate limiting с защитой от спуфинга
func RateLimitMiddleware(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Используем безопасное извлечение IP с доверенными прокси
			ip := getIPAddressSecure(r, limiter.trustedProxies)
			l := limiter.GetLimiter(ip)

			if !l.Allow() {
				response.TooManyRequests(w, "Rate limit exceeded. Please try again later.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoginRateLimiter создает rate limiter для логина (10 запросов в минуту)
// DEPRECATED: Используйте LoginRateLimiterWithProxies для защиты от спуфинга
func LoginRateLimiter() *IPRateLimiter {
	// rate.Every(6*time.Second) = 10 запросов в минуту
	return NewIPRateLimiter(rate.Every(6*time.Second), 10)
}

// LoginRateLimiterWithProxies создает rate limiter для логина с защитой от спуфинга
// trustedProxies - список доверенных прокси-серверов
func LoginRateLimiterWithProxies(trustedProxies []string) *IPRateLimiter {
	// rate.Every(6*time.Second) = 10 запросов в минуту
	return NewIPRateLimiterWithProxies(rate.Every(6*time.Second), 10, trustedProxies)
}

// TrialRequestRateLimiter создает rate limiter для trial requests (5 запросов в 10 минут)
// DEPRECATED: Используйте TrialRequestRateLimiterWithProxies для защиты от спуфинга
func TrialRequestRateLimiter() *IPRateLimiter {
	// rate.Every(2*time.Minute) = 5 запросов в 10 минут
	return NewIPRateLimiter(rate.Every(2*time.Minute), 5)
}

// TrialRequestRateLimiterWithProxies создает rate limiter для trial requests с защитой от спуфинга
// trustedProxies - список доверенных прокси-серверов
func TrialRequestRateLimiterWithProxies(trustedProxies []string) *IPRateLimiter {
	// rate.Every(2*time.Minute) = 5 запросов в 10 минут
	return NewIPRateLimiterWithProxies(rate.Every(2*time.Minute), 5, trustedProxies)
}

// startCleanupGoroutine запускает периодическую очистку неактивных limiters
// Вызывается автоматически при создании limiter'а
func (i *IPRateLimiter) startCleanupGoroutine() {
	const cleanupInterval = 5 * time.Minute // Запускать cleanup каждые 5 минут
	ticker := time.NewTicker(cleanupInterval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Периодически удаляем expired entries
				i.CleanupExpired()
			case <-i.stopChan:
				// Грациозное завершение при остановке
				log.Debug().Msg("Rate limiter cleanup goroutine stopped")
				return
			}
		}
	}()

	log.Debug().
		Dur("cleanup_interval", cleanupInterval).
		Dur("entry_ttl", i.ttl).
		Msg("Rate limiter cleanup goroutine started")
}

// Stop останавливает cleanup goroutine limiter'а
// IMPORTANT: Вызовите это при завершении приложения
func (i *IPRateLimiter) Stop() {
	select {
	case i.stopChan <- struct{}{}:
		log.Debug().Msg("Rate limiter stop signal sent")
	default:
		// Channel может быть уже закрыт
	}
}

// StartCleanupTask (DEPRECATED) запускает периодическую очистку неактивных limiters
// ВНИМАНИЕ: Это неправильно! Используйте NewIPRateLimiter или NewIPRateLimiterWithProxies
// которые автоматически запускают cleanup goroutine
func StartCleanupTask(limiter *IPRateLimiter, interval time.Duration) {
	log.Warn().Msg("StartCleanupTask is deprecated. Cleanup goroutine starts automatically on limiter creation.")
}

// UserRateLimiter управляет rate limiting на основе User ID
// Используется для payment endpoints для защиты от fraud и DoS
type UserRateLimiter struct {
	users    map[string]*LimiterEntry // ключ: user_id (string)
	mu       *sync.RWMutex
	r        rate.Limit
	b        int
	ttl      time.Duration
	stopChan chan struct{}
}

// NewUserRateLimiter создает новый user-based rate limiter
func NewUserRateLimiter(r rate.Limit, b int) *UserRateLimiter {
	limiter := &UserRateLimiter{
		users:    make(map[string]*LimiterEntry),
		mu:       &sync.RWMutex{},
		r:        r,
		b:        b,
		ttl:      1 * time.Hour,
		stopChan: make(chan struct{}),
	}
	limiter.startCleanupGoroutine()
	return limiter
}

// GetLimiter возвращает limiter для user ID и обновляет lastAccessed
func (u *UserRateLimiter) GetLimiter(userID string) *rate.Limiter {
	u.mu.Lock()
	entry, exists := u.users[userID]

	if !exists {
		u.mu.Unlock()
		return u.addUser(userID)
	}

	// Обновляем время последнего доступа
	entry.lastAccessed = time.Now()
	u.mu.Unlock()
	return entry.limiter
}

// addUser создает новый limiter для user ID
func (u *UserRateLimiter) addUser(userID string) *rate.Limiter {
	u.mu.Lock()
	defer u.mu.Unlock()

	limiter := rate.NewLimiter(u.r, u.b)
	u.users[userID] = &LimiterEntry{
		limiter:      limiter,
		lastAccessed: time.Now(),
	}

	return limiter
}

// CleanupExpired удаляет неактивные limiters на основе TTL
func (u *UserRateLimiter) CleanupExpired() {
	u.mu.Lock()
	defer u.mu.Unlock()

	now := time.Now()
	removed := 0

	for userID, entry := range u.users {
		if now.Sub(entry.lastAccessed) > u.ttl {
			delete(u.users, userID)
			removed++
		}
	}

	if removed > 0 {
		log.Debug().
			Int("removed_entries", removed).
			Int("remaining_entries", len(u.users)).
			Msg("User rate limiter cleanup: expired entries removed")
	}
}

// startCleanupGoroutine запускает периодическую очистку
func (u *UserRateLimiter) startCleanupGoroutine() {
	const cleanupInterval = 5 * time.Minute
	ticker := time.NewTicker(cleanupInterval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				u.CleanupExpired()
			case <-u.stopChan:
				log.Debug().Msg("User rate limiter cleanup goroutine shutting down")
				return
			}
		}
	}()
}

// Stop останавливает cleanup goroutine
func (u *UserRateLimiter) Stop() {
	close(u.stopChan)
}

// PaymentRateLimiter создает rate limiter для payment endpoints (10 запросов в минуту)
func PaymentRateLimiter() *UserRateLimiter {
	// rate.Every(6*time.Second) = 10 запросов в минуту
	return NewUserRateLimiter(rate.Every(6*time.Second), 10)
}

// UserRateLimitMiddleware создает middleware для user-based rate limiting
// Используется для payment endpoints и других чувствительных операций
// Требует authenticated user в контексте
func UserRateLimitMiddleware(limiter *UserRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем user из контекста
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				// Если пользователь не аутентифицирован, отказываем
				response.Unauthorized(w, "Authentication required")
				return
			}

			// Получаем limiter для этого user ID
			userID := user.ID.String() // Преобразуем UUID в string для ключа
			l := limiter.GetLimiter(userID)

			if !l.Allow() {
				response.TooManyRequests(w, "Rate limit exceeded. Please try again later.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

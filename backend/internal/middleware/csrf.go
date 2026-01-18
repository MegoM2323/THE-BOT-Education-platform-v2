package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"tutoring-platform/pkg/response"
)

// tokenEntry хранит CSRF token с timestamp истечения (TTL)
type tokenEntry struct {
	token     string
	expiresAt time.Time
}

// CSRFTokenStore хранит CSRF tokens для каждой сессии
// Использует in-memory хранилище с потокобезопасным доступом
// Вещества с TTL и автоматической очисткой истекших токенов
type CSRFTokenStore struct {
	mu                sync.RWMutex
	tokens            map[string][]tokenEntry // session_id -> []tokenEntry с TTL
	tokenTTL          time.Duration           // время жизни токена (по умолчанию 24 часа)
	maxTokensPerSess  int                     // максимум токенов на сессию (по умолчанию 10)
	cleanupInterval   time.Duration           // интервал очистки (по умолчанию 1 час)
	cleanupTicker     *time.Ticker
	stopCleanupChan   chan struct{}
	cleanupDoneWaitCh chan struct{}
}

// NewCSRFTokenStore создает новый CSRFTokenStore с TTL и cleanup
func NewCSRFTokenStore() *CSRFTokenStore {
	store := &CSRFTokenStore{
		tokens:            make(map[string][]tokenEntry),
		tokenTTL:          24 * time.Hour, // 24 часа TTL
		maxTokensPerSess:  10,             // максимум 10 токенов на сессию
		cleanupInterval:   1 * time.Hour,  // очистка каждый час
		stopCleanupChan:   make(chan struct{}),
		cleanupDoneWaitCh: make(chan struct{}),
	}

	// Запустить cleanup goroutine
	store.startCleanupRoutine()

	return store
}

// GenerateToken создает новый CSRF token для session с TTL
// Генерирует криптографически стойкий случайный токен (32 байта)
// Токен автоматически истечет через 24 часа
func (s *CSRFTokenStore) GenerateToken(sessionID string) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(bytes)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Удалить истекшие токены для сессии перед добавлением нового
	s.removeExpiredTokensLocked(sessionID)

	// Получить текущие токены для сессии
	tokens := s.tokens[sessionID]

	// Если достигнут лимит токенов на сессию - удалить самый старый
	if len(tokens) >= s.maxTokensPerSess {
		tokens = tokens[1:] // удалить первый (самый старый) токен
	}

	// Добавить новый токен с TTL
	entry := tokenEntry{
		token:     token,
		expiresAt: time.Now().Add(s.tokenTTL),
	}
	s.tokens[sessionID] = append(tokens, entry)

	return token, nil
}

// ValidateToken проверяет CSRF token с защитой от timing attacks
// Использует constant-time сравнение (crypto/subtle.ConstantTimeCompare)
// Возвращает true если токен существует, не истек и совпадает с сохраненным
func (s *CSRFTokenStore) ValidateToken(sessionID, token string) bool {
	if sessionID == "" || token == "" {
		return false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, exists := s.tokens[sessionID]
	if !exists {
		return false
	}

	now := time.Now()

	// Проверить каждый токен в списке
	for _, entry := range entries {
		// Пропустить истекшие токены
		if now.After(entry.expiresAt) {
			continue
		}

		// Используем constant-time сравнение для защиты от timing attacks
		// ConstantTimeCompare возвращает 1 если токены равны, 0 если не равны
		if subtle.ConstantTimeCompare([]byte(entry.token), []byte(token)) == 1 {
			return true
		}
	}

	return false
}

// DeleteToken удаляет все CSRF tokens для сессии (вызывается при logout)
func (s *CSRFTokenStore) DeleteToken(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tokens, sessionID)
}

// CSRFMiddleware проверяет CSRF token на state-changing requests
// Пропускает GET, HEAD, OPTIONS без проверки
// Для POST/PUT/PATCH/DELETE требует валидный X-CSRF-Token header
func CSRFMiddleware(store *CSRFTokenStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропустить safe HTTP methods (не изменяют state)
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Получить session из контекста (установлено AuthMiddleware)
			session, ok := GetSessionFromContext(r.Context())
			if !ok || session == nil {
				response.Unauthorized(w, "Authentication required")
				return
			}

			// Получить CSRF token из header
			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				response.Error(w, http.StatusForbidden, response.ErrCodeInvalidCSRF, "CSRF token missing")
				return
			}

			// Валидировать токен
			if !store.ValidateToken(session.ID.String(), token) {
				response.Error(w, http.StatusForbidden, response.ErrCodeInvalidCSRF, "Invalid CSRF token")
				return
			}

			// CSRF проверка пройдена, продолжаем
			next.ServeHTTP(w, r)
		})
	}
}

// GetSessionIDFromContext извлекает session ID из контекста
// Используется для генерации CSRF токена
func GetSessionIDFromContext(ctx context.Context) string {
	session, ok := GetSessionFromContext(ctx)
	if !ok || session == nil {
		return ""
	}
	return session.ID.String()
}

// startCleanupRoutine запускает goroutine для периодической очистки истекших токенов
// Запускается один раз при создании CSRFTokenStore
func (s *CSRFTokenStore) startCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(s.cleanupInterval)
		defer ticker.Stop()
		defer close(s.cleanupDoneWaitCh)

		for {
			select {
			case <-s.stopCleanupChan:
				// Сигнал остановки получен
				return
			case <-ticker.C:
				// Выполнить cleanup
				s.cleanupExpiredTokens()
			}
		}
	}()
}

// removeExpiredTokensLocked удаляет истекшие токены для сессии (должен быть вызван с mu.Lock)
func (s *CSRFTokenStore) removeExpiredTokensLocked(sessionID string) {
	entries, exists := s.tokens[sessionID]
	if !exists || len(entries) == 0 {
		return
	}

	now := time.Now()
	validTokens := make([]tokenEntry, 0, len(entries))

	for _, entry := range entries {
		// Оставить только невыстекшие токены
		if now.Before(entry.expiresAt) {
			validTokens = append(validTokens, entry)
		}
	}

	if len(validTokens) == 0 {
		delete(s.tokens, sessionID)
	} else {
		s.tokens[sessionID] = validTokens
	}
}

// cleanupExpiredTokens удаляет истекшие токены из всех сессий
// Вызывается периодически cleanup goroutine
func (s *CSRFTokenStore) cleanupExpiredTokens() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	sessionsToDelete := make([]string, 0)

	for sessionID, entries := range s.tokens {
		validTokens := make([]tokenEntry, 0, len(entries))

		for _, entry := range entries {
			// Оставить только невыстекшие токены
			if now.Before(entry.expiresAt) {
				validTokens = append(validTokens, entry)
			}
		}

		if len(validTokens) == 0 {
			// Сессия больше не имеет валидных токенов
			sessionsToDelete = append(sessionsToDelete, sessionID)
		} else {
			s.tokens[sessionID] = validTokens
		}
	}

	// Удалить сессии без валидных токенов
	for _, sessionID := range sessionsToDelete {
		delete(s.tokens, sessionID)
	}
}

// Stop останавливает cleanup goroutine и ждет его завершения
// Должен быть вызван перед завершением приложения для graceful shutdown
func (s *CSRFTokenStore) Stop() {
	close(s.stopCleanupChan)
	<-s.cleanupDoneWaitCh
}

// GetTokenCount возвращает количество активных токенов для сессии
// Используется для мониторинга и тестирования
func (s *CSRFTokenStore) GetTokenCount(sessionID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, exists := s.tokens[sessionID]
	if !exists {
		return 0
	}

	now := time.Now()
	count := 0

	for _, entry := range entries {
		if now.Before(entry.expiresAt) {
			count++
		}
	}

	return count
}

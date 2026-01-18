package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
)

func TestCSRFTokenStore_GenerateToken(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()
	sessionID := uuid.New().String()

	// Генерация токена должна быть успешной
	token1, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if token1 == "" {
		t.Fatal("Expected non-empty token")
	}

	// Генерация второго токена для той же сессии должна добавить его к существующему
	// (с новой моделью можно иметь несколько активных токенов)
	token2, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if token1 == token2 {
		t.Error("Expected different tokens on regeneration")
	}

	// Проверка что оба токена валидные
	if !store.ValidateToken(sessionID, token1) {
		t.Error("Expected token1 to be valid")
	}

	if !store.ValidateToken(sessionID, token2) {
		t.Error("Expected token2 to be valid")
	}

	// Проверка что количество токенов соответствует ожиданиям
	count := store.GetTokenCount(sessionID)
	if count != 2 {
		t.Errorf("Expected 2 active tokens, got %d", count)
	}
}

func TestCSRFTokenStore_ValidateToken(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()
	sessionID := uuid.New().String()

	// Валидация несуществующего токена должна вернуть false
	if store.ValidateToken(sessionID, "nonexistent") {
		t.Error("Expected validation to fail for nonexistent token")
	}

	// Генерируем токен
	token, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Валидация правильного токена должна быть успешной
	if !store.ValidateToken(sessionID, token) {
		t.Error("Expected validation to succeed for correct token")
	}

	// Валидация с неправильным session ID должна провалиться
	wrongSessionID := uuid.New().String()
	if store.ValidateToken(wrongSessionID, token) {
		t.Error("Expected validation to fail for wrong session ID")
	}

	// Валидация с неправильным токеном должна провалиться
	if store.ValidateToken(sessionID, "wrong-token") {
		t.Error("Expected validation to fail for wrong token")
	}

	// Валидация с пустыми значениями должна провалиться
	if store.ValidateToken("", token) {
		t.Error("Expected validation to fail for empty session ID")
	}
	if store.ValidateToken(sessionID, "") {
		t.Error("Expected validation to fail for empty token")
	}
}

func TestCSRFTokenStore_DeleteToken(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()
	sessionID := uuid.New().String()

	// Генерируем токен
	token, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Токен должен быть валидным
	if !store.ValidateToken(sessionID, token) {
		t.Error("Expected token to be valid before deletion")
	}

	// Удаляем токен
	store.DeleteToken(sessionID)

	// Токен больше не должен быть валидным
	if store.ValidateToken(sessionID, token) {
		t.Error("Expected token to be invalid after deletion")
	}

	// Удаление несуществующего токена не должно вызывать ошибку
	store.DeleteToken("nonexistent-session")
}

func TestCSRFMiddleware_AllowsSafeMethods(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()
	middleware := CSRFMiddleware(store)

	safeMethods := []string{"GET", "HEAD", "OPTIONS"}

	for _, method := range safeMethods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			rr := httptest.NewRecorder()

			// Handler который просто возвращает 200
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// CSRF middleware не должен блокировать safe методы
			handler := middleware(nextHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", method, rr.Code)
			}
		})
	}
}

func TestCSRFMiddleware_BlocksWithoutSession(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()
	middleware := CSRFMiddleware(store)

	unsafeMethods := []string{"POST", "PUT", "PATCH", "DELETE"}

	for _, method := range unsafeMethods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			rr := httptest.NewRecorder()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Next handler should not be called")
			})

			handler := middleware(nextHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 401 for %s without session, got %d", method, rr.Code)
			}
		})
	}
}

func TestCSRFMiddleware_BlocksWithoutToken(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()
	middleware := CSRFMiddleware(store)

	// Создаем mock session
	sessionID := uuid.New()
	session := &models.SessionWithUser{
		Session: models.Session{
			ID: sessionID,
		},
	}

	unsafeMethods := []string{"POST", "PUT", "PATCH", "DELETE"}

	for _, method := range unsafeMethods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			// Добавляем session в контекст
			ctx := context.WithValue(req.Context(), SessionContextKey, session)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Next handler should not be called without CSRF token")
			})

			handler := middleware(nextHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Errorf("Expected status 403 for %s without CSRF token, got %d", method, rr.Code)
			}
		})
	}
}

func TestCSRFMiddleware_BlocksWithInvalidToken(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()
	middleware := CSRFMiddleware(store)

	// Создаем session и генерируем валидный токен
	sessionID := uuid.New()
	session := &models.SessionWithUser{
		Session: models.Session{
			ID: sessionID,
		},
	}

	validToken, _ := store.GenerateToken(sessionID.String())

	testCases := []struct {
		name        string
		token       string
		expectedMsg string
	}{
		{
			name:        "Wrong token",
			token:       "invalid-token",
			expectedMsg: "Should block with wrong token",
		},
		{
			name:        "Token for different session",
			token:       func() string { t, _ := store.GenerateToken(uuid.New().String()); return t }(),
			expectedMsg: "Should block with token from different session",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", nil)
			req.Header.Set("X-CSRF-Token", tc.token)
			ctx := context.WithValue(req.Context(), SessionContextKey, session)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("%s: Next handler should not be called", tc.expectedMsg)
			})

			handler := middleware(nextHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Errorf("%s: Expected status 403, got %d", tc.expectedMsg, rr.Code)
			}
		})
	}

	// Проверяем что валидный токен всё ещё работает
	t.Run("Valid token passes", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("X-CSRF-Token", validToken)
		ctx := context.WithValue(req.Context(), SessionContextKey, session)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		nextCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)

		if !nextCalled {
			t.Error("Next handler should be called with valid token")
		}
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200 with valid token, got %d", rr.Code)
		}
	})
}

func TestCSRFMiddleware_AllowsWithValidToken(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()
	middleware := CSRFMiddleware(store)

	// Создаем session и генерируем токен
	sessionID := uuid.New()
	session := &models.SessionWithUser{
		Session: models.Session{
			ID: sessionID,
		},
	}

	token, err := store.GenerateToken(sessionID.String())
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	unsafeMethods := []string{"POST", "PUT", "PATCH", "DELETE"}

	for _, method := range unsafeMethods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			req.Header.Set("X-CSRF-Token", token)
			ctx := context.WithValue(req.Context(), SessionContextKey, session)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware(nextHandler)
			handler.ServeHTTP(rr, req)

			if !nextCalled {
				t.Errorf("Next handler should be called for %s with valid token", method)
			}
			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s with valid token, got %d", method, rr.Code)
			}
		})
	}
}

// TestTokenTTL проверяет истечение токена по TTL
func TestTokenTTL(t *testing.T) {
	// Создать store с коротким TTL для тестирования
	store := &CSRFTokenStore{
		tokens:            make(map[string][]tokenEntry),
		tokenTTL:          5 * time.Millisecond,
		maxTokensPerSess:  10,
		cleanupInterval:   100 * time.Millisecond,
		stopCleanupChan:   make(chan struct{}),
		cleanupDoneWaitCh: make(chan struct{}),
	}
	defer store.Stop()
	store.startCleanupRoutine()

	sessionID := uuid.New().String()
	token, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Токен должен быть валидным сразу после генерации
	if !store.ValidateToken(sessionID, token) {
		t.Fatal("Token should be valid immediately after generation")
	}

	// Ждём истечения TTL
	time.Sleep(20 * time.Millisecond)

	// Токен должен быть невалидным после истечения TTL
	if store.ValidateToken(sessionID, token) {
		t.Fatal("Token should be invalid after TTL expires")
	}
}

// TestMaxTokensPerSession проверяет лимит токенов на сессию
func TestMaxTokensPerSession(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()

	sessionID := uuid.New().String()
	tokens := make([]string, 0)

	// Сгенерировать больше токенов, чем maxTokensPerSess
	for i := 0; i < 15; i++ {
		token, err := store.GenerateToken(sessionID)
		if err != nil {
			t.Fatalf("Failed to generate token %d: %v", i, err)
		}
		tokens = append(tokens, token)
	}

	// Проверить количество активных токенов
	count := store.GetTokenCount(sessionID)
	if count > store.maxTokensPerSess {
		t.Errorf("Token count = %d, want <= %d", count, store.maxTokensPerSess)
	}

	// Только последние maxTokensPerSess токенов должны быть валидными
	expectedValid := 0
	for _, token := range tokens {
		if store.ValidateToken(sessionID, token) {
			expectedValid++
		}
	}

	if expectedValid > store.maxTokensPerSess {
		t.Errorf("Valid tokens count = %d, want <= %d", expectedValid, store.maxTokensPerSess)
	}
}

// TestCleanupExpiredTokens проверяет периодическую очистку истекших токенов
func TestCleanupExpiredTokens(t *testing.T) {
	// Создать store с коротким интервалом очистки
	store := &CSRFTokenStore{
		tokens:            make(map[string][]tokenEntry),
		tokenTTL:          5 * time.Millisecond,
		maxTokensPerSess:  10,
		cleanupInterval:   10 * time.Millisecond,
		stopCleanupChan:   make(chan struct{}),
		cleanupDoneWaitCh: make(chan struct{}),
	}
	defer store.Stop()
	store.startCleanupRoutine()

	sessionID := uuid.New().String()
	token, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Проверить что токен активен
	if count := store.GetTokenCount(sessionID); count != 1 {
		t.Errorf("Initial token count = %d, want 1", count)
	}

	// Ждём истечения TTL и срабатывания cleanup
	time.Sleep(50 * time.Millisecond)

	// Токен должен быть удален cleanup goroutine
	if store.ValidateToken(sessionID, token) {
		t.Fatal("Token should be cleaned up after expiration")
	}

	if count := store.GetTokenCount(sessionID); count != 0 {
		t.Errorf("Token count after cleanup = %d, want 0", count)
	}
}

// TestConcurrentTokenGeneration проверяет потокобезопасность при генерации токенов
func TestConcurrentTokenGeneration(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()

	sessionID := uuid.New().String()
	numGoroutines := 10
	tokensPerGoroutine := 5
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*tokensPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < tokensPerGoroutine; j++ {
				_, err := store.GenerateToken(sessionID)
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Проверить что не было ошибок
	for err := range errors {
		if err != nil {
			t.Errorf("Concurrent generation error: %v", err)
		}
	}

	// Проверить количество активных токенов (должно быть <= maxTokensPerSess)
	count := store.GetTokenCount(sessionID)
	if count > store.maxTokensPerSess {
		t.Errorf("Token count = %d, want <= %d", count, store.maxTokensPerSess)
	}
}

// TestConcurrentValidation проверяет потокобезопасность при валидации токенов
func TestConcurrentValidation(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()

	sessionID := uuid.New().String()
	token, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	numGoroutines := 20
	var wg sync.WaitGroup
	validationResults := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := store.ValidateToken(sessionID, token)
			validationResults <- result
		}()
	}

	wg.Wait()
	close(validationResults)

	// Все валидации должны вернуть true
	for result := range validationResults {
		if !result {
			t.Fatal("Concurrent validation should succeed")
		}
	}
}

// TestMultipleSessions проверяет изоляцию токенов между сессиями
func TestMultipleSessions(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()

	session1ID := uuid.New().String()
	session2ID := uuid.New().String()

	token1, err := store.GenerateToken(session1ID)
	if err != nil {
		t.Fatalf("Failed to generate token for session 1: %v", err)
	}

	token2, err := store.GenerateToken(session2ID)
	if err != nil {
		t.Fatalf("Failed to generate token for session 2: %v", err)
	}

	// Каждый токен должен быть валидным только для своей сессии
	if !store.ValidateToken(session1ID, token1) {
		t.Fatal("Token1 should be valid for session 1")
	}

	if store.ValidateToken(session1ID, token2) {
		t.Fatal("Token2 should not be valid for session 1")
	}

	if store.ValidateToken(session2ID, token1) {
		t.Fatal("Token1 should not be valid for session 2")
	}

	if !store.ValidateToken(session2ID, token2) {
		t.Fatal("Token2 should be valid for session 2")
	}
}

// TestMemoryLeakPrevention проверяет что cleanup корректно удаляет истекшие токены
func TestMemoryLeakPrevention(t *testing.T) {
	store := &CSRFTokenStore{
		tokens:            make(map[string][]tokenEntry),
		tokenTTL:          5 * time.Millisecond,
		maxTokensPerSess:  10,
		cleanupInterval:   10 * time.Millisecond,
		stopCleanupChan:   make(chan struct{}),
		cleanupDoneWaitCh: make(chan struct{}),
	}
	defer store.Stop()
	store.startCleanupRoutine()

	// Сгенерировать токены для нескольких сессий
	numSessions := 100
	for i := 0; i < numSessions; i++ {
		sessionID := uuid.New().String()
		for j := 0; j < 5; j++ {
			_, err := store.GenerateToken(sessionID)
			if err != nil {
				t.Fatalf("Failed to generate token: %v", err)
			}
		}
	}

	store.mu.RLock()
	initialMemory := len(store.tokens)
	store.mu.RUnlock()

	if initialMemory != numSessions {
		t.Errorf("Initial memory size = %d, want %d", initialMemory, numSessions)
	}

	// Ждём истечения TTL и выполнения cleanup
	time.Sleep(100 * time.Millisecond)

	// После cleanup память должна быть освобождена
	store.mu.RLock()
	finalMemory := len(store.tokens)
	store.mu.RUnlock()

	if finalMemory != 0 {
		t.Errorf("Memory after cleanup = %d, want 0", finalMemory)
	}
}

// TestMultiTabSupport проверяет что несколько вкладок (браузер) могут работать с разными CSRF токенами одновременно
// Это решает проблему T118: одна вкладка больше не может перезаписать токены другой вкладки
func TestMultiTabSupport(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()

	sessionID := uuid.New().String()

	// Симуляция: 5 вкладок в браузере (разные части приложения)
	// Каждая вкладка имеет собственный CSRF токен, полученный при загрузке страницы
	tabs := make([]string, 5)
	for i := 0; i < 5; i++ {
		token, err := store.GenerateToken(sessionID)
		if err != nil {
			t.Fatalf("Failed to generate token for tab %d: %v", i, err)
		}
		tabs[i] = token
	}

	// Все 5 токенов должны быть валидными (это ключевое улучшение)
	count := store.GetTokenCount(sessionID)
	if count != 5 {
		t.Errorf("Expected 5 active tokens for 5 tabs, got %d", count)
	}

	// Каждая вкладка может использовать свой токен
	for i, token := range tabs {
		if !store.ValidateToken(sessionID, token) {
			t.Errorf("Tab %d: Token should be valid but validation failed", i)
		}
	}

	// Новая вкладка (6-я) получает токен
	tab6Token, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Failed to generate token for tab 6: %v", err)
	}

	// Первые 5 вкладок все еще работают со своими токенами
	for i, token := range tabs {
		if !store.ValidateToken(sessionID, token) {
			t.Errorf("Tab %d: Token should still be valid after tab 6 opens", i)
		}
	}

	// 6-я вкладка работает с новым токеном
	if !store.ValidateToken(sessionID, tab6Token) {
		t.Error("Tab 6: New token should be valid")
	}

	// При достижении лимита (10 токенов), старые удаляются по одному
	// Добавляем еще 4 токена (всего 10)
	for i := 6; i < 10; i++ {
		_, err := store.GenerateToken(sessionID)
		if err != nil {
			t.Fatalf("Failed to generate token for tab %d: %v", i, err)
		}
	}

	count = store.GetTokenCount(sessionID)
	if count != 10 {
		t.Errorf("Expected 10 active tokens at limit, got %d", count)
	}

	// Добавляем 11-й токен - это должно удалить самый старый (от tab 0)
	token11, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Failed to generate token 11: %v", err)
	}

	// Проверяем что первый токен больше не валиден
	if store.ValidateToken(sessionID, tabs[0]) {
		t.Error("Tab 0 token should have been removed (FIFO removal when limit reached)")
	}

	// Остальные токены все еще валидны
	for i := 1; i < 5; i++ {
		if !store.ValidateToken(sessionID, tabs[i]) {
			t.Errorf("Tab %d token should still be valid", i)
		}
	}

	// Новый токен валиден
	if !store.ValidateToken(sessionID, token11) {
		t.Error("Token 11 should be valid")
	}

	// Проверяем количество токенов (не превышает лимит)
	count = store.GetTokenCount(sessionID)
	if count > store.maxTokensPerSess {
		t.Errorf("Token count exceeded limit: %d > %d", count, store.maxTokensPerSess)
	}
}

// TestMultiTabConcurrentAccess проверяет потокобезопасность при одновременной работе нескольких вкладок
func TestMultiTabConcurrentAccess(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()

	sessionID := uuid.New().String()
	numTabs := 10
	requestsPerTab := 20
	var wg sync.WaitGroup

	// Каждая вкладка генерирует и валидирует токены одновременно
	for tabID := 0; tabID < numTabs; tabID++ {
		wg.Add(1)
		go func(tab int) {
			defer wg.Done()
			for req := 0; req < requestsPerTab; req++ {
				// Генерируем новый токен для вкладки
				token, err := store.GenerateToken(sessionID)
				if err != nil {
					t.Errorf("Tab %d request %d: failed to generate token: %v", tab, req, err)
					return
				}

				// Сразу валидируем его
				if !store.ValidateToken(sessionID, token) {
					t.Errorf("Tab %d request %d: generated token should be valid", tab, req)
				}
			}
		}(tabID)
	}

	wg.Wait()

	// После завершения всех операций количество токенов не должно превышать лимит
	count := store.GetTokenCount(sessionID)
	if count > store.maxTokensPerSess {
		t.Errorf("Token count exceeded limit: %d > %d", count, store.maxTokensPerSess)
	}
}

// TestMultiTabWithMiddleware проверяет что middleware корректно работает с несколькими валидными токенами
func TestMultiTabWithMiddleware(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()
	middleware := CSRFMiddleware(store)

	sessionID := uuid.New()
	session := &models.SessionWithUser{
		Session: models.Session{
			ID: sessionID,
		},
	}

	// Генерируем токены для 3 разных вкладок
	tabTokens := make([]string, 3)
	for i := 0; i < 3; i++ {
		token, err := store.GenerateToken(sessionID.String())
		if err != nil {
			t.Fatalf("Failed to generate token for tab %d: %v", i, err)
		}
		tabTokens[i] = token
	}

	// Каждая вкладка может выполнить POST запрос со своим токеном
	for tabID, token := range tabTokens {
		t.Run(fmt.Sprintf("tab_%d_post_request", tabID), func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/data", nil)
			req.Header.Set("X-CSRF-Token", token)
			ctx := context.WithValue(req.Context(), SessionContextKey, session)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware(nextHandler)
			handler.ServeHTTP(rr, req)

			if !nextCalled {
				t.Errorf("Tab %d: Next handler should be called", tabID)
			}
			if rr.Code != http.StatusOK {
				t.Errorf("Tab %d: Expected status 200, got %d", tabID, rr.Code)
			}
		})
	}
}

// TestOldTokensValidUntilRotated проверяет что старые токены остаются валидными пока не удалены ротацией
func TestOldTokensValidUntilRotated(t *testing.T) {
	store := NewCSRFTokenStore()
	defer store.Stop()

	sessionID := uuid.New().String()

	// Генерируем старый токен и запоминаем его
	oldToken, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Failed to generate old token: %v", err)
	}

	// Старый токен валиден
	if !store.ValidateToken(sessionID, oldToken) {
		t.Error("Old token should be valid immediately after creation")
	}

	// Генерируем много новых токенов, чтобы заполнить хранилище
	for i := 0; i < store.maxTokensPerSess-1; i++ {
		_, err := store.GenerateToken(sessionID)
		if err != nil {
			t.Fatalf("Failed to generate token %d: %v", i, err)
		}

		// Старый токен все еще должен быть валидным, пока не достигнут лимит
		if !store.ValidateToken(sessionID, oldToken) {
			t.Errorf("Old token should still be valid after %d new tokens (haven't hit limit yet)", i+1)
		}
	}

	// Теперь хранилище полное (maxTokensPerSess токенов)
	count := store.GetTokenCount(sessionID)
	if count != store.maxTokensPerSess {
		t.Errorf("Expected %d tokens, got %d", store.maxTokensPerSess, count)
	}

	// Генерируем еще один токен - это должно вытеснить старый
	newToken, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Failed to generate final token: %v", err)
	}

	// Старый токен больше не валиден (удален при ротации)
	if store.ValidateToken(sessionID, oldToken) {
		t.Error("Old token should be invalid after being pushed out by FIFO rotation")
	}

	// Новый токен валиден
	if !store.ValidateToken(sessionID, newToken) {
		t.Error("New token should be valid")
	}
}

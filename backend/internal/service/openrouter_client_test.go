package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewOpenRouterClient проверяет создание клиента с параметрами по умолчанию
func TestNewOpenRouterClient(t *testing.T) {
	apiKey := "test-api-key"
	model := "test-model"

	client := NewOpenRouterClient(apiKey, model)

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.apiKey != apiKey {
		t.Errorf("apiKey mismatch: got %q, want %q", client.apiKey, apiKey)
	}

	if client.model != model {
		t.Errorf("model mismatch: got %q, want %q", client.model, model)
	}

	if client.timeout != defaultTimeout {
		t.Errorf("timeout mismatch: got %v, want %v", client.timeout, defaultTimeout)
	}

	if client.maxRetries != defaultMaxRetries {
		t.Errorf("maxRetries mismatch: got %d, want %d", client.maxRetries, defaultMaxRetries)
	}
}

// TestNewOpenRouterClientWithConfig проверяет создание клиента с пользовательскими параметрами
func TestNewOpenRouterClientWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		model       string
		timeout     time.Duration
		maxRetries  int
		wantTimeout time.Duration
		wantRetries int
	}{
		{
			name:        "custom timeout and retries",
			apiKey:      "key",
			model:       "model",
			timeout:     20 * time.Second,
			maxRetries:  5,
			wantTimeout: 20 * time.Second,
			wantRetries: 5,
		},
		{
			name:        "zero timeout falls back to default",
			apiKey:      "key",
			model:       "model",
			timeout:     0,
			maxRetries:  3,
			wantTimeout: defaultTimeout,
			wantRetries: 3,
		},
		{
			name:        "negative timeout falls back to default",
			apiKey:      "key",
			model:       "model",
			timeout:     -1 * time.Second,
			maxRetries:  3,
			wantTimeout: defaultTimeout,
			wantRetries: 3,
		},
		{
			name:        "negative retries falls back to default",
			apiKey:      "key",
			model:       "model",
			timeout:     30 * time.Second,
			maxRetries:  -1,
			wantTimeout: 30 * time.Second,
			wantRetries: defaultMaxRetries,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenRouterClientWithConfig(tt.apiKey, tt.model, tt.timeout, tt.maxRetries)

			if client.timeout != tt.wantTimeout {
				t.Errorf("timeout = %v, want %v", client.timeout, tt.wantTimeout)
			}

			if client.maxRetries != tt.wantRetries {
				t.Errorf("maxRetries = %d, want %d", client.maxRetries, tt.wantRetries)
			}
		})
	}
}

// TestIsRetryableError проверяет определение повторяемых ошибок
func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantRetry bool
	}{
		{
			name:      "nil error",
			err:       nil,
			wantRetry: false,
		},
		{
			name:      "context.Canceled - no retry",
			err:       context.Canceled,
			wantRetry: false,
		},
		{
			name:      "context.DeadlineExceeded - no retry",
			err:       context.DeadlineExceeded,
			wantRetry: false,
		},
		{
			name:      "connection refused - retry",
			err:       errors.New("connection refused"),
			wantRetry: true,
		},
		{
			name:      "connection reset - retry",
			err:       errors.New("connection reset"),
			wantRetry: true,
		},
		{
			name:      "broken pipe - retry",
			err:       errors.New("broken pipe"),
			wantRetry: true,
		},
		{
			name:      "EOF - retry",
			err:       errors.New("EOF"),
			wantRetry: true,
		},
		{
			name:      "invalid format - no retry",
			err:       errors.New("invalid format error"),
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableError(tt.err)
			if got != tt.wantRetry {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, got, tt.wantRetry)
			}
		})
	}
}

// TestCalculateBackoff проверяет вычисление exponential backoff
func TestCalculateBackoff(t *testing.T) {
	client := NewOpenRouterClient("key", "model")

	tests := []struct {
		name       string
		attempt    int
		minBackoff time.Duration
		maxBackoff time.Duration
	}{
		{
			name:       "attempt 0",
			attempt:    0,
			minBackoff: 0,
			maxBackoff: client.initialBackoff * 2, // включает jitter
		},
		{
			name:       "attempt 1",
			attempt:    1,
			minBackoff: client.initialBackoff,
			maxBackoff: client.initialBackoff * 4, // *2 для exponential, *2 для max jitter
		},
		{
			name:       "attempt 2",
			attempt:    2,
			minBackoff: client.initialBackoff * 2,
			maxBackoff: client.maxBackoff, // Достигли max
		},
		{
			name:       "negative attempt treated as 0",
			attempt:    -1,
			minBackoff: 0,
			maxBackoff: client.initialBackoff * 2, // включает jitter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backoff := client.calculateBackoff(tt.attempt)

			if backoff < tt.minBackoff {
				t.Errorf("backoff = %v, want >= %v", backoff, tt.minBackoff)
			}

			if backoff > tt.maxBackoff {
				t.Errorf("backoff = %v, want <= %v", backoff, tt.maxBackoff)
			}
		})
	}
}

// TestFailureMetrics проверяет отслеживание метрик ошибок
func TestFailureMetrics(t *testing.T) {
	client := NewOpenRouterClient("key", "model")

	if client.GetConsecutiveFailures() != 0 {
		t.Errorf("initial failures = %d, want 0", client.GetConsecutiveFailures())
	}

	// Регистрируем ошибки
	client.updateFailureMetrics()
	if client.GetConsecutiveFailures() != 1 {
		t.Errorf("after 1 failure = %d, want 1", client.GetConsecutiveFailures())
	}

	client.updateFailureMetrics()
	if client.GetConsecutiveFailures() != 2 {
		t.Errorf("after 2 failures = %d, want 2", client.GetConsecutiveFailures())
	}

	// Сбрасываем метрики при успехе
	client.resetFailureMetrics()
	if client.GetConsecutiveFailures() != 0 {
		t.Errorf("after reset = %d, want 0", client.GetConsecutiveFailures())
	}
}

// TestSetTimeout проверяет обновление таймаута
func TestSetTimeout(t *testing.T) {
	client := NewOpenRouterClient("key", "model")

	tests := []struct {
		name    string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "valid timeout",
			timeout: 25 * time.Second,
			wantErr: false,
		},
		{
			name:    "zero timeout error",
			timeout: 0,
			wantErr: true,
		},
		{
			name:    "negative timeout error",
			timeout: -1 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.SetTimeout(tt.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetTimeout(%v) error = %v, wantErr %v", tt.timeout, err, tt.wantErr)
			}

			if err == nil && client.timeout != tt.timeout {
				t.Errorf("timeout not updated: got %v, want %v", client.timeout, tt.timeout)
			}
		})
	}
}

// TestModerateMessageSuccess проверяет успешную модерацию
func TestModerateMessageSuccess(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем метод и path
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected /chat/completions, got %s", r.URL.Path)
		}

		// Проверяем авторизацию
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", authHeader)
		}

		// Отправляем ответ
		response := openRouterResponse{
			ID: "test-id",
			Choices: []openRouterResponseChoice{
				{
					Message: openRouterMessage{
						Role:    "assistant",
						Content: `{"blocked": false, "reason": ""}`,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Создаем клиент с кастомным базовым URL
	client := NewOpenRouterClientWithConfig("test-key", "test-model", 5*time.Second, 3)
	client.baseURL = server.URL

	// Выполняем модерацию
	ctx := context.Background()
	result, err := client.ModerateMessage(ctx, "test message")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Blocked != false {
		t.Errorf("blocked = %v, want false", result.Blocked)
	}

	if client.GetConsecutiveFailures() != 0 {
		t.Errorf("failures after success = %d, want 0", client.GetConsecutiveFailures())
	}
}

// TestModerateMessageTimeout проверяет обработку timeout
func TestModerateMessageTimeout(t *testing.T) {
	// Создаем медленный тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем клиент с коротким таймаутом
	client := NewOpenRouterClientWithConfig("test-key", "test-model", 10*time.Millisecond, 1)
	client.baseURL = server.URL

	ctx := context.Background()
	_, err := client.ModerateMessage(ctx, "test message")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Проверяем что это была повторяемая ошибка и мы попытались повторить
	errMsg := err.Error()
	if !containsPattern(errMsg, "after") {
		t.Errorf("expected retry message, got: %v", err)
	}

	// Проверяем что метрики обновлены
	if client.GetConsecutiveFailures() <= 0 {
		t.Errorf("failures = %d, want > 0", client.GetConsecutiveFailures())
	}
}

// TestModerateMessageContextCancelled проверяет отмену контекста
func TestModerateMessageContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Долгий запрос
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewOpenRouterClientWithConfig("test-key", "test-model", 5*time.Second, 3)
	client.baseURL = server.URL

	// Создаем контекст с быстрым deadline
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.ModerateMessage(ctx, "test message")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Проверяем что это была ошибка контекста
	errMsg := err.Error()
	if !containsPattern(errMsg, "context") {
		t.Errorf("expected context error, got: %v", err)
	}
}

// TestModerateMessageServerError проверяет обработку ошибок сервера
func TestModerateMessageServerError(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		shouldRetry bool
	}{
		{
			name:        "500 Internal Server Error - should retry",
			statusCode:  http.StatusInternalServerError,
			shouldRetry: true,
		},
		{
			name:        "502 Bad Gateway - should retry",
			statusCode:  http.StatusBadGateway,
			shouldRetry: true,
		},
		{
			name:        "429 Too Many Requests - should retry",
			statusCode:  http.StatusTooManyRequests,
			shouldRetry: true,
		},
		{
			name:        "400 Bad Request - should NOT retry",
			statusCode:  http.StatusBadRequest,
			shouldRetry: false,
		},
		{
			name:        "401 Unauthorized - should NOT retry",
			statusCode:  http.StatusUnauthorized,
			shouldRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"error": {"message": "test error", "code": "test_code"}}`))
			}))
			defer server.Close()

			client := NewOpenRouterClientWithConfig("test-key", "test-model", 5*time.Second, 2)
			client.baseURL = server.URL

			ctx := context.Background()
			_, err := client.ModerateMessage(ctx, "test message")

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Если shouldRetry, проверяем что было несколько попыток (сообщение должно содержать "after")
			// Если не shouldRetry, проверяем что ошибка API сразу (содержит "openrouter API error")
			errMsg := err.Error()

			if tt.shouldRetry {
				// Retryable errors (5xx, 429) должны иметь сообщение с "after"
				if !containsPattern(errMsg, "after") {
					t.Errorf("expected retry message for status %d, got: %v", tt.statusCode, err)
				}
			} else {
				// Non-retryable errors (4xx except 429) должны иметь сообщение об ошибке API
				if !containsPattern(errMsg, "openrouter API error") {
					t.Errorf("should not retry for status %d, expected API error message, got: %v", tt.statusCode, err)
				}
			}
		})
	}
}

// TestModerateMessageMalformedResponse проверяет обработку некорректного JSON ответа
func TestModerateMessageMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "test", "choices": [{"message": {"role": "assistant", "content": "invalid json"}}]}`))
	}))
	defer server.Close()

	client := NewOpenRouterClientWithConfig("test-key", "test-model", 5*time.Second, 3)
	client.baseURL = server.URL

	ctx := context.Background()
	_, err := client.ModerateMessage(ctx, "test message")

	if err == nil {
		t.Fatal("expected error for malformed response, got nil")
	}

	if !containsPattern(err.Error(), "parse moderation result") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

// TestModerateMessageEmptyChoices проверяет обработку пустого ответа
func TestModerateMessageEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := openRouterResponse{
			ID:      "test-id",
			Choices: []openRouterResponseChoice{}, // Empty
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOpenRouterClientWithConfig("test-key", "test-model", 5*time.Second, 3)
	client.baseURL = server.URL

	ctx := context.Background()
	_, err := client.ModerateMessage(ctx, "test message")

	if err == nil {
		t.Fatal("expected error for empty choices, got nil")
	}

	if !containsPattern(err.Error(), "no choices") {
		t.Errorf("expected 'no choices' error, got: %v", err)
	}
}

// TestModerateMessageRetrySuccess проверяет успешный повтор после ошибки
func TestModerateMessageRetrySuccess(t *testing.T) {
	attemptCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		// Первая попытка возвращает ошибку
		if attemptCount == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// Вторая попытка успешна
		response := openRouterResponse{
			ID: "test-id",
			Choices: []openRouterResponseChoice{
				{
					Message: openRouterMessage{
						Role:    "assistant",
						Content: `{"blocked": true, "reason": "Phone number detected"}`,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOpenRouterClientWithConfig("test-key", "test-model", 5*time.Second, 3)
	client.baseURL = server.URL

	ctx := context.Background()
	result, err := client.ModerateMessage(ctx, "test message")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if !result.Blocked {
		t.Errorf("blocked = %v, want true", result.Blocked)
	}

	if attemptCount != 2 {
		t.Errorf("expected 2 attempts, got %d", attemptCount)
	}

	if client.GetConsecutiveFailures() != 0 {
		t.Errorf("failures after success = %d, want 0", client.GetConsecutiveFailures())
	}
}

// TestModerateMessageExhaustedRetries проверяет исчерпание попыток повтора
func TestModerateMessageExhaustedRetries(t *testing.T) {
	attemptCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewOpenRouterClientWithConfig("test-key", "test-model", 5*time.Second, 2)
	client.baseURL = server.URL

	ctx := context.Background()
	_, err := client.ModerateMessage(ctx, "test message")

	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}

	// Должно быть 3 попытки (0, 1, 2)
	expectedAttempts := 3
	if attemptCount != expectedAttempts {
		t.Errorf("expected %d attempts, got %d", expectedAttempts, attemptCount)
	}

	if client.GetConsecutiveFailures() <= 0 {
		t.Errorf("failures = %d, want > 0", client.GetConsecutiveFailures())
	}
}

// BenchmarkCalculateBackoff бенчмарк для расчета backoff
func BenchmarkCalculateBackoff(b *testing.B) {
	client := NewOpenRouterClient("key", "model")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.calculateBackoff(i % 5)
	}
}

// BenchmarkModerateMessageSuccess бенчмарк для успешной модерации
func BenchmarkModerateMessageSuccess(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := openRouterResponse{
			ID: "test-id",
			Choices: []openRouterResponseChoice{
				{
					Message: openRouterMessage{
						Role:    "assistant",
						Content: `{"blocked": false, "reason": ""}`,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewOpenRouterClientWithConfig("test-key", "test-model", 5*time.Second, 1)
	client.baseURL = server.URL

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.ModerateMessage(ctx, "test message")
	}
}

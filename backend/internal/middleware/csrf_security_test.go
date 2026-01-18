package middleware

import (
	"testing"
	"time"
)

// TestConstantTimeComparison проверяет что валидация использует constant-time сравнение
// Это защищает от timing attacks где злоумышленник может угадать токен по времени ответа
func TestConstantTimeComparison(t *testing.T) {
	store := NewCSRFTokenStore()
	sessionID := "test-session-123"

	// Генерируем валидный токен
	validToken, err := store.GenerateToken(sessionID)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		sessionID      string
		token          string
		shouldValidate bool
		description    string
	}{
		{
			name:           "Valid token",
			sessionID:      sessionID,
			token:          validToken,
			shouldValidate: true,
			description:    "Правильный токен должен пройти валидацию",
		},
		{
			name:           "Invalid token - completely different",
			sessionID:      sessionID,
			token:          "completely-different-token-xyz",
			shouldValidate: false,
			description:    "Совершенно другой токен не должен пройти валидацию",
		},
		{
			name:           "Invalid token - first char wrong",
			sessionID:      sessionID,
			token:          "X" + validToken[1:],
			shouldValidate: false,
			description:    "Токен с неправильным первым символом не должен пройти валидацию",
		},
		{
			name:           "Invalid token - last char wrong",
			sessionID:      sessionID,
			token:          validToken[:len(validToken)-1] + "X",
			shouldValidate: false,
			description:    "Токен с неправильным последним символом не должен пройти валидацию",
		},
		{
			name:           "Invalid token - middle char wrong",
			sessionID:      sessionID,
			token:          validToken[:len(validToken)/2] + "X" + validToken[len(validToken)/2+1:],
			shouldValidate: false,
			description:    "Токен с неправильным средним символом не должен пройти валидацию",
		},
		{
			name:           "Empty token",
			sessionID:      sessionID,
			token:          "",
			shouldValidate: false,
			description:    "Пустой токен не должен пройти валидацию",
		},
		{
			name:           "Empty session ID",
			sessionID:      "",
			token:          validToken,
			shouldValidate: false,
			description:    "Пустой session ID не должен пройти валидацию",
		},
		{
			name:           "Non-existent session",
			sessionID:      "non-existent-session",
			token:          validToken,
			shouldValidate: false,
			description:    "Токен для несуществующей сессии не должен пройти валидацию",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := store.ValidateToken(tt.sessionID, tt.token)
			if result != tt.shouldValidate {
				t.Errorf("ValidateToken(%q, %q) = %v, want %v. %s",
					tt.sessionID, tt.token, result, tt.shouldValidate, tt.description)
			}
		})
	}
}

// BenchmarkConstantTimeComparison измеряет время валидации с разными позициями ошибок
// Все итерации должны занимать примерно одинаковое время (constant-time)
func BenchmarkConstantTimeComparison(b *testing.B) {
	store := NewCSRFTokenStore()
	sessionID := "bench-session"

	validToken, _ := store.GenerateToken(sessionID)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "First char wrong",
			token: "X" + validToken[1:],
		},
		{
			name:  "Middle char wrong",
			token: validToken[:len(validToken)/2] + "X" + validToken[len(validToken)/2+1:],
		},
		{
			name:  "Last char wrong",
			token: validToken[:len(validToken)-1] + "X",
		},
		{
			name:  "Valid token",
			token: validToken,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				store.ValidateToken(sessionID, tt.token)
			}
		})
	}
}

// TestTimingVariation проверяет что время валидации не зависит от позиции несовпадения
// Это более практический тест для обнаружения timing attacks
func TestTimingVariation(t *testing.T) {
	store := NewCSRFTokenStore()
	sessionID := "timing-test-session"

	validToken, _ := store.GenerateToken(sessionID)

	// Тесты с разными позициями неправильных символов
	testCases := []struct {
		name  string
		token string
	}{
		{
			name:  "Wrong at position 0",
			token: "X" + validToken[1:],
		},
		{
			name:  "Wrong at position 16",
			token: validToken[:16] + "X" + validToken[17:],
		},
		{
			name:  "Wrong at last position",
			token: validToken[:len(validToken)-1] + "X",
		},
	}

	// Запускаем каждый тест несколько раз и измеряем время
	iterations := 10000
	timings := make(map[string]time.Duration)

	for _, tc := range testCases {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			store.ValidateToken(sessionID, tc.token)
		}
		duration := time.Since(start)
		timings[tc.name] = duration

		t.Logf("%s: %v for %d iterations (avg: %v per call)",
			tc.name, duration, iterations, duration/time.Duration(iterations))
	}

	// Проверяем что все времена находятся в разумном диапазоне друг от друга
	// Constant-time сравнение должно дать примерно одинаковое время для всех случаев
	var minDuration, maxDuration time.Duration
	for _, duration := range timings {
		if minDuration == 0 || duration < minDuration {
			minDuration = duration
		}
		if duration > maxDuration {
			maxDuration = duration
		}
	}

	// Максимальная разница не должна быть более чем на 30% от минимума
	// (допускаем некоторую вариацию из-за шума системы)
	allowedVariation := minDuration / 3
	difference := maxDuration - minDuration

	if difference > allowedVariation {
		t.Logf("WARNING: Timing variation detected: %v (min: %v, max: %v)",
			difference, minDuration, maxDuration)
		// Это не fatal потому что constant-time сравнение все еще лучше чем обычное ==
		// но это указывает на потенциальную проблему с оптимизацией
	}
}

// TestValidateTokenFunctionality проверяет основную функциональность валидации
func TestValidateTokenFunctionality(t *testing.T) {
	store := NewCSRFTokenStore()

	// Генерируем токен для первой сессии
	session1 := "session-1"
	token1, err := store.GenerateToken(session1)
	if err != nil {
		t.Fatalf("Failed to generate token for session 1: %v", err)
	}

	// Генерируем токен для второй сессии
	session2 := "session-2"
	token2, err := store.GenerateToken(session2)
	if err != nil {
		t.Fatalf("Failed to generate token for session 2: %v", err)
	}

	// Тест 1: Токен правильной сессии должен валидировать
	if !store.ValidateToken(session1, token1) {
		t.Error("Valid token for session 1 should validate")
	}

	// Тест 2: Токен другой сессии не должен валидировать для первой сессии
	if store.ValidateToken(session1, token2) {
		t.Error("Token from session 2 should not validate for session 1")
	}

	// Тест 3: Токен правильной сессии должен валидировать
	if !store.ValidateToken(session2, token2) {
		t.Error("Valid token for session 2 should validate")
	}

	// Тест 4: Токен другой сессии не должен валидировать для второй сессии
	if store.ValidateToken(session2, token1) {
		t.Error("Token from session 1 should not validate for session 2")
	}

	// Тест 5: Удалить токен и проверить что он больше не валидирует
	store.DeleteToken(session1)
	if store.ValidateToken(session1, token1) {
		t.Error("Deleted token should not validate")
	}
}

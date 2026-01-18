package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenStore тесты для TokenStore
func TestTokenStore(t *testing.T) {
	t.Run("GenerateToken creates valid token", func(t *testing.T) {
		ts := NewTokenStore()
		userID := uuid.New()

		token, err := ts.GenerateToken(userID, 15*time.Minute)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Проверяем длину токена (32 байта в base64 = ~43 символа)
		assert.Greater(t, len(token), 40)
	})

	t.Run("ValidateToken returns correct userID", func(t *testing.T) {
		ts := NewTokenStore()
		userID := uuid.New()
		ctx := context.Background()

		token, err := ts.GenerateToken(userID, 15*time.Minute)
		require.NoError(t, err)

		validatedUserID, err := ts.ValidateToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, userID, validatedUserID)
	})

	t.Run("ValidateToken fails for invalid token", func(t *testing.T) {
		ts := NewTokenStore()
		ctx := context.Background()

		_, err := ts.ValidateToken(ctx, "invalid-token")
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("ValidateToken fails for expired token", func(t *testing.T) {
		ts := NewTokenStore()
		userID := uuid.New()
		ctx := context.Background()

		// Создаем токен с отрицательной длительностью (уже истек)
		token, err := ts.GenerateToken(userID, -1*time.Second)
		require.NoError(t, err)

		// Небольшая задержка для гарантии истечения
		time.Sleep(10 * time.Millisecond)

		_, err = ts.ValidateToken(ctx, token)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("DeleteToken removes token", func(t *testing.T) {
		ts := NewTokenStore()
		userID := uuid.New()
		ctx := context.Background()

		token, err := ts.GenerateToken(userID, 15*time.Minute)
		require.NoError(t, err)

		// Токен должен быть валидным
		_, err = ts.ValidateToken(ctx, token)
		require.NoError(t, err)

		// Удаляем токен
		err = ts.DeleteToken(ctx, token)
		require.NoError(t, err)

		// Токен больше не должен быть валидным
		_, err = ts.ValidateToken(ctx, token)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("CleanExpired removes expired tokens", func(t *testing.T) {
		ts := NewTokenStore()
		ctx := context.Background()

		// Создаем истекший токен
		expiredUserID := uuid.New()
		expiredToken, err := ts.GenerateToken(expiredUserID, -1*time.Second)
		require.NoError(t, err)

		// Создаем валидный токен
		validUserID := uuid.New()
		validToken, err := ts.GenerateToken(validUserID, 15*time.Minute)
		require.NoError(t, err)

		// Очищаем истекшие токены
		ts.CleanExpired()

		// Истекший токен должен быть удален
		_, err = ts.ValidateToken(ctx, expiredToken)
		assert.ErrorIs(t, err, ErrInvalidToken)

		// Валидный токен должен остаться
		validatedUserID, err := ts.ValidateToken(ctx, validToken)
		require.NoError(t, err)
		assert.Equal(t, validUserID, validatedUserID)
	})

	t.Run("TokenStore is thread-safe", func(t *testing.T) {
		ts := NewTokenStore()
		ctx := context.Background()
		done := make(chan bool, 10)

		// Запускаем несколько горутин параллельно
		for i := 0; i < 10; i++ {
			go func() {
				userID := uuid.New()
				token, err := ts.GenerateToken(userID, 15*time.Minute)
				assert.NoError(t, err)

				validatedUserID, err := ts.ValidateToken(ctx, token)
				assert.NoError(t, err)
				assert.Equal(t, userID, validatedUserID)

				err = ts.DeleteToken(ctx, token)
				assert.NoError(t, err)

				done <- true
			}()
		}

		// Ждем завершения всех горутин
		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("Multiple tokens for same user", func(t *testing.T) {
		ts := NewTokenStore()
		userID := uuid.New()
		ctx := context.Background()

		// Создаем два токена для одного пользователя
		token1, err := ts.GenerateToken(userID, 15*time.Minute)
		require.NoError(t, err)

		token2, err := ts.GenerateToken(userID, 15*time.Minute)
		require.NoError(t, err)

		// Токены должны быть разными
		assert.NotEqual(t, token1, token2)

		// Оба токена должны быть валидными
		validatedUserID1, err := ts.ValidateToken(ctx, token1)
		require.NoError(t, err)
		assert.Equal(t, userID, validatedUserID1)

		validatedUserID2, err := ts.ValidateToken(ctx, token2)
		require.NoError(t, err)
		assert.Equal(t, userID, validatedUserID2)
	})

	t.Run("Delete user tokens from in-memory store", func(t *testing.T) {
		ts := NewTokenStore()
		userID := uuid.New()
		anotherUserID := uuid.New()
		ctx := context.Background()

		// Создаем несколько токенов для пользователя
		token1, err := ts.GenerateToken(userID, 15*time.Minute)
		require.NoError(t, err)
		token2, err := ts.GenerateToken(userID, 15*time.Minute)
		require.NoError(t, err)

		// Создаем токен для другого пользователя
		otherToken, err := ts.GenerateToken(anotherUserID, 15*time.Minute)
		require.NoError(t, err)

		// Удаляем все токены первого пользователя (имитация очистки перед генерацией нового)
		ts.mu.Lock()
		for token, data := range ts.tokens {
			if data.UserID == userID {
				delete(ts.tokens, token)
			}
		}
		ts.mu.Unlock()

		// Токены первого пользователя должны быть удалены
		_, err = ts.ValidateToken(ctx, token1)
		assert.ErrorIs(t, err, ErrInvalidToken)
		_, err = ts.ValidateToken(ctx, token2)
		assert.ErrorIs(t, err, ErrInvalidToken)

		// Токен другого пользователя должен остаться
		validatedUserID, err := ts.ValidateToken(ctx, otherToken)
		require.NoError(t, err)
		assert.Equal(t, anotherUserID, validatedUserID)
	})
}

// TestTokenStoreIntegration интеграционные тесты
func TestTokenStoreIntegration(t *testing.T) {
	t.Run("Token lifecycle", func(t *testing.T) {
		ts := NewTokenStore()
		userID := uuid.New()
		ctx := context.Background()

		// 1. Генерация токена
		token, err := ts.GenerateToken(userID, 100*time.Millisecond)
		require.NoError(t, err)

		// 2. Немедленная валидация - успешна
		validatedUserID, err := ts.ValidateToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, userID, validatedUserID)

		// 3. Ждем истечения
		time.Sleep(150 * time.Millisecond)

		// 4. Валидация истекшего токена - ошибка
		_, err = ts.ValidateToken(ctx, token)
		assert.ErrorIs(t, err, ErrInvalidToken)

		// 5. Очистка истекших токенов
		ts.CleanExpired()

		// 6. Повторная валидация - все еще ошибка
		_, err = ts.ValidateToken(ctx, token)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("Token usage flow", func(t *testing.T) {
		ts := NewTokenStore()
		userID := uuid.New()
		ctx := context.Background()

		// 1. Генерируем токен для привязки
		token, err := ts.GenerateToken(userID, 15*time.Minute)
		require.NoError(t, err)

		// 2. Пользователь переходит по ссылке - валидируем токен
		validatedUserID, err := ts.ValidateToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, userID, validatedUserID)

		// 3. Привязка успешна - удаляем токен
		err = ts.DeleteToken(ctx, token)
		require.NoError(t, err)

		// 4. Повторное использование токена - ошибка
		_, err = ts.ValidateToken(ctx, token)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}

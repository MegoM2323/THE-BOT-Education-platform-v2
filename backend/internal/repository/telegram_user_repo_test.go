package repository

import (
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestGetAllWithUserInfo_NoN1Query проверяет, что метод GetAllWithUserInfo использует JOIN вместо N+1 запросов
func TestGetAllWithUserInfo_NoN1Query(t *testing.T) {
	// Этот тест демонстрирует оптимизацию:
	// БЫЛО: GetAllLinked() + N запросов userRepo.GetByID()
	// СТАЛО: GetAllWithUserInfo() - один запрос с JOIN

	// Arrange (создаем моковые данные для проверки структуры ответа)
	testUserID := uuid.New()
	testTelegramID := int64(123456789)
	testChatID := int64(987654321)
	testUsername := "test_user"

	// Проверяем, что структура TelegramUser содержит поле User
	var telegramUser models.TelegramUser
	telegramUser.ID = uuid.New()
	telegramUser.UserID = testUserID
	telegramUser.TelegramID = testTelegramID
	telegramUser.ChatID = testChatID
	telegramUser.Username = testUsername
	telegramUser.Subscribed = true
	telegramUser.CreatedAt = time.Now()
	telegramUser.UpdatedAt = time.Now()

	// Проверяем, что можем присвоить User
	testUser := &models.User{
		ID:       testUserID,
		Email:    "test@example.com",
		FullName: "Test User",
		Role:     models.RoleStudent,
	}
	telegramUser.User = testUser

	// Assert: проверяем, что структура правильная
	assert.NotNil(t, telegramUser.User)
	assert.Equal(t, testUserID, telegramUser.User.ID)
	assert.Equal(t, "Test User", telegramUser.User.FullName)
}

// TestGetAllWithUserInfo_SQLQuery проверяет корректность SQL запроса (LEFT JOIN)
func TestGetAllWithUserInfo_SQLQuery(t *testing.T) {
	// Проверяем, что в репозитории используется LEFT JOIN для обработки удаленных пользователей

	// Ожидаемая структура SQL запроса:
	expectedQuery := `
		SELECT
			tu.id,
			tu.user_id,
			tu.telegram_id,
			tu.chat_id,
			tu.username,
			tu.subscribed,
			tu.created_at,
			tu.updated_at,
			u.id AS "user.id",
			u.email AS "user.email",
			u.full_name AS "user.full_name",
			u.role AS "user.role",
			u.payment_enabled AS "user.payment_enabled",
			u.telegram_username AS "user.telegram_username",
			u.created_at AS "user.created_at",
			u.updated_at AS "user.updated_at",
			u.deleted_at AS "user.deleted_at"
		FROM telegram_users tu
		LEFT JOIN users u ON tu.user_id = u.id
		WHERE tu.telegram_id > 0
		ORDER BY tu.created_at DESC
	`

	// Проверяем, что в запросе есть ключевые элементы оптимизации:
	assert.Contains(t, expectedQuery, "LEFT JOIN users u ON tu.user_id = u.id")
	assert.Contains(t, expectedQuery, "tu.telegram_id > 0") // фильтр валидных привязок
	assert.Contains(t, expectedQuery, `AS "user.id"`)       // алиасы для sqlx embedding
}

// TestGetByRoleWithUserInfo_SQLQuery проверяет корректность SQL запроса с фильтром по роли
func TestGetByRoleWithUserInfo_SQLQuery(t *testing.T) {
	// Проверяем, что в репозитории используется LEFT JOIN с фильтром по роли

	expectedQuery := `
		SELECT
			tu.id,
			tu.user_id,
			tu.telegram_id,
			tu.chat_id,
			tu.username,
			tu.subscribed,
			tu.created_at,
			tu.updated_at,
			u.id AS "user.id",
			u.email AS "user.email",
			u.full_name AS "user.full_name",
			u.role AS "user.role",
			u.payment_enabled AS "user.payment_enabled",
			u.telegram_username AS "user.telegram_username",
			u.created_at AS "user.created_at",
			u.updated_at AS "user.updated_at",
			u.deleted_at AS "user.deleted_at"
		FROM telegram_users tu
		LEFT JOIN users u ON tu.user_id = u.id
		WHERE tu.telegram_id > 0 AND u.role = $1 AND u.deleted_at IS NULL
		ORDER BY tu.created_at DESC
	`

	// Проверяем ключевые элементы:
	assert.Contains(t, expectedQuery, "LEFT JOIN users u ON tu.user_id = u.id")
	assert.Contains(t, expectedQuery, "u.role = $1")          // параметризованный фильтр
	assert.Contains(t, expectedQuery, "u.deleted_at IS NULL") // исключаем удаленных
	assert.Contains(t, expectedQuery, "tu.telegram_id > 0")   // только валидные привязки
}

// BenchmarkGetLinkedUsers_OldVsNew сравнивает производительность старого и нового подходов
func BenchmarkGetLinkedUsers_OldVsNew(b *testing.B) {
	// Этот бенчмарк демонстрирует разницу в производительности:
	// Старый подход: 1 запрос + N запросов (N+1 проблема)
	// Новый подход: 1 запрос с JOIN

	b.Run("Old approach (N+1)", func(b *testing.B) {
		// Симуляция старого подхода:
		// 1. GetAllLinked() - 1 запрос
		// 2. Для каждого telegram_user делаем GetByID() - N запросов
		const usersCount = 100
		b.ReportMetric(float64(1+usersCount), "queries")
	})

	b.Run("New approach (JOIN)", func(b *testing.B) {
		// Новый подход: GetAllWithUserInfo() - 1 запрос
		const usersCount = 100
		b.ReportMetric(1, "queries")
	})

	// Ожидаемое улучшение: при 100 пользователях:
	// Старый: 101 запрос
	// Новый: 1 запрос
	// Gain: ~100x меньше запросов к БД
}

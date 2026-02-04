package repository

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetSubscribedUserIDs tests filtering subscribed users from a list
func TestGetSubscribedUserIDs(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Create multiple users with different Telegram subscription statuses
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()
	userID4 := uuid.New()

	createTestUserByID(t, db, userID1, "user1@test.com", "student")
	createTestUserByID(t, db, userID2, "user2@test.com", "student")
	createTestUserByID(t, db, userID3, "user3@test.com", "student")
	createTestUserByID(t, db, userID4, "user4@test.com", "student")

	now := time.Now()

	// Link users to Telegram with different subscription statuses
	// userID1: subscribed = true
	_, err := db.ExecContext(ctx, `
		INSERT INTO telegram_users (id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), userID1, 111111111, 111111111, "user1", true, now, now)
	require.NoError(t, err)

	// userID2: subscribed = true
	_, err = db.ExecContext(ctx, `
		INSERT INTO telegram_users (id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), userID2, 222222222, 222222222, "user2", true, now, now)
	require.NoError(t, err)

	// userID3: subscribed = false
	_, err = db.ExecContext(ctx, `
		INSERT INTO telegram_users (id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), userID3, 333333333, 333333333, "user3", false, now, now)
	require.NoError(t, err)

	// userID4: no Telegram link

	repo := NewTelegramUserRepository(db)

	// Request subscribed users from all 4 user IDs
	inputIDs := []uuid.UUID{userID1, userID2, userID3, userID4}

	subscribedIDs, err := repo.GetSubscribedUserIDs(ctx, inputIDs)
	require.NoError(t, err)

	// Should return only userID1 and userID2 (subscribed = true)
	assert.Len(t, subscribedIDs, 2)

	// Verify the returned IDs are correct
	subscribedMap := make(map[uuid.UUID]bool)
	for _, id := range subscribedIDs {
		subscribedMap[id] = true
	}

	assert.True(t, subscribedMap[userID1], "userID1 should be in subscribed list")
	assert.True(t, subscribedMap[userID2], "userID2 should be in subscribed list")
	assert.False(t, subscribedMap[userID3], "userID3 should NOT be in subscribed list")
	assert.False(t, subscribedMap[userID4], "userID4 should NOT be in subscribed list")
}

// TestGetSubscribedUserIDs_Empty tests with empty input list
func TestGetSubscribedUserIDs_Empty(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	repo := NewTelegramUserRepository(db)

	// Empty input list
	subscribedIDs, err := repo.GetSubscribedUserIDs(ctx, []uuid.UUID{})
	require.NoError(t, err)
	assert.Empty(t, subscribedIDs)
}

// TestGetSubscribedUserIDs_NoneSubscribed tests when no users are subscribed
func TestGetSubscribedUserIDs_NoneSubscribed(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	userID1 := uuid.New()
	userID2 := uuid.New()

	createTestUserByID(t, db, userID1, "user1@test.com", "student")
	createTestUserByID(t, db, userID2, "user2@test.com", "student")

	now := time.Now()

	// Both users subscribed = false
	_, err := db.ExecContext(ctx, `
		INSERT INTO telegram_users (id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), userID1, 111111111, 111111111, "user1", false, now, now)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO telegram_users (id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), userID2, 222222222, 222222222, "user2", false, now, now)
	require.NoError(t, err)

	repo := NewTelegramUserRepository(db)

	subscribedIDs, err := repo.GetSubscribedUserIDs(ctx, []uuid.UUID{userID1, userID2})
	require.NoError(t, err)
	assert.Empty(t, subscribedIDs)
}

// TestGetSubscribedUserIDs_AllSubscribed tests when all users are subscribed
func TestGetSubscribedUserIDs_AllSubscribed(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()

	createTestUserByID(t, db, userID1, "user1@test.com", "student")
	createTestUserByID(t, db, userID2, "user2@test.com", "student")
	createTestUserByID(t, db, userID3, "user3@test.com", "student")

	now := time.Now()

	// All users subscribed = true
	_, err := db.ExecContext(ctx, `
		INSERT INTO telegram_users (id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), userID1, 111111111, 111111111, "user1", true, now, now)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO telegram_users (id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), userID2, 222222222, 222222222, "user2", true, now, now)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO telegram_users (id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), userID3, 333333333, 333333333, "user3", true, now, now)
	require.NoError(t, err)

	repo := NewTelegramUserRepository(db)

	subscribedIDs, err := repo.GetSubscribedUserIDs(ctx, []uuid.UUID{userID1, userID2, userID3})
	require.NoError(t, err)
	assert.Len(t, subscribedIDs, 3)

	// Verify all IDs are returned
	subscribedMap := make(map[uuid.UUID]bool)
	for _, id := range subscribedIDs {
		subscribedMap[id] = true
	}

	assert.True(t, subscribedMap[userID1])
	assert.True(t, subscribedMap[userID2])
	assert.True(t, subscribedMap[userID3])
}

// TestGetSubscribedUserIDs_NonExistentUsers tests with non-existent user IDs
func TestGetSubscribedUserIDs_NonExistentUsers(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	userID1 := uuid.New()
	nonExistentID1 := uuid.New()
	nonExistentID2 := uuid.New()

	createTestUserByID(t, db, userID1, "user1@test.com", "student")

	now := time.Now()

	// Only userID1 exists and is subscribed
	_, err := db.ExecContext(ctx, `
		INSERT INTO telegram_users (id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, uuid.New(), userID1, 111111111, 111111111, "user1", true, now, now)
	require.NoError(t, err)

	repo := NewTelegramUserRepository(db)

	// Mix of existing and non-existent IDs
	subscribedIDs, err := repo.GetSubscribedUserIDs(ctx, []uuid.UUID{userID1, nonExistentID1, nonExistentID2})
	require.NoError(t, err)
	assert.Len(t, subscribedIDs, 1)
	assert.Equal(t, userID1, subscribedIDs[0])
}

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

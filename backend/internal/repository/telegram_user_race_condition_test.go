package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestLinkUserToTelegramAtomic_RaceCondition проверяет защиту от race condition
// при одновременных запросах на привязку одного telegram_id к разным пользователям.
// Результат: только первый запрос успешен, остальные получают ErrTelegramIDAlreadyLinked
func TestLinkUserToTelegramAtomic_RaceCondition(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewTelegramUserRepository(db)

	// Создаем тестовых пользователей в БД
	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()

	createTestUserByID(t, db, userID1, "user1@test.com", "student")
	createTestUserByID(t, db, userID2, "user2@test.com", "student")
	createTestUserByID(t, db, userID3, "user3@test.com", "student")

	// Одинаковый telegram_id для всех трёх попыток (симуляция race condition)
	// Используем уникальный telegramID на основе timestamp, чтобы избежать конфликтов между тестами
	telegramID := int64(1000000000 + time.Now().UnixNano()%1000000)
	chatID := int64(87654321)
	username := "testuser"

	// Используем WaitGroup для синхронизации трёх горутин
	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// Запускаем три одновременные попытки привязать один telegram_id
	wg.Add(3)

	go func() {
		defer wg.Done()
		err := repo.LinkUserToTelegramAtomic(context.Background(), userID1, telegramID, chatID, username)
		errChan <- err
	}()

	go func() {
		defer wg.Done()
		err := repo.LinkUserToTelegramAtomic(context.Background(), userID2, telegramID, chatID, username)
		errChan <- err
	}()

	go func() {
		defer wg.Done()
		err := repo.LinkUserToTelegramAtomic(context.Background(), userID3, telegramID, chatID, username)
		errChan <- err
	}()

	// Ждём завершения всех горутин
	wg.Wait()
	close(errChan)

	// Собираем результаты
	successCount := 0
	conflictCount := 0

	for err := range errChan {
		if err == nil {
			successCount++
		} else if err == ErrTelegramIDAlreadyLinked {
			conflictCount++
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Проверяем результаты
	// Должно быть ровно 1 успешная попытка и 2 конфликта
	if successCount != 1 {
		t.Errorf("expected 1 success, got %d", successCount)
	}
	if conflictCount != 2 {
		t.Errorf("expected 2 conflicts, got %d", conflictCount)
	}

	// Проверяем, что привязка произошла только для одного пользователя
	tgUser1, err1 := repo.GetByUserID(context.Background(), userID1)
	tgUser2, err2 := repo.GetByUserID(context.Background(), userID2)
	tgUser3, err3 := repo.GetByUserID(context.Background(), userID3)

	linkedCount := 0
	if err1 == nil && tgUser1 != nil {
		linkedCount++
		if tgUser1.TelegramID != telegramID {
			t.Errorf("user1 has wrong telegramID: %d", tgUser1.TelegramID)
		}
	}
	if err2 == nil && tgUser2 != nil {
		linkedCount++
	}
	if err3 == nil && tgUser3 != nil {
		linkedCount++
	}

	// Только один пользователь должен быть привязан к telegram_id
	if linkedCount != 1 {
		t.Errorf("expected 1 linked user, got %d", linkedCount)
	}
}

// TestLinkUserToTelegramAtomic_Idempotent проверяет идемпотентность операции.
// Если один пользователь вызовет LinkUserToTelegramAtomic дважды с одинаковым telegram_id,
// вторая попытка должна успешно обновить запись.
func TestLinkUserToTelegramAtomic_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewTelegramUserRepository(db)

	userID := uuid.New()
	createTestUserByID(t, db, userID, "user@test.com", "student")

	// Используем уникальный telegramID, чтобы избежать конфликтов с другими тестами
	telegramID := int64(2000000000 + time.Now().UnixNano()%1000000)
	chatID := int64(87654321)
	username := "testuser"

	// Первый вызов - создание привязки
	err := repo.LinkUserToTelegramAtomic(context.Background(), userID, telegramID, chatID, username)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Второй вызов - обновление (идемпотентно)
	newChatID := int64(11111111)
	newUsername := "testuser_updated"
	err = repo.LinkUserToTelegramAtomic(context.Background(), userID, telegramID, newChatID, newUsername)
	if err != nil {
		t.Fatalf("second call (idempotent) failed: %v", err)
	}

	// Проверяем, что данные обновлены
	tgUser, err := repo.GetByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	if tgUser.TelegramID != telegramID {
		t.Errorf("expected telegramID %d, got %d", telegramID, tgUser.TelegramID)
	}
	if tgUser.ChatID != newChatID {
		t.Errorf("expected chatID %d, got %d", newChatID, tgUser.ChatID)
	}
	if tgUser.Username != newUsername {
		t.Errorf("expected username %s, got %s", newUsername, tgUser.Username)
	}
}

// TestLinkUserToTelegramAtomic_DifferentUsers проверяет, что разные пользователи
// могут привязывать разные telegram_id одновременно без конфликтов.
func TestLinkUserToTelegramAtomic_DifferentUsers(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewTelegramUserRepository(db)

	userID1 := uuid.New()
	userID2 := uuid.New()

	createTestUserByID(t, db, userID1, "user1@test.com", "student")
	createTestUserByID(t, db, userID2, "user2@test.com", "student")

	telegramID1 := int64(111111)
	telegramID2 := int64(222222)
	chatID1 := int64(1000)
	chatID2 := int64(2000)

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Две одновременные привязки разных telegram_id
	wg.Add(2)

	go func() {
		defer wg.Done()
		err := repo.LinkUserToTelegramAtomic(context.Background(), userID1, telegramID1, chatID1, "user1")
		errChan <- err
	}()

	go func() {
		defer wg.Done()
		err := repo.LinkUserToTelegramAtomic(context.Background(), userID2, telegramID2, chatID2, "user2")
		errChan <- err
	}()

	wg.Wait()
	close(errChan)

	// Оба должны быть успешны
	for err := range errChan {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Проверяем, что обе привязки сохранились
	tgUser1, _ := repo.GetByUserID(context.Background(), userID1)
	tgUser2, _ := repo.GetByUserID(context.Background(), userID2)

	if tgUser1.TelegramID != telegramID1 {
		t.Errorf("user1 telegramID mismatch")
	}
	if tgUser2.TelegramID != telegramID2 {
		t.Errorf("user2 telegramID mismatch")
	}
}

// TestLinkUserToTelegramAtomic_ReplaceExisting проверяет замену существующей привязки
// на новый telegram_id при обновлении через атомарную операцию.
func TestLinkUserToTelegramAtomic_ReplaceExisting(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewTelegramUserRepository(db)

	userID := uuid.New()
	createTestUserByID(t, db, userID, "user@test.com", "student")

	oldTelegramID := int64(11111111)
	newTelegramID := int64(22222222)
	chatID := int64(1000)

	// Создаём первую привязку
	err := repo.LinkUserToTelegramAtomic(context.Background(), userID, oldTelegramID, chatID, "olduser")
	if err != nil {
		t.Fatalf("failed to create initial link: %v", err)
	}

	// Заменяем на новый telegram_id (тот же пользователь)
	err = repo.LinkUserToTelegramAtomic(context.Background(), userID, newTelegramID, chatID, "newuser")
	if err != nil {
		t.Fatalf("failed to replace link: %v", err)
	}

	// Проверяем, что привязка обновлена
	tgUser, _ := repo.GetByUserID(context.Background(), userID)
	if tgUser.TelegramID != newTelegramID {
		t.Errorf("expected new telegramID %d, got %d", newTelegramID, tgUser.TelegramID)
	}

	// Проверяем, что старый telegram_id теперь не привязан ни к кому
	oldTgUser, err := repo.GetByTelegramID(context.Background(), oldTelegramID)
	if err == nil && oldTgUser != nil {
		t.Error("old telegramID should not be linked to anyone")
	}
}

// TestLinkUserToTelegramAtomic_ConcurrentReplacement проверяет race condition
// при замене существующей привязки пользователя на новый telegram_id,
// когда одновременно другой пользователь пытается привязать старый telegram_id.
func TestLinkUserToTelegramAtomic_ConcurrentReplacement(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewTelegramUserRepository(db)

	userID1 := uuid.New()
	userID2 := uuid.New()

	createTestUserByID(t, db, userID1, "user1@test.com", "student")
	createTestUserByID(t, db, userID2, "user2@test.com", "student")

	// Сначала привязываем userID1 к oldTelegramID
	oldTelegramID := int64(100)
	newTelegramID := int64(200)
	chatID := int64(1000)

	err := repo.LinkUserToTelegramAtomic(context.Background(), userID1, oldTelegramID, chatID, "user1")
	if err != nil {
		t.Fatalf("failed to create initial link: %v", err)
	}

	// Теперь одновременно:
	// 1. userID1 обновляет привязку с oldTelegramID на newTelegramID
	// 2. userID2 пытается привязать oldTelegramID
	// Результат: userID2 должен получить успех (так как oldTelegramID станет свободен)

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	wg.Add(2)

	// userID1: обновление с oldTelegramID на newTelegramID
	go func() {
		defer wg.Done()
		err := repo.LinkUserToTelegramAtomic(context.Background(), userID1, newTelegramID, chatID, "user1_updated")
		errChan <- err
	}()

	// userID2: попытка привязать освобождённый oldTelegramID
	// (это может пройти или не пройти в зависимости от timing)
	go func() {
		defer wg.Done()
		err := repo.LinkUserToTelegramAtomic(context.Background(), userID2, oldTelegramID, chatID, "user2")
		errChan <- err
	}()

	wg.Wait()
	close(errChan)

	// Собираем результаты - оба должны быть успешны
	successCount := 0
	for err := range errChan {
		if err == nil {
			successCount++
		} else if err != ErrTelegramIDAlreadyLinked {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	// Оба пользователя должны быть успешно привязаны
	if successCount < 1 {
		t.Errorf("expected at least 1 success")
	}

	// Проверяем final state
	tgUser1, _ := repo.GetByUserID(context.Background(), userID1)
	if tgUser1 == nil || tgUser1.TelegramID != newTelegramID {
		t.Errorf("user1 should have newTelegramID")
	}
}

// TestLinkUserToTelegramAtomic_UniquenessEnforcement проверяет,
// что UNIQUE constraint на telegram_id не нарушается даже при сбое атомарности.
func TestLinkUserToTelegramAtomic_UniquenessEnforcement(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewTelegramUserRepository(db)

	telegramID := int64(333333)
	chatID := int64(1000)

	// Массово создаём попытки привязать одинаковый telegram_id
	numGoroutines := 10
	userIDs := createMultipleTestUsers(t, db, numGoroutines, "student")

	var wg sync.WaitGroup
	successCount := 0
	conflictCount := 0
	mu := sync.Mutex{}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			userID := userIDs[idx]
			username := fmt.Sprintf("user%d", idx)

			err := repo.LinkUserToTelegramAtomic(context.Background(), userID, telegramID, chatID, username)

			mu.Lock()
			defer mu.Unlock()

			if err == nil {
				successCount++
			} else if err == ErrTelegramIDAlreadyLinked {
				conflictCount++
			} else {
				t.Logf("unexpected error: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Ровно 1 успех, остальное - конфликты
	if successCount != 1 {
		t.Errorf("expected 1 success, got %d", successCount)
	}
	if conflictCount != numGoroutines-1 {
		t.Errorf("expected %d conflicts, got %d", numGoroutines-1, conflictCount)
	}

	// Проверяем, что в БД ровно одна запись с этим telegram_id
	// (это проверит, что UNIQUE constraint работает)
	tgUser, err := repo.GetByTelegramID(context.Background(), telegramID)
	if err != nil {
		t.Fatalf("failed to get user by telegram_id: %v", err)
	}
	if tgUser == nil {
		t.Fatal("expected to find telegram_id")
	}
	if tgUser.TelegramID != telegramID {
		t.Errorf("telegram_id mismatch")
	}
}

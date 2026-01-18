package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
)

// SyncTestUserRepository - мок для синхронизационных тестов
type SyncTestUserRepository struct {
	mock.Mock
}

func (m *SyncTestUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *SyncTestUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *SyncTestUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *SyncTestUserRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *SyncTestUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *SyncTestUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *SyncTestUserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *SyncTestUserRepository) List(ctx context.Context, roleFilter *models.UserRole) ([]*models.User, error) {
	args := m.Called(ctx, roleFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *SyncTestUserRepository) ListWithPagination(ctx context.Context, role *models.UserRole, offset, limit int) ([]*models.User, int, error) {
	args := m.Called(ctx, role, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Int(1), args.Error(2)
}

func (m *SyncTestUserRepository) Exists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *SyncTestUserRepository) UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error {
	args := m.Called(ctx, userID, username)
	return args.Error(0)
}

// SyncTestTelegramTokenRepository - мок для синхронизационных тестов
type SyncTestTelegramTokenRepository struct {
	mock.Mock
}

func (m *SyncTestTelegramTokenRepository) SaveToken(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error {
	args := m.Called(ctx, token, userID, expiresAt)
	return args.Error(0)
}

func (m *SyncTestTelegramTokenRepository) GetTokenUser(ctx context.Context, token string) (uuid.UUID, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *SyncTestTelegramTokenRepository) DeleteToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *SyncTestTelegramTokenRepository) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *SyncTestTelegramTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// TestLinkUserAccount_SyncsTelegramUsername - проверяет что при привязке Telegram обновляется users.telegram_username
func TestLinkUserAccount_SyncsTelegramUsername(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	token := "test-token-123"
	telegramID := int64(12345)
	chatID := int64(67890)
	username := "testuser"

	// Готовим моки
	mockUserRepo := new(SyncTestUserRepository)
	mockTelegramUserRepo := new(MockTelegramUserRepository)
	mockTelegramTokenRepo := new(SyncTestTelegramTokenRepository)

	// Expect вызовы в LinkUserAccount:
	// 1. Проверка токена в БД
	mockTelegramTokenRepo.On("GetTokenUser", ctx, token).Return(userID, nil)

	// 2. Проверка что этот Telegram ID не привязан к другому пользователю
	mockTelegramUserRepo.On("GetByTelegramID", ctx, telegramID).Return(nil, repository.ErrTelegramUserNotFound)

	// 3. Привязка в telegram_users
	mockTelegramUserRepo.On("LinkUserToTelegram", ctx, userID, telegramID, chatID, username).Return(nil)

	// 4. Синхронизация в users.telegram_username (НАША НОВАЯ ФУНКЦИОНАЛЬНОСТЬ!)
	mockUserRepo.On("UpdateTelegramUsername", ctx, userID, username).Return(nil)

	// 5. Удаление токена
	mockTelegramTokenRepo.On("DeleteToken", ctx, token).Return(nil)

	// Создаём сервис
	svc := NewTelegramService(
		mockTelegramUserRepo,
		mockTelegramTokenRepo,
		mockUserRepo,
		nil,
		0,
	)

	// Вызываем LinkUserAccount
	err := svc.LinkUserAccount(ctx, token, telegramID, chatID, username)

	// Проверяем результаты
	require.NoError(t, err)

	// Проверяем что UpdateTelegramUsername был вызван с правильными параметрами
	mockUserRepo.AssertCalled(t, "UpdateTelegramUsername", ctx, userID, username)
	mockUserRepo.AssertNumberOfCalls(t, "UpdateTelegramUsername", 1)
}

// TestLinkUserAccount_SyncFailureDoesNotBlockLinking - проверяет что ошибка синхронизации не блокирует привязку
func TestLinkUserAccount_SyncFailureDoesNotBlockLinking(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	token := "test-token-123"
	telegramID := int64(12345)
	chatID := int64(67890)
	username := "testuser"

	// Готовим моки
	mockUserRepo := new(SyncTestUserRepository)
	mockTelegramUserRepo := new(MockTelegramUserRepository)
	mockTelegramTokenRepo := new(SyncTestTelegramTokenRepository)

	// Ожидаемые вызовы
	mockTelegramTokenRepo.On("GetTokenUser", ctx, token).Return(userID, nil)
	mockTelegramUserRepo.On("GetByTelegramID", ctx, telegramID).Return(nil, repository.ErrTelegramUserNotFound)
	mockTelegramUserRepo.On("LinkUserToTelegram", ctx, userID, telegramID, chatID, username).Return(nil)

	// Синхронизация FAILS, но это не должно заблокировать операцию
	mockUserRepo.On("UpdateTelegramUsername", ctx, userID, username).Return(assert.AnError)

	mockTelegramTokenRepo.On("DeleteToken", ctx, token).Return(nil)

	// Создаём сервис
	svc := NewTelegramService(
		mockTelegramUserRepo,
		mockTelegramTokenRepo,
		mockUserRepo,
		nil,
		0,
	)

	// Вызываем LinkUserAccount - ДОЛЖНА ВЕРНУТЬ nil даже если синхронизация не удалась
	err := svc.LinkUserAccount(ctx, token, telegramID, chatID, username)

	// Привязка должна быть успешной, несмотря на ошибку синхронизации
	require.NoError(t, err)

	// Проверяем что все методы были вызваны
	mockUserRepo.AssertCalled(t, "UpdateTelegramUsername", ctx, userID, username)
	mockTelegramUserRepo.AssertCalled(t, "LinkUserToTelegram", ctx, userID, telegramID, chatID, username)
}

// TestSetUserTelegram_SyncsTelegramUsername - проверяет синхронизацию при админском установлении Telegram
// Использует атомарную операцию для защиты от race condition
func TestSetUserTelegram_SyncsTelegramUsername(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	telegramID := int64(12345)
	chatID := int64(67890)
	username := "testuser"

	// Готовим моки
	mockUserRepo := new(SyncTestUserRepository)
	mockTelegramUserRepo := new(MockTelegramUserRepository)
	mockTelegramTokenRepo := new(SyncTestTelegramTokenRepository)

	// Ожидаемые вызовы - теперь используем атомарный метод
	mockTelegramUserRepo.On("LinkUserToTelegramAtomic", ctx, userID, telegramID, chatID, username).Return(nil)

	// Синхронизация
	mockUserRepo.On("UpdateTelegramUsername", ctx, userID, username).Return(nil)

	// Создаём сервис
	svc := NewTelegramService(
		mockTelegramUserRepo,
		mockTelegramTokenRepo,
		mockUserRepo,
		nil,
		0,
	)

	// Вызываем SetUserTelegram
	err := svc.SetUserTelegram(ctx, userID, telegramID, chatID, username)

	// Проверяем результаты
	require.NoError(t, err)

	// Проверяем что атомарный метод был вызван
	mockTelegramUserRepo.AssertCalled(t, "LinkUserToTelegramAtomic", ctx, userID, telegramID, chatID, username)

	// Проверяем что UpdateTelegramUsername был вызван
	mockUserRepo.AssertCalled(t, "UpdateTelegramUsername", ctx, userID, username)
}

// TestUnlinkUser_ClearsTelegramUsername - проверяет что при отвязке Telegram очищается users.telegram_username
func TestUnlinkUser_ClearsTelegramUsername(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	// Готовим моки
	mockUserRepo := new(SyncTestUserRepository)
	mockTelegramUserRepo := new(MockTelegramUserRepository)
	mockTelegramTokenRepo := new(SyncTestTelegramTokenRepository)

	// Telegram привязка существует
	telegramUser := &models.TelegramUser{
		ID:         uuid.New(),
		UserID:     userID,
		TelegramID: 12345,
		ChatID:     67890,
		Username:   "testuser",
		Subscribed: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockTelegramUserRepo.On("GetByUserID", ctx, userID).Return(telegramUser, nil)
	mockTelegramUserRepo.On("UnlinkTelegram", ctx, userID).Return(nil)

	// При отвязке должно очиститься telegram_username (пустая строка)
	mockUserRepo.On("UpdateTelegramUsername", ctx, userID, "").Return(nil)

	// Создаём сервис
	svc := NewTelegramService(
		mockTelegramUserRepo,
		mockTelegramTokenRepo,
		mockUserRepo,
		nil,
		0,
	)

	// Вызываем UnlinkUser
	err := svc.UnlinkUser(ctx, userID)

	// Проверяем результаты
	require.NoError(t, err)

	// Проверяем что UpdateTelegramUsername был вызван с пустой строкой
	mockUserRepo.AssertCalled(t, "UpdateTelegramUsername", ctx, userID, "")
}

// TestUnlinkUser_ClearingFailureDoesNotBlockUnlinking - проверяет что ошибка при очистке не блокирует отвязку
func TestUnlinkUser_ClearingFailureDoesNotBlockUnlinking(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	// Готовим моки
	mockUserRepo := new(SyncTestUserRepository)
	mockTelegramUserRepo := new(MockTelegramUserRepository)
	mockTelegramTokenRepo := new(SyncTestTelegramTokenRepository)

	// Telegram привязка существует
	telegramUser := &models.TelegramUser{
		ID:         uuid.New(),
		UserID:     userID,
		TelegramID: 12345,
		ChatID:     67890,
		Username:   "testuser",
		Subscribed: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockTelegramUserRepo.On("GetByUserID", ctx, userID).Return(telegramUser, nil)
	mockTelegramUserRepo.On("UnlinkTelegram", ctx, userID).Return(nil)

	// Очистка FAILS, но это не должно заблокировать отвязку
	mockUserRepo.On("UpdateTelegramUsername", ctx, userID, "").Return(assert.AnError)

	// Создаём сервис
	svc := NewTelegramService(
		mockTelegramUserRepo,
		mockTelegramTokenRepo,
		mockUserRepo,
		nil,
		0,
	)

	// Вызываем UnlinkUser - ДОЛЖНА ВЕРНУТЬ nil даже если очистка не удалась
	err := svc.UnlinkUser(ctx, userID)

	// Отвязка должна быть успешной, несмотря на ошибку очистки
	require.NoError(t, err)

	// Проверяем что оба метода были вызваны
	mockTelegramUserRepo.AssertCalled(t, "UnlinkTelegram", ctx, userID)
	mockUserRepo.AssertCalled(t, "UpdateTelegramUsername", ctx, userID, "")
}

// TestSyncFlow_EndToEnd - интеграционный тест синхронизации
func TestSyncFlow_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	// Этот тест демонстрирует полный цикл:
	// 1. Пользователь создается (без telegram_username)
	// 2. Привязывает Telegram (telegram_username синхронизируется)
	// 3. Отвязывает Telegram (telegram_username очищается)

	// Поскольку у нас есть моки, этот тест использует их:
	ctx := context.Background()
	userID := uuid.New()
	token := "link-token"
	telegramID := int64(999)
	chatID := int64(888)
	username := "synctest"

	mockUserRepo := new(SyncTestUserRepository)
	mockTelegramUserRepo := new(MockTelegramUserRepository)
	mockTelegramTokenRepo := new(SyncTestTelegramTokenRepository)

	// === STEP 1: Привязка ===
	mockTelegramTokenRepo.On("GetTokenUser", ctx, token).Return(userID, nil)
	mockTelegramUserRepo.On("GetByTelegramID", ctx, telegramID).Return(nil, repository.ErrTelegramUserNotFound)
	mockTelegramUserRepo.On("LinkUserToTelegram", ctx, userID, telegramID, chatID, username).Return(nil)
	mockUserRepo.On("UpdateTelegramUsername", ctx, userID, username).Return(nil)
	mockTelegramTokenRepo.On("DeleteToken", ctx, token).Return(nil)

	svc := NewTelegramService(
		mockTelegramUserRepo,
		mockTelegramTokenRepo,
		mockUserRepo,
		nil,
		0,
	)

	err := svc.LinkUserAccount(ctx, token, telegramID, chatID, username)
	require.NoError(t, err)

	// === STEP 2: Отвязка ===
	telegramUser := &models.TelegramUser{
		ID:         uuid.New(),
		UserID:     userID,
		TelegramID: telegramID,
		ChatID:     chatID,
		Username:   username,
		Subscribed: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	mockTelegramUserRepo.On("GetByUserID", ctx, userID).Return(telegramUser, nil)
	mockTelegramUserRepo.On("UnlinkTelegram", ctx, userID).Return(nil)
	mockUserRepo.On("UpdateTelegramUsername", ctx, userID, "").Return(nil)

	err = svc.UnlinkUser(ctx, userID)
	require.NoError(t, err)

	// Проверяем что все методы были вызваны правильное количество раз
	mockUserRepo.AssertNumberOfCalls(t, "UpdateTelegramUsername", 2) // 1 для привязки, 1 для отвязки
}

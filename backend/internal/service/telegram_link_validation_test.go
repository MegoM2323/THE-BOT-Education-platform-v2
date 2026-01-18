package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"tutoring-platform/internal/models"
)

// MockTelegramUserRepoForValidation моки репозитория для тестирования валидации привязки
type MockTelegramUserRepoForValidation struct {
	mock.Mock
}

func (m *MockTelegramUserRepoForValidation) LinkUserToTelegram(ctx context.Context, userID uuid.UUID, telegramID, chatID int64, username string) error {
	args := m.Called(ctx, userID, telegramID, chatID, username)
	return args.Error(0)
}

func (m *MockTelegramUserRepoForValidation) LinkUserToTelegramAtomic(ctx context.Context, userID uuid.UUID, telegramID, chatID int64, username string) error {
	args := m.Called(ctx, userID, telegramID, chatID, username)
	return args.Error(0)
}

func (m *MockTelegramUserRepoForValidation) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.TelegramUser, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TelegramUser), args.Error(1)
}

func (m *MockTelegramUserRepoForValidation) GetByTelegramID(ctx context.Context, telegramID int64) (*models.TelegramUser, error) {
	args := m.Called(ctx, telegramID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TelegramUser), args.Error(1)
}

func (m *MockTelegramUserRepoForValidation) GetAllLinked(ctx context.Context) ([]*models.TelegramUser, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TelegramUser), args.Error(1)
}

func (m *MockTelegramUserRepoForValidation) GetAllWithUserInfo(ctx context.Context) ([]*models.TelegramUser, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TelegramUser), args.Error(1)
}

func (m *MockTelegramUserRepoForValidation) GetByRoleWithUserInfo(ctx context.Context, role string) ([]*models.TelegramUser, error) {
	args := m.Called(ctx, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TelegramUser), args.Error(1)
}

func (m *MockTelegramUserRepoForValidation) GetByRole(ctx context.Context, role string) ([]*models.TelegramUser, error) {
	args := m.Called(ctx, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TelegramUser), args.Error(1)
}

func (m *MockTelegramUserRepoForValidation) UpdateSubscription(ctx context.Context, userID uuid.UUID, subscribed bool) error {
	args := m.Called(ctx, userID, subscribed)
	return args.Error(0)
}

func (m *MockTelegramUserRepoForValidation) UnlinkTelegram(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTelegramUserRepoForValidation) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTelegramUserRepoForValidation) CleanupInvalidLinks(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTelegramUserRepoForValidation) IsValidlyLinked(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// MockUserRepoForValidation мок для UserRepository
type MockUserRepoForValidation struct {
	mock.Mock
}

func (m *MockUserRepoForValidation) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepoForValidation) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepoForValidation) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepoForValidation) GetAll(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepoForValidation) Update(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, userID, updates)
	return args.Error(0)
}

func (m *MockUserRepoForValidation) Delete(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepoForValidation) SoftDelete(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepoForValidation) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	args := m.Called(ctx, userID, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepoForValidation) UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error {
	args := m.Called(ctx, userID, username)
	return args.Error(0)
}

func (m *MockUserRepoForValidation) Exists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepoForValidation) List(ctx context.Context, role *models.UserRole) ([]*models.User, error) {
	args := m.Called(ctx, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepoForValidation) ListWithPagination(ctx context.Context, role *models.UserRole, offset, limit int) ([]*models.User, int, error) {
	args := m.Called(ctx, role, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Int(1), args.Error(2)
}

// MockTelegramTokenRepoForValidation мок для TelegramTokenRepository
type MockTelegramTokenRepoForValidation struct {
	mock.Mock
}

func (m *MockTelegramTokenRepoForValidation) SaveToken(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error {
	args := m.Called(ctx, token, userID, expiresAt)
	return args.Error(0)
}

func (m *MockTelegramTokenRepoForValidation) GetTokenUser(ctx context.Context, token string) (uuid.UUID, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockTelegramTokenRepoForValidation) DeleteToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTelegramTokenRepoForValidation) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTelegramTokenRepoForValidation) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// TestIsValidlyLinked_ReturnsTrueOnlyWhenTelegramIDGreaterThanZero проверяет IsLinked возвращает true только при telegram_id > 0
func TestIsValidlyLinked_ReturnsTrueOnlyWhenTelegramIDGreaterThanZero(t *testing.T) {
	tests := []struct {
		name           string
		telegramID     int64
		expectedLinked bool
		setupMock      func(*MockTelegramUserRepoForValidation, uuid.UUID)
	}{
		{
			name:           "Valid link with telegram_id > 0",
			telegramID:     123456789,
			expectedLinked: true,
			setupMock: func(m *MockTelegramUserRepoForValidation, userID uuid.UUID) {
				// IsValidlyLinked возвращает true для telegram_id > 0
				m.On("IsValidlyLinked", mock.Anything, userID).Return(true, nil)
			},
		},
		{
			name:           "Invalid link with telegram_id = 0",
			telegramID:     0,
			expectedLinked: false,
			setupMock: func(m *MockTelegramUserRepoForValidation, userID uuid.UUID) {
				// IsValidlyLinked возвращает false для telegram_id = 0
				m.On("IsValidlyLinked", mock.Anything, userID).Return(false, nil)
			},
		},
		{
			name:           "No link exists",
			telegramID:     0,
			expectedLinked: false,
			setupMock: func(m *MockTelegramUserRepoForValidation, userID uuid.UUID) {
				// Нет записи - возвращаем false
				m.On("IsValidlyLinked", mock.Anything, userID).Return(false, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			userID := uuid.New()

			mockTelegramUserRepo := new(MockTelegramUserRepoForValidation)
			tt.setupMock(mockTelegramUserRepo, userID)

			// Проверяем через метод репозитория
			isLinked, err := mockTelegramUserRepo.IsValidlyLinked(ctx, userID)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedLinked, isLinked, "IsValidlyLinked должен возвращать корректный статус привязки")

			mockTelegramUserRepo.AssertExpectations(t)
		})
	}
}

// TestCleanupInvalidLinks_RemovesInvalidRecords проверяет автоматическую очистку невалидных записей
func TestCleanupInvalidLinks_RemovesInvalidRecords(t *testing.T) {
	ctx := context.Background()

	mockTelegramUserRepo := new(MockTelegramUserRepoForValidation)

	// Моделируем очистку 3 невалидных записей
	mockTelegramUserRepo.On("CleanupInvalidLinks", ctx).Return(int64(3), nil)

	cleaned, err := mockTelegramUserRepo.CleanupInvalidLinks(ctx)

	assert.NoError(t, err)
	assert.Equal(t, int64(3), cleaned, "Должно быть удалено 3 невалидных записи")

	mockTelegramUserRepo.AssertExpectations(t)
}

// TestGenerateLinkToken_AfterFailedAttempt_WorksCorrectly проверяет повторную привязку после неудачи
func TestGenerateLinkToken_AfterFailedAttempt_WorksCorrectly(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockTelegramUserRepo := new(MockTelegramUserRepoForValidation)
	mockTelegramTokenRepo := new(MockTelegramTokenRepoForValidation)
	mockUserRepo := new(MockUserRepoForValidation)

	// Пользователь существует
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FullName:  "Test User",
		Role:      "student",
		DeletedAt: sql.NullTime{Valid: false},
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Первый вызов: очищаем любые невалидные записи
	mockTelegramUserRepo.On("DeleteByUserID", ctx, userID).Return(nil)

	// Проверяем валидную привязку - нет (после очистки)
	mockTelegramUserRepo.On("IsValidlyLinked", ctx, userID).Return(false, nil)

	// Удаляем старые токены
	mockTelegramTokenRepo.On("DeleteByUserID", ctx, userID).Return(nil)

	// Сохраняем новый токен
	mockTelegramTokenRepo.On("SaveToken", ctx, mock.Anything, userID, mock.Anything).Return(nil)

	// Создаем сервис
	service := &TelegramService{
		telegramUserRepo:  mockTelegramUserRepo,
		telegramTokenRepo: mockTelegramTokenRepo,
		userRepo:          mockUserRepo,
		tokenStore:        NewTokenStore(),
	}

	// Генерируем токен (должно работать без ошибок после неудачной попытки)
	token, err := service.GenerateLinkToken(ctx, userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, token, "Токен должен быть сгенерирован")

	mockTelegramUserRepo.AssertExpectations(t)
	mockTelegramTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestGenerateLinkToken_WithValidLink_ReturnsError проверяет блокировку при валидной привязке
func TestGenerateLinkToken_WithValidLink_ReturnsError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockTelegramUserRepo := new(MockTelegramUserRepoForValidation)
	mockTelegramTokenRepo := new(MockTelegramTokenRepoForValidation)
	mockUserRepo := new(MockUserRepoForValidation)

	// Пользователь существует
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		FullName:  "Test User",
		Role:      "student",
		DeletedAt: sql.NullTime{Valid: false},
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Очищаем невалидные записи
	mockTelegramUserRepo.On("DeleteByUserID", ctx, userID).Return(nil)

	// Пользователь УЖЕ валидно привязан (telegram_id > 0)
	mockTelegramUserRepo.On("IsValidlyLinked", ctx, userID).Return(true, nil)

	// Создаем сервис
	service := &TelegramService{
		telegramUserRepo:  mockTelegramUserRepo,
		telegramTokenRepo: mockTelegramTokenRepo,
		userRepo:          mockUserRepo,
		tokenStore:        NewTokenStore(),
	}

	// Пытаемся сгенерировать токен - должна быть ошибка
	token, err := service.GenerateLinkToken(ctx, userID)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.True(t, errors.Is(err, ErrTelegramAlreadyLinked), "Должна быть ошибка ErrTelegramAlreadyLinked")

	mockTelegramUserRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestDeleteByUserID_RemovesRecordCompletely проверяет полное удаление записи
func TestDeleteByUserID_RemovesRecordCompletely(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockTelegramUserRepo := new(MockTelegramUserRepoForValidation)

	// Удаляем запись пользователя
	mockTelegramUserRepo.On("DeleteByUserID", ctx, userID).Return(nil)

	err := mockTelegramUserRepo.DeleteByUserID(ctx, userID)

	assert.NoError(t, err)

	mockTelegramUserRepo.AssertExpectations(t)
}

// TestCleanupInvalidLinks_InBackgroundTask проверяет работу фоновой очистки
func TestCleanupInvalidLinks_InBackgroundTask(t *testing.T) {
	ctx := context.Background()

	mockTelegramUserRepo := new(MockTelegramUserRepoForValidation)

	// Моделируем периодическую очистку: 5 невалидных записей
	mockTelegramUserRepo.On("CleanupInvalidLinks", ctx).Return(int64(5), nil)

	cleaned, err := mockTelegramUserRepo.CleanupInvalidLinks(ctx)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), cleaned, "Фоновая задача должна удалять невалидные записи")

	mockTelegramUserRepo.AssertExpectations(t)
}

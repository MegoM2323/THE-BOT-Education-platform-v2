package service

import (
	"context"
	"errors"
	"testing"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateUserEmailValidation тесты для валидации email при обновлении пользователя
func TestUpdateUserEmailValidation(t *testing.T) {
	tests := []struct {
		name          string
		userID        uuid.UUID
		existingUsers map[uuid.UUID]*models.User
		updateEmail   *string
		expectedError error
		expectSuccess bool
	}{
		{
			name:   "valid email format update",
			userID: uuid.New(),
			existingUsers: map[uuid.UUID]*models.User{
				uuid.New(): {
					ID:       uuid.New(),
					Email:    "user1@example.com",
					FirstName: "User 1", LastName: "Lastname",
					Role:     models.RoleStudent,
				},
			},
			updateEmail:   ptrString("newemail@example.com"),
			expectedError: nil,
			expectSuccess: true,
		},
		{
			name:   "invalid email format - missing @",
			userID: uuid.New(),
			existingUsers: map[uuid.UUID]*models.User{
				uuid.New(): {
					ID:       uuid.New(),
					Email:    "user1@example.com",
					FirstName: "User 1", LastName: "Lastname",
					Role:     models.RoleStudent,
				},
			},
			updateEmail:   ptrString("invalidemail.com"),
			expectedError: models.ErrInvalidEmail,
			expectSuccess: false,
		},
		{
			name:   "invalid email format - missing domain",
			userID: uuid.New(),
			existingUsers: map[uuid.UUID]*models.User{
				uuid.New(): {
					ID:       uuid.New(),
					Email:    "user1@example.com",
					FirstName: "User 1", LastName: "Lastname",
					Role:     models.RoleStudent,
				},
			},
			updateEmail:   ptrString("user@"),
			expectedError: models.ErrInvalidEmail,
			expectSuccess: false,
		},
		{
			name:   "invalid email format - @ at beginning",
			userID: uuid.New(),
			existingUsers: map[uuid.UUID]*models.User{
				uuid.New(): {
					ID:       uuid.New(),
					Email:    "user1@example.com",
					FirstName: "User 1", LastName: "Lastname",
					Role:     models.RoleStudent,
				},
			},
			updateEmail:   ptrString("@example.com"),
			expectedError: models.ErrInvalidEmail,
			expectSuccess: false,
		},
		{
			name:   "invalid email format - missing dot in domain",
			userID: uuid.New(),
			existingUsers: map[uuid.UUID]*models.User{
				uuid.New(): {
					ID:       uuid.New(),
					Email:    "user1@example.com",
					FirstName: "User 1", LastName: "Lastname",
					Role:     models.RoleStudent,
				},
			},
			updateEmail:   ptrString("user@example"),
			expectedError: models.ErrInvalidEmail,
			expectSuccess: false,
		},
		{
			name:   "valid email with subdomain",
			userID: uuid.New(),
			existingUsers: map[uuid.UUID]*models.User{
				uuid.New(): {
					ID:       uuid.New(),
					Email:    "user1@example.com",
					FirstName: "User 1", LastName: "Lastname",
					Role:     models.RoleStudent,
				},
			},
			updateEmail:   ptrString("user@mail.example.co.uk"),
			expectedError: nil,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготавливаем mock репозиторий
			mockRepo := &MockUserRepositoryForEmail{
				users:  make(map[uuid.UUID]*models.User),
				emails: make(map[string]*models.User),
			}

			// Добавляем существующих пользователей
			for _, user := range tt.existingUsers {
				mockRepo.users[user.ID] = user
				mockRepo.emails[user.Email] = user
			}

			// Если тестируемый пользователь не в списке существующих, добавляем его
			if _, exists := mockRepo.users[tt.userID]; !exists {
				testUser := &models.User{
					ID:             tt.userID,
					Email:          "original@example.com",
					FirstName: "Test User", LastName: "Lastname",
					Role:           models.RoleStudent,
					PaymentEnabled: true,
				}
				mockRepo.users[tt.userID] = testUser
				mockRepo.emails[testUser.Email] = testUser
			}

			// Создаем сервис с mock репозиторием
			service := NewUserService(mockRepo, nil)

			// Подготавливаем запрос
			req := &models.UpdateUserRequest{
				Email: tt.updateEmail,
			}

			// Выполняем обновление
			result, err := service.UpdateUser(context.Background(), tt.userID, req)

			if tt.expectSuccess {
				require.NoError(t, err, "unexpected error: %v", err)
				require.NotNil(t, result, "result should not be nil")
				if tt.updateEmail != nil {
					assert.Equal(t, *tt.updateEmail, result.Email, "email should be updated")
				}
			} else {
				require.Error(t, err, "expected error but got none")
				if tt.expectedError != nil {
					assert.True(t, errors.Is(err, tt.expectedError), "error type mismatch: expected %v, got %v", tt.expectedError, err)
				}
			}
		})
	}
}

// TestUpdateUserEmailUniqueness тест для проверки уникальности email
func TestUpdateUserEmailUniqueness(t *testing.T) {
	ctx := context.Background()

	mockRepo := &MockUserRepositoryForEmail{
		users:  make(map[uuid.UUID]*models.User),
		emails: make(map[string]*models.User),
	}

	// Создаем двух пользователей
	user1ID := uuid.New()
	user2ID := uuid.New()

	user1 := &models.User{
		ID:       user1ID,
		Email:    "user1@example.com",
		FirstName: "User 1", LastName: "Lastname",
		Role:     models.RoleStudent,
	}

	user2 := &models.User{
		ID:       user2ID,
		Email:    "user2@example.com",
		FirstName: "User 2", LastName: "Lastname",
		Role:     models.RoleStudent,
	}

	mockRepo.users[user1ID] = user1
	mockRepo.emails["user1@example.com"] = user1
	mockRepo.users[user2ID] = user2
	mockRepo.emails["user2@example.com"] = user2

	service := NewUserService(mockRepo, nil)

	// Попытка присвоить email пользователя 1 пользователю 2 должна завершиться ошибкой
	newEmail := "user1@example.com"
	req := &models.UpdateUserRequest{
		Email: &newEmail,
	}

	_, err := service.UpdateUser(ctx, user2ID, req)
	assert.True(t, errors.Is(err, repository.ErrUserExists), "should return ErrUserExists, got %v", err)
}

// TestUpdateUserEmailSelfUpdate тест для обновления на тот же email
func TestUpdateUserEmailSelfUpdate(t *testing.T) {
	ctx := context.Background()

	mockRepo := &MockUserRepositoryForEmail{
		users:  make(map[uuid.UUID]*models.User),
		emails: make(map[string]*models.User),
	}

	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Email:    "user@example.com",
		FirstName: "User", LastName: "Lastname",
		Role:     models.RoleStudent,
	}

	mockRepo.users[userID] = user
	mockRepo.emails["user@example.com"] = user

	service := NewUserService(mockRepo, nil)

	// Попытка присвоить тот же email должна успешно завершиться
	sameEmail := "user@example.com"
	req := &models.UpdateUserRequest{
		Email: &sameEmail,
	}

	result, err := service.UpdateUser(ctx, userID, req)
	assert.NoError(t, err, "should not return error")
	assert.NotNil(t, result, "result should not be nil")
	assert.Equal(t, sameEmail, result.Email, "email should remain the same")
}

// TestUpdateUserEmailWithoutChange тест для обновления без email
func TestUpdateUserEmailWithoutChange(t *testing.T) {
	ctx := context.Background()

	mockRepo := &MockUserRepositoryForEmail{
		users:  make(map[uuid.UUID]*models.User),
		emails: make(map[string]*models.User),
	}

	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Email:    "user@example.com",
		FirstName: "User", LastName: "Lastname",
		Role:     models.RoleStudent,
	}

	mockRepo.users[userID] = user
	mockRepo.emails["user@example.com"] = user

	service := NewUserService(mockRepo, nil)

	// Обновляем только FullName без email
	newName := "Updated Name"
	req := &models.UpdateUserRequest{
		FullName: &newName,
	}

	result, err := service.UpdateUser(ctx, userID, req)
	assert.NoError(t, err, "should not return error")
	assert.NotNil(t, result, "result should not be nil")
	assert.Equal(t, "user@example.com", result.Email, "email should remain unchanged")
	assert.Equal(t, newName, result.FullName, "full name should be updated")
}

// MockUserRepositoryForEmail специальный мок для тестирования email валидации
type MockUserRepositoryForEmail struct {
	users  map[uuid.UUID]*models.User
	emails map[string]*models.User
}

func (m *MockUserRepositoryForEmail) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepositoryForEmail) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, exists := m.emails[email]
	if !exists {
		return nil, repository.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepositoryForEmail) Create(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	m.emails[user.Email] = user
	return nil
}

func (m *MockUserRepositoryForEmail) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	user, exists := m.users[id]
	if !exists {
		return repository.ErrUserNotFound
	}

	// Обновляем email в оба индекса
	if newEmail, ok := updates["email"]; ok {
		oldEmail := user.Email
		delete(m.emails, oldEmail)
		user.Email = newEmail.(string)
		m.emails[user.Email] = user
	}

	if newFullName, ok := updates["full_name"]; ok {
		user.FullName = newFullName.(string)
	}

	if newRole, ok := updates["role"]; ok {
		user.Role = newRole.(models.UserRole)
	}

	if newPaymentEnabled, ok := updates["payment_enabled"]; ok {
		user.PaymentEnabled = newPaymentEnabled.(bool)
	}

	return nil
}

func (m *MockUserRepositoryForEmail) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	user, exists := m.users[id]
	if !exists {
		return repository.ErrUserNotFound
	}
	user.PasswordHash = passwordHash
	return nil
}

func (m *MockUserRepositoryForEmail) Delete(ctx context.Context, id uuid.UUID) error {
	user, exists := m.users[id]
	if exists {
		delete(m.emails, user.Email)
		delete(m.users, id)
	}
	return nil
}

func (m *MockUserRepositoryForEmail) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return m.Delete(ctx, id)
}

func (m *MockUserRepositoryForEmail) List(ctx context.Context, roleFilter *models.UserRole) ([]*models.User, error) {
	var result []*models.User
	for _, user := range m.users {
		if roleFilter == nil || user.Role == *roleFilter {
			result = append(result, user)
		}
	}
	return result, nil
}

func (m *MockUserRepositoryForEmail) ListWithPagination(ctx context.Context, role *models.UserRole, offset, limit int) ([]*models.User, int, error) {
	var result []*models.User
	for _, user := range m.users {
		if role == nil || user.Role == *role {
			result = append(result, user)
		}
	}
	// Simple pagination
	total := len(result)
	if offset >= total {
		return []*models.User{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return result[offset:end], total, nil
}

func (m *MockUserRepositoryForEmail) Exists(ctx context.Context, email string) (bool, error) {
	_, exists := m.emails[email]
	return exists, nil
}

func (m *MockUserRepositoryForEmail) UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error {
	user, exists := m.users[userID]
	if !exists {
		return repository.ErrUserNotFound
	}
	user.TelegramUsername.String = username
	user.TelegramUsername.Valid = true
	return nil
}

// ptrString helper для создания указателя на строку
func ptrString(s string) *string {
	return &s
}

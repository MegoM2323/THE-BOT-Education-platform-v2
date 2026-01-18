package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateUser_EmailValidation тест валидации email при обновлении пользователя
func TestUpdateUser_EmailValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Создаем админа для обновления пользователей
	admin := createTestUser(t, db, "admin@test.com", "Admin User", string(models.RoleAdmin))

	// Создаем студента для тестирования обновления
	student := createTestUser(t, db, "student@test.com", "Student User", string(models.RoleStudent))

	// Создаем репозитории и сервис
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, nil)

	tests := []struct {
		name           string
		userID         string
		updateRequest  models.UpdateUserRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "valid email update",
			userID: student.ID.String(),
			updateRequest: models.UpdateUserRequest{
				Email: stringPtr("newemail@example.com"),
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:   "invalid email format - missing @",
			userID: student.ID.String(),
			updateRequest: models.UpdateUserRequest{
				Email: stringPtr("invalidemail.com"),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid email",
		},
		{
			name:   "invalid email format - missing dot",
			userID: student.ID.String(),
			updateRequest: models.UpdateUserRequest{
				Email: stringPtr("user@example"),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid email",
		},
		{
			name:   "update with same email",
			userID: student.ID.String(),
			updateRequest: models.UpdateUserRequest{
				Email: stringPtr(student.Email),
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем контекст с админом
			ctx := context.WithValue(context.Background(), middleware.UserContextKey, admin)

			// Подготавливаем тело запроса
			body, err := json.Marshal(tt.updateRequest)
			require.NoError(t, err)

			// Note: body is used in marshaling, but we call service directly
			_ = body

			// Вызываем обработчик (симулируем chi routing подставляя ID в URL)
			// Так как у нас нет настроенного chi роутера, используем service напрямую
			updateReq := tt.updateRequest
			updateReq.Sanitize()

			// Парсим ID
			targetUserID, err := uuid.Parse(tt.userID)
			require.NoError(t, err)

			// Вызываем сервис
			result, err := userService.UpdateUser(ctx, targetUserID, &updateReq)

			// Проверяем результаты
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err, "expected no error for successful update")
				assert.NotNil(t, result, "result should not be nil")
				if tt.updateRequest.Email != nil {
					assert.Equal(t, *tt.updateRequest.Email, result.Email, "email should be updated")
				}
			} else {
				assert.Error(t, err, "expected error for invalid input")
			}
		})
	}
}

// TestUpdateUser_DuplicateEmail тест проверки уникальности email при обновлении
func TestUpdateUser_DuplicateEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	student1 := createTestUser(t, db, "student1@test.com", "Student 1", string(models.RoleStudent))
	student2 := createTestUser(t, db, "student2@test.com", "Student 2", string(models.RoleStudent))

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, nil)

	newEmail := student1.Email
	updateReq := &models.UpdateUserRequest{
		Email: &newEmail,
	}

	_, err := userService.UpdateUser(ctx, student2.ID, updateReq)
	assert.Error(t, err, "expected error when assigning duplicate email")
	assert.True(t, errors.Is(err, repository.ErrUserExists), "should be ErrUserExists error")
}

// TestUpdateUser_PreservesOtherFields тест сохранения других полей при обновлении email
func TestUpdateUser_PreservesOtherFields(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	// Создаем студента
	student := createTestUser(t, db, "student@test.com", "Student User", string(models.RoleStudent))

	// Запоминаем оригинальные значения
	originalFullName := student.FullName
	originalRole := student.Role

	// Создаем репозитории и сервис
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, nil)

	// Обновляем только email
	newEmail := "newemail@example.com"
	updateReq := &models.UpdateUserRequest{
		Email: &newEmail,
	}

	result, err := userService.UpdateUser(ctx, student.ID, updateReq)
	require.NoError(t, err, "update should succeed")
	require.NotNil(t, result, "result should not be nil")

	// Проверяем что email обновился
	assert.Equal(t, newEmail, result.Email, "email should be updated")

	// Проверяем что другие поля остались без изменений
	assert.Equal(t, originalFullName, result.FullName, "full name should not change")
	assert.Equal(t, originalRole, result.Role, "role should not change")
}

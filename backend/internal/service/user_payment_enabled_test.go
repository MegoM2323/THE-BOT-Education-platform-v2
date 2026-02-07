package service

import (
	"context"
	"testing"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService_CreateUser_PaymentEnabledDefault(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	userService := NewUserService(userRepo, creditRepo)

	ctx := context.Background()

	// Создаём пользователя
	req := &models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		FirstName: "Test User", LastName: "Lastname",
		Role:     models.RoleStudent,
	}

	user, err := userService.CreateUser(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, user)

	// Проверяем что payment_enabled = true по умолчанию
	assert.True(t, user.PaymentEnabled, "PaymentEnabled should be true by default")

	// Проверяем что значение сохранилось в БД
	fetchedUser, err := userRepo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, fetchedUser.PaymentEnabled, "PaymentEnabled should be true in DB")
}

func TestUserService_UpdateUser_PaymentEnabled(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	userService := NewUserService(userRepo, creditRepo)

	ctx := context.Background()

	// Создаём пользователя
	user := createTestUser(t, db, "test@example.com", "Test User", "student")

	// Проверяем начальное значение
	assert.True(t, user.PaymentEnabled)

	// Обновляем payment_enabled на false
	falseValue := false
	updateReq := &models.UpdateUserRequest{
		PaymentEnabled: &falseValue,
	}

	updatedUser, err := userService.UpdateUser(ctx, user.ID, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updatedUser)

	// Проверяем что payment_enabled обновился
	assert.False(t, updatedUser.PaymentEnabled, "PaymentEnabled should be false after update")

	// Обновляем payment_enabled обратно на true
	trueValue := true
	updateReq2 := &models.UpdateUserRequest{
		PaymentEnabled: &trueValue,
	}

	updatedUser2, err := userService.UpdateUser(ctx, user.ID, updateReq2)
	require.NoError(t, err)
	require.NotNil(t, updatedUser2)

	// Проверяем что payment_enabled снова true
	assert.True(t, updatedUser2.PaymentEnabled, "PaymentEnabled should be true after second update")
}

func TestUserService_UpdateUser_PaymentEnabledNil(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	userService := NewUserService(userRepo, creditRepo)

	ctx := context.Background()

	// Создаём пользователя
	user := createTestUser(t, db, "test@example.com", "Test User", "student")

	// Обновляем другие поля без payment_enabled
	newName := "Updated Name"
	updateReq := &models.UpdateUserRequest{
		FullName: &newName,
		// PaymentEnabled не передаём (nil)
	}

	updatedUser, err := userService.UpdateUser(ctx, user.ID, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updatedUser)

	// Проверяем что payment_enabled не изменился
	assert.True(t, updatedUser.PaymentEnabled, "PaymentEnabled should remain true")
	assert.Equal(t, newName, updatedUser.FullName, "FullName should be updated")
}

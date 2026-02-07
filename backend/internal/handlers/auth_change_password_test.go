package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/auth"
)

// TestChangePasswordRequiresAuthentication проверяет что эндпоинт требует аутентификации
func TestChangePasswordRequiresAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	_ = ctx // Suppress unused warning

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	sessionMgr := auth.NewSessionManager("test-secret-key-32-characters!!")

	authService := service.NewAuthService(userRepo, sessionRepo, sessionMgr, 24*3600)
	userService := service.NewUserService(userRepo, creditRepo)

	authHandler := NewAuthHandler(authService, nil, 24*3600, false)
	authHandler.SetUserService(userService)

	changePasswordReq := map[string]string{
		"old_password": "OldPassword123!",
		"new_password": "NewPassword123!",
	}
	body, _ := json.Marshal(changePasswordReq)

	// Запрос БЕЗ аутентификации
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authHandler.ChangePassword(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Not authenticated")
}

// TestChangePasswordInvalidOldPassword проверяет ошибку при неверном старом пароле
func TestChangePasswordInvalidOldPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	_ = ctx // Suppress unused warning

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	sessionMgr := auth.NewSessionManager("test-secret-key-32-characters!!")

	authService := service.NewAuthService(userRepo, sessionRepo, sessionMgr, 24*3600)
	userService := service.NewUserService(userRepo, creditRepo)

	authHandler := NewAuthHandler(authService, nil, 24*3600, false)
	authHandler.SetUserService(userService)

	// Создаем тестового пользователя
	testPassword := "TestPassword123!"
	uniqueEmail := fmt.Sprintf("%s-change-pass-test@example.com", uuid.New().String()[:8])
	user, err := userService.CreateUser(ctx, &models.CreateUserRequest{
		Email:    uniqueEmail,
		Password: testPassword,
		FirstName: "Change Password Test", LastName: "Lastname",
		Role:     "student",
	})
	require.NoError(t, err)
	require.NotNil(t, user)

	// Создаем контекст с аутентифицированным пользователем
	ctxWithUser := context.WithValue(ctx, middleware.UserContextKey, user)

	changePasswordReq := map[string]string{
		"old_password": "WrongPassword123!",
		"new_password": "NewPassword123!",
	}
	body, _ := json.Marshal(changePasswordReq)

	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithUser)
	w := httptest.NewRecorder()

	authHandler.ChangePassword(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Current password is incorrect")
}

// TestChangePasswordWeakPassword проверяет ошибку при слабом пароле (< 8 символов)
func TestChangePasswordWeakPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	_ = ctx // Suppress unused warning

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	sessionMgr := auth.NewSessionManager("test-secret-key-32-characters!!")

	authService := service.NewAuthService(userRepo, sessionRepo, sessionMgr, 24*3600)
	userService := service.NewUserService(userRepo, creditRepo)

	authHandler := NewAuthHandler(authService, nil, 24*3600, false)
	authHandler.SetUserService(userService)

	// Создаем тестового пользователя
	testPassword := "TestPassword123!"
	uniqueEmail := fmt.Sprintf("%s-weak-pass-test@example.com", uuid.New().String()[:8])
	user, err := userService.CreateUser(ctx, &models.CreateUserRequest{
		Email:    uniqueEmail,
		Password: testPassword,
		FirstName: "Weak Password Test", LastName: "Lastname",
		Role:     "student",
	})
	require.NoError(t, err)
	require.NotNil(t, user)

	ctxWithUser := context.WithValue(ctx, middleware.UserContextKey, user)

	changePasswordReq := map[string]string{
		"old_password": testPassword,
		"new_password": "Short1!", // < 8 characters
	}
	body, _ := json.Marshal(changePasswordReq)

	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithUser)
	w := httptest.NewRecorder()

	authHandler.ChangePassword(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "at least 8 characters")
}

// TestChangePasswordSameAsOld проверяет ошибку когда новый пароль совпадает со старым
func TestChangePasswordSameAsOld(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	_ = ctx // Suppress unused warning

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	sessionMgr := auth.NewSessionManager("test-secret-key-32-characters!!")

	authService := service.NewAuthService(userRepo, sessionRepo, sessionMgr, 24*3600)
	userService := service.NewUserService(userRepo, creditRepo)

	authHandler := NewAuthHandler(authService, nil, 24*3600, false)
	authHandler.SetUserService(userService)

	// Создаем тестового пользователя
	testPassword := "TestPassword123!"
	uniqueEmail := fmt.Sprintf("%s-same-pass-test@example.com", uuid.New().String()[:8])
	user, err := userService.CreateUser(ctx, &models.CreateUserRequest{
		Email:    uniqueEmail,
		Password: testPassword,
		FirstName: "Same Password Test", LastName: "Lastname",
		Role:     "student",
	})
	require.NoError(t, err)
	require.NotNil(t, user)

	ctxWithUser := context.WithValue(ctx, middleware.UserContextKey, user)

	changePasswordReq := map[string]string{
		"old_password": testPassword,
		"new_password": testPassword, // Same as old
	}
	body, _ := json.Marshal(changePasswordReq)

	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithUser)
	w := httptest.NewRecorder()

	authHandler.ChangePassword(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "different from the current password")
}

// TestChangePasswordSuccess проверяет успешное изменение пароля
func TestChangePasswordSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	_ = ctx // Suppress unused warning

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	sessionMgr := auth.NewSessionManager("test-secret-key-32-characters!!")

	authService := service.NewAuthService(userRepo, sessionRepo, sessionMgr, 24*3600)
	userService := service.NewUserService(userRepo, creditRepo)

	authHandler := NewAuthHandler(authService, nil, 24*3600, false)
	authHandler.SetUserService(userService)

	// Создаем тестового пользователя
	testPassword := "TestPassword123!"
	newPassword := "NewPassword123!"
	uniqueEmail := fmt.Sprintf("%s-success-pass-test@example.com", uuid.New().String()[:8])
	user, err := userService.CreateUser(ctx, &models.CreateUserRequest{
		Email:    uniqueEmail,
		Password: testPassword,
		FirstName: "Success Password Test", LastName: "Lastname",
		Role:     "student",
	})
	require.NoError(t, err)
	require.NotNil(t, user)

	ctxWithUser := context.WithValue(ctx, middleware.UserContextKey, user)

	changePasswordReq := map[string]string{
		"old_password": testPassword,
		"new_password": newPassword,
	}
	body, _ := json.Marshal(changePasswordReq)

	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithUser)
	w := httptest.NewRecorder()

	authHandler.ChangePassword(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "successfully")

	// Проверяем что старый пароль больше не работает
	loginReq := map[string]string{
		"email":    uniqueEmail,
		"password": testPassword, // old password
	}
	loginBody, _ := json.Marshal(loginReq)

	loginRequest := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginWriter := httptest.NewRecorder()

	authHandler.Login(loginWriter, loginRequest)
	assert.Equal(t, http.StatusUnauthorized, loginWriter.Code)

	// Проверяем что новый пароль работает
	loginReq = map[string]string{
		"email":    uniqueEmail,
		"password": newPassword, // new password
	}
	loginBody, _ = json.Marshal(loginReq)

	loginRequest = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginWriter = httptest.NewRecorder()

	authHandler.Login(loginWriter, loginRequest)
	assert.Equal(t, http.StatusOK, loginWriter.Code)
}

// TestChangePasswordMissingOldPassword проверяет ошибку при отсутствии старого пароля
func TestChangePasswordMissingOldPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	_ = ctx // Suppress unused warning

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	sessionMgr := auth.NewSessionManager("test-secret-key-32-characters!!")

	authService := service.NewAuthService(userRepo, sessionRepo, sessionMgr, 24*3600)
	userService := service.NewUserService(userRepo, creditRepo)

	authHandler := NewAuthHandler(authService, nil, 24*3600, false)
	authHandler.SetUserService(userService)

	// Создаем тестового пользователя
	testPassword := "TestPassword123!"
	uniqueEmail := fmt.Sprintf("%s-missing-old@example.com", uuid.New().String()[:8])
	user, err := userService.CreateUser(ctx, &models.CreateUserRequest{
		Email:    uniqueEmail,
		Password: testPassword,
		FirstName: "Missing Old Password Test", LastName: "Lastname",
		Role:     "student",
	})
	require.NoError(t, err)

	ctxWithUser := context.WithValue(ctx, middleware.UserContextKey, user)

	changePasswordReq := map[string]string{
		"old_password": "",
		"new_password": "NewPassword123!",
	}
	body, _ := json.Marshal(changePasswordReq)

	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithUser)
	w := httptest.NewRecorder()

	authHandler.ChangePassword(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Old password is required")
}

// TestChangePasswordMissingNewPassword проверяет ошибку при отсутствии нового пароля
func TestChangePasswordMissingNewPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	_ = ctx // Suppress unused warning

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	sessionMgr := auth.NewSessionManager("test-secret-key-32-characters!!")

	authService := service.NewAuthService(userRepo, sessionRepo, sessionMgr, 24*3600)
	userService := service.NewUserService(userRepo, creditRepo)

	authHandler := NewAuthHandler(authService, nil, 24*3600, false)
	authHandler.SetUserService(userService)

	// Создаем тестового пользователя
	testPassword := "TestPassword123!"
	uniqueEmail := fmt.Sprintf("%s-missing-new@example.com", uuid.New().String()[:8])
	user, err := userService.CreateUser(ctx, &models.CreateUserRequest{
		Email:    uniqueEmail,
		Password: testPassword,
		FirstName: "Missing New Password Test", LastName: "Lastname",
		Role:     "student",
	})
	require.NoError(t, err)

	ctxWithUser := context.WithValue(ctx, middleware.UserContextKey, user)

	changePasswordReq := map[string]string{
		"old_password": testPassword,
		"new_password": "",
	}
	body, _ := json.Marshal(changePasswordReq)

	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithUser)
	w := httptest.NewRecorder()

	authHandler.ChangePassword(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "New password is required")
}

// TestChangePasswordInvalidJSON проверяет ошибку при невалидном JSON
func TestChangePasswordInvalidJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	_ = ctx // Suppress unused warning

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	sessionMgr := auth.NewSessionManager("test-secret-key-32-characters!!")

	authService := service.NewAuthService(userRepo, sessionRepo, sessionMgr, 24*3600)
	userService := service.NewUserService(userRepo, creditRepo)

	authHandler := NewAuthHandler(authService, nil, 24*3600, false)
	authHandler.SetUserService(userService)

	// Создаем тестового пользователя
	testPassword := "TestPassword123!"
	uniqueEmail := fmt.Sprintf("%s-invalid-json@example.com", uuid.New().String()[:8])
	user, err := userService.CreateUser(ctx, &models.CreateUserRequest{
		Email:    uniqueEmail,
		Password: testPassword,
		FirstName: "Invalid JSON Test", LastName: "Lastname",
		Role:     "student",
	})
	require.NoError(t, err)

	ctxWithUser := context.WithValue(ctx, middleware.UserContextKey, user)

	// Невалидный JSON
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxWithUser)
	w := httptest.NewRecorder()

	authHandler.ChangePassword(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
)

// MockPaymentSettingsService это mock для PaymentSettingsService
type MockPaymentSettingsService struct {
	mock.Mock
}

func (m *MockPaymentSettingsService) GetPaymentStatus(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPaymentSettingsService) UpdatePaymentStatus(ctx context.Context, adminID, userID uuid.UUID, enabled bool) (*models.User, error) {
	args := m.Called(ctx, adminID, userID, enabled)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockPaymentSettingsService) ListStudentsPaymentStatus(ctx context.Context, adminID uuid.UUID, filterByEnabled *bool) ([]*models.StudentPaymentStatus, error) {
	args := m.Called(ctx, adminID, filterByEnabled)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.StudentPaymentStatus), args.Error(1)
}

// TestListStudentsPaymentStatus_Success проверяет успешное получение списка студентов
func TestListStudentsPaymentStatus_Success(t *testing.T) {
	mockService := new(MockPaymentSettingsService)
	handler := NewPaymentSettingsHandler(mockService)

	adminID := uuid.New()
	students := []*models.StudentPaymentStatus{
		{
			ID:             uuid.New(),
			FullName:       "Иван Петров",
			Email:          "ivan@example.com",
			PaymentEnabled: true,
		},
		{
			ID:             uuid.New(),
			FullName:       "Мария Сидорова",
			Email:          "maria@example.com",
			PaymentEnabled: false,
		},
	}

	mockService.On("ListStudentsPaymentStatus", mock.Anything, adminID, (*bool)(nil)).Return(students, nil)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/payment-settings", nil)

	// Добавляем admin пользователя в контекст
	adminUser := &models.User{
		ID:   adminID,
		Role: models.RoleAdmin,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminUser)
	req = req.WithContext(ctx)

	// Выполняем запрос
	rr := httptest.NewRecorder()
	handler.ListStudentsPaymentStatus(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(2), data["total"])
	assert.Equal(t, float64(2), data["count"])

	mockService.AssertExpectations(t)
}

// TestListStudentsPaymentStatus_FilterEnabled проверяет фильтрацию по payment_enabled
func TestListStudentsPaymentStatus_FilterEnabled(t *testing.T) {
	mockService := new(MockPaymentSettingsService)
	handler := NewPaymentSettingsHandler(mockService)

	adminID := uuid.New()
	enabled := true
	students := []*models.StudentPaymentStatus{
		{
			ID:             uuid.New(),
			FullName:       "Иван Петров",
			Email:          "ivan@example.com",
			PaymentEnabled: true,
		},
	}

	mockService.On("ListStudentsPaymentStatus", mock.Anything, adminID, &enabled).Return(students, nil)

	// Создаем запрос с фильтром
	req := httptest.NewRequest(http.MethodGet, "/admin/payment-settings?payment_enabled=true", nil)

	adminUser := &models.User{
		ID:   adminID,
		Role: models.RoleAdmin,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminUser)
	req = req.WithContext(ctx)

	// Выполняем запрос
	rr := httptest.NewRecorder()
	handler.ListStudentsPaymentStatus(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	mockService.AssertExpectations(t)
}

// TestListStudentsPaymentStatus_NotAdmin проверяет отказ для не-админов
func TestListStudentsPaymentStatus_NotAdmin(t *testing.T) {
	mockService := new(MockPaymentSettingsService)
	handler := NewPaymentSettingsHandler(mockService)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/admin/payment-settings", nil)

	// Добавляем студента в контекст
	studentUser := &models.User{
		ID:   uuid.New(),
		Role: models.RoleStudent,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, studentUser)
	req = req.WithContext(ctx)

	// Выполняем запрос
	rr := httptest.NewRecorder()
	handler.ListStudentsPaymentStatus(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusForbidden, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	// Не вызывается сервис
	mockService.AssertNotCalled(t, "ListStudentsPaymentStatus")
}

// TestUpdatePaymentStatus_Success проверяет успешное обновление статуса
func TestUpdatePaymentStatus_Success(t *testing.T) {
	mockService := new(MockPaymentSettingsService)
	handler := NewPaymentSettingsHandler(mockService)

	adminID := uuid.New()
	studentID := uuid.New()
	enabled := true

	updatedUser := &models.User{
		ID:             studentID,
		FullName:       "Иван Петров",
		Email:          "ivan@example.com",
		Role:           models.RoleStudent,
		PaymentEnabled: enabled,
	}

	mockService.On("UpdatePaymentStatus", mock.Anything, adminID, studentID, enabled).Return(updatedUser, nil)

	// Создаем JSON body
	reqBody := UpdatePaymentStatusRequest{
		PaymentEnabled: &enabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPut, "/admin/users/"+studentID.String()+"/payment-settings", bytes.NewReader(bodyBytes))

	// Добавляем admin в контекст
	adminUser := &models.User{
		ID:   adminID,
		Role: models.RoleAdmin,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminUser)

	// Добавляем URL параметры через chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", studentID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	// Выполняем запрос
	rr := httptest.NewRecorder()
	handler.UpdatePaymentStatus(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	user := data["user"].(map[string]interface{})
	assert.Equal(t, studentID.String(), user["id"])
	assert.Equal(t, enabled, user["payment_enabled"])

	mockService.AssertExpectations(t)
}

// TestUpdatePaymentStatus_InvalidUserID проверяет обработку неверного user ID
func TestUpdatePaymentStatus_InvalidUserID(t *testing.T) {
	mockService := new(MockPaymentSettingsService)
	handler := NewPaymentSettingsHandler(mockService)

	enabled := true
	reqBody := UpdatePaymentStatusRequest{
		PaymentEnabled: &enabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Создаем запрос с невалидным UUID
	req := httptest.NewRequest(http.MethodPut, "/admin/users/invalid-uuid/payment-settings", bytes.NewReader(bodyBytes))

	adminUser := &models.User{
		ID:   uuid.New(),
		Role: models.RoleAdmin,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminUser)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	// Выполняем запрос
	rr := httptest.NewRecorder()
	handler.UpdatePaymentStatus(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	mockService.AssertNotCalled(t, "UpdatePaymentStatus")
}

// TestUpdatePaymentStatus_MissingPaymentEnabled проверяет обработку отсутствующего поля
func TestUpdatePaymentStatus_MissingPaymentEnabled(t *testing.T) {
	mockService := new(MockPaymentSettingsService)
	handler := NewPaymentSettingsHandler(mockService)

	studentID := uuid.New()

	// Создаем JSON body без payment_enabled
	reqBody := map[string]interface{}{}
	bodyBytes, _ := json.Marshal(reqBody)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodPut, "/admin/users/"+studentID.String()+"/payment-settings", bytes.NewReader(bodyBytes))

	adminUser := &models.User{
		ID:   uuid.New(),
		Role: models.RoleAdmin,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminUser)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", studentID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	// Выполняем запрос
	rr := httptest.NewRecorder()
	handler.UpdatePaymentStatus(rr, req)

	// Проверяем результат
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	mockService.AssertNotCalled(t, "UpdatePaymentStatus")
}

// TestUpdatePaymentStatus_UserNotFound проверяет обработку несуществующего пользователя
func TestUpdatePaymentStatus_UserNotFound(t *testing.T) {
	mockService := new(MockPaymentSettingsService)
	handler := NewPaymentSettingsHandler(mockService)

	adminID := uuid.New()
	studentID := uuid.New()
	enabled := true

	mockService.On("UpdatePaymentStatus", mock.Anything, adminID, studentID, enabled).Return(nil, repository.ErrUserNotFound)

	reqBody := UpdatePaymentStatusRequest{
		PaymentEnabled: &enabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/admin/users/"+studentID.String()+"/payment-settings", bytes.NewReader(bodyBytes))

	adminUser := &models.User{
		ID:   adminID,
		Role: models.RoleAdmin,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminUser)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", studentID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdatePaymentStatus(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	mockService.AssertExpectations(t)
}

// TestUpdatePaymentStatus_InvalidUserRole проверяет попытку изменить не-студента
func TestUpdatePaymentStatus_InvalidUserRole(t *testing.T) {
	mockService := new(MockPaymentSettingsService)
	handler := NewPaymentSettingsHandler(mockService)

	adminID := uuid.New()
	teacherID := uuid.New()
	enabled := false

	mockService.On("UpdatePaymentStatus", mock.Anything, adminID, teacherID, enabled).Return(nil, repository.ErrInvalidUserRole)

	reqBody := UpdatePaymentStatusRequest{
		PaymentEnabled: &enabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/admin/users/"+teacherID.String()+"/payment-settings", bytes.NewReader(bodyBytes))

	adminUser := &models.User{
		ID:   adminID,
		Role: models.RoleAdmin,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, adminUser)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", teacherID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.UpdatePaymentStatus(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))

	mockService.AssertExpectations(t)
}

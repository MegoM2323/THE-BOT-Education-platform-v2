package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
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

// MockLessonBroadcastService мок для LessonBroadcastService
type MockLessonBroadcastService struct {
	mock.Mock
}

func (m *MockLessonBroadcastService) CreateLessonBroadcast(
	ctx context.Context,
	userID uuid.UUID,
	lessonID uuid.UUID,
	message string,
	files []*multipart.FileHeader,
) (*models.LessonBroadcast, error) {
	args := m.Called(ctx, userID, lessonID, message, files)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LessonBroadcast), args.Error(1)
}

func (m *MockLessonBroadcastService) ListLessonBroadcasts(
	ctx context.Context,
	lessonID uuid.UUID,
) ([]*models.LessonBroadcast, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.LessonBroadcast), args.Error(1)
}

func (m *MockLessonBroadcastService) GetLessonBroadcast(
	ctx context.Context,
	broadcastID uuid.UUID,
) (*models.LessonBroadcast, error) {
	args := m.Called(ctx, broadcastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LessonBroadcast), args.Error(1)
}

func (m *MockLessonBroadcastService) GetBroadcastFileWithAccess(
	ctx context.Context,
	userID uuid.UUID,
	fileID uuid.UUID,
) (*models.BroadcastFile, error) {
	args := m.Called(ctx, userID, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BroadcastFile), args.Error(1)
}

// TestCreateBroadcast_Success проверяет успешное создание рассылки
func TestCreateBroadcast_Success(t *testing.T) {
	mockService := new(MockLessonBroadcastService)
	handler := NewLessonBroadcastHandler(mockService, "/tmp/uploads")

	userID := uuid.New()
	lessonID := uuid.New()
	broadcastID := uuid.New()
	message := "Тестовое сообщение"

	// Создаем multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("message", message)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/lessons/"+lessonID.String()+"/broadcasts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Добавляем пользователя в контекст (используем middleware.UserContextKey)
	user := &models.User{
		ID:   userID,
		Role: models.RoleAdmin,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)

	// Создаем роутер для chi.URLParam
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", lessonID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	expectedBroadcast := &models.LessonBroadcast{
		ID:       broadcastID,
		LessonID: lessonID,
		SenderID: userID,
		Message:  message,
		Status:   models.LessonBroadcastStatusPending,
	}

	mockService.On("CreateLessonBroadcast", mock.Anything, userID, lessonID, message, mock.AnythingOfType("[]*multipart.FileHeader")).
		Return(expectedBroadcast, nil)

	w := httptest.NewRecorder()
	handler.CreateBroadcast(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response["success"].(bool))

	mockService.AssertExpectations(t)
}

// TestCreateBroadcast_NoAuth проверяет ошибку при отсутствии аутентификации
func TestCreateBroadcast_NoAuth(t *testing.T) {
	mockService := new(MockLessonBroadcastService)
	handler := NewLessonBroadcastHandler(mockService, "/tmp/uploads")

	lessonID := uuid.New()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("message", "Test message")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/lessons/"+lessonID.String()+"/broadcasts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Добавляем chi route context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", lessonID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateBroadcast(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreateBroadcast_InvalidMessage проверяет валидацию сообщения
func TestCreateBroadcast_InvalidMessage(t *testing.T) {
	mockService := new(MockLessonBroadcastService)
	handler := NewLessonBroadcastHandler(mockService, "/tmp/uploads")

	userID := uuid.New()
	lessonID := uuid.New()

	// Пустое сообщение
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("message", "")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/lessons/"+lessonID.String()+"/broadcasts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	user := &models.User{
		ID:   userID,
		Role: models.RoleAdmin,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", lessonID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.CreateBroadcast(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestListBroadcasts_Success проверяет успешное получение списка рассылок
func TestListBroadcasts_Success(t *testing.T) {
	mockService := new(MockLessonBroadcastService)
	handler := NewLessonBroadcastHandler(mockService, "/tmp/uploads")

	userID := uuid.New()
	lessonID := uuid.New()
	broadcast1 := &models.LessonBroadcast{
		ID:       uuid.New(),
		LessonID: lessonID,
		SenderID: userID,
		Message:  "Message 1",
		Status:   models.LessonBroadcastStatusCompleted,
	}
	broadcast2 := &models.LessonBroadcast{
		ID:       uuid.New(),
		LessonID: lessonID,
		SenderID: userID,
		Message:  "Message 2",
		Status:   models.LessonBroadcastStatusPending,
	}

	mockService.On("ListLessonBroadcasts", mock.Anything, lessonID).
		Return([]*models.LessonBroadcast{broadcast1, broadcast2}, nil)

	req := httptest.NewRequest("GET", "/api/v1/lessons/"+lessonID.String()+"/broadcasts", nil)

	user := &models.User{
		ID:   userID,
		Role: models.RoleMethodologist,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", lessonID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ListBroadcasts(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response["success"].(bool))
	broadcasts := response["data"].(map[string]interface{})["broadcasts"].([]interface{})
	assert.Equal(t, 2, len(broadcasts))

	mockService.AssertExpectations(t)
}

// TestGetBroadcast_Success проверяет успешное получение деталей рассылки
func TestGetBroadcast_Success(t *testing.T) {
	mockService := new(MockLessonBroadcastService)
	handler := NewLessonBroadcastHandler(mockService, "/tmp/uploads")

	userID := uuid.New()
	lessonID := uuid.New()
	broadcastID := uuid.New()

	broadcast := &models.LessonBroadcast{
		ID:       broadcastID,
		LessonID: lessonID,
		SenderID: userID,
		Message:  "Test broadcast",
		Status:   models.LessonBroadcastStatusCompleted,
	}

	mockService.On("GetLessonBroadcast", mock.Anything, broadcastID).
		Return(broadcast, nil)

	req := httptest.NewRequest("GET", "/api/v1/lessons/"+lessonID.String()+"/broadcasts/"+broadcastID.String(), nil)

	user := &models.User{
		ID:   userID,
		Role: models.RoleMethodologist,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", lessonID.String())
	rctx.URLParams.Add("broadcast_id", broadcastID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.GetBroadcast(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response["success"].(bool))

	mockService.AssertExpectations(t)
}

// TestGetBroadcast_NotFound проверяет ошибку при несуществующей рассылке
func TestGetBroadcast_NotFound(t *testing.T) {
	mockService := new(MockLessonBroadcastService)
	handler := NewLessonBroadcastHandler(mockService, "/tmp/uploads")

	userID := uuid.New()
	lessonID := uuid.New()
	broadcastID := uuid.New()

	mockService.On("GetLessonBroadcast", mock.Anything, broadcastID).
		Return(nil, repository.ErrLessonBroadcastNotFound)

	req := httptest.NewRequest("GET", "/api/v1/lessons/"+lessonID.String()+"/broadcasts/"+broadcastID.String(), nil)

	user := &models.User{
		ID:   userID,
		Role: models.RoleMethodologist,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", lessonID.String())
	rctx.URLParams.Add("broadcast_id", broadcastID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.GetBroadcast(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

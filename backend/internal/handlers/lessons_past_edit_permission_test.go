package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/internal/validator"
)

// Тест: Админ может редактировать прошлые занятия
func TestUpdateLesson_Admin_CanEditPastLesson(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	sqlxDB := setupTestDB(t)
	defer cleanupTestDB(t, sqlxDB)
	pool := setupTestDBForHandler(t, sqlxDB)
	// Note: don't close pool - it's a shared test pool managed globally

	// Создаём репозитории
	lessonRepo := repository.NewLessonRepository(sqlxDB)
	userRepo := repository.NewUserRepository(sqlxDB)
	bookingRepo := repository.NewBookingRepository(sqlxDB)
	creditRepo := repository.NewCreditRepository(sqlxDB)

	// Создаём сервисы
	lessonService := service.NewLessonService(lessonRepo, userRepo)
	bookingService := service.NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		repository.NewCancelledBookingRepository(sqlxDB),
		validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo),
	)
	bulkEditService := service.NewBulkEditService(
		pool,
		lessonRepo,
		repository.NewLessonModificationRepository(sqlxDB),
		userRepo,
		creditRepo,
	)

	// Создаём handler
	handler := NewLessonHandler(lessonService, bookingService, bulkEditService)

	// Создаём админа и преподавателя
	admin := createTestUser(t, sqlxDB, "admin@test.com", "Admin User", string(models.RoleAdmin))
	teacher := createTestUser(t, sqlxDB, "teacher@test.com", "Teacher User", string(models.RoleMethodologist))

	// Создаём прошлое занятие (2 часа назад)
	pastTime := time.Now().Add(-2 * time.Hour)
	pastLesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   pastTime,
		EndTime:     pastTime.Add(2 * time.Hour),
		MaxStudents: 5,
		Color:       "#3b82f6",
	}
	err := lessonRepo.Create(ctx, pastLesson)
	require.NoError(t, err)

	// Обновляем занятие (изменяем subject)
	updateReq := models.UpdateLessonRequest{
		Subject: stringPtr("Обновлённое прошлое занятие"),
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/lessons/"+pastLesson.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.SetUserInContext(req.Context(), admin))

	// Роутинг с chi для правильного парсинга параметров
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", pastLesson.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateLesson(rr, req)

	// Проверяем что запрос успешен
	assert.Equal(t, http.StatusOK, rr.Code, "Admin should be able to edit past lessons")

	// Проверяем что в ответе есть warning
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "Response should contain data field")

	warningText, hasWarning := dataMap["warning"].(string)
	assert.True(t, hasWarning, "Response should contain warning for past lesson")
	assert.NotEmpty(t, warningText, "Warning text should not be empty")

	// Проверяем что занятие реально обновилось
	updatedLesson, err := lessonRepo.GetByID(ctx, pastLesson.ID)
	require.NoError(t, err)
	assert.True(t, updatedLesson.Subject.Valid)
	assert.Equal(t, "Обновлённое прошлое занятие", updatedLesson.Subject.String)
}

// Тест: Teacher не может редактировать subject занятия (только homework_text)
func TestUpdateLesson_Teacher_CannotEditSubject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	sqlxDB := setupTestDB(t)
	defer cleanupTestDB(t, sqlxDB)
	pool := setupTestDBForHandler(t, sqlxDB)
	// Note: don't close pool - it's a shared test pool managed globally

	// Создаём репозитории
	lessonRepo := repository.NewLessonRepository(sqlxDB)
	userRepo := repository.NewUserRepository(sqlxDB)
	bookingRepo := repository.NewBookingRepository(sqlxDB)
	creditRepo := repository.NewCreditRepository(sqlxDB)

	// Создаём сервисы
	lessonService := service.NewLessonService(lessonRepo, userRepo)
	bookingService := service.NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		repository.NewCancelledBookingRepository(sqlxDB),
		validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo),
	)
	bulkEditService := service.NewBulkEditService(
		pool,
		lessonRepo,
		repository.NewLessonModificationRepository(sqlxDB),
		userRepo,
		creditRepo,
	)

	// Создаём handler
	handler := NewLessonHandler(lessonService, bookingService, bulkEditService)

	// Создаём преподавателя
	teacher := createTestUser(t, sqlxDB, "teacher@test.com", "Teacher User", string(models.RoleMethodologist))

	// Создаём будущее занятие (учитель будет пытаться изменить своё занятие)
	futureTime := time.Now().Add(2 * time.Hour)
	futureLesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   futureTime,
		EndTime:     futureTime.Add(2 * time.Hour),
		MaxStudents: 5,
		Color:       "#3b82f6",
	}
	err := lessonRepo.Create(ctx, futureLesson)
	require.NoError(t, err)

	// Пытаемся обновить subject как преподаватель (должно быть запрещено)
	updateReq := models.UpdateLessonRequest{
		Subject: stringPtr("Попытка обновить"),
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/lessons/"+futureLesson.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.SetUserInContext(req.Context(), teacher))

	// Роутинг с chi
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", futureLesson.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateLesson(rr, req)

	// Ожидаем 403 Forbidden (teacher может редактировать только homework_text)
	assert.Equal(t, http.StatusForbidden, rr.Code, "Teacher should NOT be able to edit subject")
}

// Тест: Teacher МОЖЕТ редактировать homework_text своего занятия
func TestUpdateLesson_Teacher_CanEditHomeworkText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	sqlxDB := setupTestDB(t)
	defer cleanupTestDB(t, sqlxDB)
	pool := setupTestDBForHandler(t, sqlxDB)
	// Note: don't close pool - it's a shared test pool managed globally

	// Создаём репозитории
	lessonRepo := repository.NewLessonRepository(sqlxDB)
	userRepo := repository.NewUserRepository(sqlxDB)
	bookingRepo := repository.NewBookingRepository(sqlxDB)
	creditRepo := repository.NewCreditRepository(sqlxDB)

	// Создаём сервисы
	lessonService := service.NewLessonService(lessonRepo, userRepo)
	bookingService := service.NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		repository.NewCancelledBookingRepository(sqlxDB),
		validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo),
	)
	bulkEditService := service.NewBulkEditService(
		pool,
		lessonRepo,
		repository.NewLessonModificationRepository(sqlxDB),
		userRepo,
		creditRepo,
	)

	// Создаём handler
	handler := NewLessonHandler(lessonService, bookingService, bulkEditService)

	// Создаём преподавателя
	teacher := createTestUser(t, sqlxDB, "teacher@test.com", "Teacher User", string(models.RoleMethodologist))

	// Создаём будущее занятие
	futureTime := time.Now().Add(2 * time.Hour)
	futureLesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   futureTime,
		EndTime:     futureTime.Add(2 * time.Hour),
		MaxStudents: 5,
		Color:       "#3b82f6",
	}
	err := lessonRepo.Create(ctx, futureLesson)
	require.NoError(t, err)

	// Обновляем homework_text как преподаватель (должно работать)
	homeworkText := "Домашнее задание от преподавателя"
	updateReq := models.UpdateLessonRequest{
		HomeworkText: &homeworkText,
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/lessons/"+futureLesson.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.SetUserInContext(req.Context(), teacher))

	// Роутинг с chi
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", futureLesson.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateLesson(rr, req)

	// Ожидаем 200 OK
	assert.Equal(t, http.StatusOK, rr.Code, "Teacher should be able to edit homework_text")

	// Проверяем что homework_text обновился
	updatedLesson, err := lessonRepo.GetByID(ctx, futureLesson.ID)
	require.NoError(t, err)
	assert.True(t, updatedLesson.HomeworkText.Valid)
	assert.Equal(t, "Домашнее задание от преподавателя", updatedLesson.HomeworkText.String)
}

// Тест: Teacher НЕ может редактировать homework_text чужого занятия
func TestUpdateLesson_Teacher_CannotEditOtherTeacherHomework(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	sqlxDB := setupTestDB(t)
	defer cleanupTestDB(t, sqlxDB)
	pool := setupTestDBForHandler(t, sqlxDB)
	// Note: don't close pool - it's a shared test pool managed globally

	// Создаём репозитории
	lessonRepo := repository.NewLessonRepository(sqlxDB)
	userRepo := repository.NewUserRepository(sqlxDB)
	bookingRepo := repository.NewBookingRepository(sqlxDB)
	creditRepo := repository.NewCreditRepository(sqlxDB)

	// Создаём сервисы
	lessonService := service.NewLessonService(lessonRepo, userRepo)
	bookingService := service.NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		repository.NewCancelledBookingRepository(sqlxDB),
		validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo),
	)
	bulkEditService := service.NewBulkEditService(
		pool,
		lessonRepo,
		repository.NewLessonModificationRepository(sqlxDB),
		userRepo,
		creditRepo,
	)

	// Создаём handler
	handler := NewLessonHandler(lessonService, bookingService, bulkEditService)

	// Создаём двух преподавателей
	teacher1 := createTestUser(t, sqlxDB, "teacher1@test.com", "Teacher One", string(models.RoleMethodologist))
	teacher2 := createTestUser(t, sqlxDB, "teacher2@test.com", "Teacher Two", string(models.RoleMethodologist))

	// Создаём занятие teacher1
	futureTime := time.Now().Add(2 * time.Hour)
	futureLesson := &models.Lesson{
		TeacherID:   teacher1.ID,
		StartTime:   futureTime,
		EndTime:     futureTime.Add(2 * time.Hour),
		MaxStudents: 5,
		Color:       "#3b82f6",
	}
	err := lessonRepo.Create(ctx, futureLesson)
	require.NoError(t, err)

	// Teacher2 пытается обновить homework_text занятия teacher1
	homeworkText := "Попытка изменить чужое ДЗ"
	updateReq := models.UpdateLessonRequest{
		HomeworkText: &homeworkText,
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/lessons/"+futureLesson.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.SetUserInContext(req.Context(), teacher2))

	// Роутинг с chi
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", futureLesson.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateLesson(rr, req)

	// Ожидаем 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rr.Code, "Teacher should NOT be able to edit another teacher's lesson")
}

// Тест: Student не может редактировать занятия
func TestUpdateLesson_Student_CannotEditLesson(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	sqlxDB := setupTestDB(t)
	defer cleanupTestDB(t, sqlxDB)
	pool := setupTestDBForHandler(t, sqlxDB)
	// Note: don't close pool - it's a shared test pool managed globally

	// Создаём репозитории
	lessonRepo := repository.NewLessonRepository(sqlxDB)
	userRepo := repository.NewUserRepository(sqlxDB)
	bookingRepo := repository.NewBookingRepository(sqlxDB)
	creditRepo := repository.NewCreditRepository(sqlxDB)

	// Создаём сервисы
	lessonService := service.NewLessonService(lessonRepo, userRepo)
	bookingService := service.NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		repository.NewCancelledBookingRepository(sqlxDB),
		validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo),
	)
	bulkEditService := service.NewBulkEditService(
		pool,
		lessonRepo,
		repository.NewLessonModificationRepository(sqlxDB),
		userRepo,
		creditRepo,
	)

	// Создаём handler
	handler := NewLessonHandler(lessonService, bookingService, bulkEditService)

	// Создаём студента и преподавателя
	student := createTestUser(t, sqlxDB, "student@test.com", "Student User", string(models.RoleStudent))
	teacher := createTestUser(t, sqlxDB, "teacher@test.com", "Teacher User", string(models.RoleMethodologist))

	// Создаём будущее занятие
	futureTime := time.Now().Add(2 * time.Hour)
	futureLesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   futureTime,
		EndTime:     futureTime.Add(2 * time.Hour),
		MaxStudents: 5,
		Color:       "#3b82f6",
	}
	err := lessonRepo.Create(ctx, futureLesson)
	require.NoError(t, err)

	// Пытаемся обновить как студент
	updateReq := models.UpdateLessonRequest{
		Subject: stringPtr("Попытка студента"),
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/lessons/"+futureLesson.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.SetUserInContext(req.Context(), student))

	// Роутинг с chi
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", futureLesson.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateLesson(rr, req)

	// Ожидаем 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rr.Code, "Student should NOT be able to edit lessons")
}

// Тест: Админ редактирует будущее занятие (без warning)
func TestUpdateLesson_Admin_EditFutureLesson_NoWarning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	ctx := context.Background()
	sqlxDB := setupTestDB(t)
	defer cleanupTestDB(t, sqlxDB)
	pool := setupTestDBForHandler(t, sqlxDB)
	// Note: don't close pool - it's a shared test pool managed globally

	// Создаём репозитории
	lessonRepo := repository.NewLessonRepository(sqlxDB)
	userRepo := repository.NewUserRepository(sqlxDB)
	bookingRepo := repository.NewBookingRepository(sqlxDB)
	creditRepo := repository.NewCreditRepository(sqlxDB)

	// Создаём сервисы
	lessonService := service.NewLessonService(lessonRepo, userRepo)
	bookingService := service.NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		repository.NewCancelledBookingRepository(sqlxDB),
		validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo),
	)
	bulkEditService := service.NewBulkEditService(
		pool,
		lessonRepo,
		repository.NewLessonModificationRepository(sqlxDB),
		userRepo,
		creditRepo,
	)

	// Создаём handler
	handler := NewLessonHandler(lessonService, bookingService, bulkEditService)

	// Создаём админа и преподавателя
	admin := createTestUser(t, sqlxDB, "admin@test.com", "Admin User", string(models.RoleAdmin))
	teacher := createTestUser(t, sqlxDB, "teacher@test.com", "Teacher User", string(models.RoleMethodologist))

	// Создаём будущее занятие
	futureTime := time.Now().Add(2 * time.Hour)
	futureLesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   futureTime,
		EndTime:     futureTime.Add(2 * time.Hour),
		MaxStudents: 5,
		Color:       "#3b82f6",
	}
	err := lessonRepo.Create(ctx, futureLesson)
	require.NoError(t, err)

	// Обновляем занятие
	updateReq := models.UpdateLessonRequest{
		Subject: stringPtr("Обновлённое будущее занятие"),
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/lessons/"+futureLesson.ID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.SetUserInContext(req.Context(), admin))

	// Роутинг с chi
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", futureLesson.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.UpdateLesson(rr, req)

	// Проверяем успех
	assert.Equal(t, http.StatusOK, rr.Code, "Admin should be able to edit future lessons")

	// Проверяем что warning ОТСУТСТВУЕТ
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	dataMap, ok := response["data"].(map[string]interface{})
	require.True(t, ok, "Response should contain data field")

	_, hasWarning := dataMap["warning"]
	assert.False(t, hasWarning, "Response should NOT contain warning for future lesson")

	// Проверяем что занятие обновилось
	updatedLesson, err := lessonRepo.GetByID(ctx, futureLesson.ID)
	require.NoError(t, err)
	assert.True(t, updatedLesson.Subject.Valid)
	assert.Equal(t, "Обновлённое будущее занятие", updatedLesson.Subject.String)
}

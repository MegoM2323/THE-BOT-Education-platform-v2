package handlers

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/internal/validator"
)

// TestBookingHandler_AccessControl проверяет, что преподаватели не могут видеть чужие бронирования
func TestBookingHandler_AccessControl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	skipIfProduction(t)

	// Setup
	ctx := context.Background()
	sqlxDB := setupTestDB(t)
	defer cleanupTestDB(t, sqlxDB)
	pool := setupTestDBForHandler(t, sqlxDB)
	// Note: don't close pool - it's a shared test pool managed globally

	// Создаём репозитории
	bookingRepo := repository.NewBookingRepository(sqlxDB)
	lessonRepo := repository.NewLessonRepository(sqlxDB)
	creditRepo := repository.NewCreditRepository(sqlxDB)
	cancelledBookingRepo := repository.NewCancelledBookingRepository(sqlxDB)

	// Создаём валидатор и сервис
	bookingValidator := validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo)
	bookingService := service.NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		cancelledBookingRepo,
		bookingValidator,
	)

	// Создаём handler
	bookingHandler := NewBookingHandler(bookingService)

	// Создаём тестовые данные: два студента, два учителя, два занятия
	teacher1 := createTestUser(t, sqlxDB, "teacher1@test.com", "Test Teacher 1", string(models.RoleMethodologist))
	teacher2 := createTestUser(t, sqlxDB, "teacher2@test.com", "Test Teacher 2", string(models.RoleMethodologist))
	admin := createTestUser(t, sqlxDB, "admin@test.com", "Test Admin", string(models.RoleAdmin))
	student := createTestUser(t, sqlxDB, "student@test.com", "Test Student", string(models.RoleStudent))
	student2 := createTestUser(t, sqlxDB, "student2@test.com", "Test Student 2", string(models.RoleStudent))

	// Даём студентам кредиты
	addCreditToStudent(t, sqlxDB, student.ID, 100)
	addCreditToStudent(t, sqlxDB, student2.ID, 100)

	// Создаём два занятия от разных учителей (через 24 часа = можем отменить)
	futureTime := time.Now().Add(24 * time.Hour)
	lesson1ID := createLesson(t, sqlxDB, teacher1.ID, 1, futureTime)
	// Make lesson2 start after lesson1 ends (lesson1 is 2 hours, so add 2+ hours)
	lesson2StartTime := futureTime.Add(3 * time.Hour)
	lesson2ID := createLesson(t, sqlxDB, teacher2.ID, 1, lesson2StartTime)

	// Студент 1 бронирует первое занятие
	booking1, err := bookingService.CreateBooking(ctx, &models.CreateBookingRequest{
		StudentID: student.ID,
		LessonID:  lesson1ID,
		IsAdmin:   false,
	})
	require.NoError(t, err)

	// Студент 2 бронирует второе занятие
	booking2, err := bookingService.CreateBooking(ctx, &models.CreateBookingRequest{
		StudentID: student2.ID,
		LessonID:  lesson2ID,
		IsAdmin:   false,
	})
	require.NoError(t, err)

	// Table-driven tests
	tests := []struct {
		name           string
		bookingID      uuid.UUID
		user           *models.User
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Student can see own booking",
			bookingID:      booking1.ID,
			user:           student,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Student cannot see other booking",
			bookingID:      booking2.ID,
			user:           student,
			expectedStatus: http.StatusForbidden,
			expectedError:  "Unauthorized access to booking",
		},
		{
			name:           "Teacher can see own lesson booking",
			bookingID:      booking1.ID,
			user:           teacher1,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Teacher cannot see other teacher's booking",
			bookingID:      booking1.ID,
			user:           teacher2,
			expectedStatus: http.StatusForbidden,
			expectedError:  "Unauthorized access to booking",
		},
		{
			name:           "Admin can see any booking",
			bookingID:      booking1.ID,
			user:           admin,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Admin can see any booking (lesson2)",
			bookingID:      booking2.ID,
			user:           admin,
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём запрос с контекстом пользователя
			req := httptest.NewRequest("GET", "/api/v1/bookings/"+tt.bookingID.String(), nil)
			ctx := middleware.SetUserInContext(req.Context(), tt.user)

			// Add URL param for chi routing
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.bookingID.String())
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

			req = req.WithContext(ctx)

			// Создаём response writer
			w := httptest.NewRecorder()

			// Вызываем handler
			bookingHandler.GetBooking(w, req)

			// Проверяем результат
			assert.Equal(t, tt.expectedStatus, w.Code, "Unexpected status code for %s", tt.name)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Failed to unmarshal response for %s", tt.name)

				// Проверяем структуру ответа
				assert.NotNil(t, response["data"], "Response should have data field for %s", tt.name)
			} else {
				// Проверяем error message
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Failed to unmarshal error response for %s", tt.name)

				if tt.expectedError != "" {
					// Error response has structure: {"error": {"code": "...", "message": "..."}}
					errorObj, ok := response["error"].(map[string]interface{})
					assert.True(t, ok, "Response should have error field for %s", tt.name)
					if ok {
						msg, ok := errorObj["message"].(string)
						assert.True(t, ok, "Error object should have message field for %s", tt.name)
						assert.Equal(t, tt.expectedError, msg, "Unexpected error message for %s", tt.name)
					}
				}
			}
		})
	}
}

// TestBookingHandler_GetBookingStatus_AccessControl проверяет access control для GetBookingStatus
func TestBookingHandler_GetBookingStatus_AccessControl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	skipIfProduction(t)

	// Setup
	ctx := context.Background()
	sqlxDB := setupTestDB(t)
	defer cleanupTestDB(t, sqlxDB)
	pool := setupTestDBForHandler(t, sqlxDB)
	// Note: don't close pool - it's a shared test pool managed globally

	// Создаём репозитории
	bookingRepo := repository.NewBookingRepository(sqlxDB)
	lessonRepo := repository.NewLessonRepository(sqlxDB)
	creditRepo := repository.NewCreditRepository(sqlxDB)
	cancelledBookingRepo := repository.NewCancelledBookingRepository(sqlxDB)

	// Создаём валидатор и сервис
	bookingValidator := validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo)
	bookingService := service.NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		cancelledBookingRepo,
		bookingValidator,
	)

	// Создаём handler
	bookingHandler := NewBookingHandler(bookingService)

	// Создаём тестовые данные
	teacher1 := createTestUser(t, sqlxDB, "teacher1@test.com", "Test Teacher 1", string(models.RoleMethodologist))
	teacher2 := createTestUser(t, sqlxDB, "teacher2@test.com", "Test Teacher 2", string(models.RoleMethodologist))
	student := createTestUser(t, sqlxDB, "student@test.com", "Test Student", string(models.RoleStudent))
	student2 := createTestUser(t, sqlxDB, "student2@test.com", "Test Student 2", string(models.RoleStudent))

	// Даём студенту кредиты
	addCreditToStudent(t, sqlxDB, student.ID, 100)
	addCreditToStudent(t, sqlxDB, student2.ID, 100)

	// Создаём два занятия от разных учителей
	futureTime := time.Now().Add(24 * time.Hour)
	lesson1ID := createLesson(t, sqlxDB, teacher1.ID, 1, futureTime)
	lesson2StartTime := futureTime.Add(3 * time.Hour)
	lesson2ID := createLesson(t, sqlxDB, teacher2.ID, 1, lesson2StartTime)

	// Студент 1 бронирует первое занятие
	booking1, err := bookingService.CreateBooking(ctx, &models.CreateBookingRequest{
		StudentID: student.ID,
		LessonID:  lesson1ID,
		IsAdmin:   false,
	})
	require.NoError(t, err)

	// Студент 2 бронирует второе занятие
	booking2, err := bookingService.CreateBooking(ctx, &models.CreateBookingRequest{
		StudentID: student2.ID,
		LessonID:  lesson2ID,
		IsAdmin:   false,
	})
	require.NoError(t, err)

	// Table-driven tests
	tests := []struct {
		name           string
		bookingID      uuid.UUID
		user           *models.User
		expectedStatus int
	}{
		{
			name:           "Student can see own booking status",
			bookingID:      booking1.ID,
			user:           student,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Student cannot see other booking status",
			bookingID:      booking2.ID,
			user:           student,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Teacher can see own lesson booking status",
			bookingID:      booking1.ID,
			user:           teacher1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Teacher cannot see other teacher's booking status",
			bookingID:      booking1.ID,
			user:           teacher2,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём запрос с контекстом пользователя
			req := httptest.NewRequest("GET", "/api/v1/bookings/"+tt.bookingID.String()+"/status", nil)
			ctx := middleware.SetUserInContext(req.Context(), tt.user)

			// Add URL param for chi routing
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.bookingID.String())
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

			req = req.WithContext(ctx)

			// Создаём response writer
			w := httptest.NewRecorder()

			// Вызываем handler
			bookingHandler.GetBookingStatus(w, req)

			// Проверяем результат
			assert.Equal(t, tt.expectedStatus, w.Code, "Unexpected status code for %s", tt.name)
		})
	}
}

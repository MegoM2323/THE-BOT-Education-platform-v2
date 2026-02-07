package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/database"
	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/internal/validator"
	"tutoring-platform/pkg/response"
)

// skipIfProduction пропускает тест если запущен production database
func skipIfProduction(t *testing.T) {
	t.Helper()
	if os.Getenv("ENV") == "production" {
		t.Skip("Skipping test in production environment")
	}
}

// setupTestDBForHandler returns the shared test database pool.
// Uses database.GetTestPool to avoid connection exhaustion.
func setupTestDBForHandler(t *testing.T, sqlxDB *sqlx.DB) *pgxpool.Pool {
	t.Helper()

	return database.GetTestPool(t)
}

// TestBookingHandler_CreateBooking_PreviouslyCancelled проверяет HTTP handler для повторной записи
func TestBookingHandler_CreateBooking_PreviouslyCancelled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	skipIfProduction(t)

	// Setup
	ctx := context.Background()
	sqlxDB := setupTestDB(t)
	defer cleanupTestDB(t, sqlxDB)
	pool := setupTestDBForHandler(t, sqlxDB)

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
		nil, // telegramService
		nil, // userRepo
	)

	// Создаём handler
	bookingHandler := NewBookingHandler(bookingService)

	// Создаём тестовые данные: студент, учитель, занятие
	student := createTestUser(t, sqlxDB, "student@test.com", "Test Student", string(models.RoleStudent))
	teacher := createTestUser(t, sqlxDB, "teacher@test.com", "Test Teacher", string(models.RoleTeacher))

	// Даём студенту кредиты
	addCreditToStudent(t, sqlxDB, student.ID, 10)

	// Создаём занятие в будущем (через 48 часов = можем отменить)
	futureTime := time.Now().Add(48 * time.Hour)
	lessonID := createLesson(t, sqlxDB, teacher.ID, 1, futureTime)

	// Шаг 1: Студент бронирует занятие
	booking, err := bookingService.CreateBooking(ctx, &models.CreateBookingRequest{
		StudentID: student.ID,
		LessonID:  lessonID,
		IsAdmin:   false,
	})
	require.NoError(t, err)
	require.NotNil(t, booking)

	// Шаг 2: Студент отменяет занятие (> 24 часа)
	result, err := bookingService.CancelBooking(ctx, &models.CancelBookingRequest{
		BookingID: booking.ID,
		StudentID: student.ID,
		IsAdmin:   false,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, models.CancelResultSuccess, result.Status)

	// Шаг 3: Студент пытается забронировать снова через API handler
	reqBody := models.CreateBookingRequest{
		LessonID: lessonID,
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Добавляем студента в контекст (как middleware делает)
	ctx = context.WithValue(req.Context(), middleware.UserContextKey, student)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	// Выполняем запрос
	bookingHandler.CreateBooking(recorder, req)

	// Проверяем ответ
	assert.Equal(t, http.StatusForbidden, recorder.Code, "Expected 403 Forbidden")

	var resp response.ErrorResponse
	err = json.NewDecoder(recorder.Body).Decode(&resp)
	require.NoError(t, err, "Should decode error response")

	assert.False(t, resp.Success, "Response success should be false")
	assert.Equal(t, response.ErrCodeLessonPreviouslyCancelled, resp.Error.Code, "Error code should be LESSON_PREVIOUSLY_CANCELLED")
	assert.Equal(t, "Вы отписались от этого занятия и больше не можете на него записаться", resp.Error.Message, "Error message should be in Russian")
}

// TestBookingHandler_ErrorHandling проверяет обработку различных ошибок в handler
func TestBookingHandler_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	skipIfProduction(t)

	ctx := context.Background()
	sqlxDB := setupTestDB(t)
	defer cleanupTestDB(t, sqlxDB)
	pool := setupTestDBForHandler(t, sqlxDB)

	// Создаём репозитории и сервисы
	bookingRepo := repository.NewBookingRepository(sqlxDB)
	lessonRepo := repository.NewLessonRepository(sqlxDB)
	creditRepo := repository.NewCreditRepository(sqlxDB)
	cancelledBookingRepo := repository.NewCancelledBookingRepository(sqlxDB)

	bookingValidator := validator.NewBookingValidator(lessonRepo, bookingRepo, creditRepo)
	bookingService := service.NewBookingService(
		pool,
		bookingRepo,
		lessonRepo,
		creditRepo,
		cancelledBookingRepo,
		bookingValidator,
		nil, // telegramService
		nil, // userRepo
	)

	bookingHandler := NewBookingHandler(bookingService)

	// Создаём тестовые данные
	student := createTestUser(t, sqlxDB, "student2@test.com", "Student 2", string(models.RoleStudent))
	teacher := createTestUser(t, sqlxDB, "teacher2@test.com", "Teacher 2", string(models.RoleTeacher))

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T) uuid.UUID // возвращает lesson ID
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name: "Insufficient credits",
			setupFunc: func(t *testing.T) uuid.UUID {
				// Студент БЕЗ кредитов
				futureTime := time.Now().Add(48 * time.Hour)
				lessonID := createLesson(t, sqlxDB, teacher.ID, 4, futureTime)
				return lessonID
			},
			expectedStatus: http.StatusConflict,
			expectedCode:   response.ErrCodeInsufficientCredits,
			expectedMsg:    "You don't have enough credits",
		},
		{
			name: "Lesson previously cancelled",
			setupFunc: func(t *testing.T) uuid.UUID {
				// Даём 5 кредитов
				addCreditToStudent(t, sqlxDB, student.ID, 5)

				futureTime := time.Now().Add(48 * time.Hour)
				lessonID := createLesson(t, sqlxDB, teacher.ID, 1, futureTime)

				// Бронируем и отменяем
				booking, err := bookingService.CreateBooking(ctx, &models.CreateBookingRequest{
					StudentID: student.ID,
					LessonID:  lessonID,
					IsAdmin:   false,
				})
				require.NoError(t, err)

				result, err := bookingService.CancelBooking(ctx, &models.CancelBookingRequest{
					BookingID: booking.ID,
					StudentID: student.ID,
					IsAdmin:   false,
				})
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, models.CancelResultSuccess, result.Status)

				return lessonID
			},
			expectedStatus: http.StatusForbidden,
			expectedCode:   response.ErrCodeLessonPreviouslyCancelled,
			expectedMsg:    "Вы отписались от этого занятия и больше не можете на него записаться",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lessonID := tt.setupFunc(t)

			reqBody := models.CreateBookingRequest{
				LessonID: lessonID,
			}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), middleware.UserContextKey, student)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()
			bookingHandler.CreateBooking(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			var resp response.ErrorResponse
			err := json.NewDecoder(recorder.Body).Decode(&resp)
			require.NoError(t, err)

			assert.False(t, resp.Success)
			assert.Equal(t, tt.expectedCode, resp.Error.Code)
			assert.Equal(t, tt.expectedMsg, resp.Error.Message)
		})
	}
}

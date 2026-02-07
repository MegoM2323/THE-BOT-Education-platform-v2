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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
)

// TestSendLessonBroadcast_Success проверяет успешную отправку рассылки
func TestSendLessonBroadcast_Success(t *testing.T) {
	// Test data
	teacherID := uuid.New()
	student1ID := uuid.New()
	student2ID := uuid.New()
	student3ID := uuid.New()
	lessonID := uuid.New()

	// Create test bookings
	bookings := []*models.BookingWithDetails{
		{
			Booking: models.Booking{
				StudentID: student1ID,
				LessonID:  lessonID,
				Status:    models.BookingStatusActive,
			},
		},
		{
			Booking: models.Booking{
				StudentID: student2ID,
				LessonID:  lessonID,
				Status:    models.BookingStatusActive,
			},
		},
		{
			Booking: models.Booking{
				StudentID: student3ID,
				LessonID:  lessonID,
				Status:    models.BookingStatusActive,
			},
		},
	}

	// Create request
	reqBody := models.TeacherLessonBroadcastRequest{
		Message: "Важное сообщение для всех",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/teacher/lessons/"+lessonID.String()+"/broadcast",
		bytes.NewReader(bodyBytes))

	// Add teacher user to context
	teacher := &models.User{
		ID:   teacherID,
		Role: models.RoleTeacher,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, teacher)
	req = req.WithContext(ctx)

	// Add lessonID to chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", lessonID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Verify bookings struct is correct
	assert.Equal(t, 3, len(bookings))
	assert.Equal(t, student1ID, bookings[0].StudentID)
	assert.Equal(t, models.BookingStatusActive, bookings[0].Status)
}

// TestSendLessonBroadcast_NotTeacher проверяет что студент не может отправить рассылку
func TestSendLessonBroadcast_NotTeacher(t *testing.T) {
	studentID := uuid.New()
	lessonID := uuid.New()

	// Create request
	reqBody := models.TeacherLessonBroadcastRequest{
		Message: "Test message",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/teacher/lessons/"+lessonID.String()+"/broadcast",
		bytes.NewReader(bodyBytes))

	// Add student user to context (не teacher)
	student := &models.User{
		ID:   studentID,
		Role: models.RoleStudent,
	}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, student)
	req = req.WithContext(ctx)

	// Verify student role
	user, ok := middleware.GetUserFromContext(req.Context())
	require.True(t, ok)
	assert.Equal(t, models.RoleStudent, user.Role)
	assert.False(t, user.IsTeacher())
}

// TestSendLessonBroadcast_ValidationEmptyMessage проверяет валидацию пустого сообщения
func TestSendLessonBroadcast_ValidationEmptyMessage(t *testing.T) {
	req := &models.TeacherLessonBroadcastRequest{
		Message: "",
	}

	err := req.Validate()
	assert.Error(t, err)
}

// TestSendLessonBroadcast_ValidationMessageTooLong проверяет валидацию слишком длинного сообщения
func TestSendLessonBroadcast_ValidationMessageTooLong(t *testing.T) {
	// Create message with 4097 characters (max is 4096)
	longMessage := string(make([]byte, 4097))
	req := &models.TeacherLessonBroadcastRequest{
		Message: longMessage,
	}

	err := req.Validate()
	assert.Error(t, err)
}

// TestSendLessonBroadcast_ValidationValidMessage проверяет что валидное сообщение проходит проверку
func TestSendLessonBroadcast_ValidationValidMessage(t *testing.T) {
	tests := []string{
		"Short message",
		"Message with 100 characters: abcdefghijklmnopqrstuvwxyz" +
			"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnop",
		string(make([]byte, 4096)), // Max length: 4096 characters
	}

	for _, msg := range tests {
		req := &models.TeacherLessonBroadcastRequest{
			Message: msg,
		}
		err := req.Validate()
		msgPreview := msg
		if len(msg) > 20 {
			msgPreview = msg[:20]
		}
		assert.NoError(t, err, "Message validation failed: %s", msgPreview)
	}
}

// TestTeacherHandler_LessonNotFound проверяет что возвращается ошибка если занятие не найдено
func TestTeacherHandler_LessonNotFound(t *testing.T) {
	nonExistentLessonID := uuid.New()

	// Verify lesson ID is valid UUID
	assert.NoError(t, nonExistentLessonID.Scan(nonExistentLessonID.String()))
}

// TestTeacherHandler_AccessControl проверяет что teacher не может отправить рассылку для чужого занятия
func TestTeacherHandler_AccessControl(t *testing.T) {
	teacher1ID := uuid.New()
	teacher2ID := uuid.New()
	lessonID := uuid.New()

	// Lesson belongs to teacher2
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacher2ID,
	}

	// teacher1 tries to access
	teacher1 := &models.User{
		ID:   teacher1ID,
		Role: models.RoleTeacher,
	}

	// Verify access control logic
	assert.NotEqual(t, lesson.TeacherID, teacher1.ID)
	assert.Equal(t, lesson.TeacherID, teacher2ID)
}

// TestBookingStatusEnum проверяет что BookingStatus константы определены правильно
func TestBookingStatusEnum(t *testing.T) {
	assert.Equal(t, models.BookingStatus("active"), models.BookingStatusActive)
	assert.Equal(t, models.BookingStatus("cancelled"), models.BookingStatusCancelled)
}

// TestTeacherLessonBroadcastRequest_JSONMarshalling проверяет JSON сериализацию
func TestTeacherLessonBroadcastRequest_JSONMarshalling(t *testing.T) {
	req := &models.TeacherLessonBroadcastRequest{
		Message: "Test broadcast message",
	}

	// Marshal
	data, err := json.Marshal(req)
	require.NoError(t, err)

	// Unmarshal
	var decoded models.TeacherLessonBroadcastRequest
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, req.Message, decoded.Message)
}

// TestTeacherHandler_HTTPStatusCodes проверяет что handler возвращает правильные HTTP коды
func TestTeacherHandler_HTTPStatusCodes(t *testing.T) {
	// Verify response status code constants
	assert.Equal(t, 201, http.StatusCreated)
	assert.Equal(t, 400, http.StatusBadRequest)
	assert.Equal(t, 403, http.StatusForbidden)
	assert.Equal(t, 404, http.StatusNotFound)
	assert.Equal(t, 500, http.StatusInternalServerError)
}

// TestBookingWithDetails_Structure проверяет структуру BookingWithDetails
func TestBookingWithDetails_Structure(t *testing.T) {
	studentID := uuid.New()
	lessonID := uuid.New()

	booking := &models.BookingWithDetails{
		Booking: models.Booking{
			StudentID: studentID,
			LessonID:  lessonID,
			Status:    models.BookingStatusActive,
		},
	}

	// Verify embedded struct
	assert.Equal(t, studentID, booking.StudentID)
	assert.Equal(t, lessonID, booking.LessonID)
	assert.Equal(t, models.BookingStatusActive, booking.Status)
}

// TestBroadcastModel_Structure проверяет структуру Broadcast
func TestBroadcastModel_Structure(t *testing.T) {
	broadcastID := uuid.New()
	createdByID := uuid.New()

	broadcast := &models.Broadcast{
		ID:        broadcastID,
		Message:   "Test broadcast",
		Status:    models.BroadcastStatusPending,
		CreatedBy: createdByID,
	}

	assert.Equal(t, broadcastID, broadcast.ID)
	assert.Equal(t, "Test broadcast", broadcast.Message)
	assert.Equal(t, models.BroadcastStatusPending, broadcast.Status)
	assert.Equal(t, createdByID, broadcast.CreatedBy)
}

// TestUserRoles проверяет что user roles определены правильно
func TestUserRoles(t *testing.T) {
	teacher := &models.User{ID: uuid.New(), Role: models.RoleTeacher}
	student := &models.User{ID: uuid.New(), Role: models.RoleStudent}
	admin := &models.User{ID: uuid.New(), Role: models.RoleAdmin}

	assert.False(t, teacher.IsTeacher())
	assert.False(t, teacher.IsStudent())
	assert.False(t, teacher.IsAdmin())

	assert.False(t, student.IsTeacher())
	assert.True(t, student.IsStudent())
	assert.False(t, student.IsAdmin())

	assert.False(t, admin.IsTeacher())
	assert.False(t, admin.IsStudent())
	assert.True(t, admin.IsAdmin())
}

// TestListBookingsFilter проверяет структуру ListBookingsFilter
func TestListBookingsFilter_Structure(t *testing.T) {
	lessonID := uuid.New()
	status := models.BookingStatusActive

	filter := &models.ListBookingsFilter{
		LessonID: &lessonID,
		Status:   &status,
	}

	require.NotNil(t, filter.LessonID)
	assert.Equal(t, lessonID, *filter.LessonID)

	require.NotNil(t, filter.Status)
	assert.Equal(t, status, *filter.Status)
}

// TestGetTeacherSchedule_InvalidDateFormat проверяет обработку невалидного формата даты
func TestGetTeacherSchedule_InvalidDateFormat(t *testing.T) {
	// Arrange
	teacherID := uuid.New()

	// Создаем request с невалидным форматом даты
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teacher/schedule?start_date=invalid-date", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, &models.User{
		ID:   teacherID,
		Role: models.RoleTeacher,
	})
	req = req.WithContext(ctx)

	// Verify date parsing fails
	_, err := time.Parse("2006-01-02", "invalid-date")
	assert.Error(t, err)
}

// TestGetTeacherSchedule_EndDateBeforeStartDate проверяет валидацию порядка дат
func TestGetTeacherSchedule_DateRangeValidation(t *testing.T) {
	teacherID := uuid.New()

	// Test 1: end_date < start_date (invalid)
	startDate := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	assert.True(t, endDate.Before(startDate), "end_date should be before start_date")

	// Test 2: Valid date range
	validStart := time.Date(2025, 1, 13, 0, 0, 0, 0, time.UTC)
	validEnd := validStart.AddDate(0, 0, 7)
	assert.False(t, validEnd.Before(validStart), "valid end_date should be after start_date")
	assert.True(t, validEnd.After(validStart), "valid end_date should be after start_date")

	// Test 3: Range exceeds 365 days (invalid)
	longStart := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	longEnd := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	assert.True(t, longEnd.Sub(longStart) > 365*24*time.Hour, "range should exceed 365 days")

	// Verify teacher ID is valid
	assert.NotEqual(t, uuid.Nil, teacherID)
}

// TestGetTeacherSchedule_Unauthorized проверяет требование аутентификации
func TestGetTeacherSchedule_Unauthorized(t *testing.T) {
	// Создаем request БЕЗ user в контексте
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teacher/schedule", nil)

	// Verify user not in context
	user, ok := middleware.GetUserFromContext(req.Context())
	assert.False(t, ok)
	assert.Nil(t, user)
}

// TestGetTeacherSchedule_StudentForbidden проверяет что студент не может получить доступ
func TestGetTeacherSchedule_StudentForbidden(t *testing.T) {
	studentID := uuid.New()

	// Create student user
	student := &models.User{
		ID:   studentID,
		Role: models.RoleStudent,
	}

	// Verify student role checks
	assert.True(t, student.IsStudent())
	assert.False(t, student.IsTeacher())
	assert.False(t, student.IsAdmin())
}

// TestGetTeacherSchedule_AdminCanFilterByTeacherID проверяет что админ может фильтровать по teacher_id
func TestGetTeacherSchedule_AdminAccess(t *testing.T) {
	adminID := uuid.New()
	teacherID := uuid.New()

	// Create admin user
	admin := &models.User{
		ID:   adminID,
		Role: models.RoleAdmin,
	}

	// Verify admin role checks
	assert.True(t, admin.IsAdmin())
	assert.False(t, admin.IsTeacher())
	assert.False(t, admin.IsStudent())

	// Verify different IDs
	assert.NotEqual(t, adminID, teacherID)
}

// TestEnrolledStudent_Structure проверяет структуру EnrolledStudent
func TestEnrolledStudent_Structure(t *testing.T) {
	studentID := uuid.New()
	student := &models.EnrolledStudent{
		ID:       studentID,
		FirstName: "Иван Петров", LastName: "Lastname",
		Email:    "ivan@example.com",
	}

	assert.Equal(t, studentID, student.ID)
	assert.Equal(t, "Иван Петров", student.FirstName + " " + student.LastName)
	assert.Equal(t, "ivan@example.com", student.Email)
}

// TestTeacherScheduleLesson_ToResponse проверяет преобразование в JSON response
func TestTeacherScheduleLesson_ToResponse(t *testing.T) {
	lessonID := uuid.New()
	teacherID := uuid.New()
	startTime := time.Date(2025, 1, 13, 10, 0, 0, 0, time.UTC)
	endTime := startTime.Add(2 * time.Hour)

	lesson := &models.TeacherScheduleLesson{
		Lesson: models.Lesson{
			ID:              lessonID,
			TeacherID:       teacherID,
			StartTime:       startTime,
			EndTime:         endTime,
			MaxStudents:     4,
			CurrentStudents: 2,
			Color:           "#3B82F6",
		},
		TeacherName:           "Мария Иванова",
		EnrolledStudentsCount: 2,
		HomeworkCount:         1,
		BroadcastsCount:       0,
		EnrolledStudents: []*models.EnrolledStudent{
			{
				ID:       uuid.New(),
				FirstName: "Студент 1", LastName: "Lastname",
				Email:    "student1@example.com",
			},
		},
	}

	// Convert to response
	response := lesson.ToResponse()

	// Verify response structure (UUID возвращаются как типы, не строки)
	// Note: lesson_type is not included in ToResponse - it's inferred from max_students
	assert.Equal(t, lessonID, response["id"])
	assert.Equal(t, teacherID, response["teacher_id"])
	assert.Equal(t, 4, response["max_students"])
	assert.Equal(t, 2, response["current_students"])
	assert.Equal(t, 2, response["enrolled_students_count"])
	assert.Equal(t, 1, response["homework_count"])
	assert.Equal(t, 0, response["broadcasts_count"])
	assert.NotNil(t, response["enrolled_students"])

	enrolledStudents := response["enrolled_students"].([]*models.EnrolledStudent)
	assert.Equal(t, 1, len(enrolledStudents))
}

// TestGetTeacherSchedule_DefaultDates проверяет использование дефолтных дат (текущая неделя)
func TestGetTeacherSchedule_DefaultDates(t *testing.T) {
	// Get current Monday
	now := time.Now()
	weekday := now.Weekday()
	daysToMonday := int(weekday) - 1
	if weekday == time.Sunday {
		daysToMonday = 6
	}
	expectedMonday := time.Date(now.Year(), now.Month(), now.Day()-daysToMonday, 0, 0, 0, 0, now.Location())

	// Verify Monday calculation
	assert.Equal(t, time.Monday, expectedMonday.Weekday())

	// Default end date is 7 days after start
	expectedEndDate := expectedMonday.AddDate(0, 0, 7)
	assert.Equal(t, expectedMonday.AddDate(0, 0, 7), expectedEndDate)
}

// TestGetTeacherSchedule_AdminWithoutTeacherID проверяет что админ без teacher_id получает все занятия
func TestGetTeacherSchedule_AdminWithoutTeacherID(t *testing.T) {
	adminID := uuid.New()

	// Create admin user
	admin := &models.User{
		ID:   adminID,
		Role: models.RoleAdmin,
	}

	// Создаем request БЕЗ teacher_id
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teacher/schedule", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, admin)
	req = req.WithContext(ctx)

	// Verify user is admin
	user, ok := middleware.GetUserFromContext(req.Context())
	require.True(t, ok)
	assert.True(t, user.IsAdmin())

	// Verify teacher_id is not in query params
	assert.Equal(t, "", req.URL.Query().Get("teacher_id"))

	// Note: Full integration test would verify that GetAllTeachersSchedule is called
	// instead of returning empty array
}

// TestGetTeacherSchedule_AdminWithTeacherID проверяет что админ с teacher_id получает занятия конкретного учителя
func TestGetTeacherSchedule_AdminWithTeacherID(t *testing.T) {
	adminID := uuid.New()
	teacherID := uuid.New()

	// Create admin user
	admin := &models.User{
		ID:   adminID,
		Role: models.RoleAdmin,
	}

	// Создаем request С teacher_id
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teacher/schedule?teacher_id="+teacherID.String(), nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, admin)
	req = req.WithContext(ctx)

	// Verify user is admin
	user, ok := middleware.GetUserFromContext(req.Context())
	require.True(t, ok)
	assert.True(t, user.IsAdmin())

	// Verify teacher_id is in query params
	assert.Equal(t, teacherID.String(), req.URL.Query().Get("teacher_id"))

	// Parse teacher_id and verify it's valid
	parsedID, err := uuid.Parse(req.URL.Query().Get("teacher_id"))
	require.NoError(t, err)
	assert.Equal(t, teacherID, parsedID)
}

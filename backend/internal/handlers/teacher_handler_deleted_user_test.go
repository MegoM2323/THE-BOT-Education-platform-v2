package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
)

// TestSendLessonBroadcastDeletedUser проверяет что удаленный пользователь не может отправить рассылку
func TestSendLessonBroadcastDeletedUser(t *testing.T) {
	tests := []struct {
		name           string
		user           *models.User
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "deleted user cannot send broadcast",
			user: &models.User{
				ID:             uuid.New(),
				Email:          "teacher@test.com",
				FullName:       "Teacher Name",
				Role:           models.RoleMethodologist,
				PaymentEnabled: false,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(-1 * time.Hour),
					Valid: true,
				},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "User account is deleted or deactivated",
		},
		{
			name: "active teacher can proceed to role check",
			user: &models.User{
				ID:             uuid.New(),
				Email:          "teacher@test.com",
				FullName:       "Teacher Name",
				Role:           models.RoleMethodologist,
				PaymentEnabled: false,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				DeletedAt: sql.NullTime{
					Valid: false,
				},
			},
			expectedStatus: http.StatusBadRequest, // Will fail at UUID parsing since we're not providing a valid lesson ID
			expectedMsg:    "",                    // Will be different error (UUID parse)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем mock handler
			handler := NewMethodologistHandler(nil, nil, nil, nil)

			// Создаем HTTP request
			req := httptest.NewRequest(http.MethodPost, "/api/v1/teacher/lessons/invalid-id/broadcast", nil)

			// Устанавливаем пользователя в контекст
			ctx := middleware.SetUserInContext(req.Context(), tt.user)
			req = req.WithContext(ctx)

			// Создаем response writer
			w := httptest.NewRecorder()

			// Вызываем handler
			handler.SendLessonBroadcast(w, req)

			// Проверяем status код
			assert.Equal(t, tt.expectedStatus, w.Code, "expected status %d, got %d", tt.expectedStatus, w.Code)

			// Проверяем сообщение об ошибке (если это тест удаленного пользователя)
			if tt.name == "deleted user cannot send broadcast" {
				body := w.Body.String()
				assert.Contains(t, body, tt.expectedMsg, "expected message %q in response", tt.expectedMsg)
			}
		})
	}
}

// TestGetTeacherScheduleDeletedUser проверяет что удаленный пользователь не может получить расписание
func TestGetTeacherScheduleDeletedUser(t *testing.T) {
	tests := []struct {
		name           string
		user           *models.User
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "deleted teacher cannot get schedule",
			user: &models.User{
				ID:             uuid.New(),
				Email:          "teacher@test.com",
				FullName:       "Teacher Name",
				Role:           models.RoleMethodologist,
				PaymentEnabled: false,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(-1 * time.Hour),
					Valid: true,
				},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "User account is deleted or deactivated",
		},
		{
			name: "deleted admin cannot get schedule",
			user: &models.User{
				ID:             uuid.New(),
				Email:          "admin@test.com",
				FullName:       "Admin Name",
				Role:           models.RoleAdmin,
				PaymentEnabled: false,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(-1 * time.Hour),
					Valid: true,
				},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "User account is deleted or deactivated",
		},
		{
			name: "active teacher can proceed to fetch schedule",
			user: &models.User{
				ID:             uuid.New(),
				Email:          "teacher@test.com",
				FullName:       "Teacher Name",
				Role:           models.RoleMethodologist,
				PaymentEnabled: false,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				DeletedAt: sql.NullTime{
					Valid: false,
				},
			},
			expectedStatus: http.StatusInternalServerError, // Fails due to nil services in mock
			expectedMsg:    "",                             // Different error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Пропускаем тест "active teacher" так как он требует mock repository
			// и не предназначен для проверки deleted user логики
			if tt.name == "active teacher can proceed to fetch schedule" {
				t.Skip("Skipping: requires mock repository, not testing deleted user logic")
				return
			}

			// Создаем mock handler
			handler := NewMethodologistHandler(nil, nil, nil, nil)

			// Создаем HTTP request
			req := httptest.NewRequest(http.MethodGet, "/api/v1/teacher/schedule", nil)

			// Устанавливаем пользователя в контекст
			ctx := middleware.SetUserInContext(req.Context(), tt.user)
			req = req.WithContext(ctx)

			// Создаем response writer
			w := httptest.NewRecorder()

			// Вызываем handler
			handler.GetMethodologistSchedule(w, req)

			// Проверяем status код (для deleted users должен быть 401)
			assert.Equal(t, tt.expectedStatus, w.Code, "expected status %d, got %d", tt.expectedStatus, w.Code)

			// Проверяем сообщение об ошибке
			body := w.Body.String()
			assert.Contains(t, body, tt.expectedMsg, "expected message %q in response", tt.expectedMsg)
		})
	}
}

// TestUserIsDeletedCheck проверяет что метод IsDeleted() работает корректно
func TestUserIsDeletedCheck(t *testing.T) {
	tests := []struct {
		name      string
		user      *models.User
		isDeleted bool
	}{
		{
			name: "user without DeletedAt is not deleted",
			user: &models.User{
				ID:        uuid.New(),
				DeletedAt: sql.NullTime{Valid: false},
			},
			isDeleted: false,
		},
		{
			name: "user with DeletedAt is deleted",
			user: &models.User{
				ID:        uuid.New(),
				DeletedAt: sql.NullTime{Time: time.Now(), Valid: true},
			},
			isDeleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.IsDeleted()
			assert.Equal(t, tt.isDeleted, result, "expected IsDeleted() = %v, got %v", tt.isDeleted, result)
		})
	}
}

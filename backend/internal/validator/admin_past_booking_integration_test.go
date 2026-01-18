package validator

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
)

// TestAdminCanBookPastLesson_Integration проверяет полный сценарий
// где администратор добавляет студента в прошедшее занятие
func TestAdminCanBookPastLesson_Integration(t *testing.T) {
	ctx := context.Background()
	studentID := uuid.New()
	lessonID := uuid.New()

	// Создаем прошедшее занятие (вчера)
	pastLesson := &models.Lesson{
		ID:              lessonID,
		TeacherID:       uuid.New(),
		StartTime:       time.Now().Add(-25 * time.Hour),
		EndTime:         time.Now().Add(-24 * time.Hour),
		MaxStudents:     5,
		CurrentStudents: 2,
	}

	// Проверяем, что это действительно прошедшее занятие
	if !pastLesson.IsInPast() {
		t.Fatal("Test setup error: lesson should be in the past")
	}

	// Mock репозитории
	lessonRepo := &mockLessonRepo{
		lesson: pastLesson,
		err:    nil,
	}

	bookingRepo := &mockBookingRepo{
		hasConflict: false, // Нет конфликтов
		err:         nil,
	}

	validator := &BookingValidator{
		lessonRepo:  lessonRepo,
		bookingRepo: bookingRepo,
	}

	t.Run("Student cannot book past lesson", func(t *testing.T) {
		err := validator.ValidateBooking(ctx, studentID, lessonID, false)

		if err == nil {
			t.Error("Expected ErrLessonInPast, got nil")
		}

		if err != ErrLessonInPast {
			t.Errorf("Expected ErrLessonInPast, got %v", err)
		}
	})

	t.Run("Admin CAN book past lesson", func(t *testing.T) {
		err := validator.ValidateBooking(ctx, studentID, lessonID, true)

		if err != nil {
			t.Errorf("Admin should be able to book past lesson, got error: %v", err)
		}
	})
}

// TestAdminCanCancelWithin24Hours_Integration проверяет отмену бронирования
func TestAdminCanCancelWithin24Hours_Integration(t *testing.T) {
	ctx := context.Background()
	lessonID := uuid.New()

	// Создаем занятие через 12 часов (меньше 24 часов)
	lesson := &models.Lesson{
		ID:        lessonID,
		StartTime: time.Now().Add(12 * time.Hour),
		EndTime:   time.Now().Add(13 * time.Hour),
	}

	// Проверяем, что занятие находится в пределах 24 часов
	if !lesson.IsWithin24Hours() {
		t.Fatal("Test setup error: lesson should be within 24 hours")
	}

	booking := &models.Booking{
		ID:       uuid.New(),
		LessonID: lessonID,
		Status:   models.BookingStatusActive,
	}

	lessonRepo := &mockLessonRepo{
		lesson: lesson,
		err:    nil,
	}

	validator := &BookingValidator{
		lessonRepo:  lessonRepo,
		bookingRepo: nil, // Не нужен для cancellation
	}

	t.Run("Student cannot cancel within 24 hours", func(t *testing.T) {
		err := validator.ValidateCancellation(ctx, booking, false)

		if err == nil {
			t.Error("Expected ErrCannotCancelWithin24Hours, got nil")
		}

		if err != ErrCannotCancelWithin24Hours {
			t.Errorf("Expected ErrCannotCancelWithin24Hours, got %v", err)
		}
	})

	t.Run("Admin CAN cancel within 24 hours", func(t *testing.T) {
		err := validator.ValidateCancellation(ctx, booking, true)

		if err != nil {
			t.Errorf("Admin should be able to cancel within 24 hours, got error: %v", err)
		}
	})
}

// TestAdminRoleValidation проверяет, что isAdmin флаг правильно передается
func TestAdminRoleValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		isAdmin   bool
		lessonAge time.Duration // Отрицательное значение = прошлое
		wantErr   bool
	}{
		{
			name:      "admin booking past lesson (yesterday)",
			isAdmin:   true,
			lessonAge: -24 * time.Hour,
			wantErr:   false,
		},
		{
			name:      "student booking past lesson (yesterday)",
			isAdmin:   false,
			lessonAge: -24 * time.Hour,
			wantErr:   true,
		},
		{
			name:      "admin booking past lesson (1 week ago)",
			isAdmin:   true,
			lessonAge: -7 * 24 * time.Hour,
			wantErr:   false,
		},
		{
			name:      "student booking past lesson (1 week ago)",
			isAdmin:   false,
			lessonAge: -7 * 24 * time.Hour,
			wantErr:   true,
		},
		{
			name:      "admin booking future lesson (tomorrow)",
			isAdmin:   true,
			lessonAge: 24 * time.Hour,
			wantErr:   false,
		},
		{
			name:      "student booking future lesson (tomorrow)",
			isAdmin:   false,
			lessonAge: 24 * time.Hour,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lesson := &models.Lesson{
				ID:        uuid.New(),
				StartTime: time.Now().Add(tt.lessonAge),
				EndTime:   time.Now().Add(tt.lessonAge + 1*time.Hour),
			}

			lessonRepo := &mockLessonRepo{
				lesson: lesson,
			}

			bookingRepo := &mockBookingRepo{
				hasConflict: false,
			}

			validator := &BookingValidator{
				lessonRepo:  lessonRepo,
				bookingRepo: bookingRepo,
			}

			err := validator.ValidateBooking(ctx, uuid.New(), uuid.New(), tt.isAdmin)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.wantErr && err != nil && err != ErrLessonInPast {
				t.Errorf("Expected ErrLessonInPast, got %v", err)
			}
		})
	}
}

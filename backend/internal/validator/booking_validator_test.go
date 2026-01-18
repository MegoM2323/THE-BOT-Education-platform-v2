package validator

import (
	"context"
	"errors"
	"testing"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
)

// mockLessonRepo для тестирования
type mockLessonRepo struct {
	lesson *models.Lesson
	err    error
}

func (m *mockLessonRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.lesson, nil
}

// mockBookingRepo для тестирования
type mockBookingRepo struct {
	hasConflict bool
	err         error
}

func (m *mockBookingRepo) HasScheduleConflict(ctx context.Context, studentID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.hasConflict, nil
}

func TestBookingValidator_ValidateBooking(t *testing.T) {
	ctx := context.Background()
	studentID := uuid.New()
	lessonID := uuid.New()

	tests := []struct {
		name        string
		lesson      *models.Lesson
		isAdmin     bool
		hasConflict bool
		lessonErr   error
		conflictErr error
		wantErr     error
	}{
		{
			name: "student cannot book past lesson",
			lesson: &models.Lesson{
				ID:        lessonID,
				StartTime: time.Now().Add(-2 * time.Hour),
				EndTime:   time.Now().Add(-1 * time.Hour),
			},
			isAdmin: false,
			wantErr: ErrLessonInPast,
		},
		{
			name: "admin can book past lesson",
			lesson: &models.Lesson{
				ID:        lessonID,
				StartTime: time.Now().Add(-2 * time.Hour),
				EndTime:   time.Now().Add(-1 * time.Hour),
			},
			isAdmin: true,
			wantErr: nil, // Должно пройти успешно
		},
		{
			name: "student can book future lesson",
			lesson: &models.Lesson{
				ID:        lessonID,
				StartTime: time.Now().Add(1 * time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			isAdmin: false,
			wantErr: nil,
		},
		{
			name: "admin can book future lesson",
			lesson: &models.Lesson{
				ID:        lessonID,
				StartTime: time.Now().Add(1 * time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			isAdmin: true,
			wantErr: nil,
		},
		{
			name: "schedule conflict detected",
			lesson: &models.Lesson{
				ID:        lessonID,
				StartTime: time.Now().Add(1 * time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			isAdmin:     false,
			hasConflict: true,
			wantErr:     ErrScheduleConflict,
		},
		{
			name: "admin booking also checks conflicts",
			lesson: &models.Lesson{
				ID:        lessonID,
				StartTime: time.Now().Add(1 * time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			isAdmin:     true,
			hasConflict: true,
			wantErr:     ErrScheduleConflict,
		},
		{
			name:      "lesson not found",
			lesson:    nil,
			isAdmin:   false,
			lessonErr: repository.ErrLessonNotFound,
			wantErr:   repository.ErrLessonNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			lessonRepo := &mockLessonRepo{
				lesson: tt.lesson,
				err:    tt.lessonErr,
			}
			bookingRepo := &mockBookingRepo{
				hasConflict: tt.hasConflict,
				err:         tt.conflictErr,
			}

			// Create validator with mocks (теперь работает благодаря интерфейсам)
			validator := &BookingValidator{
				lessonRepo:  lessonRepo,
				bookingRepo: bookingRepo,
			}

			// Execute
			err := validator.ValidateBooking(ctx, studentID, lessonID, tt.isAdmin)

			// Verify
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					// Check if error is wrapped
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("expected error %v, got %v", tt.wantErr, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestBookingValidator_ValidateCancellation(t *testing.T) {
	ctx := context.Background()
	lessonID := uuid.New()

	tests := []struct {
		name    string
		booking *models.Booking
		lesson  *models.Lesson
		isAdmin bool
		wantErr error
	}{
		{
			name: "student cannot cancel within 24 hours",
			booking: &models.Booking{
				ID:       uuid.New(),
				LessonID: lessonID,
				Status:   models.BookingStatusActive,
			},
			lesson: &models.Lesson{
				ID:        lessonID,
				StartTime: time.Now().Add(12 * time.Hour), // Через 12 часов
				EndTime:   time.Now().Add(13 * time.Hour),
			},
			isAdmin: false,
			wantErr: ErrCannotCancelWithin24Hours,
		},
		{
			name: "admin can cancel within 24 hours",
			booking: &models.Booking{
				ID:       uuid.New(),
				LessonID: lessonID,
				Status:   models.BookingStatusActive,
			},
			lesson: &models.Lesson{
				ID:        lessonID,
				StartTime: time.Now().Add(12 * time.Hour),
				EndTime:   time.Now().Add(13 * time.Hour),
			},
			isAdmin: true,
			wantErr: nil, // Админ может отменить
		},
		{
			name: "cannot cancel inactive booking",
			booking: &models.Booking{
				ID:       uuid.New(),
				LessonID: lessonID,
				Status:   models.BookingStatusCancelled,
			},
			lesson: &models.Lesson{
				ID:        lessonID,
				StartTime: time.Now().Add(48 * time.Hour),
				EndTime:   time.Now().Add(49 * time.Hour),
			},
			isAdmin: false,
			wantErr: ErrBookingNotActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lessonRepo := &mockLessonRepo{
				lesson: tt.lesson,
			}

			validator := &BookingValidator{
				lessonRepo:  lessonRepo,
				bookingRepo: nil, // Не нужен для cancellation тестов
			}

			err := validator.ValidateCancellation(ctx, tt.booking, tt.isAdmin)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("expected error %v, got %v", tt.wantErr, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

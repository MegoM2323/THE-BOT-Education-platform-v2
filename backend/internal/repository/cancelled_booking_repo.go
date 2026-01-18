package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

// CancelledBookingRepository управляет операциями с базой данных для отменённых бронирований
type CancelledBookingRepository struct {
	db *sqlx.DB
}

// NewCancelledBookingRepository создает новый CancelledBookingRepository
func NewCancelledBookingRepository(db *sqlx.DB) *CancelledBookingRepository {
	return &CancelledBookingRepository{db: db}
}

// CreateCancelledBooking создает запись об отменённом бронировании
// Метод идемпотентен - если запись уже существует, ошибка не возвращается
func (r *CancelledBookingRepository) CreateCancelledBooking(ctx context.Context, cancelled *models.CancelledBooking) error {
	// Валидация
	if err := cancelled.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO cancelled_bookings (id, booking_id, student_id, lesson_id, cancelled_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (student_id, lesson_id) DO NOTHING
	`

	cancelled.ID = uuid.New()
	cancelled.CancelledAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		cancelled.ID,
		cancelled.BookingID,
		cancelled.StudentID,
		cancelled.LessonID,
		cancelled.CancelledAt,
	)
	if err != nil {
		// Проверяем специфичные ошибки PostgreSQL
		if pgErr, ok := err.(*pgconn.PgError); ok {
			// 23505 - unique_violation (на случай если ON CONFLICT не сработает)
			if pgErr.Code == "23505" {
				// Идемпотентность: если запись уже есть - не ошибка
				return nil
			}
		}
		return fmt.Errorf("failed to create cancelled booking: %w", err)
	}

	return nil
}

// HasCancelledBooking проверяет, отменял ли студент бронирование на этот урок
func (r *CancelledBookingRepository) HasCancelledBooking(ctx context.Context, studentID, lessonID uuid.UUID) (bool, error) {
	// Валидация входных данных
	if studentID == uuid.Nil {
		return false, models.ErrInvalidStudentID
	}
	if lessonID == uuid.Nil {
		return false, models.ErrInvalidLessonID
	}

	query := `
		SELECT EXISTS(
			SELECT 1
			FROM cancelled_bookings
			WHERE student_id = $1 AND lesson_id = $2
		)
	`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, studentID, lessonID)
	if err != nil {
		return false, fmt.Errorf("failed to check cancelled booking: %w", err)
	}

	return exists, nil
}

// ListCancelledByStudent возвращает все отменённые бронирования студента
func (r *CancelledBookingRepository) ListCancelledByStudent(ctx context.Context, studentID uuid.UUID) ([]*models.CancelledBooking, error) {
	// Валидация входных данных
	if studentID == uuid.Nil {
		return nil, models.ErrInvalidStudentID
	}

	query := `
		SELECT id, booking_id, student_id, lesson_id, cancelled_at
		FROM cancelled_bookings
		WHERE student_id = $1
		ORDER BY cancelled_at DESC
	`

	var cancelled []*models.CancelledBooking
	err := r.db.SelectContext(ctx, &cancelled, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cancelled bookings: %w", err)
	}

	return cancelled, nil
}

// DeleteCancelledBooking удаляет запись об отменённом бронировании (для админа)
func (r *CancelledBookingRepository) DeleteCancelledBooking(ctx context.Context, cancelledID uuid.UUID) error {
	// Валидация входных данных
	if cancelledID == uuid.Nil {
		return models.ErrInvalidBookingID
	}

	query := `
		DELETE FROM cancelled_bookings
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, cancelledID)
	if err != nil {
		return fmt.Errorf("failed to delete cancelled booking: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrCancelledNotFound
	}

	return nil
}

// DeleteByStudentAndLesson удаляет запись об отменённом бронировании по student_id и lesson_id (для админа)
func (r *CancelledBookingRepository) DeleteByStudentAndLesson(ctx context.Context, studentID, lessonID uuid.UUID) error {
	if studentID == uuid.Nil {
		return models.ErrInvalidStudentID
	}
	if lessonID == uuid.Nil {
		return models.ErrInvalidLessonID
	}

	query := `
		DELETE FROM cancelled_bookings
		WHERE student_id = $1 AND lesson_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, studentID, lessonID)
	if err != nil {
		return fmt.Errorf("failed to delete cancelled booking: %w", err)
	}

	return nil
}

// GetByID получает отменённое бронирование по ID (для админа)
func (r *CancelledBookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CancelledBooking, error) {
	// Валидация входных данных
	if id == uuid.Nil {
		return nil, models.ErrInvalidBookingID
	}

	query := `
		SELECT id, booking_id, student_id, lesson_id, cancelled_at
		FROM cancelled_bookings
		WHERE id = $1
	`

	var cancelled models.CancelledBooking
	err := r.db.GetContext(ctx, &cancelled, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCancelledNotFound
		}
		return nil, fmt.Errorf("failed to get cancelled booking by ID: %w", err)
	}

	return &cancelled, nil
}

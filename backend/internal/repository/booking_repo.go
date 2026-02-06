package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// BookingRepository управляет операциями с базой данных для бронирований
type BookingRepository struct {
	db *sqlx.DB
}

// NewBookingRepository создает новый BookingRepository
func NewBookingRepository(db *sqlx.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Create создает новое бронирование в рамках транзакции
func (r *BookingRepository) Create(ctx context.Context, tx pgx.Tx, booking *models.Booking) error {
	query := `
		INSERT INTO bookings (id, student_id, lesson_id, status, booked_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	booking.ID = uuid.New()
	booking.Status = models.BookingStatusActive
	booking.BookedAt = time.Now()
	booking.CreatedAt = time.Now()
	booking.UpdatedAt = time.Now()

	_, err := tx.Exec(ctx, query,
		booking.ID,
		booking.StudentID,
		booking.LessonID,
		booking.Status,
		booking.BookedAt,
		booking.CreatedAt,
		booking.UpdatedAt,
	)
	if err != nil {
		// Проверяем, является ли ошибка нарушением UNIQUE constraint
		// Это происходит, когда студент пытается забронировать уже забронированное занятие
		if IsUniqueViolationError(err) {
			return ErrDuplicateBooking
		}
		return fmt.Errorf("failed to create booking: %w", err)
	}

	return nil
}

// GetByID получает бронирование по ID
func (r *BookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Booking, error) {
	query := `
		SELECT ` + BookingSelectFields + `
		FROM bookings
		WHERE id = $1
	`

	var booking models.Booking
	err := r.db.GetContext(ctx, &booking, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to get booking by ID: %w", err)
	}

	return &booking, nil
}

// GetByIDForUpdate получает бронирование по ID с блокировкой строки для обновления
func (r *BookingRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*models.Booking, error) {
	query := `
		SELECT ` + BookingSelectFields + `
		FROM bookings
		WHERE id = $1
		FOR UPDATE
	`

	var booking models.Booking
	err := tx.QueryRow(ctx, query, id).Scan(
		&booking.ID,
		&booking.StudentID,
		&booking.LessonID,
		&booking.Status,
		&booking.BookedAt,
		&booking.CancelledAt,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to get booking by ID for update: %w", err)
	}

	return &booking, nil
}

// GetWithDetails получает бронирование с деталями занятия и преподавателя
// Фильтрует soft-deleted lessons для предотвращения exposure удаленных данных
func (r *BookingRepository) GetWithDetails(ctx context.Context, id uuid.UUID) (*models.BookingWithDetails, error) {
	query := `
		SELECT
			b.id, b.student_id, b.lesson_id, b.status, b.booked_at, b.cancelled_at, b.created_at, b.updated_at,
			l.start_time, l.end_time, l.teacher_id, l.subject, l.homework_text,
			COALESCE(NULLIF(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''), u.email) as teacher_name
		FROM bookings b
		JOIN lessons l ON b.lesson_id = l.id AND l.deleted_at IS NULL
		JOIN users u ON l.teacher_id = u.id AND u.deleted_at IS NULL
		WHERE b.id = $1
	`

	var booking models.BookingWithDetails
	err := r.db.GetContext(ctx, &booking, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to get booking with details: %w", err)
	}

	return &booking, nil
}

// List получает бронирования с опциональными фильтрами
func (r *BookingRepository) List(ctx context.Context, filter *models.ListBookingsFilter) ([]*models.BookingWithDetails, error) {
	query := `
		SELECT
			b.id, b.student_id, b.lesson_id, b.status, b.booked_at, b.cancelled_at, b.created_at, b.updated_at,
			l.start_time, l.end_time, l.teacher_id, l.subject, l.homework_text,
			COALESCE(NULLIF(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''), u.email) as teacher_name,
			COALESCE(NULLIF(TRIM(CONCAT(s.first_name, ' ', s.last_name)), ''), s.email) as student_full_name,
			s.email as student_email,
			b.created_at as booking_created_at
		FROM bookings b
		JOIN lessons l ON b.lesson_id = l.id AND l.deleted_at IS NULL
		JOIN users u ON l.teacher_id = u.id
		JOIN users s ON b.student_id = s.id
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if filter != nil {
		if filter.StudentID != nil {
			query += fmt.Sprintf(` AND b.student_id = $%d`, argIndex)
			args = append(args, *filter.StudentID)
			argIndex++
		}
		if filter.LessonID != nil {
			query += fmt.Sprintf(` AND b.lesson_id = $%d`, argIndex)
			args = append(args, *filter.LessonID)
			argIndex++
		}
		if filter.Status != nil {
			query += fmt.Sprintf(` AND b.status = $%d`, argIndex)
			args = append(args, *filter.Status)
			argIndex++
		}
		if filter.StartDate != nil {
			query += fmt.Sprintf(` AND l.start_time >= $%d`, argIndex)
			args = append(args, *filter.StartDate)
			argIndex++
		}
		if filter.EndDate != nil {
			query += fmt.Sprintf(` AND l.start_time <= $%d`, argIndex)
			args = append(args, *filter.EndDate)
			argIndex++
		}
	}

	query += ` ORDER BY l.start_time DESC`

	var bookings []*models.BookingWithDetails
	err := r.db.SelectContext(ctx, &bookings, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings: %w", err)
	}

	return bookings, nil
}

// ListWithPagination возвращает список бронирований с пагинацией
func (r *BookingRepository) ListWithPagination(ctx context.Context, filter *models.ListBookingsFilter, offset, limit int) ([]*models.BookingWithDetails, int, error) {
	// Построение базового WHERE выражения для подсчета
	whereClause := `WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if filter != nil {
		if filter.StudentID != nil {
			whereClause += fmt.Sprintf(` AND b.student_id = $%d`, argIndex)
			args = append(args, *filter.StudentID)
			argIndex++
		}
		if filter.LessonID != nil {
			whereClause += fmt.Sprintf(` AND b.lesson_id = $%d`, argIndex)
			args = append(args, *filter.LessonID)
			argIndex++
		}
		if filter.Status != nil {
			whereClause += fmt.Sprintf(` AND b.status = $%d`, argIndex)
			args = append(args, *filter.Status)
			argIndex++
		}
		if filter.StartDate != nil {
			whereClause += fmt.Sprintf(` AND l.start_time >= $%d`, argIndex)
			args = append(args, *filter.StartDate)
			argIndex++
		}
		if filter.EndDate != nil {
			whereClause += fmt.Sprintf(` AND l.start_time <= $%d`, argIndex)
			args = append(args, *filter.EndDate)
			argIndex++
		}
	}

	// Подсчет общего количества бронирований
	countQuery := `
		SELECT COUNT(*)
		FROM bookings b
		JOIN lessons l ON b.lesson_id = l.id AND l.deleted_at IS NULL
		JOIN users u ON l.teacher_id = u.id
		JOIN users s ON b.student_id = s.id
		` + whereClause

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count bookings: %w", err)
	}

	// Получение списка бронирований с LIMIT/OFFSET
	dataQuery := `
		SELECT
			b.id, b.student_id, b.lesson_id, b.status, b.booked_at, b.cancelled_at, b.created_at, b.updated_at,
			l.start_time, l.end_time, l.teacher_id, l.subject, l.homework_text,
			COALESCE(NULLIF(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''), u.email) as teacher_name,
			COALESCE(NULLIF(TRIM(CONCAT(s.first_name, ' ', s.last_name)), ''), s.email) as student_full_name,
			s.email as student_email,
			b.created_at as booking_created_at
		FROM bookings b
		JOIN lessons l ON b.lesson_id = l.id AND l.deleted_at IS NULL
		JOIN users u ON l.teacher_id = u.id
		JOIN users s ON b.student_id = s.id
		` + whereClause + ` ORDER BY l.start_time DESC LIMIT $` + fmt.Sprintf("%d", argIndex) + ` OFFSET $` + fmt.Sprintf("%d", argIndex+1)

	args = append(args, limit, offset)

	var bookings []*models.BookingWithDetails
	if err := r.db.SelectContext(ctx, &bookings, dataQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list bookings: %w", err)
	}

	if bookings == nil {
		bookings = []*models.BookingWithDetails{}
	}

	return bookings, total, nil
}

// Cancel отменяет бронирование в рамках транзакции
func (r *BookingRepository) Cancel(ctx context.Context, tx pgx.Tx, bookingID uuid.UUID) error {
	query := `
		UPDATE bookings
		SET status = $1, cancelled_at = $2, updated_at = $3
		WHERE id = $4 AND status = $5
	`

	result, err := tx.Exec(ctx, query,
		models.BookingStatusCancelled,
		time.Now(),
		time.Now(),
		bookingID,
		models.BookingStatusActive,
	)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrBookingNotActive
	}

	return nil
}

// HasScheduleConflict проверяет, есть ли у студента бронирование, конфликтующее с указанным временным диапазоном
func (r *BookingRepository) HasScheduleConflict(ctx context.Context, studentID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM bookings b
			JOIN lessons l ON b.lesson_id = l.id
			WHERE b.student_id = $1
			  AND b.status = $2
			  AND l.deleted_at IS NULL
			  AND (
				(l.start_time < $4 AND l.end_time > $3)
			  )
		)
	`

	var hasConflict bool
	err := r.db.GetContext(ctx, &hasConflict, query, studentID, models.BookingStatusActive, startTime, endTime)
	if err != nil {
		return false, fmt.Errorf("failed to check schedule conflict: %w", err)
	}

	return hasConflict, nil
}

// HasScheduleConflictInTx проверяет конфликты расписания внутри транзакции
// Это критично при template replacement чтобы видеть отменённые bookings в рамках транзакции
func (r *BookingRepository) HasScheduleConflictInTx(ctx context.Context, tx *sqlx.Tx, studentID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM bookings b
			JOIN lessons l ON b.lesson_id = l.id
			WHERE b.student_id = $1
			  AND b.status = $2
			  AND l.deleted_at IS NULL
			  AND (
				(l.start_time < $4 AND l.end_time > $3)
			  )
		)
	`

	var hasConflict bool
	err := tx.GetContext(ctx, &hasConflict, query, studentID, models.BookingStatusActive, startTime, endTime)
	if err != nil {
		return false, fmt.Errorf("failed to check schedule conflict in transaction: %w", err)
	}

	return hasConflict, nil
}

// HasScheduleConflictExcluding проверяет конфликты, исключая конкретное бронирование
func (r *BookingRepository) HasScheduleConflictExcluding(ctx context.Context, studentID uuid.UUID, excludeBookingID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM bookings b
			JOIN lessons l ON b.lesson_id = l.id
			WHERE b.student_id = $1
			  AND b.id != $2
			  AND b.status = $3
			  AND l.deleted_at IS NULL
			  AND (
				(l.start_time < $5 AND l.end_time > $4)
			  )
		)
	`

	var hasConflict bool
	err := r.db.GetContext(ctx, &hasConflict, query, studentID, excludeBookingID, models.BookingStatusActive, startTime, endTime)
	if err != nil {
		return false, fmt.Errorf("failed to check schedule conflict excluding: %w", err)
	}

	return hasConflict, nil
}

// GetActiveBookingByStudentAndLesson получает активное бронирование для студента и занятия
// Проверяет что lesson существует и не soft-deleted для обеспечения консистентности данных
func (r *BookingRepository) GetActiveBookingByStudentAndLesson(ctx context.Context, studentID, lessonID uuid.UUID) (*models.Booking, error) {
	query := `
		SELECT ` + BookingSelectFields + `
		FROM bookings b
		WHERE b.student_id = $1 AND b.lesson_id = $2 AND b.status = $3
		AND EXISTS(SELECT 1 FROM lessons WHERE id = $2 AND deleted_at IS NULL)
	`

	var booking models.Booking
	err := r.db.GetContext(ctx, &booking, query, studentID, lessonID, models.BookingStatusActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to get active booking: %w", err)
	}

	return &booking, nil
}

// GetBookingsWithLessons получает все активные бронирования студента с информацией о занятиях и учителях в одном запросе.
// Эта функция исправляет N+1 проблему: вместо N запросов для каждого занятия, делается 1 JOIN запрос.
// Возвращает только активные бронирования, упорядоченные по времени начала занятия (новые сначала).
// DISTINCT используется для исключения дубликатов если есть несколько бронирований на одно занятие.
func (r *BookingRepository) GetBookingsWithLessons(ctx context.Context, studentID uuid.UUID) ([]*models.LessonWithTeacher, error) {
	query := `
		SELECT DISTINCT
			l.id, l.teacher_id, l.start_time, l.end_time,
			l.max_students, l.current_students, l.credits_cost, l.color, l.subject, l.homework_text,
			l.created_at, l.updated_at, l.deleted_at,
			COALESCE(NULLIF(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''), u.email) as teacher_name
		FROM bookings b
		INNER JOIN lessons l ON b.lesson_id = l.id
		LEFT JOIN users u ON l.teacher_id = u.id
		WHERE b.student_id = $1 AND b.status = $2 AND l.deleted_at IS NULL
		ORDER BY l.start_time DESC
	`

	log.Debug().Str("student_id", studentID.String()).Msg("Executing GetBookingsWithLessons query")

	lessons := make([]*models.LessonWithTeacher, 0)
	err := r.db.SelectContext(ctx, &lessons, query, studentID, models.BookingStatusActive)
	if err != nil {
		log.Error().
			Str("student_id", studentID.String()).
			Err(err).
			Msg("Query failed in GetBookingsWithLessons")
		return nil, fmt.Errorf("failed to get bookings with lessons: %w", err)
	}

	if len(lessons) == 0 {
		log.Debug().Str("student_id", studentID.String()).Msg("No active lessons found for student")
		return lessons, nil
	}

	log.Debug().
		Str("student_id", studentID.String()).
		Int("lessons_count", len(lessons)).
		Msg("Retrieved lessons for student")
	return lessons, nil
}

// ReactivateBooking реактивирует отменённое бронирование (меняет status с cancelled на active)
func (r *BookingRepository) ReactivateBooking(ctx context.Context, tx pgx.Tx, studentID, lessonID uuid.UUID) (*models.Booking, error) {
	query := `
		UPDATE bookings
		SET status = $1, cancelled_at = NULL, updated_at = $2
		WHERE student_id = $3 AND lesson_id = $4 AND status = $5
		RETURNING id, student_id, lesson_id, status, booked_at, cancelled_at, created_at, updated_at
	`
	var booking models.Booking
	err := tx.QueryRow(ctx, query,
		models.BookingStatusActive,
		time.Now(),
		studentID,
		lessonID,
		models.BookingStatusCancelled,
	).Scan(&booking.ID, &booking.StudentID, &booking.LessonID, &booking.Status, &booking.BookedAt, &booking.CancelledAt, &booking.CreatedAt, &booking.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to reactivate booking: %w", err)
	}
	return &booking, nil
}

// GetBookingsForWeekTx получает bookings для lessons на конкретной неделе
// Поддерживает фильтрацию по статусу: 'active', 'cancelled', или " для всех
func (r *BookingRepository) GetBookingsForWeekTx(
	ctx context.Context,
	tx *sqlx.Tx,
	weekDate time.Time,
	statusFilter string, // 'active', 'cancelled', или '' для всех
) ([]*models.Booking, error) {
	// Нормализуем weekDate к началу дня в UTC
	weekStart := time.Date(weekDate.Year(), weekDate.Month(), weekDate.Day(), 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	query := `
		SELECT
			b.id, b.student_id, b.lesson_id, b.status,
			b.booked_at, b.cancelled_at, b.created_at, b.updated_at
		FROM bookings b
		JOIN lessons l ON b.lesson_id = l.id
		WHERE l.start_time >= $1
		  AND l.start_time < $2
		  AND l.deleted_at IS NULL
	`

	var args []interface{}
	args = append(args, weekStart, weekEnd)

	if statusFilter != "" {
		query += ` AND b.status = $3`
		args = append(args, statusFilter)
	}

	query += ` ORDER BY b.student_id, l.start_time`

	var bookings []*models.Booking
	err := tx.SelectContext(ctx, &bookings, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query bookings for week: %w", err)
	}

	return bookings, nil
}

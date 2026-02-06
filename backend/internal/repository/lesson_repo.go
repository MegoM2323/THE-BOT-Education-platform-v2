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
)

// LessonRepository управляет операциями с базой данных для занятий
type LessonRepository struct {
	db *sqlx.DB
}

// NewLessonRepository создает новый LessonRepository
func NewLessonRepository(db *sqlx.DB) *LessonRepository {
	return &LessonRepository{db: db}
}

// Create создает новое занятие
func (r *LessonRepository) Create(ctx context.Context, lesson *models.Lesson) error {
	query := `
		INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, credits_cost, color, subject, homework_text, report_text, link, is_recurring, recurring_group_id, recurring_end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	if lesson.ID == uuid.Nil {
		lesson.ID = uuid.New()
	}
	lesson.CurrentStudents = 0
	lesson.CreatedAt = time.Now()
	lesson.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		lesson.ID,
		lesson.TeacherID,
		lesson.StartTime,
		lesson.EndTime,
		lesson.MaxStudents,
		lesson.CurrentStudents,
		lesson.CreditsCost,
		lesson.Color,
		lesson.Subject,
		lesson.HomeworkText,
		lesson.ReportText,
		lesson.Link,
		lesson.IsRecurring,
		lesson.RecurringGroupID,
		lesson.RecurringEndDate,
		lesson.CreatedAt,
		lesson.UpdatedAt,
	)
	if err != nil {
		// Проверяем, является ли ошибка нарушением EXCLUDE constraint (конфликт расписания)
		if IsExclusionViolationError(err) {
			return ErrLessonOverlapConflict
		}
		return fmt.Errorf("failed to create lesson: %w", err)
	}

	return nil
}

// CreateBatchLessons создает несколько занятий в одной транзакции
func (r *LessonRepository) CreateBatchLessons(ctx context.Context, lessons []*models.Lesson) error {
	if len(lessons) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, lesson := range lessons {
		if err := r.createInTx(ctx, tx, lesson); err != nil {
			return fmt.Errorf("failed to create lesson in batch: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch lessons: %w", err)
	}

	return nil
}

// createInTx создает занятие внутри существующей транзакции sqlx
func (r *LessonRepository) createInTx(ctx context.Context, tx *sqlx.Tx, lesson *models.Lesson) error {
	query := `
		INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, credits_cost, color, subject, homework_text, report_text, link, is_recurring, recurring_group_id, recurring_end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	if lesson.ID == uuid.Nil {
		lesson.ID = uuid.New()
	}
	lesson.CurrentStudents = 0
	lesson.CreatedAt = time.Now()
	lesson.UpdatedAt = time.Now()

	_, err := tx.ExecContext(ctx, query,
		lesson.ID,
		lesson.TeacherID,
		lesson.StartTime,
		lesson.EndTime,
		lesson.MaxStudents,
		lesson.CurrentStudents,
		lesson.CreditsCost,
		lesson.Color,
		lesson.Subject,
		lesson.HomeworkText,
		lesson.ReportText,
		lesson.Link,
		lesson.IsRecurring,
		lesson.RecurringGroupID,
		lesson.RecurringEndDate,
		lesson.CreatedAt,
		lesson.UpdatedAt,
	)
	if err != nil {
		if IsExclusionViolationError(err) {
			return ErrLessonOverlapConflict
		}
		return fmt.Errorf("failed to create lesson in transaction: %w", err)
	}

	return nil
}

// GetByID получает занятие по ID
func (r *LessonRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	query := `
		SELECT ` + LessonSelectFields + `
		FROM lessons
		WHERE id = $1 AND deleted_at IS NULL
	`

	var lesson models.Lesson
	err := r.db.GetContext(ctx, &lesson, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLessonNotFound
		}
		return nil, fmt.Errorf("failed to get lesson by ID: %w", err)
	}

	return &lesson, nil
}

// GetByIDForUpdate получает занятие по ID с блокировкой строки для обновления (SELECT FOR UPDATE)
func (r *LessonRepository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*models.Lesson, error) {
	query := `
		SELECT ` + LessonSelectFields + `
		FROM lessons
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`

	var lesson models.Lesson
	err := tx.QueryRow(ctx, query, id).Scan(
		&lesson.ID,
		&lesson.TeacherID,
		&lesson.StartTime,
		&lesson.EndTime,
		&lesson.MaxStudents,
		&lesson.CurrentStudents,
		&lesson.CreditsCost,
		&lesson.Color,
		&lesson.Subject,
		&lesson.HomeworkText,
		&lesson.ReportText,
		&lesson.Link,
		&lesson.IsRecurring,
		&lesson.RecurringGroupID,
		&lesson.RecurringEndDate,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
		&lesson.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrLessonNotFound
		}
		return nil, fmt.Errorf("failed to get lesson by ID for update: %w", err)
	}

	return &lesson, nil
}

// GetWithTeacher получает занятие с информацией о преподавателе
func (r *LessonRepository) GetWithTeacher(ctx context.Context, id uuid.UUID) (*models.LessonWithTeacher, error) {
	query := `
		SELECT
			l.id, l.teacher_id, l.start_time, l.end_time,
			l.max_students, l.current_students, l.credits_cost, l.color, l.subject, l.homework_text, l.report_text, l.link,
			l.created_at, l.updated_at, l.deleted_at,
			CONCAT(u.first_name, ' ', u.last_name) as teacher_name
		FROM lessons l
		JOIN users u ON l.teacher_id = u.id
		WHERE l.id = $1 AND l.deleted_at IS NULL
	`

	var lesson models.LessonWithTeacher
	err := r.db.GetContext(ctx, &lesson, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLessonNotFound
		}
		return nil, fmt.Errorf("failed to get lesson with teacher: %w", err)
	}

	return &lesson, nil
}

// List получает занятия с опциональными фильтрами
func (r *LessonRepository) List(ctx context.Context, filter *models.ListLessonsFilter) ([]*models.LessonWithTeacher, error) {
	query := `
		SELECT DISTINCT
			l.id, l.teacher_id, l.start_time, l.end_time,
			l.max_students, l.current_students, l.credits_cost, l.color, l.subject, l.homework_text, l.link,
			l.created_at, l.updated_at, l.deleted_at,
			CONCAT(u.first_name, ' ', u.last_name) as teacher_name
		FROM lessons l
		JOIN users u ON l.teacher_id = u.id
		WHERE l.deleted_at IS NULL
	`

	args := []interface{}{}
	argIndex := 1

	if filter != nil {
		if filter.TeacherID != nil {
			query += fmt.Sprintf(` AND l.teacher_id = $%d`, argIndex)
			args = append(args, *filter.TeacherID)
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
		if filter.Available != nil && *filter.Available {
			query += ` AND l.current_students < l.max_students`
		}
	}

	query += ` ORDER BY l.start_time ASC`

	var lessons []*models.LessonWithTeacher
	err := r.db.SelectContext(ctx, &lessons, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list lessons: %w", err)
	}

	return lessons, nil
}

// GetVisibleLessons получает занятия, видимые для конкретного пользователя с учетом роли
// Правила видимости:
// - Admin: все занятия
// - Teacher: только свои занятия (teacher_id = userID)
// - Student: все групповые занятия (max_students > 1) + индивидуальные занятия (max_students = 1), на которые они записаны
func (r *LessonRepository) GetVisibleLessons(ctx context.Context, userID uuid.UUID, userRole string, filter *models.ListLessonsFilter) ([]*models.LessonWithTeacher, error) {
	var query string
	args := []interface{}{}
	argIndex := 1

	switch userRole {
	case "admin":
		// Admin sees all lessons
		query = `
			SELECT DISTINCT
				l.id, l.teacher_id, l.start_time, l.end_time,
				l.max_students, l.current_students, l.credits_cost, l.color, l.subject, l.homework_text, l.report_text, l.link,
				l.is_recurring, l.recurring_group_id, l.recurring_end_date,
				l.created_at, l.updated_at, l.deleted_at,
				CONCAT(u.first_name, ' ', u.last_name) as teacher_name
			FROM lessons l
			JOIN users u ON l.teacher_id = u.id
			WHERE l.deleted_at IS NULL
		`

	case "teacher":
		// Teachers see only their own lessons
		query = `
			SELECT DISTINCT
				l.id, l.teacher_id, l.start_time, l.end_time,
				l.max_students, l.current_students, l.credits_cost, l.color, l.subject, l.homework_text, l.report_text, l.link,
				l.is_recurring, l.recurring_group_id, l.recurring_end_date,
				l.created_at, l.updated_at, l.deleted_at,
				CONCAT(u.first_name, ' ', u.last_name) as teacher_name
			FROM lessons l
			JOIN users u ON l.teacher_id = u.id
			WHERE l.deleted_at IS NULL AND l.teacher_id = $1
		`
		args = append(args, userID)
		argIndex++

	case "student":
		// Students see:
		// 1. All group lessons (max_students > 1)
		// 2. Individual lessons (max_students = 1) they have booked (past, present, or future - any booking status)
		query = `
			SELECT DISTINCT
				l.id, l.teacher_id, l.start_time, l.end_time,
				l.max_students, l.current_students, l.credits_cost, l.color, l.subject, l.homework_text, l.report_text, l.link,
				l.is_recurring, l.recurring_group_id, l.recurring_end_date,
				l.created_at, l.updated_at, l.deleted_at,
				CONCAT(u.first_name, ' ', u.last_name) as teacher_name
			FROM lessons l
			JOIN users u ON l.teacher_id = u.id
			LEFT JOIN bookings b ON l.id = b.lesson_id AND b.student_id = $1
			WHERE l.deleted_at IS NULL
			  AND (
				l.max_students > 1
				OR (l.max_students = 1 AND b.id IS NOT NULL)
			  )
		`
		args = append(args, userID)
		argIndex++

	case "methodologist":
		// Methodologist sees all lessons (like admin, for template management)
		query = `
			SELECT DISTINCT
				l.id, l.teacher_id, l.start_time, l.end_time,
				l.max_students, l.current_students, l.credits_cost, l.color, l.subject, l.homework_text, l.report_text, l.link,
				l.is_recurring, l.recurring_group_id, l.recurring_end_date,
				l.created_at, l.updated_at, l.deleted_at,
				CONCAT(u.first_name, ' ', u.last_name) as teacher_name
			FROM lessons l
			JOIN users u ON l.teacher_id = u.id
			WHERE l.deleted_at IS NULL
		`

	default:
		return nil, fmt.Errorf("invalid user role: %s", userRole)
	}

	// Apply additional filters
	if filter != nil {
		if filter.TeacherID != nil {
			query += fmt.Sprintf(` AND l.teacher_id = $%d`, argIndex)
			args = append(args, *filter.TeacherID)
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
		if filter.Available != nil && *filter.Available {
			query += ` AND l.current_students < l.max_students`
		}
	}

	query += ` ORDER BY l.start_time ASC`

	var lessons []*models.LessonWithTeacher
	err := r.db.SelectContext(ctx, &lessons, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get visible lessons: %w", err)
	}

	return lessons, nil
}

// Update обновляет занятие
// Поддерживаемые поля: start_time, end_time, max_students, color, subject, homework_text
func (r *LessonRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Whitelist of allowed fields to prevent SQL injection
	allowedFields := map[string]bool{
		"teacher_id":    true,
		"start_time":    true,
		"end_time":      true,
		"max_students":  true,
		"credits_cost":  true,
		"color":         true,
		"subject":       true,
		"homework_text": true,
		"report_text":   true,
		"link":          true,
	}

	query := `UPDATE lessons SET updated_at = $1`
	args := []interface{}{time.Now()}
	argIndex := 2

	for field, value := range updates {
		// Validate field name against whitelist
		if !allowedFields[field] {
			return fmt.Errorf("invalid field for update: %s", field)
		}

		// Safely append field update with parameterized query
		query += fmt.Sprintf(`, %s = $%d`, field, argIndex)
		args = append(args, value)
		argIndex++
	}

	query += fmt.Sprintf(` WHERE id = $%d AND deleted_at IS NULL`, argIndex)
	args = append(args, id)

	// Check if db is nil (shouldn't happen in production, but needed for tests)
	if r.db == nil {
		// Return validation error only if we reached this point
		// (field validation already passed above)
		return fmt.Errorf("database connection not initialized")
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		// Проверяем, является ли ошибка нарушением EXCLUDE constraint (конфликт расписания)
		if IsExclusionViolationError(err) {
			return ErrLessonOverlapConflict
		}
		return fmt.Errorf("failed to update lesson: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrLessonNotFound
	}

	return nil
}

// IncrementStudents увеличивает счетчик current_students в рамках транзакции
func (r *LessonRepository) IncrementStudents(ctx context.Context, tx pgx.Tx, lessonID uuid.UUID) error {
	query := `
		UPDATE lessons
		SET current_students = current_students + 1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL AND current_students < max_students
	`

	result, err := tx.Exec(ctx, query, time.Now(), lessonID)
	if err != nil {
		return fmt.Errorf("failed to increment students: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrLessonFull
	}

	return nil
}

// DecrementStudents уменьшает счетчик current_students в рамках транзакции
// SAFETY GUARANTEE: This operation is guaranteed to never produce negative values.
// The WHERE clause includes 'current_students > 0' which prevents decrementing
// when counter is already at zero. If decrement would produce < 0, the UPDATE
// affects 0 rows and returns ErrLessonNotFound, preventing data corruption.
func (r *LessonRepository) DecrementStudents(ctx context.Context, tx pgx.Tx, lessonID uuid.UUID) error {
	query := `
		UPDATE lessons
		SET current_students = current_students - 1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL AND current_students > 0
	`

	result, err := tx.Exec(ctx, query, time.Now(), lessonID)
	if err != nil {
		return fmt.Errorf("failed to decrement students: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrLessonNotFound
	}

	return nil
}

// GetByIDs получает несколько занятий по ID с информацией о преподавателях в одном запросе
func (r *LessonRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.LessonWithTeacher, error) {
	if len(ids) == 0 {
		return []*models.LessonWithTeacher{}, nil
	}

	query := `
		SELECT
			l.id, l.teacher_id, l.start_time, l.end_time,
			l.max_students, l.current_students, l.credits_cost, l.color, l.subject, l.homework_text, l.report_text, l.link,
			l.is_recurring, l.recurring_group_id, l.recurring_end_date,
			l.created_at, l.updated_at, l.deleted_at,
			CONCAT(u.first_name, ' ', u.last_name) as teacher_name
		FROM lessons l
		JOIN users u ON l.teacher_id = u.id
		WHERE l.id = ANY($1) AND l.deleted_at IS NULL
		ORDER BY l.start_time ASC
	`

	var lessons []*models.LessonWithTeacher
	err := r.db.SelectContext(ctx, &lessons, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get lessons by IDs: %w", err)
	}

	return lessons, nil
}

// Delete выполняет мягкое удаление занятия с проверками целостности данных
func (r *LessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var hasActiveBookings bool
	err := r.db.GetContext(ctx, &hasActiveBookings,
		`SELECT EXISTS(SELECT 1 FROM bookings WHERE lesson_id = $1 AND status = 'active')`, id)
	if err != nil {
		return fmt.Errorf("check active bookings: %w", err)
	}
	if hasActiveBookings {
		return ErrLessonHasActiveBookings
	}

	var hasHomework bool
	err = r.db.GetContext(ctx, &hasHomework,
		`SELECT EXISTS(SELECT 1 FROM lesson_homework WHERE lesson_id = $1)`, id)
	if err != nil {
		return fmt.Errorf("check homework: %w", err)
	}
	if hasHomework {
		return ErrLessonHasHomework
	}

	query := `UPDATE lessons SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft delete lesson: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrLessonNotFound
	}

	return nil
}

// HardDelete выполняет полное удаление занятия из базы данных
func (r *LessonRepository) HardDelete(ctx context.Context, id uuid.UUID, force bool) error {
	if !force {
		return ErrHardDeleteRequiresForce
	}

	var hasActiveBookings bool
	err := r.db.GetContext(ctx, &hasActiveBookings,
		`SELECT EXISTS(SELECT 1 FROM bookings WHERE lesson_id = $1 AND status = 'active')`, id)
	if err != nil {
		return fmt.Errorf("check active bookings: %w", err)
	}
	if hasActiveBookings {
		return ErrLessonHasActiveBookings
	}

	var hasHomework bool
	err = r.db.GetContext(ctx, &hasHomework,
		`SELECT EXISTS(SELECT 1 FROM lesson_homework WHERE lesson_id = $1)`, id)
	if err != nil {
		return fmt.Errorf("check homework: %w", err)
	}
	if hasHomework {
		return ErrLessonHasHomework
	}

	query := `DELETE FROM lessons WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("hard delete lesson: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrLessonNotFound
	}

	return nil
}

// SyncStudentCounts синхронизирует счетчики студентов для всех занятий
// на основе COUNT активных бронирований. Вызывается администратором после изменения расписания
func (r *LessonRepository) SyncStudentCounts(ctx context.Context) error {
	query := `
		UPDATE lessons l
		SET current_students = (
			SELECT COUNT(*) FROM bookings b
			WHERE b.lesson_id = l.id
			  AND b.status = 'active'
			  AND b.deleted_at IS NULL
		),
		updated_at = $1
		WHERE l.deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to sync student counts: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// Логируем информацию о синхронизации
	fmt.Printf("[INFO] Synced student counts for %d lessons\n", rows)

	return nil
}

// GetLessonsByTimePattern finds all lessons matching a specific time pattern (for bulk edit)
// Searches for lessons from the same teacher, same day of week, same time, after a given date
func (r *LessonRepository) GetLessonsByTimePattern(ctx context.Context, teacherID uuid.UUID, dayOfWeek int, hour int, minute int, afterDate time.Time) ([]*models.Lesson, error) {
	query := `
		SELECT ` + LessonSelectFields + `
		FROM lessons
		WHERE teacher_id = $1
		  AND EXTRACT(DOW FROM start_time) = $2
		  AND EXTRACT(HOUR FROM start_time) = $3
		  AND EXTRACT(MINUTE FROM start_time) = $4
		  AND start_time > $5
		  AND deleted_at IS NULL
		ORDER BY start_time ASC
	`

	var lessons []*models.Lesson
	err := r.db.SelectContext(ctx, &lessons, query, teacherID, dayOfWeek, hour, minute, afterDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get lessons by time pattern: %w", err)
	}

	return lessons, nil
}

// IsStudentBookedForLessonTx checks if a student is already booked for a lesson within a transaction
func (r *LessonRepository) IsStudentBookedForLessonTx(ctx context.Context, tx pgx.Tx, lessonID uuid.UUID, studentID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM bookings
			WHERE lesson_id = $1 AND student_id = $2 AND status = 'active'
		)
	`

	var exists bool
	err := tx.QueryRow(ctx, query, lessonID, studentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check student booking: %w", err)
	}

	return exists, nil
}

// CreateBookingTx creates a booking within a transaction
func (r *LessonRepository) CreateBookingTx(ctx context.Context, tx pgx.Tx, booking *models.Booking) error {
	query := `
		INSERT INTO bookings (id, student_id, lesson_id, status, booked_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

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
		return fmt.Errorf("failed to create booking: %w", err)
	}

	return nil
}

// CancelBookingTx cancels a booking within a transaction
func (r *LessonRepository) CancelBookingTx(ctx context.Context, tx pgx.Tx, lessonID uuid.UUID, studentID uuid.UUID) error {
	query := `
		UPDATE bookings
		SET status = 'cancelled', cancelled_at = $1, updated_at = $2
		WHERE lesson_id = $3 AND student_id = $4 AND status = 'active'
	`

	result, err := tx.Exec(ctx, query, time.Now(), time.Now(), lessonID, studentID)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no active booking found for student on this lesson")
	}

	return nil
}

// UpdateTeacherTx updates the teacher for a lesson within a transaction
func (r *LessonRepository) UpdateTeacherTx(ctx context.Context, tx pgx.Tx, lessonID uuid.UUID, newTeacherID uuid.UUID) error {
	query := `
		UPDATE lessons
		SET teacher_id = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := tx.Exec(ctx, query, newTeacherID, time.Now(), lessonID)
	if err != nil {
		// Проверяем, является ли ошибка нарушением EXCLUDE constraint (конфликт расписания)
		if IsExclusionViolationError(err) {
			return ErrLessonOverlapConflict
		}
		return fmt.Errorf("failed to update teacher: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrLessonNotFound
	}

	return nil
}

// UpdateTimeTx updates the start and end time for a lesson within a transaction
func (r *LessonRepository) UpdateTimeTx(ctx context.Context, tx pgx.Tx, lessonID uuid.UUID, newStartTime time.Time, newEndTime time.Time) error {
	query := `
		UPDATE lessons
		SET start_time = $1, end_time = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	result, err := tx.Exec(ctx, query, newStartTime, newEndTime, time.Now(), lessonID)
	if err != nil {
		// Проверяем, является ли ошибка нарушением EXCLUDE constraint (конфликт расписания)
		if IsExclusionViolationError(err) {
			return ErrLessonOverlapConflict
		}
		return fmt.Errorf("failed to update lesson time: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrLessonNotFound
	}

	return nil
}

// UpdateMaxStudentsTx updates the max_students capacity within a transaction
func (r *LessonRepository) UpdateMaxStudentsTx(ctx context.Context, tx pgx.Tx, lessonID uuid.UUID, newMaxStudents int) error {
	query := `
		UPDATE lessons
		SET max_students = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL AND current_students <= $1
	`

	result, err := tx.Exec(ctx, query, newMaxStudents, time.Now(), lessonID)
	if err != nil {
		return fmt.Errorf("failed to update max students: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("cannot set max_students below current_students or lesson not found")
	}

	return nil
}

// CreateLessonTx creates a lesson within a transaction
func (r *LessonRepository) CreateLessonTx(ctx context.Context, tx pgx.Tx, lesson *models.Lesson) (*models.Lesson, error) {
	query := `
		INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, credits_cost, color, subject, homework_text, link,
		                     created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, teacher_id, start_time, end_time, max_students, current_students, credits_cost, color, subject, homework_text, link,
		          created_at, updated_at, deleted_at
	`

	var created models.Lesson
	err := tx.QueryRow(ctx, query,
		lesson.ID,
		lesson.TeacherID,
		lesson.StartTime,
		lesson.EndTime,
		lesson.MaxStudents,
		lesson.CurrentStudents,
		lesson.CreditsCost,
		lesson.Color,
		lesson.Subject,
		lesson.HomeworkText,
		lesson.Link,
		lesson.CreatedAt,
		lesson.UpdatedAt,
	).Scan(
		&created.ID,
		&created.TeacherID,
		&created.StartTime,
		&created.EndTime,
		&created.MaxStudents,
		&created.CurrentStudents,
		&created.CreditsCost,
		&created.Color,
		&created.Subject,
		&created.HomeworkText,
		&created.Link,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create lesson: %w", err)
	}

	return &created, nil
}

// GetBookingsByLessonTx retrieves all bookings for a specific lesson within a transaction
func (r *LessonRepository) GetBookingsByLessonTx(ctx context.Context, tx pgx.Tx, lessonID uuid.UUID) ([]*models.Booking, error) {
	query := `
		SELECT ` + BookingSelectFields + `
		FROM bookings
		WHERE lesson_id = $1 AND status = 'active'
		ORDER BY booked_at ASC
	`

	rows, err := tx.Query(ctx, query, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings for lesson: %w", err)
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		var booking models.Booking
		err := rows.Scan(
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
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading booking rows: %w", err)
	}

	return bookings, nil
}

// DeleteLessonTx soft deletes a lesson within a transaction
func (r *LessonRepository) DeleteLessonTx(ctx context.Context, tx pgx.Tx, lessonID uuid.UUID) error {
	query := `
		UPDATE lessons
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := tx.Exec(ctx, query, time.Now(), time.Now(), lessonID)
	if err != nil {
		return fmt.Errorf("failed to delete lesson: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrLessonNotFound
	}

	return nil
}

// GetTeacherSchedule получает расписание преподавателя за период с дополнительной информацией
// Возвращает lessons с количеством enrolled students, homework и broadcasts
func (r *LessonRepository) GetTeacherSchedule(ctx context.Context, teacherID uuid.UUID, startDate, endDate time.Time) ([]*models.TeacherScheduleLesson, error) {
	query := `
		SELECT
			l.id, l.teacher_id, l.start_time, l.end_time,
			l.max_students, l.current_students, l.credits_cost, l.color, l.subject, l.homework_text, l.link,
			l.created_at, l.updated_at, l.deleted_at,
			CONCAT(u.first_name, ' ', u.last_name) as teacher_name,
			COALESCE(
				(SELECT COUNT(*) FROM bookings b WHERE b.lesson_id = l.id AND b.status = 'active'),
				0
			) as enrolled_students_count,
			COALESCE(
				(SELECT COUNT(*) FROM lesson_homework lh WHERE lh.lesson_id = l.id),
				0
			) as homework_count,
			COALESCE(
				(SELECT COUNT(*) FROM lesson_broadcasts lb WHERE lb.lesson_id = l.id),
				0
			) as broadcasts_count
		FROM lessons l
		JOIN users u ON l.teacher_id = u.id
		WHERE l.teacher_id = $1
		  AND l.start_time >= $2
		  AND l.start_time <= $3
		  AND l.deleted_at IS NULL
		ORDER BY l.start_time ASC
	`

	var lessons []*models.TeacherScheduleLesson
	err := r.db.SelectContext(ctx, &lessons, query, teacherID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get teacher schedule: %w", err)
	}

	// Для каждого занятия загрузим enrolled students (с полными данными)
	for _, lesson := range lessons {
		studentsQuery := `
			SELECT u.id, u.first_name, u.last_name, u.email
			FROM users u
			INNER JOIN bookings b ON u.id = b.student_id
			WHERE b.lesson_id = $1 AND b.status = 'active'
			ORDER BY b.booked_at ASC
		`

		var students []*models.EnrolledStudent
		err := r.db.SelectContext(ctx, &students, studentsQuery, lesson.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get enrolled students for lesson %s: %w", lesson.ID, err)
		}

		lesson.EnrolledStudents = students
	}

	return lessons, nil
}

// GetLessonBookings получает информацию о студентах, записанных на занятие
func (r *LessonRepository) GetLessonBookings(ctx context.Context, lessonID uuid.UUID) ([]models.BookingInfo, error) {
	query := `
		SELECT b.student_id, CONCAT(u.first_name, ' ', u.last_name) AS student_name
		FROM bookings b
		JOIN users u ON b.student_id = u.id
		WHERE b.lesson_id = $1 AND b.status = 'active'
		ORDER BY u.first_name, u.last_name
	`

	var bookings []models.BookingInfo
	err := r.db.SelectContext(ctx, &bookings, query, lessonID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []models.BookingInfo{}, nil
		}
		return nil, fmt.Errorf("failed to get bookings for lesson %s: %w", lessonID, err)
	}

	if bookings == nil {
		bookings = []models.BookingInfo{}
	}

	return bookings, nil
}

// GetLessonBookingsForLessons получает информацию о студентах для ВСЕХ занятий в одном запросе (batch query)
// Возвращает map[lessonID][]BookingInfo для эффективной загрузки bookings вместо N+1 queries
func (r *LessonRepository) GetLessonBookingsForLessons(ctx context.Context, lessonIDs []uuid.UUID) (map[uuid.UUID][]models.BookingInfo, error) {
	if len(lessonIDs) == 0 {
		return make(map[uuid.UUID][]models.BookingInfo), nil
	}

	query := `
		SELECT b.lesson_id, b.student_id, CONCAT(u.first_name, ' ', u.last_name) AS student_name
		FROM bookings b
		JOIN users u ON b.student_id = u.id
		WHERE b.lesson_id = ANY($1) AND b.status = 'active'
		ORDER BY b.lesson_id, u.first_name, u.last_name
	`

	type bookingRow struct {
		LessonID    uuid.UUID `db:"lesson_id"`
		StudentID   uuid.UUID `db:"student_id"`
		StudentName string    `db:"student_name"`
	}

	var rows []bookingRow
	err := r.db.SelectContext(ctx, &rows, query, lessonIDs)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get bookings for lessons: %w", err)
	}

	result := make(map[uuid.UUID][]models.BookingInfo)
	for _, row := range rows {
		result[row.LessonID] = append(result[row.LessonID], models.BookingInfo{
			StudentID:   row.StudentID,
			StudentName: row.StudentName,
		})
	}

	return result, nil
}

// GetAllTeachersSchedule получает расписание ВСЕХ преподавателей за период (для админа)
// Возвращает все lessons с количеством enrolled students, homework и broadcasts
func (r *LessonRepository) GetAllTeachersSchedule(ctx context.Context, startDate, endDate time.Time) ([]*models.TeacherScheduleLesson, error) {
	query := `
		SELECT
			l.id, l.teacher_id, l.start_time, l.end_time,
			l.max_students, l.current_students, l.credits_cost, l.color, l.subject, l.homework_text, l.link,
			l.created_at, l.updated_at, l.deleted_at,
			CONCAT(u.first_name, ' ', u.last_name) as teacher_name,
			COALESCE(
				(SELECT COUNT(*) FROM bookings b WHERE b.lesson_id = l.id AND b.status = 'active'),
				0
			) as enrolled_students_count,
			COALESCE(
				(SELECT COUNT(*) FROM lesson_homework lh WHERE lh.lesson_id = l.id),
				0
			) as homework_count,
			COALESCE(
				(SELECT COUNT(*) FROM lesson_broadcasts lb WHERE lb.lesson_id = l.id),
				0
			) as broadcasts_count
		FROM lessons l
		JOIN users u ON l.teacher_id = u.id
		WHERE l.start_time >= $1
		  AND l.start_time <= $2
		  AND l.deleted_at IS NULL
		ORDER BY l.start_time ASC
	`

	var lessons []*models.TeacherScheduleLesson
	err := r.db.SelectContext(ctx, &lessons, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get all teachers schedule: %w", err)
	}

	// Для каждого занятия загрузим enrolled students (с полными данными)
	for _, lesson := range lessons {
		studentsQuery := `
			SELECT u.id, u.first_name, u.last_name, u.email
			FROM users u
			INNER JOIN bookings b ON u.id = b.student_id
			WHERE b.lesson_id = $1 AND b.status = 'active'
			ORDER BY b.booked_at ASC
		`

		var students []*models.EnrolledStudent
		err := r.db.SelectContext(ctx, &students, studentsQuery, lesson.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get enrolled students for lesson %s: %w", lesson.ID, err)
		}

		lesson.EnrolledStudents = students
	}

	return lessons, nil
}

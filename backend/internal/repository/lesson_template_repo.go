package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// LessonTemplateRepository handles lesson template operations
type LessonTemplateRepository struct {
	db *sqlx.DB
}

// NewLessonTemplateRepository creates a new LessonTemplateRepository
func NewLessonTemplateRepository(db *sqlx.DB) *LessonTemplateRepository {
	return &LessonTemplateRepository{db: db}
}

// CreateTemplate creates a new lesson template (without entries)
func (r *LessonTemplateRepository) CreateTemplate(ctx context.Context, template *models.LessonTemplate) error {
	query := `
		INSERT INTO lesson_templates (id, admin_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, admin_id, name, description, created_at, updated_at, deleted_at
	`

	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		template.ID,
		template.AdminID,
		template.Name,
		template.Description,
		template.CreatedAt,
		template.UpdatedAt,
	).Scan(
		&template.ID,
		&template.AdminID,
		&template.Name,
		&template.Description,
		&template.CreatedAt,
		&template.UpdatedAt,
		&template.DeletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create lesson template: %w", err)
	}

	return nil
}

// GetTemplateByID retrieves a template by ID (without lessons)
func (r *LessonTemplateRepository) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.LessonTemplate, error) {
	query := `
		SELECT ` + LessonTemplateSelectFields + `
		FROM lesson_templates
		WHERE id = $1 AND deleted_at IS NULL
	`

	var template models.LessonTemplate
	err := r.db.GetContext(ctx, &template, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to get template by ID: %w", err)
	}

	return &template, nil
}

// GetAllTemplates retrieves all templates for a specific admin with lesson counts
func (r *LessonTemplateRepository) GetAllTemplates(ctx context.Context, adminID uuid.UUID) ([]*models.LessonTemplate, error) {
	// Подзапрос для подсчета занятий в каждом шаблоне
	query := `
		SELECT
			lt.id, lt.admin_id, lt.name, lt.description,
			lt.created_at, lt.updated_at, lt.deleted_at,
			COALESCE((SELECT COUNT(*) FROM template_lessons tl WHERE tl.template_id = lt.id), 0) as lesson_count
		FROM lesson_templates lt
		WHERE lt.deleted_at IS NULL
		ORDER BY lt.created_at DESC
	`

	var templates []*models.LessonTemplate

	// If adminID is not nil, filter by admin
	if adminID != uuid.Nil {
		query = `
			SELECT
				lt.id, lt.admin_id, lt.name, lt.description,
				lt.created_at, lt.updated_at, lt.deleted_at,
				COALESCE((SELECT COUNT(*) FROM template_lessons tl WHERE tl.template_id = lt.id), 0) as lesson_count
			FROM lesson_templates lt
			WHERE lt.admin_id = $1 AND lt.deleted_at IS NULL
			ORDER BY lt.created_at DESC
		`
		err := r.db.SelectContext(ctx, &templates, query, adminID)
		if err != nil {
			return nil, fmt.Errorf("failed to get templates: %w", err)
		}
	} else {
		err := r.db.SelectContext(ctx, &templates, query)
		if err != nil {
			return nil, fmt.Errorf("failed to get templates: %w", err)
		}
	}

	return templates, nil
}

// GetTemplateWithLessons retrieves a template with all its lesson entries and students
func (r *LessonTemplateRepository) GetTemplateWithLessons(ctx context.Context, id uuid.UUID) (*models.LessonTemplate, error) {
	// First get the template
	template, err := r.GetTemplateByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Query to get all template lessons with teacher names
	query := `
		SELECT
			tl.id, tl.template_id, tl.day_of_week,
			tl.start_time::text as start_time,
			tl.end_time::text as end_time,
			tl.teacher_id, tl.lesson_type, tl.max_students, tl.credits_cost, tl.color, tl.subject, tl.description,
			tl.created_at, tl.updated_at,
			u.full_name as teacher_name
		FROM template_lessons tl
		JOIN users u ON tl.teacher_id = u.id
		WHERE tl.template_id = $1
		ORDER BY tl.day_of_week, tl.start_time
	`

	var lessons []*models.TemplateLessonEntry
	err = r.db.SelectContext(ctx, &lessons, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get template lessons: %w", err)
	}

	log.Debug().
		Str("template_id", id.String()).
		Int("lessons_fetched", len(lessons)).
		Msg("Template lessons query completed")

	// Log credits_cost for each lesson
	for i, lesson := range lessons {
		log.Debug().
			Str("template_id", id.String()).
			Int("lesson_index", i).
			Str("lesson_id", lesson.ID.String()).
			Int("credits_cost", lesson.CreditsCost).
			Str("teacher_name", lesson.TeacherName).
			Int("day_of_week", lesson.DayOfWeek).
			Str("start_time", lesson.StartTime).
			Msg("Loaded template lesson with credits_cost")
	}

	// For each lesson, get its students
	for _, lesson := range lessons {
		studentsQuery := `
			SELECT
				tls.id, tls.template_lesson_id, tls.student_id, tls.created_at,
				u.full_name as student_name
			FROM template_lesson_students tls
			JOIN users u ON tls.student_id = u.id
			WHERE tls.template_lesson_id = $1
			ORDER BY u.full_name
		`

		var students []*models.TemplateLessonStudent
		err = r.db.SelectContext(ctx, &students, studentsQuery, lesson.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get lesson students: %w", err)
		}

		// DEBUG: Log students loaded for each lesson
		log.Debug().
			Str("template_lesson_id", lesson.ID.String()).
			Int("students_count", len(students)).
			Msg("Loaded students for template lesson")

		lesson.Students = students
	}

	// DEBUG: Final check before returning
	log.Info().
		Str("template_id", id.String()).
		Str("template_name", template.Name).
		Int("total_lessons", len(lessons)).
		Msg("Template fully loaded with lessons and students")

	template.Lessons = lessons
	// Compute LessonCount from Lessons array (single source of truth)
	template.LessonCount = len(lessons)
	return template, nil
}

// UpdateTemplate updates a template's basic info (name, description)
func (r *LessonTemplateRepository) UpdateTemplate(ctx context.Context, id uuid.UUID, name string, description sql.NullString) error {
	query := `
		UPDATE lesson_templates
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, name, description, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

// DeleteTemplate performs soft delete on a template
func (r *LessonTemplateRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE lesson_templates
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("template not found")
	}

	return nil
}

// GetOrCreateDefaultTemplate removed - use CreateTemplate and GetAllTemplates instead
// Multiple named templates are now supported instead of single default template

// TemplateLessonRepository handles template lesson entry operations
type TemplateLessonRepository struct {
	db *sqlx.DB
}

// NewTemplateLessonRepository creates a new TemplateLessonRepository
func NewTemplateLessonRepository(db *sqlx.DB) *TemplateLessonRepository {
	return &TemplateLessonRepository{db: db}
}

// CreateTemplateLessonEntry creates a new lesson entry in a template
func (r *TemplateLessonRepository) CreateTemplateLessonEntry(ctx context.Context, entry *models.TemplateLessonEntry) error {
	query := `
		INSERT INTO template_lessons
		(id, template_id, day_of_week, start_time, end_time, teacher_id, lesson_type, max_students, credits_cost, color, subject, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, template_id, day_of_week, start_time, end_time, teacher_id, lesson_type, max_students, credits_cost, color, subject, description, created_at, updated_at
	`

	entry.ID = uuid.New()
	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	// Set default lesson_type if not provided
	if entry.LessonType == "" {
		entry.LessonType = "individual"
	}

	// Set default credits_cost if not provided (must be >= 0)
	if entry.CreditsCost < 0 {
		entry.CreditsCost = 0 // Исправляем отрицательные значения на 0
	}

	log.Debug().
		Str("lesson_id", entry.ID.String()).
		Str("template_id", entry.TemplateID.String()).
		Int("credits_cost", entry.CreditsCost).
		Msg("Creating template lesson entry with credits_cost")

	err := r.db.QueryRowContext(ctx, query,
		entry.ID,
		entry.TemplateID,
		entry.DayOfWeek,
		entry.StartTime,
		entry.EndTime,
		entry.TeacherID,
		entry.LessonType,
		entry.MaxStudents,
		entry.CreditsCost,
		entry.Color,
		entry.Subject,
		entry.Description,
		entry.CreatedAt,
		entry.UpdatedAt,
	).Scan(
		&entry.ID,
		&entry.TemplateID,
		&entry.DayOfWeek,
		&entry.StartTime,
		&entry.EndTime,
		&entry.TeacherID,
		&entry.LessonType,
		&entry.MaxStudents,
		&entry.CreditsCost,
		&entry.Color,
		&entry.Subject,
		&entry.Description,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create template lesson entry: %w", err)
	}

	return nil
}

// GetTemplateLessonByID retrieves a template lesson by ID
func (r *TemplateLessonRepository) GetTemplateLessonByID(ctx context.Context, id uuid.UUID) (*models.TemplateLessonEntry, error) {
	query := `
		SELECT
			tl.id, tl.template_id, tl.day_of_week,
			tl.start_time::text as start_time,
			tl.end_time::text as end_time,
			tl.teacher_id, tl.lesson_type, tl.max_students, tl.credits_cost, tl.color, tl.subject, tl.description,
			tl.created_at, tl.updated_at,
			u.full_name as teacher_name
		FROM template_lessons tl
		JOIN users u ON tl.teacher_id = u.id
		WHERE tl.id = $1
	`

	var entry models.TemplateLessonEntry
	err := r.db.GetContext(ctx, &entry, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template lesson not found")
		}
		return nil, fmt.Errorf("failed to get template lesson: %w", err)
	}

	return &entry, nil
}

// GetTemplateLessonsByTemplateID retrieves all lesson entries for a template
func (r *TemplateLessonRepository) GetTemplateLessonsByTemplateID(ctx context.Context, templateID uuid.UUID) ([]*models.TemplateLessonEntry, error) {
	query := `
		SELECT
			tl.id, tl.template_id, tl.day_of_week,
			tl.start_time::text as start_time,
			tl.end_time::text as end_time,
			tl.teacher_id, tl.lesson_type, tl.max_students, tl.credits_cost, tl.color, tl.subject, tl.description,
			tl.created_at, tl.updated_at,
			u.full_name as teacher_name
		FROM template_lessons tl
		JOIN users u ON tl.teacher_id = u.id
		WHERE tl.template_id = $1
		ORDER BY tl.day_of_week, tl.start_time
	`

	var entries []*models.TemplateLessonEntry
	err := r.db.SelectContext(ctx, &entries, query, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template lessons: %w", err)
	}

	return entries, nil
}

// GetTemplateLessonsByDayOfWeek retrieves template lessons filtered by day of week
func (r *TemplateLessonRepository) GetTemplateLessonsByDayOfWeek(ctx context.Context, templateID uuid.UUID, dayOfWeek int) ([]*models.TemplateLessonEntry, error) {
	query := `
		SELECT
			tl.id, tl.template_id, tl.day_of_week,
			tl.start_time::text as start_time,
			tl.end_time::text as end_time,
			tl.teacher_id, tl.lesson_type, tl.max_students, tl.credits_cost, tl.color, tl.subject, tl.description,
			tl.created_at, tl.updated_at,
			u.full_name as teacher_name
		FROM template_lessons tl
		JOIN users u ON tl.teacher_id = u.id
		WHERE tl.template_id = $1 AND tl.day_of_week = $2
		ORDER BY tl.start_time
	`

	var entries []*models.TemplateLessonEntry
	err := r.db.SelectContext(ctx, &entries, query, templateID, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get template lessons by day: %w", err)
	}

	return entries, nil
}

// UpdateTemplateLessonEntry updates a template lesson entry
func (r *TemplateLessonRepository) UpdateTemplateLessonEntry(ctx context.Context, entry *models.TemplateLessonEntry) error {
	query := `
		UPDATE template_lessons
		SET day_of_week = $1, start_time = $2, end_time = $3, teacher_id = $4,
		    lesson_type = $5, max_students = $6, credits_cost = $7, color = $8, subject = $9, description = $10, updated_at = $11
		WHERE id = $12
	`

	entry.UpdatedAt = time.Now()

	// Log credits_cost value before executing UPDATE
	log.Debug().
		Str("lesson_id", entry.ID.String()).
		Int("credits_cost_to_save", entry.CreditsCost).
		Msg("UpdateTemplateLessonEntry: saving credits_cost to database")

	result, err := r.db.ExecContext(ctx, query,
		entry.DayOfWeek,
		entry.StartTime,
		entry.EndTime,
		entry.TeacherID,
		entry.LessonType,
		entry.MaxStudents,
		entry.CreditsCost,
		entry.Color,
		entry.Subject,
		entry.Description,
		entry.UpdatedAt,
		entry.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update template lesson: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("template lesson not found")
	}

	// Log successful update
	log.Debug().
		Str("lesson_id", entry.ID.String()).
		Int("credits_cost_saved", entry.CreditsCost).
		Int("rows_affected", int(rows)).
		Msg("UpdateTemplateLessonEntry: credits_cost successfully saved")

	return nil
}

// DeleteTemplateLessonEntry hard deletes a template lesson entry (cascade)
func (r *TemplateLessonRepository) DeleteTemplateLessonEntry(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM template_lessons
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template lesson: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("template lesson not found")
	}

	return nil
}

// DeleteAllTemplateLessons deletes all lesson entries for a template
func (r *TemplateLessonRepository) DeleteAllTemplateLessons(ctx context.Context, templateID uuid.UUID) error {
	query := `
		DELETE FROM template_lessons
		WHERE template_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, templateID)
	if err != nil {
		return fmt.Errorf("failed to delete template lessons: %w", err)
	}

	return nil
}

// GetStudentCountForTemplateLessonEntry returns the current number of students assigned to a template lesson
func (r *TemplateLessonRepository) GetStudentCountForTemplateLessonEntry(ctx context.Context, templateLessonID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM template_lesson_students
		WHERE template_lesson_id = $1
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, templateLessonID)
	if err != nil {
		return 0, fmt.Errorf("failed to get student count: %w", err)
	}

	return count, nil
}

// AddStudentToTemplateLessonEntry adds a student to a template lesson entry
// Проверяет вместимость занятия перед добавлением студента
func (r *TemplateLessonRepository) AddStudentToTemplateLessonEntry(ctx context.Context, templateLessonID uuid.UUID, studentID uuid.UUID) error {
	// Получаем информацию о занятии для проверки вместимости
	lesson, err := r.GetTemplateLessonByID(ctx, templateLessonID)
	if err != nil {
		return fmt.Errorf("failed to get template lesson: %w", err)
	}

	// Получаем текущее количество студентов
	currentCount, err := r.GetStudentCountForTemplateLessonEntry(ctx, templateLessonID)
	if err != nil {
		return err
	}

	// Проверяем вместимость перед добавлением
	if currentCount >= lesson.MaxStudents {
		return fmt.Errorf("cannot add student: lesson at capacity (%d/%d)", currentCount, lesson.MaxStudents)
	}

	query := `
		INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err = r.db.ExecContext(ctx, query, uuid.New(), templateLessonID, studentID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add student to template lesson: %w", err)
	}

	return nil
}

// RemoveStudentFromTemplateLessonEntry removes a student from a template lesson entry
func (r *TemplateLessonRepository) RemoveStudentFromTemplateLessonEntry(ctx context.Context, templateLessonID uuid.UUID, studentID uuid.UUID) error {
	query := `
		DELETE FROM template_lesson_students
		WHERE template_lesson_id = $1 AND student_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, templateLessonID, studentID)
	if err != nil {
		return fmt.Errorf("failed to remove student from template lesson: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("student not found in template lesson")
	}

	return nil
}

// GetStudentsForTemplateLessonEntry retrieves all students assigned to a template lesson
func (r *TemplateLessonRepository) GetStudentsForTemplateLessonEntry(ctx context.Context, templateLessonID uuid.UUID) ([]*models.TemplateLessonStudent, error) {
	query := `
		SELECT
			tls.id, tls.template_lesson_id, tls.student_id, tls.created_at,
			u.full_name as student_name
		FROM template_lesson_students tls
		JOIN users u ON tls.student_id = u.id
		WHERE tls.template_lesson_id = $1
		ORDER BY u.full_name
	`

	var students []*models.TemplateLessonStudent
	err := r.db.SelectContext(ctx, &students, query, templateLessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template lesson students: %w", err)
	}

	return students, nil
}

// WithTransaction executes a function within a transaction for template lesson operations
func (r *TemplateLessonRepository) WithTransaction(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	// Get the underlying pgx connection
	conn, err := r.db.Connx(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer conn.Close()

	// Begin transaction
	tx, err := conn.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute function
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// TemplateApplicationRepository handles template application tracking
type TemplateApplicationRepository struct {
	db *sqlx.DB
}

// NewTemplateApplicationRepository creates a new TemplateApplicationRepository
func NewTemplateApplicationRepository(db *sqlx.DB) *TemplateApplicationRepository {
	return &TemplateApplicationRepository{db: db}
}

// CreateTemplateApplication records a template application
func (r *TemplateApplicationRepository) CreateTemplateApplication(ctx context.Context, application *models.TemplateApplication) error {
	query := `
		INSERT INTO template_applications
		(id, template_id, applied_by_id, week_start_date, applied_at, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, template_id, applied_by_id, week_start_date, applied_at, status, rolled_back_at
	`

	application.ID = uuid.New()
	application.AppliedAt = time.Now()
	application.Status = "applied"

	err := r.db.QueryRowContext(ctx, query,
		application.ID,
		application.TemplateID,
		application.AppliedByID,
		application.WeekStartDate,
		application.AppliedAt,
		application.Status,
	).Scan(
		&application.ID,
		&application.TemplateID,
		&application.AppliedByID,
		&application.WeekStartDate,
		&application.AppliedAt,
		&application.Status,
		&application.RolledBackAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create template application: %w", err)
	}

	return nil
}

// GetTemplateApplicationByID retrieves a template application by ID
func (r *TemplateApplicationRepository) GetTemplateApplicationByID(ctx context.Context, id uuid.UUID) (*models.TemplateApplication, error) {
	query := `
		SELECT
			ta.id, ta.template_id, ta.applied_by_id, ta.week_start_date,
			ta.applied_at, ta.status, ta.rolled_back_at,
			lt.name as template_name,
			u.full_name as applied_by_name
		FROM template_applications ta
		JOIN lesson_templates lt ON ta.template_id = lt.id
		JOIN users u ON ta.applied_by_id = u.id
		WHERE ta.id = $1
	`

	var application models.TemplateApplication
	err := r.db.GetContext(ctx, &application, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template application not found")
		}
		return nil, fmt.Errorf("failed to get template application: %w", err)
	}

	return &application, nil
}

// GetApplicationsByWeekStartDate retrieves all applications for a specific week
func (r *TemplateApplicationRepository) GetApplicationsByWeekStartDate(ctx context.Context, weekStartDate time.Time) ([]*models.TemplateApplication, error) {
	query := `
		SELECT
			ta.id, ta.template_id, ta.applied_by_id, ta.week_start_date,
			ta.applied_at, ta.status, ta.rolled_back_at,
			lt.name as template_name,
			u.full_name as applied_by_name
		FROM template_applications ta
		JOIN lesson_templates lt ON ta.template_id = lt.id
		JOIN users u ON ta.applied_by_id = u.id
		WHERE ta.week_start_date = $1
		ORDER BY ta.applied_at DESC
	`

	var applications []*models.TemplateApplication
	err := r.db.SelectContext(ctx, &applications, query, weekStartDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications by week: %w", err)
	}

	return applications, nil
}

// GetApplicationsByTemplate retrieves all applications of a specific template
func (r *TemplateApplicationRepository) GetApplicationsByTemplate(ctx context.Context, templateID uuid.UUID) ([]*models.TemplateApplication, error) {
	query := `
		SELECT
			ta.id, ta.template_id, ta.applied_by_id, ta.week_start_date,
			ta.applied_at, ta.status, ta.rolled_back_at,
			lt.name as template_name,
			u.full_name as applied_by_name
		FROM template_applications ta
		JOIN lesson_templates lt ON ta.template_id = lt.id
		JOIN users u ON ta.applied_by_id = u.id
		WHERE ta.template_id = $1
		ORDER BY ta.week_start_date DESC, ta.applied_at DESC
	`

	var applications []*models.TemplateApplication
	err := r.db.SelectContext(ctx, &applications, query, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications by template: %w", err)
	}

	return applications, nil
}

// GetActiveApplicationForWeek retrieves the active (not rolled back) application for a week
func (r *TemplateApplicationRepository) GetActiveApplicationForWeek(ctx context.Context, weekStartDate time.Time) (*models.TemplateApplication, error) {
	query := `
		SELECT
			ta.id, ta.template_id, ta.applied_by_id, ta.week_start_date,
			ta.applied_at, ta.status, ta.rolled_back_at,
			lt.name as template_name,
			u.full_name as applied_by_name
		FROM template_applications ta
		JOIN lesson_templates lt ON ta.template_id = lt.id
		JOIN users u ON ta.applied_by_id = u.id
		WHERE ta.week_start_date = $1 AND ta.status = 'applied'
		ORDER BY ta.applied_at DESC
		LIMIT 1
	`

	var application models.TemplateApplication
	err := r.db.GetContext(ctx, &application, query, weekStartDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active application found (not an error)
		}
		return nil, fmt.Errorf("failed to get active application: %w", err)
	}

	return &application, nil
}

// UpdateApplicationStatus updates the status of a template application
func (r *TemplateApplicationRepository) UpdateApplicationStatus(ctx context.Context, applicationID uuid.UUID, newStatus string) error {
	query := `
		UPDATE template_applications
		SET status = $1, rolled_back_at = $2
		WHERE id = $3
	`

	var rolledBackAt sql.NullTime
	if newStatus == "rolled_back" {
		rolledBackAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	result, err := r.db.ExecContext(ctx, query, newStatus, rolledBackAt, applicationID)
	if err != nil {
		return fmt.Errorf("failed to update application status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("template application not found")
	}

	return nil
}

// GetApplicationHistory retrieves application history with pagination
func (r *TemplateApplicationRepository) GetApplicationHistory(ctx context.Context, templateID uuid.UUID, limit, offset int) ([]*models.TemplateApplication, error) {
	query := `
		SELECT
			ta.id, ta.template_id, ta.applied_by_id, ta.week_start_date,
			ta.applied_at, ta.status, ta.rolled_back_at,
			lt.name as template_name,
			u.full_name as applied_by_name
		FROM template_applications ta
		JOIN lesson_templates lt ON ta.template_id = lt.id
		JOIN users u ON ta.applied_by_id = u.id
		WHERE ta.template_id = $1
		ORDER BY ta.applied_at DESC
		LIMIT $2 OFFSET $3
	`

	var applications []*models.TemplateApplication
	err := r.db.SelectContext(ctx, &applications, query, templateID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get application history: %w", err)
	}

	return applications, nil
}

// GetAllApplicationHistory retrieves all application history (no template filter)
func (r *TemplateApplicationRepository) GetAllApplicationHistory(ctx context.Context, limit, offset int) ([]*models.TemplateApplication, error) {
	query := `
		SELECT
			ta.id, ta.template_id, ta.applied_by_id, ta.week_start_date,
			ta.applied_at, ta.status, ta.rolled_back_at,
			lt.name as template_name,
			u.full_name as applied_by_name
		FROM template_applications ta
		JOIN lesson_templates lt ON ta.template_id = lt.id
		JOIN users u ON ta.applied_by_id = u.id
		ORDER BY ta.applied_at DESC
		LIMIT $1 OFFSET $2
	`

	var applications []*models.TemplateApplication
	err := r.db.SelectContext(ctx, &applications, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all application history: %w", err)
	}

	return applications, nil
}

// CountApplicationsByTemplate counts total applications for a template
func (r *TemplateApplicationRepository) CountApplicationsByTemplate(ctx context.Context, templateID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM template_applications
		WHERE template_id = $1
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, templateID)
	if err != nil {
		return 0, fmt.Errorf("failed to count applications: %w", err)
	}

	return count, nil
}

// GetLessonsCreatedFromApplication retrieves lessons created from a specific template application
func (r *TemplateApplicationRepository) GetLessonsCreatedFromApplication(ctx context.Context, applicationID uuid.UUID) ([]*models.LessonWithTeacher, error) {
	query := `
		SELECT
			l.id, l.teacher_id, l.start_time, l.end_time,
			l.max_students, l.current_students, l.color, l.subject,
			l.applied_from_template, l.template_application_id, l.created_at, l.updated_at, l.deleted_at,
			u.full_name as teacher_name
		FROM lessons l
		JOIN users u ON l.teacher_id = u.id
		WHERE l.template_application_id = $1 AND l.deleted_at IS NULL
		ORDER BY l.start_time
	`

	var lessons []*models.LessonWithTeacher
	err := r.db.SelectContext(ctx, &lessons, query, applicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lessons from application: %w", err)
	}

	return lessons, nil
}

// CountLessonsCreatedFromApplication counts lessons created from a template application
func (r *TemplateApplicationRepository) CountLessonsCreatedFromApplication(ctx context.Context, applicationID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM lessons
		WHERE template_application_id = $1 AND deleted_at IS NULL
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, applicationID)
	if err != nil {
		return 0, fmt.Errorf("failed to count lessons from application: %w", err)
	}

	return count, nil
}

// CreateTemplateApplicationTx creates a template application within a transaction
func (r *TemplateApplicationRepository) CreateTemplateApplicationTx(ctx context.Context, tx interface{}, application *models.TemplateApplication) error {
	query := `
		INSERT INTO template_applications
		(id, template_id, applied_by_id, week_start_date, applied_at, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	var err error
	switch v := tx.(type) {
	case *sqlx.Tx:
		_, err = v.ExecContext(ctx, query,
			application.ID,
			application.TemplateID,
			application.AppliedByID,
			application.WeekStartDate,
			application.AppliedAt,
			application.Status,
		)
	default:
		return fmt.Errorf("unsupported transaction type")
	}

	if err != nil {
		return fmt.Errorf("failed to create template application: %w", err)
	}

	return nil
}

// GetApplicationByTemplateAndWeekTx retrieves application by template ID and week start date within a transaction
// Returns the most recent application (by applied_at DESC) with status 'applied' or 'replaced'
func (r *TemplateApplicationRepository) GetApplicationByTemplateAndWeekTx(ctx context.Context, tx interface{}, templateID uuid.UUID, weekStartDate string) (*models.TemplateApplication, error) {
	// Валидация формата даты (YYYY-MM-DD)
	parsedDate, err := time.Parse("2006-01-02", weekStartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid week_start_date format: %w", err)
	}

	// DEBUG: Сначала получаем ВСЕ applications для этой недели независимо от статуса
	debugQuery := `
		SELECT
			ta.id, ta.template_id, ta.status, ta.week_start_date::text as week_str,
			ta.applied_at
		FROM template_applications ta
		WHERE ta.week_start_date = $1
		ORDER BY ta.applied_at DESC
	`

	type DebugApp struct {
		ID         uuid.UUID `db:"id"`
		TemplateID uuid.UUID `db:"template_id"`
		Status     string    `db:"status"`
		WeekStr    string    `db:"week_str"`
		AppliedAt  time.Time `db:"applied_at"`
	}

	var debugApps []DebugApp
	switch v := tx.(type) {
	case *sqlx.Tx:
		// Используем parsedDate для точного сравнения DATE типа
		_ = v.SelectContext(ctx, &debugApps, debugQuery, parsedDate)

		// Логируем все найденные applications для этой недели
		if len(debugApps) > 0 {
			fmt.Printf("[DEBUG] Found %d applications for week %s:\n", len(debugApps), weekStartDate)
			for i, app := range debugApps {
				matchesTemplate := "NO"
				if app.TemplateID == templateID {
					matchesTemplate = "YES"
				}
				fmt.Printf("[DEBUG]   %d. ID=%s, Template=%s (matches=%s), Status=%s, WeekStr=%s, AppliedAt=%s\n",
					i+1, app.ID, app.TemplateID, matchesTemplate, app.Status, app.WeekStr, app.AppliedAt.Format("2006-01-02 15:04:05"))
			}
		} else {
			fmt.Printf("[DEBUG] No applications found for week %s in database\n", weekStartDate)
		}
	}

	// Основной запрос: ищем application с нужным template_id и статусом 'applied' или 'replaced'
	// Используем DATE() для надёжного сравнения без timezone проблем
	query := `
		SELECT
			ta.id, ta.template_id, ta.applied_by_id, ta.week_start_date,
			ta.applied_at, ta.status, ta.rolled_back_at
		FROM template_applications ta
		WHERE ta.template_id = $1
		  AND ta.week_start_date = $2
		  AND (ta.status = 'applied' OR ta.status = 'replaced')
		ORDER BY ta.applied_at DESC
		LIMIT 1
	`

	var application models.TemplateApplication
	switch v := tx.(type) {
	case *sqlx.Tx:
		// Используем parsedDate вместо строки для точного сравнения
		err = v.GetContext(ctx, &application, query, templateID, parsedDate)

		if err == nil {
			fmt.Printf("[DEBUG] Successfully found application: ID=%s, Status=%s, WeekStartDate=%s\n",
				application.ID, application.Status, application.WeekStartDate.Format("2006-01-02"))
		}
	default:
		return nil, fmt.Errorf("unsupported transaction type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("[DEBUG] No application found with template_id=%s AND week_start_date=%s AND status IN ('applied', 'replaced')\n",
				templateID, weekStartDate)
			return nil, fmt.Errorf("no active application found for template %s on week %s (must have status 'applied' or 'replaced')", templateID, weekStartDate)
		}
		return nil, fmt.Errorf("failed to get application by template and week: %w", err)
	}

	return &application, nil
}

// MarkAsRolledBackTx marks a template application as rolled back within a transaction
func (r *TemplateApplicationRepository) MarkAsRolledBackTx(ctx context.Context, tx interface{}, applicationID uuid.UUID) error {
	query := `
		UPDATE template_applications
		SET status = 'rolled_back', rolled_back_at = $1
		WHERE id = $2
	`

	var err error
	switch v := tx.(type) {
	case *sqlx.Tx:
		_, err = v.ExecContext(ctx, query, time.Now(), applicationID)
	default:
		return fmt.Errorf("unsupported transaction type")
	}

	if err != nil {
		return fmt.Errorf("failed to mark application as rolled back: %w", err)
	}

	return nil
}

// MarkApplicationAsReplacedTx отмечает application как заменённое
// Обновляет статус на 'replaced' только если текущий статус 'applied'
// Возвращает nil если application не существует или уже не в статусе 'applied'
func (r *TemplateApplicationRepository) MarkApplicationAsReplacedTx(
	ctx context.Context,
	tx *sqlx.Tx,
	applicationID uuid.UUID,
) error {
	query := `
		UPDATE template_applications
		SET status = 'replaced', updated_at = NOW()
		WHERE id = $1 AND status = 'applied'
		RETURNING id
	`

	var returnedID uuid.UUID
	err := tx.GetContext(ctx, &returnedID, query, applicationID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Application либо не существует, либо не в статусе 'applied'
			// Не возвращаем ошибку - может быть нет application
			return nil
		}
		return fmt.Errorf("failed to mark application as replaced: %w", err)
	}

	return nil
}

// GetApplicationsByWeekDateTx получает ВСЕ applications (любого статуса) для недели
// Используется для проверки истории применений template к конкретной неделе
func (r *TemplateApplicationRepository) GetApplicationsByWeekDateTx(
	ctx context.Context,
	tx *sqlx.Tx,
	weekDate time.Time,
) ([]*models.TemplateApplication, error) {
	query := `
		SELECT
			id, template_id, applied_by_id, week_start_date,
			applied_at, rolled_back_at, status,
			created_at, updated_at
		FROM template_applications
		WHERE DATE(week_start_date) = $1
		ORDER BY applied_at DESC
	`

	var applications []*models.TemplateApplication
	err := tx.SelectContext(ctx, &applications, query, weekDate)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query applications by week: %w", err)
	}

	return applications, nil
}

// WeekLessonStats статистика по lessons на неделе
type WeekLessonStats struct {
	TotalLessons     int `db:"total_lessons"`
	TotalBookings    int `db:"total_bookings"`
	TotalCredits     int `db:"total_credits"`
	AffectedStudents int `db:"affected_students"`
}

// GetLessonStatsByWeekTx возвращает статистику по lessons на неделе
// Подсчитывает количество занятий, активных бронирований и затронутых студентов
func (r *TemplateApplicationRepository) GetLessonStatsByWeekTx(
	ctx context.Context,
	tx *sqlx.Tx,
	weekDate time.Time,
) (*WeekLessonStats, error) {
	// Нормализуем weekDate к началу дня в UTC
	weekStart := time.Date(weekDate.Year(), weekDate.Month(), weekDate.Day(), 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	query := `
		SELECT
			COUNT(DISTINCT l.id) as total_lessons,
			COUNT(DISTINCT CASE WHEN b.status = 'active' THEN b.id END) as total_bookings,
			COUNT(DISTINCT CASE WHEN b.status = 'active' THEN b.id END) as total_credits,
			COUNT(DISTINCT b.student_id) as affected_students
		FROM lessons l
		LEFT JOIN bookings b ON l.id = b.lesson_id
		WHERE l.start_time >= $1
		  AND l.start_time < $2
		  AND l.deleted_at IS NULL
	`

	var stats WeekLessonStats
	err := tx.GetContext(ctx, &stats, query, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get week lesson stats: %w", err)
	}

	return &stats, nil
}

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// TemplateService handles business logic for lesson templates
type TemplateService struct {
	db                 *sqlx.DB
	templateRepo       *repository.LessonTemplateRepository
	templateLessonRepo *repository.TemplateLessonRepository
	templateAppRepo    *repository.TemplateApplicationRepository
	lessonRepo         *repository.LessonRepository
	creditRepo         *repository.CreditRepository
	bookingRepo        *repository.BookingRepository
	userRepo           repository.UserRepository
}

// NewTemplateService creates a new TemplateService
func NewTemplateService(
	db *sqlx.DB,
	templateRepo *repository.LessonTemplateRepository,
	templateLessonRepo *repository.TemplateLessonRepository,
	templateAppRepo *repository.TemplateApplicationRepository,
	lessonRepo *repository.LessonRepository,
	creditRepo *repository.CreditRepository,
	bookingRepo *repository.BookingRepository,
	userRepo repository.UserRepository,
) *TemplateService {
	return &TemplateService{
		db:                 db,
		templateRepo:       templateRepo,
		templateLessonRepo: templateLessonRepo,
		templateAppRepo:    templateAppRepo,
		lessonRepo:         lessonRepo,
		creditRepo:         creditRepo,
		bookingRepo:        bookingRepo,
		userRepo:           userRepo,
	}
}

// validateStudentCapacity проверяет что количество студентов не превышает вместимость занятия
func validateStudentCapacity(studentCount, maxStudents int) error {
	if studentCount > maxStudents {
		return fmt.Errorf("cannot assign %d students to lesson with capacity %d", studentCount, maxStudents)
	}
	return nil
}

// CreateTemplate creates a new template with lessons and student assignments
func (s *TemplateService) CreateTemplate(ctx context.Context, adminID uuid.UUID, req *models.CreateLessonTemplateRequest) (*models.LessonTemplate, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create template
	template := &models.LessonTemplate{
		ID:        uuid.New(),
		AdminID:   adminID,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Description != nil {
		template.Description = sql.NullString{String: *req.Description, Valid: true}
	}

	// Create template in database
	if err := s.templateRepo.CreateTemplate(ctx, template); err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	// Add lessons to template
	for _, lessonReq := range req.Lessons {
		// Apply defaults
		lessonReq.ApplyDefaults()

		// Validate lesson
		if err := lessonReq.Validate(); err != nil {
			return nil, fmt.Errorf("invalid lesson in template: %w", err)
		}

		// Create template lesson entry
		entry := &models.TemplateLessonEntry{
			ID:          uuid.New(),
			TemplateID:  template.ID,
			DayOfWeek:   lessonReq.DayOfWeek,
			StartTime:   lessonReq.StartTime,
			TeacherID:   lessonReq.TeacherID,
			LessonType:  *lessonReq.LessonType, // Set from request (with default applied in ApplyDefaults)
			MaxStudents: *lessonReq.MaxStudents,
			CreditsCost: *lessonReq.CreditsCost, // Default applied in ApplyDefaults
			Color:       *lessonReq.Color,       // Default applied in ApplyDefaults
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if lessonReq.EndTime != nil {
			entry.EndTime = *lessonReq.EndTime
		}

		if lessonReq.Subject != nil {
			entry.Subject = sql.NullString{String: *lessonReq.Subject, Valid: true}
		}

		if lessonReq.Description != nil {
			entry.Description = sql.NullString{String: *lessonReq.Description, Valid: true}
		}

		// Create lesson entry
		if err := s.templateLessonRepo.CreateTemplateLessonEntry(ctx, entry); err != nil {
			return nil, fmt.Errorf("failed to create template lesson: %w", err)
		}

		// Проверяем вместимость перед добавлением студентов
		if err := validateStudentCapacity(len(lessonReq.StudentIDs), *lessonReq.MaxStudents); err != nil {
			return nil, err
		}

		// Add pre-assigned students
		for _, studentID := range lessonReq.StudentIDs {
			if err := s.templateLessonRepo.AddStudentToTemplateLessonEntry(ctx, entry.ID, studentID); err != nil {
				return nil, fmt.Errorf("failed to add student to template lesson: %w", err)
			}
		}

		// Загружаем студентов для возврата в ответе
		if len(lessonReq.StudentIDs) > 0 {
			students, err := s.templateLessonRepo.GetStudentsForTemplateLessonEntry(ctx, entry.ID)
			if err != nil {
				entry.Students = []*models.TemplateLessonStudent{}
			} else {
				entry.Students = students
			}
		}

		// Add to template's lessons collection
		template.Lessons = append(template.Lessons, entry)
	}

	// Compute LessonCount from Lessons array (single source of truth)
	template.LessonCount = len(template.Lessons)
	return template, nil
}

// UpdateTemplate updates a template's basic info (name, description)
func (s *TemplateService) UpdateTemplate(ctx context.Context, adminID uuid.UUID, templateID uuid.UUID, req *models.UpdateLessonTemplateRequest) error {
	// Load existing template
	template, err := s.templateRepo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// Get user to check role
	user, err := s.userRepo.GetByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify ownership: only creator, admin, or methodologist can modify
	if template.AdminID != adminID && !user.IsAdmin() && !user.IsMethodologist() {
		log.Warn().
			Str("admin_id", adminID.String()).
			Str("template_id", templateID.String()).
			Str("template_creator_id", template.AdminID.String()).
			Msg("Unauthorized template modification attempt: user is not the creator")
		return fmt.Errorf("not authorized to modify this template: you are not the creator")
	}

	// Update fields if provided
	var name string
	var description sql.NullString

	if req.Name != nil {
		name = *req.Name
	} else {
		name = template.Name
	}

	if req.Description != nil {
		description = sql.NullString{String: *req.Description, Valid: true}
	} else {
		description = template.Description
	}

	// Update template
	if err := s.templateRepo.UpdateTemplate(ctx, templateID, name, description); err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	// Audit log: template modification
	log.Info().
		Str("admin_id", adminID.String()).
		Str("template_id", templateID.String()).
		Str("template_name", name).
		Str("modification_type", "update_template_info").
		Msg("Template successfully updated")

	return nil
}

// DeleteTemplate soft-deletes a template
func (s *TemplateService) DeleteTemplate(ctx context.Context, adminID uuid.UUID, templateID uuid.UUID) error {
	// Load template to verify permission
	template, err := s.templateRepo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// Get user to check role
	user, err := s.userRepo.GetByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify ownership: only creator, admin, or methodologist can delete
	if template.AdminID != adminID && !user.IsAdmin() && !user.IsMethodologist() {
		log.Warn().
			Str("admin_id", adminID.String()).
			Str("template_id", templateID.String()).
			Str("template_creator_id", template.AdminID.String()).
			Msg("Unauthorized template deletion attempt: user is not the creator")
		return fmt.Errorf("not authorized to delete this template: you are not the creator")
	}

	// Soft delete
	if err := s.templateRepo.DeleteTemplate(ctx, templateID); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	// Audit log: template deletion
	log.Warn().
		Str("admin_id", adminID.String()).
		Str("template_id", templateID.String()).
		Str("template_name", template.Name).
		Str("modification_type", "delete_template").
		Msg("Template successfully deleted (soft)")

	return nil
}

// ListTemplates retrieves all active templates (admin only)
func (s *TemplateService) ListTemplates(ctx context.Context, adminID uuid.UUID) ([]*models.LessonTemplate, error) {
	templates, err := s.templateRepo.GetAllTemplates(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	return templates, nil
}

// GetTemplateWithLessons retrieves a template with all lessons and student assignments
func (s *TemplateService) GetTemplateWithLessons(ctx context.Context, templateID uuid.UUID) (*models.LessonTemplate, error) {
	template, err := s.templateRepo.GetTemplateWithLessons(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	return template, nil
}

// GetOrCreateDefaultTemplate removed - use CreateTemplate and ListTemplates instead
// Multiple named templates are now supported

// CreditRequirement представляет требование по кредитам для одного студента
type CreditRequirement struct {
	StudentID       uuid.UUID
	StudentName     string // Для детальных сообщений об ошибках
	RequiredCredits int
	CurrentBalance  int
}

// ============================================================================
// Helper Methods для Week Cleanup и Replacement
// ============================================================================

// cleanupWeekLessonsInTx отменяет и удаляет все занятия на неделе в транзакции
// Возвращает статистику: сколько bookings отменено, кредитов возвращено, lessons удалено
func (s *TemplateService) cleanupWeekLessonsInTx(
	ctx context.Context,
	tx *sqlx.Tx,
	adminID uuid.UUID,
	weekDate time.Time,
	newApplicationID uuid.UUID,
) (*models.CleanupStats, error) {
	// Получить все неудалённые lessons на неделе (с блокировкой FOR UPDATE)
	lessons, err := s.getLessonsForWeekInTx(ctx, tx, weekDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get lessons for week: %w", err)
	}

	if len(lessons) == 0 {
		// Нет занятий на неделе - ничего не отменяем
		return &models.CleanupStats{}, nil
	}

	// Получить ID-шки всех lessons
	lessonIDs := make([]uuid.UUID, len(lessons))
	for i, lesson := range lessons {
		lessonIDs[i] = lesson.ID
	}

	// Получить все active bookings для этих lessons
	bookings, err := s.getBookingsForLessonsInTx(ctx, tx, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings: %w", err)
	}

	// Отменить ВСЕ active bookings
	cancelledCount, err := s.cancelBookingsInTx(ctx, tx, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel bookings: %w", err)
	}

	// Создать credit_transactions для возврата
	refundedCount, err := s.createRefundTransactionsInTx(ctx, tx, adminID, weekDate, bookings)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund transactions: %w", err)
	}

	// Soft-delete все lessons
	deletedCount, err := s.softDeleteLessonsInTx(ctx, tx, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to soft-delete lessons: %w", err)
	}

	// Отметить старое application как 'replaced' если оно есть
	// Исключаем только что созданное application из обновления
	oldAppID, err := s.markOldApplicationAsReplacedInTx(ctx, tx, weekDate, newApplicationID)
	if err != nil {
		// CRITICAL: Must propagate errors - if DB operation fails, transaction is already aborted
		// Only ErrNoRows is acceptable (no previous application to replace)
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to mark old application as replaced: %w", err)
		}
		// No previous application - this is normal during fresh apply
		log.Info().Str("week", weekDate.Format("2006-01-02")).Msg("no previous application to replace")
	}

	return &models.CleanupStats{
		CancelledBookings:     int(cancelledCount),
		RefundedCredits:       int(refundedCount),
		DeletedLessons:        int(deletedCount),
		ReplacedApplicationID: oldAppID,
	}, nil
}

// getLessonsForWeekInTx получает все неудалённые lessons на неделе с блокировкой
func (s *TemplateService) getLessonsForWeekInTx(
	ctx context.Context,
	tx *sqlx.Tx,
	weekDate time.Time,
) ([]*models.Lesson, error) {
	// weekDate должна быть началом понедельника UTC
	weekStart := time.Date(weekDate.Year(), weekDate.Month(), weekDate.Day(), 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	query := `
		SELECT
			l.id,
			l.teacher_id,
			l.start_time,
			l.end_time,
			l.max_students,
			l.current_students,
			l.color,
			l.subject,
			l.created_at,
			l.updated_at,
			COALESCE(l.deleted_at, to_timestamp(0)) as deleted_at
		FROM lessons l
		WHERE l.start_time >= $1
		  AND l.start_time < $2
		  AND l.deleted_at IS NULL
		ORDER BY l.start_time
		FOR UPDATE
	`

	var lessons []*models.Lesson
	err := tx.SelectContext(ctx, &lessons, query, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to query lessons: %w", err)
	}

	return lessons, nil
}

// getBookingsForLessonsInTx получает все active bookings для указанных lessons
func (s *TemplateService) getBookingsForLessonsInTx(
	ctx context.Context,
	tx *sqlx.Tx,
	lessonIDs []uuid.UUID,
) ([]*models.BookingRecord, error) {
	if len(lessonIDs) == 0 {
		return []*models.BookingRecord{}, nil
	}

	query := `
		SELECT
			b.id,
			b.student_id,
			b.lesson_id,
			b.status,
			b.booked_at,
			CONCAT(u.first_name, ' ', u.last_name) as student_name,
			l.start_time
		FROM bookings b
		JOIN lessons l ON b.lesson_id = l.id
		JOIN users u ON b.student_id = u.id
		WHERE b.lesson_id = ANY($1::uuid[])
		  AND b.status = 'active'
		  AND l.deleted_at IS NULL
		ORDER BY b.student_id, l.start_time
		FOR UPDATE
	`

	var bookings []*models.BookingRecord
	err := tx.SelectContext(ctx, &bookings, query, pq.Array(lessonIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query bookings: %w", err)
	}

	return bookings, nil
}

// cancelBookingsInTx отменяет все active bookings для указанных lessons
func (s *TemplateService) cancelBookingsInTx(
	ctx context.Context,
	tx *sqlx.Tx,
	lessonIDs []uuid.UUID,
) (int64, error) {
	if len(lessonIDs) == 0 {
		return 0, nil
	}

	query := `
		UPDATE bookings
		SET status = 'cancelled', cancelled_at = NOW(), updated_at = NOW()
		WHERE lesson_id = ANY($1::uuid[])
		  AND status = 'active'
	`

	result, err := tx.ExecContext(ctx, query, pq.Array(lessonIDs))
	if err != nil {
		return 0, fmt.Errorf("failed to cancel bookings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// createRefundTransactionsInTx создаёт credit_transactions для возврата кредитов
// ATOMICITY GUARANTEE:
// - Вся функция вызывается внутри Serializable транзакции (заданной в ApplyTemplate)
// - Каждый student_id заблокирован через FOR UPDATE при SELECT
// - Используется RELATIVE UPDATE (balance = balance + $1) для защиты от race conditions
// - Если INSERT credit_transactions падает, UPDATE откатывается вместе со всей транзакцией
// - Если UPDATE падает, ничего не записывается и вся транзакция откатывается
// - Результат: ВСЕ кредиты возвращены И записаны ИЛИ НИЧЕГО не сделано (no partial states)
func (s *TemplateService) createRefundTransactionsInTx(
	ctx context.Context,
	tx *sqlx.Tx,
	adminID uuid.UUID,
	weekDate time.Time,
	bookings []*models.BookingRecord,
) (int64, error) {
	if len(bookings) == 0 {
		return 0, nil
	}

	// Группируем по студентам (1 кредит = 1 отменённое занятие)
	studentCredits := make(map[uuid.UUID]int)
	for _, b := range bookings {
		studentCredits[b.StudentID]++
	}

	// Вставляем отдельную транзакцию для каждого студента
	var totalRefundedCredits int64
	reason := fmt.Sprintf("Lesson cancelled during template replacement for week %s", weekDate.Format("2006-01-02"))

	for studentID, creditCount := range studentCredits {
		// Получаем текущий баланс с блокировкой строки (FOR UPDATE)
		// Это гарантирует что не будет concurrent update из другой транзакции
		var currentBalance int
		selectQuery := `
			SELECT balance
			FROM credits
			WHERE user_id = $1
			FOR UPDATE
		`

		err := tx.GetContext(ctx, &currentBalance, selectQuery, studentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, fmt.Errorf("кредитный счёт для студента %s не найден", studentID)
			}
			return 0, fmt.Errorf("ошибка получения баланса кредитов для студента %s: %w", studentID, err)
		}

		// Вычисляем новый баланс (для последующей записи в транзакцию)
		balanceAfter := currentBalance + creditCount

		// IMPORTANT: Используем RELATIVE UPDATE (balance = balance + $1)
		// Это обеспечивает atomicity: если функция вызывается параллельно,
		// каждый increment будет применён атомарно без потери данных
		// Пример: если два потока добавляют по 1 кредиту:
		//   Thread 1: balance = 10 -> UPDATE balance = balance + 1
		//   Thread 2: balance = 10 -> UPDATE balance = balance + 1
		// Результат: balance = 12 (оба increment учтены), а не 11 (потеря одного)
		// Но в нашем случае с FOR UPDATE row-level lock, это всё равно дополнительная гарантия
		updateQuery := `
			UPDATE credits
			SET balance = balance + $1, updated_at = NOW()
			WHERE user_id = $2
		`

		result, err := tx.ExecContext(ctx, updateQuery, creditCount, studentID)
		if err != nil {
			return 0, fmt.Errorf("ошибка обновления баланса кредитов для студента %s: %w", studentID, err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return 0, fmt.Errorf("не удалось обновить баланс кредитов для студента %s", studentID)
		}

		// Вставляем запись транзакции ПОСЛЕ успешного обновления баланса
		// Если INSERT падает - UPDATE откатывается вместе со всей транзакцией (Serializable уровень)
		// amount = положительное число (количество возвращённых кредитов)
		insertQuery := `
			INSERT INTO credit_transactions (id, user_id, amount, operation_type, reason, performed_by, balance_before, balance_after, created_at)
			VALUES ($1, $2, $3, 'refund', $4, $5, $6, $7, NOW())
		`

		_, err = tx.ExecContext(ctx, insertQuery, uuid.New(), studentID, creditCount, reason, adminID, currentBalance, balanceAfter)
		if err != nil {
			return 0, fmt.Errorf("ошибка создания записи транзакции возврата для студента %s: %w", studentID, err)
		}

		// Увеличиваем счётчик на количество кредитов (не на количество записей)
		totalRefundedCredits += int64(creditCount)
	}

	return totalRefundedCredits, nil
}

// softDeleteLessonsInTx мягко удаляет все lessons
func (s *TemplateService) softDeleteLessonsInTx(
	ctx context.Context,
	tx *sqlx.Tx,
	lessonIDs []uuid.UUID,
) (int64, error) {
	if len(lessonIDs) == 0 {
		return 0, nil
	}

	query := `
		UPDATE lessons
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = ANY($1::uuid[])
		  AND deleted_at IS NULL
	`

	result, err := tx.ExecContext(ctx, query, pq.Array(lessonIDs))
	if err != nil {
		return 0, fmt.Errorf("failed to soft-delete lessons: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// markOldApplicationAsReplacedInTx отмечает старое application как replaced
// Исключает из обновления только что созданное application (newApplicationID)
func (s *TemplateService) markOldApplicationAsReplacedInTx(
	ctx context.Context,
	tx *sqlx.Tx,
	weekDate time.Time,
	newApplicationID uuid.UUID,
) (uuid.UUID, error) {
	// Use date range comparison to avoid type mismatch issues with DATE() function
	// weekDate is always midnight UTC on a Monday, so compare normalized dates
	weekStart := time.Date(weekDate.Year(), weekDate.Month(), weekDate.Day(), 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	query := `
		UPDATE template_applications
		SET status = 'replaced', rolled_back_at = NOW(), updated_at = NOW()
		WHERE week_start_date >= $1
		  AND week_start_date < $2
		  AND status = 'applied'
		  AND id != $3
		RETURNING id
	`

	var appID uuid.UUID
	err := tx.GetContext(ctx, &appID, query, weekStart, weekEnd, newApplicationID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Нет существующего application - это нормально при чистом apply
			return uuid.UUID{}, nil
		}
		return uuid.UUID{}, fmt.Errorf("failed to mark application as replaced: %w", err)
	}

	return appID, nil
}

// validateCreditsSufficientAfterReplacementInTx проверяет хватит ли кредитов после замены
func (s *TemplateService) validateCreditsSufficientAfterReplacementInTx(
	ctx context.Context,
	tx *sqlx.Tx,
	weekDate time.Time,
	templateID uuid.UUID,
) error {
	// Это сложный запрос - проверим логику
	// Для каждого студента: current_balance + refunded - required >= 0

	weekStart := time.Date(weekDate.Year(), weekDate.Month(), weekDate.Day(), 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	query := `
		WITH affected_students AS (
			SELECT DISTINCT b.student_id
			FROM bookings b
			JOIN lessons l ON b.lesson_id = l.id
			WHERE l.start_time >= $1 AND l.start_time < $2
			  AND l.deleted_at IS NULL
			  AND b.status = 'active'
			UNION
			SELECT DISTINCT tls.student_id
			FROM template_lesson_students tls
			JOIN template_lessons tl ON tls.template_lesson_id = tl.id
			WHERE tl.template_id = $3
		),
		current_balances AS (
			SELECT
				c.user_id,
				c.balance,
				CONCAT(u.first_name, ' ', u.last_name) as full_name
			FROM credits c
			JOIN users u ON c.user_id = u.id
			WHERE c.user_id IN (SELECT student_id FROM affected_students)
			FOR UPDATE
		),
		refund_amounts AS (
			SELECT
				b.student_id,
				COUNT(*) as refund_count
			FROM bookings b
			JOIN lessons l ON b.lesson_id = l.id
			WHERE l.start_time >= $1 AND l.start_time < $2
			  AND l.deleted_at IS NULL
			  AND b.status = 'active'
			GROUP BY b.student_id
		),
		deduction_amounts AS (
			SELECT
				tls.student_id,
				COUNT(*) as deduction_count
			FROM template_lesson_students tls
			JOIN template_lessons tl ON tls.template_lesson_id = tl.id
			WHERE tl.template_id = $3
			GROUP BY tls.student_id
		),
		final_calculations AS (
			SELECT
				cb.user_id,
				cb.full_name,
				cb.balance as current_balance,
				COALESCE(ra.refund_count, 0) as refunded,
				COALESCE(da.deduction_count, 0) as required,
				cb.balance + COALESCE(ra.refund_count, 0) - COALESCE(da.deduction_count, 0) as final_balance
			FROM current_balances cb
			LEFT JOIN refund_amounts ra ON cb.user_id = ra.student_id
			LEFT JOIN deduction_amounts da ON cb.user_id = da.student_id
		)
		SELECT user_id, full_name, current_balance, refunded, required, final_balance
		FROM final_calculations
		WHERE final_balance < 0
	`

	type InsufficientCredit struct {
		UserID         uuid.UUID `db:"user_id"`
		FullName       string    `db:"full_name"`
		CurrentBalance int       `db:"current_balance"`
		Refunded       int       `db:"refunded"`
		Required       int       `db:"required"`
		FinalBalance   int       `db:"final_balance"`
	}

	var insufficients []InsufficientCredit
	err := tx.SelectContext(ctx, &insufficients, query, weekStart, weekEnd, templateID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("ошибка проверки кредитов: %w", err)
	}

	if len(insufficients) > 0 {
		var messages []string
		for _, ic := range insufficients {
			lacking := ic.Required - (ic.CurrentBalance + ic.Refunded)
			msg := fmt.Sprintf(
				"%s: баланс %d, требуется %d, не хватает %d",
				ic.FullName, ic.CurrentBalance+ic.Refunded, ic.Required, lacking,
			)
			messages = append(messages, msg)
		}
		detailedMsg := strings.Join(messages, "; ")
		return fmt.Errorf("недостаточно кредитов для применения шаблона. %s", detailedMsg)
	}

	return nil
}

// validateCreditsSufficientForReplacement проверяет хватит ли кредитов после замены ВНЕ транзакции
// Используется для предварительной валидации перед началом SERIALIZABLE транзакции
// Это предотвращает SQLSTATE 25P02 ошибки при таймауте FOR UPDATE внутри транзакции
func (s *TemplateService) validateCreditsSufficientForReplacement(
	ctx context.Context,
	weekDate time.Time,
	templateID uuid.UUID,
) error {
	// Это сложный запрос - проверим логику
	// Для каждого студента: current_balance + refunded - required >= 0

	weekStart := time.Date(weekDate.Year(), weekDate.Month(), weekDate.Day(), 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	// ВАЖНО: Запрос БЕЗ FOR UPDATE - нет блокировок, не вызывает таймауты
	query := `
		WITH affected_students AS (
			SELECT DISTINCT b.student_id
			FROM bookings b
			JOIN lessons l ON b.lesson_id = l.id
			WHERE l.start_time >= $1 AND l.start_time < $2
			  AND l.deleted_at IS NULL
			  AND b.status = 'active'
			UNION
			SELECT DISTINCT tls.student_id
			FROM template_lesson_students tls
			JOIN template_lessons tl ON tls.template_lesson_id = tl.id
			WHERE tl.template_id = $3
		),
		current_balances AS (
			SELECT
				c.user_id,
				c.balance,
				CONCAT(u.first_name, ' ', u.last_name) as full_name
			FROM credits c
			JOIN users u ON c.user_id = u.id
			WHERE c.user_id IN (SELECT student_id FROM affected_students)
		),
		refund_amounts AS (
			SELECT
				b.student_id,
				COUNT(*) as refund_count
			FROM bookings b
			JOIN lessons l ON b.lesson_id = l.id
			WHERE l.start_time >= $1 AND l.start_time < $2
			  AND l.deleted_at IS NULL
			  AND b.status = 'active'
			GROUP BY b.student_id
		),
		deduction_amounts AS (
			SELECT
				tls.student_id,
				COUNT(*) as deduction_count
			FROM template_lesson_students tls
			JOIN template_lessons tl ON tls.template_lesson_id = tl.id
			WHERE tl.template_id = $3
			GROUP BY tls.student_id
		),
		final_calculations AS (
			SELECT
				cb.user_id,
				cb.full_name,
				cb.balance as current_balance,
				COALESCE(ra.refund_count, 0) as refunded,
				COALESCE(da.deduction_count, 0) as required,
				cb.balance + COALESCE(ra.refund_count, 0) - COALESCE(da.deduction_count, 0) as final_balance
			FROM current_balances cb
			LEFT JOIN refund_amounts ra ON cb.user_id = ra.student_id
			LEFT JOIN deduction_amounts da ON cb.user_id = da.student_id
		)
		SELECT user_id, full_name, current_balance, refunded, required, final_balance
		FROM final_calculations
		WHERE final_balance < 0
	`

	type InsufficientCredit struct {
		UserID         uuid.UUID `db:"user_id"`
		FullName       string    `db:"full_name"`
		CurrentBalance int       `db:"current_balance"`
		Refunded       int       `db:"refunded"`
		Required       int       `db:"required"`
		FinalBalance   int       `db:"final_balance"`
	}

	var insufficients []InsufficientCredit
	err := s.db.SelectContext(ctx, &insufficients, query, weekStart, weekEnd, templateID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("ошибка проверки кредитов: %w", err)
	}

	if len(insufficients) > 0 {
		var messages []string
		for _, ic := range insufficients {
			lacking := ic.Required - (ic.CurrentBalance + ic.Refunded)
			msg := fmt.Sprintf(
				"%s: баланс %d, требуется %d, не хватает %d",
				ic.FullName, ic.CurrentBalance+ic.Refunded, ic.Required, lacking,
			)
			messages = append(messages, msg)
		}
		detailedMsg := strings.Join(messages, "; ")
		return fmt.Errorf("недостаточно кредитов для применения шаблона. %s", detailedMsg)
	}

	return nil
}

// ApplyTemplateToWeek applies a template to a specific calendar week
// CRITICAL: This is an atomic transaction that creates all lessons and bookings with credit deductions
func (s *TemplateService) ApplyTemplateToWeek(ctx context.Context, adminID uuid.UUID, req *models.ApplyTemplateRequest) (*models.TemplateApplication, error) {
	// Input validation
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Load the template with all lessons and student assignments
	template, err := s.templateRepo.GetTemplateWithLessons(ctx, req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// DEBUG: Log loaded template lessons count
	log.Info().
		Str("template_id", req.TemplateID.String()).
		Int("lessons_count", len(template.Lessons)).
		Msg("Template loaded for week application")

	// Validate week_start_date is a Monday
	// Используем UTC для корректного определения дня недели независимо от timezone сервера
	weekDate, err := time.Parse("2006-01-02", req.WeekStartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid week_start_date format: %w", err)
	}
	// Явно устанавливаем UTC для избежания timezone issues
	weekDate = time.Date(weekDate.Year(), weekDate.Month(), weekDate.Day(), 0, 0, 0, 0, time.UTC)

	if weekDate.Weekday() != time.Monday {
		return nil, fmt.Errorf("week_start_date must be a Monday, got %s", weekDate.Weekday())
	}

	// Проверяем есть ли существующее application для этой недели
	// Теперь вместо ошибки мы выполняем cleanup + replacement
	existingApp, err := s.templateAppRepo.GetActiveApplicationForWeek(ctx, weekDate)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing applications: %w", err)
	}
	// Запоминаем есть ли существующее application (будет использовано для cleanup)

	// === PRE-VALIDATION: Проверка кредитов ДО начала транзакции ===
	// ВАЖНО: Пропускаем простую проверку если есть existingApp (replacement),
	// потому что для replacement используется специальная проверка с учётом refund
	if existingApp == nil {
		// Собираем все требования по кредитам для всех студентов
		creditRequirements := make(map[uuid.UUID]*CreditRequirement)

		for _, templateLesson := range template.Lessons {
			for _, student := range templateLesson.Students {
				// Подсчитываем количество занятий для каждого студента
				if req, exists := creditRequirements[student.StudentID]; exists {
					req.RequiredCredits++
				} else {
					creditRequirements[student.StudentID] = &CreditRequirement{
						StudentID:       student.StudentID,
						StudentName:     student.StudentName,
						RequiredCredits: 1,
					}
				}
			}
		}

		// Проверяем баланс всех студентов ДО транзакции
		var insufficientCreditErrors []string
		for studentID, requirement := range creditRequirements {
			balance, err := s.creditRepo.GetBalance(ctx, studentID)
			if err != nil {
				return nil, fmt.Errorf("failed to check credits for student %s (%s): %w",
					requirement.StudentName, studentID, err)
			}

			// balance это *models.Credit, используем balance.Balance
			requirement.CurrentBalance = balance.Balance

			if balance.Balance < requirement.RequiredCredits {
				insufficientCreditErrors = append(insufficientCreditErrors,
					fmt.Sprintf("Студент %s (ID: %s) имеет %d кредитов, требуется %d",
						requirement.StudentName, studentID, balance.Balance, requirement.RequiredCredits))
			}
		}

		// Если у кого-то не хватает кредитов, возвращаем детальную ошибку БЕЗ начала транзакции
		if len(insufficientCreditErrors) > 0 {
			return nil, fmt.Errorf("недостаточно кредитов у студентов:\n%s",
				strings.Join(insufficientCreditErrors, "\n"))
		}
	} else {
		// === PRE-VALIDATION: Проверка кредитов для REPLACEMENT случая ===
		// Для replacement мы должны проверить баланс с учётом refund старых занятий
		// Используем новый метод БЕЗ транзакции чтобы избежать FOR UPDATE таймаутов
		if err := s.validateCreditsSufficientForReplacement(ctx, weekDate, req.TemplateID); err != nil {
			return nil, err
		}
	}

	// Если dry_run=true, возвращаем успех без реального применения
	if req.DryRun {
		// Подсчитываем количество бронирований (количество студентов во всех занятиях)
		totalBookings := 0
		for _, lesson := range template.Lessons {
			totalBookings += len(lesson.Students)
		}

		// Возвращаем предпросмотр того, что будет создано
		// Включаем lessons для того чтобы client'ы могли увидеть детали
		return &models.TemplateApplication{
			ID:                  uuid.New(),
			TemplateID:          req.TemplateID,
			AppliedByID:         adminID,
			WeekStartDate:       weekDate,
			AppliedAt:           time.Now(),
			Status:              "preview", // Явно указываем что это предпросмотр
			CreatedLessonsCount: len(template.Lessons),
			Lessons:             template.Lessons, // Добавляем lessons для детализации
			CreationStats: &models.CreationStats{
				CreatedLessons:  len(template.Lessons),
				CreatedBookings: totalBookings,
				DeductedCredits: totalBookings, // 1 кредит за 1 бронирование
			},
		}, nil
	}

	// === ATOMIC TRANSACTION STARTS ===
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err.Error() != "sql: transaction has already been committed or rolled back" {
			// Логируем ошибку rollback, но не перезаписываем оригинальную ошибку
			// В продакшене здесь должен быть logger
		}
	}()

	// Acquire advisory lock to prevent concurrent applications to same template/week
	// Uses hash of template_id + week timestamp for unique lock key
	lockKey := int64(req.TemplateID[0])<<32 | int64(weekDate.Unix()&0xFFFFFFFF)
	_, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, lockKey)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock for template application: %w", err)
	}

	// Re-check for existing application AFTER acquiring lock
	// This prevents race condition where two transactions both passed the initial check
	var existingAppIDInTx uuid.UUID
	err = tx.GetContext(ctx, &existingAppIDInTx, `
		SELECT id FROM template_applications
		WHERE week_start_date = $1 AND status IN ('applied', 'replaced')
		LIMIT 1
	`, weekDate)

	if err == nil {
		// Found existing application inside transaction
		if existingApp == nil {
			// We didn't know about this application - it was created by a concurrent transaction
			// This is a duplicate concurrent request, reject it
			return nil, fmt.Errorf("template already applied to week %s (application exists)", weekDate.Format("2006-01-02"))
		}
		// If existingApp != nil, this is an intentional replacement (user explicitly reapplying)
		// Proceed with replacement logic below
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check for existing application: %w", err)
	}
	// err == sql.ErrNoRows means no existing app exists, proceed with new application

	// Create template application record (for rollback tracking)
	application := &models.TemplateApplication{
		ID:            uuid.New(),
		TemplateID:    req.TemplateID,
		AppliedByID:   adminID,
		WeekStartDate: weekDate,
		AppliedAt:     time.Now(),
		Status:        "applied",
	}

	if err := s.templateAppRepo.CreateTemplateApplicationTx(ctx, tx, application); err != nil {
		return nil, fmt.Errorf("failed to create application record: %w", err)
	}

	// Cleanup выполняем ТОЛЬКО если есть существующее application (replacement)
	// Для новых заявок на "чистую" неделю cleanup не нужен
	var cleanupStats *models.CleanupStats
	if existingApp != nil {
		cleanupStats, err = s.cleanupWeekLessonsInTx(ctx, tx, adminID, weekDate, application.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to cleanup week: %w", err)
		}
		application.CleanupStats = cleanupStats
	} else {
		application.CleanupStats = &models.CleanupStats{}
	}

	// For each template lesson: create actual lessons and bookings
	var createdLessonsCount int
	var totalBookingsCreated int64
	var totalCreditsDeducted int64

	// DEBUG: Log before starting lesson creation loop
	log.Info().
		Int("template_lessons_to_create", len(template.Lessons)).
		Str("week_start", weekDate.Format("2006-01-02")).
		Msg("Starting lesson creation loop")

	for _, templateLesson := range template.Lessons {
		// Вычисляем дату для этого занятия
		// Frontend использует: 0=Monday, 1=Tuesday, ..., 6=Sunday
		// weekDate - это всегда понедельник выбранной недели
		// Поэтому просто добавляем day_of_week дней к понедельнику
		daysToAdd := templateLesson.DayOfWeek

		actualDate := weekDate.AddDate(0, 0, daysToAdd)

		// Parse time strings and create timestamps
		startTime, err := time.Parse("15:04:05", templateLesson.StartTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start_time: %w", err)
		}

		endTime, err := time.Parse("15:04:05", templateLesson.EndTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end_time: %w", err)
		}

		actualStartTime := time.Date(
			actualDate.Year(), actualDate.Month(), actualDate.Day(),
			startTime.Hour(), startTime.Minute(), startTime.Second(),
			0, time.UTC,
		)

		actualEndTime := time.Date(
			actualDate.Year(), actualDate.Month(), actualDate.Day(),
			endTime.Hour(), endTime.Minute(), endTime.Second(),
			0, time.UTC,
		)

		// Create the lesson
		lesson := &models.Lesson{
			ID:                    uuid.New(),
			TeacherID:             templateLesson.TeacherID,
			StartTime:             actualStartTime,
			EndTime:               actualEndTime,
			MaxStudents:           templateLesson.MaxStudents,
			CurrentStudents:       0,
			CreditsCost:           templateLesson.CreditsCost, // Copy credits_cost from template
			Color:                 templateLesson.Color,       // Copy color from template
			Subject:               templateLesson.Subject,     // Copy subject from template
			AppliedFromTemplate:   true,
			TemplateApplicationID: &application.ID,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}

		// DEBUG: Log each lesson creation
		log.Debug().
			Str("template_lesson_id", templateLesson.ID.String()).
			Int("day_of_week", templateLesson.DayOfWeek).
			Str("start_time", templateLesson.StartTime).
			Int("students_count", len(templateLesson.Students)).
			Msg("Creating lesson from template")

		// Create lesson in database
		createdLesson, err := s.createLessonInTx(ctx, tx, lesson)
		if err != nil {
			return nil, fmt.Errorf("failed to create lesson from template: %w", err)
		}

		// Add pre-assigned students (from template_lesson_students)
		for _, student := range templateLesson.Students {
			// CRITICAL: Check if student already has a booking at the same time
			// IMPORTANT: Use transaction-aware conflict check to see current state of bookings
			// including cancelled bookings from cleanup in this transaction
			hasConflict, err := s.bookingRepo.HasScheduleConflictInTx(ctx, tx, student.StudentID, actualStartTime, actualEndTime)
			if err != nil {
				// Log warning but continue - this shouldn't block template application
				log.Warn().
					Str("student_id", student.StudentID.String()).
					Str("lesson_start", actualStartTime.Format(time.RFC3339)).
					Err(err).
					Msg("Failed to check for schedule conflict, proceeding without conflict check")
			}

			// Skip booking creation if there's a conflict
			// This can happen if another process created a booking between pre-validation and now
			if hasConflict {
				log.Warn().
					Str("student_id", student.StudentID.String()).
					Str("student_name", student.StudentName).
					Str("lesson_start", actualStartTime.Format(time.RFC3339)).
					Msg("Skipped booking creation - student already has booking at this time (race condition)")
				continue
			}

			// Create booking for this student
			booking := &models.Booking{
				ID:        uuid.New(),
				StudentID: student.StudentID,
				LessonID:  createdLesson.ID,
				Status:    models.BookingStatusActive,
				BookedAt:  time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := s.createBookingInTx(ctx, tx, booking); err != nil {
				return nil, fmt.Errorf("failed to create booking for student %s: %w", student.StudentID, err)
			}

			// Deduct credit for this booking
			reason := fmt.Sprintf("Lesson booking (template %s, week %s)", req.TemplateID, req.WeekStartDate)
			if err := s.deductCreditInTx(ctx, tx, student.StudentID, adminID, booking.ID, reason); err != nil {
				return nil, fmt.Errorf("failed to deduct credit from student %s: %w", student.StudentID, err)
			}

			// Increment lesson's current_students counter
			lesson.CurrentStudents++

			// Подсчитываем статистику
			totalCreditsDeducted++
			totalBookingsCreated++
		}

		// Обновляем current_students в базе данных после добавления всех студентов
		if lesson.CurrentStudents > 0 {
			if err := s.updateLessonCurrentStudentsInTx(ctx, tx, createdLesson.ID, lesson.CurrentStudents); err != nil {
				return nil, fmt.Errorf("failed to update lesson current_students: %w", err)
			}
		}

		createdLessonsCount++
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return success response с полной статистикой
	application.CreatedLessonsCount = createdLessonsCount
	application.CreationStats = &models.CreationStats{
		CreatedLessons:  createdLessonsCount,
		CreatedBookings: int(totalBookingsCreated),
		DeductedCredits: int(totalCreditsDeducted),
	}

	return application, nil
}

// RollbackWeekToTemplate rolls back all lessons and bookings from a template application
// CRITICAL: This is an atomic transaction that deletes lessons and refunds all credits
func (s *TemplateService) RollbackWeekToTemplate(ctx context.Context, adminID uuid.UUID, weekStartDate string, templateID uuid.UUID) (*models.TemplateRollbackResponse, error) {
	// Validate date format
	weekDate, err := time.Parse("2006-01-02", weekStartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid week_start_date format: %w", err)
	}
	if weekDate.Weekday() != time.Monday {
		return nil, fmt.Errorf("week_start_date must be a Monday")
	}

	// === ATOMIC TRANSACTION STARTS ===
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err.Error() != "sql: transaction has already been committed or rolled back" {
			// Логируем ошибку rollback, но не перезаписываем оригинальную ошибку
			// В продакшене здесь должен быть logger
		}
	}()

	// Найти запись о применении шаблона для данной недели
	// Это может быть статус 'applied' (если это первая и единственная application)
	// или 'replaced' (если шаблон был применен повторно и была создана новая application)
	application, err := s.templateAppRepo.GetApplicationByTemplateAndWeekTx(ctx, tx, templateID, weekStartDate)
	if err != nil {
		// Проверяем существует ли application с любым статусом для детального сообщения об ошибке
		allApps, checkErr := s.templateAppRepo.GetApplicationsByTemplate(ctx, templateID)
		if checkErr == nil {
			// Ищем application для этой недели независимо от статуса
			for _, app := range allApps {
				appWeekStr := app.WeekStartDate.Format("2006-01-02")
				if appWeekStr == weekStartDate {
					// Нашли application для этой недели, но статус не 'applied' и не 'replaced'
					if app.Status == "rolled_back" {
						// Шаблон уже откачен - логируем попытку повторного отката
						rolledBackTime := "unknown"
						if app.RolledBackAt.Valid {
							rolledBackTime = app.RolledBackAt.Time.Format("2006-01-02 15:04:05")
						}
						log.Warn().
							Str("template_id", templateID.String()).
							Str("week_start_date", weekStartDate).
							Str("rolled_back_at", rolledBackTime).
							Str("admin_id", adminID.String()).
							Msg("Попытка повторного отката уже откатанной недели")
						return nil, fmt.Errorf("cannot rollback: week %s already rolled back at %s", weekStartDate, rolledBackTime)
					}
					// Статус существует, но это не 'applied', 'replaced' и не 'rolled_back' - неожиданное состояние
					log.Warn().
						Str("template_id", templateID.String()).
						Str("week_start_date", weekStartDate).
						Str("status", app.Status).
						Str("admin_id", adminID.String()).
						Msg("Попытка отката application с некорректным статусом")
					return nil, fmt.Errorf("cannot rollback: week %s application status is '%s' (expected 'applied' or 'replaced')", weekStartDate, app.Status)
				}
			}
		}
		// Application для этой недели не найдена вообще - шаблон не был применен
		log.Warn().
			Str("template_id", templateID.String()).
			Str("week_start_date", weekStartDate).
			Str("admin_id", adminID.String()).
			Msg("Попытка отката несуществующей application")
		return nil, fmt.Errorf("cannot rollback: no template application found for week %s (template must be applied first)", weekStartDate)
	}

	// Найти все занятия созданные из этого применения шаблона
	lessonsToDelete, err := s.getLessonsByTemplateApplicationInTx(ctx, tx, application.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find lessons from application: %w", err)
	}

	var refundedCredits int
	var cancellationErrors []string

	// For each lesson:
	for _, lesson := range lessonsToDelete {
		// Find all bookings for this lesson
		bookings, err := s.getBookingsByLessonInTx(ctx, tx, lesson.ID)
		if err != nil {
			cancellationErrors = append(cancellationErrors, fmt.Sprintf("Failed to find bookings for lesson %s: %v", lesson.ID, err))
			continue
		}

		// Cancel each booking and refund credit
		for _, booking := range bookings {
			// Cancel booking
			if err := s.cancelBookingInTx(ctx, tx, booking.ID); err != nil {
				cancellationErrors = append(cancellationErrors, fmt.Sprintf("Failed to cancel booking %s: %v", booking.ID, err))
				continue
			}

			// Refund 1 credit to student
			reason := fmt.Sprintf("Template rollback (template %s, week %s)", templateID, weekStartDate)
			if err := s.refundCreditInTx(ctx, tx, booking.StudentID, adminID, booking.ID, reason); err != nil {
				cancellationErrors = append(cancellationErrors, fmt.Sprintf("Failed to refund credit to student %s: %v", booking.StudentID, err))
				continue
			}

			refundedCredits++
		}

		// Soft delete the lesson
		if err := s.deleteLessonInTx(ctx, tx, lesson.ID); err != nil {
			cancellationErrors = append(cancellationErrors, fmt.Sprintf("Failed to delete lesson %s: %v", lesson.ID, err))
			continue
		}
	}

	// Mark application as rolled back
	if err := s.templateAppRepo.MarkAsRolledBackTx(ctx, tx, application.ID); err != nil {
		return nil, fmt.Errorf("failed to update application status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return response
	response := &models.TemplateRollbackResponse{
		ApplicationID:   application.ID,
		WeekStartDate:   weekStartDate,
		TemplateID:      templateID,
		DeletedLessons:  len(lessonsToDelete),
		RefundedCredits: refundedCredits,
		Warnings:        cancellationErrors,
	}

	return response, nil
}

// CreateTemplateLesson creates a new lesson entry in a template
func (s *TemplateService) CreateTemplateLesson(ctx context.Context, adminID uuid.UUID, templateID uuid.UUID, req *models.CreateTemplateLessonRequest) (*models.TemplateLessonEntry, error) {
	// Verify template exists
	template, err := s.templateRepo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Get user to check role
	user, err := s.userRepo.GetByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify ownership: only creator, admin, or methodologist can add lessons
	if template.AdminID != adminID && !user.IsAdmin() && !user.IsMethodologist() {
		log.Warn().
			Str("admin_id", adminID.String()).
			Str("template_id", templateID.String()).
			Str("template_creator_id", template.AdminID.String()).
			Msg("Unauthorized template lesson creation attempt: user is not the creator")
		return nil, fmt.Errorf("not authorized to modify this template: you are not the creator")
	}

	// Apply defaults
	req.ApplyDefaults()

	// Validate
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create template lesson entry
	entry := &models.TemplateLessonEntry{
		ID:          uuid.New(),
		TemplateID:  templateID,
		DayOfWeek:   req.DayOfWeek,
		StartTime:   req.StartTime,
		TeacherID:   req.TeacherID,
		LessonType:  *req.LessonType, // Set from request (with default applied in ApplyDefaults)
		MaxStudents: *req.MaxStudents,
		CreditsCost: *req.CreditsCost, // Default applied in ApplyDefaults
		Color:       *req.Color,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if req.EndTime != nil {
		entry.EndTime = *req.EndTime
	}

	if req.Subject != nil {
		entry.Subject = sql.NullString{String: *req.Subject, Valid: true}
	}

	if req.Description != nil {
		entry.Description = sql.NullString{String: *req.Description, Valid: true}
	}

	// Create lesson entry
	if err := s.templateLessonRepo.CreateTemplateLessonEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create template lesson: %w", err)
	}

	// Проверяем вместимость перед добавлением студентов
	if err := validateStudentCapacity(len(req.StudentIDs), *req.MaxStudents); err != nil {
		// Rollback: delete the created lesson
		s.templateLessonRepo.DeleteTemplateLessonEntry(ctx, entry.ID)
		return nil, err
	}

	// Add pre-assigned students
	for _, studentID := range req.StudentIDs {
		if err := s.templateLessonRepo.AddStudentToTemplateLessonEntry(ctx, entry.ID, studentID); err != nil {
			// Rollback: delete the created lesson
			s.templateLessonRepo.DeleteTemplateLessonEntry(ctx, entry.ID)
			return nil, fmt.Errorf("failed to add student to template lesson: %w", err)
		}
	}

	// Загружаем преподавателя для отображения имени
	teacher, err := s.userRepo.GetByID(ctx, req.TeacherID)
	if err == nil {
		entry.TeacherName = teacher.GetFullName()
	}

	// Загружаем студентов для возврата в ответе
	if len(req.StudentIDs) > 0 {
		students, err := s.templateLessonRepo.GetStudentsForTemplateLessonEntry(ctx, entry.ID)
		if err != nil {
			// Не критично - занятие создано, просто студенты не отобразятся в ответе
			entry.Students = []*models.TemplateLessonStudent{}
		} else {
			entry.Students = students
		}
	}

	// Audit log: template lesson creation
	log.Info().
		Str("admin_id", adminID.String()).
		Str("template_id", templateID.String()).
		Str("lesson_id", entry.ID.String()).
		Int("day_of_week", entry.DayOfWeek).
		Str("start_time", entry.StartTime).
		Int("students_count", len(req.StudentIDs)).
		Str("modification_type", "create_template_lesson").
		Msg("Template lesson successfully created")

	return entry, nil
}

// UpdateTemplateLesson updates an existing template lesson
func (s *TemplateService) UpdateTemplateLesson(ctx context.Context, adminID uuid.UUID, templateID uuid.UUID, lessonID uuid.UUID, updates map[string]interface{}) (*models.TemplateLessonEntry, error) {
	// Verify template exists
	template, err := s.templateRepo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Get user to check role
	user, err := s.userRepo.GetByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify ownership: only creator, admin, or methodologist can modify lessons
	if template.AdminID != adminID && !user.IsAdmin() && !user.IsMethodologist() {
		log.Warn().
			Str("admin_id", adminID.String()).
			Str("template_id", templateID.String()).
			Str("template_creator_id", template.AdminID.String()).
			Msg("Unauthorized template lesson update attempt: user is not the creator")
		return nil, fmt.Errorf("not authorized to modify this template: you are not the creator")
	}

	// Log incoming credits_cost value if provided
	if creditsCost, ok := updates["credits_cost"].(int); ok {
		log.Debug().
			Str("lesson_id", lessonID.String()).
			Int("credits_cost_input", creditsCost).
			Msg("UpdateTemplateLesson: received credits_cost value")
	}

	// Get existing lesson
	existingLesson, err := s.templateLessonRepo.GetTemplateLessonByID(ctx, lessonID)
	if err != nil {
		return nil, fmt.Errorf("template lesson not found: %w", err)
	}

	// Verify lesson belongs to this template
	if existingLesson.TemplateID != templateID {
		return nil, fmt.Errorf("lesson does not belong to this template")
	}

	// Apply updates
	if dayOfWeek, ok := updates["day_of_week"].(int); ok {
		existingLesson.DayOfWeek = dayOfWeek
	}
	if startTime, ok := updates["start_time"].(string); ok {
		existingLesson.StartTime = startTime
	}
	if endTime, ok := updates["end_time"].(string); ok {
		existingLesson.EndTime = endTime
	}
	if teacherID, ok := updates["teacher_id"].(uuid.UUID); ok {
		// Verify the new teacher exists and can be assigned
		teacher, err := s.userRepo.GetByID(ctx, teacherID)
		if err != nil {
			return nil, fmt.Errorf("teacher not found: %w", err)
		}

		// Check that user can be assigned as teacher (role must be teacher, admin, or methodologist)
		if !teacher.CanBeAssignedAsTeacher() {
			return nil, fmt.Errorf("user cannot be assigned as teacher (role: %s): %w", teacher.Role, models.ErrInvalidTeacherID)
		}

		existingLesson.TeacherID = teacherID
	}
	if maxStudents, ok := updates["max_students"].(int); ok {
		existingLesson.MaxStudents = maxStudents
	}
	if creditsCost, ok := updates["credits_cost"].(int); ok {
		existingLesson.CreditsCost = creditsCost
	}
	if color, ok := updates["color"].(string); ok {
		existingLesson.Color = color
	}
	if subject, ok := updates["subject"].(string); ok {
		existingLesson.Subject = sql.NullString{String: subject, Valid: true}
	}
	if description, ok := updates["description"].(string); ok {
		existingLesson.Description = sql.NullString{String: description, Valid: true}
	}

	// Validate updated lesson
	if err := existingLesson.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Log credits_cost before database update (in-memory state)
	log.Debug().
		Str("lesson_id", lessonID.String()).
		Int("credits_cost_before_update", existingLesson.CreditsCost).
		Msg("UpdateTemplateLesson: credits_cost value before database update")

	// Update lesson in database
	if err := s.templateLessonRepo.UpdateTemplateLessonEntry(ctx, existingLesson); err != nil {
		return nil, fmt.Errorf("failed to update template lesson: %w", err)
	}

	// Log credits_cost after database update
	log.Debug().
		Str("lesson_id", lessonID.String()).
		Int("credits_cost_after_update", existingLesson.CreditsCost).
		Msg("UpdateTemplateLesson: credits_cost value after database update")

	// Update student assignments if provided
	if studentIDs, ok := updates["student_ids"].([]uuid.UUID); ok {
		// Проверяем вместимость перед обновлением студентов
		if err := validateStudentCapacity(len(studentIDs), existingLesson.MaxStudents); err != nil {
			return nil, err
		}

		// Remove all existing students
		students, _ := s.templateLessonRepo.GetStudentsForTemplateLessonEntry(ctx, lessonID)
		for _, student := range students {
			s.templateLessonRepo.RemoveStudentFromTemplateLessonEntry(ctx, lessonID, student.StudentID)
		}

		// Add new students
		for _, studentID := range studentIDs {
			if err := s.templateLessonRepo.AddStudentToTemplateLessonEntry(ctx, lessonID, studentID); err != nil {
				return nil, fmt.Errorf("failed to update student assignments: %w", err)
			}
		}
	}

	// Загружаем преподавателя для отображения имени
	teacher, err := s.userRepo.GetByID(ctx, existingLesson.TeacherID)
	if err == nil {
		existingLesson.TeacherName = teacher.GetFullName()
	}

	// Загружаем студентов для возврата в ответе
	students, err := s.templateLessonRepo.GetStudentsForTemplateLessonEntry(ctx, lessonID)
	if err != nil {
		existingLesson.Students = []*models.TemplateLessonStudent{}
	} else {
		existingLesson.Students = students
	}

	// Audit log: template lesson update
	updateFields := make([]string, 0)
	if _, hasUpdate := updates["day_of_week"]; hasUpdate {
		updateFields = append(updateFields, "day_of_week")
	}
	if _, hasUpdate := updates["start_time"]; hasUpdate {
		updateFields = append(updateFields, "start_time")
	}
	if _, hasUpdate := updates["end_time"]; hasUpdate {
		updateFields = append(updateFields, "end_time")
	}
	if _, hasUpdate := updates["teacher_id"]; hasUpdate {
		updateFields = append(updateFields, "teacher_id")
	}
	if _, hasUpdate := updates["max_students"]; hasUpdate {
		updateFields = append(updateFields, "max_students")
	}
	if _, hasUpdate := updates["color"]; hasUpdate {
		updateFields = append(updateFields, "color")
	}
	if _, hasUpdate := updates["subject"]; hasUpdate {
		updateFields = append(updateFields, "subject")
	}
	if _, hasUpdate := updates["description"]; hasUpdate {
		updateFields = append(updateFields, "description")
	}
	if _, hasUpdate := updates["student_ids"]; hasUpdate {
		updateFields = append(updateFields, "student_ids")
	}
	if _, hasUpdate := updates["credits_cost"]; hasUpdate {
		updateFields = append(updateFields, "credits_cost")
	}

	log.Info().
		Str("admin_id", adminID.String()).
		Str("template_id", templateID.String()).
		Str("lesson_id", lessonID.String()).
		Strs("updated_fields", updateFields).
		Int("students_count", len(existingLesson.Students)).
		Int("credits_cost", existingLesson.CreditsCost).
		Str("modification_type", "update_template_lesson").
		Msg("Template lesson successfully updated")

	return existingLesson, nil
}

// DeleteTemplateLesson deletes a template lesson
func (s *TemplateService) DeleteTemplateLesson(ctx context.Context, adminID uuid.UUID, templateID uuid.UUID, lessonID uuid.UUID) error {
	// Verify template exists
	template, err := s.templateRepo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	// Get user to check role
	user, err := s.userRepo.GetByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify ownership: only creator, admin, or methodologist can delete lessons
	if template.AdminID != adminID && !user.IsAdmin() && !user.IsMethodologist() {
		log.Warn().
			Str("admin_id", adminID.String()).
			Str("template_id", templateID.String()).
			Str("template_creator_id", template.AdminID.String()).
			Msg("Unauthorized template lesson deletion attempt: user is not the creator")
		return fmt.Errorf("not authorized to modify this template: you are not the creator")
	}

	// Get existing lesson to verify it belongs to this template
	existingLesson, err := s.templateLessonRepo.GetTemplateLessonByID(ctx, lessonID)
	if err != nil {
		return fmt.Errorf("template lesson not found: %w", err)
	}

	// Verify lesson belongs to this template
	if existingLesson.TemplateID != templateID {
		return fmt.Errorf("lesson does not belong to this template")
	}

	// Delete lesson (cascade deletes template_lesson_students)
	if err := s.templateLessonRepo.DeleteTemplateLessonEntry(ctx, lessonID); err != nil {
		return fmt.Errorf("failed to delete template lesson: %w", err)
	}

	// Audit log: template lesson deletion
	log.Warn().
		Str("admin_id", adminID.String()).
		Str("template_id", templateID.String()).
		Str("lesson_id", lessonID.String()).
		Int("day_of_week", existingLesson.DayOfWeek).
		Str("start_time", existingLesson.StartTime).
		Str("modification_type", "delete_template_lesson").
		Msg("Template lesson successfully deleted")

	return nil
}

// ============================================================================
// Transaction Helper Methods
// ============================================================================

// createLessonInTx creates a lesson within a transaction (sqlx compatibility)
func (s *TemplateService) createLessonInTx(ctx context.Context, tx *sqlx.Tx, lesson *models.Lesson) (*models.Lesson, error) {
	query := `
		INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students,
		                     color, subject, applied_from_template, template_application_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, teacher_id, start_time, end_time, max_students, current_students,
		          color, subject, applied_from_template, template_application_id, created_at, updated_at
	`

	var created models.Lesson
	err := tx.QueryRowContext(ctx, query,
		lesson.ID,
		lesson.TeacherID,
		lesson.StartTime,
		lesson.EndTime,
		lesson.MaxStudents,
		lesson.CurrentStudents,
		lesson.Color,
		lesson.Subject,
		lesson.AppliedFromTemplate,
		lesson.TemplateApplicationID,
		lesson.CreatedAt,
		lesson.UpdatedAt,
	).Scan(
		&created.ID,
		&created.TeacherID,
		&created.StartTime,
		&created.EndTime,
		&created.MaxStudents,
		&created.CurrentStudents,
		&created.Color,
		&created.Subject,
		&created.AppliedFromTemplate,
		&created.TemplateApplicationID,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create lesson: %w", err)
	}

	// Поле deleted_at всегда NULL при создании занятия
	created.DeletedAt = sql.NullTime{Valid: false}

	return &created, nil
}

// updateLessonCurrentStudentsInTx updates the current_students count for a lesson within a transaction
func (s *TemplateService) updateLessonCurrentStudentsInTx(ctx context.Context, tx *sqlx.Tx, lessonID uuid.UUID, currentStudents int) error {
	query := `
		UPDATE lessons
		SET current_students = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err := tx.ExecContext(ctx, query, currentStudents, lessonID)
	if err != nil {
		return fmt.Errorf("failed to update lesson current_students: %w", err)
	}
	return nil
}

// createBookingInTx creates a booking within a transaction (sqlx compatibility)
func (s *TemplateService) createBookingInTx(ctx context.Context, tx *sqlx.Tx, booking *models.Booking) error {
	query := `
		INSERT INTO bookings (id, student_id, lesson_id, status, booked_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := tx.ExecContext(ctx, query,
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

// getBalanceInTx retrieves credit balance within a transaction
func (s *TemplateService) getBalanceInTx(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) (int, error) {
	query := `
		SELECT balance
		FROM credits
		WHERE user_id = $1
		FOR UPDATE
	`

	var balance int
	err := tx.GetContext(ctx, &balance, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get credit balance: %w", err)
	}

	return balance, nil
}

// deductCreditInTx deducts 1 credit from a student within a transaction
// CRITICAL: This function must be called within a serializable transaction
// to ensure atomicity of balance check + update + transaction record creation
func (s *TemplateService) deductCreditInTx(ctx context.Context, tx *sqlx.Tx, studentID uuid.UUID, performedBy uuid.UUID, bookingID uuid.UUID, reason string) error {
	const creditCost = 1 // Стоимость одного занятия в кредитах

	// Lock row and check balance in single query for atomicity
	var currentBalance int
	selectQuery := `
		SELECT balance
		FROM credits
		WHERE user_id = $1
		FOR UPDATE
	`

	err := tx.GetContext(ctx, &currentBalance, selectQuery, studentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("кредитный счёт для студента %s не найден", studentID)
		}
		return fmt.Errorf("ошибка получения баланса кредитов: %w", err)
	}

	if currentBalance < creditCost {
		return fmt.Errorf("недостаточно кредитов у студента %s (текущий баланс: %d, требуется: %d)", studentID, currentBalance, creditCost)
	}

	balanceAfter := currentBalance - creditCost

	// Update balance (row already locked by SELECT FOR UPDATE)
	updateQuery := `
		UPDATE credits
		SET balance = $1, updated_at = $2
		WHERE user_id = $3
	`

	result, err := tx.ExecContext(ctx, updateQuery, balanceAfter, time.Now(), studentID)
	if err != nil {
		return fmt.Errorf("ошибка обновления баланса кредитов: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("не удалось обновить баланс кредитов для студента %s", studentID)
	}

	// Create transaction record with POSITIVE amount (for deduct, amount shows how many credits were taken)
	// Note: operation_type='deduct' indicates the direction of the operation
	insertQuery := `
		INSERT INTO credit_transactions (id, user_id, amount, operation_type, reason, performed_by, booking_id, balance_before, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = tx.ExecContext(ctx, insertQuery,
		uuid.New(),
		studentID,
		creditCost, // Положительное значение - количество списанных кредитов
		models.OperationTypeDeduct,
		reason,
		performedBy,
		bookingID,
		currentBalance,
		balanceAfter,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("ошибка создания записи транзакции кредитов: %w", err)
	}

	return nil
}

// refundCreditInTx refunds 1 credit to a student within a transaction
// CRITICAL: This function must be called within a serializable transaction
// to ensure atomicity of balance update + transaction record creation
func (s *TemplateService) refundCreditInTx(ctx context.Context, tx *sqlx.Tx, studentID uuid.UUID, performedBy uuid.UUID, bookingID uuid.UUID, reason string) error {
	const creditRefund = 1 // Количество возвращаемых кредитов за одно занятие

	// Lock row and get current balance for atomicity
	var currentBalance int
	selectQuery := `
		SELECT balance
		FROM credits
		WHERE user_id = $1
		FOR UPDATE
	`

	err := tx.GetContext(ctx, &currentBalance, selectQuery, studentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("кредитный счёт для студента %s не найден", studentID)
		}
		return fmt.Errorf("ошибка получения баланса кредитов: %w", err)
	}

	balanceAfter := currentBalance + creditRefund

	// Update balance with explicit new value (not balance + 1) for consistency
	updateQuery := `
		UPDATE credits
		SET balance = $1, updated_at = $2
		WHERE user_id = $3
	`

	result, err := tx.ExecContext(ctx, updateQuery, balanceAfter, time.Now(), studentID)
	if err != nil {
		return fmt.Errorf("ошибка обновления баланса кредитов: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("не удалось обновить баланс кредитов для студента %s", studentID)
	}

	// Create transaction record with POSITIVE amount (for refund, amount shows how many credits were returned)
	// Note: operation_type='refund' indicates the direction of the operation
	insertQuery := `
		INSERT INTO credit_transactions (id, user_id, amount, operation_type, reason, performed_by, booking_id, balance_before, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = tx.ExecContext(ctx, insertQuery,
		uuid.New(),
		studentID,
		creditRefund, // Положительное значение - количество возвращённых кредитов
		models.OperationTypeRefund,
		reason,
		performedBy,
		bookingID,
		currentBalance,
		balanceAfter,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("ошибка создания записи транзакции кредитов: %w", err)
	}

	return nil
}

// getLessonsByTemplateApplicationInTx retrieves lessons from a template application within a transaction
func (s *TemplateService) getLessonsByTemplateApplicationInTx(ctx context.Context, tx *sqlx.Tx, applicationID uuid.UUID) ([]*models.Lesson, error) {
	query := `
		SELECT id, teacher_id, start_time, end_time,
		       max_students, current_students, applied_from_template,
		       template_application_id, created_at, updated_at, deleted_at
		FROM lessons
		WHERE template_application_id = $1 AND deleted_at IS NULL
		ORDER BY start_time ASC
	`

	var lessons []*models.Lesson
	err := tx.SelectContext(ctx, &lessons, query, applicationID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get lessons by template application: %w", err)
	}

	return lessons, nil
}

// getBookingsByLessonInTx retrieves bookings for a lesson within a transaction
func (s *TemplateService) getBookingsByLessonInTx(ctx context.Context, tx *sqlx.Tx, lessonID uuid.UUID) ([]*models.Booking, error) {
	query := `
		SELECT id, student_id, lesson_id, status, booked_at, cancelled_at, created_at, updated_at
		FROM bookings
		WHERE lesson_id = $1 AND status = 'active'
		ORDER BY booked_at ASC
	`

	var bookings []*models.Booking
	err := tx.SelectContext(ctx, &bookings, query, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings for lesson: %w", err)
	}

	return bookings, nil
}

// cancelBookingInTx cancels a booking within a transaction
func (s *TemplateService) cancelBookingInTx(ctx context.Context, tx *sqlx.Tx, bookingID uuid.UUID) error {
	query := `
		UPDATE bookings
		SET status = 'cancelled', cancelled_at = $1, updated_at = $2
		WHERE id = $3 AND status = 'active'
	`

	result, err := tx.ExecContext(ctx, query, time.Now(), time.Now(), bookingID)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("no active booking found with ID %s", bookingID)
	}

	return nil
}

// deleteLessonInTx soft deletes a lesson within a transaction
func (s *TemplateService) deleteLessonInTx(ctx context.Context, tx *sqlx.Tx, lessonID uuid.UUID) error {
	query := `
		UPDATE lessons
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := tx.ExecContext(ctx, query, time.Now(), time.Now(), lessonID)
	if err != nil {
		return fmt.Errorf("failed to delete lesson: %w", err)
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("lesson not found or already deleted: %s", lessonID)
	}

	return nil
}

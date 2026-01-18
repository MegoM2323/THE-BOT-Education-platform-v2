package service

import (
	"context"
	"fmt"

	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Mismatch описывает расхождение в счётчике студентов
type Mismatch struct {
	LessonID          uuid.UUID
	ExpectedCount     int
	ActualCount       int
	Discrepancy       int
	TeacherID         uuid.UUID
	LessonStartTime   string
	BookingsActiveSQL int
}

// VerificationService проверяет целостность счётчиков и данных в системе
type VerificationService struct {
	pool        *pgxpool.Pool
	lessonRepo  *repository.LessonRepository
	bookingRepo *repository.BookingRepository
}

// NewVerificationService создает новый VerificationService
func NewVerificationService(
	pool *pgxpool.Pool,
	lessonRepo *repository.LessonRepository,
	bookingRepo *repository.BookingRepository,
) *VerificationService {
	return &VerificationService{
		pool:        pool,
		lessonRepo:  lessonRepo,
		bookingRepo: bookingRepo,
	}
}

// CheckStudentCounterConsistency проверяет консистентность счётчика current_students для занятия
// Сравнивает текущее значение со счётом активных бронирований
// Возвращает true если счётчик корректен, false если есть расхождения
func (s *VerificationService) CheckStudentCounterConsistency(ctx context.Context, lessonID uuid.UUID) (bool, error) {
	query := `
		SELECT
			l.current_students,
			COUNT(b.id) as active_bookings
		FROM lessons l
		LEFT JOIN bookings b ON l.id = b.lesson_id AND b.status = 'active' AND b.deleted_at IS NULL
		WHERE l.id = $1
		GROUP BY l.id, l.current_students
	`

	var expectedCount int
	var actualCount int

	err := s.pool.QueryRow(ctx, query, lessonID).Scan(&expectedCount, &actualCount)
	if err != nil {
		return false, fmt.Errorf("failed to check counter consistency: %w", err)
	}

	// Счётчик консистентен если значение current_students совпадает со счётом активных бронирований
	return expectedCount == actualCount, nil
}

// AuditCounterMismatches проверяет все занятия на предмет расхождений в счётчиках
// Возвращает список всех найденных расхождений
func (s *VerificationService) AuditCounterMismatches(ctx context.Context) ([]Mismatch, error) {
	query := `
		SELECT
			l.id,
			l.current_students,
			COALESCE(COUNT(b.id), 0) as active_bookings,
			l.teacher_id,
			l.start_time::text,
			l.current_students - COALESCE(COUNT(b.id), 0) as discrepancy
		FROM lessons l
		LEFT JOIN bookings b ON l.id = b.lesson_id AND b.status = 'active' AND b.deleted_at IS NULL
		WHERE l.deleted_at IS NULL
		GROUP BY l.id, l.current_students, l.teacher_id, l.start_time
		HAVING l.current_students != COALESCE(COUNT(b.id), 0)
		ORDER BY ABS(l.current_students - COALESCE(COUNT(b.id), 0)) DESC
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to audit counter mismatches: %w", err)
	}
	defer rows.Close()

	var mismatches []Mismatch
	for rows.Next() {
		var m Mismatch
		if err := rows.Scan(
			&m.LessonID,
			&m.ExpectedCount,
			&m.ActualCount,
			&m.TeacherID,
			&m.LessonStartTime,
			&m.Discrepancy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan mismatch row: %w", err)
		}
		m.BookingsActiveSQL = m.ActualCount
		mismatches = append(mismatches, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading mismatch rows: %w", err)
	}

	return mismatches, nil
}

// RepairLessonCounter восстанавливает правильность счётчика current_students для занятия
// Пересчитывает счётчик на основе COUNT активных бронирований
// Выполняется атомарно в транзакции
func (s *VerificationService) RepairLessonCounter(ctx context.Context, lessonID uuid.UUID) error {
	txOptions := pgx.TxOptions{}
	tx, err := s.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return fmt.Errorf("failed to begin repair transaction: %w", err)
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
			log.Warn().
				Err(rollbackErr).
				Str("lesson_id", lessonID.String()).
				Msg("Failed to rollback repair transaction")
		}
	}()

	// Получаем текущий счёт активных бронирований
	var correctCount int
	countQuery := `
		SELECT COUNT(*) FROM bookings
		WHERE lesson_id = $1 AND status = 'active' AND deleted_at IS NULL
	`
	err = tx.QueryRow(ctx, countQuery, lessonID).Scan(&correctCount)
	if err != nil {
		return fmt.Errorf("failed to count active bookings: %w", err)
	}

	// Обновляем счётчик
	updateQuery := `
		UPDATE lessons
		SET current_students = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	result, err := tx.Exec(ctx, updateQuery, correctCount, lessonID)
	if err != nil {
		return fmt.Errorf("failed to update counter: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("lesson not found or already deleted")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit repair transaction: %w", err)
	}

	log.Info().
		Str("lesson_id", lessonID.String()).
		Int("new_count", correctCount).
		Msg("Lesson counter repaired successfully")

	return nil
}

// RepairAllCounters восстанавливает счётчики для всех занятий с расхождениями
// Возвращает количество исправленных занятий
func (s *VerificationService) RepairAllCounters(ctx context.Context) (int, error) {
	// Получаем список занятий с расхождениями
	mismatches, err := s.AuditCounterMismatches(ctx)
	if err != nil {
		return 0, err
	}

	if len(mismatches) == 0 {
		log.Info().Msg("No counter mismatches found")
		return 0, nil
	}

	// Восстанавливаем каждое занятие
	repaired := 0
	for _, mismatch := range mismatches {
		if err := s.RepairLessonCounter(ctx, mismatch.LessonID); err != nil {
			log.Warn().
				Err(err).
				Str("lesson_id", mismatch.LessonID.String()).
				Msg("Failed to repair lesson counter")
			continue
		}
		repaired++
	}

	log.Info().
		Int("total_mismatches", len(mismatches)).
		Int("repaired", repaired).
		Msg("Counter repair completed")

	return repaired, nil
}

// GetCounterMetrics получает статистику по счётчикам
type CounterMetrics struct {
	TotalLessons      int
	MismatchedLessons int
	MaxDiscrepancy    int
	AvgDiscrepancy    float64
}

// GetCounterMetrics возвращает метрики по состоянию счётчиков
func (s *VerificationService) GetCounterMetrics(ctx context.Context) (*CounterMetrics, error) {
	query := `
		SELECT
			COUNT(*) as total_lessons,
			SUM(CASE WHEN current_students != active_count THEN 1 ELSE 0 END) as mismatched,
			MAX(ABS(current_students - active_count)) as max_discrepancy,
			AVG(ABS(current_students - active_count)) as avg_discrepancy
		FROM (
			SELECT
				l.id,
				l.current_students,
				COALESCE(COUNT(b.id), 0) as active_count
			FROM lessons l
			LEFT JOIN bookings b ON l.id = b.lesson_id AND b.status = 'active' AND b.deleted_at IS NULL
			WHERE l.deleted_at IS NULL
			GROUP BY l.id, l.current_students
		) counts
	`

	var metrics CounterMetrics
	var avgDiscrepancy *float64

	err := s.pool.QueryRow(ctx, query).Scan(
		&metrics.TotalLessons,
		&metrics.MismatchedLessons,
		&metrics.MaxDiscrepancy,
		&avgDiscrepancy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get counter metrics: %w", err)
	}

	if avgDiscrepancy != nil {
		metrics.AvgDiscrepancy = *avgDiscrepancy
	}

	return &metrics, nil
}

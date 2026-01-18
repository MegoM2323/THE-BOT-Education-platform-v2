package repository

import (
	"context"
	"fmt"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

// SwapRepository управляет операциями с базой данных для обменов
type SwapRepository struct {
	db *sqlx.DB
}

// NewSwapRepository создает новый SwapRepository
func NewSwapRepository(db *sqlx.DB) *SwapRepository {
	return &SwapRepository{db: db}
}

// Create создает новую запись обмена в рамках транзакции
func (r *SwapRepository) Create(ctx context.Context, tx pgx.Tx, swap *models.Swap) error {
	query := `
		INSERT INTO swaps (id, student_id, old_lesson_id, new_lesson_id, old_booking_id, new_booking_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	swap.ID = uuid.New()
	swap.CreatedAt = time.Now()

	_, err := tx.Exec(ctx, query,
		swap.ID,
		swap.StudentID,
		swap.OldLessonID,
		swap.NewLessonID,
		swap.OldBookingID,
		swap.NewBookingID,
		swap.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create swap: %w", err)
	}

	return nil
}

// GetByID получает обмен по ID
func (r *SwapRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Swap, error) {
	query := `
		SELECT ` + SwapSelectFields + `
		FROM swaps
		WHERE id = $1
	`

	var swap models.Swap
	err := r.db.GetContext(ctx, &swap, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap by ID: %w", err)
	}

	return &swap, nil
}

// GetWithDetails получает обмен с деталями занятий
func (r *SwapRepository) GetWithDetails(ctx context.Context, id uuid.UUID) (*models.SwapWithDetails, error) {
	query := `
		SELECT
			s.id, s.student_id, s.old_lesson_id, s.new_lesson_id,
			s.old_booking_id, s.new_booking_id, s.created_at,
			ol.start_time as old_lesson_start_time,
			ol.end_time as old_lesson_end_time,
			nl.start_time as new_lesson_start_time,
			nl.end_time as new_lesson_end_time,
			u.full_name as student_name
		FROM swaps s
		JOIN lessons ol ON s.old_lesson_id = ol.id
		JOIN lessons nl ON s.new_lesson_id = nl.id
		JOIN users u ON s.student_id = u.id
		WHERE s.id = $1
	`

	var swap models.SwapWithDetails
	err := r.db.GetContext(ctx, &swap, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get swap with details: %w", err)
	}

	return &swap, nil
}

// List получает историю обменов с опциональными фильтрами
func (r *SwapRepository) List(ctx context.Context, filter *models.GetSwapHistoryFilter) ([]*models.SwapWithDetails, error) {
	query := `
		SELECT
			s.id, s.student_id, s.old_lesson_id, s.new_lesson_id,
			s.old_booking_id, s.new_booking_id, s.created_at,
			ol.start_time as old_lesson_start_time,
			ol.end_time as old_lesson_end_time,
			nl.start_time as new_lesson_start_time,
			nl.end_time as new_lesson_end_time,
			u.full_name as student_name
		FROM swaps s
		JOIN lessons ol ON s.old_lesson_id = ol.id
		JOIN lessons nl ON s.new_lesson_id = nl.id
		JOIN users u ON s.student_id = u.id
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if filter != nil {
		if filter.StudentID != nil {
			query += fmt.Sprintf(` AND s.student_id = $%d`, argIndex)
			args = append(args, *filter.StudentID)
			argIndex++
		}
		if filter.StartDate != nil {
			query += fmt.Sprintf(` AND s.created_at >= $%d`, argIndex)
			args = append(args, *filter.StartDate)
			argIndex++
		}
		if filter.EndDate != nil {
			query += fmt.Sprintf(` AND s.created_at <= $%d`, argIndex)
			args = append(args, *filter.EndDate)
			argIndex++
		}
	}

	query += ` ORDER BY s.created_at DESC`

	if filter != nil {
		if filter.Limit > 0 {
			query += fmt.Sprintf(` LIMIT $%d`, argIndex)
			args = append(args, filter.Limit)
			argIndex++
		}
		if filter.Offset > 0 {
			query += fmt.Sprintf(` OFFSET $%d`, argIndex)
			args = append(args, filter.Offset)
			argIndex++
		}
	}

	var swaps []*models.SwapWithDetails
	err := r.db.SelectContext(ctx, &swaps, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list swaps: %w", err)
	}

	return swaps, nil
}

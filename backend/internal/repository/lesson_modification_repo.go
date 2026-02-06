package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

// LessonModificationRepository handles lesson modification audit trail
type LessonModificationRepository struct {
	db *sqlx.DB
}

// NewLessonModificationRepository creates a new LessonModificationRepository
func NewLessonModificationRepository(db *sqlx.DB) *LessonModificationRepository {
	return &LessonModificationRepository{db: db}
}

// LogModification creates a new modification audit entry
func (r *LessonModificationRepository) LogModification(ctx context.Context, modification *models.LessonModification) error {
	query := `
		INSERT INTO lesson_modifications
		(id, original_lesson_id, modification_type, applied_by_id, applied_at, affected_lessons_count, changes_json, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, original_lesson_id, modification_type, applied_by_id, applied_at, affected_lessons_count, changes_json, notes
	`

	modification.ID = uuid.New()
	modification.AppliedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		modification.ID,
		modification.OriginalLessonID,
		modification.ModificationType,
		modification.AppliedByID,
		modification.AppliedAt,
		modification.AffectedLessonsCount,
		modification.ChangesJSON,
		modification.Notes,
	).Scan(
		&modification.ID,
		&modification.OriginalLessonID,
		&modification.ModificationType,
		&modification.AppliedByID,
		&modification.AppliedAt,
		&modification.AffectedLessonsCount,
		&modification.ChangesJSON,
		&modification.Notes,
	)

	if err != nil {
		return fmt.Errorf("failed to log modification: %w", err)
	}

	return nil
}

// GetModificationsByLesson retrieves all modifications for a specific lesson
func (r *LessonModificationRepository) GetModificationsByLesson(ctx context.Context, lessonID uuid.UUID) ([]*models.LessonModification, error) {
	query := `
		SELECT
			lm.id, lm.original_lesson_id, lm.modification_type, lm.applied_by_id,
			lm.applied_at, lm.affected_lessons_count, lm.changes_json, lm.notes,
			COALESCE(NULLIF(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''), u.email) as applied_by_name
		FROM lesson_modifications lm
		JOIN users u ON lm.applied_by_id = u.id
		WHERE lm.original_lesson_id = $1
		ORDER BY lm.applied_at DESC
	`

	var modifications []*models.LessonModification
	err := r.db.SelectContext(ctx, &modifications, query, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get modifications by lesson: %w", err)
	}

	return modifications, nil
}

// GetModificationsByType retrieves modifications filtered by type
func (r *LessonModificationRepository) GetModificationsByType(ctx context.Context, modificationType string) ([]*models.LessonModification, error) {
	query := `
		SELECT
			lm.id, lm.original_lesson_id, lm.modification_type, lm.applied_by_id,
			lm.applied_at, lm.affected_lessons_count, lm.changes_json, lm.notes,
			COALESCE(NULLIF(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''), u.email) as applied_by_name
		FROM lesson_modifications lm
		JOIN users u ON lm.applied_by_id = u.id
		WHERE lm.modification_type = $1
		ORDER BY lm.applied_at DESC
	`

	var modifications []*models.LessonModification
	err := r.db.SelectContext(ctx, &modifications, query, modificationType)
	if err != nil {
		return nil, fmt.Errorf("failed to get modifications by type: %w", err)
	}

	return modifications, nil
}

// GetRecentModifications retrieves recent modifications with limit
func (r *LessonModificationRepository) GetRecentModifications(ctx context.Context, limit int) ([]*models.LessonModification, error) {
	query := `
		SELECT
			lm.id, lm.original_lesson_id, lm.modification_type, lm.applied_by_id,
			lm.applied_at, lm.affected_lessons_count, lm.changes_json, lm.notes,
			COALESCE(NULLIF(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''), u.email) as applied_by_name
		FROM lesson_modifications lm
		JOIN users u ON lm.applied_by_id = u.id
		ORDER BY lm.applied_at DESC
		LIMIT $1
	`

	var modifications []*models.LessonModification
	err := r.db.SelectContext(ctx, &modifications, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent modifications: %w", err)
	}

	return modifications, nil
}

// GetModificationByID retrieves a specific modification by ID
func (r *LessonModificationRepository) GetModificationByID(ctx context.Context, id uuid.UUID) (*models.LessonModification, error) {
	query := `
		SELECT
			lm.id, lm.original_lesson_id, lm.modification_type, lm.applied_by_id,
			lm.applied_at, lm.affected_lessons_count, lm.changes_json, lm.notes,
			COALESCE(NULLIF(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''), u.email) as applied_by_name
		FROM lesson_modifications lm
		JOIN users u ON lm.applied_by_id = u.id
		WHERE lm.id = $1
	`

	var modification models.LessonModification
	err := r.db.GetContext(ctx, &modification, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("modification not found")
		}
		return nil, fmt.Errorf("failed to get modification: %w", err)
	}

	return &modification, nil
}

// GetModificationsByDateRange retrieves modifications within a date range
func (r *LessonModificationRepository) GetModificationsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*models.LessonModification, error) {
	query := `
		SELECT
			lm.id, lm.original_lesson_id, lm.modification_type, lm.applied_by_id,
			lm.applied_at, lm.affected_lessons_count, lm.changes_json, lm.notes,
			COALESCE(NULLIF(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''), u.email) as applied_by_name
		FROM lesson_modifications lm
		JOIN users u ON lm.applied_by_id = u.id
		WHERE lm.applied_at >= $1 AND lm.applied_at <= $2
		ORDER BY lm.applied_at DESC
	`

	var modifications []*models.LessonModification
	err := r.db.SelectContext(ctx, &modifications, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get modifications by date range: %w", err)
	}

	return modifications, nil
}

// GetModificationsByAdmin retrieves all modifications performed by a specific admin
func (r *LessonModificationRepository) GetModificationsByAdmin(ctx context.Context, adminID uuid.UUID, limit, offset int) ([]*models.LessonModification, error) {
	query := `
		SELECT
			lm.id, lm.original_lesson_id, lm.modification_type, lm.applied_by_id,
			lm.applied_at, lm.affected_lessons_count, lm.changes_json, lm.notes,
			COALESCE(NULLIF(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''), u.email) as applied_by_name
		FROM lesson_modifications lm
		JOIN users u ON lm.applied_by_id = u.id
		WHERE lm.applied_by_id = $1
		ORDER BY lm.applied_at DESC
		LIMIT $2 OFFSET $3
	`

	var modifications []*models.LessonModification
	err := r.db.SelectContext(ctx, &modifications, query, adminID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get modifications by admin: %w", err)
	}

	return modifications, nil
}

// CountModificationsByAdmin counts total modifications by an admin
func (r *LessonModificationRepository) CountModificationsByAdmin(ctx context.Context, adminID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM lesson_modifications
		WHERE applied_by_id = $1
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, adminID)
	if err != nil {
		return 0, fmt.Errorf("failed to count modifications: %w", err)
	}

	return count, nil
}

// ModificationChange represents the structure of changes stored in changes_json
type ModificationChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// CreateModificationChangesJSON creates a JSON blob for modification changes
func CreateModificationChangesJSON(changes []ModificationChange) (json.RawMessage, error) {
	jsonData, err := json.Marshal(changes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal changes: %w", err)
	}
	return json.RawMessage(jsonData), nil
}

// ParseModificationChangesJSON parses the changes_json field
func ParseModificationChangesJSON(changesJSON json.RawMessage) ([]ModificationChange, error) {
	var changes []ModificationChange
	err := json.Unmarshal(changesJSON, &changes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
	}
	return changes, nil
}

// LogModificationTx creates a new modification audit entry within a transaction
func (r *LessonModificationRepository) LogModificationTx(ctx context.Context, tx pgx.Tx, modification *models.LessonModification) error {
	query := `
		INSERT INTO lesson_modifications
		(id, original_lesson_id, modification_type, applied_by_id, applied_at, affected_lessons_count, changes_json, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	modification.ID = uuid.New()
	modification.AppliedAt = time.Now()

	err := tx.QueryRow(ctx, query,
		modification.ID,
		modification.OriginalLessonID,
		modification.ModificationType,
		modification.AppliedByID,
		modification.AppliedAt,
		modification.AffectedLessonsCount,
		modification.ChangesJSON,
		modification.Notes,
	).Scan(&modification.ID)

	if err != nil {
		return fmt.Errorf("failed to log modification in transaction: %w", err)
	}

	return nil
}

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

// BroadcastListRepository интерфейс для работы со списками рассылки
type BroadcastListRepository interface {
	Create(ctx context.Context, list *models.BroadcastList) (*models.BroadcastList, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.BroadcastList, error)
	GetAll(ctx context.Context) ([]*models.BroadcastList, error)
	Update(ctx context.Context, id uuid.UUID, name, description string, userIDs []uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetUserCount(ctx context.Context, listID uuid.UUID) (int, error)
}

// BroadcastListRepo реализация BroadcastListRepository
type BroadcastListRepo struct {
	db *sqlx.DB
}

// NewBroadcastListRepository создает новый экземпляр BroadcastListRepo
func NewBroadcastListRepository(db *sqlx.DB) BroadcastListRepository {
	return &BroadcastListRepo{db: db}
}

// Create создает новый список рассылки
func (r *BroadcastListRepo) Create(ctx context.Context, list *models.BroadcastList) (*models.BroadcastList, error) {
	query := `
		INSERT INTO broadcast_lists (id, name, description, user_ids, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, description, user_ids, created_by, created_at, updated_at
	`

	list.ID = uuid.New()
	now := time.Now()
	list.CreatedAt = now
	list.UpdatedAt = now

	err := r.db.GetContext(ctx, list, query,
		list.ID,
		list.Name,
		list.Description,
		list.UserIDs,
		list.CreatedBy,
		list.CreatedAt,
		list.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create broadcast list: %w", err)
	}

	return list, nil
}

// GetByID получает список рассылки по ID
func (r *BroadcastListRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.BroadcastList, error) {
	query := `
		SELECT id, name, description, user_ids, created_by, created_at, updated_at, deleted_at
		FROM broadcast_lists
		WHERE id = $1 AND deleted_at IS NULL
	`

	var list models.BroadcastList
	err := r.db.GetContext(ctx, &list, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBroadcastListNotFound
		}
		return nil, fmt.Errorf("failed to get broadcast list by ID: %w", err)
	}

	return &list, nil
}

// GetAll получает все не удаленные списки рассылки
func (r *BroadcastListRepo) GetAll(ctx context.Context) ([]*models.BroadcastList, error) {
	query := `
		SELECT id, name, description, user_ids, created_by, created_at, updated_at, deleted_at
		FROM broadcast_lists
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	var lists []*models.BroadcastList
	err := r.db.SelectContext(ctx, &lists, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all broadcast lists: %w", err)
	}

	return lists, nil
}

// Update обновляет список рассылки
func (r *BroadcastListRepo) Update(ctx context.Context, id uuid.UUID, name, description string, userIDs []uuid.UUID) error {
	query := `
		UPDATE broadcast_lists
		SET name = $1, description = $2, user_ids = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		name,
		description,
		models.UUIDArray(userIDs),
		time.Now(),
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update broadcast list: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrBroadcastListNotFound
	}

	return nil
}

// Delete выполняет мягкое удаление списка рассылки
func (r *BroadcastListRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE broadcast_lists
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete broadcast list: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrBroadcastListNotFound
	}

	return nil
}

// GetUserCount возвращает количество пользователей в списке рассылки
func (r *BroadcastListRepo) GetUserCount(ctx context.Context, listID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(array_length(user_ids, 1), 0) as count
		FROM broadcast_lists
		WHERE id = $1 AND deleted_at IS NULL
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, listID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrBroadcastListNotFound
		}
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}

	return count, nil
}

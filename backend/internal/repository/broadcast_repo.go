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

// BroadcastRepository интерфейс для работы с массовыми рассылками
type BroadcastRepository interface {
	Create(ctx context.Context, broadcast *models.Broadcast) (*models.Broadcast, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Broadcast, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.Broadcast, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateCounts(ctx context.Context, id uuid.UUID, sentCount, failedCount int) error
	CreateLog(ctx context.Context, log *models.BroadcastLog) error
	GetLogsByBroadcastID(ctx context.Context, broadcastID uuid.UUID) ([]*models.BroadcastLog, error)
	// Idempotency methods for delivery tracking
	GetLogByBroadcastAndUser(ctx context.Context, broadcastID, userID uuid.UUID) (*models.BroadcastLog, error)
	HasSuccessfulDelivery(ctx context.Context, broadcastID, userID uuid.UUID) (bool, error)
	UpdateLogStatus(ctx context.Context, logID uuid.UUID, status, errorMsg string) error
}

// BroadcastRepo реализация BroadcastRepository
type BroadcastRepo struct {
	db *sqlx.DB
}

// NewBroadcastRepository создает новый экземпляр BroadcastRepo
func NewBroadcastRepository(db *sqlx.DB) BroadcastRepository {
	return &BroadcastRepo{db: db}
}

// Create создает новую рассылку
func (r *BroadcastRepo) Create(ctx context.Context, broadcast *models.Broadcast) (*models.Broadcast, error) {
	query := `
		INSERT INTO broadcasts (id, list_id, message, sent_count, failed_count, status, created_by, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, list_id, message, sent_count, failed_count, status, created_by, created_at, completed_at
	`

	broadcast.ID = uuid.New()
	broadcast.CreatedAt = time.Now()
	broadcast.SentCount = 0
	broadcast.FailedCount = 0
	if broadcast.Status == "" {
		broadcast.Status = models.BroadcastStatusPending
	}

	err := r.db.GetContext(ctx, broadcast, query,
		broadcast.ID,
		broadcast.ListID,
		broadcast.Message,
		broadcast.SentCount,
		broadcast.FailedCount,
		broadcast.Status,
		broadcast.CreatedBy,
		broadcast.CreatedAt,
		broadcast.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create broadcast: %w", err)
	}

	return broadcast, nil
}

// GetByID получает рассылку по ID
func (r *BroadcastRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Broadcast, error) {
	query := `
		SELECT id, list_id, message, sent_count, failed_count, status, created_by, created_at, completed_at
		FROM broadcasts
		WHERE id = $1
	`

	var broadcast models.Broadcast
	err := r.db.GetContext(ctx, &broadcast, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBroadcastNotFound
		}
		return nil, fmt.Errorf("failed to get broadcast by ID: %w", err)
	}

	return &broadcast, nil
}

// GetAll получает все рассылки с пагинацией
func (r *BroadcastRepo) GetAll(ctx context.Context, limit, offset int) ([]*models.Broadcast, int, error) {
	// Получаем общее количество рассылок
	countQuery := `SELECT COUNT(*) FROM broadcasts`
	var total int
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count broadcasts: %w", err)
	}

	// Получаем список рассылок с пагинацией
	query := `
		SELECT id, list_id, message, sent_count, failed_count, status, created_by, created_at, completed_at
		FROM broadcasts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var broadcasts []*models.Broadcast
	err = r.db.SelectContext(ctx, &broadcasts, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get broadcasts: %w", err)
	}

	return broadcasts, total, nil
}

// UpdateStatus обновляет статус рассылки
func (r *BroadcastRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	var query string
	var args []interface{}

	// Если статус финальный, устанавливаем completed_at
	if status == models.BroadcastStatusCompleted ||
		status == models.BroadcastStatusFailed ||
		status == models.BroadcastStatusCancelled {
		query = `
			UPDATE broadcasts
			SET status = $1, completed_at = $2
			WHERE id = $3
		`
		args = []interface{}{status, time.Now(), id}
	} else {
		query = `
			UPDATE broadcasts
			SET status = $1
			WHERE id = $2
		`
		args = []interface{}{status, id}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update broadcast status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrBroadcastNotFound
	}

	return nil
}

// UpdateCounts обновляет счетчики отправленных и неудавшихся сообщений
func (r *BroadcastRepo) UpdateCounts(ctx context.Context, id uuid.UUID, sentCount, failedCount int) error {
	query := `
		UPDATE broadcasts
		SET sent_count = $1, failed_count = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, sentCount, failedCount, id)
	if err != nil {
		return fmt.Errorf("failed to update broadcast counts: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrBroadcastNotFound
	}

	return nil
}

// CreateLog создает запись в логе отправки
func (r *BroadcastRepo) CreateLog(ctx context.Context, log *models.BroadcastLog) error {
	query := `
		INSERT INTO broadcast_logs (id, broadcast_id, user_id, telegram_id, status, error, sent_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	log.ID = uuid.New()
	log.SentAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.BroadcastID,
		log.UserID,
		log.TelegramID,
		log.Status,
		log.Error,
		log.SentAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create broadcast log: %w", err)
	}

	return nil
}

// GetLogsByBroadcastID получает логи отправки для конкретной рассылки
func (r *BroadcastRepo) GetLogsByBroadcastID(ctx context.Context, broadcastID uuid.UUID) ([]*models.BroadcastLog, error) {
	query := `
		SELECT
			bl.id,
			bl.broadcast_id,
			bl.user_id,
			bl.telegram_id,
			bl.status,
			bl.error,
			bl.sent_at
		FROM broadcast_logs bl
		WHERE bl.broadcast_id = $1
		ORDER BY bl.sent_at DESC
	`

	var logs []*models.BroadcastLog
	err := r.db.SelectContext(ctx, &logs, query, broadcastID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast logs: %w", err)
	}

	return logs, nil
}

// GetLogByBroadcastAndUser получает лог отправки конкретному пользователю в рассылке
// Используется для проверки идемпотентности (не отправлять дважды одно сообщение)
func (r *BroadcastRepo) GetLogByBroadcastAndUser(ctx context.Context, broadcastID, userID uuid.UUID) (*models.BroadcastLog, error) {
	query := `
		SELECT
			bl.id,
			bl.broadcast_id,
			bl.user_id,
			bl.telegram_id,
			bl.status,
			bl.error,
			bl.sent_at
		FROM broadcast_logs bl
		WHERE bl.broadcast_id = $1 AND bl.user_id = $2
		LIMIT 1
	`

	var log models.BroadcastLog
	err := r.db.GetContext(ctx, &log, query, broadcastID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No log exists yet
		}
		return nil, fmt.Errorf("failed to get broadcast log for user: %w", err)
	}

	return &log, nil
}

// HasSuccessfulDelivery проверяет, успешно ли доставлено сообщение пользователю
// Используется для идемпотентности: если сообщение уже доставлено, не отправляем его снова
func (r *BroadcastRepo) HasSuccessfulDelivery(ctx context.Context, broadcastID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM broadcast_logs
			WHERE broadcast_id = $1 AND user_id = $2 AND status = $3
		)
	`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, broadcastID, userID, models.BroadcastLogStatusSuccess)
	if err != nil {
		return false, fmt.Errorf("failed to check delivery status: %w", err)
	}

	return exists, nil
}

// UpdateLogStatus обновляет статус записи логирования (для обновления при повторной попытке)
func (r *BroadcastRepo) UpdateLogStatus(ctx context.Context, logID uuid.UUID, status, errorMsg string) error {
	query := `
		UPDATE broadcast_logs
		SET status = $1, error = $2, sent_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, status, errorMsg, time.Now(), logID)
	if err != nil {
		return fmt.Errorf("failed to update log status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("log entry not found: %s", logID)
	}

	return nil
}

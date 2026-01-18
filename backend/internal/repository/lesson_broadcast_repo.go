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

// LessonBroadcastRepository интерфейс для работы с рассылками по урокам
type LessonBroadcastRepository interface {
	CreateBroadcast(ctx context.Context, broadcast *models.LessonBroadcast) (*models.LessonBroadcast, error)
	GetBroadcast(ctx context.Context, broadcastID uuid.UUID) (*models.LessonBroadcast, error)
	ListBroadcastsByLesson(ctx context.Context, lessonID uuid.UUID) ([]*models.LessonBroadcast, error)
	UpdateBroadcastStatus(ctx context.Context, broadcastID uuid.UUID, status string, sentCount, failedCount int) error
	AddBroadcastFile(ctx context.Context, file *models.BroadcastFile) error
	GetBroadcastFiles(ctx context.Context, broadcastID uuid.UUID) ([]*models.BroadcastFile, error)
	GetBroadcastFile(ctx context.Context, fileID uuid.UUID) (*models.BroadcastFile, error)
}

// LessonBroadcastRepo реализация LessonBroadcastRepository
type LessonBroadcastRepo struct {
	db *sqlx.DB
}

// NewLessonBroadcastRepository создает новый экземпляр LessonBroadcastRepo
func NewLessonBroadcastRepository(db *sqlx.DB) LessonBroadcastRepository {
	return &LessonBroadcastRepo{db: db}
}

// CreateBroadcast создает новую рассылку урока
func (r *LessonBroadcastRepo) CreateBroadcast(ctx context.Context, broadcast *models.LessonBroadcast) (*models.LessonBroadcast, error) {
	query := `
		INSERT INTO lesson_broadcasts (id, lesson_id, sender_id, message, status, sent_count, failed_count, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, lesson_id, sender_id, message, status, sent_count, failed_count, created_at, completed_at
	`

	broadcast.ID = uuid.New()
	broadcast.CreatedAt = time.Now()
	broadcast.SentCount = 0
	broadcast.FailedCount = 0
	if broadcast.Status == "" {
		broadcast.Status = models.LessonBroadcastStatusPending
	}

	err := r.db.GetContext(ctx, broadcast, query,
		broadcast.ID,
		broadcast.LessonID,
		broadcast.SenderID,
		broadcast.Message,
		broadcast.Status,
		broadcast.SentCount,
		broadcast.FailedCount,
		broadcast.CreatedAt,
		broadcast.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create lesson broadcast: %w", err)
	}

	return broadcast, nil
}

// GetBroadcast получает рассылку по ID
func (r *LessonBroadcastRepo) GetBroadcast(ctx context.Context, broadcastID uuid.UUID) (*models.LessonBroadcast, error) {
	query := `
		SELECT
			lb.id,
			lb.lesson_id,
			lb.sender_id,
			lb.message,
			lb.status,
			lb.sent_count,
			lb.failed_count,
			lb.created_at,
			lb.completed_at,
			COALESCE(u.full_name, 'Unknown') as sender_name
		FROM lesson_broadcasts lb
		LEFT JOIN users u ON lb.sender_id = u.id
		WHERE lb.id = $1
	`

	var broadcast models.LessonBroadcast
	err := r.db.GetContext(ctx, &broadcast, query, broadcastID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrLessonBroadcastNotFound
		}
		return nil, fmt.Errorf("failed to get lesson broadcast by ID: %w", err)
	}

	// Загружаем файлы рассылки
	files, err := r.GetBroadcastFiles(ctx, broadcastID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast files: %w", err)
	}
	broadcast.Files = files

	return &broadcast, nil
}

// ListBroadcastsByLesson получает все рассылки для конкретного урока
func (r *LessonBroadcastRepo) ListBroadcastsByLesson(ctx context.Context, lessonID uuid.UUID) ([]*models.LessonBroadcast, error) {
	query := `
		SELECT
			lb.id,
			lb.lesson_id,
			lb.sender_id,
			lb.message,
			lb.status,
			lb.sent_count,
			lb.failed_count,
			lb.created_at,
			lb.completed_at,
			COALESCE(u.full_name, 'Unknown') as sender_name
		FROM lesson_broadcasts lb
		LEFT JOIN users u ON lb.sender_id = u.id
		WHERE lb.lesson_id = $1
		ORDER BY lb.created_at DESC
	`

	var broadcasts []*models.LessonBroadcast
	err := r.db.SelectContext(ctx, &broadcasts, query, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to list broadcasts by lesson: %w", err)
	}

	// Загружаем файлы для каждой рассылки
	for _, broadcast := range broadcasts {
		files, err := r.GetBroadcastFiles(ctx, broadcast.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get broadcast files for %s: %w", broadcast.ID, err)
		}
		broadcast.Files = files
	}

	return broadcasts, nil
}

// UpdateBroadcastStatus обновляет статус рассылки и счетчики
func (r *LessonBroadcastRepo) UpdateBroadcastStatus(ctx context.Context, broadcastID uuid.UUID, status string, sentCount, failedCount int) error {
	var query string
	var args []interface{}

	// Если статус финальный (completed/failed), устанавливаем completed_at
	if status == models.LessonBroadcastStatusCompleted || status == models.LessonBroadcastStatusFailed {
		query = `
			UPDATE lesson_broadcasts
			SET status = $1, sent_count = $2, failed_count = $3, completed_at = $4
			WHERE id = $5
		`
		args = []interface{}{status, sentCount, failedCount, time.Now(), broadcastID}
	} else {
		query = `
			UPDATE lesson_broadcasts
			SET status = $1, sent_count = $2, failed_count = $3
			WHERE id = $4
		`
		args = []interface{}{status, sentCount, failedCount, broadcastID}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update lesson broadcast status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrLessonBroadcastNotFound
	}

	return nil
}

// AddBroadcastFile добавляет файловое вложение к рассылке
func (r *LessonBroadcastRepo) AddBroadcastFile(ctx context.Context, file *models.BroadcastFile) error {
	// Проверяем количество существующих файлов
	existingFiles, err := r.GetBroadcastFiles(ctx, file.BroadcastID)
	if err != nil {
		return fmt.Errorf("failed to check existing files: %w", err)
	}

	if len(existingFiles) >= models.MaxBroadcastFiles {
		return models.ErrTooManyFiles
	}

	query := `
		INSERT INTO broadcast_files (id, broadcast_id, file_name, file_path, file_size, mime_type, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	file.ID = uuid.New()
	file.UploadedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
		file.ID,
		file.BroadcastID,
		file.FileName,
		file.FilePath,
		file.FileSize,
		file.MimeType,
		file.UploadedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add broadcast file: %w", err)
	}

	return nil
}

// GetBroadcastFiles получает все файлы рассылки
func (r *LessonBroadcastRepo) GetBroadcastFiles(ctx context.Context, broadcastID uuid.UUID) ([]*models.BroadcastFile, error) {
	query := `
		SELECT id, broadcast_id, file_name, file_path, file_size, mime_type, uploaded_at
		FROM broadcast_files
		WHERE broadcast_id = $1
		ORDER BY uploaded_at ASC
	`

	var files []*models.BroadcastFile
	err := r.db.SelectContext(ctx, &files, query, broadcastID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast files: %w", err)
	}

	return files, nil
}

// GetBroadcastFile получает один файл рассылки по ID
func (r *LessonBroadcastRepo) GetBroadcastFile(ctx context.Context, fileID uuid.UUID) (*models.BroadcastFile, error) {
	query := `
		SELECT id, broadcast_id, file_name, file_path, file_size, mime_type, uploaded_at
		FROM broadcast_files
		WHERE id = $1
	`

	var file models.BroadcastFile
	err := r.db.GetContext(ctx, &file, query, fileID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("broadcast file not found")
		}
		return nil, fmt.Errorf("failed to get broadcast file: %w", err)
	}

	return &file, nil
}

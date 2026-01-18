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

// HomeworkRepository управляет операциями с базой данных для домашних заданий
type HomeworkRepository struct {
	db *sqlx.DB
}

// NewHomeworkRepository создает новый HomeworkRepository
func NewHomeworkRepository(db *sqlx.DB) *HomeworkRepository {
	return &HomeworkRepository{db: db}
}

// HomeworkSelectFields определяет поля для SELECT запросов
const HomeworkSelectFields = `
	id, lesson_id, file_name, file_path, file_size, mime_type, created_by, created_at
`

// CreateHomework создает новую запись домашнего задания
func (r *HomeworkRepository) CreateHomework(ctx context.Context, homework *models.LessonHomework) (*models.LessonHomework, error) {
	// Валидация модели перед сохранением
	if err := homework.Validate(); err != nil {
		return nil, err
	}

	query := `
		INSERT INTO lesson_homework (id, lesson_id, file_name, file_path, file_size, mime_type, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, lesson_id, file_name, file_path, file_size, mime_type, created_by, created_at
	`

	// Генерируем новый ID если не задан
	if homework.ID == uuid.Nil {
		homework.ID = uuid.New()
	}

	// Устанавливаем время создания
	homework.CreatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		homework.ID,
		homework.LessonID,
		homework.FileName,
		homework.FilePath,
		homework.FileSize,
		homework.MimeType,
		homework.CreatedBy,
		homework.CreatedAt,
	).Scan(
		&homework.ID,
		&homework.LessonID,
		&homework.FileName,
		&homework.FilePath,
		&homework.FileSize,
		&homework.MimeType,
		&homework.CreatedBy,
		&homework.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create homework: %w", err)
	}

	return homework, nil
}

// GetHomeworkByLesson получает все домашние задания для урока
func (r *HomeworkRepository) GetHomeworkByLesson(ctx context.Context, lessonID uuid.UUID) ([]*models.LessonHomework, error) {
	query := `
		SELECT ` + HomeworkSelectFields + `
		FROM lesson_homework
		WHERE lesson_id = $1
		ORDER BY created_at DESC
	`

	var homeworks []*models.LessonHomework
	err := r.db.SelectContext(ctx, &homeworks, query, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get homework by lesson: %w", err)
	}

	// Возвращаем пустой массив вместо nil если нет записей
	if homeworks == nil {
		homeworks = []*models.LessonHomework{}
	}

	return homeworks, nil
}

// GetHomeworkByID получает домашнее задание по ID
func (r *HomeworkRepository) GetHomeworkByID(ctx context.Context, homeworkID uuid.UUID) (*models.LessonHomework, error) {
	query := `
		SELECT ` + HomeworkSelectFields + `
		FROM lesson_homework
		WHERE id = $1
	`

	var homework models.LessonHomework
	err := r.db.QueryRowContext(ctx, query, homeworkID).Scan(
		&homework.ID,
		&homework.LessonID,
		&homework.FileName,
		&homework.FilePath,
		&homework.FileSize,
		&homework.MimeType,
		&homework.CreatedBy,
		&homework.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrHomeworkNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get homework by ID: %w", err)
	}

	return &homework, nil
}

// DeleteHomework удаляет домашнее задание по ID
func (r *HomeworkRepository) DeleteHomework(ctx context.Context, homeworkID uuid.UUID) error {
	query := `
		DELETE FROM lesson_homework
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, homeworkID)
	if err != nil {
		return fmt.Errorf("failed to delete homework: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrHomeworkNotFound
	}

	return nil
}

// DeleteAllByLesson удаляет все домашние задания для урока (каскадное удаление)
func (r *HomeworkRepository) DeleteAllByLesson(ctx context.Context, lessonID uuid.UUID) error {
	query := `
		DELETE FROM lesson_homework
		WHERE lesson_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, lessonID)
	if err != nil {
		return fmt.Errorf("failed to delete all homework by lesson: %w", err)
	}

	// Не проверяем RowsAffected, т.к. допустимо удаление 0 записей
	// (урок может не иметь домашних заданий)
	return nil
}

// GetHomeworkCount возвращает количество домашних заданий для урока
func (r *HomeworkRepository) GetHomeworkCount(ctx context.Context, lessonID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM lesson_homework WHERE lesson_id = $1
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, lessonID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get homework count: %w", err)
	}

	return count, nil
}

// GetTotalFileSizeByLesson возвращает суммарный размер всех файлов домашних заданий для урока
func (r *HomeworkRepository) GetTotalFileSizeByLesson(ctx context.Context, lessonID uuid.UUID) (int64, error) {
	query := `
		SELECT COALESCE(SUM(file_size), 0) FROM lesson_homework WHERE lesson_id = $1
	`

	var totalSize int64
	err := r.db.QueryRowContext(ctx, query, lessonID).Scan(&totalSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get total file size: %w", err)
	}

	return totalSize, nil
}

// UpdateHomework обновляет текстовое описание домашнего задания
func (r *HomeworkRepository) UpdateHomework(ctx context.Context, homeworkID uuid.UUID, textContent string) error {
	query := `
		UPDATE lesson_homework
		SET text_content = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, textContent, homeworkID)
	if err != nil {
		return fmt.Errorf("failed to update homework: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrHomeworkNotFound
	}

	return nil
}

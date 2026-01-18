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

// SubjectRepository интерфейс для работы с предметами
type SubjectRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subject, error)
	List(ctx context.Context) ([]*models.Subject, error)
	Create(ctx context.Context, subject *models.Subject) error
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error
	AssignToTeacher(ctx context.Context, teacherID uuid.UUID, subjectID uuid.UUID) error
	RemoveFromTeacher(ctx context.Context, teacherID uuid.UUID, subjectID uuid.UUID) error
	GetTeacherSubjects(ctx context.Context, teacherID uuid.UUID) ([]*models.TeacherSubjectWithDetails, error)
}

// SubjectRepo реализация SubjectRepository
type SubjectRepo struct {
	db *sqlx.DB
}

// NewSubjectRepository создает новый SubjectRepository
func NewSubjectRepository(db *sqlx.DB) SubjectRepository {
	return &SubjectRepo{db: db}
}

// GetByID получает предмет по ID
func (r *SubjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Subject, error) {
	query := `
		SELECT id, name, description, created_at, updated_at, deleted_at
		FROM subjects
		WHERE id = $1 AND deleted_at IS NULL
	`

	var subject models.Subject
	err := r.db.GetContext(ctx, &subject, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSubjectNotFound
		}
		return nil, fmt.Errorf("failed to get subject by ID: %w", err)
	}

	return &subject, nil
}

// List получает список всех активных предметов
func (r *SubjectRepo) List(ctx context.Context) ([]*models.Subject, error) {
	query := `
		SELECT id, name, description, created_at, updated_at, deleted_at
		FROM subjects
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`

	var subjects []*models.Subject
	err := r.db.SelectContext(ctx, &subjects, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list subjects: %w", err)
	}

	return subjects, nil
}

// Create создает новый предмет
func (r *SubjectRepo) Create(ctx context.Context, subject *models.Subject) error {
	query := `
		INSERT INTO subjects (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	subject.ID = uuid.New()
	subject.CreatedAt = time.Now()
	subject.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		subject.ID,
		subject.Name,
		subject.Description,
		subject.CreatedAt,
		subject.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create subject: %w", err)
	}

	return nil
}

// Update обновляет предмет
func (r *SubjectRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Добавляем updated_at в обновления
	updates["updated_at"] = time.Now()

	// Построение динамического SQL запроса
	query := "UPDATE subjects SET "
	args := []interface{}{}
	paramCount := 1

	for key, value := range updates {
		if key == "id" || key == "created_at" || key == "deleted_at" {
			continue
		}
		if paramCount > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", key, paramCount)
		args = append(args, value)
		paramCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL", paramCount)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update subject: %w", err)
	}

	return nil
}

// Delete выполняет мягкое удаление предмета
func (r *SubjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE subjects
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete subject: %w", err)
	}

	return nil
}

// AssignToTeacher назначает предмет преподавателю
func (r *SubjectRepo) AssignToTeacher(ctx context.Context, teacherID uuid.UUID, subjectID uuid.UUID) error {
	query := `
		INSERT INTO teacher_subjects (id, teacher_id, subject_id, assigned_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query,
		uuid.New(),
		teacherID,
		subjectID,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to assign subject to teacher: %w", err)
	}

	return nil
}

// RemoveFromTeacher удаляет предмет у преподавателя
func (r *SubjectRepo) RemoveFromTeacher(ctx context.Context, teacherID uuid.UUID, subjectID uuid.UUID) error {
	query := `
		DELETE FROM teacher_subjects
		WHERE teacher_id = $1 AND subject_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, teacherID, subjectID)
	if err != nil {
		return fmt.Errorf("failed to remove subject from teacher: %w", err)
	}

	return nil
}

// GetTeacherSubjects получает список предметов, преподаваемых учителем
func (r *SubjectRepo) GetTeacherSubjects(ctx context.Context, teacherID uuid.UUID) ([]*models.TeacherSubjectWithDetails, error) {
	query := `
		SELECT
			ts.id,
			ts.teacher_id,
			ts.subject_id,
			ts.assigned_at,
			s.name as subject_name,
			s.description as subject_description
		FROM teacher_subjects ts
		JOIN subjects s ON ts.subject_id = s.id
		WHERE ts.teacher_id = $1
		ORDER BY s.name ASC
	`

	var subjects []*models.TeacherSubjectWithDetails
	err := r.db.SelectContext(ctx, &subjects, query, teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teacher subjects: %w", err)
	}

	return subjects, nil
}

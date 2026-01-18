package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Subject представляет учебный предмет (Math, English, Russian, etc.)
type Subject struct {
	ID          uuid.UUID    `db:"id" json:"id"`
	Name        string       `db:"name" json:"name"`
	Description string       `db:"description" json:"description"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt   sql.NullTime `db:"deleted_at" json:"deleted_at,omitempty"`
}

// CreateSubjectRequest представляет запрос на создание нового предмета
type CreateSubjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateSubjectRequest представляет запрос на обновление предмета
type UpdateSubjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// SubjectWithTeacherCount представляет предмет с количеством преподавателей
type SubjectWithTeacherCount struct {
	Subject
	TeacherCount int `db:"teacher_count" json:"teacher_count"`
}

// Value implements the driver.Valuer interface for database storage
func (s *Subject) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for database retrieval
func (s *Subject) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return ErrInvalidSubject
	}
}

// MarshalJSON customizes JSON marshaling for Subject to handle sql.NullTime fields
func (s *Subject) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"id":          s.ID,
		"name":        s.Name,
		"description": s.Description,
		"created_at":  s.CreatedAt,
		"updated_at":  s.UpdatedAt,
	}

	// Handle DeletedAt - convert sql.NullTime to ISO8601 string or omit if null
	if s.DeletedAt.Valid {
		data["deleted_at"] = s.DeletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	return json.Marshal(data)
}

// IsDeleted проверяет, удален ли предмет (мягкое удаление)
func (s *Subject) IsDeleted() bool {
	return s.DeletedAt.Valid
}

// Validate выполняет валидацию CreateSubjectRequest
func (r *CreateSubjectRequest) Validate() error {
	// Проверка названия предмета
	if r.Name == "" {
		return ErrInvalidSubjectName
	}
	if len(r.Name) < 2 || len(r.Name) > 100 {
		return ErrInvalidSubjectName
	}

	// Описание опционально, но если указано - проверяем длину
	if len(r.Description) > 500 {
		return ErrSubjectDescriptionTooLong
	}

	return nil
}

// Validate выполняет валидацию UpdateSubjectRequest
func (r *UpdateSubjectRequest) Validate() error {
	// Проверяем название, если оно указано
	if r.Name != nil {
		if *r.Name == "" {
			return ErrInvalidSubjectName
		}
		if len(*r.Name) < 2 || len(*r.Name) > 100 {
			return ErrInvalidSubjectName
		}
	}

	// Проверяем описание, если оно указано
	if r.Description != nil && len(*r.Description) > 500 {
		return ErrSubjectDescriptionTooLong
	}

	return nil
}

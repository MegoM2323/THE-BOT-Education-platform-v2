package models

import (
	"time"

	"github.com/google/uuid"
)

// TeacherSubject представляет связь между преподавателем и предметом
type TeacherSubject struct {
	ID         uuid.UUID `db:"id" json:"id"`
	TeacherID  uuid.UUID `db:"teacher_id" json:"teacher_id"`
	SubjectID  uuid.UUID `db:"subject_id" json:"subject_id"`
	AssignedAt time.Time `db:"assigned_at" json:"assigned_at"`
}

// TeacherSubjectWithDetails представляет связь преподавателя с деталями предмета
type TeacherSubjectWithDetails struct {
	TeacherSubject
	SubjectName        string `db:"subject_name" json:"subject_name"`
	SubjectDescription string `db:"subject_description" json:"subject_description"`
}

// AssignSubjectRequest представляет запрос на назначение предмета преподавателю
type AssignSubjectRequest struct {
	TeacherID uuid.UUID `json:"teacher_id"`
	SubjectID uuid.UUID `json:"subject_id"`
}

// RemoveSubjectRequest представляет запрос на удаление предмета у преподавателя
type RemoveSubjectRequest struct {
	TeacherID uuid.UUID `json:"teacher_id"`
	SubjectID uuid.UUID `json:"subject_id"`
}

// Validate выполняет валидацию AssignSubjectRequest
func (r *AssignSubjectRequest) Validate() error {
	// Проверка ID преподавателя
	if r.TeacherID == uuid.Nil {
		return ErrInvalidTeacherID
	}

	// Проверка ID предмета
	if r.SubjectID == uuid.Nil {
		return ErrInvalidSubjectID
	}

	return nil
}

// Validate выполняет валидацию RemoveSubjectRequest
func (r *RemoveSubjectRequest) Validate() error {
	// Проверка ID преподавателя
	if r.TeacherID == uuid.Nil {
		return ErrInvalidTeacherID
	}

	// Проверка ID предмета
	if r.SubjectID == uuid.Nil {
		return ErrInvalidSubjectID
	}

	return nil
}

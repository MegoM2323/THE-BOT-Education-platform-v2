package models

import (
	"fmt"

	"github.com/google/uuid"
)

// ApplyToAllSubsequentRequest represents a request to apply a modification to a lesson and all its subsequent occurrences
type ApplyToAllSubsequentRequest struct {
	LessonID         uuid.UUID  `json:"lesson_id"`
	ModificationType string     `json:"modification_type"` // add_student, remove_student, change_teacher, change_time, change_capacity
	StudentID        *uuid.UUID `json:"student_id,omitempty"`
	TeacherID        *uuid.UUID `json:"teacher_id,omitempty"`
	NewStartTime     *string    `json:"new_start_time,omitempty"` // ISO format
	NewMaxStudents   *int       `json:"new_max_students,omitempty"`
}

// Validate validates the ApplyToAllSubsequentRequest based on modification type
func (r *ApplyToAllSubsequentRequest) Validate() error {
	if r.LessonID == uuid.Nil {
		return fmt.Errorf("lesson_id is required")
	}

	if r.ModificationType == "" {
		return fmt.Errorf("modification_type is required")
	}

	switch r.ModificationType {
	case "add_student", "remove_student":
		if r.StudentID == nil || *r.StudentID == uuid.Nil {
			return fmt.Errorf("student_id is required for %s", r.ModificationType)
		}
	case "change_teacher":
		if r.TeacherID == nil || *r.TeacherID == uuid.Nil {
			return fmt.Errorf("teacher_id is required for change_teacher")
		}
	case "change_time":
		if r.NewStartTime == nil || *r.NewStartTime == "" {
			return fmt.Errorf("new_start_time is required for change_time")
		}
	case "change_capacity":
		if r.NewMaxStudents == nil || *r.NewMaxStudents < 1 {
			return fmt.Errorf("new_max_students must be >= 1 for change_capacity")
		}
	default:
		return fmt.Errorf("invalid modification_type: %s (must be add_student, remove_student, change_teacher, change_time, or change_capacity)", r.ModificationType)
	}

	return nil
}

// ApplyToAllSubsequentResponse represents the response after applying a modification
type ApplyToAllSubsequentResponse struct {
	Success              bool                `json:"success"`
	Modification         *LessonModification `json:"modification"`
	AffectedLessonsCount int                 `json:"affected_lessons_count"`
	Message              string              `json:"message"`
}

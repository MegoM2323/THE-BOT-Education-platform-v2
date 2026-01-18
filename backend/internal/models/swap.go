package models

import (
	"time"

	"github.com/google/uuid"
)

// Swap представляет операцию обмена урока
type Swap struct {
	ID           uuid.UUID `db:"id" json:"id"`
	StudentID    uuid.UUID `db:"student_id" json:"student_id"`
	OldLessonID  uuid.UUID `db:"old_lesson_id" json:"old_lesson_id"`
	NewLessonID  uuid.UUID `db:"new_lesson_id" json:"new_lesson_id"`
	OldBookingID uuid.UUID `db:"old_booking_id" json:"old_booking_id"`
	NewBookingID uuid.UUID `db:"new_booking_id" json:"new_booking_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// SwapWithDetails представляет обмен с деталями уроков
type SwapWithDetails struct {
	Swap
	OldLessonStartTime time.Time `db:"old_lesson_start_time" json:"old_lesson_start_time"`
	OldLessonEndTime   time.Time `db:"old_lesson_end_time" json:"old_lesson_end_time"`
	NewLessonStartTime time.Time `db:"new_lesson_start_time" json:"new_lesson_start_time"`
	NewLessonEndTime   time.Time `db:"new_lesson_end_time" json:"new_lesson_end_time"`
	StudentName        string    `db:"student_name" json:"student_name"`
}

// PerformSwapRequest представляет запрос на выполнение обмена урока
type PerformSwapRequest struct {
	StudentID   uuid.UUID `json:"student_id"`
	OldLessonID uuid.UUID `json:"old_lesson_id"`
	NewLessonID uuid.UUID `json:"new_lesson_id"`
}

// ValidateSwapRequest представляет запрос на валидацию обмена перед выполнением
type ValidateSwapRequest struct {
	StudentID   uuid.UUID `json:"student_id"`
	OldLessonID uuid.UUID `json:"old_lesson_id"`
	NewLessonID uuid.UUID `json:"new_lesson_id"`
}

// ValidateSwapResponse представляет ответ валидации обмена
type ValidateSwapResponse struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Details struct {
		OldLesson *LessonWithTeacher `json:"old_lesson,omitempty"`
		NewLesson *LessonWithTeacher `json:"new_lesson,omitempty"`
	} `json:"details,omitempty"`
}

// GetSwapHistoryFilter представляет фильтры для истории обменов
type GetSwapHistoryFilter struct {
	StudentID *uuid.UUID `json:"student_id,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// Validate выполняет валидацию PerformSwapRequest
func (r *PerformSwapRequest) Validate() error {
	if r.StudentID == uuid.Nil {
		return ErrInvalidStudentID
	}
	if r.OldLessonID == uuid.Nil {
		return ErrInvalidOldLessonID
	}
	if r.NewLessonID == uuid.Nil {
		return ErrInvalidNewLessonID
	}
	if r.OldLessonID == r.NewLessonID {
		return ErrSameLessonSwap
	}
	return nil
}

// Validate выполняет валидацию ValidateSwapRequest
func (r *ValidateSwapRequest) Validate() error {
	if r.StudentID == uuid.Nil {
		return ErrInvalidStudentID
	}
	if r.OldLessonID == uuid.Nil {
		return ErrInvalidOldLessonID
	}
	if r.NewLessonID == uuid.Nil {
		return ErrInvalidNewLessonID
	}
	if r.OldLessonID == r.NewLessonID {
		return ErrSameLessonSwap
	}
	return nil
}

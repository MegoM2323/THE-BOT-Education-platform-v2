package models

import (
	"time"

	"github.com/google/uuid"
)

// CancelledBooking представляет запись об отменённом бронировании
// для предотвращения повторной записи студента на тот же урок
type CancelledBooking struct {
	ID          uuid.UUID `db:"id" json:"id"`
	BookingID   uuid.UUID `db:"booking_id" json:"booking_id"`
	StudentID   uuid.UUID `db:"student_id" json:"student_id"`
	LessonID    uuid.UUID `db:"lesson_id" json:"lesson_id"`
	CancelledAt time.Time `db:"cancelled_at" json:"cancelled_at"`
}

// Validate выполняет валидацию CancelledBooking
func (cb *CancelledBooking) Validate() error {
	if cb.BookingID == uuid.Nil {
		return ErrInvalidBookingID
	}
	if cb.StudentID == uuid.Nil {
		return ErrInvalidStudentID
	}
	if cb.LessonID == uuid.Nil {
		return ErrInvalidLessonID
	}
	return nil
}

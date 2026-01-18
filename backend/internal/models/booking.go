package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// BookingStatus представляет статус бронирования
type BookingStatus string

const (
	BookingStatusActive    BookingStatus = "active"
	BookingStatusCancelled BookingStatus = "cancelled"
)

// CancelBookingResultStatus представляет результат операции отмены
type CancelBookingResultStatus string

const (
	// CancelResultSuccess успешная отмена активного бронирования
	CancelResultSuccess CancelBookingResultStatus = "success"
	// CancelResultAlreadyCancelled бронирование уже было отменено ранее (идемпотентность)
	CancelResultAlreadyCancelled CancelBookingResultStatus = "already_cancelled"
)

// CancelBookingResult представляет результат операции отмены бронирования с информацией о статусе
type CancelBookingResult struct {
	// Status: "success" если было активное бронирование и успешно отменено
	// Status: "already_cancelled" если бронирование было уже отменено (идемпотентное поведение)
	Status  CancelBookingResultStatus `json:"status"`
	Message string                    `json:"message"`
}

// Booking представляет бронирование урока студентом
type Booking struct {
	ID          uuid.UUID     `db:"id" json:"id"`
	StudentID   uuid.UUID     `db:"student_id" json:"student_id"`
	LessonID    uuid.UUID     `db:"lesson_id" json:"lesson_id"`
	Status      BookingStatus `db:"status" json:"status"`
	BookedAt    time.Time     `db:"booked_at" json:"booked_at"`
	CancelledAt sql.NullTime  `db:"cancelled_at" json:"cancelled_at,omitempty"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
}

// BookingWithDetails представляет бронирование с деталями урока, преподавателя и студента
type BookingWithDetails struct {
	Booking
	// Поля из таблицы lessons (start_time и end_time всегда NOT NULL в БД)
	StartTime time.Time `db:"start_time" json:"start_time"`
	EndTime   time.Time `db:"end_time" json:"end_time"`
	TeacherID uuid.UUID `db:"teacher_id" json:"teacher_id"`
	// Subject и HomeworkText для отображения ДЗ у студентов
	Subject      sql.NullString `db:"subject" json:"subject,omitempty"`
	HomeworkText sql.NullString `db:"homework_text" json:"homework_text,omitempty"`
	// Поля из таблицы users (учитель)
	TeacherName string `db:"teacher_name" json:"teacher_name,omitempty"`
	// Поля из таблицы users (студент)
	StudentFullName string `db:"student_full_name" json:"student_name,omitempty"`
	StudentEmail    string `db:"student_email" json:"student_email,omitempty"`
	// booking_created_at всегда NOT NULL в БД (это b.created_at)
	BookingCreatedAt time.Time `db:"booking_created_at" json:"booking_created_at"`
}

// CreateBookingRequest представляет запрос на создание нового бронирования
type CreateBookingRequest struct {
	StudentID uuid.UUID `json:"student_id"`
	LessonID  uuid.UUID `json:"lesson_id"`
	IsAdmin   bool      `json:"is_admin,omitempty"` // Skip credit check for admin bookings
	AdminID   uuid.UUID `json:"admin_id,omitempty"` // Admin who created the booking
}

// CancelBookingRequest представляет запрос на отмену бронирования
type CancelBookingRequest struct {
	BookingID uuid.UUID `json:"booking_id"`
	StudentID uuid.UUID `json:"student_id"`
	IsAdmin   bool      `json:"is_admin"` // Флаг для разрешения админам удалять любые бронирования
}

// BookingResponse представляет стандартный ответ для бронирования в API
type BookingResponse struct {
	ID               uuid.UUID     `json:"id"`
	StudentID        uuid.UUID     `json:"student_id"`
	LessonID         uuid.UUID     `json:"lesson_id"`
	Status           BookingStatus `json:"status"`
	BookedAt         time.Time     `json:"booked_at"`
	CancelledAt      *time.Time    `json:"cancelled_at,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	StartTime        *time.Time    `json:"start_time,omitempty"`
	EndTime          *time.Time    `json:"end_time,omitempty"`
	TeacherID        *uuid.UUID    `json:"teacher_id,omitempty"`
	TeacherName      string        `json:"teacher_name,omitempty"`
	Subject          string        `json:"subject,omitempty"`
	HomeworkText     string        `json:"homework_text,omitempty"`
	StudentFullName  string        `json:"student_name,omitempty"`
	StudentEmail     string        `json:"student_email,omitempty"`
	BookingCreatedAt *time.Time    `json:"booking_created_at,omitempty"`
}

// ListBookingsFilter представляет фильтры для списка бронирований
type ListBookingsFilter struct {
	StudentID *uuid.UUID     `json:"student_id,omitempty"`
	LessonID  *uuid.UUID     `json:"lesson_id,omitempty"`
	Status    *BookingStatus `json:"status,omitempty"`
	StartDate *time.Time     `json:"start_date,omitempty"`
	EndDate   *time.Time     `json:"end_date,omitempty"`
}

// IsActive проверяет, активно ли бронирование
func (b *Booking) IsActive() bool {
	return b.Status == BookingStatusActive
}

// IsCancelled проверяет, отменено ли бронирование
func (b *Booking) IsCancelled() bool {
	return b.Status == BookingStatusCancelled
}

// Validate выполняет валидацию CreateBookingRequest
func (r *CreateBookingRequest) Validate() error {
	if r.StudentID == uuid.Nil {
		return ErrInvalidStudentID
	}
	if r.LessonID == uuid.Nil {
		return ErrInvalidLessonID
	}
	return nil
}

// Validate выполняет валидацию CancelBookingRequest
func (r *CancelBookingRequest) Validate() error {
	if r.BookingID == uuid.Nil {
		return ErrInvalidBookingID
	}
	if r.StudentID == uuid.Nil {
		return ErrInvalidStudentID
	}
	return nil
}

// ToBookingResponse конвертирует BookingWithDetails в BookingResponse
func (b *BookingWithDetails) ToBookingResponse() *BookingResponse {
	var cancelledAt *time.Time
	if b.CancelledAt.Valid {
		cancelledAt = &b.CancelledAt.Time
	}

	var startTime *time.Time
	if !b.StartTime.IsZero() {
		startTime = &b.StartTime
	}

	var endTime *time.Time
	if !b.EndTime.IsZero() {
		endTime = &b.EndTime
	}

	var teacherID *uuid.UUID
	if b.TeacherID != uuid.Nil {
		teacherID = &b.TeacherID
	}

	var bookingCreatedAt *time.Time
	if !b.BookingCreatedAt.IsZero() {
		bookingCreatedAt = &b.BookingCreatedAt
	}

	var subject string
	if b.Subject.Valid {
		subject = b.Subject.String
	}

	var homeworkText string
	if b.HomeworkText.Valid {
		homeworkText = b.HomeworkText.String
	}

	return &BookingResponse{
		ID:               b.ID,
		StudentID:        b.StudentID,
		LessonID:         b.LessonID,
		Status:           b.Status,
		BookedAt:         b.BookedAt,
		CancelledAt:      cancelledAt,
		CreatedAt:        b.CreatedAt,
		UpdatedAt:        b.UpdatedAt,
		StartTime:        startTime,
		EndTime:          endTime,
		TeacherID:        teacherID,
		TeacherName:      b.TeacherName,
		Subject:          subject,
		HomeworkText:     homeworkText,
		StudentFullName:  b.StudentFullName,
		StudentEmail:     b.StudentEmail,
		BookingCreatedAt: bookingCreatedAt,
	}
}

// ToBookingResponse конвертирует простое Booking в BookingResponse
func (b *Booking) ToBookingResponse() *BookingResponse {
	var cancelledAt *time.Time
	if b.CancelledAt.Valid {
		cancelledAt = &b.CancelledAt.Time
	}

	return &BookingResponse{
		ID:          b.ID,
		StudentID:   b.StudentID,
		LessonID:    b.LessonID,
		Status:      b.Status,
		BookedAt:    b.BookedAt,
		CancelledAt: cancelledAt,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

// MarshalJSON marshals BookingWithDetails to JSON, converting sql.NullString to string
func (b *BookingWithDetails) MarshalJSON() ([]byte, error) {
	type BookingWithDetailsAlias BookingWithDetails

	var cancelledAt *time.Time
	if b.CancelledAt.Valid {
		cancelledAt = &b.CancelledAt.Time
	}

	subject := ""
	if b.Subject.Valid {
		subject = b.Subject.String
	}

	homeworkText := ""
	if b.HomeworkText.Valid {
		homeworkText = b.HomeworkText.String
	}

	return json.Marshal(&struct {
		CancelledAt  *time.Time `json:"cancelled_at,omitempty"`
		Subject      string     `json:"subject,omitempty"`
		HomeworkText string     `json:"homework_text,omitempty"`
		*BookingWithDetailsAlias
	}{
		CancelledAt:             cancelledAt,
		Subject:                 subject,
		HomeworkText:            homeworkText,
		BookingWithDetailsAlias: (*BookingWithDetailsAlias)(b),
	})
}

package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// BookingInfo представляет информацию о студенте, записанном на занятие
type BookingInfo struct {
	StudentID   uuid.UUID `json:"student_id"`
	StudentName string    `json:"student_name"`
}

// LessonType определяет тип урока
type LessonType string

const (
	// LessonTypeIndividual представляет индивидуальный урок (1 студент)
	LessonTypeIndividual LessonType = "individual"
	// LessonTypeGroup представляет групповой урок (2+ студентов)
	LessonTypeGroup LessonType = "group"
)

// Lesson представляет урок, созданный преподавателем
type Lesson struct {
	ID                    uuid.UUID      `db:"id" json:"id"`
	TeacherID             uuid.UUID      `db:"teacher_id" json:"teacher_id"`
	StartTime             time.Time      `db:"start_time" json:"start_time"`
	EndTime               time.Time      `db:"end_time" json:"end_time"`
	MaxStudents           int            `db:"max_students" json:"max_students"`
	CurrentStudents       int            `db:"current_students" json:"current_students"`
	CreditsCost           int            `db:"credits_cost" json:"credits_cost"`
	Color                 string         `db:"color" json:"color"`
	Subject               sql.NullString `db:"subject" json:"subject,omitempty"`
	HomeworkText          sql.NullString `db:"homework_text" json:"homework_text,omitempty"`
	ReportText            sql.NullString `db:"report_text" json:"report_text,omitempty"`
	Link                  sql.NullString `db:"link" json:"link,omitempty"`
	AppliedFromTemplate   bool           `db:"applied_from_template" json:"applied_from_template"`
	TemplateApplicationID *uuid.UUID     `db:"template_application_id" json:"template_application_id,omitempty"`
	IsRecurring           bool           `db:"is_recurring" json:"is_recurring"`
	RecurringGroupID      *uuid.UUID     `db:"recurring_group_id" json:"recurring_group_id,omitempty"`
	RecurringEndDate      sql.NullTime   `db:"recurring_end_date" json:"recurring_end_date,omitempty"`
	CreatedAt             time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt             sql.NullTime   `db:"deleted_at" json:"deleted_at,omitempty"`
}

// LessonWithTeacher представляет урок с информацией о преподавателе
type LessonWithTeacher struct {
	Lesson
	TeacherName string `db:"teacher_name" json:"teacher_name"`
}

// LessonResponse представляет урок в API response с вычисляемыми полями
type LessonResponse struct {
	ID                    uuid.UUID     `json:"id"`
	TeacherID             uuid.UUID     `json:"teacher_id"`
	TeacherName           string        `json:"teacher_name"`
	StartTime             time.Time     `json:"start_time"`
	EndTime               time.Time     `json:"end_time"`
	MaxStudents           int           `json:"max_students"`
	CurrentStudents       int           `json:"current_students"`
	CreditsCost           int           `json:"credits_cost"`
	Color                 string        `json:"color"`
	Subject               string        `json:"subject"`
	HomeworkText          string        `json:"homework_text,omitempty"`
	ReportText            string        `json:"report_text,omitempty"`
	Link                  string        `json:"link,omitempty"`
	AppliedFromTemplate   bool          `json:"applied_from_template"`
	TemplateApplicationID string        `json:"template_application_id,omitempty"`
	IsPast                bool          `json:"is_past"` // Вычисляемое поле: занятие уже прошло
	Bookings              []BookingInfo `json:"bookings,omitempty"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
	DeletedAt             *time.Time    `json:"deleted_at,omitempty"`
}

// ToResponse преобразует LessonWithTeacher в LessonResponse с вычисляемым полем is_past
func (l *LessonWithTeacher) ToResponse() *LessonResponse {
	subject := ""
	if l.Subject.Valid {
		subject = l.Subject.String
	}

	homeworkText := ""
	if l.HomeworkText.Valid {
		homeworkText = l.HomeworkText.String
	}

	reportText := ""
	if l.ReportText.Valid {
		reportText = l.ReportText.String
	}

	link := ""
	if l.Link.Valid {
		link = l.Link.String
	}

	deletedAt := (*time.Time)(nil)
	if l.DeletedAt.Valid {
		deletedAt = &l.DeletedAt.Time
	}

	templateAppID := ""
	if l.TemplateApplicationID != nil {
		templateAppID = l.TemplateApplicationID.String()
	}

	return &LessonResponse{
		ID:                    l.ID,
		TeacherID:             l.TeacherID,
		TeacherName:           l.TeacherName,
		StartTime:             l.StartTime,
		EndTime:               l.EndTime,
		MaxStudents:           l.MaxStudents,
		CurrentStudents:       l.CurrentStudents,
		CreditsCost:           l.CreditsCost,
		Color:                 l.Color,
		Subject:               subject,
		HomeworkText:          homeworkText,
		ReportText:            reportText,
		Link:                  link,
		AppliedFromTemplate:   l.AppliedFromTemplate,
		TemplateApplicationID: templateAppID,
		IsPast:                l.IsInPast(),
		CreatedAt:             l.CreatedAt,
		UpdatedAt:             l.UpdatedAt,
		DeletedAt:             deletedAt,
	}
}

// ToResponseWithoutTeacher преобразует Lesson в LessonResponse (без teacher_name)
func (l *Lesson) ToResponseWithoutTeacher() *LessonResponse {
	subject := ""
	if l.Subject.Valid {
		subject = l.Subject.String
	}

	homeworkText := ""
	if l.HomeworkText.Valid {
		homeworkText = l.HomeworkText.String
	}

	reportText := ""
	if l.ReportText.Valid {
		reportText = l.ReportText.String
	}

	link := ""
	if l.Link.Valid {
		link = l.Link.String
	}

	deletedAt := (*time.Time)(nil)
	if l.DeletedAt.Valid {
		deletedAt = &l.DeletedAt.Time
	}

	templateAppID := ""
	if l.TemplateApplicationID != nil {
		templateAppID = l.TemplateApplicationID.String()
	}

	return &LessonResponse{
		ID:                    l.ID,
		TeacherID:             l.TeacherID,
		TeacherName:           "", // Не загружено
		StartTime:             l.StartTime,
		EndTime:               l.EndTime,
		MaxStudents:           l.MaxStudents,
		CurrentStudents:       l.CurrentStudents,
		CreditsCost:           l.CreditsCost,
		Color:                 l.Color,
		Subject:               subject,
		HomeworkText:          homeworkText,
		ReportText:            reportText,
		Link:                  link,
		AppliedFromTemplate:   l.AppliedFromTemplate,
		TemplateApplicationID: templateAppID,
		IsPast:                l.IsInPast(),
		CreatedAt:             l.CreatedAt,
		UpdatedAt:             l.UpdatedAt,
		DeletedAt:             deletedAt,
	}
}

// CreateLessonRequest представляет запрос на создание нового урока
type CreateLessonRequest struct {
	TeacherID        uuid.UUID   `json:"teacher_id"`
	StartTime        time.Time   `json:"start_time"`
	EndTime          time.Time   `json:"end_time"`                // Required: время окончания
	MaxStudents      int         `json:"max_students"`            // Required: максимум студентов
	CreditsCost      int         `json:"credits_cost"`            // Required: стоимость в кредитах
	Color            string      `json:"color"`                   // Required: цвет занятия
	LessonType       *LessonType `json:"lesson_type,omitempty"`   // Optional: defaults to individual
	Subject          *string     `json:"subject,omitempty"`       // Optional: lesson subject/topic
	HomeworkText     *string     `json:"homework_text,omitempty"` // Optional: homework text instructions
	Link             *string     `json:"link,omitempty"`          // Optional: link to meeting/resources
	StudentIDs       []uuid.UUID `json:"student_ids,omitempty"`   // Optional: students to enroll on creation
	IsRecurring      bool        `json:"is_recurring,omitempty"`  // Optional: создать повторяющееся занятие
	RecurringWeeks   *int        `json:"recurring_weeks,omitempty"`   // Optional: количество недель (4, 8, 12)
	RecurringEndDate *time.Time  `json:"recurring_end_date,omitempty"` // Optional: дата окончания повторений
}

// UpdateLessonRequest представляет запрос на обновление урока
type UpdateLessonRequest struct {
	TeacherID    *uuid.UUID  `json:"teacher_id,omitempty"`
	StartTime    *time.Time  `json:"start_time,omitempty"`
	EndTime      *time.Time  `json:"end_time,omitempty"`
	LessonType   *LessonType `json:"lesson_type,omitempty"`
	MaxStudents  *int        `json:"max_students,omitempty"`
	CreditsCost  *int        `json:"credits_cost,omitempty"`
	Color        *string     `json:"color,omitempty"`
	Subject      *string     `json:"subject,omitempty"`
	HomeworkText *string     `json:"homework_text,omitempty"`
	ReportText   *string     `json:"report_text,omitempty"`
	Link         *string     `json:"link,omitempty"`
	ApplyToFuture *bool      `json:"apply_to_future,omitempty"` // Optional: применить ко всем будущим занятиям серии
}

// ListLessonsFilter представляет фильтры для списка уроков
type ListLessonsFilter struct {
	TeacherID *uuid.UUID `json:"teacher_id,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Available *bool      `json:"available,omitempty"` // Есть свободные места
}

// IsDeleted проверяет, удален ли урок (мягкое удаление)
func (l *Lesson) IsDeleted() bool {
	return l.DeletedAt.Valid
}

// IsFull проверяет, достиг ли урок максимальной вместимости
func (l *Lesson) IsFull() bool {
	return l.CurrentStudents >= l.MaxStudents
}

// HasAvailableSpots проверяет, есть ли свободные места на уроке
func (l *Lesson) HasAvailableSpots() bool {
	return l.CurrentStudents < l.MaxStudents
}

// IsInPast проверяет, находится ли время начала урока в прошлом
func (l *Lesson) IsInPast() bool {
	return l.StartTime.Before(time.Now())
}

// IsWithin24Hours проверяет, начинается ли урок в течение 24 часов
func (l *Lesson) IsWithin24Hours() bool {
	return time.Until(l.StartTime) < 24*time.Hour
}

// ValidateLessonTypeChange проверяет, может ли урок изменить свой тип
// Возвращает ошибку, если попытка изменить групповой урок с несколькими студентами на индивидуальный
func (l *Lesson) ValidateLessonTypeChange(newLessonType LessonType) error {
	// Cannot change to individual if current students > 1
	if newLessonType == LessonTypeIndividual && l.CurrentStudents > 1 {
		return ErrCannotChangeToIndividual
	}
	return nil
}

// lessonMarshalHelper is a helper function to handle common JSON marshaling logic
// It converts sql.NullString and sql.NullTime fields to proper JSON types
func lessonMarshalHelper(lesson *Lesson, teacherName *string) ([]byte, error) {
	type LessonAlias Lesson
	subject := ""
	if lesson.Subject.Valid {
		subject = lesson.Subject.String
	}
	homeworkText := ""
	if lesson.HomeworkText.Valid {
		homeworkText = lesson.HomeworkText.String
	}
	link := ""
	if lesson.Link.Valid {
		link = lesson.Link.String
	}
	deletedAt := (*time.Time)(nil)
	if lesson.DeletedAt.Valid {
		deletedAt = &lesson.DeletedAt.Time
	}
	templateAppID := ""
	if lesson.TemplateApplicationID != nil {
		templateAppID = lesson.TemplateApplicationID.String()
	}

	// For LessonWithTeacher: include teacher_name field
	if teacherName != nil {
		return json.Marshal(&struct {
			Subject               string     `json:"subject"`
			HomeworkText          string     `json:"homework_text,omitempty"`
			Link                  string     `json:"link,omitempty"`
			DeletedAt             *time.Time `json:"deleted_at,omitempty"`
			TemplateApplicationID string     `json:"template_application_id,omitempty"`
			TeacherName           string     `json:"teacher_name"`
			*LessonAlias
		}{
			Subject:               subject,
			HomeworkText:          homeworkText,
			Link:                  link,
			DeletedAt:             deletedAt,
			TemplateApplicationID: templateAppID,
			TeacherName:           *teacherName,
			LessonAlias:           (*LessonAlias)(lesson),
		})
	}

	// For Lesson only: exclude teacher_name
	return json.Marshal(&struct {
		Subject               string     `json:"subject"`
		HomeworkText          string     `json:"homework_text,omitempty"`
		Link                  string     `json:"link,omitempty"`
		DeletedAt             *time.Time `json:"deleted_at,omitempty"`
		TemplateApplicationID string     `json:"template_application_id,omitempty"`
		*LessonAlias
	}{
		Subject:               subject,
		HomeworkText:          homeworkText,
		Link:                  link,
		DeletedAt:             deletedAt,
		TemplateApplicationID: templateAppID,
		LessonAlias:           (*LessonAlias)(lesson),
	})
}

// MarshalJSON marshals Lesson to JSON, converting sql.NullString to string
func (l *Lesson) MarshalJSON() ([]byte, error) {
	return lessonMarshalHelper(l, nil)
}

// MarshalJSON marshals LessonWithTeacher to JSON, including the teacher_name field
func (l *LessonWithTeacher) MarshalJSON() ([]byte, error) {
	return lessonMarshalHelper(&l.Lesson, &l.TeacherName)
}

// ApplyDefaults applies default values for optional fields in CreateLessonRequest
func (r *CreateLessonRequest) ApplyDefaults() {
	// Infer LessonType from MaxStudents if not set
	if r.LessonType == nil {
		if r.MaxStudents == 1 {
			lessonType := LessonTypeIndividual
			r.LessonType = &lessonType
		} else if r.MaxStudents >= 4 {
			lessonType := LessonTypeGroup
			r.LessonType = &lessonType
		}
		// If max_students is 2 or 3, don't set lesson type (will be caught by validation)
	}
}

// Validate выполняет валидацию CreateLessonRequest
// NOTE: This is called AFTER ApplyDefaults in production, but tests may call it directly
func (r *CreateLessonRequest) Validate() error {
	if r.TeacherID == uuid.Nil {
		return ErrInvalidTeacherID
	}
	if r.StartTime.IsZero() {
		return ErrInvalidLessonTime
	}

	// Validate EndTime
	if r.EndTime.IsZero() {
		return ErrInvalidLessonTime
	}
	if r.EndTime.Before(r.StartTime) || r.EndTime.Equal(r.StartTime) {
		return ErrInvalidLessonTime
	}

	// Validate MaxStudents
	if r.MaxStudents <= 0 {
		return ErrInvalidMaxStudents
	}

	// Validate CreditsCost
	if r.CreditsCost < 0 {
		return ErrInvalidMaxStudents // Можно добавить специальную ошибку
	}

	// Validate Color (should not be empty)
	if r.Color == "" {
		return ErrInvalidMaxStudents // Можно добавить специальную ошибку
	}

	// Validate constraints based on lesson type
	if r.LessonType != nil {
		if *r.LessonType == LessonTypeIndividual {
			// Individual lessons MUST have exactly 1 student
			if r.MaxStudents != 1 {
				return ErrIndividualLessonMaxStudents
			}
		} else if *r.LessonType == LessonTypeGroup {
			// Group lessons MUST have at least 4 students
			if r.MaxStudents < 4 {
				return ErrGroupLessonMinStudents
			}
		}
	} else {
		// No lesson type set - infer from max_students and validate
		// Valid: maxStudents = 1 (individual) or maxStudents >= 4 (group)
		// Invalid: maxStudents in [2, 3] - treated as failed group lesson
		if r.MaxStudents == 1 {
			// Valid for individual
		} else if r.MaxStudents >= 4 {
			// Valid for group
		} else {
			// maxStudents in [2, 3]: invalid - group lessons require at least 4
			return ErrGroupLessonMinStudents
		}
	}

	return nil
}

// Validate выполняет валидацию UpdateLessonRequest
func (r *UpdateLessonRequest) Validate() error {
	// Проверяем max_students, если он указан
	if r.MaxStudents != nil {
		if *r.MaxStudents <= 0 {
			return ErrInvalidMaxStudents
		}

		// Validate constraints for group lessons (must have at least 4)
		if r.LessonType != nil && *r.LessonType == LessonTypeGroup {
			if *r.MaxStudents < 4 {
				return ErrGroupLessonMinStudents
			}
		}
	}

	// Проверяем время, если оба параметра указаны
	if r.StartTime != nil && r.EndTime != nil {
		if r.EndTime.Before(*r.StartTime) || r.EndTime.Equal(*r.StartTime) {
			return ErrInvalidLessonTime
		}
	}

	return nil
}

// EnrolledStudent представляет студента, записанного на занятие (для teacher schedule API)
type EnrolledStudent struct {
	ID       uuid.UUID `db:"id" json:"id"`
	FullName string    `db:"full_name" json:"full_name"`
	Email    string    `db:"email" json:"email"`
}

// TeacherScheduleLesson представляет занятие в календаре преподавателя с дополнительной информацией
type TeacherScheduleLesson struct {
	Lesson
	TeacherName           string             `db:"teacher_name" json:"teacher_name"`
	EnrolledStudentsCount int                `db:"enrolled_students_count" json:"enrolled_students_count"`
	HomeworkCount         int                `db:"homework_count" json:"homework_count"`
	BroadcastsCount       int                `db:"broadcasts_count" json:"broadcasts_count"`
	EnrolledStudents      []*EnrolledStudent `json:"enrolled_students"`
	IsPast                bool               `json:"is_past"` // Вычисляемое поле
}

// ToResponse преобразует TeacherScheduleLesson в JSON response с вычисляемым полем is_past
func (l *TeacherScheduleLesson) ToResponse() map[string]interface{} {
	subject := ""
	if l.Subject.Valid {
		subject = l.Subject.String
	}

	homeworkText := ""
	if l.HomeworkText.Valid {
		homeworkText = l.HomeworkText.String
	}

	link := ""
	if l.Link.Valid {
		link = l.Link.String
	}

	templateAppID := ""
	if l.TemplateApplicationID != nil {
		templateAppID = l.TemplateApplicationID.String()
	}

	// Обрабатываем nil slice
	enrolledStudents := l.EnrolledStudents
	if enrolledStudents == nil {
		enrolledStudents = []*EnrolledStudent{}
	}

	return map[string]interface{}{
		"id":                      l.ID,
		"teacher_id":              l.TeacherID,
		"start_time":              l.StartTime,
		"end_time":                l.EndTime,
		"max_students":            l.MaxStudents,
		"current_students":        l.CurrentStudents,
		"credits_cost":            l.CreditsCost,
		"color":                   l.Color,
		"subject":                 subject,
		"homework_text":           homeworkText,
		"link":                    link,
		"applied_from_template":   l.AppliedFromTemplate,
		"template_application_id": templateAppID,
		"is_past":                 l.StartTime.Before(time.Now()),
		"enrolled_students_count": l.EnrolledStudentsCount,
		"enrolled_students":       enrolledStudents,
		"homework_count":          l.HomeworkCount,
		"broadcasts_count":        l.BroadcastsCount,
	}
}

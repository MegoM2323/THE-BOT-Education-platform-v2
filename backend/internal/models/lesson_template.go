package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TemplateApplicationStatus represents valid status values for TemplateApplication
type TemplateApplicationStatus string

// Valid status constants for TemplateApplication
const (
	TemplateApplicationStatusApplied    TemplateApplicationStatus = "applied"
	TemplateApplicationStatusReplaced   TemplateApplicationStatus = "replaced"
	TemplateApplicationStatusRolledBack TemplateApplicationStatus = "rolled_back"
	TemplateApplicationStatusPreview    TemplateApplicationStatus = "preview" // Dry-run preview
)

// IsValid checks if the status is a valid TemplateApplicationStatus
func (s TemplateApplicationStatus) IsValid() bool {
	switch s {
	case TemplateApplicationStatusApplied,
		TemplateApplicationStatusReplaced,
		TemplateApplicationStatusRolledBack,
		TemplateApplicationStatusPreview:
		return true
	default:
		return false
	}
}

// String returns the string representation of TemplateApplicationStatus
func (s TemplateApplicationStatus) String() string {
	return string(s)
}

// LessonTemplate represents a weekly schedule template created by admin
type LessonTemplate struct {
	ID          uuid.UUID      `db:"id" json:"id"`
	AdminID     uuid.UUID      `db:"admin_id" json:"admin_id"`
	Name        string         `db:"name" json:"name"`
	Description sql.NullString `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at" json:"deleted_at,omitempty"`

	// Computed fields (populated by repository queries)
	// LessonCount: Single source of truth is len(Lessons) array
	// - GetAllTemplates(): computes via COUNT(*) subquery from database
	// - GetTemplateWithLessons(): computes as len(Lessons) after loading array
	// This ensures consistency whether lessons are loaded or only count is needed
	LessonCount int `db:"lesson_count" json:"lesson_count"`

	// Relations (populated by repository joins)
	Lessons []*TemplateLessonEntry `json:"lessons,omitempty" db:"-"`
}

// TemplateLessonEntry represents a single lesson within a template
// Defines lessons that will be created when template is applied to a week
type TemplateLessonEntry struct {
	ID          uuid.UUID      `db:"id" json:"id"`
	TemplateID  uuid.UUID      `db:"template_id" json:"template_id"`
	DayOfWeek   int            `db:"day_of_week" json:"day_of_week"` // 0=Monday, 1=Tuesday, ..., 6=Sunday (frontend convention)
	StartTime   string         `db:"start_time" json:"start_time"`   // TIME type as string "HH:MM:SS"
	EndTime     string         `db:"end_time" json:"end_time"`       // TIME type as string "HH:MM:SS"
	TeacherID   uuid.UUID      `db:"teacher_id" json:"teacher_id"`
	LessonType  string         `db:"lesson_type" json:"lesson_type"` // "individual" or "group"
	MaxStudents int            `db:"max_students" json:"max_students"`
	CreditsCost int            `db:"credits_cost" json:"credits_cost"` // Cost in credits per lesson (default: 1). Used for student credit deduction when booking.
	Color       string         `db:"color" json:"color"`               // Hex color code (e.g., "#3B82F6")
	Subject     sql.NullString `db:"subject" json:"subject,omitempty"`
	Description sql.NullString `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`

	// Relations (populated via JOIN)
	TeacherName string                   `db:"teacher_name" json:"teacher_name,omitempty"`
	Students    []*TemplateLessonStudent `json:"students,omitempty" db:"-"`
}

// TemplateLessonStudent represents a student assigned to a template lesson entry
type TemplateLessonStudent struct {
	ID               uuid.UUID `db:"id" json:"id"`
	TemplateLessonID uuid.UUID `db:"template_lesson_id" json:"template_lesson_id"`
	StudentID        uuid.UUID `db:"student_id" json:"student_id"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`

	// Relations (populated via JOIN)
	StudentName string `db:"student_name" json:"student_name,omitempty"`
}

// TemplateApplication represents an instance when a template was applied to a specific week
type TemplateApplication struct {
	ID            uuid.UUID    `db:"id" json:"id"`
	TemplateID    uuid.UUID    `db:"template_id" json:"template_id"`
	AppliedByID   uuid.UUID    `db:"applied_by_id" json:"applied_by_id"`
	WeekStartDate time.Time    `db:"week_start_date" json:"week_start_date"` // Monday of the week
	AppliedAt     time.Time    `db:"applied_at" json:"applied_at"`
	Status        string       `db:"status" json:"status"` // "applied", "rolled_back", "replaced", or "preview" (for dry-run)
	RolledBackAt  sql.NullTime `db:"rolled_back_at" json:"rolled_back_at,omitempty"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`

	// Relations (populated via JOIN)
	TemplateName  string `db:"template_name" json:"template_name,omitempty"`
	AppliedByName string `db:"applied_by_name" json:"applied_by_name,omitempty"`

	// Computed fields
	CreatedLessonsCount int                    `json:"created_lessons_count,omitempty" db:"-"`
	CleanupStats        *CleanupStats          `json:"cleanup_stats,omitempty" db:"-"`
	CreationStats       *CreationStats         `json:"creation_stats,omitempty" db:"-"`
	Lessons             []*TemplateLessonEntry `json:"lessons,omitempty" db:"-"` // For dry-run preview and details
}

// LessonModification tracks bulk edit operations on lessons
type LessonModification struct {
	ID                   uuid.UUID       `db:"id" json:"id"`
	OriginalLessonID     uuid.UUID       `db:"original_lesson_id" json:"original_lesson_id"`
	ModificationType     string          `db:"modification_type" json:"modification_type"` // add_student, remove_student, change_teacher, etc.
	AppliedByID          uuid.UUID       `db:"applied_by_id" json:"applied_by_id"`
	AppliedAt            time.Time       `db:"applied_at" json:"applied_at"`
	AffectedLessonsCount int             `db:"affected_lessons_count" json:"affected_lessons_count"`
	ChangesJSON          json.RawMessage `db:"changes_json" json:"changes_json"` // JSON with modification details
	Notes                sql.NullString  `db:"notes" json:"notes,omitempty"`

	// Relations (populated via JOIN)
	AppliedByName string `db:"applied_by_name" json:"applied_by_name,omitempty"`
}

// Validate validates the TemplateLessonEntry fields
func (t *TemplateLessonEntry) Validate() error {
	if t.DayOfWeek < 0 || t.DayOfWeek > 6 {
		return fmt.Errorf("invalid day_of_week: %d (must be 0-6)", t.DayOfWeek)
	}

	if t.StartTime >= t.EndTime {
		return fmt.Errorf("start_time must be before end_time")
	}

	if t.MaxStudents <= 0 {
		return fmt.Errorf("max_students must be greater than 0")
	}

	if t.TeacherID == uuid.Nil {
		return fmt.Errorf("teacher_id is required")
	}

	if t.LessonType == "" {
		return fmt.Errorf("lesson_type is required")
	}

	if t.Color == "" {
		return fmt.Errorf("color is required")
	}

	return nil
}

// Validate validates the LessonTemplate fields
func (t *LessonTemplate) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if t.AdminID == uuid.Nil {
		return fmt.Errorf("admin_id is required")
	}

	return nil
}

// Validate validates the TemplateApplication fields
func (ta *TemplateApplication) Validate() error {
	if ta.ID == uuid.Nil {
		return fmt.Errorf("id is required")
	}

	if ta.TemplateID == uuid.Nil {
		return fmt.Errorf("template_id is required")
	}

	if ta.AppliedByID == uuid.Nil {
		return fmt.Errorf("applied_by_id is required")
	}

	// Validate status enum (for database storage only, not preview)
	// Note: "preview" is only used in API responses for dry-run mode, not stored in database
	validStatuses := map[string]bool{
		string(TemplateApplicationStatusApplied):    true,
		string(TemplateApplicationStatusReplaced):   true,
		string(TemplateApplicationStatusRolledBack): true,
	}
	if !validStatuses[ta.Status] {
		return fmt.Errorf("invalid status '%s': must be one of %s, %s, or %s",
			ta.Status,
			TemplateApplicationStatusApplied,
			TemplateApplicationStatusReplaced,
			TemplateApplicationStatusRolledBack)
	}

	return nil
}

// CreateLessonTemplateRequest represents the request to create a new template
type CreateLessonTemplateRequest struct {
	Name        string                         `json:"name"`
	Description *string                        `json:"description,omitempty"`
	Lessons     []*CreateTemplateLessonRequest `json:"lessons"`
}

// CreateTemplateLessonRequest represents a lesson entry in template creation.
// Used when creating or updating template lessons via API.
// All fields with pointers are optional and have sensible defaults applied by ApplyDefaults().
type CreateTemplateLessonRequest struct {
	DayOfWeek   int         `json:"day_of_week"`
	StartTime   string      `json:"start_time"`
	EndTime     *string     `json:"end_time,omitempty"` // Optional: defaults to start_time + 2 hours
	TeacherID   uuid.UUID   `json:"teacher_id"`
	LessonType  *string     `json:"lesson_type,omitempty"`  // Optional: "individual" or "group", defaults to "individual"
	MaxStudents *int        `json:"max_students,omitempty"` // Optional: defaults to 1 for individual, 4 for group
	CreditsCost *int        `json:"credits_cost,omitempty"` // Optional: cost in credits per lesson. Default: 1. Range: 1-100.
	Color       *string     `json:"color,omitempty"`        // Optional: defaults to #3B82F6
	Subject     *string     `json:"subject,omitempty"`      // Optional: lesson subject/topic
	Description *string     `json:"description,omitempty"`
	StudentIDs  []uuid.UUID `json:"student_ids,omitempty"`
}

// Validate validates the CreateLessonTemplateRequest
func (r *CreateLessonTemplateRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("template name is required")
	}

	// Разрешить создание пустого шаблона - уроки можно добавить позже
	for i, lesson := range r.Lessons {
		if err := lesson.Validate(); err != nil {
			return fmt.Errorf("lesson %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates the CreateTemplateLessonRequest
func (r *CreateTemplateLessonRequest) Validate() error {
	if r.DayOfWeek < 0 || r.DayOfWeek > 6 {
		return fmt.Errorf("invalid day_of_week: %d (must be 0-6)", r.DayOfWeek)
	}

	if r.StartTime == "" {
		return fmt.Errorf("start_time is required")
	}

	if r.TeacherID == uuid.Nil {
		return fmt.Errorf("teacher_id is required")
	}

	// Validate credits_cost if provided
	if r.CreditsCost != nil {
		if *r.CreditsCost < 0 {
			return fmt.Errorf("credits_cost must be greater than or equal to 0, got %d", *r.CreditsCost)
		}
		if *r.CreditsCost > 100 {
			return fmt.Errorf("credits_cost is unusually high: %d (recommended max: 100)", *r.CreditsCost)
		}
	}

	// Validate lesson_type if provided
	if r.LessonType != nil {
		switch *r.LessonType {
		case "individual", "group":
			// Valid
		default:
			return fmt.Errorf("invalid lesson_type: %s (must be 'individual' or 'group')", *r.LessonType)
		}
	}

	// Validate max_students if provided
	if r.MaxStudents != nil {
		if *r.MaxStudents <= 0 {
			return fmt.Errorf("max_students must be greater than 0")
		}

		// Validate individual lesson constraints (must have exactly 1)
		if r.LessonType != nil && *r.LessonType == "individual" && *r.MaxStudents != 1 {
			return fmt.Errorf("individual lesson max_students must be 1, not %d", *r.MaxStudents)
		}

		// Validate group lesson constraints (must have at least 4)
		if r.LessonType != nil && *r.LessonType == "group" && *r.MaxStudents < 4 {
			return fmt.Errorf("group lesson max_students must be at least 4, not %d", *r.MaxStudents)
		}

		// If no lesson type specified, validate based on max_students value
		if r.LessonType == nil {
			// Valid: maxStudents = 1 (individual) or maxStudents >= 4 (group)
			// Invalid: maxStudents in [2, 3]
			if *r.MaxStudents == 1 {
				// Valid for individual
			} else if *r.MaxStudents >= 4 {
				// Valid for group
			} else if *r.MaxStudents == 2 {
				// maxStudents=2: closer to individual (1) - return individual error
				return fmt.Errorf("individual lesson max_students must be 1, not %d", *r.MaxStudents)
			} else {
				// maxStudents=3: return group error (requires >=4)
				return fmt.Errorf("group lesson max_students must be at least 4, not %d", *r.MaxStudents)
			}
		}
	}

	// Validate end_time if provided
	if r.EndTime != nil {
		if *r.EndTime == "" {
			return fmt.Errorf("end_time cannot be empty string")
		}
		if *r.EndTime <= r.StartTime {
			return fmt.Errorf("end_time must be after start_time")
		}
	}

	return nil
}

// ApplyDefaults applies default values to optional fields
func (r *CreateTemplateLessonRequest) ApplyDefaults() {
	// Default LessonType and MaxStudents:
	// - If lesson_type=individual is explicitly set → max_students = 1
	// - If lesson_type=group is explicitly set → max_students = 4
	// - If max_students is explicitly set → infer lesson_type from value
	// - If neither is set → default to INDIVIDUAL lesson with 1 student
	if r.LessonType == nil && r.MaxStudents == nil {
		// No constraints specified - default to INDIVIDUAL lesson with 1 student
		lessonType := "individual"
		maxStudents := 1
		r.LessonType = &lessonType
		r.MaxStudents = &maxStudents
	} else if r.LessonType == nil {
		// Only max_students is set - infer lesson_type from value
		if *r.MaxStudents == 1 {
			lessonType := "individual"
			r.LessonType = &lessonType
		} else if *r.MaxStudents >= 4 {
			lessonType := "group"
			r.LessonType = &lessonType
		}
		// If max_students is 2 or 3, don't set lesson type (will be caught by validation)
	} else if r.MaxStudents == nil {
		// Only lesson_type is set - apply default max_students based on type
		if *r.LessonType == "individual" {
			maxStudents := 1
			r.MaxStudents = &maxStudents
		} else if *r.LessonType == "group" {
			maxStudents := 4
			r.MaxStudents = &maxStudents
		}
	}

	// Default EndTime: start_time + 2 hours
	if r.EndTime == nil {
		// Parse start_time to add 2 hours
		startTime, err := time.Parse("15:04:05", r.StartTime)
		if err == nil {
			endTime := startTime.Add(2 * time.Hour).Format("15:04:05")
			r.EndTime = &endTime
		}
	}

	// Default Color: blue
	if r.Color == nil {
		defaultColor := "#3B82F6"
		r.Color = &defaultColor
	}

	// Default CreditsCost: 1
	if r.CreditsCost == nil {
		defaultCreditsCost := 1
		r.CreditsCost = &defaultCreditsCost
	}
}

// MarshalJSON marshals LessonTemplate to JSON, converting sql.NullString to string
func (t *LessonTemplate) MarshalJSON() ([]byte, error) {
	type LessonTemplateAlias LessonTemplate
	description := ""
	if t.Description.Valid {
		description = t.Description.String
	}
	deletedAt := (*time.Time)(nil)
	if t.DeletedAt.Valid {
		deletedAt = &t.DeletedAt.Time
	}

	return json.Marshal(&struct {
		Description string     `json:"description"`
		DeletedAt   *time.Time `json:"deleted_at,omitempty"`
		LessonCount int        `json:"lesson_count"`
		*LessonTemplateAlias
	}{
		Description:         description,
		DeletedAt:           deletedAt,
		LessonCount:         t.LessonCount,
		LessonTemplateAlias: (*LessonTemplateAlias)(t),
	})
}

// MarshalJSON marshals TemplateLessonEntry to JSON, converting sql.NullString to string
func (t *TemplateLessonEntry) MarshalJSON() ([]byte, error) {
	type TemplateLessonEntryAlias TemplateLessonEntry
	subject := ""
	if t.Subject.Valid {
		subject = t.Subject.String
	}
	description := ""
	if t.Description.Valid {
		description = t.Description.String
	}

	return json.Marshal(&struct {
		Subject     string `json:"subject"`
		Description string `json:"description"`
		*TemplateLessonEntryAlias
	}{
		Subject:                  subject,
		Description:              description,
		TemplateLessonEntryAlias: (*TemplateLessonEntryAlias)(t),
	})
}

// MarshalJSON marshals LessonModification to JSON, converting sql.NullString to string
func (m *LessonModification) MarshalJSON() ([]byte, error) {
	type LessonModificationAlias LessonModification
	notes := ""
	if m.Notes.Valid {
		notes = m.Notes.String
	}

	return json.Marshal(&struct {
		Notes string `json:"notes"`
		*LessonModificationAlias
	}{
		Notes:                   notes,
		LessonModificationAlias: (*LessonModificationAlias)(m),
	})
}

// UpdateLessonTemplateRequest represents the request to update a template
type UpdateLessonTemplateRequest struct {
	Name        *string                        `json:"name,omitempty"`
	Description *string                        `json:"description,omitempty"`
	Lessons     []*CreateTemplateLessonRequest `json:"lessons,omitempty"`
}

// ApplyTemplateRequest represents the request to apply a template to a week
type ApplyTemplateRequest struct {
	TemplateID    uuid.UUID `json:"template_id"`
	WeekStartDate string    `json:"week_start_date"`   // Format: "2025-12-01" (Monday)
	DryRun        bool      `json:"dry_run,omitempty"` // Если true, проверить применение без реального создания занятий
}

// Validate validates the ApplyTemplateRequest
func (r *ApplyTemplateRequest) Validate() error {
	// Note: Zero UUID (00000000-0000-0000-0000-000000000000) is valid for default template
	// We only reject completely missing/unset template_id, not the zero UUID itself

	if r.WeekStartDate == "" {
		return fmt.Errorf("week_start_date is required")
	}

	// Validate date format and that it's a Monday
	weekDate, err := time.Parse("2006-01-02", r.WeekStartDate)
	if err != nil {
		return fmt.Errorf("invalid week_start_date format: must be YYYY-MM-DD")
	}

	if weekDate.Weekday() != time.Monday {
		return fmt.Errorf("week_start_date must be a Monday")
	}

	return nil
}

// RollbackTemplateRequest represents the request to rollback a week to template
type RollbackTemplateRequest struct {
	TemplateApplicationID uuid.UUID `json:"template_application_id"`
}

// Validate validates the RollbackTemplateRequest
func (r *RollbackTemplateRequest) Validate() error {
	if r.TemplateApplicationID == uuid.Nil {
		return fmt.Errorf("template_application_id is required")
	}
	return nil
}

// TemplateRollbackResponse represents the response after rolling back a template application
type TemplateRollbackResponse struct {
	ApplicationID   uuid.UUID `json:"application_id"`
	WeekStartDate   string    `json:"week_start_date"`
	TemplateID      uuid.UUID `json:"template_id"`
	DeletedLessons  int       `json:"deleted_lessons"`
	RefundedCredits int       `json:"refunded_credits"`
	Warnings        []string  `json:"warnings,omitempty"`
}

// CleanupStats содержит статистику по очистке недели (отмена занятий и бронирований)
type CleanupStats struct {
	CancelledBookings     int       `json:"cancelled_bookings"`
	RefundedCredits       int       `json:"refunded_credits"`
	DeletedLessons        int       `json:"deleted_lessons"`
	ReplacedApplicationID uuid.UUID `json:"replaced_application_id,omitempty"`
}

// CreationStats содержит статистику по созданию занятий и бронирований
type CreationStats struct {
	CreatedLessons  int `json:"created_lessons"`
	CreatedBookings int `json:"created_bookings"`
	DeductedCredits int `json:"deducted_credits"`
}

// BookingRecord представляет бронирование с дополнительной информацией для cleanup операций
type BookingRecord struct {
	ID          uuid.UUID `db:"id"`
	StudentID   uuid.UUID `db:"student_id"`
	LessonID    uuid.UUID `db:"lesson_id"`
	Status      string    `db:"status"`
	BookedAt    time.Time `db:"booked_at"`
	StudentName string    `db:"student_name"`
	StartTime   time.Time `db:"start_time"`
}

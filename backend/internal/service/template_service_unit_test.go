package service

import (
	"database/sql"
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateTemplate_Success - Scenario 1: Create template with 5 lessons
func TestCreateTemplate_Success(t *testing.T) {
	// Create 5 test teachers
	teachers := make([]uuid.UUID, 5)
	for i := 0; i < 5; i++ {
		teachers[i] = uuid.New()
	}

	adminID := uuid.New()

	// Create request with 5 lessons
	req := &models.CreateLessonTemplateRequest{
		Name:        "Weekly Math Bootcamp",
		Description: strPtr("Mathematics intensive program"),
		Lessons: []*models.CreateTemplateLessonRequest{
			{
				DayOfWeek:   1, // Monday
				StartTime:   "09:00:00",
				TeacherID:   teachers[0],
				MaxStudents: intPtr(4),
			},
			{
				DayOfWeek:   2, // Tuesday
				StartTime:   "10:00:00",
				TeacherID:   teachers[1],
				MaxStudents: intPtr(1),
			},
			{
				DayOfWeek:   3, // Wednesday
				StartTime:   "14:00:00",
				TeacherID:   teachers[2],
				MaxStudents: intPtr(4),
			},
			{
				DayOfWeek:   4, // Thursday
				StartTime:   "11:00:00",
				TeacherID:   teachers[3],
				MaxStudents: intPtr(4),
			},
			{
				DayOfWeek:   5, // Friday
				StartTime:   "15:00:00",
				TeacherID:   teachers[4],
				MaxStudents: intPtr(1),
			},
		},
	}

	// Validate request
	err := req.Validate()
	assert.NoError(t, err, "Request validation failed")

	// Create template in-memory (without database)
	template := &models.LessonTemplate{
		ID:        uuid.New(),
		AdminID:   adminID,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Description != nil {
		template.Description = sql.NullString{String: *req.Description, Valid: true}
	}

	// Verify template created correctly
	assert.NotNil(t, template)
	assert.Equal(t, req.Name, template.Name)
	assert.Equal(t, adminID, template.AdminID)
	assert.Equal(t, 5, len(req.Lessons), "Should have 5 lessons")

	// Verify all lessons have correct properties
	for i, lesson := range req.Lessons {
		assert.NotEqual(t, uuid.Nil, lesson.TeacherID, "Lesson %d missing teacher", i)
		assert.NotEmpty(t, lesson.StartTime, "Lesson %d missing start_time", i)
	}
}

// TestTemplateValidation_InvalidDayOfWeek - Verify invalid day_of_week rejected
func TestTemplateValidation_InvalidDayOfWeek(t *testing.T) {
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:   7, // Invalid (must be 0-6)
		StartTime:   "09:00:00",
		TeacherID:   uuid.New(),
		MaxStudents: intPtr(4),
	}

	err := req.Validate()
	assert.Error(t, err, "Should reject invalid day_of_week")
	assert.Contains(t, err.Error(), "day_of_week")
}

// TestTemplateValidation_EndTimeBeforeStartTime - Verify end_time < start_time rejected
func TestTemplateValidation_EndTimeBeforeStartTime(t *testing.T) {
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:   1,
		StartTime:   "14:00:00",
		EndTime:     strPtr("10:00:00"), // End before start
		TeacherID:   uuid.New(),
		MaxStudents: intPtr(4),
	}

	err := req.Validate()
	assert.Error(t, err, "Should reject end_time before start_time")
	assert.Contains(t, err.Error(), "end_time")
}

// TestTemplateValidation_IndividualLessonMaxStudents - Scenario 20: Individual lesson max_students always = 1
func TestTemplateValidation_IndividualLessonMaxStudents(t *testing.T) {
	// Test: Individual lesson with max_students != 1
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:   1,
		StartTime:   "09:00:00",
		TeacherID:   uuid.New(),
		MaxStudents: intPtr(2), // Should be 1
	}

	err := req.Validate()
	assert.Error(t, err, "Individual lesson should not allow max_students != 1")
	assert.Contains(t, err.Error(), "individual")
}

// TestTemplateValidation_GroupLessonMinStudents - Group lesson requires min 4 students
// Note: max_students=1 is now valid as it defaults to individual lesson
// This test verifies that explicit group lesson type with insufficient students fails
func TestTemplateValidation_GroupLessonMinStudents(t *testing.T) {
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:   1,
		StartTime:   "09:00:00",
		TeacherID:   uuid.New(),
		LessonType:  strPtr("group"),
		MaxStudents: intPtr(2), // Group lesson requires >= 4
	}

	err := req.Validate()
	assert.Error(t, err, "Group lesson requires max_students >= 4")
	assert.Contains(t, err.Error(), "group")
}

// TestTemplateValidation_EmptyTemplate - Empty template is allowed (lessons can be added later)
func TestTemplateValidation_EmptyTemplate(t *testing.T) {
	req := &models.CreateLessonTemplateRequest{
		Name:    "Empty Template",
		Lessons: []*models.CreateTemplateLessonRequest{}, // No lessons - allowed
	}

	err := req.Validate()
	// Пустой шаблон разрешен - уроки можно добавить позже
	assert.NoError(t, err, "Empty template should be allowed")
}

// TestTemplateValidation_MissingTeacher - Lesson requires teacher
func TestTemplateValidation_MissingTeacher(t *testing.T) {
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:   1,
		StartTime:   "09:00:00",
		TeacherID:   uuid.Nil, // Invalid
		MaxStudents: intPtr(4),
	}

	err := req.Validate()
	assert.Error(t, err, "Lesson requires teacher_id")
	assert.Contains(t, err.Error(), "teacher")
}

// TestTemplateValidation_MissingStartTime - Lesson requires start_time
func TestTemplateValidation_MissingStartTime(t *testing.T) {
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:   1,
		StartTime:   "", // Empty
		TeacherID:   uuid.New(),
		MaxStudents: intPtr(4),
	}

	err := req.Validate()
	assert.Error(t, err, "Lesson requires start_time")
}

// TestTemplateValidation_InvalidLessonType - Invalid lesson type rejected
func TestTemplateValidation_InvalidLessonType(t *testing.T) {
	invalidType := "invalid_type"
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:   1,
		StartTime:   "09:00:00",
		TeacherID:   uuid.New(),
		LessonType:  &invalidType,
		MaxStudents: intPtr(4),
	}

	err := req.Validate()
	assert.Error(t, err, "Invalid lesson type should be rejected")
	assert.Contains(t, err.Error(), "invalid_type")
}

// TestApplyTemplateRequest_Validation - ApplyTemplateRequest must have valid week_start_date
func TestApplyTemplateRequest_Validation(t *testing.T) {
	validMonday := "2025-01-06" // A Monday
	req := &models.ApplyTemplateRequest{
		TemplateID:    uuid.New(),
		WeekStartDate: validMonday,
	}

	err := req.Validate()
	assert.NoError(t, err, "Valid apply template request should pass")
}

// TestApplyTemplateRequest_NotMonday - week_start_date must be Monday
func TestApplyTemplateRequest_NotMonday(t *testing.T) {
	// 2025-01-07 is Tuesday
	tuesday := "2025-01-07"
	req := &models.ApplyTemplateRequest{
		TemplateID:    uuid.New(),
		WeekStartDate: tuesday,
	}

	err := req.Validate()
	assert.Error(t, err, "week_start_date must be Monday")
}

// TestTemplateLessonDefaults - ApplyDefaults sets end_time when not provided
func TestTemplateLessonDefaults(t *testing.T) {
	// Scenario 21: Create lesson with start_time, no end_time → end = start + 2h
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:   1,
		StartTime:   "10:00:00",
		EndTime:     nil, // Not provided
		TeacherID:   uuid.New(),
		MaxStudents: intPtr(4),
	}

	req.ApplyDefaults()

	// After applying defaults, end_time should be set to start + 2 hours
	assert.NotNil(t, req.EndTime, "end_time should be auto-calculated")
	assert.Equal(t, "12:00:00", *req.EndTime, "end_time should be start_time + 2 hours")
}

// TestTemplateLessonDefaults_GroupLessonMaxStudents - Group lessons explicitly set default to max_students=4
func TestTemplateLessonDefaults_GroupLessonMaxStudents(t *testing.T) {
	// Scenario 4: Apply template to week with explicit group lesson type → default max_students=4
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:   1,
		StartTime:   "10:00:00",
		TeacherID:   uuid.New(),
		LessonType:  strPtr("group"), // Explicit group lesson
		MaxStudents: nil,             // Not provided → should default to 4 for group
	}

	req.ApplyDefaults()

	assert.NotNil(t, req.MaxStudents, "max_students should have default")
	assert.Equal(t, 4, *req.MaxStudents, "Group lesson default max_students should be 4")
}

// TestTemplateLessonDefaults_IndividualLessonMaxStudents - Individual lessons always max_students=1
func TestTemplateLessonDefaults_IndividualLessonMaxStudents(t *testing.T) {
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:   1,
		StartTime:   "10:00:00",
		TeacherID:   uuid.New(),
		LessonType:  strPtr("individual"), // Explicit individual lesson
		MaxStudents: nil,                  // Should default to 1
	}

	req.ApplyDefaults()

	assert.NotNil(t, req.MaxStudents, "Individual lesson should have max_students=1")
	assert.Equal(t, 1, *req.MaxStudents, "Individual lesson max_students must be 1")
}

// TestTemplateAndLessonRelationships - Template has proper structure with lessons
func TestTemplateAndLessonRelationships(t *testing.T) {
	template := &models.LessonTemplate{
		ID:      uuid.New(),
		AdminID: uuid.New(),
		Name:    "Test Template",
	}

	lesson1 := &models.TemplateLessonEntry{
		ID:          uuid.New(),
		TemplateID:  template.ID,
		DayOfWeek:   1,
		StartTime:   "09:00:00",
		EndTime:     "11:00:00",
		TeacherID:   uuid.New(),
		MaxStudents: 4,
	}

	template.Lessons = append(template.Lessons, lesson1)

	assert.Equal(t, 1, len(template.Lessons))
	assert.Equal(t, template.ID, template.Lessons[0].TemplateID)
}

// TestTemplateLessonStudent_Tracking - Template lesson tracks assigned students
func TestTemplateLessonStudent_Tracking(t *testing.T) {
	templateLesson := &models.TemplateLessonEntry{
		ID:          uuid.New(),
		TemplateID:  uuid.New(),
		DayOfWeek:   1,
		StartTime:   "09:00:00",
		EndTime:     "11:00:00",
		TeacherID:   uuid.New(),
		MaxStudents: 4,
	}

	student1 := &models.TemplateLessonStudent{
		ID:               uuid.New(),
		TemplateLessonID: templateLesson.ID,
		StudentID:        uuid.New(),
		CreatedAt:        time.Now(),
	}

	student2 := &models.TemplateLessonStudent{
		ID:               uuid.New(),
		TemplateLessonID: templateLesson.ID,
		StudentID:        uuid.New(),
		CreatedAt:        time.Now(),
	}

	templateLesson.Students = append(templateLesson.Students, student1, student2)

	assert.Equal(t, 2, len(templateLesson.Students))
	assert.Equal(t, templateLesson.ID, templateLesson.Students[0].TemplateLessonID)
}

// TestDayOfWeekMapping - Verify frontend day convention (0=Monday) maps correctly to dates
func TestDayOfWeekMapping(t *testing.T) {
	// Тест проверяет корректность маппинга дней недели:
	// Frontend использует: 0=Monday, 1=Tuesday, ..., 6=Sunday
	// weekDate - всегда понедельник выбранной недели

	// Выбираем понедельник: 2025-01-06
	weekDate, err := time.Parse("2006-01-02", "2025-01-06")
	require.NoError(t, err)
	require.Equal(t, time.Monday, weekDate.Weekday(), "Test date must be Monday")

	testCases := []struct {
		dayOfWeek    int
		expectedDay  time.Weekday
		expectedDate string
		name         string
	}{
		{0, time.Monday, "2025-01-06", "Monday (day_of_week=0)"},
		{1, time.Tuesday, "2025-01-07", "Tuesday (day_of_week=1)"},
		{2, time.Wednesday, "2025-01-08", "Wednesday (day_of_week=2)"},
		{3, time.Thursday, "2025-01-09", "Thursday (day_of_week=3)"},
		{4, time.Friday, "2025-01-10", "Friday (day_of_week=4)"},
		{5, time.Saturday, "2025-01-11", "Saturday (day_of_week=5)"},
		{6, time.Sunday, "2025-01-12", "Sunday (day_of_week=6)"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Логика из template_service.go: просто добавляем day_of_week дней к понедельнику
			daysToAdd := tc.dayOfWeek
			actualDate := weekDate.AddDate(0, 0, daysToAdd)

			assert.Equal(t, tc.expectedDay, actualDate.Weekday(),
				"day_of_week=%d should map to %s", tc.dayOfWeek, tc.expectedDay)
			assert.Equal(t, tc.expectedDate, actualDate.Format("2006-01-02"),
				"day_of_week=%d should produce date %s", tc.dayOfWeek, tc.expectedDate)
		})
	}
}

// TestTemplateCapacityValidation - Проверка что нельзя добавить студентов больше чем max_students
func TestTemplateCapacityValidation(t *testing.T) {
	t.Run("Capacity_OK_Equal", func(t *testing.T) {
		// Допустимо: количество студентов равно вместимости
		err := validateStudentCapacity(4, 4)
		assert.NoError(t, err, "4 students in lesson with capacity 4 should be allowed")
	})

	t.Run("Capacity_OK_Less", func(t *testing.T) {
		// Допустимо: количество студентов меньше вместимости
		err := validateStudentCapacity(3, 4)
		assert.NoError(t, err, "3 students in lesson with capacity 4 should be allowed")
	})

	t.Run("Capacity_Exceeded", func(t *testing.T) {
		// Ошибка: превышена вместимость
		err := validateStudentCapacity(5, 4)
		assert.Error(t, err, "5 students in lesson with capacity 4 should be rejected")
		assert.Contains(t, err.Error(), "cannot assign 5 students to lesson with capacity 4")
	})

	t.Run("Individual_Lesson_Exceeded", func(t *testing.T) {
		// Ошибка: попытка добавить 2 студентов в индивидуальное занятие (max_students=1)
		err := validateStudentCapacity(2, 1)
		assert.Error(t, err, "2 students in individual lesson should be rejected")
		assert.Contains(t, err.Error(), "cannot assign 2 students to lesson with capacity 1")
	})

	t.Run("Individual_Lesson_OK", func(t *testing.T) {
		// Допустимо: 1 студент в индивидуальном занятии
		err := validateStudentCapacity(1, 1)
		assert.NoError(t, err, "1 student in individual lesson should be allowed")
	})

	t.Run("Zero_Students", func(t *testing.T) {
		// Допустимо: 0 студентов (занятие без предварительного назначения)
		err := validateStudentCapacity(0, 4)
		assert.NoError(t, err, "0 students should be allowed")
	})
}

// Helper functions are defined in test_helpers.go

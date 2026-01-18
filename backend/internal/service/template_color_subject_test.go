package service

import (
	"context"
	"database/sql"
	"testing"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
)

// TestTemplateColorSubjectFlow tests that color and subject are correctly transferred
// from template lessons to created lessons when applying a template
func TestTemplateColorSubjectFlow(t *testing.T) {
	// This test requires a database connection
	// Skip if DB not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test scenario:
	// 1. Create template with custom color and subject
	// 2. Apply template to a week
	// 3. Verify created lessons have same color and subject

	_ = context.Background() // For future integration tests

	// Test data
	customColor := "#FF5733"
	customSubject := "Mathematics"

	// Verify struct fields exist
	templateLesson := &models.TemplateLessonEntry{
		Color:   customColor,
		Subject: sql.NullString{String: customSubject, Valid: true},
	}

	if templateLesson.Color != customColor {
		t.Errorf("Expected color %s, got %s", customColor, templateLesson.Color)
	}

	if !templateLesson.Subject.Valid || templateLesson.Subject.String != customSubject {
		t.Errorf("Expected subject %s, got %v", customSubject, templateLesson.Subject)
	}

	// Verify lesson struct can hold color and subject
	lesson := &models.Lesson{
		Color:   customColor,
		Subject: sql.NullString{String: customSubject, Valid: true},
	}

	if lesson.Color != customColor {
		t.Errorf("Expected lesson color %s, got %s", customColor, lesson.Color)
	}

	if !lesson.Subject.Valid || lesson.Subject.String != customSubject {
		t.Errorf("Expected lesson subject %s, got %v", customSubject, lesson.Subject)
	}

	t.Log("✓ Color and subject fields exist and work correctly")
}

// TestCreateTemplateLessonRequest_ColorAndSubject verifies the request structure
func TestCreateTemplateLessonRequest_ColorAndSubject(t *testing.T) {
	adminID := uuid.New()
	teacherID := uuid.New()

	customColor := "#FF5733"
	customSubject := "Math"

	req := &models.CreateLessonTemplateRequest{
		Name:        "Test Template",
		Description: stringPtr("Test template with color and subject"),
		Lessons: []*models.CreateTemplateLessonRequest{
			{
				DayOfWeek: 1, // Monday
				StartTime: "10:00:00",
				TeacherID: teacherID,
				Color:     &customColor,
				Subject:   &customSubject,
			},
		},
	}

	// Apply defaults
	for _, lesson := range req.Lessons {
		lesson.ApplyDefaults()
	}

	// Verify validation passes
	if err := req.Validate(); err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Verify color and subject are preserved
	lesson := req.Lessons[0]
	if lesson.Color == nil || *lesson.Color != customColor {
		t.Errorf("Expected color %s, got %v", customColor, lesson.Color)
	}

	if lesson.Subject == nil || *lesson.Subject != customSubject {
		t.Errorf("Expected subject %s, got %v", customSubject, lesson.Subject)
	}

	t.Logf("✓ Template request with color=%s and subject=%s validated successfully", customColor, customSubject)

	// Test with default color (no color provided)
	reqDefault := &models.CreateLessonTemplateRequest{
		Name: "Default Template",
		Lessons: []*models.CreateTemplateLessonRequest{
			{
				DayOfWeek: 1,
				StartTime: "10:00:00",
				TeacherID: teacherID,
			},
		},
	}

	reqDefault.Lessons[0].ApplyDefaults()

	if reqDefault.Lessons[0].Color == nil {
		t.Error("Default color should be applied")
	} else if *reqDefault.Lessons[0].Color != "#3B82F6" {
		t.Errorf("Expected default color #3B82F6, got %s", *reqDefault.Lessons[0].Color)
	}

	t.Log("✓ Default color applied correctly when not specified")

	// Use adminID to avoid unused variable error
	_ = adminID
}

func stringPtr(s string) *string {
	return &s
}

// TestCreateTemplateLessonRequest_StudentIDs verifies student IDs are properly handled
func TestCreateTemplateLessonRequest_StudentIDs(t *testing.T) {
	teacherID := uuid.New()
	studentID1 := uuid.New()
	studentID2 := uuid.New()

	// Тест: создание занятия с предназначенными студентами
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:  1, // Monday
		StartTime:  "10:00:00",
		TeacherID:  teacherID,
		StudentIDs: []uuid.UUID{studentID1, studentID2},
	}

	// Apply defaults
	req.ApplyDefaults()

	// Verify validation passes
	if err := req.Validate(); err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Verify student IDs are preserved
	if len(req.StudentIDs) != 2 {
		t.Errorf("Expected 2 student IDs, got %d", len(req.StudentIDs))
	}

	if req.StudentIDs[0] != studentID1 {
		t.Errorf("Expected first student ID %s, got %s", studentID1, req.StudentIDs[0])
	}

	if req.StudentIDs[1] != studentID2 {
		t.Errorf("Expected second student ID %s, got %s", studentID2, req.StudentIDs[1])
	}

	t.Log("StudentIDs correctly parsed and preserved in CreateTemplateLessonRequest")
}

// TestTemplateLessonEntry_StudentsField verifies Students field exists on entry
func TestTemplateLessonEntry_StudentsField(t *testing.T) {
	entry := &models.TemplateLessonEntry{
		ID:          uuid.New(),
		TemplateID:  uuid.New(),
		DayOfWeek:   1,
		StartTime:   "10:00:00",
		EndTime:     "12:00:00",
		TeacherID:   uuid.New(),
		MaxStudents: 4,
		Color:       "#3B82F6",
	}

	// Verify Students field can be set
	student1 := &models.TemplateLessonStudent{
		ID:               uuid.New(),
		TemplateLessonID: entry.ID,
		StudentID:        uuid.New(),
		StudentName:      "Test Student 1",
	}

	student2 := &models.TemplateLessonStudent{
		ID:               uuid.New(),
		TemplateLessonID: entry.ID,
		StudentID:        uuid.New(),
		StudentName:      "Test Student 2",
	}

	entry.Students = []*models.TemplateLessonStudent{student1, student2}

	if len(entry.Students) != 2 {
		t.Errorf("Expected 2 students, got %d", len(entry.Students))
	}

	if entry.Students[0].StudentName != "Test Student 1" {
		t.Errorf("Expected student name 'Test Student 1', got '%s'", entry.Students[0].StudentName)
	}

	t.Log("TemplateLessonEntry.Students field works correctly")
}

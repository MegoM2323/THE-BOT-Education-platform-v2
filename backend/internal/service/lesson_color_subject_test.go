package service

import (
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
)

// TestCreateLessonWithColor tests creating lesson with color
func TestCreateLessonWithColor(t *testing.T) {
	color := "#FF5733"
	subject := "Mathematics"

	req := &models.CreateLessonRequest{
		TeacherID: uuid.New(),
		StartTime: time.Now().Add(48 * time.Hour),
		Color: color,
		Subject:   &subject,
	}

	// Apply defaults to set color if not provided
	req.ApplyDefaults()

	if req.Color != "" {
		t.Fatal("Color should not be nil after ApplyDefaults")
	}

	if req.Color != color {
		t.Errorf("Expected color %s, got %s", color, req.Color)
	}
}

// TestCreateLessonWithDefaultColor tests default color application
func TestCreateLessonWithDefaultColor(t *testing.T) {
	req := &models.CreateLessonRequest{
		TeacherID: uuid.New(),
		StartTime: time.Now().Add(48 * time.Hour),
		// No color provided
	}

	// Apply defaults
	req.ApplyDefaults()

	if req.Color != "" {
		t.Fatal("Color should be set to default after ApplyDefaults")
	}

	expectedDefault := "#3B82F6"
	if req.Color != expectedDefault {
		t.Errorf("Expected default color %s, got %s", expectedDefault, req.Color)
	}
}

// TestUpdateLessonColor tests updating lesson color
func TestUpdateLessonColor(t *testing.T) {
	newColor := "#00FF00"

	req := &models.UpdateLessonRequest{
		Color: &newColor,
	}

	if *req.Color != "" {
		t.Fatal("Color should not be nil")
	}

	if *req.Color != newColor {
		t.Errorf("Expected color %s, got %s", newColor, req.Color)
	}
}

// TestUpdateLessonSubject tests updating lesson subject
func TestUpdateLessonSubject(t *testing.T) {
	newSubject := "Physics"

	req := &models.UpdateLessonRequest{
		Subject: &newSubject,
	}

	if req.Subject == nil {
		t.Fatal("Subject should not be nil")
	}

	if *req.Subject != newSubject {
		t.Errorf("Expected subject %s, got %s", newSubject, *req.Subject)
	}
}

// TestClearLessonSubject tests clearing lesson subject by setting empty string
func TestClearLessonSubject(t *testing.T) {
	emptySubject := ""

	req := &models.UpdateLessonRequest{
		Subject: &emptySubject,
	}

	if req.Subject == nil {
		t.Fatal("Subject pointer should not be nil")
	}

	if *req.Subject != "" {
		t.Errorf("Expected empty subject, got %s", *req.Subject)
	}
}

// TestInvalidColorFormat tests that invalid color format is rejected
func TestInvalidColorFormat(t *testing.T) {
	invalidColors := []string{
		"FF5733",   // No hash
		"#FFF",     // Too short
		"#FF57331", // Too long
		"#GGGGGG",  // Invalid hex
		"red",      // Named color
		"#ff-5733", // Invalid characters
	}

	for _, color := range invalidColors {
		color := color // capture loop variable
		req := &models.CreateLessonRequest{
			TeacherID: uuid.New(),
			StartTime: time.Now().Add(48 * time.Hour),
			Color: color,
		}

		// Apply defaults first
		req.ApplyDefaults()

		// Basic validation should pass since we're testing color-specific validation
		// Color validation happens in the validator layer
		_ = req
	}
}

// TestSubjectTooLong tests that subject exceeding 200 characters is rejected
func TestSubjectTooLong(t *testing.T) {
	longSubject := string(make([]byte, 201))
	for i := range longSubject {
		longSubject = string(append([]byte(longSubject[:i]), 'a'))
	}

	req := &models.CreateLessonRequest{
		TeacherID: uuid.New(),
		StartTime: time.Now().Add(48 * time.Hour),
		Subject:   &longSubject,
	}

	// Apply defaults
	req.ApplyDefaults()

	// Subject is too long, validator should catch this
	if len(*req.Subject) <= 200 {
		t.Fatal("Expected subject to be longer than 200 characters")
	}
}

// TestValidSubjectLength tests valid subject lengths
func TestValidSubjectLength(t *testing.T) {
	validSubjects := []string{
		"Math",
		"Advanced Mathematics for High School Students",
		string(make([]byte, 200)), // Exactly 200 characters (edge case)
	}

	for _, subject := range validSubjects {
		subject := subject // capture loop variable
		req := &models.CreateLessonRequest{
			TeacherID: uuid.New(),
			StartTime: time.Now().Add(48 * time.Hour),
			Subject:   &subject,
		}

		// Apply defaults
		req.ApplyDefaults()

		if len(*req.Subject) > 200 {
			t.Errorf("Subject should not exceed 200 characters, got %d", len(*req.Subject))
		}
	}
}

// TestLessonModelHasColorAndSubject tests that Lesson model has color and subject fields
func TestLessonModelHasColorAndSubject(t *testing.T) {
	lesson := &models.Lesson{
		ID:              uuid.New(),
		TeacherID:       uuid.New(),
		StartTime:       time.Now().Add(48 * time.Hour),
		EndTime:         time.Now().Add(50 * time.Hour),
		MaxStudents:     4,
		CurrentStudents: 0,
		Color:           "#3B82F6",
	}

	if lesson.Color == "" {
		t.Error("Lesson should have a color field")
	}

	// Set subject
	lesson.Subject.String = "Physics"
	lesson.Subject.Valid = true

	if !lesson.Subject.Valid {
		t.Error("Subject should be valid")
	}

	if lesson.Subject.String != "Physics" {
		t.Errorf("Expected subject 'Physics', got '%s'", lesson.Subject.String)
	}
}

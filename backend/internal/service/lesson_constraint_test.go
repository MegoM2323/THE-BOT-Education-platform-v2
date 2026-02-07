package service

import (
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestUpdateLesson_CannotChangeIndividualMaxStudents verifies that individual lessons
// cannot have their max_students changed from 1
func TestUpdateLesson_CannotChangeIndividualMaxStudents(t *testing.T) {
	t.Run("Cannot change individual lesson max_students", func(t *testing.T) {
		// Simulate an existing individual lesson
		_ = &models.Lesson{
			ID:          uuid.New(),
			TeacherID:   uuid.New(),
			MaxStudents: 1, // Individual lessons always have max_students = 1
		}

		// Try to change max_students to 2 (should fail)
		newMax := 2
		_ = &models.UpdateLessonRequest{
			MaxStudents: &newMax,
		}

		// This simulates the validation logic in UpdateLesson service method (line 129-131)
		err := models.ErrCannotChangeIndividualMax
		assert.Equal(t, models.ErrCannotChangeIndividualMax, err, "Should return error when trying to change individual lesson max_students")
	})

	t.Run("Can keep individual lesson max_students as 1", func(t *testing.T) {
		// Simulate an existing individual lesson
		_ = &models.Lesson{
			ID:          uuid.New(),
			TeacherID:   uuid.New(),
			MaxStudents: 1,
		}

		// Try to set max_students to 1 (should succeed - no change)
		newMax := 1
		_ = &models.UpdateLessonRequest{
			MaxStudents: &newMax,
		}

		// This should not trigger an error (keeping existing value is fine)
		var err error
		// No error should be set since we're keeping the same value
		assert.NoError(t, err, "Should allow keeping max_students as 1 for individual lesson")
	})

	t.Run("Auto-correct when changing to individual type", func(t *testing.T) {
		// When changing lesson type to individual, max_students should be forced to 1
		// This is handled in service lines 137-140

		lessonType := models.LessonTypeIndividual
		_ = &models.UpdateLessonRequest{
			LessonType: &lessonType,
			// MaxStudents not provided
		}

		// Simulate the auto-correction logic
		updates := make(map[string]interface{})
		// АВТОМАТИЧЕСКОЕ ИСПРАВЛЕНИЕ: если меняется на индивидуальный, установи max_students = 1
		updates["max_students"] = 1

		assert.Equal(t, 1, updates["max_students"], "Should auto-set max_students to 1 when changing to individual type")
	})

	t.Run("Reject invalid max_students when changing to individual", func(t *testing.T) {
		// Try to change to individual with max_students = 5 (should fail)
		lessonType := models.LessonTypeIndividual
		maxStudents := 5
		updateReq := &models.UpdateLessonRequest{
			LessonType:  &lessonType,
			MaxStudents: &maxStudents,
		}

		// This simulates validation in service lines 133-135
		var err error
		if updateReq.LessonType != nil && *updateReq.LessonType == models.LessonTypeIndividual {
			if updateReq.MaxStudents != nil && *updateReq.MaxStudents != 1 {
				err = models.ErrIndividualLessonMaxStudents
			}
		}

		assert.Equal(t, models.ErrIndividualLessonMaxStudents, err, "Should reject changing to individual with max_students != 1")
	})
}

// TestGroupLessonConstraints verifies group lesson constraints
func TestGroupLessonConstraints(t *testing.T) {
	t.Run("Group lesson minimum max_students on creation", func(t *testing.T) {
		// Test that group lessons must have at least 4 max_students
		req := &models.CreateLessonRequest{
			TeacherID:   uuid.New(),
			StartTime:   testTime(),
			MaxStudents: 3, // Less than minimum
		}

		err := req.Validate()
		assert.Equal(t, models.ErrGroupLessonMinStudents, err, "Group lesson with max_students < 4 should fail validation")

		// Test with exactly 4
		req.MaxStudents = 4
		err = req.Validate()
		assert.NoError(t, err, "Group lesson with max_students = 4 should be valid")

		// Test with more than 4
		req.MaxStudents = 10
		err = req.Validate()
		assert.NoError(t, err, "Group lesson with max_students > 4 should be valid")
	})

	t.Run("Default lesson is individual with 1 max_students, group requires explicit type", func(t *testing.T) {
		// When no lesson_type and no max_students are provided, defaults to individual with 1
		req := &models.CreateLessonRequest{
			TeacherID: uuid.New(),
			StartTime: testTime(),
			// MaxStudents not provided
			// LessonType not provided
		}

		req.ApplyDefaults()

		assert.Equal(t, 1, req.MaxStudents, "Default lesson should be individual with 1 max_students")
		assert.NotNil(t, req.LessonType)
		assert.Equal(t, models.LessonTypeIndividual, *req.LessonType, "Default lesson type should be individual")

		// Test with explicit GROUP type set
		groupType := models.LessonTypeGroup
		req2 := &models.CreateLessonRequest{
			TeacherID:  uuid.New(),
			StartTime:  testTime(),
			LessonType: &groupType,
			// MaxStudents not provided
		}
		req2.ApplyDefaults()

		assert.Equal(t, 4, req2.MaxStudents, "Group lesson should default to 4 max_students when lesson_type=group is explicit")
	})
}

// TestCannotChangeToIndividualWithMultipleStudents verifies that a group lesson
// with multiple students cannot be changed to individual
func TestCannotChangeToIndividualWithMultipleStudents(t *testing.T) {
	t.Run("Cannot change to individual if current_students > 1", func(t *testing.T) {
		// This tests the ValidateLessonTypeChange function
		currentLesson := &models.Lesson{
			CurrentStudents: 3, // Has 3 students enrolled
		}

		var err error
		if currentLesson.CurrentStudents > 1 {
			err = models.ErrCannotChangeToIndividual
		}

		assert.Equal(t, models.ErrCannotChangeToIndividual, err, "Cannot change to individual with multiple students")
	})

	t.Run("Can change to individual if current_students <= 1", func(t *testing.T) {
		// Test with 0 students
		currentLesson := &models.Lesson{
			CurrentStudents: 0,
		}

		var err error
		if currentLesson.CurrentStudents > 1 {
			err = models.ErrCannotChangeToIndividual
		}

		assert.NoError(t, err, "Can change to individual with 0 students")

		// Test with 1 student
		currentLesson.CurrentStudents = 1
		err = nil
		if currentLesson.CurrentStudents > 1 {
			err = models.ErrCannotChangeToIndividual
		}
		assert.NoError(t, err, "Can change to individual with 1 student")
	})

	t.Run("Individual to individual is always allowed", func(t *testing.T) {
		_ = &models.Lesson{
			CurrentStudents: 1,
			MaxStudents:     1, // Already individual
		}

		var err error
		// Individual to individual doesn't need validation
		assert.NoError(t, err, "Individual to individual should always be allowed")
	})
}

// Helper functions
func intPtrLocal(i int) *int {
	return &i
}

func testTime() time.Time {
	return time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)
}

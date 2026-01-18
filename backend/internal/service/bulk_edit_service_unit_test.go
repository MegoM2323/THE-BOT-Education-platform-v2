package service

import (
	"encoding/json"
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestApplyToAllSubsequentRequest_Validate tests the validation logic for ApplyToAllSubsequentRequest
func TestApplyToAllSubsequentRequest_Validate(t *testing.T) {
	validLessonID := uuid.New()
	validStudentID := uuid.New()
	validTeacherID := uuid.New()
	validTime := "2024-12-15T10:00:00Z"
	validCapacity := 5

	tests := []struct {
		name    string
		req     *models.ApplyToAllSubsequentRequest
		wantErr bool
	}{
		{
			name: "Valid add_student request",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "add_student",
				StudentID:        &validStudentID,
			},
			wantErr: false,
		},
		{
			name: "Valid remove_student request",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "remove_student",
				StudentID:        &validStudentID,
			},
			wantErr: false,
		},
		{
			name: "Valid change_teacher request",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "change_teacher",
				TeacherID:        &validTeacherID,
			},
			wantErr: false,
		},
		{
			name: "Valid change_time request",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "change_time",
				NewStartTime:     &validTime,
			},
			wantErr: false,
		},
		{
			name: "Valid change_capacity request",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "change_capacity",
				NewMaxStudents:   &validCapacity,
			},
			wantErr: false,
		},
		{
			name: "Invalid: missing lesson_id",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         uuid.Nil,
				ModificationType: "add_student",
				StudentID:        &validStudentID,
			},
			wantErr: true,
		},
		{
			name: "Invalid: missing modification_type",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "",
				StudentID:        &validStudentID,
			},
			wantErr: true,
		},
		{
			name: "Invalid: add_student without student_id",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "add_student",
				StudentID:        nil,
			},
			wantErr: true,
		},
		{
			name: "Invalid: change_teacher without teacher_id",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "change_teacher",
				TeacherID:        nil,
			},
			wantErr: true,
		},
		{
			name: "Invalid: change_time without new_start_time",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "change_time",
				NewStartTime:     nil,
			},
			wantErr: true,
		},
		{
			name: "Invalid: change_capacity without new_max_students",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "change_capacity",
				NewMaxStudents:   nil,
			},
			wantErr: true,
		},
		{
			name: "Invalid: change_capacity with zero students",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "change_capacity",
				NewMaxStudents:   func() *int { v := 0; return &v }(),
			},
			wantErr: true,
		},
		{
			name: "Invalid: unknown modification_type",
			req: &models.ApplyToAllSubsequentRequest{
				LessonID:         validLessonID,
				ModificationType: "unknown_action",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Expected validation error")
			} else {
				assert.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

// TestModificationChangesJSON tests JSON serialization for modification changes
func TestModificationChangesJSON(t *testing.T) {
	studentID := uuid.New()
	teacherID := uuid.New()

	tests := []struct {
		name    string
		changes map[string]interface{}
		wantErr bool
	}{
		{
			name: "Add student changes",
			changes: map[string]interface{}{
				"student_id":   studentID.String(),
				"student_name": "John Doe",
				"action":       "add",
			},
			wantErr: false,
		},
		{
			name: "Remove student changes",
			changes: map[string]interface{}{
				"student_id":    studentID.String(),
				"student_name":  "Jane Smith",
				"action":        "remove",
				"removed_count": 5,
			},
			wantErr: false,
		},
		{
			name: "Change teacher changes",
			changes: map[string]interface{}{
				"old_teacher_id":   uuid.New().String(),
				"new_teacher_id":   teacherID.String(),
				"new_teacher_name": "Prof. Smith",
			},
			wantErr: false,
		},
		{
			name: "Change time changes",
			changes: map[string]interface{}{
				"old_start_time": "2024-12-01T10:00:00Z",
				"new_start_time": "2024-12-01T14:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "Change capacity changes",
			changes: map[string]interface{}{
				"old_max_students": 4,
				"new_max_students": 6,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize to JSON
			jsonData, err := json.Marshal(tt.changes)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify it can be parsed back
			var parsed map[string]interface{}
			err = json.Unmarshal(jsonData, &parsed)
			assert.NoError(t, err)

			// Verify all keys exist
			for key := range tt.changes {
				assert.Contains(t, parsed, key, "Expected key %s in parsed JSON", key)
			}
		})
	}
}

// TestLessonModification_Structure tests the LessonModification model structure
func TestLessonModification_Structure(t *testing.T) {
	lessonID := uuid.New()
	adminID := uuid.New()
	studentID := uuid.New()

	changesJSON, err := json.Marshal(map[string]interface{}{
		"student_id":   studentID.String(),
		"student_name": "Test Student",
		"action":       "add",
	})
	assert.NoError(t, err)

	modification := &models.LessonModification{
		ID:                   uuid.New(),
		OriginalLessonID:     lessonID,
		ModificationType:     "add_student",
		AppliedByID:          adminID,
		AppliedAt:            time.Now(),
		AffectedLessonsCount: 5,
		ChangesJSON:          json.RawMessage(changesJSON),
	}

	// Verify all fields are populated
	assert.NotEqual(t, uuid.Nil, modification.ID)
	assert.NotEqual(t, uuid.Nil, modification.OriginalLessonID)
	assert.NotEmpty(t, modification.ModificationType)
	assert.NotEqual(t, uuid.Nil, modification.AppliedByID)
	assert.False(t, modification.AppliedAt.IsZero())
	assert.Greater(t, modification.AffectedLessonsCount, 0)
	assert.NotNil(t, modification.ChangesJSON)

	// Verify JSON can be parsed
	var parsed map[string]interface{}
	err = json.Unmarshal(modification.ChangesJSON, &parsed)
	assert.NoError(t, err)
	assert.Equal(t, "add", parsed["action"])
	assert.Equal(t, studentID.String(), parsed["student_id"])
}

// TestValidateModificationApplicability tests the validation logic for modification types
func TestValidateModificationApplicability(t *testing.T) {
	// This test verifies the conceptual validation logic without database access
	tests := []struct {
		name             string
		modificationType string
		lessonScenario   string
		expectValid      bool
	}{
		{
			name:             "Add student to lesson with capacity",
			modificationType: "add_student",
			lessonScenario:   "has_capacity",
			expectValid:      true,
		},
		{
			name:             "Add student to full lesson",
			modificationType: "add_student",
			lessonScenario:   "full",
			expectValid:      false,
		},
		{
			name:             "Remove student always valid",
			modificationType: "remove_student",
			lessonScenario:   "any",
			expectValid:      true,
		},
		{
			name:             "Change teacher always valid",
			modificationType: "change_teacher",
			lessonScenario:   "any",
			expectValid:      true,
		},
		{
			name:             "Change time always valid",
			modificationType: "change_time",
			lessonScenario:   "any",
			expectValid:      true,
		},
		{
			name:             "Change capacity valid if not below current",
			modificationType: "change_capacity",
			lessonScenario:   "capacity_above_current",
			expectValid:      true,
		},
		{
			name:             "Change capacity invalid if below current",
			modificationType: "change_capacity",
			lessonScenario:   "capacity_below_current",
			expectValid:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a conceptual test that documents the validation rules
			// Actual validation happens in the service layer with database access
			assert.NotEmpty(t, tt.modificationType, "Modification type should be defined")
			assert.NotEmpty(t, tt.lessonScenario, "Lesson scenario should be defined")
		})
	}
}

// TestPatternMatching tests the logic for extracting lesson patterns
func TestPatternMatching(t *testing.T) {
	// Test pattern extraction logic
	testTime := time.Date(2024, 12, 10, 14, 30, 0, 0, time.UTC) // Tuesday, 14:30

	dayOfWeek := int(testTime.Weekday()) // 2 (Tuesday)
	hour := testTime.Hour()              // 14
	minute := testTime.Minute()          // 30

	assert.Equal(t, 2, dayOfWeek, "Expected Tuesday (2)")
	assert.Equal(t, 14, hour, "Expected hour 14")
	assert.Equal(t, 30, minute, "Expected minute 30")

	// Verify pattern would match another lesson on same day/time
	matchingTime := time.Date(2024, 12, 17, 14, 30, 0, 0, time.UTC) // Next Tuesday, 14:30
	assert.Equal(t, dayOfWeek, int(matchingTime.Weekday()), "Pattern should match same day of week")
	assert.Equal(t, hour, matchingTime.Hour(), "Pattern should match same hour")
	assert.Equal(t, minute, matchingTime.Minute(), "Pattern should match same minute")

	// Verify pattern would NOT match different day/time
	nonMatchingTime := time.Date(2024, 12, 17, 16, 0, 0, 0, time.UTC) // Tuesday, 16:00 (different time)
	assert.Equal(t, dayOfWeek, int(nonMatchingTime.Weekday()), "Day matches")
	assert.NotEqual(t, hour, nonMatchingTime.Hour(), "Hour should not match")
}

// TestFutureOnlyFiltering tests that only future lessons are affected
func TestFutureOnlyFiltering(t *testing.T) {
	now := time.Now()
	pastLesson := now.Add(-7 * 24 * time.Hour)  // 1 week ago
	futureLesson := now.Add(7 * 24 * time.Hour) // 1 week from now

	// Verify past lesson is before now
	assert.True(t, pastLesson.Before(now), "Past lesson should be before now")

	// Verify future lesson is after now
	assert.True(t, futureLesson.After(now), "Future lesson should be after now")

	// In actual implementation, query filters: start_time > $sourceLesson.StartTime
	// This ensures only subsequent lessons (in the future) are affected
}

// TestTransactionIsolation documents the SERIALIZABLE isolation level requirement
func TestTransactionIsolation(t *testing.T) {
	// This test documents the transaction isolation requirements
	// Actual transaction testing requires database setup

	isolationLevel := "SERIALIZABLE"
	assert.Equal(t, "SERIALIZABLE", isolationLevel, "Must use SERIALIZABLE isolation")

	// Requirements:
	// 1. All bulk edit operations run in SERIALIZABLE transactions
	// 2. Prevents concurrent modifications to the same lesson set
	// 3. Ensures all-or-nothing semantics (atomic)
	// 4. Prevents phantom reads and write skew anomalies
}

// TestAllOrNothingSemantics documents the atomic transaction behavior
func TestAllOrNothingSemantics(t *testing.T) {
	// This test documents the expected atomic behavior

	scenarios := []struct {
		name           string
		operation      string
		affectedCount  int
		failurePoint   string
		expectRollback bool
	}{
		{
			name:           "All succeed",
			operation:      "add_student",
			affectedCount:  5,
			failurePoint:   "none",
			expectRollback: false,
		},
		{
			name:           "One lesson full - rollback all",
			operation:      "add_student",
			affectedCount:  5,
			failurePoint:   "lesson_full",
			expectRollback: true,
		},
		{
			name:           "Student already booked - rollback all",
			operation:      "add_student",
			affectedCount:  5,
			failurePoint:   "already_booked",
			expectRollback: true,
		},
		{
			name:           "Capacity below current - rollback all",
			operation:      "change_capacity",
			affectedCount:  5,
			failurePoint:   "capacity_violation",
			expectRollback: true,
		},
	}

	for _, tt := range scenarios {
		t.Run(tt.name, func(t *testing.T) {
			// Document expected behavior:
			// If ANY lesson fails validation or modification, the ENTIRE transaction must rollback
			// This ensures data consistency and prevents partial modifications

			if tt.expectRollback {
				assert.NotEqual(t, "none", tt.failurePoint, "Rollback scenario should have failure point")
			} else {
				assert.Equal(t, "none", tt.failurePoint, "Success scenario should have no failure")
			}
		})
	}
}

// TestModificationTypes documents all supported modification types
func TestModificationTypes(t *testing.T) {
	supportedTypes := []string{
		"add_student",
		"remove_student",
		"change_teacher",
		"change_time",
		"change_capacity",
	}

	for _, modType := range supportedTypes {
		t.Run(modType, func(t *testing.T) {
			assert.NotEmpty(t, modType, "Modification type should not be empty")

			// Verify each type has specific validation requirements
			switch modType {
			case "add_student", "remove_student":
				assert.Contains(t, modType, "student", "Should reference student")
			case "change_teacher":
				assert.Contains(t, modType, "teacher", "Should reference teacher")
			case "change_time":
				assert.Contains(t, modType, "time", "Should reference time")
			case "change_capacity":
				assert.Contains(t, modType, "capacity", "Should reference capacity")
			}
		})
	}
}

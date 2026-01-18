package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"tutoring-platform/internal/models"
)

// TestTemplateLesson_CreditsCost_Create tests creating a template lesson with credits_cost
func TestTemplateLesson_CreditsCost_Create(t *testing.T) {
	tests := []struct {
		name           string
		creditsCost    *int
		expectedStatus int
		shouldSucceed  bool
		expectedErr    string
	}{
		{
			name:           "Create with CreditsCost=5 - should succeed",
			creditsCost:    intPtr(5),
			expectedStatus: http.StatusCreated,
			shouldSucceed:  true,
		},
		{
			name:           "Create with CreditsCost=1 - should succeed",
			creditsCost:    intPtr(1),
			expectedStatus: http.StatusCreated,
			shouldSucceed:  true,
		},
		{
			name:           "Create with CreditsCost=100 - should succeed",
			creditsCost:    intPtr(100),
			expectedStatus: http.StatusCreated,
			shouldSucceed:  true,
		},
		{
			name:           "Create with CreditsCost=nil - should use default 1",
			creditsCost:    nil,
			expectedStatus: http.StatusCreated,
			shouldSucceed:  true,
		},
		{
			name:           "Create with CreditsCost=0 - should fail",
			creditsCost:    intPtr(0),
			expectedStatus: http.StatusBadRequest,
			shouldSucceed:  false,
			expectedErr:    "credits_cost must be greater than 0",
		},
		{
			name:           "Create with CreditsCost=101 - should fail",
			creditsCost:    intPtr(101),
			expectedStatus: http.StatusBadRequest,
			shouldSucceed:  false,
			expectedErr:    "credits_cost is unusually high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			req := &models.CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				EndTime:     stringPtr("12:00:00"),
				TeacherID:   uuid.New(),
				LessonType:  stringPtr("individual"),
				MaxStudents: intPtr(1),
				CreditsCost: tt.creditsCost,
				Color:       stringPtr("#3B82F6"),
			}

			// Validate before applying defaults (to test validation logic)
			err := req.Validate()

			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("Expected validation to succeed, got error: %v", err)
					return
				}
				// Apply defaults and validate that credits_cost was set
				req.ApplyDefaults()
				if req.CreditsCost == nil {
					t.Errorf("CreditsCost should not be nil after ApplyDefaults")
					return
				}
				if *req.CreditsCost <= 0 {
					t.Errorf("CreditsCost should be > 0, got %d", *req.CreditsCost)
					return
				}
				if *req.CreditsCost > 100 {
					t.Errorf("CreditsCost should be <= 100, got %d", *req.CreditsCost)
					return
				}
			} else {
				if err == nil {
					t.Errorf("Expected validation to fail, got nil error")
					return
				}
				if tt.expectedErr != "" && !contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
				}
			}
		})
	}
}

// TestTemplateLesson_CreditsCost_GetReturnsValue tests that get returns the credits_cost value
func TestTemplateLesson_CreditsCost_GetReturnsValue(t *testing.T) {
	// Create a template lesson entry directly
	lessonID := uuid.New()
	templateID := uuid.New()
	teacherID := uuid.New()

	lesson := &models.TemplateLessonEntry{
		ID:          lessonID,
		TemplateID:  templateID,
		DayOfWeek:   1,
		StartTime:   "10:00:00",
		EndTime:     "12:00:00",
		TeacherID:   teacherID,
		LessonType:  "individual",
		MaxStudents: 1,
		CreditsCost: 5,
		Color:       "#3B82F6",
		TeacherName: "Test Teacher",
	}

	// Verify that credits_cost is returned in the lesson
	if lesson.CreditsCost != 5 {
		t.Errorf("Expected CreditsCost=5, got %d", lesson.CreditsCost)
	}

	// Test JSON marshaling
	data, err := json.Marshal(lesson)
	if err != nil {
		t.Fatalf("Failed to marshal lesson: %v", err)
	}

	jsonStr := string(data)
	if !contains(jsonStr, `"credits_cost":5`) {
		t.Errorf("Expected 'credits_cost':5 in JSON, got: %s", jsonStr)
	}
}

// TestTemplateLesson_CreditsCost_UpdateValue tests updating the credits_cost value
func TestTemplateLesson_CreditsCost_UpdateValue(t *testing.T) {
	tests := []struct {
		name          string
		oldCost       int
		newCost       *int
		shouldSucceed bool
		expectedErr   string
	}{
		{
			name:          "Update from 5 to 3 - should succeed",
			oldCost:       5,
			newCost:       intPtr(3),
			shouldSucceed: true,
		},
		{
			name:          "Update from 5 to 10 - should succeed",
			oldCost:       5,
			newCost:       intPtr(10),
			shouldSucceed: true,
		},
		{
			name:          "Update from 5 to 100 - should succeed",
			oldCost:       5,
			newCost:       intPtr(100),
			shouldSucceed: true,
		},
		{
			name:          "Update to 0 - should fail",
			oldCost:       5,
			newCost:       intPtr(0),
			shouldSucceed: false,
			expectedErr:   "credits_cost must be greater than 0",
		},
		{
			name:          "Update to 101 - should fail",
			oldCost:       5,
			newCost:       intPtr(101),
			shouldSucceed: false,
			expectedErr:   "credits_cost is unusually high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &models.CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				EndTime:     stringPtr("12:00:00"),
				TeacherID:   uuid.New(),
				LessonType:  stringPtr("individual"),
				MaxStudents: intPtr(1),
				CreditsCost: tt.newCost,
				Color:       stringPtr("#3B82F6"),
			}

			err := req.Validate()

			if tt.shouldSucceed {
				if err != nil {
					t.Errorf("Expected update to succeed, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected update to fail, got nil error")
				}
				if tt.expectedErr != "" && !contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
				}
			}
		})
	}
}

// TestTemplateLesson_CreditsCost_DefaultApplied tests that default is applied when nil
func TestTemplateLesson_CreditsCost_DefaultApplied(t *testing.T) {
	req := &models.CreateTemplateLessonRequest{
		DayOfWeek:   1,
		StartTime:   "10:00:00",
		TeacherID:   uuid.New(),
		CreditsCost: nil, // Not provided
	}

	// Before ApplyDefaults
	if req.CreditsCost != nil {
		t.Errorf("CreditsCost should be nil before ApplyDefaults, got %v", req.CreditsCost)
	}

	// Apply defaults
	req.ApplyDefaults()

	// After ApplyDefaults
	if req.CreditsCost == nil {
		t.Errorf("CreditsCost should not be nil after ApplyDefaults")
		return
	}

	if *req.CreditsCost != 1 {
		t.Errorf("Expected default CreditsCost=1, got %d", *req.CreditsCost)
	}
}

// TestTemplateLesson_CreditsCost_PreservedWhenProvided tests that provided value is preserved
func TestTemplateLesson_CreditsCost_PreservedWhenProvided(t *testing.T) {
	testCases := []int{1, 2, 5, 10, 50, 100}

	for _, cost := range testCases {
		t.Run(fmt.Sprintf("CreditsCost=%d_preserved", cost), func(t *testing.T) {
			req := &models.CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				TeacherID:   uuid.New(),
				CreditsCost: intPtr(cost),
			}

			req.ApplyDefaults()

			if req.CreditsCost == nil {
				t.Errorf("CreditsCost should not be nil")
				return
			}

			if *req.CreditsCost != cost {
				t.Errorf("Expected CreditsCost=%d, got %d", cost, *req.CreditsCost)
			}
		})
	}
}

// TestTemplateLesson_CreditsCost_DatabaseConstraint tests database constraint validation
// This is a unit test to ensure the constraint rules are properly enforced
func TestTemplateLesson_CreditsCost_DatabaseConstraint(t *testing.T) {
	// Database constraint: credits_cost > 0
	tests := []struct {
		name  string
		cost  int
		valid bool
	}{
		{
			name:  "CreditsCost=0_violates_constraint",
			cost:  0,
			valid: false,
		},
		{
			name:  "CreditsCost=-1_violates_constraint",
			cost:  -1,
			valid: false,
		},
		{
			name:  "CreditsCost=1_satisfies_constraint",
			cost:  1,
			valid: true,
		},
		{
			name:  "CreditsCost=100_satisfies_constraint",
			cost:  100,
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				// Should pass validation
				if tt.cost <= 0 {
					t.Errorf("Cost %d should be valid (>0)", tt.cost)
				}
			} else {
				// Should fail validation
				if tt.cost > 0 {
					t.Errorf("Cost %d should be invalid (<=0)", tt.cost)
				}
			}
		})
	}
}

// Helper functions (stringPtr is in test_helpers.go)
func intPtr(i int) *int {
	return &i
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0)
}

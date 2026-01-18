package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateTemplateLessonRequest_ApplyDefaults_MaxStudents(t *testing.T) {
	tests := []struct {
		name                string
		req                 *CreateTemplateLessonRequest
		expectedMaxStudents int
		expectedLessonType  string
	}{
		{
			name: "Default: neither lesson_type nor max_students - should default to individual with 1 student",
			req: &CreateTemplateLessonRequest{
				DayOfWeek: 1,
				StartTime: "10:00:00",
			},
			expectedMaxStudents: 1,
			expectedLessonType:  "individual",
		},
		{
			name: "Explicit individual lesson_type - should set max_students to 1",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:  1,
				StartTime:  "10:00:00",
				LessonType: stringPtr("individual"),
			},
			expectedMaxStudents: 1,
			expectedLessonType:  "individual",
		},
		{
			name: "Explicit group lesson_type - should set max_students to 4",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:  1,
				StartTime:  "10:00:00",
				LessonType: stringPtr("group"),
			},
			expectedMaxStudents: 4,
			expectedLessonType:  "group",
		},
		{
			name: "Explicit max_students=1 - should infer individual lesson_type",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				MaxStudents: intPtr(1),
			},
			expectedMaxStudents: 1,
			expectedLessonType:  "individual",
		},
		{
			name: "Explicit max_students=4 - should infer group lesson_type",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				MaxStudents: intPtr(4),
			},
			expectedMaxStudents: 4,
			expectedLessonType:  "group",
		},
		{
			name: "Explicit max_students=6 - should infer group lesson_type",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				MaxStudents: intPtr(6),
			},
			expectedMaxStudents: 6,
			expectedLessonType:  "group",
		},
		{
			name: "Both lesson_type and max_students explicit - should preserve both",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				LessonType:  stringPtr("group"),
				MaxStudents: intPtr(5),
			},
			expectedMaxStudents: 5,
			expectedLessonType:  "group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.req.ApplyDefaults()

			if tt.req.MaxStudents == nil {
				t.Errorf("MaxStudents should not be nil after ApplyDefaults")
				return
			}

			if *tt.req.MaxStudents != tt.expectedMaxStudents {
				t.Errorf("Expected max_students=%d, got %d", tt.expectedMaxStudents, *tt.req.MaxStudents)
			}

			if tt.req.LessonType == nil {
				t.Errorf("LessonType should not be nil after ApplyDefaults")
				return
			}

			if *tt.req.LessonType != tt.expectedLessonType {
				t.Errorf("Expected lesson_type=%s, got %s", tt.expectedLessonType, *tt.req.LessonType)
			}
		})
	}
}

func TestCreateTemplateLessonRequest_ApplyDefaults_Color(t *testing.T) {
	tests := []struct {
		name          string
		req           *CreateTemplateLessonRequest
		expectedColor string
	}{
		{
			name: "Default color applied when not provided",
			req: &CreateTemplateLessonRequest{
				DayOfWeek: 1,
				StartTime: "10:00:00",
			},
			expectedColor: "#3B82F6",
		},
		{
			name: "Custom color preserved",
			req: &CreateTemplateLessonRequest{
				DayOfWeek: 1,
				StartTime: "10:00:00",
				Color:     stringPtr("#FF5733"),
			},
			expectedColor: "#FF5733",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.req.ApplyDefaults()

			if tt.req.Color == nil {
				t.Errorf("Color should not be nil after ApplyDefaults")
				return
			}

			if *tt.req.Color != tt.expectedColor {
				t.Errorf("Expected color %s, got %s", tt.expectedColor, *tt.req.Color)
			}
		})
	}
}

func TestCreateTemplateLessonRequest_Subject(t *testing.T) {
	tests := []struct {
		name            string
		req             *CreateTemplateLessonRequest
		expectedSubject *string
	}{
		{
			name: "Subject is optional",
			req: &CreateTemplateLessonRequest{
				DayOfWeek: 1,
				StartTime: "10:00:00",
			},
			expectedSubject: nil,
		},
		{
			name: "Subject preserved when provided",
			req: &CreateTemplateLessonRequest{
				DayOfWeek: 1,
				StartTime: "10:00:00",
				Subject:   stringPtr("Math"),
			},
			expectedSubject: stringPtr("Math"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.req.ApplyDefaults()

			if tt.expectedSubject == nil && tt.req.Subject != nil {
				t.Errorf("Expected subject to be nil, got %s", *tt.req.Subject)
			}

			if tt.expectedSubject != nil {
				if tt.req.Subject == nil {
					t.Errorf("Expected subject %s, got nil", *tt.expectedSubject)
					return
				}
				if *tt.req.Subject != *tt.expectedSubject {
					t.Errorf("Expected subject %s, got %s", *tt.expectedSubject, *tt.req.Subject)
				}
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func TestLessonTemplate_MarshalJSON_LessonCount(t *testing.T) {
	template := &LessonTemplate{
		Name:        "Test Template",
		LessonCount: 5,
	}

	data, err := template.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	jsonStr := string(data)

	// Проверяем что lesson_count присутствует в JSON
	if !contains(jsonStr, `"lesson_count":5`) {
		t.Errorf("Expected lesson_count:5 in JSON, got: %s", jsonStr)
	}
}

func TestLessonTemplate_MarshalJSON_ZeroLessonCount(t *testing.T) {
	template := &LessonTemplate{
		Name:        "Empty Template",
		LessonCount: 0,
	}

	data, err := template.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	jsonStr := string(data)

	// Проверяем что lesson_count:0 присутствует в JSON (нулевое значение не должно пропускаться)
	if !contains(jsonStr, `"lesson_count":0`) {
		t.Errorf("Expected lesson_count:0 in JSON, got: %s", jsonStr)
	}
}

func TestCreateTemplateLessonRequest_Validate_CreditsCost(t *testing.T) {
	tests := []struct {
		name           string
		req            *CreateTemplateLessonRequest
		expectedErr    bool
		expectedErrMsg string
	}{
		{
			name: "nil CreditsCost - should pass validation (default 1 will be applied)",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:  1,
				StartTime:  "10:00:00",
				TeacherID:  uuid.New(),
				LessonType: stringPtr("individual"),
			},
			expectedErr: false,
		},
		{
			name: "CreditsCost = 0 - should pass validation (free lessons allowed)",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				TeacherID:   uuid.New(),
				CreditsCost: intPtr(0),
			},
			expectedErr: false,
		},
		{
			name: "CreditsCost = -1 - should fail validation",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				TeacherID:   uuid.New(),
				CreditsCost: intPtr(-1),
			},
			expectedErr:    true,
			expectedErrMsg: "credits_cost must be greater than or equal to 0",
		},
		{
			name: "CreditsCost = 1 - should pass validation",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				TeacherID:   uuid.New(),
				CreditsCost: intPtr(1),
			},
			expectedErr: false,
		},
		{
			name: "CreditsCost = 5 - should pass validation",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				TeacherID:   uuid.New(),
				CreditsCost: intPtr(5),
			},
			expectedErr: false,
		},
		{
			name: "CreditsCost = 50 - should pass validation",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				TeacherID:   uuid.New(),
				CreditsCost: intPtr(50),
			},
			expectedErr: false,
		},
		{
			name: "CreditsCost = 100 - should pass validation (max recommended)",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				TeacherID:   uuid.New(),
				CreditsCost: intPtr(100),
			},
			expectedErr: false,
		},
		{
			name: "CreditsCost = 101 - should fail validation (exceeds recommended max)",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				TeacherID:   uuid.New(),
				CreditsCost: intPtr(101),
			},
			expectedErr:    true,
			expectedErrMsg: "credits_cost is unusually high",
		},
		{
			name: "CreditsCost = 1000 - should fail validation",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				TeacherID:   uuid.New(),
				CreditsCost: intPtr(1000),
			},
			expectedErr:    true,
			expectedErrMsg: "credits_cost is unusually high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.expectedErr {
				if err == nil {
					t.Errorf("Expected error for CreditsCost validation, got nil")
					return
				}
				if !contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Expected error message containing '%s', got '%s'", tt.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for CreditsCost validation, got: %v", err)
				}
			}
		})
	}
}

func TestCreateTemplateLessonRequest_ApplyDefaults_CreditsCost(t *testing.T) {
	tests := []struct {
		name            string
		req             *CreateTemplateLessonRequest
		expectedCost    int
		shouldBeApplied bool
	}{
		{
			name: "nil CreditsCost - should apply default 1",
			req: &CreateTemplateLessonRequest{
				DayOfWeek: 1,
				StartTime: "10:00:00",
			},
			expectedCost:    1,
			shouldBeApplied: true,
		},
		{
			name: "CreditsCost = 5 - should not change",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				CreditsCost: intPtr(5),
			},
			expectedCost:    5,
			shouldBeApplied: false,
		},
		{
			name: "CreditsCost = 100 - should not change",
			req: &CreateTemplateLessonRequest{
				DayOfWeek:   1,
				StartTime:   "10:00:00",
				CreditsCost: intPtr(100),
			},
			expectedCost:    100,
			shouldBeApplied: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.req.ApplyDefaults()

			if tt.req.CreditsCost == nil {
				t.Errorf("CreditsCost should not be nil after ApplyDefaults")
				return
			}

			if *tt.req.CreditsCost != tt.expectedCost {
				t.Errorf("Expected CreditsCost=%d, got %d", tt.expectedCost, *tt.req.CreditsCost)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

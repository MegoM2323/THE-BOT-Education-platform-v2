package models

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestApplyDefaults_EndTime tests that end_time defaults to start_time + 2 hours
func TestApplyDefaults_EndTime(t *testing.T) {
	startTime := time.Now()
	req := &CreateLessonRequest{
		TeacherID: uuid.New(),
		StartTime: startTime,
		// EndTime is nil
		// MaxStudents is nil
	}

	req.ApplyDefaults()

	if req.EndTime == nil {
		t.Fatal("Expected EndTime to be set after ApplyDefaults")
	}

	expectedEndTime := startTime.Add(2 * time.Hour)
	if !req.EndTime.Equal(expectedEndTime) {
		t.Errorf("Expected EndTime=%v, got %v", expectedEndTime, *req.EndTime)
	}
}

// TestApplyDefaults_MaxStudents_Individual tests that individual lessons default to 1
func TestApplyDefaults_MaxStudents_Individual(t *testing.T) {
	req := &CreateLessonRequest{
		TeacherID: uuid.New(),
		StartTime: time.Now(),
	}

	req.ApplyDefaults()

	if req.MaxStudents == nil {
		t.Fatal("Expected MaxStudents to be set after ApplyDefaults")
	}

	if *req.MaxStudents != 1 {
		t.Errorf("Expected MaxStudents=1 for individual lesson, got %d", *req.MaxStudents)
	}
}

// TestApplyDefaults_MaxStudents_Group tests that group lessons default to 4 when lesson_type is explicitly set to Group
func TestApplyDefaults_MaxStudents_Group(t *testing.T) {
	lessonType := LessonTypeGroup
	req := &CreateLessonRequest{
		TeacherID:  uuid.New(),
		StartTime:  time.Now(),
		LessonType: &lessonType,
	}

	req.ApplyDefaults()

	if req.MaxStudents == nil {
		t.Fatal("Expected MaxStudents to be set after ApplyDefaults")
	}

	if *req.MaxStudents != 4 {
		t.Errorf("Expected MaxStudents=4 for group lesson, got %d", *req.MaxStudents)
	}
}

// TestApplyDefaults_MaxStudents_DefaultIndividual tests that default lesson (no type specified) is individual with max_students=1
func TestApplyDefaults_MaxStudents_DefaultIndividual(t *testing.T) {
	req := &CreateLessonRequest{
		TeacherID: uuid.New(),
		StartTime: time.Now(),
		// Neither LessonType nor MaxStudents specified
	}

	req.ApplyDefaults()

	if req.MaxStudents == nil {
		t.Fatal("Expected MaxStudents to be set after ApplyDefaults")
	}

	if *req.MaxStudents != 1 {
		t.Errorf("Expected MaxStudents=1 for default lesson, got %d", *req.MaxStudents)
	}

	if req.LessonType == nil {
		t.Fatal("Expected LessonType to be set after ApplyDefaults")
	}

	if *req.LessonType != LessonTypeIndividual {
		t.Errorf("Expected LessonType=Individual for default lesson, got %s", *req.LessonType)
	}
}

// TestApplyDefaults_DoesNotOverrideProvidedValues tests that defaults don't override user values
func TestApplyDefaults_DoesNotOverrideProvidedValues(t *testing.T) {
	startTime := time.Now()
	customEndTime := startTime.Add(3 * time.Hour)
	customMaxStudents := 6

	req := &CreateLessonRequest{
		TeacherID:   uuid.New(),
		StartTime:   startTime,
		EndTime:     &customEndTime,
		MaxStudents: &customMaxStudents,
	}

	req.ApplyDefaults()

	if !req.EndTime.Equal(customEndTime) {
		t.Errorf("Expected EndTime to remain %v, got %v", customEndTime, *req.EndTime)
	}

	if *req.MaxStudents != customMaxStudents {
		t.Errorf("Expected MaxStudents to remain %d, got %d", customMaxStudents, *req.MaxStudents)
	}
}

// TestValidate_EndTimeOptional tests that validation passes when EndTime is nil
func TestValidate_EndTimeOptional(t *testing.T) {
	maxStudents := 4
	req := &CreateLessonRequest{
		TeacherID:   uuid.New(),
		StartTime:   time.Now(),
		MaxStudents: &maxStudents,
		// EndTime is nil - should be valid
	}

	if err := req.Validate(); err != nil {
		t.Errorf("Expected validation to pass with nil EndTime, got error: %v", err)
	}
}

// TestValidate_MaxStudentsOptional tests that validation passes when MaxStudents is nil
func TestValidate_MaxStudentsOptional(t *testing.T) {
	endTime := time.Now().Add(2 * time.Hour)
	req := &CreateLessonRequest{
		TeacherID: uuid.New(),
		StartTime: time.Now(),
		EndTime:   &endTime,
		// MaxStudents is nil - should be valid
	}

	if err := req.Validate(); err != nil {
		t.Errorf("Expected validation to pass with nil MaxStudents, got error: %v", err)
	}
}

// TestValidate_EndTimeMustBeAfterStartTime tests that end_time > start_time is enforced
func TestValidate_EndTimeMustBeAfterStartTime(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(-1 * time.Hour) // Invalid: before start
	maxStudents := 4

	req := &CreateLessonRequest{
		TeacherID:   uuid.New(),
		StartTime:   startTime,
		EndTime:     &endTime,
		MaxStudents: &maxStudents,
	}

	if err := req.Validate(); err != ErrInvalidLessonTime {
		t.Errorf("Expected ErrInvalidLessonTime, got %v", err)
	}
}

// TestValidate_IndividualLessonMustHaveMaxStudents1 tests that individual lessons must have max_students = 1
func TestValidate_IndividualLessonMustHaveMaxStudents1(t *testing.T) {
	endTime := time.Now().Add(2 * time.Hour)
	maxStudents := 2 // Invalid: 2 is neither 1 (individual) nor >= 4 (group)
	lessonType := LessonTypeIndividual

	req := &CreateLessonRequest{
		TeacherID:   uuid.New(),
		StartTime:   time.Now(),
		EndTime:     &endTime,
		MaxStudents: &maxStudents,
		LessonType:  &lessonType,
	}

	if err := req.Validate(); err != ErrIndividualLessonMaxStudents {
		t.Errorf("Expected ErrIndividualLessonMaxStudents, got %v", err)
	}
}

// TestValidate_GroupLessonMinimum4Students tests that group lessons must have at least 4 students on creation
func TestValidate_GroupLessonMinimum4Students(t *testing.T) {
	endTime := time.Now().Add(2 * time.Hour)

	tests := []struct {
		name        string
		maxStudents int
		wantErr     error
	}{
		{"1 student", 1, nil},                        // Valid (individual lesson)
		{"2 students", 2, ErrGroupLessonMinStudents}, // Invalid: 2 is neither 1 nor >= 4
		{"3 students", 3, ErrGroupLessonMinStudents}, // Invalid: 3 is neither 1 nor >= 4
		{"4 students", 4, nil},                       // Valid (group lesson)
		{"6 students", 6, nil},                       // Valid (can create with more)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreateLessonRequest{
				TeacherID:   uuid.New(),
				StartTime:   time.Now(),
				EndTime:     &endTime,
				MaxStudents: &tt.maxStudents,
			}

			err := req.Validate()
			if err != tt.wantErr {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

// TestValidate_ValidIndividualLesson tests that a properly constructed individual lesson passes validation
func TestValidate_ValidIndividualLesson(t *testing.T) {
	endTime := time.Now().Add(2 * time.Hour)
	maxStudents := 1

	req := &CreateLessonRequest{
		TeacherID:   uuid.New(),
		StartTime:   time.Now(),
		EndTime:     &endTime,
		MaxStudents: &maxStudents,
	}

	if err := req.Validate(); err != nil {
		t.Errorf("Expected validation to pass for valid individual lesson, got error: %v", err)
	}
}

// TestValidate_ValidGroupLesson tests that a properly constructed group lesson passes validation
func TestValidate_ValidGroupLesson(t *testing.T) {
	endTime := time.Now().Add(2 * time.Hour)
	maxStudents := 4

	req := &CreateLessonRequest{
		TeacherID:   uuid.New(),
		StartTime:   time.Now(),
		EndTime:     &endTime,
		MaxStudents: &maxStudents,
	}

	if err := req.Validate(); err != nil {
		t.Errorf("Expected validation to pass for valid group lesson, got error: %v", err)
	}
}

// TestFullFlow_ApplyDefaultsThenValidate tests the complete flow: ApplyDefaults -> Validate
func TestFullFlow_ApplyDefaultsThenValidate(t *testing.T) {
	tests := []struct {
		name       string
		lessonType LessonType
		wantValid  bool
	}{
		{
			name:       "Individual lesson with defaults",
			lessonType: LessonTypeIndividual,
			wantValid:  true,
		},
		{
			name:       "Group lesson with defaults",
			lessonType: LessonTypeGroup,
			wantValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lessonType := tt.lessonType
			req := &CreateLessonRequest{
				TeacherID:  uuid.New(),
				LessonType: &lessonType,
				StartTime:  time.Now(),
				// EndTime and MaxStudents are nil
			}

			req.ApplyDefaults()

			err := req.Validate()
			if tt.wantValid && err != nil {
				t.Errorf("Expected validation to pass, got error: %v", err)
			}
			if !tt.wantValid && err == nil {
				t.Errorf("Expected validation to fail, but it passed")
			}
		})
	}
}

// TestLesson_MarshalJSON_ConvertsSqlNullString tests Lesson JSON serialization with sql.NullString fields
func TestLesson_MarshalJSON_ConvertsSqlNullString(t *testing.T) {
	tests := []struct {
		name          string
		lesson        *Lesson
		checkFields   map[string]interface{} // Fields to verify in JSON
		shouldExist   []string               // Fields that should exist in JSON
		shouldNotHave []string               // Fields that should NOT exist in JSON (omitempty)
	}{
		{
			name: "Valid subject and homework",
			lesson: &Lesson{
				ID:              uuid.New(),
				TeacherID:       uuid.New(),
				StartTime:       time.Now(),
				EndTime:         time.Now().Add(2 * time.Hour),
				MaxStudents:     4,
				CurrentStudents: 0,
				Color:           "#3B82F6",
				Subject:         sql.NullString{String: "Mathematics", Valid: true},
				HomeworkText:    sql.NullString{String: "Chapter 5", Valid: true},
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				DeletedAt:       sql.NullTime{Valid: false},
			},
			shouldExist: []string{"id", "teacher_id", "subject", "homework_text"},
		},
		{
			name: "Null subject and homework",
			lesson: &Lesson{
				ID:              uuid.New(),
				TeacherID:       uuid.New(),
				StartTime:       time.Now(),
				EndTime:         time.Now().Add(2 * time.Hour),
				MaxStudents:     4,
				CurrentStudents: 0,
				Color:           "#3B82F6",
				Subject:         sql.NullString{Valid: false},
				HomeworkText:    sql.NullString{Valid: false},
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				DeletedAt:       sql.NullTime{Valid: false},
			},
			shouldExist: []string{"id", "teacher_id"},
		},
		{
			name: "With deleted_at timestamp",
			lesson: &Lesson{
				ID:              uuid.New(),
				TeacherID:       uuid.New(),
				StartTime:       time.Now(),
				EndTime:         time.Now().Add(2 * time.Hour),
				MaxStudents:     4,
				CurrentStudents: 0,
				Color:           "#3B82F6",
				Subject:         sql.NullString{Valid: false},
				HomeworkText:    sql.NullString{Valid: false},
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				DeletedAt:       sql.NullTime{Time: time.Now(), Valid: true},
			},
			shouldExist: []string{"id", "deleted_at"},
		},
		{
			name: "With recurring_group_id",
			lesson: &Lesson{
				ID:               uuid.New(),
				TeacherID:        uuid.New(),
				StartTime:        time.Now(),
				EndTime:          time.Now().Add(2 * time.Hour),
				MaxStudents:      4,
				CurrentStudents:  0,
				Color:            "#3B82F6",
				Subject:          sql.NullString{Valid: false},
				HomeworkText:     sql.NullString{Valid: false},
				IsRecurring:      true,
				RecurringGroupID: func() *uuid.UUID { u := uuid.New(); return &u }(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				DeletedAt:        sql.NullTime{Valid: false},
			},
			shouldExist: []string{"recurring_group_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.lesson)
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// Check that required fields exist
			for _, field := range tt.shouldExist {
				if _, exists := result[field]; !exists {
					t.Errorf("Expected field '%s' to exist in JSON output", field)
				}
			}

			// Check that sql.NullString was properly converted
			if tt.lesson.Subject.Valid {
				subject, ok := result["subject"]
				if !ok {
					t.Error("Expected 'subject' field to exist")
				}
				if subjectStr, ok := subject.(string); ok {
					if subjectStr != tt.lesson.Subject.String {
						t.Errorf("Expected subject=%s, got %s", tt.lesson.Subject.String, subjectStr)
					}
				} else {
					t.Error("Expected 'subject' to be string, got other type")
				}
			}

			// Check that HomeworkText was properly converted
			if tt.lesson.HomeworkText.Valid {
				homework, ok := result["homework_text"]
				if !ok {
					t.Error("Expected 'homework_text' field to exist")
				}
				if hwStr, ok := homework.(string); ok {
					if hwStr != tt.lesson.HomeworkText.String {
						t.Errorf("Expected homework_text=%s, got %s", tt.lesson.HomeworkText.String, hwStr)
					}
				} else {
					t.Error("Expected 'homework_text' to be string, got other type")
				}
			}

			// Check that DeletedAt was properly converted
			if tt.lesson.DeletedAt.Valid {
				deletedAt, ok := result["deleted_at"]
				if !ok {
					t.Error("Expected 'deleted_at' field to exist")
				}
				if deletedAt == nil {
					t.Error("Expected 'deleted_at' to not be nil")
				}
			}
		})
	}
}

// TestLessonWithTeacher_MarshalJSON_InheritsBase tests that LessonWithTeacher properly inherits MarshalJSON behavior
func TestLessonWithTeacher_MarshalJSON_InheritsBase(t *testing.T) {
	tests := []struct {
		name           string
		lessonWTeacher *LessonWithTeacher
		shouldExist    []string
	}{
		{
			name: "LessonWithTeacher serializes with teacher_name",
			lessonWTeacher: &LessonWithTeacher{
				Lesson: Lesson{
					ID:              uuid.New(),
					TeacherID:       uuid.New(),
					StartTime:       time.Now(),
					EndTime:         time.Now().Add(2 * time.Hour),
					MaxStudents:     4,
					CurrentStudents: 0,
					Color:           "#3B82F6",
					Subject:         sql.NullString{String: "Physics", Valid: true},
					HomeworkText:    sql.NullString{String: "Exercise 10", Valid: true},
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
					DeletedAt:       sql.NullTime{Valid: false},
				},
				TeacherName: "John Doe",
			},
			shouldExist: []string{"id", "teacher_name", "subject", "homework_text"},
		},
		{
			name: "LessonWithTeacher without optional fields",
			lessonWTeacher: &LessonWithTeacher{
				Lesson: Lesson{
					ID:              uuid.New(),
					TeacherID:       uuid.New(),
					StartTime:       time.Now(),
					EndTime:         time.Now().Add(2 * time.Hour),
					MaxStudents:     1,
					CurrentStudents: 1,
					Color:           "#FF5733",
					Subject:         sql.NullString{Valid: false},
					HomeworkText:    sql.NullString{Valid: false},
					CreatedAt:       time.Now(),
					UpdatedAt:       time.Now(),
					DeletedAt:       sql.NullTime{Valid: false},
				},
				TeacherName: "Jane Smith",
			},
			shouldExist: []string{"id", "teacher_name", "teacher_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.lessonWTeacher)
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// Check that required fields exist
			for _, field := range tt.shouldExist {
				if _, exists := result[field]; !exists {
					t.Errorf("Expected field '%s' to exist in JSON output", field)
				}
			}

			// Verify teacher_name is properly included
			if teacherName, ok := result["teacher_name"].(string); ok {
				if teacherName != tt.lessonWTeacher.TeacherName {
					t.Errorf("Expected teacher_name=%s, got %s", tt.lessonWTeacher.TeacherName, teacherName)
				}
			} else {
				t.Error("Expected 'teacher_name' to be string")
			}

			// Verify that sql.NullString fields are converted
			if tt.lessonWTeacher.Subject.Valid {
				if subject, ok := result["subject"].(string); ok {
					if subject != tt.lessonWTeacher.Subject.String {
						t.Errorf("Expected subject=%s, got %s", tt.lessonWTeacher.Subject.String, subject)
					}
				}
			}
		})
	}
}

// TestLesson_MarshalJSON_Consistency tests that Lesson serialization is consistent
func TestLesson_MarshalJSON_Consistency(t *testing.T) {
	lesson := &Lesson{
		ID:              uuid.New(),
		TeacherID:       uuid.New(),
		StartTime:       time.Now().Round(time.Second),
		EndTime:         time.Now().Add(2 * time.Hour).Round(time.Second),
		MaxStudents:     4,
		CurrentStudents: 2,
		Color:           "#3B82F6",
		Subject:         sql.NullString{String: "Biology", Valid: true},
		HomeworkText:    sql.NullString{String: "Read chapter 3", Valid: true},
		CreatedAt:       time.Now().Round(time.Second),
		UpdatedAt:       time.Now().Round(time.Second),
		DeletedAt:       sql.NullTime{Valid: false},
	}

	// Marshal twice
	data1, err := json.Marshal(lesson)
	if err != nil {
		t.Fatalf("First MarshalJSON failed: %v", err)
	}

	data2, err := json.Marshal(lesson)
	if err != nil {
		t.Fatalf("Second MarshalJSON failed: %v", err)
	}

	// Compare JSON outputs
	if string(data1) != string(data2) {
		t.Error("MarshalJSON produced different outputs for the same input")
		t.Logf("First:  %s", data1)
		t.Logf("Second: %s", data2)
	}
}

// TestLessonMarshalJSON_TypeCorrectness verifies all fields have correct JSON types
func TestLessonMarshalJSON_TypeCorrectness(t *testing.T) {
	recurringID := uuid.New()
	lesson := &Lesson{
		ID:               uuid.New(),
		TeacherID:        uuid.New(),
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(2 * time.Hour),
		MaxStudents:      4,
		CurrentStudents:  2,
		Color:            "#3B82F6",
		Subject:          sql.NullString{String: "Art", Valid: true},
		HomeworkText:     sql.NullString{String: "Sketch", Valid: true},
		IsRecurring:      true,
		RecurringGroupID: &recurringID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DeletedAt:        sql.NullTime{Time: time.Now(), Valid: true},
	}

	data, err := json.Marshal(lesson)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Type checks for all fields
	typeChecks := []struct {
		field        string
		expectedType string
	}{
		{"id", "string"}, // UUID marshals to string
		{"teacher_id", "string"},
		{"max_students", "float64"}, // Numbers marshal to float64 in interface{}
		{"current_students", "float64"},
		{"color", "string"},
		{"subject", "string"},
		{"homework_text", "string"},
		{"is_recurring", "bool"},
		{"recurring_group_id", "string"},
		{"created_at", "string"}, // Time marshals to string
		{"updated_at", "string"},
		{"deleted_at", "string"},
	}

	for _, tc := range typeChecks {
		val, exists := result[tc.field]
		if !exists {
			// Some fields might be omitted if empty, that's ok
			continue
		}
		if val == nil {
			continue // Null values are acceptable
		}

		var actualType string
		switch val.(type) {
		case string:
			actualType = "string"
		case float64:
			actualType = "float64"
		case bool:
			actualType = "bool"
		default:
			actualType = "unknown"
		}

		if actualType != tc.expectedType {
			t.Errorf("Field '%s': expected type %s, got %s", tc.field, tc.expectedType, actualType)
		}
	}
}

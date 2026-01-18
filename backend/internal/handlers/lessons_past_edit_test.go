package handlers

import (
	"testing"
	"time"

	"tutoring-platform/internal/models"
)

// Test: Lesson.IsInPast() correctly identifies past lessons
func TestLesson_IsInPast(t *testing.T) {
	tests := []struct {
		name      string
		startTime time.Time
		want      bool
	}{
		{
			name:      "Past lesson (24 hours ago)",
			startTime: time.Now().Add(-24 * time.Hour),
			want:      true,
		},
		{
			name:      "Past lesson (1 hour ago)",
			startTime: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "Future lesson (24 hours ahead)",
			startTime: time.Now().Add(24 * time.Hour),
			want:      false,
		},
		{
			name:      "Future lesson (1 minute ahead)",
			startTime: time.Now().Add(1 * time.Minute),
			want:      false,
		},
		{
			name:      "Lesson starting now (edge case)",
			startTime: time.Now().Add(1 * time.Millisecond), // Чуть в будущем
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lesson := &models.Lesson{
				StartTime: tt.startTime,
			}

			got := lesson.IsInPast()
			if got != tt.want {
				t.Errorf("IsInPast() = %v, want %v (startTime=%v, now=%v)",
					got, tt.want, tt.startTime, time.Now())
			}
		})
	}
}

// Test: LessonResponse.ToResponse includes is_past field
func TestLessonWithTeacher_ToResponse_IncludesIsPast(t *testing.T) {
	tests := []struct {
		name      string
		startTime time.Time
		wantPast  bool
	}{
		{
			name:      "Past lesson returns is_past=true",
			startTime: time.Now().Add(-2 * time.Hour),
			wantPast:  true,
		},
		{
			name:      "Future lesson returns is_past=false",
			startTime: time.Now().Add(2 * time.Hour),
			wantPast:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lesson := &models.LessonWithTeacher{
				Lesson: models.Lesson{
					StartTime: tt.startTime,
					EndTime:   tt.startTime.Add(2 * time.Hour),
				},
				TeacherName: "Test Teacher",
			}

			response := lesson.ToResponse()

			if response.IsPast != tt.wantPast {
				t.Errorf("ToResponse().IsPast = %v, want %v", response.IsPast, tt.wantPast)
			}
		})
	}
}

// Test: LessonResponse always has is_past field (not omitempty)
func TestLessonResponse_IsPastFieldPresent(t *testing.T) {
	lesson := &models.LessonWithTeacher{
		Lesson: models.Lesson{
			StartTime: time.Now().Add(24 * time.Hour),
			EndTime:   time.Now().Add(26 * time.Hour),
		},
		TeacherName: "Test",
	}

	response := lesson.ToResponse()

	// Проверяем что поле существует и false
	if response.IsPast != false {
		t.Errorf("Future lesson should have IsPast=false, got %v", response.IsPast)
	}
}

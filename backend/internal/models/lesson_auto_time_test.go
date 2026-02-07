package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestCreateLessonRequest_ApplyDefaults_AutoEndTime tests automatic end_time calculation
func TestCreateLessonRequest_ApplyDefaults_AutoEndTime(t *testing.T) {
	tests := []struct {
		name            string
		startTime       time.Time
		endTime         *time.Time
		expectedEndDiff time.Duration
	}{
		{
			name:            "Auto-calculate end_time when not provided",
			startTime:       time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
			endTime:         nil,
			expectedEndDiff: 2 * time.Hour,
		},
		{
			name:            "Keep provided end_time when specified",
			startTime:       time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
			endTime:         ptrTime(time.Date(2025, 12, 1, 13, 0, 0, 0, time.UTC)),
			expectedEndDiff: 3 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreateLessonRequest{
				TeacherID: uuid.New(),
				StartTime: tt.startTime,
			}
			if tt.endTime != nil {
				req.EndTime = *tt.endTime
			}

			// Apply defaults (this is what the service does before validation)
			req.ApplyDefaults()

			// Verify end_time is set correctly
			actualDiff := req.EndTime.Sub(req.StartTime)
			assert.Equal(t, tt.expectedEndDiff, actualDiff, "End time should be %v after start time", tt.expectedEndDiff)
		})
	}
}

// TestCreateLessonRequest_Validation_EndTimeGreaterThanStartTime tests validation
func TestCreateLessonRequest_Validation_EndTimeGreaterThanStartTime(t *testing.T) {
	tests := []struct {
		name        string
		startTime   time.Time
		endTime     *time.Time
		shouldError bool
		expectedErr error
	}{
		{
			name:        "Valid: end_time after start_time",
			startTime:   time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
			endTime:     ptrTime(time.Date(2025, 12, 1, 12, 0, 0, 0, time.UTC)),
			shouldError: false,
		},
		{
			name:        "Valid: end_time not provided (will be auto-calculated)",
			startTime:   time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
			endTime:     nil,
			shouldError: false,
		},
		{
			name:        "Invalid: end_time before start_time",
			startTime:   time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
			endTime:     ptrTime(time.Date(2025, 12, 1, 9, 0, 0, 0, time.UTC)),
			shouldError: true,
			expectedErr: ErrInvalidLessonTime,
		},
		{
			name:        "Invalid: end_time equal to start_time",
			startTime:   time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC),
			endTime:     ptrTime(time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)),
			shouldError: true,
			expectedErr: ErrInvalidLessonTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreateLessonRequest{
				TeacherID: uuid.New(),
				StartTime: tt.startTime,
			}
			if tt.endTime != nil {
				req.EndTime = *tt.endTime
			}

			// Apply defaults first (like the service does)
			req.ApplyDefaults()

			// Then validate
			err := req.Validate()

			if tt.shouldError {
				assert.Error(t, err, "Expected validation error")
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr, "Expected specific error type")
				}
			} else {
				assert.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

// TestUpdateLessonRequest_Validation_EndTimeGreaterThanStartTime tests update validation
func TestUpdateLessonRequest_Validation_EndTimeGreaterThanStartTime(t *testing.T) {
	tests := []struct {
		name        string
		startTime   *time.Time
		endTime     *time.Time
		shouldError bool
		expectedErr error
	}{
		{
			name:        "Valid: both times provided, end > start",
			startTime:   ptrTime(time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)),
			endTime:     ptrTime(time.Date(2025, 12, 1, 12, 0, 0, 0, time.UTC)),
			shouldError: false,
		},
		{
			name:        "Valid: only start_time provided (end_time will auto-calculate)",
			startTime:   ptrTime(time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)),
			endTime:     nil,
			shouldError: false,
		},
		{
			name:        "Valid: only end_time provided",
			startTime:   nil,
			endTime:     ptrTime(time.Date(2025, 12, 1, 12, 0, 0, 0, time.UTC)),
			shouldError: false,
		},
		{
			name:        "Valid: neither time provided",
			startTime:   nil,
			endTime:     nil,
			shouldError: false,
		},
		{
			name:        "Invalid: end_time before start_time",
			startTime:   ptrTime(time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)),
			endTime:     ptrTime(time.Date(2025, 12, 1, 9, 0, 0, 0, time.UTC)),
			shouldError: true,
			expectedErr: ErrInvalidLessonTime,
		},
		{
			name:        "Invalid: end_time equal to start_time",
			startTime:   ptrTime(time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)),
			endTime:     ptrTime(time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)),
			shouldError: true,
			expectedErr: ErrInvalidLessonTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &UpdateLessonRequest{
				StartTime: tt.startTime,
				EndTime:   tt.endTime,
			}

			err := req.Validate()

			if tt.shouldError {
				assert.Error(t, err, "Expected validation error")
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr, "Expected specific error type")
				}
			} else {
				assert.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

// TestApplyDefaults_Integration tests that defaults work correctly with all fields
func TestApplyDefaults_Integration(t *testing.T) {
	startTime := time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name                string
		endTime             *time.Time
		maxStudents         int
		expectedEndDiff     time.Duration
		expectedMaxStudents int
	}{
		{
			name:                "Auto end_time with max_students=1",
			endTime:             nil,
			maxStudents:         1,
			expectedEndDiff:     2 * time.Hour,
			expectedMaxStudents: 1,
		},
		{
			name:                "Custom end_time and max_students",
			endTime:             ptrTime(startTime.Add(3 * time.Hour)),
			maxStudents:         6,
			expectedEndDiff:     3 * time.Hour,
			expectedMaxStudents: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreateLessonRequest{
				TeacherID:   uuid.New(),
				StartTime:   startTime,
				MaxStudents: tt.maxStudents,
			}
			if tt.endTime != nil {
				req.EndTime = *tt.endTime
			}

			req.ApplyDefaults()

			// Verify end_time
			actualEndDiff := req.EndTime.Sub(req.StartTime)
			assert.Equal(t, tt.expectedEndDiff, actualEndDiff, "End time should be correct")

			// Verify max_students
			assert.Equal(t, tt.expectedMaxStudents, req.MaxStudents, "MaxStudents should be correct")

			// Verify request is valid
			err := req.Validate()
			assert.NoError(t, err, "Request should be valid after defaults")
		})
	}
}

// Helper functions
func ptrTime(t time.Time) *time.Time {
	return &t
}

func ptrInt(i int) *int {
	return &i
}

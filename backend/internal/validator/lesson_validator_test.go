package validator

import (
	"testing"

	"tutoring-platform/internal/models"
)

func TestLessonValidator_ValidateColor(t *testing.T) {
	validator := NewLessonValidator()

	tests := []struct {
		name    string
		color   string
		wantErr bool
	}{
		{
			name:    "Valid hex color uppercase",
			color:   "#FF5733",
			wantErr: false,
		},
		{
			name:    "Valid hex color lowercase",
			color:   "#3b82f6",
			wantErr: false,
		},
		{
			name:    "Valid hex color mixed case",
			color:   "#3B82F6",
			wantErr: false,
		},
		{
			name:    "Invalid - no hash",
			color:   "FF5733",
			wantErr: true,
		},
		{
			name:    "Invalid - too short",
			color:   "#FFF",
			wantErr: true,
		},
		{
			name:    "Invalid - too long",
			color:   "#FF57331",
			wantErr: true,
		},
		{
			name:    "Invalid - non-hex characters",
			color:   "#GG5733",
			wantErr: true,
		},
		{
			name:    "Invalid - empty string",
			color:   "",
			wantErr: true,
		},
		{
			name:    "Invalid - only hash",
			color:   "#",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateColor(tt.color)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateColor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err != models.ErrInvalidColor {
				t.Errorf("ValidateColor() expected ErrInvalidColor, got %v", err)
			}
		})
	}
}

func TestLessonValidator_ValidateSubject(t *testing.T) {
	validator := NewLessonValidator()

	tests := []struct {
		name    string
		subject string
		wantErr bool
	}{
		{
			name:    "Valid subject",
			subject: "Mathematics",
			wantErr: false,
		},
		{
			name:    "Valid empty subject",
			subject: "",
			wantErr: false,
		},
		{
			name:    "Valid long subject",
			subject: "Advanced Mathematics for High School Students - Calculus and Trigonometry",
			wantErr: false,
		},
		{
			name:    "Invalid - too long (201 characters)",
			subject: "This is a very long subject name that exceeds the maximum allowed length of 200 characters and should fail validation because it is way too long to be a reasonable subject name for a lesson in any context whatsoever really.",
			wantErr: true,
		},
		{
			name:    "Valid - exactly 200 characters",
			subject: "This is exactly 200 characters long subject name that should pass validation because it is exactly at the limit of what is allowed for a subject name in our system which is quite generous actually...",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSubject(tt.subject)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSubject() error = %v, wantErr %v (subject length: %d)", err, tt.wantErr, len(tt.subject))
			}
			if err != nil && err != models.ErrSubjectTooLong {
				t.Errorf("ValidateSubject() expected ErrSubjectTooLong, got %v", err)
			}
		})
	}
}

func TestLessonValidator_ValidateCreateLessonRequest(t *testing.T) {
	validator := NewLessonValidator()

	validColor := "#3B82F6"
	invalidColor := "invalid"
	validSubject := "Math"
	tooLongSubject := string(make([]byte, 201))

	tests := []struct {
		name    string
		req     *models.CreateLessonRequest
		wantErr bool
	}{
		{
			name: "Valid request with color and subject",
			req: &models.CreateLessonRequest{
				Color:   validColor,
				Subject: &validSubject,
			},
			wantErr: false,
		},
		{
			name: "Valid request with color only",
			req: &models.CreateLessonRequest{
				Color: validColor,
			},
			wantErr: false,
		},
		{
			name: "Valid request with empty color and subject",
			req: &models.CreateLessonRequest{
				Color:   "",
				Subject: nil,
			},
			wantErr: false,
		},
		{
			name: "Invalid color",
			req: &models.CreateLessonRequest{
				Color:   invalidColor,
				Subject: &validSubject,
			},
			wantErr: true,
		},
		{
			name: "Invalid subject - too long",
			req: &models.CreateLessonRequest{
				Color:   validColor,
				Subject: &tooLongSubject,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCreateLessonRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateLessonRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLessonValidator_ValidateUpdateLessonRequest(t *testing.T) {
	validator := NewLessonValidator()

	validColor := "#FF5733"
	invalidColor := "#ZZZ"
	validSubject := "English"
	tooLongSubject := string(make([]byte, 201))

	tests := []struct {
		name    string
		req     *models.UpdateLessonRequest
		wantErr bool
	}{
		{
			name: "Valid update with color and subject",
			req: &models.UpdateLessonRequest{
				Color:   &validColor,
				Subject: &validSubject,
			},
			wantErr: false,
		},
		{
			name: "Valid update with color only",
			req: &models.UpdateLessonRequest{
				Color: &validColor,
			},
			wantErr: false,
		},
		{
			name: "Valid update with nil fields",
			req: &models.UpdateLessonRequest{
				Color:   nil,
				Subject: nil,
			},
			wantErr: false,
		},
		{
			name: "Invalid color",
			req: &models.UpdateLessonRequest{
				Color:   &invalidColor,
				Subject: &validSubject,
			},
			wantErr: true,
		},
		{
			name: "Invalid subject - too long",
			req: &models.UpdateLessonRequest{
				Color:   &validColor,
				Subject: &tooLongSubject,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUpdateLessonRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdateLessonRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

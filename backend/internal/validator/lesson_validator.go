package validator

import (
	"regexp"
	"strings"

	"tutoring-platform/internal/models"
)

// LessonValidator handles lesson-specific validation
type LessonValidator struct{}

// NewLessonValidator creates a new LessonValidator
func NewLessonValidator() *LessonValidator {
	return &LessonValidator{}
}

// hexColorRegex matches #RRGGBB hex color format
var hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// ValidateColor validates that the color is in valid hex format (#RRGGBB)
func (v *LessonValidator) ValidateColor(color string) error {
	if color == "" {
		return models.ErrInvalidColor
	}

	// Trim whitespace
	color = strings.TrimSpace(color)

	// Check hex format: #RRGGBB
	if !hexColorRegex.MatchString(color) {
		return models.ErrInvalidColor
	}

	return nil
}

// ValidateSubject validates the lesson subject/topic
func (v *LessonValidator) ValidateSubject(subject string) error {
	if subject == "" {
		// Empty subject is allowed
		return nil
	}

	// Check length (max 200 characters)
	if len(subject) > 200 {
		return models.ErrSubjectTooLong
	}

	return nil
}

// ValidateCreateLessonRequest validates color and subject in CreateLessonRequest
func (v *LessonValidator) ValidateCreateLessonRequest(req *models.CreateLessonRequest) error {
	// Validate color (required)
	if err := v.ValidateColor(req.Color); err != nil {
		return err
	}

	// Validate subject if provided (optional)
	if req.Subject != nil {
		if err := v.ValidateSubject(*req.Subject); err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdateLessonRequest validates color and subject in UpdateLessonRequest
func (v *LessonValidator) ValidateUpdateLessonRequest(req *models.UpdateLessonRequest) error {
	// Validate color if provided
	if req.Color != nil {
		if err := v.ValidateColor(*req.Color); err != nil {
			return err
		}
	}

	// Validate subject if provided
	if req.Subject != nil {
		if err := v.ValidateSubject(*req.Subject); err != nil {
			return err
		}
	}

	return nil
}

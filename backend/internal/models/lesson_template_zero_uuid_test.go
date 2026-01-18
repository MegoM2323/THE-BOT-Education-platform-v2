package models

import (
	"testing"

	"github.com/google/uuid"
)

// TestApplyTemplateRequest_ZeroUUID verifies that the zero UUID is accepted for default template
func TestApplyTemplateRequest_ZeroUUID(t *testing.T) {
	// Zero UUID is the special ID for singleton default template
	zeroUUID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	req := &ApplyTemplateRequest{
		TemplateID:    zeroUUID,
		WeekStartDate: "2025-12-08", // Monday
	}

	err := req.Validate()
	if err != nil {
		t.Errorf("Expected zero UUID to be valid for default template, got error: %v", err)
	}
}

// TestApplyTemplateRequest_NonZeroUUID verifies that non-zero UUIDs also work
func TestApplyTemplateRequest_NonZeroUUID(t *testing.T) {
	// Regular UUID should also work
	regularUUID := uuid.New()

	req := &ApplyTemplateRequest{
		TemplateID:    regularUUID,
		WeekStartDate: "2025-12-08", // Monday
	}

	err := req.Validate()
	if err != nil {
		t.Errorf("Expected regular UUID to be valid, got error: %v", err)
	}
}

package models

import (
	"testing"

	"github.com/google/uuid"
)

// TestTemplateApplicationStatusIsValid tests the IsValid() method
func TestTemplateApplicationStatusIsValid(t *testing.T) {
	tests := []struct {
		name   string
		status TemplateApplicationStatus
		want   bool
	}{
		{
			name:   "valid: applied",
			status: TemplateApplicationStatusApplied,
			want:   true,
		},
		{
			name:   "valid: replaced",
			status: TemplateApplicationStatusReplaced,
			want:   true,
		},
		{
			name:   "valid: rolled_back",
			status: TemplateApplicationStatusRolledBack,
			want:   true,
		},
		{
			name:   "valid: preview (for API responses)",
			status: TemplateApplicationStatusPreview,
			want:   true,
		},
		{
			name:   "invalid: empty string",
			status: "",
			want:   false,
		},
		{
			name:   "invalid: pending",
			status: "pending",
			want:   false,
		},
		{
			name:   "invalid: approved",
			status: "approved",
			want:   false,
		},
		{
			name:   "invalid: rejected",
			status: "rejected",
			want:   false,
		},
		{
			name:   "invalid: random string",
			status: "invalid_status",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTemplateApplicationStatusString tests the String() method
func TestTemplateApplicationStatusString(t *testing.T) {
	tests := []struct {
		name   string
		status TemplateApplicationStatus
		want   string
	}{
		{
			name:   "applied",
			status: TemplateApplicationStatusApplied,
			want:   "applied",
		},
		{
			name:   "replaced",
			status: TemplateApplicationStatusReplaced,
			want:   "replaced",
		},
		{
			name:   "rolled_back",
			status: TemplateApplicationStatusRolledBack,
			want:   "rolled_back",
		},
		{
			name:   "preview",
			status: TemplateApplicationStatusPreview,
			want:   "preview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestTemplateApplicationValidate tests the Validate() method
func TestTemplateApplicationValidate(t *testing.T) {
	templateID := uuid.New()
	appliedByID := uuid.New()

	tests := []struct {
		name    string
		ta      *TemplateApplication
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid: applied status",
			ta: &TemplateApplication{
				ID:          uuid.New(),
				TemplateID:  templateID,
				AppliedByID: appliedByID,
				Status:      "applied",
			},
			wantErr: false,
		},
		{
			name: "valid: replaced status",
			ta: &TemplateApplication{
				ID:          uuid.New(),
				TemplateID:  templateID,
				AppliedByID: appliedByID,
				Status:      "replaced",
			},
			wantErr: false,
		},
		{
			name: "valid: rolled_back status",
			ta: &TemplateApplication{
				ID:          uuid.New(),
				TemplateID:  templateID,
				AppliedByID: appliedByID,
				Status:      "rolled_back",
			},
			wantErr: false,
		},
		{
			name: "invalid: nil id",
			ta: &TemplateApplication{
				ID:          uuid.Nil,
				TemplateID:  templateID,
				AppliedByID: appliedByID,
				Status:      "applied",
			},
			wantErr: true,
			errMsg:  "id is required",
		},
		{
			name: "invalid: nil template_id",
			ta: &TemplateApplication{
				ID:          uuid.New(),
				TemplateID:  uuid.Nil,
				AppliedByID: appliedByID,
				Status:      "applied",
			},
			wantErr: true,
			errMsg:  "template_id is required",
		},
		{
			name: "invalid: nil applied_by_id",
			ta: &TemplateApplication{
				ID:          uuid.New(),
				TemplateID:  templateID,
				AppliedByID: uuid.Nil,
				Status:      "applied",
			},
			wantErr: true,
			errMsg:  "applied_by_id is required",
		},
		{
			name: "invalid: pending status",
			ta: &TemplateApplication{
				ID:          uuid.New(),
				TemplateID:  templateID,
				AppliedByID: appliedByID,
				Status:      "pending",
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name: "invalid: approved status",
			ta: &TemplateApplication{
				ID:          uuid.New(),
				TemplateID:  templateID,
				AppliedByID: appliedByID,
				Status:      "approved",
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name: "invalid: rejected status",
			ta: &TemplateApplication{
				ID:          uuid.New(),
				TemplateID:  templateID,
				AppliedByID: appliedByID,
				Status:      "rejected",
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name: "invalid: unknown status",
			ta: &TemplateApplication{
				ID:          uuid.New(),
				TemplateID:  templateID,
				AppliedByID: appliedByID,
				Status:      "unknown_status",
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name: "note: preview status is rejected by Validate (only for API responses)",
			ta: &TemplateApplication{
				ID:          uuid.New(),
				TemplateID:  templateID,
				AppliedByID: appliedByID,
				Status:      "preview",
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ta.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

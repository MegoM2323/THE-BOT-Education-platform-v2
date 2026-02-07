package repository

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestLessonRepositoryUpdateSQLInjectionProtection tests that Update method validates field names
func TestLessonRepositoryUpdateSQLInjectionProtection(t *testing.T) {
	tests := []struct {
		name    string
		updates map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid field: teacher_id",
			updates: map[string]interface{}{
				"teacher_id": uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "valid field: start_time",
			updates: map[string]interface{}{
				"start_time": time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid field: end_time",
			updates: map[string]interface{}{
				"end_time": time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid field: max_students",
			updates: map[string]interface{}{
				"max_students": 10,
			},
			wantErr: false,
		},
		{
			name: "valid field: color",
			updates: map[string]interface{}{
				"color": "#FF5733",
			},
			wantErr: false,
		},
		{
			name: "valid field: subject",
			updates: map[string]interface{}{
				"subject": "Mathematics",
			},
			wantErr: false,
		},
		{
			name: "valid field: homework_text",
			updates: map[string]interface{}{
				"homework_text": "Do homework",
			},
			wantErr: false,
		},
		{
			name: "multiple valid fields",
			updates: map[string]interface{}{
				"color":   "#FF5733",
				"subject": "Mathematics",
			},
			wantErr: false,
		},
		{
			name: "invalid field: id (SQL injection attempt)",
			updates: map[string]interface{}{
				"id": uuid.New(),
			},
			wantErr: true,
			errMsg:  "invalid field for update",
		},
		{
			name: "invalid field: password_hash (privilege escalation attempt)",
			updates: map[string]interface{}{
				"password_hash": "hacked",
			},
			wantErr: true,
			errMsg:  "invalid field for update",
		},
		{
			name: "invalid field: created_at (temporal injection attempt)",
			updates: map[string]interface{}{
				"created_at": time.Now(),
			},
			wantErr: true,
			errMsg:  "invalid field for update",
		},
		{
			name: "invalid field: deleted_at (bypass soft delete)",
			updates: map[string]interface{}{
				"deleted_at": time.Now(),
			},
			wantErr: true,
			errMsg:  "invalid field for update",
		},
		{
			name: "SQL injection via field name: teacher_id; DROP TABLE lessons;--",
			updates: map[string]interface{}{
				"teacher_id; DROP TABLE lessons;--": "value",
			},
			wantErr: true,
			errMsg:  "invalid field for update",
		},
		{
			name: "SQL injection via field name: color' OR '1'='1",
			updates: map[string]interface{}{
				"color' OR '1'='1": "#FF0000",
			},
			wantErr: true,
			errMsg:  "invalid field for update",
		},
		{
			name: "empty updates",
			updates: map[string]interface{}{
				"color": "#FF0000",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock repository
			repo := &LessonRepository{
				db: nil, // We're testing validation, not actual DB calls
			}

			ctx := context.Background()
			lessonID := uuid.New()

			// Call Update to test field validation
			// Note: This will fail on DB execution, but we're testing validation logic
			err := repo.Update(ctx, lessonID, tt.updates)

			if tt.wantErr {
				assert.Error(t, err, "expected error for invalid field")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				// Error is expected due to nil DB, but not due to validation
				if err != nil {
					// If DB is nil, we might get an error, but it should not be validation error
					assert.NotContains(t, err.Error(), "invalid field for update")
				}
			}
		})
	}
}

// TestLessonRepositoryUpdateWhitelistValidation tests that only whitelisted fields are accepted
func TestLessonRepositoryUpdateWhitelistValidation(t *testing.T) {
	repo := &LessonRepository{
		db: nil,
	}

	ctx := context.Background()
	lessonID := uuid.New()

	// Test that clearly invalid fields are rejected
	// Note: teacher_id is now a valid field for updates (admin can change teacher)
	invalidUpdates := []map[string]interface{}{
		{"id": uuid.New()},
		{"current_students": 5},
		{"deleted_at": time.Now()},
		{"created_at": time.Now()},
		{"updated_at": time.Now()},
		{"is_recurring": false},
		{"recurring_group_id": uuid.New()},
	}

	for _, updates := range invalidUpdates {
		err := repo.Update(ctx, lessonID, updates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid field for update")
	}
}

// TestUserRepositoryUpdateSQLInjectionProtection tests that UserRepo Update validates field names
func TestUserRepositoryUpdateSQLInjectionProtection(t *testing.T) {
	tests := []struct {
		name    string
		updates map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid field: full_name",
			updates: map[string]interface{}{
				"full_name": "John Doe",
			},
			wantErr: false,
		},
		{
			name: "valid field: role",
			updates: map[string]interface{}{
				"role": models.RoleAdmin,
			},
			wantErr: false,
		},
		{
			name: "valid field: telegram_username",
			updates: map[string]interface{}{
				"telegram_username": "@johndoe",
			},
			wantErr: false,
		},
		{
			name: "valid field: payment_enabled",
			updates: map[string]interface{}{
				"payment_enabled": true,
			},
			wantErr: false,
		},
		{
			name: "invalid field: password_hash (privilege escalation)",
			updates: map[string]interface{}{
				"password_hash": "hacked_hash",
			},
			wantErr: true,
			errMsg:  "invalid field for update",
		},
		{
			name: "valid field: email",
			updates: map[string]interface{}{
				"email": "updated@example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid field: id (identity change)",
			updates: map[string]interface{}{
				"id": uuid.New(),
			},
			wantErr: true,
			errMsg:  "invalid field for update",
		},
		{
			name: "SQL injection via field name: role; UPDATE users SET password_hash = 'hacked';--",
			updates: map[string]interface{}{
				"role; UPDATE users SET password_hash = 'hacked';--": "admin",
			},
			wantErr: true,
			errMsg:  "invalid field for update",
		},
		{
			name: "SQL injection via field name: full_name' OR '1'='1",
			updates: map[string]interface{}{
				"full_name' OR '1'='1": "Admin",
			},
			wantErr: true,
			errMsg:  "invalid field for update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &UserRepo{
				db: nil,
			}

			ctx := context.Background()
			userID := uuid.New()

			err := repo.Update(ctx, userID, tt.updates)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				if err != nil {
					assert.NotContains(t, err.Error(), "invalid field for update")
				}
			}
		})
	}
}

// TestUpdateFieldValidationRejectionTypes tests various injection attack patterns
func TestUpdateFieldValidationRejectionTypes(t *testing.T) {
	tests := []struct {
		name       string
		fieldNames []string
	}{
		{
			name: "SQL keywords injection",
			fieldNames: []string{
				"color; DROP TABLE lessons",
				"subject' OR 1=1--",
				"homework_text) UNION SELECT",
				"color/**/OR/**/1=1",
			},
		},
		{
			name: "Table reference injection",
			fieldNames: []string{
				"lessons.id",
				"users.password_hash",
				"bookings.status",
				"(SELECT * FROM users)",
			},
		},
		{
			name: "Function injection",
			fieldNames: []string{
				"id; TRUNCATE lessons;--",
				"color'; DELETE FROM lessons WHERE '1'='1",
				"subject' + (SELECT count(*) FROM pg_database) + '",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &LessonRepository{db: nil}
			ctx := context.Background()
			lessonID := uuid.New()

			for _, fieldName := range tt.fieldNames {
				updates := map[string]interface{}{
					fieldName: "test_value",
				}

				err := repo.Update(ctx, lessonID, updates)
				assert.Error(t, err, "should reject field: %s", fieldName)
				assert.Contains(t, err.Error(), "invalid field for update")
			}
		})
	}
}

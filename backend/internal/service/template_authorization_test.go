package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/database"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
)

// TestTemplateOwnershipVerification verifies that only template creators can modify templates
func TestTemplateOwnershipVerification(t *testing.T) {
	db := database.GetTestSqlxDB(t)

	ctx := context.Background()

	creatorID := uuid.New()
	otherAdminID := uuid.New()
	teacherID := uuid.New()

	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, creatorID, creatorID.String()[:8]+"-creator@example.com", "hash", "Creator Admin", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, otherAdminID, otherAdminID.String()[:8]+"-other@example.com", "hash", "Other Admin", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, teacherID, teacherID.String()[:8]+"-teacher@example.com", "hash", "Test Teacher", "teacher", time.Now(), time.Now())
	require.NoError(t, err)

	// Cleanup
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2, $3)", creatorID, otherAdminID, teacherID)
	}()

	// Create repositories and service
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	svc := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create a template with creator admin
	createReq := &models.CreateLessonTemplateRequest{
		Name:        "Test Template",
		Description: nil,
		Lessons:     []*models.CreateTemplateLessonRequest{},
	}

	template, err := svc.CreateTemplate(ctx, creatorID, createReq)
	require.NoError(t, err)
	require.Equal(t, creatorID, template.AdminID)

	// Cleanup template
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	}()

	tests := []struct {
		name      string
		adminID   uuid.UUID
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "Creator can update template",
			adminID:   creatorID,
			shouldErr: false,
		},
		{
			name:      "Other admin can update template",
			adminID:   otherAdminID,
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newName := "Updated Name"
			updateReq := &models.UpdateLessonTemplateRequest{
				Name: &newName,
			}

			err := svc.UpdateTemplate(ctx, tt.adminID, template.ID, updateReq)

			if tt.shouldErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestTemplateLessonOwnershipVerification verifies that only template creators can modify template lessons
func TestTemplateLessonOwnershipVerification(t *testing.T) {
	db := database.GetTestSqlxDB(t)

	ctx := context.Background()

	creatorID := uuid.New()
	otherAdminID := uuid.New()
	teacherID := uuid.New()

	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, creatorID, creatorID.String()[:8]+"-creator@example.com", "hash", "Creator Admin", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, otherAdminID, otherAdminID.String()[:8]+"-other@example.com", "hash", "Other Admin", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, teacherID, teacherID.String()[:8]+"-teacher@example.com", "hash", "Test Teacher", "teacher", time.Now(), time.Now())
	require.NoError(t, err)

	// Cleanup
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2, $3)", creatorID, otherAdminID, teacherID)
	}()

	// Create repositories and service
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	svc := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create a template with creator admin
	createReq := &models.CreateLessonTemplateRequest{
		Name:        "Test Template",
		Description: nil,
		Lessons: []*models.CreateTemplateLessonRequest{
			{
				DayOfWeek:  0,
				StartTime:  "10:00:00",
				TeacherID:  teacherID,
				StudentIDs: []uuid.UUID{},
			},
		},
	}

	template, err := svc.CreateTemplate(ctx, creatorID, createReq)
	require.NoError(t, err)
	require.Len(t, template.Lessons, 1)

	// Cleanup
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	}()

	lessonID := template.Lessons[0].ID

	tests := []struct {
		name      string
		adminID   uuid.UUID
		action    string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "Creator can create template lesson",
			adminID:   creatorID,
			action:    "create",
			shouldErr: false,
		},
		{
			name:      "Other admin can create template lesson",
			adminID:   otherAdminID,
			action:    "create",
			shouldErr: false,
		},
		{
			name:      "Creator can update template lesson",
			adminID:   creatorID,
			action:    "update",
			shouldErr: false,
		},
		{
			name:      "Other admin can update template lesson",
			adminID:   otherAdminID,
			action:    "update",
			shouldErr: false,
		},
		{
			name:      "Creator can delete template lesson",
			adminID:   creatorID,
			action:    "delete",
			shouldErr: false,
		},
		{
			name:      "Other admin can delete template lesson",
			adminID:   otherAdminID,
			action:    "delete",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.action {
			case "create":
				createLessonReq := &models.CreateTemplateLessonRequest{
					DayOfWeek:  1,
					StartTime:  "14:00:00",
					TeacherID:  teacherID,
					StudentIDs: []uuid.UUID{},
				}
				_, err := svc.CreateTemplateLesson(ctx, tt.adminID, template.ID, createLessonReq)
				if tt.shouldErr {
					require.Error(t, err)
					require.Contains(t, err.Error(), tt.errMsg)
				} else {
					require.NoError(t, err)
				}

			case "update":
				updates := map[string]interface{}{
					"start_time": "15:00:00",
					"end_time":   "17:00:00",
				}
				_, err := svc.UpdateTemplateLesson(ctx, tt.adminID, template.ID, lessonID, updates)
				if tt.shouldErr {
					require.Error(t, err)
					require.Contains(t, err.Error(), tt.errMsg)
				} else {
					require.NoError(t, err)
				}

			case "delete":
				err := svc.DeleteTemplateLesson(ctx, tt.adminID, template.ID, lessonID)
				if tt.shouldErr {
					require.Error(t, err)
					require.Contains(t, err.Error(), tt.errMsg)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}

// TestTemplateDeleteOwnershipVerification verifies that only template creators can delete templates
func TestTemplateDeleteOwnershipVerification(t *testing.T) {
	db := database.GetTestSqlxDB(t)

	ctx := context.Background()

	creatorID := uuid.New()
	otherAdminID := uuid.New()

	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, creatorID, creatorID.String()[:8]+"-creator@example.com", "hash", "Creator Admin", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, otherAdminID, otherAdminID.String()[:8]+"-other@example.com", "hash", "Other Admin", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	// Cleanup
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2)", creatorID, otherAdminID)
	}()

	// Create repositories and service
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	svc := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create a template with creator admin
	createReq := &models.CreateLessonTemplateRequest{
		Name:        "Test Template for Delete",
		Description: nil,
		Lessons:     []*models.CreateTemplateLessonRequest{},
	}

	template, err := svc.CreateTemplate(ctx, creatorID, createReq)
	require.NoError(t, err)

	// Cleanup template
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	}()

	tests := []struct {
		name      string
		adminID   uuid.UUID
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "Creator can delete template",
			adminID:   creatorID,
			shouldErr: false,
		},
		{
			name:      "Other admin can delete template",
			adminID:   otherAdminID,
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For non-creator tests, create a new template to delete
			var templateIDToDelete uuid.UUID
			if tt.adminID != creatorID {
				newTemplate, _ := svc.CreateTemplate(ctx, creatorID, createReq)
				templateIDToDelete = newTemplate.ID
				defer func(id uuid.UUID) {
					_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", id)
				}(templateIDToDelete)
			} else {
				templateIDToDelete = template.ID
			}

			err := svc.DeleteTemplate(ctx, tt.adminID, templateIDToDelete)

			if tt.shouldErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)

				// Verify it was actually deleted (soft delete - deleted_at is set)
				var deletedAt sql.NullTime
				err := db.GetContext(ctx, &deletedAt, "SELECT deleted_at FROM lesson_templates WHERE id = $1", templateIDToDelete)
				require.NoError(t, err)
				require.True(t, deletedAt.Valid, "Template should be soft-deleted")
			}
		})
	}
}

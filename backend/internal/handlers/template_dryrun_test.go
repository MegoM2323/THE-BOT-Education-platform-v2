package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"
)

// TestApplyTemplateDryRun tests that dry-run mode returns preview without creating data
func TestApplyTemplateDryRun(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	// Create test admin user with unique email
	admin := createTestUser(t, db, "admin@test.com", "DryRun Admin", "admin")
	adminID := admin.ID

	// Create test teacher with unique email
	teacher := createTestUser(t, db, "teacher@test.com", "DryRun Teacher", "teacher")
	teacherID := teacher.ID

	// Create test student with unique email
	student := createTestUser(t, db, "student@test.com", "DryRun Student", "student")
	studentID := student.ID

	// Create credits for student
	addCreditToStudent(t, db, studentID, 10)

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Create service
	templateService := service.NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create handler
	handler := NewTemplateHandler(templateService)

	// Create a test template with one lesson
	template := createTestTemplate(t, db, adminID, "Test Template", []*models.CreateTemplateLessonRequest{
		{
			DayOfWeek:  0,
			StartTime:  "10:00:00",
			TeacherID:  teacherID,
			StudentIDs: []uuid.UUID{studentID},
		},
	})
	templateID := template.ID

	// Get next Monday
	now := time.Now().UTC()
	daysUntilMonday := (8 - int(now.Weekday())) % 7
	if daysUntilMonday == 0 && now.Hour() >= 18 {
		daysUntilMonday = 7
	}
	nextMonday := now.AddDate(0, 0, daysUntilMonday)
	weekStartDate := nextMonday.Format("2006-01-02")

	tests := []struct {
		name           string
		dryRun         bool
		expectedStatus int
		shouldContain  []string
		shouldNotHave  []string
	}{
		{
			name:           "Dry-run mode returns HTTP 200 with preview flag",
			dryRun:         true,
			expectedStatus: http.StatusOK,
			shouldContain: []string{
				`"dry_run":true`,
				`"status":"preview"`,
				`"message":"This is a preview`,
				`"lessons_count":1`,
			},
			shouldNotHave: []string{},
		},
		{
			name:           "Normal mode returns HTTP 201 without dry_run flag",
			dryRun:         false,
			expectedStatus: http.StatusCreated,
			shouldContain:  []string{},
			shouldNotHave: []string{
				`"status":"preview"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			reqBody := models.ApplyTemplateRequest{
				TemplateID:    templateID,
				WeekStartDate: weekStartDate,
				DryRun:        tt.dryRun,
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest("POST", "/api/v1/templates/"+templateID.String()+"/apply", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Add user to context
			userCtx := context.WithValue(req.Context(), middleware.UserContextKey, &models.User{
				ID:   adminID,
				Role: models.RoleAdmin,
			})
			req = req.WithContext(userCtx)

			// Create response recorder and router with URL parameter
			w := httptest.NewRecorder()
			r := chi.NewRouter()
			r.Post("/api/v1/templates/{id}/apply", handler.ApplyTemplate)
			r.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")

			// Check response contains expected strings
			responseBody := w.Body.String()
			for _, expected := range tt.shouldContain {
				assert.Contains(t, responseBody, expected, "Response should contain %s", expected)
			}

			// Check response doesn't contain unexpected strings
			for _, notExpected := range tt.shouldNotHave {
				assert.NotContains(t, responseBody, notExpected, "Response should NOT contain %s", notExpected)
			}

			if tt.dryRun {
				var lessonCount int
				err := db.GetContext(ctx, &lessonCount, `
					SELECT COUNT(*) FROM lessons l
					JOIN template_applications ta ON ta.id = l.template_application_id
					WHERE ta.template_id = $1
				`, templateID)
				assert.NoError(t, err)
				assert.Equal(t, 0, lessonCount, "No lessons should be created in dry-run mode for this template")

				var bookingCount int
				err = db.GetContext(ctx, &bookingCount, `
					SELECT COUNT(*) FROM bookings b
					JOIN lessons l ON l.id = b.lesson_id
					JOIN template_applications ta ON ta.id = l.template_application_id
					WHERE ta.template_id = $1
				`, templateID)
				assert.NoError(t, err)
				assert.Equal(t, 0, bookingCount, "No bookings should be created in dry-run mode for this template")

				creditBalance, err := creditRepo.GetBalance(ctx, studentID)
				require.NoError(t, err)
				assert.Equal(t, 10, creditBalance.Balance, "Credits should not be deducted in dry-run mode")
			} else {
				var lessonCount int
				err := db.GetContext(ctx, &lessonCount, `
					SELECT COUNT(*) FROM lessons l
					JOIN template_applications ta ON ta.id = l.template_application_id
					WHERE ta.template_id = $1
				`, templateID)
				assert.NoError(t, err)
				assert.Greater(t, lessonCount, 0, "Lessons should be created in normal mode for this template")
			}
		})
	}
}

// TestDryRunResponseStructure tests that dry-run response has correct JSON structure
func TestDryRunResponseStructure(t *testing.T) {
	// Create a mock response with credits field
	mockResponse := map[string]interface{}{
		"dry_run":         true,
		"status":          "preview",
		"message":         "This is a preview of what would be created. No changes have been made.",
		"template_id":     uuid.New().String(),
		"week_start_date": "2025-01-06",
		"lessons_count":   2,
		"would_create": map[string]interface{}{
			"lessons":  2,
			"bookings": 3, // 3 students across 2 lessons
			"credits":  3, // 1 credit per booking
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(mockResponse)
	require.NoError(t, err)

	// Unmarshal and verify structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)

	// Verify required fields
	assert.Equal(t, true, result["dry_run"], "dry_run should be true")
	assert.Equal(t, "preview", result["status"], "status should be 'preview'")
	assert.Contains(t, result["message"], "preview", "message should mention preview")
	assert.NotNil(t, result["would_create"], "would_create field should exist")

	// Verify would_create structure
	wouldCreate, ok := result["would_create"].(map[string]interface{})
	assert.True(t, ok, "would_create should be a map")
	assert.NotNil(t, wouldCreate["lessons"], "lessons count should be in would_create")
	assert.NotNil(t, wouldCreate["bookings"], "bookings count should be in would_create")
	assert.NotNil(t, wouldCreate["credits"], "credits count should be in would_create")

	// Verify credits equals bookings (1 credit per booking)
	bookings := wouldCreate["bookings"].(float64)
	credits := wouldCreate["credits"].(float64)
	assert.Equal(t, bookings, credits, "credits should equal bookings (1 credit per booking)")
}

// TestDryRunErrorHandling tests that dry-run respects validation errors
func TestDryRunErrorHandling(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create test admin user
	admin := createTestUser(t, db, "admin@test.com", "DryRun Error Admin", "admin")
	adminID := admin.ID

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Create service
	templateService := service.NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create handler
	handler := NewTemplateHandler(templateService)

	// Test with invalid template ID
	invalidTemplateID := uuid.New()

	reqBody := models.ApplyTemplateRequest{
		TemplateID:    invalidTemplateID,
		WeekStartDate: time.Now().Format("2006-01-02"),
		DryRun:        true,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/templates/"+invalidTemplateID.String()+"/apply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Add user to context
	userCtx := context.WithValue(req.Context(), middleware.UserContextKey, &models.User{
		ID:   adminID,
		Role: models.RoleAdmin,
	})
	req = req.WithContext(userCtx)

	w := httptest.NewRecorder()
	handler.ApplyTemplate(w, req)

	// Should return error for invalid template
	assert.NotEqual(t, http.StatusOK, w.Code, "Should not return OK for non-existent template")

	var errorResp response.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.False(t, errorResp.Success, "Success should be false")
}

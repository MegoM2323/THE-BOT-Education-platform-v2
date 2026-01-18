package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
)

// TestTemplateRollbackAPI_Success tests successful rollback via HTTP API
func TestTemplateRollbackAPI_Success(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)

	// Create service and handler
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)
	templateService := service.NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)
	handler := NewTemplateHandler(templateService)

	// Create test users
	admin := createTestUser(t, db, "admin@test.com", "Admin User", "admin")
	teacher := createTestUser(t, db, "teacher@test.com", "Teacher User", "teacher")
	student := createTestUser(t, db, "student@test.com", "Student User", "student")

	// Add credits
	addCreditToStudent(t, db, student.ID, 1)

	// Create and apply template
	template := createTestTemplate(t, db, admin.ID, "Test Template", []*models.CreateTemplateLessonRequest{
		{
			DayOfWeek:  1,
			StartTime:  "10:00:00",
			TeacherID:  teacher.ID,
			StudentIDs: []uuid.UUID{student.ID},
		},
	})

	weekStartDate := getNextMonday().Format("2006-01-02")
	applyReq := &models.ApplyTemplateRequest{
		TemplateID:    template.ID,
		WeekStartDate: weekStartDate,
	}

	ctx := context.Background()
	_, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, applyReq)
	require.NoError(t, err)

	// Prepare rollback request
	reqBody := map[string]string{
		"week_start_date": weekStartDate,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+template.ID.String()+"/rollback", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Add admin user to context
	ctx = context.WithValue(req.Context(), middleware.UserContextKey, admin)
	req = req.WithContext(ctx)

	// Create router with URL parameter
	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/api/v1/templates/{id}/rollback", handler.RollbackTemplate)
	r.ServeHTTP(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["deleted_lessons"])
	assert.Equal(t, float64(1), data["refunded_credits"])
	assert.Equal(t, weekStartDate, data["week_start_date"])

	// Verify credit refunded
	balance := getCreditBalance(t, db, student.ID)
	assert.Equal(t, 1, balance)
}

// TestTemplateRollbackAPI_Unauthorized tests rollback without authentication
func TestTemplateRollbackAPI_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create service and handler
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)

	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)
	templateService := service.NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)
	handler := NewTemplateHandler(templateService)

	// Prepare request without user context
	templateID := uuid.New()
	reqBody := map[string]string{
		"week_start_date": "2025-12-01",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID.String()+"/rollback", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.RollbackTemplate(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestTemplateRollbackAPI_NonAdmin tests rollback by non-admin user
func TestTemplateRollbackAPI_NonAdmin(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create service and handler
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)

	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)
	templateService := service.NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)
	handler := NewTemplateHandler(templateService)

	// Create student user
	student := createTestUser(t, db, "student@test.com", "Student User", "student")

	// Prepare request
	templateID := uuid.New()
	reqBody := map[string]string{
		"week_start_date": "2025-12-01",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID.String()+"/rollback", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Add student user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, student)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.RollbackTemplate(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// TestTemplateRollbackAPI_InvalidTemplateID tests rollback with invalid template ID
func TestTemplateRollbackAPI_InvalidTemplateID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create service and handler
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)

	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)
	templateService := service.NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)
	handler := NewTemplateHandler(templateService)

	// Create admin
	admin := createTestUser(t, db, "admin@test.com", "Admin User", "admin")

	// Prepare request with invalid UUID
	reqBody := map[string]string{
		"week_start_date": "2025-12-01",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/invalid-uuid/rollback", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Add admin user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, admin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/api/v1/templates/{id}/rollback", handler.RollbackTemplate)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestTemplateRollbackAPI_MissingWeekDate tests rollback without week_start_date
func TestTemplateRollbackAPI_MissingWeekDate(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create service and handler
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)

	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)
	templateService := service.NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)
	handler := NewTemplateHandler(templateService)

	// Create admin
	admin := createTestUser(t, db, "admin@test.com", "Admin User", "admin")

	// Prepare request without week_start_date
	templateID := uuid.New()
	reqBody := map[string]string{}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID.String()+"/rollback", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Add admin user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, admin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/api/v1/templates/{id}/rollback", handler.RollbackTemplate)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestTemplateRollbackAPI_NotApplied tests rollback when template not applied to week
func TestTemplateRollbackAPI_NotApplied(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)

	// Create service and handler
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)
	templateService := service.NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)
	handler := NewTemplateHandler(templateService)

	// Create admin and template (not applied)
	admin := createTestUser(t, db, "admin@test.com", "Admin User", "admin")
	template := createTestTemplate(t, db, admin.ID, "Test Template", []*models.CreateTemplateLessonRequest{})

	// Prepare rollback request
	weekStartDate := getNextMonday().Format("2006-01-02")
	reqBody := map[string]string{
		"week_start_date": weekStartDate,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+template.ID.String()+"/rollback", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Add admin user to context
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, admin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/api/v1/templates/{id}/rollback", handler.RollbackTemplate)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestTemplateRollbackAPI_AlreadyRolledBack tests rollback of already rolled back application
func TestTemplateRollbackAPI_AlreadyRolledBack(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)

	// Create service and handler
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)
	templateService := service.NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)
	handler := NewTemplateHandler(templateService)

	// Create test users
	admin := createTestUser(t, db, "admin@test.com", "Admin User", "admin")
	teacher := createTestUser(t, db, "teacher@test.com", "Teacher User", "teacher")
	student := createTestUser(t, db, "student@test.com", "Student User", "student")

	// Add credits
	addCreditToStudent(t, db, student.ID, 1)

	// Create and apply template
	template := createTestTemplate(t, db, admin.ID, "Test Template", []*models.CreateTemplateLessonRequest{
		{
			DayOfWeek:  1,
			StartTime:  "10:00:00",
			TeacherID:  teacher.ID,
			StudentIDs: []uuid.UUID{student.ID},
		},
	})

	weekStartDate := getNextMonday().Format("2006-01-02")
	applyReq := &models.ApplyTemplateRequest{
		TemplateID:    template.ID,
		WeekStartDate: weekStartDate,
	}

	ctx := context.Background()
	_, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, applyReq)
	require.NoError(t, err)

	// First rollback
	_, err = templateService.RollbackWeekToTemplate(ctx, admin.ID, weekStartDate, template.ID)
	require.NoError(t, err)

	// Try second rollback via API
	reqBody := map[string]string{
		"week_start_date": weekStartDate,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+template.ID.String()+"/rollback", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Add admin user to context
	ctx = context.WithValue(req.Context(), middleware.UserContextKey, admin)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Post("/api/v1/templates/{id}/rollback", handler.RollbackTemplate)
	r.ServeHTTP(rr, req)

	// Should return 404 (no active application found)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// ============================================================================
// Helper Functions - Defined in test_helpers.go to avoid duplication
// ============================================================================
// The following helper functions are defined in test_helpers.go:
// - setupTestDB()
// - cleanupTestDB()
// - createTestUser()
// - addCreditToStudent()
// - getCreditBalance()
// - createTestTemplate()
// - getNextMonday()

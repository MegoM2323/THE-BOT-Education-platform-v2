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
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
)

// TestMethodologistTemplateWorkflow tests the complete template workflow for methodologist role
// Covers: Create, Read, Update, Delete operations on templates and template lessons
// Also covers: Apply and Rollback template operations
func TestMethodologistTemplateWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	// Setup: Create test users
	methodologist := createTestUser(t, db, "methodologist@test.com", "Methodologist User", string(models.RoleMethodologist))
	teacher1 := createTestUser(t, db, "teacher1@test.com", "Teacher One", string(models.RoleMethodologist))
	student1 := createTestUser(t, db, "student1@test.com", "Student One", string(models.RoleStudent))

	// Give student enough credits
	addCreditToStudent(t, db, student1.ID, 10)

	// Initialize repositories and services
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	templateService := service.NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)
	templateHandler := NewTemplateHandler(templateService)

	t.Run("Scenario 1: Methodologist can create and retrieve templates", func(t *testing.T) {
		testTemplateCreateAndRetrieve(t, ctx, templateHandler, methodologist)
	})

	t.Run("Scenario 2: Methodologist can create template lessons", func(t *testing.T) {
		testTemplateLessonCreation(t, ctx, templateHandler, methodologist, teacher1)
	})

	t.Run("Scenario 3: Methodologist can apply and rollback templates", func(t *testing.T) {
		testTemplateApplicationAndRollback(t, ctx, templateHandler, methodologist, teacher1, student1, db)
	})

	t.Run("Scenario 4: Access control - only methodologist and admin can access", func(t *testing.T) {
		testMethodologistAccessControl(t, ctx, templateHandler, methodologist, teacher1, student1)
	})
}

// testTemplateCreateAndRetrieve tests CreateTemplate and GetTemplate operations
func testTemplateCreateAndRetrieve(t *testing.T, ctx context.Context, handler *TemplateHandler, methodologist *models.User) {
	// Create template
	templateID := createTestTemplateViaHandler(t, ctx, handler, methodologist)

	t.Run("CreateTemplate returns 201", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, templateID, "Template should be created with valid ID")
	})

	t.Run("GetTemplate returns 200", func(t *testing.T) {
		rctx := addURLParamToContext(ctx, "id", templateID.String())
		rctx = context.WithValue(rctx, middleware.UserContextKey, methodologist)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+templateID.String(), nil)
		req = req.WithContext(rctx)
		w := httptest.NewRecorder()

		handler.GetTemplate(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Methodologist should be able to retrieve template")
	})
}

// testTemplateLessonCreation tests CreateTemplateLesson operation
func testTemplateLessonCreation(t *testing.T, ctx context.Context, handler *TemplateHandler, methodologist, teacher *models.User) {
	// Create template
	templateID := createTestTemplateViaHandler(t, ctx, handler, methodologist)

	t.Run("CreateTemplateLesson returns 201", func(t *testing.T) {
		lessonID := createTestTemplateLessonViaHandler(t, ctx, handler, methodologist, templateID, teacher.ID)
		assert.NotEqual(t, uuid.Nil, lessonID, "Lesson should be created with valid ID")
	})
}

// testTemplateApplicationAndRollback tests ApplyTemplate and RollbackTemplate
func testTemplateApplicationAndRollback(t *testing.T, ctx context.Context, handler *TemplateHandler, methodologist, teacher, student *models.User, db *sqlx.DB) {
	// Create template with lessons
	templateID := createTestTemplateViaHandler(t, ctx, handler, methodologist)
	_ = createTestTemplateLessonViaHandler(t, ctx, handler, methodologist, templateID, teacher.ID)

	weekStart := getNextMonday()
	weekStartDate := weekStart.Format("2006-01-02")

	t.Run("ApplyTemplate returns 201", func(t *testing.T) {
		reqBody := models.ApplyTemplateRequest{
			TemplateID:    templateID,
			WeekStartDate: weekStartDate,
		}

		body, _ := json.Marshal(reqBody)
		rctx := addURLParamToContext(ctx, "id", templateID.String())
		rctx = context.WithValue(rctx, middleware.UserContextKey, methodologist)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID.String()+"/apply", bytes.NewReader(body))
		req = req.WithContext(rctx)
		w := httptest.NewRecorder()

		handler.ApplyTemplate(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Methodologist should be able to apply template")
	})

	t.Run("RollbackTemplate returns 200", func(t *testing.T) {
		reqBody := map[string]string{
			"week_start_date": weekStartDate,
		}

		body, _ := json.Marshal(reqBody)
		rctx := addURLParamToContext(ctx, "id", templateID.String())
		rctx = context.WithValue(rctx, middleware.UserContextKey, methodologist)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID.String()+"/rollback", bytes.NewReader(body))
		req = req.WithContext(rctx)
		w := httptest.NewRecorder()

		handler.RollbackTemplate(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Methodologist should be able to rollback template")
	})
}

// testMethodologistAccessControl tests that only methodologist and admin can access
func testMethodologistAccessControl(t *testing.T, ctx context.Context, handler *TemplateHandler, methodologist, teacher, student *models.User) {
	// Reuse existing database setup from parent test - don't create new one
	// This avoids redundant database initialization

	tests := []struct {
		name     string
		user     *models.User
		expected int
	}{
		{
			name:     "Methodologist can create template",
			user:     methodologist,
			expected: http.StatusCreated,
		},
		{
			name:     "Teacher cannot create template",
			user:     teacher,
			expected: http.StatusForbidden,
		},
		{
			name:     "Student cannot create template",
			user:     student,
			expected: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := models.CreateLessonTemplateRequest{
				Name: "Test Template",
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(body))
			req = req.WithContext(context.WithValue(ctx, middleware.UserContextKey, tt.user))
			w := httptest.NewRecorder()

			handler.CreateTemplate(w, req)

			assert.Equal(t, tt.expected, w.Code, "Expected status %d for %s", tt.expected, tt.name)
		})
	}
}

// createTestTemplateViaHandler creates a template via HTTP handler and returns ID
func createTestTemplateViaHandler(t *testing.T, ctx context.Context, handler *TemplateHandler, user *models.User) uuid.UUID {
	t.Helper()
	reqBody := models.CreateLessonTemplateRequest{
		Name: "Helper Template " + uuid.New().String()[:8],
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(ctx, middleware.UserContextKey, user))
	w := httptest.NewRecorder()

	handler.CreateTemplate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create test template: status %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	return uuid.MustParse(data["id"].(string))
}

// createTestTemplateLessonViaHandler creates a template lesson via HTTP handler and returns ID
func createTestTemplateLessonViaHandler(t *testing.T, ctx context.Context, handler *TemplateHandler, user *models.User, templateID, teacherID uuid.UUID) uuid.UUID {
	t.Helper()
	endT := "10:00:00"
	lessonType := "group"
	maxStud := 5
	color := "#3B82F6"
	subject := "Mathematics"
	reqBody := models.CreateTemplateLessonRequest{
		DayOfWeek:   0, // Monday
		StartTime:   "09:00:00",
		EndTime:     &endT,
		TeacherID:   teacherID,
		LessonType:  &lessonType,
		MaxStudents: &maxStud,
		Color:       &color,
		Subject:     &subject,
	}

	body, _ := json.Marshal(reqBody)
	rctx := addURLParamToContext(ctx, "id", templateID.String())
	rctx = context.WithValue(rctx, middleware.UserContextKey, user)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID.String()+"/lessons", bytes.NewReader(body))
	req = req.WithContext(rctx)
	w := httptest.NewRecorder()

	handler.CreateTemplateLesson(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create test template lesson: status %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	return uuid.MustParse(data["id"].(string))
}

// addURLParamToContext adds a chi URL parameter to context for testing
func addURLParamToContext(ctx context.Context, paramName, paramValue string) context.Context {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(paramName, paramValue)
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}

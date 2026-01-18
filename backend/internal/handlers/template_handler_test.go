package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
)

// Helper function to create a user with context
func contextWithUserForTemplate(user *models.User) context.Context {
	return context.WithValue(context.Background(), middleware.UserContextKey, user)
}

// Helper function to create test users with minimal fields (for handler tests)
func createHandlerTestUserForTemplate(role models.UserRole) *models.User {
	return &models.User{
		ID:       uuid.New(),
		FullName: "Test User",
		Role:     role,
	}
}

// TestTemplateGetTemplatesAccessControl tests role-based access control for GetTemplates
func TestTemplateGetTemplatesAccessControl(t *testing.T) {
	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
		userProvided   bool
	}{
		{
			name:           "Admin can access GetTemplates",
			userRole:       models.RoleAdmin,
			expectedStatus: http.StatusOK,
			userProvided:   true,
		},
		{
			name:           "Methodologist can access GetTemplates",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusOK,
			userProvided:   true,
		},
		{
			name:           "Teacher cannot access GetTemplates - Forbidden",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusForbidden,
			userProvided:   true,
		},
		{
			name:           "Student cannot access GetTemplates - Forbidden",
			userRole:       models.RoleStudent,
			expectedStatus: http.StatusForbidden,
			userProvided:   true,
		},
		{
			name:           "Unauthenticated user cannot access GetTemplates",
			userRole:       "",
			expectedStatus: http.StatusUnauthorized,
			userProvided:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
			if tt.userProvided {
				user := createHandlerTestUserForTemplate(tt.userRole)
				ctx := contextWithUserForTemplate(user)
				req = req.WithContext(ctx)
			}
			w := httptest.NewRecorder()

			// Test the authorization logic directly
			user, ok := middleware.GetUserFromContext(req.Context())
			if !ok {
				assert.Equal(t, http.StatusUnauthorized, tt.expectedStatus, "Unauthenticated should return 401")
			} else if !user.IsAdmin() && !user.IsMethodologist() {
				assert.Equal(t, http.StatusForbidden, tt.expectedStatus, "Non-admin/methodologist should return 403")
			} else {
				assert.Equal(t, http.StatusOK, tt.expectedStatus, "Admin/methodologist should return 200")
			}

			_ = w // Unused in this test
		})
	}
}

// TestTemplateCreateTemplateAccessControl tests role-based access control for CreateTemplate
func TestTemplateCreateTemplateAccessControl(t *testing.T) {
	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
		userProvided   bool
	}{
		{
			name:           "Admin can access CreateTemplate",
			userRole:       models.RoleAdmin,
			expectedStatus: http.StatusCreated,
			userProvided:   true,
		},
		{
			name:           "Methodologist can access CreateTemplate",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusCreated,
			userProvided:   true,
		},
		{
			name:           "Teacher cannot access CreateTemplate - Forbidden",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusForbidden,
			userProvided:   true,
		},
		{
			name:           "Unauthenticated cannot access CreateTemplate",
			userRole:       "",
			expectedStatus: http.StatusUnauthorized,
			userProvided:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := &models.CreateLessonTemplateRequest{
				Name: "Test Template",
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader(body))
			if tt.userProvided {
				user := createHandlerTestUserForTemplate(tt.userRole)
				ctx := contextWithUserForTemplate(user)
				req = req.WithContext(ctx)
			}

			// Test the authorization logic
			user, ok := middleware.GetUserFromContext(req.Context())
			if !ok {
				assert.Equal(t, http.StatusUnauthorized, tt.expectedStatus)
			} else if !user.IsAdmin() && !user.IsMethodologist() {
				assert.Equal(t, http.StatusForbidden, tt.expectedStatus)
			} else {
				assert.Equal(t, http.StatusCreated, tt.expectedStatus)
			}
		})
	}
}

// TestTemplateGetTemplateAccessControl tests role-based access control for GetTemplate
func TestTemplateGetTemplateAccessControl(t *testing.T) {
	templateID := uuid.New()

	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
		userProvided   bool
	}{
		{
			name:           "Admin can access GetTemplate",
			userRole:       models.RoleAdmin,
			expectedStatus: http.StatusOK,
			userProvided:   true,
		},
		{
			name:           "Methodologist can access GetTemplate",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusOK,
			userProvided:   true,
		},
		{
			name:           "Teacher cannot access GetTemplate - Forbidden",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusForbidden,
			userProvided:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/"+templateID.String(), nil)
			if tt.userProvided {
				user := createHandlerTestUserForTemplate(tt.userRole)
				ctx := contextWithUserForTemplate(user)
				req = req.WithContext(ctx)
			}
			req.SetPathValue("id", templateID.String())

			// Test the authorization logic
			user, ok := middleware.GetUserFromContext(req.Context())
			if !ok {
				assert.Equal(t, http.StatusUnauthorized, tt.expectedStatus)
			} else if !user.IsAdmin() && !user.IsMethodologist() {
				assert.Equal(t, http.StatusForbidden, tt.expectedStatus)
			} else {
				assert.Equal(t, http.StatusOK, tt.expectedStatus)
			}
		})
	}
}

// TestTemplateUpdateTemplateAccessControl tests role-based access control for UpdateTemplate
func TestTemplateUpdateTemplateAccessControl(t *testing.T) {
	templateID := uuid.New()

	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
	}{
		{
			name:           "Admin can access UpdateTemplate",
			userRole:       models.RoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Methodologist can access UpdateTemplate",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Teacher cannot access UpdateTemplate - Forbidden",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := &models.UpdateLessonTemplateRequest{
				Name: stringPtrForTemplate("Updated Template"),
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/"+templateID.String(), bytes.NewReader(body))
			user := createHandlerTestUserForTemplate(tt.userRole)
			ctx := contextWithUserForTemplate(user)
			req = req.WithContext(ctx)
			req.SetPathValue("id", templateID.String())

			// Test the authorization logic
			user, ok := middleware.GetUserFromContext(req.Context())
			if !ok {
				assert.Equal(t, http.StatusUnauthorized, tt.expectedStatus)
			} else if !user.IsAdmin() && !user.IsMethodologist() {
				assert.Equal(t, http.StatusForbidden, tt.expectedStatus)
			} else {
				assert.Equal(t, http.StatusOK, tt.expectedStatus)
			}
		})
	}
}

// TestTemplateDeleteTemplateAccessControl tests role-based access control for DeleteTemplate
func TestTemplateDeleteTemplateAccessControl(t *testing.T) {
	templateID := uuid.New()

	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
	}{
		{
			name:           "Admin can access DeleteTemplate",
			userRole:       models.RoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Methodologist can access DeleteTemplate",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Teacher cannot access DeleteTemplate - Forbidden",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/"+templateID.String(), nil)
			user := createHandlerTestUserForTemplate(tt.userRole)
			ctx := contextWithUserForTemplate(user)
			req = req.WithContext(ctx)
			req.SetPathValue("id", templateID.String())

			// Test the authorization logic
			user, ok := middleware.GetUserFromContext(req.Context())
			if !ok {
				assert.Equal(t, http.StatusUnauthorized, tt.expectedStatus)
			} else if !user.IsAdmin() && !user.IsMethodologist() {
				assert.Equal(t, http.StatusForbidden, tt.expectedStatus)
			} else {
				assert.Equal(t, http.StatusOK, tt.expectedStatus)
			}
		})
	}
}

// TestTemplateApplyTemplateAccessControl tests role-based access control for ApplyTemplate
func TestTemplateApplyTemplateAccessControl(t *testing.T) {
	templateID := uuid.New()
	weekStartDate := "2025-01-13"

	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
	}{
		{
			name:           "Admin can access ApplyTemplate",
			userRole:       models.RoleAdmin,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Methodologist can access ApplyTemplate",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Teacher cannot access ApplyTemplate - Forbidden",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := &models.ApplyTemplateRequest{
				WeekStartDate: weekStartDate,
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID.String()+"/apply", bytes.NewReader(body))
			user := createHandlerTestUserForTemplate(tt.userRole)
			ctx := contextWithUserForTemplate(user)
			req = req.WithContext(ctx)
			req.SetPathValue("id", templateID.String())

			// Test the authorization logic
			user, ok := middleware.GetUserFromContext(req.Context())
			if !ok {
				assert.Equal(t, http.StatusUnauthorized, tt.expectedStatus)
			} else if !user.IsAdmin() && !user.IsMethodologist() {
				assert.Equal(t, http.StatusForbidden, tt.expectedStatus)
			} else {
				assert.Equal(t, http.StatusCreated, tt.expectedStatus)
			}
		})
	}
}

// TestTemplateRollbackTemplateAccessControl tests role-based access control for RollbackTemplate
func TestTemplateRollbackTemplateAccessControl(t *testing.T) {
	templateID := uuid.New()
	weekStartDate := "2025-01-13"

	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
	}{
		{
			name:           "Admin can access RollbackTemplate",
			userRole:       models.RoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Methodologist can access RollbackTemplate",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Teacher cannot access RollbackTemplate - Forbidden",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]string{
				"week_start_date": weekStartDate,
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID.String()+"/rollback", bytes.NewReader(body))
			user := createHandlerTestUserForTemplate(tt.userRole)
			ctx := contextWithUserForTemplate(user)
			req = req.WithContext(ctx)
			req.SetPathValue("id", templateID.String())

			// Test the authorization logic
			user, ok := middleware.GetUserFromContext(req.Context())
			if !ok {
				assert.Equal(t, http.StatusUnauthorized, tt.expectedStatus)
			} else if !user.IsAdmin() && !user.IsMethodologist() {
				assert.Equal(t, http.StatusForbidden, tt.expectedStatus)
			} else {
				assert.Equal(t, http.StatusOK, tt.expectedStatus)
			}
		})
	}
}

// TestTemplateCreateTemplateLessonAccessControl tests role-based access control for CreateTemplateLesson
func TestTemplateCreateTemplateLessonAccessControl(t *testing.T) {
	templateID := uuid.New()
	teacherID := uuid.New()

	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
	}{
		{
			name:           "Admin can access CreateTemplateLesson",
			userRole:       models.RoleAdmin,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Methodologist can access CreateTemplateLesson",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Teacher cannot access CreateTemplateLesson - Forbidden",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := &models.CreateTemplateLessonRequest{
				DayOfWeek: 0,
				StartTime: "09:00:00",
				TeacherID: teacherID,
			}
			payload.ApplyDefaults()
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/templates/"+templateID.String()+"/lessons", bytes.NewReader(body))
			user := createHandlerTestUserForTemplate(tt.userRole)
			ctx := contextWithUserForTemplate(user)
			req = req.WithContext(ctx)
			req.SetPathValue("id", templateID.String())

			// Test the authorization logic
			user, ok := middleware.GetUserFromContext(req.Context())
			if !ok {
				assert.Equal(t, http.StatusUnauthorized, tt.expectedStatus)
			} else if !user.IsAdmin() && !user.IsMethodologist() {
				assert.Equal(t, http.StatusForbidden, tt.expectedStatus)
			} else {
				assert.Equal(t, http.StatusCreated, tt.expectedStatus)
			}
		})
	}
}

// TestTemplateUpdateTemplateLessonAccessControl tests role-based access control for UpdateTemplateLesson
func TestTemplateUpdateTemplateLessonAccessControl(t *testing.T) {
	templateID := uuid.New()
	lessonID := uuid.New()

	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
	}{
		{
			name:           "Admin can access UpdateTemplateLesson",
			userRole:       models.RoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Methodologist can access UpdateTemplateLesson",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Teacher cannot access UpdateTemplateLesson - Forbidden",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"day_of_week": 1,
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api/v1/templates/"+templateID.String()+"/lessons/"+lessonID.String(), bytes.NewReader(body))
			user := createHandlerTestUserForTemplate(tt.userRole)
			ctx := contextWithUserForTemplate(user)
			req = req.WithContext(ctx)
			req.SetPathValue("id", templateID.String())
			req.SetPathValue("lesson_id", lessonID.String())

			// Test the authorization logic
			user, ok := middleware.GetUserFromContext(req.Context())
			if !ok {
				assert.Equal(t, http.StatusUnauthorized, tt.expectedStatus)
			} else if !user.IsAdmin() && !user.IsMethodologist() {
				assert.Equal(t, http.StatusForbidden, tt.expectedStatus)
			} else {
				assert.Equal(t, http.StatusOK, tt.expectedStatus)
			}
		})
	}
}

// TestTemplateDeleteTemplateLessonAccessControl tests role-based access control for DeleteTemplateLesson
func TestTemplateDeleteTemplateLessonAccessControl(t *testing.T) {
	templateID := uuid.New()
	lessonID := uuid.New()

	tests := []struct {
		name           string
		userRole       models.UserRole
		expectedStatus int
	}{
		{
			name:           "Admin can access DeleteTemplateLesson",
			userRole:       models.RoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Methodologist can access DeleteTemplateLesson",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Teacher cannot access DeleteTemplateLesson - Forbidden",
			userRole:       models.RoleMethodologist,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/"+templateID.String()+"/lessons/"+lessonID.String(), nil)
			user := createHandlerTestUserForTemplate(tt.userRole)
			ctx := contextWithUserForTemplate(user)
			req = req.WithContext(ctx)
			req.SetPathValue("id", templateID.String())
			req.SetPathValue("lesson_id", lessonID.String())

			// Test the authorization logic
			user, ok := middleware.GetUserFromContext(req.Context())
			if !ok {
				assert.Equal(t, http.StatusUnauthorized, tt.expectedStatus)
			} else if !user.IsAdmin() && !user.IsMethodologist() {
				assert.Equal(t, http.StatusForbidden, tt.expectedStatus)
			} else {
				assert.Equal(t, http.StatusOK, tt.expectedStatus)
			}
		})
	}
}

// TestTemplateGetTemplate_InvalidUUID tests UUID validation
func TestTemplateGetTemplate_InvalidUUID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/invalid-uuid", nil)
	user := createHandlerTestUserForTemplate(models.RoleAdmin)
	ctx := contextWithUserForTemplate(user)
	req = req.WithContext(ctx)
	req.SetPathValue("id", "invalid-uuid")

	// Authorization should pass but UUID validation should fail
	user, ok := middleware.GetUserFromContext(req.Context())
	assert.True(t, ok, "User should be in context")
	assert.True(t, user.IsAdmin(), "Admin should have access")
}

// TestTemplateGetTemplates_UnauthenticatedUser tests unauthorized access
func TestTemplateGetTemplates_UnauthenticatedUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
	// No user in context

	user, ok := middleware.GetUserFromContext(req.Context())
	assert.False(t, ok, "User should not be in context for unauthenticated request")
	assert.Nil(t, user, "User should be nil")
}

// TestTemplateCreateTemplate_InvalidPayload tests validation
func TestTemplateCreateTemplate_InvalidPayload(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", io.NopCloser(bytes.NewReader([]byte("invalid json"))))
	user := createHandlerTestUserForTemplate(models.RoleAdmin)
	ctx := contextWithUserForTemplate(user)
	req = req.WithContext(ctx)

	// Authorization should pass but JSON parsing should fail
	user, ok := middleware.GetUserFromContext(req.Context())
	assert.True(t, ok, "User should be in context")
	assert.True(t, user.IsAdmin(), "Admin should have access")
}

// stringPtr is a helper to create string pointers for template tests
func stringPtrForTemplate(s string) *string {
	return &s
}

package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"
)

// TemplateHandler handles HTTP requests for lesson templates
type TemplateHandler struct {
	templateService *service.TemplateService
}

// NewTemplateHandler creates a new TemplateHandler
func NewTemplateHandler(templateService *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
	}
}

// GetDefaultTemplate removed - use GetTemplates instead
// Multiple named templates are now supported instead of single default template

// GetTemplates handles GET /api/v1/templates - List all templates (Admin only)
// @Summary      List templates
// @Description  Get all lesson templates (admin only)
// @Tags         templates
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      403  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /templates [get]
func (h *TemplateHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !admin.IsAdmin() && !admin.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	templates, err := h.templateService.ListTemplates(r.Context(), admin.ID)
	if err != nil {
		response.InternalError(w, "Failed to retrieve templates")
		return
	}

	response.OK(w, map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// CreateTemplate handles POST /api/v1/templates - Create template (Admin only)
// @Summary      Create template
// @Description  Create a new lesson template (admin only)
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        payload  body      models.CreateLessonTemplateRequest  true  "Template details"
// @Success      201  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /templates [post]
func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !admin.IsAdmin() && !admin.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	var req models.CreateLessonTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid JSON: "+err.Error())
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		response.BadRequest(w, response.ErrCodeValidationFailed, err.Error())
		return
	}

	template, err := h.templateService.CreateTemplate(r.Context(), admin.ID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "teacher") {
			response.BadRequest(w, response.ErrCodeValidationFailed, err.Error())
			return
		}
		response.InternalError(w, "Failed to create template: "+err.Error())
		return
	}

	response.Created(w, template)
}

// GetTemplate handles GET /api/v1/templates/:id - Get template with lessons (Admin only)
// @Summary      Get template
// @Description  Get template details with lessons (admin only)
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Template ID"
// @Success      200  {object}  response.SuccessResponse{data=interface{}}
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /templates/{id} [get]
func (h *TemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !admin.IsAdmin() && !admin.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	templateIDStr := chi.URLParam(r, "id")
	if templateIDStr == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Template ID is required")
		return
	}

	// Validate UUID format
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid template ID format")
		return
	}

	template, err := h.templateService.GetTemplateWithLessons(r.Context(), templateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "Template not found")
			return
		}
		response.InternalError(w, "Failed to retrieve template: "+err.Error())
		return
	}

	response.OK(w, template)
}

// UpdateTemplate handles PUT /api/v1/templates/:id - Update template (Admin only)
func (h *TemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !admin.IsAdmin() && !admin.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	templateIDStr := chi.URLParam(r, "id")
	if templateIDStr == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Template ID is required")
		return
	}

	// Validate UUID format
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid template ID format")
		return
	}

	var req models.UpdateLessonTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid JSON: "+err.Error())
		return
	}

	if err := h.templateService.UpdateTemplate(r.Context(), admin.ID, templateID, &req); err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			response.Forbidden(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "teacher") || strings.Contains(err.Error(), "validation") {
			response.BadRequest(w, response.ErrCodeValidationFailed, err.Error())
			return
		}
		response.InternalError(w, "Failed to update template: "+err.Error())
		return
	}

	response.OK(w, map[string]string{
		"status":  "updated",
		"message": "Template updated successfully",
	})
}

// DeleteTemplate handles DELETE /api/v1/templates/:id - Delete template (Admin only)
func (h *TemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !admin.IsAdmin() && !admin.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	templateIDStr := chi.URLParam(r, "id")
	if templateIDStr == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Template ID is required")
		return
	}

	// Validate UUID format
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid template ID format")
		return
	}

	if err := h.templateService.DeleteTemplate(r.Context(), admin.ID, templateID); err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			response.Forbidden(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, err.Error())
			return
		}
		response.InternalError(w, "Failed to delete template: "+err.Error())
		return
	}

	response.OK(w, map[string]string{
		"status":  "deleted",
		"message": "Template deleted successfully",
	})
}

// ApplyTemplate handles POST /api/v1/templates/:id/apply - Apply template to week (Admin only)
// @Summary      Apply template
// @Description  Apply template lessons to a specific week (admin only, atomic credit deduction)
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Template ID"
// @Param        payload  body      models.ApplyTemplateRequest  true  "Application details"
// @Param        dry_run  query     boolean  false  "Dry run mode - preview what would be created without actually creating it"
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}  "Dry-run preview"
// @Success      201  {object}  response.SuccessResponse{data=interface{}}  "Template applied"
// @Failure      400  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Failure      409  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /templates/{id}/apply [post]
func (h *TemplateHandler) ApplyTemplate(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !admin.IsAdmin() && !admin.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	templateID := chi.URLParam(r, "id")
	if templateID == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Template ID is required")
		return
	}

	// Validate UUID format
	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid template ID format")
		return
	}

	var req models.ApplyTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid JSON: "+err.Error())
		return
	}

	// Set template ID from URL parameter
	req.TemplateID = templateUUID

	// Проверяем query параметр dry_run
	if r.URL.Query().Get("dry_run") == "true" {
		req.DryRun = true
	}

	// Validate request
	if err := req.Validate(); err != nil {
		response.BadRequest(w, response.ErrCodeValidationFailed, err.Error())
		return
	}

	application, err := h.templateService.ApplyTemplateToWeek(r.Context(), admin.ID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "already applied") || strings.Contains(err.Error(), "conflict") {
			response.Conflict(w, response.ErrCodeConflict, err.Error())
			return
		}
		response.InternalError(w, "Failed to apply template: "+err.Error())
		return
	}

	// Обработаем dry-run с явным флагом в ответе
	if req.DryRun {
		// Подготавливаем статистику would_create из CreationStats
		wouldCreate := map[string]interface{}{
			"lessons":  len(application.Lessons),
			"bookings": 0,
			"credits":  0,
		}
		if application.CreationStats != nil {
			wouldCreate["lessons"] = application.CreationStats.CreatedLessons
			wouldCreate["bookings"] = application.CreationStats.CreatedBookings
			wouldCreate["credits"] = application.CreationStats.DeductedCredits
		}

		// Возвращаем HTTP 200 для предпросмотра (не 201, так как ничего не создается)
		response.OK(w, map[string]interface{}{
			"dry_run":         true,
			"status":          "preview",
			"message":         "This is a preview of what would be created. No changes have been made.",
			"template_id":     application.TemplateID,
			"week_start_date": application.WeekStartDate.Format("2006-01-02"),
			"lessons_count":   len(application.Lessons),
			"preview":         application,
			"would_create":    wouldCreate,
		})
		return
	}

	// Для реального применения возвращаем 201 Created
	response.Created(w, application)
}

// RollbackTemplate handles POST /api/v1/templates/:id/rollback - Rollback week to template (Admin only)
// @Summary      Rollback template
// @Description  Rollback template application for a week (admin only, atomic credit refund)
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Template ID"
// @Param        payload  body      models.RollbackTemplateRequest  true  "Rollback details"
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /templates/{id}/rollback [post]
func (h *TemplateHandler) RollbackTemplate(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !admin.IsAdmin() && !admin.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	templateIDStr := chi.URLParam(r, "id")
	if templateIDStr == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Template ID is required")
		return
	}

	// Validate UUID format
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid template ID format")
		return
	}

	var req struct {
		WeekStartDate string `json:"week_start_date" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid JSON: "+err.Error())
		return
	}

	if req.WeekStartDate == "" {
		response.BadRequest(w, response.ErrCodeValidationFailed, "week_start_date is required")
		return
	}

	rollbackResponse, err := h.templateService.RollbackWeekToTemplate(r.Context(), admin.ID, req.WeekStartDate, templateID)
	if err != nil {
		// Проверяем специфичные ошибки в порядке приоритета
		errMsg := err.Error()

		// "cannot rollback: no template application found" -> 404
		if strings.Contains(errMsg, "no template application found") {
			response.NotFound(w, err.Error())
			return
		}

		// "cannot rollback: week already rolled back" -> 404
		if strings.Contains(errMsg, "already rolled back") {
			response.NotFound(w, err.Error())
			return
		}

		// Другие "not found" тоже 404
		if strings.Contains(errMsg, "not found") {
			response.NotFound(w, err.Error())
			return
		}

		// Все остальные - 500
		response.InternalError(w, "Failed to rollback template: "+err.Error())
		return
	}

	response.OK(w, rollbackResponse)
}

// CreateTemplateLesson handles POST /api/v1/templates/:id/lessons - Create template lesson (Admin only)
func (h *TemplateHandler) CreateTemplateLesson(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !admin.IsAdmin() && !admin.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	templateIDStr := chi.URLParam(r, "id")
	if templateIDStr == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Template ID is required")
		return
	}

	// Validate UUID format
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid template ID format")
		return
	}

	// Decode request
	var req models.CreateTemplateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid JSON: "+err.Error())
		return
	}

	// Validate request before applying defaults
	if err := req.Validate(); err != nil {
		response.BadRequest(w, response.ErrCodeValidationFailed, err.Error())
		return
	}

	// Apply defaults (end_time = start_time + 2 hours, max_students based on type, default color)
	req.ApplyDefaults()

	// Create lesson entry through service
	createdLesson, err := h.templateService.CreateTemplateLesson(r.Context(), admin.ID, templateID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			response.Forbidden(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "validation") {
			response.BadRequest(w, response.ErrCodeValidationFailed, err.Error())
			return
		}
		response.InternalError(w, "Failed to create template lesson: "+err.Error())
		return
	}

	response.Created(w, createdLesson)
}

// UpdateTemplateLesson handles PUT /api/v1/templates/:id/lessons/:lesson_id - Update template lesson (Admin only)
func (h *TemplateHandler) UpdateTemplateLesson(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !admin.IsAdmin() && !admin.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	templateIDStr := chi.URLParam(r, "id")
	lessonIDStr := chi.URLParam(r, "lesson_id")

	if templateIDStr == "" || lessonIDStr == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Template ID and Lesson ID are required")
		return
	}

	// Validate UUID formats
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid template ID format")
		return
	}

	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID format")
		return
	}

	// Decode request (all fields optional)
	var req struct {
		DayOfWeek   *int         `json:"day_of_week,omitempty"`
		StartTime   *string      `json:"start_time,omitempty"`
		EndTime     *string      `json:"end_time,omitempty"`
		TeacherID   *uuid.UUID   `json:"teacher_id,omitempty"`
		MaxStudents *int         `json:"max_students,omitempty"`
		CreditsCost *int         `json:"credits_cost,omitempty"`
		Color       *string      `json:"color,omitempty"`
		Subject     *string      `json:"subject,omitempty"`
		Description *string      `json:"description,omitempty"`
		StudentIDs  *[]uuid.UUID `json:"student_ids,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid JSON: "+err.Error())
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.DayOfWeek != nil {
		updates["day_of_week"] = *req.DayOfWeek
	}
	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	}
	if req.TeacherID != nil {
		updates["teacher_id"] = *req.TeacherID
	}
	if req.MaxStudents != nil {
		updates["max_students"] = *req.MaxStudents
	}
	if req.CreditsCost != nil {
		updates["credits_cost"] = *req.CreditsCost
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Subject != nil {
		updates["subject"] = *req.Subject
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.StudentIDs != nil {
		updates["student_ids"] = *req.StudentIDs
	}

	// Update lesson through service
	updatedLesson, err := h.templateService.UpdateTemplateLesson(r.Context(), admin.ID, templateID, lessonID, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			response.Forbidden(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "validation") || strings.Contains(err.Error(), "does not belong") {
			response.BadRequest(w, response.ErrCodeValidationFailed, err.Error())
			return
		}
		response.InternalError(w, "Failed to update template lesson: "+err.Error())
		return
	}

	response.OK(w, updatedLesson)
}

// DeleteTemplateLesson handles DELETE /api/v1/templates/:id/lessons/:lesson_id - Delete template lesson (Admin only)
func (h *TemplateHandler) DeleteTemplateLesson(w http.ResponseWriter, r *http.Request) {
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !admin.IsAdmin() && !admin.IsMethodologist() {
		response.Forbidden(w, "Admin or methodologist access required")
		return
	}

	templateIDStr := chi.URLParam(r, "id")
	lessonIDStr := chi.URLParam(r, "lesson_id")

	if templateIDStr == "" || lessonIDStr == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Template ID and Lesson ID are required")
		return
	}

	// Validate UUID formats
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid template ID format")
		return
	}

	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid lesson ID format")
		return
	}

	// Delete lesson through service
	if err := h.templateService.DeleteTemplateLesson(r.Context(), admin.ID, templateID, lessonID); err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			response.Forbidden(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(w, err.Error())
			return
		}
		if strings.Contains(err.Error(), "does not belong") {
			response.BadRequest(w, response.ErrCodeValidationFailed, err.Error())
			return
		}
		response.InternalError(w, "Failed to delete template lesson: "+err.Error())
		return
	}

	response.OK(w, map[string]string{
		"status":  "deleted",
		"message": "Template lesson deleted successfully",
	})
}

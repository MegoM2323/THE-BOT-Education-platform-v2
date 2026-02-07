package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/pkg/response"
)

// SubjectsHandler обрабатывает эндпоинты предметов
type SubjectsHandler struct {
	subjectRepo repository.SubjectRepository
}

// NewSubjectsHandler создает новый SubjectsHandler
func NewSubjectsHandler(subjectRepo repository.SubjectRepository) *SubjectsHandler {
	return &SubjectsHandler{
		subjectRepo: subjectRepo,
	}
}

// GetSubjects обрабатывает GET /api/v1/subjects
// @Summary      List all subjects
// @Description  Get list of all available subjects
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse{data=[]models.Subject}
// @Failure      500  {object}  response.ErrorResponse
// @Router       /subjects [get]
func (h *SubjectsHandler) GetSubjects(w http.ResponseWriter, r *http.Request) {
	subjects, err := h.subjectRepo.List(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to retrieve subjects")
		return
	}

	response.OK(w, map[string]interface{}{
		"subjects": subjects,
	})
}

// GetSubject обрабатывает GET /api/v1/subjects/:id
// @Summary      Get subject by ID
// @Description  Get detailed information about a specific subject
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Param        id   path     string  true  "Subject ID"
// @Success      200  {object}  response.SuccessResponse{data=models.Subject}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /subjects/{id} [get]
func (h *SubjectsHandler) GetSubject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid subject ID")
		return
	}

	subject, err := h.subjectRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrSubjectNotFound) {
			response.NotFound(w, "Subject not found")
			return
		}
		response.InternalError(w, "Failed to retrieve subject")
		return
	}

	response.OK(w, subject)
}

// CreateSubject обрабатывает POST /api/v1/subjects (admin only)
// @Summary      Create subject
// @Description  Create a new subject (admin only)
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreateSubjectRequest  true  "Subject data"
// @Success      201  {object}  response.SuccessResponse{data=models.Subject}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /subjects [post]
func (h *SubjectsHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	var req models.CreateSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, err.Error())
		return
	}

	subject := &models.Subject{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.subjectRepo.Create(r.Context(), subject); err != nil {
		response.InternalError(w, "Failed to create subject")
		return
	}

	w.WriteHeader(http.StatusCreated)
	response.OK(w, subject)
}

// UpdateSubject обрабатывает PUT /api/v1/subjects/:id (admin only)
// @Summary      Update subject
// @Description  Update subject information (admin only)
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Param        id    path      string                         true  "Subject ID"
// @Param        body  body      models.UpdateSubjectRequest    true  "Updated subject data"
// @Success      200   {object}  response.SuccessResponse{data=models.Subject}
// @Failure      400   {object}  response.ErrorResponse
// @Failure      403   {object}  response.ErrorResponse
// @Failure      404   {object}  response.ErrorResponse
// @Failure      500   {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /subjects/{id} [put]
func (h *SubjectsHandler) UpdateSubject(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid subject ID")
		return
	}

	// Проверяем, что предмет существует
	subject, err := h.subjectRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrSubjectNotFound) {
			response.NotFound(w, "Subject not found")
			return
		}
		response.InternalError(w, "Failed to retrieve subject")
		return
	}

	var req models.UpdateSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, err.Error())
		return
	}

	// Построение обновлений
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
		subject.Name = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
		subject.Description = *req.Description
	}

	if err := h.subjectRepo.Update(r.Context(), id, updates); err != nil {
		response.InternalError(w, "Failed to update subject")
		return
	}

	response.OK(w, subject)
}

// DeleteSubject обрабатывает DELETE /api/v1/subjects/:id (admin only)
// @Summary      Delete subject
// @Description  Delete a subject (soft delete, admin only)
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Param        id   path     string  true  "Subject ID"
// @Success      204
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /subjects/{id} [delete]
func (h *SubjectsHandler) DeleteSubject(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid subject ID")
		return
	}

	// Проверяем, что предмет существует
	_, err = h.subjectRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrSubjectNotFound) {
			response.NotFound(w, "Subject not found")
			return
		}
		response.InternalError(w, "Failed to retrieve subject")
		return
	}

	if err := h.subjectRepo.Delete(r.Context(), id); err != nil {
		response.InternalError(w, "Failed to delete subject")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTeacherSubjects обрабатывает GET /api/v1/teachers/:id/subjects
// @Summary      Get teacher subjects
// @Description  Get list of subjects taught by a specific teacher
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Param        id   path     string  true  "Teacher ID"
// @Success      200  {object}  response.SuccessResponse{data=[]models.TeacherSubjectWithDetails}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /teachers/{id}/subjects [get]
func (h *SubjectsHandler) GetTeacherSubjects(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid teacher ID")
		return
	}

	subjects, err := h.subjectRepo.GetTeacherSubjects(r.Context(), id)
	if err != nil {
		response.InternalError(w, "Failed to retrieve teacher subjects")
		return
	}

	response.OK(w, map[string]interface{}{
		"subjects": subjects,
	})
}

// AssignSubjectToTeacher обрабатывает POST /api/v1/teachers/:id/subjects (admin only)
// @Summary      Assign subject to teacher
// @Description  Assign a subject to a teacher (admin only)
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Param        id    path      string                             true  "Teacher ID"
// @Param        body  body      models.AssignSubjectRequest        true  "Subject assignment data"
// @Success      201   {object}  response.SuccessResponse{data=models.TeacherSubject}
// @Failure      400   {object}  response.ErrorResponse
// @Failure      403   {object}  response.ErrorResponse
// @Failure      500   {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /teachers/{id}/subjects [post]
func (h *SubjectsHandler) AssignSubjectToTeacher(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	idStr := chi.URLParam(r, "id")
	teacherID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid teacher ID")
		return
	}

	var req models.AssignSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, err.Error())
		return
	}

	if err := h.subjectRepo.AssignToTeacher(r.Context(), teacherID, req.SubjectID); err != nil {
		// Проверяем на конфликты
		if repository.IsUniqueViolationError(err) {
			response.BadRequest(w, response.ErrCodeConflict, "Subject already assigned to this teacher")
			return
		}
		response.InternalError(w, "Failed to assign subject")
		return
	}

	w.WriteHeader(http.StatusCreated)
	response.OK(w, map[string]interface{}{
		"message": "Subject assigned successfully",
	})
}

// RemoveSubjectFromTeacher обрабатывает DELETE /api/v1/teachers/:id/subjects/:subjectId (admin only)
// @Summary      Remove subject from teacher
// @Description  Remove a subject from a teacher (admin only)
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Param        id         path     string  true  "Teacher ID"
// @Param        subjectId  path     string  true  "Subject ID"
// @Success      204
// @Failure      403  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /teachers/{id}/subjects/{subjectId} [delete]
func (h *SubjectsHandler) RemoveSubjectFromTeacher(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	idStr := chi.URLParam(r, "id")
	teacherID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid teacher ID")
		return
	}

	subjectIdStr := chi.URLParam(r, "subjectId")
	subjectID, err := uuid.Parse(subjectIdStr)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid subject ID")
		return
	}

	if err := h.subjectRepo.RemoveFromTeacher(r.Context(), teacherID, subjectID); err != nil {
		response.InternalError(w, "Failed to remove subject")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetMySubjects обрабатывает GET /api/v1/my-subjects
// @Summary      Get current user's subjects
// @Description  Get list of subjects for the currently logged-in teacher
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.SuccessResponse{data=[]models.TeacherSubjectWithDetails}
// @Failure      403  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /my-subjects [get]
func (h *SubjectsHandler) GetMySubjects(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsTeacher() {
		response.Forbidden(w, "Teacher access required")
		return
	}

	subjects, err := h.subjectRepo.GetTeacherSubjects(r.Context(), user.ID)
	if err != nil {
		response.InternalError(w, "Failed to retrieve your subjects")
		return
	}

	response.OK(w, map[string]interface{}{
		"subjects": subjects,
	})
}

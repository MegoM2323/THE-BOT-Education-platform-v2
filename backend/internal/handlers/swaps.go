package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/internal/validator"
	"tutoring-platform/pkg/response"
)

// SwapHandler обрабатывает эндпоинты обменов
type SwapHandler struct {
	swapService *service.SwapService
}

// NewSwapHandler создает новый SwapHandler
func NewSwapHandler(swapService *service.SwapService) *SwapHandler {
	return &SwapHandler{
		swapService: swapService,
	}
}

// PerformSwap обрабатывает POST /api/v1/swaps
func (h *SwapHandler) PerformSwap(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только студенты могут выполнять обмены
	if !user.IsStudent() {
		response.Forbidden(w, "Only students can perform lesson swaps")
		return
	}

	var reqBody struct {
		OldLessonID uuid.UUID `json:"old_lesson_id"`
		NewLessonID uuid.UUID `json:"new_lesson_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Создаем запрос на обмен
	req := &models.PerformSwapRequest{
		StudentID:   user.ID,
		OldLessonID: reqBody.OldLessonID,
		NewLessonID: reqBody.NewLessonID,
	}

	// Выполняем обмен
	swap, err := h.swapService.PerformSwap(r.Context(), req)
	if err != nil {
		h.handleSwapError(w, err)
		return
	}

	response.Created(w, map[string]interface{}{
		"swap":    swap,
		"message": "Lesson swap completed successfully",
	})
}

// ValidateSwap обрабатывает POST /api/v1/swaps/validate
func (h *SwapHandler) ValidateSwap(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	var reqBody struct {
		OldLessonID uuid.UUID `json:"old_lesson_id"`
		NewLessonID uuid.UUID `json:"new_lesson_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Создаем запрос на валидацию
	req := &models.ValidateSwapRequest{
		StudentID:   user.ID,
		OldLessonID: reqBody.OldLessonID,
		NewLessonID: reqBody.NewLessonID,
	}

	// Валидируем обмен
	validationResponse, err := h.swapService.ValidateSwap(r.Context(), req)
	if err != nil {
		response.InternalError(w, "Failed to validate swap")
		return
	}

	if !validationResponse.Valid {
		response.OK(w, map[string]interface{}{
			"valid":  false,
			"errors": validationResponse.Errors,
		})
		return
	}

	response.OK(w, map[string]interface{}{
		"valid":      true,
		"old_lesson": validationResponse.Details.OldLesson,
		"new_lesson": validationResponse.Details.NewLesson,
	})
}

// GetSwapHistory обрабатывает GET /api/v1/swaps/history
func (h *SwapHandler) GetSwapHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	filter := &models.GetSwapHistoryFilter{}

	// Студенты видят только свою историю обменов
	if user.IsStudent() {
		filter.StudentID = &user.ID
	} else if user.IsAdmin() {
		// Админы могут фильтровать по student_id, если указан
		if studentIDStr := r.URL.Query().Get("student_id"); studentIDStr != "" {
			studentID, err := uuid.Parse(studentIDStr)
			if err == nil {
				filter.StudentID = &studentID
			}
		}
	}
	// Преподаватели в настоящее время не имеют доступа к истории обменов (не реализовано в модели)

	// Парсим опциональные фильтры по дате
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			// Добавляем 1 день, чтобы включить конечную дату
			endDate = endDate.Add(24 * time.Hour)
			filter.EndDate = &endDate
		}
	}

	swaps, err := h.swapService.GetSwapHistory(r.Context(), filter)
	if err != nil {
		response.InternalError(w, "Failed to retrieve swap history")
		return
	}

	response.OK(w, map[string]interface{}{
		"swaps": swaps,
		"count": len(swaps),
	})
}

// handleSwapError обрабатывает специфичные для обменов ошибки
func (h *SwapHandler) handleSwapError(w http.ResponseWriter, err error) {
	// Проверяем ошибки models (валидация запроса)
	if errors.Is(err, models.ErrInvalidStudentID) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid student ID")
		return
	}
	if errors.Is(err, models.ErrInvalidOldLessonID) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid old lesson ID")
		return
	}
	if errors.Is(err, models.ErrInvalidNewLessonID) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid new lesson ID")
		return
	}
	if errors.Is(err, models.ErrSameLessonSwap) {
		response.BadRequest(w, response.ErrCodeInvalidSwap, "Cannot swap to the same lesson")
		return
	}

	// Проверяем ошибки repository
	if errors.Is(err, repository.ErrLessonNotFound) {
		response.NotFound(w, "Lesson not found")
		return
	}
	if errors.Is(err, repository.ErrLessonFull) {
		response.Conflict(w, response.ErrCodeLessonFull, "The new lesson is already full")
		return
	}

	// Проверяем ошибки validator
	if errors.Is(err, validator.ErrNoActiveBooking) {
		response.Conflict(w, response.ErrCodeInvalidSwap, "No active booking found for the old lesson")
		return
	}
	if errors.Is(err, validator.ErrNewLessonNotAvailable) {
		response.Conflict(w, response.ErrCodeLessonFull, "The new lesson is not available for booking")
		return
	}
	if errors.Is(err, validator.ErrSwapScheduleConflict) {
		response.Conflict(w, response.ErrCodeScheduleConflict, "You have a conflicting booking at the time of the new lesson")
		return
	}
	if errors.Is(err, validator.ErrSwapTooLate) {
		response.Conflict(w, response.ErrCodeBookingTooLate, "Swaps must be made at least 24 hours before both lessons")
		return
	}
	if errors.Is(err, validator.ErrSwapLessonInPast) {
		response.Conflict(w, response.ErrCodeInvalidSwap, "Cannot swap to a lesson in the past")
		return
	}

	// Неизвестная ошибка
	response.InternalError(w, "An error occurred processing your swap request")
}

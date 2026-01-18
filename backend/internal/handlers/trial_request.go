package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"log"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"
)

// TrialRequestHandler обрабатывает эндпоинты пробных запросов
type TrialRequestHandler struct {
	service *service.TrialRequestService
}

// NewTrialRequestHandler создает новый TrialRequestHandler
func NewTrialRequestHandler(service *service.TrialRequestService) *TrialRequestHandler {
	return &TrialRequestHandler{
		service: service,
	}
}

// CreateTrialRequest обрабатывает POST /api/v1/landing/trial-request
// Это публичный эндпоинт (аутентификация не требуется)
func (h *TrialRequestHandler) CreateTrialRequest(w http.ResponseWriter, r *http.Request) {
	var input models.CreateTrialRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Создаем пробный запрос
	trialRequest, err := h.service.CreateTrialRequest(r.Context(), &input)
	if err != nil {
		// Проверяем на ошибки валидации models
		if errors.Is(err, models.ErrInvalidName) {
			response.BadRequest(w, response.ErrCodeValidationFailed, models.ErrInvalidName.Error())
			return
		}
		if errors.Is(err, models.ErrInvalidPhone) {
			response.BadRequest(w, response.ErrCodeValidationFailed, models.ErrInvalidPhone.Error())
			return
		}
		if errors.Is(err, models.ErrInvalidTelegram) {
			response.BadRequest(w, response.ErrCodeValidationFailed, models.ErrInvalidTelegram.Error())
			return
		}
		if errors.Is(err, models.ErrInvalidEmailFormat) {
			response.BadRequest(w, response.ErrCodeValidationFailed, models.ErrInvalidEmailFormat.Error())
			return
		}

		// Неизвестная ошибка
		log.Printf("ERROR: Failed to create trial request: %v", err)
		response.InternalError(w, "Failed to create trial request")
		return
	}

	response.Created(w, map[string]interface{}{
		"trial_request": trialRequest,
		"message":       "Trial request submitted successfully",
	})
}

// GetTrialRequests обрабатывает GET /api/v1/trial-requests
// Это эндпоинт только для админов
func (h *TrialRequestHandler) GetTrialRequests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.service.GetAllTrialRequests(r.Context())
	if err != nil {
		log.Printf("ERROR: Failed to retrieve trial requests: %v", err)
		response.InternalError(w, "Failed to retrieve trial requests")
		return
	}

	response.OK(w, map[string]interface{}{
		"trial_requests": requests,
		"count":          len(requests),
	})
}

// GetAllTrialRequests обрабатывает GET /api/v1/admin/landing/trial-requests
// Это эндпоинт только для админов (старый алиас для совместимости)
func (h *TrialRequestHandler) GetAllTrialRequests(w http.ResponseWriter, r *http.Request) {
	h.GetTrialRequests(w, r)
}

// GetTrialRequestByID обрабатывает GET /api/v1/admin/landing/trial-requests/:id
// Это эндпоинт только для админов
func (h *TrialRequestHandler) GetTrialRequestByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid trial request ID")
		return
	}

	trialRequest, err := h.service.GetTrialRequestByID(r.Context(), id)
	if err != nil {
		log.Printf("ERROR: Failed to get trial request by ID %d: %v", id, err)
		response.NotFound(w, "Trial request not found")
		return
	}

	response.OK(w, map[string]interface{}{
		"trial_request": trialRequest,
	})
}

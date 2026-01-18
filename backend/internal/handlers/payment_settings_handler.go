package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/pkg/response"
)

// PaymentSettingsServiceInterface определяет интерфейс для работы с настройками платежей
type PaymentSettingsServiceInterface interface {
	GetPaymentStatus(ctx context.Context, userID uuid.UUID) (bool, error)
	UpdatePaymentStatus(ctx context.Context, adminID, userID uuid.UUID, enabled bool) (*models.User, error)
	ListStudentsPaymentStatus(ctx context.Context, adminID uuid.UUID, filterByEnabled *bool) ([]*models.StudentPaymentStatus, error)
}

// PaymentSettingsHandler обрабатывает эндпоинты управления платежами студентов
type PaymentSettingsHandler struct {
	paymentSettingsService PaymentSettingsServiceInterface
}

// NewPaymentSettingsHandler создает новый PaymentSettingsHandler
func NewPaymentSettingsHandler(paymentSettingsService PaymentSettingsServiceInterface) *PaymentSettingsHandler {
	return &PaymentSettingsHandler{
		paymentSettingsService: paymentSettingsService,
	}
}

// ListStudentsPaymentStatus обрабатывает GET /api/v1/admin/payment-settings
// @Summary      List students payment status
// @Description  Get list of all students with payment settings (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        payment_enabled  query  string  false  "Filter by payment status (true/false)"
// @Param        limit            query  int     false  "Limit results (default 100)"
// @Param        offset           query  int     false  "Offset results (default 0)"
// @Success      200  {object}  response.SuccessResponse{data=[]models.StudentPaymentStatus}
// @Failure      403  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /admin/payment-settings [get]
func (h *PaymentSettingsHandler) ListStudentsPaymentStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только админы могут получать список настроек платежей
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	// Опциональный фильтр по payment_enabled
	var filterByEnabled *bool
	paymentEnabledStr := r.URL.Query().Get("payment_enabled")
	if paymentEnabledStr != "" {
		if paymentEnabledStr == "true" {
			enabled := true
			filterByEnabled = &enabled
		} else if paymentEnabledStr == "false" {
			disabled := false
			filterByEnabled = &disabled
		}
	}

	// Парсим limit и offset (для будущей пагинации)
	limit := 100
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Получаем список статусов платежей студентов
	students, err := h.paymentSettingsService.ListStudentsPaymentStatus(r.Context(), user.ID, filterByEnabled)
	if err != nil {
		h.handlePaymentSettingsError(w, err)
		return
	}

	// Применяем пагинацию в памяти (для простоты, т.к. количество студентов не критично)
	start := offset
	end := offset + limit
	if start > len(students) {
		start = len(students)
	}
	if end > len(students) {
		end = len(students)
	}

	response.OK(w, map[string]interface{}{
		"students": students[start:end],
		"total":    len(students),
		"count":    end - start,
	})
}

// UpdatePaymentStatus обрабатывает PUT /api/v1/admin/users/:id/payment-settings
// @Summary      Update student payment status
// @Description  Enable or disable payments for a student (admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id       path      string                                true  "User ID"
// @Param        payload  body      UpdatePaymentStatusRequest            true  "Payment settings"
// @Success      200      {object}  response.SuccessResponse{data=models.User}
// @Failure      400      {object}  response.ErrorResponse
// @Failure      403      {object}  response.ErrorResponse
// @Failure      404      {object}  response.ErrorResponse
// @Failure      500      {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /admin/users/{id}/payment-settings [put]
func (h *PaymentSettingsHandler) UpdatePaymentStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только админы могут изменять настройки платежей
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid user ID")
		return
	}

	var req UpdatePaymentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Валидация: payment_enabled должен быть явно указан
	if req.PaymentEnabled == nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "payment_enabled field is required")
		return
	}

	// Обновляем статус платежей
	updatedUser, err := h.paymentSettingsService.UpdatePaymentStatus(r.Context(), user.ID, userID, *req.PaymentEnabled)
	if err != nil {
		h.handlePaymentSettingsError(w, err)
		return
	}

	response.OK(w, map[string]interface{}{
		"user": updatedUser,
	})
}

// UpdatePaymentStatusRequest представляет запрос на обновление статуса платежей
type UpdatePaymentStatusRequest struct {
	PaymentEnabled *bool `json:"payment_enabled"`
}

// handlePaymentSettingsError обрабатывает специфичные для payment settings ошибки
func (h *PaymentSettingsHandler) handlePaymentSettingsError(w http.ResponseWriter, err error) {
	// Проверяем ошибки авторизации
	if err == repository.ErrUnauthorized {
		response.Forbidden(w, "Admin access required")
		return
	}

	// Проверяем ошибки repository
	if err == repository.ErrUserNotFound {
		response.NotFound(w, "User not found")
		return
	}

	// Проверяем ошибку неверной роли
	if err == repository.ErrInvalidUserRole {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Payment settings can only be changed for students")
		return
	}

	// Неизвестная ошибка - логируем полную информацию для отладки
	// В production окружении подробная информация не передается клиенту
	fmt.Printf("[ERROR] handlePaymentSettingsError: %v\n", err)
	response.InternalError(w, "An error occurred processing your request")
}

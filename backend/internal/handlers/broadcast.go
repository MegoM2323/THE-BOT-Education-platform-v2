package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"log"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"
)

// BroadcastHandler обрабатывает админские эндпоинты для массовых рассылок
type BroadcastHandler struct {
	broadcastService *service.BroadcastService
	telegramService  *service.TelegramService
}

// NewBroadcastHandler создает новый BroadcastHandler
func NewBroadcastHandler(broadcastService *service.BroadcastService, telegramService *service.TelegramService) *BroadcastHandler {
	return &BroadcastHandler{
		broadcastService: broadcastService,
		telegramService:  telegramService,
	}
}

// GetLinkedUsers обрабатывает GET /api/v1/admin/telegram/users
// Возвращает список пользователей с привязанным Telegram (только для админов)
func (h *BroadcastHandler) GetLinkedUsers(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	// Проверяем что Telegram сервис настроен
	if h.telegramService == nil {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Telegram bot is not configured. Please set TELEGRAM_BOT_TOKEN in environment variables.")
		return
	}

	// Опциональный фильтр по роли
	roleFilter := r.URL.Query().Get("role")

	// Получаем список пользователей с привязанным Telegram
	telegramUsers, err := h.telegramService.GetLinkedUsers(r.Context(), roleFilter)
	if err != nil {
		log.Printf("[ERROR] GetLinkedUsers failed: %v", err)
		response.InternalError(w, "Failed to retrieve linked users")
		return
	}

	// Трансформируем ответ в формат, ожидаемый frontend
	// Frontend ожидает User объекты с вложенным полем telegram
	type TelegramInfo struct {
		Username   string `json:"username"`
		TelegramID int64  `json:"telegram_id"`
		LinkedAt   string `json:"linked_at"`
	}

	type UserWithTelegram struct {
		ID       string        `json:"id"`
		Email    string        `json:"email"`
		FullName string        `json:"full_name"`
		Role     string        `json:"role"`
		Telegram *TelegramInfo `json:"telegram"`
	}

	var responseUsers []UserWithTelegram
	for _, tu := range telegramUsers {
		if tu.User == nil {
			continue
		}
		responseUsers = append(responseUsers, UserWithTelegram{
			ID:       tu.User.ID.String(),
			Email:    tu.User.Email,
			FullName: tu.User.GetFullName(),
			Role:     string(tu.User.Role),
			Telegram: &TelegramInfo{
				Username:   tu.Username,
				TelegramID: tu.TelegramID,
				LinkedAt:   tu.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			},
		})
	}

	response.OK(w, map[string]interface{}{
		"users": responseUsers,
		"count": len(responseUsers),
	})
}

// CreateBroadcastList обрабатывает POST /api/v1/admin/telegram/lists
// Создает новый список рассылки
func (h *BroadcastHandler) CreateBroadcastList(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	var req models.CreateBroadcastListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Валидация запроса
	if err := req.Validate(); err != nil {
		h.handleBroadcastError(w, err)
		return
	}

	// Создаем список рассылки
	list, err := h.broadcastService.CreateBroadcastList(
		r.Context(),
		req.Name,
		req.Description,
		req.UserIDs,
		user.ID,
	)
	if err != nil {
		log.Printf("ERROR: Failed to create broadcast list: %v", err)
		h.handleBroadcastError(w, err)
		return
	}

	response.Created(w, map[string]interface{}{
		"list": list,
	})
}

// GetBroadcastLists обрабатывает GET /api/v1/admin/telegram/lists
// Возвращает все списки рассылки
func (h *BroadcastHandler) GetBroadcastLists(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	// Получаем все списки рассылки
	lists, err := h.broadcastService.GetBroadcastLists(r.Context())
	if err != nil {
		log.Printf("ERROR: Failed to retrieve broadcast lists: %v", err)
		response.InternalError(w, "Failed to retrieve broadcast lists")
		return
	}

	response.OK(w, map[string]interface{}{
		"lists": lists,
		"count": len(lists),
	})
}

// GetBroadcastListByID обрабатывает GET /api/v1/admin/telegram/lists/{id}
// Возвращает конкретный список рассылки по ID
func (h *BroadcastHandler) GetBroadcastListByID(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	// Парсим ID из URL
	listID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid list ID")
		return
	}

	// Получаем список рассылки
	list, err := h.broadcastService.GetBroadcastListByID(r.Context(), listID)
	if err != nil {
		log.Printf("ERROR: Failed to get broadcast list by ID %s: %v", listID, err)
		h.handleBroadcastError(w, err)
		return
	}

	response.OK(w, map[string]interface{}{
		"list": list,
	})
}

// UpdateBroadcastList обрабатывает PUT /api/v1/admin/telegram/lists/{id}
// Обновляет список рассылки
func (h *BroadcastHandler) UpdateBroadcastList(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	// Парсим ID из URL
	listID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid list ID")
		return
	}

	var req models.UpdateBroadcastListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Валидация запроса
	if err := req.Validate(); err != nil {
		h.handleBroadcastError(w, err)
		return
	}

	// Обновляем список рассылки
	err = h.broadcastService.UpdateBroadcastList(
		r.Context(),
		listID,
		req.Name,
		req.Description,
		req.UserIDs,
	)
	if err != nil {
		log.Printf("ERROR: Failed to update broadcast list %s: %v", listID, err)
		h.handleBroadcastError(w, err)
		return
	}

	// Получаем обновленный список
	updatedList, err := h.broadcastService.GetBroadcastListByID(r.Context(), listID)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve updated broadcast list %s: %v", listID, err)
		h.handleBroadcastError(w, err)
		return
	}

	response.OK(w, map[string]interface{}{
		"list": updatedList,
	})
}

// DeleteBroadcastList обрабатывает DELETE /api/v1/admin/telegram/lists/{id}
// Удаляет список рассылки (мягкое удаление)
func (h *BroadcastHandler) DeleteBroadcastList(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	// Парсим ID из URL
	listID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid list ID")
		return
	}

	// Удаляем список рассылки
	if err := h.broadcastService.DeleteBroadcastList(r.Context(), listID); err != nil {
		log.Printf("ERROR: Failed to delete broadcast list %s: %v", listID, err)
		h.handleBroadcastError(w, err)
		return
	}

	response.OK(w, map[string]string{
		"message": "Broadcast list deleted successfully",
	})
}

// SendBroadcast обрабатывает POST /api/v1/admin/telegram/broadcast
// Создает и отправляет массовую рассылку
func (h *BroadcastHandler) SendBroadcast(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	// Check if Telegram bot is configured
	if h.telegramService == nil {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Telegram bot is not configured. Please set TELEGRAM_BOT_TOKEN in environment variables.")
		return
	}

	var req models.SendBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Валидация запроса
	if err := req.Validate(); err != nil {
		h.handleBroadcastError(w, err)
		return
	}

	// Создаем broadcast в зависимости от того, указан list_id или user_ids
	var broadcast *models.Broadcast
	var err error

	if req.ListID != nil {
		// Используем существующую логику с list_id
		broadcast, err = h.broadcastService.CreateBroadcast(
			r.Context(),
			*req.ListID,
			req.Message,
			user.ID,
		)
		if err != nil {
			log.Printf("ERROR: Failed to create broadcast: %v", err)
			h.handleBroadcastError(w, err)
			return
		}
	} else {
		// Используем новую логику с user_ids
		broadcast, err = h.broadcastService.CreateBroadcastForUsers(
			r.Context(),
			req.UserIDs,
			req.Message,
			user.ID,
		)
		if err != nil {
			log.Printf("ERROR: Failed to create broadcast for users: %v", err)
			h.handleBroadcastError(w, err)
			return
		}
	}

	// Запускаем отправку (проверяем что Telegram настроен синхронно)
	if err := h.broadcastService.SendBroadcast(r.Context(), broadcast.ID); err != nil {
		log.Printf("ERROR: Failed to send broadcast %s: %v", broadcast.ID, err)
		h.handleBroadcastError(w, err)
		return
	}

	response.Created(w, map[string]interface{}{
		"broadcast": broadcast,
		"message":   "Broadcast started successfully",
	})
}

// GetBroadcasts обрабатывает GET /api/v1/admin/telegram/broadcasts
// Возвращает список всех рассылок с пагинацией
func (h *BroadcastHandler) GetBroadcasts(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	// Парсим query параметры для пагинации
	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Получаем список рассылок
	broadcasts, totalCount, err := h.broadcastService.GetBroadcasts(r.Context(), limit, offset)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve broadcasts: %v", err)
		response.InternalError(w, "Failed to retrieve broadcasts")
		return
	}

	response.OK(w, map[string]interface{}{
		"broadcasts": broadcasts,
		"count":      len(broadcasts),
		"total":      totalCount,
		"limit":      limit,
		"offset":     offset,
	})
}

// DeliveryInfo структура для отображения статуса доставки во frontend
type DeliveryInfo struct {
	UserName   string `json:"user_name"`
	TelegramID int64  `json:"telegram_id"`
	Status     string `json:"status"`
	SentAt     string `json:"sent_at,omitempty"`
	Error      string `json:"error,omitempty"`
}

// GetBroadcastDetails обрабатывает GET /api/v1/admin/telegram/broadcasts/{id}
// Возвращает детали конкретной рассылки с логами
func (h *BroadcastHandler) GetBroadcastDetails(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	// Парсим ID из URL
	broadcastID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid broadcast ID")
		return
	}

	// Получаем детали рассылки
	broadcast, logs, err := h.broadcastService.GetBroadcastDetails(r.Context(), broadcastID)
	if err != nil {
		log.Printf("ERROR: Failed to get broadcast details for ID %s: %v", broadcastID, err)
		h.handleBroadcastError(w, err)
		return
	}

	// Получаем список рассылки для total_recipients и list_name
	var totalRecipients int
	var listName string
	if broadcast.ListID != nil {
		list, err := h.broadcastService.GetBroadcastListByID(r.Context(), *broadcast.ListID)
		if err == nil && list != nil {
			totalRecipients = len(list.UserIDs)
			listName = list.Name
		}
	}

	// Преобразуем logs в deliveries для frontend
	deliveries := make([]DeliveryInfo, 0, len(logs))
	for _, logEntry := range logs {
		delivery := DeliveryInfo{
			TelegramID: logEntry.TelegramID,
			Status:     logEntry.Status,
			Error:      logEntry.Error,
		}
		if !logEntry.SentAt.IsZero() {
			delivery.SentAt = logEntry.SentAt.Format("2006-01-02T15:04:05Z07:00")
		}
		deliveries = append(deliveries, delivery)
	}

	// Формируем плоский ответ для frontend
	createdAt := ""
	if !broadcast.CreatedAt.IsZero() {
		createdAt = broadcast.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	response.OK(w, map[string]interface{}{
		"id":               broadcast.ID,
		"list_id":          broadcast.ListID,
		"list_name":        listName,
		"message":          broadcast.Message,
		"status":           broadcast.Status,
		"sent_count":       broadcast.SentCount,
		"failed_count":     broadcast.FailedCount,
		"total_recipients": totalRecipients,
		"created_at":       createdAt,
		"deliveries":       deliveries,
	})
}

// CancelBroadcast обрабатывает POST /api/v1/admin/telegram/broadcasts/{id}/cancel
// Отменяет запланированную или выполняющуюся рассылку
func (h *BroadcastHandler) CancelBroadcast(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	if !user.IsAdmin() && !user.IsTeacher() {
		response.Forbidden(w, "Admin or teacher access required")
		return
	}

	// Парсим ID из URL
	broadcastID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid broadcast ID")
		return
	}

	// Отменяем рассылку
	err = h.broadcastService.CancelBroadcast(r.Context(), broadcastID)
	if err != nil {
		log.Printf("ERROR: Failed to cancel broadcast %s: %v", broadcastID, err)
		h.handleBroadcastError(w, err)
		return
	}

	// Получаем обновленную рассылку
	updatedBroadcast, _, err := h.broadcastService.GetBroadcastDetails(r.Context(), broadcastID)
	if err != nil {
		log.Printf("ERROR: Failed to get updated broadcast details for ID %s: %v", broadcastID, err)
		h.handleBroadcastError(w, err)
		return
	}

	response.OK(w, map[string]interface{}{
		"broadcast": updatedBroadcast,
		"message":   "Broadcast cancelled successfully",
	})
}

// handleBroadcastError обрабатывает специфичные для broadcast ошибки
func (h *BroadcastHandler) handleBroadcastError(w http.ResponseWriter, err error) {
	// Проверяем ошибки repository
	if errors.Is(err, repository.ErrBroadcastListNotFound) {
		response.NotFound(w, "Broadcast list not found")
		return
	}
	if errors.Is(err, repository.ErrBroadcastNotFound) {
		response.NotFound(w, "Broadcast not found")
		return
	}

	// Проверяем ошибки models (валидация)
	if errors.Is(err, models.ErrInvalidBroadcastName) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Broadcast list name must be at least 3 characters")
		return
	}
	if errors.Is(err, models.ErrInvalidBroadcastUsers) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Broadcast list must contain at least one user")
		return
	}
	if errors.Is(err, models.ErrInvalidBroadcastListID) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid broadcast list ID")
		return
	}
	if errors.Is(err, models.ErrInvalidBroadcastMessage) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Broadcast message is required")
		return
	}
	if errors.Is(err, models.ErrBroadcastMessageTooLong) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Broadcast message must not exceed 4096 characters")
		return
	}

	// Проверяем ошибки service
	if errors.Is(err, service.ErrBroadcastAlreadyInProgress) {
		response.Conflict(w, response.ErrCodeConflict, "Broadcast already in progress or completed")
		return
	}
	if errors.Is(err, service.ErrBroadcastCannotCancel) {
		response.Conflict(w, response.ErrCodeConflict, "Cannot cancel broadcast in current status")
		return
	}
	if errors.Is(err, service.ErrUsersNotLinkedToTelegram) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Some users are not linked to Telegram")
		return
	}
	if errors.Is(err, service.ErrTelegramNotConfigured) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Telegram bot is not configured. Please set TELEGRAM_BOT_TOKEN in environment variables.")
		return
	}
	if errors.Is(err, service.ErrInvalidFilePath) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid file path: path traversal attempts are not allowed")
		return
	}

	// Проверяем бизнес-логические ошибки
	if err.Error() == "broadcast list is deleted" {
		response.Conflict(w, response.ErrCodeConflict, "Broadcast list is deleted")
		return
	}
	if err.Error() == "broadcast already completed" || err.Error() == "broadcast already cancelled" {
		response.Conflict(w, response.ErrCodeConflict, err.Error())
		return
	}

	// Неизвестная ошибка
	response.InternalError(w, "An error occurred processing your request")
}

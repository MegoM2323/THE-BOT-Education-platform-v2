package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"log"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"
	"tutoring-platform/pkg/telegram"
)

// AdminTelegramHandler обрабатывает эндпоинты администратора для управления Telegram пользователями
type AdminTelegramHandler struct {
	telegramService *service.TelegramService
	userService     *service.UserService
	telegramRepo    repository.TelegramUserRepository
}

// NewAdminTelegramHandler создает новый AdminTelegramHandler
func NewAdminTelegramHandler(
	telegramService *service.TelegramService,
	userService *service.UserService,
	telegramRepo repository.TelegramUserRepository,
) *AdminTelegramHandler {
	return &AdminTelegramHandler{
		telegramService: telegramService,
		userService:     userService,
		telegramRepo:    telegramRepo,
	}
}

// ListUsersWithTelegram обрабатывает GET /admin/users/telegram
// Возвращает список всех пользователей с информацией о их Telegram статусе
func (h *AdminTelegramHandler) ListUsersWithTelegram(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Проверяем, что пользователь админ
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	searchQuery := r.URL.Query().Get("search")
	linked := r.URL.Query().Get("linked")
	subscribed := r.URL.Query().Get("subscribed")

	usersWithTelegram, err := h.userService.GetAllUsersWithTelegramInfo(r.Context())
	if err != nil {
		log.Printf("ERROR: Failed to retrieve users with telegram info: %v", err)
		response.InternalError(w, "Failed to retrieve users")
		return
	}

	var result []map[string]interface{}
	for _, userMap := range usersWithTelegram {
		telegramLinked, _ := userMap["telegram_linked"].(bool)
		telegramUsername, _ := userMap["telegram_username"].(string)
		isSubscribed, _ := userMap["subscribed"].(bool)

		if linked == "true" && !telegramLinked {
			continue
		}
		if linked == "false" && telegramLinked {
			continue
		}

		if subscribed == "true" && (!telegramLinked || !isSubscribed) {
			continue
		}
		if subscribed == "false" && telegramLinked && isSubscribed {
			continue
		}

		if searchQuery != "" {
			if !telegramLinked || len(telegramUsername) < len(searchQuery) {
				continue
			}
			match := false
			for i := 0; i <= len(telegramUsername)-len(searchQuery); i++ {
				if telegramUsername[i:i+len(searchQuery)] == searchQuery {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		result = append(result, userMap)
	}

	response.OK(w, map[string]interface{}{
		"users": result,
		"count": len(result),
	})
}

// UnlinkUserTelegram обрабатывает DELETE /admin/users/{id}/telegram
// Отвязывает Telegram аккаунт пользователя
func (h *AdminTelegramHandler) UnlinkUserTelegram(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Проверяем, что пользователь админ
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid user ID")
		return
	}

	// Проверяем, что администратор не отвязывает свой собственный Telegram
	if userID == user.ID {
		response.Conflict(w, response.ErrCodeConflict, "Cannot unlink your own telegram account")
		return
	}

	// Отвязываем Telegram
	if err := h.telegramService.UnlinkUser(r.Context(), userID); err != nil {
		log.Printf("ERROR: Failed to unlink telegram for user %s: %v", userID, err)

		if errors.Is(err, repository.ErrTelegramUserNotFound) {
			response.NotFound(w, "Telegram account not linked")
			return
		}

		response.InternalError(w, "Failed to unlink telegram account")
		return
	}

	log.Println("INFO")

	response.OK(w, map[string]interface{}{
		"message": "Telegram account unlinked successfully",
		"user_id": userID.String(),
	})
}

// SetUserTelegram обрабатывает PUT /admin/users/{id}/telegram
// Устанавливает или обновляет Telegram username пользователя
func (h *AdminTelegramHandler) SetUserTelegram(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Проверяем, что пользователь админ
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid user ID")
		return
	}

	// Парсим тело запроса
	var req struct {
		TelegramUsername string `json:"telegram_username"`
		TelegramID       int64  `json:"telegram_id"`
		ChatID           int64  `json:"chat_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Валидируем входные данные
	if req.TelegramUsername == "" {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Telegram username is required")
		return
	}
	if req.TelegramID == 0 {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Telegram ID is required")
		return
	}
	if req.ChatID == 0 {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Chat ID is required")
		return
	}

	// Проверяем, что пользователь существует
	targetUser, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	// Привязываем Telegram через сервис (с проверкой конфликтов)
	if err := h.telegramService.SetUserTelegram(r.Context(), userID, req.TelegramID, req.ChatID, req.TelegramUsername); err != nil {
		log.Printf("ERROR: Failed to set telegram for user %s: %v", userID, err)

		if errors.Is(err, service.ErrTelegramIDAlreadyLinked) {
			response.Conflict(w, response.ErrCodeAlreadyExists, "Telegram account already linked to another user")
			return
		}

		response.InternalError(w, "Failed to set telegram account")
		return
	}

	log.Println("INFO")

	response.OK(w, map[string]interface{}{
		"message":           "Telegram account set successfully",
		"user_id":           userID.String(),
		"telegram_username": req.TelegramUsername,
		"telegram_id":       req.TelegramID,
		"chat_id":           req.ChatID,
		"full_name":         targetUser.GetFullName(),
		"email":             targetUser.Email,
	})
}

// SendMessageToUser обрабатывает POST /admin/users/{id}/telegram/message
// Отправляет сообщение пользователю в Telegram
func (h *AdminTelegramHandler) SendMessageToUser(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Проверяем, что пользователь админ
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid user ID")
		return
	}

	// Парсим тело запроса
	var req struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Валидируем сообщение
	if req.Message == "" {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Message is required")
		return
	}
	if len(req.Message) > 4096 {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Message is too long (max 4096 characters)")
		return
	}

	// Получаем Telegram информацию пользователя
	telegramUser, err := h.telegramRepo.GetByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrTelegramUserNotFound) {
			response.NotFound(w, "User does not have telegram account linked")
			return
		}
		response.InternalError(w, "Failed to get telegram user info")
		return
	}

	// Отправляем сообщение
	if err := h.telegramService.SendMessage(r.Context(), telegramUser.ChatID, req.Message); err != nil {
		log.Printf("ERROR: Failed to send message to user %s: %v", userID, err)

		// Check if it's a Telegram API error
		if telegramErr, ok := err.(*telegram.TelegramError); ok {
			if telegramErr.ErrorCode == 403 {
				response.Conflict(w, response.ErrCodeConflict, "Bot is blocked by user")
				return
			}
		}

		response.InternalError(w, "Failed to send message to user")
		return
	}

	log.Println("INFO")

	response.OK(w, map[string]interface{}{
		"message": "Message sent successfully",
		"user_id": userID.String(),
	})
}

// GetUserTelegramInfo обрабатывает GET /admin/users/{id}/telegram
// Возвращает детальную информацию о Telegram привязке пользователя
func (h *AdminTelegramHandler) GetUserTelegramInfo(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Проверяем, что пользователь админ
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid user ID")
		return
	}

	// Получаем Telegram информацию
	telegramUser, err := h.telegramRepo.GetByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrTelegramUserNotFound) {
			response.OK(w, map[string]interface{}{
				"linked": false,
			})
			return
		}
		response.InternalError(w, "Failed to get telegram user info")
		return
	}

	// Получаем информацию обычного пользователя для контекста
	targetUser, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	response.OK(w, map[string]interface{}{
		"user": map[string]interface{}{
			"id":        targetUser.ID.String(),
			"email":     targetUser.Email,
			"full_name": targetUser.GetFullName(),
			"role":      string(targetUser.Role),
		},
		"telegram": map[string]interface{}{
			"linked":      true,
			"username":    telegramUser.Username,
			"telegram_id": telegramUser.TelegramID,
			"chat_id":     telegramUser.ChatID,
			"subscribed":  telegramUser.Subscribed,
			"linked_at":   telegramUser.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

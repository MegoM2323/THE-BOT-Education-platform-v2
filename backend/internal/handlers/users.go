package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/pagination"
	"tutoring-platform/pkg/response"
)

// UserHandler обрабатывает эндпоинты пользователей
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler создает новый UserHandler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUsers обрабатывает GET /api/v1/users
// @Summary      List users
// @Description  Get list of users (admin: all, non-admin: filtered by role only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        role  query  string  false  "Filter by role (student, teacher, admin)"
// @Param        page  query  int  false  "Page number (default: 1)"
// @Param        per_page  query  int  false  "Items per page (default: 20, max: 100)"
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      403  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /users [get]
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Опциональный фильтр по роли
	var roleFilter *models.UserRole
	roleStr := r.URL.Query().Get("role")
	if roleStr != "" {
		role := models.UserRole(roleStr)
		if role == models.RoleStudent || role == models.RoleAdmin || role == models.RoleTeacher {
			roleFilter = &role
		}
	}

	// Админы и методисты могут получить полный список без фильтров
	if !user.IsAdmin() && !user.IsTeacher() {
		// Не-админы и не-методисты должны указать role
		if roleFilter == nil {
			response.Forbidden(w, "Admin or teacher access required for unfiltered user list")
			return
		}
		// Разрешаем только запросы на студентов или методистов
		if *roleFilter != models.RoleStudent && *roleFilter != models.RoleTeacher {
			response.Forbidden(w, "Only role=student or role=teacher allowed for non-admin users")
			return
		}
	}

	// Парсим параметры пагинации
	params := pagination.ParseParams(r)

	users, total, err := h.userService.ListUsersWithPagination(r.Context(), roleFilter, params.Offset, params.PerPage)
	if err != nil {
		response.InternalError(w, "Failed to retrieve users")
		return
	}

	if users == nil {
		users = []*models.User{}
	}

	response.OK(w, pagination.NewResponse(map[string]interface{}{
		"users": users,
	}, params.Page, params.PerPage, total))
}

// CreateUser обрабатывает POST /api/v1/users
// @Summary      Create user
// @Description  Create a new user (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        payload  body      models.CreateUserRequest  true  "User details"
// @Success      201  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только админы могут создавать пользователей
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Создаем пользователя
	newUser, err := h.userService.CreateUser(r.Context(), &req)
	if err != nil {
		h.handleUserError(w, err)
		return
	}

	response.Created(w, map[string]interface{}{
		"user": newUser,
	})
}

// GetUser обрабатывает GET /api/v1/users/:id
// @Summary      Get user details
// @Description  Get details for a specific user (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "User ID"
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только админы могут получать детали пользователя
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid user ID")
		return
	}

	targetUser, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	response.OK(w, map[string]interface{}{
		"user": targetUser,
	})
}

// UpdateUser обрабатывает PUT /api/v1/users/:id
// @Summary      Update user
// @Description  Update user details (admin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "User ID"
// @Param        payload  body      models.UpdateUserRequest  true  "Updated user details"
// @Success      200  {object}  response.SuccessResponse{data=map[string]interface{}}
// @Failure      400  {object}  response.ErrorResponse
// @Failure      403  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только админы могут обновлять пользователей
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid user ID")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Обновляем пользователя
	updatedUser, err := h.userService.UpdateUser(r.Context(), userID, &req)
	if err != nil {
		h.handleUserError(w, err)
		return
	}

	response.OK(w, map[string]interface{}{
		"user": updatedUser,
	})
}

// DeleteUser обрабатывает DELETE /api/v1/users/:id
// @Summary      Delete user
// @Description  Delete a user (admin only, soft delete)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "User ID"
// @Success      200  {object}  response.SuccessResponse{data=interface{}}
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Security     SessionAuth
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Только админы могут удалять пользователей
	if !user.IsAdmin() {
		response.Forbidden(w, "Admin access required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid user ID")
		return
	}

	// Предотвращаем удаление админом самого себя
	if userID == user.ID {
		response.Conflict(w, response.ErrCodeConflict, "Cannot delete your own account")
		return
	}

	// Удаляем пользователя
	if err := h.userService.DeleteUser(r.Context(), userID); err != nil {
		h.handleUserError(w, err)
		return
	}

	response.OK(w, map[string]string{
		"message": "User deleted successfully",
	})
}

// handleUserError обрабатывает специфичные для пользователей ошибки
// Возвращает одинаковые сообщения об ошибках для существующих и несуществующих пользователей
// чтобы предотвратить user enumeration атаки
func (h *UserHandler) handleUserError(w http.ResponseWriter, err error) {
	// Проверяем ошибки repository - НЕ раскрываем информацию о существовании пользователя
	if errors.Is(err, repository.ErrUserExists) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Unable to complete this operation")
		return
	}
	if errors.Is(err, repository.ErrUserNotFound) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Unable to complete this operation")
		return
	}

	// Проверяем ошибки зависимостей при удалении
	if errors.Is(err, repository.ErrUserHasActiveLessons) {
		response.Conflict(w, response.ErrCodeConflict, "Невозможно удалить преподавателя с активными занятиями")
		return
	}
	if errors.Is(err, repository.ErrUserHasActiveBookings) {
		response.Conflict(w, response.ErrCodeConflict, "Невозможно удалить студента с активными бронированиями")
		return
	}
	if errors.Is(err, repository.ErrUserHasPayments) {
		response.Conflict(w, response.ErrCodeConflict, "Невозможно удалить пользователя с историей платежей")
		return
	}

	// Проверяем ошибки models (валидация)
	if errors.Is(err, models.ErrInvalidEmail) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid email address. Must be in format: user@example.com")
		return
	}
	if errors.Is(err, models.ErrPasswordTooShort) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Password must be at least 8 characters long")
		return
	}
	if errors.Is(err, models.ErrInvalidFullName) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Full name is required and must be at least 2 characters")
		return
	}
	if errors.Is(err, models.ErrInvalidRole) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid role specified. Must be one of: student, teacher, admin")
		return
	}
	if errors.Is(err, models.ErrInvalidTelegramHandle) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid Telegram username. Must be 3-32 characters, containing only letters, numbers and underscores")
		return
	}

	// Неизвестная ошибка - логируем полную информацию для отладки
	// В production окружении подробная информация не передается клиенту
	log.Error().Err(err).Msg("Unhandled error in user operation")
	response.InternalError(w, "An error occurred processing your request")
}

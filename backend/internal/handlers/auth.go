package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/auth"
	"tutoring-platform/pkg/response"
)

// AuthHandler обрабатывает эндпоинты аутентификации
type AuthHandler struct {
	authService   *service.AuthService
	userService   *service.UserService
	creditService *service.CreditService
	sessionMaxAge int
	isProduction  bool
	sameSite      http.SameSite
	csrfStore     *middleware.CSRFTokenStore
}

// NewAuthHandler создает новый AuthHandler
func NewAuthHandler(authService *service.AuthService, creditService *service.CreditService, sessionMaxAge int, isProduction bool) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		creditService: creditService,
		sessionMaxAge: sessionMaxAge,
		isProduction:  isProduction,
		sameSite:      http.SameSiteLaxMode, // Default to Lax for better compatibility
	}
}

// NewAuthHandlerWithSameSite создает новый AuthHandler с указанным SameSite
func NewAuthHandlerWithSameSite(authService *service.AuthService, creditService *service.CreditService, sessionMaxAge int, isProduction bool, sameSite string) *AuthHandler {
	var sameSiteMode http.SameSite
	switch sameSite {
	case "Strict":
		sameSiteMode = http.SameSiteStrictMode
	case "None":
		sameSiteMode = http.SameSiteNoneMode
	default:
		sameSiteMode = http.SameSiteLaxMode // Default to Lax for better compatibility
	}
	return &AuthHandler{
		authService:   authService,
		creditService: creditService,
		sessionMaxAge: sessionMaxAge,
		isProduction:  isProduction,
		sameSite:      sameSiteMode,
	}
}

// SetUserService устанавливает userService (используется для инициализации после создания handler)
func (h *AuthHandler) SetUserService(userService *service.UserService) {
	h.userService = userService
}

// SetCSRFStore устанавливает CSRF token store
func (h *AuthHandler) SetCSRFStore(store *middleware.CSRFTokenStore) {
	h.csrfStore = store
}

// GetCSRFToken возвращает CSRF token для текущей сессии
func (h *AuthHandler) GetCSRFToken(w http.ResponseWriter, r *http.Request) {
	// Получаем session из контекста
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok || session == nil {
		response.Unauthorized(w, "Session not found")
		return
	}

	// Генерируем CSRF token для сессии
	token, err := h.csrfStore.GenerateToken(session.ID.String())
	if err != nil {
		response.InternalError(w, "Failed to generate CSRF token")
		return
	}

	response.OK(w, map[string]string{
		"csrf_token": token,
	})
}

// Register обрабатывает POST /api/v1/auth/register
// Публичный эндпоинт для регистрации нового пользователя через email и пароль
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Валидация
	if req.Email == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Email is required")
		return
	}
	if req.Password == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Password is required")
		return
	}
	if req.FullName == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Full name is required")
		return
	}

	// Вызвать сервис для регистрации
	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	// Получаем IP-адрес и user agent для сессии
	ipAddress := r.RemoteAddr
	userAgent := r.Header.Get("User-Agent")

	// Создаем сессию для новоиспеченного пользователя и получаем объект сессии
	sessionToken, session, err := h.authService.CreateSessionWithData(r.Context(), user.ID, ipAddress, userAgent)
	if err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Failed to create session: "+err.Error())
		return
	}

	// Устанавливаем session cookie
	cookieOpts := auth.GetDefaultCookieOptions(sessionToken, h.sessionMaxAge, h.isProduction)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieOpts.Name,
		Value:    cookieOpts.Value,
		Path:     cookieOpts.Path,
		MaxAge:   cookieOpts.MaxAge,
		Secure:   cookieOpts.Secure,
		HttpOnly: cookieOpts.HTTPOnly,
		SameSite: h.sameSite, // Use configured SameSite mode
	})

	// Генерируем CSRF token для новой сессии
	var csrfToken string
	if h.csrfStore != nil {
		token, err := h.csrfStore.GenerateToken(session.ID.String())
		if err != nil {
			response.InternalError(w, "Failed to generate CSRF token")
			return
		}
		csrfToken = token

		// Устанавливаем CSRF token в заголовок для фронтенда
		w.Header().Set("X-CSRF-Token", csrfToken)
	}

	response.Created(w, map[string]interface{}{
		"user":       user,
		"token":      sessionToken,
		"csrf_token": csrfToken,
	})
}

// Login обрабатывает POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Получаем IP-адрес и user agent
	ipAddress := r.RemoteAddr
	userAgent := r.Header.Get("User-Agent")

	// Аутентифицируем пользователя
	loginResp, err := h.authService.Login(r.Context(), &req, ipAddress, userAgent)
	if err != nil {
		response.Unauthorized(w, err.Error())
		return
	}

	// Устанавливаем session cookie
	cookieOpts := auth.GetDefaultCookieOptions(loginResp.SessionToken, h.sessionMaxAge, h.isProduction)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieOpts.Name,
		Value:    cookieOpts.Value,
		Path:     cookieOpts.Path,
		MaxAge:   cookieOpts.MaxAge,
		Secure:   cookieOpts.Secure,
		HttpOnly: cookieOpts.HTTPOnly,
		SameSite: h.sameSite, // Use configured SameSite mode
	})

	// Генерируем CSRF token для новой сессии
	var csrfToken string
	if h.csrfStore != nil {
		sessionID := loginResp.Session.ID.String()
		token, err := h.csrfStore.GenerateToken(sessionID)
		if err != nil {
			response.InternalError(w, "Failed to generate CSRF token")
			return
		}
		csrfToken = token

		// Устанавливаем CSRF token в заголовок для фронтенда
		w.Header().Set("X-CSRF-Token", csrfToken)
	}

	response.OK(w, map[string]interface{}{
		"user":       loginResp.User,
		"token":      loginResp.SessionToken,
		"csrf_token": csrfToken,
	})
}

// Logout обрабатывает POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Not authenticated")
		return
	}

	// Удаляем CSRF token из хранилища
	if h.csrfStore != nil {
		h.csrfStore.DeleteToken(session.ID.String())
	}

	// Удаляем сессию
	if err := h.authService.Logout(r.Context(), session.ID); err != nil {
		response.InternalError(w, "Failed to logout")
		return
	}

	// Очищаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	response.OK(w, map[string]string{
		"message": "Logged out successfully",
	})
}

// GetMe обрабатывает GET /api/v1/auth/me
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Not authenticated")
		return
	}

	// Получаем баланс кредитов для текущего пользователя
	respBody := map[string]interface{}{
		"user": user,
	}

	// Получаем баланс кредитов если creditService доступен
	if h.creditService != nil {
		credits, err := h.creditService.GetBalance(r.Context(), user.ID)
		if err == nil && credits != nil {
			respBody["balance"] = credits.Balance
		}
	}

	// Добавляем payment_enabled для студентов (уже включено в user object через MarshalJSON)
	// User.PaymentEnabled уже сериализуется в JSON автоматически

	response.OK(w, respBody)
}

// RegisterViaTelegram обрабатывает POST /api/v1/auth/register-telegram
// Публичный эндпоинт для регистрации студента через Telegram
func (h *AuthHandler) RegisterViaTelegram(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterViaTelegramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Валидация
	if req.TelegramUsername == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Telegram username is required")
		return
	}

	// Вызвать сервис для регистрации через Telegram
	user, err := h.authService.RegisterViaTelegram(r.Context(), &req)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	response.Created(w, map[string]interface{}{
		"user": user,
	})
}

// UpdateProfile обрабатывает PUT /api/v1/auth/profile
// Эндпоинт позволяет пользователю обновлять свой профиль (email, имя, telegram_username)
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Not authenticated")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Обновляем профиль текущего пользователя
	updatedUser, err := h.userService.UpdateUser(r.Context(), user.ID, &req)
	if err != nil {
		// Проверяем типы ошибок
		if err == models.ErrInvalidTelegramHandle {
			response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid Telegram username. Must be 3-32 characters, containing only letters, numbers and underscores")
			return
		}
		if err == models.ErrInvalidEmail {
			response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid email address. Must be in format: user@example.com")
			return
		}
		if err == models.ErrInvalidFullName {
			response.BadRequest(w, response.ErrCodeValidationFailed, "Full name is required and must be at least 2 characters")
			return
		}

		response.InternalError(w, "Failed to update profile")
		return
	}

	response.OK(w, map[string]interface{}{
		"user": updatedUser,
	})
}

// ChangePassword обрабатывает POST /api/v1/auth/change-password
// Эндпоинт позволяет пользователю изменить свой пароль
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Not authenticated")
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Invalid request body")
		return
	}

	// Валидация входных данных
	if req.OldPassword == "" {
		response.BadRequest(w, response.ErrCodeInvalidInput, "Old password is required")
		return
	}
	if req.NewPassword == "" {
		response.BadRequest(w, response.ErrCodeValidationFailed, "New password is required")
		return
	}
	if len(req.NewPassword) < 8 {
		response.BadRequest(w, response.ErrCodeValidationFailed, "New password must be at least 8 characters long")
		return
	}
	if req.OldPassword == req.NewPassword {
		response.BadRequest(w, response.ErrCodeValidationFailed, "New password must be different from the current password")
		return
	}

	// Изменяем пароль
	if err := h.authService.ChangePassword(r.Context(), user.ID, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "current password is incorrect" {
			response.Unauthorized(w, "Current password is incorrect")
			return
		}
		response.InternalError(w, "Failed to change password")
		return
	}

	response.OK(w, map[string]string{
		"message": "Password changed successfully",
	})
}

// handleAuthError обрабатывает ошибки аутентификации и регистрации
// Стандартизирует ответы об ошибках чтобы предотвратить user enumeration атаки
func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error) {
	// Проверяем типы ошибок валидации
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
	if errors.Is(err, models.ErrInvalidTelegramHandle) {
		response.BadRequest(w, response.ErrCodeValidationFailed, "Invalid Telegram username. Must be 3-32 characters, containing only letters, numbers and underscores")
		return
	}

	// Для любых других ошибок (включая "user already exists") возвращаем генерическое сообщение
	// чтобы не дать злоумышленнику возможность определить существует ли пользователь
	response.BadRequest(w, response.ErrCodeValidationFailed, "Unable to complete this operation")
}

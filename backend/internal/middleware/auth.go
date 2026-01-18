package middleware

import (
	"context"
	"net/http"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/service"
	"tutoring-platform/pkg/response"
)

// ContextKey тип для ключей контекста
type ContextKey string

const (
	// UserContextKey ключ контекста для текущего пользователя
	UserContextKey ContextKey = "user"
	// SessionContextKey ключ контекста для текущей сессии
	SessionContextKey ContextKey = "session"
)

// AuthMiddleware обрабатывает аутентификацию путем валидации сессионных cookies
type AuthMiddleware struct {
	authService   *service.AuthService
	isProduction  bool
	sameSite      http.SameSite
	sessionMaxAge int // MaxAge для cookie в секундах
}

// NewAuthMiddleware создает новый AuthMiddleware
func NewAuthMiddleware(authService *service.AuthService, isProduction bool) *AuthMiddleware {
	return &AuthMiddleware{
		authService:   authService,
		isProduction:  isProduction,
		sameSite:      http.SameSiteLaxMode, // Default to Lax for better compatibility
		sessionMaxAge: 7 * 24 * 60 * 60,     // 7 дней по умолчанию
	}
}

// NewAuthMiddlewareWithSameSite создает новый AuthMiddleware с указанным SameSite
func NewAuthMiddlewareWithSameSite(authService *service.AuthService, isProduction bool, sameSite string) *AuthMiddleware {
	var sameSiteMode http.SameSite
	switch sameSite {
	case "Strict":
		sameSiteMode = http.SameSiteStrictMode
	case "None":
		sameSiteMode = http.SameSiteNoneMode
	default:
		sameSiteMode = http.SameSiteLaxMode // Default to Lax for better compatibility
	}
	return &AuthMiddleware{
		authService:   authService,
		isProduction:  isProduction,
		sameSite:      sameSiteMode,
		sessionMaxAge: 7 * 24 * 60 * 60, // 7 дней по умолчанию
	}
}

// NewAuthMiddlewareWithSessionMaxAge создает новый AuthMiddleware с указанным SameSite и sessionMaxAge
func NewAuthMiddlewareWithSessionMaxAge(authService *service.AuthService, isProduction bool, sameSite string, sessionMaxAge int) *AuthMiddleware {
	var sameSiteMode http.SameSite
	switch sameSite {
	case "Strict":
		sameSiteMode = http.SameSiteStrictMode
	case "None":
		sameSiteMode = http.SameSiteNoneMode
	default:
		sameSiteMode = http.SameSiteLaxMode
	}
	return &AuthMiddleware{
		authService:   authService,
		isProduction:  isProduction,
		sameSite:      sameSiteMode,
		sessionMaxAge: sessionMaxAge,
	}
}

// Authenticate middleware, который валидирует сессионный cookie
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем сессионный cookie
		cookie, err := r.Cookie("session")
		if err != nil {
			response.Unauthorized(w, "Authentication required")
			return
		}

		// Валидируем и получаем сессию с встроенной проверкой атомарности
		// и буфером времени для предотвращения race condition
		session, err := m.authService.ValidateSessionWithBuffer(r.Context(), cookie.Value)
		if err != nil {
			response.Unauthorized(w, "Invalid or expired session")
			return
		}

		// Получаем пользователя из сессии (уже загружено в ValidateSessionWithBuffer)
		user, err := m.authService.GetUserByID(r.Context(), session.UserID)
		if err != nil {
			response.Unauthorized(w, "User not found")
			return
		}

		// Проверяем что пользователь не удален (мягкое удаление)
		if user.IsDeleted() {
			response.Unauthorized(w, "User account is deleted or deactivated")
			return
		}

		// Пытаемся продлить сессию (автоматическое продление при каждом запросе)
		// ValidateSessionWithBuffer уже выполнил предварительную проверку близости к истечению
		newToken, err := m.authService.RefreshSessionToken(r.Context(), cookie.Value)
		if err == nil && newToken != cookie.Value {
			// Сессия была продлена, устанавливаем новый cookie
			setCookie := &http.Cookie{
				Name:     "session",
				Value:    newToken,
				Path:     "/",
				MaxAge:   m.sessionMaxAge,
				HttpOnly: true,
				Secure:   m.isProduction,
				SameSite: m.sameSite,
			}
			http.SetCookie(w, setCookie)
		}
		// Если продление не требуется или произошла ошибка, продолжаем с текущей сессией

		// Добавляем пользователя и сессию в контекст
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, SessionContextKey, session)

		// Вызываем следующий обработчик
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthenticate middleware, который валидирует сессионный cookie, если он присутствует
func (m *AuthMiddleware) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем сессионный cookie
		cookie, err := r.Cookie("session")
		if err != nil {
			// Нет сессионного cookie, продолжаем без аутентификации
			next.ServeHTTP(w, r)
			return
		}

		// Валидируем и получаем сессию с встроенной проверкой атомарности
		// и буфером времени для предотвращения race condition
		session, err := m.authService.ValidateSessionWithBuffer(r.Context(), cookie.Value)
		if err != nil {
			// Невалидная сессия, продолжаем без аутентификации
			next.ServeHTTP(w, r)
			return
		}

		// Получаем пользователя из сессии
		user, err := m.authService.GetUserByID(r.Context(), session.UserID)
		if err != nil {
			// Пользователь не найден, продолжаем без аутентификации
			next.ServeHTTP(w, r)
			return
		}

		// Проверяем что пользователь не удален (мягкое удаление)
		if user.IsDeleted() {
			// Пользователь удален, продолжаем без аутентификации
			next.ServeHTTP(w, r)
			return
		}

		// Пытаемся продлить сессию (автоматическое продление при каждом запросе)
		// ValidateSessionWithBuffer уже выполнил предварительную проверку близости к истечению
		newToken, err := m.authService.RefreshSessionToken(r.Context(), cookie.Value)
		if err == nil && newToken != cookie.Value {
			// Сессия была продлена, устанавливаем новый cookie
			setCookie := &http.Cookie{
				Name:     "session",
				Value:    newToken,
				Path:     "/",
				MaxAge:   m.sessionMaxAge,
				HttpOnly: true,
				Secure:   m.isProduction,
				SameSite: m.sameSite,
			}
			http.SetCookie(w, setCookie)
		}
		// Если продление не требуется или произошла ошибка, продолжаем с текущей сессией

		// Добавляем пользователя и сессию в контекст
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, SessionContextKey, session)

		// Вызываем следующий обработчик
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext извлекает текущего пользователя из контекста
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

// GetSessionFromContext извлекает текущую сессию из контекста
func GetSessionFromContext(ctx context.Context) (*models.SessionWithUser, bool) {
	session, ok := ctx.Value(SessionContextKey).(*models.SessionWithUser)
	return session, ok
}

// SetUserInContext устанавливает пользователя в контекст (для тестов)
func SetUserInContext(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// SetSessionInContext устанавливает сессию в контекст (для тестов)
func SetSessionInContext(ctx context.Context, session *models.SessionWithUser) context.Context {
	return context.WithValue(ctx, SessionContextKey, session)
}

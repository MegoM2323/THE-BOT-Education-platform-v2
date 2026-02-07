package middleware

import (
	"net/http"

	"tutoring-platform/internal/models"
	"tutoring-platform/pkg/response"
)

// RequireRole middleware, который проверяет, имеет ли пользователь требуемую роль
func RequireRole(roles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем пользователя из контекста
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				response.Unauthorized(w, "Authentication required")
				return
			}

			// Проверяем, имеет ли пользователь одну из требуемых ролей
			hasRole := false
			for _, role := range roles {
				if user.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				response.Forbidden(w, "Insufficient permissions")
				return
			}

			// Вызываем следующий обработчик
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin удобный middleware, требующий роль администратора
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(models.RoleAdmin)(next)
}

// RequireTeacher удобный middleware, требующий роль методиста
func RequireTeacher(next http.Handler) http.Handler {
	return RequireRole(models.RoleTeacher)(next)
}

// RequireStudent удобный middleware, требующий роль студента
func RequireStudent(next http.Handler) http.Handler {
	return RequireRole(models.RoleStudent)(next)
}

// RequireTeacherOrAdmin удобный middleware, требующий роль методиста или администратора
func RequireTeacherOrAdmin(next http.Handler) http.Handler {
	return RequireRole(models.RoleTeacher, models.RoleAdmin)(next)
}

// RequireAdminOrTeacher удобный middleware, требующий роль администратора или методиста
func RequireAdminOrTeacher(next http.Handler) http.Handler {
	return RequireRole(models.RoleAdmin, models.RoleTeacher)(next)
}

// RequireAdminOnly удобный middleware, требующий ТОЛЬКО роль администратора (не методист)
func RequireAdminOnly(next http.Handler) http.Handler {
	return RequireRole(models.RoleAdmin)(next)
}

// RequireStudentOrAdmin удобный middleware, требующий роль студента или администратора
func RequireStudentOrAdmin(next http.Handler) http.Handler {
	return RequireRole(models.RoleStudent, models.RoleAdmin)(next)
}

package models

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"tutoring-platform/pkg/sanitize"

	"github.com/google/uuid"
)

// UserRole представляет роль пользователя в системе
type UserRole string

const (
	RoleStudent       UserRole = "student"
	RoleTeacher UserRole = "teacher"
	RoleAdmin         UserRole = "admin"
)

// User представляет учетную запись пользователя в системе
type User struct {
	ID                     uuid.UUID      `db:"id" json:"id"`
	Email                  string         `db:"email" json:"email"`
	PasswordHash           string         `db:"password_hash" json:"-"` // Никогда не показывать хеш пароля в JSON
	FirstName              string         `db:"first_name" json:"first_name"`
	LastName               string         `db:"last_name" json:"last_name"`
	Role                   UserRole       `db:"role" json:"role"`
	PaymentEnabled         bool           `db:"payment_enabled" json:"payment_enabled"`
	TelegramUsername       sql.NullString `db:"telegram_username" json:"telegram_username,omitempty"` // Telegram username
	ParentTelegramUsername sql.NullString `db:"parent_telegram_username" json:"parent_telegram_username,omitempty"`
	ParentChatID           sql.NullInt64  `db:"parent_chat_id" json:"-"`
	TelegramLinked         bool           `db:"-" json:"telegram_linked,omitempty"` // Вычисляемое поле, не хранится в БД
	CreatedAt              time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt              sql.NullTime   `db:"deleted_at" json:"deleted_at,omitempty"`
}

// GetFullName возвращает полное имя пользователя
func (u *User) GetFullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

// CreateUserRequest представляет запрос на создание нового пользователя
type CreateUserRequest struct {
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	FullName  string   `json:"full_name"`
	Role      UserRole `json:"role"`
}

// UpdateUserRequest представляет запрос на обновление пользователя
type UpdateUserRequest struct {
	Email                  *string   `json:"email,omitempty"`
	FirstName              *string   `json:"first_name,omitempty"`
	LastName               *string   `json:"last_name,omitempty"`
	FullName               *string   `json:"full_name,omitempty"`
	Role                   *UserRole `json:"role,omitempty"`
	PaymentEnabled         *bool     `json:"payment_enabled,omitempty"`
	TelegramUsername       *string   `json:"telegram_username,omitempty"`
	ParentTelegramUsername *string   `json:"parent_telegram_username,omitempty"`
	ParentChatID           *int64    `json:"parent_chat_id,omitempty"`
	Password               *string   `json:"password,omitempty"` // Новый пароль (для админа)
}

// ChangePasswordRequest представляет запрос на смену пароля
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// RegisterRequest представляет запрос на регистрацию нового пользователя
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
}

// RegisterViaTelegramRequest представляет запрос на регистрацию через Telegram
type RegisterViaTelegramRequest struct {
	TelegramUsername string `json:"telegram_username"`
}

// MarshalJSON customizes JSON marshaling for User to handle sql.NullString and sql.NullTime fields
func (u *User) MarshalJSON() ([]byte, error) {
	// Create a map to build the JSON response
	data := map[string]interface{}{
		"id":              u.ID,
		"email":           u.Email,
		"first_name":      u.FirstName,
		"last_name":       u.LastName,
		"full_name":       u.GetFullName(),
		"role":            u.Role,
		"payment_enabled": u.PaymentEnabled,
		"created_at":      u.CreatedAt,
		"updated_at":      u.UpdatedAt,
		"telegram_linked": u.TelegramLinked,
	}

	// Handle TelegramUsername - convert sql.NullString to string or null
	if u.TelegramUsername.Valid && u.TelegramUsername.String != "" {
		data["telegram_username"] = u.TelegramUsername.String
	}

	// Handle ParentTelegramUsername - convert sql.NullString to string or null
	if u.ParentTelegramUsername.Valid && u.ParentTelegramUsername.String != "" {
		data["parent_telegram_username"] = u.ParentTelegramUsername.String
	}

	// Handle DeletedAt - convert sql.NullTime to ISO8601 string or omit if null
	if u.DeletedAt.Valid {
		data["deleted_at"] = u.DeletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	return json.Marshal(data)
}

// IsDeleted проверяет, удален ли пользователь (мягкое удаление)
func (u *User) IsDeleted() bool {
	return u.DeletedAt.Valid
}

// IsStudent проверяет, является ли пользователь студентом
func (u *User) IsStudent() bool {
	return u.Role == RoleStudent
}

// IsAdmin проверяет, является ли пользователь администратором
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsTeacher проверяет, является ли пользователь методистом
func (u *User) IsTeacher() bool {
	return u.Role == RoleTeacher
}

// CanBeAssignedAsTeacher проверяет, может ли пользователь быть назначен преподавателем занятия
// Допустимые роли: admin, teacher
func (u *User) CanBeAssignedAsTeacher() bool {
	return u.IsAdmin() || u.IsTeacher()
}

// Validate выполняет валидацию CreateUserRequest
func (r *CreateUserRequest) Validate() error {
	// Санитизация входных данных
	r.Email = sanitize.Email(r.Email)
	r.FirstName = sanitize.Name(r.FirstName)
	r.LastName = sanitize.Name(r.LastName)
	r.FullName = sanitize.Name(r.FullName)

	// Валидация email
	if r.Email == "" {
		return ErrInvalidEmail
	}
	// Простая проверка формата email
	if !isValidEmail(r.Email) {
		return ErrInvalidEmail
	}

	// Валидация пароля
	if r.Password == "" {
		return ErrPasswordTooShort
	}
	if len(r.Password) < 8 {
		return ErrPasswordTooShort
	}

	// Валидация имени (поддерживаем и старый FullName, и новые FirstName/LastName)
	if r.FirstName == "" && r.LastName == "" && r.FullName == "" {
		return ErrInvalidFullName
	}
	if r.FirstName != "" && len(r.FirstName) < 2 {
		return ErrInvalidFullName
	}
	if r.FullName != "" && len(r.FullName) < 2 {
		return ErrInvalidFullName
	}

	// Валидация роли
	if r.Role != RoleStudent && r.Role != RoleAdmin && r.Role != RoleTeacher {
		return ErrInvalidRole
	}

	return nil
}

// isValidEmail выполняет базовую валидацию email адреса
func isValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	// Проверяем наличие @ и точки
	atIndex := -1
	lastDotIndex := -1

	for i, char := range email {
		if char == '@' {
			if atIndex != -1 {
				return false // Несколько @
			}
			atIndex = i
		}
		if char == '.' {
			lastDotIndex = i
		}
	}

	// Должна быть ровно одна @
	if atIndex == -1 {
		return false
	}

	// Проверяем структуру: user@domain.ext
	if atIndex == 0 || atIndex == len(email)-1 {
		return false // @ не может быть в начале или конце
	}

	if lastDotIndex == -1 || lastDotIndex <= atIndex {
		return false // Должна быть точка после @
	}

	if lastDotIndex == len(email)-1 {
		return false // Точка не может быть в конце
	}

	return true
}

// Sanitize очищает входные данные UpdateUserRequest
func (r *UpdateUserRequest) Sanitize() {
	if r.Email != nil {
		email := sanitize.Email(*r.Email)
		r.Email = &email
	}
	if r.FirstName != nil {
		firstName := sanitize.Name(*r.FirstName)
		r.FirstName = &firstName
	}
	if r.LastName != nil {
		lastName := sanitize.Name(*r.LastName)
		r.LastName = &lastName
	}
	if r.FullName != nil {
		fullName := sanitize.Name(*r.FullName)
		r.FullName = &fullName
	}
	if r.TelegramUsername != nil {
		*r.TelegramUsername = sanitize.TelegramUsername(*r.TelegramUsername)
	}
	if r.ParentTelegramUsername != nil {
		*r.ParentTelegramUsername = sanitize.TelegramUsername(*r.ParentTelegramUsername)
	}
}

// ValidateTelegramUsername проверяет корректность имени пользователя Telegram
// Telegram username: 5-32 символа, начинается с буквы, содержит буквы, цифры, подчёркивание
func isValidTelegramUsername(username string) bool {
	if len(username) < 5 || len(username) > 32 {
		return false
	}

	// Telegram username должен начинаться с буквы
	if len(username) > 0 && !((username[0] >= 'a' && username[0] <= 'z') ||
		(username[0] >= 'A' && username[0] <= 'Z')) {
		return false
	}

	// Telegram username может содержать только буквы, цифры и подчеркивание
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}

	return true
}

// StudentPaymentStatus представляет статус платежей студента для админа
type StudentPaymentStatus struct {
	ID             uuid.UUID `json:"id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	PaymentEnabled bool      `json:"payment_enabled"`
	UpdatedAt      time.Time `json:"updated_at"`
}

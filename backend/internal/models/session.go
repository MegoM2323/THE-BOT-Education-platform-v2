package models

import (
	"time"

	"github.com/google/uuid"
)

// AuthFailureReason представляет причину неудачной попытки входа
type AuthFailureReason string

const (
	// AuthFailureReasonInvalidPassword - неверный пароль
	AuthFailureReasonInvalidPassword AuthFailureReason = "invalid_password"
	// AuthFailureReasonUserNotFound - пользователь не найден
	AuthFailureReasonUserNotFound AuthFailureReason = "user_not_found"
	// AuthFailureReasonAccountLocked - аккаунт заблокирован
	AuthFailureReasonAccountLocked AuthFailureReason = "account_locked"
	// AuthFailureReasonAccountDeleted - аккаунт удален
	AuthFailureReasonAccountDeleted AuthFailureReason = "account_deleted"
)

// Session представляет активную сессию пользователя
type Session struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	IPAddress string    `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent string    `db:"user_agent" json:"user_agent,omitempty"`
}

// SessionWithUser представляет сессию с информацией о пользователе
type SessionWithUser struct {
	Session
	UserEmail string   `db:"user_email" json:"user_email"`
	UserName  string   `db:"user_name" json:"user_name"`
	UserRole  UserRole `db:"user_role" json:"user_role"`
}

// CreateSessionRequest представляет запрос на создание новой сессии
type CreateSessionRequest struct {
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// IsExpired проверяет, истек ли срок действия сессии
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid проверяет, действительна ли сессия (не истекла)
func (s *Session) IsValid() bool {
	return !s.IsExpired()
}

// TimeUntilExpiry возвращает длительность до истечения срока действия сессии
func (s *Session) TimeUntilExpiry() time.Duration {
	return time.Until(s.ExpiresAt)
}

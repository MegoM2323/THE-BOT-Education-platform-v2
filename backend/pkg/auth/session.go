package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidSession возвращается при неудачной валидации сессии
	ErrInvalidSession = errors.New("invalid session")
	// ErrExpiredSession возвращается когда сессия истекла
	ErrExpiredSession = errors.New("session has expired")
)

// SessionManager управляет операциями с cookie сессий
type SessionManager struct {
	secret []byte
}

// NewSessionManager создает новый SessionManager
func NewSessionManager(secret string) *SessionManager {
	return &SessionManager{
		secret: []byte(secret),
	}
}

// SessionData представляет данные, хранящиеся в cookie сессии
type SessionData struct {
	SessionID uuid.UUID `json:"session_id"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CreateSessionToken создает подписанный токен сессии
func (sm *SessionManager) CreateSessionToken(sessionID, userID uuid.UUID, expiresAt time.Time) (string, error) {
	data := SessionData{
		SessionID: sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	// Сериализуем данные сессии в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Кодируем в base64
	encodedData := base64.URLEncoding.EncodeToString(jsonData)

	// Создаем HMAC подпись
	signature := sm.sign(encodedData)

	// Объединяем данные и подпись
	token := fmt.Sprintf("%s.%s", encodedData, signature)

	return token, nil
}

// ValidateSessionToken валидирует и извлекает данные сессии из токена
// ВАЖНО: При истечении токена возвращает данные И ошибку ErrExpiredSession
// Это позволяет вызывающему коду проверить сессию по БД (где она могла быть продлена)
func (sm *SessionManager) ValidateSessionToken(token string) (*SessionData, error) {
	// Разделяем токен на данные и подпись
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidSession
	}

	encodedData := parts[0]
	signature := parts[1]

	// Проверяем подпись
	expectedSignature := sm.sign(encodedData)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, ErrInvalidSession
	}

	// Декодируем base64
	jsonData, err := base64.URLEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session data: %w", err)
	}

	// Десериализуем JSON
	var data SessionData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	// Проверяем срок действия токена
	// ВАЖНО: Возвращаем данные даже при истечении - сессия в БД может быть продлена
	if time.Now().After(data.ExpiresAt) {
		return &data, ErrExpiredSession
	}

	return &data, nil
}

// sign создает HMAC подпись для переданных данных
func (sm *SessionManager) sign(data string) string {
	h := hmac.New(sha256.New, sm.secret)
	h.Write([]byte(data))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// CookieOptions представляет опции для установки cookie сессий
type CookieOptions struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HTTPOnly bool
	SameSite string // "Strict", "Lax", "None"
}

// GetDefaultCookieOptions возвращает стандартные опции cookie
func GetDefaultCookieOptions(token string, maxAge int, isProduction bool) *CookieOptions {
	// Default to Lax for better compatibility with redirects and navigation
	sameSite := "Lax"
	if isProduction {
		sameSite = "Strict" // Use Strict in production for better CSRF protection
	}
	return &CookieOptions{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Domain:   "",
		MaxAge:   maxAge,
		Secure:   isProduction, // Безопасный режим только в production (HTTPS)
		HTTPOnly: true,
		SameSite: sameSite,
	}
}

// GetCookieOptionsWithSameSite возвращает опции cookie с указанным SameSite
func GetCookieOptionsWithSameSite(token string, maxAge int, isProduction bool, sameSite string) *CookieOptions {
	return &CookieOptions{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Domain:   "",
		MaxAge:   maxAge,
		Secure:   isProduction, // Безопасный режим только в production (HTTPS)
		HTTPOnly: true,
		SameSite: sameSite,
	}
}

// FormatSetCookie форматирует значение заголовка Set-Cookie
func FormatSetCookie(opts *CookieOptions) string {
	cookie := fmt.Sprintf("%s=%s; Path=%s; Max-Age=%d", opts.Name, opts.Value, opts.Path, opts.MaxAge)

	if opts.Domain != "" {
		cookie += fmt.Sprintf("; Domain=%s", opts.Domain)
	}

	if opts.Secure {
		cookie += "; Secure"
	}

	if opts.HTTPOnly {
		cookie += "; HttpOnly"
	}

	if opts.SameSite != "" {
		cookie += fmt.Sprintf("; SameSite=%s", opts.SameSite)
	}

	return cookie
}

// FormatDeleteCookie форматирует заголовок Set-Cookie для удаления cookie
func FormatDeleteCookie(name string) string {
	return fmt.Sprintf("%s=; Path=/; Max-Age=0", name)
}

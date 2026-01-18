package models

import (
	"time"
	"unicode/utf8"

	"tutoring-platform/pkg/sanitize"
)

// TrialRequest представляет запрос на пробный урок с лендинговой страницы
type TrialRequest struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Phone     string    `db:"phone" json:"phone"`
	Telegram  string    `db:"telegram" json:"telegram"`
	Email     *string   `db:"email" json:"email,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CreateTrialRequestInput представляет входные данные для создания запроса на пробный урок
type CreateTrialRequestInput struct {
	Name     string  `json:"name"`
	Phone    string  `json:"phone"`
	Telegram string  `json:"telegram"`
	Email    *string `json:"email,omitempty"`
}

// Validate выполняет валидацию CreateTrialRequestInput
func (r *CreateTrialRequestInput) Validate() error {
	// Санитизация входных данных
	r.Name = sanitize.Name(r.Name)
	r.Phone = sanitize.Input(r.Phone)
	r.Telegram = sanitize.Input(r.Telegram)
	if r.Email != nil {
		email := sanitize.Email(*r.Email)
		r.Email = &email
	}

	nameLen := utf8.RuneCountInString(r.Name)
	if r.Name == "" || nameLen < 2 || nameLen > 100 {
		return ErrInvalidName
	}
	if r.Phone == "" {
		return ErrInvalidPhone
	}
	// Сначала нормализуем telegram (удаляем @ если есть), затем валидируем длину
	normalizedTelegram := r.Telegram
	if len(normalizedTelegram) > 0 && normalizedTelegram[0] == '@' {
		normalizedTelegram = normalizedTelegram[1:]
	}
	telegramLen := utf8.RuneCountInString(normalizedTelegram)
	if normalizedTelegram == "" || telegramLen < 3 || telegramLen > 50 {
		return ErrInvalidTelegram
	}
	return nil
}

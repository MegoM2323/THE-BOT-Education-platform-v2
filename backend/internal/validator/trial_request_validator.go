package validator

import (
	"regexp"
	"strings"

	"tutoring-platform/internal/models"
)

// TrialRequestValidator валидирует данные пробного запроса
type TrialRequestValidator struct {
	phoneRegex *regexp.Regexp
	emailRegex *regexp.Regexp
}

// NewTrialRequestValidator создает новый TrialRequestValidator
func NewTrialRequestValidator() *TrialRequestValidator {
	// Регулярное выражение для телефона: поддерживает различные форматы, включая +7, 8, с/без пробелов и тире
	phoneRegex := regexp.MustCompile(`^[\+]?[(]?[0-9]{1,3}[)]?[-\s\.]?[(]?[0-9]{1,4}[)]?[-\s\.]?[0-9]{1,4}[-\s\.]?[0-9]{1,9}$`)

	// Регулярное выражение для email: базовая валидация email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	return &TrialRequestValidator{
		phoneRegex: phoneRegex,
		emailRegex: emailRegex,
	}
}

// Validate валидирует входные данные пробного запроса
func (v *TrialRequestValidator) Validate(input *models.CreateTrialRequestInput) error {
	// Базовая валидация
	if err := input.Validate(); err != nil {
		return err
	}

	// Проверяем формат телефона
	cleanPhone := strings.ReplaceAll(input.Phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	if !v.phoneRegex.MatchString(cleanPhone) {
		return models.ErrInvalidPhone
	}

	// Извлекаем только цифры для проверки минимальной длины
	phoneDigits := regexp.MustCompile(`[^0-9]`).ReplaceAllString(input.Phone, "")
	if len(phoneDigits) < 10 {
		return models.ErrInvalidPhone
	}

	// Нормализуем telegram (удаляем @ в начале, если есть)
	input.Telegram = strings.TrimPrefix(input.Telegram, "@")

	// Проверяем email, если предоставлен
	if input.Email != nil && *input.Email != "" {
		if !v.emailRegex.MatchString(*input.Email) {
			return models.ErrInvalidEmailFormat
		}
	}

	return nil
}

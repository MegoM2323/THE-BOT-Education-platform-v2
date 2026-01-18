package utils

import (
	"fmt"

	"github.com/google/uuid"
)

// MaskUserID маскирует UUID пользователя для безопасного логирования
// Показывает только первые 8 символов UUID + ***
// Пример: "d3c8c7a6-1234-5678-abcd-ef1234567890" -> "d3c8c7a6***"
func MaskUserID(id uuid.UUID) string {
	s := id.String()
	if len(s) > 8 {
		return s[:8] + "***"
	}
	return "***"
}

// MaskAmount маскирует сумму денег/кредитов для логирования
// Возвращает диапазон вместо точной суммы для безопасности
// Примеры:
//   - 50 кредитов -> "0-100"
//   - 500 кредитов -> "100-1000"
//   - 5000 кредитов -> "1000+"
func MaskAmount(amount int) string {
	if amount < 0 {
		amount = -amount // Работаем с абсолютным значением
	}

	if amount < 100 {
		return "0-100"
	} else if amount < 1000 {
		return "100-1000"
	}
	return "1000+"
}

// MaskEmail маскирует email для логирования
// Показывает только первый символ + *** + домен
// Пример: "user@example.com" -> "u***@example.com"
func MaskEmail(email string) string {
	if len(email) == 0 {
		return "***"
	}

	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}

	if atIndex <= 0 {
		return "***"
	}

	return fmt.Sprintf("%c***%s", rune(email[0]), email[atIndex:])
}

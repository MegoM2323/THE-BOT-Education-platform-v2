package sanitize

import (
	"html"
	"strings"
)

// Input очищает пользовательский ввод от потенциально опасных символов
// Защита от XSS атак путем экранирования HTML-тегов
func Input(s string) string {
	// Удаляем leading/trailing пробелы
	s = strings.TrimSpace(s)

	// Удаляем табуляции и переводы строк для single-line полей
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")

	// Нормализуем множественные пробелы в один
	s = strings.Join(strings.Fields(s), " ")

	// Экранируем HTML-специальные символы
	s = html.EscapeString(s)

	return s
}

// MultilineInput очищает многострочный ввод (например, descriptions)
func MultilineInput(s string) string {
	// Удаляем leading/trailing пробелы
	s = strings.TrimSpace(s)

	// Заменяем табуляции на пробелы
	s = strings.ReplaceAll(s, "\t", " ")

	// Нормализуем переводы строк
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	// Экранируем HTML-специальные символы
	s = html.EscapeString(s)

	return s
}

// Email нормализует email адрес (приводит к lowercase)
func Email(email string) string {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	return email
}

// Name очищает имена пользователей
func Name(name string) string {
	// Удаляем leading/trailing пробелы
	name = strings.TrimSpace(name)

	// Экранируем HTML-специальные символы
	name = html.EscapeString(name)

	// Удаляем множественные пробелы
	name = strings.Join(strings.Fields(name), " ")

	return name
}

// Reason очищает причины транзакций кредитов
// Поддерживает multiline текст (для длинных пояснений)
func Reason(reason string) string {
	return MultilineInput(reason)
}

// Description очищает описания (для lessons и других объектов)
func Description(description string) string {
	return MultilineInput(description)
}

// TelegramUsername очищает и нормализует Telegram username
// Сохраняет @ в начале (стандарт Telegram), приводит к lowercase
func TelegramUsername(username string) string {
	// Удаляем leading/trailing пробелы
	username = strings.TrimSpace(username)

	// Приводим к lowercase
	username = strings.ToLower(username)

	return username
}

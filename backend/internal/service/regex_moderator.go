package service

import (
	"regexp"
	"strings"
)

// RegexModerator предоставляет fallback модерацию на основе regex паттернов
type RegexModerator struct {
	phonePattern    *regexp.Regexp
	emailPattern    *regexp.Regexp
	socialPattern   *regexp.Regexp
	platformPattern *regexp.Regexp
	obfuscatedPhone *regexp.Regexp
}

// NewRegexModerator создает новый RegexModerator с заранее скомпилированными паттернами
func NewRegexModerator() *RegexModerator {
	return &RegexModerator{
		// Телефонные номера: +7, 8, 7, скобки, дефисы
		phonePattern: regexp.MustCompile(`(?i)(\+?7|8)[\s\-\(]?\d{3}[\s\-\)]?\d{3}[\s\-]?\d{2}[\s\-]?\d{2}`),

		// Email адреса
		emailPattern: regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`),

		// Социальные сети и мессенджеры
		socialPattern: regexp.MustCompile(`(?i)(вконтакте|vkontakte|vk\.com|instagram|insta|телеграм|telegram|@\w+|whatsapp|ватсап|viber|вайбер|facebook|фейсбук)`),

		// Платформы для видеозвонков
		platformPattern: regexp.MustCompile(`(?i)(zoom|скайп|skype|discord|дискорд|google meet|гугл мит|teams|тимс|microsoft teams)`),

		// Обфусцированные телефонные номера (написанные словами)
		obfuscatedPhone: regexp.MustCompile(`(?i)(восемь|девять|семь|ноль|один|два|три|четыре|пять|шесть)[\s\-]*(девять|восемь|семь|ноль|один|два|три|четыре|пять|шесть)`),
	}
}

// Check проверяет сообщение на наличие запрещенных паттернов
func (m *RegexModerator) Check(message string) (blocked bool, reason string) {
	message = strings.ToLower(message)

	// Проверка телефонных номеров
	if m.phonePattern.MatchString(message) {
		return true, "Message contains phone number"
	}

	// Проверка email адресов
	if m.emailPattern.MatchString(message) {
		return true, "Message contains email address"
	}

	// Проверка социальных сетей
	if m.socialPattern.MatchString(message) {
		return true, "Message contains social media reference"
	}

	// Проверка платформ для видеозвонков
	if m.platformPattern.MatchString(message) {
		return true, "Message contains video call platform reference"
	}

	// Проверка обфусцированных телефонов
	if m.obfuscatedPhone.MatchString(message) {
		return true, "Message contains obfuscated contact information"
	}

	return false, ""
}

package config

import (
	"testing"
)

// TestTelegramWebhookSecretValidation проверяет, что пустой webhook secret отклоняется при использовании webhook
func TestTelegramWebhookSecretValidation(t *testing.T) {
	tests := []struct {
		name            string
		useWebhook      bool
		webhookSecret   string
		botToken        string
		adminTelegramID int64
		webhookURL      string
		expectError     bool
		errorContains   string
		description     string
	}{
		{
			name:            "webhook_enabled_empty_secret_rejected",
			useWebhook:      true,
			webhookSecret:   "", // ПУСТОЙ СЕКРЕТ = ОШИБКА!
			botToken:        "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
			adminTelegramID: 12345,
			webhookURL:      "https://example.com/webhook",
			expectError:     true,
			errorContains:   "TELEGRAM_WEBHOOK_SECRET обязателен при использовании webhook",
			description:     "Empty webhook secret must be rejected when webhook is enabled",
		},
		{
			name:            "webhook_enabled_short_secret_rejected",
			useWebhook:      true,
			webhookSecret:   "short-secret", // 12 символов < 32
			botToken:        "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
			adminTelegramID: 12345,
			webhookURL:      "https://example.com/webhook",
			expectError:     true,
			errorContains:   "TELEGRAM_WEBHOOK_SECRET должен быть не менее 32 символов",
			description:     "Webhook secret shorter than 32 chars must be rejected",
		},
		{
			name:            "webhook_enabled_valid_secret_accepted",
			useWebhook:      true,
			webhookSecret:   "this-is-a-very-secure-webhook-secret-with-32-chars", // >= 32 символов
			botToken:        "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
			adminTelegramID: 12345,
			webhookURL:      "https://example.com/webhook",
			expectError:     false,
			description:     "Valid webhook secret (32+ chars) must be accepted",
		},
		{
			name:            "webhook_disabled_empty_secret_allowed",
			useWebhook:      false,
			webhookSecret:   "", // Для polling mode пустой секрет OK
			botToken:        "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
			adminTelegramID: 12345,
			webhookURL:      "",
			expectError:     false,
			description:     "Empty webhook secret is OK when webhook is disabled",
		},
		{
			name:            "telegram_disabled_any_config_allowed",
			useWebhook:      true,
			webhookSecret:   "", // Telegram отключен - любая конфиг OK
			botToken:        "", // Пустой токен = Telegram отключен
			adminTelegramID: 0,
			webhookURL:      "",
			expectError:     false,
			description:     "Any Telegram config is OK when Telegram is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					Name:     "test_db",
					User:     "test_user",
					Password: "test_pass",
					SSLMode:  "disable",
				},
				Server: ServerConfig{
					Port: "8080",
					Env:  "development",
				},
				Session: SessionConfig{
					Secret:   "test-session-secret-very-long-with-good-entropy-!@#$",
					MaxAge:   3600,
					HTTPOnly: true,
					SameSite: "Lax",
				},
				Telegram: TelegramConfig{
					BotToken:        tt.botToken,
					AdminTelegramID: tt.adminTelegramID,
					WebhookURL:      tt.webhookURL,
					WebhookSecret:   tt.webhookSecret,
					UseWebhook:      tt.useWebhook,
				},
			}

			err := config.Validate()

			if (err != nil) != tt.expectError {
				t.Errorf("%s: expectError=%v, got error=%v, err=%v", tt.description, tt.expectError, err != nil, err)
			}

			if tt.expectError && err != nil && tt.errorContains != "" {
				if !contains(err.Error(), tt.errorContains) {
					t.Errorf("%s: expected error to contain '%s', got '%s'", tt.description, tt.errorContains, err.Error())
				}
			}
		})
	}
}

// TestTelegramWebhookSecretMinimumLength проверяет, что минимальная длина 32 символа соблюдается
func TestTelegramWebhookSecretMinimumLength(t *testing.T) {
	tests := []struct {
		name         string
		secretLength int
		expectError  bool
		description  string
	}{
		{
			name:         "31_chars_rejected",
			secretLength: 31,
			expectError:  true,
			description:  "31-char secret should be rejected",
		},
		{
			name:         "32_chars_accepted",
			secretLength: 32,
			expectError:  false,
			description:  "32-char secret should be accepted",
		},
		{
			name:         "33_chars_accepted",
			secretLength: 33,
			expectError:  false,
			description:  "33-char secret should be accepted",
		},
		{
			name:         "64_chars_accepted",
			secretLength: 64,
			expectError:  false,
			description:  "64-char secret should be accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём секрет нужной длины
			secret := generateSecretOfLength(tt.secretLength)

			config := &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					Name:     "test_db",
					User:     "test_user",
					Password: "test_pass",
					SSLMode:  "disable",
				},
				Server: ServerConfig{
					Port: "8080",
					Env:  "development",
				},
				Session: SessionConfig{
					Secret:   "test-session-secret-very-long-with-good-entropy-!@#$",
					MaxAge:   3600,
					HTTPOnly: true,
					SameSite: "Lax",
				},
				Telegram: TelegramConfig{
					BotToken:        "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
					AdminTelegramID: 12345,
					WebhookURL:      "https://example.com/webhook",
					WebhookSecret:   secret,
					UseWebhook:      true,
				},
			}

			err := config.Validate()

			if (err != nil) != tt.expectError {
				t.Errorf("%s: expectError=%v, got error=%v", tt.description, tt.expectError, err != nil)
			}
		})
	}
}

// generateSecretOfLength генерирует строку нужной длины
func generateSecretOfLength(length int) string {
	secret := ""
	for i := 0; i < length; i++ {
		secret += "a"
	}
	return secret
}

// contains проверяет, содержит ли строка substring
func contains(str, substring string) bool {
	for i := 0; i <= len(str)-len(substring); i++ {
		if str[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}

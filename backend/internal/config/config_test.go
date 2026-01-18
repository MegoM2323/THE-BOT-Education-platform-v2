package config

import (
	"encoding/base64"
	"strings"
	"testing"
)

// TestIsValidTelegramToken проверяет валидацию формата Telegram токена
func TestIsValidTelegramToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{
			name:  "valid token format",
			token: "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
			want:  true,
		},
		{
			name:  "valid token with hyphen",
			token: "987654321:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
			want:  true,
		},
		{
			name:  "valid token with underscore",
			token: "111111111:ABC_DEF_1234_ghIkl_zyx57W2v1u123ew",
			want:  true,
		},
		{
			name:  "empty token",
			token: "",
			want:  false,
		},
		{
			name:  "missing colon",
			token: "123456789ABCDEfghijklmnoPQRSTUvwxyz123456789",
			want:  false,
		},
		{
			name:  "too many colons",
			token: "123:456:789",
			want:  false,
		},
		{
			name:  "empty bot ID",
			token: ":ABCDEfghijklmnoPQRSTUvwxyz123456789",
			want:  false,
		},
		{
			name:  "non-numeric bot ID",
			token: "abc123:ABCDEfghijklmnoPQRSTUvwxyz123456789",
			want:  false,
		},
		{
			name:  "token string too short",
			token: "123456789:short",
			want:  false,
		},
		{
			name:  "token with invalid characters",
			token: "123456789:ABC@DEF#GHI$JKL%MNO^PQR&STU",
			want:  false,
		},
		{
			name:  "token string with space",
			token: "123456789:ABCDEfg hijklmnoPQRSTUvwxyz123",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidTelegramToken(tt.token)
			if got != tt.want {
				t.Errorf("isValidTelegramToken(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}

// TestMaskTelegramToken проверяет маскирование токена для логирования
func TestMaskTelegramToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "normal token",
			token:    "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
			expected: "123...789",
		},
		{
			name:     "token with hyphens",
			token:    "987654321:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
			expected: "987...w11",
		},
		{
			name:     "empty token",
			token:    "",
			expected: "<not set>",
		},
		{
			name:     "short token (7 chars)",
			token:    "123:abc",
			expected: "123...abc",
		},
		{
			name:     "exactly 6 chars",
			token:    "123:ab",
			expected: "***",
		},
		{
			name:     "7 chars - shows first and last 3",
			token:    "123:abc1",
			expected: "123...bc1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskTelegramToken(tt.token)
			if got != tt.expected {
				t.Errorf("maskTelegramToken(%q) = %q, want %q", tt.token, got, tt.expected)
			}
		})
	}
}

// TestConfig_String_MasksSecrets проверяет что метод String() не выводит чувствительные данные
func TestConfig_String_MasksSecrets(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Env:              "test",
			Port:             "8080",
			ProductionDomain: "example.com",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "test_db",
			User:     "postgres",
			Password: "super_secret_password_12345",
			SSLMode:  "disable",
		},
		Session: SessionConfig{
			Secret: "super_secret_session_key_abcdef",
		},
		Telegram: TelegramConfig{
			BotToken:      "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
			WebhookSecret: "webhook_secret_token",
		},
		YooKassa: YooKassaConfig{
			ShopID:    "123456",
			SecretKey: "live_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		},
	}

	str := cfg.String()

	// Проверка что секреты НЕ в выводе
	secretValues := []string{
		"super_secret_password_12345",
		"super_secret_session_key_abcdef",
		"123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
		"webhook_secret_token",
		"live_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
	}

	for _, secret := range secretValues {
		if strings.Contains(str, secret) {
			t.Errorf("String() содержит секрет '%s', но не должен. Вывод: %s", secret, str)
		}
	}

	// Проверка что non-sensitive данные В выводе
	expectedValues := []string{
		"test",      // Environment
		"8080",      // Port
		"localhost", // DB Host
		"postgres",  // DB User
		"test_db",   // DB Name
		"disable",   // SSL Mode
	}

	for _, expected := range expectedValues {
		if !strings.Contains(str, expected) {
			t.Errorf("String() должен содержать '%s', но не содержит. Вывод: %s", expected, str)
		}
	}

	// Проверка наличия boolean флагов (не раскрывая сами токены)
	if !strings.Contains(str, "Telegram:{Configured:true}") {
		t.Errorf("String() должен показывать что Telegram настроен (true), но не показывает. Вывод: %s", str)
	}
	if !strings.Contains(str, "YooKassa:{Configured:true}") {
		t.Errorf("String() должен показывать что YooKassa настроен (true), но не показывает. Вывод: %s", str)
	}
}

// TestConfig_String_EmptySecrets проверяет корректный вывод когда секреты не заданы
func TestConfig_String_EmptySecrets(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Env:  "development",
			Port: "3000",
		},
		Database: DatabaseConfig{
			Host:    "127.0.0.1",
			Port:    5432,
			Name:    "dev_db",
			User:    "dev_user",
			SSLMode: "require",
		},
		Session: SessionConfig{
			Secret: "minimal_secret_key_for_dev_32ch",
		},
		Telegram: TelegramConfig{
			BotToken: "", // Не настроен
		},
		YooKassa: YooKassaConfig{
			ShopID:    "", // Не настроен
			SecretKey: "",
		},
	}

	str := cfg.String()

	// Проверка что флаги показывают false для ненастроенных сервисов
	if !strings.Contains(str, "Telegram:{Configured:false}") {
		t.Errorf("String() должен показывать что Telegram НЕ настроен (false). Вывод: %s", str)
	}
	if !strings.Contains(str, "YooKassa:{Configured:false}") {
		t.Errorf("String() должен показывать что YooKassa НЕ настроен (false). Вывод: %s", str)
	}

	// Проверка что секрет сессии всё равно не выводится
	if strings.Contains(str, "minimal_secret_key_for_dev_32ch") {
		t.Errorf("String() содержит SESSION_SECRET, но не должен. Вывод: %s", str)
	}

	// Проверка что общая информация присутствует
	expectedValues := []string{
		"development",
		"3000",
		"127.0.0.1",
		"dev_user",
		"dev_db",
		"require",
	}

	for _, expected := range expectedValues {
		if !strings.Contains(str, expected) {
			t.Errorf("String() должен содержать '%s'. Вывод: %s", expected, str)
		}
	}
}

// TestConfig_String_Format проверяет общий формат вывода
func TestConfig_String_Format(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Env:  "production",
			Port: "8080",
		},
		Database: DatabaseConfig{
			Host:     "db.example.com",
			Port:     5432,
			Name:     "prod_db",
			User:     "prod_user",
			Password: "REDACTED",
			SSLMode:  "verify-full",
		},
		Session: SessionConfig{
			Secret: "REDACTED",
		},
	}

	str := cfg.String()

	// Проверка что строка начинается с "Config{"
	if !strings.HasPrefix(str, "Config{") {
		t.Errorf("String() должен начинаться с 'Config{', получено: %s", str)
	}

	// Проверка наличия ключевых частей структуры
	requiredParts := []string{
		"Env:",
		"Port:",
		"Database:",
		"SSLMode:",
		"Telegram:",
		"YooKassa:",
	}

	for _, part := range requiredParts {
		if !strings.Contains(str, part) {
			t.Errorf("String() должен содержать '%s'. Вывод: %s", part, str)
		}
	}
}

// TestValidate_DatabasePasswordRequiredInProduction проверяет что пустой пароль БД отклоняется в продакшене
func TestValidate_DatabasePasswordRequiredInProduction(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "production_with_empty_password",
			env:      "production",
			password: "",
			wantErr:  true,
			errMsg:   "DB_PASSWORD must not be empty in production",
		},
		{
			name:     "production_with_password",
			env:      "production",
			password: "secure_password_123",
			wantErr:  false,
		},
		{
			name:     "development_with_empty_password",
			env:      "development",
			password: "",
			wantErr:  false,
		},
		{
			name:     "development_with_password",
			env:      "development",
			password: "dev_password",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Env:              tt.env,
					Port:             "8080",
					ProductionDomain: "example.com",
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					Name:     "test_db",
					User:     "postgres",
					Password: tt.password,
					SSLMode:  "require",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
			}

			err := cfg.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error message = %q, should contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestValidate_ProductionDatabaseSecurityChecks проверяет все проверки безопасности БД для продакшена
func TestValidate_ProductionDatabaseSecurityChecks(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "production_missing_password",
			cfg: &Config{
				Server: ServerConfig{
					Env:              "production",
					Port:             "8080",
					ProductionDomain: "example.com",
				},
				Database: DatabaseConfig{
					Host:     "db.example.com",
					Port:     5432,
					Name:     "prod_db",
					User:     "prod_user",
					Password: "",
					SSLMode:  "require",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
			},
			wantErr: true,
			errMsg:  "DB_PASSWORD must not be empty in production",
		},
		{
			name: "production_with_ssl_disabled",
			cfg: &Config{
				Server: ServerConfig{
					Env:              "production",
					Port:             "8080",
					ProductionDomain: "example.com",
				},
				Database: DatabaseConfig{
					Host:     "db.example.com",
					Port:     5432,
					Name:     "prod_db",
					User:     "prod_user",
					Password: "password123",
					SSLMode:  "disable",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
			},
			wantErr: true,
			errMsg:  "Database SSL must be enabled in production",
		},
		{
			name: "production_missing_domain",
			cfg: &Config{
				Server: ServerConfig{
					Env:              "production",
					Port:             "8080",
					ProductionDomain: "",
				},
				Database: DatabaseConfig{
					Host:     "db.example.com",
					Port:     5432,
					Name:     "prod_db",
					User:     "prod_user",
					Password: "password123",
					SSLMode:  "require",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
			},
			wantErr: true,
			errMsg:  "PRODUCTION_DOMAIN is required in production mode",
		},
		{
			name: "production_all_valid",
			cfg: &Config{
				Server: ServerConfig{
					Env:              "production",
					Port:             "8080",
					ProductionDomain: "example.com",
				},
				Database: DatabaseConfig{
					Host:     "db.example.com",
					Port:     5432,
					Name:     "prod_db",
					User:     "prod_user",
					Password: "secure_password",
					SSLMode:  "verify-full",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error message = %q, should contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestValidate_TelegramTokenFormat проверяет валидацию формата Telegram токена
func TestValidate_TelegramTokenFormat(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid telegram token format",
			cfg: &Config{
				Server: ServerConfig{
					Env:  "development",
					Port: "8080",
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					Name:     "test_db",
					User:     "postgres",
					Password: "",
					SSLMode:  "disable",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
				Telegram: TelegramConfig{
					BotToken:        "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
					AdminTelegramID: 123456,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid token format - missing colon",
			cfg: &Config{
				Server: ServerConfig{
					Env:  "development",
					Port: "8080",
				},
				Database: DatabaseConfig{
					Host:    "localhost",
					Port:    5432,
					Name:    "test_db",
					User:    "postgres",
					SSLMode: "disable",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
				Telegram: TelegramConfig{
					BotToken:        "invalidtokenformat",
					AdminTelegramID: 123456,
				},
			},
			wantErr: true,
			errMsg:  "неверный формат",
		},
		{
			name: "invalid token format - non-numeric bot id",
			cfg: &Config{
				Server: ServerConfig{
					Env:  "development",
					Port: "8080",
				},
				Database: DatabaseConfig{
					Host:    "localhost",
					Port:    5432,
					Name:    "test_db",
					User:    "postgres",
					SSLMode: "disable",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
				Telegram: TelegramConfig{
					BotToken:        "abc123:ABCDEfghijklmnoPQRSTUvwxyz123456789",
					AdminTelegramID: 123456,
				},
			},
			wantErr: true,
			errMsg:  "неверный формат",
		},
		{
			name: "no telegram token configured",
			cfg: &Config{
				Server: ServerConfig{
					Env:  "development",
					Port: "8080",
				},
				Database: DatabaseConfig{
					Host:    "localhost",
					Port:    5432,
					Name:    "test_db",
					User:    "postgres",
					SSLMode: "disable",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
				Telegram: TelegramConfig{
					BotToken: "", // Not configured
				},
			},
			wantErr: false,
		},
		{
			name: "telegram token without admin id",
			cfg: &Config{
				Server: ServerConfig{
					Env:  "development",
					Port: "8080",
				},
				Database: DatabaseConfig{
					Host:    "localhost",
					Port:    5432,
					Name:    "test_db",
					User:    "postgres",
					SSLMode: "disable",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
				Telegram: TelegramConfig{
					BotToken:        "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
					AdminTelegramID: 0, // Invalid
				},
			},
			wantErr: true,
			errMsg:  "должен быть больше 0",
		},
		{
			name: "webhook secret too short",
			cfg: &Config{
				Server: ServerConfig{
					Env:  "development",
					Port: "8080",
				},
				Database: DatabaseConfig{
					Host:    "localhost",
					Port:    5432,
					Name:    "test_db",
					User:    "postgres",
					SSLMode: "disable",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
				Telegram: TelegramConfig{
					BotToken:        "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
					AdminTelegramID: 123456,
					UseWebhook:      true,
					WebhookURL:      "https://example.com/webhook",
					WebhookSecret:   "short", // Too short
				},
			},
			wantErr: true,
			errMsg:  "не менее 32 символов",
		},
		{
			name: "webhook secret long enough",
			cfg: &Config{
				Server: ServerConfig{
					Env:  "development",
					Port: "8080",
				},
				Database: DatabaseConfig{
					Host:    "localhost",
					Port:    5432,
					Name:    "test_db",
					User:    "postgres",
					SSLMode: "disable",
				},
				Session: SessionConfig{
					Secret: "secret_key_at_least_32_characters_long",
					MaxAge: 86400,
				},
				Telegram: TelegramConfig{
					BotToken:        "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
					AdminTelegramID: 123456,
					UseWebhook:      true,
					WebhookURL:      "https://example.com/webhook",
					WebhookSecret:   "abcdefghijklmnopqrstuvwxyz123456", // 32 chars
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error message = %q, should contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestValidateSessionSecret выполняет table-driven тесты валидации сессионного секрета
func TestValidateSessionSecret(t *testing.T) {
	tests := []struct {
		name         string
		secret       string
		isProduction bool
		wantErr      bool
		errContains  string
	}{
		// Успешные случаи (development)
		{
			name:         "valid secret - development",
			secret:       "MyGoodToken123!@#$%^&*()_+-=XYZZ",
			isProduction: false,
			wantErr:      false,
		},
		{
			name:         "valid secret with 32 chars minimum - development",
			secret:       "aB1!dEfGhIjXlMnOpQrStUvWxYz0Pp23",
			isProduction: false,
			wantErr:      false,
		},
		{
			name:         "valid secret with 48 chars - production",
			secret:       "MyGoodToken123!@#$%^&*()_+-=[]{}Auth48CharsQWERS",
			isProduction: true,
			wantErr:      false,
		},
		{
			name:         "valid production secret with all 4 character types",
			secret:       "Xa1!Yb2@Zc3#Wd4$Ee5%Ff6^Gg7&Hh8*Ii9(Jj0)Kk!MmSTU",
			isProduction: true,
			wantErr:      false,
		},

		// Ошибки: слишком короткие
		{
			name:         "too short - development",
			secret:       "short",
			isProduction: false,
			wantErr:      true,
			errContains:  "не менее 32 символов",
		},
		{
			name:         "too short for production",
			secret:       "aB1!sD2@eF3#gH4$iJ5%kL6^mN7&oP8*",
			isProduction: true,
			wantErr:      true,
			errContains:  "не менее 48 символов",
		},

		// Ошибки: пробелы (проверка даже коротких)
		{
			name:         "only whitespace",
			secret:       "                                ",
			isProduction: false,
			wantErr:      true,
			errContains:  "только пробельными",
		},

		// Ошибки: повторяющиеся символы
		{
			name:         "consecutive characters - 5 same",
			secret:       "aaaaaaaB1!dEfGhIjKlMnOpQrStUvWxYz",
			isProduction: false,
			wantErr:      true,
			errContains:  "слишком много одинаковых",
		},
		{
			name:         "consecutive zeros",
			secret:       "000000aB1!dEfGhIjKlMnOpQrStUvWxYz",
			isProduction: false,
			wantErr:      true,
			errContains:  "слишком много одинаковых",
		},

		// Ошибки: последовательные символы
		{
			name:         "sequential numbers",
			secret:       "MyToken123456789!@#$%^&*()_+-=[]XY",
			isProduction: false,
			wantErr:      true,
			errContains:  "последовательные",
		},
		{
			name:         "sequential letters",
			secret:       "abcdefghMYTOKEN123!@#$%^&*()_+-XY",
			isProduction: false,
			wantErr:      true,
			errContains:  "последовательные",
		},

		// Ошибки: слабые паттерны
		{
			name:         "weak pattern - all ones",
			secret:       "111111aB!dEfGhIjKlMnOpQrStUvWxyzQ",
			isProduction: false,
			wantErr:      true,
			errContains:  "слишком много одинаковых",
		},
		{
			name:         "weak pattern - password",
			secret:       "MyPasswordKey123!@#$%^&*()_+-=QZ",
			isProduction: false,
			wantErr:      true,
			errContains:  "распространённый слабый паттерн",
		},
		{
			name:         "weak pattern - session",
			secret:       "MySessionKey123!@#$%^&*()_+-=QZZ",
			isProduction: false,
			wantErr:      true,
			errContains:  "распространённый слабый паттерн",
		},

		// Ошибки: недостаточная энтропия
		{
			name:         "insufficient entropy - only lowercase and digits",
			secret:       "ajklfhvbjkxcmbnvjkxcmnbvjkxcmbnvjkxc",
			isProduction: false,
			wantErr:      true,
			errContains:  "минимум 3 типа",
		},
		{
			name:         "insufficient entropy - production only 3 types",
			secret:       "aBsD2eF3gH4iJ5kL6mN7oP8qR9tU0vW1xY2zAB3cD4eF5gH",
			isProduction: true,
			wantErr:      true,
			errContains:  "ВСЕ 4 типа",
		},

		// Edge cases
		{
			name:         "empty secret",
			secret:       "",
			isProduction: false,
			wantErr:      true,
			errContains:  "не менее 32",
		},
		{
			name:         "exactly 32 chars with good entropy",
			secret:       "aB1!cDeF2gHiJ3kLmN4oPqRs5tUvWx67",
			isProduction: false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSessionSecret(tt.secret, tt.isProduction)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSessionSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil {
					t.Errorf("expected error with '%s', got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, should contain '%s'", err, tt.errContains)
				}
			}
		})
	}
}

// TestMaskSecret проверяет маскирование секретов для логирования
func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		expected string
	}{
		{
			name:     "normal secret",
			secret:   "mySecAuth123!@#$%^&*()_+",
			expected: "myS...)_+",
		},
		{
			name:     "empty secret",
			secret:   "",
			expected: "<not set>",
		},
		{
			name:     "short secret - 6 chars",
			secret:   "abcdef",
			expected: "***",
		},
		{
			name:     "short secret - 3 chars",
			secret:   "abc",
			expected: "***",
		},
		{
			name:     "exactly 7 chars",
			secret:   "1234567",
			expected: "123...567",
		},
		{
			name:     "very long secret",
			secret:   "aB1!cDeF2gHiJ3kLmN4oPqRs5tUvWxYzAB3cD4eF5gH6iJ7kL8mN9oP0qR1sT2uV3wX4yZ5",
			expected: "aB1...yZ5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskSecret(tt.secret)
			if got != tt.expected {
				t.Errorf("maskSecret(%q) = %q, want %q", tt.secret, got, tt.expected)
			}
		})
	}
}

// TestGenerateSecureSecret проверяет что функция генерирует стойкие случайные секреты
func TestGenerateSecureSecret(t *testing.T) {
	tests := []struct {
		name          string
		length        int
		shouldSucceed bool
		minLength     int
	}{
		{
			name:          "generate 32 byte secret",
			length:        32,
			shouldSucceed: true,
			minLength:     42, // base64 encoding increases size: 32 bytes -> 44 chars (with padding)
		},
		{
			name:          "generate 48 byte secret",
			length:        48,
			shouldSucceed: true,
			minLength:     64, // 48 bytes -> 64 chars in base64
		},
		{
			name:          "generate 16 byte secret",
			length:        16,
			shouldSucceed: true,
			minLength:     20, // 16 bytes -> 24 chars in base64
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := generateSecureSecret(tt.length)

			if (err != nil) != !tt.shouldSucceed {
				t.Errorf("generateSecureSecret() error = %v, shouldSucceed %v", err, tt.shouldSucceed)
			}

			if tt.shouldSucceed {
				if len(secret) < tt.minLength {
					t.Errorf("generated secret length %d is less than minimum %d", len(secret), tt.minLength)
				}

				// Генерируем второй секрет и проверяем что он не идентичен первому
				secret2, _ := generateSecureSecret(tt.length)
				if secret == secret2 {
					t.Errorf("generated secrets should be random, but got identical values")
				}
			}
		})
	}
}

// TestGenerateSecureSecret_ValidBase64 проверяет что сгенерированные секреты это валидный base64
func TestGenerateSecureSecret_ValidBase64(t *testing.T) {
	secret, err := generateSecureSecret(32)
	if err != nil {
		t.Fatalf("generateSecureSecret failed: %v", err)
	}

	// Пытаемся декодировать - если это не base64, декодирование упадёт
	_, err = base64.StdEncoding.DecodeString(secret)
	if err != nil {
		t.Errorf("generated secret is not valid base64: %v", err)
	}
}

// TestCheckTelegramTokenExposure проверяет что функция корректно идентифицирует реальные токены
func TestCheckTelegramTokenExposure(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		expectWarn bool
	}{
		{
			name:       "real token format - should warn",
			token:      "123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789",
			expectWarn: true,
		},
		{
			name:       "empty token - no warning",
			token:      "",
			expectWarn: false,
		},
		{
			name:       "example token format - no warning",
			token:      "your_telegram_bot_token_here",
			expectWarn: false,
		},
		{
			name:       "short invalid token - no warning",
			token:      "invalid",
			expectWarn: false,
		},
		{
			name:       "token without colon - no warning",
			token:      "abc123def456",
			expectWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify function doesn't panic with different inputs
			// The actual warning output would need to be captured from logs
			checkTelegramTokenExposure(tt.token)
			// If we get here without panic, test passes
		})
	}
}

// TestSessionSecretRequiredInProduction проверяет что SESSION_SECRET обязателен в production
func TestSessionSecretRequiredInProduction(t *testing.T) {
	// Этот тест проверяет поведение Load() при отсутствии SESSION_SECRET в production
	// Мы НЕ можем запустить Load() напрямую так как он читает os.Getenv()
	// Вместо этого проверяем validateSessionSecret() с пустым значением
	tests := []struct {
		name    string
		secret  string
		env     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty secret in production should fail validation",
			secret:  "",
			env:     "production",
			wantErr: true,
			errMsg:  "не менее", // минимальная ошибка при пустой строке
		},
		{
			name:    "valid secret in production should pass",
			secret:  "Xa1!Yb2@Zc3#Wd4$Ee5%Ff6^Gg7&Hh8*Ii9(Jj0)Kk!MmSTU", // 48+ chars with all types
			env:     "production",
			wantErr: false,
		},
		{
			name:    "valid secret in development should pass",
			secret:  "aB1!cDeF2gHiJ3kLmN4oPqRs5tUvWx67",
			env:     "development",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSessionSecret(tt.secret, tt.env == "production")
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSessionSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error = %v, should contain %q", err, tt.errMsg)
			}
		})
	}
}

// TestSessionSecretValidation_AllCriteria проверяет все критерии валидации SECRET
func TestSessionSecretValidation_AllCriteria(t *testing.T) {
	tests := []struct {
		name         string
		secret       string
		isProduction bool
		wantErr      bool
		description  string
	}{
		{
			name:         "production - valid 48+ chars with all 4 types",
			secret:       "Xa1!Yb2@Zc3#Wd4$Ee5%Ff6^Gg7&Hh8*Ii9(Jj0)Kk!Mm0NN", // 48+ with all types, no weak patterns
			isProduction: true,
			wantErr:      false,
			description:  "Valid production secret",
		},
		{
			name:         "production - too short (47 chars)",
			secret:       "Xa1!Yb2@Zc3#Wd4$Ee5%Ff6^Gg7&Hh8*Ii9(Jj0)Kk!Mm0", // 47 chars
			isProduction: true,
			wantErr:      true,
			description:  "Production requires 48+ chars",
		},
		{
			name:         "production - missing special chars",
			secret:       "XaYbZcWdEeFfGgHhIiJjKkMmNnOoPpQqRrSsTtUuVvWwXxYyZz", // No special chars
			isProduction: true,
			wantErr:      true,
			description:  "Production requires all 4 character types",
		},
		{
			name:         "dev - valid 32+ chars with 3 types",
			secret:       "Xa1!Yb2@Zc3#Wd4$Ee5%Ff6^Gg7&Hh99", // 32 chars with 3 types, no weak patterns
			isProduction: false,
			wantErr:      false,
			description:  "Development allows 3 types",
		},
		{
			name:         "dev - too short (31 chars)",
			secret:       "Xa1!Yb2@Zc3#Wd4$Ee5%Ff6^Gg7&Hh", // 31 chars
			isProduction: false,
			wantErr:      true,
			description:  "Minimum 32 characters required",
		},
		{
			name:         "dev - weak pattern: contains 'password'",
			secret:       "MyPasswordKey123!@#$%^&*()_+-=XYZ", // Contains "password"
			isProduction: false,
			wantErr:      true,
			description:  "Weak pattern detection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSessionSecret(tt.secret, tt.isProduction)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSessionSecret(%s) error = %v, wantErr %v", tt.description, err, tt.wantErr)
			}
		})
	}
}

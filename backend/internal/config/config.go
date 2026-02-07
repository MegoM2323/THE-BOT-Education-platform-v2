package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Config содержит всю конфигурацию приложения
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Session  SessionConfig
	Telegram TelegramConfig
	YooKassa YooKassaConfig
}

// DatabaseConfig содержит конфигурацию подключения к базе данных
type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// ServerConfig содержит конфигурацию сервера
type ServerConfig struct {
	Port             string
	Env              string   // development, production
	ProductionDomain string   // Domain for production environment
	TrustedProxies   []string // Список доверенных прокси-серверов (для X-Forwarded-For)
}

// SessionConfig содержит конфигурацию управления сессиями
type SessionConfig struct {
	Secret   string
	MaxAge   time.Duration
	Secure   bool // Устанавливается в true для продакшена (только HTTPS)
	HTTPOnly bool
	SameSite string // "Strict", "Lax", "None"
}

// TelegramConfig содержит конфигурацию Telegram бота
type TelegramConfig struct {
	BotToken        string
	AdminTelegramID int64
	WebhookURL      string
	WebhookSecret   string // Secret token for webhook signature verification
	UseWebhook      bool
	ProxyURL        string // SOCKS5 proxy URL (например socks5://127.0.0.1:1080)
}

// maskTelegramToken маскирует токен для логирования (показывает только первые и последние символы)
// Формат маски: XXX...XXX где X это первые/последние 3 символа (при длине > 6)
// Для коротких токенов (<=6 символов) показывает полный token с маскировкой компонентов
func maskTelegramToken(token string) string {
	if token == "" {
		return "<not set>"
	}
	// Если токен слишком короткий (6 или менее), скрываем полностью
	if len(token) <= 6 {
		return "***"
	}
	// Показываем первые 3 и последние 3 символа
	return token[:3] + "..." + token[len(token)-3:]
}

// YooKassaConfig содержит конфигурацию YooKassa платежной системы
type YooKassaConfig struct {
	ShopID    string
	SecretKey string
	ReturnURL string
}

// isValidTelegramToken проверяет что токен соответствует формату Telegram бота
// Telegram токены имеют формат: <bot_id>:<token_string>
// Пример: 123456789:ABCDEfghijklmnoPQRSTUvwxyz123456789
func isValidTelegramToken(token string) bool {
	// Токен не должен быть пустым
	if token == "" {
		return false
	}

	// Проверяем наличие ':' разделителя
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return false
	}

	botID := parts[0]
	tokenStr := parts[1]

	// Bot ID должен быть числовым (только цифры)
	if botID == "" {
		return false
	}
	for _, ch := range botID {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	// Token string должен иметь минимальную длину (обычно 35+ символов)
	// и содержать только буквы, цифры, '-', '_'
	if len(tokenStr) < 20 {
		return false
	}

	for _, ch := range tokenStr {
		if !((ch >= 'A' && ch <= 'Z') ||
			(ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_') {
			return false
		}
	}

	return true
}

// validateSessionSecret выполняет комплексную валидацию сессионного секрета
// Проверяет:
// - Минимальную длину (32 байта)
// - Отсутствие слабых паттернов (повторяющиеся символы, последовательные символы)
// - Энтропию (разнообразие символов)
func validateSessionSecret(secret string, isProduction bool) error {
	const minLength = 32

	// Проверка минимальной длины
	if len(secret) < minLength {
		return fmt.Errorf("SESSION_SECRET должен быть не менее %d символов (текущая длина: %d)", minLength, len(secret))
	}

	// Проверка на пустой или все пробелы
	if strings.TrimSpace(secret) == "" {
		return fmt.Errorf("SESSION_SECRET не может быть только пробельными символами")
	}

	// Проверка на повторяющиеся символы (более 4 одинаковых подряд = weak pattern)
	// Пример: "aaaaaaa" или "1111111"
	for i := 0; i < len(secret)-4; i++ {
		if secret[i] == secret[i+1] && secret[i+1] == secret[i+2] &&
			secret[i+2] == secret[i+3] && secret[i+3] == secret[i+4] {
			return fmt.Errorf("SESSION_SECRET содержит слишком много одинаковых символов подряд (более 4 одинаковых)")
		}
	}

	// Проверка на полностью последовательные паттерны
	// Примеры: "123456789", "abcdefgh", "ABCDEFGH"
	sequentialPatterns := []string{
		"01234567", "12345678", "23456789", "34567890",
		"abcdefgh", "bcdefghi", "cdefghij", "defghijk", "efghijkl", "fghijklm",
		"ABCDEFGH", "BCDEFGHI", "CDEFGHIJ", "DEFGHIJK", "EFGHIJKL", "FGHIJKLM",
	}
	lowerSecretSeq := strings.ToLower(secret)
	for _, pattern := range sequentialPatterns {
		if strings.Contains(lowerSecretSeq, pattern) {
			return fmt.Errorf("SESSION_SECRET содержит последовательные символы (например, '123456' или 'abcdef')")
		}
	}

	// Проверка на распространённые слабые паттерны
	weakPatterns := []string{
		"000000", "111111", "222222", "333333", "444444", "555555", "666666", "777777", "888888", "999999",
		"password", "secret", "key", "session", "test123", "admin123",
	}
	lowerSecret := strings.ToLower(secret)
	for _, pattern := range weakPatterns {
		if strings.Contains(lowerSecret, pattern) {
			return fmt.Errorf("SESSION_SECRET содержит распространённый слабый паттерн: '%s'", pattern)
		}
	}

	// Проверка энтропии (минимальное разнообразие символов)
	// Должны быть буквы, цифры и спецсимволы
	var hasLower, hasUpper, hasDigit, hasSpecial bool
	specialChars := "!@#$%^&*()_+-=[]{};\\'\":\\|,.<>?/~`"

	for _, ch := range secret {
		switch {
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		case strings.ContainsRune(specialChars, ch):
			hasSpecial = true
		}
	}

	characterTypeCount := 0
	if hasLower {
		characterTypeCount++
	}
	if hasUpper {
		characterTypeCount++
	}
	if hasDigit {
		characterTypeCount++
	}
	if hasSpecial {
		characterTypeCount++
	}

	// Требуем минимум 3 типа символов для достаточной энтропии
	if characterTypeCount < 3 {
		return fmt.Errorf("SESSION_SECRET должен содержать минимум 3 типа символов (прописные, строчные, цифры, спецсимволы). Текущих типов: %d", characterTypeCount)
	}

	// В production режиме требуем более строгие критерии
	if isProduction {
		// В production требуем все 4 типа символов
		if characterTypeCount < 4 {
			return fmt.Errorf("SESSION_SECRET в production должен содержать ВСЕ 4 типа символов (прописные, строчные, цифры, спецсимволы). Текущих типов: %d", characterTypeCount)
		}
		// В production требуем минимум 48 символов для дополнительной безопасности
		const productionMinLength = 48
		if len(secret) < productionMinLength {
			return fmt.Errorf("SESSION_SECRET в production должен быть не менее %d символов (текущая длина: %d)", productionMinLength, len(secret))
		}
	}

	return nil
}

// maskSecret маскирует секрет для безопасного логирования
func maskSecret(secret string) string {
	if secret == "" {
		return "<not set>"
	}
	if len(secret) <= 6 {
		return "***"
	}
	// Показываем первые 3 и последние 3 символа
	return secret[:3] + "..." + secret[len(secret)-3:]
}

// generateSecureSecret генерирует криптографически стойкий случайный секрет
// length - требуемая длина секрета в байтах (рекомендуется 32 для dev, 48 для production)
// Возвращает base64-кодированную строку
func generateSecureSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// checkTelegramTokenExposure проверяет потенциальное раскрытие токена Telegram
// Проверяет общие способы утечки: логи, сохранение в файлах, коммиты в git
// CRITICAL: Telegram bot token должен быть защищен как пароль
func checkTelegramTokenExposure(token string) {
	if token == "" {
		return
	}

	// Проверяем на наличие токена в типичных местах утечки
	// Эта проверка предупреждает разработчика, если они случайно записали реальный токен

	// Pattern для обнаружения реальных Telegram токенов (не примеров)
	// Пример токена: 123456789:ABCD...
	isRealToken := regexp.MustCompile(`^\d{7,10}:[A-Za-z0-9_-]{25,}$`).MatchString(token)

	// Проверяем на наличие типичных паттернов сохранения
	if isRealToken {
		// Проверяем на наличие токена в переменной окружения (OK)
		// Но если токен выглядит реальным, выдаём предупреждение в логи
		log.Println("[SECURITY WARNING] Telegram bot token is configured. Ensure:")
		log.Println("  1. Token is loaded from secure vault or environment variable")
		log.Println("  2. Token is NEVER committed to version control (.gitignore check)")
		log.Println("  3. Token has minimal required permissions (send messages only)")
		log.Println("  4. Token rotation is planned (recommend rotation every 90 days)")
	}
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	// Загружаем конфигурацию базы данных
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("некорректный DB_PORT: %w", err)
	}

	// Загружаем максимальный срок действия сессии (в секундах)
	// По умолчанию 7 дней (604800 секунд) для удобства пользователей
	sessionMaxAgeSeconds, err := strconv.Atoi(getEnv("SESSION_MAX_AGE", "604800"))
	if err != nil {
		return nil, fmt.Errorf("некорректный SESSION_MAX_AGE: %w", err)
	}

	// Определяем окружение
	env := getEnv("ENV", "development")
	isProduction := env == "production"

	// Проверяем секретный ключ сессии
	sessionSecret := getEnv("SESSION_SECRET", "")

	if sessionSecret == "" {
		if isProduction {
			// В production режиме обязательно требуем явно установленный секрет
			return nil, fmt.Errorf("CRITICAL SECURITY: SESSION_SECRET is required in production. " +
				"Generate with: openssl rand -base64 48 (48 bytes for production, 32+ for dev). " +
				"Set unique secret for each production environment")
		}

		// В development режиме генерируем случайный секрет с предупреждением
		log.Println("[SECURITY WARNING] SESSION_SECRET not set in development. Generating temporary random secret.")
		log.Println("For consistent development sessions, set SESSION_SECRET in .env")
		log.Println("Generate with: openssl rand -base64 48")

		generatedSecret, err := generateSecureSecret(32) // 32 bytes для development
		if err != nil {
			return nil, fmt.Errorf("failed to generate SESSION_SECRET: %w", err)
		}
		sessionSecret = generatedSecret
		log.Printf("[SECURITY WARNING] Generated temporary SESSION_SECRET: %s (save this to .env if you want persistent sessions)\n", maskSecret(sessionSecret))
	}

	// Валидируем сессионный секрет с учётом окружения
	if err := validateSessionSecret(sessionSecret, isProduction); err != nil {
		return nil, fmt.Errorf("SESSION_SECRET validation failed: %w", err)
	}

	// Загружаем конфигурацию Telegram
	telegramBotToken := getEnv("TELEGRAM_BOT_TOKEN", "")
	adminTelegramID, _ := strconv.ParseInt(getEnv("ADMIN_TELEGRAM_ID", "0"), 10, 64)
	useWebhook := getEnv("TELEGRAM_USE_WEBHOOK", "false") == "true"

	// CRITICAL SECURITY: Проверяем потенциальное раскрытие токена
	checkTelegramTokenExposure(telegramBotToken)

	// Determine default SameSite based on environment
	defaultSameSite := "Lax"
	if isProduction {
		defaultSameSite = "Strict"
	}

	// Загружаем доверенные прокси-серверы (разделены запятыми)
	trustedProxies := []string{}
	if proxiesStr := getEnv("TRUSTED_PROXIES", ""); proxiesStr != "" {
		for _, proxy := range strings.Split(proxiesStr, ",") {
			if trimmed := strings.TrimSpace(proxy); trimmed != "" {
				trustedProxies = append(trustedProxies, trimmed)
			}
		}
	}
	// По умолчанию доверяем localhost для development
	if len(trustedProxies) == 0 && !isProduction {
		trustedProxies = []string{"127.0.0.1", "localhost", "::1"}
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			Name:     getEnv("DB_NAME", "tutoring_platform"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSL_MODE", "require"),
		},
		Server: ServerConfig{
			Port:             getEnv("SERVER_PORT", "8080"),
			Env:              env,
			ProductionDomain: getEnv("PRODUCTION_DOMAIN", ""),
			TrustedProxies:   trustedProxies,
		},
		Session: SessionConfig{
			Secret:   sessionSecret,
			MaxAge:   time.Duration(sessionMaxAgeSeconds) * time.Second,
			Secure:   isProduction, // Безопасно только в продакшене
			HTTPOnly: true,
			SameSite: getEnv("SESSION_SAME_SITE", defaultSameSite),
		},
		Telegram: TelegramConfig{
			BotToken:        telegramBotToken,
			AdminTelegramID: adminTelegramID,
			WebhookURL:      getEnv("TELEGRAM_WEBHOOK_URL", ""),
			WebhookSecret:   getEnv("TELEGRAM_WEBHOOK_SECRET", ""),
			UseWebhook:      useWebhook,
			ProxyURL:        getEnv("TELEGRAM_PROXY", ""),
		},
		YooKassa: YooKassaConfig{
			ShopID:    getEnv("YOOKASSA_SHOP_ID", ""),
			SecretKey: getEnv("YOOKASSA_SECRET_KEY", ""),
			ReturnURL: getEnv("YOOKASSA_RETURN_URL", "http://localhost:5173/payment-success"),
		},
	}

	// Валидируем конфигурацию
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("некорректная конфигурация: %w", err)
	}

	return config, nil
}

// Validate выполняет валидацию конфигурации
func (c *Config) Validate() error {
	// Валидируем конфигурацию базы данных
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST обязателен")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME обязательно")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER обязателен")
	}

	// Database password validation - CRITICAL SECURITY CHECK
	if c.IsProduction() {
		// CRITICAL: In production, password MUST NOT be empty
		// Empty password allows unauthenticated database access
		if c.Database.Password == "" {
			return fmt.Errorf("CRITICAL SECURITY: DB_PASSWORD must not be empty in production. Empty password allows unauthorized database access")
		}
	}

	// Database safety checks
	if c.IsDevelopment() {
		// In development, only allow localhost databases or Docker service name "postgres"
		if c.Database.Host != "localhost" && c.Database.Host != "127.0.0.1" && c.Database.Host != "postgres" {
			return fmt.Errorf("SAFETY: Cannot connect to remote database %s in development mode. Use localhost or Docker service name only", c.Database.Host)
		}
	}

	if c.IsProduction() {
		// In production, require production domain
		if c.Server.ProductionDomain == "" {
			return fmt.Errorf("PRODUCTION_DOMAIN is required in production mode")
		}
	}

	// Валидируем конфигурацию сервера
	if c.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT обязателен")
	}

	// Валидируем конфигурацию сессии
	if c.Session.Secret == "" {
		return fmt.Errorf("SESSION_SECRET обязателен")
	}
	// SESSION_SECRET уже валидирован в Load(), здесь только проверяем MaxAge
	if c.Session.MaxAge <= 0 {
		return fmt.Errorf("SESSION_MAX_AGE должен быть больше 0")
	}

	// Валидируем конфигурацию Telegram (если бот включен)
	if c.Telegram.BotToken != "" {
		// Валидируем формат токена Telegram
		if !isValidTelegramToken(c.Telegram.BotToken) {
			maskedToken := maskTelegramToken(c.Telegram.BotToken)
			return fmt.Errorf("TELEGRAM_BOT_TOKEN имеет неверный формат (token: %s)", maskedToken)
		}
		if c.Telegram.AdminTelegramID <= 0 {
			return fmt.Errorf("ADMIN_TELEGRAM_ID должен быть больше 0")
		}
		if c.Telegram.UseWebhook && c.Telegram.WebhookURL == "" {
			return fmt.Errorf("TELEGRAM_WEBHOOK_URL обязателен при использовании webhook")
		}
		// CRITICAL SECURITY: Webhook secret MUST be set and have minimum length
		// Empty secret allows ANY attacker to forge Telegram webhook calls
		if c.Telegram.UseWebhook {
			if c.Telegram.WebhookSecret == "" {
				return fmt.Errorf("CRITICAL SECURITY: TELEGRAM_WEBHOOK_SECRET обязателен при использовании webhook. Пустой секрет позволяет подделать вызовы webhook")
			}
			if len(c.Telegram.WebhookSecret) < 32 {
				return fmt.Errorf("CRITICAL SECURITY: TELEGRAM_WEBHOOK_SECRET должен быть не менее 32 символов для безопасности (текущая длина: %d)", len(c.Telegram.WebhookSecret))
			}
		}
	}

	return nil
}

// GetDSN возвращает строку подключения PostgreSQL
func (c *DatabaseConfig) GetDSN() string {
	// Строим DSN, пропуская пароль если он пустой (для peer auth)
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Name,
		c.SSLMode,
	)

	// Добавляем пароль только если он не пустой
	if c.Password != "" {
		dsn = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host,
			c.Port,
			c.User,
			c.Password,
			c.Name,
			c.SSLMode,
		)
	}

	return dsn
}

// IsProduction возвращает true, если окружение - продакшен
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// IsDevelopment возвращает true, если окружение - разработка
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// GetBaseURL возвращает базовый URL для приложения
func (c *Config) GetBaseURL() string {
	if c.IsProduction() && c.Server.ProductionDomain != "" {
		return "https://" + c.Server.ProductionDomain
	}
	return "http://localhost:" + c.Server.Port
}

// GetWebhookURL возвращает URL для Telegram webhook
func (c *Config) GetWebhookURL() string {
	if c.IsProduction() && c.Server.ProductionDomain != "" {
		domain := c.Server.ProductionDomain
		// Remove https:// prefix if already present (from deploy script)
		if strings.HasPrefix(domain, "https://") {
			domain = strings.TrimPrefix(domain, "https://")
		} else if strings.HasPrefix(domain, "http://") {
			domain = strings.TrimPrefix(domain, "http://")
		}
		return fmt.Sprintf("https://%s/api/v1/telegram/webhook", domain)
	}
	return c.Telegram.WebhookURL
}

// String возвращает строковое представление конфигурации с маскировкой секретов
// ВАЖНО: никогда не логирует актуальные значения секретов (пароли, токены, ключи)
func (c *Config) String() string {
	maskSecret := func(secret string) string {
		if secret == "" {
			return "<not set>"
		}
		return "***"
	}

	telegramConfigured := c.Telegram.BotToken != ""
	yookassaConfigured := c.YooKassa.ShopID != "" && c.YooKassa.SecretKey != ""

	return fmt.Sprintf(
		"Config{Database:{Host:%s Port:%d Name:%s User:%s Password:%s SSLMode:%s} "+
			"Server:{Port:%s Env:%s ProductionDomain:%s} "+
			"Session:{Secret:%s MaxAge:%v Secure:%v HTTPOnly:%v SameSite:%s} "+
			"Telegram:{Configured:%v} "+
			"YooKassa:{Configured:%v}}",
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.User,
		maskSecret(c.Database.Password),
		c.Database.SSLMode,
		c.Server.Port,
		c.Server.Env,
		c.Server.ProductionDomain,
		maskSecret(c.Session.Secret),
		c.Session.MaxAge,
		c.Session.Secure,
		c.Session.HTTPOnly,
		c.Session.SameSite,
		telegramConfigured,
		yookassaConfigured,
	)
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

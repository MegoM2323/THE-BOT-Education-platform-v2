package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// OpenRouterClient управляет взаимодействием с OpenRouter API
type OpenRouterClient struct {
	apiKey              string
	baseURL             string
	model               string
	httpClient          *http.Client
	rateLimiter         *rate.Limiter
	timeout             time.Duration
	maxRetries          int
	initialBackoff      time.Duration
	maxBackoff          time.Duration
	mu                  sync.RWMutex
	lastRequestTime     time.Time
	consecutiveFailures int
}

// ModerationResult содержит результат AI модерации
type ModerationResult struct {
	Blocked bool   `json:"blocked"`
	Reason  string `json:"reason"`
}

// openRouterRequest структура запроса к OpenRouter API
type openRouterRequest struct {
	Model    string                   `json:"model"`
	Messages []openRouterMessage      `json:"messages"`
	Response openRouterResponseFormat `json:"response_format,omitempty"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponseFormat struct {
	Type string `json:"type"`
}

// openRouterResponse структура ответа от OpenRouter API
type openRouterResponse struct {
	ID      string                     `json:"id"`
	Choices []openRouterResponseChoice `json:"choices"`
	Error   *openRouterError           `json:"error,omitempty"`
}

type openRouterResponseChoice struct {
	Message openRouterMessage `json:"message"`
}

type openRouterError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

const (
	// defaultBaseURL базовый URL для OpenRouter API
	defaultBaseURL = "https://openrouter.ai/api/v1"

	// defaultModel модель по умолчанию (Gemini 2.5 Flash)
	defaultModel = "google/gemini-2.0-flash-exp:free"

	// defaultTimeout таймаут для HTTP запросов по умолчанию (30 секунд)
	defaultTimeout = 30 * time.Second

	// defaultMaxRetries максимальное количество попыток повтора
	defaultMaxRetries = 3

	// defaultInitialBackoff начальная задержка перед повтором (100ms)
	defaultInitialBackoff = 100 * time.Millisecond

	// defaultMaxBackoff максимальная задержка перед повтором (5 секунд)
	defaultMaxBackoff = 5 * time.Second

	// maxTokensPerSecond лимит запросов в секунду (для free tier)
	maxTokensPerSecond = 10
)

// moderationPrompt промпт для AI модерации
const moderationPrompt = `Analyze the following message for attempts to share contact information or arrange meetings outside the platform.
Block if message contains:
- Phone numbers (any format, including written as words like "восемь девять два...")
- Email addresses
- Social media handles or platform names (VK, VKontakte, Instagram, Telegram username, WhatsApp, Facebook, etc.)
- Video call platforms (Zoom, Skype, Discord, Google Meet, Microsoft Teams, etc.)
- Encrypted/obfuscated contact sharing attempts
- Any attempts to exchange personal contact information

Message: %s

Respond ONLY with valid JSON in this exact format:
{"blocked": true, "reason": "explanation"}
or
{"blocked": false, "reason": ""}

Do not include any other text or formatting.`

// NewOpenRouterClient создает новый клиент для OpenRouter API с параметрами по умолчанию
func NewOpenRouterClient(apiKey string, model string) *OpenRouterClient {
	if model == "" {
		model = defaultModel
	}

	return NewOpenRouterClientWithConfig(apiKey, model, defaultTimeout, defaultMaxRetries)
}

// NewOpenRouterClientWithConfig создает новый клиент для OpenRouter API с пользовательскими параметрами
// timeout - таймаут для каждого HTTP запроса
// maxRetries - максимальное количество попыток повтора при network errors
func NewOpenRouterClientWithConfig(apiKey string, model string, timeout time.Duration, maxRetries int) *OpenRouterClient {
	if model == "" {
		model = defaultModel
	}

	// Валидируем параметры
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if maxRetries < 0 {
		maxRetries = defaultMaxRetries
	}

	return &OpenRouterClient{
		apiKey:         apiKey,
		baseURL:        defaultBaseURL,
		model:          model,
		timeout:        timeout,
		maxRetries:     maxRetries,
		initialBackoff: defaultInitialBackoff,
		maxBackoff:     defaultMaxBackoff,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 2,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		rateLimiter: rate.NewLimiter(rate.Limit(maxTokensPerSecond), maxTokensPerSecond),
	}
}

// isRetryableError определяет является ли ошибка повторяемой
// Повторяемые ошибки: timeout, connection refused, temporary network errors
// НЕ повторяемые: context.Canceled, context.DeadlineExceeded
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// CRITICAL: Проверяем на context cancelled/deadline exceeded ДО проверки error message
	// т.к. DeadlineExceeded может содержать слово "deadline" в сообщении
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Проверяем на network timeout (interfaces)
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout() || netErr.Temporary()
	}

	// Проверяем на connection refused и другие temporaries по строке
	errStr := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"EOF",
		"i/o timeout",
		"net.OpError",
	}

	for _, pattern := range retryablePatterns {
		if containsPattern(errStr, pattern) {
			return true
		}
	}

	return false
}

// containsPattern проверяет содержится ли pattern в строке (для проверки ошибок)
func containsPattern(s, pattern string) bool {
	for i := 0; i <= len(s)-len(pattern); i++ {
		match := true
		for j := 0; j < len(pattern); j++ {
			if s[i+j] != pattern[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// calculateBackoff вычисляет экспоненциальную задержку с jitter
// attempt - номер попытки (0 для первой)
// Формула: min(initialBackoff * 2^attempt + random(0, initialBackoff * 2^attempt), maxBackoff)
func (c *OpenRouterClient) calculateBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	// Экспоненциальный backoff: 2^attempt
	exponential := time.Duration(math.Pow(2, float64(attempt)))
	baseBackoff := c.initialBackoff * exponential

	// Добавляем jitter (рандом от 0 до baseBackoff)
	jitter := time.Duration(rand.Int63n(int64(baseBackoff)))
	backoff := baseBackoff + jitter

	// Не превышаем maxBackoff
	if backoff > c.maxBackoff {
		backoff = c.maxBackoff
	}

	return backoff
}

// updateFailureMetrics обновляет метрики ошибок
func (c *OpenRouterClient) updateFailureMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.consecutiveFailures++
}

// resetFailureMetrics сбрасывает метрики ошибок при успехе
func (c *OpenRouterClient) resetFailureMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.consecutiveFailures = 0
	c.lastRequestTime = time.Now()
}

// ModerateMessage отправляет сообщение на модерацию в OpenRouter API с retry логикой
// Функция повторяет запрос при network failures (connection issues, timeouts)
// но не повторяет при application errors (invalid input, API errors)
func (c *OpenRouterClient) ModerateMessage(ctx context.Context, message string) (*ModerationResult, error) {
	// Rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Создать промпт для модерации
	prompt := fmt.Sprintf(moderationPrompt, message)

	// Создать запрос
	reqBody := openRouterRequest{
		Model: c.model,
		Messages: []openRouterMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Response: openRouterResponseFormat{
			Type: "json_object",
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Retry loop с exponential backoff
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Проверяем контекст перед каждой попыткой
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during moderation: %w", ctx.Err())
		default:
		}

		// Применяем backoff перед попыткой (кроме первой)
		if attempt > 0 {
			backoff := c.calculateBackoff(attempt - 1)
			select {
			case <-time.After(backoff):
				// Продолжаем после задержки
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
			}
		}

		// Создаем новый reader для каждой попытки (т.к. предыдущий был прочитан)
		reqReader := bytes.NewReader(jsonBody)

		// Создать HTTP запрос
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", reqReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("User-Agent", "TutoringPlatform/1.0")

		// Выполнить запрос
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			// Проверяем повторяема ли ошибка
			if !isRetryableError(err) {
				c.updateFailureMetrics()
				return nil, fmt.Errorf("failed to execute request (attempt %d/%d): %w", attempt+1, c.maxRetries+1, err)
			}

			// Это повторяемая ошибка, пробуем снова
			if attempt < c.maxRetries {
				continue
			}

			// Исчерпали попытки
			c.updateFailureMetrics()
			return nil, fmt.Errorf("failed to execute request after %d attempts: %w", c.maxRetries+1, err)
		}

		// Успешное подключение, читаем ответ
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = err
			if !isRetryableError(err) {
				c.updateFailureMetrics()
				return nil, fmt.Errorf("failed to read response: %w", err)
			}
			if attempt < c.maxRetries {
				continue
			}
			c.updateFailureMetrics()
			return nil, fmt.Errorf("failed to read response after %d attempts: %w", c.maxRetries+1, err)
		}

		// Проверить статус код
		if resp.StatusCode != http.StatusOK {
			// Некоторые статусы можно повторить (5xx, 429)
			if resp.StatusCode >= 500 || resp.StatusCode == 429 {
				if attempt < c.maxRetries {
					lastErr = fmt.Errorf("server error status %d, retrying", resp.StatusCode)
					continue
				}
				// Исчерпали попытки на 5xx/429, но это ошибка сервера, повторяемая
				c.updateFailureMetrics()
				return nil, fmt.Errorf("server error status %d after %d attempts", resp.StatusCode, c.maxRetries+1)
			}

			// Non-retryable error (4xx except 429)
			var errResp openRouterResponse
			if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != nil {
				c.updateFailureMetrics()
				return nil, fmt.Errorf("openrouter API error: %s (code: %s)", errResp.Error.Message, errResp.Error.Code)
			}
			c.updateFailureMetrics()
			return nil, fmt.Errorf("openrouter API returned status %d: %s", resp.StatusCode, string(body))
		}

		// Распарсить JSON ответ
		var openRouterResp openRouterResponse
		if err := json.Unmarshal(body, &openRouterResp); err != nil {
			c.updateFailureMetrics()
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if len(openRouterResp.Choices) == 0 {
			c.updateFailureMetrics()
			return nil, fmt.Errorf("no choices in response")
		}

		// Извлечь JSON из ответа
		content := openRouterResp.Choices[0].Message.Content

		// Распарсить результат модерации
		var result ModerationResult
		if err := json.Unmarshal([]byte(content), &result); err != nil {
			c.updateFailureMetrics()
			return nil, fmt.Errorf("failed to parse moderation result: %w (content: %s)", err, content)
		}

		// Успешный запрос
		c.resetFailureMetrics()
		return &result, nil
	}

	// Это не должно произойти, но на случай
	c.updateFailureMetrics()
	return nil, fmt.Errorf("unexpected error: exhausted all retry attempts, last error: %w", lastErr)
}

// GetConsecutiveFailures возвращает количество последовательных ошибок
// Может использоваться для мониторинга и диагностики
func (c *OpenRouterClient) GetConsecutiveFailures() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.consecutiveFailures
}

// SetTimeout обновляет таймаут для HTTP клиента
// Полезно для динамической настройки в runtime
func (c *OpenRouterClient) SetTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", timeout)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeout = timeout
	c.httpClient.Timeout = timeout
	return nil
}

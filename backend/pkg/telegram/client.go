package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/proxy"
)

// Client представляет HTTP клиент для работы с Telegram Bot API
type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
	proxyURL   string // SOCKS5 proxy URL

	// Polling state
	pollingMu     sync.Mutex
	pollingCancel context.CancelFunc
	pollingWg     sync.WaitGroup
}

// TelegramError представляет ошибку Telegram API
type TelegramError struct {
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

// Error реализует интерфейс error для TelegramError
func (e *TelegramError) Error() string {
	return fmt.Sprintf("telegram API error %d: %s", e.ErrorCode, e.Description)
}

// APIResponse представляет стандартный ответ от Telegram API
type APIResponse struct {
	Ok          bool            `json:"ok"`
	Result      json.RawMessage `json:"result,omitempty"`
	ErrorCode   int             `json:"error_code,omitempty"`
	Description string          `json:"description,omitempty"`
}

// NewClient создает новый клиент для работы с Telegram Bot API
func NewClient(token string) *Client {
	return NewClientWithProxy(token, "")
}

// NewClientWithProxy создает клиент с поддержкой SOCKS5 proxy
// proxyURL формат: socks5://user:password@host:port или socks5://host:port
func NewClientWithProxy(token, proxyURL string) *Client {
	transport := createTransport(proxyURL)

	return &Client{
		token:    token,
		baseURL:  fmt.Sprintf("https://api.telegram.org/bot%s/", token),
		proxyURL: proxyURL,
		httpClient: &http.Client{
			Timeout:   30 * time.Second, // Увеличен для работы через proxy
			Transport: transport,
		},
	}
}

// createTransport создает http.Transport с опциональной поддержкой SOCKS5 proxy
func createTransport(proxyURL string) *http.Transport {
	if proxyURL == "" {
		// Прямое подключение без proxy
		return &http.Transport{
			Proxy: nil,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}
	}

	// Парсим URL proxy
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		log.Error().Err(err).Str("proxy_url", proxyURL).Msg("Некорректный URL proxy, используем прямое соединение")
		return &http.Transport{Proxy: nil}
	}

	// Создаем SOCKS5 dialer
	var auth *proxy.Auth
	if parsedURL.User != nil {
		password, _ := parsedURL.User.Password()
		auth = &proxy.Auth{
			User:     parsedURL.User.Username(),
			Password: password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", parsedURL.Host, auth, proxy.Direct)
	if err != nil {
		log.Error().Err(err).Str("proxy_url", proxyURL).Msg("Не удалось создать SOCKS5 dialer, используем прямое соединение")
		return &http.Transport{Proxy: nil}
	}

	// Создаем transport с SOCKS5 dialer
	contextDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		log.Error().Msg("SOCKS5 dialer не поддерживает DialContext, используем прямое соединение")
		return &http.Transport{Proxy: nil}
	}

	log.Info().Str("proxy", parsedURL.Host).Msg("Telegram клиент использует SOCKS5 proxy")

	return &http.Transport{
		DialContext: contextDialer.DialContext,
	}
}

// doRequest выполняет HTTP запрос к Telegram API с обработкой ошибок и retry логикой
func (c *Client) doRequest(method string, payload interface{}) (*APIResponse, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + method
	maxRetries := 3
	baseDelay := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodPost, url, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		// Обработка rate limiting (429 Too Many Requests)
		if resp.StatusCode == http.StatusTooManyRequests {
			if attempt < maxRetries-1 {
				// Exponential backoff
				delay := baseDelay * time.Duration(1<<attempt)
				time.Sleep(delay)
				continue
			}
			return nil, &TelegramError{
				ErrorCode:   429,
				Description: "Too many requests, retry limit exceeded",
			}
		}

		// Парсинг ответа
		var apiResp APIResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		// Обработка ошибок API
		if !apiResp.Ok {
			return nil, &TelegramError{
				ErrorCode:   apiResp.ErrorCode,
				Description: apiResp.Description,
			}
		}

		// Обработка HTTP статусов
		if resp.StatusCode >= 400 {
			return nil, &TelegramError{
				ErrorCode:   resp.StatusCode,
				Description: fmt.Sprintf("HTTP error: %s", resp.Status),
			}
		}

		return &apiResp, nil
	}

	return nil, fmt.Errorf("max retries exceeded")
}

// SendMessage отправляет текстовое сообщение в указанный чат
func (c *Client) SendMessage(chatID int64, text string) error {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	_, err := c.doRequest("sendMessage", payload)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// MediaFile представляет файл для отправки в медиа группе
type MediaFile struct {
	FileName string // Отображаемое имя файла
	FilePath string // Путь к файлу на диске
}

// SendDocument отправляет файл (документ) в указанный чат через multipart/form-data
// fileName - имя файла для отображения в Telegram
// filePath - полный путь к файлу на диске
func (c *Client) SendDocument(chatID int64, fileName, filePath string) error {
	// Telegram API требует multipart/form-data для загрузки файлов
	// Используем упрощенный подход с retry логикой
	maxRetries := 3
	baseDelay := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := c.sendDocumentAttempt(chatID, fileName, filePath)
		if err == nil {
			return nil
		}

		// Проверяем тип ошибки
		if telegramErr, ok := err.(*TelegramError); ok {
			switch telegramErr.ErrorCode {
			case 403, 400:
				// Forbidden или Bad Request - не ретраим
				return err
			case 429:
				// Too Many Requests - exponential backoff
				if attempt < maxRetries-1 {
					delay := baseDelay * time.Duration(1<<attempt)
					log.Info().Dur("backoff", delay).Msg("Rate limit hit, backing off")
					time.Sleep(delay)
					continue
				}
			}
		}

		// Для других ошибок (network errors) - retry
		if attempt < maxRetries-1 {
			time.Sleep(baseDelay)
			continue
		}

		return fmt.Errorf("max retries exceeded: %w", err)
	}

	return fmt.Errorf("failed to send document after %d attempts", maxRetries)
}

// SendMediaGroup отправляет группу документов (до 10 файлов) в указанный чат
// Telegram API позволяет отправить до 10 медиа файлов одним запросом
// Возвращает количество успешно отправленных файлов и ошибку
func (c *Client) SendMediaGroup(chatID int64, files []MediaFile) (int, error) {
	if len(files) == 0 {
		return 0, nil
	}

	if len(files) > 10 {
		return 0, fmt.Errorf("слишком много файлов: %d (максимум 10)", len(files))
	}

	// Telegram sendMediaGroup требует JSON структуру с массивом InputMedia
	// Для документов используем InputMediaDocument
	// Но проблема: нельзя загрузить файлы через JSON - нужен multipart/form-data
	// Решение: отправляем файлы по одному с retry логикой
	// При partial failure продолжаем отправку остальных файлов

	maxRetries := 3
	baseDelay := time.Second
	successCount := 0
	var lastError error

	for i, file := range files {
		var sent bool
		for attempt := 0; attempt < maxRetries; attempt++ {
			err := c.sendDocumentAttempt(chatID, file.FileName, file.FilePath)
			if err == nil {
				successCount++
				sent = true
				log.Info().
					Int("file_index", i+1).
					Int("total_files", len(files)).
					Str("file_name", file.FileName).
					Msg("Файл успешно отправлен в медиа группе")
				break
			}

			lastError = err

			// Проверяем тип ошибки
			if telegramErr, ok := err.(*TelegramError); ok {
				switch telegramErr.ErrorCode {
				case 403:
					// Forbidden - бот заблокирован, прерываем всю отправку
					log.Error().
						Int64("chat_id", chatID).
						Str("file_name", file.FileName).
						Msg("Бот заблокирован пользователем, отправка прервана")
					return successCount, fmt.Errorf("бот заблокирован пользователем: %w", err)
				case 400:
					// Bad Request - проблема с файлом, пропускаем этот файл
					log.Warn().
						Str("file_name", file.FileName).
						Err(err).
						Msg("Невалидный файл, пропускаем")
					break // Прерываем retry цикл для этого файла
				case 429:
					// Too Many Requests - exponential backoff и retry
					if attempt < maxRetries-1 {
						delay := baseDelay * time.Duration(1<<attempt)
						log.Info().
							Dur("backoff", delay).
							Int("attempt", attempt+1).
							Str("file_name", file.FileName).
							Msg("Rate limit достигнут, ожидание перед повтором")
						time.Sleep(delay)
						continue
					}
				}
			}

			// Для других ошибок (network errors) - retry
			if attempt < maxRetries-1 {
				log.Warn().
					Err(err).
					Int("attempt", attempt+1).
					Str("file_name", file.FileName).
					Msg("Ошибка отправки файла, повторная попытка")
				time.Sleep(baseDelay)
				continue
			}
		}

		if !sent {
			log.Error().
				Str("file_name", file.FileName).
				Err(lastError).
				Msg("Не удалось отправить файл после всех попыток")
		}
	}

	// Если не отправлен ни один файл - возвращаем ошибку
	if successCount == 0 {
		return 0, fmt.Errorf("не удалось отправить ни одного файла из %d: %w", len(files), lastError)
	}

	// Если отправлены не все файлы - логируем warning, но не возвращаем ошибку
	if successCount < len(files) {
		log.Warn().
			Int("success_count", successCount).
			Int("total_files", len(files)).
			Msg("Частичная отправка медиа группы: некоторые файлы не были доставлены")
	}

	return successCount, nil
}

// sendDocumentAttempt выполняет одну попытку отправки документа
func (c *Client) sendDocumentAttempt(chatID int64, fileName, filePath string) error {
	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Получаем информацию о файле
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Создаем multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Добавляем поле chat_id
	if err := writer.WriteField("chat_id", fmt.Sprintf("%d", chatID)); err != nil {
		return fmt.Errorf("failed to write chat_id field: %w", err)
	}

	// Добавляем файл как document
	part, err := writer.CreateFormFile("document", fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	// Закрываем writer чтобы записать boundary
	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Отправляем запрос
	url := c.baseURL + "sendDocument"
	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Парсим ответ
	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Обработка ошибок API
	if !apiResp.Ok {
		return &TelegramError{
			ErrorCode:   apiResp.ErrorCode,
			Description: apiResp.Description,
		}
	}

	// Обработка HTTP статусов
	if resp.StatusCode >= 400 {
		return &TelegramError{
			ErrorCode:   resp.StatusCode,
			Description: fmt.Sprintf("HTTP error: %s", resp.Status),
		}
	}

	log.Info().
		Int64("chat_id", chatID).
		Str("file_name", fileName).
		Int64("file_size", fileInfo.Size()).
		Msg("Document sent successfully")

	return nil
}

// GetMe получает информацию о боте
func (c *Client) GetMe() (*User, error) {
	resp, err := c.doRequest("getMe", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}

	var user User
	if err := json.Unmarshal(resp.Result, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// SetWebhook устанавливает webhook URL для получения обновлений
// secret_token используется для подписи всех webhook обновлений
func (c *Client) SetWebhook(url string, secretToken string) error {
	payload := map[string]interface{}{
		"url": url,
	}

	// Передаём secret_token при регистрации webhook
	// Telegram будет использовать его для подписи каждого обновления
	if secretToken != "" {
		payload["secret_token"] = secretToken
	}

	_, err := c.doRequest("setWebhook", payload)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	return nil
}

// DeleteWebhook удаляет webhook
func (c *Client) DeleteWebhook() error {
	_, err := c.doRequest("deleteWebhook", nil)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	return nil
}

// GetUpdates получает обновления от Telegram через long polling
// offset - ID первого обновления (используйте last_update_id + 1)
// timeout - таймаут в секундах для long polling (рекомендуется 30)
func (c *Client) GetUpdates(offset int64, timeout int) ([]*Update, error) {
	payload := map[string]interface{}{
		"timeout": timeout,
	}
	if offset > 0 {
		payload["offset"] = offset
	}

	// Используем отдельный клиент с увеличенным timeout для long polling
	// Telegram рекомендует timeout на 5-10 секунд больше чем запрошенный
	// Используем тот же proxy что и основной клиент
	transport := createTransport(c.proxyURL)
	longPollingClient := &http.Client{
		Timeout:   time.Duration(timeout+15) * time.Second,
		Transport: transport,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := c.baseURL + "getUpdates"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := longPollingClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Ok {
		return nil, &TelegramError{
			ErrorCode:   apiResp.ErrorCode,
			Description: apiResp.Description,
		}
	}

	var updates []*Update
	if err := json.Unmarshal(apiResp.Result, &updates); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updates: %w", err)
	}

	return updates, nil
}

// UpdateHandler тип функции-обработчика обновлений
type UpdateHandler func(ctx context.Context, update *Update) error

// StartPolling запускает long polling для получения обновлений
// handler вызывается для каждого полученного обновления
// Возвращает немедленно, polling работает в отдельной горутине
func (c *Client) StartPolling(handler UpdateHandler) error {
	c.pollingMu.Lock()
	defer c.pollingMu.Unlock()

	// Проверяем, не запущен ли уже polling
	if c.pollingCancel != nil {
		return fmt.Errorf("polling already started")
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.pollingCancel = cancel

	c.pollingWg.Add(1)
	go c.pollingLoop(ctx, handler)

	log.Info().Msg("Telegram polling started")
	return nil
}

// StopPolling останавливает polling и ждёт завершения горутины
func (c *Client) StopPolling() {
	c.pollingMu.Lock()
	if c.pollingCancel != nil {
		c.pollingCancel()
		c.pollingCancel = nil
	}
	c.pollingMu.Unlock()

	// Ждём завершения горутины polling
	c.pollingWg.Wait()
	log.Info().Msg("Telegram polling stopped")
}

// pollingLoop основной цикл polling
func (c *Client) pollingLoop(ctx context.Context, handler UpdateHandler) {
	defer c.pollingWg.Done()

	var offset int64 = 0
	consecutiveErrors := 0
	maxConsecutiveErrors := 10
	baseBackoff := time.Second

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Telegram polling loop stopped by context")
			return
		default:
			// Получаем обновления с таймаутом 30 секунд
			updates, err := c.GetUpdates(offset, 30)
			if err != nil {
				consecutiveErrors++
				log.Warn().Err(err).Int("consecutive_errors", consecutiveErrors).Msg("Failed to get Telegram updates")

				if consecutiveErrors >= maxConsecutiveErrors {
					log.Error().Msg("Too many consecutive errors in Telegram polling, stopping")
					return
				}

				// Exponential backoff: 1s, 2s, 4s, 8s, до 30s макс
				backoff := baseBackoff * time.Duration(1<<consecutiveErrors)
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}

				select {
				case <-ctx.Done():
					return
				case <-time.After(backoff):
					continue
				}
			}

			// Сбрасываем счётчик ошибок при успехе
			consecutiveErrors = 0

			// Обрабатываем каждое обновление
			for _, update := range updates {
				// Обновляем offset для следующего запроса
				if update.UpdateID >= offset {
					offset = update.UpdateID + 1
				}

				// Обрабатываем обновление в отдельной горутине
				go func(u *Update) {
					handlerCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()

					if err := handler(handlerCtx, u); err != nil {
						log.Warn().Err(err).Int64("update_id", u.UpdateID).Msg("Failed to handle Telegram update")
					}
				}(update)
			}
		}
	}
}

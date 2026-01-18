package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// YooKassaClientInterface определяет интерфейс для работы с YooKassa API
type YooKassaClientInterface interface {
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*YooKassaPaymentResponse, error)
	GetPayment(ctx context.Context, paymentID string) (*YooKassaPaymentResponse, error)
}

// YooKassaClient представляет клиент для работы с API YooKassa
type YooKassaClient struct {
	shopID     string
	secretKey  string
	baseURL    string
	httpClient *http.Client
}

// NewYooKassaClient создает новый YooKassa клиент
func NewYooKassaClient(shopID, secretKey string) *YooKassaClient {
	return &YooKassaClient{
		shopID:    shopID,
		secretKey: secretKey,
		baseURL:   "https://api.yookassa.ru/v3",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreatePaymentRequest представляет запрос на создание платежа в YooKassa
type CreatePaymentRequest struct {
	Amount         Amount       `json:"amount"`
	Capture        bool         `json:"capture"`
	Confirmation   Confirmation `json:"confirmation"`
	Description    string       `json:"description"`
	Metadata       Metadata     `json:"metadata,omitempty"`
	IdempotencyKey string       `json:"-"` // В header, не в body
}

// Amount представляет сумму платежа
type Amount struct {
	Value    string `json:"value"`    // Сумма в виде строки, например "100.00"
	Currency string `json:"currency"` // "RUB"
}

// Confirmation представляет настройки подтверждения платежа
type Confirmation struct {
	Type      string `json:"type"`       // "redirect"
	ReturnURL string `json:"return_url"` // URL для возврата после оплаты
}

// Metadata представляет дополнительные данные платежа
type Metadata struct {
	PaymentID string `json:"payment_id,omitempty"` // ID платежа в нашей системе
}

// YooKassaPaymentResponse представляет ответ от YooKassa при создании платежа
type YooKassaPaymentResponse struct {
	ID           string                   `json:"id"`
	Status       string                   `json:"status"`
	Paid         bool                     `json:"paid"`
	Amount       Amount                   `json:"amount"`
	Confirmation YooKassaConfirmationResp `json:"confirmation"`
	Metadata     Metadata                 `json:"metadata"`
}

// YooKassaConfirmationResp представляет данные подтверждения в ответе
type YooKassaConfirmationResp struct {
	Type            string `json:"type"`
	ConfirmationURL string `json:"confirmation_url"`
}

// YooKassaErrorResponse представляет ошибку от YooKassa
type YooKassaErrorResponse struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Parameter   string `json:"parameter,omitempty"`
}

// CreatePayment создает платеж в YooKassa
func (c *YooKassaClient) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*YooKassaPaymentResponse, error) {
	// Формируем JSON payload
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment request: %w", err)
	}

	// Создаем HTTP запрос
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/payments", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotence-Key", req.IdempotencyKey)

	// HTTP Basic Auth: base64(shopID:secretKey)
	auth := base64.StdEncoding.EncodeToString([]byte(c.shopID + ":" + c.secretKey))
	httpReq.Header.Set("Authorization", "Basic "+auth)

	// Выполняем запрос
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Проверяем код ответа
	if resp.StatusCode != http.StatusOK {
		var errResp YooKassaErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("YooKassa API error (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("YooKassa API error: %s - %s", errResp.Code, errResp.Description)
	}

	// Парсим успешный ответ
	var paymentResp YooKassaPaymentResponse
	if err := json.Unmarshal(body, &paymentResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment response: %w", err)
	}

	return &paymentResp, nil
}

// GetPayment получает информацию о платеже из YooKassa
func (c *YooKassaClient) GetPayment(ctx context.Context, paymentID string) (*YooKassaPaymentResponse, error) {
	// Создаем HTTP запрос
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/payments/"+paymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// HTTP Basic Auth
	auth := base64.StdEncoding.EncodeToString([]byte(c.shopID + ":" + c.secretKey))
	httpReq.Header.Set("Authorization", "Basic "+auth)

	// Выполняем запрос
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Проверяем код ответа
	if resp.StatusCode != http.StatusOK {
		var errResp YooKassaErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("YooKassa API error (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("YooKassa API error: %s - %s", errResp.Code, errResp.Description)
	}

	// Парсим ответ
	var paymentResp YooKassaPaymentResponse
	if err := json.Unmarshal(body, &paymentResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment response: %w", err)
	}

	return &paymentResp, nil
}

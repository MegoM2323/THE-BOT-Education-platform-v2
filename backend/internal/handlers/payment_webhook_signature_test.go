package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tutoring-platform/internal/config"
	"tutoring-platform/internal/models"

	"github.com/google/uuid"
)

// MockPaymentServiceForTests implements PaymentServiceInterface for testing
type MockPaymentServiceForTests struct {
	CreatePaymentFunc              func(ctx context.Context, userID uuid.UUID, req *models.CreatePaymentRequest) (*models.PaymentResponse, error)
	GetPaymentHistoryFunc          func(ctx context.Context, userID uuid.UUID) ([]*models.Payment, error)
	ProcessPaymentSuccessFunc      func(ctx context.Context, paymentID string) error
	ProcessPaymentCancellationFunc func(ctx context.Context, paymentID string) error
}

func (m *MockPaymentServiceForTests) CreatePayment(ctx context.Context, userID uuid.UUID, req *models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	if m.CreatePaymentFunc != nil {
		return m.CreatePaymentFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *MockPaymentServiceForTests) GetPaymentHistory(ctx context.Context, userID uuid.UUID) ([]*models.Payment, error) {
	if m.GetPaymentHistoryFunc != nil {
		return m.GetPaymentHistoryFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockPaymentServiceForTests) ProcessPaymentSuccess(ctx context.Context, paymentID string) error {
	if m.ProcessPaymentSuccessFunc != nil {
		return m.ProcessPaymentSuccessFunc(ctx, paymentID)
	}
	return nil
}

func (m *MockPaymentServiceForTests) ProcessPaymentCancellation(ctx context.Context, paymentID string) error {
	if m.ProcessPaymentCancellationFunc != nil {
		return m.ProcessPaymentCancellationFunc(ctx, paymentID)
	}
	return nil
}

// TestYooKassaWebhook_ValidSignature проверяет обработку webhook с валидной подписью
func TestYooKassaWebhook_ValidSignature(t *testing.T) {
	secretKey := "test_secret_key"
	cfg := &config.Config{
		YooKassa: config.YooKassaConfig{
			ShopID:    "test_shop_id",
			SecretKey: secretKey,
			ReturnURL: "http://localhost:5173/payment-success",
		},
	}

	// Note: We pass nil for paymentService since signature tests don't exercise the service
	handler := NewPaymentHandler(nil, cfg)

	// Создаем payload webhook
	webhook := models.YooKassaWebhookRequest{
		Type:  "notification",
		Event: "payment.succeeded",
		Object: models.YooKassaPaymentObject{
			ID:     "test_payment_123",
			Status: "succeeded",
			Paid:   true,
			Amount: struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			}{
				Value:    "2800.00",
				Currency: "RUB",
			},
			Metadata: struct {
				PaymentID string `json:"payment_id"`
			}{
				PaymentID: "test_payment_123",
			},
		},
	}

	body, _ := json.Marshal(webhook)

	// Вычисляем корректную подпись
	h256 := hmac.New(sha256.New, []byte(secretKey))
	h256.Write(body)
	signature := hex.EncodeToString(h256.Sum(nil))

	// Создаем HTTP запрос с корректной подписью
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Yookassa-Shop-Api-Signature-SHA256", signature)

	w := httptest.NewRecorder()
	handler.YooKassaWebhook(w, req)

	// Проверяем успешный ответ (200 OK)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// TestYooKassaWebhook_InvalidSignature проверяет отклонение webhook с неверной подписью
func TestYooKassaWebhook_InvalidSignature(t *testing.T) {
	secretKey := "test_secret_key"
	cfg := &config.Config{
		YooKassa: config.YooKassaConfig{
			ShopID:    "test_shop_id",
			SecretKey: secretKey,
			ReturnURL: "http://localhost:5173/payment-success",
		},
	}

	// Note: We pass nil for paymentService since signature tests don't exercise the service
	handler := NewPaymentHandler(nil, cfg)

	// Создаем payload webhook
	webhook := models.YooKassaWebhookRequest{
		Type:  "notification",
		Event: "payment.succeeded",
		Object: models.YooKassaPaymentObject{
			ID:     "test_payment_456",
			Status: "succeeded",
			Paid:   true,
		},
	}

	body, _ := json.Marshal(webhook)

	// Используем неверную подпись
	invalidSignature := "invalid_signature_1234567890abcdef1234567890abcdef1234567890ab"

	// Создаем HTTP запрос с неверной подписью
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Yookassa-Shop-Api-Signature-SHA256", invalidSignature)

	w := httptest.NewRecorder()
	handler.YooKassaWebhook(w, req)

	// Проверяем ошибку 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Проверяем, что ответ содержит информацию об ошибке подписи
	var errorResponse struct {
		Success bool `json:"success"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	json.NewDecoder(w.Body).Decode(&errorResponse)
	if errorResponse.Error.Code != "INVALID_SIGNATURE" {
		t.Errorf("Expected error code INVALID_SIGNATURE, got %s", errorResponse.Error.Code)
	}
}

// TestYooKassaWebhook_MissingSignatureHeader проверяет отклонение webhook без заголовка подписи
func TestYooKassaWebhook_MissingSignatureHeader(t *testing.T) {
	secretKey := "test_secret_key"
	cfg := &config.Config{
		YooKassa: config.YooKassaConfig{
			ShopID:    "test_shop_id",
			SecretKey: secretKey,
			ReturnURL: "http://localhost:5173/payment-success",
		},
	}

	// Note: We pass nil for paymentService since signature tests don't exercise the service
	handler := NewPaymentHandler(nil, cfg)

	// Создаем payload webhook
	webhook := models.YooKassaWebhookRequest{
		Type:  "notification",
		Event: "payment.succeeded",
		Object: models.YooKassaPaymentObject{
			ID: "test_payment_789",
		},
	}

	body, _ := json.Marshal(webhook)

	// Создаем HTTP запрос БЕЗ заголовка подписи
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// НЕ устанавливаем заголовок X-Yookassa-Shop-Api-Signature-SHA256

	w := httptest.NewRecorder()
	handler.YooKassaWebhook(w, req)

	// Проверяем ошибку 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// TestYooKassaWebhook_MissingSecretKey проверяет обработку ошибки конфигурации
func TestYooKassaWebhook_MissingSecretKey(t *testing.T) {
	// Конфигурация БЕЗ секретного ключа
	cfg := &config.Config{
		YooKassa: config.YooKassaConfig{
			ShopID:    "test_shop_id",
			SecretKey: "", // Пустой ключ - ошибка конфигурации
			ReturnURL: "http://localhost:5173/payment-success",
		},
	}

	// Note: We pass nil for paymentService since signature tests don't exercise the service
	handler := NewPaymentHandler(nil, cfg)

	webhook := models.YooKassaWebhookRequest{
		Type:  "notification",
		Event: "payment.succeeded",
		Object: models.YooKassaPaymentObject{
			ID: "test_payment_999",
		},
	}

	body, _ := json.Marshal(webhook)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Yookassa-Shop-Api-Signature-SHA256", "some_signature")

	w := httptest.NewRecorder()
	handler.YooKassaWebhook(w, req)

	// Проверяем ошибку 401 (так как конфигурация неполна)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// TestYooKassaWebhook_PaymentCancelled проверяет обработку webhook отмены платежа с валидной подписью
func TestYooKassaWebhook_PaymentCancelled(t *testing.T) {
	secretKey := "test_secret_key"
	cfg := &config.Config{
		YooKassa: config.YooKassaConfig{
			ShopID:    "test_shop_id",
			SecretKey: secretKey,
			ReturnURL: "http://localhost:5173/payment-success",
		},
	}

	paymentIDCapture := ""
	mockPaymentService := &MockPaymentServiceForTests{
		ProcessPaymentCancellationFunc: func(ctx context.Context, paymentID string) error {
			paymentIDCapture = paymentID
			return nil
		},
	}

	handler := NewPaymentHandler(mockPaymentService, cfg)

	webhook := models.YooKassaWebhookRequest{
		Type:  "notification",
		Event: "payment.canceled",
		Object: models.YooKassaPaymentObject{
			ID:     "cancelled_payment_123",
			Status: "canceled",
		},
	}

	body, _ := json.Marshal(webhook)

	// Вычисляем корректную подпись
	h256 := hmac.New(sha256.New, []byte(secretKey))
	h256.Write(body)
	signature := hex.EncodeToString(h256.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Yookassa-Shop-Api-Signature-SHA256", signature)

	w := httptest.NewRecorder()
	handler.YooKassaWebhook(w, req)

	// Проверяем успешный ответ (200 OK)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Проверяем, что был вызван правильный обработчик отмены
	if paymentIDCapture != "cancelled_payment_123" {
		t.Errorf("Expected payment ID 'cancelled_payment_123', got '%s'", paymentIDCapture)
	}
}

// TestYooKassaWebhook_UnknownEvent проверяет обработку неизвестного типа события
func TestYooKassaWebhook_UnknownEvent(t *testing.T) {
	secretKey := "test_secret_key"
	cfg := &config.Config{
		YooKassa: config.YooKassaConfig{
			ShopID:    "test_shop_id",
			SecretKey: secretKey,
			ReturnURL: "http://localhost:5173/payment-success",
		},
	}

	// Note: We pass nil for paymentService since unknown events don't trigger service calls
	handler := NewPaymentHandler(nil, cfg)

	webhook := models.YooKassaWebhookRequest{
		Type:  "notification",
		Event: "some.unknown.event",
		Object: models.YooKassaPaymentObject{
			ID: "unknown_event_payment",
		},
	}

	body, _ := json.Marshal(webhook)

	// Вычисляем корректную подпись
	h256 := hmac.New(sha256.New, []byte(secretKey))
	h256.Write(body)
	signature := hex.EncodeToString(h256.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Yookassa-Shop-Api-Signature-SHA256", signature)

	w := httptest.NewRecorder()
	handler.YooKassaWebhook(w, req)

	// Проверяем успешный ответ (200 OK) - даже для неизвестного события
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

// BenchmarkYooKassaSignatureVerification проверяет производительность верификации подписи
func BenchmarkYooKassaSignatureVerification(b *testing.B) {
	secretKey := "test_secret_key_for_benchmark"
	cfg := &config.Config{
		YooKassa: config.YooKassaConfig{
			ShopID:    "test_shop_id",
			SecretKey: secretKey,
			ReturnURL: "http://localhost:5173/payment-success",
		},
	}

	// Note: We pass nil for paymentService since benchmark only tests signature verification
	handler := NewPaymentHandler(nil, cfg)

	webhook := models.YooKassaWebhookRequest{
		Type:  "notification",
		Event: "payment.succeeded",
		Object: models.YooKassaPaymentObject{
			ID:     "benchmark_payment",
			Status: "succeeded",
			Paid:   true,
		},
	}

	body, _ := json.Marshal(webhook)

	// Вычисляем подпись один раз
	h256 := hmac.New(sha256.New, []byte(secretKey))
	h256.Write(body)
	signature := hex.EncodeToString(h256.Sum(nil))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhook", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Yookassa-Shop-Api-Signature-SHA256", signature)

		w := httptest.NewRecorder()
		handler.YooKassaWebhook(w, req)
	}
}

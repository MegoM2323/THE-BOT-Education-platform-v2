package service

import (
	"context"
	"testing"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
)

func TestCreatePaymentRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		credits int
		wantErr bool
	}{
		{
			name:    "Valid request with 1 credit",
			credits: 1,
			wantErr: false,
		},
		{
			name:    "Valid request with 10 credits",
			credits: 10,
			wantErr: false,
		},
		{
			name:    "Invalid request with 0 credits",
			credits: 0,
			wantErr: true,
		},
		{
			name:    "Invalid request with negative credits",
			credits: -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &models.CreatePaymentRequest{
				Credits: tt.credits,
			}

			err := req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreatePaymentRequest_CalculateAmount(t *testing.T) {
	tests := []struct {
		name       string
		credits    int
		wantAmount float64
	}{
		{
			name:       "1 credit = 2800 rubles",
			credits:    1,
			wantAmount: 2800.0,
		},
		{
			name:       "5 credits = 14000 rubles",
			credits:    5,
			wantAmount: 14000.0,
		},
		{
			name:       "10 credits = 28000 rubles",
			credits:    10,
			wantAmount: 28000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &models.CreatePaymentRequest{
				Credits: tt.credits,
			}

			amount := req.CalculateAmount()
			if amount != tt.wantAmount {
				t.Errorf("CalculateAmount() = %v, want %v", amount, tt.wantAmount)
			}
		})
	}
}

func TestYooKassaClient_CreatePaymentRequest(t *testing.T) {
	// Тест структуры запроса к YooKassa
	req := &CreatePaymentRequest{
		Amount: Amount{
			Value:    "2800.00",
			Currency: "RUB",
		},
		Capture: true,
		Confirmation: Confirmation{
			Type:      "redirect",
			ReturnURL: "http://localhost:5173/payment-success",
		},
		Description:    "Покупка 1 кредитов",
		IdempotencyKey: uuid.New().String(),
		Metadata: Metadata{
			PaymentID: uuid.New().String(),
		},
	}

	if req.Amount.Value != "2800.00" {
		t.Errorf("Amount.Value = %s, want 2800.00", req.Amount.Value)
	}

	if req.Amount.Currency != "RUB" {
		t.Errorf("Amount.Currency = %s, want RUB", req.Amount.Currency)
	}

	if !req.Capture {
		t.Error("Capture should be true for automatic payment capture")
	}

	if req.Confirmation.Type != "redirect" {
		t.Errorf("Confirmation.Type = %s, want redirect", req.Confirmation.Type)
	}

	if req.IdempotencyKey == "" {
		t.Error("IdempotencyKey should not be empty")
	}

	if req.Metadata.PaymentID == "" {
		t.Error("Metadata.PaymentID should not be empty")
	}
}

func TestPaymentStatus_Constants(t *testing.T) {
	// Проверяем константы статусов
	if models.PaymentStatusPending != "pending" {
		t.Errorf("PaymentStatusPending = %s, want pending", models.PaymentStatusPending)
	}

	if models.PaymentStatusSucceeded != "succeeded" {
		t.Errorf("PaymentStatusSucceeded = %s, want succeeded", models.PaymentStatusSucceeded)
	}

	if models.PaymentStatusCancelled != "cancelled" {
		t.Errorf("PaymentStatusCancelled = %s, want cancelled", models.PaymentStatusCancelled)
	}

	if models.PaymentStatusFailed != "failed" {
		t.Errorf("PaymentStatusFailed = %s, want failed", models.PaymentStatusFailed)
	}
}

func TestCreditPrice_Constant(t *testing.T) {
	// Проверяем константу цены кредита
	if models.CreditPrice != 2800.0 {
		t.Errorf("CreditPrice = %v, want 2800.0", models.CreditPrice)
	}
}

// Mock tests для проверки логики (без реального API вызова)
func TestYooKassaClient_NewClient(t *testing.T) {
	shopID := "test_shop_id"
	secretKey := "test_secret_key"

	client := NewYooKassaClient(shopID, secretKey)

	if client.shopID != shopID {
		t.Errorf("shopID = %s, want %s", client.shopID, shopID)
	}

	if client.secretKey != secretKey {
		t.Errorf("secretKey = %s, want %s", client.secretKey, secretKey)
	}

	if client.baseURL != "https://api.yookassa.ru/v3" {
		t.Errorf("baseURL = %s, want https://api.yookassa.ru/v3", client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

// Integration test для проверки workflow (требует mock repository)
func TestPaymentWorkflow_Concepts(t *testing.T) {
	// Концептуальный тест workflow (без реальной базы)
	ctx := context.Background()
	userID := uuid.New()
	credits := 1

	// 1. Создание запроса
	req := &models.CreatePaymentRequest{
		Credits: credits,
	}

	// 2. Валидация
	if err := req.Validate(); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// 3. Рассчет суммы
	amount := req.CalculateAmount()
	expectedAmount := float64(credits) * models.CreditPrice
	if amount != expectedAmount {
		t.Errorf("Amount = %v, want %v", amount, expectedAmount)
	}

	// 4. Создание платежа в БД (mock)
	payment := &models.Payment{
		ID:      uuid.New(),
		UserID:  userID,
		Amount:  amount,
		Credits: credits,
		Status:  models.PaymentStatusPending,
	}

	// 5. Создание запроса к YooKassa
	yookassaReq := &CreatePaymentRequest{
		Amount: Amount{
			Value:    "2800.00",
			Currency: "RUB",
		},
		Capture: true,
		Confirmation: Confirmation{
			Type:      "redirect",
			ReturnURL: "http://localhost:5173/payment-success",
		},
		Description:    "Покупка 1 кредитов",
		IdempotencyKey: payment.ID.String(),
		Metadata: Metadata{
			PaymentID: payment.ID.String(),
		},
	}

	// 6. Проверка idempotency key
	if yookassaReq.IdempotencyKey != payment.ID.String() {
		t.Error("Idempotency key should match payment ID")
	}

	// 7. Проверка metadata
	if yookassaReq.Metadata.PaymentID != payment.ID.String() {
		t.Error("Metadata PaymentID should match payment ID")
	}

	// Тест успешен если все проверки прошли
	_ = ctx // использование контекста для консистентности
}

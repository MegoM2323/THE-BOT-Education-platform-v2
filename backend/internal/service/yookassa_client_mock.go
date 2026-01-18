package service

import (
	"context"

	"github.com/google/uuid"
)

// MockYooKassaClient - mock для тестирования без реальных вызовов к YooKassa API
type MockYooKassaClient struct {
	// Можно добавить поля для отслеживания вызовов в тестах
	CreatePaymentCalled bool
	LastRequest         *CreatePaymentRequest
}

// CreatePayment - mock implementation
func (m *MockYooKassaClient) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*YooKassaPaymentResponse, error) {
	m.CreatePaymentCalled = true
	m.LastRequest = req

	// Возвращаем mock ответ с подтверждающим URL
	return &YooKassaPaymentResponse{
		ID:     "mock_yookassa_" + uuid.New().String(),
		Status: "pending",
		Paid:   false,
		Amount: req.Amount,
		Confirmation: YooKassaConfirmationResp{
			Type:            "redirect",
			ConfirmationURL: "https://yoomoney.ru/checkout/payments/" + uuid.New().String(),
		},
		Metadata: req.Metadata,
	}, nil
}

// GetPayment - mock implementation
func (m *MockYooKassaClient) GetPayment(ctx context.Context, paymentID string) (*YooKassaPaymentResponse, error) {
	return &YooKassaPaymentResponse{
		ID:     paymentID,
		Status: "succeeded",
		Paid:   true,
		Amount: Amount{
			Value:    "2800.00",
			Currency: "RUB",
		},
	}, nil
}

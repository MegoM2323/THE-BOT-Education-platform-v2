package service

import (
	"context"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
)

// MockPaymentService - mock для тестирования PaymentHandler без реальных вызовов к сервису
type MockPaymentService struct {
	CreatePaymentFunc              func(ctx context.Context, userID uuid.UUID, req *models.CreatePaymentRequest) (*models.PaymentResponse, error)
	GetPaymentHistoryFunc          func(ctx context.Context, userID uuid.UUID) ([]*models.Payment, error)
	ProcessPaymentSuccessFunc      func(ctx context.Context, paymentID string) error
	ProcessPaymentCancellationFunc func(ctx context.Context, paymentID string) error
}

// CreatePayment - mock implementation
func (m *MockPaymentService) CreatePayment(ctx context.Context, userID uuid.UUID, req *models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	if m.CreatePaymentFunc != nil {
		return m.CreatePaymentFunc(ctx, userID, req)
	}
	return nil, nil
}

// GetPaymentHistory - mock implementation
func (m *MockPaymentService) GetPaymentHistory(ctx context.Context, userID uuid.UUID) ([]*models.Payment, error) {
	if m.GetPaymentHistoryFunc != nil {
		return m.GetPaymentHistoryFunc(ctx, userID)
	}
	return nil, nil
}

// ProcessPaymentSuccess - mock implementation
func (m *MockPaymentService) ProcessPaymentSuccess(ctx context.Context, paymentID string) error {
	if m.ProcessPaymentSuccessFunc != nil {
		return m.ProcessPaymentSuccessFunc(ctx, paymentID)
	}
	return nil
}

// ProcessPaymentCancellation - mock implementation
func (m *MockPaymentService) ProcessPaymentCancellation(ctx context.Context, paymentID string) error {
	if m.ProcessPaymentCancellationFunc != nil {
		return m.ProcessPaymentCancellationFunc(ctx, paymentID)
	}
	return nil
}

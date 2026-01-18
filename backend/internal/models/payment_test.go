package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreatePaymentRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		credits int
		wantErr bool
	}{
		{
			name:    "valid request - 1 credit (minimum)",
			credits: 1,
			wantErr: false,
		},
		{
			name:    "valid request - 5 credits",
			credits: 5,
			wantErr: false,
		},
		{
			name:    "valid request - 10 credits",
			credits: 10,
			wantErr: false,
		},
		{
			name:    "valid request - 100 credits (maximum)",
			credits: 100,
			wantErr: false,
		},
		{
			name:    "invalid - zero credits",
			credits: 0,
			wantErr: true,
		},
		{
			name:    "invalid - negative credits",
			credits: -1,
			wantErr: true,
		},
		{
			name:    "invalid - 101 credits (exceeds maximum)",
			credits: 101,
			wantErr: true,
		},
		{
			name:    "invalid - 500 credits (far exceeds maximum)",
			credits: 500,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreatePaymentRequest{
				Credits: tt.credits,
			}
			err := req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != ErrInvalidCreditAmount {
				t.Errorf("Expected ErrInvalidCreditAmount, got %v", err)
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
			name:       "1 credit = 2800 RUB",
			credits:    1,
			wantAmount: 2800.0,
		},
		{
			name:       "5 credits = 14000 RUB",
			credits:    5,
			wantAmount: 14000.0,
		},
		{
			name:       "10 credits = 28000 RUB",
			credits:    10,
			wantAmount: 28000.0,
		},
		{
			name:       "100 credits = 280000 RUB (maximum)",
			credits:    100,
			wantAmount: 280000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &CreatePaymentRequest{
				Credits: tt.credits,
			}
			amount := req.CalculateAmount()
			if amount != tt.wantAmount {
				t.Errorf("CalculateAmount() = %v, want %v", amount, tt.wantAmount)
			}
		})
	}
}

func TestPayment_StatusMethods(t *testing.T) {
	payment := &Payment{
		ID:     uuid.New(),
		Status: PaymentStatusPending,
	}

	if !payment.IsPending() {
		t.Errorf("IsPending() = false, want true")
	}
	if payment.IsSucceeded() {
		t.Errorf("IsSucceeded() = true, want false")
	}
	if payment.IsCancelled() {
		t.Errorf("IsCancelled() = true, want false")
	}

	payment.Status = PaymentStatusSucceeded
	if payment.IsPending() {
		t.Errorf("IsPending() = true, want false")
	}
	if !payment.IsSucceeded() {
		t.Errorf("IsSucceeded() = false, want true")
	}

	payment.Status = PaymentStatusCancelled
	if !payment.IsCancelled() {
		t.Errorf("IsCancelled() = false, want true")
	}
}

func TestPaymentHistoryFilter_Validate(t *testing.T) {
	userID := uuid.New()
	pending := PaymentStatusPending
	succeeded := PaymentStatusSucceeded

	tests := []struct {
		name    string
		filter  *PaymentHistoryFilter
		wantErr bool
	}{
		{
			name: "valid filter with all fields",
			filter: &PaymentHistoryFilter{
				UserID: &userID,
				Status: &pending,
				Limit:  10,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "valid filter minimal",
			filter: &PaymentHistoryFilter{
				Limit: 20,
			},
			wantErr: false,
		},
		{
			name: "default limit applied",
			filter: &PaymentHistoryFilter{
				Limit: 0,
			},
			wantErr: false,
		},
		{
			name: "valid status succeeded",
			filter: &PaymentHistoryFilter{
				Status: &succeeded,
				Limit:  10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Проверка установки значений по умолчанию
			if tt.filter.Limit == 0 {
				t.Errorf("Validate() did not set default Limit")
			}
			if tt.filter.Offset < 0 {
				t.Errorf("Validate() did not fix negative Offset")
			}
		})
	}
}

package models

import (
	"testing"

	"github.com/google/uuid"
)

// TestAddCreditsRequest_ValidateAmountRange проверяет валидацию диапазона Amount для AddCreditsRequest
func TestAddCreditsRequest_ValidateAmountRange(t *testing.T) {
	adminID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name    string
		amount  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid amount at lower bound",
			amount:  1,
			wantErr: false,
		},
		{
			name:    "valid amount at upper bound",
			amount:  100,
			wantErr: false,
		},
		{
			name:    "valid amount in middle range",
			amount:  50,
			wantErr: false,
		},
		{
			name:    "invalid amount zero",
			amount:  0,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
		{
			name:    "invalid amount negative",
			amount:  -1,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
		{
			name:    "invalid amount exceeds upper bound",
			amount:  101,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
		{
			name:    "invalid amount far exceeds upper bound",
			amount:  1000,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
		{
			name:    "invalid amount large negative",
			amount:  -100,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &AddCreditsRequest{
				UserID:      userID,
				Amount:      tt.amount,
				Reason:      "Test reason",
				PerformedBy: adminID,
			}

			err := req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, expected %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestDeductCreditsRequest_ValidateAmountRange проверяет валидацию диапазона Amount для DeductCreditsRequest
func TestDeductCreditsRequest_ValidateAmountRange(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name    string
		amount  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid amount at lower bound",
			amount:  1,
			wantErr: false,
		},
		{
			name:    "valid amount at upper bound",
			amount:  100,
			wantErr: false,
		},
		{
			name:    "valid amount in middle range",
			amount:  75,
			wantErr: false,
		},
		{
			name:    "invalid amount zero",
			amount:  0,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
		{
			name:    "invalid amount negative",
			amount:  -1,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
		{
			name:    "invalid amount exceeds upper bound",
			amount:  101,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
		{
			name:    "invalid amount far exceeds upper bound",
			amount:  500,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &DeductCreditsRequest{
				UserID: userID,
				Amount: tt.amount,
				Reason: "Test reason",
			}

			err := req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, expected %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestRefundCreditsRequest_ValidateAmountRange проверяет валидацию диапазона Amount для RefundCreditsRequest
func TestRefundCreditsRequest_ValidateAmountRange(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name    string
		amount  int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid amount at lower bound",
			amount:  1,
			wantErr: false,
		},
		{
			name:    "valid amount at upper bound",
			amount:  100,
			wantErr: false,
		},
		{
			name:    "valid amount in middle range",
			amount:  25,
			wantErr: false,
		},
		{
			name:    "invalid amount zero",
			amount:  0,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
		{
			name:    "invalid amount negative",
			amount:  -1,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
		{
			name:    "invalid amount exceeds upper bound",
			amount:  101,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
		{
			name:    "invalid amount far exceeds upper bound",
			amount:  9999,
			wantErr: true,
			errMsg:  "количество кредитов должно быть от 1 до 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &RefundCreditsRequest{
				UserID: userID,
				Amount: tt.amount,
				Reason: "Test reason",
			}

			err := req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, expected %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestAddCreditsRequest_ValidateAllFields проверяет полную валидацию AddCreditsRequest
func TestAddCreditsRequest_ValidateAllFields(t *testing.T) {
	validUserID := uuid.New()
	validAdminID := uuid.New()

	tests := []struct {
		name    string
		request *AddCreditsRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &AddCreditsRequest{
				UserID:      validUserID,
				Amount:      50,
				Reason:      "Monthly credits",
				PerformedBy: validAdminID,
			},
			wantErr: false,
		},
		{
			name: "invalid user ID",
			request: &AddCreditsRequest{
				UserID:      uuid.Nil,
				Amount:      50,
				Reason:      "Monthly credits",
				PerformedBy: validAdminID,
			},
			wantErr: true,
		},
		{
			name: "invalid performed by ID",
			request: &AddCreditsRequest{
				UserID:      validUserID,
				Amount:      50,
				Reason:      "Monthly credits",
				PerformedBy: uuid.Nil,
			},
			wantErr: true,
		},
		{
			name: "empty reason",
			request: &AddCreditsRequest{
				UserID:      validUserID,
				Amount:      50,
				Reason:      "",
				PerformedBy: validAdminID,
			},
			wantErr: true,
		},
		{
			name: "amount exceeds upper bound",
			request: &AddCreditsRequest{
				UserID:      validUserID,
				Amount:      150,
				Reason:      "Monthly credits",
				PerformedBy: validAdminID,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDeductCreditsRequest_ValidateAllFields проверяет полную валидацию DeductCreditsRequest
func TestDeductCreditsRequest_ValidateAllFields(t *testing.T) {
	validUserID := uuid.New()

	tests := []struct {
		name    string
		request *DeductCreditsRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &DeductCreditsRequest{
				UserID: validUserID,
				Amount: 50,
				Reason: "Lesson payment",
			},
			wantErr: false,
		},
		{
			name: "invalid user ID",
			request: &DeductCreditsRequest{
				UserID: uuid.Nil,
				Amount: 50,
				Reason: "Lesson payment",
			},
			wantErr: true,
		},
		{
			name: "empty reason",
			request: &DeductCreditsRequest{
				UserID: validUserID,
				Amount: 50,
				Reason: "",
			},
			wantErr: true,
		},
		{
			name: "amount exceeds upper bound",
			request: &DeductCreditsRequest{
				UserID: validUserID,
				Amount: 250,
				Reason: "Lesson payment",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRefundCreditsRequest_ValidateAllFields проверяет полную валидацию RefundCreditsRequest
func TestRefundCreditsRequest_ValidateAllFields(t *testing.T) {
	validUserID := uuid.New()

	tests := []struct {
		name    string
		request *RefundCreditsRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &RefundCreditsRequest{
				UserID: validUserID,
				Amount: 50,
				Reason: "Cancelled lesson",
			},
			wantErr: false,
		},
		{
			name: "invalid user ID",
			request: &RefundCreditsRequest{
				UserID: uuid.Nil,
				Amount: 50,
				Reason: "Cancelled lesson",
			},
			wantErr: true,
		},
		{
			name: "empty reason",
			request: &RefundCreditsRequest{
				UserID: validUserID,
				Amount: 50,
				Reason: "",
			},
			wantErr: true,
		},
		{
			name: "amount exceeds upper bound",
			request: &RefundCreditsRequest{
				UserID: validUserID,
				Amount: 200,
				Reason: "Cancelled lesson",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCredit_HasSufficientBalance проверяет проверку достаточности баланса
func TestCredit_HasSufficientBalance(t *testing.T) {
	tests := []struct {
		name       string
		balance    int
		required   int
		sufficient bool
	}{
		{
			name:       "sufficient balance",
			balance:    100,
			required:   50,
			sufficient: true,
		},
		{
			name:       "exactly sufficient balance",
			balance:    50,
			required:   50,
			sufficient: true,
		},
		{
			name:       "insufficient balance",
			balance:    25,
			required:   50,
			sufficient: false,
		},
		{
			name:       "zero balance insufficient",
			balance:    0,
			required:   1,
			sufficient: false,
		},
		{
			name:       "zero required sufficient",
			balance:    0,
			required:   0,
			sufficient: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credit := &Credit{Balance: tt.balance}
			got := credit.HasSufficientBalance(tt.required)
			if got != tt.sufficient {
				t.Errorf("HasSufficientBalance(%d) = %v, want %v", tt.required, got, tt.sufficient)
			}
		})
	}
}

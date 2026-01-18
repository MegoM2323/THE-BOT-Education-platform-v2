package models

import (
	"time"

	"tutoring-platform/pkg/sanitize"

	"github.com/google/uuid"
)

// OperationType представляет тип операции с кредитами
type OperationType string

const (
	OperationTypeAdd    OperationType = "add"
	OperationTypeDeduct OperationType = "deduct"
	OperationTypeRefund OperationType = "refund"
)

// Credit представляет баланс кредитов пользователя
type Credit struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Balance   int       `db:"balance" json:"balance"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// CreditTransaction представляет одну транзакцию с кредитами
type CreditTransaction struct {
	ID            uuid.UUID     `db:"id" json:"id"`
	UserID        uuid.UUID     `db:"user_id" json:"user_id"`
	Amount        int           `db:"amount" json:"amount"`
	OperationType OperationType `db:"operation_type" json:"operation_type"`
	Reason        string        `db:"reason" json:"reason"`
	PerformedBy   uuid.NullUUID `db:"performed_by" json:"performed_by,omitempty"`
	BookingID     uuid.NullUUID `db:"booking_id" json:"booking_id,omitempty"`
	BalanceBefore int           `db:"balance_before" json:"balance_before"`
	BalanceAfter  int           `db:"balance_after" json:"balance_after"`
	CreatedAt     time.Time     `db:"created_at" json:"created_at"`
}

// CreditTransactionWithUser представляет транзакцию с информацией о пользователе
type CreditTransactionWithUser struct {
	CreditTransaction
	UserEmail   string `db:"user_email" json:"user_email"`
	UserName    string `db:"user_name" json:"user_name"`
	PerformedBy string `db:"performed_by_email" json:"performed_by_email,omitempty"`
}

// AddCreditsRequest представляет запрос на добавление кредитов пользователю
type AddCreditsRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	Amount      int       `json:"amount"`
	Reason      string    `json:"reason"`
	PerformedBy uuid.UUID `json:"performed_by"` // Администратор, выполнивший операцию
}

// DeductCreditsRequest представляет запрос на списание кредитов у пользователя
type DeductCreditsRequest struct {
	UserID      uuid.UUID  `json:"user_id"`
	Amount      int        `json:"amount"`
	Reason      string     `json:"reason"`
	BookingID   *uuid.UUID `json:"booking_id,omitempty"`
	PerformedBy uuid.UUID  `json:"performed_by,omitempty"` // Администратор или система, выполнившие операцию
}

// RefundCreditsRequest представляет запрос на возврат кредитов пользователю
type RefundCreditsRequest struct {
	UserID      uuid.UUID  `json:"user_id"`
	Amount      int        `json:"amount"`
	Reason      string     `json:"reason"`
	BookingID   *uuid.UUID `json:"booking_id,omitempty"`
	PerformedBy uuid.UUID  `json:"performed_by,omitempty"` // Администратор или система, выполнившие операцию
}

// Константы для пагинации истории транзакций
// DefaultTransactionLimit = 50 транзакций на страницу по умолчанию
// MaxTransactionLimit = 500 транзакций максимум (предотвращение исчерпания памяти)
// MaxBalance = 10000 максимальный лимит баланса кредитов (бизнес-логика)
const (
	DefaultTransactionLimit = 50    // Лимит по умолчанию
	MaxTransactionLimit     = 500   // Максимальный лимит
	MaxBalance              = 10000 // Максимальный баланс кредитов
)

// GetCreditHistoryFilter представляет фильтры для истории кредитов
// Limit и Offset применяются для пагинации:
// - Если Limit не указан или = 0, применяется DefaultTransactionLimit (50)
// - Если Limit > MaxTransactionLimit (500), капируется до MaxTransactionLimit
// - Offset = 0 по умолчанию (первая страница)
type GetCreditHistoryFilter struct {
	UserID        *uuid.UUID     `json:"user_id,omitempty"`
	OperationType *OperationType `json:"operation_type,omitempty"`
	StartDate     *time.Time     `json:"start_date,omitempty"`
	EndDate       *time.Time     `json:"end_date,omitempty"`
	Limit         int            `json:"limit,omitempty"`
	Offset        int            `json:"offset,omitempty"`
}

// HasSufficientBalance проверяет, достаточен ли баланс кредитов
func (c *Credit) HasSufficientBalance(amount int) bool {
	return c.Balance >= amount
}

// Validate выполняет валидацию AddCreditsRequest
func (r *AddCreditsRequest) Validate() error {
	// Санитизация входных данных
	r.Reason = sanitize.Reason(r.Reason)

	if r.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	if r.Amount <= 0 || r.Amount > 100 {
		return ErrInvalidCreditAmount
	}
	if r.Reason == "" {
		return ErrInvalidReason
	}
	if r.PerformedBy == uuid.Nil {
		return ErrInvalidPerformedBy
	}
	return nil
}

// Validate выполняет валидацию DeductCreditsRequest
func (r *DeductCreditsRequest) Validate() error {
	// Санитизация входных данных
	r.Reason = sanitize.Reason(r.Reason)

	if r.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	if r.Amount <= 0 || r.Amount > 100 {
		return ErrInvalidCreditAmount
	}
	if r.Reason == "" {
		return ErrInvalidReason
	}
	return nil
}

// Validate выполняет валидацию RefundCreditsRequest
func (r *RefundCreditsRequest) Validate() error {
	// Санитизация входных данных
	r.Reason = sanitize.Reason(r.Reason)

	if r.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	if r.Amount <= 0 || r.Amount > 100 {
		return ErrInvalidCreditAmount
	}
	if r.Reason == "" {
		return ErrInvalidReason
	}
	return nil
}

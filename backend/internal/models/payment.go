package models

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus представляет статус платежа
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"   // Ожидание оплаты
	PaymentStatusSucceeded PaymentStatus = "succeeded" // Успешно оплачен
	PaymentStatusCancelled PaymentStatus = "cancelled" // Отменен (cancelled в БД)
	PaymentStatusFailed    PaymentStatus = "failed"    // Ошибка при оплате
)

// Константы для работы с платежами
const (
	CreditPrice      = 2800.0 // 1 кредит = 2800₽
	MinCreditsAmount = 1      // Минимальное количество кредитов для покупки
	MaxCreditsAmount = 100    // Максимальное количество кредитов для покупки
)

// Payment представляет запись о платеже
type Payment struct {
	ID                uuid.UUID     `json:"id" db:"id"`
	UserID            uuid.UUID     `json:"user_id" db:"user_id"`
	YooKassaPaymentID string        `json:"yookassa_payment_id,omitempty" db:"yookassa_payment_id"`
	Amount            float64       `json:"amount" db:"amount"`   // Сумма в рублях
	Credits           int           `json:"credits" db:"credits"` // Количество кредитов
	Status            PaymentStatus `json:"status" db:"status"`   // Статус платежа
	ConfirmationURL   string        `json:"confirmation_url,omitempty" db:"confirmation_url"`
	IdempotencyKey    string        `json:"idempotency_key,omitempty" db:"idempotency_key"` // Ключ для предотвращения дубликатов
	ProcessedAt       *time.Time    `json:"processed_at,omitempty" db:"processed_at"`       // Время успешной обработки
	CreatedAt         time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at" db:"updated_at"`
}

// CreatePaymentRequest представляет запрос на создание платежа
type CreatePaymentRequest struct {
	Credits int `json:"credits"` // Количество кредитов для покупки
}

// Validate проверяет валидность запроса на создание платежа
func (r *CreatePaymentRequest) Validate() error {
	if r.Credits < MinCreditsAmount || r.Credits > MaxCreditsAmount {
		return ErrInvalidCreditAmount
	}
	return nil
}

// CalculateAmount вычисляет сумму платежа по количеству кредитов
func (r *CreatePaymentRequest) CalculateAmount() float64 {
	return float64(r.Credits) * CreditPrice
}

// PaymentResponse представляет ответ при создании платежа
type PaymentResponse struct {
	PaymentID       uuid.UUID `json:"payment_id"`
	Amount          float64   `json:"amount"`
	Credits         int       `json:"credits"`
	ConfirmationURL string    `json:"confirmation_url"`
}

// YooKassaWebhookRequest представляет webhook от YooKassa
type YooKassaWebhookRequest struct {
	Type   string                `json:"type"`
	Event  string                `json:"event"`
	Object YooKassaPaymentObject `json:"object"`
}

// YooKassaPaymentObject представляет объект платежа в webhook
type YooKassaPaymentObject struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Paid   bool   `json:"paid"`
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Metadata struct {
		PaymentID string `json:"payment_id"`
	} `json:"metadata"`
}

// PaymentHistoryFilter представляет фильтры для истории платежей
type PaymentHistoryFilter struct {
	UserID    *uuid.UUID     `json:"user_id,omitempty"`
	Status    *PaymentStatus `json:"status,omitempty"`
	StartDate *time.Time     `json:"start_date,omitempty"`
	EndDate   *time.Time     `json:"end_date,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	Offset    int            `json:"offset,omitempty"`
}

// Validate выполняет валидацию PaymentHistoryFilter
func (f *PaymentHistoryFilter) Validate() error {
	if f.Status != nil {
		status := *f.Status
		if status != PaymentStatusPending && status != PaymentStatusSucceeded && status != PaymentStatusCancelled && status != PaymentStatusFailed {
			return ErrInvalidPaymentStatus
		}
	}

	if f.Limit <= 0 {
		f.Limit = 20 // По умолчанию
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	return nil
}

// PaymentWithUser представляет платеж с информацией о пользователе
type PaymentWithUser struct {
	Payment
	UserEmail string `db:"user_email" json:"user_email"`
	UserName  string `db:"user_name" json:"user_name"`
}

// IsPending проверяет, находится ли платеж в статусе ожидания
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsSucceeded проверяет, успешно ли завершен платеж
func (p *Payment) IsSucceeded() bool {
	return p.Status == PaymentStatusSucceeded
}

// IsCancelled проверяет, отменен ли платеж
func (p *Payment) IsCancelled() bool {
	return p.Status == PaymentStatusCancelled
}

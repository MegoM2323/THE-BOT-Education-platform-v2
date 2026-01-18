package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

var (
	// ErrNotFound возвращается когда запись не найдена
	ErrNotFound = errors.New("payment not found")
)

// PaymentRepository предоставляет методы для работы с платежами
type PaymentRepository struct {
	db *sqlx.DB
}

// NewPaymentRepository создает новый PaymentRepository
func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create создает новую запись платежа
func (r *PaymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	query := `
		INSERT INTO payments (id, user_id, yookassa_payment_id, amount, credits, status, confirmation_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	payment.CreatedAt = now
	payment.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.UserID,
		payment.YooKassaPaymentID,
		payment.Amount,
		payment.Credits,
		payment.Status,
		payment.ConfirmationURL,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

// UpdateStatus обновляет статус платежа
func (r *PaymentRepository) UpdateStatus(ctx context.Context, paymentID uuid.UUID, status models.PaymentStatus, yookassaID string) error {
	query := `
		UPDATE payments
		SET status = $1, yookassa_payment_id = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, status, yookassaID, time.Now(), paymentID)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

// UpdateConfirmationURL обновляет confirmation_url платежа
func (r *PaymentRepository) UpdateConfirmationURL(ctx context.Context, paymentID uuid.UUID, confirmationURL string) error {
	query := `
		UPDATE payments
		SET confirmation_url = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, confirmationURL, time.Now(), paymentID)
	if err != nil {
		return fmt.Errorf("failed to update confirmation URL: %w", err)
	}

	return nil
}

// GetByID получает платеж по ID
func (r *PaymentRepository) GetByID(ctx context.Context, paymentID uuid.UUID) (*models.Payment, error) {
	query := `
		SELECT ` + PaymentSelectFields + `
		FROM payments
		WHERE id = $1
	`

	var payment models.Payment
	err := r.db.GetContext(ctx, &payment, query, paymentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment by ID: %w", err)
	}

	return &payment, nil
}

// GetByYooKassaID получает платеж по YooKassa payment ID
func (r *PaymentRepository) GetByYooKassaID(ctx context.Context, yookassaID string) (*models.Payment, error) {
	query := `
		SELECT ` + PaymentSelectFields + `
		FROM payments
		WHERE yookassa_payment_id = $1
	`

	var payment models.Payment
	err := r.db.GetContext(ctx, &payment, query, yookassaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment by YooKassa ID: %w", err)
	}

	return &payment, nil
}

// ListByUser возвращает список платежей пользователя
func (r *PaymentRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Payment, error) {
	query := `
		SELECT ` + PaymentSelectFields + `
		FROM payments
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var payments []*models.Payment
	err := r.db.SelectContext(ctx, &payments, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments by user: %w", err)
	}

	return payments, nil
}

// CheckIdempotency проверяет, был ли платеж с таким ключом уже обработан
// Возвращает платеж и флаг: (платеж, был_обработан, ошибка)
func (r *PaymentRepository) CheckIdempotency(ctx context.Context, idempotencyKey string) (*models.Payment, bool, error) {
	query := `
		SELECT ` + PaymentSelectFields + `
		FROM payments
		WHERE idempotency_key = $1
	`

	var payment models.Payment
	err := r.db.GetContext(ctx, &payment, query, idempotencyKey)
	if err != nil {
		if err == sql.ErrNoRows {
			// Платеж с таким ключом еще не существует
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to check idempotency: %w", err)
	}

	// Платеж существует, проверяем был ли он обработан
	isProcessed := payment.ProcessedAt != nil
	return &payment, isProcessed, nil
}

// UpdateStatusWithTx обновляет статус платежа используя переданную транзакцию
func (r *PaymentRepository) UpdateStatusWithTx(ctx context.Context, tx pgx.Tx, paymentID uuid.UUID, status models.PaymentStatus, yookassaID string) error {
	query := `
		UPDATE payments
		SET status = $1, yookassa_payment_id = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := tx.Exec(ctx, query, status, yookassaID, time.Now(), paymentID)
	if err != nil {
		return fmt.Errorf("failed to update payment status in tx: %w", err)
	}

	return nil
}

// MarkAsProcessedWithTx отмечает платеж как обработанный (кредиты добавлены)
func (r *PaymentRepository) MarkAsProcessedWithTx(ctx context.Context, tx pgx.Tx, paymentID uuid.UUID, idempotencyKey string) error {
	now := time.Now()
	query := `
		UPDATE payments
		SET processed_at = $1, idempotency_key = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := tx.Exec(ctx, query, now, idempotencyKey, now, paymentID)
	if err != nil {
		return fmt.Errorf("failed to mark payment as processed: %w", err)
	}

	return nil
}

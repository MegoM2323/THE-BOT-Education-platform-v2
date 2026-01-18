package service

import (
	"context"
	"fmt"
	"log"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/internal/utils"
	"tutoring-platform/pkg/metrics"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PaymentService обрабатывает бизнес-логику платежей
type PaymentService struct {
	pool           *pgxpool.Pool
	paymentRepo    *repository.PaymentRepository
	creditService  *CreditService
	yookassaClient YooKassaClientInterface
	userRepo       repository.UserRepository
	returnURL      string // URL для возврата после оплаты
}

// NewPaymentService создает новый PaymentService
func NewPaymentService(
	pool *pgxpool.Pool,
	paymentRepo *repository.PaymentRepository,
	creditService *CreditService,
	yookassaClient YooKassaClientInterface,
	userRepo repository.UserRepository,
	returnURL string,
) *PaymentService {
	return &PaymentService{
		pool:           pool,
		paymentRepo:    paymentRepo,
		creditService:  creditService,
		yookassaClient: yookassaClient,
		userRepo:       userRepo,
		returnURL:      returnURL,
	}
}

// CreatePayment создает новый платеж
func (s *PaymentService) CreatePayment(ctx context.Context, userID uuid.UUID, req *models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	// Валидация запроса
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Проверяем что платежи включены для пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if !user.PaymentEnabled {
		return nil, repository.ErrPaymentDisabledForUser
	}

	// Рассчитываем сумму
	amount := req.CalculateAmount()

	// Создаем запись платежа в БД со статусом pending
	payment := &models.Payment{
		ID:      uuid.New(),
		UserID:  userID,
		Amount:  amount,
		Credits: req.Credits,
		Status:  models.PaymentStatusPending,
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Формируем запрос к YooKassa
	yookassaReq := &CreatePaymentRequest{
		Amount: Amount{
			Value:    fmt.Sprintf("%.2f", amount),
			Currency: "RUB",
		},
		Capture: true, // Автоматическое подтверждение платежа
		Confirmation: Confirmation{
			Type:      "redirect",
			ReturnURL: s.returnURL,
		},
		Description: fmt.Sprintf("Покупка %d кредитов", req.Credits),
		Metadata: Metadata{
			PaymentID: payment.ID.String(),
		},
		IdempotencyKey: payment.ID.String(), // Используем ID платежа как idempotency key
	}

	// Вызываем YooKassa API
	yookassaResp, err := s.yookassaClient.CreatePayment(ctx, yookassaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment in YooKassa: %w", err)
	}

	// Обновляем запись платежа: yookassa_payment_id и confirmation_url
	if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, models.PaymentStatusPending, yookassaResp.ID); err != nil {
		return nil, fmt.Errorf("failed to update payment with YooKassa ID: %w", err)
	}

	if err := s.paymentRepo.UpdateConfirmationURL(ctx, payment.ID, yookassaResp.Confirmation.ConfirmationURL); err != nil {
		return nil, fmt.Errorf("failed to update confirmation URL: %w", err)
	}

	// Формируем ответ
	response := &models.PaymentResponse{
		PaymentID:       payment.ID,
		Amount:          amount,
		Credits:         req.Credits,
		ConfirmationURL: yookassaResp.Confirmation.ConfirmationURL,
	}

	return response, nil
}

// GetPaymentHistory возвращает историю платежей пользователя
func (s *PaymentService) GetPaymentHistory(ctx context.Context, userID uuid.UUID) ([]*models.Payment, error) {
	payments, err := s.paymentRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment history: %w", err)
	}

	return payments, nil
}

// ProcessPaymentSuccess обрабатывает успешный платеж (вызывается из webhook)
// Операция полностью атомарна:
// 1. Проверяется идемпотентность (был ли платеж уже обработан)
// 2. Обновляется статус платежа и добавляются кредиты в единой транзакции
// 3. При ошибке вся операция откатывается
func (s *PaymentService) ProcessPaymentSuccess(ctx context.Context, yookassaPaymentID string) error {
	// Получаем платеж по YooKassa ID
	payment, err := s.paymentRepo.GetByYooKassaID(ctx, yookassaPaymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment by YooKassa ID: %w", err)
	}

	// Проверяем идемпотентность ДО начала транзакции
	// Если платеж уже обработан (ProcessedAt != nil), пропускаем
	if payment.ProcessedAt != nil {
		log.Printf("Payment %s already processed at %v, skipping duplicate webhook", payment.ID, payment.ProcessedAt)
		return nil
	}

	// Идемпотency key - используем YooKassa payment ID для гарантии уникальности
	idempotencyKey := yookassaPaymentID

	// Начинаем транзакцию для атомарной обработки платежа + кредитов
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Гарантированный rollback при ошибке или успешный commit
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		// Игнорируем ошибку "tx is closed" - это нормально при успешном commit
		if rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
			log.Printf("failed to rollback transaction for payment %s: %v", payment.ID, rollbackErr)
		}
	}()

	// Обновляем статус платежа на succeeded ВНУТРИ транзакции
	if err := s.paymentRepo.UpdateStatusWithTx(ctx, tx, payment.ID, models.PaymentStatusSucceeded, yookassaPaymentID); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Добавляем кредиты пользователю ВНУТРИ ЭТОЙ ЖЕ транзакции
	// Это гарантирует, что если кредиты не добавятся, то статус платежа не обновится
	addCreditsReq := &models.AddCreditsRequest{
		UserID:      payment.UserID,
		Amount:      payment.Credits,
		Reason:      fmt.Sprintf("Пополнение через платеж #%s", payment.ID.String()),
		PerformedBy: payment.UserID, // Система от имени пользователя
	}

	if err := s.creditService.AddCreditsWithTx(ctx, tx, addCreditsReq); err != nil {
		return fmt.Errorf("failed to add credits in transaction: %w", err)
	}

	// Отмечаем платеж как обработанный (processed_at + idempotency_key)
	if err := s.paymentRepo.MarkAsProcessedWithTx(ctx, tx, payment.ID, idempotencyKey); err != nil {
		return fmt.Errorf("failed to mark payment as processed: %w", err)
	}

	// Фиксируем транзакцию - если здесь ошибка, defer откатит все операции
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// ТОЛЬКО ПОСЛЕ УСПЕШНОГО COMMIT обновляем метрики
	metrics.PaymentsProcessed.WithLabelValues("succeeded").Inc()
	metrics.PaymentAmountTotal.Add(float64(payment.Amount))

	log.Printf("Payment %s successfully processed: user=%s, credits=%d, amount=%s",
		payment.ID, utils.MaskUserID(payment.UserID), payment.Credits, utils.MaskAmount(int(payment.Amount)))

	return nil
}

// ProcessPaymentCancellation обрабатывает отмену платежа (вызывается из webhook)
func (s *PaymentService) ProcessPaymentCancellation(ctx context.Context, yookassaPaymentID string) error {
	// Получаем платеж по YooKassa ID
	payment, err := s.paymentRepo.GetByYooKassaID(ctx, yookassaPaymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment by YooKassa ID: %w", err)
	}

	// Обновляем статус на canceled
	if err := s.paymentRepo.UpdateStatus(ctx, payment.ID, models.PaymentStatusCancelled, yookassaPaymentID); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Обновляем метрики отмененного платежа
	metrics.PaymentsProcessed.WithLabelValues("cancelled").Inc()

	return nil
}

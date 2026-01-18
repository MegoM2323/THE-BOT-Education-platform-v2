package service

import (
	"context"
	"fmt"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/pkg/metrics"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// CreditService обрабатывает бизнес-логику для кредитов
type CreditService struct {
	pool       *pgxpool.Pool
	creditRepo *repository.CreditRepository
}

// NewCreditService создает новый CreditService
func NewCreditService(pool *pgxpool.Pool, creditRepo *repository.CreditRepository) *CreditService {
	return &CreditService{
		pool:       pool,
		creditRepo: creditRepo,
	}
}

// withSerializableTx выполняет операцию в транзакции с уровнем изоляции SERIALIZABLE
// SERIALIZABLE гарантирует полную изоляцию транзакций для финансовых операций с кредитами
// Если fn возвращает ошибку, транзакция откатывается
// Если commit не удаётся, возвращается ошибка commit
func (s *CreditService) withSerializableTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	// Начинаем транзакцию с уровнем изоляции SERIALIZABLE для предотвращения race conditions
	txOptions := pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	}
	tx, err := s.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return fmt.Errorf("failed to begin serializable transaction: %w", err)
	}

	// defer выполнится в конце функции
	// Если Commit уже был вызван успешно, Rollback вернёт ошибку "tx is closed"
	// которую мы игнорируем. Если был Commit с ошибкой, Rollback откатит транзакцию.
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		// Игнорируем ошибку "tx is closed" - это нормально, когда commit был успешен
		if rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
			log.Warn().
				Err(rollbackErr).
				Str("component", "CreditService").
				Msg("Failed to rollback transaction in withSerializableTx")
		}
	}()

	// Выполняем операцию в контексте транзакции
	if err := fn(ctx, tx); err != nil {
		// Транзакция откатится в defer
		return err
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		// defer откатит (вернёт ошибку "tx is closed" которую игнорируем)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetBalance получает баланс кредитов для пользователя
func (s *CreditService) GetBalance(ctx context.Context, userID uuid.UUID) (*models.Credit, error) {
	return s.creditRepo.GetBalance(ctx, userID)
}

// GetBalanceOptimized получает баланс кредитов для пользователя - оптимизированная версия
// Используется для частых запросов (например, от sidebar) для минимизации сетевого трафика
// Возвращает только число (баланс), без лишних данных
func (s *CreditService) GetBalanceOptimized(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.creditRepo.GetBalanceOptimized(ctx, userID)
}

// AddCredits добавляет кредиты на счет пользователя (операция администратора)
// Операция атомарна: либо обновляется баланс И создается запись транзакции,
// либо откатывается вся операция на случай ошибки
// Использует SERIALIZABLE isolation level для предотвращения race conditions
func (s *CreditService) AddCredits(ctx context.Context, req *models.AddCreditsRequest) error {
	// Проверяем запрос ДО начала транзакции
	if err := req.Validate(); err != nil {
		return fmt.Errorf("AddCredits validation failed: %w", err)
	}

	// Выполняем операцию в защищённой транзакции с SERIALIZABLE isolation
	if err := s.withSerializableTx(ctx, func(txCtx context.Context, tx pgx.Tx) error {
		// Блокируем и получаем баланс кредитов с FOR UPDATE
		credit, err := s.creditRepo.GetBalanceForUpdate(txCtx, tx, req.UserID)
		if err != nil {
			// КРИТИЧНО: Если записи о кредитах нет, это критичная ошибка
			// (должна быть инициализирована при создании пользователя)
			return fmt.Errorf("failed to get credit balance for user %s - credit account may not be initialized: %w", req.UserID.String(), err)
		}

		// Вычисляем новый баланс
		newBalance := credit.Balance + req.Amount

		// Проверяем что новый баланс не превысит максимальный лимит
		// Максимальный лимит = 10000 кредитов (бизнес-логика для предотвращения накопления)
		if newBalance > models.MaxBalance {
			return models.ErrBalanceExceeded
		}

		// Обновляем баланс
		if err := s.creditRepo.UpdateBalance(txCtx, tx, req.UserID, newBalance); err != nil {
			return fmt.Errorf("AddCredits: failed to update credit balance: %w", err)
		}

		// Создаем запись транзакции
		transaction := &models.CreditTransaction{
			UserID:        req.UserID,
			Amount:        req.Amount,
			OperationType: models.OperationTypeAdd,
			Reason:        req.Reason,
			PerformedBy:   uuid.NullUUID{UUID: req.PerformedBy, Valid: true},
			BalanceBefore: credit.Balance,
			BalanceAfter:  newBalance,
		}
		if err := s.creditRepo.CreateTransaction(txCtx, tx, transaction); err != nil {
			return fmt.Errorf("AddCredits: failed to create credit transaction: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	// Обновляем метрики добавления кредитов (ТОЛЬКО после успешного commit)
	metrics.CreditsAdded.Add(float64(req.Amount))

	return nil
}

// AddCreditsWithTx добавляет кредиты в рамках переданной транзакции
// Используется для операций платежей, где нужно гарантировать атомарность
func (s *CreditService) AddCreditsWithTx(ctx context.Context, tx pgx.Tx, req *models.AddCreditsRequest) error {
	// Проверяем запрос
	if err := req.Validate(); err != nil {
		return fmt.Errorf("AddCreditsWithTx validation failed: %w", err)
	}

	// Блокируем и получаем баланс кредитов в рамках транзакции
	credit, err := s.creditRepo.GetBalanceForUpdate(ctx, tx, req.UserID)
	if err != nil {
		return fmt.Errorf("AddCreditsWithTx: failed to get credit balance for user %s: %w", req.UserID.String(), err)
	}

	// Обновляем баланс
	newBalance := credit.Balance + req.Amount

	// Проверяем что новый баланс не превысит максимальный лимит
	if newBalance > models.MaxBalance {
		return models.ErrBalanceExceeded
	}

	if err := s.creditRepo.UpdateBalance(ctx, tx, req.UserID, newBalance); err != nil {
		return fmt.Errorf("AddCreditsWithTx: failed to update credit balance: %w", err)
	}

	// Создаем запись транзакции
	transaction := &models.CreditTransaction{
		UserID:        req.UserID,
		Amount:        req.Amount,
		OperationType: models.OperationTypeAdd,
		Reason:        req.Reason,
		PerformedBy:   uuid.NullUUID{UUID: req.PerformedBy, Valid: true},
		BalanceBefore: credit.Balance,
		BalanceAfter:  newBalance,
	}
	if err := s.creditRepo.CreateTransaction(ctx, tx, transaction); err != nil {
		return fmt.Errorf("AddCreditsWithTx: failed to create credit transaction: %w", err)
	}

	return nil
}

// DeductCredits списывает кредиты со счета пользователя (операция администратора)
// Операция атомарна: баланс проверяется и обновляется в одной транзакции,
// гарантируя отсутствие race condition при одновременных операциях
// Использует SERIALIZABLE isolation level для предотвращения race conditions
// ГАРАНТИЯ: баланс никогда не станет отрицательным
func (s *CreditService) DeductCredits(ctx context.Context, req *models.DeductCreditsRequest) error {
	// Проверяем запрос ДО начала транзакции
	if err := req.Validate(); err != nil {
		return fmt.Errorf("DeductCredits validation failed: %w", err)
	}

	// Выполняем операцию в защищённой транзакции с SERIALIZABLE isolation
	if err := s.withSerializableTx(ctx, func(txCtx context.Context, tx pgx.Tx) error {
		// Блокируем и получаем баланс кредитов с FOR UPDATE
		credit, err := s.creditRepo.GetBalanceForUpdate(txCtx, tx, req.UserID)
		if err != nil {
			// КРИТИЧНО: Если записи о кредитах нет, это критичная ошибка
			return fmt.Errorf("failed to get credit balance for user %s - credit account may not be initialized: %w", req.UserID.String(), err)
		}

		// Проверяем достаточность баланса ДО обновления (в рамках транзакции с блокировкой)
		if !credit.HasSufficientBalance(req.Amount) {
			return fmt.Errorf("DeductCredits: %w (have: %d, need: %d)", repository.ErrInsufficientCredits, credit.Balance, req.Amount)
		}

		// Вычисляем новый баланс
		newBalance := credit.Balance - req.Amount

		// Дополнительная проверка на отрицательный баланс (защита от ошибок в логике)
		if newBalance < 0 {
			return fmt.Errorf("DeductCredits: internal error - balance would become negative (have: %d, deduct: %d)", credit.Balance, req.Amount)
		}

		// Обновляем баланс
		if err := s.creditRepo.UpdateBalance(txCtx, tx, req.UserID, newBalance); err != nil {
			return fmt.Errorf("DeductCredits: failed to update credit balance: %w", err)
		}

		// Создаем запись транзакции
		transaction := &models.CreditTransaction{
			UserID:        req.UserID,
			Amount:        -req.Amount,
			OperationType: models.OperationTypeDeduct,
			Reason:        req.Reason,
			BalanceBefore: credit.Balance,
			BalanceAfter:  newBalance,
		}
		if req.BookingID != nil {
			transaction.BookingID = uuid.NullUUID{UUID: *req.BookingID, Valid: true}
		}
		if err := s.creditRepo.CreateTransaction(txCtx, tx, transaction); err != nil {
			return fmt.Errorf("DeductCredits: failed to create credit transaction: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	// Обновляем метрики списания кредитов (ТОЛЬКО после успешного commit)
	// Tracks всех операций списания, включая отмену бронирований и ручные операции администраторов
	metrics.CreditsDeducted.Add(float64(req.Amount))

	return nil
}

// RefundCredits возвращает кредиты на счет пользователя
// Операция атомарна: баланс и запись транзакции обновляются вместе
// Использует SERIALIZABLE isolation level для предотвращения race conditions
func (s *CreditService) RefundCredits(ctx context.Context, req *models.RefundCreditsRequest) error {
	// Проверяем запрос ДО начала транзакции
	if err := req.Validate(); err != nil {
		return fmt.Errorf("RefundCredits validation failed: %w", err)
	}

	// Выполняем операцию в защищённой транзакции с SERIALIZABLE isolation
	if err := s.withSerializableTx(ctx, func(txCtx context.Context, tx pgx.Tx) error {
		// Блокируем и получаем баланс кредитов с FOR UPDATE
		credit, err := s.creditRepo.GetBalanceForUpdate(txCtx, tx, req.UserID)
		if err != nil {
			// КРИТИЧНО: Если записи о кредитах нет, это критичная ошибка
			return fmt.Errorf("failed to get credit balance for user %s - credit account may not be initialized: %w", req.UserID.String(), err)
		}

		// Вычисляем новый баланс
		newBalance := credit.Balance + req.Amount

		// Проверяем что новый баланс не превысит максимальный лимит
		// Максимальный лимит = 10000 кредитов (бизнес-логика для предотвращения накопления)
		if newBalance > models.MaxBalance {
			return models.ErrBalanceExceeded
		}

		// Обновляем баланс
		if err := s.creditRepo.UpdateBalance(txCtx, tx, req.UserID, newBalance); err != nil {
			return fmt.Errorf("RefundCredits: failed to update credit balance: %w", err)
		}

		// Создаем запись транзакции
		transaction := &models.CreditTransaction{
			UserID:        req.UserID,
			Amount:        req.Amount,
			OperationType: models.OperationTypeRefund,
			Reason:        req.Reason,
			BalanceBefore: credit.Balance,
			BalanceAfter:  newBalance,
		}
		if req.BookingID != nil {
			transaction.BookingID = uuid.NullUUID{UUID: *req.BookingID, Valid: true}
		}
		if err := s.creditRepo.CreateTransaction(txCtx, tx, transaction); err != nil {
			return fmt.Errorf("RefundCredits: failed to create credit transaction: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	// Обновляем метрики возврата кредитов (ТОЛЬКО после успешного commit)
	// Tracks всех операций возврата, включая отмену бронирований и ручные операции администраторов
	metrics.CreditsRefunded.Add(float64(req.Amount))

	return nil
}

// GetTransactionHistory получает историю транзакций кредитов
func (s *CreditService) GetTransactionHistory(ctx context.Context, filter *models.GetCreditHistoryFilter) ([]*models.CreditTransactionWithUser, error) {
	return s.creditRepo.GetTransactionHistory(ctx, filter)
}

// GetAllStudentCredits получает все балансы кредитов для студентов
func (s *CreditService) GetAllStudentCredits(ctx context.Context) ([]map[string]interface{}, error) {
	return s.creditRepo.GetAllStudentCredits(ctx)
}

// GetAllStudentCreditsWithPagination получает кредиты студентов с пагинацией
func (s *CreditService) GetAllStudentCreditsWithPagination(ctx context.Context, offset, limit int) ([]map[string]interface{}, int, error) {
	return s.creditRepo.GetAllStudentCreditsWithPagination(ctx, offset, limit)
}

// GetAllStudentCreditsNoPagination получает все балансы кредитов студентов без пагинации
// Используется в админ-панели для полного списка всех студентов и их кредитов
func (s *CreditService) GetAllStudentCreditsNoPagination(ctx context.Context) ([]map[string]interface{}, error) {
	return s.creditRepo.GetAllStudentCreditsNoPagination(ctx)
}

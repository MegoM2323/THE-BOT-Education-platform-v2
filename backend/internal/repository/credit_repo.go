package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

// CreditRepository управляет операциями с базой данных для кредитов
type CreditRepository struct {
	db *sqlx.DB
}

// NewCreditRepository создает новый CreditRepository
func NewCreditRepository(db *sqlx.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

// GetBalance получает баланс кредитов для пользователя
// Если пользователь не найден в таблице credits, возвращает баланс = 0 (не ошибку!)
func (r *CreditRepository) GetBalance(ctx context.Context, userID uuid.UUID) (*models.Credit, error) {
	query := `
		SELECT ` + CreditSelectFields + `
		FROM credits
		WHERE user_id = $1
	`

	var credit models.Credit
	err := r.db.GetContext(ctx, &credit, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Для новых пользователей возвращаем нулевой баланс, а не ошибку
			// Это позволяет корректно проверять кредиты при применении шаблонов
			return &models.Credit{
				ID:        uuid.New(),
				UserID:    userID,
				Balance:   0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get credit balance: %w", err)
	}

	return &credit, nil
}

// GetBalanceOptimized получает баланс кредитов для пользователя - оптимизированная версия для sidebar
// Возвращает только user_id и balance (минимум данных для быстрого ответа)
// Используется для частых запросов от sidebar с кэшированием
func (r *CreditRepository) GetBalanceOptimized(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT balance
		FROM credits
		WHERE user_id = $1
	`

	var balance int
	err := r.db.GetContext(ctx, &balance, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Для новых пользователей возвращаем 0 баланс, а не ошибку
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get credit balance: %w", err)
	}

	return balance, nil
}

// GetBalanceForUpdate получает баланс кредитов с блокировкой строки для обновления
// Если запись кредитов не существует, создает её с balance = 0 и возвращает
// Это обеспечивает корректную работу для новых пользователей без записи credits
func (r *CreditRepository) GetBalanceForUpdate(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (*models.Credit, error) {
	// Сначала пытаемся создать запись если её нет (ON CONFLICT DO NOTHING)
	// Это обеспечивает атомарность - запись всегда будет существовать перед блокировкой
	insertQuery := `
		INSERT INTO credits (id, user_id, balance, created_at, updated_at)
		VALUES ($1, $2, 0, $3, $4)
		ON CONFLICT (user_id) DO NOTHING
	`
	now := time.Now()
	_, err := tx.Exec(ctx, insertQuery, uuid.New(), userID, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure credit record exists: %w", err)
	}

	// Теперь блокируем и получаем запись (она гарантированно существует)
	selectQuery := `
		SELECT ` + CreditSelectFields + `
		FROM credits
		WHERE user_id = $1
		FOR UPDATE
	`

	var credit models.Credit
	err = tx.QueryRow(ctx, selectQuery, userID).Scan(
		&credit.ID,
		&credit.UserID,
		&credit.Balance,
		&credit.CreatedAt,
		&credit.UpdatedAt,
	)
	if err != nil {
		// Эта ошибка не должна возникать после INSERT, но обрабатываем на всякий случай
		if err == pgx.ErrNoRows {
			return nil, ErrCreditNotFound
		}
		return nil, fmt.Errorf("failed to get credit balance for update: %w", err)
	}

	return &credit, nil
}

// UpdateBalance обновляет баланс кредитов в рамках транзакции
// SAFETY: Этот метод предполагает что:
// 1. Вызывается внутри транзакции с адекватным уровнем изоляции (минимум READ_COMMITTED)
// 2. Caller получил строку с FOR UPDATE блокировкой перед вызовом этого метода
// 3. Caller вычислил newBalance на основе currentBalance полученного с блокировкой
// Пример правильного использования (см. createRefundTransactionsInTx):
//
//	SELECT balance FROM credits WHERE user_id = $1 FOR UPDATE
//	newBalance := currentBalance + refundAmount
//	UpdateBalance(ctx, tx, userID, newBalance)
//	INSERT INTO credit_transactions ...
//
// Если эти условия не соблюдены, используйте относительное обновление:
//
//	UPDATE credits SET balance = balance + $1 WHERE user_id = $2
func (r *CreditRepository) UpdateBalance(ctx context.Context, tx pgx.Tx, userID uuid.UUID, newBalance int) error {
	query := `
		UPDATE credits
		SET balance = $1, updated_at = $2
		WHERE user_id = $3
	`

	result, err := tx.Exec(ctx, query, newBalance, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update credit balance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCreditNotFound
	}

	return nil
}

// CreateTransaction создает запись о транзакции кредитов
// ВАЖНО: Чувствительные данные (user_id, amount, balance) НЕ должны логироваться в открытом виде
// Используйте функции маскирования из internal/utils для логирования:
// - utils.MaskUserID() для скрытия user_id
// - utils.MaskAmount() для скрытия сумм
func (r *CreditRepository) CreateTransaction(ctx context.Context, tx pgx.Tx, transaction *models.CreditTransaction) error {
	query := `
		INSERT INTO credit_transactions (id, user_id, amount, operation_type, reason, performed_by, booking_id, balance_before, balance_after, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	transaction.ID = uuid.New()
	transaction.CreatedAt = time.Now()

	var performedBy interface{} = nil
	if transaction.PerformedBy.Valid {
		performedBy = transaction.PerformedBy.UUID
	}

	var bookingID interface{} = nil
	if transaction.BookingID.Valid {
		bookingID = transaction.BookingID.UUID
	}

	_, err := tx.Exec(ctx, query,
		transaction.ID,
		transaction.UserID,
		transaction.Amount,
		transaction.OperationType,
		transaction.Reason,
		performedBy,
		bookingID,
		transaction.BalanceBefore,
		transaction.BalanceAfter,
		transaction.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create credit transaction: %w", err)
	}

	return nil
}

// GetTransactionHistory получает историю транзакций кредитов с пагинацией
// ВСЕГДА применяет LIMIT и OFFSET для предотвращения исчерпания памяти:
// - Если filter == nil или filter.Limit == 0, применяется DefaultTransactionLimit (50)
// - Если filter.Limit > MaxTransactionLimit (500), капируется до MaxTransactionLimit
// - Результат всегда содержит <= 500 строк в памяти
func (r *CreditRepository) GetTransactionHistory(ctx context.Context, filter *models.GetCreditHistoryFilter) ([]*models.CreditTransactionWithUser, error) {
	query := `
		SELECT
			ct.id, ct.user_id, ct.amount, ct.operation_type, ct.reason,
			ct.booking_id, ct.balance_before, ct.balance_after, ct.created_at,
			u.email as user_email, u.full_name as user_name,
			COALESCE(p.email, '') as performed_by_email
		FROM credit_transactions ct
		JOIN users u ON ct.user_id = u.id
		LEFT JOIN users p ON ct.performed_by = p.id
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	// Применяем фильтры
	if filter != nil {
		if filter.UserID != nil {
			query += fmt.Sprintf(` AND ct.user_id = $%d`, argIndex)
			args = append(args, *filter.UserID)
			argIndex++
		}
		if filter.OperationType != nil {
			query += fmt.Sprintf(` AND ct.operation_type = $%d`, argIndex)
			args = append(args, *filter.OperationType)
			argIndex++
		}
		if filter.StartDate != nil {
			query += fmt.Sprintf(` AND ct.created_at >= $%d`, argIndex)
			args = append(args, *filter.StartDate)
			argIndex++
		}
		if filter.EndDate != nil {
			query += fmt.Sprintf(` AND ct.created_at <= $%d`, argIndex)
			args = append(args, *filter.EndDate)
			argIndex++
		}
	}

	// Order by created_at DESC, then by id DESC for stable pagination
	// When multiple transactions have the same created_at timestamp, id ensures consistent ordering
	query += ` ORDER BY ct.created_at DESC, ct.id DESC`

	// Определяем лимит с дефолтом и максимумом
	limit := models.DefaultTransactionLimit
	if filter != nil && filter.Limit > 0 {
		limit = filter.Limit
	}
	// Капируем лимит до максимума
	if limit > models.MaxTransactionLimit {
		limit = models.MaxTransactionLimit
	}

	// Определяем offset (по умолчанию 0)
	offset := 0
	if filter != nil && filter.Offset > 0 {
		offset = filter.Offset
	}

	// Применяем LIMIT и OFFSET (обязательно!)
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	args = append(args, limit, offset)

	var transactions []*models.CreditTransactionWithUser
	err := r.db.SelectContext(ctx, &transactions, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}

	// Гарантируем, что результат всегда slice (не nil)
	if transactions == nil {
		transactions = []*models.CreditTransactionWithUser{}
	}

	return transactions, nil
}

// GetTransactionByBooking получает транзакции кредитов для конкретного бронирования
func (r *CreditRepository) GetTransactionByBooking(ctx context.Context, bookingID uuid.UUID) ([]*models.CreditTransaction, error) {
	query := `
		SELECT id, user_id, amount, operation_type, reason, performed_by, booking_id, created_at
		FROM credit_transactions
		WHERE booking_id = $1
		ORDER BY created_at DESC, id DESC
	`

	var transactions []*models.CreditTransaction
	err := r.db.SelectContext(ctx, &transactions, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by booking: %w", err)
	}

	return transactions, nil
}

// CreateCredit создает новый аккаунт кредитов для пользователя
// Использует ON CONFLICT для безопасной обработки дублирования
// Возвращает ErrDuplicateCredit если кредиты для пользователя уже существуют
func (r *CreditRepository) CreateCredit(ctx context.Context, userID uuid.UUID, initialBalance int) error {
	query := `
		INSERT INTO credits (id, user_id, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO NOTHING
	`

	result, err := r.db.ExecContext(ctx, query,
		uuid.New(),
		userID,
		initialBalance,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create credit account: %w", err)
	}

	// Проверяем, была ли строка вставлена (0 значит дублирование из-за ON CONFLICT)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrDuplicateCredit
	}

	return nil
}

// GetAllStudentCredits получает все балансы кредитов для студентов
// Returns: slice of {user_id, email, full_name, balance}
func (r *CreditRepository) GetAllStudentCredits(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT
			u.id as user_id,
			u.email,
			u.full_name,
			COALESCE(c.balance, 0) as balance
		FROM users u
		LEFT JOIN credits c ON u.id = c.user_id
		WHERE u.deleted_at IS NULL
			AND u.role = 'student'
		ORDER BY u.full_name
	`

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get student credits: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		result := make(map[string]interface{})
		err := rows.MapScan(result)
		if err != nil {
			return nil, fmt.Errorf("failed to scan student credit: %w", err)
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading student credits: %w", err)
	}

	return results, nil
}

// GetAllStudentCreditsWithPagination получает балансы кредитов для студентов с пагинацией
// Returns: slice of {user_id, email, full_name, balance} и общее количество студентов
func (r *CreditRepository) GetAllStudentCreditsWithPagination(ctx context.Context, offset, limit int) ([]map[string]interface{}, int, error) {
	// Подсчет общего количества студентов
	countQuery := `
		SELECT COUNT(*)
		FROM users u
		WHERE u.deleted_at IS NULL
			AND u.role = 'student'
	`

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count students: %w", err)
	}

	// Получение списка студентов и их балансов с LIMIT/OFFSET
	query := `
		SELECT
			u.id as user_id,
			u.email,
			u.full_name,
			COALESCE(c.balance, 0) as balance
		FROM users u
		LEFT JOIN credits c ON u.id = c.user_id
		WHERE u.deleted_at IS NULL
			AND u.role = 'student'
		ORDER BY u.full_name
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get student credits: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		result := make(map[string]interface{})
		err := rows.MapScan(result)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan student credit: %w", err)
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error reading student credits: %w", err)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}

	return results, total, nil
}

// GetAllStudentCreditsNoPagination получает ВСЕ балансы кредитов для студентов БЕЗ пагинации
// Используется для экспорта или обработки всех студентов без ограничений
// Returns: slice of {user_id, email, full_name, balance}, отсортированный по full_name
// ВАЖНО: Может вернуть большой объем данных, используйте с осторожностью на больших базах
func (r *CreditRepository) GetAllStudentCreditsNoPagination(ctx context.Context) ([]map[string]interface{}, error) {
	query := `
		SELECT
			u.id as user_id,
			u.email,
			u.full_name,
			COALESCE(c.balance, 0) as balance
		FROM users u
		LEFT JOIN credits c ON u.id = c.user_id
		WHERE u.deleted_at IS NULL
			AND u.role = 'student'
		ORDER BY u.full_name ASC
	`

	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get student credits: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		result := make(map[string]interface{})
		err := rows.MapScan(result)
		if err != nil {
			return nil, fmt.Errorf("failed to scan student credit: %w", err)
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading student credits: %w", err)
	}

	// Гарантируем что результат всегда slice, не nil
	if results == nil {
		results = []map[string]interface{}{}
	}

	return results, nil
}

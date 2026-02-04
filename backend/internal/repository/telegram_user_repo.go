package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// TelegramUserRepository интерфейс для работы с привязками пользователей к Telegram
type TelegramUserRepository interface {
	LinkUserToTelegram(ctx context.Context, userID uuid.UUID, telegramID, chatID int64, username string) error
	LinkUserToTelegramAtomic(ctx context.Context, userID uuid.UUID, telegramID, chatID int64, username string) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.TelegramUser, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*models.TelegramUser, error)
	GetAllLinked(ctx context.Context) ([]*models.TelegramUser, error)
	GetAllWithUserInfo(ctx context.Context) ([]*models.TelegramUser, error)
	GetByRoleWithUserInfo(ctx context.Context, role string) ([]*models.TelegramUser, error)
	GetByRole(ctx context.Context, role string) ([]*models.TelegramUser, error)
	GetSubscribedUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error)
	UpdateSubscription(ctx context.Context, userID uuid.UUID, subscribed bool) error
	UnlinkTelegram(ctx context.Context, userID uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	CleanupInvalidLinks(ctx context.Context) (int64, error)
	IsValidlyLinked(ctx context.Context, userID uuid.UUID) (bool, error)
}

// TelegramUserRepo реализация TelegramUserRepository
type TelegramUserRepo struct {
	db *sqlx.DB
}

// NewTelegramUserRepository создает новый экземпляр TelegramUserRepo
func NewTelegramUserRepository(db *sqlx.DB) TelegramUserRepository {
	return &TelegramUserRepo{db: db}
}

// LinkUserToTelegram привязывает пользователя к Telegram аккаунту
// Использует SELECT и DELETE/INSERT для безопасной работы без ON CONFLICT
func (r *TelegramUserRepo) LinkUserToTelegram(ctx context.Context, userID uuid.UUID, telegramID, chatID int64, username string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем, есть ли уже запись для этого пользователя
	var existingID sql.NullString
	checkQuery := `SELECT id FROM telegram_users WHERE user_id = $1`
	err = tx.QueryRowContext(ctx, checkQuery, userID).Scan(&existingID)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	now := time.Now()

	if existingID.Valid {
		// Обновляем существующую запись
		updateQuery := `
			UPDATE telegram_users
			SET telegram_id = $1, chat_id = $2, username = $3, updated_at = $4
			WHERE user_id = $5
		`
		_, err = tx.ExecContext(ctx, updateQuery, telegramID, chatID, username, now, userID)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				if pqErr.Constraint == "telegram_users_telegram_id_key" {
					return ErrTelegramUserAlreadyLinked
				}
			}
			return fmt.Errorf("failed to update telegram user: %w", err)
		}
	} else {
		// Вставляем новую запись
		insertQuery := `
			INSERT INTO telegram_users (id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = tx.ExecContext(ctx, insertQuery,
			uuid.New(),
			userID,
			telegramID,
			chatID,
			username,
			true, // по умолчанию подписан на уведомления
			now,
			now,
		)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				if pqErr.Constraint == "telegram_users_telegram_id_key" {
					return ErrTelegramUserAlreadyLinked
				}
			}
			return fmt.Errorf("failed to link user to telegram: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// LinkUserToTelegramAtomic привязывает пользователя к Telegram аккаунту атомарно с проверкой на дублирование.
// Использует SELECT FOR UPDATE для защиты от race condition при одновременных запросах.
// Гарантирует, что:
// 1. Только один пользователь может привязать данный telegram_id (UNIQUE constraint в БД)
// 2. При race condition (два запроса одновременно) второй получит ErrTelegramIDAlreadyLinked
// 3. При обновлении существующей привязки - успешно (идемпотентно)
func (r *TelegramUserRepo) LinkUserToTelegramAtomic(ctx context.Context, userID uuid.UUID, telegramID, chatID int64, username string) error {
	// Используем транзакцию для атомарности операций
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Шаг 1: Проверяем, привязан ли telegram_id к другому пользователю (с row lock)
	var existingUserID sql.NullString
	lockQuery := `
		SELECT user_id FROM telegram_users
		WHERE telegram_id = $1
		FOR UPDATE
	`
	err = tx.QueryRowContext(ctx, lockQuery, telegramID).Scan(&existingUserID)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check telegram_id lock: %w", err)
	}

	// Если telegram_id привязан к другому пользователю - ошибка
	if err == nil && existingUserID.Valid {
		existingID, _ := uuid.Parse(existingUserID.String)
		if existingID != userID {
			return ErrTelegramIDAlreadyLinked
		}
	}

	// Шаг 2: Проверяем, привязан ли уже текущий пользователь (с row lock)
	var existingID sql.NullString
	userLockQuery := `
		SELECT id FROM telegram_users
		WHERE user_id = $1
		FOR UPDATE
	`
	err = tx.QueryRowContext(ctx, userLockQuery, userID).Scan(&existingID)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check user lock: %w", err)
	}

	// Шаг 3: Вставляем или обновляем привязку
	now := time.Now()

	// Если запись уже существует для этого пользователя - обновляем
	if err == nil && existingID.Valid {
		updateQuery := `
			UPDATE telegram_users
			SET telegram_id = $1, chat_id = $2, username = $3, updated_at = $4
			WHERE user_id = $5
		`
		_, err = tx.ExecContext(ctx, updateQuery, telegramID, chatID, username, now, userID)
		if err != nil {
			// Проверка на duplicate key для telegram_id (могла произойти race condition)
			if isDuplicateKeyError(err, "telegram_users_telegram_id_key") {
				return ErrTelegramIDAlreadyLinked
			}
			return fmt.Errorf("failed to update telegram user: %w", err)
		}
	} else {
		// Если записи нет - вставляем новую
		insertQuery := `
			INSERT INTO telegram_users (id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = tx.ExecContext(ctx, insertQuery,
			uuid.New(),
			userID,
			telegramID,
			chatID,
			username,
			true, // по умолчанию подписан на уведомления
			now,
			now,
		)

		if err != nil {
			// Проверка на duplicate key для telegram_id (могла произойти race condition)
			if isDuplicateKeyError(err, "telegram_users_telegram_id_key") {
				return ErrTelegramIDAlreadyLinked
			}
			return fmt.Errorf("failed to link user to telegram: %w", err)
		}
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByUserID получает привязку Telegram по ID пользователя
func (r *TelegramUserRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.TelegramUser, error) {
	query := `
		SELECT id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at
		FROM telegram_users
		WHERE user_id = $1
	`

	var telegramUser models.TelegramUser
	err := r.db.GetContext(ctx, &telegramUser, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTelegramUserNotFound
		}
		return nil, fmt.Errorf("failed to get telegram user by user ID: %w", err)
	}

	return &telegramUser, nil
}

// GetByTelegramID получает привязку Telegram по Telegram ID
func (r *TelegramUserRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*models.TelegramUser, error) {
	query := `
		SELECT id, user_id, telegram_id, chat_id, username, subscribed, created_at, updated_at
		FROM telegram_users
		WHERE telegram_id = $1
	`

	var telegramUser models.TelegramUser
	err := r.db.GetContext(ctx, &telegramUser, query, telegramID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTelegramUserNotFound
		}
		return nil, fmt.Errorf("failed to get telegram user by telegram ID: %w", err)
	}

	return &telegramUser, nil
}

// GetAllLinked получает все привязанные Telegram аккаунты (без JOIN, обогащение данных в service layer)
func (r *TelegramUserRepo) GetAllLinked(ctx context.Context) ([]*models.TelegramUser, error) {
	query := `
		SELECT
			tu.id,
			tu.user_id,
			tu.telegram_id,
			tu.chat_id,
			tu.username,
			tu.subscribed,
			tu.created_at,
			tu.updated_at
		FROM telegram_users tu
		INNER JOIN users u ON tu.user_id = u.id
		WHERE u.deleted_at IS NULL
		ORDER BY tu.created_at DESC
	`

	var telegramUsers []*models.TelegramUser
	err := r.db.SelectContext(ctx, &telegramUsers, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all linked telegram users: %w", err)
	}

	return telegramUsers, nil
}

// GetByRole получает привязанные Telegram аккаунты пользователей с определенной ролью (без JOIN, обогащение в service layer)
func (r *TelegramUserRepo) GetByRole(ctx context.Context, role string) ([]*models.TelegramUser, error) {
	query := `
		SELECT
			tu.id,
			tu.user_id,
			tu.telegram_id,
			tu.chat_id,
			tu.username,
			tu.subscribed,
			tu.created_at,
			tu.updated_at
		FROM telegram_users tu
		INNER JOIN users u ON tu.user_id = u.id
		WHERE u.role = $1 AND u.deleted_at IS NULL
		ORDER BY tu.created_at DESC
	`

	var telegramUsers []*models.TelegramUser
	err := r.db.SelectContext(ctx, &telegramUsers, query, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get telegram users by role: %w", err)
	}

	return telegramUsers, nil
}

// UpdateSubscription обновляет статус подписки на уведомления
func (r *TelegramUserRepo) UpdateSubscription(ctx context.Context, userID uuid.UUID, subscribed bool) error {
	query := `
		UPDATE telegram_users
		SET subscribed = $1, updated_at = $2
		WHERE user_id = $3
	`

	result, err := r.db.ExecContext(ctx, query, subscribed, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTelegramUserNotFound
	}

	return nil
}

// UnlinkTelegram удаляет привязку пользователя к Telegram
func (r *TelegramUserRepo) UnlinkTelegram(ctx context.Context, userID uuid.UUID) error {
	query := `
		DELETE FROM telegram_users
		WHERE user_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to unlink telegram: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTelegramUserNotFound
	}

	return nil
}

// DeleteByUserID полностью удаляет запись пользователя из telegram_users (для очистки невалидных записей)
func (r *TelegramUserRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		DELETE FROM telegram_users
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete telegram user: %w", err)
	}

	return nil
}

// CleanupInvalidLinks удаляет все записи с невалидным telegram_id (0 или NULL)
// Возвращает количество удаленных записей
func (r *TelegramUserRepo) CleanupInvalidLinks(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM telegram_users
		WHERE telegram_id IS NULL OR telegram_id = 0
	`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup invalid links: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// IsValidlyLinked проверяет, валидно ли привязан пользователь (telegram_id > 0)
func (r *TelegramUserRepo) IsValidlyLinked(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM telegram_users
			WHERE user_id = $1 AND telegram_id > 0
		)
	`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check valid link: %w", err)
	}

	return exists, nil
}

// GetAllWithUserInfo получает все привязанные Telegram аккаунты с полной информацией о пользователях
// Использует LEFT JOIN для обработки случаев, когда пользователь удален
func (r *TelegramUserRepo) GetAllWithUserInfo(ctx context.Context) ([]*models.TelegramUser, error) {
	query := `
		SELECT
			tu.id,
			tu.user_id,
			tu.telegram_id,
			tu.chat_id,
			tu.username,
			tu.subscribed,
			tu.created_at,
			tu.updated_at,
			u.id AS "user.id",
			u.email AS "user.email",
			u.full_name AS "user.full_name",
			u.role AS "user.role",
			u.payment_enabled AS "user.payment_enabled",
			u.telegram_username AS "user.telegram_username",
			u.created_at AS "user.created_at",
			u.updated_at AS "user.updated_at",
			u.deleted_at AS "user.deleted_at"
		FROM telegram_users tu
		LEFT JOIN users u ON tu.user_id = u.id
		WHERE tu.telegram_id > 0
		ORDER BY tu.created_at DESC
	`

	// Структура для сканирования результатов с вложенными данными
	type telegramUserWithUser struct {
		models.TelegramUser
		User models.User `db:"user"`
	}

	var results []telegramUserWithUser
	err := r.db.SelectContext(ctx, &results, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all linked telegram users with user info: %w", err)
	}

	// Преобразуем результаты в нужный формат
	telegramUsers := make([]*models.TelegramUser, len(results))
	for i := range results {
		telegramUsers[i] = &results[i].TelegramUser
		// Если пользователь найден (не удален), присваиваем данные
		if results[i].User.ID != uuid.Nil {
			telegramUsers[i].User = &results[i].User
		} else {
			// Если пользователь удален или не найден, оставляем User = nil
			telegramUsers[i].User = nil
		}
	}

	return telegramUsers, nil
}

// GetByRoleWithUserInfo получает привязанные Telegram аккаунты пользователей с определенной ролью и полной информацией
// Использует LEFT JOIN для обработки случаев, когда пользователь удален
func (r *TelegramUserRepo) GetByRoleWithUserInfo(ctx context.Context, role string) ([]*models.TelegramUser, error) {
	query := `
		SELECT
			tu.id,
			tu.user_id,
			tu.telegram_id,
			tu.chat_id,
			tu.username,
			tu.subscribed,
			tu.created_at,
			tu.updated_at,
			u.id AS "user.id",
			u.email AS "user.email",
			u.full_name AS "user.full_name",
			u.role AS "user.role",
			u.payment_enabled AS "user.payment_enabled",
			u.telegram_username AS "user.telegram_username",
			u.created_at AS "user.created_at",
			u.updated_at AS "user.updated_at",
			u.deleted_at AS "user.deleted_at"
		FROM telegram_users tu
		LEFT JOIN users u ON tu.user_id = u.id
		WHERE tu.telegram_id > 0 AND u.role = $1 AND u.deleted_at IS NULL
		ORDER BY tu.created_at DESC
	`

	// Структура для сканирования результатов с вложенными данными
	type telegramUserWithUser struct {
		models.TelegramUser
		User models.User `db:"user"`
	}

	var results []telegramUserWithUser
	err := r.db.SelectContext(ctx, &results, query, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get telegram users by role with user info: %w", err)
	}

	// Преобразуем результаты в нужный формат
	telegramUsers := make([]*models.TelegramUser, len(results))
	for i := range results {
		telegramUsers[i] = &results[i].TelegramUser
		// Если пользователь найден (не удален), присваиваем данные
		if results[i].User.ID != uuid.Nil {
			telegramUsers[i].User = &results[i].User
		} else {
			// Если пользователь удален или не найден, оставляем User = nil
			telegramUsers[i].User = nil
		}
	}

	return telegramUsers, nil
}

// GetSubscribedUserIDs возвращает список userIDs, которые привязаны к Telegram и подписаны (subscribed=true)
func (r *TelegramUserRepo) GetSubscribedUserIDs(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(userIDs) == 0 {
		return []uuid.UUID{}, nil
	}

	query := `SELECT user_id FROM telegram_users WHERE user_id = ANY($1) AND subscribed = true`

	var subscribedIDs []uuid.UUID
	err := r.db.SelectContext(ctx, &subscribedIDs, query, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribed user IDs: %w", err)
	}

	return subscribedIDs, nil
}

// isDuplicateKeyError проверяет, является ли ошибка нарушением уникального ключа
// Поддерживает как lib/pq (prod), так и pgx (tests) драйверы
func isDuplicateKeyError(err error, constraintName string) bool {
	if err == nil {
		return false
	}

	// Проверяем pgx драйвер (используется в тестах)
	if pgxErr, ok := err.(*pgconn.PgError); ok {
		if pgxErr.Code == "23505" {
			if constraintName == "" || pgxErr.ConstraintName == constraintName {
				return true
			}
		}
		return false
	}

	// Проверяем lib/pq драйвер (может использоваться в production)
	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code == "23505" {
			if constraintName == "" || pqErr.Constraint == constraintName {
				return true
			}
		}
		return false
	}

	// Fallback: проверяем текст ошибки для случаев обертывания
	errStr := err.Error()
	if strings.Contains(errStr, "23505") {
		if constraintName == "" || strings.Contains(errStr, constraintName) {
			return true
		}
	}

	return false
}

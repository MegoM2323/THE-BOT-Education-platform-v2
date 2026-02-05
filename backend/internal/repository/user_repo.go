package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, roleFilter *models.UserRole) ([]*models.User, error)
	ListWithPagination(ctx context.Context, roleFilter *models.UserRole, offset, limit int) ([]*models.User, int, error)
	Exists(ctx context.Context, email string) (bool, error)
	UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error
	GetParentChatIDsByStudentIDs(ctx context.Context, studentIDs []uuid.UUID) (map[uuid.UUID]int64, error)
}

// UserRepo реализация UserRepository
type UserRepo struct {
	db *sqlx.DB
}

// NewUserRepository создает новый UserRepository
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &UserRepo{db: db}
}

// Create создает нового пользователя
func (r *UserRepo) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, full_name, role, payment_enabled, telegram_username, parent_telegram_username, parent_chat_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.PaymentEnabled,
		user.TelegramUsername,
		user.ParentTelegramUsername,
		user.ParentChatID,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID получает пользователя по ID
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT ` + UserSelectFields + `
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetByEmail получает пользователя по email
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT ` + UserSelectFields + `
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// List получает всех пользователей с опциональным фильтром по роли
func (r *UserRepo) List(ctx context.Context, role *models.UserRole) ([]*models.User, error) {
	query := `
		SELECT ` + UserSelectFields + `
		FROM users
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	if role != nil {
		query += ` AND role = $1`
		args = append(args, *role)
	}

	query += ` ORDER BY created_at DESC`

	var users []*models.User
	err := r.db.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// ListWithPagination возвращает список пользователей с пагинацией
func (r *UserRepo) ListWithPagination(ctx context.Context, role *models.UserRole, offset, limit int) ([]*models.User, int, error) {
	// Валидация параметров пагинации
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 20 // значение по умолчанию
	}
	if limit > 100 {
		limit = 100 // максимальный лимит
	}

	// Получаем полное количество записей (без LIMIT/OFFSET)
	countQuery := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	countArgs := []interface{}{}

	if role != nil {
		countQuery += ` AND role = $1`
		countArgs = append(countArgs, *role)
	}

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Получаем отсортированный список с LIMIT/OFFSET
	query := `
		SELECT ` + UserSelectFields + `
		FROM users
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	if role != nil {
		query += ` AND role = $1`
		args = append(args, *role)
	}

	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	var users []*models.User
	if err := r.db.SelectContext(ctx, &users, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	if users == nil {
		users = []*models.User{}
	}

	return users, total, nil
}

// Update обновляет пользователя
// Поддерживаемые поля: email, full_name, role, telegram_username, payment_enabled
func (r *UserRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Whitelist of allowed fields to prevent SQL injection
	allowedFields := map[string]bool{
		"email":                    true,
		"full_name":                true,
		"role":                     true,
		"telegram_username":        true,
		"parent_telegram_username": true,
		"parent_chat_id":           true,
		"payment_enabled":          true,
	}

	// Валидация полей до начала транзакции
	for field := range updates {
		if !allowedFields[field] {
			return fmt.Errorf("invalid field for update: %s", field)
		}
	}

	// Check if db is nil (shouldn't happen in production, but needed for tests)
	if r.db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Начинаем транзакцию для атомарности операции
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем, что пользователь существует и не удален
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`
	if err := tx.QueryRowContext(ctx, checkQuery, id).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	query := `UPDATE users SET updated_at = $1`
	args := []interface{}{time.Now()}
	argIndex := 2

	for field, value := range updates {
		// Safely append field update with parameterized query
		query += fmt.Sprintf(`, %s = $%d`, field, argIndex)
		args = append(args, value)
		argIndex++
	}

	query += fmt.Sprintf(` WHERE id = $%d AND deleted_at IS NULL`, argIndex)
	args = append(args, id)

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdatePassword обновляет хеш пароля пользователя
func (r *UserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	// Начинаем транзакцию для атомарности операции
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем, что пользователь существует и не удален
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`
	if err := tx.QueryRowContext(ctx, checkQuery, id).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := tx.ExecContext(ctx, query, passwordHash, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete выполняет мягкое удаление пользователя с проверкой зависимостей
func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var user struct {
		Role models.UserRole `db:"role"`
	}
	checkQuery := `SELECT role FROM users WHERE id = $1 AND deleted_at IS NULL`
	if err := tx.QueryRowContext(ctx, checkQuery, id).Scan(&user.Role); err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user role: %w", err)
	}

	if user.Role == models.RoleMethodologist {
		var hasActiveLessons bool
		err := tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM lessons WHERE teacher_id = $1 AND deleted_at IS NULL)`, id).Scan(&hasActiveLessons)
		if err != nil {
			return fmt.Errorf("check lessons: %w", err)
		}
		if hasActiveLessons {
			return ErrUserHasActiveLessons
		}
	}

	if user.Role == models.RoleStudent {
		var hasActiveBookings bool
		err := tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM bookings WHERE student_id = $1 AND status = 'active')`, id).Scan(&hasActiveBookings)
		if err != nil {
			return fmt.Errorf("check bookings: %w", err)
		}
		if hasActiveBookings {
			return ErrUserHasActiveBookings
		}
	}

	var hasPayments bool
	err = tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM payments WHERE user_id = $1)`, id).Scan(&hasPayments)
	if err != nil {
		return fmt.Errorf("check payments: %w", err)
	}
	if hasPayments {
		return ErrUserHasPayments
	}

	now := time.Now()
	query := `
		UPDATE users
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := tx.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SoftDelete выполняет мягкое удаление пользователя без проверок зависимостей
func (r *UserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Exists проверяет, существует ли пользователь с указанным email
func (r *UserRepo) Exists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// UpdateTelegramUsername обновляет telegram_username пользователя (синхронизация с telegram_users.username)
// Возвращает ErrTelegramUsernameInUse если Telegram имя уже используется другим пользователем
func (r *UserRepo) UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error {
	// Начинаем транзакцию для атомарности операции
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем, что пользователь существует и не удален
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`
	if err := tx.QueryRowContext(ctx, checkQuery, userID).Scan(&exists); err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	query := `
		UPDATE users
		SET telegram_username = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := tx.ExecContext(ctx, query, username, time.Now(), userID)
	if err != nil {
		// Проверяем, является ли ошибка нарушением UNIQUE constraint на telegram_username
		if IsUniqueViolationError(err) {
			return ErrTelegramUsernameInUse
		}
		return fmt.Errorf("failed to update telegram username: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrUserNotFound
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetParentChatIDsByStudentIDs получает parent_chat_id для списка студентов
func (r *UserRepo) GetParentChatIDsByStudentIDs(ctx context.Context, studentIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	query := `
		SELECT id, parent_chat_id
		FROM users
		WHERE id = ANY($1)
		  AND parent_chat_id IS NOT NULL
		  AND deleted_at IS NULL
	`
	rows, err := r.db.QueryContext(ctx, query, studentIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query parent chat IDs: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]int64)
	for rows.Next() {
		var id uuid.UUID
		var parentChatID sql.NullInt64
		if err := rows.Scan(&id, &parentChatID); err != nil {
			return nil, fmt.Errorf("failed to scan parent chat ID: %w", err)
		}
		if parentChatID.Valid {
			result[id] = parentChatID.Int64
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return result, nil
}

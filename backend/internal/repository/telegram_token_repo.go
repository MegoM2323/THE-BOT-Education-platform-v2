package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// TelegramTokenRepository интерфейс для работы с токенами привязки Telegram
type TelegramTokenRepository interface {
	SaveToken(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error
	GetTokenUser(ctx context.Context, token string) (uuid.UUID, error)
	DeleteToken(ctx context.Context, token string) error
	DeleteExpiredTokens(ctx context.Context) (int64, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

// TelegramTokenRepo реализация TelegramTokenRepository
type TelegramTokenRepo struct {
	db *sqlx.DB
}

// NewTelegramTokenRepository создает новый экземпляр TelegramTokenRepo
func NewTelegramTokenRepository(db *sqlx.DB) TelegramTokenRepository {
	return &TelegramTokenRepo{db: db}
}

// SaveToken сохраняет новый токен привязки
func (r *TelegramTokenRepo) SaveToken(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error {
	query := `
		INSERT INTO telegram_tokens (id, token, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		uuid.New(),
		token,
		userID,
		expiresAt,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// GetTokenUser получает UUID пользователя по токену и проверяет его валидность
func (r *TelegramTokenRepo) GetTokenUser(ctx context.Context, token string) (uuid.UUID, error) {
	query := `
		SELECT user_id
		FROM telegram_tokens
		WHERE token = $1 AND expires_at > CURRENT_TIMESTAMP
	`

	var userID uuid.UUID
	err := r.db.GetContext(ctx, &userID, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("invalid or expired token")
		}
		return uuid.Nil, fmt.Errorf("failed to get token user: %w", err)
	}

	return userID, nil
}

// DeleteToken удаляет токен по его значению
func (r *TelegramTokenRepo) DeleteToken(ctx context.Context, token string) error {
	query := `DELETE FROM telegram_tokens WHERE token = $1`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

// DeleteExpiredTokens удаляет все истекшие токены и возвращает их количество
func (r *TelegramTokenRepo) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	query := `DELETE FROM telegram_tokens WHERE expires_at <= CURRENT_TIMESTAMP`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// DeleteByUserID удаляет все токены пользователя
func (r *TelegramTokenRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM telegram_tokens WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete tokens for user: %w", err)
	}

	return nil
}

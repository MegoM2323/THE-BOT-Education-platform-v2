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

// SessionRepository управляет операциями с базой данных для сессий
type SessionRepository struct {
	db *sqlx.DB
}

// NewSessionRepository создает новый SessionRepository
func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create создает новую сессию
func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, created_at, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	session.ID = uuid.New()
	session.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.CreatedAt,
		session.ExpiresAt,
		session.IPAddress,
		session.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID получает сессию по ID
func (r *SessionRepository) GetByID(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	query := `
		SELECT id, user_id, created_at, expires_at, ip_address, user_agent
		FROM sessions
		WHERE id = $1
	`

	var session models.Session
	err := r.db.GetContext(ctx, &session, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}

	return &session, nil
}

// GetWithUser получает сессию с информацией о пользователе
func (r *SessionRepository) GetWithUser(ctx context.Context, sessionID uuid.UUID) (*models.SessionWithUser, error) {
	query := `
		SELECT
			s.id, s.user_id, s.created_at, s.expires_at, s.ip_address, s.user_agent,
			u.email as user_email, u.full_name as user_name, u.role as user_role
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = $1 AND u.deleted_at IS NULL
	`

	var session models.SessionWithUser
	err := r.db.GetContext(ctx, &session, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session with user: %w", err)
	}

	return &session, nil
}

// Delete удаляет сессию по ID
func (r *SessionRepository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// DeleteByUserID удаляет все сессии пользователя
func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions by user ID: %w", err)
	}

	return nil
}

// DeleteExpired удаляет все истекшие сессии
func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < $1`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

// ListByUserID получает все активные сессии пользователя
func (r *SessionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	query := `
		SELECT id, user_id, created_at, expires_at, ip_address, user_agent
		FROM sessions
		WHERE user_id = $1 AND expires_at > $2
		ORDER BY created_at DESC
	`

	var sessions []*models.Session
	err := r.db.SelectContext(ctx, &sessions, query, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions by user ID: %w", err)
	}

	return sessions, nil
}

// UpdateExpiry обновляет время истечения сессии
func (r *SessionRepository) UpdateExpiry(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	query := `
		UPDATE sessions
		SET expires_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, expiresAt, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session expiry: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// AuthFailureRepository управляет записями о неудачных попытках входа
type AuthFailureRepository interface {
	// RecordFailure записывает неудачную попытку входа
	RecordFailure(ctx context.Context, email string, ipAddress string, reason models.AuthFailureReason, userAgent *string) error

	// CountRecentFailures подсчитывает количество неудачных попыток за последний период
	CountRecentFailures(ctx context.Context, email string, window time.Duration) (int, error)

	// CountRecentFailuresByIP подсчитывает количество неудачных попыток с IP за последний период
	CountRecentFailuresByIP(ctx context.Context, ipAddress string, window time.Duration) (int, error)

	// ClearFailures очищает все неудачные попытки для пользователя
	ClearFailures(ctx context.Context, email string) error

	// CleanupOldFailures удаляет старые записи о неудачных попытках
	CleanupOldFailures(ctx context.Context, olderThan time.Duration) error
}

// AuthFailureRepositoryImpl реализует AuthFailureRepository
type AuthFailureRepositoryImpl struct {
	db *sqlx.DB
}

// NewAuthFailureRepository создает новый AuthFailureRepository
func NewAuthFailureRepository(db *sqlx.DB) AuthFailureRepository {
	return &AuthFailureRepositoryImpl{db: db}
}

// RecordFailure записывает неудачную попытку входа
func (r *AuthFailureRepositoryImpl) RecordFailure(ctx context.Context, email string, ipAddress string, reason models.AuthFailureReason, userAgent *string) error {
	query := `
		INSERT INTO auth_failures (id, email, ip_address, reason, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	id := uuid.New()
	createdAt := time.Now()

	_, err := r.db.ExecContext(ctx, query, id, email, ipAddress, string(reason), userAgent, createdAt)
	if err != nil {
		return fmt.Errorf("failed to record auth failure: %w", err)
	}

	return nil
}

// CountRecentFailures подсчитывает количество неудачных попыток за последний период
func (r *AuthFailureRepositoryImpl) CountRecentFailures(ctx context.Context, email string, window time.Duration) (int, error) {
	query := `
		SELECT COUNT(*) as count
		FROM auth_failures
		WHERE email = $1 AND created_at > $2
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, email, time.Now().Add(-window))
	if err != nil {
		return 0, fmt.Errorf("failed to count recent failures: %w", err)
	}

	return count, nil
}

// CountRecentFailuresByIP подсчитывает количество неудачных попыток с IP за последний период
func (r *AuthFailureRepositoryImpl) CountRecentFailuresByIP(ctx context.Context, ipAddress string, window time.Duration) (int, error) {
	query := `
		SELECT COUNT(*) as count
		FROM auth_failures
		WHERE ip_address = $1 AND created_at > $2
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, ipAddress, time.Now().Add(-window))
	if err != nil {
		return 0, fmt.Errorf("failed to count recent failures by IP: %w", err)
	}

	return count, nil
}

// ClearFailures очищает все неудачные попытки для пользователя
func (r *AuthFailureRepositoryImpl) ClearFailures(ctx context.Context, email string) error {
	query := `DELETE FROM auth_failures WHERE email = $1`

	_, err := r.db.ExecContext(ctx, query, email)
	if err != nil {
		return fmt.Errorf("failed to clear failures: %w", err)
	}

	return nil
}

// CleanupOldFailures удаляет старые записи о неудачных попытках
func (r *AuthFailureRepositoryImpl) CleanupOldFailures(ctx context.Context, olderThan time.Duration) error {
	query := `DELETE FROM auth_failures WHERE created_at < $1`

	_, err := r.db.ExecContext(ctx, query, time.Now().Add(-olderThan))
	if err != nil {
		return fmt.Errorf("failed to cleanup old failures: %w", err)
	}

	return nil
}

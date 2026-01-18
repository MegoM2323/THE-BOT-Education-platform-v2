package repository

import (
	"context"
	"fmt"

	"tutoring-platform/internal/models"

	"github.com/jmoiron/sqlx"
)

// TrialRequestRepository управляет операциями с базой данных для заявок на пробное занятие
type TrialRequestRepository struct {
	db *sqlx.DB
}

// NewTrialRequestRepository создает новый TrialRequestRepository
func NewTrialRequestRepository(db *sqlx.DB) *TrialRequestRepository {
	return &TrialRequestRepository{db: db}
}

// Create создает новую заявку на пробное занятие
func (r *TrialRequestRepository) Create(ctx context.Context, input *models.CreateTrialRequestInput) (*models.TrialRequest, error) {
	query := `
		INSERT INTO trial_requests (name, phone, telegram, email)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, phone, telegram, email, created_at
	`

	var trialRequest models.TrialRequest
	err := r.db.QueryRowContext(ctx, query,
		input.Name,
		input.Phone,
		input.Telegram,
		input.Email,
	).Scan(
		&trialRequest.ID,
		&trialRequest.Name,
		&trialRequest.Phone,
		&trialRequest.Telegram,
		&trialRequest.Email,
		&trialRequest.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create trial request: %w", err)
	}

	return &trialRequest, nil
}

// GetAll получает все заявки на пробное занятие, упорядоченные по дате создания (новые первыми)
func (r *TrialRequestRepository) GetAll(ctx context.Context) ([]*models.TrialRequest, error) {
	query := `
		SELECT id, name, phone, telegram, email, created_at
		FROM trial_requests
		ORDER BY created_at DESC
	`

	var requests []*models.TrialRequest
	err := r.db.SelectContext(ctx, &requests, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get trial requests: %w", err)
	}

	return requests, nil
}

// GetByID получает заявку на пробное занятие по ID
func (r *TrialRequestRepository) GetByID(ctx context.Context, id int64) (*models.TrialRequest, error) {
	query := `
		SELECT id, name, phone, telegram, email, created_at
		FROM trial_requests
		WHERE id = $1
	`

	var trialRequest models.TrialRequest
	err := r.db.GetContext(ctx, &trialRequest, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get trial request by ID: %w", err)
	}

	return &trialRequest, nil
}

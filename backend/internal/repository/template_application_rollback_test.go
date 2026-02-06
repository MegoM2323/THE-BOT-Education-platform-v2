package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetApplicationByTemplateAndWeekTx_Success проверяет успешное нахождение application
func TestGetApplicationByTemplateAndWeekTx_Success(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	repo := NewTemplateApplicationRepository(db)

	// Создаем тестовые данные
	templateID := uuid.New()
	appliedByID := uuid.New()
	weekStartDate := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC) // Monday

	// Создаем пользователя (admin)
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, appliedByID, "test-rollback-admin@example.com", "hash", "Test", "Admin", "admin")
	require.NoError(t, err)

	// Создаем template
	_, err = db.ExecContext(ctx, `
		INSERT INTO lesson_templates (id, name, description, admin_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, templateID, "Test Rollback Template", "Template for testing rollback query", appliedByID)
	require.NoError(t, err)

	// Создаем application со статусом 'applied'
	applicationID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO template_applications
		(id, template_id, applied_by_id, week_start_date, applied_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), $5, NOW(), NOW())
	`, applicationID, templateID, appliedByID, weekStartDate, "applied")
	require.NoError(t, err)

	// Запускаем транзакцию для теста
	tx, err := db.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	// Тестируем поиск по строке даты
	weekStartDateStr := weekStartDate.Format("2006-01-02")
	app, err := repo.GetApplicationByTemplateAndWeekTx(ctx, tx, templateID, weekStartDateStr)

	// Проверки
	assert.NoError(t, err)
	require.NotNil(t, app)
	assert.Equal(t, applicationID, app.ID)
	assert.Equal(t, "applied", app.Status)
	assert.Equal(t, weekStartDate.Format("2006-01-02"), app.WeekStartDate.Format("2006-01-02"))
}

// TestGetApplicationByTemplateAndWeekTx_ReplacedStatus проверяет поиск application со статусом 'replaced'
func TestGetApplicationByTemplateAndWeekTx_ReplacedStatus(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	repo := NewTemplateApplicationRepository(db)

	templateID := uuid.New()
	appliedByID := uuid.New()
	weekStartDate := time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC)

	// Создаем пользователя
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, appliedByID, "test-replaced-admin@example.com", "hash", "Test Admin 2", "admin")
	require.NoError(t, err)

	// Создаем template
	_, err = db.ExecContext(ctx, `
		INSERT INTO lesson_templates (id, name, description, admin_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, templateID, "Test Replaced Template", "Template for testing replaced status", appliedByID)
	require.NoError(t, err)

	// Создаем первую application (старую, которая будет replaced)
	firstAppID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO template_applications
		(id, template_id, applied_by_id, week_start_date, applied_at, status, rolled_back_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW() - INTERVAL '1 hour', $5, NOW() - INTERVAL '1 hour', NOW(), NOW())
	`, firstAppID, templateID, appliedByID, weekStartDate, "replaced")
	require.NoError(t, err)

	// Создаем вторую application (новую, которая должна быть найдена)
	secondAppID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO template_applications
		(id, template_id, applied_by_id, week_start_date, applied_at, status, rolled_back_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), $5, NOW(), NOW(), NOW())
	`, secondAppID, templateID, appliedByID, weekStartDate, "replaced")
	require.NoError(t, err)

	tx, err := db.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	weekStartDateStr := weekStartDate.Format("2006-01-02")
	app, err := repo.GetApplicationByTemplateAndWeekTx(ctx, tx, templateID, weekStartDateStr)

	// Должна вернуться ПОСЛЕДНЯЯ application (по applied_at DESC)
	assert.NoError(t, err)
	require.NotNil(t, app)
	assert.Equal(t, secondAppID, app.ID, "Should return the most recent application")
	assert.Equal(t, "replaced", app.Status)

}

// TestGetApplicationByTemplateAndWeekTx_RolledBackNotFound проверяет что rolled_back application не находится
func TestGetApplicationByTemplateAndWeekTx_RolledBackNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	repo := NewTemplateApplicationRepository(db)

	templateID := uuid.New()
	appliedByID := uuid.New()
	weekStartDate := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)

	// Создаем пользователя
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, appliedByID, "test-rolledback-admin@example.com", "hash", "Test Admin 3", "admin")
	require.NoError(t, err)

	// Создаем template
	_, err = db.ExecContext(ctx, `
		INSERT INTO lesson_templates (id, name, description, admin_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, templateID, "Test Rolled Back Template", "Template for testing rolled_back status", appliedByID)
	require.NoError(t, err)

	// Создаем application со статусом 'rolled_back' (уже откаченная)
	rolledBackAppID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO template_applications
		(id, template_id, applied_by_id, week_start_date, applied_at, status, rolled_back_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), $5, NOW(), NOW(), NOW())
	`, rolledBackAppID, templateID, appliedByID, weekStartDate, "rolled_back")
	require.NoError(t, err)

	tx, err := db.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	weekStartDateStr := weekStartDate.Format("2006-01-02")
	app, err := repo.GetApplicationByTemplateAndWeekTx(ctx, tx, templateID, weekStartDateStr)

	// Должна быть ошибка "no active application found"
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "no active application found")
}

// TestGetApplicationByTemplateAndWeekTx_WrongTemplateID проверяет поиск с неверным template_id
func TestGetApplicationByTemplateAndWeekTx_WrongTemplateID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	repo := NewTemplateApplicationRepository(db)

	correctTemplateID := uuid.New()
	wrongTemplateID := uuid.New()
	appliedByID := uuid.New()
	weekStartDate := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)

	// Создаем пользователя
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, appliedByID, "test-wrong-template-admin@example.com", "hash", "Test Admin 4", "admin")
	require.NoError(t, err)

	// Создаем template
	_, err = db.ExecContext(ctx, `
		INSERT INTO lesson_templates (id, name, description, admin_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, correctTemplateID, "Correct Template", "Correct template", appliedByID)
	require.NoError(t, err)

	// Создаем application с correctTemplateID
	appID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO template_applications
		(id, template_id, applied_by_id, week_start_date, applied_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), $5, NOW(), NOW())
	`, appID, correctTemplateID, appliedByID, weekStartDate, "applied")
	require.NoError(t, err)

	tx, err := db.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	// Ищем с wrongTemplateID - должна быть ошибка
	weekStartDateStr := weekStartDate.Format("2006-01-02")
	app, err := repo.GetApplicationByTemplateAndWeekTx(ctx, tx, wrongTemplateID, weekStartDateStr)

	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "no active application found")
}

// TestGetApplicationByTemplateAndWeekTx_InvalidDateFormat проверяет валидацию формата даты
func TestGetApplicationByTemplateAndWeekTx_InvalidDateFormat(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	repo := NewTemplateApplicationRepository(db)

	tx, err := db.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	invalidDates := []string{
		"2026/03/09", // Wrong separator
		"09-03-2026", // Wrong order
		"2026-3-9",   // Missing leading zeros
		"not-a-date", // Invalid format
		"",           // Empty string
	}

	for _, invalidDate := range invalidDates {
		app, err := repo.GetApplicationByTemplateAndWeekTx(ctx, tx, uuid.New(), invalidDate)
		assert.Error(t, err, "Should fail for date: %s", invalidDate)
		assert.Nil(t, app)
		assert.Contains(t, err.Error(), "invalid week_start_date format")
	}
}

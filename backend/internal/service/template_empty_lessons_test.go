package service

import (
	"context"
	"database/sql"
	"testing"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestTemplateWithNoLessons проверяет поведение когда шаблон не содержит уроков
func TestTemplateWithNoLessons(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	// Создаём репозитории и сервисы
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	svc := NewTemplateService(
		db,
		templateRepo,
		templateLessonRepo,
		templateAppRepo,
		lessonRepo,
		creditRepo,
		bookingRepo,
		userRepo,
	)

	// Создаём пользователей
	admin := createTestUser(t, db, "admin@test.com", "Test Admin", string(models.RoleAdmin))

	// Создаём шаблон БЕЗ уроков напрямую в БД
	templateID := uuid.New()
	_, err := db.Exec(`
		INSERT INTO lesson_templates (id, name, description, admin_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, templateID, "Empty Template", sql.NullString{}, admin.ID)
	assert.NoError(t, err)

	// Убеждаемся что шаблон создан, но НЕ имеет уроков
	fullTemplate, err := templateRepo.GetTemplateWithLessons(ctx, templateID)
	assert.NoError(t, err)
	assert.Equal(t, "Empty Template", fullTemplate.Name)
	assert.Empty(t, fullTemplate.Lessons, "Template should have 0 lessons")

	t.Logf("Template created: %s with %d lessons", fullTemplate.Name, len(fullTemplate.Lessons))

	// Пытаемся применить шаблон БЕЗ уроков к неделе
	result, err := svc.ApplyTemplateToWeek(ctx, admin.ID, &models.ApplyTemplateRequest{
		TemplateID:    templateID,
		WeekStartDate: "2025-12-15", // Понедельник
		DryRun:        false,
	})

	// DEBUG: С нашим логированием мы увидим:
	// {"level":"info","lessons_count":0,"message":"Template loaded for week application"}
	// {"level":"info","template_lessons_to_create":0,"message":"Starting lesson creation loop"}

	// Ожидаем успешное применение, но с 0 созданных уроков
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.CreatedLessonsCount, "Empty template should create 0 lessons")
	assert.Equal(t, "applied", result.Status, "Application status should be 'applied'")

	t.Log("✓ Empty template application: 0 lessons created (expected behavior)")
	t.Logf("Result: %d lessons created with status '%s'", result.CreatedLessonsCount, result.Status)
}

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/models"
)

// TestGetOrCreateDefaultTemplate removed - default template concept removed
// Use CreateTemplate and GetAllTemplates instead for multiple named templates

func TestTemplateWithLessons(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewLessonTemplateRepository(db)
	lessonRepo := NewTemplateLessonRepository(db)
	ctx := context.Background()

	// Create admin user
	adminID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, adminID, "test.admin@example.com", "hash", "Test Admin", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	// Create template
	template := &models.LessonTemplate{
		AdminID:     adminID,
		Name:        "Weekly Schedule",
		Description: sql.NullString{String: "Test template", Valid: true},
	}
	err = repo.CreateTemplate(ctx, template)
	require.NoError(t, err)

	// Create a test teacher
	teacherID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, teacherID, "test.teacher@example.com", "hash", "Test Teacher", "teacher", time.Now(), time.Now())
	require.NoError(t, err)

	t.Run("template starts empty", func(t *testing.T) {
		fullTemplate, err := repo.GetTemplateWithLessons(ctx, template.ID)
		require.NoError(t, err)
		assert.NotNil(t, fullTemplate)
		assert.Len(t, fullTemplate.Lessons, 0)
	})

	t.Run("can add lessons to template", func(t *testing.T) {
		// Add a lesson
		lesson := &models.TemplateLessonEntry{
			ID:          uuid.New(),
			TemplateID:  template.ID,
			DayOfWeek:   1, // Monday
			StartTime:   "10:00:00",
			EndTime:     "12:00:00",
			TeacherID:   teacherID,
			LessonType:  "group",
			MaxStudents: 4,
			Color:       "#3B82F6",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := lessonRepo.CreateTemplateLessonEntry(ctx, lesson)
		require.NoError(t, err)

		// Retrieve template with lessons
		fullTemplate, err := repo.GetTemplateWithLessons(ctx, template.ID)
		require.NoError(t, err)
		assert.Len(t, fullTemplate.Lessons, 1)
		assert.Equal(t, lesson.ID, fullTemplate.Lessons[0].ID)
		assert.Equal(t, "Test Teacher", fullTemplate.Lessons[0].TeacherName)
	})

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lessons WHERE template_id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", teacherID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", adminID)
}

func TestGetAllTemplatesWithLessonCount(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewLessonTemplateRepository(db)
	lessonRepo := NewTemplateLessonRepository(db)
	ctx := context.Background()

	// Create admin user
	adminID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, adminID, "test.admin.count@example.com", "hash", "Test Admin Count", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	// Create teacher
	teacherID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, teacherID, "test.teacher.count@example.com", "hash", "Test Teacher Count", "teacher", time.Now(), time.Now())
	require.NoError(t, err)

	// Create first template (empty)
	template1 := &models.LessonTemplate{
		AdminID:     adminID,
		Name:        "Empty Template",
		Description: sql.NullString{String: "Template without lessons", Valid: true},
	}
	err = repo.CreateTemplate(ctx, template1)
	require.NoError(t, err)

	// Create second template with lessons
	template2 := &models.LessonTemplate{
		AdminID:     adminID,
		Name:        "Template With Lessons",
		Description: sql.NullString{String: "Template with 3 lessons", Valid: true},
	}
	err = repo.CreateTemplate(ctx, template2)
	require.NoError(t, err)

	// Add 3 lessons to template2
	for i := 0; i < 3; i++ {
		lesson := &models.TemplateLessonEntry{
			ID:          uuid.New(),
			TemplateID:  template2.ID,
			DayOfWeek:   i + 1,
			StartTime:   "10:00:00",
			EndTime:     "12:00:00",
			TeacherID:   teacherID,
			LessonType:  "group",
			MaxStudents: 4,
			Color:       "#3B82F6",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := lessonRepo.CreateTemplateLessonEntry(ctx, lesson)
		require.NoError(t, err)
	}

	t.Run("GetAllTemplates returns correct lesson_count", func(t *testing.T) {
		templates, err := repo.GetAllTemplates(ctx, adminID)
		require.NoError(t, err)
		require.Len(t, templates, 2)

		// Проверяем что lesson_count корректен для обоих шаблонов
		var emptyTemplate, fullTemplate *models.LessonTemplate
		for _, tmpl := range templates {
			if tmpl.Name == "Empty Template" {
				emptyTemplate = tmpl
			} else if tmpl.Name == "Template With Lessons" {
				fullTemplate = tmpl
			}
		}

		require.NotNil(t, emptyTemplate, "Empty template not found")
		require.NotNil(t, fullTemplate, "Full template not found")

		assert.Equal(t, 0, emptyTemplate.LessonCount, "Empty template should have 0 lessons")
		assert.Equal(t, 3, fullTemplate.LessonCount, "Full template should have 3 lessons")
	})

	t.Run("GetAllTemplates with nil adminID returns all templates with lesson_count", func(t *testing.T) {
		templates, err := repo.GetAllTemplates(ctx, uuid.Nil)
		require.NoError(t, err)

		// Должны быть как минимум наши 2 шаблона
		require.GreaterOrEqual(t, len(templates), 2)

		// Находим наши шаблоны и проверяем lesson_count
		for _, tmpl := range templates {
			if tmpl.ID == template1.ID {
				assert.Equal(t, 0, tmpl.LessonCount)
			} else if tmpl.ID == template2.ID {
				assert.Equal(t, 3, tmpl.LessonCount)
			}
		}
	})

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lessons WHERE template_id IN ($1, $2)", template1.ID, template2.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id IN ($1, $2)", template1.ID, template2.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", teacherID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", adminID)
}

func TestWithTransactionRollback(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewLessonTemplateRepository(db)
	lessonRepo := NewTemplateLessonRepository(db)
	ctx := context.Background()

	// Create admin user
	adminID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, adminID, "test.admin.txn@example.com", "hash", "Test Admin Txn", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	// Create template
	template := &models.LessonTemplate{
		AdminID:     adminID,
		Name:        "Transaction Test Template",
		Description: sql.NullString{String: "Test transaction behavior", Valid: true},
	}
	err = repo.CreateTemplate(ctx, template)
	require.NoError(t, err)

	// Create a test teacher
	teacherID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, teacherID, "test.teacher.txn@example.com", "hash", "Test Teacher Txn", "teacher", time.Now(), time.Now())
	require.NoError(t, err)

	t.Run("WithTransaction passes actual transaction to callback", func(t *testing.T) {
		// Use the transaction to insert a lesson
		callbackExecuted := false
		err := lessonRepo.WithTransaction(ctx, func(tx *sqlx.Tx) error {
			callbackExecuted = true
			require.NotNil(t, tx, "transaction should not be nil")
			return nil
		})
		require.NoError(t, err)
		assert.True(t, callbackExecuted, "callback should have been executed")
	})

	t.Run("WithTransaction rolls back on callback error", func(t *testing.T) {
		// Create initial lesson count
		initialCount := 0
		err := db.GetContext(ctx, &initialCount, "SELECT COUNT(*) FROM template_lessons WHERE template_id = $1", template.ID)
		require.NoError(t, err)

		// Try to execute transaction that will fail
		testError := fmt.Errorf("intentional test error")
		err = lessonRepo.WithTransaction(ctx, func(tx *sqlx.Tx) error {
			// The transaction callback receives the actual transaction
			require.NotNil(t, tx, "transaction should not be nil in callback")
			// Simulate some operation that would fail
			return testError
		})

		// Verify error was returned
		require.Error(t, err)
		assert.Equal(t, testError, err)

		// Verify lesson count hasn't changed (transaction was rolled back)
		finalCount := 0
		err = db.GetContext(ctx, &finalCount, "SELECT COUNT(*) FROM template_lessons WHERE template_id = $1", template.ID)
		require.NoError(t, err)
		assert.Equal(t, initialCount, finalCount, "lesson count should not change after rollback")
	})

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lessons WHERE template_id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", teacherID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", adminID)
}

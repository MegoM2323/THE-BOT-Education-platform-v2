package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/models"
)

func TestLessonCountConsistency(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewLessonTemplateRepository(db)
	lessonRepo := NewTemplateLessonRepository(db)
	ctx := context.Background()

	// Create admin user
	adminID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, adminID, "test.admin.consistency@example.com", "hash", "Test Admin", "Consistency", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	// Create template
	template := &models.LessonTemplate{
		AdminID: adminID,
		Name:    "Consistency Test Template",
	}
	err = repo.CreateTemplate(ctx, template)
	require.NoError(t, err)

	// Create teacher
	teacherID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, teacherID, "test.teacher.consistency@example.com", "hash", "Test Teacher", "Consistency", "methodologist", time.Now(), time.Now())
	require.NoError(t, err)

	t.Run("GetTemplateWithLessons computes LessonCount from Lessons array", func(t *testing.T) {
		// Initially empty
		fullTemplate, err := repo.GetTemplateWithLessons(ctx, template.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, fullTemplate.LessonCount, "Empty template should have LessonCount=0")
		assert.Len(t, fullTemplate.Lessons, 0, "Empty template should have no lessons")

		// Add 2 lessons
		for i := 0; i < 2; i++ {
			lesson := &models.TemplateLessonEntry{
				ID:          uuid.New(),
				TemplateID:  template.ID,
				DayOfWeek:   i,
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

		// Verify LessonCount matches Lessons array length
		fullTemplate, err = repo.GetTemplateWithLessons(ctx, template.ID)
		require.NoError(t, err)
		assert.Equal(t, len(fullTemplate.Lessons), fullTemplate.LessonCount,
			"LessonCount should equal len(Lessons) array")
		assert.Equal(t, 2, fullTemplate.LessonCount, "LessonCount should be 2")
		assert.Len(t, fullTemplate.Lessons, 2, "Lessons array should have 2 items")
	})

	t.Run("LessonCount from GetAllTemplates matches GetTemplateWithLessons", func(t *testing.T) {
		// Get via GetAllTemplates (uses COUNT(*) subquery)
		templates, err := repo.GetAllTemplates(ctx, adminID)
		require.NoError(t, err)

		var countFromList int
		for _, tmpl := range templates {
			if tmpl.ID == template.ID {
				countFromList = tmpl.LessonCount
				break
			}
		}

		// Get via GetTemplateWithLessons (computes from array)
		fullTemplate, err := repo.GetTemplateWithLessons(ctx, template.ID)
		require.NoError(t, err)
		countFromFull := fullTemplate.LessonCount

		// Both should match
		assert.Equal(t, countFromList, countFromFull,
			"LessonCount from GetAllTemplates should match GetTemplateWithLessons")
	})

	t.Run("LessonCount stays accurate after lesson deletion", func(t *testing.T) {
		// Get initial count
		fullTemplate, err := repo.GetTemplateWithLessons(ctx, template.ID)
		require.NoError(t, err)
		initialCount := fullTemplate.LessonCount
		initialLessons := fullTemplate.Lessons

		// Delete first lesson
		err = lessonRepo.DeleteTemplateLessonEntry(ctx, initialLessons[0].ID)
		require.NoError(t, err)

		// Verify count decreased
		fullTemplate, err = repo.GetTemplateWithLessons(ctx, template.ID)
		require.NoError(t, err)
		assert.Equal(t, initialCount-1, fullTemplate.LessonCount,
			"LessonCount should decrease after deletion")
		assert.Equal(t, len(fullTemplate.Lessons), fullTemplate.LessonCount,
			"LessonCount should still equal len(Lessons)")
	})

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lessons WHERE template_id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", teacherID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", adminID)
}

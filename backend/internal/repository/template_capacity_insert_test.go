package repository

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/models"
)

// TestTemplateLessonCapacityOnInsert validates that capacity check works on INSERT
// This test verifies fix for FB016: Capacity validation only on UPDATE trigger
func TestTemplateLessonCapacityOnInsert_DirectSQL(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	// Setup: Create admin, teacher, and students
	adminID := uuid.New()
	teacherID := uuid.New()
	student1ID := uuid.New()
	student2ID := uuid.New()
	student3ID := uuid.New()

	// Create users
	createTestUserSQLX(t, db, adminID, "admin.capacity@example.com", "admin")
	createTestUserSQLX(t, db, teacherID, "teacher.capacity@example.com", "teacher")
	createTestUserSQLX(t, db, student1ID, "student1.capacity@example.com", "student")
	createTestUserSQLX(t, db, student2ID, "student2.capacity@example.com", "student")
	createTestUserSQLX(t, db, student3ID, "student3.capacity@example.com", "student")

	// Create template
	templateID := uuid.New()
	err := db.QueryRowxContext(ctx, `
		INSERT INTO lesson_templates (id, admin_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, templateID, adminID, "Capacity Test Template", time.Now(), time.Now()).Scan(&templateID)
	require.NoError(t, err)

	// Create template lesson with max_students = 2
	lessonID := uuid.New()
	err = db.QueryRowxContext(ctx, `
		INSERT INTO template_lessons
		(id, template_id, day_of_week, start_time, end_time, teacher_id, lesson_type, max_students, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, lessonID, templateID, 1, "10:00:00", "12:00:00", teacherID, "group", 2, "#3B82F6", time.Now(), time.Now()).Scan(&lessonID)
	require.NoError(t, err)

	// Test 1: Add first student (should succeed)
	var studentID uuid.UUID
	err = db.QueryRowxContext(ctx, `
		INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, uuid.New(), lessonID, student1ID, time.Now()).Scan(&studentID)
	require.NoError(t, err, "Adding first student should succeed")

	// Verify count
	var count int
	err = db.QueryRowxContext(ctx, `
		SELECT COUNT(*) FROM template_lesson_students WHERE template_lesson_id = $1
	`, lessonID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Should have 1 student")

	// Test 2: Add second student (should succeed - at capacity)
	err = db.QueryRowxContext(ctx, `
		INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, uuid.New(), lessonID, student2ID, time.Now()).Scan(&studentID)
	require.NoError(t, err, "Adding second student should succeed (at capacity)")

	// Verify count
	err = db.QueryRowxContext(ctx, `
		SELECT COUNT(*) FROM template_lesson_students WHERE template_lesson_id = $1
	`, lessonID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count, "Should have 2 students")

	// Test 3: Try to add third student (should FAIL - exceeds capacity)
	err = db.QueryRowxContext(ctx, `
		INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, uuid.New(), lessonID, student3ID, time.Now()).Scan(&studentID)

	// This should fail with capacity error
	assert.Error(t, err, "Adding third student should fail - exceeds capacity of 2")
	if err != nil {
		assert.Contains(t, err.Error(), "capacity", "Error should mention capacity constraint")
	}

	// Verify final count still 2 (insert was rejected)
	err = db.QueryRowxContext(ctx, `
		SELECT COUNT(*) FROM template_lesson_students WHERE template_lesson_id = $1
	`, lessonID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count, "Should still have 2 students after failed insert")

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lesson_students WHERE template_lesson_id = $1", lessonID)
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lessons WHERE id = $1", lessonID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", templateID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2, $3, $4, $5)", adminID, teacherID, student1ID, student2ID, student3ID)
}

// TestTemplateLessonCapacityOnInsert_RepositoryLayer validates capacity check through repository
func TestTemplateLessonCapacityOnInsert_RepositoryLayer(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	lessonRepo := NewTemplateLessonRepository(db)
	ctx := context.Background()

	// Setup
	adminID := uuid.New()
	teacherID := uuid.New()
	student1ID := uuid.New()
	student2ID := uuid.New()
	student3ID := uuid.New()

	createTestUserSQLX(t, db, adminID, "admin.repocap@example.com", "admin")
	createTestUserSQLX(t, db, teacherID, "teacher.repocap@example.com", "teacher")
	createTestUserSQLX(t, db, student1ID, "student1.repocap@example.com", "student")
	createTestUserSQLX(t, db, student2ID, "student2.repocap@example.com", "student")
	createTestUserSQLX(t, db, student3ID, "student3.repocap@example.com", "student")

	// Create template
	templateRepo := NewLessonTemplateRepository(db)
	template := &models.LessonTemplate{
		AdminID:     adminID,
		Name:        "Repository Capacity Test",
		Description: sql.NullString{String: "Test repository layer capacity", Valid: true},
	}
	err := templateRepo.CreateTemplate(ctx, template)
	require.NoError(t, err)

	// Create template lesson with capacity = 2
	lesson := &models.TemplateLessonEntry{
		ID:          uuid.New(),
		TemplateID:  template.ID,
		DayOfWeek:   3,
		StartTime:   "14:00:00",
		EndTime:     "16:00:00",
		TeacherID:   teacherID,
		LessonType:  "group",
		MaxStudents: 2,
		Color:       "#F59E0B",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = lessonRepo.CreateTemplateLessonEntry(ctx, lesson)
	require.NoError(t, err)

	// Test adding students through repository
	// Student 1: should succeed
	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, student1ID)
	require.NoError(t, err, "First student should be added")

	// Student 2: should succeed (at capacity)
	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, student2ID)
	require.NoError(t, err, "Second student should be added at capacity")

	// Student 3: should fail (exceeds capacity)
	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, student3ID)
	require.Error(t, err, "Third student should be rejected - exceeds capacity")

	// Verify final count
	students, err := lessonRepo.GetStudentsForTemplateLessonEntry(ctx, lesson.ID)
	require.NoError(t, err)
	assert.Len(t, students, 2, "Should have exactly 2 students after failed add attempt")

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lesson_students WHERE template_lesson_id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lessons WHERE id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2, $3, $4, $5)", adminID, teacherID, student1ID, student2ID, student3ID)
}

// TestTemplateLessonCapacityOnUpdate validates capacity check still works on UPDATE
func TestTemplateLessonCapacityOnUpdate_NoReduction(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	lessonRepo := NewTemplateLessonRepository(db)
	ctx := context.Background()

	// Setup
	adminID := uuid.New()
	teacherID := uuid.New()
	student1ID := uuid.New()
	student2ID := uuid.New()

	createTestUserSQLX(t, db, adminID, "admin.updatecap@example.com", "admin")
	createTestUserSQLX(t, db, teacherID, "teacher.updatecap@example.com", "teacher")
	createTestUserSQLX(t, db, student1ID, "student1.updatecap@example.com", "student")
	createTestUserSQLX(t, db, student2ID, "student2.updatecap@example.com", "student")

	// Create template and lesson with capacity = 3
	templateRepo := NewLessonTemplateRepository(db)
	template := &models.LessonTemplate{
		AdminID:     adminID,
		Name:        "Update Capacity Test",
		Description: sql.NullString{String: "Test update capacity check", Valid: true},
	}
	err := templateRepo.CreateTemplate(ctx, template)
	require.NoError(t, err)

	lesson := &models.TemplateLessonEntry{
		ID:          uuid.New(),
		TemplateID:  template.ID,
		DayOfWeek:   4,
		StartTime:   "09:00:00",
		EndTime:     "11:00:00",
		TeacherID:   teacherID,
		LessonType:  "group",
		MaxStudents: 3,
		Color:       "#8B5CF6",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = lessonRepo.CreateTemplateLessonEntry(ctx, lesson)
	require.NoError(t, err)

	// Add 2 students
	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, student1ID)
	require.NoError(t, err)
	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, student2ID)
	require.NoError(t, err)

	// Verify count
	students, err := lessonRepo.GetStudentsForTemplateLessonEntry(ctx, lesson.ID)
	require.NoError(t, err)
	assert.Len(t, students, 2, "Should have 2 students")

	// Try to reduce max_students to 1 (should fail - 2 students already assigned)
	var lessonUpdateID uuid.UUID
	err = db.QueryRowContext(ctx, `
		UPDATE template_lessons SET max_students = $1 WHERE id = $2
		RETURNING id
	`, 1, lesson.ID).Scan(&lessonUpdateID)

	assert.Error(t, err, "Reducing max_students below current student count should fail")
	if err != nil {
		// Check for either "capacity" or the actual error message pattern
		errMsg := err.Error()
		assert.True(t,
			(contains(errMsg, "capacity") || contains(errMsg, "Cannot set max_students")),
			"Error should mention capacity constraint or max_students limitation. Got: %s", errMsg)
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lesson_students WHERE template_lesson_id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lessons WHERE id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2, $3, $4)", adminID, teacherID, student1ID, student2ID)
}

// Helper function to create test users
func createTestUserSQLX(t *testing.T, db *sqlx.DB, id uuid.UUID, email, role string) {
	_, err := db.Exec(`
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, email, "hash", "Test User", role, time.Now(), time.Now())
	require.NoError(t, err, "Failed to create test user: %s", email)
}

// Helper function for substring matching
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

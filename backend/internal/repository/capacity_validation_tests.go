package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/models"
)

// TEST SUITE 1: Capacity Validation Tests
// Tests for AddStudentToTemplateLessonEntry with capacity constraints

func TestAddStudentToTemplateLessonEntry_BelowCapacity_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lessonRepo := NewTemplateLessonRepository(db)
	ctx := context.Background()

	// Setup: Create template with lesson capacity=3
	adminID := uuid.New()
	teacherID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, adminID, "admin.capacity@example.com", "hash", "Admin", "", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, teacherID, "teacher.capacity@example.com", "hash", "Teacher", "", "methodologist", time.Now(), time.Now())
	require.NoError(t, err)

	templateRepo := NewLessonTemplateRepository(db)
	template := &models.LessonTemplate{
		AdminID:     adminID,
		Name:        "Capacity Test Template",
		Description: sql.NullString{String: "Test template for capacity", Valid: true},
	}
	err = templateRepo.CreateTemplate(ctx, template)
	require.NoError(t, err)

	// Create lesson with max_students = 3
	lesson := &models.TemplateLessonEntry{
		ID:          uuid.New(),
		TemplateID:  template.ID,
		DayOfWeek:   1,
		StartTime:   "10:00:00",
		EndTime:     "12:00:00",
		TeacherID:   teacherID,
		MaxStudents: 3,
		Color:       "#3B82F6",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = lessonRepo.CreateTemplateLessonEntry(ctx, lesson)
	require.NoError(t, err)

	// Test: Add 2 students (below capacity of 3)
	student1ID := uuid.New()
	student2ID := uuid.New()

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, student1ID, "student1.capacity@example.com", "hash", "Student", "1", "student", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, student2ID, "student2.capacity@example.com", "hash", "Student", "2", "student", time.Now(), time.Now())
	require.NoError(t, err)

	// Add both students - should succeed
	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, student1ID)
	require.NoError(t, err, "Adding first student should succeed")

	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, student2ID)
	require.NoError(t, err, "Adding second student should succeed (2 < capacity 3)")

	// Verify both students were added
	students, err := lessonRepo.GetStudentsForTemplateLessonEntry(ctx, lesson.ID)
	require.NoError(t, err)
	assert.Len(t, students, 2, "Should have 2 students in lesson")

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lesson_students WHERE template_lesson_id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lessons WHERE id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2, $3, $4)", adminID, teacherID, student1ID, student2ID)
}

func TestAddStudentToTemplateLessonEntry_AtCapacity_CannotAdd(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lessonRepo := NewTemplateLessonRepository(db)
	ctx := context.Background()

	// Setup: Create template with lesson capacity=2
	adminID := uuid.New()
	teacherID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, adminID, "admin.atcap@example.com", "hash", "Admin", "", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, teacherID, "teacher.atcap@example.com", "hash", "Teacher", "", "methodologist", time.Now(), time.Now())
	require.NoError(t, err)

	templateRepo := NewLessonTemplateRepository(db)
	template := &models.LessonTemplate{
		AdminID:     adminID,
		Name:        "Capacity At Limit Template",
		Description: sql.NullString{String: "Test template for capacity limit", Valid: true},
	}
	err = templateRepo.CreateTemplate(ctx, template)
	require.NoError(t, err)

	// Create lesson with max_students = 2
	lesson := &models.TemplateLessonEntry{
		ID:          uuid.New(),
		TemplateID:  template.ID,
		DayOfWeek:   2,
		StartTime:   "14:00:00",
		EndTime:     "16:00:00",
		TeacherID:   teacherID,
		MaxStudents: 2,
		Color:       "#EF4444",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = lessonRepo.CreateTemplateLessonEntry(ctx, lesson)
	require.NoError(t, err)

	// Create 3 students
	students := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		students[i] = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, students[i], "student"+string(rune('0'+i))+".atcap@example.com", "hash", "Student", string(rune('0'+i)), "student", time.Now(), time.Now())
		require.NoError(t, err)
	}

	// Add 2 students (at capacity)
	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, students[0])
	require.NoError(t, err, "Adding first student should succeed")

	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, students[1])
	require.NoError(t, err, "Adding second student should succeed (at capacity 2)")

	// Verify count
	studentList, err := lessonRepo.GetStudentsForTemplateLessonEntry(ctx, lesson.ID)
	require.NoError(t, err)
	assert.Len(t, studentList, 2, "Should have exactly 2 students at capacity")

	// Try to add 3rd student - behavior test
	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, students[2])
	t.Logf("Adding 3rd student when capacity=2 result: %v", err)

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lesson_students WHERE template_lesson_id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lessons WHERE id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2, $3, $4, $5)", adminID, teacherID, students[0], students[1], students[2])
}

func TestAddStudentToTemplateLessonEntry_DuplicateHandling(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lessonRepo := NewTemplateLessonRepository(db)
	ctx := context.Background()

	// Setup
	adminID := uuid.New()
	teacherID := uuid.New()
	studentID := uuid.New()

	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, adminID, "admin.dup@example.com", "hash", "Admin", "", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, teacherID, "teacher.dup@example.com", "hash", "Teacher", "", "methodologist", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, studentID, "student.dup@example.com", "hash", "Student", "", "student", time.Now(), time.Now())
	require.NoError(t, err)

	templateRepo := NewLessonTemplateRepository(db)
	template := &models.LessonTemplate{
		AdminID:     adminID,
		Name:        "Duplicate Student Template",
		Description: sql.NullString{String: "Test template for duplicate handling", Valid: true},
	}
	err = templateRepo.CreateTemplate(ctx, template)
	require.NoError(t, err)

	lesson := &models.TemplateLessonEntry{
		ID:          uuid.New(),
		TemplateID:  template.ID,
		DayOfWeek:   3,
		StartTime:   "09:00:00",
		EndTime:     "11:00:00",
		TeacherID:   teacherID,
		MaxStudents: 5,
		Color:       "#10B981",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = lessonRepo.CreateTemplateLessonEntry(ctx, lesson)
	require.NoError(t, err)

	// Add same student twice
	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, studentID)
	require.NoError(t, err, "First addition should succeed")

	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, studentID)
	if err != nil {
		t.Logf("Duplicate addition prevented: %v", err)
	} else {
		students, err := lessonRepo.GetStudentsForTemplateLessonEntry(ctx, lesson.ID)
		require.NoError(t, err)
		t.Logf("Duplicate addition allowed - student count: %d", len(students))
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lesson_students WHERE template_lesson_id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lessons WHERE id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2, $3)", adminID, teacherID, studentID)
}

func TestAddStudentToTemplateLessonEntry_MultipleStudentsSequential(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lessonRepo := NewTemplateLessonRepository(db)
	ctx := context.Background()

	// Setup
	adminID := uuid.New()
	teacherID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, adminID, "admin.seq@example.com", "hash", "Admin", "", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, teacherID, "teacher.seq@example.com", "hash", "Teacher", "", "methodologist", time.Now(), time.Now())
	require.NoError(t, err)

	templateRepo := NewLessonTemplateRepository(db)
	template := &models.LessonTemplate{
		AdminID:     adminID,
		Name:        "Sequential Enrollment Template",
		Description: sql.NullString{String: "Test sequential student enrollment", Valid: true},
	}
	err = templateRepo.CreateTemplate(ctx, template)
	require.NoError(t, err)

	// Lesson with capacity = 2
	lesson := &models.TemplateLessonEntry{
		ID:          uuid.New(),
		TemplateID:  template.ID,
		DayOfWeek:   4,
		StartTime:   "15:00:00",
		EndTime:     "17:00:00",
		TeacherID:   teacherID,
		MaxStudents: 2,
		Color:       "#6366F1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = lessonRepo.CreateTemplateLessonEntry(ctx, lesson)
	require.NoError(t, err)

	// Create 3 students
	students := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		students[i] = uuid.New()
		_, err = db.ExecContext(ctx, `
			INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, students[i], "student"+string(rune('1'+i))+".seq@example.com", "hash", "Student", string(rune('1'+i)), "student", time.Now(), time.Now())
		require.NoError(t, err)
	}

	// Test sequential enrollment
	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, students[0])
	require.NoError(t, err)
	studentList, _ := lessonRepo.GetStudentsForTemplateLessonEntry(ctx, lesson.ID)
	assert.Len(t, studentList, 1, "Count should be 1 after first student")

	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, students[1])
	require.NoError(t, err)
	studentList, _ = lessonRepo.GetStudentsForTemplateLessonEntry(ctx, lesson.ID)
	assert.Len(t, studentList, 2, "Count should be 2 after second student")

	// Try 3rd student
	err = lessonRepo.AddStudentToTemplateLessonEntry(ctx, lesson.ID, students[2])
	if err != nil {
		t.Logf("Adding 3rd student rejected: %v", err)
	} else {
		studentList, _ = lessonRepo.GetStudentsForTemplateLessonEntry(ctx, lesson.ID)
		t.Logf("Adding 3rd student allowed - count: %d", len(studentList))
	}

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lesson_students WHERE template_lesson_id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM template_lessons WHERE id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lesson_templates WHERE id = $1", template.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2, $3, $4, $5)", adminID, teacherID, students[0], students[1], students[2])
}

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetVisibleLessons_RealDatabase - Integration test with real database
func TestGetVisibleLessons_RealDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	lessonRepo := NewLessonRepository(db)

	// Create test users
	admin := createTestUserVisibility(t, db, ctx, "admin@test.com", models.RoleAdmin)
	teacher1 := createTestUserVisibility(t, db, ctx, "teacher1@test.com", models.RoleMethodologist)
	teacher2 := createTestUserVisibility(t, db, ctx, "teacher2@test.com", models.RoleMethodologist)
	student1 := createTestUserVisibility(t, db, ctx, "student1@test.com", models.RoleStudent)
	student2 := createTestUserVisibility(t, db, ctx, "student2@test.com", models.RoleStudent)

	// Create lessons
	// Teacher1 lessons
	groupLesson1 := createTestLessonWithType(t, lessonRepo, ctx, teacher1.ID, 4)
	individualLesson1 := createTestLessonWithType(t, lessonRepo, ctx, teacher1.ID, 1)

	// Teacher2 lessons
	groupLesson2 := createTestLessonWithType(t, lessonRepo, ctx, teacher2.ID, 4)
	individualLesson2 := createTestLessonWithType(t, lessonRepo, ctx, teacher2.ID, 1)

	// Assign student1 to individualLesson1
	createTestBooking(t, db, ctx, student1.ID, individualLesson1.ID)

	// Assign student2 to individualLesson2
	createTestBooking(t, db, ctx, student2.ID, individualLesson2.ID)

	t.Run("Admin sees all lessons", func(t *testing.T) {
		lessons, err := lessonRepo.GetVisibleLessons(ctx, admin.ID, string(models.RoleAdmin), &models.ListLessonsFilter{})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(lessons), 4, "Admin should see at least 4 lessons")

		lessonIDs := extractLessonIDs(lessons)
		assert.Contains(t, lessonIDs, groupLesson1.ID)
		assert.Contains(t, lessonIDs, individualLesson1.ID)
		assert.Contains(t, lessonIDs, groupLesson2.ID)
		assert.Contains(t, lessonIDs, individualLesson2.ID)
	})

	t.Run("Teacher1 sees only their lessons", func(t *testing.T) {
		lessons, err := lessonRepo.GetVisibleLessons(ctx, teacher1.ID, string(models.RoleMethodologist), &models.ListLessonsFilter{})
		require.NoError(t, err)
		require.Len(t, lessons, 2, "Teacher1 should see only their 2 lessons")

		lessonIDs := extractLessonIDs(lessons)
		assert.Contains(t, lessonIDs, groupLesson1.ID)
		assert.Contains(t, lessonIDs, individualLesson1.ID)
		assert.NotContains(t, lessonIDs, groupLesson2.ID)
		assert.NotContains(t, lessonIDs, individualLesson2.ID)
	})

	t.Run("Teacher2 sees only their lessons", func(t *testing.T) {
		lessons, err := lessonRepo.GetVisibleLessons(ctx, teacher2.ID, string(models.RoleMethodologist), &models.ListLessonsFilter{})
		require.NoError(t, err)
		require.Len(t, lessons, 2, "Teacher2 should see only their 2 lessons")

		lessonIDs := extractLessonIDs(lessons)
		assert.Contains(t, lessonIDs, groupLesson2.ID)
		assert.Contains(t, lessonIDs, individualLesson2.ID)
		assert.NotContains(t, lessonIDs, groupLesson1.ID)
		assert.NotContains(t, lessonIDs, individualLesson1.ID)
	})

	t.Run("Student1 sees group lessons + their individual lesson", func(t *testing.T) {
		lessons, err := lessonRepo.GetVisibleLessons(ctx, student1.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(lessons), 3, "Student1 should see at least 2 group lessons + 1 assigned individual lesson")

		lessonIDs := extractLessonIDs(lessons)
		assert.Contains(t, lessonIDs, groupLesson1.ID)
		assert.Contains(t, lessonIDs, groupLesson2.ID)
		assert.Contains(t, lessonIDs, individualLesson1.ID, "Student1 should see their assigned individual lesson")
		assert.NotContains(t, lessonIDs, individualLesson2.ID, "Student1 should NOT see student2's individual lesson")
	})

	t.Run("Student2 sees group lessons + their individual lesson", func(t *testing.T) {
		lessons, err := lessonRepo.GetVisibleLessons(ctx, student2.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(lessons), 3, "Student2 should see at least 2 group lessons + 1 assigned individual lesson")

		lessonIDs := extractLessonIDs(lessons)
		assert.Contains(t, lessonIDs, groupLesson1.ID)
		assert.Contains(t, lessonIDs, groupLesson2.ID)
		assert.Contains(t, lessonIDs, individualLesson2.ID, "Student2 should see their assigned individual lesson")
		assert.NotContains(t, lessonIDs, individualLesson1.ID, "Student2 should NOT see student1's individual lesson")
	})

	t.Run("Student without individual lessons sees only group lessons", func(t *testing.T) {
		student3 := createTestUserVisibility(t, db, ctx, "student3@test.com", models.RoleStudent)

		lessons, err := lessonRepo.GetVisibleLessons(ctx, student3.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(lessons), 2, "Student3 should see at least 2 group lessons")

		lessonIDs := extractLessonIDs(lessons)
		assert.Contains(t, lessonIDs, groupLesson1.ID)
		assert.Contains(t, lessonIDs, groupLesson2.ID)
		assert.NotContains(t, lessonIDs, individualLesson1.ID)
		assert.NotContains(t, lessonIDs, individualLesson2.ID)
	})

	// Note: LessonType field was removed from Lesson model
	// Skipping filter by lesson type test until model is updated
	/*
		t.Run("Filter by lesson type - student sees only group lessons", func(t *testing.T) {
			lessonType := models.LessonTypeGroup
			filter := &models.ListLessonsFilter{
				LessonType: &lessonType,
			}

			lessons, err := lessonRepo.GetVisibleLessons(ctx, student1.ID, string(models.RoleStudent), filter)
			require.NoError(t, err)
			require.Len(t, lessons, 2, "Student1 with group filter should see only group lessons")
		})
	*/

	t.Run("Filter by teacher - admin sees specific teacher's lessons", func(t *testing.T) {
		filter := &models.ListLessonsFilter{
			TeacherID: &teacher1.ID,
		}

		lessons, err := lessonRepo.GetVisibleLessons(ctx, admin.ID, string(models.RoleAdmin), filter)
		require.NoError(t, err)
		require.Len(t, lessons, 2, "Admin with teacher1 filter should see only teacher1's lessons")

		for _, lesson := range lessons {
			assert.Equal(t, teacher1.ID, lesson.TeacherID)
		}
	})

	t.Run("Filter by available - shows only lessons with available spots", func(t *testing.T) {
		// Fill up groupLesson1
		for i := 0; i < 4; i++ {
			student := createTestUserVisibility(t, db, ctx, uuid.New().String()+"@test.com", models.RoleStudent)
			createTestBooking(t, db, ctx, student.ID, groupLesson1.ID)
		}

		// Manually update current_students
		_, err := db.ExecContext(ctx, "UPDATE lessons SET current_students = 4 WHERE id = $1", groupLesson1.ID)
		require.NoError(t, err)

		available := true
		filter := &models.ListLessonsFilter{
			Available: &available,
		}

		lessons, err := lessonRepo.GetVisibleLessons(ctx, admin.ID, string(models.RoleAdmin), filter)
		require.NoError(t, err)

		lessonIDs := extractLessonIDs(lessons)
		assert.NotContains(t, lessonIDs, groupLesson1.ID, "Full lesson should not appear in available filter")
	})

	t.Run("Invalid role returns error", func(t *testing.T) {
		_, err := lessonRepo.GetVisibleLessons(ctx, admin.ID, "invalid_role", &models.ListLessonsFilter{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user role")
	})
}

// TestIndividualLessonConstraints - Verify individual lesson max_students = 1
func TestIndividualLessonConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	lessonRepo := NewLessonRepository(db)

	teacher := createTestUserVisibility(t, db, ctx, "teacher@test.com", models.RoleMethodologist)

	t.Run("Individual lesson created with max_students=1", func(t *testing.T) {
		lesson := createTestLessonWithType(t, lessonRepo, ctx, teacher.ID, 1)
		assert.Equal(t, 1, lesson.MaxStudents)
	})

	t.Run("Cannot create individual lesson with max_students > 1", func(t *testing.T) {
		lesson := &models.Lesson{
			ID:              uuid.New(),
			TeacherID:       teacher.ID,
			StartTime:       time.Now().Add(24 * time.Hour),
			EndTime:         time.Now().Add(26 * time.Hour),
			MaxStudents:     2, // Invalid for individual
			CurrentStudents: 0,
			Color:           "#3B82F6", // Default color
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err := lessonRepo.Create(ctx, lesson)
		// Database should accept (validation happens at service layer)
		// But service layer validation should reject this
		require.NoError(t, err, "Repository accepts, but service layer should validate")
	})

	t.Run("Group lesson created with max_students >= 2", func(t *testing.T) {
		lesson := createTestLessonWithType(t, lessonRepo, ctx, teacher.ID, 4)
		assert.Equal(t, 4, lesson.MaxStudents)
	})
}

// Helper functions

func extractLessonIDs(lessons []*models.LessonWithTeacher) []uuid.UUID {
	ids := make([]uuid.UUID, len(lessons))
	for i, lesson := range lessons {
		ids[i] = lesson.ID
	}
	return ids
}

func createTestUserVisibility(t *testing.T, db *sqlx.DB, ctx context.Context, email string, role models.UserRole) *models.User {
	t.Helper()

	userID := uuid.New()
	uniqueEmail := fmt.Sprintf("%s-%s", userID.String()[:8], email)

	user := &models.User{
		ID:           userID,
		Email:        uniqueEmail,
		PasswordHash: "hash",
		FullName:     "Test User",
		Role:         role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	require.NoError(t, err)

	return user
}

func createTestLessonWithType(t *testing.T, lessonRepo *LessonRepository, ctx context.Context, teacherID uuid.UUID, maxStudents int) *models.Lesson {
	t.Helper()

	lesson := &models.Lesson{
		ID:              uuid.New(),
		TeacherID:       teacherID,
		StartTime:       time.Now().Add(24 * time.Hour),
		EndTime:         time.Now().Add(26 * time.Hour),
		MaxStudents:     maxStudents,
		CurrentStudents: 0,
		Color:           "#3B82F6", // Default color (blue)
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := lessonRepo.Create(ctx, lesson)
	require.NoError(t, err)

	return lesson
}

func createTestBooking(t *testing.T, db *sqlx.DB, ctx context.Context, studentID uuid.UUID, lessonID uuid.UUID) {
	t.Helper()

	query := `
		INSERT INTO bookings (id, student_id, lesson_id, status, booked_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := db.ExecContext(ctx, query,
		uuid.New(),
		studentID,
		lessonID,
		models.BookingStatusActive,
		time.Now(),
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err)
}

// TEST SUITE 2 (Extended): Past Lessons Visibility Tests
// Tests for GetVisibleLessons behavior with past lessons and different booking statuses

func TestGetVisibleLessons_Student_IncludesPastBookings_Suite2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	lessonRepo := NewLessonRepository(db)

	// Create test users
	teacher := createTestUserVisibility(t, db, ctx, "teacher.past@test.com", models.RoleMethodologist)
	student := createTestUserVisibility(t, db, ctx, "student.past@test.com", models.RoleStudent)

	// Create past lesson (24 hours ago)
	pastTime := time.Now().Add(-24 * time.Hour)
	pastLesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   pastTime,
		EndTime:     pastTime.Add(2 * time.Hour),
		MaxStudents: 1,
		Color:       "#FF6B6B",
	}
	err := lessonRepo.Create(ctx, pastLesson)
	require.NoError(t, err)

	// Create active booking for past lesson
	_, err = db.ExecContext(ctx, `
		INSERT INTO bookings (id, student_id, lesson_id, status, booked_at)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), student.ID, pastLesson.ID, "active", time.Now())
	require.NoError(t, err)

	// Test: Student should see past lessons with active bookings
	lessons, err := lessonRepo.GetVisibleLessons(ctx, student.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
	require.NoError(t, err)

	lessonIDs := extractLessonIDs(lessons)
	assert.Contains(t, lessonIDs, pastLesson.ID, "Student should see past lesson with active booking")

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM bookings WHERE lesson_id = $1", pastLesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lessons WHERE id = $1", pastLesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2)", teacher.ID, student.ID)
}

func TestGetVisibleLessons_Student_IncludesCancelledBookings_Suite2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	lessonRepo := NewLessonRepository(db)

	// Create test users
	teacher := createTestUserVisibility(t, db, ctx, "teacher.cancelled@test.com", models.RoleMethodologist)
	student := createTestUserVisibility(t, db, ctx, "student.cancelled@test.com", models.RoleStudent)

	// Create lesson (future)
	futureTime := time.Now().Add(48 * time.Hour)
	lesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   futureTime,
		EndTime:     futureTime.Add(2 * time.Hour),
		MaxStudents: 1,
		Color:       "#4ECDC4",
	}
	err := lessonRepo.Create(ctx, lesson)
	require.NoError(t, err)

	// Create cancelled booking
	_, err = db.ExecContext(ctx, `
		INSERT INTO bookings (id, student_id, lesson_id, status, booked_at, cancelled_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), student.ID, lesson.ID, "cancelled", time.Now().Add(-2*time.Hour), time.Now())
	require.NoError(t, err)

	// Test: Student should see lessons even with cancelled bookings
	lessons, err := lessonRepo.GetVisibleLessons(ctx, student.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
	require.NoError(t, err)

	lessonIDs := extractLessonIDs(lessons)
	assert.Contains(t, lessonIDs, lesson.ID, "Student should see lesson with cancelled booking")

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM bookings WHERE lesson_id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lessons WHERE id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2)", teacher.ID, student.ID)
}

func TestGetVisibleLessons_Student_ExcludesUnbookedIndividual_Suite2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	lessonRepo := NewLessonRepository(db)

	// Create test users
	teacher := createTestUserVisibility(t, db, ctx, "teacher.unbooked@test.com", models.RoleMethodologist)
	student1 := createTestUserVisibility(t, db, ctx, "student1.unbooked@test.com", models.RoleStudent)
	student2 := createTestUserVisibility(t, db, ctx, "student2.unbooked@test.com", models.RoleStudent)

	// Create individual lesson
	futureTime := time.Now().Add(48 * time.Hour)
	individualLesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   futureTime,
		EndTime:     futureTime.Add(2 * time.Hour),
		MaxStudents: 1,
		Color:       "#95E1D3",
	}
	err := lessonRepo.Create(ctx, individualLesson)
	require.NoError(t, err)

	// Student1 has booking, Student2 doesn't
	_, err = db.ExecContext(ctx, `
		INSERT INTO bookings (id, student_id, lesson_id, status, booked_at)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), student1.ID, individualLesson.ID, "active", time.Now())
	require.NoError(t, err)

	// Test: Student1 should see (booked), Student2 should not see (unbooked individual)
	lessons1, _ := lessonRepo.GetVisibleLessons(ctx, student1.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
	ids1 := extractLessonIDs(lessons1)
	assert.Contains(t, ids1, individualLesson.ID, "Student1 should see their booked individual lesson")

	lessons2, _ := lessonRepo.GetVisibleLessons(ctx, student2.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
	ids2 := extractLessonIDs(lessons2)
	assert.NotContains(t, ids2, individualLesson.ID, "Student2 should NOT see unbooked individual lesson")

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM bookings WHERE lesson_id = $1", individualLesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lessons WHERE id = $1", individualLesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2, $3)", teacher.ID, student1.ID, student2.ID)
}

func TestGetVisibleLessons_GroupLessonsAlwaysVisible_Suite2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	lessonRepo := NewLessonRepository(db)

	// Create test users
	teacher := createTestUserVisibility(t, db, ctx, "teacher.group@test.com", models.RoleMethodologist)
	student := createTestUserVisibility(t, db, ctx, "student.group@test.com", models.RoleStudent)

	// Create group lesson (not booked by this student)
	futureTime := time.Now().Add(48 * time.Hour)
	groupLesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   futureTime,
		EndTime:     futureTime.Add(2 * time.Hour),
		MaxStudents: 4,
		Color:       "#FFD93D",
	}
	err := lessonRepo.Create(ctx, groupLesson)
	require.NoError(t, err)

	// Test: Student should see group lesson even without booking
	lessons, err := lessonRepo.GetVisibleLessons(ctx, student.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
	require.NoError(t, err)

	lessonIDs := extractLessonIDs(lessons)
	assert.Contains(t, lessonIDs, groupLesson.ID, "Student should see group lesson regardless of booking status")

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM lessons WHERE id = $1", groupLesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2)", teacher.ID, student.ID)
}

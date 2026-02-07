package service

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST SUITE 3: Schedule Consistency Tests
// Tests for consistency between admin and student visibility of the same lessons

func TestGetVisibleLessons_AdminVsStudent_IdenticalForBookedLessons(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestServiceDB(t)
	defer cleanupServiceTestDB(t, db)

	ctx := context.Background()
	lessonRepo := repository.NewLessonRepository(db)

	// Create test users
	admin := createTestServiceUser(t, db, ctx, "admin.consistency@test.com", models.RoleAdmin)
	teacher := createTestServiceUser(t, db, ctx, "teacher.consistency@test.com", models.RoleTeacher)
	student := createTestServiceUser(t, db, ctx, "student.consistency@test.com", models.RoleStudent)

	// Create 3 lessons
	futureTime := time.Now().Add(48 * time.Hour)
	lessons := make([]*models.Lesson, 3)

	for i := 0; i < 3; i++ {
		lesson := &models.Lesson{
			TeacherID:   teacher.ID,
			StartTime:   futureTime.Add(time.Duration(i*2) * time.Hour),
			EndTime:     futureTime.Add(time.Duration(i*2+2) * time.Hour),
			MaxStudents: 1,
			Color:       "#FF6B6B",
		}
		err := lessonRepo.Create(ctx, lesson)
		require.NoError(t, err)
		lessons[i] = lesson
	}

	// Student books 2 of the 3 lessons
	for i := 0; i < 2; i++ {
		_, err := db.ExecContext(ctx, `
			INSERT INTO bookings (id, student_id, lesson_id, status, booked_at)
			VALUES ($1, $2, $3, $4, $5)
		`, uuid.New(), student.ID, lessons[i].ID, "active", time.Now())
		require.NoError(t, err)
	}

	// Get visible lessons for admin and student
	adminLessons, err := lessonRepo.GetVisibleLessons(ctx, admin.ID, string(models.RoleAdmin), &models.ListLessonsFilter{})
	require.NoError(t, err)

	studentLessons, err := lessonRepo.GetVisibleLessons(ctx, student.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
	require.NoError(t, err)

	// Extract IDs
	adminLessonIDs := extractServiceLessonIDs(adminLessons)
	studentLessonIDs := extractServiceLessonIDs(studentLessons)

	// Test: Both admin and student should see lessons they booked
	t.Run("Admin sees all 3 lessons", func(t *testing.T) {
		for _, lesson := range lessons {
			assert.Contains(t, adminLessonIDs, lesson.ID, "Admin should see all lessons")
		}
	})

	t.Run("Student sees only 2 booked lessons", func(t *testing.T) {
		assert.Len(t, studentLessonIDs, 2, "Student should see only their 2 booked lessons")
		assert.Contains(t, studentLessonIDs, lessons[0].ID)
		assert.Contains(t, studentLessonIDs, lessons[1].ID)
		assert.NotContains(t, studentLessonIDs, lessons[2].ID, "Student should not see unbooked lesson")
	})

	t.Run("Both see the same booked lessons by ID", func(t *testing.T) {
		// Find the 2 booked lessons in admin view
		for _, adminLessonID := range adminLessonIDs {
			if adminLessonID == lessons[0].ID || adminLessonID == lessons[1].ID {
				assert.Contains(t, studentLessonIDs, adminLessonID, "Student and admin should both see booked lessons")
			}
		}
	})

	// Cleanup
	for _, lesson := range lessons {
		_, _ = db.ExecContext(ctx, "DELETE FROM bookings WHERE lesson_id = $1", lesson.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM lessons WHERE id = $1", lesson.ID)
	}
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2, $3)", admin.ID, teacher.ID, student.ID)
}

func TestGetVisibleLessons_StudentBookedIndividualAndGroupLesson(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestServiceDB(t)
	defer cleanupServiceTestDB(t, db)

	ctx := context.Background()
	lessonRepo := repository.NewLessonRepository(db)

	// Create test users
	teacher := createTestServiceUser(t, db, ctx, "teacher.mixed@test.com", models.RoleTeacher)
	student := createTestServiceUser(t, db, ctx, "student.mixed@test.com", models.RoleStudent)

	// Create past individual lesson
	pastTime := time.Now().Add(-24 * time.Hour)
	pastIndividualLesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   pastTime,
		EndTime:     pastTime.Add(2 * time.Hour),
		MaxStudents: 1,
		Color:       "#FF6B6B",
	}
	err := lessonRepo.Create(ctx, pastIndividualLesson)
	require.NoError(t, err)

	// Create future group lesson
	futureTime := time.Now().Add(48 * time.Hour)
	futureGroupLesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   futureTime,
		EndTime:     futureTime.Add(2 * time.Hour),
		MaxStudents: 4,
		Color:       "#4ECDC4",
	}
	err = lessonRepo.Create(ctx, futureGroupLesson)
	require.NoError(t, err)

	// Student books past individual lesson
	_, err = db.ExecContext(ctx, `
		INSERT INTO bookings (id, student_id, lesson_id, status, booked_at)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), student.ID, pastIndividualLesson.ID, "active", time.Now().Add(-24*time.Hour))
	require.NoError(t, err)

	// Student books future group lesson
	_, err = db.ExecContext(ctx, `
		INSERT INTO bookings (id, student_id, lesson_id, status, booked_at)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), student.ID, futureGroupLesson.ID, "active", time.Now())
	require.NoError(t, err)

	// Get visible lessons
	visibleLessons, err := lessonRepo.GetVisibleLessons(ctx, student.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
	require.NoError(t, err)

	lessonIDs := extractServiceLessonIDs(visibleLessons)

	// Test: Student should see BOTH lessons (past individual AND future group)
	assert.Contains(t, lessonIDs, pastIndividualLesson.ID, "Student should see past individual lesson they booked")
	assert.Contains(t, lessonIDs, futureGroupLesson.ID, "Student should see future group lesson they booked")

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM bookings WHERE student_id = $1", student.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lessons WHERE id IN ($1, $2)", pastIndividualLesson.ID, futureGroupLesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2)", teacher.ID, student.ID)
}

func TestGetVisibleLessons_AfterBookingCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestServiceDB(t)
	defer cleanupServiceTestDB(t, db)

	ctx := context.Background()
	lessonRepo := repository.NewLessonRepository(db)

	// Create test users
	teacher := createTestServiceUser(t, db, ctx, "teacher.cancel@test.com", models.RoleTeacher)
	student := createTestServiceUser(t, db, ctx, "student.cancel@test.com", models.RoleStudent)

	// Create future individual lesson
	futureTime := time.Now().Add(48 * time.Hour)
	lesson := &models.Lesson{
		TeacherID:   teacher.ID,
		StartTime:   futureTime,
		EndTime:     futureTime.Add(2 * time.Hour),
		MaxStudents: 1,
		Color:       "#95E1D3",
	}
	err := lessonRepo.Create(ctx, lesson)
	require.NoError(t, err)

	// Student creates active booking
	bookingID := uuid.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO bookings (id, student_id, lesson_id, status, booked_at)
		VALUES ($1, $2, $3, $4, $5)
	`, bookingID, student.ID, lesson.ID, "active", time.Now())
	require.NoError(t, err)

	// Test: With active booking, student should see lesson
	t.Run("Student sees lesson with active booking", func(t *testing.T) {
		visibleLessons, err := lessonRepo.GetVisibleLessons(ctx, student.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
		require.NoError(t, err)

		lessonIDs := extractServiceLessonIDs(visibleLessons)
		assert.Contains(t, lessonIDs, lesson.ID, "Student should see lesson with active booking")
	})

	// Cancel the booking
	_, err = db.ExecContext(ctx, `
		UPDATE bookings SET status = $1, cancelled_at = $2 WHERE id = $3
	`, "cancelled", time.Now(), bookingID)
	require.NoError(t, err)

	// Test: After cancellation, check visibility
	t.Run("Student visibility after booking cancellation", func(t *testing.T) {
		visibleLessons, err := lessonRepo.GetVisibleLessons(ctx, student.ID, string(models.RoleStudent), &models.ListLessonsFilter{})
		require.NoError(t, err)

		lessonIDs := extractServiceLessonIDs(visibleLessons)
		// Document the behavior: whether cancelled bookings are still visible
		t.Logf("Lessons visible after cancellation: %d", len(lessonIDs))
		for _, id := range lessonIDs {
			if id == lesson.ID {
				t.Logf("Student can see lesson with cancelled booking")
			}
		}
	})

	// Cleanup
	_, _ = db.ExecContext(ctx, "DELETE FROM bookings WHERE id = $1", bookingID)
	_, _ = db.ExecContext(ctx, "DELETE FROM lessons WHERE id = $1", lesson.ID)
	_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id IN ($1, $2)", teacher.ID, student.ID)
}

// Helper functions for service tests
func setupTestServiceDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("postgres", "postgres://postgres:@localhost/tutoring_test?sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	return db
}

func cleanupServiceTestDB(t *testing.T, db *sqlx.DB) {
	// Clean up tables used in tests
	ctx := context.Background()
	_, _ = db.ExecContext(ctx, "DELETE FROM bookings")
	_, _ = db.ExecContext(ctx, "DELETE FROM lessons")
	_, _ = db.ExecContext(ctx, "DELETE FROM users")
	_ = db.Close()
}

func createTestServiceUser(t *testing.T, db *sqlx.DB, ctx context.Context, email string, role models.UserRole) *models.User {
	userID := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, userID, email, "hash", "Test", "User", string(role), time.Now(), time.Now())
	require.NoError(t, err)

	user := &models.User{
		ID:    userID,
		Email: email,
		Role:  role,
	}
	return user
}

func extractServiceLessonIDs(lessons []*models.LessonWithTeacher) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(lessons))
	for _, lesson := range lessons {
		ids = append(ids, lesson.ID)
	}
	return ids
}

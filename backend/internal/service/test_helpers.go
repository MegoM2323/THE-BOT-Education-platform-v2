package service

import (
	"testing"
	"time"

	"tutoring-platform/internal/database"
	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection for integration tests
// CRITICAL: This function hardcodes 'tutoring_platform_test' to prevent accidental production data deletion
// NOTE: Uses GetTestSqlxDB which automatically applies all migrations
func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	// Use GetTestSqlxDB which automatically applies migrations and verifies schema
	db := database.GetTestSqlxDB(t)

	// Clean up existing data
	cleanupAllTables(t, db)

	return db
}

// cleanupTestDB cleans up test data but does NOT close the connection
// (connection is managed by GetTestSqlxDB as a shared resource)
func cleanupTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()

	if db != nil {
		cleanupAllTables(t, db)
		// DO NOT close the connection - it's a shared resource managed by GetTestSqlxDB
	}
}

// cleanupAllTables deletes all data from test tables
// CRITICAL: This function contains safety checks to prevent accidental production data deletion
func cleanupAllTables(t *testing.T, db *sqlx.DB) {
	t.Helper()

	// CRITICAL SAFETY CHECK: Verify we're on test database before deleting anything
	var currentDB string
	err := db.Get(&currentDB, "SELECT current_database()")
	if err != nil {
		t.Fatalf("CRITICAL: Failed to verify database name before cleanup: %v", err)
	}
	if currentDB != "tutoring_platform_test" {
		t.Fatalf("CRITICAL SAFETY VIOLATION: cleanupAllTables attempted on database '%s' instead of 'tutoring_platform_test'. "+
			"This would DELETE ALL DATA! Aborting.", currentDB)
	}

	// Order matters due to foreign key constraints
	tables := []string{
		"lesson_modifications",
		"credit_transactions",
		"swaps",
		"bookings",
		"lessons",
		"template_lesson_students",
		"template_lessons",
		"template_applications",
		"lesson_templates",
		"broadcasts",
		"broadcast_lists",
		"chat_rooms",
		"messages",
		"telegram_users",
		"credits",
		"sessions",
		"trial_requests",
		"users",
	}

	for _, table := range tables {
		// Use TRUNCATE CASCADE to remove all dependent records
		_, _ = db.Exec("TRUNCATE TABLE " + table + " CASCADE")
	}
}

// createTestUser creates a test user in the database
func createTestUser(t *testing.T, db *sqlx.DB, email, fullName, role string) *models.User {
	user := &models.User{
		ID:             uuid.New(),
		Email:          email,
		PasswordHash:   "hashed_password",
		FullName:       fullName,
		Role:           models.UserRole(role),
		PaymentEnabled: true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err := db.Exec(`
		INSERT INTO users (id, email, password_hash, full_name, role, payment_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.Email, user.PasswordHash, user.FullName, user.Role, user.PaymentEnabled, user.CreatedAt, user.UpdatedAt)
	require.NoError(t, err, "Failed to create test user")

	// Примечание: кредиты для студентов создаются автоматически триггером create_credits_on_student_insert
	// Для других ролей создаём вручную
	if role != "student" {
		_, err = db.Exec(`
			INSERT INTO credits (id, user_id, balance, created_at, updated_at)
			VALUES ($1, $2, 0, $3, $4)
			ON CONFLICT (user_id) DO NOTHING
		`, uuid.New(), user.ID, time.Now(), time.Now())
		require.NoError(t, err, "Failed to create credits")
	}

	return user
}

// addCreditToStudent adds credits to a student's account
func addCreditToStudent(t *testing.T, db *sqlx.DB, studentID uuid.UUID, amount int) {
	_, err := db.Exec(`
		UPDATE credits
		SET balance = balance + $1, updated_at = $2
		WHERE user_id = $3
	`, amount, time.Now(), studentID)
	require.NoError(t, err, "Failed to add credits to student")
}

// getCreditBalance retrieves the current credit balance for a user
func getCreditBalance(t *testing.T, db *sqlx.DB, userID uuid.UUID) int {
	var balance int
	err := db.Get(&balance, "SELECT balance FROM credits WHERE user_id = $1", userID)
	require.NoError(t, err, "Failed to get credit balance")
	return balance
}

// createTestTemplate creates a test template in the database
func createTestTemplate(t *testing.T, db *sqlx.DB, adminID uuid.UUID, name string, lessons []*models.CreateTemplateLessonRequest) *models.LessonTemplate {
	templateID := uuid.New()

	_, err := db.Exec(`
		INSERT INTO lesson_templates (id, admin_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, templateID, adminID, name, time.Now(), time.Now())
	require.NoError(t, err, "Failed to create test template")

	template := &models.LessonTemplate{
		ID:        templateID,
		AdminID:   adminID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, lesson := range lessons {
		lesson.ApplyDefaults()

		entryID := uuid.New()
		_, err = db.Exec(`
			INSERT INTO template_lessons (id, template_id, day_of_week, start_time, end_time, teacher_id, lesson_type, max_students, color, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, entryID, templateID, lesson.DayOfWeek, lesson.StartTime, *lesson.EndTime, lesson.TeacherID, *lesson.LessonType, *lesson.MaxStudents, *lesson.Color, time.Now(), time.Now())
		require.NoError(t, err, "Failed to insert template lesson")

		for _, studentID := range lesson.StudentIDs {
			_, err = db.Exec(`
				INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at)
				VALUES ($1, $2, $3, $4)
			`, uuid.New(), entryID, studentID, time.Now())
			require.NoError(t, err, "Failed to insert template student assignment")
		}

		entry := &models.TemplateLessonEntry{
			ID:          entryID,
			TemplateID:  templateID,
			DayOfWeek:   lesson.DayOfWeek,
			StartTime:   lesson.StartTime,
			EndTime:     *lesson.EndTime,
			TeacherID:   lesson.TeacherID,
			LessonType:  *lesson.LessonType,
			MaxStudents: *lesson.MaxStudents,
			Color:       *lesson.Color,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		template.Lessons = append(template.Lessons, entry)
	}

	return template
}

// getNextMonday returns the next Monday at 00:00
func getNextMonday() time.Time {
	now := time.Now()
	daysUntilMonday := (int(time.Monday) - int(now.Weekday()) + 7) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	return now.AddDate(0, 0, daysUntilMonday)
}

// Helper functions for template tests
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

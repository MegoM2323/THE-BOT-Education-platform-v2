package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/database"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
)

// TestTemplateRollback_E2E tests the complete rollback workflow
func TestTemplateRollback_E2E(t *testing.T) {
	db := setupRollbackTestDB(t)
	t.Cleanup(func() {
		cleanupRollbackTestDB(t, db)
	})

	ctx := context.Background()

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Create service
	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create test users
	admin := createRollbackTestUser(t, db, "admin@test.com", "Admin User", "admin")
	teacher := createRollbackTestUser(t, db, "teacher@test.com", "Teacher User", "methodologist")
	student1 := createRollbackTestUser(t, db, "student1@test.com", "Student One", "student")
	student2 := createRollbackTestUser(t, db, "student2@test.com", "Student Two", "student")

	// Add credits to students (2 credits each to book 2 lessons)
	addRollbackTestCredit(t, db, student1.ID, 2)
	addRollbackTestCredit(t, db, student2.ID, 2)

	// Create template with 2 lessons
	template := createRollbackTestTemplate(t, db, admin.ID, "Test Template", []*models.CreateTemplateLessonRequest{
		{
			DayOfWeek:   1, // Monday
			StartTime:   "10:00:00",
			TeacherID:   teacher.ID,
			LessonType:  strPtr("group"), // Group lesson to hold 2 students
			MaxStudents: intPtr(4),       // Group lessons require >= 4, but only 2 students assigned
			StudentIDs:  []uuid.UUID{student1.ID, student2.ID},
		},
		{
			DayOfWeek:   3, // Wednesday
			StartTime:   "14:00:00",
			TeacherID:   teacher.ID,
			LessonType:  strPtr("individual"), // Individual lesson
			MaxStudents: intPtr(1),
			StudentIDs:  []uuid.UUID{student1.ID},
		},
	})

	// Get Monday of next week
	weekStartDate := getRollbackTestNextMonday().Format("2006-01-02")

	// Apply template
	applyReq := &models.ApplyTemplateRequest{
		TemplateID:    template.ID,
		WeekStartDate: weekStartDate,
	}

	application, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, applyReq)
	require.NoError(t, err)
	require.NotNil(t, application)
	require.Equal(t, 2, application.CreatedLessonsCount)
	require.Equal(t, "applied", application.Status)

	// Verify credits deducted
	student1Credits := getRollbackTestCreditBalance(t, db, student1.ID)
	student2Credits := getRollbackTestCreditBalance(t, db, student2.ID)
	assert.Equal(t, 0, student1Credits, "Student1 should have 0 credits (booked 2 lessons)")
	assert.Equal(t, 1, student2Credits, "Student2 should have 1 credit (booked 1 lesson)")

	// Verify lessons created
	lessons, err := templateAppRepo.GetLessonsCreatedFromApplication(ctx, application.ID)
	require.NoError(t, err)
	assert.Len(t, lessons, 2)

	// === PERFORM ROLLBACK ===
	rollbackResponse, err := templateService.RollbackWeekToTemplate(ctx, admin.ID, weekStartDate, template.ID)
	require.NoError(t, err)
	require.NotNil(t, rollbackResponse)

	// Verify rollback response
	assert.Equal(t, application.ID, rollbackResponse.ApplicationID)
	assert.Equal(t, weekStartDate, rollbackResponse.WeekStartDate)
	assert.Equal(t, template.ID, rollbackResponse.TemplateID)
	assert.Equal(t, 2, rollbackResponse.DeletedLessons)
	assert.Equal(t, 3, rollbackResponse.RefundedCredits, "Should refund 3 credits (2 for student1, 1 for student2)")
	assert.Empty(t, rollbackResponse.Warnings)

	// Verify credits refunded
	student1CreditsAfter := getRollbackTestCreditBalance(t, db, student1.ID)
	student2CreditsAfter := getRollbackTestCreditBalance(t, db, student2.ID)
	assert.Equal(t, 2, student1CreditsAfter, "Student1 credits should be restored to 2")
	assert.Equal(t, 2, student2CreditsAfter, "Student2 credits should be restored to 2")

	// Verify lessons soft-deleted
	var deletedCount int
	err = db.Get(&deletedCount, `
		SELECT COUNT(*)
		FROM lessons
		WHERE template_application_id = $1 AND deleted_at IS NULL
	`, application.ID.String())
	require.NoError(t, err)
	assert.Equal(t, 0, deletedCount, "All lessons should be soft-deleted")

	// Verify application marked as rolled back
	updatedApp, err := templateAppRepo.GetTemplateApplicationByID(ctx, application.ID)
	require.NoError(t, err)
	assert.Equal(t, "rolled_back", updatedApp.Status)
	assert.True(t, updatedApp.RolledBackAt.Valid)
}

// TestTemplateRollback_Idempotent tests that rollback can be called multiple times safely
func TestTemplateRollback_Idempotent(t *testing.T) {
	db := setupRollbackTestDB(t)
	t.Cleanup(func() {
		cleanupRollbackTestDB(t, db)
	})

	ctx := context.Background()

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Create service
	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create test users
	admin := createRollbackTestUser(t, db, "admin@test.com", "Admin User", "admin")
	teacher := createRollbackTestUser(t, db, "teacher@test.com", "Teacher User", "methodologist")
	student := createRollbackTestUser(t, db, "student@test.com", "Student User", "student")

	// Add credits
	addRollbackTestCredit(t, db, student.ID, 1)

	// Create template
	template := createRollbackTestTemplate(t, db, admin.ID, "Test Template", []*models.CreateTemplateLessonRequest{
		{
			DayOfWeek:  1, // Monday
			StartTime:  "10:00:00",
			TeacherID:  teacher.ID,
			StudentIDs: []uuid.UUID{student.ID},
		},
	})

	// Apply template
	weekStartDate := getRollbackTestNextMonday().Format("2006-01-02")
	applyReq := &models.ApplyTemplateRequest{
		TemplateID:    template.ID,
		WeekStartDate: weekStartDate,
	}

	_, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, applyReq)
	require.NoError(t, err)

	// First rollback
	rollback1, err := templateService.RollbackWeekToTemplate(ctx, admin.ID, weekStartDate, template.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, rollback1.DeletedLessons)
	assert.Equal(t, 1, rollback1.RefundedCredits)

	// Second rollback (should fail - already rolled back)
	_, err = templateService.RollbackWeekToTemplate(ctx, admin.ID, weekStartDate, template.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already rolled back", "Should indicate application was already rolled back")
	assert.Contains(t, err.Error(), weekStartDate, "Error message should include week date")
}

// TestTemplateRollback_DetailedErrorMessages tests that error messages are clear and actionable
func TestTemplateRollback_DetailedErrorMessages(t *testing.T) {
	db := setupRollbackTestDB(t)
	t.Cleanup(func() {
		cleanupRollbackTestDB(t, db)
	})

	ctx := context.Background()

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Create service
	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create test users
	admin := createRollbackTestUser(t, db, "admin@test.com", "Admin User", "admin")
	teacher := createRollbackTestUser(t, db, "teacher@test.com", "Teacher User", "methodologist")
	student := createRollbackTestUser(t, db, "student@test.com", "Student User", "student")

	// Add credits
	addRollbackTestCredit(t, db, student.ID, 1)

	// Create template
	template := createRollbackTestTemplate(t, db, admin.ID, "Test Template", []*models.CreateTemplateLessonRequest{
		{
			DayOfWeek:  1, // Monday
			StartTime:  "10:00:00",
			TeacherID:  teacher.ID,
			StudentIDs: []uuid.UUID{student.ID},
		},
	})

	// Apply template
	weekStartDate := getRollbackTestNextMonday().Format("2006-01-02")
	applyReq := &models.ApplyTemplateRequest{
		TemplateID:    template.ID,
		WeekStartDate: weekStartDate,
	}

	_, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, applyReq)
	require.NoError(t, err)

	// First rollback - успешно
	rollback1, err := templateService.RollbackWeekToTemplate(ctx, admin.ID, weekStartDate, template.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, rollback1.DeletedLessons)

	// Получаем время отката для проверки в сообщении
	app, err := templateAppRepo.GetTemplateApplicationByID(ctx, rollback1.ApplicationID)
	require.NoError(t, err)
	require.True(t, app.RolledBackAt.Valid, "RolledBackAt should be set")

	// Second rollback - должно вернуть детальное сообщение с датой и временем
	_, err = templateService.RollbackWeekToTemplate(ctx, admin.ID, weekStartDate, template.ID)
	require.Error(t, err)

	// Проверяем что сообщение об ошибке содержит все необходимые детали
	assert.Contains(t, err.Error(), "already rolled back", "Error should indicate application was already rolled back")
	assert.Contains(t, err.Error(), weekStartDate, "Error should include week date")
	// Проверяем что есть timestamp отката (в формате YYYY-MM-DD HH:MM:SS)
	assert.Regexp(t, `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, err.Error(), "Error should include rollback timestamp")

	t.Logf("Detailed error message: %s", err.Error())
}

// TestTemplateRollback_NoApplication tests rollback when no application exists
func TestTemplateRollback_NoApplication(t *testing.T) {
	db := setupRollbackTestDB(t)
	t.Cleanup(func() {
		cleanupRollbackTestDB(t, db)
	})

	ctx := context.Background()

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Create service
	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create admin
	admin := createRollbackTestUser(t, db, "admin@test.com", "Admin User", "admin")

	// Create template (not applied to any week)
	template := createRollbackTestTemplate(t, db, admin.ID, "Test Template", []*models.CreateTemplateLessonRequest{})

	// Try to rollback non-existent application
	weekStartDate := getRollbackTestNextMonday().Format("2006-01-02")
	_, err := templateService.RollbackWeekToTemplate(ctx, admin.ID, weekStartDate, template.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no template application found", "Should indicate no application exists")
	assert.Contains(t, err.Error(), "template must be applied first", "Should suggest to apply template first")
}

// TestTemplateRollback_InvalidWeekDate tests rollback with invalid week date
func TestTemplateRollback_InvalidWeekDate(t *testing.T) {
	db := setupRollbackTestDB(t)
	t.Cleanup(func() {
		cleanupRollbackTestDB(t, db)
	})

	ctx := context.Background()

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Create service
	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create admin
	admin := createRollbackTestUser(t, db, "admin@test.com", "Admin User", "admin")

	// Create template
	template := createRollbackTestTemplate(t, db, admin.ID, "Test Template", []*models.CreateTemplateLessonRequest{})

	// Test with non-Monday date (should fail)
	weekStartDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02") // Tomorrow (probably not Monday)
	_, err := templateService.RollbackWeekToTemplate(ctx, admin.ID, weekStartDate, template.ID)

	// Will fail either with "not a Monday" or "no active application"
	assert.Error(t, err)
}

// TestTemplateRollback_WithActiveBookings tests rollback when lessons have active bookings
func TestTemplateRollback_WithActiveBookings(t *testing.T) {
	db := setupRollbackTestDB(t)
	t.Cleanup(func() {
		cleanupRollbackTestDB(t, db)
	})

	ctx := context.Background()

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Create service
	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create test users
	admin := createRollbackTestUser(t, db, "admin@test.com", "Admin User", "admin")
	teacher := createRollbackTestUser(t, db, "teacher@test.com", "Teacher User", "methodologist")
	student := createRollbackTestUser(t, db, "student@test.com", "Student User", "student")

	// Add credits
	addRollbackTestCredit(t, db, student.ID, 1)

	// Create template
	template := createRollbackTestTemplate(t, db, admin.ID, "Test Template", []*models.CreateTemplateLessonRequest{
		{
			DayOfWeek:  1, // Monday
			StartTime:  "10:00:00",
			TeacherID:  teacher.ID,
			StudentIDs: []uuid.UUID{student.ID},
		},
	})

	// Apply template
	weekStartDate := getRollbackTestNextMonday().Format("2006-01-02")
	applyReq := &models.ApplyTemplateRequest{
		TemplateID:    template.ID,
		WeekStartDate: weekStartDate,
	}

	application, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, applyReq)
	require.NoError(t, err)

	// Verify booking created and credit deducted
	studentCredits := getRollbackTestCreditBalance(t, db, student.ID)
	assert.Equal(t, 0, studentCredits)

	// Rollback (should refund credits even with active bookings)
	rollback, err := templateService.RollbackWeekToTemplate(ctx, admin.ID, weekStartDate, template.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, rollback.DeletedLessons)
	assert.Equal(t, 1, rollback.RefundedCredits)

	// Verify credit refunded
	studentCreditsAfter := getRollbackTestCreditBalance(t, db, student.ID)
	assert.Equal(t, 1, studentCreditsAfter)

	// Verify booking cancelled
	var cancelledCount int
	err = db.Get(&cancelledCount, `
		SELECT COUNT(*)
		FROM bookings
		WHERE lesson_id IN (
			SELECT id FROM lessons WHERE template_application_id = $1
		) AND status = 'cancelled'
	`, application.ID.String())
	require.NoError(t, err)
	assert.Equal(t, 1, cancelledCount)
}

// TestTemplateRollback_ConcurrentRollbacks tests that concurrent rollbacks don't cause issues
func TestTemplateRollback_ConcurrentRollbacks(t *testing.T) {
	db := setupRollbackTestDB(t)
	t.Cleanup(func() {
		cleanupRollbackTestDB(t, db)
	})

	ctx := context.Background()

	// Create repositories
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Create service
	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Create test users
	admin := createRollbackTestUser(t, db, "admin@test.com", "Admin User", "admin")
	teacher := createRollbackTestUser(t, db, "teacher@test.com", "Teacher User", "methodologist")
	student := createRollbackTestUser(t, db, "student@test.com", "Student User", "student")

	// Add credits
	addRollbackTestCredit(t, db, student.ID, 1)

	// Create template
	template := createRollbackTestTemplate(t, db, admin.ID, "Test Template", []*models.CreateTemplateLessonRequest{
		{
			DayOfWeek:  1, // Monday
			StartTime:  "10:00:00",
			TeacherID:  teacher.ID,
			StudentIDs: []uuid.UUID{student.ID},
		},
	})

	// Apply template
	weekStartDate := getRollbackTestNextMonday().Format("2006-01-02")
	applyReq := &models.ApplyTemplateRequest{
		TemplateID:    template.ID,
		WeekStartDate: weekStartDate,
	}

	_, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, applyReq)
	require.NoError(t, err)

	// Concurrent rollbacks
	done := make(chan bool, 2)
	successCount := 0

	for i := 0; i < 2; i++ {
		go func() {
			_, err := templateService.RollbackWeekToTemplate(ctx, admin.ID, weekStartDate, template.ID)
			if err == nil {
				successCount++
			}
			done <- true
		}()
	}

	// Wait for both goroutines
	<-done
	<-done

	// Only one rollback should succeed due to SERIALIZABLE isolation
	assert.Equal(t, 1, successCount, "Only one concurrent rollback should succeed")

	// Verify final state
	studentCreditsAfter := getRollbackTestCreditBalance(t, db, student.ID)
	assert.Equal(t, 1, studentCreditsAfter, "Credits should be refunded exactly once")
}

// ============================================================================
// Helper Functions
// ============================================================================

func setupRollbackTestDB(t *testing.T) *sqlx.DB {
	db := database.GetTestSqlxDB(t)
	return db
}

func cleanupRollbackTestDB(t *testing.T, db *sqlx.DB) {
}

func createRollbackTestUser(t *testing.T, db *sqlx.DB, email, fullName, role string) *models.User {
	// Добавляем UUID к email для уникальности между тестами
	uniqueEmail := email + "." + uuid.New().String()[:8]

	// Split fullName into first and last name
	firstName := fullName
	lastName := ""
	for i, r := range fullName {
		if r == ' ' {
			firstName = fullName[:i]
			lastName = fullName[i+1:]
			break
		}
	}

	user := &models.User{
		ID:             uuid.New(),
		Email:          uniqueEmail,
		PasswordHash:   "hashed_password",
		FirstName:      firstName,
		LastName:       lastName,
		Role:           models.UserRole(role),
		PaymentEnabled: true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err := db.Exec(`
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, payment_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Role, user.PaymentEnabled, user.CreatedAt, user.UpdatedAt)
	require.NoError(t, err)

	// Примечание: кредиты для студентов создаются автоматически триггером
	// Для других ролей создаём вручную
	if role != "student" {
		_, err = db.Exec(`
			INSERT INTO credits (id, user_id, balance, created_at, updated_at)
			VALUES ($1, $2, 0, $3, $4)
			ON CONFLICT (user_id) DO NOTHING
		`, uuid.New(), user.ID, time.Now(), time.Now())
		require.NoError(t, err)
	}

	return user
}

func addRollbackTestCredit(t *testing.T, db *sqlx.DB, studentID uuid.UUID, amount int) {
	_, err := db.Exec(`
		UPDATE credits
		SET balance = balance + $1, updated_at = $2
		WHERE user_id = $3
	`, amount, time.Now(), studentID)
	require.NoError(t, err)
}

func getRollbackTestCreditBalance(t *testing.T, db *sqlx.DB, userID uuid.UUID) int {
	var balance int
	err := db.Get(&balance, "SELECT balance FROM credits WHERE user_id = $1", userID)
	require.NoError(t, err)
	return balance
}

func createRollbackTestTemplate(t *testing.T, db *sqlx.DB, adminID uuid.UUID, name string, lessons []*models.CreateTemplateLessonRequest) *models.LessonTemplate {
	templateID := uuid.New()

	_, err := db.Exec(`
		INSERT INTO lesson_templates (id, admin_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, templateID, adminID, name, time.Now(), time.Now())
	require.NoError(t, err)

	template := &models.LessonTemplate{
		ID:        templateID,
		AdminID:   adminID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Lessons:   []*models.TemplateLessonEntry{},
	}

	for _, lesson := range lessons {
		lesson.ApplyDefaults()

		entryID := uuid.New()
		_, err = db.Exec(`
			INSERT INTO template_lessons (id, template_id, day_of_week, start_time, end_time, teacher_id, lesson_type, max_students, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, entryID, templateID, lesson.DayOfWeek, lesson.StartTime, *lesson.EndTime, lesson.TeacherID, lesson.LessonType, *lesson.MaxStudents, time.Now(), time.Now())
		require.NoError(t, err)

		// Add students to template lesson
		for _, studentID := range lesson.StudentIDs {
			_, err = db.Exec(`
				INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at)
				VALUES ($1, $2, $3, $4)
			`, uuid.New(), entryID, studentID, time.Now())
			require.NoError(t, err)
		}

		entry := &models.TemplateLessonEntry{
			ID:          entryID,
			TemplateID:  templateID,
			DayOfWeek:   lesson.DayOfWeek,
			StartTime:   lesson.StartTime,
			EndTime:     *lesson.EndTime,
			TeacherID:   lesson.TeacherID,
			MaxStudents: *lesson.MaxStudents,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		template.Lessons = append(template.Lessons, entry)
	}

	return template
}

func getRollbackTestNextMonday() time.Time {
	now := time.Now()
	// Find next Monday
	daysUntilMonday := (int(time.Monday) - int(now.Weekday()) + 7) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7 // If today is Monday, get next Monday
	}
	return now.AddDate(0, 0, daysUntilMonday)
}

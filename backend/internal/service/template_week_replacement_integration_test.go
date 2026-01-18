package service

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTemplateWithLessons создает шаблон с указанными занятиями и студентами
func createTemplateWithLessons(
	t *testing.T,
	db *sqlx.DB,
	adminID uuid.UUID,
	templateName string,
	lessonsConfig []struct {
		dayOfWeek   int
		startTime   string
		endTime     string
		teacherID   uuid.UUID
		students    []uuid.UUID
		maxStudents int
	},
) uuid.UUID {
	templateID := uuid.New()

	// Создаем шаблон
	_, err := db.Exec(`
		INSERT INTO lesson_templates (id, admin_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, templateID, adminID, templateName, time.Now(), time.Now())
	require.NoError(t, err)

	// Добавляем занятия к шаблону
	for _, lessonCfg := range lessonsConfig {
		lessonID := uuid.New()

		_, err := db.Exec(`
			INSERT INTO template_lessons
			(id, template_id, day_of_week, start_time, end_time, teacher_id, lesson_type, max_students, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, lessonID, templateID, lessonCfg.dayOfWeek, lessonCfg.startTime, lessonCfg.endTime,
			lessonCfg.teacherID, "group", lessonCfg.maxStudents, time.Now(), time.Now())
		require.NoError(t, err)

		// Добавляем студентов к занятию
		for _, studentID := range lessonCfg.students {
			_, err := db.Exec(`
				INSERT INTO template_lesson_students (id, template_lesson_id, student_id, created_at)
				VALUES ($1, $2, $3, $4)
			`, uuid.New(), lessonID, studentID, time.Now())
			require.NoError(t, err)
		}
	}

	return templateID
}

// TestTemplateReplacement_CleanApply проверяет применение шаблона к пустой неделе
// Сценарий 1.1: Clean Apply (no existing lessons)
func TestTemplateReplacement_CleanApply(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Пропускаем если БД недоступна
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	// Setup: Создаем пользователей
	admin := createTestUser(t, db, "admin-replace@test.com", "Admin Replace", string(models.RoleAdmin))
	teacher := createTestUser(t, db, "teacher-replace@test.com", "Teacher Replace", string(models.RoleMethodologist))
	student1 := createTestUser(t, db, "student-replace1@test.com", "Student Replace 1", string(models.RoleStudent))
	student2 := createTestUser(t, db, "student-replace2@test.com", "Student Replace 2", string(models.RoleStudent))

	// Даем студентам кредиты
	addCreditToStudent(t, db, student1.ID, 20)
	addCreditToStudent(t, db, student2.ID, 20)

	// Создаем шаблон с 5 занятиями
	templateID := createTemplateWithLessons(t, db, admin.ID, "Clean Apply Template", []struct {
		dayOfWeek   int
		startTime   string
		endTime     string
		teacherID   uuid.UUID
		students    []uuid.UUID
		maxStudents int
	}{
		{0, "09:00:00", "11:00:00", teacher.ID, []uuid.UUID{student1.ID, student2.ID}, 4},
		{1, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student1.ID, student2.ID}, 4},
		{2, "11:00:00", "13:00:00", teacher.ID, []uuid.UUID{student1.ID, student2.ID}, 4},
		{3, "14:00:00", "16:00:00", teacher.ID, []uuid.UUID{student1.ID, student2.ID}, 4},
		{4, "15:00:00", "17:00:00", teacher.ID, []uuid.UUID{student1.ID, student2.ID}, 4},
	})

	// Инициализируем сервисы
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Применяем шаблон к пустой неделе
	weekDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC) // Понедельник

	req := &models.ApplyTemplateRequest{
		TemplateID:    templateID,
		WeekStartDate: "2026-02-02",
		DryRun:        false,
	}

	// Action: Применяем шаблон
	app, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, req)

	// Assertions
	assert.NoError(t, err, "Apply должен успешно завершиться")
	assert.NotNil(t, app, "Application должна быть создана")
	assert.Equal(t, "applied", app.Status, "Status должен быть 'applied'")

	// Проверяем что cleanup был выполнен, но ничего не удалено (чистая неделя)
	assert.NotNil(t, app.CleanupStats, "CleanupStats должен быть не nil")
	assert.Equal(t, 0, app.CleanupStats.CancelledBookings, "Cancelled bookings should be 0 for clean apply")
	assert.Equal(t, 0, app.CleanupStats.DeletedLessons, "Deleted lessons should be 0 for clean apply")

	// Проверяем creation stats
	assert.NotNil(t, app.CreationStats, "CreationStats должны быть nil")
	assert.Equal(t, 5, app.CreationStats.CreatedLessons, "Должно быть создано 5 занятий")
	assert.Equal(t, 10, app.CreationStats.CreatedBookings, "Должно быть 10 бронирований (5 lessons * 2 students)")
	assert.Equal(t, 10, app.CreationStats.DeductedCredits, "Должно быть списано 10 кредитов")

	// Проверяем что lessons созданы в БД
	var lessonCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM lessons
		WHERE start_time >= $1 AND start_time < $2 AND deleted_at IS NULL
	`, weekDate, weekDate.AddDate(0, 0, 7)).Scan(&lessonCount)
	assert.NoError(t, err)
	assert.Equal(t, 5, lessonCount, "В БД должно быть 5 несданных занятий")

	// Проверяем что bookings созданы
	var bookingCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM bookings
		WHERE status = 'active'
	`).Scan(&bookingCount)
	assert.NoError(t, err)
	assert.Equal(t, 10, bookingCount, "В БД должно быть 10 активных бронирований")

	// Проверяем баланс кредитов (каждый студент потратил 5 кредитов)
	var balance1, balance2 int
	err = db.QueryRow(`SELECT balance FROM credits WHERE user_id = $1`, student1.ID).Scan(&balance1)
	assert.NoError(t, err)
	assert.Equal(t, 15, balance1, "Студент 1 должен иметь 15 кредитов (20 - 5)")

	err = db.QueryRow(`SELECT balance FROM credits WHERE user_id = $1`, student2.ID).Scan(&balance2)
	assert.NoError(t, err)
	assert.Equal(t, 15, balance2, "Студент 2 должен иметь 15 кредитов (20 - 5)")

	t.Log("✓ Test 1.1 PASSED: Clean Apply")
}

// TestTemplateReplacement_ReplaceWithExisting проверяет замену существующих занятий
// Сценарий 1.2: Replace Apply (with existing lessons)
func TestTemplateReplacement_ReplaceWithExisting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	// Setup: Создаем пользователей
	admin := createTestUser(t, db, "admin-replace2@test.com", "Admin Replace 2", string(models.RoleAdmin))
	teacher1 := createTestUser(t, db, "teacher-replace1a@test.com", "Teacher Replace 1A", string(models.RoleMethodologist))
	teacher2 := createTestUser(t, db, "teacher-replace2a@test.com", "Teacher Replace 2A", string(models.RoleMethodologist))
	student1 := createTestUser(t, db, "student-replace1a@test.com", "Student Replace 1A", string(models.RoleStudent))
	student2 := createTestUser(t, db, "student-replace2a@test.com", "Student Replace 2A", string(models.RoleStudent))
	student3 := createTestUser(t, db, "student-replace3a@test.com", "Student Replace 3A", string(models.RoleStudent))

	// Даем кредиты
	addCreditToStudent(t, db, student1.ID, 30)
	addCreditToStudent(t, db, student2.ID, 30)
	addCreditToStudent(t, db, student3.ID, 30)

	// Инициализируем сервисы
	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	weekDate := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)

	// STEP 1: Применяем первый шаблон (3 занятия, 2 студента = 6 бронирований)
	template1ID := createTemplateWithLessons(t, db, admin.ID, "Template A", []struct {
		dayOfWeek   int
		startTime   string
		endTime     string
		teacherID   uuid.UUID
		students    []uuid.UUID
		maxStudents int
	}{
		{0, "09:00:00", "11:00:00", teacher1.ID, []uuid.UUID{student1.ID, student2.ID}, 4},
		{2, "10:00:00", "12:00:00", teacher1.ID, []uuid.UUID{student1.ID, student2.ID}, 4},
		{4, "14:00:00", "16:00:00", teacher1.ID, []uuid.UUID{student1.ID, student2.ID}, 4},
	})

	req1 := &models.ApplyTemplateRequest{
		TemplateID:    template1ID,
		WeekStartDate: "2026-02-09",
		DryRun:        false,
	}

	app1, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, req1)
	require.NoError(t, err, "First apply should succeed")
	require.NotNil(t, app1, "app1 should not be nil")
	assert.NotNil(t, app1.CleanupStats, "First apply should have cleanup stats (empty)")
	assert.Equal(t, 0, app1.CleanupStats.DeletedLessons, "First apply should have 0 deleted lessons")
	assert.Equal(t, 6, app1.CreationStats.CreatedBookings, "First apply: 6 bookings (3 lessons * 2 students)")

	// STEP 2: Проверяем начальное состояние БД
	var lessonsAfterFirst, bookingsAfterFirst int
	db.QueryRow(`SELECT COUNT(*) FROM lessons WHERE start_time >= $1 AND start_time < $2 AND deleted_at IS NULL`,
		weekDate, weekDate.AddDate(0, 0, 7)).Scan(&lessonsAfterFirst)
	db.QueryRow(`SELECT COUNT(*) FROM bookings WHERE status = 'active'`).Scan(&bookingsAfterFirst)

	assert.Equal(t, 3, lessonsAfterFirst, "After first apply: 3 lessons")
	assert.Equal(t, 6, bookingsAfterFirst, "After first apply: 6 bookings")

	// STEP 3: Применяем второй шаблон для замены (4 занятия, 3 студента = 12 бронирований)
	// Используем разные времена чтобы не было конфликтов расписания
	template2ID := createTemplateWithLessons(t, db, admin.ID, "Template B", []struct {
		dayOfWeek   int
		startTime   string
		endTime     string
		teacherID   uuid.UUID
		students    []uuid.UUID
		maxStudents int
	}{
		{0, "14:00:00", "16:00:00", teacher2.ID, []uuid.UUID{student1.ID, student2.ID, student3.ID}, 5}, // День 0 но в 14:00 (не конфликт с 09:00)
		{1, "09:00:00", "11:00:00", teacher2.ID, []uuid.UUID{student1.ID, student2.ID, student3.ID}, 5}, // День 1 (нет конфликта)
		{3, "09:00:00", "11:00:00", teacher2.ID, []uuid.UUID{student1.ID, student2.ID, student3.ID}, 5}, // День 3 (нет конфликта)
		{5, "09:00:00", "11:00:00", teacher2.ID, []uuid.UUID{student1.ID, student2.ID, student3.ID}, 5}, // День 5 (нет конфликта)
	})

	req2 := &models.ApplyTemplateRequest{
		TemplateID:    template2ID,
		WeekStartDate: "2026-02-09",
		DryRun:        false,
	}

	app2, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, req2)

	// ASSERTIONS для замены
	assert.NoError(t, err, "Replace apply should succeed")
	assert.NotNil(t, app2.CleanupStats, "Replace should have cleanup stats")
	assert.Equal(t, 6, app2.CleanupStats.CancelledBookings, "Should cancel 6 old bookings")
	assert.Equal(t, 6, app2.CleanupStats.RefundedCredits, "Should refund 6 credits")
	assert.Equal(t, 3, app2.CleanupStats.DeletedLessons, "Should delete 3 old lessons")

	assert.NotNil(t, app2.CreationStats, "Replace should have creation stats")
	assert.Equal(t, 4, app2.CreationStats.CreatedLessons, "Should create 4 new lessons")
	assert.Equal(t, 12, app2.CreationStats.CreatedBookings, "Should create 12 new bookings")
	assert.Equal(t, 12, app2.CreationStats.DeductedCredits, "Should deduct 12 credits")

	// Net credit change: 12 deducted - 6 refunded = -6 (net deduction)
	netChange := app2.CreationStats.DeductedCredits - app2.CleanupStats.RefundedCredits
	assert.Equal(t, 6, netChange, "Net credit change should be -6")

	// Проверяем что старые lessons помечены как deleted
	var deletedCount int
	db.QueryRow(`SELECT COUNT(*) FROM lessons WHERE start_time >= $1 AND start_time < $2 AND deleted_at IS NOT NULL`,
		weekDate, weekDate.AddDate(0, 0, 7)).Scan(&deletedCount)
	assert.Equal(t, 3, deletedCount, "Old 3 lessons should be soft-deleted")

	// Проверяем что новые lessons созданы
	var newLessonsCount int
	db.QueryRow(`SELECT COUNT(*) FROM lessons WHERE start_time >= $1 AND start_time < $2 AND deleted_at IS NULL`,
		weekDate, weekDate.AddDate(0, 0, 7)).Scan(&newLessonsCount)
	assert.Equal(t, 4, newLessonsCount, "New 4 lessons should be created")

	// Проверяем что старые bookings отменены
	var activeBokingsCount int
	db.QueryRow(`SELECT COUNT(*) FROM bookings WHERE status = 'active'`).Scan(&activeBokingsCount)
	assert.Equal(t, 12, activeBokingsCount, "Should have exactly 12 active bookings (new ones)")

	// Проверяем баланс кредитов для каждого студента
	// Template A: 3 lessons * 2 students (student1, student2) = 6 bookings total = 3 per student
	// Cleanup: Refund 3 to student1, 3 to student2
	// Template B: 4 lessons * 3 students (student1, student2, student3) = 12 bookings total = 4 per student
	// Net for student1&2: -3 (Template A) + 3 (refund) - 4 (Template B) = -4 net
	// Net for student3: -4 (Template B) = -4 net
	var balance1, balance2, balance3 int
	db.QueryRow(`SELECT balance FROM credits WHERE user_id = $1`, student1.ID).Scan(&balance1)
	db.QueryRow(`SELECT balance FROM credits WHERE user_id = $1`, student2.ID).Scan(&balance2)
	db.QueryRow(`SELECT balance FROM credits WHERE user_id = $1`, student3.ID).Scan(&balance3)

	assert.Equal(t, 26, balance1, "Student1 balance: 30 - 3 + 3 - 4 = 26")
	assert.Equal(t, 26, balance2, "Student2 balance: 30 - 3 + 3 - 4 = 26")
	assert.Equal(t, 26, balance3, "Student3 balance: 30 - 4 = 26 (wasn't in first template)")

	t.Log("✓ Test 1.2 PASSED: Replace Apply with Existing")
}

// TestTemplateReplacement_InsufficientCreditsAfterReplacement проверяет отказ при недостатке кредитов после замены
// Сценарий 2.2: Insufficient Credits After Replacement
func TestTemplateReplacement_InsufficientCreditsAfterReplacement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	// Setup
	admin := createTestUser(t, db, "admin-insuff@test.com", "Admin Insuff", string(models.RoleAdmin))
	teacher := createTestUser(t, db, "teacher-insuff@test.com", "Teacher Insuff", string(models.RoleMethodologist))
	student1 := createTestUser(t, db, "student-insuff1@test.com", "Student Insuff 1", string(models.RoleStudent))

	// Даем студенту ровно столько, чтобы хватило на первый шаблон
	// Первый: 2 lessons * 1 student = 2 credits (8 всего)
	// Второй: 5 lessons * 1 student = 5 credits требуется
	// После refund: 8 + 2 = 10
	// После requirement: 10 - 5 = 5 (достаточно)
	// Но если новых требуется 15: 10 - 15 = -5 (недостаточно)
	addCreditToStudent(t, db, student1.ID, 10)

	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	weekDate := time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC)

	// STEP 1: Применяем первый шаблон (2 занятия = 2 бронирования)
	template1ID := createTemplateWithLessons(t, db, admin.ID, "Template Small", []struct {
		dayOfWeek   int
		startTime   string
		endTime     string
		teacherID   uuid.UUID
		students    []uuid.UUID
		maxStudents int
	}{
		{0, "09:00:00", "11:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{2, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
	})

	req1 := &models.ApplyTemplateRequest{
		TemplateID:    template1ID,
		WeekStartDate: "2026-02-16",
		DryRun:        false,
	}

	app1, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, req1)
	require.NoError(t, err, "First apply should succeed")
	require.NotNil(t, app1, "app1 should not be nil")
	assert.Equal(t, 2, app1.CreationStats.CreatedBookings, "First: 2 bookings")

	// После первого: 10 - 2 = 8 credits

	// STEP 2: Пытаемся применить шаблон требующий 15 кредитов (когда будет refund 2)
	// После refund: 8 + 2 = 10
	// Требуется: 15 bookings = 15 credits
	// Итого: 10 - 15 = -5 (недостаточно!)
	template2ID := createTemplateWithLessons(t, db, admin.ID, "Template Large", []struct {
		dayOfWeek   int
		startTime   string
		endTime     string
		teacherID   uuid.UUID
		students    []uuid.UUID
		maxStudents int
	}{
		// 3 lessons на каждый день = 15 total bookings
		{0, "08:00:00", "10:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{0, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{0, "12:00:00", "14:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{1, "08:00:00", "10:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{1, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{1, "12:00:00", "14:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{2, "08:00:00", "10:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{2, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{2, "12:00:00", "14:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{3, "08:00:00", "10:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{3, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{3, "12:00:00", "14:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{4, "08:00:00", "10:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{4, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
		{4, "12:00:00", "14:00:00", teacher.ID, []uuid.UUID{student1.ID}, 4},
	})

	req2 := &models.ApplyTemplateRequest{
		TemplateID:    template2ID,
		WeekStartDate: "2026-02-16",
		DryRun:        false,
	}

	app2, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, req2)

	// ASSERTIONS
	assert.Error(t, err, "Apply should FAIL with insufficient credits error")
	assert.Nil(t, app2, "Application should not be created")
	assert.Contains(t, err.Error(), "недостаточно", "Error should mention insufficient credits (Russian)")

	// Проверяем что БД не изменилась (atomicity)
	var lessonCount, bookingCount int
	db.QueryRow(`SELECT COUNT(*) FROM lessons WHERE start_time >= $1 AND start_time < $2 AND deleted_at IS NULL`,
		weekDate, weekDate.AddDate(0, 0, 7)).Scan(&lessonCount)
	db.QueryRow(`SELECT COUNT(*) FROM bookings WHERE status = 'active'`).Scan(&bookingCount)

	assert.Equal(t, 2, lessonCount, "Should still have 2 old lessons (transaction rolled back)")
	assert.Equal(t, 2, bookingCount, "Should still have 2 old bookings (transaction rolled back)")

	// Проверяем баланс не изменился
	var balance int
	db.QueryRow(`SELECT balance FROM credits WHERE user_id = $1`, student1.ID).Scan(&balance)
	assert.Equal(t, 8, balance, "Balance should be unchanged (transaction rolled back)")

	t.Log("✓ Test 2.2 PASSED: Insufficient Credits After Replacement")
}

// TestTemplateReplacement_NetCreditCalculation проверяет корректность расчета кредитов
func TestTemplateReplacement_NetCreditCalculation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	// Setup
	admin := createTestUser(t, db, "admin-net@test.com", "Admin Net", string(models.RoleAdmin))
	teacher := createTestUser(t, db, "teacher-net@test.com", "Teacher Net", string(models.RoleMethodologist))
	student := createTestUser(t, db, "student-net@test.com", "Student Net", string(models.RoleStudent))

	// Даем студенту 20 кредитов
	addCreditToStudent(t, db, student.ID, 20)

	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// STEP 1: Применяем шаблон с 4 lessons = 4 bookings, cost = 4 credits
	// Balance: 20 - 4 = 16
	template1ID := createTemplateWithLessons(t, db, admin.ID, "Template 4", []struct {
		dayOfWeek   int
		startTime   string
		endTime     string
		teacherID   uuid.UUID
		students    []uuid.UUID
		maxStudents int
	}{
		{0, "09:00:00", "11:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
		{1, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
		{2, "11:00:00", "13:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
		{3, "12:00:00", "14:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
	})

	req1 := &models.ApplyTemplateRequest{
		TemplateID:    template1ID,
		WeekStartDate: "2026-02-23",
		DryRun:        false,
	}

	app1, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, req1)
	assert.NoError(t, err)
	assert.Equal(t, 4, app1.CreationStats.DeductedCredits)

	// STEP 2: Заменяем на 8 lessons = 8 bookings, cost = 8 credits
	// Refund: 4, Deduct: 8, Net: -4
	// Balance: 16 + 4 - 8 = 12
	// Важно: Используем неконфликтующие времена - все времена НЕ пересекаются
	template2ID := createTemplateWithLessons(t, db, admin.ID, "Template 8", []struct {
		dayOfWeek   int
		startTime   string
		endTime     string
		teacherID   uuid.UUID
		students    []uuid.UUID
		maxStudents int
	}{
		{0, "08:00:00", "10:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
		{0, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
		{1, "08:00:00", "10:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
		{1, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
		{2, "08:00:00", "10:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
		{2, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
		{3, "08:00:00", "10:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
		{4, "08:00:00", "10:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
	})

	req2 := &models.ApplyTemplateRequest{
		TemplateID:    template2ID,
		WeekStartDate: "2026-02-23",
		DryRun:        false,
	}

	app2, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, req2)
	assert.NoError(t, err)
	assert.NotNil(t, app2.CleanupStats)
	assert.Equal(t, 4, app2.CleanupStats.RefundedCredits)
	assert.Equal(t, 8, app2.CreationStats.DeductedCredits)

	// Net credit change
	netChange := app2.CreationStats.DeductedCredits - app2.CleanupStats.RefundedCredits
	assert.Equal(t, 4, netChange, "Net deduction should be 4 (8 - 4)")

	// Проверяем итоговый баланс
	var balance int
	db.QueryRow(`SELECT balance FROM credits WHERE user_id = $1`, student.ID).Scan(&balance)
	assert.Equal(t, 12, balance, "Balance should be 12 (20 - 4 - 4)")

	t.Log("✓ Test 3: Net Credit Calculation")
}

// TestTemplateReplacement_ApplicationStatusTracking проверяет отслеживание статуса applications
func TestTemplateReplacement_ApplicationStatusTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()

	// Setup
	admin := createTestUser(t, db, "admin-status@test.com", "Admin Status", string(models.RoleAdmin))
	teacher := createTestUser(t, db, "teacher-status@test.com", "Teacher Status", string(models.RoleMethodologist))
	student := createTestUser(t, db, "student-status@test.com", "Student Status", string(models.RoleStudent))

	addCreditToStudent(t, db, student.ID, 30)

	templateRepo := repository.NewLessonTemplateRepository(db)
	templateLessonRepo := repository.NewTemplateLessonRepository(db)
	templateAppRepo := repository.NewTemplateApplicationRepository(db)
	lessonRepo := repository.NewLessonRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	userRepo := repository.NewUserRepository(db)

	templateService := NewTemplateService(db, templateRepo, templateLessonRepo, templateAppRepo, lessonRepo, creditRepo, bookingRepo, userRepo)

	// Создаем два шаблона
	template1ID := createTemplateWithLessons(t, db, admin.ID, "Template S1", []struct {
		dayOfWeek   int
		startTime   string
		endTime     string
		teacherID   uuid.UUID
		students    []uuid.UUID
		maxStudents int
	}{
		{0, "09:00:00", "11:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
		{2, "10:00:00", "12:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
	})

	template2ID := createTemplateWithLessons(t, db, admin.ID, "Template S2", []struct {
		dayOfWeek   int
		startTime   string
		endTime     string
		teacherID   uuid.UUID
		students    []uuid.UUID
		maxStudents int
	}{
		{1, "09:00:00", "11:00:00", teacher.ID, []uuid.UUID{student.ID}, 4},
	})

	// STEP 1: Применяем первый шаблон
	req1 := &models.ApplyTemplateRequest{
		TemplateID:    template1ID,
		WeekStartDate: "2026-03-02",
		DryRun:        false,
	}

	app1, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, req1)
	assert.NoError(t, err)
	assert.Equal(t, "applied", app1.Status)

	// Проверяем что first application имеет статус 'applied'
	var app1Status string
	db.QueryRow(`SELECT status FROM template_applications WHERE id = $1`, app1.ID).Scan(&app1Status)
	assert.Equal(t, "applied", app1Status, "First application should have status 'applied'")

	// STEP 2: Применяем второй шаблон (replacement)
	req2 := &models.ApplyTemplateRequest{
		TemplateID:    template2ID,
		WeekStartDate: "2026-03-02",
		DryRun:        false,
	}

	app2, err := templateService.ApplyTemplateToWeek(ctx, admin.ID, req2)
	assert.NoError(t, err)
	assert.Equal(t, "applied", app2.Status)

	// Проверяем что old application теперь имеет статус 'replaced'
	db.QueryRow(`SELECT status FROM template_applications WHERE id = $1`, app1.ID).Scan(&app1Status)
	assert.Equal(t, "replaced", app1Status, "Old application should now have status 'replaced'")

	// Проверяем что old application id сохранен в новой application
	assert.NotNil(t, app2.CleanupStats)
	assert.Equal(t, app1.ID, app2.CleanupStats.ReplacedApplicationID, "New application should reference old one")

	t.Log("✓ Test 4.3: Application Status Tracking")
}

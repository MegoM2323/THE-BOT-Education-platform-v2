package handlers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/database"
	"tutoring-platform/internal/models"
	"tutoring-platform/pkg/hash"
)

// stringPtr возвращает указатель на копию строки
// Используется в тестах для инициализации полей *string
func stringPtr(s string) *string {
	return &s
}

// setupTestDB инициализирует тестовую БД и возвращает подключение.
// Использует общий пул соединений для избегания исчерпания лимита.
// Note: НЕ очищает таблицы здесь, т.к. это вызывает race conditions при параллельном запуске тестов.
// Каждый тест должен использовать уникальные данные (UUID в email) для изоляции.
func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	db := database.GetTestSqlxDB(t)

	return db
}

// cleanupTestDB no-op для совместимости.
// НЕ очищает таблицы, т.к. тесты используют уникальные данные и не конфликтуют.
// Примечание: не закрывает общий пул - управляется глобально.
func cleanupTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()
}

// cleanupAllTables no-op для совместимости.
// НЕ очищает таблицы, т.к. тесты используют уникальные данные.
func cleanupAllTables(t *testing.T, db *sqlx.DB) {
	t.Helper()
}

// createTestUser создает тестового пользователя с указанной ролью.
// Всегда генерирует уникальный email на основе UUID для избежания конфликтов.
func createTestUser(t *testing.T, db *sqlx.DB, email, fullName, role string) *models.User {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID := uuid.New()
	uniqueEmail := fmt.Sprintf("%s-%s", userID.String()[:8], email)

	passwordHash, err := hash.HashPassword("TestPass123!")
	require.NoError(t, err, "Failed to hash password")

	firstName := ""
	lastName := ""
	if fullName != "" {
		parts := strings.SplitN(strings.TrimSpace(fullName), " ", 2)
		if len(parts) > 0 {
			firstName = parts[0]
		}
		if len(parts) > 1 {
			lastName = parts[1]
		}
	}

	user := &models.User{
		ID:           userID,
		Email:        uniqueEmail,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         models.UserRole(role),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, payment_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		false,
		user.CreatedAt,
		user.UpdatedAt,
	)
	require.NoError(t, err, "Failed to create test user")

	return user
}

// addCreditToStudent добавляет кредиты студенту
func addCreditToStudent(t *testing.T, db *sqlx.DB, studentID uuid.UUID, amount int) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First try to get existing record
	var existingID string
	query := `SELECT id FROM credits WHERE user_id = $1`
	err := db.GetContext(ctx, &existingID, query, studentID)

	if err != nil {
		// Record doesn't exist, create new one
		createQuery := `
			INSERT INTO credits (id, user_id, balance, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err := db.ExecContext(ctx, createQuery,
			uuid.New(),
			studentID,
			amount,
			time.Now(),
			time.Now(),
		)
		require.NoError(t, err, "Failed to create credits for student")
	} else {
		// Record exists, update balance
		updateQuery := `UPDATE credits SET balance = balance + $1, updated_at = $2 WHERE user_id = $3`
		_, err := db.ExecContext(ctx, updateQuery, amount, time.Now(), studentID)
		require.NoError(t, err, "Failed to update credits for student")
	}
}

// createLesson создает тестовое занятие для учителя
func createLesson(t *testing.T, db *sqlx.DB, teacherID uuid.UUID, maxStudents int, startTime time.Time) uuid.UUID {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lessonID := uuid.New()
	endTime := startTime.Add(2 * time.Hour)
	color := "#3B82F6"

	query := `
		INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := db.ExecContext(ctx, query,
		lessonID,
		teacherID,
		startTime,
		endTime,
		maxStudents,
		0, // current_students
		color,
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err, "Failed to create test lesson")

	return lessonID
}

// createLessonWithTime создает тестовое занятие с указанными часом и минутами
func createLessonWithTime(t *testing.T, db *sqlx.DB, teacherID uuid.UUID, maxStudents int, date time.Time, hour, minute int) uuid.UUID {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lessonID := uuid.New()
	startTime := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, date.Location())
	endTime := startTime.Add(2 * time.Hour)
	color := "#3B82F6"

	query := `
		INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := db.ExecContext(ctx, query,
		lessonID,
		teacherID,
		startTime,
		endTime,
		maxStudents,
		0, // current_students
		color,
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err, "Failed to create test lesson with time")

	return lessonID
}

// createMatchingLessons создает несколько занятий в один и тот же день недели и время (для bulk edit тестов)
func createMatchingLessons(t *testing.T, db *sqlx.DB, teacherID uuid.UUID, count int) []uuid.UUID {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var lessonIDs []uuid.UUID
	baseTime := time.Now().AddDate(0, 0, 1) // Start tomorrow
	// Align to 9:00 AM
	baseTime = time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 9, 0, 0, 0, baseTime.Location())

	for i := 0; i < count; i++ {
		lessonID := uuid.New()
		startTime := baseTime.AddDate(0, 0, i*7) // Add 7 days for each lesson (same day of week)
		endTime := startTime.Add(2 * time.Hour)
		color := "#3B82F6"

		query := `
			INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, color, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`

		_, err := db.ExecContext(ctx, query,
			lessonID,
			teacherID,
			startTime,
			endTime,
			4, // max_students
			0, // current_students
			color,
			time.Now(),
			time.Now(),
		)
		require.NoError(t, err, "Failed to create matching lesson")

		lessonIDs = append(lessonIDs, lessonID)
	}

	return lessonIDs
}

// addStudentToLesson добавляет студента на занятие (создает booking)
func addStudentToLesson(t *testing.T, db *sqlx.DB, lessonID, studentID uuid.UUID) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bookingID := uuid.New()

	query := `
		INSERT INTO bookings (id, lesson_id, student_id, status, booked_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := db.ExecContext(ctx, query,
		bookingID,
		lessonID,
		studentID,
		"active", // status - must be 'active' or 'cancelled'
		time.Now(),
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err, "Failed to add student to lesson")

	// Update current_students count
	updateQuery := `
		UPDATE lessons SET current_students = current_students + 1 WHERE id = $1
	`
	_, err = db.ExecContext(ctx, updateQuery, lessonID)
	require.NoError(t, err, "Failed to update lesson student count")
}

var weekCounter int64 = 0
var weekMu sync.Mutex

// getNextMonday returns a unique Monday for each test call.
// Uses an incrementing counter to ensure no two tests use the same week.
func getNextMonday() time.Time {
	weekMu.Lock()
	weekCounter++
	offset := weekCounter
	weekMu.Unlock()

	now := time.Now()
	daysUntilMonday := (time.Monday - now.Weekday() + 7) % 7
	if daysUntilMonday == 0 && now.Hour() > 0 {
		daysUntilMonday = 7
	}
	baseMonday := now.AddDate(0, 0, int(daysUntilMonday))
	uniqueMonday := baseMonday.AddDate(0, 0, int(offset)*7)
	return time.Date(uniqueMonday.Year(), uniqueMonday.Month(), uniqueMonday.Day(), 0, 0, 0, 0, uniqueMonday.Location())
}

// getCreditBalance получает текущий баланс кредитов студента
func getCreditBalance(t *testing.T, db *sqlx.DB, studentID uuid.UUID) int {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var balance int
	query := `SELECT balance FROM credits WHERE user_id = $1`
	err := db.GetContext(ctx, &balance, query, studentID)
	require.NoError(t, err, "Failed to get credit balance")
	return balance
}

// REMOVED: createTestTemplate function - templates are no longer supported

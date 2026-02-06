package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"tutoring-platform/internal/database"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// setupTestDB returns the shared test database connection.
// Uses the global pool from database.GetTestSqlxDB to avoid connection exhaustion.
// Note: НЕ очищает таблицы здесь, т.к. это вызывает race conditions при параллельном запуске тестов.
// Каждый тест должен использовать уникальные данные (UUID в email) для изоляции.
func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	db := database.GetTestSqlxDB(t)

	return db
}

// cleanupTestDB no-op для совместимости.
// НЕ очищает таблицы, т.к. тесты используют уникальные данные и не конфликтуют.
// Note: Does NOT close the connection - it's managed by the shared pool.
func cleanupTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()
}

// cleanupAllTables no-op для совместимости.
// НЕ очищает таблицы, т.к. тесты используют уникальные данные.
func cleanupAllTables(t *testing.T, db *sqlx.DB) {
	t.Helper()
}

// createTestUserByID creates a test user in the database with a specific UUID.
// Uses unique email based on userID to prevent conflicts in parallel tests.
func createTestUserByID(t *testing.T, db *sqlx.DB, userID uuid.UUID, email string, role string) {
	t.Helper()

	ctx := context.Background()
	now := time.Now()

	uniqueEmail := fmt.Sprintf("%s-%s", userID.String()[:8], email)

	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := db.ExecContext(ctx, query, userID, uniqueEmail, "test_hash", "Test", "User", role, now, now)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
}

// createMultipleTestUsers creates multiple test users with unique UUIDs and emails.
// Email format: {uuid[:8]}-testuser{index}@test.com to prevent conflicts in parallel tests.
func createMultipleTestUsers(t *testing.T, db *sqlx.DB, count int, role string) []uuid.UUID {
	t.Helper()

	userIDs := make([]uuid.UUID, count)
	for i := 0; i < count; i++ {
		userID := uuid.New()
		userIDs[i] = userID
		email := fmt.Sprintf("testuser%d@test.com", i)
		createTestUserByID(t, db, userID, email, role)
	}
	return userIDs
}

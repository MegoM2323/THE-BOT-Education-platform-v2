package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"tutoring-platform/internal/database"
	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Test Setup Helpers ====================

// setupLessonBroadcastTestDB returns the shared test database connections.
// Uses database.GetTestSqlxDB and database.GetTestPool to avoid connection exhaustion.
// Note: НЕ очищает таблицы здесь, т.к. это вызывает race conditions при параллельном запуске тестов.
func setupLessonBroadcastTestDB(t *testing.T) (*sqlx.DB, *pgxpool.Pool) {
	t.Helper()

	pool := database.GetTestPool(t)
	sqlxDB := database.GetTestSqlxDB(t)

	return sqlxDB, pool
}

// cleanupLessonBroadcastTestDB no-op для совместимости.
// НЕ очищает таблицы, т.к. тесты используют уникальные данные и не конфликтуют.
func cleanupLessonBroadcastTestDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
}

// createTestUser создает тестового пользователя
func createTestUserForBroadcast(t *testing.T, pool *pgxpool.Pool, role string) uuid.UUID {
	t.Helper()

	ctx := context.Background()
	userID := uuid.New()

	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := pool.Exec(ctx, query,
		userID,
		"user_"+userID.String()+"@test.com",
		"hashed_password",
		"Test",
		"User",
		role,
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err)

	return userID
}

// createTestLesson создает тестовое занятие
func createTestLesson(t *testing.T, pool *pgxpool.Pool, teacherID uuid.UUID) uuid.UUID {
	t.Helper()

	ctx := context.Background()
	lessonID := uuid.New()

	query := `
		INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	_, err := pool.Exec(ctx, query,
		lessonID,
		teacherID,
		startTime,
		endTime,
		4,
		0,
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err)

	return lessonID
}

// ==================== Tests ====================

func TestLessonBroadcastRepo_CreateBroadcast(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	repo := NewLessonBroadcastRepository(sqlxDB)

	teacherID := createTestUserForBroadcast(t, pool, "teacher")
	lessonID := createTestLesson(t, pool, teacherID)

	broadcast := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Важное объявление для студентов",
	}

	result, err := repo.CreateBroadcast(context.Background(), broadcast)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.ID)
	assert.Equal(t, lessonID, result.LessonID)
	assert.Equal(t, teacherID, result.SenderID)
	assert.Equal(t, "Важное объявление для студентов", result.Message)
	assert.Equal(t, models.LessonBroadcastStatusPending, result.Status)
	assert.Equal(t, 0, result.SentCount)
	assert.Equal(t, 0, result.FailedCount)
	assert.False(t, result.CreatedAt.IsZero())
	assert.Nil(t, result.CompletedAt)
}

func TestLessonBroadcastRepo_GetBroadcast(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	repo := NewLessonBroadcastRepository(sqlxDB)

	teacherID := createTestUserForBroadcast(t, pool, "teacher")
	lessonID := createTestLesson(t, pool, teacherID)

	// Создаем рассылку
	broadcast := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Тестовое сообщение",
	}

	created, err := repo.CreateBroadcast(context.Background(), broadcast)
	require.NoError(t, err)

	// Получаем рассылку
	result, err := repo.GetBroadcast(context.Background(), created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, lessonID, result.LessonID)
	assert.Equal(t, teacherID, result.SenderID)
	assert.Equal(t, "Тестовое сообщение", result.Message)
	assert.Equal(t, models.LessonBroadcastStatusPending, result.Status)
	// Проверяем что sender_name возвращается корректно
	assert.Equal(t, "Test User", result.SenderName)
}

func TestLessonBroadcastRepo_GetBroadcast_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	repo := NewLessonBroadcastRepository(sqlxDB)

	result, err := repo.GetBroadcast(context.Background(), uuid.New())

	assert.ErrorIs(t, err, ErrLessonBroadcastNotFound)
	assert.Nil(t, result)
}

func TestLessonBroadcastRepo_ListBroadcastsByLesson(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	repo := NewLessonBroadcastRepository(sqlxDB)

	teacherID := createTestUserForBroadcast(t, pool, "teacher")
	lessonID := createTestLesson(t, pool, teacherID)

	// Создаем две рассылки для одного урока
	broadcast1 := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Сообщение 1",
	}
	_, err := repo.CreateBroadcast(context.Background(), broadcast1)
	require.NoError(t, err)

	broadcast2 := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Сообщение 2",
	}
	_, err = repo.CreateBroadcast(context.Background(), broadcast2)
	require.NoError(t, err)

	// Получаем список рассылок
	result, err := repo.ListBroadcastsByLesson(context.Background(), lessonID)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Сообщение 2", result[0].Message) // Newest first
	assert.Equal(t, "Сообщение 1", result[1].Message)
	// Проверяем что sender_name возвращается корректно
	assert.Equal(t, "Test User", result[0].SenderName)
	assert.Equal(t, "Test User", result[1].SenderName)
}

func TestLessonBroadcastRepo_ListBroadcastsByLesson_WithSenderName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	repo := NewLessonBroadcastRepository(sqlxDB)

	// Создаем двух разных пользователей (преподавателя и админа)
	ctx := context.Background()

	teacherID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, teacherID, "teacher@test.com", "hash", "Иван", "Иванов", "teacher", time.Now(), time.Now())
	require.NoError(t, err)

	adminID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, adminID, "admin@test.com", "hash", "Петр", "Петров", "admin", time.Now(), time.Now())
	require.NoError(t, err)

	lessonID := createTestLesson(t, pool, teacherID)

	// Создаем рассылку от преподавателя
	broadcast1 := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Сообщение от преподавателя",
	}
	_, err = repo.CreateBroadcast(context.Background(), broadcast1)
	require.NoError(t, err)

	// Создаем рассылку от админа
	broadcast2 := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: adminID,
		Message:  "Сообщение от администратора",
	}
	_, err = repo.CreateBroadcast(context.Background(), broadcast2)
	require.NoError(t, err)

	// Получаем список рассылок
	result, err := repo.ListBroadcastsByLesson(context.Background(), lessonID)

	require.NoError(t, err)
	assert.Len(t, result, 2)

	// Проверяем что sender_name корректно возвращается для обоих отправителей
	assert.Equal(t, "Сообщение от администратора", result[0].Message) // Newest first
	assert.Equal(t, "Петр Петров", result[0].SenderName)
	assert.Equal(t, adminID, result[0].SenderID)

	assert.Equal(t, "Сообщение от преподавателя", result[1].Message)
	assert.Equal(t, "Иван Иванов", result[1].SenderName)
	assert.Equal(t, teacherID, result[1].SenderID)
}

func TestLessonBroadcastRepo_UpdateBroadcastStatus_Completed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	repo := NewLessonBroadcastRepository(sqlxDB)

	teacherID := createTestUserForBroadcast(t, pool, "teacher")
	lessonID := createTestLesson(t, pool, teacherID)

	// Создаем рассылку
	broadcast := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Тестовое сообщение",
	}
	created, err := repo.CreateBroadcast(context.Background(), broadcast)
	require.NoError(t, err)

	// Обновляем статус на completed
	err = repo.UpdateBroadcastStatus(
		context.Background(),
		created.ID,
		models.LessonBroadcastStatusCompleted,
		10, // sent_count
		2,  // failed_count
	)
	require.NoError(t, err)

	// Проверяем обновление
	updated, err := repo.GetBroadcast(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, models.LessonBroadcastStatusCompleted, updated.Status)
	assert.Equal(t, 10, updated.SentCount)
	assert.Equal(t, 2, updated.FailedCount)
	assert.NotNil(t, updated.CompletedAt)
}

func TestLessonBroadcastRepo_UpdateBroadcastStatus_Sending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	repo := NewLessonBroadcastRepository(sqlxDB)

	teacherID := createTestUserForBroadcast(t, pool, "teacher")
	lessonID := createTestLesson(t, pool, teacherID)

	// Создаем рассылку
	broadcast := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Тестовое сообщение",
	}
	created, err := repo.CreateBroadcast(context.Background(), broadcast)
	require.NoError(t, err)

	// Обновляем статус на sending
	err = repo.UpdateBroadcastStatus(
		context.Background(),
		created.ID,
		models.LessonBroadcastStatusSending,
		5, // sent_count
		1, // failed_count
	)
	require.NoError(t, err)

	// Проверяем обновление
	updated, err := repo.GetBroadcast(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, models.LessonBroadcastStatusSending, updated.Status)
	assert.Equal(t, 5, updated.SentCount)
	assert.Equal(t, 1, updated.FailedCount)
	assert.Nil(t, updated.CompletedAt) // Не должно быть completed_at для промежуточного статуса
}

func TestLessonBroadcastRepo_AddBroadcastFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	repo := NewLessonBroadcastRepository(sqlxDB)

	teacherID := createTestUserForBroadcast(t, pool, "teacher")
	lessonID := createTestLesson(t, pool, teacherID)

	// Создаем рассылку
	broadcast := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Тестовое сообщение",
	}
	created, err := repo.CreateBroadcast(context.Background(), broadcast)
	require.NoError(t, err)

	// Добавляем файл
	file := &models.BroadcastFile{
		BroadcastID: created.ID,
		FileName:    "document.pdf",
		FilePath:    "/uploads/abc123.pdf",
		FileSize:    2048000,
		MimeType:    "application/pdf",
	}

	err = repo.AddBroadcastFile(context.Background(), file)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, file.ID)
	assert.False(t, file.UploadedAt.IsZero())
}

func TestLessonBroadcastRepo_AddBroadcastFile_TooManyFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	repo := NewLessonBroadcastRepository(sqlxDB)

	teacherID := createTestUserForBroadcast(t, pool, "teacher")
	lessonID := createTestLesson(t, pool, teacherID)

	// Создаем рассылку
	broadcast := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Тестовое сообщение",
	}
	created, err := repo.CreateBroadcast(context.Background(), broadcast)
	require.NoError(t, err)

	// Добавляем 10 файлов (максимум)
	for i := 1; i <= 10; i++ {
		file := &models.BroadcastFile{
			BroadcastID: created.ID,
			FileName:    fmt.Sprintf("file%d.pdf", i),
			FilePath:    fmt.Sprintf("/uploads/file%d.pdf", i),
			FileSize:    1024000,
			MimeType:    "application/pdf",
		}
		err = repo.AddBroadcastFile(context.Background(), file)
		require.NoError(t, err)
	}

	// Пытаемся добавить 11-й файл (должно вернуть ошибку)
	file11 := &models.BroadcastFile{
		BroadcastID: created.ID,
		FileName:    "file11.pdf",
		FilePath:    "/uploads/file11.pdf",
		FileSize:    1024000,
		MimeType:    "application/pdf",
	}

	err = repo.AddBroadcastFile(context.Background(), file11)

	assert.ErrorIs(t, err, models.ErrTooManyFiles)
}

func TestLessonBroadcastRepo_GetBroadcastFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	repo := NewLessonBroadcastRepository(sqlxDB)

	teacherID := createTestUserForBroadcast(t, pool, "teacher")
	lessonID := createTestLesson(t, pool, teacherID)

	// Создаем рассылку
	broadcast := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Тестовое сообщение",
	}
	created, err := repo.CreateBroadcast(context.Background(), broadcast)
	require.NoError(t, err)

	// Добавляем два файла
	file1 := &models.BroadcastFile{
		BroadcastID: created.ID,
		FileName:    "document1.pdf",
		FilePath:    "/uploads/1.pdf",
		FileSize:    2048000,
		MimeType:    "application/pdf",
	}
	err = repo.AddBroadcastFile(context.Background(), file1)
	require.NoError(t, err)

	file2 := &models.BroadcastFile{
		BroadcastID: created.ID,
		FileName:    "image.png",
		FilePath:    "/uploads/2.png",
		FileSize:    512000,
		MimeType:    "image/png",
	}
	err = repo.AddBroadcastFile(context.Background(), file2)
	require.NoError(t, err)

	// Получаем файлы
	result, err := repo.GetBroadcastFiles(context.Background(), created.ID)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "document1.pdf", result[0].FileName)
	assert.Equal(t, "image.png", result[1].FileName)
}

// TestLessonDelete_SoftDeleteDoesNotRemoveBroadcasts проверяет что soft delete lesson'а
// НЕ удаляет broadcast'ы (это логическое удаление, данные остаются в БД)
func TestLessonDelete_SoftDeleteDoesNotRemoveBroadcasts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	ctx := context.Background()
	lessonRepo := NewLessonRepository(sqlxDB)
	broadcastRepo := NewLessonBroadcastRepository(sqlxDB)

	// Создаем учителя
	teacherID := createTestUserForBroadcast(t, pool, "teacher")

	// Создаем урок
	lessonID := createTestLesson(t, pool, teacherID)

	// Создаем 3 broadcast'а для этого урока
	broadcasts := make([]*models.LessonBroadcast, 3)
	for i := 0; i < 3; i++ {
		broadcast := &models.LessonBroadcast{
			LessonID: lessonID,
			SenderID: teacherID,
			Message:  fmt.Sprintf("Тестовое сообщение %d", i+1),
		}
		created, err := broadcastRepo.CreateBroadcast(ctx, broadcast)
		require.NoError(t, err)
		broadcasts[i] = created
	}

	// Проверяем что broadcast'ы созданы
	createdBroadcasts, err := broadcastRepo.ListBroadcastsByLesson(ctx, lessonID)
	require.NoError(t, err)
	assert.Len(t, createdBroadcasts, 3)

	// Soft delete урока - это UPDATE deleted_at, не DELETE
	err = lessonRepo.Delete(ctx, lessonID)
	require.NoError(t, err)

	// При soft delete broadcast'ы НЕ удаляются каскадно (ON DELETE CASCADE работает только на DELETE)
	// Broadcasts остаются в БД для истории и архива
	stillExistingBroadcasts, err := broadcastRepo.ListBroadcastsByLesson(ctx, lessonID)
	require.NoError(t, err)
	assert.Len(t, stillExistingBroadcasts, 3, "Broadcasts should remain after soft delete (no cascade for UPDATE)")
}

// TestLessonHardDelete_CascadeDeletesBroadcasts проверяет что hard delete lesson'а
// также каскадно удаляет все broadcast'ы
func TestLessonHardDelete_CascadeDeletesBroadcasts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	sqlxDB, pool := setupLessonBroadcastTestDB(t)
	defer cleanupLessonBroadcastTestDB(t, pool)

	ctx := context.Background()
	lessonRepo := NewLessonRepository(sqlxDB)
	broadcastRepo := NewLessonBroadcastRepository(sqlxDB)

	// Создаем учителя
	teacherID := createTestUserForBroadcast(t, pool, "teacher")

	// Создаем урок
	lessonID := createTestLesson(t, pool, teacherID)

	// Создаем broadcast с файлами
	broadcast := &models.LessonBroadcast{
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Broadcast с файлами",
	}
	created, err := broadcastRepo.CreateBroadcast(ctx, broadcast)
	require.NoError(t, err)

	// Добавляем файл к broadcast'у
	file := &models.BroadcastFile{
		BroadcastID: created.ID,
		FileName:    "test.pdf",
		FilePath:    "/uploads/test.pdf",
		FileSize:    1024000,
		MimeType:    "application/pdf",
	}
	err = broadcastRepo.AddBroadcastFile(ctx, file)
	require.NoError(t, err)

	// Проверяем что файл добавлен
	files, err := broadcastRepo.GetBroadcastFiles(ctx, created.ID)
	require.NoError(t, err)
	assert.Len(t, files, 1)

	// Soft delete урока сначала (чтобы можно было hard delete)
	err = lessonRepo.Delete(ctx, lessonID)
	require.NoError(t, err)

	// Hard delete урока
	err = lessonRepo.HardDelete(ctx, lessonID, true)
	require.NoError(t, err)

	// Проверяем что broadcast полностью удален (вместе с файлами)
	// Это подтверждает что ON DELETE CASCADE работает правильно
	deletedBroadcasts, err := broadcastRepo.ListBroadcastsByLesson(ctx, lessonID)
	require.NoError(t, err)
	assert.Len(t, deletedBroadcasts, 0, "Broadcasts should be hard-deleted when lesson is hard-deleted")
}

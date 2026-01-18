package repository

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/database"
	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupHomeworkTestDB returns a HomeworkRepository using the shared test database pool.
// Uses database.GetTestSqlxDB to avoid connection exhaustion.
// Note: НЕ очищает таблицы здесь, т.к. это вызывает race conditions при параллельном запуске тестов.
func setupHomeworkTestDB(t *testing.T) *HomeworkRepository {
	t.Helper()

	db := database.GetTestSqlxDB(t)

	return NewHomeworkRepository(db)
}

// cleanupHomeworkTestDB no-op для совместимости.
// НЕ очищает таблицы, т.к. тесты используют уникальные данные и не конфликтуют.
func cleanupHomeworkTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()
}

// createHomeworkTestUser создает тестового пользователя с уникальным email
func createHomeworkTestUser(t *testing.T, db *sqlx.DB, role string) uuid.UUID {
	t.Helper()

	userID := uuid.New()
	email := "test-" + userID.String() + "@example.com" // Уникальный email
	query := `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`

	_, err := db.Exec(query, userID, email, "hash", "Test User", role)
	require.NoError(t, err, "Failed to create test user")

	return userID
}

// createHomeworkTestLesson создает тестовый урок
func createHomeworkTestLesson(t *testing.T, db *sqlx.DB, teacherID uuid.UUID) uuid.UUID {
	t.Helper()

	lessonID := uuid.New()
	query := `
		INSERT INTO lessons (id, teacher_id, start_time, end_time, max_students, current_students, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`

	startTime := time.Now().Add(48 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	_, err := db.Exec(query, lessonID, teacherID, startTime, endTime, 4, 0)
	require.NoError(t, err, "Failed to create test lesson")

	return lessonID
}

func TestHomeworkRepository_CreateHomework(t *testing.T) {
	repo := setupHomeworkTestDB(t)
	ctx := context.Background()

	// Создаем тестового пользователя (teacher)
	teacherID := createHomeworkTestUser(t, repo.db, "teacher")

	// Создаем тестовый урок
	lessonID := createHomeworkTestLesson(t, repo.db, teacherID)

	t.Run("Successful creation", func(t *testing.T) {
		homework := &models.LessonHomework{
			LessonID:  lessonID,
			FileName:  "homework.pdf",
			FilePath:  "/uploads/homework.pdf",
			FileSize:  1024,
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		}

		result, err := repo.CreateHomework(ctx, homework)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, homework.LessonID, result.LessonID)
		assert.Equal(t, homework.FileName, result.FileName)
		assert.Equal(t, homework.FilePath, result.FilePath)
		assert.Equal(t, homework.FileSize, result.FileSize)
		assert.Equal(t, homework.MimeType, result.MimeType)
		assert.Equal(t, homework.CreatedBy, result.CreatedBy)
		assert.False(t, result.CreatedAt.IsZero())
	})

	t.Run("Invalid lesson ID", func(t *testing.T) {
		homework := &models.LessonHomework{
			LessonID:  uuid.Nil,
			FileName:  "homework.pdf",
			FilePath:  "/uploads/homework.pdf",
			FileSize:  1024,
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		}

		_, err := repo.CreateHomework(ctx, homework)
		assert.Error(t, err)
		assert.Equal(t, models.ErrInvalidLessonID, err)
	})

	t.Run("Empty file name", func(t *testing.T) {
		homework := &models.LessonHomework{
			LessonID:  lessonID,
			FileName:  "",
			FilePath:  "/uploads/homework.pdf",
			FileSize:  1024,
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		}

		_, err := repo.CreateHomework(ctx, homework)
		assert.Error(t, err)
		assert.Equal(t, models.ErrInvalidFileName, err)
	})

	t.Run("File size too large", func(t *testing.T) {
		homework := &models.LessonHomework{
			LessonID:  lessonID,
			FileName:  "homework.pdf",
			FilePath:  "/uploads/homework.pdf",
			FileSize:  10485761, // 10MB + 1 byte
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		}

		_, err := repo.CreateHomework(ctx, homework)
		assert.Error(t, err)
		assert.Equal(t, models.ErrInvalidFileSize, err)
	})

	t.Run("Invalid MIME type", func(t *testing.T) {
		homework := &models.LessonHomework{
			LessonID:  lessonID,
			FileName:  "homework.exe",
			FilePath:  "/uploads/homework.exe",
			FileSize:  1024,
			MimeType:  "application/x-msdownload",
			CreatedBy: teacherID,
		}

		_, err := repo.CreateHomework(ctx, homework)
		assert.Error(t, err)
		assert.Equal(t, models.ErrMimeTypeNotAllowed, err)
	})

	t.Run("File name too long", func(t *testing.T) {
		longName := ""
		for i := 0; i < 260; i++ {
			longName += "a"
		}

		homework := &models.LessonHomework{
			LessonID:  lessonID,
			FileName:  longName,
			FilePath:  "/uploads/homework.pdf",
			FileSize:  1024,
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		}

		_, err := repo.CreateHomework(ctx, homework)
		assert.Error(t, err)
		assert.Equal(t, models.ErrFileNameTooLong, err)
	})

	t.Run("Invalid created_by", func(t *testing.T) {
		homework := &models.LessonHomework{
			LessonID:  lessonID,
			FileName:  "homework.pdf",
			FilePath:  "/uploads/homework.pdf",
			FileSize:  1024,
			MimeType:  "application/pdf",
			CreatedBy: uuid.Nil,
		}

		_, err := repo.CreateHomework(ctx, homework)
		assert.Error(t, err)
		assert.Equal(t, models.ErrInvalidUserID, err)
	})
}

func TestHomeworkRepository_GetHomeworkByLesson(t *testing.T) {
	repo := setupHomeworkTestDB(t)
	ctx := context.Background()

	teacherID := createHomeworkTestUser(t, repo.db, "teacher")
	lessonID := createHomeworkTestLesson(t, repo.db, teacherID)

	t.Run("Get all homework for lesson", func(t *testing.T) {
		// Создаем несколько домашних заданий
		homework1 := &models.LessonHomework{
			LessonID:  lessonID,
			FileName:  "homework1.pdf",
			FilePath:  "/uploads/homework1.pdf",
			FileSize:  1024,
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		}
		_, err := repo.CreateHomework(ctx, homework1)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond) // Небольшая задержка для правильной сортировки

		homework2 := &models.LessonHomework{
			LessonID:  lessonID,
			FileName:  "homework2.pdf",
			FilePath:  "/uploads/homework2.pdf",
			FileSize:  2048,
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		}
		_, err = repo.CreateHomework(ctx, homework2)
		require.NoError(t, err)

		// Получаем все домашние задания
		homeworks, err := repo.GetHomeworkByLesson(ctx, lessonID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(homeworks), 2)

		// Проверяем сортировку по created_at DESC (новые первыми)
		if len(homeworks) >= 2 {
			assert.True(t, homeworks[0].CreatedAt.After(homeworks[1].CreatedAt) ||
				homeworks[0].CreatedAt.Equal(homeworks[1].CreatedAt))
		}
	})

	t.Run("Get homework for lesson with no homework", func(t *testing.T) {
		emptyLessonID := uuid.New()
		homeworks, err := repo.GetHomeworkByLesson(ctx, emptyLessonID)
		require.NoError(t, err)
		assert.NotNil(t, homeworks)
		assert.Equal(t, 0, len(homeworks))
	})
}

func TestHomeworkRepository_GetHomeworkByID(t *testing.T) {
	repo := setupHomeworkTestDB(t)
	ctx := context.Background()

	teacherID := createHomeworkTestUser(t, repo.db, "teacher")
	lessonID := createHomeworkTestLesson(t, repo.db, teacherID)

	t.Run("Get existing homework", func(t *testing.T) {
		homework := &models.LessonHomework{
			LessonID:  lessonID,
			FileName:  "homework.pdf",
			FilePath:  "/uploads/homework.pdf",
			FileSize:  1024,
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		}

		created, err := repo.CreateHomework(ctx, homework)
		require.NoError(t, err)

		// Получаем по ID
		result, err := repo.GetHomeworkByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, created.LessonID, result.LessonID)
		assert.Equal(t, created.FileName, result.FileName)
		assert.Equal(t, created.FileSize, result.FileSize)
		assert.Equal(t, created.MimeType, result.MimeType)
	})

	t.Run("Get non-existing homework", func(t *testing.T) {
		nonExistingID := uuid.New()
		_, err := repo.GetHomeworkByID(ctx, nonExistingID)
		assert.Error(t, err)
		assert.Equal(t, ErrHomeworkNotFound, err)
	})
}

func TestHomeworkRepository_DeleteHomework(t *testing.T) {
	repo := setupHomeworkTestDB(t)
	ctx := context.Background()

	teacherID := createHomeworkTestUser(t, repo.db, "teacher")
	lessonID := createHomeworkTestLesson(t, repo.db, teacherID)

	t.Run("Delete existing homework", func(t *testing.T) {
		homework := &models.LessonHomework{
			LessonID:  lessonID,
			FileName:  "homework.pdf",
			FilePath:  "/uploads/homework.pdf",
			FileSize:  1024,
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		}

		created, err := repo.CreateHomework(ctx, homework)
		require.NoError(t, err)

		// Удаляем
		err = repo.DeleteHomework(ctx, created.ID)
		require.NoError(t, err)

		// Проверяем что удалено
		_, err = repo.GetHomeworkByID(ctx, created.ID)
		assert.Error(t, err)
		assert.Equal(t, ErrHomeworkNotFound, err)
	})

	t.Run("Delete non-existing homework", func(t *testing.T) {
		nonExistingID := uuid.New()
		err := repo.DeleteHomework(ctx, nonExistingID)
		assert.Error(t, err)
		assert.Equal(t, ErrHomeworkNotFound, err)
	})
}

func TestHomeworkRepository_DeleteAllByLesson(t *testing.T) {
	repo := setupHomeworkTestDB(t)
	ctx := context.Background()

	teacherID := createHomeworkTestUser(t, repo.db, "teacher")
	lessonID := createHomeworkTestLesson(t, repo.db, teacherID)

	t.Run("Delete all homework for lesson", func(t *testing.T) {
		// Создаем несколько домашних заданий
		for i := 0; i < 3; i++ {
			homework := &models.LessonHomework{
				LessonID:  lessonID,
				FileName:  "homework.pdf",
				FilePath:  "/uploads/homework.pdf",
				FileSize:  1024,
				MimeType:  "application/pdf",
				CreatedBy: teacherID,
			}
			_, err := repo.CreateHomework(ctx, homework)
			require.NoError(t, err)
		}

		// Проверяем что созданы
		homeworks, err := repo.GetHomeworkByLesson(ctx, lessonID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(homeworks), 3)

		// Удаляем все
		err = repo.DeleteAllByLesson(ctx, lessonID)
		require.NoError(t, err)

		// Проверяем что удалены
		homeworks, err = repo.GetHomeworkByLesson(ctx, lessonID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(homeworks))
	})

	t.Run("Delete for lesson with no homework", func(t *testing.T) {
		emptyLessonID := uuid.New()
		err := repo.DeleteAllByLesson(ctx, emptyLessonID)
		require.NoError(t, err) // Не должно быть ошибки
	})
}

func TestHomeworkRepository_GetHomeworkCount(t *testing.T) {
	repo := setupHomeworkTestDB(t)
	ctx := context.Background()

	teacherID := createHomeworkTestUser(t, repo.db, "teacher")
	lessonID := createHomeworkTestLesson(t, repo.db, teacherID)

	t.Run("Count homework", func(t *testing.T) {
		// Изначально 0
		count, err := repo.GetHomeworkCount(ctx, lessonID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Создаем несколько домашних заданий
		for i := 0; i < 5; i++ {
			homework := &models.LessonHomework{
				LessonID:  lessonID,
				FileName:  "homework.pdf",
				FilePath:  "/uploads/homework.pdf",
				FileSize:  1024,
				MimeType:  "application/pdf",
				CreatedBy: teacherID,
			}
			_, err := repo.CreateHomework(ctx, homework)
			require.NoError(t, err)
		}

		// Проверяем количество
		count, err = repo.GetHomeworkCount(ctx, lessonID)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
	})
}

func TestHomeworkRepository_GetTotalFileSizeByLesson(t *testing.T) {
	repo := setupHomeworkTestDB(t)
	ctx := context.Background()

	teacherID := createHomeworkTestUser(t, repo.db, "teacher")
	lessonID := createHomeworkTestLesson(t, repo.db, teacherID)

	t.Run("Calculate total file size", func(t *testing.T) {
		// Изначально 0
		totalSize, err := repo.GetTotalFileSizeByLesson(ctx, lessonID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), totalSize)

		// Создаем домашние задания с разными размерами
		sizes := []int64{1024, 2048, 4096, 8192}
		expectedTotal := int64(0)

		for _, size := range sizes {
			homework := &models.LessonHomework{
				LessonID:  lessonID,
				FileName:  "homework.pdf",
				FilePath:  "/uploads/homework.pdf",
				FileSize:  size,
				MimeType:  "application/pdf",
				CreatedBy: teacherID,
			}
			_, err := repo.CreateHomework(ctx, homework)
			require.NoError(t, err)
			expectedTotal += size
		}

		// Проверяем суммарный размер
		totalSize, err = repo.GetTotalFileSizeByLesson(ctx, lessonID)
		require.NoError(t, err)
		assert.Equal(t, expectedTotal, totalSize)
	})
}

func TestHomeworkModel_AllowedMimeTypes(t *testing.T) {
	repo := setupHomeworkTestDB(t)
	ctx := context.Background()

	teacherID := createHomeworkTestUser(t, repo.db, "teacher")
	lessonID := createHomeworkTestLesson(t, repo.db, teacherID)

	allowedMimeTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"image/png",
		"image/jpeg",
		"video/mp4",
		"audio/mpeg",
		"text/plain",
		"application/zip",
	}

	for _, mimeType := range allowedMimeTypes {
		t.Run("Allowed MIME type: "+mimeType, func(t *testing.T) {
			homework := &models.LessonHomework{
				LessonID:  lessonID,
				FileName:  "file.ext",
				FilePath:  "/uploads/file.ext",
				FileSize:  1024,
				MimeType:  mimeType,
				CreatedBy: teacherID,
			}

			_, err := repo.CreateHomework(ctx, homework)
			assert.NoError(t, err)
		})
	}

	t.Run("Disallowed MIME types", func(t *testing.T) {
		disallowedMimeTypes := []string{
			"application/x-msdownload", // .exe
			"application/x-sh",         // shell script
			"text/javascript",
			"application/x-httpd-php",
		}

		for _, mimeType := range disallowedMimeTypes {
			homework := &models.LessonHomework{
				LessonID:  lessonID,
				FileName:  "file.ext",
				FilePath:  "/uploads/file.ext",
				FileSize:  1024,
				MimeType:  mimeType,
				CreatedBy: teacherID,
			}

			_, err := repo.CreateHomework(ctx, homework)
			assert.Error(t, err)
			assert.Equal(t, models.ErrMimeTypeNotAllowed, err)
		}
	})
}

func TestHomeworkModel_HelperMethods(t *testing.T) {
	t.Run("IsImage", func(t *testing.T) {
		homework := &models.LessonHomework{
			MimeType: "image/png",
		}
		assert.True(t, homework.IsImage())

		homework.MimeType = "application/pdf"
		assert.False(t, homework.IsImage())
	})

	t.Run("IsVideo", func(t *testing.T) {
		homework := &models.LessonHomework{
			MimeType: "video/mp4",
		}
		assert.True(t, homework.IsVideo())

		homework.MimeType = "image/png"
		assert.False(t, homework.IsVideo())
	})

	t.Run("IsDocument", func(t *testing.T) {
		homework := &models.LessonHomework{
			MimeType: "application/pdf",
		}
		assert.True(t, homework.IsDocument())

		homework.MimeType = "video/mp4"
		assert.False(t, homework.IsDocument())
	})

	t.Run("GetFileExtension", func(t *testing.T) {
		tests := []struct {
			mimeType string
			expected string
		}{
			{"application/pdf", ".pdf"},
			{"image/png", ".png"},
			{"video/mp4", ".mp4"},
			{"audio/mpeg", ".mp3"},
			{"text/plain", ".txt"},
			{"unknown/type", ""},
		}

		for _, tt := range tests {
			homework := &models.LessonHomework{
				MimeType: tt.mimeType,
			}
			assert.Equal(t, tt.expected, homework.GetFileExtension())
		}
	})
}

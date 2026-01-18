package repository

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestLessonColor_CreateAndRetrieve проверяет что color сохраняется и возвращается корректно
func TestLessonColor_CreateAndRetrieve(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	lessonRepo := NewLessonRepository(db)

	// Создаём тестового учителя
	teacher := createTestUserVisibility(t, db, ctx, "teacher@test.com", models.RoleTeacher)

	tests := []struct {
		name          string
		color         string
		expectedColor string
	}{
		{
			name:          "Create lesson with custom color #FF5733",
			color:         "#FF5733",
			expectedColor: "#FF5733",
		},
		{
			name:          "Create lesson with default color #3B82F6",
			color:         "#3B82F6",
			expectedColor: "#3B82F6",
		},
		{
			name:          "Create lesson with green color #00FF00",
			color:         "#00FF00",
			expectedColor: "#00FF00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Создаём занятие с указанным цветом
			lesson := &models.Lesson{
				ID:              uuid.New(),
				TeacherID:       teacher.ID,
				StartTime:       time.Now().Add(24 * time.Hour),
				EndTime:         time.Now().Add(26 * time.Hour),
				MaxStudents:     4,
				CurrentStudents: 0,
				Color:           tt.color,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			err := lessonRepo.Create(ctx, lesson)
			require.NoError(t, err, "Failed to create lesson")

			// 2. Получаем занятие через GetByID
			retrieved, err := lessonRepo.GetByID(ctx, lesson.ID)
			require.NoError(t, err, "Failed to retrieve lesson by ID")
			require.Equal(t, tt.expectedColor, retrieved.Color, "Color should match")

			// 3. Получаем занятие через GetWithTeacher
			withTeacher, err := lessonRepo.GetWithTeacher(ctx, lesson.ID)
			require.NoError(t, err, "Failed to retrieve lesson with teacher")
			require.Equal(t, tt.expectedColor, withTeacher.Color, "Color should match in GetWithTeacher")

			// 4. Проверяем что занятие появляется в List
			lessons, err := lessonRepo.List(ctx, nil)
			require.NoError(t, err, "Failed to list lessons")

			found := false
			for _, l := range lessons {
				if l.ID == lesson.ID {
					found = true
					require.Equal(t, tt.expectedColor, l.Color, "Color should match in List")
				}
			}
			require.True(t, found, "Lesson should be present in list")
		})
	}
}

// TestLessonColor_Update проверяет что можно обновить цвет занятия
func TestLessonColor_Update(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	ctx := context.Background()
	lessonRepo := NewLessonRepository(db)

	// Создаём тестового учителя
	teacher := createTestUserVisibility(t, db, ctx, "teacher2@test.com", models.RoleTeacher)

	// Создаём занятие с дефолтным цветом
	lesson := &models.Lesson{
		ID:              uuid.New(),
		TeacherID:       teacher.ID,
		StartTime:       time.Now().Add(24 * time.Hour),
		EndTime:         time.Now().Add(26 * time.Hour),
		MaxStudents:     1,
		CurrentStudents: 0,
		Color:           "#3B82F6",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := lessonRepo.Create(ctx, lesson)
	require.NoError(t, err, "Failed to create lesson")

	// Обновляем цвет
	newColor := "#FF5733"
	err = lessonRepo.Update(ctx, lesson.ID, map[string]interface{}{
		"color": newColor,
	})
	require.NoError(t, err, "Failed to update lesson color")

	// Проверяем что цвет обновился
	updated, err := lessonRepo.GetByID(ctx, lesson.ID)
	require.NoError(t, err, "Failed to retrieve updated lesson")
	require.Equal(t, newColor, updated.Color, "Color should be updated")
}

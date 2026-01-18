package repository

import (
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestMarkApplicationAsReplacedTx проверяет корректность обновления статуса application
func TestMarkApplicationAsReplacedTx(t *testing.T) {
	// Этот тест требует реального подключения к БД для работы с транзакциями
	// В production среде используйте test database или mock
	t.Skip("Skipping integration test - requires database connection")

	// Пример использования метода:
	// ctx := context.Background()
	// tx, err := db.Beginx()
	// require.NoError(t, err)
	// defer tx.Rollback()
	//
	// repo := NewTemplateApplicationRepository(db)
	// err = repo.MarkApplicationAsReplacedTx(ctx, tx, applicationID)
	// assert.NoError(t, err)
	//
	// tx.Commit()
}

// TestGetApplicationsByWeekDateTx проверяет получение всех applications для недели
func TestGetApplicationsByWeekDateTx(t *testing.T) {
	t.Skip("Skipping integration test - requires database connection")

	// Пример использования:
	// weekDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC) // Monday
	// tx, _ := db.Beginx()
	// defer tx.Rollback()
	//
	// repo := NewTemplateApplicationRepository(db)
	// apps, err := repo.GetApplicationsByWeekDateTx(ctx, tx, weekDate)
	// assert.NoError(t, err)
	// assert.NotNil(t, apps)
}

// TestGetLessonStatsByWeekTx проверяет корректность подсчёта статистики
func TestGetLessonStatsByWeekTx(t *testing.T) {
	t.Skip("Skipping integration test - requires database connection")

	// Пример использования:
	// weekDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)
	// tx, _ := db.Beginx()
	// defer tx.Rollback()
	//
	// repo := NewTemplateApplicationRepository(db)
	// stats, err := repo.GetLessonStatsByWeekTx(ctx, tx, weekDate)
	// assert.NoError(t, err)
	// assert.GreaterOrEqual(t, stats.TotalLessons, 0)
	// assert.GreaterOrEqual(t, stats.TotalBookings, 0)
}

// TestWeekLessonStatsStructure проверяет структуру WeekLessonStats
func TestWeekLessonStatsStructure(t *testing.T) {
	stats := &WeekLessonStats{
		TotalLessons:     5,
		TotalBookings:    10,
		TotalCredits:     10,
		AffectedStudents: 3,
	}

	assert.Equal(t, 5, stats.TotalLessons)
	assert.Equal(t, 10, stats.TotalBookings)
	assert.Equal(t, 10, stats.TotalCredits)
	assert.Equal(t, 3, stats.AffectedStudents)
}

// TestTemplateApplicationWithStats проверяет модель TemplateApplication с новыми полями
func TestTemplateApplicationWithStats(t *testing.T) {
	app := &models.TemplateApplication{
		ID:            uuid.New(),
		TemplateID:    uuid.New(),
		AppliedByID:   uuid.New(),
		WeekStartDate: time.Now(),
		AppliedAt:     time.Now(),
		Status:        "replaced",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		CleanupStats: &models.CleanupStats{
			CancelledBookings:     5,
			RefundedCredits:       5,
			DeletedLessons:        3,
			ReplacedApplicationID: uuid.New(),
		},
		CreationStats: &models.CreationStats{
			CreatedLessons:  4,
			CreatedBookings: 6,
			DeductedCredits: 6,
		},
	}

	assert.Equal(t, "replaced", app.Status)
	assert.NotNil(t, app.CleanupStats)
	assert.NotNil(t, app.CreationStats)
	assert.Equal(t, 5, app.CleanupStats.CancelledBookings)
	assert.Equal(t, 4, app.CreationStats.CreatedLessons)
}

// BenchmarkGetLessonStatsByWeek бенчмарк для проверки производительности запроса статистики
func BenchmarkGetLessonStatsByWeek(b *testing.B) {
	b.Skip("Skipping benchmark - requires database connection")

	// Пример использования для performance тестирования:
	// ctx := context.Background()
	// weekDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)
	//
	// b.ResetTimer()
	// for i := 0; i < b.N; i++ {
	//     tx, _ := db.Beginx()
	//     repo := NewTemplateApplicationRepository(db)
	//     _, err := repo.GetLessonStatsByWeekTx(ctx, tx, weekDate)
	//     if err != nil {
	//         b.Fatal(err)
	//     }
	//     tx.Rollback()
	// }
}

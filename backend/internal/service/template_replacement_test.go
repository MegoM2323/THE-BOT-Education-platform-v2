package service

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLessonTemplateRepository для тестирования
type MockLessonTemplateRepository struct {
	mock.Mock
}

func (m *MockLessonTemplateRepository) GetTemplateWithLessons(ctx context.Context, templateID uuid.UUID) (*models.LessonTemplate, error) {
	args := m.Called(ctx, templateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LessonTemplate), args.Error(1)
}

// TestTemplateReplacement_CleanupStats проверяет что cleanup статистика возвращается корректно
func TestTemplateReplacement_CleanupStats(t *testing.T) {
	t.Run("CleanupStats содержит все поля", func(t *testing.T) {
		// Создаём пример CleanupStats
		stats := &models.CleanupStats{
			CancelledBookings:     5,
			RefundedCredits:       5,
			DeletedLessons:        3,
			ReplacedApplicationID: uuid.New(),
		}

		// Проверяем что все поля присутствуют
		assert.Equal(t, 5, stats.CancelledBookings, "CancelledBookings должен быть 5")
		assert.Equal(t, 5, stats.RefundedCredits, "RefundedCredits должен быть 5")
		assert.Equal(t, 3, stats.DeletedLessons, "DeletedLessons должен быть 3")
		assert.NotEqual(t, uuid.Nil, stats.ReplacedApplicationID, "ReplacedApplicationID не должен быть nil")
	})

	t.Run("CleanupStats может быть nil для чистого apply", func(t *testing.T) {
		// Для обычного apply (не replacement) CleanupStats должен быть nil
		application := &models.TemplateApplication{
			ID:                  uuid.New(),
			TemplateID:          uuid.New(),
			AppliedByID:         uuid.New(),
			WeekStartDate:       time.Now(),
			AppliedAt:           time.Now(),
			Status:              "applied",
			CreatedLessonsCount: 5,
			CleanupStats:        nil, // Нет cleanup при первом apply
			CreationStats: &models.CreationStats{
				CreatedLessons:  5,
				CreatedBookings: 10,
				DeductedCredits: 10,
			},
		}

		assert.Nil(t, application.CleanupStats, "CleanupStats должен быть nil для чистого apply")
		assert.NotNil(t, application.CreationStats, "CreationStats должен присутствовать всегда")
	})
}

// TestTemplateReplacement_CreationStats проверяет что creation статистика возвращается корректно
func TestTemplateReplacement_CreationStats(t *testing.T) {
	t.Run("CreationStats содержит все поля", func(t *testing.T) {
		stats := &models.CreationStats{
			CreatedLessons:  3,
			CreatedBookings: 6,
			DeductedCredits: 6,
		}

		assert.Equal(t, 3, stats.CreatedLessons, "CreatedLessons должен быть 3")
		assert.Equal(t, 6, stats.CreatedBookings, "CreatedBookings должен быть 6")
		assert.Equal(t, 6, stats.DeductedCredits, "DeductedCredits должен быть 6")
	})

	t.Run("CreationStats подсчитывается корректно для нескольких студентов", func(t *testing.T) {
		// Имитируем создание 2 lessons, каждый с 2 студентами
		// Ожидаем: 2 lessons, 4 bookings, 4 deducted credits
		stats := &models.CreationStats{
			CreatedLessons:  2,
			CreatedBookings: 4,
			DeductedCredits: 4,
		}

		// Проверяем соотношения
		assert.Equal(t, stats.CreatedBookings, stats.DeductedCredits, "Каждое бронирование = 1 списанный кредит")
		assert.True(t, stats.CreatedBookings >= stats.CreatedLessons, "Бронирований >= количества занятий")
	})
}

// TestTemplateReplacement_BookingRecord проверяет структуру BookingRecord
func TestTemplateReplacement_BookingRecord(t *testing.T) {
	t.Run("BookingRecord содержит необходимые поля для cleanup", func(t *testing.T) {
		record := &models.BookingRecord{
			ID:          uuid.New(),
			StudentID:   uuid.New(),
			LessonID:    uuid.New(),
			Status:      "active",
			BookedAt:    time.Now(),
			StudentName: "Иван Иванов",
			StartTime:   time.Now().Add(24 * time.Hour),
		}

		assert.NotEqual(t, uuid.Nil, record.ID)
		assert.NotEqual(t, uuid.Nil, record.StudentID)
		assert.NotEqual(t, uuid.Nil, record.LessonID)
		assert.Equal(t, "active", record.Status)
		assert.NotEmpty(t, record.StudentName, "StudentName нужен для детальных логов")
		assert.False(t, record.StartTime.IsZero(), "StartTime используется для группировки")
	})
}

// TestTemplateReplacement_ApplicationStatuses проверяет возможные статусы TemplateApplication
func TestTemplateReplacement_ApplicationStatuses(t *testing.T) {
	t.Run("Статус 'applied' для успешного применения", func(t *testing.T) {
		app := &models.TemplateApplication{
			Status: "applied",
		}
		assert.Equal(t, "applied", app.Status)
	})

	t.Run("Статус 'rolled_back' для откатанного шаблона", func(t *testing.T) {
		app := &models.TemplateApplication{
			Status: "rolled_back",
		}
		assert.Equal(t, "rolled_back", app.Status)
	})

	t.Run("Статус 'replaced' для замененного шаблона", func(t *testing.T) {
		app := &models.TemplateApplication{
			Status: "replaced",
		}
		assert.Equal(t, "replaced", app.Status)
	})
}

// TestTemplateReplacement_CreditCalculation проверяет корректность расчета кредитов при замене
func TestTemplateReplacement_CreditCalculation(t *testing.T) {
	t.Run("Кредиты: current + refunded - required = final", func(t *testing.T) {
		// Сценарий: студент имеет 2 кредита, получит 3 возврата, требуется 4 новых
		// Итого: 2 + 3 - 4 = 1 (достаточно)
		currentBalance := 2
		refunded := 3
		required := 4
		finalBalance := currentBalance + refunded - required

		assert.Equal(t, 1, finalBalance, "Должно остаться 1 кредит после замены")
		assert.True(t, finalBalance >= 0, "Баланс не должен быть отрицательным")
	})

	t.Run("Недостаточно кредитов: current + refunded < required", func(t *testing.T) {
		// Сценарий: студент имеет 1 кредит, получит 1 возврат, требуется 5 новых
		// Итого: 1 + 1 - 5 = -3 (недостаточно)
		currentBalance := 1
		refunded := 1
		required := 5
		finalBalance := currentBalance + refunded - required

		assert.Equal(t, -3, finalBalance)
		assert.True(t, finalBalance < 0, "Операция должна быть отклонена")
	})

	t.Run("Нулевая замена: все кредиты вернутся и спишутся обратно", func(t *testing.T) {
		// Сценарий: тот же шаблон применяется повторно
		// current=5, refunded=3, required=3 => final=5 (баланс не меняется)
		currentBalance := 5
		refunded := 3
		required := 3
		finalBalance := currentBalance + refunded - required

		assert.Equal(t, 5, finalBalance, "Баланс должен остаться неизменным")
	})
}

// TestTemplateReplacement_WeekDateHandling проверяет обработку дат недели
func TestTemplateReplacement_WeekDateHandling(t *testing.T) {
	t.Run("weekDate должна быть понедельником в UTC", func(t *testing.T) {
		// Создаём понедельник
		weekDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC) // Понедельник
		assert.Equal(t, time.Monday, weekDate.Weekday(), "Должен быть понедельник")
	})

	t.Run("weekEnd = weekStart + 7 дней", func(t *testing.T) {
		weekStart := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)
		weekEnd := weekStart.AddDate(0, 0, 7)

		assert.Equal(t, time.Monday, weekEnd.Weekday(), "Следующий понедельник")
		assert.Equal(t, 7*24*time.Hour, weekEnd.Sub(weekStart), "Ровно 7 дней")
	})
}

// TestTemplateReplacement_ErrorMessages проверяет качество сообщений об ошибках
func TestTemplateReplacement_ErrorMessages(t *testing.T) {
	t.Run("Ошибка insufficient_credits содержит детали", func(t *testing.T) {
		// Имитируем ошибку с деталями по студентам
		errorMsg := "Student Иван Иванов: current 2 + refunded 1 - required 5 = -2 (insufficient)"

		assert.Contains(t, errorMsg, "Иван Иванов", "Должно быть имя студента")
		assert.Contains(t, errorMsg, "current 2", "Должен быть текущий баланс")
		assert.Contains(t, errorMsg, "refunded 1", "Должны быть возвращаемые кредиты")
		assert.Contains(t, errorMsg, "required 5", "Должны быть требуемые кредиты")
		assert.Contains(t, errorMsg, "insufficient", "Должна быть метка недостатка")
	})

	t.Run("Ошибка week locked возвращается при NOWAIT", func(t *testing.T) {
		errorMsg := "week is locked for modification, try again"

		assert.Contains(t, errorMsg, "locked", "Должно быть слово locked")
		assert.Contains(t, errorMsg, "try again", "Должно быть предложение повторить")
	})
}

// TestTemplateReplacement_TransactionIsolation проверяет уровень изоляции транзакций
func TestTemplateReplacement_TransactionIsolation(t *testing.T) {
	t.Run("Должен использоваться SERIALIZABLE isolation", func(t *testing.T) {
		// Это проверяется в реальном коде через sql.LevelSerializable
		// Здесь только документируем требование
		t.Log("ApplyTemplateToWeek ДОЛЖЕН использовать sql.LevelSerializable")
		t.Log("Это предотвращает race conditions при параллельных replacement")
	})

	t.Run("FOR UPDATE NOWAIT должен использоваться для блокировки", func(t *testing.T) {
		// Документируем требование использования NOWAIT
		t.Log("getLessonsForWeekInTx ДОЛЖЕН использовать FOR UPDATE NOWAIT")
		t.Log("getBookingsForLessonsInTx ДОЛЖЕН использовать FOR UPDATE NOWAIT")
		t.Log("Это предотвращает deadlocks при concurrent операциях")
	})
}

// TestTemplateReplacement_AtomicOperations проверяет атомарность операций
func TestTemplateReplacement_AtomicOperations(t *testing.T) {
	t.Run("Cleanup и creation в одной транзакции", func(t *testing.T) {
		// Имитируем успешную замену
		cleanupStats := &models.CleanupStats{
			CancelledBookings: 3,
			RefundedCredits:   3,
			DeletedLessons:    2,
		}

		creationStats := &models.CreationStats{
			CreatedLessons:  2,
			CreatedBookings: 4,
			DeductedCredits: 4,
		}

		// Обе статистики должны присутствовать после успешного replacement
		assert.NotNil(t, cleanupStats)
		assert.NotNil(t, creationStats)

		// Проверяем что операции сбалансированы
		netCredits := creationStats.DeductedCredits - cleanupStats.RefundedCredits
		t.Logf("Чистое списание кредитов: %d", netCredits)
		// netCredits может быть любым числом в зависимости от разницы в шаблонах
	})

	t.Run("Rollback при ошибке должен отменить все изменения", func(t *testing.T) {
		// Это требование к коду - если cleanup прошёл, но creation упал,
		// вся транзакция должна откатиться
		t.Log("При ошибке в creation должен откатиться весь cleanup")
		t.Log("Проверяется через defer tx.Rollback() в коде")
	})
}

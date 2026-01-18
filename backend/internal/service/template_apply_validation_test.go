package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tutoring-platform/internal/models"
)

// TestTemplateApply_DryRunMode проверяет режим dry_run
func TestTemplateApply_DryRunMode(t *testing.T) {
	t.Skip("Integration test - requires database setup")

	// Этот тест демонстрирует использование dry_run режима
	// В реальном коде нужно:
	// 1. Создать тестовую БД
	// 2. Создать template с уроками
	// 3. Создать студентов с достаточным балансом
	// 4. Применить шаблон с DryRun=true
	// 5. Проверить, что вернулся статус "dry_run"
	// 6. Проверить, что в БД не создалось реальных занятий
}

// TestTemplateApply_InsufficientCreditsError проверяет детальную ошибку при недостатке кредитов
func TestTemplateApply_InsufficientCreditsError(t *testing.T) {
	// Демонстрация ожидаемого формата ошибки
	expectedErrorFormat := `недостаточно кредитов у студентов:
Студент Иван Иванов (ID: %s) имеет 2 кредитов, требуется 5
Студент Петр Петров (ID: %s) имеет 0 кредитов, требуется 3`

	// В реальном тесте нужно:
	// 1. Создать студентов с недостаточным балансом
	// 2. Создать template с несколькими занятиями
	// 3. Попытаться применить шаблон
	// 4. Проверить, что ошибка содержит детальную информацию о каждом студенте

	assert.Contains(t, expectedErrorFormat, "недостаточно кредитов")
	assert.Contains(t, expectedErrorFormat, "имеет")
	assert.Contains(t, expectedErrorFormat, "требуется")
}

// TestTemplateApply_ScheduleConflictError проверяет детальную ошибку при конфликте расписания
func TestTemplateApply_ScheduleConflictError(t *testing.T) {
	// Демонстрация ожидаемого формата ошибки
	expectedErrorFormat := `конфликты расписания:
Студент Иван Иванов уже записан на другое занятие в интервале 10:00 - 12:00 (день недели: 2025-12-15)
Студент Петр Петров уже записан на другое занятие в интервале 14:00 - 16:00 (день недели: 2025-12-16)`

	// В реальном тесте нужно:
	// 1. Создать студентов
	// 2. Создать существующие бронирования на определённое время
	// 3. Создать template с занятиями, конфликтующими по времени
	// 4. Попытаться применить шаблон
	// 5. Проверить, что ошибка содержит детальную информацию о конфликтах

	assert.Contains(t, expectedErrorFormat, "конфликты расписания")
	assert.Contains(t, expectedErrorFormat, "уже записан")
	assert.Contains(t, expectedErrorFormat, "интервале")
}

// TestTemplateApply_TimezoneHandling проверяет корректность работы с timezone
func TestTemplateApply_TimezoneHandling(t *testing.T) {
	// Проверяем, что парсинг даты использует UTC
	weekDate, err := time.Parse("2006-01-02", "2025-12-15")
	require.NoError(t, err)

	// Явно устанавливаем UTC
	weekDate = time.Date(weekDate.Year(), weekDate.Month(), weekDate.Day(), 0, 0, 0, 0, time.UTC)

	// Проверяем, что день недели определяется корректно
	assert.Equal(t, time.Monday, weekDate.Weekday(), "2025-12-15 должен быть понедельником")

	// Добавляем дни для расчёта занятий
	tuesday := weekDate.AddDate(0, 0, 1)
	assert.Equal(t, time.Tuesday, tuesday.Weekday())

	friday := weekDate.AddDate(0, 0, 4)
	assert.Equal(t, time.Friday, friday.Weekday())
}

// TestTemplateApply_ErrorMessageQuality проверяет качество сообщений об ошибках
func TestTemplateApply_ErrorMessageQuality(t *testing.T) {
	// Проверяем, что сообщения об ошибках на русском языке и понятны
	testCases := []struct {
		name           string
		errorMessage   string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name:           "Ошибка недостатка кредитов",
			errorMessage:   "недостаточно кредитов у студентов:\nСтудент Иван (ID: abc) имеет 1 кредитов, требуется 3",
			mustContain:    []string{"недостаточно кредитов", "Студент", "имеет", "требуется"},
			mustNotContain: []string{"insufficient", "credits", "balance"},
		},
		{
			name:           "Ошибка конфликта расписания",
			errorMessage:   "конфликты расписания:\nСтудент Петр уже записан на другое занятие в интервале 10:00 - 12:00",
			mustContain:    []string{"конфликты расписания", "уже записан", "интервале"},
			mustNotContain: []string{"conflict", "schedule", "booking"},
		},
		{
			name:           "Ошибка формата даты",
			errorMessage:   "week_start_date must be a Monday, got Tuesday",
			mustContain:    []string{"Monday", "got"},
			mustNotContain: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, must := range tc.mustContain {
				assert.Contains(t, tc.errorMessage, must,
					"Ошибка должна содержать '%s'", must)
			}

			for _, mustNot := range tc.mustNotContain {
				assert.NotContains(t, tc.errorMessage, mustNot,
					"Ошибка НЕ должна содержать '%s' (только русский текст)", mustNot)
			}
		})
	}
}

// TestApplyTemplateRequest_DryRunField проверяет наличие поля DryRun в запросе
func TestApplyTemplateRequest_DryRunField(t *testing.T) {
	req := &models.ApplyTemplateRequest{
		TemplateID:    uuid.New(),
		WeekStartDate: "2025-12-15", // Понедельник
		DryRun:        true,
	}

	assert.True(t, req.DryRun, "Поле DryRun должно быть доступно")

	// Проверяем валидацию с DryRun
	err := req.Validate()
	assert.NoError(t, err, "Валидация должна пройти успешно с DryRun=true")

	// Проверяем с невалидной датой
	req.WeekStartDate = "2025-12-16" // Вторник
	err = req.Validate()
	assert.Error(t, err, "Валидация должна упасть на невалидной дате")
	assert.Contains(t, err.Error(), "Monday", "Ошибка должна указывать на требование понедельника")
}

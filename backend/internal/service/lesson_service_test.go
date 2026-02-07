package service

import (
	"testing"
	"time"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Тесты валидации типа занятия

func TestValidateLessonTypeChange_GroupToIndividual_NoStudents(t *testing.T) {
	// Создаем урок типа group с 0 студентами
	lesson := &models.Lesson{
		CurrentStudents: 0, // Нет студентов
	}

	// Действие - валидация смены типа на individual
	err := lesson.ValidateLessonTypeChange(models.LessonTypeIndividual)

	// Проверка - должна быть успешной
	assert.NoError(t, err)
	t.Log("TestValidateLessonTypeChange_GroupToIndividual_NoStudents: PASSED - Смена на individual с 0 студентов разрешена")
}

func TestValidateLessonTypeChange_GroupToIndividual_OneStudent(t *testing.T) {
	// Создаем урок типа group с 1 студентом
	lesson := &models.Lesson{
		CurrentStudents: 1, // Один студент
	}

	// Действие - валидация смены типа на individual
	err := lesson.ValidateLessonTypeChange(models.LessonTypeIndividual)

	// Проверка - должна быть успешной (1 студент = допустимо)
	assert.NoError(t, err)
	t.Log("TestValidateLessonTypeChange_GroupToIndividual_OneStudent: PASSED - Смена на individual с 1 студентом разрешена")
}

func TestValidateLessonTypeChange_GroupToIndividual_MultipleStudents(t *testing.T) {
	// Создаем урок типа group с 2+ студентами
	lesson := &models.Lesson{
		CurrentStudents: 2, // ДВА студента - ошибка!
	}

	// Действие - валидация смены типа на individual
	err := lesson.ValidateLessonTypeChange(models.LessonTypeIndividual)

	// Проверка - должна быть ошибка
	assert.Error(t, err)
	assert.Equal(t, models.ErrCannotChangeToIndividual, err)
	t.Log("TestValidateLessonTypeChange_GroupToIndividual_MultipleStudents: PASSED - Ошибка при попытке изменить на individual с 2+ студентами")
}

func TestCreateLesson_InvalidMaxStudents_IndividualWith2(t *testing.T) {
	// Попытка создать individual с max_students = 2
	// With explicit lesson_type=individual and max_students=2, should fail with ErrIndividualLessonMaxStudents
	endTime := time.Now().Add(50 * time.Hour)
	lessonType := models.LessonTypeIndividual
	createReq := &models.CreateLessonRequest{
		TeacherID:   uuid.New(),
		StartTime:   time.Now().Add(48 * time.Hour),
		EndTime:     endTime,
		MaxStudents: 2, // НЕПРАВИЛЬНО! Individual требует ровно 1
		LessonType:  &lessonType,  // Явно указываем тип individual
	}

	// Действие - валидация должна отклонить
	err := createReq.Validate()

	// Проверка
	assert.Error(t, err)
	assert.Equal(t, models.ErrIndividualLessonMaxStudents, err)
	t.Log("TestCreateLesson_InvalidMaxStudents_IndividualWith2: PASSED - Валидация отклонила individual с max_students = 2")
}

func TestCreateLesson_ValidIndividual(t *testing.T) {
	// Создать valid request для individual
	endTime := time.Now().Add(50 * time.Hour)
	createReq := &models.CreateLessonRequest{
		TeacherID:   uuid.New(),
		StartTime:   time.Now().Add(48 * time.Hour),
		EndTime:     endTime,
		MaxStudents: 1, // Правильно!
	}

	// Действие - валидация
	err := createReq.Validate()

	// Проверка
	assert.NoError(t, err)
	t.Log("TestCreateLesson_ValidIndividual: PASSED - Valid individual урок с max_students = 1 принят")
}

func TestValidateLessonTypeChange_IndividualToIndividual(t *testing.T) {
	// Если меняется тип на individual, но уже он individual - должна быть успешной
	lesson := &models.Lesson{
		CurrentStudents: 1,
	}

	// Действие - валидация смены типа на individual (он уже individual)
	err := lesson.ValidateLessonTypeChange(models.LessonTypeIndividual)

	// Проверка - должна быть успешной
	assert.NoError(t, err)
	t.Log("TestValidateLessonTypeChange_IndividualToIndividual: PASSED - Смена individual на individual разрешена")
}

func TestCreateLesson_GroupValidation(t *testing.T) {
	// Проверяем что group урок может иметь > 1 студента
	endTime := time.Now().Add(50 * time.Hour)
	createReq := &models.CreateLessonRequest{
		TeacherID:   uuid.New(),
		StartTime:   time.Now().Add(48 * time.Hour),
		EndTime:     endTime,
		MaxStudents: 5, // Правильно для group
	}

	// Действие - валидация
	err := createReq.Validate()

	// Проверка
	assert.NoError(t, err)
	t.Log("TestCreateLesson_GroupValidation: PASSED - Group урок с max_students = 5 принят")
}

func TestCreateLesson_GroupWith1Student(t *testing.T) {
	// Group урок с 1 студентом - ЗАПРЕЩЕН при создании (требуется минимум 4)
	endTime := time.Now().Add(50 * time.Hour)
	lessonType := models.LessonTypeGroup
	createReq := &models.CreateLessonRequest{
		TeacherID:   uuid.New(),
		StartTime:   time.Now().Add(48 * time.Hour),
		EndTime:     endTime,
		LessonType:  &lessonType,  // Explicitly group lesson
		MaxStudents: 1, // Group требует минимум 4 студента на создание
	}

	// Действие - валидация
	err := createReq.Validate()

	// Проверка - должна быть ошибка (group требует >= 4)
	assert.Error(t, err, "Group урок с max_students < 4 на создание должен отклониться")
	assert.Equal(t, models.ErrGroupLessonMinStudents, err)
	t.Log("TestCreateLesson_GroupWith1Student: PASSED - Group урок с max_students = 1 отклонен")
}

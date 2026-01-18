package service

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ==================== Mocks ====================

type mockHomeworkRepository struct {
	mock.Mock
}

func (m *mockHomeworkRepository) CreateHomework(ctx context.Context, homework *models.LessonHomework) (*models.LessonHomework, error) {
	args := m.Called(ctx, homework)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LessonHomework), args.Error(1)
}

func (m *mockHomeworkRepository) GetHomeworkByLesson(ctx context.Context, lessonID uuid.UUID) ([]*models.LessonHomework, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.LessonHomework), args.Error(1)
}

func (m *mockHomeworkRepository) GetHomeworkByID(ctx context.Context, homeworkID uuid.UUID) (*models.LessonHomework, error) {
	args := m.Called(ctx, homeworkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LessonHomework), args.Error(1)
}

func (m *mockHomeworkRepository) DeleteHomework(ctx context.Context, homeworkID uuid.UUID) error {
	args := m.Called(ctx, homeworkID)
	return args.Error(0)
}

func (m *mockHomeworkRepository) UpdateHomework(ctx context.Context, homeworkID uuid.UUID, textContent string) error {
	args := m.Called(ctx, homeworkID, textContent)
	return args.Error(0)
}

type mockLessonRepository struct {
	mock.Mock
}

func (m *mockLessonRepository) GetByID(ctx context.Context, lessonID uuid.UUID) (*models.Lesson, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Lesson), args.Error(1)
}

type mockBookingRepository struct {
	mock.Mock
}

func (m *mockBookingRepository) GetActiveBookingByStudentAndLesson(ctx context.Context, studentID, lessonID uuid.UUID) (*models.Booking, error) {
	args := m.Called(ctx, studentID, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Booking), args.Error(1)
}

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// ==================== Test CreateHomework ====================

func TestCreateHomework_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	lessonID := uuid.New()
	teacherID := userID

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем урок
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID,
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Мокаем пользователя (admin)
	user := &models.User{
		ID:   userID,
		Role: models.RoleAdmin,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Создаем тестовый файл
	fileContent := []byte("test file content")
	file := bytes.NewReader(fileContent)

	req := &models.CreateHomeworkRequest{
		LessonID: lessonID,
		FileName: "test.pdf",
		FileSize: 1024,
		MimeType: "application/pdf",
	}

	// Переменная для хранения созданного homework
	var createdHomework *models.LessonHomework

	// Мокаем создание записи
	mockHomeworkRepo.On("CreateHomework", ctx, mock.MatchedBy(func(hw *models.LessonHomework) bool {
		// Проверяем что homework имеет правильные поля
		return hw.LessonID == lessonID && hw.FileName == "test.pdf" && hw.FileSize == 1024
	})).Return(&models.LessonHomework{
		ID:        uuid.New(),
		LessonID:  lessonID,
		FileName:  "test.pdf",
		FileSize:  1024,
		MimeType:  "application/pdf",
		CreatedBy: userID,
	}, nil).Run(func(args mock.Arguments) {
		createdHomework = args.Get(1).(*models.LessonHomework)
	})

	// Выполняем тест
	result, err := service.CreateHomework(ctx, userID, file, req)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, lessonID, result.LessonID)
	assert.Equal(t, "test.pdf", result.FileName)

	// Проверяем что файл был создан через service
	assert.NotEmpty(t, createdHomework.FilePath)
	assert.FileExists(t, createdHomework.FilePath)

	// Очистка: удаляем созданный файл
	os.Remove(createdHomework.FilePath)
	// Удаляем директорию если пустая
	os.Remove("uploads/homework")

	mockLessonRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
}

func TestCreateHomework_LessonNotFound(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем урок не найден
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(nil, repository.ErrLessonNotFound)

	fileContent := []byte("test file content")
	file := bytes.NewReader(fileContent)

	req := &models.CreateHomeworkRequest{
		LessonID: lessonID,
		FileName: "test.pdf",
		FileSize: 1024,
		MimeType: "application/pdf",
	}

	// Выполняем тест
	result, err := service.CreateHomework(ctx, userID, file, req)

	// Проверяем результат
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repository.ErrLessonNotFound, err)
}

// ==================== Test GetHomeworkByLesson for Past Lessons ====================

// TestGetHomeworkByLesson_TeacherCanAccessPastLessonHomework проверяет что преподаватель
// может получать ДЗ для прошедших занятий
func TestGetHomeworkByLesson_TeacherCanAccessPastLessonHomework(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Создаем занятие которое уже прошло (в прошлом)
	pastLesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID,
		StartTime: time.Now().Add(-24 * time.Hour),
		EndTime:   time.Now().Add(-23 * time.Hour),
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(pastLesson, nil)

	// Мокаем преподавателя
	teacher := &models.User{
		ID:   teacherID,
		Role: models.RoleTeacher,
	}
	mockUserRepo.On("GetByID", ctx, teacherID).Return(teacher, nil)

	// Мокаем существующие ДЗ
	homeworks := []*models.LessonHomework{
		{
			ID:        uuid.New(),
			LessonID:  lessonID,
			FileName:  "homework.pdf",
			FileSize:  5000,
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		},
	}
	mockHomeworkRepo.On("GetHomeworkByLesson", ctx, lessonID).Return(homeworks, nil)

	// Выполняем тест
	result, err := service.GetHomeworkByLesson(ctx, teacherID, lessonID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "homework.pdf", result[0].FileName)

	mockLessonRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
}

// TestGetHomeworkByLesson_AdminCanAccessPastLessonHomework проверяет что администратор
// может получать ДЗ для прошедших занятий
func TestGetHomeworkByLesson_AdminCanAccessPastLessonHomework(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	teacherID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Создаем занятие которое уже прошло
	pastLesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID,
		StartTime: time.Now().Add(-48 * time.Hour),
		EndTime:   time.Now().Add(-47 * time.Hour),
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(pastLesson, nil)

	// Мокаем администратора
	admin := &models.User{
		ID:   adminID,
		Role: models.RoleAdmin,
	}
	mockUserRepo.On("GetByID", ctx, adminID).Return(admin, nil)

	// Мокаем существующие ДЗ
	homeworks := []*models.LessonHomework{
		{
			ID:        uuid.New(),
			LessonID:  lessonID,
			FileName:  "homework.pdf",
			FileSize:  5000,
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		},
	}
	mockHomeworkRepo.On("GetHomeworkByLesson", ctx, lessonID).Return(homeworks, nil)

	// Выполняем тест
	result, err := service.GetHomeworkByLesson(ctx, adminID, lessonID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)

	mockLessonRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
}

// TestGetHomeworkByLesson_StudentCanAccessPastGroupLesson проверяет что студент
// может получать ДЗ прошедших групповых занятий
func TestGetHomeworkByLesson_StudentCanAccessPastGroupLesson(t *testing.T) {
	ctx := context.Background()
	studentID := uuid.New()
	teacherID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Создаем прошедшее групповое занятие
	pastGroupLesson := &models.Lesson{
		ID:          lessonID,
		TeacherID:   teacherID,
		StartTime:   time.Now().Add(-24 * time.Hour),
		EndTime:     time.Now().Add(-23 * time.Hour),
		MaxStudents: 2, // Групповое занятие (более 1 студента)
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(pastGroupLesson, nil)

	// Мокаем студента
	student := &models.User{
		ID:   studentID,
		Role: models.RoleStudent,
	}
	mockUserRepo.On("GetByID", ctx, studentID).Return(student, nil)

	// Мокаем существующие ДЗ
	homeworks := []*models.LessonHomework{
		{
			ID:        uuid.New(),
			LessonID:  lessonID,
			FileName:  "homework.pdf",
			FileSize:  5000,
			MimeType:  "application/pdf",
			CreatedBy: teacherID,
		},
	}
	mockHomeworkRepo.On("GetHomeworkByLesson", ctx, lessonID).Return(homeworks, nil)

	// Выполняем тест
	result, err := service.GetHomeworkByLesson(ctx, studentID, lessonID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)

	mockLessonRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
}

func TestCreateHomework_PermissionDenied(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	lessonID := uuid.New()
	teacherID := uuid.New() // другой преподаватель

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем урок
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID,
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Мокаем пользователя (student, не admin и не teacher этого урока)
	user := &models.User{
		ID:   userID,
		Role: models.RoleStudent,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	fileContent := []byte("test file content")
	file := bytes.NewReader(fileContent)

	req := &models.CreateHomeworkRequest{
		LessonID: lessonID,
		FileName: "test.pdf",
		FileSize: 1024,
		MimeType: "application/pdf",
	}

	// Выполняем тест
	result, err := service.CreateHomework(ctx, userID, file, req)

	// Проверяем результат
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repository.ErrUnauthorized, err)

	mockLessonRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// ==================== Test GetHomeworkByLesson ====================

func TestGetHomeworkByLesson_AdminSuccess(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем пользователя (admin)
	user := &models.User{
		ID:   userID,
		Role: models.RoleAdmin,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Мокаем урок
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: uuid.New(),
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Мокаем список домашних заданий
	homeworks := []*models.LessonHomework{
		{
			ID:       uuid.New(),
			LessonID: lessonID,
			FileName: "homework1.pdf",
		},
		{
			ID:       uuid.New(),
			LessonID: lessonID,
			FileName: "homework2.pdf",
		},
	}
	mockHomeworkRepo.On("GetHomeworkByLesson", ctx, lessonID).Return(homeworks, nil)

	// Выполняем тест
	result, err := service.GetHomeworkByLesson(ctx, userID, lessonID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	mockUserRepo.AssertExpectations(t)
	mockLessonRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
}

func TestGetHomeworkByLesson_TeacherSuccess(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем пользователя (teacher)
	user := &models.User{
		ID:   userID,
		Role: models.RoleTeacher,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Мокаем урок (teacher этого урока)
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: userID,
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Мокаем список домашних заданий
	homeworks := []*models.LessonHomework{
		{
			ID:       uuid.New(),
			LessonID: lessonID,
			FileName: "homework1.pdf",
		},
	}
	mockHomeworkRepo.On("GetHomeworkByLesson", ctx, lessonID).Return(homeworks, nil)

	// Выполняем тест
	result, err := service.GetHomeworkByLesson(ctx, userID, lessonID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)

	mockUserRepo.AssertExpectations(t)
	mockLessonRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
}

func TestGetHomeworkByLesson_StudentGroupLessonSuccess(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем пользователя (student)
	user := &models.User{
		ID:   userID,
		Role: models.RoleStudent,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Мокаем урок (групповой)
	lesson := &models.Lesson{
		ID:          lessonID,
		TeacherID:   uuid.New(),
		MaxStudents: 2, // Групповое занятие (более 1 студента)
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Мокаем список домашних заданий
	homeworks := []*models.LessonHomework{
		{
			ID:       uuid.New(),
			LessonID: lessonID,
			FileName: "homework1.pdf",
		},
	}
	mockHomeworkRepo.On("GetHomeworkByLesson", ctx, lessonID).Return(homeworks, nil)

	// Выполняем тест
	result, err := service.GetHomeworkByLesson(ctx, userID, lessonID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)

	mockUserRepo.AssertExpectations(t)
	mockLessonRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
}

func TestGetHomeworkByLesson_StudentIndividualLessonWithBooking(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем пользователя (student)
	user := &models.User{
		ID:   userID,
		Role: models.RoleStudent,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Мокаем урок (индивидуальный)
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: uuid.New(),
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Мокаем наличие записи на урок
	booking := &models.Booking{
		ID:        uuid.New(),
		StudentID: userID,
		LessonID:  lessonID,
	}
	mockBookingRepo.On("GetActiveBookingByStudentAndLesson", ctx, userID, lessonID).Return(booking, nil)

	// Мокаем список домашних заданий
	homeworks := []*models.LessonHomework{
		{
			ID:       uuid.New(),
			LessonID: lessonID,
			FileName: "homework1.pdf",
		},
	}
	mockHomeworkRepo.On("GetHomeworkByLesson", ctx, lessonID).Return(homeworks, nil)

	// Выполняем тест
	result, err := service.GetHomeworkByLesson(ctx, userID, lessonID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)

	mockUserRepo.AssertExpectations(t)
	mockLessonRepo.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
}

func TestGetHomeworkByLesson_StudentUnauthorized(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем пользователя (student)
	user := &models.User{
		ID:   userID,
		Role: models.RoleStudent,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Мокаем урок (индивидуальный)
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: uuid.New(),
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Мокаем отсутствие записи на урок
	mockBookingRepo.On("GetActiveBookingByStudentAndLesson", ctx, userID, lessonID).Return(nil, repository.ErrBookingNotFound)

	// Выполняем тест
	result, err := service.GetHomeworkByLesson(ctx, userID, lessonID)

	// Проверяем результат
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repository.ErrUnauthorized, err)

	mockUserRepo.AssertExpectations(t)
	mockLessonRepo.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
}

// ==================== Test DeleteHomework ====================

func TestDeleteHomework_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	homeworkID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Создаем тестовый файл
	tmpFile, err := os.CreateTemp("", "homework_test_*.pdf")
	assert.NoError(t, err)
	tmpFile.Close()

	// Мокаем пользователя (admin)
	user := &models.User{
		ID:   userID,
		Role: models.RoleAdmin,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Мокаем домашнее задание
	homework := &models.LessonHomework{
		ID:        homeworkID,
		LessonID:  lessonID,
		FileName:  "test.pdf",
		FilePath:  tmpFile.Name(),
		CreatedBy: uuid.New(),
	}
	mockHomeworkRepo.On("GetHomeworkByID", ctx, homeworkID).Return(homework, nil)

	// Мокаем урок
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: uuid.New(),
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Мокаем удаление из БД
	mockHomeworkRepo.On("DeleteHomework", ctx, homeworkID).Return(nil)

	// Выполняем тест
	err = service.DeleteHomework(ctx, userID, homeworkID)

	// Проверяем результат
	assert.NoError(t, err)

	// Проверяем что файл удален
	_, err = os.Stat(tmpFile.Name())
	assert.True(t, os.IsNotExist(err))

	mockUserRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
	mockLessonRepo.AssertExpectations(t)
}

func TestDeleteHomework_PermissionDenied(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	homeworkID := uuid.New()
	lessonID := uuid.New()
	creatorID := uuid.New() // другой пользователь
	teacherID := uuid.New() // другой преподаватель

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем пользователя (student, не admin и не creator и не teacher)
	user := &models.User{
		ID:   userID,
		Role: models.RoleStudent,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Мокаем домашнее задание
	homework := &models.LessonHomework{
		ID:        homeworkID,
		LessonID:  lessonID,
		FileName:  "test.pdf",
		FilePath:  "/tmp/test.pdf",
		CreatedBy: creatorID,
	}
	mockHomeworkRepo.On("GetHomeworkByID", ctx, homeworkID).Return(homework, nil)

	// Мокаем урок (teacher - другой пользователь)
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID,
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Выполняем тест
	err := service.DeleteHomework(ctx, userID, homeworkID)

	// Проверяем результат
	assert.Error(t, err)
	assert.Equal(t, repository.ErrUnauthorized, err)

	mockUserRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
	mockLessonRepo.AssertExpectations(t)
}

func TestDeleteHomework_HomeworkNotFound(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	homeworkID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем пользователя
	user := &models.User{
		ID:   userID,
		Role: models.RoleAdmin,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Мокаем домашнее задание не найдено
	mockHomeworkRepo.On("GetHomeworkByID", ctx, homeworkID).Return(nil, repository.ErrHomeworkNotFound)

	// Выполняем тест
	err := service.DeleteHomework(ctx, userID, homeworkID)

	// Проверяем результат
	assert.Error(t, err)
	assert.Equal(t, repository.ErrHomeworkNotFound, err)

	mockUserRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
}

// TestDeleteHomework_TeacherCanDeleteAdminFile проверяет что преподаватель урока
// может удалить файл ДЗ, созданный администратором
func TestDeleteHomework_TeacherCanDeleteAdminFile(t *testing.T) {
	ctx := context.Background()
	teacherID := uuid.New()
	adminID := uuid.New()
	homeworkID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Создаем тестовый файл
	tmpFile, err := os.CreateTemp("", "homework_test_*.pdf")
	assert.NoError(t, err)
	tmpFile.Close()

	// Мокаем пользователя (teacher)
	user := &models.User{
		ID:   teacherID,
		Role: models.RoleTeacher,
	}
	mockUserRepo.On("GetByID", ctx, teacherID).Return(user, nil)

	// Мокаем домашнее задание (создано admin)
	homework := &models.LessonHomework{
		ID:        homeworkID,
		LessonID:  lessonID,
		FileName:  "test.pdf",
		FilePath:  tmpFile.Name(),
		CreatedBy: adminID,
	}
	mockHomeworkRepo.On("GetHomeworkByID", ctx, homeworkID).Return(homework, nil)

	// Мокаем урок (teacher этого урока)
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID,
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Мокаем удаление из БД
	mockHomeworkRepo.On("DeleteHomework", ctx, homeworkID).Return(nil)

	// Выполняем тест
	err = service.DeleteHomework(ctx, teacherID, homeworkID)

	// Проверяем результат
	assert.NoError(t, err)

	// Проверяем что файл удален
	_, err = os.Stat(tmpFile.Name())
	assert.True(t, os.IsNotExist(err))

	mockUserRepo.AssertExpectations(t)
	mockHomeworkRepo.AssertExpectations(t)
	mockLessonRepo.AssertExpectations(t)
}

// ==================== Test GetHomeworkByIDWithAccess ====================

func TestGetHomeworkByIDWithAccess_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	homeworkID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем домашнее задание
	homework := &models.LessonHomework{
		ID:       homeworkID,
		LessonID: lessonID,
		FileName: "test.pdf",
	}
	mockHomeworkRepo.On("GetHomeworkByID", ctx, homeworkID).Return(homework, nil)

	// Мокаем пользователя (admin)
	user := &models.User{
		ID:   userID,
		Role: models.RoleAdmin,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Мокаем урок
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: uuid.New(),
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Мокаем список домашних заданий
	homeworks := []*models.LessonHomework{homework}
	mockHomeworkRepo.On("GetHomeworkByLesson", ctx, lessonID).Return(homeworks, nil)

	// Выполняем тест
	result, err := service.GetHomeworkByIDWithAccess(ctx, userID, homeworkID)

	// Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, homeworkID, result.ID)

	mockHomeworkRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockLessonRepo.AssertExpectations(t)
}

func TestGetHomeworkByIDWithAccess_Unauthorized(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	homeworkID := uuid.New()
	lessonID := uuid.New()

	mockHomeworkRepo := new(mockHomeworkRepository)
	mockLessonRepo := new(mockLessonRepository)
	mockBookingRepo := new(mockBookingRepository)
	mockUserRepo := new(mockUserRepository)

	service := NewHomeworkService(mockHomeworkRepo, mockLessonRepo, mockBookingRepo, mockUserRepo)

	// Мокаем домашнее задание
	homework := &models.LessonHomework{
		ID:       homeworkID,
		LessonID: lessonID,
		FileName: "test.pdf",
	}
	mockHomeworkRepo.On("GetHomeworkByID", ctx, homeworkID).Return(homework, nil)

	// Мокаем пользователя (student без доступа)
	user := &models.User{
		ID:   userID,
		Role: models.RoleStudent,
	}
	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Мокаем урок (индивидуальный)
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: uuid.New(),
	}
	mockLessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)

	// Мокаем отсутствие записи
	mockBookingRepo.On("GetActiveBookingByStudentAndLesson", ctx, userID, lessonID).Return(nil, repository.ErrBookingNotFound)

	// Выполняем тест
	result, err := service.GetHomeworkByIDWithAccess(ctx, userID, homeworkID)

	// Проверяем результат
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repository.ErrUnauthorized, err)

	mockHomeworkRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockLessonRepo.AssertExpectations(t)
	mockBookingRepo.AssertExpectations(t)
}

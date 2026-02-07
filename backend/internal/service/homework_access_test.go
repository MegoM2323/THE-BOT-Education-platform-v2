package service

import (
	"context"
	"testing"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHomeworkRepository - мок для HomeworkRepository
type MockHomeworkRepository struct {
	mock.Mock
}

func (m *MockHomeworkRepository) CreateHomework(ctx context.Context, homework *models.LessonHomework) (*models.LessonHomework, error) {
	args := m.Called(ctx, homework)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LessonHomework), args.Error(1)
}

func (m *MockHomeworkRepository) GetHomeworkByLesson(ctx context.Context, lessonID uuid.UUID) ([]*models.LessonHomework, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.LessonHomework), args.Error(1)
}

func (m *MockHomeworkRepository) GetHomeworkByID(ctx context.Context, homeworkID uuid.UUID) (*models.LessonHomework, error) {
	args := m.Called(ctx, homeworkID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LessonHomework), args.Error(1)
}

func (m *MockHomeworkRepository) DeleteHomework(ctx context.Context, homeworkID uuid.UUID) error {
	args := m.Called(ctx, homeworkID)
	return args.Error(0)
}

func (m *MockHomeworkRepository) UpdateHomework(ctx context.Context, homeworkID uuid.UUID, textContent string) error {
	args := m.Called(ctx, homeworkID, textContent)
	return args.Error(0)
}

// MockHomeworkLessonRepository - мок для LessonRepository
type MockHomeworkLessonRepository struct {
	mock.Mock
}

func (m *MockHomeworkLessonRepository) GetByID(ctx context.Context, lessonID uuid.UUID) (*models.Lesson, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Lesson), args.Error(1)
}

// MockHomeworkBookingRepository - мок для BookingRepository
type MockHomeworkBookingRepository struct {
	mock.Mock
}

func (m *MockHomeworkBookingRepository) GetActiveBookingByStudentAndLesson(ctx context.Context, studentID, lessonID uuid.UUID) (*models.Booking, error) {
	args := m.Called(ctx, studentID, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Booking), args.Error(1)
}

// MockHomeworkUserRepository - мок для UserRepository
type MockHomeworkUserRepository struct {
	mock.Mock
}

func (m *MockHomeworkUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// TestHomeworkAccess_StudentCanViewHomeworkWhenBooked - тест доступа студента к ДЗ когда он записан на урок
func TestHomeworkAccess_StudentCanViewHomeworkWhenBooked(t *testing.T) {
	ctx := context.Background()

	// Создаем моки
	homeworkRepo := new(MockHomeworkRepository)
	lessonRepo := new(MockHomeworkLessonRepository)
	bookingRepo := new(MockHomeworkBookingRepository)
	userRepo := new(MockHomeworkUserRepository)

	// ID сущностей
	studentID := uuid.New()
	teacherID := uuid.New()
	lessonID := uuid.New()
	homeworkID := uuid.New()

	// Создаем студента
	student := &models.User{
		ID:        studentID,
		Email:     "student@example.com",
		FirstName: "Студент",
		LastName:  "Иванов",
		Role:      models.RoleStudent,
	}

	// Создаем индивидуальный урок
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID,
	}

	// Создаем активное бронирование
	booking := &models.Booking{
		ID:        uuid.New(),
		StudentID: studentID,
		LessonID:  lessonID,
		Status:    models.BookingStatusActive,
	}

	// Создаем домашнее задание
	homework := &models.LessonHomework{
		ID:        homeworkID,
		LessonID:  lessonID,
		FileName:  "homework.pdf",
		FilePath:  "/tmp/homework.pdf",
		FileSize:  1024,
		MimeType:  "application/pdf",
		CreatedBy: teacherID,
	}

	homeworkList := []*models.LessonHomework{homework}

	// Настраиваем ожидания моков
	userRepo.On("GetByID", ctx, studentID).Return(student, nil)
	lessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)
	bookingRepo.On("GetActiveBookingByStudentAndLesson", ctx, studentID, lessonID).Return(booking, nil)
	homeworkRepo.On("GetHomeworkByLesson", ctx, lessonID).Return(homeworkList, nil)

	// Создаем сервис
	service := NewHomeworkService(homeworkRepo, lessonRepo, bookingRepo, userRepo)

	// Вызываем метод
	result, err := service.GetHomeworkByLesson(ctx, studentID, lessonID)

	// Проверяем результат
	assert.NoError(t, err, "Студент должен иметь доступ к ДЗ урока, на который он записан")
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, homeworkID, result[0].ID)
	assert.Equal(t, "homework.pdf", result[0].FileName)

	// Проверяем что все моки были вызваны
	userRepo.AssertExpectations(t)
	lessonRepo.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
	homeworkRepo.AssertExpectations(t)
}

// TestHomeworkAccess_StudentCannotViewHomeworkWhenNotBooked - тест что студент НЕ видит ДЗ когда он не записан
func TestHomeworkAccess_StudentCannotViewHomeworkWhenNotBooked(t *testing.T) {
	ctx := context.Background()

	// Создаем моки
	homeworkRepo := new(MockHomeworkRepository)
	lessonRepo := new(MockHomeworkLessonRepository)
	bookingRepo := new(MockHomeworkBookingRepository)
	userRepo := new(MockHomeworkUserRepository)

	// ID сущностей
	studentID := uuid.New()
	teacherID := uuid.New()
	lessonID := uuid.New()

	// Создаем студента
	student := &models.User{
		ID:        studentID,
		Email:     "student@example.com",
		FirstName: "Студент",
		LastName:  "Иванов",
		Role:      models.RoleStudent,
	}

	// Создаем индивидуальный урок
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID,
	}

	// Настраиваем ожидания моков (студент НЕ записан на урок)
	userRepo.On("GetByID", ctx, studentID).Return(student, nil)
	lessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)
	bookingRepo.On("GetActiveBookingByStudentAndLesson", ctx, studentID, lessonID).Return(nil, repository.ErrBookingNotFound)

	// Создаем сервис
	service := NewHomeworkService(homeworkRepo, lessonRepo, bookingRepo, userRepo)

	// Вызываем метод
	result, err := service.GetHomeworkByLesson(ctx, studentID, lessonID)

	// Проверяем результат
	assert.Error(t, err, "Студент НЕ должен иметь доступ к ДЗ урока, на который он не записан")
	assert.Equal(t, repository.ErrUnauthorized, err)
	assert.Nil(t, result)

	// Проверяем что все моки были вызваны
	userRepo.AssertExpectations(t)
	lessonRepo.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
}

// TestHomeworkAccess_StudentCanViewGroupLessonHomework - тест что студент видит ДЗ группового урока без бронирования
func TestHomeworkAccess_StudentCanViewGroupLessonHomework(t *testing.T) {
	ctx := context.Background()

	// Создаем моки
	homeworkRepo := new(MockHomeworkRepository)
	lessonRepo := new(MockHomeworkLessonRepository)
	bookingRepo := new(MockHomeworkBookingRepository)
	userRepo := new(MockHomeworkUserRepository)

	// ID сущностей
	studentID := uuid.New()
	teacherID := uuid.New()
	lessonID := uuid.New()
	homeworkID := uuid.New()

	// Создаем студента
	student := &models.User{
		ID:        studentID,
		Email:     "student@example.com",
		FirstName: "Студент",
		LastName:  "Иванов",
		Role:      models.RoleStudent,
	}

	// Создаем ГРУППОВОЙ урок
	lesson := &models.Lesson{
		ID:          lessonID,
		TeacherID:   teacherID,
		MaxStudents: 2,
	}

	// Создаем домашнее задание
	homework := &models.LessonHomework{
		ID:        homeworkID,
		LessonID:  lessonID,
		FileName:  "group_homework.pdf",
		FilePath:  "/tmp/group_homework.pdf",
		FileSize:  2048,
		MimeType:  "application/pdf",
		CreatedBy: teacherID,
	}

	homeworkList := []*models.LessonHomework{homework}

	// Настраиваем ожидания моков
	userRepo.On("GetByID", ctx, studentID).Return(student, nil)
	lessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)
	// Для группового урока НЕ проверяем бронирование
	homeworkRepo.On("GetHomeworkByLesson", ctx, lessonID).Return(homeworkList, nil)

	// Создаем сервис
	service := NewHomeworkService(homeworkRepo, lessonRepo, bookingRepo, userRepo)

	// Вызываем метод
	result, err := service.GetHomeworkByLesson(ctx, studentID, lessonID)

	// Проверяем результат
	assert.NoError(t, err, "Студент должен иметь доступ к ДЗ группового урока даже без бронирования")
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, homeworkID, result[0].ID)
	assert.Equal(t, "group_homework.pdf", result[0].FileName)

	// Проверяем что все моки были вызваны
	userRepo.AssertExpectations(t)
	lessonRepo.AssertExpectations(t)
	homeworkRepo.AssertExpectations(t)
	// bookingRepo НЕ должен был вызываться для группового урока
	bookingRepo.AssertNotCalled(t, "GetActiveBookingByStudentAndLesson")
}

// TestHomeworkAccess_AdminCanViewAll - тест что админ видит всё
func TestHomeworkAccess_AdminCanViewAll(t *testing.T) {
	ctx := context.Background()

	// Создаем моки
	homeworkRepo := new(MockHomeworkRepository)
	lessonRepo := new(MockHomeworkLessonRepository)
	bookingRepo := new(MockHomeworkBookingRepository)
	userRepo := new(MockHomeworkUserRepository)

	// ID сущностей
	adminID := uuid.New()
	teacherID := uuid.New()
	lessonID := uuid.New()
	homeworkID := uuid.New()

	// Создаем админа
	admin := &models.User{
		ID:        adminID,
		Email:     "admin@example.com",
		FirstName: "Админ",
		LastName:  "Админов",
		Role:      models.RoleAdmin,
	}

	// Создаем индивидуальный урок
	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID,
	}

	// Создаем домашнее задание
	homework := &models.LessonHomework{
		ID:        homeworkID,
		LessonID:  lessonID,
		FileName:  "admin_homework.pdf",
		FilePath:  "/tmp/admin_homework.pdf",
		FileSize:  512,
		MimeType:  "application/pdf",
		CreatedBy: teacherID,
	}

	homeworkList := []*models.LessonHomework{homework}

	// Настраиваем ожидания моков
	userRepo.On("GetByID", ctx, adminID).Return(admin, nil)
	lessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)
	homeworkRepo.On("GetHomeworkByLesson", ctx, lessonID).Return(homeworkList, nil)

	// Создаем сервис
	service := NewHomeworkService(homeworkRepo, lessonRepo, bookingRepo, userRepo)

	// Вызываем метод
	result, err := service.GetHomeworkByLesson(ctx, adminID, lessonID)

	// Проверяем результат
	assert.NoError(t, err, "Админ должен иметь доступ к любому ДЗ")
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, homeworkID, result[0].ID)

	// Проверяем что все моки были вызваны
	userRepo.AssertExpectations(t)
	lessonRepo.AssertExpectations(t)
	homeworkRepo.AssertExpectations(t)
	// bookingRepo НЕ должен был вызываться для админа
	bookingRepo.AssertNotCalled(t, "GetActiveBookingByStudentAndLesson")
}

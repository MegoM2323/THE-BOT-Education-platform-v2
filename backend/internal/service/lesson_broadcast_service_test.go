package service

import (
	"context"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Моки для тестирования

type MockLessonBroadcastRepo struct {
	mock.Mock
}

func (m *MockLessonBroadcastRepo) CreateBroadcast(ctx context.Context, broadcast *models.LessonBroadcast) (*models.LessonBroadcast, error) {
	args := m.Called(ctx, broadcast)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LessonBroadcast), args.Error(1)
}

func (m *MockLessonBroadcastRepo) GetBroadcast(ctx context.Context, broadcastID uuid.UUID) (*models.LessonBroadcast, error) {
	args := m.Called(ctx, broadcastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LessonBroadcast), args.Error(1)
}

func (m *MockLessonBroadcastRepo) ListBroadcastsByLesson(ctx context.Context, lessonID uuid.UUID) ([]*models.LessonBroadcast, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.LessonBroadcast), args.Error(1)
}

func (m *MockLessonBroadcastRepo) UpdateBroadcastStatus(ctx context.Context, broadcastID uuid.UUID, status string, sentCount, failedCount int) error {
	args := m.Called(ctx, broadcastID, status, sentCount, failedCount)
	return args.Error(0)
}

func (m *MockLessonBroadcastRepo) AddBroadcastFile(ctx context.Context, file *models.BroadcastFile) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockLessonBroadcastRepo) GetBroadcastFiles(ctx context.Context, broadcastID uuid.UUID) ([]*models.BroadcastFile, error) {
	args := m.Called(ctx, broadcastID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.BroadcastFile), args.Error(1)
}

func (m *MockLessonBroadcastRepo) GetBroadcastFile(ctx context.Context, id uuid.UUID) (*models.BroadcastFile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BroadcastFile), args.Error(1)
}

type MockLessonRepo struct {
	mock.Mock
}

func (m *MockLessonRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Lesson, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Lesson), args.Error(1)
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepo) List(ctx context.Context, roleFilter *models.UserRole) ([]*models.User, error) {
	args := m.Called(ctx, roleFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepo) ListWithPagination(ctx context.Context, role *models.UserRole, offset, limit int) ([]*models.User, int, error) {
	args := m.Called(ctx, role, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Int(1), args.Error(2)
}

func (m *MockUserRepo) Exists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepo) UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error {
	args := m.Called(ctx, userID, username)
	return args.Error(0)
}

func (m *MockUserRepo) GetAllUsersWithTelegramInfo(ctx context.Context) ([]map[string]interface{}, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

// Тесты CreateLessonBroadcast

func TestCreateLessonBroadcast_InvalidMessage_Empty(t *testing.T) {
	// Setup
	ctx := context.Background()
	service := NewLessonBroadcastService(nil, nil, nil, nil, nil, nil, "")

	teacherID := uuid.New()
	lessonID := uuid.New()

	// Execute & Assert
	_, err := service.CreateLessonBroadcast(ctx, teacherID, lessonID, "", nil)
	assert.ErrorIs(t, err, ErrInvalidMessage)
}

func TestCreateLessonBroadcast_InvalidMessage_TooLong(t *testing.T) {
	// Setup
	ctx := context.Background()
	service := NewLessonBroadcastService(nil, nil, nil, nil, nil, nil, "")

	teacherID := uuid.New()
	lessonID := uuid.New()
	longMessage := strings.Repeat("a", models.MaxBroadcastMessageLen+1)

	// Execute & Assert
	_, err := service.CreateLessonBroadcast(ctx, teacherID, lessonID, longMessage, nil)
	assert.ErrorIs(t, err, ErrInvalidMessage)
}

func TestCreateLessonBroadcast_TooManyFiles(t *testing.T) {
	// Setup
	ctx := context.Background()
	service := NewLessonBroadcastService(nil, nil, nil, nil, nil, nil, "")

	teacherID := uuid.New()
	lessonID := uuid.New()

	// Создаём 4 файла (больше максимума)
	files := make([]*multipart.FileHeader, models.MaxBroadcastFiles+1)

	// Execute & Assert
	_, err := service.CreateLessonBroadcast(ctx, teacherID, lessonID, "Test", files)
	assert.ErrorIs(t, err, ErrTooManyFiles)
}

func TestCreateLessonBroadcast_LessonNotFound(t *testing.T) {
	// Setup
	ctx := context.Background()
	lessonRepo := new(MockLessonRepo)
	service := NewLessonBroadcastService(nil, nil, lessonRepo, nil, nil, nil, "")

	teacherID := uuid.New()
	lessonID := uuid.New()

	lessonRepo.On("GetByID", ctx, lessonID).Return(nil, repository.ErrLessonNotFound)

	// Execute & Assert
	_, err := service.CreateLessonBroadcast(ctx, teacherID, lessonID, "Test message", nil)
	assert.ErrorIs(t, err, repository.ErrLessonNotFound)

	lessonRepo.AssertExpectations(t)
}

func TestCreateLessonBroadcast_Unauthorized_Student(t *testing.T) {
	// Setup
	ctx := context.Background()
	lessonRepo := new(MockLessonRepo)
	userRepo := new(MockUserRepo)
	service := NewLessonBroadcastService(nil, nil, lessonRepo, userRepo, nil, nil, "")

	teacherID := uuid.New()
	studentID := uuid.New()
	lessonID := uuid.New()

	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID,
	}

	student := &models.User{
		ID:   studentID,
		Role: models.RoleStudent, // Not teacher, not admin
	}

	lessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)
	userRepo.On("GetByID", ctx, studentID).Return(student, nil)

	// Execute & Assert
	_, err := service.CreateLessonBroadcast(ctx, studentID, lessonID, "Test", nil)
	assert.ErrorIs(t, err, repository.ErrUnauthorized)

	lessonRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestCreateLessonBroadcast_Success_Teacher(t *testing.T) {
	// Setup
	ctx := context.Background()
	broadcastRepo := new(MockLessonBroadcastRepo)
	lessonRepo := new(MockLessonRepo)
	userRepo := new(MockUserRepo)

	tmpDir := t.TempDir()
	service := NewLessonBroadcastService(
		nil, // db not needed
		broadcastRepo,
		lessonRepo,
		userRepo,
		nil, // telegram repo
		nil, // telegram client
		tmpDir,
	)

	teacherID := uuid.New()
	lessonID := uuid.New()
	broadcastID := uuid.New()

	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID,
		StartTime: time.Now().Add(24 * time.Hour),
	}

	teacher := &models.User{
		ID:   teacherID,
		Role: models.RoleTeacher,
	}

	broadcast := &models.LessonBroadcast{
		ID:       broadcastID,
		LessonID: lessonID,
		SenderID: teacherID,
		Message:  "Test message",
		Status:   models.LessonBroadcastStatusPending,
	}

	lessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)
	userRepo.On("GetByID", ctx, teacherID).Return(teacher, nil)
	broadcastRepo.On("CreateBroadcast", ctx, mock.AnythingOfType("*models.LessonBroadcast")).Return(broadcast, nil)
	broadcastRepo.On("GetBroadcastFiles", ctx, broadcastID).Return([]*models.BroadcastFile{}, nil)

	// Mock для async goroutine (может быть вызван)
	broadcastRepo.On("GetBroadcast", mock.Anything, broadcastID).Return(broadcast, nil).Maybe()
	broadcastRepo.On("UpdateBroadcastStatus", mock.Anything, broadcastID, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	lessonRepo.On("GetByID", mock.Anything, lessonID).Return(lesson, nil).Maybe()

	// Execute
	result, err := service.CreateLessonBroadcast(ctx, teacherID, lessonID, "Test message", nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, broadcastID, result.ID)
	assert.Equal(t, models.LessonBroadcastStatusPending, result.Status)

	// Даём время горутине завершиться
	time.Sleep(100 * time.Millisecond)

	broadcastRepo.AssertExpectations(t)
	lessonRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestCreateLessonBroadcast_Success_Admin(t *testing.T) {
	// Setup
	ctx := context.Background()
	broadcastRepo := new(MockLessonBroadcastRepo)
	lessonRepo := new(MockLessonRepo)
	userRepo := new(MockUserRepo)

	tmpDir := t.TempDir()
	service := NewLessonBroadcastService(
		nil,
		broadcastRepo,
		lessonRepo,
		userRepo,
		nil,
		nil,
		tmpDir,
	)

	adminID := uuid.New()
	teacherID := uuid.New()
	lessonID := uuid.New()
	broadcastID := uuid.New()

	lesson := &models.Lesson{
		ID:        lessonID,
		TeacherID: teacherID, // Different from admin
	}

	admin := &models.User{
		ID:   adminID,
		Role: models.RoleAdmin, // Admin can broadcast any lesson
	}

	broadcast := &models.LessonBroadcast{
		ID:       broadcastID,
		LessonID: lessonID,
		SenderID: adminID,
		Message:  "Admin broadcast",
		Status:   models.LessonBroadcastStatusPending,
	}

	lessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)
	userRepo.On("GetByID", ctx, adminID).Return(admin, nil)
	broadcastRepo.On("CreateBroadcast", ctx, mock.AnythingOfType("*models.LessonBroadcast")).Return(broadcast, nil)
	broadcastRepo.On("GetBroadcastFiles", ctx, broadcastID).Return([]*models.BroadcastFile{}, nil)

	// Mock для async goroutine (может быть вызван)
	broadcastRepo.On("GetBroadcast", mock.Anything, broadcastID).Return(broadcast, nil).Maybe()
	broadcastRepo.On("UpdateBroadcastStatus", mock.Anything, broadcastID, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	lessonRepo.On("GetByID", mock.Anything, lessonID).Return(lesson, nil).Maybe()

	// Execute
	result, err := service.CreateLessonBroadcast(ctx, adminID, lessonID, "Admin broadcast", nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Даём время горутине завершиться
	time.Sleep(100 * time.Millisecond)

	lessonRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	broadcastRepo.AssertExpectations(t)
}

// Тесты ListLessonBroadcasts

func TestListLessonBroadcasts_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	broadcastRepo := new(MockLessonBroadcastRepo)
	lessonRepo := new(MockLessonRepo)
	service := NewLessonBroadcastService(nil, broadcastRepo, lessonRepo, nil, nil, nil, "")

	lessonID := uuid.New()
	lesson := &models.Lesson{ID: lessonID}

	broadcasts := []*models.LessonBroadcast{
		{
			ID:       uuid.New(),
			LessonID: lessonID,
			Message:  "Broadcast 1",
		},
		{
			ID:       uuid.New(),
			LessonID: lessonID,
			Message:  "Broadcast 2",
		},
	}

	lessonRepo.On("GetByID", ctx, lessonID).Return(lesson, nil)
	broadcastRepo.On("ListBroadcastsByLesson", ctx, lessonID).Return(broadcasts, nil)

	// Execute
	result, err := service.ListLessonBroadcasts(ctx, lessonID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Broadcast 1", result[0].Message)

	lessonRepo.AssertExpectations(t)
	broadcastRepo.AssertExpectations(t)
}

func TestListLessonBroadcasts_LessonNotFound(t *testing.T) {
	// Setup
	ctx := context.Background()
	lessonRepo := new(MockLessonRepo)
	service := NewLessonBroadcastService(nil, nil, lessonRepo, nil, nil, nil, "")

	lessonID := uuid.New()
	lessonRepo.On("GetByID", ctx, lessonID).Return(nil, repository.ErrLessonNotFound)

	// Execute & Assert
	_, err := service.ListLessonBroadcasts(ctx, lessonID)
	assert.ErrorIs(t, err, repository.ErrLessonNotFound)

	lessonRepo.AssertExpectations(t)
}

// Тесты GetLessonBroadcast

func TestGetLessonBroadcast_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	broadcastRepo := new(MockLessonBroadcastRepo)
	service := NewLessonBroadcastService(nil, broadcastRepo, nil, nil, nil, nil, "")

	broadcastID := uuid.New()
	broadcast := &models.LessonBroadcast{
		ID:      broadcastID,
		Message: "Test broadcast",
		Files: []*models.BroadcastFile{
			{FileName: "test.pdf"},
		},
	}

	broadcastRepo.On("GetBroadcast", ctx, broadcastID).Return(broadcast, nil)

	// Execute
	result, err := service.GetLessonBroadcast(ctx, broadcastID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test broadcast", result.Message)
	assert.Len(t, result.Files, 1)

	broadcastRepo.AssertExpectations(t)
}

func TestGetLessonBroadcast_NotFound(t *testing.T) {
	// Setup
	ctx := context.Background()
	broadcastRepo := new(MockLessonBroadcastRepo)
	service := NewLessonBroadcastService(nil, broadcastRepo, nil, nil, nil, nil, "")

	broadcastID := uuid.New()
	broadcastRepo.On("GetBroadcast", ctx, broadcastID).Return(nil, repository.ErrLessonBroadcastNotFound)

	// Execute & Assert
	_, err := service.GetLessonBroadcast(ctx, broadcastID)
	assert.ErrorIs(t, err, repository.ErrLessonBroadcastNotFound)

	broadcastRepo.AssertExpectations(t)
}

// Тесты SendBroadcastAsync - проверяем что метод не паникует

func TestSendBroadcastAsync_BroadcastNotFound(t *testing.T) {
	// Setup
	ctx := context.Background()
	broadcastRepo := new(MockLessonBroadcastRepo)

	// Создаём mock Telegram client (nil client приводит к раннему выходу)
	service := NewLessonBroadcastService(nil, broadcastRepo, nil, nil, nil, nil, "")

	broadcastID := uuid.New()

	// Telegram client == nil, поэтому GetBroadcast не будет вызван
	// Проверяем что не паникует
	assert.NotPanics(t, func() {
		service.SendBroadcastAsync(ctx, broadcastID)
	})

	// Не проверяем expectations т.к. метод вернётся раньше из-за nil telegramClient
}

func TestSendBroadcastAsync_NotPendingStatus(t *testing.T) {
	// Setup
	ctx := context.Background()
	broadcastRepo := new(MockLessonBroadcastRepo)
	service := NewLessonBroadcastService(nil, broadcastRepo, nil, nil, nil, nil, "")

	broadcastID := uuid.New()

	// Telegram client == nil, поэтому метод вернётся раньше
	// Проверяем что не паникует
	assert.NotPanics(t, func() {
		service.SendBroadcastAsync(ctx, broadcastID)
	})

	// Не проверяем expectations т.к. метод вернётся раньше из-за nil telegramClient
}

// Тесты sendMessage

// Removed TestSendMessage_Success - it duplicates test in broadcast_service_test.go

func TestSendMessage_TelegramClientNotConfigured(t *testing.T) {
	// Setup
	service := NewLessonBroadcastService(nil, nil, nil, nil, nil, nil, "")

	// Execute
	err := service.sendMessage(12345, "Test")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "telegram client not configured")
}

func TestSendFile_TelegramClientNotConfigured(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	service := NewLessonBroadcastService(nil, nil, nil, nil, nil, nil, tmpDir)

	file := &models.BroadcastFile{
		FileName: "test.pdf",
		FilePath: "test.pdf",
	}

	// Execute
	err := service.sendFile(12345, file)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "telegram client not configured")
}

// Тесты finalizeBroadcast

func TestFinalizeBroadcast_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	broadcastRepo := new(MockLessonBroadcastRepo)
	service := NewLessonBroadcastService(nil, broadcastRepo, nil, nil, nil, nil, "")

	broadcastID := uuid.New()
	broadcastRepo.On("UpdateBroadcastStatus", ctx, broadcastID, models.LessonBroadcastStatusCompleted, 5, 2).Return(nil)

	// Execute - should not panic
	assert.NotPanics(t, func() {
		service.finalizeBroadcast(ctx, broadcastID, 5, 2, models.LessonBroadcastStatusCompleted)
	})

	broadcastRepo.AssertExpectations(t)
}

func TestFinalizeBroadcast_FailedStatus(t *testing.T) {
	// Setup
	ctx := context.Background()
	broadcastRepo := new(MockLessonBroadcastRepo)
	service := NewLessonBroadcastService(nil, broadcastRepo, nil, nil, nil, nil, "")

	broadcastID := uuid.New()
	broadcastRepo.On("UpdateBroadcastStatus", ctx, broadcastID, models.LessonBroadcastStatusFailed, 0, 3).Return(nil)

	// Execute - should not panic
	assert.NotPanics(t, func() {
		service.finalizeBroadcast(ctx, broadcastID, 0, 3, models.LessonBroadcastStatusFailed)
	})

	broadcastRepo.AssertExpectations(t)
}

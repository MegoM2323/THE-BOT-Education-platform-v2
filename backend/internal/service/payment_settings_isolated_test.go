package service

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// LocalMockUserRepo - изолированный мок для тестирования payment settings service
type LocalMockUserRepo struct {
	mock.Mock
}

func (m *LocalMockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *LocalMockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *LocalMockUserRepo) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *LocalMockUserRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *LocalMockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *LocalMockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *LocalMockUserRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *LocalMockUserRepo) List(ctx context.Context, roleFilter *models.UserRole) ([]*models.User, error) {
	args := m.Called(ctx, roleFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *LocalMockUserRepo) ListWithPagination(ctx context.Context, role *models.UserRole, offset, limit int) ([]*models.User, int, error) {
	args := m.Called(ctx, role, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Int(1), args.Error(2)
}

func (m *LocalMockUserRepo) Exists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *LocalMockUserRepo) UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error {
	args := m.Called(ctx, userID, username)
	return args.Error(0)
}

func TestPaymentSettings_GetStatus_Success(t *testing.T) {
	mockRepo := new(LocalMockUserRepo)
	service := NewPaymentSettingsService(mockRepo)
	ctx := context.Background()
	userID := uuid.New()

	expectedUser := &models.User{
		ID:             userID,
		Email:          "student@example.com",
		FirstName: "Test Student", LastName: "Lastname",
		Role:           models.RoleStudent,
		PaymentEnabled: true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

	enabled, err := service.GetPaymentStatus(ctx, userID)

	assert.NoError(t, err)
	assert.True(t, enabled)
	mockRepo.AssertExpectations(t)
}

func TestPaymentSettings_UpdateStatus_Success(t *testing.T) {
	mockRepo := new(LocalMockUserRepo)
	service := NewPaymentSettingsService(mockRepo)
	ctx := context.Background()
	adminID := uuid.New()
	studentID := uuid.New()

	admin := &models.User{
		ID:       adminID,
		Email:    "admin@example.com",
		FirstName: "Admin User", LastName: "Lastname",
		Role:     models.RoleAdmin,
	}

	student := &models.User{
		ID:             studentID,
		Email:          "student@example.com",
		FirstName: "Student User", LastName: "Lastname",
		Role:           models.RoleStudent,
		PaymentEnabled: false,
	}

	updatedStudent := &models.User{
		ID:             studentID,
		Email:          "student@example.com",
		FirstName: "Student User", LastName: "Lastname",
		Role:           models.RoleStudent,
		PaymentEnabled: true,
		UpdatedAt:      time.Now(),
	}

	mockRepo.On("GetByID", ctx, adminID).Return(admin, nil)
	mockRepo.On("GetByID", ctx, studentID).Return(student, nil).Once()
	mockRepo.On("Update", ctx, studentID, map[string]interface{}{"payment_enabled": true}).Return(nil)
	mockRepo.On("GetByID", ctx, studentID).Return(updatedStudent, nil).Once()

	result, err := service.UpdatePaymentStatus(ctx, adminID, studentID, true)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.PaymentEnabled)
	mockRepo.AssertExpectations(t)
}

func TestPaymentSettings_UpdateStatus_NonAdminDenied(t *testing.T) {
	mockRepo := new(LocalMockUserRepo)
	service := NewPaymentSettingsService(mockRepo)
	ctx := context.Background()
	teacherID := uuid.New()
	studentID := uuid.New()

	teacher := &models.User{
		ID:       teacherID,
		Email:    "teacher@example.com",
		FirstName: "Teacher User", LastName: "Lastname",
		Role:     models.RoleTeacher,
	}

	mockRepo.On("GetByID", ctx, teacherID).Return(teacher, nil)

	result, err := service.UpdatePaymentStatus(ctx, teacherID, studentID, true)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repository.ErrUnauthorized, err)
	mockRepo.AssertExpectations(t)
}

func TestPaymentSettings_UpdateStatus_NonStudentTargetDenied(t *testing.T) {
	mockRepo := new(LocalMockUserRepo)
	service := NewPaymentSettingsService(mockRepo)
	ctx := context.Background()
	adminID := uuid.New()
	teacherID := uuid.New()

	admin := &models.User{
		ID:       adminID,
		Email:    "admin@example.com",
		FirstName: "Admin User", LastName: "Lastname",
		Role:     models.RoleAdmin,
	}

	teacher := &models.User{
		ID:       teacherID,
		Email:    "teacher@example.com",
		FirstName: "Teacher User", LastName: "Lastname",
		Role:     models.RoleTeacher,
	}

	mockRepo.On("GetByID", ctx, adminID).Return(admin, nil)
	mockRepo.On("GetByID", ctx, teacherID).Return(teacher, nil)

	result, err := service.UpdatePaymentStatus(ctx, adminID, teacherID, false)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repository.ErrInvalidUserRole, err)
	mockRepo.AssertExpectations(t)
}

func TestPaymentSettings_ListStudents_Success(t *testing.T) {
	mockRepo := new(LocalMockUserRepo)
	service := NewPaymentSettingsService(mockRepo)
	ctx := context.Background()
	adminID := uuid.New()

	admin := &models.User{
		ID:       adminID,
		Email:    "admin@example.com",
		FirstName: "Admin User", LastName: "Lastname",
		Role:     models.RoleAdmin,
	}

	students := []*models.User{
		{
			ID:             uuid.New(),
			Email:          "charlie@example.com",
			FirstName: "Charlie Student", LastName: "Lastname",
			Role:           models.RoleStudent,
			PaymentEnabled: true,
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			Email:          "alice@example.com",
			FirstName: "Alice Student", LastName: "Lastname",
			Role:           models.RoleStudent,
			PaymentEnabled: false,
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			Email:          "bob@example.com",
			FirstName: "Bob Student", LastName: "Lastname",
			Role:           models.RoleStudent,
			PaymentEnabled: true,
			UpdatedAt:      time.Now(),
		},
	}

	studentRole := models.RoleStudent
	mockRepo.On("GetByID", ctx, adminID).Return(admin, nil)
	mockRepo.On("List", ctx, &studentRole).Return(students, nil)

	result, err := service.ListStudentsPaymentStatus(ctx, adminID, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)

	// Проверяем сортировку по имени
	assert.Equal(t, "Alice Student", result[0].FullName)
	assert.Equal(t, "Bob Student", result[1].FullName)
	assert.Equal(t, "Charlie Student", result[2].FullName)

	mockRepo.AssertExpectations(t)
}

func TestPaymentSettings_ListStudents_WithFilter(t *testing.T) {
	mockRepo := new(LocalMockUserRepo)
	service := NewPaymentSettingsService(mockRepo)
	ctx := context.Background()
	adminID := uuid.New()
	filterEnabled := false

	admin := &models.User{
		ID:       adminID,
		Email:    "admin@example.com",
		FirstName: "Admin User", LastName: "Lastname",
		Role:     models.RoleAdmin,
	}

	students := []*models.User{
		{
			ID:             uuid.New(),
			Email:          "charlie@example.com",
			FirstName: "Charlie Student", LastName: "Lastname",
			Role:           models.RoleStudent,
			PaymentEnabled: true,
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			Email:          "alice@example.com",
			FirstName: "Alice Student", LastName: "Lastname",
			Role:           models.RoleStudent,
			PaymentEnabled: false,
			UpdatedAt:      time.Now(),
		},
	}

	studentRole := models.RoleStudent
	mockRepo.On("GetByID", ctx, adminID).Return(admin, nil)
	mockRepo.On("List", ctx, &studentRole).Return(students, nil)

	result, err := service.ListStudentsPaymentStatus(ctx, adminID, &filterEnabled)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.False(t, result[0].PaymentEnabled)

	mockRepo.AssertExpectations(t)
}

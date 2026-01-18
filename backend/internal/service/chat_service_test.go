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

// ==================== Mocks ====================

// MockChatRepository мок репозитория чатов
type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) GetOrCreateRoom(ctx context.Context, teacherID, studentID uuid.UUID) (*models.ChatRoom, error) {
	args := m.Called(ctx, teacherID, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChatRoom), args.Error(1)
}

func (m *MockChatRepository) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*models.ChatRoom, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChatRoom), args.Error(1)
}

func (m *MockChatRepository) ListRoomsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*models.ChatRoom, error) {
	args := m.Called(ctx, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ChatRoom), args.Error(1)
}

func (m *MockChatRepository) ListRoomsByStudent(ctx context.Context, studentID uuid.UUID) ([]*models.ChatRoom, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ChatRoom), args.Error(1)
}

func (m *MockChatRepository) CreateMessage(ctx context.Context, msg *models.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockChatRepository) UpdateMessageStatus(ctx context.Context, msgID uuid.UUID, status string) error {
	args := m.Called(ctx, msgID, status)
	return args.Error(0)
}

func (m *MockChatRepository) GetMessagesByRoom(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	args := m.Called(ctx, roomID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockChatRepository) GetMessageByID(ctx context.Context, msgID uuid.UUID) (*models.Message, error) {
	args := m.Called(ctx, msgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Message), args.Error(1)
}

func (m *MockChatRepository) GetAttachmentsByMessage(ctx context.Context, msgID uuid.UUID) ([]*models.FileAttachment, error) {
	args := m.Called(ctx, msgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FileAttachment), args.Error(1)
}

func (m *MockChatRepository) UpdateLastMessageAt(ctx context.Context, roomID uuid.UUID, messageTime time.Time) error {
	args := m.Called(ctx, roomID, messageTime)
	return args.Error(0)
}

func (m *MockChatRepository) GetPendingMessages(ctx context.Context) ([]*models.Message, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockChatRepository) CreateAttachment(ctx context.Context, att *models.FileAttachment) error {
	args := m.Called(ctx, att)
	return args.Error(0)
}

func (m *MockChatRepository) GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*models.FileAttachment, error) {
	args := m.Called(ctx, attachmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FileAttachment), args.Error(1)
}

func (m *MockChatRepository) ListAllRooms(ctx context.Context) ([]repository.ChatRoomWithDetails, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.ChatRoomWithDetails), args.Error(1)
}

func (m *MockChatRepository) SoftDeleteMessage(ctx context.Context, msgID uuid.UUID) error {
	args := m.Called(ctx, msgID)
	return args.Error(0)
}

// MockChatUserRepository мок для userRepository (только GetByID метод)
type MockChatUserRepository struct {
	mock.Mock
}

func (m *MockChatUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// ==================== Tests ====================

func TestChatService_GetOrCreateRoom(t *testing.T) {
	t.Run("success - teacher creates room with student", func(t *testing.T) {
		mockChatRepo := new(MockChatRepository)
		mockUserRepo := new(MockChatUserRepository)
		service := NewChatService(mockChatRepo, mockUserRepo, nil)

		teacherID := uuid.New()
		studentID := uuid.New()

		teacher := &models.User{
			ID:   teacherID,
			Role: models.RoleMethodologist,
		}

		student := &models.User{
			ID:   studentID,
			Role: models.RoleStudent,
		}

		expectedRoom := &models.ChatRoom{
			ID:        uuid.New(),
			TeacherID: teacherID,
			StudentID: studentID,
		}

		mockUserRepo.On("GetByID", mock.Anything, teacherID).Return(teacher, nil)
		mockUserRepo.On("GetByID", mock.Anything, studentID).Return(student, nil)
		mockChatRepo.On("GetOrCreateRoom", mock.Anything, teacherID, studentID).Return(expectedRoom, nil)

		room, err := service.GetOrCreateRoom(context.Background(), teacherID, studentID)

		assert.NoError(t, err)
		assert.NotNil(t, room)
		assert.Equal(t, expectedRoom.ID, room.ID)
		assert.Equal(t, teacherID, room.TeacherID)
		assert.Equal(t, studentID, room.StudentID)

		mockUserRepo.AssertExpectations(t)
		mockChatRepo.AssertExpectations(t)
	})

	t.Run("success - student creates room with teacher", func(t *testing.T) {
		mockChatRepo := new(MockChatRepository)
		mockUserRepo := new(MockChatUserRepository)
		service := NewChatService(mockChatRepo, mockUserRepo, nil)

		studentID := uuid.New()
		teacherID := uuid.New()

		student := &models.User{
			ID:   studentID,
			Role: models.RoleStudent,
		}

		teacher := &models.User{
			ID:   teacherID,
			Role: models.RoleMethodologist,
		}

		expectedRoom := &models.ChatRoom{
			ID:        uuid.New(),
			TeacherID: teacherID,
			StudentID: studentID,
		}

		mockUserRepo.On("GetByID", mock.Anything, studentID).Return(student, nil)
		mockUserRepo.On("GetByID", mock.Anything, teacherID).Return(teacher, nil)
		mockChatRepo.On("GetOrCreateRoom", mock.Anything, teacherID, studentID).Return(expectedRoom, nil)

		room, err := service.GetOrCreateRoom(context.Background(), studentID, teacherID)

		assert.NoError(t, err)
		assert.NotNil(t, room)
		assert.Equal(t, teacherID, room.TeacherID)
		assert.Equal(t, studentID, room.StudentID)

		mockUserRepo.AssertExpectations(t)
		mockChatRepo.AssertExpectations(t)
	})

	t.Run("error - cannot chat with self", func(t *testing.T) {
		mockChatRepo := new(MockChatRepository)
		mockUserRepo := new(MockChatUserRepository)
		service := NewChatService(mockChatRepo, mockUserRepo, nil)

		userID := uuid.New()

		room, err := service.GetOrCreateRoom(context.Background(), userID, userID)

		assert.Error(t, err)
		assert.Nil(t, room)
		assert.Equal(t, models.ErrCannotChatWithSelf, err)

		mockUserRepo.AssertNotCalled(t, "GetByID")
		mockChatRepo.AssertNotCalled(t, "GetOrCreateRoom")
	})

	t.Run("error - students cannot chat with each other", func(t *testing.T) {
		mockChatRepo := new(MockChatRepository)
		mockUserRepo := new(MockChatUserRepository)
		service := NewChatService(mockChatRepo, mockUserRepo, nil)

		student1ID := uuid.New()
		student2ID := uuid.New()

		student1 := &models.User{
			ID:   student1ID,
			Role: models.RoleStudent,
		}

		student2 := &models.User{
			ID:   student2ID,
			Role: models.RoleStudent,
		}

		mockUserRepo.On("GetByID", mock.Anything, student1ID).Return(student1, nil)
		mockUserRepo.On("GetByID", mock.Anything, student2ID).Return(student2, nil)

		room, err := service.GetOrCreateRoom(context.Background(), student1ID, student2ID)

		assert.Error(t, err)
		assert.Nil(t, room)
		assert.Contains(t, err.Error(), "students cannot chat with each other")

		mockUserRepo.AssertExpectations(t)
		mockChatRepo.AssertNotCalled(t, "GetOrCreateRoom")
	})
}

// См. BACKLOG.md: Исправить async moderation test
// TestChatService_SendMessage skipped - race condition с async moderation

func TestChatService_GetChatHistory(t *testing.T) {
	t.Run("success - participant gets messages", func(t *testing.T) {
		mockChatRepo := new(MockChatRepository)
		mockUserRepo := new(MockChatUserRepository)
		service := NewChatService(mockChatRepo, mockUserRepo, nil)

		roomID := uuid.New()
		userID := uuid.New()
		otherUserID := uuid.New()

		room := &models.ChatRoom{
			ID:        roomID,
			TeacherID: userID,
			StudentID: otherUserID,
		}

		expectedMessages := []*models.Message{
			{
				ID:          uuid.New(),
				RoomID:      roomID,
				SenderID:    userID,
				MessageText: "Message 1",
				Status:      string(models.MessageStatusDelivered),
			},
			{
				ID:          uuid.New(),
				RoomID:      roomID,
				SenderID:    otherUserID,
				MessageText: "Message 2",
				Status:      string(models.MessageStatusDelivered),
			},
		}

		req := &models.GetMessagesRequest{
			RoomID: roomID,
			Limit:  50,
			Offset: 0,
		}

		mockChatRepo.On("GetRoomByID", mock.Anything, roomID).Return(room, nil)
		mockChatRepo.On("GetMessagesByRoom", mock.Anything, roomID, 50, 0).Return(expectedMessages, nil)
		mockChatRepo.On("GetAttachmentsByMessage", mock.Anything, mock.Anything).Return([]*models.FileAttachment{}, nil)

		messages, err := service.GetChatHistory(context.Background(), userID, string(models.RoleMethodologist), req)

		assert.NoError(t, err)
		assert.NotNil(t, messages)
		assert.Len(t, messages, 2)

		mockChatRepo.AssertExpectations(t)
	})

	t.Run("error - non-participant tries to get messages", func(t *testing.T) {
		mockChatRepo := new(MockChatRepository)
		mockUserRepo := new(MockChatUserRepository)
		service := NewChatService(mockChatRepo, mockUserRepo, nil)

		roomID := uuid.New()
		teacherID := uuid.New()
		studentID := uuid.New()
		intruderID := uuid.New()

		room := &models.ChatRoom{
			ID:        roomID,
			TeacherID: teacherID,
			StudentID: studentID,
		}

		req := &models.GetMessagesRequest{
			RoomID: roomID,
			Limit:  50,
			Offset: 0,
		}

		mockChatRepo.On("GetRoomByID", mock.Anything, roomID).Return(room, nil)

		messages, err := service.GetChatHistory(context.Background(), intruderID, string(models.RoleStudent), req)

		assert.Error(t, err)
		assert.Nil(t, messages)
		assert.Equal(t, repository.ErrUnauthorized, err)

		mockChatRepo.AssertExpectations(t)
		mockChatRepo.AssertNotCalled(t, "GetMessagesByRoom")
	})

	t.Run("success - admin can read any chat", func(t *testing.T) {
		mockChatRepo := new(MockChatRepository)
		mockUserRepo := new(MockChatUserRepository)
		service := NewChatService(mockChatRepo, mockUserRepo, nil)

		roomID := uuid.New()
		teacherID := uuid.New()
		studentID := uuid.New()
		adminID := uuid.New()

		room := &models.ChatRoom{
			ID:        roomID,
			TeacherID: teacherID,
			StudentID: studentID,
		}

		expectedMessages := []*models.Message{
			{
				ID:          uuid.New(),
				RoomID:      roomID,
				SenderID:    teacherID,
				MessageText: "Message from teacher",
				Status:      string(models.MessageStatusDelivered),
			},
		}

		req := &models.GetMessagesRequest{
			RoomID: roomID,
			Limit:  50,
			Offset: 0,
		}

		mockChatRepo.On("GetRoomByID", mock.Anything, roomID).Return(room, nil)
		mockChatRepo.On("GetMessagesByRoom", mock.Anything, roomID, 50, 0).Return(expectedMessages, nil)
		mockChatRepo.On("GetAttachmentsByMessage", mock.Anything, mock.Anything).Return([]*models.FileAttachment{}, nil)

		messages, err := service.GetChatHistory(context.Background(), adminID, string(models.RoleAdmin), req)

		assert.NoError(t, err)
		assert.NotNil(t, messages)
		assert.Len(t, messages, 1)

		mockChatRepo.AssertExpectations(t)
	})
}

func TestChatService_GetUserChats(t *testing.T) {
	t.Run("success - teacher gets rooms", func(t *testing.T) {
		mockChatRepo := new(MockChatRepository)
		mockUserRepo := new(MockChatUserRepository)
		service := NewChatService(mockChatRepo, mockUserRepo, nil)

		teacherID := uuid.New()

		expectedRooms := []*models.ChatRoom{
			{
				ID:        uuid.New(),
				TeacherID: teacherID,
				StudentID: uuid.New(),
			},
			{
				ID:        uuid.New(),
				TeacherID: teacherID,
				StudentID: uuid.New(),
			},
		}

		mockChatRepo.On("ListRoomsByTeacher", mock.Anything, teacherID).Return(expectedRooms, nil)

		rooms, err := service.GetUserChats(context.Background(), teacherID, string(models.RoleMethodologist))

		assert.NoError(t, err)
		assert.NotNil(t, rooms)
		assert.Len(t, rooms, 2)

		mockChatRepo.AssertExpectations(t)
	})

	t.Run("success - student gets rooms", func(t *testing.T) {
		mockChatRepo := new(MockChatRepository)
		mockUserRepo := new(MockChatUserRepository)
		service := NewChatService(mockChatRepo, mockUserRepo, nil)

		studentID := uuid.New()

		expectedRooms := []*models.ChatRoom{
			{
				ID:        uuid.New(),
				TeacherID: uuid.New(),
				StudentID: studentID,
			},
		}

		mockChatRepo.On("ListRoomsByStudent", mock.Anything, studentID).Return(expectedRooms, nil)

		rooms, err := service.GetUserChats(context.Background(), studentID, string(models.RoleStudent))

		assert.NoError(t, err)
		assert.NotNil(t, rooms)
		assert.Len(t, rooms, 1)

		mockChatRepo.AssertExpectations(t)
	})
}

// TestGetMessageWithAttachments проверяет получение сообщения с вложениями
func TestGetMessageWithAttachments(t *testing.T) {
	ctx := context.Background()

	// Mock репозитории
	chatRepo := new(MockChatRepository)
	userRepo := new(MockUserRepository)
	chatService := NewChatService(chatRepo, userRepo, nil)

	// Данные для теста
	teacherID := uuid.New()
	studentID := uuid.New()
	roomID := uuid.New()
	messageID := uuid.New()

	room := &models.ChatRoom{
		ID:        roomID,
		TeacherID: teacherID,
		StudentID: studentID,
	}

	message := &models.Message{
		ID:          messageID,
		RoomID:      roomID,
		SenderID:    studentID,
		MessageText: "Тестовое сообщение",
		Status:      string(models.MessageStatusDelivered),
		CreatedAt:   time.Now(),
	}

	attachments := []*models.FileAttachment{
		{
			ID:        uuid.New(),
			MessageID: messageID,
			FileName:  "test.txt",
			FilePath:  "/tmp/test.txt",
			FileSize:  1024,
			MimeType:  "text/plain",
		},
	}

	// Настраиваем моки
	chatRepo.On("GetMessageByID", ctx, messageID).Return(message, nil)
	chatRepo.On("GetRoomByID", ctx, roomID).Return(room, nil)
	chatRepo.On("GetAttachmentsByMessage", ctx, messageID).Return(attachments, nil)

	// Тест: студент получает свое сообщение
	result, err := chatService.GetMessageWithAttachments(ctx, studentID, messageID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, messageID, result.ID)
	assert.Equal(t, "Тестовое сообщение", result.MessageText)
	assert.Len(t, result.Attachments, 1)
	assert.Equal(t, "test.txt", result.Attachments[0].FileName)

	chatRepo.AssertExpectations(t)
}

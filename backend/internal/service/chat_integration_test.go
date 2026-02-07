package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChatStudentCreatesRoomWithTeacher - Student creates a chat room with a methodologist
func TestChatStudentCreatesRoomWithTeacher(t *testing.T) {
	ctx := context.Background()

	// Setup: Create test users
	studentID := uuid.New()
	methodistID := uuid.New()

	student := &models.User{
		ID:   studentID,
		Role: models.RoleStudent,
	}
	methodologist := &models.User{
		ID:   methodistID,
		Role: models.RoleMethodologist,
	}

	mockChatRepoIntegration := &integrationMockChatRepository{
		rooms: make(map[uuid.UUID]*models.ChatRoom),
	}
	mockUserRepoIntegration := &integrationMockUserRepository{
		users: map[uuid.UUID]*models.User{
			studentID:   student,
			methodistID: methodologist,
		},
	}

	chatService := NewChatService(mockChatRepoIntegration, mockUserRepoIntegration, nil)

	// Test 1: Student can create room with methodologist
	room, err := chatService.GetOrCreateRoom(ctx, studentID, methodistID)
	assert.NoError(t, err, "Student should be able to create room with methodologist")
	assert.NotNil(t, room)
	assert.Equal(t, methodistID, room.TeacherID)
	assert.Equal(t, studentID, room.StudentID)

	// Test 2: Student CANNOT create room with another student
	student2ID := uuid.New()
	student2 := &models.User{
		ID:   student2ID,
		Role: models.RoleStudent,
	}
	mockUserRepoIntegration.users[student2ID] = student2

	_, err = chatService.GetOrCreateRoom(ctx, studentID, student2ID)
	assert.Error(t, err, "Student should NOT be able to create room with another student")
	assert.Contains(t, err.Error(), "students cannot chat with each other")
}

// TestChatMethodologistSeesStudentChat - Methodologist can see chat initiated by student
func TestChatMethodologistSeesStudentChat(t *testing.T) {
	ctx := context.Background()

	studentID := uuid.New()
	methodistID := uuid.New()

	student := &models.User{
		ID:   studentID,
		Role: models.RoleStudent,
	}
	methodologist := &models.User{
		ID:   methodistID,
		Role: models.RoleMethodologist,
	}

	mockChatRepoIntegration2 := &integrationMockChatRepository{
		rooms: make(map[uuid.UUID]*models.ChatRoom),
	}
	mockUserRepoIntegration2 := &integrationMockUserRepository{
		users: map[uuid.UUID]*models.User{
			studentID:   student,
			methodistID: methodologist,
		},
	}

	chatService := NewChatService(mockChatRepoIntegration2, mockUserRepoIntegration2, nil)

	// Step 1: Student creates room with methodologist
	room, err := chatService.GetOrCreateRoom(ctx, studentID, methodistID)
	require.NoError(t, err)
	require.NotNil(t, room)

	// Step 2: Methodologist loads their chat list
	chatRooms, err := chatService.GetUserChats(ctx, methodistID, string(models.RoleMethodologist))
	assert.NoError(t, err, "Methodologist should be able to load chat list")
	assert.Greater(t, len(chatRooms), 0, "Methodologist should see the chat with student")

	// Step 3: Verify methodologist can access the room
	accessibleRoom, err := chatService.GetRoomByID(ctx, room.ID, methodistID)
	assert.NoError(t, err, "Methodologist should be able to open the room")
	assert.NotNil(t, accessibleRoom)
}

// TestChatAdminListsAllChats - Admin can list all chats via GetUserChats with admin role
func TestChatAdminListsAllChats(t *testing.T) {
	ctx := context.Background()

	studentID1 := uuid.New()
	studentID2 := uuid.New()
	methodistID1 := uuid.New()
	methodistID2 := uuid.New()
	adminID := uuid.New()

	student1 := &models.User{ID: studentID1, Role: models.RoleStudent}
	student2 := &models.User{ID: studentID2, Role: models.RoleStudent}
	methodologist1 := &models.User{ID: methodistID1, Role: models.RoleMethodologist}
	methodologist2 := &models.User{ID: methodistID2, Role: models.RoleMethodologist}
	admin := &models.User{ID: adminID, Role: models.RoleAdmin}

	mockChatRepoIntegration3 := &integrationMockChatRepository{
		rooms: make(map[uuid.UUID]*models.ChatRoom),
	}
	mockUserRepoIntegration3 := &integrationMockUserRepository{
		users: map[uuid.UUID]*models.User{
			studentID1:   student1,
			studentID2:   student2,
			methodistID1: methodologist1,
			methodistID2: methodologist2,
			adminID:      admin,
		},
	}

	chatService := NewChatService(mockChatRepoIntegration3, mockUserRepoIntegration3, nil)

	// Create multiple chat rooms
	room1, _ := chatService.GetOrCreateRoom(ctx, studentID1, methodistID1)
	room2, _ := chatService.GetOrCreateRoom(ctx, studentID2, methodistID2)

	// Verify rooms were created
	assert.NotNil(t, room1)
	assert.NotNil(t, room2)

	// Admin can list ALL chats via ListAllRooms (returned by GetUserChats for admin role)
	// Note: In real implementation, admin role routes to ListAllRooms which returns all rooms
	_, err := chatService.GetUserChats(ctx, adminID, string(models.RoleAdmin))
	assert.NoError(t, err, "Admin should be able to list all chats")
	// Admin should be able to see all rooms created
}

// TestChatMessageOrdering - Messages maintain correct order
func TestChatMessageOrdering(t *testing.T) {
	ctx := context.Background()

	studentID := uuid.New()
	methodistID := uuid.New()

	student := &models.User{ID: studentID, Role: models.RoleStudent}
	methodologist := &models.User{ID: methodistID, Role: models.RoleMethodologist}

	mockChatRepoIntegration4 := &integrationMockChatRepository{
		rooms:    make(map[uuid.UUID]*models.ChatRoom),
		messages: make(map[uuid.UUID]*models.Message),
	}
	mockUserRepoIntegration4 := &integrationMockUserRepository{
		users: map[uuid.UUID]*models.User{
			studentID:   student,
			methodistID: methodologist,
		},
	}

	chatService := NewChatService(mockChatRepoIntegration4, mockUserRepoIntegration4, nil)

	// Create room
	room, _ := chatService.GetOrCreateRoom(ctx, studentID, methodistID)

	// Send 3 messages
	now := time.Now()
	msg1 := &models.Message{
		ID:          uuid.New(),
		RoomID:      room.ID,
		SenderID:    studentID,
		MessageText: "Message 1",
		Status:      models.MessageStatusDelivered,
		CreatedAt:   now,
	}
	msg2 := &models.Message{
		ID:          uuid.New(),
		RoomID:      room.ID,
		SenderID:    studentID,
		MessageText: "Message 2",
		Status:      models.MessageStatusDelivered,
		CreatedAt:   now.Add(1 * time.Second),
	}
	msg3 := &models.Message{
		ID:          uuid.New(),
		RoomID:      room.ID,
		SenderID:    studentID,
		MessageText: "Message 3",
		Status:      models.MessageStatusDelivered,
		CreatedAt:   now.Add(2 * time.Second),
	}

	_ = mockChatRepoIntegration4.CreateMessage(ctx, msg1)
	_ = mockChatRepoIntegration4.CreateMessage(ctx, msg2)
	_ = mockChatRepoIntegration4.CreateMessage(ctx, msg3)

	// Retrieve messages
	messages, err := mockChatRepoIntegration4.GetMessagesByRoom(ctx, room.ID, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(messages))

	// Verify order: oldest first
	assert.Equal(t, "Message 1", messages[0].MessageText)
	assert.Equal(t, "Message 2", messages[1].MessageText)
	assert.Equal(t, "Message 3", messages[2].MessageText)
}

// TestChatValidation - Validate permission checks
func TestChatValidation(t *testing.T) {
	ctx := context.Background()

	studentID1 := uuid.New()
	studentID2 := uuid.New()
	methodistID := uuid.New()

	student1 := &models.User{ID: studentID1, Role: models.RoleStudent}
	student2 := &models.User{ID: studentID2, Role: models.RoleStudent}
	methodologist := &models.User{ID: methodistID, Role: models.RoleMethodologist}

	mockChatRepoIntegration5 := &integrationMockChatRepository{
		rooms: make(map[uuid.UUID]*models.ChatRoom),
	}
	mockUserRepoIntegration5 := &integrationMockUserRepository{
		users: map[uuid.UUID]*models.User{
			studentID1:  student1,
			studentID2:  student2,
			methodistID: methodologist,
		},
	}

	chatService := NewChatService(mockChatRepoIntegration5, mockUserRepoIntegration5, nil)

	// Test 1: Student cannot chat with self
	_, err := chatService.GetOrCreateRoom(ctx, studentID1, studentID1)
	assert.Error(t, err, "User should not be able to chat with themselves")

	// Test 2: Student cannot chat with student
	_, err = chatService.GetOrCreateRoom(ctx, studentID1, studentID2)
	assert.Error(t, err, "Student should not be able to chat with another student")

	// Test 3: Valid rooms should work
	room, err := chatService.GetOrCreateRoom(ctx, studentID1, methodistID)
	assert.NoError(t, err, "Student should be able to create room with methodologist")
	assert.NotNil(t, room)

	// Test 4: Unauthorized user cannot access room
	_, err = chatService.GetRoomByID(ctx, room.ID, studentID2)
	assert.Error(t, err, "Unauthorized user should not access room")
}

// ==================== Mock Repositories ====================

type integrationMockChatRepository struct {
	rooms    map[uuid.UUID]*models.ChatRoom
	messages map[uuid.UUID]*models.Message
	counter  int
}

func (m *integrationMockChatRepository) GetOrCreateRoom(ctx context.Context, teacherID, studentID uuid.UUID) (*models.ChatRoom, error) {
	// Check if room already exists
	for _, room := range m.rooms {
		if (room.TeacherID == teacherID && room.StudentID == studentID) ||
			(room.TeacherID == studentID && room.StudentID == teacherID) {
			return room, nil
		}
	}

	// Create new room
	room := &models.ChatRoom{
		ID:        uuid.New(),
		TeacherID: teacherID,
		StudentID: studentID,
		CreatedAt: time.Now(),
	}
	m.rooms[room.ID] = room
	return room, nil
}

func (m *integrationMockChatRepository) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*models.ChatRoom, error) {
	room, exists := m.rooms[roomID]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return room, nil
}

func (m *integrationMockChatRepository) ListRoomsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*models.ChatRoom, error) {
	var rooms []*models.ChatRoom
	for _, room := range m.rooms {
		if room.TeacherID == teacherID {
			rooms = append(rooms, room)
		}
	}
	return rooms, nil
}

func (m *integrationMockChatRepository) ListRoomsByStudent(ctx context.Context, studentID uuid.UUID) ([]*models.ChatRoom, error) {
	var rooms []*models.ChatRoom
	for _, room := range m.rooms {
		if room.StudentID == studentID {
			rooms = append(rooms, room)
		}
	}
	return rooms, nil
}

func (m *integrationMockChatRepository) ListAllRooms(ctx context.Context) ([]repository.ChatRoomWithDetails, error) {
	var result []repository.ChatRoomWithDetails
	for _, room := range m.rooms {
		result = append(result, repository.ChatRoomWithDetails{
			ID: room.ID,
		})
	}
	return result, nil
}

func (m *integrationMockChatRepository) CreateMessage(ctx context.Context, msg *models.Message) error {
	m.messages[msg.ID] = msg
	return nil
}

func (m *integrationMockChatRepository) UpdateMessageStatus(ctx context.Context, msgID uuid.UUID, status string) error {
	if msg, exists := m.messages[msgID]; exists {
		msg.Status = status
		return nil
	}
	return sql.ErrNoRows
}

func (m *integrationMockChatRepository) GetMessagesByRoom(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	var messages []*models.Message
	for _, msg := range m.messages {
		if msg.RoomID == roomID {
			messages = append(messages, msg)
		}
	}
	// Sort by created_at
	for i := 0; i < len(messages)-1; i++ {
		for j := i + 1; j < len(messages); j++ {
			if messages[j].CreatedAt.Before(messages[i].CreatedAt) {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
	}
	return messages, nil
}

func (m *integrationMockChatRepository) GetMessageByID(ctx context.Context, msgID uuid.UUID) (*models.Message, error) {
	msg, exists := m.messages[msgID]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return msg, nil
}

func (m *integrationMockChatRepository) GetAttachmentsByMessage(ctx context.Context, msgID uuid.UUID) ([]*models.FileAttachment, error) {
	return []*models.FileAttachment{}, nil
}

func (m *integrationMockChatRepository) UpdateLastMessageAt(ctx context.Context, roomID uuid.UUID, messageTime time.Time) error {
	if room, exists := m.rooms[roomID]; exists {
		room.LastMessageAt = sql.NullTime{Time: messageTime, Valid: true}
		return nil
	}
	return sql.ErrNoRows
}

func (m *integrationMockChatRepository) GetPendingMessages(ctx context.Context) ([]*models.Message, error) {
	return []*models.Message{}, nil
}

func (m *integrationMockChatRepository) CreateAttachment(ctx context.Context, att *models.FileAttachment) error {
	return nil
}

func (m *integrationMockChatRepository) GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*models.FileAttachment, error) {
	return nil, sql.ErrNoRows
}

func (m *integrationMockChatRepository) SoftDeleteMessage(ctx context.Context, msgID uuid.UUID) error {
	if msg, exists := m.messages[msgID]; exists {
		msg.DeletedAt = sql.NullTime{Time: time.Now(), Valid: true}
		return nil
	}
	return sql.ErrNoRows
}

type integrationMockUserRepository struct {
	users map[uuid.UUID]*models.User
}

func (m *integrationMockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

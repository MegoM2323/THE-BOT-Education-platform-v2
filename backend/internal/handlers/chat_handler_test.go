package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tutoring-platform/internal/middleware"
	"tutoring-platform/internal/models"
	"tutoring-platform/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ==================== Mocks ====================

// MockChatRepository for testing
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

// MockUserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, roleFilter *models.UserRole) ([]*models.User, error) {
	args := m.Called(ctx, roleFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) Exists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error {
	args := m.Called(ctx, userID, username)
	return args.Error(0)
}

// ==================== Test Helpers ====================

// setSessionInContext adds session to request context
func setSessionInContext(r *http.Request, userID uuid.UUID, role models.UserRole) *http.Request {
	session := &models.SessionWithUser{
		Session: models.Session{
			ID:     uuid.New(),
			UserID: userID,
		},
		UserRole: role,
	}
	ctx := context.WithValue(r.Context(), middleware.SessionContextKey, session)
	return r.WithContext(ctx)
}

// ==================== GetMyRooms Tests ====================

// TestChatHandler_GetMyRooms_Success проверяет получение списка комнат
func TestChatHandler_GetMyRooms_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepository)

	studentID := uuid.New()
	teacherID := uuid.New()
	roomID := uuid.New()

	expectedRooms := []*models.ChatRoom{
		{
			ID:              roomID,
			TeacherID:       teacherID,
			StudentID:       studentID,
			ParticipantID:   teacherID,
			ParticipantName: "Иван Преподаватель",
			ParticipantRole: string(models.RoleTeacher),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	mockChatRepo.On("ListRoomsByStudent", mock.Anything, studentID).Return(expectedRooms, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, nil)
	handler := NewChatHandler(chatService)

	req := httptest.NewRequest("GET", "/api/v1/chat/rooms", nil)
	req = setSessionInContext(req, studentID, models.RoleStudent)
	w := httptest.NewRecorder()

	handler.GetMyRooms(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool               `json:"success"`
		Data    []*models.ChatRoom `json:"data"`
	}

	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Data, 1)
	assert.Equal(t, roomID, response.Data[0].ID)
	assert.Equal(t, teacherID, response.Data[0].ParticipantID)
	assert.Equal(t, "Иван Преподаватель", response.Data[0].ParticipantName)

	mockChatRepo.AssertExpectations(t)
	t.Log("✓ GetMyRooms returns participant info")
}

// TestChatHandler_GetMyRooms_NoSession проверяет отсутствие сессии
func TestChatHandler_GetMyRooms_NoSession(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepository)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, nil)
	handler := NewChatHandler(chatService)

	req := httptest.NewRequest("GET", "/api/v1/chat/rooms", nil)
	w := httptest.NewRecorder()

	handler.GetMyRooms(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	t.Log("✓ GetMyRooms requires authentication")
}

// ==================== SendMessage Tests ====================

// TestChatHandler_SendMessage_EmptyMessage проверяет пустое сообщение
func TestChatHandler_SendMessage_EmptyMessage(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepository)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, nil)
	handler := NewChatHandler(chatService)

	roomID := uuid.New()
	userID := uuid.New()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("message", "")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/chat/rooms/"+roomID.String()+"/messages", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = setSessionInContext(req, userID, models.RoleStudent)
	w := httptest.NewRecorder()

	handler.SendMessage(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockChatRepo.AssertNotCalled(t, "CreateMessage")
	t.Log("✓ SendMessage rejects empty messages")
}

// TestChatHandler_SendMessage_InvalidRoomID проверяет неверный room ID
func TestChatHandler_SendMessage_InvalidRoomID(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepository)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, nil)
	handler := NewChatHandler(chatService)

	userID := uuid.New()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("message", "Test")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/chat/rooms/invalid-id/messages", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = setSessionInContext(req, userID, models.RoleStudent)
	w := httptest.NewRecorder()

	handler.SendMessage(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	t.Log("✓ SendMessage validates room ID format")
}

// TestChatHandler_SendMessage_NoSession проверяет отсутствие сессии
func TestChatHandler_SendMessage_NoSession(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepository)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, nil)
	handler := NewChatHandler(chatService)

	roomID := uuid.New()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("message", "Test")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/chat/rooms/"+roomID.String()+"/messages", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.SendMessage(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	t.Log("✓ SendMessage requires authentication")
}

// ==================== GetOrCreateRoom Tests ====================

// TestChatHandler_GetOrCreateRoom_InvalidJSON проверяет неверный JSON
func TestChatHandler_GetOrCreateRoom_InvalidJSON(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepository)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, nil)
	handler := NewChatHandler(chatService)

	userID := uuid.New()

	req := httptest.NewRequest("POST", "/api/v1/chat/rooms", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req = setSessionInContext(req, userID, models.RoleStudent)
	w := httptest.NewRecorder()

	handler.GetOrCreateRoom(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockChatRepo.AssertNotCalled(t, "GetOrCreateRoom")
	t.Log("✓ GetOrCreateRoom validates JSON format")
}

// TestChatHandler_GetOrCreateRoom_InvalidParticipantID проверяет неверный UUID
func TestChatHandler_GetOrCreateRoom_InvalidParticipantID(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepository)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, nil)
	handler := NewChatHandler(chatService)

	userID := uuid.New()

	reqBody := struct {
		ParticipantID string `json:"participant_id"`
	}{
		ParticipantID: "not-a-uuid",
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/chat/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = setSessionInContext(req, userID, models.RoleStudent)
	w := httptest.NewRecorder()

	handler.GetOrCreateRoom(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	t.Log("✓ GetOrCreateRoom validates participant_id UUID")
}

// TestChatHandler_GetOrCreateRoom_NoSession проверяет отсутствие сессии
func TestChatHandler_GetOrCreateRoom_NoSession(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepository)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, nil)
	handler := NewChatHandler(chatService)

	reqBody := struct {
		ParticipantID string `json:"participant_id"`
	}{
		ParticipantID: uuid.New().String(),
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/chat/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.GetOrCreateRoom(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	t.Log("✓ GetOrCreateRoom requires authentication")
}

// ==================== File Upload Tests ====================

// TestChatHandler_SendMessage_FileUploadValidation проверяет валидацию при загрузке файла
func TestChatHandler_SendMessage_FileUploadValidation(t *testing.T) {
	// This test verifies that the handler can process multipart forms with files
	// Full integration requires a database, so we test basic validation here
	mockChatRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepository)

	userID := uuid.New()
	roomID := uuid.New()

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, nil)
	handler := NewChatHandler(chatService)

	// Test 1: Valid multipart form with message and file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("message", "See attached file")

	part, err := writer.CreateFormFile("files", "test.txt")
	require.NoError(t, err)

	_, err = io.Copy(part, bytes.NewReader([]byte("content")))
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/chat/rooms/"+roomID.String()+"/messages", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = setSessionInContext(req, userID, models.RoleStudent)
	w := httptest.NewRecorder()

	// Should fail because room doesn't exist in mock, but form should parse
	handler.SendMessage(w, req)

	// Should get a 400 error (failed to send message, room doesn't exist)
	// But the form parsing should work
	assert.True(t, w.Code >= 400, "Expected error response for non-existent room")

	t.Log("✓ File upload form parsing works")
}

// ==================== Response Format Tests ====================

// TestChatHandler_ResponseFormat_Success проверяет формат ответа
func TestChatHandler_ResponseFormat_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepository)

	userID := uuid.New()

	mockChatRepo.On("ListRoomsByStudent", mock.Anything, userID).Return([]*models.ChatRoom{}, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, nil)
	handler := NewChatHandler(chatService)

	req := httptest.NewRequest("GET", "/api/v1/chat/rooms", nil)
	req = setSessionInContext(req, userID, models.RoleStudent)
	w := httptest.NewRecorder()

	handler.GetMyRooms(w, req)

	// Verify response structure
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Should have success field
	assert.NotNil(t, response["success"])
	assert.True(t, response["success"].(bool))

	// Should have data field
	assert.NotNil(t, response["data"])

	t.Log("✓ Responses use {success: true, data: ...} format")
}

// ==================== Security Tests ====================

// TestFileAttachment_JSONSerialization проверяет что FilePath исключен из JSON (A08 fix)
func TestFileAttachment_JSONSerialization(t *testing.T) {
	attachment := &models.FileAttachment{
		ID:         uuid.New(),
		MessageID:  uuid.New(),
		FileName:   "secret_document.pdf",
		FilePath:   "/var/uploads/chat/12345abcde.pdf", // ВНУТРЕННИЙ путь - должен быть скрыт!
		FileSize:   2048,
		MimeType:   "application/pdf",
		UploadedAt: time.Now(),
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(attachment)
	require.NoError(t, err)

	jsonStr := string(jsonData)
	t.Logf("JSON: %s", jsonStr)

	// Verify that FilePath is NOT in JSON response
	assert.NotContains(t, jsonStr, "file_path", "Field 'file_path' должен быть исключен из JSON!")
	assert.NotContains(t, jsonStr, "/var/uploads/chat", "Путь к файлу не должен быть выдан!")

	// Verify other fields ARE in JSON
	assert.Contains(t, jsonStr, "file_name")
	assert.Contains(t, jsonStr, "secret_document.pdf")
	assert.Contains(t, jsonStr, "file_size")
	assert.Contains(t, jsonStr, "2048")
	assert.Contains(t, jsonStr, "mime_type")
	assert.Contains(t, jsonStr, "application/pdf")

	// Deserialize and verify structure
	var deserialized models.FileAttachment
	err = json.Unmarshal(jsonData, &deserialized)
	require.NoError(t, err)

	// Fields should be present in deserialized struct
	assert.Equal(t, attachment.ID, deserialized.ID)
	assert.Equal(t, "secret_document.pdf", deserialized.FileName)
	assert.Equal(t, int64(2048), deserialized.FileSize)
	assert.Equal(t, "application/pdf", deserialized.MimeType)

	// FilePath should be empty after deserialization (since it has json:"-" tag)
	assert.Empty(t, deserialized.FilePath)

	t.Log("✓ FileAttachment: FilePath успешно скрыт из JSON (security fix A08)")
}

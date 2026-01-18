package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChatRepo - мок для ChatRepository
type MockChatRepo struct {
	mock.Mock
}

func (m *MockChatRepo) CreateAttachment(ctx context.Context, att *models.FileAttachment) error {
	args := m.Called(ctx, att)
	return args.Error(0)
}

func (m *MockChatRepo) GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*models.FileAttachment, error) {
	args := m.Called(ctx, attachmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FileAttachment), args.Error(1)
}

// Реализация остальных методов интерфейса (не используются в этих тестах)
func (m *MockChatRepo) GetOrCreateRoom(ctx context.Context, teacherID, studentID uuid.UUID) (*models.ChatRoom, error) {
	return nil, nil
}
func (m *MockChatRepo) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*models.ChatRoom, error) {
	return nil, nil
}
func (m *MockChatRepo) ListRoomsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]*models.ChatRoom, error) {
	return nil, nil
}
func (m *MockChatRepo) ListRoomsByStudent(ctx context.Context, studentID uuid.UUID) ([]*models.ChatRoom, error) {
	return nil, nil
}
func (m *MockChatRepo) CreateMessage(ctx context.Context, msg *models.Message) error {
	return nil
}
func (m *MockChatRepo) UpdateMessageStatus(ctx context.Context, msgID uuid.UUID, status string) error {
	return nil
}
func (m *MockChatRepo) GetMessagesByRoom(ctx context.Context, roomID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	return nil, nil
}
func (m *MockChatRepo) GetMessageByID(ctx context.Context, msgID uuid.UUID) (*models.Message, error) {
	return nil, nil
}
func (m *MockChatRepo) GetAttachmentsByMessage(ctx context.Context, msgID uuid.UUID) ([]*models.FileAttachment, error) {
	return nil, nil
}
func (m *MockChatRepo) UpdateLastMessageAt(ctx context.Context, roomID uuid.UUID, messageTime any) error {
	return nil
}
func (m *MockChatRepo) GetPendingMessages(ctx context.Context) ([]*models.Message, error) {
	return nil, nil
}

// TestUploadFile_Success проверяет успешную загрузку файла
func TestUploadFile_Success(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir := t.TempDir()

	mockRepo := new(MockChatRepo)
	service := NewFileUploadService(mockRepo)
	service.uploadDir = tempDir

	// Создаем тестовый PDF файл (минимальный валидный PDF)
	pdfContent := []byte("%PDF-1.4\n%Test PDF\n%%EOF")
	fileReader := bytes.NewReader(pdfContent)

	messageID := uuid.New()
	roomID := uuid.New()

	// Мокируем успешное создание записи в БД
	mockRepo.On("CreateAttachment", mock.Anything, mock.MatchedBy(func(att *models.FileAttachment) bool {
		return att.MessageID == messageID && att.MimeType == "application/pdf"
	})).Return(nil)

	// Загружаем файл
	attachment, err := service.UploadFile(context.Background(), fileReader, "test.pdf", int64(len(pdfContent)), messageID, roomID)

	assert.NoError(t, err)
	assert.NotNil(t, attachment)
	assert.Equal(t, messageID, attachment.MessageID)
	assert.Equal(t, "test.pdf", attachment.FileName)
	assert.Equal(t, "application/pdf", attachment.MimeType)
	assert.Equal(t, int64(len(pdfContent)), attachment.FileSize)

	// Проверяем что файл существует на диске
	assert.FileExists(t, attachment.FilePath)

	// Проверяем содержимое файла
	savedContent, err := os.ReadFile(attachment.FilePath)
	assert.NoError(t, err)
	assert.Equal(t, pdfContent, savedContent)

	mockRepo.AssertExpectations(t)
}

// TestUploadFile_FileTooLarge проверяет валидацию размера файла
func TestUploadFile_FileTooLarge(t *testing.T) {
	tempDir := t.TempDir()

	mockRepo := new(MockChatRepo)
	service := NewFileUploadService(mockRepo)
	service.uploadDir = tempDir

	// Размер превышает максимальный
	fileSize := int64(MaxFileSize + 1)
	fileReader := bytes.NewReader([]byte("test"))

	messageID := uuid.New()
	roomID := uuid.New()

	// Загружаем файл
	_, err := service.UploadFile(context.Background(), fileReader, "large.pdf", fileSize, messageID, roomID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file too large")

	mockRepo.AssertNotCalled(t, "CreateAttachment")
}

// TestUploadFile_InvalidMimeType проверяет валидацию MIME-типа
func TestUploadFile_InvalidMimeType(t *testing.T) {
	tempDir := t.TempDir()

	mockRepo := new(MockChatRepo)
	service := NewFileUploadService(mockRepo)
	service.uploadDir = tempDir

	// Создаем файл с недопустимым MIME-типом (исполняемый файл)
	exeContent := []byte("MZ\x90\x00") // Начало PE executable
	fileReader := bytes.NewReader(exeContent)

	messageID := uuid.New()
	roomID := uuid.New()

	// Загружаем файл
	_, err := service.UploadFile(context.Background(), fileReader, "malware.exe", int64(len(exeContent)), messageID, roomID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file type")

	mockRepo.AssertNotCalled(t, "CreateAttachment")
}

// TestUploadFile_DatabaseError проверяет обработку ошибки БД
func TestUploadFile_DatabaseError(t *testing.T) {
	tempDir := t.TempDir()

	mockRepo := new(MockChatRepo)
	service := NewFileUploadService(mockRepo)
	service.uploadDir = tempDir

	pdfContent := []byte("%PDF-1.4\n%Test PDF\n%%EOF")
	fileReader := bytes.NewReader(pdfContent)

	messageID := uuid.New()
	roomID := uuid.UUID{}

	// Мокируем ошибку БД
	dbError := errors.New("database connection lost")
	mockRepo.On("CreateAttachment", mock.Anything, mock.Anything).Return(dbError)

	// Загружаем файл
	_, err := service.UploadFile(context.Background(), fileReader, "test.pdf", int64(len(pdfContent)), messageID, roomID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create attachment record")

	// Проверяем что файл был удален после ошибки БД
	filePath := filepath.Join(tempDir, roomID.String(), messageID.String())
	entries, _ := os.ReadDir(filePath)
	assert.Empty(t, entries, "File should be deleted after database error")

	mockRepo.AssertExpectations(t)
}

// TestGetFilePath_Success проверяет успешное получение пути к файлу
func TestGetFilePath_Success(t *testing.T) {
	tempDir := t.TempDir()

	mockRepo := new(MockChatRepo)
	service := NewFileUploadService(mockRepo)
	service.uploadDir = tempDir

	attachmentID := uuid.New()
	filePath := filepath.Join(tempDir, "test.pdf")

	// Создаем файл
	err := os.WriteFile(filePath, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Мокируем успешное получение из БД
	mockRepo.On("GetAttachmentByID", mock.Anything, attachmentID).Return(&models.FileAttachment{
		ID:       attachmentID,
		FilePath: filePath,
	}, nil)

	// Получаем путь
	path, err := service.GetFilePath(context.Background(), attachmentID)

	assert.NoError(t, err)
	assert.Equal(t, filePath, path)

	mockRepo.AssertExpectations(t)
}

// TestGetFilePath_FileNotFound проверяет обработку отсутствующего файла
func TestGetFilePath_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	mockRepo := new(MockChatRepo)
	service := NewFileUploadService(mockRepo)
	service.uploadDir = tempDir

	attachmentID := uuid.New()
	nonExistentPath := filepath.Join(tempDir, "nonexistent.pdf")

	// Мокируем успешное получение из БД, но файл не существует
	mockRepo.On("GetAttachmentByID", mock.Anything, attachmentID).Return(&models.FileAttachment{
		ID:       attachmentID,
		FilePath: nonExistentPath,
	}, nil)

	// Получаем путь
	_, err := service.GetFilePath(context.Background(), attachmentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found on disk")

	mockRepo.AssertExpectations(t)
}

// TestDeleteFile_Success проверяет успешное удаление файла
func TestDeleteFile_Success(t *testing.T) {
	tempDir := t.TempDir()

	mockRepo := new(MockChatRepo)
	service := NewFileUploadService(mockRepo)
	service.uploadDir = tempDir

	attachmentID := uuid.New()
	filePath := filepath.Join(tempDir, "test.pdf")

	// Создаем файл
	err := os.WriteFile(filePath, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Мокируем успешное получение из БД
	mockRepo.On("GetAttachmentByID", mock.Anything, attachmentID).Return(&models.FileAttachment{
		ID:       attachmentID,
		FilePath: filePath,
	}, nil)

	// Удаляем файл
	err = service.DeleteFile(context.Background(), attachmentID)

	assert.NoError(t, err)
	assert.NoFileExists(t, filePath)

	mockRepo.AssertExpectations(t)
}

// TestValidateFilename проверяет валидацию имен файлов
func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Valid filename",
			filename: "document.pdf",
			wantErr:  false,
		},
		{
			name:     "Valid filename with spaces",
			filename: "my document 2024.docx",
			wantErr:  false,
		},
		{
			name:     "Empty filename",
			filename: "",
			wantErr:  true,
			errMsg:   "filename is empty",
		},
		{
			name:     "Path traversal with ..",
			filename: "../../../etc/passwd",
			wantErr:  true,
			errMsg:   "contains path separators",
		},
		{
			name:     "Path with forward slash",
			filename: "path/to/file.pdf",
			wantErr:  true,
			errMsg:   "contains path separators",
		},
		{
			name:     "Path with backslash",
			filename: "path\\to\\file.pdf",
			wantErr:  true,
			errMsg:   "contains path separators",
		},
		{
			name:     "Filename too long",
			filename: string(make([]byte, 256)), // 256 символов
			wantErr:  true,
			errMsg:   "filename too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilename(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUploadFile_SupportedMimeTypes проверяет все поддерживаемые MIME-типы
func TestUploadFile_SupportedMimeTypes(t *testing.T) {
	tests := []struct {
		name         string
		content      []byte
		filename     string
		expectedType string
	}{
		{
			name:         "PDF file",
			content:      []byte("%PDF-1.4\n%%EOF"),
			filename:     "document.pdf",
			expectedType: "application/pdf",
		},
		{
			name:         "JPEG file",
			content:      []byte("\xFF\xD8\xFF\xE0\x00\x10JFIF"),
			filename:     "photo.jpg",
			expectedType: "image/jpeg",
		},
		{
			name:         "PNG file",
			content:      []byte("\x89PNG\r\n\x1a\n"),
			filename:     "image.png",
			expectedType: "image/png",
		},
		{
			name:         "GIF file",
			content:      []byte("GIF89a"),
			filename:     "animation.gif",
			expectedType: "image/gif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			mockRepo := new(MockChatRepo)
			service := NewFileUploadService(mockRepo)
			service.uploadDir = tempDir

			fileReader := bytes.NewReader(tt.content)
			messageID := uuid.New()
			roomID := uuid.New()

			mockRepo.On("CreateAttachment", mock.Anything, mock.MatchedBy(func(att *models.FileAttachment) bool {
				return att.MimeType == tt.expectedType
			})).Return(nil)

			attachment, err := service.UploadFile(context.Background(), fileReader, tt.filename, int64(len(tt.content)), messageID, roomID)

			assert.NoError(t, err)
			assert.NotNil(t, attachment)
			assert.Equal(t, tt.expectedType, attachment.MimeType)

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUploadFile_ZeroSize проверяет валидацию нулевого размера файла
func TestUploadFile_ZeroSize(t *testing.T) {
	tempDir := t.TempDir()

	mockRepo := new(MockChatRepo)
	service := NewFileUploadService(mockRepo)
	service.uploadDir = tempDir

	fileReader := bytes.NewReader([]byte{})
	messageID := uuid.New()
	roomID := uuid.New()

	_, err := service.UploadFile(context.Background(), fileReader, "empty.pdf", 0, messageID, roomID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file size must be positive")

	mockRepo.AssertNotCalled(t, "CreateAttachment")
}

// TestUploadFile_ReadError проверяет обработку ошибки чтения файла
func TestUploadFile_ReadError(t *testing.T) {
	tempDir := t.TempDir()

	mockRepo := new(MockChatRepo)
	service := NewFileUploadService(mockRepo)
	service.uploadDir = tempDir

	// Создаем reader который возвращает ошибку
	errorReader := &errorReader{err: errors.New("read error")}

	messageID := uuid.New()
	roomID := uuid.New()

	_, err := service.UploadFile(context.Background(), errorReader, "test.pdf", 100, messageID, roomID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")

	mockRepo.AssertNotCalled(t, "CreateAttachment")
}

// errorReader - вспомогательный reader который всегда возвращает ошибку
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

var _ io.Reader = (*errorReader)(nil)

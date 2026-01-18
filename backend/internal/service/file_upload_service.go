package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"tutoring-platform/internal/models"

	"github.com/google/uuid"
)

const (
	// Максимальный размер файла: 10MB
	MaxFileSize = 10 * 1024 * 1024 // 10MB

	// Базовая директория для загрузки файлов
	BaseUploadDir = "uploads/chat"

	// Размер буфера для детекции MIME-типа
	MIMEDetectionBufferSize = 512
)

// Разрешенные MIME-типы для загрузки
var AllowedMimeTypes = map[string]bool{
	"application/pdf": true, // PDF
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // DOCX
	"image/jpeg": true, // JPEG
	"image/png":  true, // PNG
	"image/gif":  true, // GIF
	"image/webp": true, // WebP
}

// fileUploadRepository - интерфейс для dependency injection
type fileUploadRepository interface {
	CreateAttachment(ctx context.Context, att *models.FileAttachment) error
	GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*models.FileAttachment, error)
}

// FileUploadService управляет загрузкой и хранением файлов
type FileUploadService struct {
	chatRepo  fileUploadRepository
	uploadDir string
}

// NewFileUploadService создает новый FileUploadService
func NewFileUploadService(chatRepo fileUploadRepository) *FileUploadService {
	return &FileUploadService{
		chatRepo:  chatRepo,
		uploadDir: BaseUploadDir,
	}
}

// UploadFile загружает файл на диск и создает запись в БД
func (s *FileUploadService) UploadFile(ctx context.Context, file io.Reader, originalFilename string, fileSize int64, messageID uuid.UUID, roomID uuid.UUID) (*models.FileAttachment, error) {
	// 1. Валидация размера файла
	if fileSize > MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d bytes)", fileSize, MaxFileSize)
	}

	if fileSize <= 0 {
		return nil, fmt.Errorf("file size must be positive")
	}

	// 2. Читаем первые 512 байт для детекции MIME-типа
	buffer := make([]byte, MIMEDetectionBufferSize)
	n, err := io.ReadAtLeast(file, buffer, MIMEDetectionBufferSize)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read file for MIME detection: %w", err)
	}

	// Детекция MIME-типа
	mimeType := http.DetectContentType(buffer[:n])

	// Валидация MIME-типа
	if !AllowedMimeTypes[mimeType] {
		return nil, fmt.Errorf("invalid file type: %s (allowed: PDF, DOCX, JPEG, PNG, GIF, WebP)", mimeType)
	}

	// 3. Создаем директорию: uploads/chat/{room_id}/{message_id}/
	dirPath := filepath.Join(s.uploadDir, roomID.String(), messageID.String())
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// 4. Генерируем безопасное имя файла (uuid + расширение)
	fileExt := filepath.Ext(originalFilename)
	if fileExt == "" {
		// Если расширения нет, пытаемся получить из MIME-типа
		exts, err := mime.ExtensionsByType(mimeType)
		if err == nil && len(exts) > 0 {
			fileExt = exts[0]
		}
	}

	safeFilename := uuid.New().String() + fileExt
	filePath := filepath.Join(dirPath, safeFilename)

	// 5. Сохраняем файл на диск
	outFile, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Сначала записываем прочитанный буфер
	if _, err := outFile.Write(buffer[:n]); err != nil {
		return nil, fmt.Errorf("failed to write buffer to file: %w", err)
	}

	// Затем копируем остаток файла
	written, err := io.Copy(outFile, file)
	if err != nil {
		// Удаляем частично записанный файл
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	totalWritten := int64(n) + written

	// Проверяем что записанный размер соответствует заявленному
	if totalWritten != fileSize {
		os.Remove(filePath)
		return nil, fmt.Errorf("file size mismatch: expected %d, written %d", fileSize, totalWritten)
	}

	// 6. Создаем запись в БД
	attachment := &models.FileAttachment{
		MessageID: messageID,
		FileName:  originalFilename,
		FilePath:  filePath,
		FileSize:  fileSize,
		MimeType:  mimeType,
	}

	if err := s.chatRepo.CreateAttachment(ctx, attachment); err != nil {
		// Удаляем файл если не удалось создать запись в БД
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to create attachment record: %w", err)
	}

	return attachment, nil
}

// GetFilePath возвращает путь к файлу по ID вложения
func (s *FileUploadService) GetFilePath(ctx context.Context, attachmentID uuid.UUID) (string, error) {
	attachment, err := s.chatRepo.GetAttachmentByID(ctx, attachmentID)
	if err != nil {
		return "", fmt.Errorf("failed to get attachment: %w", err)
	}

	// Проверяем что файл существует
	if _, err := os.Stat(attachment.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found on disk: %s", attachment.FilePath)
	}

	return attachment.FilePath, nil
}

// GetAttachment возвращает информацию о вложении
func (s *FileUploadService) GetAttachment(ctx context.Context, attachmentID uuid.UUID) (*models.FileAttachment, error) {
	return s.chatRepo.GetAttachmentByID(ctx, attachmentID)
}

// DeleteFile удаляет файл с диска и запись из БД
func (s *FileUploadService) DeleteFile(ctx context.Context, attachmentID uuid.UUID) error {
	// Получаем информацию о файле
	attachment, err := s.chatRepo.GetAttachmentByID(ctx, attachmentID)
	if err != nil {
		return fmt.Errorf("failed to get attachment: %w", err)
	}

	// Удаляем файл с диска (игнорируем ошибку если файл уже удален)
	if err := os.Remove(attachment.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Пытаемся удалить пустые директории
	s.cleanupEmptyDirs(filepath.Dir(attachment.FilePath))

	// См. BACKLOG.md: Soft delete для вложений (deleted_at вместо физического удаления)

	return nil
}

// cleanupEmptyDirs удаляет пустые директории (message_id и room_id)
func (s *FileUploadService) cleanupEmptyDirs(dirPath string) {
	// Проверяем что директория пустая
	entries, err := os.ReadDir(dirPath)
	if err != nil || len(entries) > 0 {
		return
	}

	// Удаляем директорию сообщения
	os.Remove(dirPath)

	// Пытаемся удалить директорию комнаты
	roomDirPath := filepath.Dir(dirPath)
	entries, err = os.ReadDir(roomDirPath)
	if err != nil || len(entries) > 0 {
		return
	}

	os.Remove(roomDirPath)
}

// ValidateFilename проверяет безопасность имени файла
func ValidateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename is empty")
	}

	// Проверяем на path traversal атаки
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("invalid filename: contains path separators")
	}

	// Проверяем длину имени файла
	if len(filename) > 255 {
		return fmt.Errorf("filename too long (max 255 characters)")
	}

	return nil
}

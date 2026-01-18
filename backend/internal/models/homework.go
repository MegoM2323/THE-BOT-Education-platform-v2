package models

import (
	"time"

	"github.com/google/uuid"
)

// LessonHomework представляет файл домашнего задания, прикрепленный к уроку
type LessonHomework struct {
	ID          uuid.UUID `db:"id" json:"id"`
	LessonID    uuid.UUID `db:"lesson_id" json:"lesson_id"`
	FileName    string    `db:"file_name" json:"file_name"`
	FilePath    string    `db:"file_path" json:"file_path"`
	FileSize    int64     `db:"file_size" json:"file_size"`       // В байтах, макс 10MB
	MimeType    string    `db:"mime_type" json:"mime_type"`       // MIME тип файла
	TextContent string    `db:"text_content" json:"text_content"` // Текстовое описание домашнего задания
	CreatedBy   uuid.UUID `db:"created_by" json:"created_by"`     // Кто загрузил файл (teacher/admin)
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// CreateHomeworkRequest представляет запрос на создание домашнего задания
type CreateHomeworkRequest struct {
	LessonID    uuid.UUID `json:"lesson_id"`
	FileName    string    `json:"file_name"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	TextContent string    `json:"text_content"`
}

// UpdateHomeworkRequest представляет запрос на обновление описания домашнего задания
type UpdateHomeworkRequest struct {
	TextContent string `json:"text_content"`
}

// Validate выполняет валидацию UpdateHomeworkRequest
func (r *UpdateHomeworkRequest) Validate() error {
	// TextContent может быть пустым, но не nil
	if len(r.TextContent) > 10000 {
		return ErrHomeworkContentTooLong
	}
	return nil
}

// Validate выполняет валидацию CreateHomeworkRequest
func (r *CreateHomeworkRequest) Validate() error {
	// Проверка ID урока
	if r.LessonID == uuid.Nil {
		return ErrInvalidLessonID
	}

	// Проверка имени файла
	if r.FileName == "" {
		return ErrInvalidFileName
	}
	if len(r.FileName) > 255 {
		return ErrFileNameTooLong
	}

	// Проверка размера файла: > 0 и <= 10MB
	if r.FileSize <= 0 || r.FileSize > 10485760 {
		return ErrInvalidFileSize
	}

	// Проверка MIME типа
	if r.MimeType == "" {
		return ErrInvalidMimeType
	}
	if !isAllowedMimeType(r.MimeType) {
		return ErrMimeTypeNotAllowed
	}

	return nil
}

// Validate выполняет валидацию LessonHomework модели перед сохранением
func (h *LessonHomework) Validate() error {
	// Проверка ID урока
	if h.LessonID == uuid.Nil {
		return ErrInvalidLessonID
	}

	// Проверка имени файла
	if h.FileName == "" {
		return ErrInvalidFileName
	}
	if len(h.FileName) > 255 {
		return ErrFileNameTooLong
	}

	// Проверка размера файла: > 0 и <= 10MB
	if h.FileSize <= 0 || h.FileSize > 10485760 {
		return ErrInvalidFileSize
	}

	// Проверка MIME типа
	if h.MimeType == "" {
		return ErrInvalidMimeType
	}
	if !isAllowedMimeType(h.MimeType) {
		return ErrMimeTypeNotAllowed
	}

	// Проверка created_by
	if h.CreatedBy == uuid.Nil {
		return ErrInvalidUserID
	}

	return nil
}

// isAllowedMimeType проверяет, разрешен ли MIME тип для загрузки
func isAllowedMimeType(mimeType string) bool {
	allowedTypes := map[string]bool{
		// Документы
		"application/pdf":    true,
		"application/msword": true, // .doc
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // .docx
		"application/vnd.ms-excel": true, // .xls
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true, // .xlsx
		"application/vnd.ms-powerpoint":                                             true, // .ppt
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true, // .pptx
		"text/plain": true,

		// Изображения
		"image/png":  true,
		"image/jpeg": true,
		"image/jpg":  true,
		"image/gif":  true,
		"image/webp": true,

		// Видео
		"video/mp4":       true,
		"video/mpeg":      true,
		"video/webm":      true,
		"video/quicktime": true, // .mov

		// Аудио
		"audio/mpeg": true, // .mp3
		"audio/wav":  true,
		"audio/ogg":  true,

		// Архивы
		"application/zip":              true,
		"application/x-rar-compressed": true,
		"application/x-7z-compressed":  true,
	}

	return allowedTypes[mimeType]
}

// GetFileExtension возвращает расширение файла по MIME типу
func (h *LessonHomework) GetFileExtension() string {
	extensions := map[string]string{
		"application/pdf":    ".pdf",
		"application/msword": ".doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
		"image/png":  ".png",
		"image/jpeg": ".jpg",
		"text/plain": ".txt",
		"video/mp4":  ".mp4",
		"audio/mpeg": ".mp3",
	}

	if ext, ok := extensions[h.MimeType]; ok {
		return ext
	}
	return ""
}

// IsImage проверяет, является ли файл изображением
func (h *LessonHomework) IsImage() bool {
	return h.MimeType == "image/png" ||
		h.MimeType == "image/jpeg" ||
		h.MimeType == "image/jpg" ||
		h.MimeType == "image/gif" ||
		h.MimeType == "image/webp"
}

// IsVideo проверяет, является ли файл видео
func (h *LessonHomework) IsVideo() bool {
	return h.MimeType == "video/mp4" ||
		h.MimeType == "video/mpeg" ||
		h.MimeType == "video/webm" ||
		h.MimeType == "video/quicktime"
}

// IsDocument проверяет, является ли файл документом
func (h *LessonHomework) IsDocument() bool {
	return h.MimeType == "application/pdf" ||
		h.MimeType == "application/msword" ||
		h.MimeType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" ||
		h.MimeType == "text/plain"
}

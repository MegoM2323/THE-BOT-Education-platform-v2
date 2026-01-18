package models

import (
	"time"

	"github.com/google/uuid"
)

// LessonBroadcast представляет рассылку сообщений студентам урока
type LessonBroadcast struct {
	ID          uuid.UUID        `db:"id" json:"id"`
	LessonID    uuid.UUID        `db:"lesson_id" json:"lesson_id"`
	SenderID    uuid.UUID        `db:"sender_id" json:"sender_id"`
	SenderName  string           `db:"sender_name" json:"sender_name"` // Имя отправителя рассылки
	Message     string           `db:"message" json:"message"`
	Status      string           `db:"status" json:"status"`
	SentCount   int              `db:"sent_count" json:"sent_count"`
	FailedCount int              `db:"failed_count" json:"failed_count"`
	CreatedAt   time.Time        `db:"created_at" json:"created_at"`
	CompletedAt *time.Time       `db:"completed_at" json:"completed_at,omitempty"`
	Files       []*BroadcastFile `db:"-" json:"files,omitempty"`
}

// BroadcastFile представляет файловое вложение к рассылке урока
type BroadcastFile struct {
	ID          uuid.UUID `db:"id" json:"id"`
	BroadcastID uuid.UUID `db:"broadcast_id" json:"broadcast_id"`
	FileName    string    `db:"file_name" json:"file_name"`
	FilePath    string    `db:"file_path" json:"file_path"`
	FileSize    int64     `db:"file_size" json:"file_size"`
	MimeType    string    `db:"mime_type" json:"mime_type"`
	UploadedAt  time.Time `db:"uploaded_at" json:"uploaded_at"`
}

// CreateLessonBroadcastRequest представляет запрос на создание рассылки урока
type CreateLessonBroadcastRequest struct {
	LessonID uuid.UUID `json:"lesson_id" validate:"required"`
	Message  string    `json:"message" validate:"required,min=1,max=4096"`
}

// Константы статусов рассылки урока
const (
	LessonBroadcastStatusPending   = "pending"
	LessonBroadcastStatusSending   = "sending"
	LessonBroadcastStatusCompleted = "completed"
	LessonBroadcastStatusFailed    = "failed"
)

// Константы ограничений
const (
	MaxBroadcastFiles      = 10       // Максимум 10 файлов на рассылку (лимит Telegram API)
	MaxBroadcastFileSize   = 10485760 // 10MB в байтах
	MaxBroadcastMessageLen = 4096     // Максимальная длина сообщения
	MinBroadcastMessageLen = 1        // Минимальная длина сообщения
)

// Validate выполняет валидацию LessonBroadcast
func (lb *LessonBroadcast) Validate() error {
	// Валидация LessonID
	if lb.LessonID == uuid.Nil {
		return ErrInvalidLessonID
	}

	// Валидация SenderID
	if lb.SenderID == uuid.Nil {
		return ErrInvalidUserID
	}

	// Валидация Message
	if len(lb.Message) < MinBroadcastMessageLen {
		return ErrInvalidBroadcastMessage
	}
	if len(lb.Message) > MaxBroadcastMessageLen {
		return ErrBroadcastMessageTooLong
	}

	// Валидация Status
	if !isValidBroadcastStatus(lb.Status) {
		return ErrInvalidBroadcastStatus
	}

	// Валидация файлов
	if len(lb.Files) > MaxBroadcastFiles {
		return ErrTooManyFiles
	}

	return nil
}

// Validate выполняет валидацию BroadcastFile
func (bf *BroadcastFile) Validate() error {
	// Валидация BroadcastID
	if bf.BroadcastID == uuid.Nil {
		return ErrInvalidBroadcastID
	}

	// Валидация FileName
	if bf.FileName == "" {
		return ErrInvalidFileName
	}

	// Валидация FileSize
	if bf.FileSize <= 0 || bf.FileSize > MaxBroadcastFileSize {
		return ErrInvalidFileSize
	}

	// Валидация MimeType
	if bf.MimeType == "" {
		return ErrInvalidMimeType
	}

	return nil
}

// Validate выполняет валидацию CreateLessonBroadcastRequest
func (r *CreateLessonBroadcastRequest) Validate() error {
	// Валидация LessonID
	if r.LessonID == uuid.Nil {
		return ErrInvalidLessonID
	}

	// Валидация Message
	if len(r.Message) < MinBroadcastMessageLen {
		return ErrInvalidBroadcastMessage
	}
	if len(r.Message) > MaxBroadcastMessageLen {
		return ErrBroadcastMessageTooLong
	}

	return nil
}

// isValidBroadcastStatus проверяет корректность статуса рассылки
func isValidBroadcastStatus(status string) bool {
	switch status {
	case LessonBroadcastStatusPending,
		LessonBroadcastStatusSending,
		LessonBroadcastStatusCompleted,
		LessonBroadcastStatusFailed:
		return true
	default:
		return false
	}
}

// IsPending проверяет, находится ли рассылка в статусе ожидания
func (lb *LessonBroadcast) IsPending() bool {
	return lb.Status == LessonBroadcastStatusPending
}

// IsSending проверяет, отправляется ли рассылка в данный момент
func (lb *LessonBroadcast) IsSending() bool {
	return lb.Status == LessonBroadcastStatusSending
}

// IsCompleted проверяет, завершена ли рассылка
func (lb *LessonBroadcast) IsCompleted() bool {
	return lb.Status == LessonBroadcastStatusCompleted
}

// IsFailed проверяет, провалилась ли рассылка
func (lb *LessonBroadcast) IsFailed() bool {
	return lb.Status == LessonBroadcastStatusFailed
}

// IsFinalStatus проверяет, находится ли рассылка в финальном статусе
func (lb *LessonBroadcast) IsFinalStatus() bool {
	return lb.IsCompleted() || lb.IsFailed()
}
